package mikrotik

import (
	"context"
	"fmt"

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
	records, err := p.client.GetDNSRecords(DNSRecordFilter{})
	if err != nil {
		return nil, err
	}

	// Filter managed records
	filteredRecords := p.filterManagedRecords(records)
	log.Debugf("Filtered down to %d managed DNS records", len(filteredRecords))

	// Aggregate records to endpoints
	endpoints, err := p.aggregateRecordsToEndpoints(filteredRecords)
	if err != nil {
		log.Errorf("Failed to aggregate DNS records to endpoints: %v", err)
		return nil, err
	}

	log.Debugf("Returned %d endpoints after domain filtering", len(endpoints))
	return endpoints, nil
}

// ApplyChanges applies a given set of changes in the DNS provider.
func (p *MikrotikProvider) ApplyChanges(ctx context.Context, changes *plan.Changes) error {
	// Create new endpoints
	for _, endpoint := range changes.Create {
		_, err := p.client.CreateRecordsFromEndpoint(endpoint)
		if err != nil {
			log.Errorf("Failed to create DNS records for endpoint %s: %v", endpoint.DNSName, err)
			return err
		}
	}

	// Delete endpoints (Delete only - handle Updates separately)
	for _, endpoint := range changes.Delete {
		if err := p.client.DeleteRecordsFromEndpoint(endpoint); err != nil {
			log.Errorf("Failed to delete DNS records for endpoint %s: %v", endpoint.DNSName, err)
			return err
		}
	}

	// We assume that UpdateOld and UpdateNew are aligned by index.
	if len(changes.UpdateOld) > 0 || len(changes.UpdateNew) > 0 {

		if len(changes.UpdateOld) != len(changes.UpdateNew) {
			log.Errorf("Mismatched UpdateOld and UpdateNew lengths: %d vs %d", len(changes.UpdateOld), len(changes.UpdateNew))
			return fmt.Errorf("mismatched UpdateOld and UpdateNew lengths: %d vs %d", len(changes.UpdateOld), len(changes.UpdateNew))
		}

		// Process matched pairs with smart updates
		for key, oldEndpoint := range changes.UpdateOld {
			newEndpoint := changes.UpdateNew[key]
			// check name and type for sanity
			if oldEndpoint.DNSName != newEndpoint.DNSName || oldEndpoint.RecordType != newEndpoint.RecordType {
				log.Errorf("Mismatched UpdateOld and UpdateNew endpoints at index %d: %v vs %v", key, oldEndpoint, newEndpoint)
				return fmt.Errorf("mismatched UpdateOld and UpdateNew endpoints at index %d: %v vs %v", key, oldEndpoint, newEndpoint)
			}
			// if metadata are same, do smart update
			if p.compareEndpointsBesidesTargets(oldEndpoint, newEndpoint) {
				log.Infof("Performing smart update for endpoint %s", newEndpoint.DNSName)
				if err := p.smartUpdateEndpoint(oldEndpoint, newEndpoint); err != nil {
					log.Errorf("Failed to update DNS records for endpoint %s: %v", newEndpoint.DNSName, err)
					return err
				}
			} else {
				log.Infof("Performing full replacement update for endpoint %s", newEndpoint.DNSName)
				// Full replacement: delete old and create new
				if err := p.client.DeleteRecordsFromEndpoint(oldEndpoint); err != nil {
					log.Errorf("Failed to delete DNS records for endpoint %s during update: %v", oldEndpoint.DNSName, err)
					return err
				}
				_, err := p.client.CreateRecordsFromEndpoint(newEndpoint)
				if err != nil {
					log.Errorf("Failed to create DNS records for endpoint %s during update: %v", newEndpoint.DNSName, err)
					return err
				}
			}
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
		err := p.client.DeleteRecordsFromEndpoint(deleteEndpoint)
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

		_, err := p.client.CreateRecordsFromEndpoint(addEndpoint)
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
func (p *MikrotikProvider) compareEndpointsBesidesTargets(a *endpoint.Endpoint, b *endpoint.Endpoint) bool {
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

// filterManagedRecords filters DNS records based on the provider's domain filter.
// TODO: filter by other criteria if needed (e.g., comment prefix)
func (p *MikrotikProvider) filterManagedRecords(records []DNSRecord) []DNSRecord {
	var filtered []DNSRecord
	for _, record := range records {
		if p.domainFilter != nil && !p.domainFilter.Match(record.Name) {
			log.Debugf("Skipping record %s as it does not match domain filter", record.Name)
			continue
		}
		filtered = append(filtered, record)
	}
	return filtered
}

// aggregateRecordsToEndpoints groups DNS records by common properties and converts them to ExternalDNS endpoints
func (p *MikrotikProvider) aggregateRecordsToEndpoints(records []DNSRecord) ([]*endpoint.Endpoint, error) {
	log.Debugf("Aggregating %d DNS records to endpoints", len(records))

	// Group records by all fields that should be identical for aggregation
	recordGroups := make(map[string][]*DNSRecord)
	for i := range records {
		record := &records[i]

		// Group by all fields that should be identical for aggregation
		groupKey := fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s:%s",
			record.Name, record.Type, record.TTL, record.Comment,
			record.Regexp, record.MatchSubdomain, record.AddressList, record.Disabled)

		recordGroups[groupKey] = append(recordGroups[groupKey], record)
		log.Debugf("Added record %s (ID: %s) to group %s", record.Name, record.ID, groupKey)
	}

	var endpoints []*endpoint.Endpoint
	for groupKey, groupRecords := range recordGroups {
		if len(groupRecords) == 0 {
			continue
		}

		// Use the first record as the template for the base endpoint
		template := groupRecords[0]
		ttl, err := MikrotikTTLtoEndpointTTL(template.TTL)
		if err != nil {
			log.Warnf("Failed to convert TTL for group %s: %v", groupKey, err)
			continue
		}

		baseEndpoint := &endpoint.Endpoint{
			DNSName:    template.Name,
			RecordType: template.Type,
			RecordTTL:  ttl,
		}

		// Add provider-specific properties from the template
		// Do not pass default values to avoid unnecessary updates
		if template.Comment != "" && template.Comment != p.client.DefaultComment {
			baseEndpoint.ProviderSpecific = append(baseEndpoint.ProviderSpecific, endpoint.ProviderSpecificProperty{Name: "comment", Value: template.Comment})
		}
		if template.Disabled != "" && template.Disabled != "false" {
			baseEndpoint.ProviderSpecific = append(baseEndpoint.ProviderSpecific, endpoint.ProviderSpecificProperty{Name: "disabled", Value: template.Disabled})
		}
		if template.Regexp != "" {
			baseEndpoint.ProviderSpecific = append(baseEndpoint.ProviderSpecific, endpoint.ProviderSpecificProperty{Name: "regexp", Value: template.Regexp})
		}
		if template.MatchSubdomain != "" && template.MatchSubdomain != "false" {
			baseEndpoint.ProviderSpecific = append(baseEndpoint.ProviderSpecific, endpoint.ProviderSpecificProperty{Name: "match-subdomain", Value: template.MatchSubdomain})
		}
		if template.AddressList != "" {
			baseEndpoint.ProviderSpecific = append(baseEndpoint.ProviderSpecific, endpoint.ProviderSpecificProperty{Name: "address-list", Value: template.AddressList})
		}

		// Aggregate all targets from the records in the group
		var targets []string
		for _, record := range groupRecords {
			target, err := record.toExternalDNSTarget()
			if err != nil {
				log.Warnf("Failed to convert record %s to target for group %s: %v", record.ID, groupKey, err)
				continue
			}
			targets = append(targets, target)
		}

		if len(targets) == 0 {
			log.Warnf("No valid targets found for group %s", groupKey)
			continue
		}

		baseEndpoint.Targets = targets
		endpoints = append(endpoints, baseEndpoint)
	}

	log.Debugf("Aggregated %d record groups to %d endpoints", len(recordGroups), len(endpoints))
	return endpoints, nil
}
