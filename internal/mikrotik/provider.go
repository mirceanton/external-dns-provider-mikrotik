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
	domainFilter endpoint.DomainFilter
}

// NewMikrotikProvider initializes a new DNSProvider.
func NewMikrotikProvider(domainFilter endpoint.DomainFilter, config *Config) (provider.Provider, error) {
	c, err := NewMikrotikClient(config)

	if err != nil {
		return nil, fmt.Errorf("failed to create the MikroTik client: %w", err)
	}

	p := &MikrotikProvider{
		client:       c,
		domainFilter: domainFilter,
	}

	return p, nil
}

// Records returns the list of all DNS records.
func (p *MikrotikProvider) Records(ctx context.Context) ([]*endpoint.Endpoint, error) {
	records, err := p.client.GetAllDNSRecords()
	if err != nil {
		return nil, err
	}

	var endpoints []*endpoint.Endpoint
	for _, record := range records {
		ep, _ := record.toExternalDNSEndpoint()

		if !p.domainFilter.Match(ep.DNSName) {
			continue
		}

		endpoints = append(endpoints, ep)
	}

	return endpoints, nil
}

// When a new DNSEndpoint is created, it can omit the ProviderSpecific field "disabled".
// However, when the same DNSEndpoint is updated, the "disabled" field is set to "false"
// by default by the Mikrotik API.
// This causes the same endpoint to be deleted and recreated constantly.
// This function checks the `updateNew` and `updateOld` fields of the `plan.Changes` struct and
// removes the `updateNew` entries that are the same as the `updateOld` entries, just with the disabled
// field set to false instead of it being omitted.
func cleanupChanges(changes *plan.Changes) *plan.Changes {
	index := 0

	for _, newEndpoint := range changes.UpdateNew {
		for _, oldEndpoint := range changes.UpdateOld {
			if newEndpoint.DNSName == oldEndpoint.DNSName && newEndpoint.RecordType == oldEndpoint.RecordType {
				newDisabled, _ := newEndpoint.GetProviderSpecificProperty("disabled")
				oldDisabled, _ := oldEndpoint.GetProviderSpecificProperty("disabled")
				if newDisabled == "false" && oldDisabled == "" {
					log.Debugf("Ignoring update for %s because it is disabled", newEndpoint.DNSName)
					changes.UpdateNew = append(changes.UpdateNew[:index], changes.UpdateNew[index+1:]...)
				}
			}
		}
		index++
	}

	return changes
}

// ApplyChanges applies a given set of changes in the DNS provider.
func (p *MikrotikProvider) ApplyChanges(ctx context.Context, changes *plan.Changes) error {
	changes = cleanupChanges(changes)

	for _, endpoint := range append(changes.UpdateOld, changes.Delete...) {
		if err := p.client.DeleteDNSRecord(endpoint); err != nil {
			return err
		}
	}

	for _, endpoint := range append(changes.Create, changes.UpdateNew...) {
		if _, err := p.client.CreateDNSRecord(endpoint); err != nil {
			return err
		}
	}

	return nil
}

// GetDomainFilter returns the domain filter for the provider.
func (p *MikrotikProvider) GetDomainFilter() endpoint.DomainFilter {
	return p.domainFilter
}
