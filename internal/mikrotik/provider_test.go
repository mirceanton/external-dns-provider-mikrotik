package mikrotik

import (
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
func NewEndpoint(dnsName, target string, ttl int64, providerSpecificProps []map[string]string) *endpoint.Endpoint {
	e := &endpoint.Endpoint{
		DNSName:   dnsName,
		Targets:   endpoint.NewTargets(target),
		RecordTTL: endpoint.TTL(ttl),
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
			&MikrotikDefaults{
				TTL:     defaultTTL,
				Comment: defaultComment,
			},
			nil,
			nil,
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
			endpoint:      NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"comment": "direct-comment"}}),
			property:      "comment",
			expectedValue: "direct-comment",
		},
		{
			name:          "Prefixed property exists",
			provider:      mikrotikProvider,
			endpoint:      NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"webhook/comment": "prefixed-comment"}}),
			property:      "comment",
			expectedValue: "prefixed-comment",
		},
		{
			name:          "Both properties exist - direct takes precedence",
			provider:      mikrotikProvider,
			endpoint:      NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"comment": "direct-comment"}, {"webhook/comment": "prefixed-comment"}}),
			property:      "comment",
			expectedValue: "direct-comment",
		},
		{
			name:          "Property does not exist",
			provider:      mikrotikProvider,
			endpoint:      NewEndpoint("example.com", "192.0.2.1", 3600, nil),
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

func TestCompareEndpoints(t *testing.T) {
	mikrotikProvider := &MikrotikProvider{
		client: &MikrotikApiClient{
			&MikrotikDefaults{
				TTL:     int64(defaultTTL),
				Comment: defaultComment,
			},
			nil,
			nil,
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
			endpointA:     NewEndpoint("example.com", "192.0.2.1", 3600, nil),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", 3600, nil),
			expectedMatch: true,
		},
		{
			name:          "Matching provider-specific",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"comment": "some-comment"}, {"disabled": "true"}}),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"comment": "some-comment"}, {"disabled": "true"}}),
			expectedMatch: true,
		},

		// EDGE CASES
		{
			name:          "Match-Subdomain: 'false' and unspecified should match",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "192.0.2.1", 3600, nil),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"match-subdomain": "false"}}),
			expectedMatch: true,
		},
		{
			name:          "Match-Subdomain: 'false' and empty should match",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"match-subdomain": ""}}),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"match-subdomain": "false"}}),
			expectedMatch: true,
		},
		{
			name:          "Disabled: 'false' and unspecified should match",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "192.0.2.1", 3600, nil),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"disabled": "false"}}),
			expectedMatch: true,
		},
		{
			name:          "Disabled: 'false' and empty should match",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"disabled": ""}}),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"disabled": "false"}}),
			expectedMatch: true,
		},
		{
			name:          "TTL: Default and zero should match",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "192.0.2.1", 0, nil),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", defaultTTL, nil),
			expectedMatch: true,
		},
		{
			name:          "Comment: Default and empty should match",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"comment": ""}}),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"comment": defaultComment}}),
			expectedMatch: true,
		},
		{
			name:          "Comment: Default and unspecified should match",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "192.0.2.1", 3600, nil),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"comment": defaultComment}}),
			expectedMatch: true,
		},

		// MISMATCH CASES
		{
			name:          "Mismatch in DNSName",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "192.0.2.1", 3600, nil),
			endpointB:     NewEndpoint("different.org", "192.0.2.1", 3600, nil),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in Target",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "1.2.3.4", 3600, nil),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", 3600, nil),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in TTL (X != Y)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "192.0.2.1", 5, nil),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", 15, nil),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in TTL (0 != X)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "192.0.2.1", 0, nil),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", 15, nil),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in TTL (Default != X)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "192.0.2.1", defaultTTL, nil),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", 15, nil),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in comment (something != empty)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"comment": "some-comment"}}),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"comment": ""}}),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in comment (something != unspecified)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"comment": "some-comment"}}),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", 3600, nil),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in comment (something != default)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"comment": "some-comment"}}),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"comment": defaultComment}}),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in comment (something != something else)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"comment": "some-comment"}}),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"comment": "other-comment"}}),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in match-subdomain (true != false)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"match-subdomain": "true"}}),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"match-subdomain": "false"}}),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in match-subdomain (true != empty)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"match-subdomain": "true"}}),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"match-subdomain": ""}}),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in match-subdomain (true != unspecified)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"match-subdomain": "true"}}),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", 3600, nil),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in disabled (true != false)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"disabled": "true"}}),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"disabled": "false"}}),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in disabled (true != empty)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"disabled": "true"}}),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"disabled": ""}}),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in disabled (true != unspecified)",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"disabled": "true"}}),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", 3600, nil),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in address-list",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"address-list": "1.2.3.4"}}),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"address-list": "2.3.4.5"}}),
			expectedMatch: false,
		},
		{
			name:          "Mismatch in regexp",
			provider:      mikrotikProvider,
			endpointA:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"regexp": ".*"}}),
			endpointB:     NewEndpoint("example.com", "192.0.2.1", 3600, []map[string]string{{"regexp": "diff.*"}}),
			expectedMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := tt.provider.compareEndpoints(tt.endpointA, tt.endpointB)
			if match != tt.expectedMatch {
				t.Errorf("Expected %v, got %v", tt.expectedMatch, match)
			}
		})
	}
}

func TestListContains(t *testing.T) {
	defaultTTL := 1800
	mikrotikProvider := &MikrotikProvider{
		client: &MikrotikApiClient{
			&MikrotikDefaults{
				TTL: int64(defaultTTL),
			},
			nil,
			nil,
		},
	}
	tests := []struct {
		name          string
		provider      *MikrotikProvider
		haystack      []*endpoint.Endpoint
		needle        *endpoint.Endpoint
		expectContain bool
	}{
		{
			name:     "Needle exists in haystack",
			provider: mikrotikProvider,
			haystack: []*endpoint.Endpoint{
				NewEndpoint("example1.com", "192.0.2.1", 3600, nil),
				NewEndpoint("example2.com", "192.0.2.2", 3600, nil),
				NewEndpoint("example3.com", "192.0.2.3", 3600, nil),
			},
			needle:        NewEndpoint("example2.com", "192.0.2.2", 3600, nil),
			expectContain: true,
		},
		{
			name:     "Needle does not exist in haystack",
			provider: mikrotikProvider,
			haystack: []*endpoint.Endpoint{
				NewEndpoint("example1.com", "192.0.2.1", 3600, nil),
				NewEndpoint("example2.com", "192.0.2.2", 3600, nil),
				NewEndpoint("example3.com", "192.0.2.3", 3600, nil),
			},
			needle:        NewEndpoint("example5.com", "192.0.2.5", 3600, nil),
			expectContain: false,
		},
		{
			name:          "Haystack is empty",
			provider:      mikrotikProvider,
			haystack:      []*endpoint.Endpoint{},
			needle:        NewEndpoint("example5.com", "192.0.2.5", 3600, nil),
			expectContain: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contains := tt.provider.listContains(tt.haystack, tt.needle)
			if contains != tt.expectContain {
				t.Errorf("Expected %v, got %v", tt.expectContain, contains)
			}
		})
	}
}

func TestChanges(t *testing.T) {
	mikrotikProvider := &MikrotikProvider{
		client: &MikrotikApiClient{
			&MikrotikDefaults{
				TTL:     int64(defaultTTL),
				Comment: defaultComment,
			},
			nil,
			nil,
		},
	}

	tests := []struct {
		name            string
		provider        *MikrotikProvider
		inputChanges    *plan.Changes
		expectedChanges *plan.Changes
	}{
		{
			name:     "Multiple matching records - all should be cleaned up",
			provider: mikrotikProvider,
			inputChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					NewEndpoint("example1.com", "192.0.2.1", 3600, []map[string]string{{"comment": "test comment"}, {"disabled": "true"}}),
					NewEndpoint("example2.com", "192.0.2.2", 3600, []map[string]string{{"match-subdomain": "*.example.com"}, {"address-list": "secondary"}}),
				},
				UpdateNew: []*endpoint.Endpoint{
					NewEndpoint("example1.com", "192.0.2.1", 3600, []map[string]string{{"comment": "test comment"}, {"disabled": "true"}}),
					NewEndpoint("example2.com", "192.0.2.2", 3600, []map[string]string{{"match-subdomain": "*.example.com"}, {"address-list": "secondary"}}),
				},
			},
			expectedChanges: &plan.Changes{},
		},
		{
			name:     "Some matching, some different - only partial cleanup",
			provider: mikrotikProvider,
			inputChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					NewEndpoint("matching.com", "1.1.1.1", 3600, []map[string]string{{"comment": "some comment"}}),
					NewEndpoint("different.org", "2.2.2.2", 3600, []map[string]string{{"match-subdomain": "*.old.com"}}),
				},
				UpdateNew: []*endpoint.Endpoint{
					NewEndpoint("matching.com", "1.1.1.1", 3600, []map[string]string{{"comment": "some comment"}}),
					NewEndpoint("different.org", "2.2.2.2", 3600, []map[string]string{{"match-subdomain": "*.new.ro"}}),
				},
			},
			expectedChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					// NewEndpoint("matching.com", "1.1.1.1", 3600, []map[string]string{{"comment": "some comment"}}), // this gets removed because it is the same
					NewEndpoint("different.org", "2.2.2.2", 3600, []map[string]string{{"match-subdomain": "*.old.com"}}),
				},
				UpdateNew: []*endpoint.Endpoint{
					// NewEndpoint("matching.com", "1.1.1.1", 3600, []map[string]string{{"comment": "some comment"}}), // this gets removed because it is the same
					NewEndpoint("different.org", "2.2.2.2", 3600, []map[string]string{{"match-subdomain": "*.new.ro"}}),
				},
			},
		},
		{
			name:     "Different comments across multiple records - no cleanup",
			provider: mikrotikProvider,
			inputChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					NewEndpoint("different.com", "1.1.1.1", 3600, []map[string]string{{"comment": "some comment"}}),
					NewEndpoint("different.org", "2.2.2.2", 3600, []map[string]string{{"match-subdomain": "*.old.com"}}),
				},
				UpdateNew: []*endpoint.Endpoint{
					NewEndpoint("different.com", "1.1.1.1", 3600, []map[string]string{{"comment": "new comment"}}),
					NewEndpoint("different.org", "2.2.2.2", 3600, []map[string]string{{"match-subdomain": "*.new.ro"}}),
				},
			},
			expectedChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					NewEndpoint("different.com", "1.1.1.1", 3600, []map[string]string{{"comment": "some comment"}}),
					NewEndpoint("different.org", "2.2.2.2", 3600, []map[string]string{{"match-subdomain": "*.old.com"}}),
				},
				UpdateNew: []*endpoint.Endpoint{
					NewEndpoint("different.com", "1.1.1.1", 3600, []map[string]string{{"comment": "new comment"}}),
					NewEndpoint("different.org", "2.2.2.2", 3600, []map[string]string{{"match-subdomain": "*.new.ro"}}),
				},
			},
		},
		{
			name:     "Default TTL is enforced on Creation",
			provider: mikrotikProvider,
			inputChanges: &plan.Changes{
				Create: []*endpoint.Endpoint{
					NewEndpoint("zero.com", "1.1.1.1", 0, nil),
					NewEndpoint("default.com", "1.1.1.1", defaultTTL, nil),
					NewEndpoint("something.com", "1.1.1.1", 5, nil),
				},
			},
			expectedChanges: &plan.Changes{
				Create: []*endpoint.Endpoint{
					NewEndpoint("zero.com", "1.1.1.1", defaultTTL, nil),
					NewEndpoint("default.com", "1.1.1.1", defaultTTL, nil),
					NewEndpoint("something.com", "1.1.1.1", 5, nil),
				},
			},
		},
		{
			name:     "Default TTL is enforced on Update",
			provider: mikrotikProvider,
			inputChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					NewEndpoint("zero-to-default.com", "1.1.1.1", 0, nil),
					NewEndpoint("zero-to-something.com", "1.1.1.1", 0, nil),
					NewEndpoint("default-to-zero.com", "1.1.1.1", defaultTTL, nil),
					NewEndpoint("default-to-something.com", "1.1.1.1", defaultTTL, nil),
					NewEndpoint("something-to-zero.com", "1.1.1.1", 5, nil),
					NewEndpoint("something-to-default.com", "1.1.1.1", 5, nil),
					NewEndpoint("something-to-someting-else.com", "1.1.1.1", 5, nil),
				},
				UpdateNew: []*endpoint.Endpoint{
					NewEndpoint("zero-to-default.com", "1.1.1.1", defaultTTL, nil),
					NewEndpoint("zero-to-something.com", "1.1.1.1", 5, nil),
					NewEndpoint("default-to-zero.com", "1.1.1.1", 0, nil),
					NewEndpoint("default-to-something.com", "1.1.1.1", 5, nil),
					NewEndpoint("something-to-zero.com", "1.1.1.1", 0, nil),
					NewEndpoint("something-to-default.com", "1.1.1.1", defaultTTL, nil),
					NewEndpoint("something-to-someting-else.com", "1.1.1.1", 7, nil),
				},
			},
			expectedChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					// NewEndpoint("zero-to-default.com", "1.1.1.1", 0, nil), // removed by filter
					NewEndpoint("zero-to-something.com", "1.1.1.1", 0, nil),
					// NewEndpoint("default-to-zero.com", "1.1.1.1", defaultTTL, nil), // removed by filter
					NewEndpoint("default-to-something.com", "1.1.1.1", defaultTTL, nil),
					NewEndpoint("something-to-zero.com", "1.1.1.1", 5, nil),
					NewEndpoint("something-to-default.com", "1.1.1.1", 5, nil),
					NewEndpoint("something-to-someting-else.com", "1.1.1.1", 5, nil),
				},
				UpdateNew: []*endpoint.Endpoint{
					// NewEndpoint("zero-to-default.com", "1.1.1.1", defaultTTL, nil), // removed by filter
					NewEndpoint("zero-to-something.com", "1.1.1.1", 5, nil),
					// NewEndpoint("default-to-zero.com", "1.1.1.1", 0, nil), // removed by filter
					NewEndpoint("default-to-something.com", "1.1.1.1", 5, nil),
					NewEndpoint("something-to-zero.com", "1.1.1.1", 0, nil),
					NewEndpoint("something-to-default.com", "1.1.1.1", defaultTTL, nil),
					NewEndpoint("something-to-someting-else.com", "1.1.1.1", 7, nil),
				},
			},
		},
		{
			name:     "Default Comment is enforced on Creation",
			provider: mikrotikProvider,
			inputChanges: &plan.Changes{
				Create: []*endpoint.Endpoint{
					NewEndpoint("unset.com", "1.1.1.1", 5, nil),
					NewEndpoint("empty.com", "1.1.1.1", 5, []map[string]string{{"comment": ""}}),
					NewEndpoint("default.com", "1.1.1.1", 5, []map[string]string{{"comment": defaultComment}}),
					NewEndpoint("something.com", "1.1.1.1", 5, []map[string]string{{"comment": "something"}}),
				},
			},
			expectedChanges: &plan.Changes{
				Create: []*endpoint.Endpoint{
					NewEndpoint("unset.com", "1.1.1.1", 5, []map[string]string{{"comment": defaultComment}}),
					NewEndpoint("empty.com", "1.1.1.1", 5, []map[string]string{{"comment": defaultComment}}),
					NewEndpoint("default.com", "1.1.1.1", 5, []map[string]string{{"comment": defaultComment}}),
					NewEndpoint("something.com", "1.1.1.1", 5, []map[string]string{{"comment": "something"}}),
				},
			},
		},
		{
			name:     "Default Comment is enforced on Update",
			provider: mikrotikProvider,
			inputChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					NewEndpoint("unset-to-empty.com", "1.1.1.1", 0, nil),
					NewEndpoint("unset-to-default.com", "1.1.1.1", 0, nil),
					NewEndpoint("unset-to-something.com", "1.1.1.1", 0, nil),

					NewEndpoint("empty-to-unset.com", "1.1.1.1", 0, []map[string]string{{"comment": ""}}),
					NewEndpoint("empty-to-default.com", "1.1.1.1", 0, []map[string]string{{"comment": ""}}),
					NewEndpoint("empty-to-something.com", "1.1.1.1", 0, []map[string]string{{"comment": ""}}),

					NewEndpoint("default-to-unset.com", "1.1.1.1", 0, []map[string]string{{"comment": defaultComment}}),
					NewEndpoint("default-to-empty.com", "1.1.1.1", 0, []map[string]string{{"comment": defaultComment}}),
					NewEndpoint("default-to-something.com", "1.1.1.1", 0, []map[string]string{{"comment": defaultComment}}),

					NewEndpoint("something-to-unset.com", "1.1.1.1", 5, []map[string]string{{"comment": "something"}}),
					NewEndpoint("something-to-empty.com", "1.1.1.1", 5, []map[string]string{{"comment": "something"}}),
					NewEndpoint("something-to-default.com", "1.1.1.1", 5, []map[string]string{{"comment": "something"}}),
					NewEndpoint("something-to-someting-else.com", "1.1.1.1", 5, []map[string]string{{"comment": "something"}}),
				},
				UpdateNew: []*endpoint.Endpoint{
					NewEndpoint("unset-to-empty.com", "1.1.1.1", 0, []map[string]string{{"comment": ""}}),
					NewEndpoint("unset-to-default.com", "1.1.1.1", 0, []map[string]string{{"comment": defaultComment}}),
					NewEndpoint("unset-to-something.com", "1.1.1.1", 0, []map[string]string{{"comment": "something"}}),

					NewEndpoint("empty-to-unset.com", "1.1.1.1", 0, []map[string]string{{"comment": ""}}),
					NewEndpoint("empty-to-default.com", "1.1.1.1", 0, []map[string]string{{"comment": defaultComment}}),
					NewEndpoint("empty-to-something.com", "1.1.1.1", 0, []map[string]string{{"comment": "something"}}),

					NewEndpoint("default-to-unset.com", "1.1.1.1", 0, []map[string]string{{"comment": defaultComment}}),
					NewEndpoint("default-to-empty.com", "1.1.1.1", 0, []map[string]string{{"comment": ""}}),
					NewEndpoint("default-to-something.com", "1.1.1.1", 0, []map[string]string{{"comment": "something"}}),

					NewEndpoint("something-to-unset.com", "1.1.1.1", 5, nil),
					NewEndpoint("something-to-empty.com", "1.1.1.1", 5, []map[string]string{{"comment": ""}}),
					NewEndpoint("something-to-default.com", "1.1.1.1", 5, []map[string]string{{"comment": defaultComment}}),
					NewEndpoint("something-to-someting-else.com", "1.1.1.1", 5, []map[string]string{{"comment": "something-else"}}),
				},
			},
			expectedChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					// NewEndpoint("unset-to-empty.com", "1.1.1.1", 0, nil),
					// NewEndpoint("unset-to-default.com", "1.1.1.1", 0, nil),
					NewEndpoint("unset-to-something.com", "1.1.1.1", 0, nil),

					// NewEndpoint("empty-to-unset.com", "1.1.1.1", 0, []map[string]string{{"comment": ""}}),
					// NewEndpoint("empty-to-default.com", "1.1.1.1", 0, []map[string]string{{"comment": ""}}),
					NewEndpoint("empty-to-something.com", "1.1.1.1", 0, []map[string]string{{"comment": ""}}),

					// NewEndpoint("default-to-unset.com", "1.1.1.1", 0, []map[string]string{{"comment": defaultComment}}),
					// NewEndpoint("default-to-empty.com", "1.1.1.1", 0, []map[string]string{{"comment": defaultComment}}),
					NewEndpoint("default-to-something.com", "1.1.1.1", 0, []map[string]string{{"comment": defaultComment}}),

					NewEndpoint("something-to-unset.com", "1.1.1.1", 5, []map[string]string{{"comment": "something"}}),
					NewEndpoint("something-to-empty.com", "1.1.1.1", 5, []map[string]string{{"comment": "something"}}),
					NewEndpoint("something-to-default.com", "1.1.1.1", 5, []map[string]string{{"comment": "something"}}),
					NewEndpoint("something-to-someting-else.com", "1.1.1.1", 5, []map[string]string{{"comment": "something"}}),
				},
				UpdateNew: []*endpoint.Endpoint{
					// NewEndpoint("unset-to-empty.com", "1.1.1.1", 0, []map[string]string{{"comment": ""}}),
					// NewEndpoint("unset-to-default.com", "1.1.1.1", 0, []map[string]string{{"comment": defaultComment}}),
					NewEndpoint("unset-to-something.com", "1.1.1.1", 0, []map[string]string{{"comment": "something"}}),

					// NewEndpoint("empty-to-unset.com", "1.1.1.1", 0, []map[string]string{{"comment": ""}}),
					// NewEndpoint("empty-to-default.com", "1.1.1.1", 0, []map[string]string{{"comment": defaultComment}}),
					NewEndpoint("empty-to-something.com", "1.1.1.1", 0, []map[string]string{{"comment": "something"}}),

					// NewEndpoint("default-to-unset.com", "1.1.1.1", 0, []map[string]string{{"comment": defaultComment}}),
					// NewEndpoint("default-to-empty.com", "1.1.1.1", 0, []map[string]string{{"comment": ""}}),
					NewEndpoint("default-to-something.com", "1.1.1.1", 0, []map[string]string{{"comment": "something"}}),

					NewEndpoint("something-to-unset.com", "1.1.1.1", 5, []map[string]string{{"comment": defaultComment}}),
					NewEndpoint("something-to-empty.com", "1.1.1.1", 5, []map[string]string{{"comment": defaultComment}}),
					NewEndpoint("something-to-default.com", "1.1.1.1", 5, []map[string]string{{"comment": defaultComment}}),
					NewEndpoint("something-to-someting-else.com", "1.1.1.1", 5, []map[string]string{{"comment": "something-else"}}),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputChanges := tt.provider.changes(tt.inputChanges)

			if len(outputChanges.UpdateOld) != len(tt.expectedChanges.UpdateOld) {
				t.Errorf("Expected UpdateOld length %d, got %d", len(tt.expectedChanges.UpdateOld), len(outputChanges.UpdateOld))
			}
			if len(outputChanges.UpdateNew) != len(tt.expectedChanges.UpdateNew) {
				t.Errorf("Expected UpdateNew length %d, got %d", len(tt.expectedChanges.UpdateNew), len(outputChanges.UpdateNew))
			}
			if len(outputChanges.Create) != len(tt.expectedChanges.Create) {
				t.Errorf("Expected Create length %d, got %d", len(tt.expectedChanges.Create), len(outputChanges.Create))
			}

			for i := range tt.expectedChanges.UpdateOld {
				if !mikrotikProvider.compareEndpoints(outputChanges.UpdateOld[i], tt.expectedChanges.UpdateOld[i]) {
					t.Errorf("Expected UpdateOld endpoint: %v , got %v", tt.expectedChanges.UpdateOld[i], outputChanges.UpdateOld[i])
				}
			}
			for i := range tt.expectedChanges.UpdateNew {
				if !mikrotikProvider.compareEndpoints(outputChanges.UpdateNew[i], tt.expectedChanges.UpdateNew[i]) {
					t.Errorf("Expected UpdateNew endpoint: %v , got %v", tt.expectedChanges.UpdateNew[i], outputChanges.UpdateNew[i])
				}
			}
			for i := range outputChanges.Create {
				if !mikrotikProvider.compareEndpoints(outputChanges.Create[i], tt.expectedChanges.Create[i]) {
					t.Errorf("Expected Create endpoint TTL %d, got %d", 0, outputChanges.Create[i].RecordTTL)
				}
			}
		})
	}
}
