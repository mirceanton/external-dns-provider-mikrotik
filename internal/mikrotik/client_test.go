// client_test.go
package mikrotik

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
			json.NewEncoder(w).Encode(mockServerInfo)
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
