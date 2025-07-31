package client

import (
	"net/http"
	"testing"
)

func TestSessionAuth_ApplyAuth(t *testing.T) {
	tests := []struct {
		name       string
		cookieFile string
	}{
		{
			name:       "empty cookie file",
			cookieFile: "",
		},
		{
			name:       "with cookie file",
			cookieFile: "/tmp/cookies.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := &SessionAuth{
				CookieFile: tt.cookieFile,
			}

			req, _ := http.NewRequest("GET", "http://example.com", nil)
			err := auth.ApplyAuth(req)

			// SessionAuth.ApplyAuth should never return error as it just relies on cookies
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestAPIKeyAuth_ApplyAuth(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
	}{
		{
			name:   "standard api key",
			apiKey: "test-api-key-123",
		},
		{
			name:   "empty api key",
			apiKey: "",
		},
		{
			name:   "special characters",
			apiKey: "key-with-special-chars!@#",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := &APIKeyAuth{APIKey: tt.apiKey}
			req, _ := http.NewRequest("GET", "http://example.com", nil)

			err := auth.ApplyAuth(req)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if req.Header.Get("X-N8N-API-KEY") != tt.apiKey {
				t.Errorf("Expected X-N8N-API-KEY header to be %q, got %q", tt.apiKey, req.Header.Get("X-N8N-API-KEY"))
			}
		})
	}
}

func TestBasicAuth_ApplyAuth(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
	}{
		{
			name:     "standard credentials",
			email:    "test@example.com",
			password: "password123",
		},
		{
			name:     "empty email",
			email:    "",
			password: "password123",
		},
		{
			name:     "empty password",
			email:    "test@example.com",
			password: "",
		},
		{
			name:     "special characters in password",
			email:    "test@example.com",
			password: "pass!@#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := &BasicAuth{
				Email:    tt.email,
				Password: tt.password,
			}
			req, _ := http.NewRequest("GET", "http://example.com", nil)

			err := auth.ApplyAuth(req)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			username, password, ok := req.BasicAuth()
			if !ok {
				t.Error("Expected basic auth to be set")
			}

			if username != tt.email {
				t.Errorf("Expected username to be %q, got %q", tt.email, username)
			}

			if password != tt.password {
				t.Errorf("Expected password to be %q, got %q", tt.password, password)
			}
		})
	}
}
