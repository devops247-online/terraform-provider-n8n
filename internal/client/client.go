package client

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
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
	CookieJar  http.CookieJar
	CookieFile string
}

func (a *SessionAuth) ApplyAuth(req *http.Request) error {
	// Session authentication is handled via cookies in the HTTP client
	// No additional headers needed as cookies are automatically sent
	return nil
}

// validateCookieFilePath validates that the cookie file path is safe to open
func validateCookieFilePath(cookieFile string) error {
	if cookieFile == "" {
		return fmt.Errorf("cookie file path cannot be empty")
	}

	// Clean the path to resolve any .. or . components
	cleanPath := filepath.Clean(cookieFile)

	// Check for path traversal attempts
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("cookie file path contains invalid path traversal: %s", cookieFile)
	}

	if err := validateAbsolutePath(cleanPath, cookieFile); err != nil {
		return err
	}

	return validateFileExtension(cleanPath)
}

// validateAbsolutePath checks if absolute paths are in allowed directories
func validateAbsolutePath(cleanPath, originalPath string) error {
	if !filepath.IsAbs(cleanPath) {
		return nil
	}

	allowedDirs := getAllowedDirectories()
	for _, allowedDir := range allowedDirs {
		if strings.HasPrefix(cleanPath, filepath.Clean(allowedDir)) {
			return nil
		}
	}

	return fmt.Errorf("cookie file path outside allowed directories: %s", originalPath)
}

// getAllowedDirectories returns list of safe directories for cookie files
func getAllowedDirectories() []string {
	allowedDirs := []string{"/tmp", "/var/tmp", os.TempDir()}

	if homeDir, err := os.UserHomeDir(); err == nil {
		allowedDirs = append(allowedDirs, homeDir)
	}
	if cwd, err := os.Getwd(); err == nil {
		allowedDirs = append(allowedDirs, cwd)
	}

	return allowedDirs
}

// validateFileExtension checks if the file extension is allowed
func validateFileExtension(cleanPath string) error {
	ext := filepath.Ext(cleanPath)
	allowedExts := []string{".txt", ".cookies", ".cookie", ""}

	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			return nil
		}
	}

	return fmt.Errorf("cookie file has invalid extension: %s (allowed: .txt, .cookies, .cookie, or no extension)", ext)
}

// LoadCookiesFromFile loads cookies from a Netscape format cookie file
func LoadCookiesFromFile(cookieFile string, targetURL *url.URL) (http.CookieJar, error) {
	// Validate the cookie file path for security
	if err := validateCookieFilePath(cookieFile); err != nil {
		return nil, fmt.Errorf("invalid cookie file path: %w", err)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	// Use the cleaned path
	cleanPath := filepath.Clean(cookieFile)
	file, err := os.Open(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open cookie file: %w", err)
	}
	defer file.Close()

	var cookies []*http.Cookie
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse Netscape cookie format: domain \t flag \t path \t secure \t expiration \t name \t value
		parts := strings.Split(line, "\t")
		if len(parts) < 7 {
			continue
		}

		domain := parts[0]
		path := parts[2]
		secure := parts[3] == "TRUE"
		expiration := parts[4]
		name := parts[5]
		value := parts[6]

		// Convert expiration timestamp
		var expires time.Time
		if expiration != "0" {
			if timestamp, err := strconv.ParseInt(expiration, 10, 64); err == nil {
				expires = time.Unix(timestamp, 0)
			}
		}

		// Skip expired cookies
		if !expires.IsZero() && expires.Before(time.Now()) {
			continue
		}

		cookie := &http.Cookie{
			Name:     name,
			Value:    value,
			Domain:   strings.TrimPrefix(domain, "."),
			Path:     path,
			Expires:  expires,
			Secure:   secure,
			HttpOnly: strings.HasPrefix(domain, "#HttpOnly_"),
		}

		cookies = append(cookies, cookie)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading cookie file: %w", err)
	}

	// Set cookies in jar
	if len(cookies) > 0 {
		jar.SetCookies(targetURL, cookies)
	}

	return jar, nil
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

	// Configure TLS settings
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			// InsecureSkipVerify should only be used for development/testing environments
			// with self-signed certificates. In production, proper certificate validation
			// should be used to prevent man-in-the-middle attacks.
			InsecureSkipVerify: config.InsecureSkipVerify, // #nosec G402 - Configurable for development environments
		},
	}

	httpClient := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	// If using session authentication, set up cookie jar
	if sessionAuth, ok := config.Auth.(*SessionAuth); ok && sessionAuth.CookieFile != "" {
		cookieJar, err := LoadCookiesFromFile(sessionAuth.CookieFile, baseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to load cookies from file: %w", err)
		}
		httpClient.Jar = cookieJar
		sessionAuth.CookieJar = cookieJar
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

		// Ensure response body is properly closed
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				c.logger.Logf("Warning: failed to close response body: %v", closeErr)
			}
		}()

		respBody, err := io.ReadAll(resp.Body)
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
		strings.Contains(err.Error(), "connection reset") ||
		strings.Contains(err.Error(), "network is unreachable")
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
