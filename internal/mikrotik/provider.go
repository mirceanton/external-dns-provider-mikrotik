package mikrotik

import (
	"context"
	"fmt"

	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
	"sigs.k8s.io/external-dns/provider"
)

type MikrotikProvider struct {
	provider.BaseProvider

	client       *MikrotikApiClient
	domainFilter endpoint.DomainFilter
}

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

func (p *MikrotikProvider) Records(ctx context.Context) ([]*endpoint.Endpoint, error) {
	records, err := p.client.GetAll()
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

func (p *MikrotikProvider) ApplyChanges(ctx context.Context, changes *plan.Changes) error {
	for _, endpoint := range append(changes.UpdateOld, changes.Delete...) {
		if err := p.client.Delete(endpoint); err != nil {
			return err
		}
	}

	for _, endpoint := range append(changes.Create, changes.UpdateNew...) {
		if _, err := p.client.Create(endpoint); err != nil {
			return err
		}
	}

	return nil
}

func (p *MikrotikProvider) GetDomainFilter() endpoint.DomainFilter {
	return p.domainFilter
}
