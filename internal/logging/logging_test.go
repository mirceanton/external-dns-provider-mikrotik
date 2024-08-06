package logging

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSetLogFormat(t *testing.T) {
	// Disable log output for tests
	log.SetOutput(os.NewFile(0, os.DevNull))

	// Define test cases
	tests := []struct {
		envValue string
		expected log.Formatter
	}{
		{"text", &log.TextFormatter{}},
		{"Text", &log.TextFormatter{}},
		{"TEXT", &log.TextFormatter{}},
		{"json", &log.JSONFormatter{}},
		{"Json", &log.JSONFormatter{}},
		{"JSON", &log.JSONFormatter{}},
		{"unknown", &log.TextFormatter{}},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.envValue, func(t *testing.T) {
			os.Setenv("LOG_FORMAT", tt.envValue)
			setLogFormat()
			assert.IsType(t, tt.expected, log.StandardLogger().Formatter)
		})
	}

	// Reset log output
	log.SetOutput(os.Stdout)
}

func TestSetLogLevel(t *testing.T) {
	// Disable log output for tests
	log.SetOutput(os.NewFile(0, os.DevNull))

	// Define test cases
	tests := []struct {
		envValue string
		expected log.Level
	}{
		{"debug", log.DebugLevel},
		{"Debug", log.DebugLevel},
		{"DEBUG", log.DebugLevel},
		{"info", log.InfoLevel},
		{"Info", log.InfoLevel},
		{"INFO", log.InfoLevel},
		{"warn", log.WarnLevel},
		{"Warn", log.WarnLevel},
		{"WARN", log.WarnLevel},
		{"error", log.ErrorLevel},
		{"Error", log.ErrorLevel},
		{"ERROR", log.ErrorLevel},
		{"unknown", log.InfoLevel},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.envValue, func(t *testing.T) {
			os.Setenv("LOG_LEVEL", tt.envValue)
			setLogLevel()
			assert.Equal(t, tt.expected, log.GetLevel())
		})
	}

	// Reset log output
	log.SetOutput(os.Stdout)
}

func TestInit(t *testing.T) {
	os.Setenv("LOG_FORMAT", "json")
	os.Setenv("LOG_LEVEL", "debug")

	Init()

	assert.IsType(t, &log.JSONFormatter{}, log.StandardLogger().Formatter)
	assert.Equal(t, log.DebugLevel, log.GetLevel())
}
