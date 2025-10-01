package mikrotik

import (
	"context"
	"fmt"
	"sort"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
	"sigs.k8s.io/external-dns/provider"
)

// DNS Provider for working with mikrotik
type MikrotikProvider struct {
	provider.BaseProvider

	client       *MikrotikApiClient
	domainFilter *endpoint.DomainFilter
}

// NewMikrotikProvider initializes a new DNSProvider, of the Mikrotik variety
func NewMikrotikProvider(domainFilter *endpoint.DomainFilter, defaults *MikrotikDefaults, config *MikrotikConnectionConfig) (provider.Provider, error) {
	// Create the Mikrotik API Client
	client, err := NewMikrotikClient(config, defaults)
	if err != nil {
		return nil, fmt.Errorf("failed to create the MikroTik client: %w", err)
	}

	// Ensure the Client can connect to the API by fetching system info
	info, err := client.GetSystemInfo()
	if err != nil {
		log.Errorf("failed to connect to the MikroTik RouterOS API Endpoint: %v", err)
		return nil, err
	}
	log.Infof("connected to board %s running RouterOS version %s (%s)", info.BoardName, info.Version, info.ArchitectureName)

	// If the client connects properly, create the DNS Provider
	p := &MikrotikProvider{
		client:       client,
		domainFilter: domainFilter,
	}

	return p, nil
}

// Records returns the list of all DNS records.
func (p *MikrotikProvider) Records(ctx context.Context) ([]*endpoint.Endpoint, error) {
	// Get all managed records (no name filter)
	records, err := p.client.GetDNSRecordsByNameAndType("", "")
	if err != nil {
		return nil, err
	}

	// Use the new aggregation logic to combine multiple records into endpoints
	endpoints, err := AggregateRecordsToEndpoints(records, p.client.DefaultComment)
	if err != nil {
		log.Errorf("Failed to aggregate DNS records to endpoints: %v", err)
		return nil, err
	}

	// Filter endpoints by domain filter
	var filteredEndpoints []*endpoint.Endpoint
	for _, ep := range endpoints {
		if !p.domainFilter.Match(ep.DNSName) {
			continue
		}

		filteredEndpoints = append(filteredEndpoints, ep)
	}

	log.Debugf("Returned %d endpoints after domain filtering", len(filteredEndpoints))
	return filteredEndpoints, nil
}

// ApplyChanges applies a given set of changes in the DNS provider.
func (p *MikrotikProvider) ApplyChanges(ctx context.Context, changes *plan.Changes) error {
	changes = p.changes(changes)

	// SECURITY: Verify all endpoints are within allowed domain scope before making any changes
	allEndpoints := append(changes.UpdateOld, changes.Delete...)
	allEndpoints = append(allEndpoints, changes.Create...)
	allEndpoints = append(allEndpoints, changes.UpdateNew...)

	for _, endpoint := range allEndpoints {
		if !p.domainFilter.Match(endpoint.DNSName) {
			log.Errorf("SECURITY: Attempted to manage DNS record outside allowed domain scope: %s", endpoint.DNSName)
			return fmt.Errorf("security violation: DNS name '%s' is not within allowed domain filter", endpoint.DNSName)
		}
	}

	// Delete endpoints (Delete only - handle Updates separately) - now safe after domain verification
	for _, endpoint := range changes.Delete {
		if err := p.client.DeleteDNSRecords(endpoint); err != nil {
			log.Errorf("Failed to delete DNS records for endpoint %s: %v", endpoint.DNSName, err)
			return err
		}
	}

	// Handle Updates with smart differential updates
	if len(changes.UpdateOld) > 0 || len(changes.UpdateNew) > 0 {
		// Create maps for matching old and new endpoints by DNS name and record type
		oldEndpoints := make(map[string]*endpoint.Endpoint)
		newEndpoints := make(map[string]*endpoint.Endpoint)

		// Build map of old endpoints
		for _, oldEndpoint := range changes.UpdateOld {
			key := fmt.Sprintf("%s:%s", oldEndpoint.DNSName, oldEndpoint.RecordType)
			oldEndpoints[key] = oldEndpoint
		}

		// Build map of new endpoints
		for _, newEndpoint := range changes.UpdateNew {
			key := fmt.Sprintf("%s:%s", newEndpoint.DNSName, newEndpoint.RecordType)
			newEndpoints[key] = newEndpoint
		}

		// Find matched pairs for smart updates
		processedKeys := make(map[string]bool)

		// Process matched pairs with smart updates
		for key, oldEndpoint := range oldEndpoints {
			if newEndpoint, exists := newEndpoints[key]; exists {
				if err := p.smartUpdateEndpoint(oldEndpoint, newEndpoint); err != nil {
					log.Errorf("Failed to update DNS records for endpoint %s: %v", newEndpoint.DNSName, err)
					return err
				}
				processedKeys[key] = true
			}
		}

		// Delete unmatched old endpoints
		for key, oldEndpoint := range oldEndpoints {
			if !processedKeys[key] {
				if err := p.client.DeleteDNSRecords(oldEndpoint); err != nil {
					log.Errorf("Failed to delete unmatched old DNS records for endpoint %s: %v", oldEndpoint.DNSName, err)
					return err
				}
			}
		}

		// Create unmatched new endpoints
		for key, newEndpoint := range newEndpoints {
			if !processedKeys[key] {
				_, err := p.client.CreateDNSRecords(newEndpoint)
				if err != nil {
					log.Errorf("Failed to create unmatched new DNS records for endpoint %s: %v", newEndpoint.DNSName, err)
					return err
				}
			}
		}
	}

	// Create new endpoints - now safe after domain verification
	for _, endpoint := range changes.Create {
		_, err := p.client.CreateDNSRecords(endpoint)
		if err != nil {
			log.Errorf("Failed to create DNS records for endpoint %s: %v", endpoint.DNSName, err)
			return err
		}
	}

	return nil
}

// smartUpdateEndpoint performs differential updates, only modifying changed targets
func (p *MikrotikProvider) smartUpdateEndpoint(oldEndpoint, newEndpoint *endpoint.Endpoint) error {
	log.Debugf("Smart update: comparing old endpoint %s with new endpoint", oldEndpoint.DNSName)

	// Build maps of old and new targets
	oldTargets := make(map[string]bool) // target -> exists
	for _, target := range oldEndpoint.Targets {
		oldTargets[target] = true
	}

	newTargets := make(map[string]bool) // target -> exists
	for _, target := range newEndpoint.Targets {
		newTargets[target] = true
	}

	log.Debugf("Old targets: %v, New targets: %v", oldEndpoint.Targets, newEndpoint.Targets)

	// Find targets to delete (in old but not in new)
	var toDelete []string
	for target := range oldTargets {
		if !newTargets[target] {
			toDelete = append(toDelete, target)
		}
	}

	// Find targets to add (in new but not in old)
	var toAdd []string
	for target := range newTargets {
		if !oldTargets[target] {
			toAdd = append(toAdd, target)
		}
	}

	log.Infof("Smart update for %s: %d targets to delete, %d targets to add", newEndpoint.DNSName, len(toDelete), len(toAdd))

	// Delete obsolete targets using batch deletion
	if len(toDelete) > 0 {
		// Create a temporary endpoint for batch deletion of obsolete targets
		deleteEndpoint := &endpoint.Endpoint{
			DNSName:    newEndpoint.DNSName,
			RecordType: newEndpoint.RecordType,
			Targets:    toDelete,
		}

		log.Debugf("Batch deleting %d obsolete targets for %s", len(toDelete), newEndpoint.DNSName)
		err := p.client.DeleteDNSRecords(deleteEndpoint)
		if err != nil {
			return fmt.Errorf("failed to batch delete obsolete targets for %s: %w", newEndpoint.DNSName, err)
		}
	}

	// Add new targets
	if len(toAdd) > 0 {
		// Create a new endpoint with only the new targets
		addEndpoint := &endpoint.Endpoint{
			DNSName:          newEndpoint.DNSName,
			RecordType:       newEndpoint.RecordType,
			RecordTTL:        newEndpoint.RecordTTL,
			Targets:          toAdd,
			ProviderSpecific: newEndpoint.ProviderSpecific,
		}

		_, err := p.client.CreateDNSRecords(addEndpoint)
		if err != nil {
			return fmt.Errorf("failed to create new targets: %w", err)
		}
	}

	return nil
}

// GetDomainFilter returns the domain filter for the provider.
func (p *MikrotikProvider) GetDomainFilter() endpoint.DomainFilterInterface {
	return p.domainFilter
}

// ================================================================================================
// UTILS
// ================================================================================================
// getProviderSpecific retrieves a provider-specific property from the endpoint, looking both values
// that could come from annotations (i.e. webhook/%s) as well as values from CRD (i.e. %s).
// If the property is not found, it returns the specified default value.
func (p *MikrotikProvider) getProviderSpecificOrDefault(ep *endpoint.Endpoint, ps string, defaultValue string) string {
	value, valueExists := ep.GetProviderSpecificProperty(ps)
	if valueExists && value != "" {
		log.Debugf("Found provider-specific property '%s' with value: %s", ps, value)
		return value
	}

	log.Debugf("Provider-specific property '%s' not found, checking for webhook/%s", ps, ps)
	value, valueExists = ep.GetProviderSpecificProperty(fmt.Sprintf("webhook/%s", ps))
	if valueExists && value != "" {
		log.Debugf("Found provider-specific property 'webhook/%s' with value: %s", ps, value)
		return value
	}

	log.Debugf("Property '%s' not found, returning default value: %s", ps, defaultValue)
	return defaultValue
}

// compareEndpoints compares two endpoints to determine if they are identical, keeping in mind empty/default states.
func (p *MikrotikProvider) compareEndpoints(a *endpoint.Endpoint, b *endpoint.Endpoint) bool {
	log.Debugf("Comparing endpoint a: %v", a)
	log.Debugf("Against endpoint b: %v", b)

	if a.DNSName != b.DNSName {
		log.Debugf("DNSName mismatch: %v != %v", a.DNSName, b.DNSName)
		return false
	}

	if a.RecordType != b.RecordType {
		log.Debugf("RecordType mismatch: %v != %v", a.RecordType, b.RecordType)
		return false
	}

	// Compare all targets, not just the first one
	if len(a.Targets) != len(b.Targets) {
		log.Debugf("Targets length mismatch: %d != %d", len(a.Targets), len(b.Targets))
		return false
	}

	// Create sorted copies of targets for comparison
	aTargets := make([]string, len(a.Targets))
	bTargets := make([]string, len(b.Targets))
	copy(aTargets, a.Targets)
	copy(bTargets, b.Targets)

	// Sort targets to ensure consistent comparison
	sort.Strings(aTargets)
	sort.Strings(bTargets)

	// Compare sorted targets
	for i := 0; i < len(aTargets); i++ {
		if aTargets[i] != bTargets[i] {
			log.Debugf("Target[%d] mismatch: %v != %v", i, aTargets[i], bTargets[i])
			return false
		}
	}

	aRelevantTTL := a.RecordTTL != 0 && a.RecordTTL != endpoint.TTL(p.client.DefaultTTL)
	bRelevantTTL := b.RecordTTL != 0 && b.RecordTTL != endpoint.TTL(p.client.DefaultTTL)
	if a.RecordTTL != b.RecordTTL && (aRelevantTTL || bRelevantTTL) {
		log.Debugf("RecordTTL mismatch: %v != %v", a.RecordTTL, b.RecordTTL)
		return false
	}

	aComment := p.getProviderSpecificOrDefault(a, "comment", "")
	bComment := p.getProviderSpecificOrDefault(b, "comment", "")
	aRelevantComment := aComment != "" && aComment != p.client.DefaultComment
	bRelevantComment := bComment != "" && bComment != p.client.DefaultComment
	if aComment != bComment && (aRelevantComment || bRelevantComment) {
		log.Debugf("Comment mismatch: %v != %v", aComment, bComment)
		return false
	}

	aMatchSubdomain := p.getProviderSpecificOrDefault(a, "match-subdomain", "false")
	bMatchSubdomain := p.getProviderSpecificOrDefault(b, "match-subdomain", "false")
	if aMatchSubdomain != bMatchSubdomain {
		log.Debugf("MatchSubdomain mismatch: %v != %v", aMatchSubdomain, bMatchSubdomain)
		return false
	}

	aDisabled := p.getProviderSpecificOrDefault(a, "disabled", "false")
	bDisabled := p.getProviderSpecificOrDefault(b, "disabled", "false")
	if aDisabled != bDisabled {
		log.Debugf("Disabled mismatch: %v != %v", aDisabled, bDisabled)
		return false
	}

	aAddressList := p.getProviderSpecificOrDefault(a, "address-list", "")
	bAddressList := p.getProviderSpecificOrDefault(b, "address-list", "")
	if aAddressList != bAddressList {
		log.Debugf("AddressList mismatch: %v != %v", aAddressList, bAddressList)
		return false
	}

	aRegexp := p.getProviderSpecificOrDefault(a, "regexp", "")
	bRegexp := p.getProviderSpecificOrDefault(b, "regexp", "")
	if aRegexp != bRegexp {
		log.Debugf("Regexp mismatch: %v != %v", aRegexp, bRegexp)
		return false
	}

	log.Debugf("Endpoints match successfully.")
	return true
}

func (p *MikrotikProvider) listContains(haystack []*endpoint.Endpoint, needle *endpoint.Endpoint) bool {
	for _, v := range haystack {
		if p.compareEndpoints(needle, v) {
			return true
		}
	}
	return false
}

// changes processes and filters the changes plan for updates.
// It adjusts TTL for created endpoints and removes duplicate updates from the plan.
func (p *MikrotikProvider) changes(changes *plan.Changes) *plan.Changes {
	log.Debug("Starting to process changes plan.")

	// Initialize new plan -> we don't really need to worry about Delete changes.
	// Only updates are sketchy
	newChanges := &plan.Changes{
		Create:    []*endpoint.Endpoint{},
		Delete:    changes.Delete,
		UpdateOld: []*endpoint.Endpoint{},
		UpdateNew: []*endpoint.Endpoint{},
	}

	log.Debugf("Initial changes - Create: %d, Delete: %d, UpdateOld: %d, UpdateNew: %d", len(changes.Create), len(changes.Delete), len(changes.UpdateOld), len(changes.UpdateNew))

	// Process Create changes
	for _, create := range changes.Create {
		// Enforce Default TTL
		if !create.RecordTTL.IsConfigured() {
			log.Debugf("Setting default TTL for created endpoint: %v", create)
			create.RecordTTL = endpoint.TTL(p.client.DefaultTTL)
		}

		// Enforce Default Comment
		if p.client.DefaultComment != "" {
			if p.getProviderSpecificOrDefault(create, "comment", "") == "" {
				log.Debugf("Setting default comment for created endpoint: %v", create)
				create.SetProviderSpecificProperty("comment", p.client.DefaultComment)
			}
		}

		newChanges.Create = append(newChanges.Create, create)
	}

	// Identify duplicates in Update changes
	duplicates := []*endpoint.Endpoint{}
	for _, old := range changes.UpdateOld {
		for _, new := range changes.UpdateNew {
			if p.compareEndpoints(old, new) {
				log.Debugf("Found duplicate update for endpoint: %v", old)
				duplicates = append(duplicates, old)
			}
		}
	}

	// Filter out duplicates from UpdateOld
	for _, old := range changes.UpdateOld {
		if !p.listContains(duplicates, old) {
			log.Debugf("Adding non-duplicate UpdateOld endpoint: %v", old)
			newChanges.UpdateOld = append(newChanges.UpdateOld, old)
		}
	}

	// Filter out duplicates from UpdateNew
	for _, new := range changes.UpdateNew {
		if !p.listContains(duplicates, new) {
			log.Debugf("Adding non-duplicate UpdateNew endpoint: %v", new)

			// Enforce Default TTL
			if !new.RecordTTL.IsConfigured() {
				log.Debugf("Setting default TTL for UpdateNew endpoint: %v", new)
				new.RecordTTL = endpoint.TTL(p.client.DefaultTTL)
			}

			// Enforce Default Comment
			if p.client.DefaultComment != "" {
				if p.getProviderSpecificOrDefault(new, "comment", "") == "" {
					log.Debugf("Setting default comment for UpdateNew endpoint: %v", new)
					new.SetProviderSpecificProperty("comment", p.client.DefaultComment)
				}
			}

			newChanges.UpdateNew = append(newChanges.UpdateNew, new)
		}
	}

	log.Debugf("Processed changes - Create: %d, Delete: %d, UpdateOld: %d, UpdateNew: %d", len(newChanges.Create), len(newChanges.Delete), len(newChanges.UpdateOld), len(newChanges.UpdateNew))
	log.Debug("Finished processing changes plan.")
	return newChanges
}
