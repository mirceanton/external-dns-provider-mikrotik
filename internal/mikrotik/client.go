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
	"slices"

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
	resp, err := c.doRequest(http.MethodGet, "system/resource", "", nil)
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
	log.Debugf("fetching DNS records matching Name='%s' and Type='%s'", filter.Name, filter.Type)

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

// DeleteRecordsFromEndpoint deletes all DNS records associated with an endpoint
func (c *MikrotikApiClient) DeleteRecordsFromEndpoint(ep *endpoint.Endpoint) error {
	log.Infof("deleting DNS records for endpoint: %+v", ep)

	if len(ep.Targets) == 0 {
		log.Warnf("no targets specified for endpoint %s, nothing to delete", ep.DNSName)
		return nil
	}

	// Find records that match this endpoint
	allRecords, err := c.GetDNSRecords(DNSRecordFilter{Name: ep.DNSName, Type: ep.RecordType})
	if err != nil {
		return fmt.Errorf("failed to get DNS records for %s::%s: %w", ep.RecordType, ep.DNSName, err)
	}

	// Match records to delete based on targets
	// TODO: maybe we can do this filtering server-side?
	for _, record := range allRecords {
		log.Debugf("Checking record: %+v", record)

		recordTarget, err := record.toExternalDNSTarget()
		if err != nil {
			log.Errorf("error converting record to external-dns target: %v", err)
			continue
		}

		if slices.Contains(ep.Targets, recordTarget) {
			// TODO: Consider also matching by TTL and providerSpecific if provided in the endpoint
			err := c.deleteDNSRecord(&record)
			if err != nil {
				log.Errorf("error deleting DNS record %s: %v", record.ID, err)
				return err
			}
		}
	}

	return nil
}

// CreateRecordsFromEndpoint creates multiple DNS records in batch
func (c *MikrotikApiClient) CreateRecordsFromEndpoint(ep *endpoint.Endpoint) ([]*DNSRecord, error) {
	log.Infof("creating DNS records for endpoint: %+v", ep)

	if len(ep.Targets) == 0 {
		log.Warnf("no targets specified for endpoint %s, nothing to delete", ep.DNSName)
		return nil, nil
	}

	// Convert endpoint to multiple DNS records
	records, err := NewDNSRecords(ep)
	if err != nil {
		return nil, fmt.Errorf("failed to convert endpoint to DNS records: %w", err)
	}

	createdRecords := []*DNSRecord{}
	for _, record := range records {
		created, err := c.createDNSRecord(record)
		if err != nil {
			return nil, fmt.Errorf("failed to create DNS record: %w", err)
		}
		createdRecords = append(createdRecords, created)
	}

	log.Infof("successfully created %d DNS records", len(createdRecords))
	return createdRecords, nil
}

// createDNSRecord creates a single DNS record
func (c *MikrotikApiClient) createDNSRecord(record *DNSRecord) (*DNSRecord, error) {
	log.Infof("creating DNS record: %+v", record)

	// Enforce Default TTL
	if record.TTL == "0s" && c.DefaultTTL > 0 {
		log.Debugf("Setting default TTL for created record: %+v", record)
		record.TTL, _ = EndpointTTLtoMikrotikTTL(endpoint.TTL(c.DefaultTTL))
	}

	// Enforce Default Comment
	if c.DefaultComment != "" {
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
		return nil, fmt.Errorf("error marshalling DNS record: %w", err)
	}

	// Send the request
	resp, err := c.doRequest(http.MethodPut, "ip/dns/static", "", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating DNS record: %w", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var createdRecord DNSRecord
	if err = json.NewDecoder(resp.Body).Decode(&createdRecord); err != nil {
		return nil, fmt.Errorf("error decoding response body for record: %w", err)
	}
	log.Debugf("created record: %+v", createdRecord)

	return &createdRecord, nil
}

// deleteDNSRecord deletes a single DNS record
func (c *MikrotikApiClient) deleteDNSRecord(record *DNSRecord) error {
	log.Infof("deleting DNS record (ID: %s)", record.ID)

	resp, err := c.doRequest(http.MethodDelete, fmt.Sprintf("ip/dns/static/%s", record.ID), "", nil)
	if err != nil {
		log.Errorf("error deleting DNS record %s: %v", record.ID, err)
		return err
	}
	defer resp.Body.Close()
	log.Debugf("record deleted successfully: %s", record.ID)

	return nil
}

// doRequest sends an HTTP request to the MikroTik API with credentials
// queryString will be appended to the path as-is (should already be encoded)
func (c *MikrotikApiClient) doRequest(method, path string, queryString string, body io.Reader) (*http.Response, error) {
	// Build URL with query parameters
	baseURL := fmt.Sprintf("%s/rest/%s", c.BaseUrl, path)

	// Add query parameters if provided
	if queryString != "" {
		baseURL += "?" + queryString
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
		_ = resp.Body.Close()
		log.Errorf("request failed with status %s, response: %s", resp.Status, string(respBody))
		return nil, fmt.Errorf("request failed: %s", resp.Status)
	}
	log.Debugf("request succeeded with status %s", resp.Status)

	return resp, nil
}
