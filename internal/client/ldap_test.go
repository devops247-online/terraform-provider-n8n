package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetLDAPConfig(t *testing.T) {
	// Mock response
	mockConfig := LDAPConfig{
		ServerURL:              "ldap://ldap.example.com:389",
		BindDN:                 "cn=admin,dc=example,dc=com",
		SearchBase:             "ou=users,dc=example,dc=com",
		SearchFilter:           "(uid={{username}})",
		UserIDAttribute:        "uid",
		UserEmailAttribute:     "mail",
		UserFirstNameAttribute: "givenName",
		UserLastNameAttribute:  "sn",
		TLSEnabled:             false,
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/ldap/config" {
			t.Errorf("Expected path /api/v1/ldap/config, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockConfig)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    &APIKeyAuth{APIKey: "test-key"},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test GetLDAPConfig
	result, err := client.GetLDAPConfig()
	if err != nil {
		t.Fatalf("GetLDAPConfig failed: %v", err)
	}

	if result.ServerURL != "ldap://ldap.example.com:389" {
		t.Errorf("Expected server URL 'ldap://ldap.example.com:389', got '%s'", result.ServerURL)
	}
	if result.BindDN != "cn=admin,dc=example,dc=com" {
		t.Errorf("Expected bind DN 'cn=admin,dc=example,dc=com', got '%s'", result.BindDN)
	}
	if result.UserIDAttribute != "uid" {
		t.Errorf("Expected user ID attribute 'uid', got '%s'", result.UserIDAttribute)
	}
}

func TestClient_UpdateLDAPConfig(t *testing.T) {
	// Mock request/response
	inputConfig := &LDAPConfig{
		ServerURL:              "ldaps://ldap.example.com:636",
		BindDN:                 "cn=admin,dc=example,dc=com",
		BindPassword:           "secret123",
		SearchBase:             "ou=users,dc=example,dc=com",
		SearchFilter:           "(uid={{username}})",
		UserIDAttribute:        "uid",
		UserEmailAttribute:     "mail",
		UserFirstNameAttribute: "givenName",
		UserLastNameAttribute:  "sn",
		TLSEnabled:             true,
	}

	mockResponse := LDAPConfig{
		ServerURL:              "ldaps://ldap.example.com:636",
		BindDN:                 "cn=admin,dc=example,dc=com",
		SearchBase:             "ou=users,dc=example,dc=com",
		SearchFilter:           "(uid={{username}})",
		UserIDAttribute:        "uid",
		UserEmailAttribute:     "mail",
		UserFirstNameAttribute: "givenName",
		UserLastNameAttribute:  "sn",
		TLSEnabled:             true,
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/ldap/config" {
			t.Errorf("Expected path /api/v1/ldap/config, got %s", r.URL.Path)
		}

		// Verify request body
		var requestBody LDAPConfig
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		if requestBody.ServerURL != "ldaps://ldap.example.com:636" {
			t.Errorf("Expected server URL 'ldaps://ldap.example.com:636', got '%s'", requestBody.ServerURL)
		}
		if requestBody.BindPassword != "secret123" {
			t.Errorf("Expected bind password 'secret123', got '%s'", requestBody.BindPassword)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    &APIKeyAuth{APIKey: "test-key"},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test UpdateLDAPConfig
	result, err := client.UpdateLDAPConfig(inputConfig)
	if err != nil {
		t.Fatalf("UpdateLDAPConfig failed: %v", err)
	}

	if result.ServerURL != "ldaps://ldap.example.com:636" {
		t.Errorf("Expected server URL 'ldaps://ldap.example.com:636', got '%s'", result.ServerURL)
	}
	if result.TLSEnabled != true {
		t.Errorf("Expected TLS enabled to be true, got %v", result.TLSEnabled)
	}
}

func TestClient_TestLDAPConnection(t *testing.T) {
	// Mock response
	mockResult := LDAPTestResult{
		Success: true,
		Message: "LDAP connection successful",
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/ldap/test" {
			t.Errorf("Expected path /api/v1/ldap/test, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockResult)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    &APIKeyAuth{APIKey: "test-key"},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test TestLDAPConnection
	result, err := client.TestLDAPConnection()
	if err != nil {
		t.Fatalf("TestLDAPConnection failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success to be true, got %v", result.Success)
	}
	if result.Message != "LDAP connection successful" {
		t.Errorf("Expected message 'LDAP connection successful', got '%s'", result.Message)
	}
}

func TestClient_TestLDAPConnectionWithConfig(t *testing.T) {
	// Mock request/response
	inputConfig := &LDAPConfig{
		ServerURL:    "ldap://test.example.com:389",
		BindDN:       "cn=test,dc=example,dc=com",
		BindPassword: "testpass",
	}

	mockResult := LDAPTestResult{
		Success: false,
		Message: "Connection timeout",
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/ldap/test" {
			t.Errorf("Expected path /api/v1/ldap/test, got %s", r.URL.Path)
		}

		// Verify request body
		var requestBody LDAPConfig
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		if requestBody.ServerURL != "ldap://test.example.com:389" {
			t.Errorf("Expected server URL 'ldap://test.example.com:389', got '%s'", requestBody.ServerURL)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockResult)
	}))
	defer server.Close()

	// Create client
	client, err := NewClient(&Config{
		BaseURL: server.URL,
		Auth:    &APIKeyAuth{APIKey: "test-key"},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test TestLDAPConnectionWithConfig
	result, err := client.TestLDAPConnectionWithConfig(inputConfig)
	if err != nil {
		t.Fatalf("TestLDAPConnectionWithConfig failed: %v", err)
	}

	if result.Success {
		t.Errorf("Expected success to be false, got %v", result.Success)
	}
	if result.Message != "Connection timeout" {
		t.Errorf("Expected message 'Connection timeout', got '%s'", result.Message)
	}
}

func TestClient_UpdateLDAPConfig_ValidationErrors(t *testing.T) {
	// Create client
	client, err := NewClient(&Config{
		BaseURL: "http://example.com",
		Auth:    &APIKeyAuth{APIKey: "test-key"},
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test nil config
	_, err = client.UpdateLDAPConfig(nil)
	if err == nil {
		t.Error("Expected error for nil config, got nil")
	}

	// Test missing server URL
	_, err = client.UpdateLDAPConfig(&LDAPConfig{
		BindDN:       "cn=admin,dc=example,dc=com",
		BindPassword: "secret",
	})
	if err == nil {
		t.Error("Expected error for missing server URL, got nil")
	}

	// Test missing bind DN
	_, err = client.UpdateLDAPConfig(&LDAPConfig{
		ServerURL:    "ldap://ldap.example.com:389",
		BindPassword: "secret",
	})
	if err == nil {
		t.Error("Expected error for missing bind DN, got nil")
	}

	// Test missing bind password
	_, err = client.UpdateLDAPConfig(&LDAPConfig{
		ServerURL: "ldap://ldap.example.com:389",
		BindDN:    "cn=admin,dc=example,dc=com",
	})
	if err == nil {
		t.Error("Expected error for missing bind password, got nil")
	}
}
