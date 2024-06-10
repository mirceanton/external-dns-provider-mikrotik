package mikrotik

import (
	"fmt"

	"sigs.k8s.io/external-dns/endpoint"
)

// NewDNSRecord converts an ExternalDNS Endpoint to a Mikrotik DNSRecord
func NewRecordFromEndpoint(endpoint *endpoint.Endpoint) (*DNSRecord, error) {
	record := DNSRecord{
		Name:    endpoint.DNSName,
		Type:    endpoint.RecordType,
		Comment: "Managed by ExternalDNS",
	}

	switch endpoint.RecordType {
	case "A", "AAAA":
		record.Address = endpoint.Targets[0]
	case "CNAME":
		record.CName = endpoint.Targets[0]
	case "TXT":
		record.Text = endpoint.Targets[0]

	default:
		return nil, fmt.Errorf("unsupported DNS type: %s", endpoint.RecordType)
	}

	return &record, nil
}

func NewEndpointFromRecord(record DNSRecord) (*endpoint.Endpoint, error) {
	var ep endpoint.Endpoint
	switch record.Type {
	case "", "A", "AAAA": // "" means A record because mikrotik is weird like that... :P
		ep = endpoint.Endpoint{
			DNSName:    record.Name,
			RecordType: record.Type,
			Targets:    endpoint.NewTargets(record.Address),
		}
	case "CNAME":
		ep = endpoint.Endpoint{
			DNSName:    record.Name,
			RecordType: record.Type,
			Targets:    endpoint.NewTargets(record.CName),
		}
	case "TXT":
		ep = endpoint.Endpoint{
			DNSName:    record.Name,
			RecordType: record.Type,
			Targets:    endpoint.NewTargets(record.Text),
		}

	default:
		return nil, fmt.Errorf("unsupported DNS type: %s", record.Type)
	}

	return &ep, nil
}
