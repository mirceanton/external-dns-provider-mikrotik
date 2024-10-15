package mikrotik

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/external-dns/endpoint"
)

func TestMikrotikTTLtoEndpointTTL(t *testing.T) {
	tests := []struct {
		name        string
		inputTTL    string
		expectedTTL endpoint.TTL
		expectError bool
	}{
		{"Valid TTL with days", "1d5h20m15s", endpoint.TTL(105615), false},
		{"Valid TTL with hours", "2h30m", endpoint.TTL(9000), false},
		{"Valid TTL with minutes and seconds", "45m15s", endpoint.TTL(2715), false},
		{"Valid TTL with only days", "3d", endpoint.TTL(259200), false},
		{"Valid TTL with decimal days", "1.5d", endpoint.TTL(129600), false},
		{"Valid TTL with decimal hours", "2.5h", endpoint.TTL(9000), false},
		{"Invalid TTL string", "invalid", 0, true},
		{"Invalid unit", "1x", 0, true},
		{"Empty TTL string", "", endpoint.TTL(0), false},
		{"TTL with zero seconds", "0s", endpoint.TTL(0), false},
		{"TTL with negative value", "-1h", 0, true},
		{"TTL with unexpected characters", "1h30m20x", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ttl, err := mikrotikTTLtoEndpointTTL(tt.inputTTL)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTTL, ttl)
			}
		})
	}
}

func TestEndpointTTLtoMikrotikTTL(t *testing.T) {
	tests := []struct {
		name        string
		inputTTL    endpoint.TTL
		expectedTTL string
		expectError bool
	}{
		{"TTL with days, hours, minutes, and seconds", endpoint.TTL(105615), "1d5h20m15s", false},
		{"TTL with hours and minutes", endpoint.TTL(9000), "2h30m", false},
		{"TTL with minutes and seconds", endpoint.TTL(2715), "45m15s", false},
		{"TTL with only days", endpoint.TTL(259200), "3d", false},
		{"TTL with decimal days", endpoint.TTL(129600), "1d12h", false},
		{"TTL zero", endpoint.TTL(0), "0s", false},
		{"TTL negative", endpoint.TTL(-3600), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ttlStr, err := endpointTTLtoMikrotikTTL(tt.inputTTL)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, "", ttlStr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTTL, ttlStr)
			}
		})
	}
}

func TestDNSRecordToExternalDNSEndpoint(t *testing.T) {
	tests := []struct {
		name        string
		record      *DNSRecord
		expected    *endpoint.Endpoint
		expectError bool
	}{
		{
			name: "Valid A record",
			record: &DNSRecord{
				Name:    "example.com",
				Type:    "A",
				Address: "192.0.2.1",
				TTL:     "1h",
			},
			expected: &endpoint.Endpoint{
				DNSName:    "example.com",
				RecordType: "A",
				Targets:    endpoint.NewTargets("192.0.2.1"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expectError: false,
		},
		{
			name: "Valid AAAA record",
			record: &DNSRecord{
				Name:    "ipv6.example.com",
				Type:    "AAAA",
				Address: "2001:db8::1",
				TTL:     "2h",
			},
			expected: &endpoint.Endpoint{
				DNSName:    "ipv6.example.com",
				RecordType: "AAAA",
				Targets:    endpoint.NewTargets("2001:db8::1"),
				RecordTTL:  endpoint.TTL(7200),
			},
			expectError: false,
		},
		{
			name: "Valid CNAME record",
			record: &DNSRecord{
				Name:  "www.example.com",
				Type:  "CNAME",
				CName: "example.com",
				TTL:   "30m",
			},
			expected: &endpoint.Endpoint{
				DNSName:    "www.example.com",
				RecordType: "CNAME",
				Targets:    endpoint.NewTargets("example.com"),
				RecordTTL:  endpoint.TTL(1800),
			},
			expectError: false,
		},
		{
			name: "Valid TXT record",
			record: &DNSRecord{
				Name: "example.com",
				Type: "TXT",
				Text: "v=spf1 include:example.com ~all",
				TTL:  "10m",
			},
			expected: &endpoint.Endpoint{
				DNSName:    "example.com",
				RecordType: "TXT",
				Targets:    endpoint.NewTargets("v=spf1 include:example.com ~all"),
				RecordTTL:  endpoint.TTL(600),
			},
			expectError: false,
		},

		{
			name: "Record with match-subdomain",
			record: &DNSRecord{
				Name:           "example.com",
				Type:           "CNAME",
				CName:          "example.org",
				TTL:            "30m",
				MatchSubdomain: "yes",
			},
			expected: &endpoint.Endpoint{
				DNSName:    "example.com",
				RecordType: "CNAME",
				Targets:    endpoint.NewTargets("example.org"),
				RecordTTL:  endpoint.TTL(1800),
				ProviderSpecific: endpoint.ProviderSpecific{
					{Name: "match-subdomain", Value: "yes"},
				},
			},
			expectError: false,
		},
		{
			name: "Record with address-list",
			record: &DNSRecord{
				Name:        "blocked.example.com",
				Type:        "A",
				Address:     "192.0.2.123",
				TTL:         "1h",
				AddressList: "blocked",
			},
			expected: &endpoint.Endpoint{
				DNSName:    "blocked.example.com",
				RecordType: "A",
				Targets:    endpoint.NewTargets("192.0.2.123"),
				RecordTTL:  endpoint.TTL(3600),
				ProviderSpecific: endpoint.ProviderSpecific{
					{Name: "address-list", Value: "blocked"},
				},
			},
			expectError: false,
		},
		{
			name: "Invalid TTL in DNSRecord",
			record: &DNSRecord{
				Name:    "example.com",
				Type:    "A",
				Address: "192.0.2.1",
				TTL:     "invalid",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Unsupported record type",
			record: &DNSRecord{
				Name: "example.com",
				Type: "MX",
				TTL:  "1h",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Provider-specific properties",
			record: &DNSRecord{
				Name:           "example.com",
				Type:           "TXT",
				Text:           "some text",
				TTL:            "10m",
				Comment:        "Test comment",
				Regexp:         "^www\\.",
				MatchSubdomain: "yes",
				AddressList:    "list1",
			},
			expected: &endpoint.Endpoint{
				DNSName:    "example.com",
				RecordType: "TXT",
				Targets:    endpoint.NewTargets("some text"),
				RecordTTL:  endpoint.TTL(600),
				ProviderSpecific: endpoint.ProviderSpecific{
					{Name: "comment", Value: "Test comment"},
					{Name: "regexp", Value: "^www\\."},
					{Name: "match-subdomain", Value: "yes"},
					{Name: "address-list", Value: "list1"},
				},
			},
			expectError: false,
		},
		{
			name: "Empty Type (should default to 'A')",
			record: &DNSRecord{
				Name:    "example.com",
				Address: "192.0.2.1",
				TTL:     "1h",
			},
			expected: &endpoint.Endpoint{
				DNSName:    "example.com",
				RecordType: "A",
				Targets:    endpoint.NewTargets("192.0.2.1"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expectError: false,
		},
		{
			name: "Empty TTL (should default to 0)",
			record: &DNSRecord{
				Name:    "example.com",
				Type:    "A",
				Address: "192.0.2.1",
				// TTL is empty
			},
			expected: &endpoint.Endpoint{
				DNSName:    "example.com",
				RecordType: "A",
				Targets:    endpoint.NewTargets("192.0.2.1"),
				RecordTTL:  endpoint.TTL(0),
			},
			expectError: false,
		},
		{
			name: "Invalid A record (empty address)",
			record: &DNSRecord{
				Name:    "invalid.example.com",
				Type:    "A",
				Address: "",
				TTL:     "1h",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid AAAA record (empty address)",
			record: &DNSRecord{
				Name:    "invalid.example.com",
				Type:    "AAAA",
				Address: "",
				TTL:     "1h",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid CNAME record (empty cname)",
			record: &DNSRecord{
				Name:  "invalid.example.com",
				Type:  "CNAME",
				CName: "",
				TTL:   "30m",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid TXT record (empty text)",
			record: &DNSRecord{
				Name: "invalid.example.com",
				Type: "TXT",
				Text: "",
				TTL:  "10m",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Record with empty targets",
			record: &DNSRecord{
				Name: "empty.example.com",
				Type: "A",
				TTL:  "1h",
				// Address is empty
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint, err := tt.record.toExternalDNSEndpoint()
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, endpoint)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.DNSName, endpoint.DNSName)
				assert.Equal(t, tt.expected.RecordType, endpoint.RecordType)
				assert.Equal(t, tt.expected.Targets, endpoint.Targets)
				assert.Equal(t, tt.expected.RecordTTL, endpoint.RecordTTL)

				// Check provider-specific properties
				assert.Equal(t, len(tt.expected.ProviderSpecific), len(endpoint.ProviderSpecific))
				for _, expectedPS := range tt.expected.ProviderSpecific {
					found := false
					for _, actualPS := range endpoint.ProviderSpecific {
						if expectedPS.Name == actualPS.Name {
							assert.Equal(t, expectedPS.Value, actualPS.Value)
							found = true
							break
						}
					}
					assert.True(t, found, "ProviderSpecific property '%s' not found", expectedPS.Name)
				}
			}
		})
	}
}

func TestExternalDNSEndpointToDNSRecord(t *testing.T) {
	tests := []struct {
		name        string
		endpoint    *endpoint.Endpoint
		expected    *DNSRecord
		expectError bool
	}{
		// Valid A record
		{
			name: "Valid A record",
			endpoint: &endpoint.Endpoint{
				DNSName:    "example.com",
				RecordType: "A",
				Targets:    endpoint.NewTargets("192.0.2.1"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expected: &DNSRecord{
				Name:    "example.com",
				Type:    "A",
				Address: "192.0.2.1",
				TTL:     "1h",
			},
			expectError: false,
		},
		// Valid AAAA record
		{
			name: "Valid AAAA record",
			endpoint: &endpoint.Endpoint{
				DNSName:    "ipv6.example.com",
				RecordType: "AAAA",
				Targets:    endpoint.NewTargets("2001:db8::1"),
				RecordTTL:  endpoint.TTL(7200),
			},
			expected: &DNSRecord{
				Name:    "ipv6.example.com",
				Type:    "AAAA",
				Address: "2001:db8::1",
				TTL:     "2h",
			},
			expectError: false,
		},
		// Valid CNAME record
		{
			name: "Valid CNAME record",
			endpoint: &endpoint.Endpoint{
				DNSName:    "www.example.com",
				RecordType: "CNAME",
				Targets:    endpoint.NewTargets("example.com"),
				RecordTTL:  endpoint.TTL(1800),
			},
			expected: &DNSRecord{
				Name:  "www.example.com",
				Type:  "CNAME",
				CName: "example.com",
				TTL:   "30m",
			},
			expectError: false,
		},
		// Valid TXT record
		{
			name: "Valid TXT record",
			endpoint: &endpoint.Endpoint{
				DNSName:    "example.com",
				RecordType: "TXT",
				Targets:    endpoint.NewTargets("v=spf1 include:example.com ~all"),
				RecordTTL:  endpoint.TTL(600),
			},
			expected: &DNSRecord{
				Name: "example.com",
				Type: "TXT",
				Text: "v=spf1 include:example.com ~all",
				TTL:  "10m",
			},
			expectError: false,
		},
		// Valid A record with multiple targets (should use first target)
		{
			name: "Valid A record with multiple targets",
			endpoint: &endpoint.Endpoint{
				DNSName:    "multi.example.com",
				RecordType: "A",
				Targets:    endpoint.NewTargets("192.0.2.1", "192.0.2.2"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expected: &DNSRecord{
				Name:    "multi.example.com",
				Type:    "A",
				Address: "192.0.2.1", // Should use the first target
				TTL:     "1h",
			},
			expectError: false,
		},
		// Valid record with provider-specific properties
		{
			name: "Valid record with provider-specific properties",
			endpoint: &endpoint.Endpoint{
				DNSName:    "provider.example.com",
				RecordType: "A",
				Targets:    endpoint.NewTargets("192.0.2.3"),
				RecordTTL:  endpoint.TTL(3600),
				ProviderSpecific: endpoint.ProviderSpecific{
					{Name: "comment", Value: "Test comment"},
					{Name: "regexp", Value: "^www\\."},
					{Name: "match-subdomain", Value: "yes"},
					{Name: "address-list", Value: "list1"},
				},
			},
			expected: &DNSRecord{
				Name:           "provider.example.com",
				Type:           "A",
				Address:        "192.0.2.3",
				TTL:            "1h",
				Comment:        "Test comment",
				Regexp:         "^www\\.",
				MatchSubdomain: "yes",
				AddressList:    "list1",
			},
			expectError: false,
		},
		// Invalid A record (invalid IP address)
		{
			name: "Invalid A record (invalid IP address)",
			endpoint: &endpoint.Endpoint{
				DNSName:    "invalid.example.com",
				RecordType: "A",
				Targets:    endpoint.NewTargets("999.999.999.999"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expected:    nil,
			expectError: true,
		},
		// Invalid AAAA record (invalid IPv6 address)
		{
			name: "Invalid AAAA record (invalid IPv6 address)",
			endpoint: &endpoint.Endpoint{
				DNSName:    "invalid.example.com",
				RecordType: "AAAA",
				Targets:    endpoint.NewTargets("gggg::1"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expected:    nil,
			expectError: true,
		},
		// Invalid CNAME record (empty target)
		{
			name: "Invalid CNAME record (empty target)",
			endpoint: &endpoint.Endpoint{
				DNSName:    "invalid.example.com",
				RecordType: "CNAME",
				Targets:    endpoint.NewTargets(""),
				RecordTTL:  endpoint.TTL(1800),
			},
			expected:    nil,
			expectError: true,
		},
		// Invalid TXT record (empty text)
		{
			name: "Invalid TXT record (empty text)",
			endpoint: &endpoint.Endpoint{
				DNSName:    "invalid.example.com",
				RecordType: "TXT",
				Targets:    endpoint.NewTargets(""),
				RecordTTL:  endpoint.TTL(600),
			},
			expected:    nil,
			expectError: true,
		},
		// Record with empty targets
		{
			name: "Record with empty targets",
			endpoint: &endpoint.Endpoint{
				DNSName:    "empty.example.com",
				RecordType: "A",
				Targets:    endpoint.Targets{},
				RecordTTL:  endpoint.TTL(3600),
			},
			expected:    nil,
			expectError: true,
		},
		// Unsupported record type
		{
			name: "Unsupported record type",
			endpoint: &endpoint.Endpoint{
				DNSName:    "unsupported.example.com",
				RecordType: "MX",
				Targets:    endpoint.NewTargets("mail.example.com"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expected:    nil,
			expectError: true,
		},
		// Invalid provider-specific configuration
		{
			name: "Invalid provider-specific configuration",
			endpoint: &endpoint.Endpoint{
				DNSName:    "example.com",
				RecordType: "TXT",
				Targets:    endpoint.NewTargets("some text"),
				ProviderSpecific: endpoint.ProviderSpecific{
					{Name: "unsupported", Value: "value"},
				},
			},
			expected:    nil,
			expectError: true,
		},
		// Empty DNSName
		{
			name: "Empty DNSName",
			endpoint: &endpoint.Endpoint{
				DNSName:    "",
				RecordType: "A",
				Targets:    endpoint.NewTargets("192.0.2.1"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expected:    nil,
			expectError: true,
		},
		// Empty RecordType
		{
			name: "Empty RecordType",
			endpoint: &endpoint.Endpoint{
				DNSName:    "example.com",
				RecordType: "",
				Targets:    endpoint.NewTargets("192.0.2.1"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expected:    nil,
			expectError: true,
		},
		// Empty TTL (should default to "0s")
		{
			name: "Empty TTL (should default to '0s')",
			endpoint: &endpoint.Endpoint{
				DNSName:    "example.com",
				RecordType: "A",
				Targets:    endpoint.NewTargets("192.0.2.1"),
				// RecordTTL is zero value
			},
			expected: &DNSRecord{
				Name:    "example.com",
				Type:    "A",
				Address: "192.0.2.1",
				TTL:     "0s",
			},
			expectError: false,
		},
		// Invalid TTL value
		{
			name: "Invalid TTL value (negative)",
			endpoint: &endpoint.Endpoint{
				DNSName:    "example.com",
				RecordType: "A",
				Targets:    endpoint.NewTargets("192.0.2.1"),
				RecordTTL:  endpoint.TTL(-1),
			},
			expected:    nil,
			expectError: true,
		},
		// Setting match-subdomain via provider-specific
		{
			name: "Setting match-subdomain via provider-specific",
			endpoint: &endpoint.Endpoint{
				DNSName:    "example.com",
				RecordType: "CNAME",
				Targets:    endpoint.NewTargets("example.org"),
				RecordTTL:  endpoint.TTL(1800),
				ProviderSpecific: endpoint.ProviderSpecific{
					{Name: "match-subdomain", Value: "yes"},
				},
			},
			expected: &DNSRecord{
				Name:           "example.com",
				Type:           "CNAME",
				CName:          "example.org",
				TTL:            "30m",
				MatchSubdomain: "yes",
			},
			expectError: false,
		},
		// Setting address-list via provider-specific
		{
			name: "Setting address-list via provider-specific",
			endpoint: &endpoint.Endpoint{
				DNSName:    "blocked.example.com",
				RecordType: "A",
				Targets:    endpoint.NewTargets("192.0.2.123"),
				RecordTTL:  endpoint.TTL(3600),
				ProviderSpecific: endpoint.ProviderSpecific{
					{Name: "address-list", Value: "blocked"},
				},
			},
			expected: &DNSRecord{
				Name:        "blocked.example.com",
				Type:        "A",
				Address:     "192.0.2.123",
				TTL:         "1h",
				AddressList: "blocked",
			},
			expectError: false,
		},
		// Multiple provider-specific properties
		{
			name: "Multiple provider-specific properties",
			endpoint: &endpoint.Endpoint{
				DNSName:    "multi.example.com",
				RecordType: "TXT",
				Targets:    endpoint.NewTargets("some text"),
				RecordTTL:  endpoint.TTL(600),
				ProviderSpecific: endpoint.ProviderSpecific{
					{Name: "comment", Value: "Test comment"},
					{Name: "address-list", Value: "list1"},
					{Name: "match-subdomain", Value: "yes"},
				},
			},
			expected: &DNSRecord{
				Name:           "multi.example.com",
				Type:           "TXT",
				Text:           "some text",
				TTL:            "10m",
				Comment:        "Test comment",
				AddressList:    "list1",
				MatchSubdomain: "yes",
			},
			expectError: false,
		},
		// Invalid provider-specific name
		{
			name: "Invalid provider-specific name",
			endpoint: &endpoint.Endpoint{
				DNSName:    "invalid.example.com",
				RecordType: "A",
				Targets:    endpoint.NewTargets("192.0.2.1"),
				ProviderSpecific: endpoint.ProviderSpecific{
					{Name: "invalid-name", Value: "value"},
				},
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record, err := NewDNSRecord(tt.endpoint)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, record)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, record)
				assert.Equal(t, tt.expected.Name, record.Name)
				assert.Equal(t, tt.expected.Type, record.Type)
				assert.Equal(t, tt.expected.TTL, record.TTL)

				// Verify the record content based on type
				switch record.Type {
				case "A", "AAAA":
					assert.Equal(t, tt.expected.Address, record.Address)
				case "CNAME":
					assert.Equal(t, tt.expected.CName, record.CName)
				case "TXT":
					assert.Equal(t, tt.expected.Text, record.Text)
				}

				// Check provider-specific properties
				assert.Equal(t, tt.expected.Comment, record.Comment)
				assert.Equal(t, tt.expected.Regexp, record.Regexp)
				assert.Equal(t, tt.expected.MatchSubdomain, record.MatchSubdomain)
				assert.Equal(t, tt.expected.AddressList, record.AddressList)
			}
		})
	}
}
