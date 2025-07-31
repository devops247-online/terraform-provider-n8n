package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestClient_GetCredentials(t *testing.T) {
	expectedCredentials := []Credential{
		{
			ID:   "1",
			Name: "Test Credential 1",
			Type: "oauth2Api",
		},
		{
			ID:   "2",
			Name: "Test Credential 2",
			Type: "httpBasicAuth",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		if r.URL.Path != "/api/v1/credentials" {
			t.Errorf("Expected path /api/v1/credentials, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		response := CredentialListResponse{
			Data: expectedCredentials,
		}
		_ = json.NewEncoder(w).Encode(response)
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

	result, err := client.GetCredentials(nil)
	if err != nil {
		t.Errorf("GetCredentials() error = %v", err)
	}

	if len(result.Data) != 2 {
		t.Errorf("GetCredentials() returned %d credentials, expected 2", len(result.Data))
	}

	if result.Data[0].ID != "1" {
		t.Errorf("GetCredentials() first credential ID = %s, expected 1", result.Data[0].ID)
	}
}

func TestClient_GetCredentialsWithOptions(t *testing.T) {
	expectedQuery := url.Values{
		"type":  []string{"oauth2Api"},
		"limit": []string{"10"},
	}

	response := CredentialListResponse{
		Data: []Credential{},
	}

	server := TestServer(ListTestHandler(t, expectedQuery, response))
	defer server.Close()

	client := CreateTestClient(t, server.URL)

	options := &CredentialListOptions{
		Type:  "oauth2Api",
		Limit: 10,
	}

	_, err := client.GetCredentials(options)
	if err != nil {
		t.Errorf("GetCredentials() error = %v", err)
	}
}

func TestClient_GetCredential(t *testing.T) {
	expectedCredential := &Credential{
		ID:   "test-id",
		Name: "Test Credential",
		Type: "oauth2Api",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		if r.URL.Path != "/api/v1/credentials/test-id" {
			t.Errorf("Expected path /api/v1/credentials/test-id, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(expectedCredential)
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

	result, err := client.GetCredential("test-id")
	if err != nil {
		t.Errorf("GetCredential() error = %v", err)
	}

	if result.ID != "test-id" {
		t.Errorf("GetCredential() ID = %s, expected test-id", result.ID)
	}

	if result.Name != "Test Credential" {
		t.Errorf("GetCredential() Name = %s, expected Test Credential", result.Name)
	}
}

func TestClient_GetCredentialEmptyID(t *testing.T) {
	config := &Config{
		BaseURL: "https://example.com",
		Auth:    &APIKeyAuth{APIKey: "test-key"},
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = client.GetCredential("")
	if err == nil {
		t.Error("GetCredential() with empty ID should return error")
	}
}

func TestClient_CreateCredential(t *testing.T) {
	credential := &Credential{
		Name: "New Credential",
		Type: "oauth2Api",
		Data: map[string]interface{}{
			"clientId": "test-client-id",
		},
	}

	expectedResult := &Credential{
		ID:   "new-id",
		Name: "New Credential",
		Type: "oauth2Api",
		Data: map[string]interface{}{
			"clientId": "test-client-id",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/api/v1/credentials" {
			t.Errorf("Expected path /api/v1/credentials, got %s", r.URL.Path)
		}

		var receivedCredential Credential
		_ = json.NewDecoder(r.Body).Decode(&receivedCredential)

		if receivedCredential.Name != "New Credential" {
			t.Errorf("Expected credential name 'New Credential', got %s", receivedCredential.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(expectedResult)
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

	result, err := client.CreateCredential(credential)
	if err != nil {
		t.Errorf("CreateCredential() error = %v", err)
	}

	if result.ID != "new-id" {
		t.Errorf("CreateCredential() ID = %s, expected new-id", result.ID)
	}
}

func TestClient_CreateCredentialValidation(t *testing.T) {
	config := &Config{
		BaseURL: "https://example.com",
		Auth:    &APIKeyAuth{APIKey: "test-key"},
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// Test nil credential
	_, err = client.CreateCredential(nil)
	if err == nil {
		t.Error("CreateCredential() with nil credential should return error")
	}

	// Test empty name
	_, err = client.CreateCredential(&Credential{Type: "oauth2Api"})
	if err == nil {
		t.Error("CreateCredential() with empty name should return error")
	}

	// Test empty type
	_, err = client.CreateCredential(&Credential{Name: "Test"})
	if err == nil {
		t.Error("CreateCredential() with empty type should return error")
	}
}

func TestClient_UpdateCredential(t *testing.T) {
	credential := &Credential{
		Name: "Updated Credential",
		Type: "oauth2Api",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}

		if r.URL.Path != "/api/v1/credentials/test-id" {
			t.Errorf("Expected path /api/v1/credentials/test-id, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(credential)
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

	_, err = client.UpdateCredential("test-id", credential)
	if err != nil {
		t.Errorf("UpdateCredential() error = %v", err)
	}
}

func TestClient_DeleteCredential(t *testing.T) {
	server := TestServer(DeleteTestHandler(t, "/api/v1/credentials/test-id"))
	defer server.Close()

	client := CreateTestClient(t, server.URL)

	err := client.DeleteCredential("test-id")
	if err != nil {
		t.Errorf("DeleteCredential() error = %v", err)
	}
}
