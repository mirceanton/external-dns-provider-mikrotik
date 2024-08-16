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

// ApplyChanges applies a given set of changes in the DNS provider.
func (p *MikrotikProvider) ApplyChanges(ctx context.Context, changes *plan.Changes) error {
	// if changes.UpdateNew has a record that is in changes.UpdateOld, but old has no provider specific "disabled",
	// and new has provider specific "disabled" set to "false", then we should ignore the record
	for _, newEndpoint := range changes.UpdateNew {
		for _, oldEndpoint := range changes.UpdateOld {
			if newEndpoint.DNSName == oldEndpoint.DNSName && newEndpoint.RecordType == oldEndpoint.RecordType {
				newDisabled, _ := newEndpoint.GetProviderSpecificProperty("disabled")
				oldDisabled, _ := oldEndpoint.GetProviderSpecificProperty("disabled")
				if newDisabled == "false" && oldDisabled == "" {
					log.Debugf("Ignoring update for %s because it is disabled", newEndpoint.DNSName)
					// TODO: remove the record from changes.UpdateNew and changes.UpdateOld
				}

			}
		}
	}

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
