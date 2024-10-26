package mikrotik

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/external-dns/endpoint"
)

// ================================================================================================
// Test Validation Functions
// ================================================================================================
func TestValidateIPv4(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		expectError bool
	}{
		{"Valid IPv4 address", "192.168.1.1", false},
		{"Invalid IPv4 address", "256.256.256.256", true},
		{"Looks like IPv6", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", true},
		{"Empty address", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIPv4(tt.address)
			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v for address: %s", tt.expectError, err, tt.address)
			}
		})
	}
}

func TestValidateIPv6(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		expectError bool
	}{
		{"Valid IPv6 address", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", false},
		{"Invalid IPv6 address", "1200:0000:AB00:1234:0000:2552:7777:1313:3", true},
		{"Looks like IPv4", "192.168.1.1", true},
		{"Empty address", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIPv6(tt.address)
			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v for address: %s", tt.expectError, err, tt.address)
			}
		})
	}
}

func TestValidateTXT(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		expectError bool
	}{
		{"Valid TXT record", "This is a valid TXT record", false},
		{"Empty TXT record", "", true},
		{"Single space TXT record", " ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTXT(tt.text)
			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v for TXT record: %s", tt.expectError, err, tt.text)
			}
		})
	}
}

func TestValidateDomain(t *testing.T) {
	tests := []struct {
		name        string
		domain      string
		expectError bool
	}{
		{"Valid domain", "example.com", false},
		{"Invalid domain with underscores", "example_domain.com", true},
		{"Too long domain", strings.Repeat("a", 255) + ".com", true},
		{"Empty domain", "", true},
		{"Invalid domain format", "invalid_domain", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDomain(tt.domain)
			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v for domain: %s", tt.expectError, err, tt.domain)
			}
		})
	}
}

func TestValidateMXPreference(t *testing.T) {
	tests := []struct {
		name        string
		preference  string
		expectError bool
	}{
		{"Valid MX preference", "10", false},
		{"Empty MX preference", "", true},
		{"Negative MX preference", "-10", true},
		{"Too high MX preference", "70000", true},
		{"Not a number", "abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUnsignedInteger(tt.preference)
			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v for MX preference: %s", tt.expectError, err, tt.preference)
			}
		})
	}
}

// ================================================================================================
// Test TTL Conversion Functions
// ================================================================================================
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

// ================================================================================================
// Test DNS Record Conversion Functions
// ================================================================================================
func TestDNSRecordToExternalDNSEndpoint(t *testing.T) {
	tests := []struct {
		name        string
		record      *DNSRecord
		expected    *endpoint.Endpoint
		expectError bool
	}{
		// ===============================================================
		// A RECORD TEST CASES
		// ===============================================================
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
			name: "Invalid A record (malformed address)",
			record: &DNSRecord{
				Name:    "invalid.example.com",
				Type:    "A",
				Address: "999.999.999.999",
				TTL:     "1h",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid A record (IPv6 address)",
			record: &DNSRecord{
				Name:    "invalid.example.com",
				Type:    "A",
				Address: "2001:db8::1",
				TTL:     "1h",
			},
			expected:    nil,
			expectError: true,
		},

		// ===============================================================
		// AAAA RECORD TEST CASES
		// ===============================================================
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
			name: "Invalid AAAA record (IPv4 address)",
			record: &DNSRecord{
				Name:    "invalid.example.com",
				Type:    "AAAA",
				Address: "1.2.3.4",
				TTL:     "1h",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid AAAA record (malformed address)",
			record: &DNSRecord{
				Name:    "invalid.example.com",
				Type:    "AAAA",
				Address: "1200:0000:AB00:1234:0000:2552:7777:1313:3:31",
				TTL:     "1h",
			},
			expected:    nil,
			expectError: true,
		},

		// ===============================================================
		// CNAME RECORD TEST CASES
		// ===============================================================
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
			name: "Invalid CNAME record (malformed domain)",
			record: &DNSRecord{
				Name:  "invalid.example.com",
				Type:  "CNAME",
				CName: "sub......domain...here-",
				TTL:   "30m",
			},
			expected:    nil,
			expectError: true,
		},

		// ===============================================================
		// TXT RECORD TEST CASES
		// ===============================================================
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

		// ===============================================================
		// MX RECORD TEST CASES
		// ===============================================================
		{
			name: "Valid MX record",
			record: &DNSRecord{
				Name:         "mx.example.com",
				Type:         "MX",
				MXExchange:   "mailhost1.example.com",
				MXPreference: "10",
				TTL:          "10m",
			},
			expected: &endpoint.Endpoint{
				DNSName:    "mx.example.com",
				RecordType: "MX",
				Targets:    endpoint.NewTargets("10 mailhost1.example.com"),
				RecordTTL:  endpoint.TTL(600),
			},
			expectError: false,
		},
		{
			name: "Invalid MX record (empty preference)",
			record: &DNSRecord{
				Name:         "invalid.example.com",
				Type:         "MX",
				MXExchange:   "mailhost1.example.com",
				MXPreference: "",
				TTL:          "10m",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid MX record (negative preference)",
			record: &DNSRecord{
				Name:         "invalid.example.com",
				Type:         "MX",
				MXExchange:   "mailhost1.example.com",
				MXPreference: "-10",
				TTL:          "10m",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid MX record (too large preference)",
			record: &DNSRecord{
				Name:         "invalid.example.com",
				Type:         "MX",
				MXExchange:   "mailhost1.example.com",
				MXPreference: "70000",
				TTL:          "10m",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid MX record (non-numeric preference)",
			record: &DNSRecord{
				Name:         "invalid.example.com",
				Type:         "MX",
				MXExchange:   "mailhost1.example.com",
				MXPreference: "invalid",
				TTL:          "10m",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid MX record (empty exchange)",
			record: &DNSRecord{
				Name:         "invalid.example.com",
				Type:         "MX",
				MXExchange:   "",
				MXPreference: "1",
				TTL:          "10m",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid MX record (exchange missing TLD)",
			record: &DNSRecord{
				Name:         "invalid.example.com",
				Type:         "MX",
				MXExchange:   "invalid_domain",
				MXPreference: "10",
				TTL:          "10m",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid MX record (malformed domain)",
			record: &DNSRecord{
				Name:         "invalid.example.com",
				Type:         "MX",
				MXExchange:   "sub....domain.com",
				MXPreference: "10",
				TTL:          "10m",
			},
			expected:    nil,
			expectError: true,
		},

		// ===============================================================
		// SRV RECORD TEST CASES
		// ===============================================================
		{
			name: "Valid SRV record",
			record: &DNSRecord{
				Name:        "_sip._tcp.example.com",
				Type:        "SRV",
				SrvPriority: "10",
				SrvWeight:   "20",
				SrvPort:     "5060",
				SrvTarget:   "sipserver.example.com",
				TTL:         "1h",
			},
			expected: &endpoint.Endpoint{
				DNSName:    "_sip._tcp.example.com",
				RecordType: "SRV",
				Targets:    endpoint.NewTargets("10 20 5060 sipserver.example.com"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expectError: false,
		},
		{
			name: "Valid SRV record with minimum priority and weight",
			record: &DNSRecord{
				Name:        "_sip._tcp.example.com",
				Type:        "SRV",
				SrvPriority: "0",
				SrvWeight:   "0",
				SrvPort:     "443",
				SrvTarget:   "example.com",
				TTL:         "1h",
			},
			expected: &endpoint.Endpoint{
				DNSName:    "_sip._tcp.example.com",
				RecordType: "SRV",
				Targets:    endpoint.NewTargets("0 0 443 example.com"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expectError: false,
		},
		{
			name: "Valid SRV record with maximum priority and weight",
			record: &DNSRecord{
				Name:        "_sip._tcp.example.com",
				Type:        "SRV",
				SrvPriority: "65535",
				SrvWeight:   "65535",
				SrvPort:     "53",
				SrvTarget:   "domain.com",
				TTL:         "1h",
			},
			expected: &endpoint.Endpoint{
				DNSName:    "_sip._tcp.example.com",
				RecordType: "SRV",
				Targets:    endpoint.NewTargets("65535 65535 53 domain.com"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expectError: false,
		},
		{
			name: "Invalid SRV record (negative priority)",
			record: &DNSRecord{
				Name:        "_sip._tcp.example.com",
				Type:        "SRV",
				SrvPriority: "-1",
				SrvWeight:   "20",
				SrvPort:     "80",
				SrvTarget:   "sipserver.example.com",
				TTL:         "1h",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid SRV record (negative weight)",
			record: &DNSRecord{
				Name:        "_sip._tcp.example.com",
				Type:        "SRV",
				SrvPriority: "10",
				SrvWeight:   "-5",
				SrvPort:     "80",
				SrvTarget:   "sipserver.example.com",
				TTL:         "1h",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid SRV record (port out of range)",
			record: &DNSRecord{
				Name:        "_sip._tcp.example.com",
				Type:        "SRV",
				SrvPriority: "10",
				SrvWeight:   "20",
				SrvPort:     "70000",
				SrvTarget:   "sipserver.example.com",
				TTL:         "1h",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid SRV record (empty target)",
			record: &DNSRecord{
				Name:        "_sip._tcp.example.com",
				Type:        "SRV",
				SrvPriority: "10",
				SrvWeight:   "20",
				SrvPort:     "80",
				SrvTarget:   "",
				TTL:         "1h",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid SRV record (malformed target domain)",
			record: &DNSRecord{
				Name:        "_sip._tcp.example.com",
				Type:        "SRV",
				SrvPriority: "10",
				SrvWeight:   "20",
				SrvPort:     "80",
				SrvTarget:   "invalid_domain..com",
				TTL:         "1h",
			},
			expected:    nil,
			expectError: true,
		},

		// ===============================================================
		// PROVIDER-SPECIFIC DATA TEST CASES
		// ===============================================================
		{
			name: "Valid Provider-specific properties",
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
		// TODO: invalid provider specific

		// ===============================================================
		// DEFAULT VALUES FOR UNSET FIELDS TEST CASES
		// ===============================================================
		{
			name: "Empty Type (should default to 'A')",
			record: &DNSRecord{
				Name: "example.com",
				//! Type is empty
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
				//! TTL is empty
			},
			expected: &endpoint.Endpoint{
				DNSName:    "example.com",
				RecordType: "A",
				Targets:    endpoint.NewTargets("192.0.2.1"),
				RecordTTL:  endpoint.TTL(0),
			},
			expectError: false,
		},

		// ===============================================================
		// GENERIC ERROR CASES
		// ===============================================================
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
				Type: "FWD",
				TTL:  "1h",
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
				//! Address is empty
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
		// ===============================================================
		// A RECORD TEST CASES
		// ===============================================================
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
		{
			name: "Invalid A record (emopty address)",
			endpoint: &endpoint.Endpoint{
				DNSName:    "invalid.example.com",
				RecordType: "A",
				Targets:    endpoint.NewTargets(),
				RecordTTL:  endpoint.TTL(3600),
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid A record (malformed address)",
			endpoint: &endpoint.Endpoint{
				DNSName:    "invalid.example.com",
				RecordType: "A",
				Targets:    endpoint.NewTargets("999.999.999.999"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid A record (IPv6 address)",
			endpoint: &endpoint.Endpoint{
				DNSName:    "invalid.example.com",
				RecordType: "A",
				Targets:    endpoint.NewTargets("2001:db8::1"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expected:    nil,
			expectError: true,
		},

		// ===============================================================
		// AAAA RECORD TEST CASES
		// ===============================================================
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
		{
			name: "Valid AAAA record with multiple targets",
			endpoint: &endpoint.Endpoint{
				DNSName:    "multi.example.com",
				RecordType: "AAAA",
				Targets:    endpoint.NewTargets("2001:db8::1", "2001:db8::2"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expected: &DNSRecord{
				Name:    "multi.example.com",
				Type:    "AAAA",
				Address: "2001:db8::1", // Should use the first target
				TTL:     "1h",
			},
			expectError: false,
		},
		{
			name: "Invalid AAAA record (empty address)",
			endpoint: &endpoint.Endpoint{
				DNSName:    "multi.example.com",
				RecordType: "AAAA",
				Targets:    endpoint.NewTargets(""),
				RecordTTL:  endpoint.TTL(3600),
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid AAAA record (malformed address)",
			endpoint: &endpoint.Endpoint{
				DNSName:    "invalid.example.com",
				RecordType: "AAAA",
				Targets:    endpoint.NewTargets("gggg::1"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid AAAA record (IPv4 address)",
			endpoint: &endpoint.Endpoint{
				DNSName:    "invalid.example.com",
				RecordType: "AAAA",
				Targets:    endpoint.NewTargets("1.2.3.4"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expected:    nil,
			expectError: true,
		},

		// ===============================================================
		// CNAME RECORD TEST CASES
		// ===============================================================
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
		{
			name: "Invalid CNAME record (malformed target)",
			endpoint: &endpoint.Endpoint{
				DNSName:    "invalid.example.com",
				RecordType: "CNAME",
				Targets:    endpoint.NewTargets("sub...............domain"),
				RecordTTL:  endpoint.TTL(1800),
			},
			expected:    nil,
			expectError: true,
		},

		// ===============================================================
		// TXT RECORD TEST CASES
		// ===============================================================
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

		// ===============================================================
		// MX RECORD TEST CASES
		// ===============================================================
		{
			name: "Valid MX record",
			endpoint: &endpoint.Endpoint{
				DNSName:    "mx.example.com",
				RecordType: "MX",
				Targets:    endpoint.NewTargets("10 mailhost1.example.com"),
				RecordTTL:  endpoint.TTL(600),
			},
			expected: &DNSRecord{
				Name:         "mx.example.com",
				Type:         "MX",
				MXExchange:   "mailhost1.example.com",
				MXPreference: "10",
				TTL:          "10m",
			},
			expectError: false,
		},
		{
			name: "Invalid MX record (empty preference)",
			endpoint: &endpoint.Endpoint{
				DNSName:    "mx.example.com",
				RecordType: "MX",
				Targets:    endpoint.NewTargets(" mailhost1.example.com"),
				RecordTTL:  endpoint.TTL(600),
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid MX record (negative preference)",
			endpoint: &endpoint.Endpoint{
				DNSName:    "mx.example.com",
				RecordType: "MX",
				Targets:    endpoint.NewTargets("-10 mailhost1.example.com"),
				RecordTTL:  endpoint.TTL(600),
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid MX record (too large preference)",
			endpoint: &endpoint.Endpoint{
				DNSName:    "mx.example.com",
				RecordType: "MX",
				Targets:    endpoint.NewTargets("70000 mailhost1.example.com"),
				RecordTTL:  endpoint.TTL(600),
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid MX record (non-numeric preference)",
			endpoint: &endpoint.Endpoint{
				DNSName:    "mx.example.com",
				RecordType: "MX",
				Targets:    endpoint.NewTargets("123 "),
				RecordTTL:  endpoint.TTL(600),
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid MX record (malformed exchange)",
			endpoint: &endpoint.Endpoint{
				DNSName:    "mx.example.com",
				RecordType: "MX",
				Targets:    endpoint.NewTargets("123 sub....domain.com"),
				RecordTTL:  endpoint.TTL(600),
			},
			expected:    nil,
			expectError: true,
		},

		// ===============================================================
		// SRV RECORD TEST CASES
		// ===============================================================
		{
			name: "Valid SRV record",
			endpoint: &endpoint.Endpoint{
				DNSName:    "_sip._tcp.example.com",
				RecordType: "SRV",
				Targets:    endpoint.NewTargets("10 20 5060 sipserver.example.com"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expected: &DNSRecord{
				Name:        "_sip._tcp.example.com",
				Type:        "SRV",
				SrvPriority: "10",
				SrvWeight:   "20",
				SrvPort:     "5060",
				SrvTarget:   "sipserver.example.com",
				TTL:         "1h",
			},
			expectError: false,
		},
		{
			name: "Valid SRV record with lowest priority and weight",
			endpoint: &endpoint.Endpoint{
				DNSName:    "_sip._tcp.example.com",
				RecordType: "SRV",
				Targets:    endpoint.NewTargets("0 0 80 example.com"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expected: &DNSRecord{
				Name:        "_sip._tcp.example.com",
				Type:        "SRV",
				SrvPriority: "0",
				SrvWeight:   "0",
				SrvPort:     "80",
				SrvTarget:   "example.com",
				TTL:         "1h",
			},
			expectError: false,
		},
		{
			name: "Valid SRV record with maximum values",
			endpoint: &endpoint.Endpoint{
				DNSName:    "_sip._tcp.example.com",
				RecordType: "SRV",
				Targets:    endpoint.NewTargets("65535 65535 65535 domain.com"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expected: &DNSRecord{
				Name:        "_sip._tcp.example.com",
				Type:        "SRV",
				SrvPriority: "65535",
				SrvWeight:   "65535",
				SrvPort:     "65535",
				SrvTarget:   ".",
				TTL:         "1h",
			},
			expectError: false,
		},
		{
			name: "Invalid SRV record (negative priority)",
			endpoint: &endpoint.Endpoint{
				DNSName:    "_sip._tcp.example.com",
				RecordType: "SRV",
				Targets:    endpoint.NewTargets("-1 20 80 sipserver.example.com"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid SRV record (negative weight)",
			endpoint: &endpoint.Endpoint{
				DNSName:    "_sip._tcp.example.com",
				RecordType: "SRV",
				Targets:    endpoint.NewTargets("10 -5 80 sipserver.example.com"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid SRV record (port out of range)",
			endpoint: &endpoint.Endpoint{
				DNSName:    "_sip._tcp.example.com",
				RecordType: "SRV",
				Targets:    endpoint.NewTargets("10 20 70000 sipserver.example.com"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid SRV record (empty target)",
			endpoint: &endpoint.Endpoint{
				DNSName:    "_sip._tcp.example.com",
				RecordType: "SRV",
				Targets:    endpoint.NewTargets("10 20 80"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid SRV record (malformed target domain)",
			endpoint: &endpoint.Endpoint{
				DNSName:    "_sip._tcp.example.com",
				RecordType: "SRV",
				Targets:    endpoint.NewTargets("10 20 80 invalid_domain..com"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expected:    nil,
			expectError: true,
		},

		// ===============================================================
		// PROVIDER-SPECIFIC DATA TEST CASES
		// ===============================================================
		{
			name: "Invalid provider-specific configuration (unknown field)",
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
		{
			name: "Multiple provider-specific properties",
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

		// ===============================================================
		// DEFAULT VALUES FOR UNSET FIELDS TEST CASES
		// ===============================================================
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

		// ===============================================================
		// GENERIC ERROR CASES
		// ===============================================================
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
		{
			name: "Unsupported record type",
			endpoint: &endpoint.Endpoint{
				DNSName:    "unsupported.example.com",
				RecordType: "FWD",
				Targets:    endpoint.NewTargets("example.com"),
				RecordTTL:  endpoint.TTL(3600),
			},
			expected:    nil,
			expectError: true,
		},
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
