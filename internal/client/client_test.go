package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config with API key",
			config: &Config{
				BaseURL: "https://example.com",
				Auth:    &APIKeyAuth{APIKey: "test-key"},
			},
			wantErr: false,
		},
		{
			name: "valid config with basic auth",
			config: &Config{
				BaseURL: "https://example.com",
				Auth:    &BasicAuth{Email: "test@example.com", Password: "password"},
			},
			wantErr: false,
		},
		{
			name: "missing base URL",
			config: &Config{
				Auth: &APIKeyAuth{APIKey: "test-key"},
			},
			wantErr: true,
		},
		{
			name: "missing auth",
			config: &Config{
				BaseURL: "https://example.com",
			},
			wantErr: true,
		},
		{
			name: "invalid base URL",
			config: &Config{
				BaseURL: ":/invalid-url",
				Auth:    &APIKeyAuth{APIKey: "test-key"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client without error")
			}
		})
	}
}

func TestAPIKeyAuth(t *testing.T) {
	auth := &APIKeyAuth{APIKey: "test-key"}
	req, _ := http.NewRequest("GET", "https://example.com", nil)

	err := auth.ApplyAuth(req)
	if err != nil {
		t.Errorf("APIKeyAuth.ApplyAuth() error = %v", err)
	}

	if got := req.Header.Get("X-N8N-API-KEY"); got != "test-key" {
		t.Errorf("APIKeyAuth.ApplyAuth() header = %v, want %v", got, "test-key")
	}
}

func TestBasicAuth(t *testing.T) {
	auth := &BasicAuth{Email: "test@example.com", Password: "password"}
	req, _ := http.NewRequest("GET", "https://example.com", nil)

	err := auth.ApplyAuth(req)
	if err != nil {
		t.Errorf("BasicAuth.ApplyAuth() error = %v", err)
	}

	username, password, ok := req.BasicAuth()
	if !ok {
		t.Error("BasicAuth.ApplyAuth() did not set basic auth")
	}
	if username != "test@example.com" {
		t.Errorf("BasicAuth.ApplyAuth() username = %v, want %v", username, "test@example.com")
	}
	if password != "password" {
		t.Errorf("BasicAuth.ApplyAuth() password = %v, want %v", password, "password")
	}
}

func TestClient_doRequest(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-N8N-API-KEY") != "test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"code": 401, "message": "Unauthorized"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id": "1", "name": "test"}`))
	}))
	defer server.Close()

	config := &Config{
		BaseURL: server.URL,
		Auth:    &APIKeyAuth{APIKey: "test-key"},
		Timeout: 5 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	var result map[string]interface{}
	err = client.Get("test", &result)
	if err != nil {
		t.Errorf("Client.Get() error = %v", err)
	}

	if result["id"] != "1" {
		t.Errorf("Client.Get() result id = %v, want %v", result["id"], "1")
	}
}

func TestClient_ErrorHandling(t *testing.T) {
	// Create a test server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code": 400, "message": "Bad Request", "details": "Test error"}`))
	}))
	defer server.Close()

	config := &Config{
		BaseURL: server.URL,
		Auth:    &APIKeyAuth{APIKey: "test-key"},
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	var result map[string]interface{}
	err = client.Get("test", &result)

	if err == nil {
		t.Error("Client.Get() expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Errorf("Client.Get() error type = %T, want *APIError", err)
	}

	if apiErr.Code != 400 {
		t.Errorf("APIError.Code = %v, want %v", apiErr.Code, 400)
	}

	if apiErr.Message != "Bad Request" {
		t.Errorf("APIError.Message = %v, want %v", apiErr.Message, "Bad Request")
	}
}

func TestClient_RetryLogic(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	config := &Config{
		BaseURL: server.URL,
		Auth:    &APIKeyAuth{APIKey: "test-key"},
		RetryConfig: RetryConfig{
			MaxRetries: 3,
			BaseDelay:  10 * time.Millisecond,
			MaxDelay:   100 * time.Millisecond,
		},
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	var result map[string]interface{}
	err = client.Get("test", &result)
	if err != nil {
		t.Errorf("Client.Get() with retries error = %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}

	if result["success"] != true {
		t.Errorf("Expected success=true, got %v", result["success"])
	}
}

func TestClient_RetryExhaustion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	config := &Config{
		BaseURL: server.URL,
		Auth:    &APIKeyAuth{APIKey: "test-key"},
		RetryConfig: RetryConfig{
			MaxRetries: 2,
			BaseDelay:  10 * time.Millisecond,
			MaxDelay:   100 * time.Millisecond,
		},
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	var result map[string]interface{}
	err = client.Get("test", &result)
	if err == nil {
		t.Error("Expected error after retry exhaustion")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Errorf("Expected APIError, got %T", err)
	}

	if apiErr.Code != 500 {
		t.Errorf("Expected status code 500, got %d", apiErr.Code)
	}
}

func TestClient_LoggingConfiguration(t *testing.T) {
	var loggedMessages []string
	testLogger := &TestLogger{
		messages: &loggedMessages,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"test": "response"}`))
	}))
	defer server.Close()

	config := &Config{
		BaseURL: server.URL,
		Auth:    &APIKeyAuth{APIKey: "test-key"},
		Logger:  testLogger,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	var result map[string]interface{}
	err = client.Get("test", &result)
	if err != nil {
		t.Errorf("Client.Get() error = %v", err)
	}

	if len(loggedMessages) < 2 {
		t.Errorf("Expected at least 2 log messages, got %d", len(loggedMessages))
	}

	// Check that request and response were logged
	foundRequestLog := false
	foundResponseLog := false
	for _, msg := range loggedMessages {
		if strings.Contains(msg, "n8n API request:") {
			foundRequestLog = true
		}
		if strings.Contains(msg, "n8n API response:") {
			foundResponseLog = true
		}
	}

	if !foundRequestLog {
		t.Error("Expected request log message")
	}
	if !foundResponseLog {
		t.Error("Expected response log message")
	}
}

func TestClient_BackoffCalculation(t *testing.T) {
	config := &Config{
		BaseURL: "https://example.com",
		Auth:    &APIKeyAuth{APIKey: "test-key"},
		RetryConfig: RetryConfig{
			MaxRetries: 3,
			BaseDelay:  100 * time.Millisecond,
			MaxDelay:   1 * time.Second,
		},
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// Test backoff calculation
	delay0 := client.calculateBackoff(0)
	delay1 := client.calculateBackoff(1)
	delay2 := client.calculateBackoff(2)
	delay10 := client.calculateBackoff(10) // Should hit max delay

	if delay0 != 100*time.Millisecond {
		t.Errorf("Expected delay0 = 100ms, got %v", delay0)
	}

	if delay1 != 200*time.Millisecond {
		t.Errorf("Expected delay1 = 200ms, got %v", delay1)
	}

	if delay2 != 400*time.Millisecond {
		t.Errorf("Expected delay2 = 400ms, got %v", delay2)
	}

	if delay10 != 1*time.Second {
		t.Errorf("Expected delay10 = 1s (max delay), got %v", delay10)
	}
}

// TestLogger implements Logger for testing
type TestLogger struct {
	messages *[]string
}

func (l *TestLogger) Logf(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	*l.messages = append(*l.messages, message)
}
