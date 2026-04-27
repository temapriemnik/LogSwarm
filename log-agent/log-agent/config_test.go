package main

import (
	"os"
	"testing"
	"time"
)

var testTime = time.Date(2026, 4, 26, 12, 0, 0, 0, time.UTC)

func TestConfigFromEnv(t *testing.T) {
	os.Setenv("API_KEY", "test-key-123")
	os.Setenv("NATS_URL", "nats://custom:4222")
	defer os.Unsetenv("API_KEY")
	defer os.Unsetenv("NATS_URL")

	cfg := Config{
		APIKey:  os.Getenv("API_KEY"),
		NATSURL: getEnv("NATS_URL", "nats://localhost:4222"),
	}

	if cfg.APIKey != "test-key-123" {
		t.Errorf("expected API_KEY 'test-key-123', got '%s'", cfg.APIKey)
	}

	if cfg.NATSURL != "nats://custom:4222" {
		t.Errorf("expected NATS_URL 'nats://custom:4222', got '%s'", cfg.NATSURL)
	}
}

func TestConfigDefaults(t *testing.T) {
	os.Unsetenv("API_KEY")
	os.Unsetenv("NATS_URL")

	natsURL := getEnv("NATS_URL", "nats://localhost:4222")

	if natsURL != "nats://localhost:4222" {
		t.Errorf("expected default NATS_URL, got '%s'", natsURL)
	}
}

func TestLogMessageJSON(t *testing.T) {
	msg := LogMessage{
		APIToken:  "my-api-key",
		Log:       "container: test log message",
		Timestamp: testTime,
	}

	if msg.APIToken != "my-api-key" {
		t.Errorf("expected api_key 'my-api-key', got '%s'", msg.APIToken)
	}
	if msg.Log != "container: test log message" {
		t.Errorf("expected log 'container: test log message', got '%s'", msg.Log)
	}
}