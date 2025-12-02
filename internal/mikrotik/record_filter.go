package mikrotik

import "net/url"

// DNSRecordFilter represents the filtering criteria for DNS records in MikroTik RouterOS.
type DNSRecordFilter struct {
	Name string
	Type string
}

// toQueryParams converts a DNSRecordFilter to a query string for the RouterOS API.
func (f DNSRecordFilter) toQueryParams() string {
	recordType := f.Type
	if recordType == "" {
		recordType = "A,AAAA,CNAME,TXT,MX,SRV,NS"
	}
	query := "type=" + recordType

	if f.Name != "" {
		query += "&name=" + url.QueryEscape(f.Name)
	}

	return query
}
