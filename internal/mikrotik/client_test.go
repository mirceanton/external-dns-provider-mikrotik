// client_test.go
package mikrotik

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestGetSystemInfo(t *testing.T) { // Set up your mock server
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

		// Return dummy data
		if r.URL.Path == "/rest/system/resource" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockServerInfo)
			return
		}

		// 404 on anything else
		http.NotFound(w, r)
	}))
	defer server.Close()

	// Set up the client with the test server's URL
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

	// Call GetSystemInfo
	info, err := client.GetSystemInfo()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the returned system info
	if info.ArchitectureName != mockServerInfo.ArchitectureName {
		t.Errorf("Expected ArchitectureName %s, got %s", mockServerInfo.ArchitectureName, info.ArchitectureName)
	}
	if info.Version != mockServerInfo.Version {
		t.Errorf("Expected Version %s, got %s", mockServerInfo.Version, info.Version)
	}
}

func TestGetSystemInfo_Unauthorized(t *testing.T) {
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

	// Set up the client with incorrect credentials
	config := &Config{
		BaseUrl:       server.URL + "/rest",
		Username:      "wronguser",
		Password:      "wrongpass",
		SkipTLSVerify: true,
	}

	client, err := NewMikrotikClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Call GetSystemInfo and expect an error
	_, err = client.GetSystemInfo()
	if err == nil {
		t.Fatalf("Expected error due to unauthorized access, got none")
	}

	// Optionally, check that the error message contains expected text
	if !strings.Contains(err.Error(), "request failed") {
		t.Errorf("Expected error message to contain 'request failed', got %v", err)
	}
}
