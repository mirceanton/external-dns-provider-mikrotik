package mikrotik

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
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
	Comment        string `json:"comment,omitempty"`         // provider-specific
	Regexp         string `json:"regexp,omitempty"`          // provider-specific
	MatchSubdomain string `json:"match-subdomain,omitempty"` // provider-specific
	AddressList    string `json:"address-list,omitempty"`    // provider-specific

	// Disabled       string `json:"disabled,omitempty"`        // provider-specific

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
	log.Debugf("Converting ExternalDNS endpoint to MikrotikDNS: %v", endpoint)

	if endpoint.DNSName == "" {
		return nil, fmt.Errorf("DNS name is required")
	}

	if endpoint.RecordType == "" {
		return nil, fmt.Errorf("record type is required")
	}

	if len(endpoint.Targets) == 0 || endpoint.Targets[0] == "" {
		return nil, fmt.Errorf("no target provided for DNS record")
	}

	record := &DNSRecord{Name: endpoint.DNSName}
	log.Debugf("Name set to: %s", record.Name)

	record.Type = endpoint.RecordType
	log.Debugf("Type set to: %s", record.Type)

	switch record.Type {
	case "A":
		record.Address = endpoint.Targets[0]
		if net.ParseIP(record.Address) == nil || strings.Contains(record.Address, ":") {
			return nil, fmt.Errorf("invalid IPv4 address: %s", record.Address)
		}
		log.Debugf("Address set to: %s", record.Address)
	case "AAAA":
		record.Address = endpoint.Targets[0]
		if net.ParseIP(record.Address) == nil || !strings.Contains(record.Address, ":") {
			return nil, fmt.Errorf("invalid IPv6 address: %s", record.Address)
		}
		log.Debugf("Address set to: %s", record.Address)
	case "CNAME":
		record.CName = endpoint.Targets[0]
		if record.CName == "" {
			return nil, fmt.Errorf("CNAME target cannot be empty")
		}
		log.Debugf("CName set to: %s", record.CName)
	case "TXT":
		record.Text = endpoint.Targets[0]
		if record.Text == "" {
			return nil, fmt.Errorf("TXT record text cannot be empty")
		}
		log.Debugf("Text set to: %s", record.Text)
	default:
		return nil, fmt.Errorf("unsupported DNS type: %s", endpoint.RecordType)
	}

	ttl, err := endpointTTLtoMikrotikTTL(endpoint.RecordTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to convert TTL: %v", err)
	}
	record.TTL = ttl
	log.Debugf("TTL set to: %s", record.TTL)

	for _, providerSpecific := range endpoint.ProviderSpecific {
		switch providerSpecific.Name {
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
				"unsupported provider specific configuration '%s' for DNS Record of type %s",
				providerSpecific.Name,
				record.Type,
			)
		}
	}

	log.Debugf("Converted ExternalDNS endpoint to MikrotikDNS: %v", record)
	return record, nil
}

// toExternalDNSEndpoint converts a Mikrotik DNSRecord to an ExternalDNS Endpoint
func (r *DNSRecord) toExternalDNSEndpoint() (*endpoint.Endpoint, error) {
	log.Debugf("Converting MikrotikDNS record to ExternalDNS: %v", r)

	if r.Type == "" {
		log.Debugf("Record type not set. Using default value 'A'")
		r.Type = "A"
	}

	ep := endpoint.Endpoint{
		DNSName:    r.Name,
		RecordType: r.Type,
	}

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
	if len(ep.Targets) == 0 || ep.Targets[0] == "" {
		return nil, fmt.Errorf("no target provided for DNS record")
	}

	ttl, err := mikrotikTTLtoEndpointTTL(r.TTL)
	if err != nil {
		return nil, fmt.Errorf("failed to convert MikrotikDNS record to ExternalDNS: %v", err)
	}
	ep.RecordTTL = ttl

	if r.Comment != "" {
		ep.ProviderSpecific = append(ep.ProviderSpecific, endpoint.ProviderSpecificProperty{
			Name:  "comment",
			Value: r.Comment,
		})
	}
	if r.Regexp != "" {
		ep.ProviderSpecific = append(ep.ProviderSpecific, endpoint.ProviderSpecificProperty{
			Name:  "regexp",
			Value: r.Regexp,
		})
	}
	if r.MatchSubdomain != "" {
		ep.ProviderSpecific = append(ep.ProviderSpecific, endpoint.ProviderSpecificProperty{
			Name:  "match-subdomain",
			Value: r.MatchSubdomain,
		})
	}
	if r.AddressList != "" {
		ep.ProviderSpecific = append(ep.ProviderSpecific, endpoint.ProviderSpecificProperty{
			Name:  "address-list",
			Value: r.AddressList,
		})
	}

	log.Debugf("Converted MikrotikDNS record to ExternalDNS: %v", ep)
	return &ep, nil
}

// mikrotikTTLtoEndpointTTL converts a Mikrotik TTL to an ExternalDNS TTL
func mikrotikTTLtoEndpointTTL(ttl string) (endpoint.TTL, error) {
	log.Debugf("Converting Mikrotik TTL to Endpoint TTL: %s", ttl)

	if ttl == "" {
		log.Warnf("Found a Mikrotik Endpoint with no TTL?! Setting TTL to 0")
		ttl = "0s"
	}

	// Define the unit multipliers in seconds
	unitMap := map[string]float64{
		"d": 86400, // seconds in a day
		"h": 3600,  // seconds in an hour
		"m": 60,    // seconds in a minute
		"s": 1,     // seconds in a second
	}

	// Regular expression to match number-unit pairs, including negative numbers
	re := regexp.MustCompile(`(-?\d*\.?\d+)([dhms])`)

	matches := re.FindAllStringSubmatch(ttl, -1)
	if matches == nil {
		return 0, fmt.Errorf("invalid duration string: '%s'", ttl)
	}

	// Reconstruct the matched parts to validate the entire input
	var matchedString string
	for _, match := range matches {
		matchedString += match[0]
	}

	// Remove any whitespace for accurate comparison
	trimmedInput := strings.ReplaceAll(ttl, " ", "")
	if matchedString != trimmedInput {
		return 0, fmt.Errorf("invalid characters in duration string: '%s'", ttl)
	}

	var totalSeconds float64

	for _, match := range matches {
		valueStr := match[1]
		unitStr := match[2]

		multiplier, ok := unitMap[unitStr]
		if !ok {
			return 0, fmt.Errorf("invalid unit '%s' in duration", unitStr)
		}

		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number '%s' in duration", valueStr)
		}

		if value < 0 {
			return 0, fmt.Errorf("negative values are not allowed: '%s'", valueStr+unitStr)
		}

		totalSeconds += value * multiplier
	}

	duration := time.Duration(totalSeconds * float64(time.Second))
	log.Debugf("Parsed duration: %v", duration)

	log.Debugf("Converted TTL: %v", duration.Seconds())
	return endpoint.TTL(duration.Seconds()), nil
}

// endpointTTLtoMikrotikTTL converts an ExternalDNS TTL to a Mikrotik TTL.
// If no TTL is configured in the ExternalDNS endpoint, the default TTL is used.
func endpointTTLtoMikrotikTTL(ttl endpoint.TTL) (string, error) {
	log.Debugf("Converting Endpoint TTL to Mikrotik TTL: %v", ttl)

	if ttl < 0 {
		return "", fmt.Errorf("negative TTL values are not allowed: %v", ttl)
	}

	totalSeconds := int64(ttl)
	days := totalSeconds / 86400
	remainder := totalSeconds % 86400

	hours := remainder / 3600
	remainder %= 3600

	minutes := remainder / 60
	seconds := remainder % 60

	var parts []string

	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if seconds > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%ds", seconds))
	}

	durationStr := strings.Join(parts, "")
	log.Debugf("Converted TTL: %v", durationStr)
	return durationStr, nil
}
