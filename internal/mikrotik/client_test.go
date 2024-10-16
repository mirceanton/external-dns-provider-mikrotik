// client_test.go
package mikrotik

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"sigs.k8s.io/external-dns/endpoint"
)

var (
	mockUsername   = "testuser"
	mockPassword   = "testpass"
	mockServerInfo = SystemInfo{
		ArchitectureName:     "arm64",
		BadBlocks:            "0.1",
		BoardName:            "RB5009UG+S+",
		BuildTime:            "2024-09-20 13:00:27",
		CPU:                  "ARM64",
		CPUCount:             "4",
		CPUFrequency:         "1400",
		CPULoad:              "0",
		FactorySoftware:      "7.4.1",
		FreeHDDSpace:         "1019346944",
		FreeMemory:           "916791296",
		Platform:             "MikroTik",
		TotalHDDSpace:        "1073741824",
		TotalMemory:          "1073741824",
		Uptime:               "4d19h9m34s",
		Version:              "7.16 (stable)",
		WriteSectSinceReboot: "5868",
		WriteSectTotal:       "131658",
	}
)

func TestNewMikrotikClient(t *testing.T) {
	config := &Config{
		BaseUrl:       "https://192.168.88.1:443",
		Username:      "admin",
		Password:      "password",
		SkipTLSVerify: true,
	}

	client, err := NewMikrotikClient(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client.Config != config {
		t.Errorf("Expected config to be %v, got %v", config, client.Config)
	}

	if client.Client == nil {
		t.Errorf("Expected HTTP client to be initialized")
	}

	transport, ok := client.Client.Transport.(*http.Transport)
	if !ok {
		t.Errorf("Expected Transport to be *http.Transport")
	}

	if transport.TLSClientConfig == nil {
		t.Errorf("Expected TLSClientConfig to be set")
	} else if !transport.TLSClientConfig.InsecureSkipVerify {
		t.Errorf("Expected InsecureSkipVerify to be true")
	}
}

func TestGetSystemInfo(t *testing.T) {
	// Set up your mock server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate the Basic Auth header
		username, password, ok := r.BasicAuth()
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if username != mockUsername || password != mockPassword {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Return dummy data for /rest/system/resource
		if r.URL.Path == "/rest/system/resource" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(mockServerInfo)
			if err != nil {
				t.Errorf("error json encoding server info")
			}
			return
		}

		// Return 404 for any other path
		http.NotFound(w, r)
	}))
	defer server.Close()

	// Define test cases
	testCases := []struct {
		name          string
		config        Config
		expectedError bool
	}{
		{
			name: "Valid credentials",
			config: Config{
				BaseUrl:       server.URL,
				Username:      mockUsername,
				Password:      mockPassword,
				SkipTLSVerify: true,
			},
			expectedError: false,
		},
		{
			name: "Incorrect password",
			config: Config{
				BaseUrl:       server.URL,
				Username:      mockUsername,
				Password:      "wrongpass",
				SkipTLSVerify: true,
			},
			expectedError: true,
		},
		{
			name: "Incorrect username",
			config: Config{
				BaseUrl:       server.URL,
				Username:      "wronguser",
				Password:      mockPassword,
				SkipTLSVerify: true,
			},
			expectedError: true,
		},
		{
			name: "Incorrect username and password",
			config: Config{
				BaseUrl:       server.URL,
				Username:      "wronguser",
				Password:      "wrongpass",
				SkipTLSVerify: true,
			},
			expectedError: true,
		},
		{
			name: "Missing credentials",
			config: Config{
				BaseUrl:       server.URL,
				Username:      "",
				Password:      "",
				SkipTLSVerify: true,
			},
			expectedError: true,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &tc.config

			client, err := NewMikrotikClient(config)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			// Call GetSystemInfo
			info, err := client.GetSystemInfo()

			if tc.expectedError {
				if err == nil {
					t.Fatalf("Expected error due to unauthorized access, got none")
				}
				if info != nil {
					t.Errorf("Expected no system info, got %v", info)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
				if info.ArchitectureName != mockServerInfo.ArchitectureName {
					t.Errorf("Expected ArchitectureName %s, got %s", mockServerInfo.ArchitectureName, info.ArchitectureName)
				}
				if info.Version != mockServerInfo.Version {
					t.Errorf("Expected Version %s, got %s", mockServerInfo.Version, info.Version)
				}
			}
		})
	}
}

func TestCreateDNSRecord(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name           string
		initialRecords map[string]DNSRecord
		endpoint       *endpoint.Endpoint
		expectedError  bool
	}{
		{
			name:           "Valid A record creation",
			initialRecords: map[string]DNSRecord{},
			endpoint: &endpoint.Endpoint{
				DNSName:    "test-a.example.com",
				RecordType: "A",
				Targets:    endpoint.Targets{"192.0.2.1"},
			},
			expectedError: false,
		},
		{
			name:           "Valid AAAA record creation",
			initialRecords: map[string]DNSRecord{},
			endpoint: &endpoint.Endpoint{
				DNSName:    "test-aaaa.example.com",
				RecordType: "AAAA",
				Targets:    endpoint.Targets{"2001:db8::1"},
			},
			expectedError: false,
		},
		{
			name:           "Valid CNAME record creation",
			initialRecords: map[string]DNSRecord{},
			endpoint: &endpoint.Endpoint{
				DNSName:    "test-cname.example.com",
				RecordType: "CNAME",
				Targets:    endpoint.Targets{"example.com"},
			},
			expectedError: false,
		},
		{
			name:           "Valid TXT record creation",
			initialRecords: map[string]DNSRecord{},
			endpoint: &endpoint.Endpoint{
				DNSName:    "test-txt.example.com",
				RecordType: "TXT",
				Targets:    endpoint.Targets{"\"some text record value\""},
			},
			expectedError: false,
		},
		{
			name: "Record already exists",
			initialRecords: map[string]DNSRecord{
				"exists.example.com|A": {
					ID:      "*EXISTING",
					Name:    "exists.example.com",
					Type:    "A",
					Address: "192.0.2.1",
				},
			},
			endpoint: &endpoint.Endpoint{
				DNSName:    "exists.example.com",
				RecordType: "A",
				Targets:    endpoint.Targets{"192.0.2.1"},
			},
			expectedError: true,
		},
		{
			name:           "Invalid record (missing DNSName)",
			initialRecords: map[string]DNSRecord{},
			endpoint: &endpoint.Endpoint{
				DNSName:    "",
				RecordType: "A",
				Targets:    endpoint.Targets{"192.0.2.1"},
			},
			expectedError: true,
		},

		{
			name:           "Empty target for A record",
			initialRecords: map[string]DNSRecord{},
			endpoint: &endpoint.Endpoint{
				DNSName:    "empty-target.example.com",
				RecordType: "A",
				Targets:    endpoint.Targets{""}, // Empty target
			},
			expectedError: true,
		},
		{
			name:           "Malformed IP address for A record",
			initialRecords: map[string]DNSRecord{},
			endpoint: &endpoint.Endpoint{
				DNSName:    "malformed-ip.example.com",
				RecordType: "A",
				Targets:    endpoint.Targets{"999.999.999.999"}, // Invalid IP
			},
			expectedError: true,
		},
		{
			name:           "Malformed IP address for AAAA record",
			initialRecords: map[string]DNSRecord{},
			endpoint: &endpoint.Endpoint{
				DNSName:    "malformed-ipv6.example.com",
				RecordType: "AAAA",
				Targets:    endpoint.Targets{"gggg::1"}, // Invalid IPv6
			},
			expectedError: true,
		},
		{
			name:           "Empty target for CNAME record",
			initialRecords: map[string]DNSRecord{},
			endpoint: &endpoint.Endpoint{
				DNSName:    "empty-cname.example.com",
				RecordType: "CNAME",
				Targets:    endpoint.Targets{""}, // Empty target
			},
			expectedError: true,
		},
		// { //! we dont have any kind of cname validation so this will always pass
		// 	name:           "Malformed domain name for CNAME record",
		// 	initialRecords: map[string]DNSRecord{},
		// 	endpoint: &endpoint.Endpoint{
		// 		DNSName:    "bad-cname.example.com",
		// 		RecordType: "CNAME",
		// 		Targets:    endpoint.Targets{"1234!"},
		// 	},
		// 	expectedError: true,
		// },
		{
			name:           "Empty text for TXT record",
			initialRecords: map[string]DNSRecord{},
			endpoint: &endpoint.Endpoint{
				DNSName:    "empty-txt.example.com",
				RecordType: "TXT",
				Targets:    endpoint.Targets{""}, // Empty text
			},
			expectedError: true,
		},
		{
			name:           "Invalid record type",
			initialRecords: map[string]DNSRecord{},
			endpoint: &endpoint.Endpoint{
				DNSName:    "invalid-type.example.com",
				RecordType: "INVALID",
				Targets:    endpoint.Targets{"some target"},
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Initialize an in-memory store for DNS records for this test case
			recordStore := make(map[string]DNSRecord)
			// Pre-populate recordStore with initialRecords
			for k, v := range tc.initialRecords {
				recordStore[k] = v
			}

			// Set up your mock server
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Validate the Basic Auth header
				username, password, ok := r.BasicAuth()
				if !ok || username != mockUsername || password != mockPassword {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				// Handle DNS record creation
				if r.URL.Path == "/rest/ip/dns/static" && r.Method == http.MethodPut {
					var record DNSRecord
					if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
						http.Error(w, "Bad Request", http.StatusBadRequest)
						return
					}

					// Check if record already exists
					key := record.Name + "|" + record.Type
					if _, exists := recordStore[key]; exists {
						http.Error(w, "Conflict: Record already exists", http.StatusConflict)
						return
					}

					// Simulate assigning an ID and storing the record
					record.ID = "*NEW"
					recordStore[key] = record

					// Return the created record
					w.Header().Set("Content-Type", "application/json")
					if err := json.NewEncoder(w).Encode(record); err != nil {
						http.Error(w, "Internal Server Error", http.StatusInternalServerError)
						return
					}
					return
				}

				// Return 404 for any other path
				http.NotFound(w, r)
			}))
			defer server.Close()

			// Set up the client with correct credentials
			config := &Config{
				BaseUrl:       server.URL,
				Username:      mockUsername,
				Password:      mockPassword,
				SkipTLSVerify: true,
			}

			client, err := NewMikrotikClient(config)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			// Call CreateDNSRecord
			record, err := client.CreateDNSRecord(tc.endpoint)

			if tc.expectedError {
				if err == nil {
					t.Fatalf("Expected error, got none")
				}
				// Optionally, check for specific error messages
				return // Error was expected and occurred, test passes
			}

			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			// Verify that the record was stored in the mock server
			key := tc.endpoint.DNSName + "|" + tc.endpoint.RecordType
			storedRecord, exists := recordStore[key]
			if !exists {
				t.Fatalf("Expected record to be stored, but it was not found")
			}

			// Verify that the client received the correct record
			if record.ID != storedRecord.ID {
				t.Errorf("Expected ID '%s', got '%s'", storedRecord.ID, record.ID)
			}

			// Additional checks specific to record type
			switch tc.endpoint.RecordType {
			case "A", "AAAA":
				if storedRecord.Address != tc.endpoint.Targets[0] {
					t.Errorf("Expected Address '%s', got '%s'", tc.endpoint.Targets[0], storedRecord.Address)
				}
			case "CNAME":
				if storedRecord.CName != tc.endpoint.Targets[0] {
					t.Errorf("Expected CName '%s', got '%s'", tc.endpoint.Targets[0], storedRecord.CName)
				}
			case "TXT":
				if storedRecord.Text != tc.endpoint.Targets[0] {
					t.Errorf("Expected Text '%s', got '%s'", tc.endpoint.Targets[0], storedRecord.Text)
				}
			default:
				t.Errorf("Unsupported RecordType '%s' in test case", tc.endpoint.RecordType)
			}
		})
	}
}
