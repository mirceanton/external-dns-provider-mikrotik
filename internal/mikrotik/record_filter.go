package mikrotik

import "net/url"

// DNSRecordFilter represents the filtering criteria for DNS records in MikroTik RouterOS.
type DNSRecordFilter struct {
	Name string
	Type string
}

// toQueryParams converts a DNSRecordFilter to an encoded query string for the RouterOS API.
func (f DNSRecordFilter) toQueryParams() string {
	params := url.Values{}

	if f.Name != "" {
		params.Set("name", f.Name)
	}

	recordType := f.Type
	if recordType == "" {
		recordType = "A,AAAA,CNAME,TXT,MX,SRV,NS"
	}
	params.Set("type", recordType)

	return params.Encode()
}
