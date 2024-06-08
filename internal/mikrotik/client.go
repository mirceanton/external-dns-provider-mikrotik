// Rest API Docs: https://help.mikrotik.com/docs/display/ROS/REST+API

package mikrotik

import (
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

type MikrotikApiClient struct {
	*Config
	*http.Client
}

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

func (c *MikrotikApiClient) GetSystemInfo() (*SystemInfo, error) {
	log.Infof("Fetching system information.")

	// no logging here since we have logs in doRequest
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

func (c *MikrotikApiClient) Create(endpoint *endpoint.Endpoint) (*DNSRecord, error) {
	log.Infof("Creating DNS record: %+v", endpoint)

	log.Debugf(fmt.Sprintf("Received request for object %s", endpoint))
	return nil, nil

	// jsonBody, err := json.Marshal(DNSRecord{
	// 	Name:    endpoint.DNSName,
	// 	Address: endpoint.Targets[0],
	// 	Type:    endpoint.RecordType,
	// 	// TTL:     endpoint.RecordTTL,  //FIXME
	// 	Comment: "Managed by ExternalDNS",
	// })
	// if err != nil {
	// 	log.Errorf("Error marshalling DNS record: %v", err)
	// 	return nil, err
	// }

	// // no logging here since we have logs in doRequest
	// resp, err := c.doRequest(http.MethodPut, "ip/dns/static", bytes.NewReader(jsonBody))
	// if err != nil {
	// 	return nil, err
	// }

	// defer resp.Body.Close()

	// var record DNSRecord
	// if err = json.NewDecoder(resp.Body).Decode(&record); err != nil {
	// 	log.Errorf("Error decoding response body: %v", err)
	// 	return nil, err
	// }
	// log.Debugf("Created record: %+v", record)

	// return &record, nil
}

func (c *MikrotikApiClient) GetAll() ([]DNSRecord, error) {
	log.Infof("Fetching all DNS records")

	// no logging here since we have logs in doRequest
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
	log.Debugf("Retrieved records: %+v", records)

	return records, nil
}

func (c *MikrotikApiClient) Delete(endpoint *endpoint.Endpoint) error {
	log.Infof("Deleting DNS record: %+v", endpoint)

	// no logging here since we have logs in Search
	record, err := c.Search(endpoint.DNSName, endpoint.RecordType)
	if err != nil {
		return err
	}

	// no logging here since we have logs in doRequest
	_, err = c.doRequest(http.MethodDelete, fmt.Sprintf("ip/dns/static/%s", record.ID), nil)
	if err != nil {
		return err
	}
	log.Infof("Deleted record: %+v", record)

	return nil
}

func (c *MikrotikApiClient) Search(key, recordType string) (*DNSRecord, error) {
	log.Infof("Searching for DNS record: Key: %s, RecordType: %s", key, recordType)

	searchParams := fmt.Sprintf("name=%s", key)
	if recordType != "A" {
		searchParams = fmt.Sprintf("%s&type=%s", searchParams, recordType)
	}
	log.Debugf("Search params: %s", searchParams)

	// no logging here since we have logs in doRequest
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

	log.Debugf("Found record: %+v", record)

	return &record[0], nil
}

// utility function used to send requests to the mikrotik api with credentials injected in the header
func (c *MikrotikApiClient) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	endpoint_url := fmt.Sprintf("https://%s:%s/rest/%s", c.Config.Host, c.Config.Port, path)
	log.Infof("Sending %s request to: %s", method, endpoint_url)

	req, err := http.NewRequest(method, endpoint_url, body)
	if err != nil {
		log.Errorf("Failed to create HTTP request: %v", err)
		return nil, err
	}

	req.SetBasicAuth(c.Config.Username, c.Config.Password)
	log.Debugf("Auth header set on request.")

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
	log.Infof("Request succeeded with status %s", resp.Status)

	return resp, nil
}
