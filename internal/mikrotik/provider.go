package mikrotik

import (
	"context"
	"fmt"

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

// Records returns the list of HostOverride records in Opnsense Unbound.
func (p *MikrotikProvider) Records(ctx context.Context) ([]*endpoint.Endpoint, error) {
	records, err := p.client.GetAllDNSRecords()
	if err != nil {
		return nil, err
	}

	var endpoints []*endpoint.Endpoint
	for _, record := range records {
		ep, _ := NewEndpointFromRecord(record)

		if !p.domainFilter.Match(ep.DNSName) {
			continue
		}

		endpoints = append(endpoints, ep)
	}

	return endpoints, nil
}

// ApplyChanges applies a given set of changes in the DNS provider.
func (p *MikrotikProvider) ApplyChanges(ctx context.Context, changes *plan.Changes) error {
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
