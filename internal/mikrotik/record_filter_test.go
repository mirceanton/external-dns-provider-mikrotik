package mikrotik

import (
	"net/url"
	"testing"
)

func TestDNSRecordFilter_toQueryParams(t *testing.T) {
	tests := []struct {
		name         string
		filter       DNSRecordFilter
		wantName     string
		wantTypeVals string
	}{
		{
			name:         "default type when empty",
			filter:       DNSRecordFilter{Name: "", Type: ""},
			wantName:     "",
			wantTypeVals: "A,AAAA,CNAME,TXT,MX,SRV,NS",
		},
		{
			name:         "custom single type",
			filter:       DNSRecordFilter{Name: "", Type: "A"},
			wantName:     "",
			wantTypeVals: "A",
		},
		{
			name:         "custom multi type",
			filter:       DNSRecordFilter{Name: "example.com", Type: "A,AAAA"},
			wantName:     "example.com",
			wantTypeVals: "A,AAAA",
		},
		{
			name:         "only name set",
			filter:       DNSRecordFilter{Name: "host.example.com", Type: ""},
			wantName:     "host.example.com",
			wantTypeVals: "A,AAAA,CNAME,TXT,MX,SRV,NS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.toQueryParams()

			vals, err := url.ParseQuery(got)
			if err != nil {
				t.Fatalf("failed to parse query params: %v", err)
			}

			if want := tt.wantName; want != "" {
				if vals.Get("name") != want {
					t.Fatalf("name param = %q, want %q", vals.Get("name"), want)
				}
			} else {
				if vals.Get("name") != "" {
					t.Fatalf("expected no name param, got %q", vals.Get("name"))
				}
			}

			if vals.Get("type") != tt.wantTypeVals {
				t.Fatalf("type param = %q, want %q", vals.Get("type"), tt.wantTypeVals)
			}
		})
	}
}
