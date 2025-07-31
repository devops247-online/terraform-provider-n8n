package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestServer creates a test HTTP server for client testing
func TestServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// CreateTestClient creates a client configured for testing with the given server URL
func CreateTestClient(t *testing.T, serverURL string) *Client {
	t.Helper()

	config := &Config{
		BaseURL: serverURL,
		Auth:    &APIKeyAuth{APIKey: "test-key"},
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	return client
}

// DeleteTestHandler creates a generic DELETE request handler for testing
func DeleteTestHandler(t *testing.T, expectedPath string) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}

		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// ListTestHandler creates a generic list request handler that validates query parameters
func ListTestHandler(t *testing.T, expectedQuery url.Values, responseData interface{}) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		// Validate expected query parameters
		for key, expectedValues := range expectedQuery {
			if len(expectedValues) > 0 {
				if query.Get(key) != expectedValues[0] {
					t.Errorf("Expected %s=%s, got %s", key, expectedValues[0], query.Get(key))
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(responseData); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}
}
