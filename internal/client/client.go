package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client represents the n8n API client
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	auth       AuthMethod
}

// Config holds configuration for the n8n client
type Config struct {
	BaseURL            string
	Auth               AuthMethod
	InsecureSkipVerify bool
	Timeout            time.Duration
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

	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
		auth:       config.Auth,
	}, nil
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	// Construct full URL
	fullURL := c.baseURL.ResolveReference(&url.URL{Path: path})

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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle error responses
	if resp.StatusCode >= 400 {
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

// Get performs a GET request
func (c *Client) Get(path string, result interface{}) error {
	return c.doRequest("GET", path, nil, result)
}

// Post performs a POST request
func (c *Client) Post(path string, body interface{}, result interface{}) error {
	return c.doRequest("POST", path, body, result)
}

// Put performs a PUT request
func (c *Client) Put(path string, body interface{}, result interface{}) error {
	return c.doRequest("PUT", path, body, result)
}

// Delete performs a DELETE request
func (c *Client) Delete(path string) error {
	return c.doRequest("DELETE", path, nil, nil)
}
