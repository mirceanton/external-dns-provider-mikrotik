package mikrotik

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
)

const (
	defaultTTL     = 1800
	defaultComment = "default comment"
	defaultPS      = "default"
)

// Helper function to create endpoints for brevity
func NewEndpoint(dnsName string, targets []string, ttl int64, providerSpecificProps []map[string]string) *endpoint.Endpoint {
	e := &endpoint.Endpoint{
		DNSName:    dnsName,
		Targets:    endpoint.NewTargets(targets...),
		RecordType: "A", // Default to A record type
		RecordTTL:  endpoint.TTL(ttl),
	}
	for _, prop := range providerSpecificProps {
		for key, value := range prop {
			e.SetProviderSpecificProperty(key, value)
		}
	}
	return e
}

// Helper function to create endpoints with custom record type
func NewEndpointWithType(dnsName, target, recordType string, ttl int64, providerSpecificProps []map[string]string) *endpoint.Endpoint {
	e := &endpoint.Endpoint{
		DNSName:    dnsName,
		Targets:    endpoint.NewTargets(target),
		RecordType: recordType,
		RecordTTL:  endpoint.TTL(ttl),
	}
	for _, prop := range providerSpecificProps {
		for key, value := range prop {
			e.SetProviderSpecificProperty(key, value)
		}
	}
	return e
}

func TestGetProviderSpecificOrDefault(t *testing.T) {
	mikrotikProvider := &MikrotikProvider{
		client: &MikrotikApiClient{
			MikrotikDefaults: &MikrotikDefaults{
				DefaultTTL:     defaultTTL,
				DefaultComment: defaultComment,
			},
			MikrotikConnectionConfig: nil,
			Client:                   nil,
		},
	}
	tests := []struct {
		name          string
		provider      *MikrotikProvider
		endpoint      *endpoint.Endpoint
		property      string
		expectedValue string
	}{
		{
			name:          "Direct property exists",
			provider:      mikrotikProvider,
			endpoint:      NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"comment": "direct-comment"}}),
			property:      "comment",
			expectedValue: "direct-comment",
		},
		{
			name:          "Prefixed property exists",
			provider:      mikrotikProvider,
			endpoint:      NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"webhook/comment": "prefixed-comment"}}),
			property:      "comment",
			expectedValue: "prefixed-comment",
		},
		{
			name:          "Both properties exist - direct takes precedence",
			provider:      mikrotikProvider,
			endpoint:      NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"comment": "direct-comment"}, {"webhook/comment": "prefixed-comment"}}),
			property:      "comment",
			expectedValue: "direct-comment",
		},
		{
			name:          "Property does not exist",
			provider:      mikrotikProvider,
			endpoint:      NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, nil),
			property:      "comment",
			expectedValue: defaultPS,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := tt.provider.getProviderSpecificOrDefault(tt.endpoint, tt.property, defaultPS)
			if value != tt.expectedValue {
				t.Errorf("Expected %q, got %q", tt.expectedValue, value)
			}
		})
	}
}

func TestCompareEndpointsBesidesTargets(t *testing.T) {
	mikrotikProvider := &MikrotikProvider{
		client: &MikrotikApiClient{
			MikrotikDefaults: &MikrotikDefaults{
				DefaultTTL:     int64(defaultTTL),
				DefaultComment: defaultComment,
			},
			MikrotikConnectionConfig: nil,
			Client:                   nil,
		},
	}
	tests := []struct {
		name          string
		provider      *MikrotikProvider
		endpointA     *endpoint.Endpoint
		endpointB     *endpoint.Endpoint
		expectedMatch bool
	}{
		// MATCHING CASES
		{
			name:          "Matching basic properties",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, nil),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, nil),
			expectedMatch: true,
		},
		{
			name:          "Matching provider-specific",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"comment": "some-comment"}, {"disabled": "true"}}),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"comment": "some-comment"}, {"disabled": "true"}}),
			expectedMatch: true,
		},

		// EDGE CASES
		{
			name:          "Match-Subdomain: 'false' and unspecified should match",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, nil),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"match-subdomain": "false"}}),
			expectedMatch: true,
		},
		{
			name:          "Match-Subdomain: 'false' and empty should match",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"match-subdomain": ""}}),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"match-subdomain": "false"}}),
			expectedMatch: true,
		},
		{
			name:          "Disabled: 'false' and unspecified should match",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, nil),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"disabled": "false"}}),
			expectedMatch: true,
		},
		{
			name:          "Disabled: 'false' and empty should match",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"disabled": ""}}),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"disabled": "false"}}),
			expectedMatch: true,
		},
		{
			name:          "TTL: Default and zero should match",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, 0, nil),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, defaultTTL, nil),
			expectedMatch: true,
		},
		{
			name:          "Comment: Default and empty should match",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"comment": ""}}),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"comment": defaultComment}}),
			expectedMatch: true,
		},
		{
			name:          "Comment: Default and unspecified should match",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, nil),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"comment": defaultComment}}),
			expectedMatch: true,
		},

		// MISMATCH CASES
		{
			name:          "Mismatch in DNSName",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, nil),
			endpointB:     NewEndpoint("different.org", []string{"192.0.2.1"}, 3600, nil),
			expectedMatch: false,
		},
		{
			name:          "Should ignore mismatch in Target",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"1.2.3.4"}, 3600, nil),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, nil),
			expectedMatch: true,
		},
		{
			name:          "Mismatch in TTL (X != Y)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, 5, nil),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, 15, nil),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in TTL (0 != X)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, 0, nil),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, 15, nil),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in TTL (Default != X)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, defaultTTL, nil),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, 15, nil),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in comment (something != empty)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"comment": "some-comment"}}),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"comment": ""}}),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in comment (something != unspecified)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"comment": "some-comment"}}),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, nil),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in comment (something != default)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"comment": "some-comment"}}),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"comment": defaultComment}}),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in comment (something != something else)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"comment": "some-comment"}}),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"comment": "other-comment"}}),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in match-subdomain (true != false)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"match-subdomain": "true"}}),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"match-subdomain": "false"}}),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in match-subdomain (true != empty)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"match-subdomain": "true"}}),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"match-subdomain": ""}}),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in match-subdomain (true != unspecified)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"match-subdomain": "true"}}),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, nil),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in disabled (true != false)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"disabled": "true"}}),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"disabled": "false"}}),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in disabled (true != empty)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"disabled": "true"}}),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"disabled": ""}}),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in disabled (true != unspecified)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"disabled": "true"}}),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, nil),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in address-list",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"address-list": "1.2.3.4"}}),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"address-list": "2.3.4.5"}}),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in regexp",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"regexp": ".*"}}),
			endpointB:     NewEndpoint("example.com", []string{"192.0.2.1"}, 3600, []map[string]string{{"regexp": "diff.*"}}),
			expectedMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := tt.provider.compareEndpointsBesidesTargets(tt.endpointA, tt.endpointB)
			if match != tt.expectedMatch {
				t.Errorf("Expected %v, got %v", tt.expectedMatch, match)
			}
		})
	}
}

func TestNewMikrotikProvider(t *testing.T) {
	mockServerInfo := MikrotikSystemInfo{
		ArchitectureName: "arm64",
		BoardName:        "RB5009UG+S+",
		Version:          "7.16 (stable)",
	}

	tests := []struct {
		name          string
		expectError   bool
		simulateError bool
	}{
		{
			name:          "Successful provider creation",
			expectError:   false,
			simulateError: false,
		},
		{
			name:          "Failed to connect to API",
			expectError:   true,
			simulateError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulate connection error
				if tt.simulateError {
					http.Error(w, "Connection failed", http.StatusInternalServerError)
					return
				}

				// Basic Auth validation
				username, password, ok := r.BasicAuth()
				if !ok || username != "testuser" || password != "testpass" {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				// Return system info for /rest/system/resource
				if r.URL.Path == "/rest/system/resource" && r.Method == http.MethodGet {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(mockServerInfo)
					return
				}

				http.NotFound(w, r)
			}))
			defer server.Close()

			domainFilter := endpoint.NewDomainFilter([]string{"example.com"})
			defaults := &MikrotikDefaults{
				DefaultTTL:     3600,
				DefaultComment: "external-dns",
			}
			config := &MikrotikConnectionConfig{
				BaseUrl:       server.URL,
				Username:      "testuser",
				Password:      "testpass",
				SkipTLSVerify: true,
			}

			provider, err := NewMikrotikProvider(domainFilter, defaults, config)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error, got none")
				}
				if provider != nil {
					t.Errorf("Expected nil provider on error, got %v", provider)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
				if provider == nil {
					t.Fatal("Expected provider, got nil")
				}

				// Verify provider properties
				mikrotikProvider, ok := provider.(*MikrotikProvider)
				if !ok {
					t.Fatal("Expected MikrotikProvider type")
				}
				if mikrotikProvider.client == nil {
					t.Error("Expected client to be set")
				}
				if mikrotikProvider.domainFilter != domainFilter {
					t.Error("Expected domainFilter to be set correctly")
				}
			}
		})
	}
}

func TestMikrotikProvider_Records(t *testing.T) {
	mockRecords := []DNSRecord{
		{
			ID:      "*1",
			Name:    "example.com",
			Type:    "A",
			Address: "1.2.3.4",
			Comment: "external-dns",
			TTL:     "1h",
		},
		{
			ID:      "*2",
			Name:    "test.example.com",
			Type:    "CNAME",
			CName:   "target.example.com",
			Comment: "external-dns",
			TTL:     "30m",
		},
		{
			ID:      "*3",
			Name:    "other.org",
			Type:    "A",
			Address: "5.6.7.8",
			Comment: "external-dns",
			TTL:     "2h",
		},
	}

	tests := []struct {
		name              string
		domainFilter      []string
		expectError       bool
		simulateAPIError  bool
		expectedEndpoints int
	}{
		{
			name:              "Successful records retrieval with domain filtering",
			domainFilter:      []string{"example.com"},
			expectError:       false,
			simulateAPIError:  false,
			expectedEndpoints: 2, // Only example.com and test.example.com should match
		},
		{
			name:              "Successful records retrieval without filtering",
			domainFilter:      []string{},
			expectError:       false,
			simulateAPIError:  false,
			expectedEndpoints: 3, // All records should be returned
		},
		{
			name:              "API error during records retrieval",
			domainFilter:      []string{"example.com"},
			expectError:       true,
			simulateAPIError:  true,
			expectedEndpoints: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulate API error
				if tt.simulateAPIError {
					http.Error(w, "API Error", http.StatusInternalServerError)
					return
				}

				// Basic Auth validation
				username, password, ok := r.BasicAuth()
				if !ok || username != "testuser" || password != "testpass" {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				// Return DNS records for /rest/ip/dns/static
				if r.URL.Path == "/rest/ip/dns/static" && r.Method == http.MethodGet {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(mockRecords)
					return
				}

				http.NotFound(w, r)
			}))
			defer server.Close()

			// Create provider
			domainFilter := endpoint.NewDomainFilter(tt.domainFilter)
			defaults := &MikrotikDefaults{
				DefaultTTL:     3600,
				DefaultComment: "external-dns",
			}
			config := &MikrotikConnectionConfig{
				BaseUrl:       server.URL,
				Username:      "testuser",
				Password:      "testpass",
				SkipTLSVerify: true,
			}

			client, err := NewMikrotikClient(config, defaults)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			provider := &MikrotikProvider{
				client:       client,
				domainFilter: domainFilter,
			}

			// Test Records method
			ctx := context.Background()
			endpoints, err := provider.Records(ctx)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error, got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
				if len(endpoints) != tt.expectedEndpoints {
					t.Errorf("Expected %d endpoints, got %d", tt.expectedEndpoints, len(endpoints))
				}

				// Verify domain filtering worked correctly
				for _, ep := range endpoints {
					if len(tt.domainFilter) > 0 && !domainFilter.Match(ep.DNSName) {
						t.Errorf("Endpoint %s should have been filtered out", ep.DNSName)
					}
				}
			}
		})
	}
}

func TestMikrotikProvider_ApplyChanges(t *testing.T) {
	tests := []struct {
		name                string
		changes             *plan.Changes
		expectError         bool
		simulateAPIError    bool
		expectedPutCalls    int // Expected number of PUT (create) calls
		expectedDeleteCalls int // Expected number of DELETE calls
	}{
		{
			name: "Successful create operation",
			changes: &plan.Changes{
				Create: []*endpoint.Endpoint{
					NewEndpoint("new.example.com", []string{"1.1.1.1"}, 3600, nil),
				},
			},
			expectError:         false,
			simulateAPIError:    false,
			expectedPutCalls:    1,
			expectedDeleteCalls: 0,
		},
		{
			name: "Successful delete operation",
			changes: &plan.Changes{
				Delete: []*endpoint.Endpoint{
					NewEndpoint("delete.example.com", []string{"1.1.1.1"}, 3600, nil),
				},
			},
			expectError:         false,
			simulateAPIError:    false,
			expectedPutCalls:    0,
			expectedDeleteCalls: 1,
		},
		{
			name: "Successful update operation",
			changes: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					NewEndpoint("update.example.com", []string{"1.1.1.1"}, 3600, nil),
				},
				UpdateNew: []*endpoint.Endpoint{
					NewEndpoint("update.example.com", []string{"2.2.2.2"}, 3600, nil),
				},
			},
			expectError:         false,
			simulateAPIError:    false,
			expectedPutCalls:    1, // Smart update might create new targets
			expectedDeleteCalls: 1, // Smart update might delete old targets
		},
		{
			name: "Domain filter security violation",
			changes: &plan.Changes{
				Create: []*endpoint.Endpoint{
					NewEndpoint("malicious.attacker.com", []string{"1.1.1.1"}, 3600, nil),
				},
			},
			expectError:         true,
			simulateAPIError:    false,
			expectedPutCalls:    0,
			expectedDeleteCalls: 0,
		},
		{
			name: "Update with overlapping records - should skip identical ones",
			changes: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					// This record is identical in both old and new - should be skipped
					NewEndpoint("identical.example.com", []string{"1.1.1.1"}, 3600, []map[string]string{{"comment": "same-comment"}}),
					// This record actually changes targets - should be processed
					NewEndpoint("changing.example.com", []string{"1.1.1.1"}, 3600, []map[string]string{{"comment": "old-comment"}}),
				},
				UpdateNew: []*endpoint.Endpoint{
					// This record is identical to the old one - should be skipped
					NewEndpoint("identical.example.com", []string{"1.1.1.1"}, 3600, []map[string]string{{"comment": "same-comment"}}),
					// This record has different target - should be processed
					NewEndpoint("changing.example.com", []string{"2.2.2.2"}, 3600, []map[string]string{{"comment": "new-comment"}}),
				},
			},
			expectError:         false,
			simulateAPIError:    false,
			expectedPutCalls:    1, // smartUpdate creates new target - note: the identical record should be filtered out by changes()
			expectedDeleteCalls: 1, // Note: DELETE might be 0 if no existing record is found to delete
		},
		{
			name: "Update with all identical records - should skip all operations",
			changes: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					NewEndpoint("same1.example.com", []string{"1.1.1.1"}, 3600, []map[string]string{{"comment": "test"}}),
					NewEndpoint("same2.example.com", []string{"2.2.2.2"}, 1800, []map[string]string{{"disabled": "false"}}),
					NewEndpoint("multi2.example.com", []string{"1.1.1.1", "2.2.2.2"}, 1800, []map[string]string{{"disabled": "false"}}),
				},
				UpdateNew: []*endpoint.Endpoint{
					NewEndpoint("same1.example.com", []string{"1.1.1.1"}, 3600, []map[string]string{{"comment": "test"}}),
					NewEndpoint("same2.example.com", []string{"2.2.2.2"}, 1800, []map[string]string{{"disabled": "false"}}),
					NewEndpoint("multi2.example.com", []string{"1.1.1.1", "2.2.2.2"}, 1800, []map[string]string{{"disabled": "false"}}),
				},
			},
			expectError:         false,
			simulateAPIError:    false,
			expectedPutCalls:    0,
			expectedDeleteCalls: 0,
		},
		{
			name: "Update multiple targets",
			changes: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					NewEndpoint("multi1.example.com", []string{"1.1.1.1", "2.2.2.2"}, 3600, []map[string]string{{"comment": "test"}}),
				},
				UpdateNew: []*endpoint.Endpoint{
					NewEndpoint("multi1.example.com", []string{"2.2.2.2", "3.3.3.3", "4.4.4.4"}, 3600, []map[string]string{{"comment": "test"}}),
				},
			},
			expectError:         false,
			simulateAPIError:    false,
			expectedPutCalls:    2,
			expectedDeleteCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Counters to track API calls
			var putCallCount, deleteCallCount int

			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulate API error
				if tt.simulateAPIError {
					http.Error(w, "API Error", http.StatusInternalServerError)
					return
				}

				// Basic Auth validation
				username, password, ok := r.BasicAuth()
				if !ok || username != "testuser" || password != "testpass" {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				// Mock DNS records for GET requests (needed for delete and update)
				if r.URL.Path == "/rest/ip/dns/static" && r.Method == http.MethodGet {
					allRecords := []DNSRecord{
						{
							ID:      "*1",
							Name:    "delete.example.com",
							Type:    "A",
							Address: "2.2.2.2",
							Comment: "external-dns",
						},
						{
							ID:      "*1b",
							Name:    "delete.example.com",
							Type:    "A",
							Address: "1.1.1.1",
							Comment: "external-dns",
						},
						{
							ID:      "*2",
							Name:    "update.example.com",
							Type:    "A",
							Address: "1.1.1.1",
							Comment: "external-dns",
						},
						{
							ID:      "*3",
							Name:    "identical.example.com",
							Type:    "A",
							Address: "1.1.1.1",
							Comment: "same-comment",
						},
						{
							ID:      "*4",
							Name:    "changing.example.com",
							Type:    "A",
							Address: "1.1.1.1",
							Comment: "old-comment",
						},
						{
							ID:      "*5",
							Name:    "same1.example.com",
							Type:    "A",
							Address: "1.1.1.1",
							Comment: "test",
						},
						{
							ID:      "*6",
							Name:    "same2.example.com",
							Type:    "A",
							Address: "2.2.2.2",
							Comment: "external-dns",
						},
						{
							ID:      "*7",
							Name:    "multi1.example.com",
							Type:    "A",
							Address: "1.1.1.1",
							Comment: "external-dns",
						},
						{
							ID:      "*8",
							Name:    "multi1.example.com",
							Type:    "A",
							Address: "2.2.2.2",
							Comment: "external-dns",
						},
						{
							ID:      "*9",
							Name:    "multi2.example.com",
							Type:    "A",
							Address: "1.1.1.1",
							Comment: "external-dns",
						},
						{
							ID:      "*10",
							Name:    "multi2.example.com",
							Type:    "A",
							Address: "2.2.2.2",
							Comment: "external-dns",
						},
					}

					// Filter records based on query parameters
					query := r.URL.Query()
					nameFilter := query.Get("name")
					commentFilter := query.Get("comment")

					var mockRecords []DNSRecord
					for _, record := range allRecords {
						// Apply name filter if specified
						if nameFilter != "" && record.Name != nameFilter {
							continue
						}
						// Apply comment filter if specified
						if commentFilter != "" && record.Comment != commentFilter {
							continue
						}
						mockRecords = append(mockRecords, record)
					}

					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(mockRecords)
					return
				}

				// Handle PUT requests (create)
				if r.Method == http.MethodPut && r.URL.Path == "/rest/ip/dns/static" {
					putCallCount++
					var record DNSRecord
					json.NewDecoder(r.Body).Decode(&record)
					record.ID = "*new"
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(record)
					return
				}

				// Handle DELETE requests
				if r.Method == http.MethodDelete {
					deleteCallCount++
					w.WriteHeader(http.StatusOK)
					return
				}

				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			// Create provider with domain filter
			domainFilter := endpoint.NewDomainFilter([]string{"example.com"})
			defaults := &MikrotikDefaults{
				DefaultTTL:     3600,
				DefaultComment: "external-dns",
			}
			config := &MikrotikConnectionConfig{
				BaseUrl:       server.URL,
				Username:      "testuser",
				Password:      "testpass",
				SkipTLSVerify: true,
			}

			client, err := NewMikrotikClient(config, defaults)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			provider := &MikrotikProvider{
				client:       client,
				domainFilter: domainFilter,
			}

			// Test ApplyChanges method
			ctx := context.Background()
			err = provider.ApplyChanges(ctx, tt.changes)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error, got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}

				// Verify API call counts to ensure duplicates were skipped
				if putCallCount != tt.expectedPutCalls {
					t.Errorf("Expected %d PUT calls, got %d", tt.expectedPutCalls, putCallCount)
				}
				if deleteCallCount != tt.expectedDeleteCalls {
					t.Errorf("Expected %d DELETE calls, got %d", tt.expectedDeleteCalls, deleteCallCount)
				}
			}
		})
	}
}
