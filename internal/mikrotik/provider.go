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

	// Aggregate records to endpoints
	endpoints, err := p.aggregateRecordsToEndpoints(filteredRecords)
	if err != nil {
		log.Errorf("Failed to aggregate DNS records to endpoints: %v", err)
		return nil, err
	}

	return endpoints, nil
}

// ApplyChanges applies a given set of changes in the DNS provider.
func (p *MikrotikProvider) ApplyChanges(ctx context.Context, changes *plan.Changes) error {
	changes, err := p.filterChanges(changes)
	if err != nil {
		return fmt.Errorf("failed to process changes: %w", err)
	}

	for _, endpoint := range append(changes.UpdateOld, changes.Delete...) {
		if err := p.client.DeleteRecordsFromEndpoint(endpoint); err != nil {
			return err
		}
	}

	for _, endpoint := range append(changes.Create, changes.UpdateNew...) {
		if _, err := p.client.CreateRecordsFromEndpoint(endpoint); err != nil {
			return err
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
func (p *MikrotikProvider) compareEndpointsMetadata(a *endpoint.Endpoint, b *endpoint.Endpoint) bool {
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
func (p *MikrotikProvider) filterManagedRecords(records []DNSRecord) []DNSRecord {
	if p.domainFilter == nil {
		log.Debug("No domain filter set, returning all records")
		return records
	}

	var filtered []DNSRecord
	for _, record := range records {
		if !p.domainFilter.Match(record.Name) {
			log.Debugf("Skipping record %s as it does not match domain filter", record.Name)
			continue
		}
		filtered = append(filtered, record)
	}

	log.Debugf("Filtered records: %d out of %d match domain filter", len(filtered), len(records))
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
	log.Debugf("Grouped records into %d groups", len(recordGroups))

	// Convert each group to an endpoint
	var endpoints []*endpoint.Endpoint
	for groupKey, groupRecords := range recordGroups {
		log.Debugf("Processing group %s with %d records", groupKey, len(groupRecords))

		// Use the first record as the template for the base endpoint
		template := groupRecords[0]

		ttl, err := MikrotikTTLtoEndpointTTL(template.TTL)
		if err != nil {
			return nil, fmt.Errorf("invalid TTL in record group %s: %w", groupKey, err)
		}

		baseEndpoint := &endpoint.Endpoint{
			DNSName:    template.Name,
			RecordType: template.Type,
			RecordTTL:  ttl,
		}

		// Add provider-specific properties from the template
		if template.Comment != "" {
			baseEndpoint.ProviderSpecific = append(
				baseEndpoint.ProviderSpecific,
				endpoint.ProviderSpecificProperty{Name: "comment", Value: template.Comment},
			)
		}
		if template.Disabled != "" {
			baseEndpoint.ProviderSpecific = append(
				baseEndpoint.ProviderSpecific,
				endpoint.ProviderSpecificProperty{Name: "disabled", Value: template.Disabled},
			)
		}
		if template.Regexp != "" {
			baseEndpoint.ProviderSpecific = append(
				baseEndpoint.ProviderSpecific,
				endpoint.ProviderSpecificProperty{Name: "regexp", Value: template.Regexp},
			)
		}
		if template.MatchSubdomain != "" {
			baseEndpoint.ProviderSpecific = append(
				baseEndpoint.ProviderSpecific,
				endpoint.ProviderSpecificProperty{Name: "match-subdomain", Value: template.MatchSubdomain},
			)
		}
		if template.AddressList != "" {
			baseEndpoint.ProviderSpecific = append(
				baseEndpoint.ProviderSpecific,
				endpoint.ProviderSpecificProperty{Name: "address-list", Value: template.AddressList},
			)
		}

		// Aggregate all targets from the records in the group
		var targets []string
		for _, record := range groupRecords {
			target, err := record.toExternalDNSTarget()
			if err != nil {
				return nil, fmt.Errorf("failed to convert record %+v to target: %w", record, err)
			}
			targets = append(targets, target)
		}

		baseEndpoint.Targets = targets
		log.Debugf("Created endpoint for group %s: %+v", groupKey, baseEndpoint)

		endpoints = append(endpoints, baseEndpoint)
	}

	log.Debugf("Aggregated %d record groups to %d endpoints", len(recordGroups), len(endpoints))
	return endpoints, nil
}

// filterChanges processes the given plan.Changes to optimize updates by splitting them into deletes and creates
func (p *MikrotikProvider) filterChanges(changes *plan.Changes) (*plan.Changes, error) {
	if len(changes.UpdateOld) == 0 || len(changes.UpdateNew) == 0 {
		return changes, nil
	}

	if len(changes.UpdateOld) != len(changes.UpdateNew) {
		return nil, fmt.Errorf("mismatched UpdateOld and UpdateNew lengths: %d vs %d", len(changes.UpdateOld), len(changes.UpdateNew))
	}

	newChanges := &plan.Changes{
		Create:    changes.Create,
		UpdateOld: []*endpoint.Endpoint{},
		UpdateNew: []*endpoint.Endpoint{},
		Delete:    changes.Delete,
	}

	// Process matched pairs (updateOld and updateNew are matched by index)
	for key, oldEndpoint := range changes.UpdateOld {
		newEndpoint := changes.UpdateNew[key]

		// sanity check: name and type must match
		if oldEndpoint.DNSName != newEndpoint.DNSName || oldEndpoint.RecordType != newEndpoint.RecordType {
			return nil, fmt.Errorf("mismatched UpdateOld and UpdateNew endpoints at index %d: %v vs %v", key, oldEndpoint, newEndpoint)
		}

		// if metadata matches, do partial update
		if p.compareEndpointsMetadata(oldEndpoint, newEndpoint) {
			deleteEndpoint, createEndpoint := p.diffEndpoints(oldEndpoint, newEndpoint)
			if deleteEndpoint != nil {
				newChanges.Delete = append(newChanges.Delete, deleteEndpoint)
			}
			if createEndpoint != nil {
				newChanges.Create = append(newChanges.Create, createEndpoint)
			}
		} else {
			newChanges.Delete = append(newChanges.Delete, oldEndpoint)
			newChanges.Create = append(newChanges.Create, newEndpoint)
		}
	}

	return newChanges, nil
}

// diffEndpoints computes the difference between two endpoints and returns two new endpoints:
// one for targets to delete and one for targets to add. If there are no targets to delete or add,
// the corresponding endpoint will be nil.
func (p *MikrotikProvider) diffEndpoints(oldEndpoint, newEndpoint *endpoint.Endpoint) (*endpoint.Endpoint, *endpoint.Endpoint) {
	// Build maps of old and new targets
	oldTargets := make(map[string]bool) // target -> exists
	for _, target := range oldEndpoint.Targets {
		oldTargets[target] = true
	}
	log.Debugf("Old targets: %v", oldEndpoint.Targets)

	newTargets := make(map[string]bool) // target -> exists
	for _, target := range newEndpoint.Targets {
		newTargets[target] = true
	}
	log.Debugf("New targets: %v", newEndpoint.Targets)

	// Find targets to delete (in old but not in new)
	var toDelete []string
	for target := range oldTargets {
		if !newTargets[target] {
			toDelete = append(toDelete, target)
		}
	}
	log.Debugf("Targets to delete: %v", toDelete)

	// Find targets to add (in new but not in old)
	var toAdd []string
	for target := range newTargets {
		if !oldTargets[target] {
			toAdd = append(toAdd, target)
		}
	}
	log.Debugf("Targets to add: %v", toAdd)

	endpointToDelete := &endpoint.Endpoint{
		DNSName:    oldEndpoint.DNSName,
		RecordType: oldEndpoint.RecordType,
		Targets:    toDelete,
	}
	if len(toDelete) == 0 {
		log.Debug("No targets to delete, returning nil for delete endpoint")
		endpointToDelete = nil
	}

	endpointToAdd := &endpoint.Endpoint{
		DNSName:    newEndpoint.DNSName,
		RecordType: newEndpoint.RecordType,
		Targets:    toAdd,
	}
	if len(toAdd) == 0 {
		log.Debug("No targets to add, returning nil for add endpoint")
		endpointToAdd = nil
	}

	return endpointToDelete, endpointToAdd
}
