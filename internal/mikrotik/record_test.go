package mikrotik

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/external-dns/endpoint"
)

func TestRecordJSONMarshaling(t *testing.T) {
	record := DNSRecord{
		ID:             "1",
		Address:        "192.168.1.1",
		CName:          "example.com",
		ForwardTo:      "forward.example.com",
		MXExchange:     "mail.example.com",
		Name:           "example",
		SrvPort:        "8080",
		SrvTarget:      "target.example.com",
		Text:           "some text",
		Type:           "A",
		AddressList:    "list",
		Comment:        "a comment",
		Disabled:       "false",
		MatchSubdomain: "sub.example.com",
		MXPreference:   "10",
		NS:             "ns.example.com",
		Regexp:         ".*",
		SrvPriority:    "1",
		SrvWeight:      "5",
	}

	data, err := json.Marshal(record)
	assert.NoError(t, err)

	var unmarshaledRecord DNSRecord
	err = json.Unmarshal(data, &unmarshaledRecord)
	assert.NoError(t, err)

	assert.Equal(t, record, unmarshaledRecord)
}

func TestRecordJSONUnmarshaling(t *testing.T) {
	// Define test cases
	tests := []struct {
		name           string
		data           []byte
		expectedRecord DNSRecord
	}{
		{
			name: "Complete record",
			data: []byte(`{
				".id": "1",
				"address": "192.168.1.1",
				"cname": "example.com",
				"forward-to": "forward.example.com",
				"mx-exchange": "mail.example.com",
				"name": "example",
				"srv-port": "8080",
				"srv-target": "target.example.com",
				"text": "some text",
				"type": "A",
				"address-list": "list",
				"comment": "a comment",
				"disabled": "false",
				"match-subdomain": "sub.example.com",
				"mx-preference": "10",
				"ns": "ns.example.com",
				"regexp": ".*",
				"srv-priority": "1",
				"srv-weight": "5"
			}`),
			expectedRecord: DNSRecord{
				ID:             "1",
				Address:        "192.168.1.1",
				CName:          "example.com",
				ForwardTo:      "forward.example.com",
				MXExchange:     "mail.example.com",
				Name:           "example",
				SrvPort:        "8080",
				SrvTarget:      "target.example.com",
				Text:           "some text",
				Type:           "A",
				AddressList:    "list",
				Comment:        "a comment",
				Disabled:       "false",
				MatchSubdomain: "sub.example.com",
				MXPreference:   "10",
				NS:             "ns.example.com",
				Regexp:         ".*",
				SrvPriority:    "1",
				SrvWeight:      "5",
			},
		},
		{
			name: "Simple record",
			data: []byte(`{
				"name": "example.com",
				"address": "192.168.1.1",
				"type": "A"
			}`),
			expectedRecord: DNSRecord{
				Name:    "example.com",
				Address: "192.168.1.1",
				Type:    "A",
			},
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var record DNSRecord
			err := json.Unmarshal(tt.data, &record)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedRecord, record)
		})
	}
}

func TestNewRecordFromEndpoint(t *testing.T) {
	// Define test cases
	tests := []struct {
		name           string
		endpoint       *endpoint.Endpoint
		shouldError    bool
		expectedRecord *DNSRecord
	}{
		{
			name:        "Basic A record",
			shouldError: false,
			endpoint: &endpoint.Endpoint{
				DNSName:    "example.com",
				Targets:    []string{"1.2.3.4"},
				RecordType: "A",
			},
			expectedRecord: &DNSRecord{
				Name:    "example.com",
				Type:    "A",
				Address: "1.2.3.4",
				Comment: "Managed by ExternalDNS",
			},
		},
		{
			name:        "Basic CNAME record",
			shouldError: false,
			endpoint: &endpoint.Endpoint{
				DNSName:    "example.com",
				Targets:    []string{"cname.example.com"},
				RecordType: "CNAME",
			},
			expectedRecord: &DNSRecord{
				Name:    "example.com",
				Type:    "CNAME",
				CName:   "cname.example.com",
				Comment: "Managed by ExternalDNS",
			},
		},
		{
			name:        "Basic TXT record",
			shouldError: false,
			endpoint: &endpoint.Endpoint{
				DNSName:    "example.com",
				Targets:    []string{"some text"},
				RecordType: "TXT",
			},
			expectedRecord: &DNSRecord{
				Name:    "example.com",
				Type:    "TXT",
				Text:    "some text",
				Comment: "Managed by ExternalDNS",
			},
		},
		{
			name:        "Basic AAAA record",
			shouldError: false,
			endpoint: &endpoint.Endpoint{
				DNSName:    "example.com",
				Targets:    []string{"2001:db8::1"},
				RecordType: "AAAA",
			},
			expectedRecord: &DNSRecord{
				Name:    "example.com",
				Type:    "AAAA",
				Address: "2001:db8::1",
				Comment: "Managed by ExternalDNS",
			},
		},
		{
			name:        "Unsupported record type",
			shouldError: true,
			endpoint: &endpoint.Endpoint{
				DNSName:    "example.com",
				Targets:    []string{""},
				RecordType: "SRV",
			},
			expectedRecord: nil,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record, err := NewRecordFromEndpoint(tt.endpoint)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedRecord, record)
		})
	}
}

func TestNewEndpointFromRecord(t *testing.T) {
	// Define test cases
	tests := []struct {
		name             string
		shouldError      bool
		record           DNSRecord
		expectedEndpoint *endpoint.Endpoint
	}{
		{
			name:        "Basic A record",
			shouldError: false,
			record: DNSRecord{
				Name:    "example.com",
				Type:    "A",
				Address: "192.168.1.1",
			},
			expectedEndpoint: &endpoint.Endpoint{
				DNSName:    "example.com",
				Targets:    endpoint.NewTargets("192.168.1.1"),
				RecordType: "A",
			},
		},
		{
			name:        "Basic CNAME record",
			shouldError: false,
			record: DNSRecord{
				Name:  "example.com",
				Type:  "CNAME",
				CName: "cname.example.com",
			},
			expectedEndpoint: &endpoint.Endpoint{
				DNSName:    "example.com",
				Targets:    endpoint.NewTargets("cname.example.com"),
				RecordType: "CNAME",
			},
		},
		{
			name:        "Basic TXT record",
			shouldError: false,
			record: DNSRecord{
				Name: "example.com",
				Type: "TXT",
				Text: "some text",
			},
			expectedEndpoint: &endpoint.Endpoint{
				DNSName:    "example.com",
				Targets:    endpoint.NewTargets("some text"),
				RecordType: "TXT",
			},
		},
		{
			name:        "Basic AAAA record",
			shouldError: false,
			record: DNSRecord{
				Name:    "example.com",
				Type:    "AAAA",
				Address: "2001:db8::1",
			},
			expectedEndpoint: &endpoint.Endpoint{
				DNSName:    "example.com",
				Targets:    endpoint.NewTargets("2001:db8::1"),
				RecordType: "AAAA",
			},
		},
		{
			name:        "Unsupported record type",
			shouldError: true,
			record: DNSRecord{
				Name: "example.com",
				Type: "SRV",
			},
			expectedEndpoint: nil,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint, err := NewEndpointFromRecord(tt.record)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedEndpoint, endpoint)
		})
	}
}
