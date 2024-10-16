// client_test.go
package mikrotik

import (
	"net/http"
	"testing"
)

func TestNewMikrotikClient(t *testing.T) {
	config := &Config{
		Host:          "test-host",
		Port:          "443",
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
