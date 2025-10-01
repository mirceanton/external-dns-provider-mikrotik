// client_test.go
package mikrotik

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"sigs.k8s.io/external-dns/endpoint"
)

var (
	mockUsername = "testuser"
	mockPassword = "testpass"
)

func TestNewMikrotikClient(t *testing.T) {
	config := &MikrotikConnectionConfig{
		BaseUrl:       "https://192.168.88.1:443",
		Username:      "admin",
		Password:      "password",
		SkipTLSVerify: true,
	}

	defaults := &MikrotikDefaults{
		DefaultTTL:     1900,
		,
	}

	client, err := NewMikrotikClient(config, defaults)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client.MikrotikConnectionConfig != config {
		t.Errorf("Expected config to be %v, got %v", config, client.MikrotikConnectionConfig)
	}

	if client.MikrotikDefaults != defaults {
		t.Errorf("Expected defaults to be %v, got %v", defaults, client.MikrotikDefaults)
	}

	if client.Client == nil {
		t.Errorf("Expected HTTP client to be initialized")
	}

	transport, ok := client.Transport.(*http.Transport)
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
	mockServerInfo := MikrotikSystemInfo{
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

	// Set up mock server
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
		config        MikrotikConnectionConfig
		defaults      MikrotikDefaults
		expectedError bool
	}{
		{
			name: "Valid credentials",
			config: MikrotikConnectionConfig{
				BaseUrl:       server.URL,
				Username:      mockUsername,
				Password:      mockPassword,
				SkipTLSVerify: true,
			},
			defaults:      MikrotikDefaults{},
			expectedError: false,
		},
		{
			name: "Incorrect password",
			config: MikrotikConnectionConfig{
				BaseUrl:       server.URL,
				Username:      mockUsername,
				Password:      "wrongpass",
				SkipTLSVerify: true,
			},
			defaults:      MikrotikDefaults{},
			expectedError: true,
		},
		{
			name: "Incorrect username",
			config: MikrotikConnectionConfig{
				BaseUrl:       server.URL,
				Username:      "wronguser",
				Password:      mockPassword,
				SkipTLSVerify: true,
			},
			defaults:      MikrotikDefaults{},
			expectedError: true,
		},
		{
			name: "Incorrect username and password",
			config: MikrotikConnectionConfig{
				BaseUrl:       server.URL,
				Username:      "wronguser",
				Password:      "wrongpass",
				SkipTLSVerify: true,
			},
			defaults:      MikrotikDefaults{},
			expectedError: true,
		},
		{
			name: "Missing credentials",
			config: MikrotikConnectionConfig{
				BaseUrl:       server.URL,
				Username:      "",
				Password:      "",
				SkipTLSVerify: true,
			},
			defaults:      MikrotikDefaults{},
			expectedError: true,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &tc.config
			defaults := &tc.defaults

			client, err := NewMikrotikClient(config, defaults)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}
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
				// i think there's no point in checking any more fields
			}
		})
	}
}

func TestGetDNSRecordsByName(t *testing.T) {
	testCases := []struct {
		name          string
		targetName    string
		records       []DNSRecord
		expectedCount int
		expectError   bool
		expectedQuery string
	}{
		{
			name:       "Filter by specific name",
			targetName: "example.com",
			records: []DNSRecord{
				{
					ID:      "*1",
					Address: "1.2.3.4",
					Name:    "example.com",
					TTL:     "1h",
					Type:    "A",
				},
				{
					ID:    "*2",
					CName: "example.com",
					Name:  "www.example.com",
					TTL:   "1h",
					Type:  "CNAME",
				},
			},
			expectedCount: 1, // Only records with exact name match
			expectedQuery: "ip/dns/static?type=A,AAAA,CNAME,TXT,MX,SRV,NS&name=example.com",
		},
		{
			name:       "Get all managed records (empty name)",
			targetName: "",
			records: []DNSRecord{
				{
					ID:      "*1",
					Address: "1.2.3.4",
					Name:    "example.com",
					TTL:     "1h",
					Type:    "A",
				},
				{
					ID:      "*2",
					Address: "5.6.7.8",
					Name:    "test.com",
					TTL:     "1h",
					Type:    "A",
				},
			},
			expectedCount: 2, // All managed records
			expectedQuery: "ip/dns/static?type=A,AAAA,CNAME,TXT,MX,SRV,NS",
		},
		{
			name:          "No matching records",
			targetName:    "nonexistent.com",
			records:       []DNSRecord{},
			expectedCount: 0,
			expectedQuery: "ip/dns/static?type=A,AAAA,CNAME,TXT,MX,SRV,NS&name=nonexistent.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Basic Auth validation
				username, password, ok := r.BasicAuth()
				if !ok || username != mockUsername || password != mockPassword {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				// Verify the query path matches expected
				if r.URL.Path != "/rest/ip/dns/static" {
					t.Errorf("Expected path '/rest/ip/dns/static', got '%s'", r.URL.Path)
				}

				// Check query parameters
				if tc.targetName != "" {
					nameParam := r.URL.Query().Get("name")
					if nameParam != tc.targetName {
						t.Errorf("Expected name parameter '%s', got '%s'", tc.targetName, nameParam)
					}
				}

				// Return filtered records based on query
				var filteredRecords []DNSRecord
				for _, record := range tc.records {
					if tc.targetName == "" || record.Name == tc.targetName {
						filteredRecords = append(filteredRecords, record)
					}
				}

				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(filteredRecords); err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
			}))
			defer server.Close()

			config := &MikrotikConnectionConfig{
				BaseUrl:       server.URL,
				Username:      mockUsername,
				Password:      mockPassword,
				SkipTLSVerify: true,
			}
			defaults := &MikrotikDefaults{
				DefaultComment: "external-dns",
			}
			client, err := NewMikrotikClient(config, defaults)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			records, err := client.GetDNSRecordsByNameAndType(tc.targetName, "")

			if tc.expectError {
				if err == nil {
					t.Fatalf("Expected error, got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}

				if len(records) != tc.expectedCount {
					t.Errorf("Expected %d records, got %d", tc.expectedCount, len(records))
				}
			}
		})
	}
}

func TestDeleteDNSRecords(t *testing.T) {
	testCases := []struct {
		name              string
		endpoint          *endpoint.Endpoint
		existingRecords   []DNSRecord
		defaultComment    string
		expectError       bool
		expectedDeletions int
	}{
		{
			name: "Successful deletion of single record",
			endpoint: &endpoint.Endpoint{
				DNSName:    "example.com",
				RecordType: "A",
				Targets:    []string{"1.2.3.4"},
			},
			existingRecords: []DNSRecord{
				{
					ID:      "*1",
					Name:    "example.com",
					Type:    "A",
					Address: "1.2.3.4",
					Comment: "external-dns",
				},
			},
			defaultComment:    "external-dns",
			expectError:       false,
			expectedDeletions: 1,
		},
		{
			name: "Successful deletion of multiple records",
			endpoint: &endpoint.Endpoint{
				DNSName:    "multi.example.com",
				RecordType: "A",
				Targets:    []string{"1.2.3.4", "5.6.7.8"},
			},
			existingRecords: []DNSRecord{
				{
					ID:      "*1",
					Name:    "multi.example.com",
					Type:    "A",
					Address: "1.2.3.4",
					Comment: "external-dns",
				},
				{
					ID:      "*2",
					Name:    "multi.example.com",
					Type:    "A",
					Address: "5.6.7.8",
					Comment: "external-dns",
				},
			},
			defaultComment:    "external-dns",
			expectError:       false,
			expectedDeletions: 2,
		},
		{
			name: "No records found to delete",
			endpoint: &endpoint.Endpoint{
				DNSName:    "nonexistent.com",
				RecordType: "A",
				Targets:    []string{"1.2.3.4"},
			},
			existingRecords:   []DNSRecord{},
			defaultComment:    "external-dns",
			expectError:       false,
			expectedDeletions: 0,
		},
		{
			name: "Partial deletion - delete only specific targets",
			endpoint: &endpoint.Endpoint{
				DNSName:    "partial.example.com",
				RecordType: "A",
				Targets:    []string{"1.2.3.4"}, // Only delete this specific target
			},
			existingRecords: []DNSRecord{
				{
					ID:      "*1",
					Name:    "partial.example.com",
					Type:    "A",
					Address: "1.2.3.4", // This should be deleted
					Comment: "external-dns",
				},
				{
					ID:      "*2",
					Name:    "partial.example.com",
					Type:    "A",
					Address: "5.6.7.8", // This should NOT be deleted (different target)
					Comment: "external-dns",
				},
			},
			defaultComment:    "external-dns",
			expectError:       false,
			expectedDeletions: 1, // Only one record should be deleted
		},
		{
			name: "Partial deletion - delete multiple specific targets",
			endpoint: &endpoint.Endpoint{
				DNSName:    "multi-partial.example.com",
				RecordType: "A",
				Targets:    []string{"1.2.3.4", "5.6.7.8"}, // Delete these two targets
			},
			existingRecords: []DNSRecord{
				{
					ID:      "*1",
					Name:    "multi-partial.example.com",
					Type:    "A",
					Address: "1.2.3.4", // Should be deleted
					Comment: "external-dns",
				},
				{
					ID:      "*2",
					Name:    "multi-partial.example.com",
					Type:    "A",
					Address: "5.6.7.8", // Should be deleted
					Comment: "external-dns",
				},
				{
					ID:      "*3",
					Name:    "multi-partial.example.com",
					Type:    "A",
					Address: "9.10.11.12", // Should NOT be deleted (not in targets)
					Comment: "external-dns",
				},
			},
			defaultComment:    "external-dns",
			expectError:       false,
			expectedDeletions: 2, // Only two records should be deleted
		},
		{
			name: "Partial deletion - target not found",
			endpoint: &endpoint.Endpoint{
				DNSName:    "notfound.example.com",
				RecordType: "A",
				Targets:    []string{"1.2.3.4"}, // This target doesn't exist
			},
			existingRecords: []DNSRecord{
				{
					ID:      "*1",
					Name:    "notfound.example.com",
					Type:    "A",
					Address: "5.6.7.8", // Different target exists
					Comment: "external-dns",
				},
			},
			defaultComment:    "external-dns",
			expectError:       false,
			expectedDeletions: 0, // No records should be deleted
		},
		{
			name: "Partial deletion - mixed CNAME targets",
			endpoint: &endpoint.Endpoint{
				DNSName:    "cname-partial.example.com",
				RecordType: "CNAME",
				Targets:    []string{"target1.example.com"}, // Only delete this CNAME target
			},
			existingRecords: []DNSRecord{
				{
					ID:      "*1",
					Name:    "cname-partial.example.com",
					Type:    "CNAME",
					CName:   "target1.example.com", // Should be deleted
					Comment: "external-dns",
				},
				{
					ID:      "*2",
					Name:    "cname-partial.example.com",
					Type:    "CNAME",
					CName:   "target2.example.com", // Should NOT be deleted
					Comment: "external-dns",
				},
			},
			defaultComment:    "external-dns",
			expectError:       false,
			expectedDeletions: 1,
		},
		{
			name: "Partial deletion - some targets exist, some don't",
			endpoint: &endpoint.Endpoint{
				DNSName:    "mixed-exist.example.com",
				RecordType: "A",
				Targets:    []string{"1.2.3.4", "5.6.7.8", "9.10.11.12"}, // Mix of existing and non-existing
			},
			existingRecords: []DNSRecord{
				{
					ID:      "*1",
					Name:    "mixed-exist.example.com",
					Type:    "A",
					Address: "1.2.3.4", // Exists, should be deleted
					Comment: "external-dns",
				},
				{
					ID:      "*2",
					Name:    "mixed-exist.example.com",
					Type:    "A",
					Address: "9.10.11.12", // Exists, should be deleted
					Comment: "external-dns",
				},
				// Note: 5.6.7.8 doesn't exist in records
			},
			defaultComment:    "external-dns",
			expectError:       false,
			expectedDeletions: 2, // Only the two existing records should be deleted
		},
		{
			name: "Partial deletion - different record types with same name",
			endpoint: &endpoint.Endpoint{
				DNSName:    "mixed-types.example.com",
				RecordType: "A", // Only targeting A records
				Targets:    []string{"1.2.3.4"},
			},
			existingRecords: []DNSRecord{
				{
					ID:      "*1",
					Name:    "mixed-types.example.com",
					Type:    "A",
					Address: "1.2.3.4", // Should be deleted
					Comment: "external-dns",
				},
				{
					ID:      "*2",
					Name:    "mixed-types.example.com",
					Type:    "CNAME",
					CName:   "target.example.com", // Should NOT be deleted (different type)
					Comment: "external-dns",
				},
			},
			defaultComment:    "external-dns",
			expectError:       false,
			expectedDeletions: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deletedCount := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Basic Auth validation
				username, password, ok := r.BasicAuth()
				if !ok || username != mockUsername || password != mockPassword {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				// Handle GET requests to /rest/ip/dns/static (for GetDNSRecordsByName)
				if r.Method == http.MethodGet && r.URL.Path == "/rest/ip/dns/static" {
					w.Header().Set("Content-Type", "application/json")
					if err := json.NewEncoder(w).Encode(tc.existingRecords); err != nil {
						http.Error(w, "Internal Server Error", http.StatusInternalServerError)
						return
					}
					return
				}

				// Handle DELETE requests to /rest/ip/dns/static/*
				if r.Method == http.MethodDelete && len(r.URL.Path) > len("/rest/ip/dns/static/") &&
					r.URL.Path[:len("/rest/ip/dns/static/")] == "/rest/ip/dns/static/" {
					deletedCount++
					w.WriteHeader(http.StatusOK)
					return
				}

				// Return 404 for any other path
				http.NotFound(w, r)
			}))
			defer server.Close()

			// Set up the client
			config := &MikrotikConnectionConfig{
				BaseUrl:       server.URL,
				Username:      mockUsername,
				Password:      mockPassword,
				SkipTLSVerify: true,
			}
			defaults := &MikrotikDefaults{
				DefaultComment: tc.defaultComment,
			}
			client, err := NewMikrotikClient(config, defaults)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			err = client.DeleteDNSRecords(tc.endpoint)

			if tc.expectError {
				if err == nil {
					t.Fatalf("Expected error, got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
				if deletedCount != tc.expectedDeletions {
					t.Errorf("Expected %d deletions, got %d", tc.expectedDeletions, deletedCount)
				}
			}
		})
	}
}

func TestCreateDNSRecords(t *testing.T) {
	testCases := []struct {
		name            string
		endpoint        *endpoint.Endpoint
		expectError     bool
		expectedRecords int
		statusCode      int
	}{
		{
			name: "Successful creation of single A record",
			endpoint: &endpoint.Endpoint{
				DNSName:    "example.com",
				RecordType: "A",
				Targets:    []string{"1.2.3.4"},
				RecordTTL:  endpoint.TTL(3600),
			},
			expectError:     false,
			expectedRecords: 1,
			statusCode:      http.StatusOK,
		},
		{
			name: "Successful creation of multiple A records",
			endpoint: &endpoint.Endpoint{
				DNSName:    "multi.example.com",
				RecordType: "A",
				Targets:    []string{"1.2.3.4", "5.6.7.8"},
				RecordTTL:  endpoint.TTL(3600),
			},
			expectError:     false,
			expectedRecords: 2,
			statusCode:      http.StatusOK,
		},
		{
			name: "API error during creation",
			endpoint: &endpoint.Endpoint{
				DNSName:    "error.example.com",
				RecordType: "A",
				Targets:    []string{"1.2.3.4"},
				RecordTTL:  endpoint.TTL(3600),
			},
			expectError:     true,
			expectedRecords: 0,
			statusCode:      http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			createdCount := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Basic Auth validation
				username, password, ok := r.BasicAuth()
				if !ok || username != mockUsername || password != mockPassword {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				// Handle PUT requests to /rest/ip/dns/static
				if r.Method == http.MethodPut && r.URL.Path == "/rest/ip/dns/static" {
					if tc.statusCode == http.StatusInternalServerError {
						http.Error(w, "Internal Server Error", http.StatusInternalServerError)
						return
					}

					// Parse request body to get record data
					var record DNSRecord
					if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
						http.Error(w, "Bad Request", http.StatusBadRequest)
						return
					}

					createdCount++
					record.ID = "*1"

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

			// Set up the client
			config := &MikrotikConnectionConfig{
				BaseUrl:       server.URL,
				Username:      mockUsername,
				Password:      mockPassword,
				SkipTLSVerify: true,
			}
			defaults := &MikrotikDefaults{
				DefaultTTL:     3600,
				DefaultComment: "external-dns",
			}
			client, err := NewMikrotikClient(config, defaults)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			records, err := client.CreateDNSRecords(tc.endpoint)

			if tc.expectError {
				if err == nil {
					t.Fatalf("Expected error, got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
				if len(records) != tc.expectedRecords {
					t.Errorf("Expected %d records created, got %d", tc.expectedRecords, len(records))
				}
			}
		})
	}
}

func TestCreateSingleDNSRecord(t *testing.T) {
	testCases := []struct {
		name        string
		record      *DNSRecord
		expectError bool
		statusCode  int
	}{
		{
			name: "Successful creation",
			record: &DNSRecord{
				Name:    "example.com",
				Type:    "A",
				Address: "1.2.3.4",
				TTL:     "1h",
				Comment: "test",
			},
			expectError: false,
			statusCode:  http.StatusOK,
		},
		{
			name: "Server error",
			record: &DNSRecord{
				Name:    "error.com",
				Type:    "A",
				Address: "1.2.3.4",
				TTL:     "1h",
				Comment: "test",
			},
			expectError: true,
			statusCode:  http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Basic Auth validation
				username, password, ok := r.BasicAuth()
				if !ok || username != mockUsername || password != mockPassword {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				// Handle PUT requests to /rest/ip/dns/static
				if r.Method == http.MethodPut && r.URL.Path == "/rest/ip/dns/static" {
					if tc.statusCode != http.StatusOK {
						http.Error(w, http.StatusText(tc.statusCode), tc.statusCode)
						return
					}

					// Parse and echo back the record with an ID
					var record DNSRecord
					if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
						http.Error(w, "Bad Request", http.StatusBadRequest)
						return
					}
					record.ID = "*1"

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

			// Set up the client
			config := &MikrotikConnectionConfig{
				BaseUrl:       server.URL,
				Username:      mockUsername,
				Password:      mockPassword,
				SkipTLSVerify: true,
			}
			defaults := &MikrotikDefaults{}
			client, err := NewMikrotikClient(config, defaults)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			createdRecord, err := client.createSingleDNSRecord(tc.record)

			if tc.expectError {
				if err == nil {
					t.Fatalf("Expected error, got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
				if createdRecord == nil {
					t.Fatal("Expected created record, got nil")
				}
				if createdRecord.ID != "*1" {
					t.Errorf("Expected ID '*1', got %s", createdRecord.ID)
				}
				if createdRecord.Name != tc.record.Name {
					t.Errorf("Expected Name %s, got %s", tc.record.Name, createdRecord.Name)
				}
			}
		})
	}
}

func TestDoRequest(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Successful GET request",
			method:         "GET",
			path:           "system/resource",
			body:           "",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Successful POST request with body",
			method:         "POST",
			path:           "ip/dns/static",
			body:           `{"name":"test.com","type":"A","address":"1.2.3.4"}`,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "404 Not Found",
			method:         "GET",
			path:           "nonexistent/path",
			body:           "",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "401 Unauthorized",
			method:         "GET",
			path:           "unauthorized",
			body:           "",
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
		{
			name:           "500 Internal Server Error",
			method:         "GET",
			path:           "error",
			body:           "",
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Handle special paths that simulate different status codes
				if r.URL.Path == "/rest/unauthorized" {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				if r.URL.Path == "/rest/error" {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				if r.URL.Path == "/rest/nonexistent/path" {
					http.NotFound(w, r)
					return
				}

				// Basic Auth validation for normal requests
				username, password, ok := r.BasicAuth()
				if !ok || username != mockUsername || password != mockPassword {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				// Handle valid requests
				if r.URL.Path == "/rest/system/resource" {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(`{"version":"7.16"}`))
					return
				}
				if r.URL.Path == "/rest/ip/dns/static" {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(`{"id":"*1","name":"test.com"}`))
					return
				}

				http.NotFound(w, r)
			}))
			defer server.Close()

			// Set up the client
			config := &MikrotikConnectionConfig{
				BaseUrl:       server.URL,
				Username:      mockUsername,
				Password:      mockPassword,
				SkipTLSVerify: true,
			}
			defaults := &MikrotikDefaults{}
			client, err := NewMikrotikClient(config, defaults)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			var bodyReader io.Reader
			if tc.body != "" {
				bodyReader = bytes.NewReader([]byte(tc.body))
			}

			resp, err := client.doRequest(tc.method, tc.path, nil, bodyReader)

			if tc.expectError {
				if err == nil {
					t.Fatalf("Expected error, got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
				if resp == nil {
					t.Fatal("Expected response, got nil")
				}
				defer resp.Body.Close()
				if resp.StatusCode != tc.expectedStatus {
					t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
				}
			}
		})
	}
}
