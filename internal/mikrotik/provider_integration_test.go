//go:build integration

package mikrotik

import (
	"context"
	"os"
	"testing"

	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
)

func newIntegrationProvider(t *testing.T) *MikrotikProvider {
	t.Helper()
	baseURL := os.Getenv("MIKROTIK_BASEURL")
	if baseURL == "" {
		t.Skip("MIKROTIK_BASEURL not set, skipping integration test")
	}
	config := &MikrotikConnectionConfig{
		BaseUrl:       baseURL,
		Username:      os.Getenv("MIKROTIK_USERNAME"),
		Password:      os.Getenv("MIKROTIK_PASSWORD"),
		SkipTLSVerify: os.Getenv("MIKROTIK_SKIP_TLS_VERIFY") == "true",
	}
	defaults := &MikrotikDefaults{DefaultTTL: 3600, DefaultComment: ""}
	client, err := NewMikrotikClient(config, defaults)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return &MikrotikProvider{
		client:       client,
		domainFilter: endpoint.NewDomainFilter([]string{"integration.test"}),
	}
}

func integrationApplyChanges(t *testing.T, p *MikrotikProvider, changes *plan.Changes) {
	t.Helper()
	if err := p.ApplyChanges(context.Background(), changes); err != nil {
		t.Fatalf("ApplyChanges failed: %v", err)
	}
}

func integrationAssertRecordExists(t *testing.T, p *MikrotikProvider, name, recordType, target string) {
	t.Helper()
	endpoints, err := p.Records(context.Background())
	if err != nil {
		t.Fatalf("Records() failed: %v", err)
	}
	for _, ep := range endpoints {
		if ep.DNSName == name && ep.RecordType == recordType {
			for _, tgt := range ep.Targets {
				if tgt == target {
					return
				}
			}
		}
	}
	t.Errorf("expected record %s %s -> %s not found in Records()", recordType, name, target)
}

func integrationAssertTargetAbsent(t *testing.T, p *MikrotikProvider, name, recordType, target string) {
	t.Helper()
	endpoints, err := p.Records(context.Background())
	if err != nil {
		t.Fatalf("Records() failed: %v", err)
	}
	for _, ep := range endpoints {
		if ep.DNSName == name && ep.RecordType == recordType {
			for _, tgt := range ep.Targets {
				if tgt == target {
					t.Errorf("target %s for record %s %s should be absent but was found in Records()", target, recordType, name)
					return
				}
			}
		}
	}
}

func integrationAssertRecordAbsent(t *testing.T, p *MikrotikProvider, name, recordType string) {
	t.Helper()
	endpoints, err := p.Records(context.Background())
	if err != nil {
		t.Fatalf("Records() failed: %v", err)
	}
	for _, ep := range endpoints {
		if ep.DNSName == name && ep.RecordType == recordType {
			t.Errorf("record %s %s should be absent but was found in Records()", recordType, name)
		}
	}
}

func TestIntegration_Records_Empty(t *testing.T) {
	p := newIntegrationProvider(t)
	endpoints, err := p.Records(context.Background())
	if err != nil {
		t.Fatalf("Records() error: %v", err)
	}
	if len(endpoints) != 0 {
		t.Errorf("expected 0 records on fresh RouterOS in integration.test zone, got %d: %v", len(endpoints), endpoints)
	}
}

func TestIntegration_Create_A(t *testing.T) {
	p := newIntegrationProvider(t)
	ep := NewEndpoint("new.integration.test", []string{"1.1.1.1"}, "A", 3600, nil)
	t.Cleanup(func() {
		integrationApplyChanges(t, p, &plan.Changes{Delete: []*endpoint.Endpoint{ep}})
	})
	integrationApplyChanges(t, p, &plan.Changes{Create: []*endpoint.Endpoint{ep}})
	integrationAssertRecordExists(t, p, "new.integration.test", "A", "1.1.1.1")
}

func TestIntegration_Create_AAAA(t *testing.T) {
	p := newIntegrationProvider(t)
	ep := NewEndpoint("aaaa.integration.test", []string{"2001:db8::1"}, "AAAA", 3600, nil)
	t.Cleanup(func() {
		integrationApplyChanges(t, p, &plan.Changes{Delete: []*endpoint.Endpoint{ep}})
	})
	integrationApplyChanges(t, p, &plan.Changes{Create: []*endpoint.Endpoint{ep}})
	integrationAssertRecordExists(t, p, "aaaa.integration.test", "AAAA", "2001:db8::1")
}

func TestIntegration_Create_CNAME(t *testing.T) {
	p := newIntegrationProvider(t)
	ep := NewEndpoint("cname.integration.test", []string{"target.example.com"}, "CNAME", 3600, nil)
	t.Cleanup(func() {
		integrationApplyChanges(t, p, &plan.Changes{Delete: []*endpoint.Endpoint{ep}})
	})
	integrationApplyChanges(t, p, &plan.Changes{Create: []*endpoint.Endpoint{ep}})
	integrationAssertRecordExists(t, p, "cname.integration.test", "CNAME", "target.example.com")
}

func TestIntegration_Create_TXT(t *testing.T) {
	p := newIntegrationProvider(t)
	ep := NewEndpoint("txt.integration.test", []string{"v=spf1 include:example.com ~all"}, "TXT", 3600, nil)
	t.Cleanup(func() {
		integrationApplyChanges(t, p, &plan.Changes{Delete: []*endpoint.Endpoint{ep}})
	})
	integrationApplyChanges(t, p, &plan.Changes{Create: []*endpoint.Endpoint{ep}})
	integrationAssertRecordExists(t, p, "txt.integration.test", "TXT", "v=spf1 include:example.com ~all")
}

func TestIntegration_Create_MX(t *testing.T) {
	p := newIntegrationProvider(t)
	ep := NewEndpoint("mx.integration.test", []string{"10 mail.example.com"}, "MX", 3600, nil)
	t.Cleanup(func() {
		integrationApplyChanges(t, p, &plan.Changes{Delete: []*endpoint.Endpoint{ep}})
	})
	integrationApplyChanges(t, p, &plan.Changes{Create: []*endpoint.Endpoint{ep}})
	integrationAssertRecordExists(t, p, "mx.integration.test", "MX", "10 mail.example.com")
}

func TestIntegration_Create_SRV(t *testing.T) {
	p := newIntegrationProvider(t)
	ep := NewEndpoint("sip.tcp.integration.test", []string{"10 20 5060 sip.example.com"}, "SRV", 3600, nil)
	t.Cleanup(func() {
		integrationApplyChanges(t, p, &plan.Changes{Delete: []*endpoint.Endpoint{ep}})
	})
	integrationApplyChanges(t, p, &plan.Changes{Create: []*endpoint.Endpoint{ep}})
	integrationAssertRecordExists(t, p, "sip.tcp.integration.test", "SRV", "10 20 5060 sip.example.com")
}

func TestIntegration_Create_NS(t *testing.T) {
	p := newIntegrationProvider(t)
	ep := NewEndpoint("ns.integration.test", []string{"ns1.example.com"}, "NS", 3600, nil)
	t.Cleanup(func() {
		integrationApplyChanges(t, p, &plan.Changes{Delete: []*endpoint.Endpoint{ep}})
	})
	integrationApplyChanges(t, p, &plan.Changes{Create: []*endpoint.Endpoint{ep}})
	integrationAssertRecordExists(t, p, "ns.integration.test", "NS", "ns1.example.com")
}

func TestIntegration_Delete(t *testing.T) {
	p := newIntegrationProvider(t)
	ep := NewEndpoint("delete.integration.test", []string{"1.1.1.1"}, "A", 3600, nil)

	integrationApplyChanges(t, p, &plan.Changes{Create: []*endpoint.Endpoint{ep}})
	integrationAssertRecordExists(t, p, "delete.integration.test", "A", "1.1.1.1")

	integrationApplyChanges(t, p, &plan.Changes{Delete: []*endpoint.Endpoint{ep}})
	integrationAssertRecordAbsent(t, p, "delete.integration.test", "A")
}

func TestIntegration_Update(t *testing.T) {
	p := newIntegrationProvider(t)
	oldEp := NewEndpoint("update.integration.test", []string{"1.1.1.1"}, "A", 3600, nil)
	newEp := NewEndpoint("update.integration.test", []string{"2.2.2.2"}, "A", 3600, nil)
	t.Cleanup(func() {
		integrationApplyChanges(t, p, &plan.Changes{Delete: []*endpoint.Endpoint{newEp}})
	})

	integrationApplyChanges(t, p, &plan.Changes{Create: []*endpoint.Endpoint{oldEp}})
	integrationAssertRecordExists(t, p, "update.integration.test", "A", "1.1.1.1")

	integrationApplyChanges(t, p, &plan.Changes{
		UpdateOld: []*endpoint.Endpoint{oldEp},
		UpdateNew: []*endpoint.Endpoint{newEp},
	})
	integrationAssertRecordExists(t, p, "update.integration.test", "A", "2.2.2.2")
	integrationAssertTargetAbsent(t, p, "update.integration.test", "A", "1.1.1.1")
}

func TestIntegration_Update_MultipleTargets(t *testing.T) {
	p := newIntegrationProvider(t)
	oldEp := NewEndpoint("multi.integration.test", []string{"1.1.1.1", "2.2.2.2"}, "A", 3600, nil)
	newEp := NewEndpoint("multi.integration.test", []string{"2.2.2.2", "3.3.3.3", "4.4.4.4"}, "A", 3600, nil)
	t.Cleanup(func() {
		integrationApplyChanges(t, p, &plan.Changes{Delete: []*endpoint.Endpoint{newEp}})
	})

	integrationApplyChanges(t, p, &plan.Changes{Create: []*endpoint.Endpoint{oldEp}})
	integrationAssertRecordExists(t, p, "multi.integration.test", "A", "1.1.1.1")
	integrationAssertRecordExists(t, p, "multi.integration.test", "A", "2.2.2.2")

	integrationApplyChanges(t, p, &plan.Changes{
		UpdateOld: []*endpoint.Endpoint{oldEp},
		UpdateNew: []*endpoint.Endpoint{newEp},
	})
	integrationAssertRecordExists(t, p, "multi.integration.test", "A", "2.2.2.2")
	integrationAssertRecordExists(t, p, "multi.integration.test", "A", "3.3.3.3")
	integrationAssertRecordExists(t, p, "multi.integration.test", "A", "4.4.4.4")
	integrationAssertTargetAbsent(t, p, "multi.integration.test", "A", "1.1.1.1")
}
