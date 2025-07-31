package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client represents the n8n API client
type Client struct {
	baseURL     *url.URL
	httpClient  *http.Client
	auth        AuthMethod
	logger      Logger
	retryConfig RetryConfig
}

// Logger interface for logging requests and responses
type Logger interface {
	Logf(format string, args ...any)
}

// DefaultLogger implements Logger using the standard log package
type DefaultLogger struct{}

func (l *DefaultLogger) Logf(format string, args ...any) {
	log.Printf(format, args...)
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// Config holds configuration for the n8n client
type Config struct {
	BaseURL            string
	Auth               AuthMethod
	InsecureSkipVerify bool
	Timeout            time.Duration
	Logger             Logger
	RetryConfig        RetryConfig
	CookieFile         string // Path to cookie file for session authentication
}

// AuthMethod interface for different authentication methods
type AuthMethod interface {
	ApplyAuth(*http.Request) error
}

// APIKeyAuth implements API key authentication
type APIKeyAuth struct {
	APIKey string
}

func (a *APIKeyAuth) ApplyAuth(req *http.Request) error {
	req.Header.Set("X-N8N-API-KEY", a.APIKey)
	return nil
}

// BasicAuth implements basic authentication
type BasicAuth struct {
	Email    string
	Password string
}

func (a *BasicAuth) ApplyAuth(req *http.Request) error {
	req.SetBasicAuth(a.Email, a.Password)
	return nil
}

// SessionAuth implements session-based authentication using cookies
type SessionAuth struct {
	CookieJar http.CookieJar
}

func (a *SessionAuth) ApplyAuth(req *http.Request) error {
	// Session authentication is handled via cookies in the HTTP client
	// No additional headers needed as cookies are automatically sent
	return nil
}

// APIError represents an error response from the n8n API
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("n8n API error (code %d): %s - %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("n8n API error (code %d): %s", e.Code, e.Message)
}

// NewClient creates a new n8n API client
func NewClient(config *Config) (*Client, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	if config.Auth == nil {
		return nil, fmt.Errorf("authentication method is required")
	}

	baseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	// Ensure the base URL has a trailing slash and api path
	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}
	if !strings.HasSuffix(baseURL.Path, "api/v1/") {
		baseURL.Path += "api/v1/"
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.InsecureSkipVerify,
		},
	}

	httpClient := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	logger := config.Logger
	if logger == nil {
		logger = &DefaultLogger{}
	}

	retryConfig := config.RetryConfig
	if retryConfig.MaxRetries == 0 {
		retryConfig.MaxRetries = 3
	}
	if retryConfig.BaseDelay == 0 {
		retryConfig.BaseDelay = 100 * time.Millisecond
	}
	if retryConfig.MaxDelay == 0 {
		retryConfig.MaxDelay = 5 * time.Second
	}

	return &Client{
		baseURL:     baseURL,
		httpClient:  httpClient,
		auth:        config.Auth,
		logger:      logger,
		retryConfig: retryConfig,
	}, nil
}

// doRequest performs an HTTP request with authentication, retries, and logging
func (c *Client) doRequest(method, path string, body any, result any) error {
	var jsonData []byte
	var err error

	if body != nil {
		jsonData, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	// Construct full URL
	var fullURL *url.URL
	if strings.Contains(path, "?") {
		// Path contains query parameters, parse it properly
		pathURL, err := url.Parse(path)
		if err != nil {
			return fmt.Errorf("failed to parse path with query: %w", err)
		}
		fullURL = c.baseURL.ResolveReference(pathURL)
	} else {
		// Simple path without query parameters
		fullURL = c.baseURL.ResolveReference(&url.URL{Path: path})
	}

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		var reqBody io.Reader
		if jsonData != nil {
			reqBody = bytes.NewBuffer(jsonData)
		}

		req, err := http.NewRequest(method, fullURL.String(), reqBody)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		// Apply authentication
		if err := c.auth.ApplyAuth(req); err != nil {
			return fmt.Errorf("failed to apply authentication: %w", err)
		}

		// Log request
		c.logger.Logf("n8n API request: %s %s (attempt %d/%d)", method, fullURL.String(), attempt+1, c.retryConfig.MaxRetries+1)
		if len(jsonData) > 0 {
			c.logger.Logf("n8n API request body: %s", string(jsonData))
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if attempt < c.retryConfig.MaxRetries && isRetryableError(err) {
				delay := c.calculateBackoff(attempt)
				c.logger.Logf("n8n API request failed, retrying in %v: %v", delay, err)
				time.Sleep(delay)
				continue
			}
			return fmt.Errorf("request failed: %w", err)
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		// Log response
		c.logger.Logf("n8n API response: %d %s", resp.StatusCode, resp.Status)
		if len(respBody) > 0 {
			c.logger.Logf("n8n API response body: %s", string(respBody))
		}

		// Handle error responses
		if resp.StatusCode >= 400 {
			// Check if this is a retryable HTTP error
			if attempt < c.retryConfig.MaxRetries && isRetryableHTTPStatus(resp.StatusCode) {
				delay := c.calculateBackoff(attempt)
				c.logger.Logf("n8n API request failed with status %d, retrying in %v", resp.StatusCode, delay)
				time.Sleep(delay)
				continue
			}

			var apiErr APIError
			if err := json.Unmarshal(respBody, &apiErr); err != nil {
				// If we can't parse the error response, create a generic error
				return &APIError{
					Code:    resp.StatusCode,
					Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)),
				}
			}
			apiErr.Code = resp.StatusCode
			return &apiErr
		}

		// Parse successful response
		if result != nil && len(respBody) > 0 {
			if err := json.Unmarshal(respBody, result); err != nil {
				return fmt.Errorf("failed to unmarshal response: %w", err)
			}
		}

		return nil
	}

	return fmt.Errorf("max retries exceeded")
}

// calculateBackoff calculates exponential backoff delay
func (c *Client) calculateBackoff(attempt int) time.Duration {
	delay := time.Duration(float64(c.retryConfig.BaseDelay) * math.Pow(2, float64(attempt)))
	return min(delay, c.retryConfig.MaxDelay)
}

// isRetryableError determines if an error is retryable
func isRetryableError(err error) bool {
	// Network errors are generally retryable
	return strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "connection reset")
}

// isRetryableHTTPStatus determines if an HTTP status code is retryable
func isRetryableHTTPStatus(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests ||
		statusCode == http.StatusInternalServerError ||
		statusCode == http.StatusBadGateway ||
		statusCode == http.StatusServiceUnavailable ||
		statusCode == http.StatusGatewayTimeout
}

// Get performs a GET request
func (c *Client) Get(path string, result any) error {
	return c.doRequest("GET", path, nil, result)
}

// Post performs a POST request
func (c *Client) Post(path string, body any, result any) error {
	return c.doRequest("POST", path, body, result)
}

// Put performs a PUT request
func (c *Client) Put(path string, body any, result any) error {
	return c.doRequest("PUT", path, body, result)
}

// Delete performs a DELETE request
func (c *Client) Delete(path string) error {
	return c.doRequest("DELETE", path, nil, nil)
}

// PaginationInfo holds pagination metadata
type PaginationInfo struct {
	Limit      int    `json:"limit,omitempty"`
	Offset     int    `json:"offset,omitempty"`
	Total      int    `json:"total,omitempty"`
	NextCursor string `json:"nextCursor,omitempty"`
	HasNext    bool   `json:"hasNext,omitempty"`
}

// GetWithPagination performs a GET request with pagination support
func (c *Client) GetWithPagination(path string, result any) (*PaginationInfo, error) {
	err := c.doRequest("GET", path, nil, result)
	if err != nil {
		return nil, err
	}

	// Try to extract pagination info from the response
	// This is a best-effort approach since different endpoints might return different formats
	pagination := &PaginationInfo{}

	// If result is a map, try to extract pagination fields
	if resultMap, ok := result.(*map[string]any); ok {
		if nextCursor, exists := (*resultMap)["nextCursor"]; exists {
			if cursorStr, ok := nextCursor.(string); ok {
				pagination.NextCursor = cursorStr
				pagination.HasNext = cursorStr != ""
			}
		}
		if total, exists := (*resultMap)["total"]; exists {
			if totalFloat, ok := total.(float64); ok {
				pagination.Total = int(totalFloat)
			}
		}
	}

	return pagination, nil
}
