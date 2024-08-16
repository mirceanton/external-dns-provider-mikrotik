package mikrotik

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/external-dns/endpoint"
)

func TestRecordConversion(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		input    endpoint.Endpoint
		expected DNSRecord
	}{
		{
			name: "Basic A Record Conversion",
			input: endpoint.Endpoint{
				DNSName:    "google.com",
				RecordType: "A",
				RecordTTL:  300,
				Targets:    endpoint.Targets{"8.8.8.8"},
			},
			expected: DNSRecord{
				Name:    "google.com",
				Type:    "A",
				TTL:     "5m0s",
				Address: "8.8.8.8",
			},
		},
		{
			name: "Basic AAAA Record Conversion",
			input: endpoint.Endpoint{
				DNSName:    "google.com",
				RecordType: "AAAA",
				RecordTTL:  180,
				Targets:    endpoint.Targets{"2607:f8b0:400a:800::200e"},
			},
			expected: DNSRecord{
				Name:    "google.com",
				Type:    "AAAA",
				TTL:     "3m0s",
				Address: "2607:f8b0:400a:800::200e",
			},
		},
		{
			name: "Basic CNAME Record Conversion",
			input: endpoint.Endpoint{
				DNSName:    "google.com",
				RecordType: "CNAME",
				RecordTTL:  120,
				Targets:    endpoint.Targets{"maps.google.com"},
			},
			expected: DNSRecord{
				Name:  "google.com",
				Type:  "CNAME",
				TTL:   "2m0s",
				CName: "maps.google.com",
			},
		},
		{
			name: "Basic TXT Record Conversion",
			input: endpoint.Endpoint{
				DNSName:    "google.com",
				RecordType: "TXT",
				RecordTTL:  100,
				Targets:    endpoint.Targets{"lorem ipsum dolor sit amet"},
			},
			expected: DNSRecord{
				Name: "google.com",
				Type: "TXT",
				TTL:  "1m40s",
				Text: "lorem ipsum dolor sit amet",
			},
		},
		{
			name: "Complete Record Conversion",
			input: endpoint.Endpoint{
				DNSName:    "google.com",
				RecordType: "A",
				RecordTTL:  10,
				Targets:    endpoint.Targets{"1.1.1.1"},
				ProviderSpecific: []endpoint.ProviderSpecificProperty{
					{Name: "disabled", Value: "true"},
					{Name: "comment", Value: "test comment"},
					{Name: "regexp", Value: "test regexp"},
					{Name: "match-subdomain", Value: "true"},
					{Name: "address-list", Value: "test address list"},
				},
			},
			expected: DNSRecord{
				Name:           "google.com",
				Type:           "A",
				TTL:            "10s",
				Address:        "1.1.1.1",
				Disabled:       "true",
				Comment:        "test comment",
				Regexp:         "test regexp",
				MatchSubdomain: "true",
				AddressList:    "test address list",
			},
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record, err := NewDNSRecord(&tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, *record)

			endpoint, err := record.toExternalDNSEndpoint()
			assert.NoError(t, err)
			assert.Equal(t, tt.input, *endpoint)
		})
	}
}

func TestNonSupportedRecordType(t *testing.T) {
	endpoint := endpoint.Endpoint{
		DNSName:    "name",
		RecordType: "unsupported",
		RecordTTL:  0,
		Targets:    endpoint.Targets{""},
	}

	record, err := NewDNSRecord(&endpoint)
	assert.Nil(t, record)
	assert.Error(t, err)
}

func TestEmptyRecordType(t *testing.T) {
	record := &DNSRecord{
		Name:    "name",
		Type:    "",
		TTL:     "10s",
		Address: "1.1.1.1",
	}

	endpoint, err := record.toExternalDNSEndpoint()
	assert.NoError(t, err)
	assert.Equal(t, endpoint.RecordType, "A")
}
