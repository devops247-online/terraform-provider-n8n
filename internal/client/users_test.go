package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_GetUsers(t *testing.T) {
	expectedUsers := []User{
		{
			ID:        "1",
			Email:     "user1@example.com",
			FirstName: "User",
			LastName:  "One",
			Role:      "member",
		},
		{
			ID:        "2",
			Email:     "user2@example.com",
			FirstName: "User",
			LastName:  "Two",
			Role:      "admin",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		if r.URL.Path != "/api/v1/users" {
			t.Errorf("Expected path /api/v1/users, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		response := UserListResponse{
			Data: expectedUsers,
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

	result, err := client.GetUsers(nil)
	if err != nil {
		t.Errorf("GetUsers() error = %v", err)
	}

	if len(result.Data) != 2 {
		t.Errorf("GetUsers() returned %d users, expected 2", len(result.Data))
	}

	if result.Data[0].ID != "1" {
		t.Errorf("GetUsers() first user ID = %s, expected 1", result.Data[0].ID)
	}
}

func TestClient_GetUsersWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		if query.Get("role") != "admin" {
			t.Errorf("Expected role=admin, got %s", query.Get("role"))
		}

		if query.Get("limit") != "5" {
			t.Errorf("Expected limit=5, got %s", query.Get("limit"))
		}

		w.Header().Set("Content-Type", "application/json")
		response := UserListResponse{
			Data: []User{},
		}
		_ = json.NewEncoder(w).Encode(response)
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

	options := &UserListOptions{
		Role:  "admin",
		Limit: 5,
	}

	_, err = client.GetUsers(options)
	if err != nil {
		t.Errorf("GetUsers() error = %v", err)
	}
}

func TestClient_GetUser(t *testing.T) {
	expectedUser := &User{
		ID:        "test-id",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Role:      "member",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		if r.URL.Path != "/api/v1/users/test-id" {
			t.Errorf("Expected path /api/v1/users/test-id, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(expectedUser)
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

	result, err := client.GetUser("test-id")
	if err != nil {
		t.Errorf("GetUser() error = %v", err)
	}

	if result.ID != "test-id" {
		t.Errorf("GetUser() ID = %s, expected test-id", result.ID)
	}

	if result.Email != "test@example.com" {
		t.Errorf("GetUser() Email = %s, expected test@example.com", result.Email)
	}
}

func TestClient_GetUserEmptyID(t *testing.T) {
	config := &Config{
		BaseURL: "https://example.com",
		Auth:    &APIKeyAuth{APIKey: "test-key"},
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = client.GetUser("")
	if err == nil {
		t.Error("GetUser() with empty ID should return error")
	}
}

func TestClient_CreateUser(t *testing.T) {
	userReq := &CreateUserRequest{
		Email:     "newuser@example.com",
		FirstName: "New",
		LastName:  "User",
		Role:      "member",
		Password:  "password123",
	}

	expectedUser := User{
		ID:        "new-id",
		Email:     "newuser@example.com",
		FirstName: "New",
		LastName:  "User",
		Role:      "member",
	}

	// n8n API returns array of {user: User, error: string} objects
	type CreateUserResponse struct {
		User  User   `json:"user"`
		Error string `json:"error"`
	}

	expectedResult := []CreateUserResponse{
		{
			User:  expectedUser,
			Error: "",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/api/v1/users" {
			t.Errorf("Expected path /api/v1/users, got %s", r.URL.Path)
		}

		var receivedUserReqArray []*CreateUserRequest
		_ = json.NewDecoder(r.Body).Decode(&receivedUserReqArray)

		if len(receivedUserReqArray) == 0 {
			t.Errorf("Expected array of users, got empty array")
		}

		if len(receivedUserReqArray) > 0 && receivedUserReqArray[0].Email != "newuser@example.com" {
			t.Errorf("Expected user email 'newuser@example.com', got %s", receivedUserReqArray[0].Email)
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

	result, err := client.CreateUser(userReq)
	if err != nil {
		t.Errorf("CreateUser() error = %v", err)
	}

	if result.ID != "new-id" {
		t.Errorf("CreateUser() ID = %s, expected new-id", result.ID)
	}
}

func TestClient_CreateUserValidation(t *testing.T) {
	config := &Config{
		BaseURL: "https://example.com",
		Auth:    &APIKeyAuth{APIKey: "test-key"},
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// Test nil user request
	_, err = client.CreateUser(nil)
	if err == nil {
		t.Error("CreateUser() with nil user request should return error")
	}

	// Test empty email
	_, err = client.CreateUser(&CreateUserRequest{FirstName: "Test"})
	if err == nil {
		t.Error("CreateUser() with empty email should return error")
	}
}

func TestClient_UpdateUser(t *testing.T) {
	user := &User{
		Email:     "updated@example.com",
		FirstName: "Updated",
		LastName:  "User",
		Role:      "admin",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}

		if r.URL.Path != "/api/v1/users/test-id" {
			t.Errorf("Expected path /api/v1/users/test-id, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(user)
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

	_, err = client.UpdateUser("test-id", user)
	if err != nil {
		t.Errorf("UpdateUser() error = %v", err)
	}
}

func TestClient_DeleteUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}

		if r.URL.Path != "/api/v1/users/test-id" {
			t.Errorf("Expected path /api/v1/users/test-id, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
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

	err = client.DeleteUser("test-id")
	if err != nil {
		t.Errorf("DeleteUser() error = %v", err)
	}
}
