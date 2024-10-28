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

// NewMikrotikProvider initializes a new DNSProvider, of the Mikrotik variety
func NewMikrotikProvider(domainFilter endpoint.DomainFilter, config *MikrotikConnectionConfig) (provider.Provider, error) {
	// Create the Mikrotik API Client
	client, err := NewMikrotikClient(config)
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

func cleanupChanges(changes *plan.Changes) *plan.Changes {
	for oldIndex, old := range changes.UpdateOld {
		for newIndex, new := range changes.UpdateNew {
			if old.DNSName == new.DNSName && old.Targets[0] == new.Targets[0] && old.RecordTTL == new.RecordTTL {
				needUpdate := false

				oldComment, _ := old.GetProviderSpecificProperty("comment")
				newComment, _ := new.GetProviderSpecificProperty("webhook/comment")
				if oldComment != newComment {
					needUpdate = true
				}

				oldMatchSubdomain, _ := old.GetProviderSpecificProperty("match-subdomain")
				newMatchSubdomain, _ := new.GetProviderSpecificProperty("webhook/match-subdomain")
				if oldMatchSubdomain != newMatchSubdomain {
					needUpdate = true
				}

				oldAddressList, _ := old.GetProviderSpecificProperty("address-list")
				newAddressList, _ := new.GetProviderSpecificProperty("webhook/address-list")
				if oldAddressList != newAddressList {
					needUpdate = true
				}

				oldRegexp, _ := old.GetProviderSpecificProperty("regexp")
				newRegexp, _ := new.GetProviderSpecificProperty("webhook/regexp")
				if oldRegexp != newRegexp {
					needUpdate = true
				}

				if !needUpdate {
					changes.UpdateOld = append(changes.UpdateOld[:oldIndex], changes.UpdateOld[oldIndex+1:]...)
					changes.UpdateNew = append(changes.UpdateNew[:newIndex], changes.UpdateNew[newIndex+1:]...)
				}
			}
		}
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
func (p *MikrotikProvider) GetDomainFilter() endpoint.DomainFilterInterface {
	return p.domainFilter
}
