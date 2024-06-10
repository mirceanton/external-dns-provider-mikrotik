package mikrotik

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/external-dns/endpoint"
)

// NewDNSRecord converts an ExternalDNS Endpoint to a Mikrotik DNSRecord
func NewRecordFromEndpoint(endpoint *endpoint.Endpoint) (*DNSRecord, error) {
	jsonBody, err := json.Marshal(endpoint)
	if err != nil {
		log.Errorf("Error marshalling endpoint: %v", err)
		return nil, err
	}
	log.Debugf("Endpoint to parse: %s", string(jsonBody))

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

	jsonBody, err = json.Marshal(record)
	if err != nil {
		log.Errorf("Error marshalling dns record: %v", err)
		return nil, err
	}
	log.Debugf("Record parsed: %s", string(jsonBody))

	return &record, nil
}

func NewEndpointFromRecord(record DNSRecord) (*endpoint.Endpoint, error) {
	jsonBody, err := json.Marshal(record)
	if err != nil {
		log.Errorf("Error marshalling record: %v", err)
		return nil, err
	}
	log.Debugf("Record to parse: %s", string(jsonBody))

	var ep endpoint.Endpoint
	if record.Type == "" {
		record.Type = "A"
	}
	switch record.Type {
	case "A", "AAAA":
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
