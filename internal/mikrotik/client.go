// Rest API Docs: https://help.mikrotik.com/docs/display/ROS/REST+API

package mikrotik

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"slices"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/publicsuffix"
	"sigs.k8s.io/external-dns/endpoint"
)

type MikrotikDefaults struct {
	DefaultTTL     int64  `env:"MIKROTIK_DEFAULT_TTL" envDefault:"3600"`
	DefaultComment string `env:"MIKROTIK_DEFAULT_COMMENT" envDefault:""`
}

// MikrotikConnectionConfig holds the connection details for the API client
type MikrotikConnectionConfig struct {
	BaseUrl       string `env:"MIKROTIK_BASEURL,notEmpty"`
	Username      string `env:"MIKROTIK_USERNAME,notEmpty"`
	Password      string `env:"MIKROTIK_PASSWORD,notEmpty"`
	SkipTLSVerify bool   `env:"MIKROTIK_SKIP_TLS_VERIFY" envDefault:"false"`
}

// MikrotikApiClient encapsulates the client configuration and HTTP client
type MikrotikApiClient struct {
	*MikrotikDefaults
	*MikrotikConnectionConfig
	*http.Client
}

// MikrotikSystemInfo represents MikroTik system information
// https://help.mikrotik.com/docs/display/ROS/Resource
type MikrotikSystemInfo struct {
	ArchitectureName     string `json:"architecture-name"`
	BadBlocks            string `json:"bad-blocks"`
	BoardName            string `json:"board-name"`
	BuildTime            string `json:"build-time"`
	CPU                  string `json:"cpu"`
	CPUCount             string `json:"cpu-count"`
	CPUFrequency         string `json:"cpu-frequency"`
	CPULoad              string `json:"cpu-load"`
	FactorySoftware      string `json:"factory-software"`
	FreeHDDSpace         string `json:"free-hdd-space"`
	FreeMemory           string `json:"free-memory"`
	Platform             string `json:"platform"`
	TotalHDDSpace        string `json:"total-hdd-space"`
	TotalMemory          string `json:"total-memory"`
	Uptime               string `json:"uptime"`
	Version              string `json:"version"`
	WriteSectSinceReboot string `json:"write-sect-since-reboot"`
	WriteSectTotal       string `json:"write-sect-total"`
}

// NewMikrotikClient creates a new instance of MikrotikApiClient
func NewMikrotikClient(config *MikrotikConnectionConfig, defaults *MikrotikDefaults) (*MikrotikApiClient, error) {
	log.Infof("creating a new Mikrotik API Client")

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Errorf("failed to create cookie jar: %v", err)
		return nil, err
	}

	client := &MikrotikApiClient{
		MikrotikDefaults:         defaults,
		MikrotikConnectionConfig: config,
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: config.SkipTLSVerify,
				},
			},
			Jar: jar,
		},
	}

	return client, nil
}

// GetSystemInfo fetches system information from the MikroTik API
func (c *MikrotikApiClient) GetSystemInfo() (*MikrotikSystemInfo, error) {
	log.Debugf("fetching system information.")

	// Send the request
	resp, err := c.doRequest(http.MethodGet, "system/resource", nil, nil)
	if err != nil {
		log.Errorf("error fetching system info: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Parse the response
	var info MikrotikSystemInfo
	if err = json.NewDecoder(resp.Body).Decode(&info); err != nil {
		log.Errorf("error decoding response body: %v", err)
		return nil, err
	}
	log.Debugf("got system info: %+v", info)

	return &info, nil
}

// GetDNSRecordsByName fetches DNS records filtered by name and type from the MikroTik API
func (c *MikrotikApiClient) GetDNSRecords(filter DNSRecordFilter) ([]DNSRecord, error) {
	log.Debugf("fetching DNS records")

	// Send the request
	resp, err := c.doRequest(http.MethodGet, "ip/dns/static", filter.toQueryParams(), nil)
	if err != nil {
		log.Errorf("error fetching DNS records: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Parse the response
	var records []DNSRecord
	if err = json.NewDecoder(resp.Body).Decode(&records); err != nil {
		log.Errorf("error decoding response body: %v", err)
		return nil, err
	}

	log.Debugf("fetched %d DNS records using server-side filtering", len(records))
	return records, nil
}

// DeleteDNSRecords deletes all DNS records associated with an endpoint
func (c *MikrotikApiClient) DeleteDNSRecords(ep *endpoint.Endpoint) error {
	log.Infof("deleting DNS records for endpoint: %+v", ep)

	if len(ep.Targets) == 0 {
		log.Warnf("no targets specified for endpoint %s, nothing to delete", ep.DNSName)
		return nil
	}

	// Find records that match this endpoint using server-side filtering for better performance
	allRecords, err := c.GetDNSRecords(DNSRecordFilter{
		Name: ep.DNSName,
		Type: ep.RecordType,
	})
	if err != nil {
		return fmt.Errorf("failed to get DNS records for %s::%s: %w", ep.RecordType, ep.DNSName, err)
	}

	// Find matching records based on targets
	var recordsToDelete []DNSRecord
	for _, record := range allRecords {
		recordTarget, err := record.toExternalDNSTarget()
		if err != nil {
			log.Warnf("Skipping record with unsupported type '%s': %+v", record.Type, record)
			continue
		}
		if recordTarget == "" {
			log.Warnf("Skipping record with empty target: %+v", record)
			continue
		}
		log.Debugf("Checking record: Name='%s', Type='%s', Target='%s'", record.Name, record.Type, recordTarget)

		// Check if this record's target is in the list of targets to delete
		if slices.Contains(ep.Targets, recordTarget) {
			// TODO: Consider also matching by TTL and providerSpecific if provided in the endpoint
			log.Debugf("Target matches: '%s', adding to delete list", recordTarget)
			recordsToDelete = append(recordsToDelete, record)
		}
	}

	if len(recordsToDelete) == 0 {
		log.Warnf("No DNS records found to delete for endpoint %s", ep.DNSName)
		return nil
	}

	// Delete records directly using their fixed IDs from the initial query
	for i, record := range recordsToDelete {
		log.Debugf("deleting DNS record %d/%d: %s (ID: %s)", i+1, len(recordsToDelete), record.Name, record.ID)

		// Perform the actual deletion using the original record ID
		resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("ip/dns/static/%s", record.ID), nil, nil)
		if err != nil {
			log.Errorf("error deleting DNS record %s: %v", record.ID, err)
			return err
		}
		resp.Body.Close()
		log.Debugf("record deleted successfully: %s", record.ID)
	}

	log.Infof("successfully deleted %d DNS records", len(recordsToDelete))
	return nil
}

// CreateDNSRecords creates multiple DNS records in batch
func (c *MikrotikApiClient) CreateDNSRecords(ep *endpoint.Endpoint) ([]*DNSRecord, error) {
	log.Infof("creating DNS records for endpoint: %+v", ep)

	// Convert endpoint to multiple DNS records
	records, err := NewDNSRecords(ep)
	if err != nil {
		return nil, fmt.Errorf("failed to convert endpoint to DNS records: %w", err)
	}

	var createdRecords []*DNSRecord
	var lastError error

	for i, record := range records {
		log.Debugf("creating DNS record %d/%d: %+v", i+1, len(records), record)

		// Enforce Default TTL
		if record.TTL == "0s" && c.DefaultTTL > 0 {
			log.Debugf("Setting default TTL for created record: %+v", record)
			record.TTL, _ = EndpointTTLtoMikrotikTTL(endpoint.TTL(c.DefaultTTL))
		}

		// Enforce Default Comment
		if c.DefaultComment != "" {
			log.Debugf("Default comment configured. Checking records comment...")
			if record.Comment != "" {
				log.Debugf("Record already has a comment, skipping default comment: %+v", record)
			} else {
				log.Debugf("Setting default comment for created record: %+v", record)
				record.Comment = c.DefaultComment
			}
		}

		// Serialize the data to JSON to be sent to the API
		jsonBody, err := json.Marshal(record)
		if err != nil {
			log.Errorf("error marshalling DNS record: %v, continuing with next record", err)
			lastError = err
			continue
		}

		// Send the request
		resp, err := c.doRequest(http.MethodPut, "ip/dns/static", nil, bytes.NewReader(jsonBody))
		if err != nil {
			// Keep track of the last error but continue with the next record
			// This will be handled in the next webhook synchronization
			log.Errorf("error creating DNS record %d: %v, continuing with next record", i+1, err)
			lastError = err
			continue
		}
		defer resp.Body.Close()

		// Parse the response
		var createdRecord DNSRecord
		if err = json.NewDecoder(resp.Body).Decode(&createdRecord); err != nil {
			log.Errorf("Error decoding response body: %v, continuing with next record", err)
			lastError = err
			continue
		}
		log.Debugf("created record: %+v", createdRecord)

		createdRecords = append(createdRecords, &createdRecord)
	}

	log.Infof("successfully created %d DNS records", len(createdRecords))

	// If no records were successfully created and we have errors, return the last error
	if len(createdRecords) == 0 && lastError != nil {
		return nil, lastError
	}

	return createdRecords, nil
}

type DNSRecordFilter struct {
	Name string
	Type string
}

// Convert a DNSRecordFilter to URL Query Parameters for the RouterOS API
func (f DNSRecordFilter) toQueryParams() url.Values {
	queryParams := url.Values{}

	if f.Name != "" {
		queryParams.Set("name", f.Name)
	}

	if f.Type != "" {
		queryParams.Set("type", f.Type)
	} else {
		// Default to common record types if no type specified
		queryParams.Set("type", "A,AAAA,CNAME,TXT,MX,SRV,NS")
	}

	return queryParams
}

// encodeQueryParams custom encodes query parameters for MikroTik API
// Special handling: type parameter commas should not be URL-encoded and spaces should be %20
func (c *MikrotikApiClient) encodeQueryParams(params url.Values) string {
	if len(params) == 0 {
		return ""
	}

	var parts []string
	for key, values := range params {
		for _, value := range values {
			if key == "type" {
				// For type parameter, don't encode commas
				parts = append(parts, fmt.Sprintf("type=%s", value))
			} else {
				// For other parameters, replace '+' with '%20' for spaces
				parts = append(parts, fmt.Sprintf("%s=%s", url.QueryEscape(key), strings.ReplaceAll(url.QueryEscape(value), "+", "%20")))
			}
		}
	}
	return strings.Join(parts, "&")
}

// doRequest sends an HTTP request to the MikroTik API with credentials
// queryParams will be URL-encoded and appended to the path
func (c *MikrotikApiClient) doRequest(method, path string, queryParams url.Values, body io.Reader) (*http.Response, error) {
	// Build URL with query parameters
	baseURL := fmt.Sprintf("%s/rest/%s", c.BaseUrl, path)

	// Add query parameters if provided
	if len(queryParams) > 0 {
		baseURL += "?" + c.encodeQueryParams(queryParams)
	}

	log.Debugf("sending %s request to: %s", method, baseURL)

	req, err := http.NewRequest(method, baseURL, body)
	if err != nil {
		log.Errorf("failed to create HTTP request: %v", err)
		return nil, err
	}

	req.SetBasicAuth(c.Username, c.Password)

	resp, err := c.Do(req)
	if err != nil {
		log.Errorf("error sending HTTP request: %v", err)
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		respBody, _ := io.ReadAll(resp.Body)
		log.Errorf("request failed with status %s, response: %s", resp.Status, string(respBody))
		return nil, fmt.Errorf("request failed: %s", resp.Status)
	}
	log.Debugf("request succeeded with status %s", resp.Status)

	return resp, nil
}
