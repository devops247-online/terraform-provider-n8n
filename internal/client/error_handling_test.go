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
	if testing.Short() {
		t.Skip("Skipping network error tests in short mode")
	}
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
	if testing.Short() {
		t.Skip("Skipping timeout handling test in short mode")
	}
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
	if testing.Short() {
		t.Skip("Skipping large response test in short mode")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Write a large but valid JSON response
		var items []string
		for i := 0; i < 1000; i++ {
			items = append(items, fmt.Sprintf(`{"id": "item-%d", "name": "Item %d"}`, i, i))
		}
		response := fmt.Sprintf(`{"data": [%s]}`, strings.Join(items, ","))
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

func TestClient_MalformedErrorResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		desc       string
	}{
		{
			name:       "invalid JSON error response",
			statusCode: http.StatusBadRequest,
			response:   `{invalid json}`,
			desc:       "should create generic error for invalid JSON",
		},
		{
			name:       "empty error response",
			statusCode: http.StatusInternalServerError,
			response:   ``,
			desc:       "should create generic error for empty response",
		},
		{
			name:       "non-JSON error response",
			statusCode: http.StatusServiceUnavailable,
			response:   `Service Temporarily Unavailable`,
			desc:       "should create generic error for plain text",
		},
		{
			name:       "partial JSON error response",
			statusCode: http.StatusBadRequest,
			response:   `{"code": 400}`,
			desc:       "should handle missing message field",
		},
		{
			name:       "error response with extra fields",
			statusCode: http.StatusUnprocessableEntity,
			response:   `{"code": 422, "message": "Validation failed", "details": "Name is required", "field": "name", "extra": "ignored"}`,
			desc:       "should handle extra fields gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := CreateTestClient(t, server.URL)

			var result interface{}
			err := client.doRequest("GET", "/test", nil, &result)

			if err == nil {
				t.Error("Expected error but got none")
				return
			}

			apiErr, ok := err.(*APIError)
			if !ok {
				t.Errorf("Expected APIError type, got %T", err)
				return
			}

			if apiErr.Code != tt.statusCode {
				t.Errorf("Expected status code %d, got %d", tt.statusCode, apiErr.Code)
			}

			// Error should contain meaningful information
			if apiErr.Error() == "" {
				t.Error("Expected non-empty error message")
			}

			t.Logf("Test case: %s - Error: %s", tt.desc, apiErr.Error())
		})
	}
}

func TestClient_RetryableHTTPStatuses(t *testing.T) {
	retryableStatuses := []int{
		http.StatusTooManyRequests,     // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout,      // 504
	}

	for _, statusCode := range retryableStatuses {
		t.Run(fmt.Sprintf("retryable_status_%d", statusCode), func(t *testing.T) {
			if !isRetryableHTTPStatus(statusCode) {
				t.Errorf("Status code %d should be retryable", statusCode)
			}
		})
	}

	nonRetryableStatuses := []int{
		http.StatusBadRequest,          // 400
		http.StatusUnauthorized,        // 401
		http.StatusForbidden,           // 403
		http.StatusNotFound,            // 404
		http.StatusConflict,            // 409
		http.StatusUnprocessableEntity, // 422
	}

	for _, statusCode := range nonRetryableStatuses {
		t.Run(fmt.Sprintf("non_retryable_status_%d", statusCode), func(t *testing.T) {
			if isRetryableHTTPStatus(statusCode) {
				t.Errorf("Status code %d should not be retryable", statusCode)
			}
		})
	}
}

func TestClient_RetryableNetworkErrors(t *testing.T) {
	retryableErrors := []string{
		"connection timeout",
		"read timeout",
		"write timeout",
		"connection refused",
		"connection reset by peer",
		"network is unreachable",
	}

	for _, errorMsg := range retryableErrors {
		t.Run(fmt.Sprintf("retryable_error_%s", strings.ReplaceAll(errorMsg, " ", "_")), func(t *testing.T) {
			testErr := fmt.Errorf("%s", errorMsg)
			if !isRetryableError(testErr) {
				t.Errorf("Error %q should be retryable", errorMsg)
			}
		})
	}

	nonRetryableErrors := []string{
		"invalid URL",
		"certificate verify failed",
		"unsupported protocol scheme",
	}

	for _, errorMsg := range nonRetryableErrors {
		t.Run(fmt.Sprintf("non_retryable_error_%s", strings.ReplaceAll(errorMsg, " ", "_")), func(t *testing.T) {
			testErr := fmt.Errorf("%s", errorMsg)
			if isRetryableError(testErr) {
				t.Errorf("Error %q should not be retryable", errorMsg)
			}
		})
	}
}

func TestClient_ExponentialBackoff(t *testing.T) {
	config := &Config{
		BaseURL: "https://example.com",
		Auth:    &APIKeyAuth{APIKey: "test-key"},
		RetryConfig: RetryConfig{
			MaxRetries: 5,
			BaseDelay:  50 * time.Millisecond,
			MaxDelay:   2 * time.Second,
		},
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	tests := []struct {
		attempt      int
		expectedMin  time.Duration
		expectedMax  time.Duration
		shouldHitMax bool
	}{
		{0, 50 * time.Millisecond, 50 * time.Millisecond, false},
		{1, 100 * time.Millisecond, 100 * time.Millisecond, false},
		{2, 200 * time.Millisecond, 200 * time.Millisecond, false},
		{3, 400 * time.Millisecond, 400 * time.Millisecond, false},
		{4, 800 * time.Millisecond, 800 * time.Millisecond, false},
		{5, 1600 * time.Millisecond, 2 * time.Second, true}, // Should hit max
		{10, 2 * time.Second, 2 * time.Second, true},        // Should hit max
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("attempt_%d", tt.attempt), func(t *testing.T) {
			delay := client.calculateBackoff(tt.attempt)

			if tt.shouldHitMax {
				if delay != config.RetryConfig.MaxDelay {
					t.Errorf("Expected max delay %v for attempt %d, got %v", config.RetryConfig.MaxDelay, tt.attempt, delay)
				}
			} else {
				if delay < tt.expectedMin || delay > tt.expectedMax {
					t.Errorf("Expected delay between %v and %v for attempt %d, got %v", tt.expectedMin, tt.expectedMax, tt.attempt, delay)
				}
			}
		})
	}
}

func TestClient_RetryExhaustionDetailed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping detailed retry test in short mode")
	}
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"code": 500, "message": "Server Error"}`))
	}))
	defer server.Close()

	config := &Config{
		BaseURL: server.URL,
		Auth:    &APIKeyAuth{APIKey: "test-key"},
		RetryConfig: RetryConfig{
			MaxRetries: 3,
			BaseDelay:  1 * time.Millisecond, // Very fast for testing
			MaxDelay:   10 * time.Millisecond,
		},
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	var result interface{}
	err = client.doRequest("GET", "/test", nil, &result)

	if err == nil {
		t.Error("Expected error after retry exhaustion")
	}

	expectedAttempts := config.RetryConfig.MaxRetries + 1 // Initial + retries
	if attemptCount != expectedAttempts {
		t.Errorf("Expected %d attempts, got %d", expectedAttempts, attemptCount)
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Errorf("Expected APIError, got %T", err)
	} else if apiErr.Code != 500 {
		t.Errorf("Expected status code 500, got %d", apiErr.Code)
	}
}

func TestClient_PartialRetrySuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping partial retry success test in short mode")
	}
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			// First two attempts fail
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"code": 503, "message": "Service Unavailable"}`))
			return
		}
		// Third attempt succeeds
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true, "attempt": 3}`))
	}))
	defer server.Close()

	config := &Config{
		BaseURL: server.URL,
		Auth:    &APIKeyAuth{APIKey: "test-key"},
		RetryConfig: RetryConfig{
			MaxRetries: 3,
			BaseDelay:  1 * time.Millisecond,
			MaxDelay:   10 * time.Millisecond,
		},
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	var result map[string]interface{}
	err = client.doRequest("GET", "/test", nil, &result)

	if err != nil {
		t.Errorf("Unexpected error after successful retry: %v", err)
	}

	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}

	if result["success"] != true {
		t.Errorf("Expected success=true, got %v", result["success"])
	}
}

func TestClient_RequestBodyMarshaling(t *testing.T) {
	tests := []struct {
		name    string
		body    interface{}
		wantErr bool
		desc    string
	}{
		{
			name:    "valid struct",
			body:    map[string]string{"key": "value"},
			wantErr: false,
			desc:    "should marshal valid struct",
		},
		{
			name:    "nil body",
			body:    nil,
			wantErr: false,
			desc:    "should handle nil body",
		},
		{
			name:    "unmarshalable type",
			body:    make(chan int),
			wantErr: true,
			desc:    "should fail for unmarshalable types",
		},
		{
			name:    "function type",
			body:    func() {},
			wantErr: true,
			desc:    "should fail for function types",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	client := CreateTestClient(t, server.URL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result interface{}
			err := client.doRequest("POST", "/test", tt.body, &result)

			if tt.wantErr && err == nil {
				t.Error("Expected marshaling error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			t.Logf("Test case: %s", tt.desc)
		})
	}
}

func TestClient_ResponseBodyReading(t *testing.T) {
	tests := []struct {
		name         string
		response     string
		contentType  string
		expectResult bool
		desc         string
	}{
		{
			name:         "valid JSON response",
			response:     `{"data": "test"}`,
			contentType:  "application/json",
			expectResult: true,
			desc:         "should parse valid JSON",
		},
		{
			name:         "empty response with result pointer",
			response:     ``,
			contentType:  "application/json",
			expectResult: false,
			desc:         "should handle empty response gracefully",
		},
		{
			name:         "whitespace only response",
			response:     `   `,
			contentType:  "application/json",
			expectResult: false,
			desc:         "should handle whitespace-only response",
		},
		{
			name:         "large JSON response",
			response:     strings.Repeat(`{"key": "value"},`, 1000),
			contentType:  "application/json",
			expectResult: false, // Will be invalid JSON due to trailing comma
			desc:         "should handle large responses",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := CreateTestClient(t, server.URL)

			var result map[string]interface{}
			err := client.doRequest("GET", "/test", nil, &result)

			if !tt.expectResult && err == nil && len(tt.response) > 0 && strings.TrimSpace(tt.response) != "" {
				// Should have error for invalid JSON (except empty responses)
				t.Error("Expected JSON parsing error but got none")
			}

			t.Logf("Test case: %s - Error: %v", tt.desc, err)
		})
	}
}

func TestClient_ConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent requests test in short mode")
	}
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"request": %d}`, requestCount)))
	}))
	defer server.Close()

	client := CreateTestClient(t, server.URL)

	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			var result map[string]interface{}
			err := client.doRequest("GET", fmt.Sprintf("/test-%d", id), nil, &result)
			results <- err
		}(i)
	}

	// Collect results
	errorCount := 0
	for i := 0; i < numGoroutines; i++ {
		if err := <-results; err != nil {
			errorCount++
			t.Errorf("Concurrent request %d failed: %v", i, err)
		}
	}

	if errorCount > 0 {
		t.Errorf("%d out of %d concurrent requests failed", errorCount, numGoroutines)
	}

	// All requests should have been processed
	if requestCount != numGoroutines {
		t.Errorf("Expected %d requests, got %d", numGoroutines, requestCount)
	}
}

func TestClient_PathResolution(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		expectedPath string
		wantErr      bool
	}{
		{
			name:         "simple path",
			path:         "workflows",
			expectedPath: "/api/v1/workflows",
			wantErr:      false,
		},
		{
			name:         "path with query parameters",
			path:         "workflows?active=true&limit=10",
			expectedPath: "/api/v1/workflows",
			wantErr:      false,
		},
		{
			name:         "path with ID",
			path:         "workflows/123",
			expectedPath: "/api/v1/workflows/123",
			wantErr:      false,
		},
		{
			name:         "path with special characters",
			path:         "workflows/test%20workflow",
			expectedPath: "/api/v1/workflows/test%20workflow",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasPrefix(r.URL.Path, tt.expectedPath) {
					t.Errorf("Expected path to start with %q, got %q", tt.expectedPath, r.URL.Path)
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			}))
			defer server.Close()

			client := CreateTestClient(t, server.URL)

			var result interface{}
			err := client.doRequest("GET", tt.path, nil, &result)

			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
