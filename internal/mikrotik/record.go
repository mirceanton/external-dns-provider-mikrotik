package mikrotik

import (
	"fmt"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/external-dns/endpoint"
)

// DNSRecord represents a MikroTik DNS record in the format used by the API
// https://help.mikrotik.com/docs/display/ROS/DNS#DNS-DNSStatic
type DNSRecord struct {
	// Common fields for all record types
	ID             string `json:".id,omitempty"`             // only fetched from API
	Name           string `json:"name"`                      // endpoint.DNSName
	Type           string `json:"type"`                      // endpoint.RecordType
	TTL            string `json:"ttl,omitempty"`             // endpoint.RecordTTL
	Disabled       string `json:"disabled,omitempty"`        // provider-specific
	Comment        string `json:"comment,omitempty"`         // provider-specific
	Regexp         string `json:"regexp,omitempty"`          // provider-specific
	MatchSubdomain string `json:"match-subdomain,omitempty"` // provider-specific
	AddressList    string `json:"address-list,omitempty"`    // provider-specific

	// Record specific fields
	Address string `json:"address,omitempty"` // A, AAAA -> endpoint.Targets[0]
	CName   string `json:"cname,omitempty"`   // CNAME -> endpoint.Targets[0]
	Text    string `json:"text,omitempty"`    // TXT -> endpoint.Targets[0]

	// Additional fields for other record types that are not currently supported
	// MXExchange   string `json:"mx-exchange,omitempty"`   // MX -> provider-specific
	// MXPreference string `json:"mx-preference,omitempty"` // MX -> provider-specific
	// SrvPort      string `json:"srv-port,omitempty"`      // SRV -> provider-specific
	// SrvTarget    string `json:"srv-target,omitempty"`    // SRV -> provider-specific
	// SrvPriority  string `json:"srv-priority,omitempty"`  // SRV -> provider-specific
	// SrvWeight    string `json:"srv-weight,omitempty"`    // SRV -> provider-specific
	// NS           string `json:"ns,omitempty"`            // NS -> provider-specific
	// ForwardTo    string `json:"forward-to,omitempty"`    // FWD
}

// NewDNSRecord converts an ExternalDNS Endpoint to a Mikrotik DNSRecord
func NewDNSRecord(endpoint *endpoint.Endpoint) (*DNSRecord, error) {
	log.Debugf("converting ExternalDNS endpoint to MikrotikDNS: %v", endpoint)

	record := &DNSRecord{Name: endpoint.DNSName}
	log.Debugf("Name set to: %s", record.Name)

	recordType := endpoint.RecordType
	if recordType == "" {
		recordType = "A"
	}
	record.Type = recordType
	log.Debugf("Type set to: %s", record.Type)

	switch recordType {
	case "A", "AAAA":
		record.Address = endpoint.Targets[0]
		log.Debugf("Address set to: %s", record.Address)
	case "CNAME":
		record.CName = endpoint.Targets[0]
		log.Debugf("CName set to: %s", record.CName)
	case "TXT":
		record.Text = endpoint.Targets[0]
		log.Debugf("Text set to: %s", record.Text)

	default:
		return nil, fmt.Errorf("unsupported DNS type: %s", endpoint.RecordType)
	}

	ttl, err := endpointTTLtoMikrotikTTL(endpoint.RecordTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to convert ExternalDNS endpoint to MikrotikDNS: %v", err)
	}
	record.TTL = ttl
	log.Debugf("TTL set to: %s", record.TTL)

	for _, providerSpecific := range endpoint.ProviderSpecific {
		switch providerSpecific.Name {
		case "disabled":
			record.Disabled = providerSpecific.Value
			log.Debugf("Disabled set to: %s", record.Disabled)
		case "comment":
			record.Comment = providerSpecific.Value
			log.Debugf("Comment set to: %s", record.Comment)
		case "regexp":
			record.Regexp = providerSpecific.Value
			log.Debugf("Regexp set to: %s", record.Regexp)
		case "match-subdomain":
			record.MatchSubdomain = providerSpecific.Value
			log.Debugf("MatchSubdomain set to: %s", record.MatchSubdomain)
		case "address-list":
			record.AddressList = providerSpecific.Value
			log.Debugf("AddressList set to: %s", record.AddressList)

		default:
			return nil, fmt.Errorf(
				"unsupported provider specific configuration '%s' for DNS Record of type %s ",
				providerSpecific.Name,
				recordType,
			)
		}
	}

	if record.Comment == "" {
		log.Debugf("Comment not set. Using default value from environment variable MIKROTIK_DEFAULT_COMMENT")
		record.Comment = os.Getenv("MIKROTIK_DEFAULT_COMMENT")
		log.Debugf("Comment set to: %s", record.Comment)
	}

	if record.Disabled == "" {
		log.Debugf("Disabled not set on ExternalDNS side. Setting to 'false'")
		record.Disabled = "false"
		log.Debugf("Disabled set to: %s", record.Disabled)
	}

	log.Debugf("Converted ExternalDNS endpoint to MikrotikDNS: %s", record.toString())
	return record, nil
}

func (r *DNSRecord) toExternalDNSEndpoint() (*endpoint.Endpoint, error) {
	log.Debugf("converting MikrotikDNS record to ExternalDNS: %v", r.toString())

	if r.Type == "" {
		log.Debugf("Record type not set. Using default value 'A'")
		r.Type = "A"
	}

	ep := endpoint.Endpoint{
		DNSName:    r.Name,
		RecordType: r.Type,
	}
	log.Debugf("DNSName set to: %s", ep.DNSName)
	log.Debugf("RecordType set to: %s", ep.RecordType)

	switch ep.RecordType {
	case "A", "AAAA":
		ep.Targets = endpoint.NewTargets(r.Address)
	case "CNAME":
		ep.Targets = endpoint.NewTargets(r.CName)
	case "TXT":
		ep.Targets = endpoint.NewTargets(r.Text)

	default:
		return nil, fmt.Errorf("unsupported DNS type: %s", ep.RecordType)
	}
	log.Debugf("Targets set to: %v", ep.Targets)

	ttl, err := mikrotikTTLtoEndpointTTL(r.TTL)
	if err != nil {
		return nil, fmt.Errorf("failed to convert MikrotikDNS record to ExternalDNS: %v", err)
	}
	ep.RecordTTL = ttl
	log.Debugf("RecordTTL set to: %v", ep.RecordTTL)

	if r.Disabled != "" {
		ep.ProviderSpecific = append(ep.ProviderSpecific, endpoint.ProviderSpecificProperty{
			Name:  "disabled",
			Value: r.Disabled,
		})
		log.Debugf("Disabled set to: %s", r.Disabled)
	}
	if r.Comment != "" {
		ep.ProviderSpecific = append(ep.ProviderSpecific, endpoint.ProviderSpecificProperty{
			Name:  "comment",
			Value: r.Comment,
		})
		log.Debugf("Comment set to: %s", r.Comment)
	}
	if r.Regexp != "" {
		ep.ProviderSpecific = append(ep.ProviderSpecific, endpoint.ProviderSpecificProperty{
			Name:  "regexp",
			Value: r.Regexp,
		})
		log.Debugf("Regexp set to: %s", r.Regexp)
	}
	if r.MatchSubdomain != "" {
		ep.ProviderSpecific = append(ep.ProviderSpecific, endpoint.ProviderSpecificProperty{
			Name:  "match-subdomain",
			Value: r.MatchSubdomain,
		})
		log.Debugf("MatchSubdomain set to: %s", r.MatchSubdomain)
	}
	if r.AddressList != "" {
		ep.ProviderSpecific = append(ep.ProviderSpecific, endpoint.ProviderSpecificProperty{
			Name:  "address-list",
			Value: r.AddressList,
		})
		log.Debugf("AddressList set to: %s", r.AddressList)
	}

	log.Debugf("Converted MikrotikDNS record to ExternalDNS: %v", ep)
	return &ep, nil
}

// MikrotikTTLtoEndpointTTL converts a Mikrotik TTL to an ExternalDNS TTL
func mikrotikTTLtoEndpointTTL(ttl string) (endpoint.TTL, error) {
	log.Debugf("Converting Mikrotik TTL to Endpoint TTL: %s", ttl)

	// i think this should realistically never happen. if it does, it's perhaps a bug?
	// Mikrotik sets TTL by default on all records, so it should always be set
	if ttl == "" {
		log.Warnf("Found a Mikrotik Endpoint with no TTL?! Setting TTL to 0")
		ttl = "0s"
	}

	duration, err := time.ParseDuration(ttl)
	if err != nil {
		return endpoint.TTL(0), fmt.Errorf("failed to parse duration: %v", err)
	}
	log.Debugf("Parsed duration: %v", duration)

	log.Debugf("Converted TTL: %v", duration.Seconds())
	return endpoint.TTL(duration.Seconds()), nil
}

// EndpointTTLtoMikrotikTTL converts an ExternalDNS TTL to a Mikrotik TTL.
// If no TTL is configured in the ExternalDNS endpoint, the default TTL is used.
func endpointTTLtoMikrotikTTL(ttl endpoint.TTL) (string, error) {
	log.Debugf("Converting Endpoint TTL to Mikrotil TTL: %v", ttl)

	ttlString := os.Getenv("MIKROTIK_DEFAULT_TTL")
	if ttlString == "" {
		log.Debugf("No default TTL set in environment variable MIKROTIK_DEFAULT_TTL. Using '0s' as default.")
		ttlString = "0s"
	}

	if ttl.IsConfigured() {
		log.Debugf("ExternalDNS Endpoint has TTL defined: %v", ttl)
		ttlString = strconv.FormatInt(int64(ttl), 10) + "s"
		log.Debugf("Using TTL from endpoint: %s", ttlString)
	} else {
		log.Debugf("No TTL configured in the ExternalDNS endpoint. Using default TTL: %s", ttlString)
	}

	duration, err := time.ParseDuration(ttlString)
	if err != nil {
		return "", fmt.Errorf("failed to parse TTL: %v", err)
	}

	log.Debugf("Converted TTL: %v", duration.String())
	return duration.String(), nil
}
