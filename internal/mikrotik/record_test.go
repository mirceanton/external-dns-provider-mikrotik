package mikrotik

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecordJSONMarshaling(t *testing.T) {
	record := DNSRecord{
		ID:             "1",
		Address:        "192.168.1.1",
		CName:          "example.com",
		ForwardTo:      "forward.example.com",
		MXExchange:     "mail.example.com",
		Name:           "example",
		SrvPort:        "8080",
		SrvTarget:      "target.example.com",
		Text:           "some text",
		Type:           "A",
		AddressList:    "list",
		Comment:        "a comment",
		Disabled:       "false",
		MatchSubdomain: "sub.example.com",
		MXPreference:   "10",
		NS:             "ns.example.com",
		Regexp:         ".*",
		SrvPriority:    "1",
		SrvWeight:      "5",
	}

	data, err := json.Marshal(record)
	assert.NoError(t, err)

	var unmarshaledRecord DNSRecord
	err = json.Unmarshal(data, &unmarshaledRecord)
	assert.NoError(t, err)

	assert.Equal(t, record, unmarshaledRecord)
}

func TestRecordJSONUnmarshaling(t *testing.T) {
	data := []byte(`{
        ".id": "1",
        "address": "192.168.1.1",
        "cname": "example.com",
        "forward-to": "forward.example.com",
        "mx-exchange": "mail.example.com",
        "name": "example",
        "srv-port": "8080",
        "srv-target": "target.example.com",
        "text": "some text",
        "type": "A",
        "address-list": "list",
        "comment": "a comment",
        "disabled": "false",
        "match-subdomain": "sub.example.com",
        "mx-preference": "10",
        "ns": "ns.example.com",
        "regexp": ".*",
        "srv-priority": "1",
        "srv-weight": "5"
    }`)

	var record DNSRecord
	err := json.Unmarshal(data, &record)
	assert.NoError(t, err)

	expectedRecord := DNSRecord{
		ID:             "1",
		Address:        "192.168.1.1",
		CName:          "example.com",
		ForwardTo:      "forward.example.com",
		MXExchange:     "mail.example.com",
		Name:           "example",
		SrvPort:        "8080",
		SrvTarget:      "target.example.com",
		Text:           "some text",
		Type:           "A",
		AddressList:    "list",
		Comment:        "a comment",
		Disabled:       "false",
		MatchSubdomain: "sub.example.com",
		MXPreference:   "10",
		NS:             "ns.example.com",
		Regexp:         ".*",
		SrvPriority:    "1",
		SrvWeight:      "5",
	}

	assert.Equal(t, expectedRecord, record)
}
