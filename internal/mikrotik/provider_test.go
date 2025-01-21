package mikrotik

import (
	"testing"

	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
)

func TestGetProviderSpecificOrDefault(t *testing.T) {
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
		endpoint      *endpoint.Endpoint
		property      string
		expectedValue string
	}{
		{
			name:     "Direct property exists",
			provider: mikrotikProvider,
			endpoint: &endpoint.Endpoint{
				ProviderSpecific: endpoint.ProviderSpecific{
					{Name: "comment", Value: "direct-comment"},
				},
			},
			property:      "comment",
			expectedValue: "direct-comment",
		},
		{
			name:     "Prefixed property exists",
			provider: mikrotikProvider,
			endpoint: &endpoint.Endpoint{
				ProviderSpecific: endpoint.ProviderSpecific{
					{Name: "webhook/comment", Value: "prefixed-comment"},
				},
			},
			property:      "comment",
			expectedValue: "prefixed-comment",
		},
		{
			name:     "Both properties exist - direct takes precedence",
			provider: mikrotikProvider,
			endpoint: &endpoint.Endpoint{
				ProviderSpecific: endpoint.ProviderSpecific{
					{Name: "comment", Value: "direct-comment"},
					{Name: "webhook/comment", Value: "prefixed-comment"},
				},
			},
			property:      "comment",
			expectedValue: "direct-comment",
		},
		{
			name:     "Neither property exists",
			provider: mikrotikProvider,
			endpoint: &endpoint.Endpoint{
				ProviderSpecific: endpoint.ProviderSpecific{},
			},
			property:      "comment",
			expectedValue: "",
		},
		{
			name:     "Weong key selected",
			provider: mikrotikProvider,
			endpoint: &endpoint.Endpoint{
				ProviderSpecific: endpoint.ProviderSpecific{
					{Name: "comment", Value: "direct-comment"},
				},
			},
			property:      "address-list",
			expectedValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := tt.provider.getProviderSpecificOrDefault(tt.endpoint, tt.property, "")
			if value != tt.expectedValue {
				t.Errorf("Expected %q, got %q", tt.expectedValue, value)
			}
		})
	}
}

func TestCompareEndpoints(t *testing.T) {
	defaultTTL := 1800
	defaultComment := "test comment"
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
			name:     "Matching basic properties",
			provider: mikrotikProvider,
			endpointA: &endpoint.Endpoint{
				DNSName:   "example.com",
				Targets:   endpoint.NewTargets("192.0.2.1"),
				RecordTTL: endpoint.TTL(3600),
			},
			endpointB: &endpoint.Endpoint{
				DNSName:   "example.com",
				Targets:   endpoint.NewTargets("192.0.2.1"),
				RecordTTL: endpoint.TTL(3600),
			},
			expectedMatch: true,
		},
		{
			name:     "Matching provider-specific",
			provider: mikrotikProvider,
			endpointA: &endpoint.Endpoint{
				DNSName:   "example.com",
				Targets:   endpoint.NewTargets("192.0.2.1"),
				RecordTTL: endpoint.TTL(3600),
				ProviderSpecific: endpoint.ProviderSpecific{
					{Name: "comment", Value: "match"},
					{Name: "match-subdomain", Value: "true"},
					{Name: "address-list", Value: "default"},
					{Name: "regexp", Value: ".*"},
				},
			},
			endpointB: &endpoint.Endpoint{
				DNSName:   "example.com",
				Targets:   endpoint.NewTargets("192.0.2.1"),
				RecordTTL: endpoint.TTL(3600),
				ProviderSpecific: endpoint.ProviderSpecific{
					{Name: "webhook/comment", Value: "match"},
					{Name: "webhook/match-subdomain", Value: "true"},
					{Name: "webhook/address-list", Value: "default"},
					{Name: "webhook/regexp", Value: ".*"},
				},
			},
			expectedMatch: true,
		},

		// EDGE CASES
		{
			name:     "Match-Subdomain: 'false' and unspecified should match",
			provider: mikrotikProvider,
			endpointA: &endpoint.Endpoint{
				DNSName:   "example.com",
				Targets:   endpoint.NewTargets("192.0.2.1"),
				RecordTTL: endpoint.TTL(3600),
				ProviderSpecific: endpoint.ProviderSpecific{
					{Name: "match-subdomain", Value: "false"},
				},
			},
			endpointB: &endpoint.Endpoint{
				DNSName:          "example.com",
				Targets:          endpoint.NewTargets("192.0.2.1"),
				RecordTTL:        endpoint.TTL(3600),
				ProviderSpecific: endpoint.ProviderSpecific{}, // unspecified match-subdomain
			},
			expectedMatch: true,
		},
		{
			name:     "Disabled: 'false' and unspecified should match",
			provider: mikrotikProvider,
			endpointA: &endpoint.Endpoint{
				DNSName:   "example.com",
				Targets:   endpoint.NewTargets("192.0.2.1"),
				RecordTTL: endpoint.TTL(3600),
				ProviderSpecific: endpoint.ProviderSpecific{
					{Name: "disabled", Value: "false"},
				},
			},
			endpointB: &endpoint.Endpoint{
				DNSName:          "example.com",
				Targets:          endpoint.NewTargets("192.0.2.1"),
				RecordTTL:        endpoint.TTL(3600),
				ProviderSpecific: endpoint.ProviderSpecific{}, // unspecified disabled
			},
			expectedMatch: true,
		},
		{
			name:     "0 TTL and Default TTL should match",
			provider: mikrotikProvider,
			endpointA: &endpoint.Endpoint{
				DNSName:   "example.com",
				Targets:   endpoint.NewTargets("192.0.2.1"),
				RecordTTL: endpoint.TTL(0),
			},
			endpointB: &endpoint.Endpoint{
				DNSName:   "example.com",
				Targets:   endpoint.NewTargets("192.0.2.1"),
				RecordTTL: endpoint.TTL(defaultTTL),
			},
			expectedMatch: true,
		},
		{
			name:     "No comment and Default comment should match",
			provider: mikrotikProvider,
			endpointA: &endpoint.Endpoint{
				DNSName: "example.com",
				Targets: endpoint.NewTargets("192.0.2.1"),
				ProviderSpecific: endpoint.ProviderSpecific{
					{Name: "comment", Value: defaultComment},
				},
			},
			endpointB: &endpoint.Endpoint{
				DNSName: "example.com",
				Targets: endpoint.NewTargets("192.0.2.1"),
			},
			expectedMatch: true,
		},

		// MISMATCH CASES
		{
			name:     "Provider-specific properties do not match",
			provider: mikrotikProvider,
			endpointA: &endpoint.Endpoint{
				DNSName:   "example.com",
				Targets:   endpoint.NewTargets("192.0.2.1"),
				RecordTTL: endpoint.TTL(3600),
				ProviderSpecific: endpoint.ProviderSpecific{
					{Name: "comment", Value: "mismatch"},
				},
			},
			endpointB: &endpoint.Endpoint{
				DNSName:   "example.com",
				Targets:   endpoint.NewTargets("192.0.2.1"),
				RecordTTL: endpoint.TTL(3600),
				ProviderSpecific: endpoint.ProviderSpecific{
					{Name: "webhook/comment", Value: "different"},
				},
			},
			expectedMatch: false,
		},
		{
			name:     "Mismatch in DNSName",
			provider: mikrotikProvider,
			endpointA: &endpoint.Endpoint{
				DNSName:   "example1.com",
				Targets:   endpoint.NewTargets("192.0.2.1"),
				RecordTTL: endpoint.TTL(3600),
			},
			endpointB: &endpoint.Endpoint{
				DNSName:   "example2.com",
				Targets:   endpoint.NewTargets("192.0.2.1"),
				RecordTTL: endpoint.TTL(3600),
			},
			expectedMatch: false,
		},
		{
			name:     "Mismatch in Target",
			provider: mikrotikProvider,
			endpointA: &endpoint.Endpoint{
				DNSName:   "example.com",
				Targets:   endpoint.NewTargets("192.0.2.1"),
				RecordTTL: endpoint.TTL(3600),
			},
			endpointB: &endpoint.Endpoint{
				DNSName:   "example.com",
				Targets:   endpoint.NewTargets("192.0.2.2"),
				RecordTTL: endpoint.TTL(3600),
			},
			expectedMatch: false,
		},
		{
			name:     "Mismatch in TTL (X != Y)",
			provider: mikrotikProvider,
			endpointA: &endpoint.Endpoint{
				DNSName:   "example.com",
				Targets:   endpoint.NewTargets("192.0.2.1"),
				RecordTTL: endpoint.TTL(3600),
			},
			endpointB: &endpoint.Endpoint{
				DNSName:   "example.com",
				Targets:   endpoint.NewTargets("192.0.2.1"),
				RecordTTL: endpoint.TTL(3601),
			},
			expectedMatch: false,
		},
		{
			name:     "Mismatch in TTL (0 != X)",
			provider: mikrotikProvider,
			endpointA: &endpoint.Endpoint{
				DNSName:   "example.com",
				Targets:   endpoint.NewTargets("192.0.2.1"),
				RecordTTL: endpoint.TTL(0),
			},
			endpointB: &endpoint.Endpoint{
				DNSName:   "example.com",
				Targets:   endpoint.NewTargets("192.0.2.1"),
				RecordTTL: endpoint.TTL(5),
			},
			expectedMatch: true,
		},
		{
			name:     "Mismatch in TTL (Default != X)",
			provider: mikrotikProvider,
			endpointA: &endpoint.Endpoint{
				DNSName:   "example.com",
				Targets:   endpoint.NewTargets("192.0.2.1"),
				RecordTTL: endpoint.TTL(defaultTTL),
			},
			endpointB: &endpoint.Endpoint{
				DNSName:   "example.com",
				Targets:   endpoint.NewTargets("192.0.2.1"),
				RecordTTL: endpoint.TTL(5),
			},
			expectedMatch: true,
		},
		{
			name:     "Mismatch in comment (something != nothing)",
			provider: mikrotikProvider,
			endpointA: &endpoint.Endpoint{
				DNSName:   "example.com",
				Targets:   endpoint.NewTargets("192.0.2.1"),
				RecordTTL: endpoint.TTL(3600),
				ProviderSpecific: endpoint.ProviderSpecific{
					{Name: "comment", Value: "mismatch"},
				},
			},
			endpointB: &endpoint.Endpoint{
				DNSName:   "example.com",
				Targets:   endpoint.NewTargets("192.0.2.1"),
				RecordTTL: endpoint.TTL(3600),
			},
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
				{
					DNSName:   "example1.com",
					Targets:   endpoint.NewTargets("192.2.2.1"),
					RecordTTL: endpoint.TTL(36),
				},
				{
					DNSName:   "example.com",
					Targets:   endpoint.NewTargets("192.0.2.1"),
					RecordTTL: endpoint.TTL(3600),
					ProviderSpecific: endpoint.ProviderSpecific{
						{Name: "comment", Value: "test"},
					},
				},
				{
					DNSName:   "example2.com",
					Targets:   endpoint.NewTargets("192.1.2.1"),
					RecordTTL: endpoint.TTL(360),
				},
			},
			needle: &endpoint.Endpoint{
				DNSName:   "example.com",
				Targets:   endpoint.NewTargets("192.0.2.1"),
				RecordTTL: endpoint.TTL(3600),
				ProviderSpecific: endpoint.ProviderSpecific{
					{Name: "webhook/comment", Value: "test"},
				},
			},
			expectContain: true,
		},
		{
			name:     "Needle does not exist in haystack",
			provider: mikrotikProvider,
			haystack: []*endpoint.Endpoint{
				{
					DNSName:   "example1.com",
					Targets:   endpoint.NewTargets("192.0.2.1"),
					RecordTTL: endpoint.TTL(3600),
				},
				{
					DNSName:   "example2.com",
					Targets:   endpoint.NewTargets("192.0.2.1"),
					RecordTTL: endpoint.TTL(3600),
				},
				{
					DNSName:   "example3.com",
					Targets:   endpoint.NewTargets("192.0.2.1"),
					RecordTTL: endpoint.TTL(3600),
				},
			},
			needle: &endpoint.Endpoint{
				DNSName:   "example.org",
				Targets:   endpoint.NewTargets("192.0.2.1"),
				RecordTTL: endpoint.TTL(3600),
			},
			expectContain: false,
		},
		{
			name:     "Haystack is empty",
			provider: mikrotikProvider,
			haystack: []*endpoint.Endpoint{},
			needle: &endpoint.Endpoint{
				DNSName:   "example.com",
				Targets:   endpoint.NewTargets("192.0.2.1"),
				RecordTTL: endpoint.TTL(3600),
			},
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
	defaultTTL := 1800
	newDefaultTTL := 111111
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
					{
						DNSName:   "example.com",
						Targets:   endpoint.NewTargets("1.1.1.1"),
						RecordTTL: endpoint.TTL(3600),
						ProviderSpecific: endpoint.ProviderSpecific{
							{Name: "comment", Value: "test comment"},
							{Name: "address-list", Value: "main"},
							{Name: "match-subdomain", Value: ".*"},
						},
					},
					{
						DNSName:   "example.org",
						Targets:   endpoint.NewTargets("2.2.2.2"),
						RecordTTL: endpoint.TTL(300),
						ProviderSpecific: endpoint.ProviderSpecific{
							{Name: "comment", Value: "another comment"},
							{Name: "address-list", Value: "secondary"},
							{Name: "match-subdomain", Value: "*.example.com"},
						},
					},
				},
				UpdateNew: []*endpoint.Endpoint{
					{
						DNSName:   "example.com",
						Targets:   endpoint.NewTargets("1.1.1.1"),
						RecordTTL: endpoint.TTL(3600),
						ProviderSpecific: endpoint.ProviderSpecific{
							{Name: "webhook/comment", Value: "test comment"},
							{Name: "address-list", Value: "main"},
							{Name: "match-subdomain", Value: ".*"},
						},
					},
					{
						DNSName:   "example.org",
						Targets:   endpoint.NewTargets("2.2.2.2"),
						RecordTTL: endpoint.TTL(300),
						ProviderSpecific: endpoint.ProviderSpecific{
							{Name: "webhook/comment", Value: "another comment"},
							{Name: "address-list", Value: "secondary"},
							{Name: "match-subdomain", Value: "*.example.com"},
						},
					},
				},
			},
			expectedChanges: &plan.Changes{},
		},
		{
			name:     "Some matching, some different - only partial cleanup",
			provider: mikrotikProvider,
			inputChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					{
						DNSName:   "matching.com",
						Targets:   endpoint.NewTargets("1.1.1.1"),
						RecordTTL: endpoint.TTL(3600),
						ProviderSpecific: endpoint.ProviderSpecific{
							{Name: "comment", Value: "old comment"},
						},
					},
					{
						DNSName:   "different.org",
						Targets:   endpoint.NewTargets("2.2.2.2"),
						RecordTTL: endpoint.TTL(300),
						ProviderSpecific: endpoint.ProviderSpecific{
							{Name: "comment", Value: "old comment"},
						},
					},
				},
				UpdateNew: []*endpoint.Endpoint{
					{
						DNSName:   "matching.com",
						Targets:   endpoint.NewTargets("1.1.1.1"),
						RecordTTL: endpoint.TTL(3600),
						ProviderSpecific: endpoint.ProviderSpecific{
							{Name: "webhook/comment", Value: "old comment"},
						},
					},
					{
						DNSName:   "different.org",
						Targets:   endpoint.NewTargets("2.2.2.2"),
						RecordTTL: endpoint.TTL(300),
						ProviderSpecific: endpoint.ProviderSpecific{
							{Name: "webhook/comment", Value: "new comment"},
						},
					},
				},
			},
			expectedChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					{
						DNSName:   "different.org",
						Targets:   endpoint.NewTargets("2.2.2.2"),
						RecordTTL: endpoint.TTL(300),
						ProviderSpecific: endpoint.ProviderSpecific{
							{Name: "comment", Value: "old comment"},
						},
					},
				},
				UpdateNew: []*endpoint.Endpoint{
					{
						DNSName:   "different.org",
						Targets:   endpoint.NewTargets("2.2.2.2"),
						RecordTTL: endpoint.TTL(300),
						ProviderSpecific: endpoint.ProviderSpecific{
							{Name: "comment", Value: "new comment"},
						},
					},
				},
			},
		},
		{
			name:     "Different comments across multiple records - no cleanup",
			provider: mikrotikProvider,
			inputChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					{
						DNSName:   "example.com",
						Targets:   endpoint.NewTargets("1.1.1.1"),
						RecordTTL: endpoint.TTL(3600),
						ProviderSpecific: endpoint.ProviderSpecific{
							{Name: "comment", Value: "old comment"},
						},
					},
					{
						DNSName:   "example.net",
						Targets:   endpoint.NewTargets("3.3.3.3"),
						RecordTTL: endpoint.TTL(120),
					},
				},
				UpdateNew: []*endpoint.Endpoint{
					{
						DNSName:   "example.com",
						Targets:   endpoint.NewTargets("1.1.1.1"),
						RecordTTL: endpoint.TTL(3600),
						ProviderSpecific: endpoint.ProviderSpecific{
							{Name: "webhook/comment", Value: "new comment"},
						},
					},
					{
						DNSName:   "example.net",
						Targets:   endpoint.NewTargets("3.3.3.3"),
						RecordTTL: endpoint.TTL(120),
						ProviderSpecific: endpoint.ProviderSpecific{
							{Name: "comment", Value: "new comment"},
						},
					},
				},
			},
			expectedChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					{
						DNSName:   "example.com",
						Targets:   endpoint.NewTargets("1.1.1.1"),
						RecordTTL: endpoint.TTL(3600),
						ProviderSpecific: endpoint.ProviderSpecific{
							{Name: "comment", Value: "old comment"},
						},
					},
					{
						DNSName:   "example.net",
						Targets:   endpoint.NewTargets("3.3.3.3"),
						RecordTTL: endpoint.TTL(120),
					},
				},
				UpdateNew: []*endpoint.Endpoint{
					{
						DNSName:   "example.com",
						Targets:   endpoint.NewTargets("1.1.1.1"),
						RecordTTL: endpoint.TTL(3600),
						ProviderSpecific: endpoint.ProviderSpecific{
							{Name: "webhook/comment", Value: "new comment"},
						},
					},
					{
						DNSName:   "example.net",
						Targets:   endpoint.NewTargets("3.3.3.3"),
						RecordTTL: endpoint.TTL(120),
						ProviderSpecific: endpoint.ProviderSpecific{
							{Name: "comment", Value: "new comment"},
						},
					},
				},
			},
		},
		{
			name:     "Create record with zero value in TTL",
			provider: mikrotikProvider,
			inputChanges: &plan.Changes{
				Create: []*endpoint.Endpoint{
					{
						DNSName:   "example.org",
						Targets:   endpoint.NewTargets("2.2.2.2"),
						RecordTTL: endpoint.TTL(0),
					},
				},
			},
			expectedChanges: &plan.Changes{
				Create: []*endpoint.Endpoint{
					{
						DNSName:   "example.org",
						Targets:   endpoint.NewTargets("2.2.2.2"),
						RecordTTL: endpoint.TTL(mikrotikProvider.client.TTL),
					},
				},
			},
		},
		{
			name:     "Update record with zero value in TTL",
			provider: mikrotikProvider,
			inputChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					{
						DNSName:   "example.org",
						Targets:   endpoint.NewTargets("2.2.2.2"),
						RecordTTL: endpoint.TTL(defaultTTL),
					},
					{
						DNSName:   "example.org",
						Targets:   endpoint.NewTargets("2.2.2.2"),
						RecordTTL: endpoint.TTL(0),
					},
				},
				UpdateNew: []*endpoint.Endpoint{
					{
						DNSName:   "example.org",
						Targets:   endpoint.NewTargets("2.2.2.2"),
						RecordTTL: endpoint.TTL(0),
					},
					{
						DNSName:   "example.org",
						Targets:   endpoint.NewTargets("2.2.2.2"),
						RecordTTL: endpoint.TTL(mikrotikProvider.client.TTL),
					},
				},
			},
			expectedChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{},
				UpdateNew: []*endpoint.Endpoint{},
			},
		},
		{
			name:     "Update 0 -> default TTL => no changes",
			provider: mikrotikProvider,
			inputChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					{
						DNSName:   "example.org",
						Targets:   endpoint.NewTargets("2.2.2.2"),
						RecordTTL: endpoint.TTL(0),
					},
				},
				UpdateNew: []*endpoint.Endpoint{
					{
						DNSName:   "example.org",
						Targets:   endpoint.NewTargets("2.2.2.2"),
						RecordTTL: endpoint.TTL(defaultTTL),
					},
				},
			},
			expectedChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{},
				UpdateNew: []*endpoint.Endpoint{},
			},
		},
		{
			name:     "Update X -> default TTL => changes",
			provider: mikrotikProvider,
			inputChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					{
						DNSName:   "example.org",
						Targets:   endpoint.NewTargets("2.2.2.2"),
						RecordTTL: endpoint.TTL(5),
					},
				},
				UpdateNew: []*endpoint.Endpoint{
					{
						DNSName:   "example.org",
						Targets:   endpoint.NewTargets("2.2.2.2"),
						RecordTTL: endpoint.TTL(defaultTTL),
					},
				},
			},
			expectedChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					{
						DNSName:   "example.org",
						Targets:   endpoint.NewTargets("2.2.2.2"),
						RecordTTL: endpoint.TTL(5),
					},
				},
				UpdateNew: []*endpoint.Endpoint{
					{
						DNSName:   "example.org",
						Targets:   endpoint.NewTargets("2.2.2.2"),
						RecordTTL: endpoint.TTL(defaultTTL),
					},
				},
			},
		},
		{
			name: "Update default TTL -> X => changes",
			provider: &MikrotikProvider{
				client: &MikrotikApiClient{
					&MikrotikDefaults{
						TTL: int64(newDefaultTTL),
					},
					nil,
					nil,
				},
			},
			inputChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					{
						DNSName:   "example.org",
						Targets:   endpoint.NewTargets("2.2.2.2"),
						RecordTTL: endpoint.TTL(defaultTTL),
					},
				},
				UpdateNew: []*endpoint.Endpoint{
					{
						DNSName:   "example.org",
						Targets:   endpoint.NewTargets("2.2.2.2"),
						RecordTTL: endpoint.TTL(newDefaultTTL),
					},
				},
			},
			expectedChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					{
						DNSName:   "example.org",
						Targets:   endpoint.NewTargets("2.2.2.2"),
						RecordTTL: endpoint.TTL(defaultTTL),
					},
				},
				UpdateNew: []*endpoint.Endpoint{
					{
						DNSName:   "example.org",
						Targets:   endpoint.NewTargets("2.2.2.2"),
						RecordTTL: endpoint.TTL(newDefaultTTL),
					},
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
					t.Errorf("Expected endpoint: %v , got %v", tt.expectedChanges.UpdateOld[i], outputChanges.UpdateOld[i])
				}
			}

			for i := range tt.expectedChanges.UpdateNew {
				if !mikrotikProvider.compareEndpoints(outputChanges.UpdateNew[i], tt.expectedChanges.UpdateNew[i]) {
					t.Errorf("Expected endpoint: %v , got %v", tt.expectedChanges.UpdateNew[i], outputChanges.UpdateNew[i])
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
