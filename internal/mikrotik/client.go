package mikrotik

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
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

func newMikrotikClient(config *Config) (*MikrotikApiClient, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
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

	if err := client.SystemInfo(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *MikrotikApiClient) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	endpoint_url := fmt.Sprintf("https://%s:%s/rest/%s", c.Config.Host, c.Config.Port, path)

	req, err := http.NewRequest(method, endpoint_url, body)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.Config.Username, c.Config.Password)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Errorf("request failed: %s, response: %s", resp.Status, string(respBody))
		return nil, fmt.Errorf("request failed: %s", resp.Status)
	}

	return resp, nil
}

// perform a dummy request to the system info endpoint to validate the credentials
func (c *MikrotikApiClient) SystemInfo() error {
	_, err := c.doRequest(http.MethodGet, "system/resource", nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *MikrotikApiClient) Create(endpoint *endpoint.Endpoint) (*DNSRecord, error) {
	jsonBody, err := json.Marshal(DNSRecord{
		Name:    endpoint.DNSName,
		Type:    endpoint.RecordType,
		Address: endpoint.Targets[0],
		TTL:     endpoint.TTL,
		Comment: "Managed by ExternalDNS",
	})
	if err != nil {
		return nil, err
	}
	resp, err := c.doRequest(http.MethodPut, "ip/dns/static", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var record DNSRecord
	if err = json.NewDecoder(resp.Body).Decode(&record); err != nil {
		return nil, err
	}

	log.Debugf("created record: %+v", record)
	return &record, nil
}

func (c *MikrotikApiClient) GetAll() ([]DNSRecord, error) {
	resp, err := c.doRequest(http.MethodGet, "ip/dns/static", nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var records []DNSRecord
	if err = json.NewDecoder(resp.Body).Decode(&records); err != nil {
		return nil, err
	}

	log.Debugf("retrieved records: %+v", records)
	return records, nil
}

func (c *MikrotikApiClient) Delete(endpoint *endpoint.Endpoint) error {
	record, err := c.Search(endpoint.DNSName, endpoint.RecordType)
	if err != nil {
		return err
	}

	_, err = c.doRequest(http.MethodDelete, fmt.Sprintf("ip/dns/static/%s", record.ID), nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *MikrotikApiClient) Search(key, recordType string) (*DNSRecord, error) {
	searchParams := fmt.Sprintf("name=%s", key)
	if recordType != "A" {
		searchParams = fmt.Sprintf("%s&type=%s", searchParams, recordType)
	}

	resp, err := c.doRequest(http.MethodGet, fmt.Sprintf("ip/dns/static?%s", searchParams), nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var record DNSRecord
	if err = json.NewDecoder(resp.Body).Decode(&record); err != nil {
		return nil, err
	}

	log.Debugf("found record: %+v", record)
	return &record, nil
}
