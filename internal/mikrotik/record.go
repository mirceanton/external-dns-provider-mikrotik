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
	Disabled       string `json:"disabled,omitempty"`        // provider-specific

	// Record specific fields
	Address      string `json:"address,omitempty"`       // A, AAAA -> endpoint.Targets[0]
	CName        string `json:"cname,omitempty"`         // CNAME -> endpoint.Targets[0]
	Text         string `json:"text,omitempty"`          // TXT -> endpoint.Targets[0]
	MXExchange   string `json:"mx-exchange,omitempty"`   // MX -> provider-specific
	MXPreference string `json:"mx-preference,omitempty"` // MX -> provider-specific
	SrvPort      string `json:"srv-port,omitempty"`      // SRV -> provider-specific
	SrvTarget    string `json:"srv-target,omitempty"`    // SRV -> provider-specific
	SrvPriority  string `json:"srv-priority,omitempty"`  // SRV -> provider-specific
	SrvWeight    string `json:"srv-weight,omitempty"`    // SRV -> provider-specific
	NS           string `json:"ns,omitempty"`            // NS -> provider-specific

	// Additional fields for other record types that are not currently supported
	// ForwardTo    string `json:"forward-to,omitempty"`    // FWD
}

// NewDNSRecords converts an ExternalDNS Endpoint to multiple Mikrotik DNSRecords (one per target)
func NewDNSRecords(ep *endpoint.Endpoint) ([]*DNSRecord, error) {
	log.Debugf("Converting ExternalDNS endpoint to MikrotikDNS records: %+v", ep)

	// Sanity checks for common fields
	if ep.DNSName == "" {
		return nil, fmt.Errorf("DNS name is required")
	}
	if ep.RecordType == "" {
		return nil, fmt.Errorf("record type is required")
	}
	if len(ep.Targets) == 0 {
		return nil, fmt.Errorf("no targets provided for DNS record")
	}

	// Convert ExternalDNS TTL to Mikrotik TTL once
	ttl, err := EndpointTTLtoMikrotikTTL(ep.RecordTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to convert TTL: %w", err)
	}

	// Initialize a base record with common properties
	baseRecord := DNSRecord{
		Name: ep.DNSName,
		Type: ep.RecordType,
		TTL:  ttl,
	}

	// Process provider-specific properties once
	for _, providerSpecific := range ep.ProviderSpecific {
		switch providerSpecific.Name {
		case "comment", "webhook/comment":
			baseRecord.Comment = providerSpecific.Value
		case "disabled", "webhook/disabled":
			baseRecord.Disabled = providerSpecific.Value
		case "regexp", "webhook/regexp":
			baseRecord.Regexp = providerSpecific.Value
		case "match-subdomain", "webhook/match-subdomain":
			baseRecord.MatchSubdomain = providerSpecific.Value
		case "address-list", "webhook/address-list":
			baseRecord.AddressList = providerSpecific.Value
		default:
			log.Debugf("Encountered unknown provider-specific configuration '%s: %s' for DNS Record of type %s", providerSpecific.Name, providerSpecific.Value, baseRecord.Type)
		}
	}

	var records []*DNSRecord
	for i, target := range ep.Targets {
		if target == "" {
			log.Warnf("Skipping empty target at index %d for endpoint %s", i, ep.DNSName)
			continue
		}

		// Create a new record by copying the base record
		record := baseRecord

		// Set target-specific fields based on record type
		switch record.Type {
		case "A":
			if err := validateIPv4(target); err != nil {
				return nil, fmt.Errorf("invalid A record target %s: %w", target, err)
			}
			record.Address = target
		case "AAAA":
			if err := validateIPv6(target); err != nil {
				return nil, fmt.Errorf("invalid AAAA record target %s: %w", target, err)
			}
			record.Address = target
		case "CNAME":
			if err := validateDomain(target); err != nil {
				return nil, fmt.Errorf("invalid CNAME record target %s: %w", target, err)
			}
			record.CName = target
		case "TXT":
			if err := validateTXT(target); err != nil {
				return nil, fmt.Errorf("invalid TXT record target %s: %w", target, err)
			}
			record.Text = target
		case "MX":
			preference, exchange, err := parseMX(target)
			if err != nil {
				return nil, fmt.Errorf("invalid MX record target %s: %w", target, err)
			}
			record.MXPreference = preference
			record.MXExchange = exchange
		case "SRV":
			priority, weight, port, srvTarget, err := parseSRV(target)
			if err != nil {
				return nil, fmt.Errorf("invalid SRV record target %s: %w", target, err)
			}
			record.SrvPriority = priority
			record.SrvWeight = weight
			record.SrvPort = port
			record.SrvTarget = srvTarget
		case "NS":
			if err := validateDomain(target); err != nil {
				return nil, fmt.Errorf("invalid NS record target %s: %w", target, err)
			}
			record.NS = target
		default:
			return nil, fmt.Errorf("unsupported DNS type: %s", ep.RecordType)
		}

		records = append(records, &record)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no valid targets found for DNS record")
	}

	log.Debugf("Created %d DNS records from endpoint %s", len(records), ep.DNSName)
	return records, nil
}

// toExternalDNSTarget converts a Mikrotik DNSRecord to an ExternalDNS target string
func (r *DNSRecord) toExternalDNSTarget() (string, error) {
	log.Debugf("Converting MikrotikDNS record to ExternalDNS target: %+v", r)

	// Record-specific data to target string
	switch r.Type {
	case "A":
		if err := validateIPv4(r.Address); err != nil {
			return "", err
		}
		return r.Address, nil
	case "AAAA":
		if err := validateIPv6(r.Address); err != nil {
			return "", err
		}
		return r.Address, nil
	case "CNAME":
		if err := validateDomain(r.CName); err != nil {
			return "", err
		}
		return r.CName, nil
	case "TXT":
		if err := validateTXT(r.Text); err != nil {
			return "", err
		}
		return r.Text, nil
	case "MX":
		if err := validateDomain(r.MXExchange); err != nil {
			return "", err
		}
		if err := validateUnsignedInteger(r.MXPreference); err != nil {
			return "", err
		}
		return fmt.Sprintf("%s %s", r.MXPreference, r.MXExchange), nil
	case "SRV":
		if err := validateUnsignedInteger(r.SrvPort); err != nil {
			return "", err
		}
		if err := validateUnsignedInteger(r.SrvPriority); err != nil {
			return "", err
		}
		if err := validateUnsignedInteger(r.SrvWeight); err != nil {
			return "", err
		}
		if err := validateDomain(r.SrvTarget); err != nil {
			return "", err
		}
		return fmt.Sprintf("%s %s %s %s", r.SrvPriority, r.SrvWeight, r.SrvPort, r.SrvTarget), nil
	case "NS":
		if err := validateDomain(r.NS); err != nil {
			return "", err
		}
		return r.NS, nil
	default:
		return "", fmt.Errorf("unsupported DNS type: %s", r.Type)
	}
}

// ================================================================================================
// UTILS
// ================================================================================================
// MikrotikTTLtoEndpointTTL converts a Mikrotik TTL to an ExternalDNS TTL
func MikrotikTTLtoEndpointTTL(ttl string) (endpoint.TTL, error) {
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

// EndpointTTLtoMikrotikTTL converts an ExternalDNS TTL to a Mikrotik TTL.
// If no TTL is configured in the ExternalDNS endpoint, the default TTL is used.
func EndpointTTLtoMikrotikTTL(ttl endpoint.TTL) (string, error) {
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

// validateIPv4 checks if the provided address is a valid IPv4 address.
func validateIPv4(address string) error {
	if net.ParseIP(address) == nil {
		return fmt.Errorf("invalid IP address: %s", address)
	}

	if strings.Contains(address, ":") {
		return fmt.Errorf("provided address looks like an IPv6 address: %s", address)
	}

	return nil
}

// validateIPv6 checks if the provided address is a valid IPv6 address.
func validateIPv6(address string) error {
	if net.ParseIP(address) == nil {
		return fmt.Errorf("invalid IP address: %s", address)
	}

	if !strings.Contains(address, ":") {
		return fmt.Errorf("provided address looks like an IPv4 address: %s", address)
	}

	return nil
}

// validateTXT checks if the provided TXT record text is valid.
func validateTXT(text string) error {
	if text == "" {
		return fmt.Errorf("TXT record text cannot be empty")
	}
	//? TODO: add more validation here?
	return nil
}

// validateDomain checks if the provided domain is semantically valid.
func validateDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("a domain cannot be empty")
	}

	if len(domain) > 253 {
		return fmt.Errorf("invalid domain, length exceeds 253 characters")
	}

	domainRegex := `^(?i:[a-z0-9]([-a-z0-9]{0,61}[a-z0-9])?\.)+[a-z]{2,}$`
	matched, err := regexp.MatchString(domainRegex, domain)
	if err != nil || !matched {
		return fmt.Errorf("invalid domain: %s", domain)
	}

	return nil
}

// validateUnsignedInteger checks if the provided value is a number between 0 and 65535.
func validateUnsignedInteger(value string) error {
	if value == "" {
		return fmt.Errorf("value cannot be empty")
	}
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("value cannot be converted to int: %s", value)
	}

	if intVal < 0 || intVal > 65535 {
		return fmt.Errorf("value must be between 0 and 65535: %s", value)
	}
	return nil
}

// parseMX parses and validates an MX record
func parseMX(data string) (string, string, error) {
	data_split := strings.Split(data, " ")
	if len(data_split) != 2 {
		return "", "", fmt.Errorf("malformed MX record %s", data)
	}

	// Extract and Validate MX Preference
	preference := data_split[0]
	if err := validateUnsignedInteger(preference); err != nil {
		return "", "", fmt.Errorf("failed to validate MX preference: %v", err)
	}

	// Extract and Validate MX Exchange
	exchange := data_split[1]
	if err := validateDomain(exchange); err != nil {
		return "", "", fmt.Errorf("failed to validate MX exchange: %v", err)
	}

	return preference, exchange, nil
}

func parseSRV(data string) (string, string, string, string, error) {
	data_split := strings.Split(data, " ")
	if len(data_split) != 4 {
		return "", "", "", "", fmt.Errorf("malformed SRV record %s", data)
	}

	// Extract and Validate SRV Priority
	priority := data_split[0]
	if err := validateUnsignedInteger(priority); err != nil {
		return "", "", "", "", fmt.Errorf("failed to validate SRV priority: %v", err)
	}

	// Extract and Validate SRV weight
	weight := data_split[1]
	if err := validateUnsignedInteger(weight); err != nil {
		return "", "", "", "", fmt.Errorf("failed to validate SRV weight: %v", err)
	}

	// Extract and Validate SRV port
	port := data_split[2]
	if err := validateUnsignedInteger(port); err != nil {
		return "", "", "", "", fmt.Errorf("failed to validate SRV port: %v", err)
	}

	// Extract and Validate SRV target
	target := data_split[3]
	if err := validateDomain(target); err != nil {
		return "", "", "", "", fmt.Errorf("failed to validate SRV target: %v", err)
	}

	return priority, weight, port, target, nil
}
