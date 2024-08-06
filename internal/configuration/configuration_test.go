package configuration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInitDefaults(t *testing.T) {
	// Set environment variables for testing
	t.Setenv("SERVER_READ_TIMEOUT", "1s")
	t.Setenv("SERVER_WRITE_TIMEOUT", "1s")

	cfg := Init()

	assert.Equal(t, "localhost", cfg.ServerHost)
	assert.Equal(t, 8888, cfg.ServerPort)
	assert.Equal(t, 1*time.Second, cfg.ServerReadTimeout)
	assert.Equal(t, 1*time.Second, cfg.ServerWriteTimeout)
	assert.Equal(t, []string(nil), cfg.DomainFilter)
	assert.Equal(t, []string(nil), cfg.ExcludeDomains)
	assert.Equal(t, "", cfg.RegexDomainFilter)
	assert.Equal(t, "", cfg.RegexDomainExclusion)
}

func TestInitWithEnvVariables(t *testing.T) {
	// Set environment variables for testing
	t.Setenv("SERVER_HOST", "127.0.0.1")
	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("SERVER_READ_TIMEOUT", "2s")
	t.Setenv("SERVER_WRITE_TIMEOUT", "2s")
	t.Setenv("DOMAIN_FILTER", "example.com,example.org")
	t.Setenv("EXCLUDE_DOMAIN_FILTER", "exclude.com,exclude.org")
	t.Setenv("REGEXP_DOMAIN_FILTER", ".*\\.example\\.com")
	t.Setenv("REGEXP_DOMAIN_FILTER_EXCLUSION", ".*\\.exclude\\.com")

	cfg := Init()

	assert.Equal(t, "127.0.0.1", cfg.ServerHost)
	assert.Equal(t, 9090, cfg.ServerPort)
	assert.Equal(t, 2*time.Second, cfg.ServerReadTimeout)
	assert.Equal(t, 2*time.Second, cfg.ServerWriteTimeout)
	assert.Equal(t, []string{"example.com", "example.org"}, cfg.DomainFilter)
	assert.Equal(t, []string{"exclude.com", "exclude.org"}, cfg.ExcludeDomains)
	assert.Equal(t, ".*\\.example\\.com", cfg.RegexDomainFilter)
	assert.Equal(t, ".*\\.exclude\\.com", cfg.RegexDomainExclusion)
}
