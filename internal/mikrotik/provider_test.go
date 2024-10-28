package mikrotik

import (
	"testing"

	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
)

func TestCleanupChanges(t *testing.T) {
	tests := []struct {
		name            string
		inputChanges    *plan.Changes
		expectedChanges *plan.Changes
	}{
		{
			name: "Matching ProviderSpecific properties - should clean up",
			inputChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					{
						DNSName:   "example.com",
						Targets:   endpoint.NewTargets("1.1.1.1"),
						RecordTTL: endpoint.TTL(3600),
						ProviderSpecific: endpoint.ProviderSpecific{
							{Name: "comment", Value: "test comment"},
							{Name: "match-subdomain", Value: "true"},
							{Name: "address-list", Value: "main"},
							{Name: "regexp", Value: ".*"},
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
							{Name: "webhook/match-subdomain", Value: "true"},
							{Name: "webhook/address-list", Value: "main"},
							{Name: "webhook/regexp", Value: ".*"},
						},
					},
				},
			},
			expectedChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{}, // should be empty after cleanup
				UpdateNew: []*endpoint.Endpoint{}, // should be empty after cleanup
			},
		},
		{
			name: "Different ProviderSpecific - should not clean up",
			inputChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					{
						DNSName:   "example.com",
						Targets:   endpoint.NewTargets("1.1.1.1"),
						RecordTTL: endpoint.TTL(3600),
						ProviderSpecific: endpoint.ProviderSpecific{
							{Name: "comment", Value: "old comment"},
							{Name: "match-subdomain", Value: "true"},
							{Name: "address-list", Value: "main"},
							{Name: "regexp", Value: ".*"},
						},
					},
				},
				UpdateNew: []*endpoint.Endpoint{
					{
						DNSName:   "example.com",
						Targets:   endpoint.NewTargets("1.1.1.1"),
						RecordTTL: endpoint.TTL(3600),
						ProviderSpecific: endpoint.ProviderSpecific{
							{Name: "webhook/comment", Value: "new comment"},
							{Name: "webhook/match-subdomain", Value: "true"},
							{Name: "webhook/address-list", Value: "main"},
							{Name: "webhook/regexp", Value: ".*"},
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
							{Name: "match-subdomain", Value: "true"},
							{Name: "address-list", Value: "main"},
							{Name: "regexp", Value: ".*"},
						},
					},
				},
				UpdateNew: []*endpoint.Endpoint{
					{
						DNSName:   "example.com",
						Targets:   endpoint.NewTargets("1.1.1.1"),
						RecordTTL: endpoint.TTL(3600),
						ProviderSpecific: endpoint.ProviderSpecific{
							{Name: "webhook/comment", Value: "new comment"},
							{Name: "webhook/match-subdomain", Value: "true"},
							{Name: "webhook/address-list", Value: "main"},
							{Name: "webhook/regexp", Value: ".*"},
						},
					},
				},
			},
		},
		{
			name: "Different DNSName, same Targets and TTL - should not clean up",
			inputChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					{
						DNSName:   "different.com",
						Targets:   endpoint.NewTargets("1.1.1.1"),
						RecordTTL: endpoint.TTL(3600),
					},
				},
				UpdateNew: []*endpoint.Endpoint{
					{
						DNSName:   "example.com",
						Targets:   endpoint.NewTargets("1.1.1.1"),
						RecordTTL: endpoint.TTL(3600),
					},
				},
			},
			expectedChanges: &plan.Changes{
				UpdateOld: []*endpoint.Endpoint{
					{
						DNSName:   "different.com",
						Targets:   endpoint.NewTargets("1.1.1.1"),
						RecordTTL: endpoint.TTL(3600),
					},
				},
				UpdateNew: []*endpoint.Endpoint{
					{
						DNSName:   "example.com",
						Targets:   endpoint.NewTargets("1.1.1.1"),
						RecordTTL: endpoint.TTL(3600),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute the cleanup function
			outputChanges := cleanupChanges(tt.inputChanges)

			// Compare UpdateOld
			if len(outputChanges.UpdateOld) != len(tt.expectedChanges.UpdateOld) {
				t.Errorf("Expected UpdateOld length %d, got %d", len(tt.expectedChanges.UpdateOld), len(outputChanges.UpdateOld))
			} else {
				for i := range tt.expectedChanges.UpdateOld {
					if outputChanges.UpdateOld[i].DNSName != tt.expectedChanges.UpdateOld[i].DNSName {
						t.Errorf("Expected UpdateOld[%d].DNSName %s, got %s", i, tt.expectedChanges.UpdateOld[i].DNSName, outputChanges.UpdateOld[i].DNSName)
					}
				}
			}

			// Compare UpdateNew
			if len(outputChanges.UpdateNew) != len(tt.expectedChanges.UpdateNew) {
				t.Errorf("Expected UpdateNew length %d, got %d", len(tt.expectedChanges.UpdateNew), len(outputChanges.UpdateNew))
			} else {
				for i := range tt.expectedChanges.UpdateNew {
					if outputChanges.UpdateNew[i].DNSName != tt.expectedChanges.UpdateNew[i].DNSName {
						t.Errorf("Expected UpdateNew[%d].DNSName %s, got %s", i, tt.expectedChanges.UpdateNew[i].DNSName, outputChanges.UpdateNew[i].DNSName)
					}
				}
			}
		})
	}
}
