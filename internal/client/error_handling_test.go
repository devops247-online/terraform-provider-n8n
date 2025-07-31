package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClient_HTTPErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		wantError  bool
		errorType  string
	}{
		{
			name:       "401 Unauthorized",
			statusCode: http.StatusUnauthorized,
			response:   `{"code": 401, "message": "Unauthorized"}`,
			wantError:  true,
			errorType:  "unauthorized",
		},
		{
			name:       "403 Forbidden",
			statusCode: http.StatusForbidden,
			response:   `{"code": 403, "message": "Forbidden"}`,
			wantError:  true,
			errorType:  "forbidden",
		},
		{
			name:       "404 Not Found",
			statusCode: http.StatusNotFound,
			response:   `{"code": 404, "message": "Not Found"}`,
			wantError:  true,
			errorType:  "not found",
		},
		{
			name:       "422 Unprocessable Entity",
			statusCode: http.StatusUnprocessableEntity,
			response:   `{"code": 422, "message": "Validation failed"}`,
			wantError:  true,
			errorType:  "validation",
		},
		{
			name:       "500 Internal Server Error",
			statusCode: http.StatusInternalServerError,
			response:   `{"code": 500, "message": "Internal Server Error"}`,
			wantError:  true,
			errorType:  "server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := CreateTestClient(t, server.URL)

			var result interface{}
			err := client.doRequest("GET", "/test", nil, &result)

			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tt.wantError && err != nil {
				errorMsg := strings.ToLower(err.Error())
				if !strings.Contains(errorMsg, tt.errorType) && !strings.Contains(errorMsg, fmt.Sprintf("%d", tt.statusCode)) {
					t.Errorf("Expected error to contain %q or status code %d, got %q", tt.errorType, tt.statusCode, err.Error())
				}
			}
		})
	}
}

func TestClient_NetworkErrors(t *testing.T) {
	tests := []struct {
		name      string
		serverURL string
		wantError bool
	}{
		{
			name:      "invalid URL",
			serverURL: "invalid-url",
			wantError: true,
		},
		{
			name:      "connection refused",
			serverURL: "http://localhost:9999",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				BaseURL: tt.serverURL,
				Auth:    &APIKeyAuth{APIKey: "test-key"},
			}

			client, err := NewClient(config)
			if tt.name == "invalid URL" && err != nil {
				// Invalid URL should fail at client creation
				return
			}
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			var result interface{}
			err = client.doRequest("GET", "/test", nil, &result)

			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestClient_TimeoutHandling(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
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

	// Set a very short timeout
	client.httpClient.Timeout = 10 * time.Millisecond

	var result interface{}
	err = client.doRequest("GET", "/test", nil, &result)

	if err == nil {
		t.Error("Expected timeout error but got none")
	}

	if !strings.Contains(strings.ToLower(err.Error()), "timeout") && !strings.Contains(strings.ToLower(err.Error()), "deadline") {
		t.Errorf("Expected timeout error, got %v", err)
	}
}

func TestClient_InvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	client := CreateTestClient(t, server.URL)

	var result interface{}
	err := client.doRequest("GET", "/test", nil, &result)

	if err == nil {
		t.Error("Expected JSON parse error but got none")
	}

	if !strings.Contains(strings.ToLower(err.Error()), "json") && !strings.Contains(strings.ToLower(err.Error()), "unmarshal") {
		t.Errorf("Expected JSON error, got %v", err)
	}
}

func TestClient_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Empty response body
	}))
	defer server.Close()

	client := CreateTestClient(t, server.URL)

	var result interface{}
	err := client.doRequest("GET", "/test", nil, &result)

	// Empty response should not cause an error when result is nil
	if err != nil {
		t.Errorf("Unexpected error for empty response: %v", err)
	}
}

func TestClient_LargeResponse(t *testing.T) {
	// Create a large JSON response
	largeData := make([]map[string]string, 1000)
	for i := 0; i < 1000; i++ {
		largeData[i] = map[string]string{
			"id":   fmt.Sprintf("item-%d", i),
			"name": fmt.Sprintf("Item %d with some long description", i),
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Write a large response
		response := fmt.Sprintf(`{"data": %s}`, strings.Repeat(`{"id": "test", "name": "test"}`, 10000))
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client := CreateTestClient(t, server.URL)

	var result interface{}
	err := client.doRequest("GET", "/test", nil, &result)

	if err != nil {
		t.Errorf("Unexpected error for large response: %v", err)
	}
}

func TestClient_RequestWithContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that request context exists
		if r.Context() == nil {
			t.Error("Expected request to have context")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	client := CreateTestClient(t, server.URL)

	var result interface{}
	err := client.doRequest("GET", "/test", nil, &result)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
