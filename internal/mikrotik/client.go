// Rest API Docs: https://help.mikrotik.com/docs/display/ROS/REST+API

package mikrotik

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/publicsuffix"
	"sigs.k8s.io/external-dns/endpoint"
)

// NewMikrotikClient creates a new instance of MikrotikApiClient
func NewMikrotikClient(config *Config) (*MikrotikApiClient, error) {
	log.Infof("Creating a new Mikrotik API Client")

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Errorf("Failed to create cookie jar: %v", err)
		return nil, err
	}

	client := &MikrotikApiClient{
		Config: config,
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: config.SkipTLSVerify,
				},
			},
			Jar: jar,
		},
	}

	log.Infof("Connecting to the MikroTik RouterOS API Endpoint...")
	info, err := client.GetSystemInfo()
	if err != nil {
		log.Errorf("Failed to connect to the MikroTik RouterOS API Endpoint: %v", err)
		return nil, err
	}

	log.Infof("Connected to board %s running RouterOS version %s (%s)", info.BoardName, info.Version, info.ArchitectureName)
	return client, nil
}

// NewDNSRecord converts an ExternalDNS Endpoint to a Mikrotik DNSRecord
func NewDNSRecord(endpoint *endpoint.Endpoint) (*DNSRecord, error) {
	record := DNSRecord{
		Address: endpoint.Targets[0],
		Name:    endpoint.DNSName,
		Type:    endpoint.RecordType,
		Comment: "Managed by ExternalDNS",
	}

	for _, prop := range endpoint.ProviderSpecific {
		switch prop.Name {
		case "cname":
			record.CName = prop.Value
		case "forward-to":
			record.ForwardTo = prop.Value
		case "mx-exchange":
			record.MXExchange = prop.Value
		case "srv-target":
			record.SrvTarget = prop.Value
		case "text":
			record.Text = prop.Value
		case "address-list":
			record.AddressList = prop.Value
		case "ns":
			record.NS = prop.Value
		case "regexp":
			record.Regexp = prop.Value
		case "srv-port":
			record.SrvPort = prop.Value
		case "mx-preference":
			record.MXPreference = prop.Value
		case "srv-priority":
			record.SrvPriority = prop.Value
		case "srv-weight":
			record.SrvWeight = prop.Value
		case "disabled":
			record.Disabled = prop.Value
		case "ttl":
			record.TTL = prop.Value
		}
	}

	return &record, nil
}

// GetSystemInfo fetches system information from the MikroTik API
func (c *MikrotikApiClient) GetSystemInfo() (*SystemInfo, error) {
	log.Infof("Fetching system information.")

	resp, err := c.doRequest(http.MethodGet, "system/resource", nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var info SystemInfo
	if err = json.NewDecoder(resp.Body).Decode(&info); err != nil {
		log.Errorf("Error decoding response body: %v", err)
		return nil, err
	}
	log.Debugf("Got system info: %+v", info)

	return &info, nil
}

// Create sends a request to create a new DNS record
func (c *MikrotikApiClient) Create(endpoint *endpoint.Endpoint) (*DNSRecord, error) {
	log.Infof("Creating DNS record: %+v", endpoint)

	record, err := NewDNSRecord(endpoint)
	if err != nil {
		log.Errorf("Error converting ExternalDNS endpoint to Mikrotik DNS Record: %v", err)
		return nil, err
	}
	jsonBody, err := json.Marshal(record)
	if err != nil {
		log.Errorf("Error marshalling DNS record: %v", err)
		return nil, err
	}

	log.Debugf("JSON body being sent: %s", string(jsonBody))
	resp, err := c.doRequest(http.MethodPut, "ip/dns/static", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&record); err != nil {
		log.Errorf("Error decoding response body: %v", err)
		return nil, err
	}
	log.Infof("Created record: %+v", record)

	return record, nil
}

// GetAll fetches all DNS records from the MikroTik API
func (c *MikrotikApiClient) GetAll() ([]DNSRecord, error) {
	log.Infof("Fetching all DNS records")

	resp, err := c.doRequest(http.MethodGet, "ip/dns/static", nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var records []DNSRecord
	if err = json.NewDecoder(resp.Body).Decode(&records); err != nil {
		log.Errorf("Error decoding response body: %v", err)
		return nil, err
	}

	return records, nil
}

// Delete sends a request to delete a DNS record
func (c *MikrotikApiClient) Delete(endpoint *endpoint.Endpoint) error {
	log.Infof("Deleting DNS record: %+v", endpoint)

	record, err := c.Search(endpoint.DNSName, endpoint.RecordType)
	if err != nil {
		return err
	}

	_, err = c.doRequest(http.MethodDelete, fmt.Sprintf("ip/dns/static/%s", record.ID), nil)
	if err != nil {
		return err
	}
	log.Infof("Deleted record: %+v", record)

	return nil
}

// Search searches for a DNS record by key and type
func (c *MikrotikApiClient) Search(key, recordType string) (*DNSRecord, error) {
	log.Infof("Searching for DNS record: Key: %s, RecordType: %s", key, recordType)

	searchParams := fmt.Sprintf("name=%s", key)
	if recordType != "A" {
		searchParams = fmt.Sprintf("%s&type=%s", searchParams, recordType)
	}
	log.Debugf("Search params: %s", searchParams)

	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("ip/dns/static?%s", searchParams), nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var record []DNSRecord
	if err = json.NewDecoder(resp.Body).Decode(&record); err != nil {
		log.Errorf("Error decoding response body: %v", err)
		return nil, err
	}
	if len(record) == 0 {
		return nil, errors.New("record list is empty")
	}

	log.Infof("Found record: %+v", record)

	return &record[0], nil
}

// doRequest sends an HTTP request to the MikroTik API with credentials
func (c *MikrotikApiClient) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	endpoint_url := fmt.Sprintf("https://%s:%s/rest/%s", c.Config.Host, c.Config.Port, path)
	log.Debugf("Sending %s request to: %s", method, endpoint_url)

	req, err := http.NewRequest(method, endpoint_url, body)
	if err != nil {
		log.Errorf("Failed to create HTTP request: %v", err)
		return nil, err
	}

	req.SetBasicAuth(c.Config.Username, c.Config.Password)

	resp, err := c.Client.Do(req)
	if err != nil {
		log.Errorf("HTTP request failed: %v", err)
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		respBody, _ := io.ReadAll(resp.Body)
		log.Errorf("Request failed with status %s, response: %s", resp.Status, string(respBody))
		return nil, fmt.Errorf("request failed: %s", resp.Status)
	}
	log.Debugf("Request succeeded with status %s", resp.Status)

	return resp, nil
}
