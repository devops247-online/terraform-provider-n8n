package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_GetProjects(t *testing.T) {
	// Mock response
	mockResponse := ProjectListResponse{
		Data: []Project{
			{
				ID:          "proj-1",
				Name:        "Test Project",
				Description: "A test project",
				OwnerID:     "user-1",
				MemberCount: 2,
			},
		},
		NextCursor: "",
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/projects" {
			t.Errorf("Expected path /api/v1/projects, got %s", r.URL.Path)
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

	// Test GetProjects
	result, err := client.GetProjects(nil)
	if err != nil {
		t.Fatalf("GetProjects failed: %v", err)
	}

	if len(result.Data) != 1 {
		t.Errorf("Expected 1 project, got %d", len(result.Data))
	}

	project := result.Data[0]
	if project.ID != "proj-1" {
		t.Errorf("Expected project ID 'proj-1', got '%s'", project.ID)
	}
	if project.Name != "Test Project" {
		t.Errorf("Expected project name 'Test Project', got '%s'", project.Name)
	}
}

func TestClient_GetProject(t *testing.T) {
	// Mock response
	mockProject := Project{
		ID:          "proj-1",
		Name:        "Test Project",
		Description: "A test project",
		OwnerID:     "user-1",
		MemberCount: 2,
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/projects/proj-1" {
			t.Errorf("Expected path /api/v1/projects/proj-1, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockProject)
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

	// Test GetProject
	result, err := client.GetProject("proj-1")
	if err != nil {
		t.Fatalf("GetProject failed: %v", err)
	}

	if result.ID != "proj-1" {
		t.Errorf("Expected project ID 'proj-1', got '%s'", result.ID)
	}
	if result.Name != "Test Project" {
		t.Errorf("Expected project name 'Test Project', got '%s'", result.Name)
	}
}

func TestClient_CreateProject(t *testing.T) {
	// Mock request/response
	inputProject := &Project{
		Name:        "New Project",
		Description: "A new project",
	}

	mockResponse := Project{
		ID:          "proj-2",
		Name:        "New Project",
		Description: "A new project",
		OwnerID:     "user-1",
		MemberCount: 1,
		CreatedAt:   &time.Time{},
		UpdatedAt:   &time.Time{},
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/projects" {
			t.Errorf("Expected path /api/v1/projects, got %s", r.URL.Path)
		}

		// Verify request body
		var requestBody Project
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		if requestBody.Name != "New Project" {
			t.Errorf("Expected project name 'New Project', got '%s'", requestBody.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
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

	// Test CreateProject
	result, err := client.CreateProject(inputProject)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	if result.ID != "proj-2" {
		t.Errorf("Expected project ID 'proj-2', got '%s'", result.ID)
	}
	if result.Name != "New Project" {
		t.Errorf("Expected project name 'New Project', got '%s'", result.Name)
	}
}

func TestClient_UpdateProject(t *testing.T) {
	// Mock request/response
	inputProject := &Project{
		Name:        "Updated Project",
		Description: "An updated project",
	}

	mockResponse := Project{
		ID:          "proj-1",
		Name:        "Updated Project",
		Description: "An updated project",
		OwnerID:     "user-1",
		MemberCount: 2,
		UpdatedAt:   &time.Time{},
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/projects/proj-1" {
			t.Errorf("Expected path /api/v1/projects/proj-1, got %s", r.URL.Path)
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

	// Test UpdateProject
	result, err := client.UpdateProject("proj-1", inputProject)
	if err != nil {
		t.Fatalf("UpdateProject failed: %v", err)
	}

	if result.Name != "Updated Project" {
		t.Errorf("Expected project name 'Updated Project', got '%s'", result.Name)
	}
}

func TestClient_DeleteProject(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/projects/proj-1" {
			t.Errorf("Expected path /api/v1/projects/proj-1, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
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

	// Test DeleteProject
	err = client.DeleteProject("proj-1")
	if err != nil {
		t.Fatalf("DeleteProject failed: %v", err)
	}
}

func TestClient_GetProjectUsers(t *testing.T) {
	// Mock response
	mockUsers := []ProjectUser{
		{
			ID:        "pu-1",
			ProjectID: "proj-1",
			UserID:    "user-1",
			Role:      "admin",
			AddedAt:   &time.Time{},
		},
		{
			ID:        "pu-2",
			ProjectID: "proj-1",
			UserID:    "user-2",
			Role:      "editor",
			AddedAt:   &time.Time{},
		},
	}

	mockResponse := struct {
		Data []ProjectUser `json:"data"`
	}{
		Data: mockUsers,
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/projects/proj-1/users" {
			t.Errorf("Expected path /api/v1/projects/proj-1/users, got %s", r.URL.Path)
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

	// Test GetProjectUsers
	result, err := client.GetProjectUsers("proj-1")
	if err != nil {
		t.Fatalf("GetProjectUsers failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 project users, got %d", len(result))
	}

	if result[0].UserID != "user-1" {
		t.Errorf("Expected user ID 'user-1', got '%s'", result[0].UserID)
	}
	if result[0].Role != "admin" {
		t.Errorf("Expected role 'admin', got '%s'", result[0].Role)
	}
}

func TestClient_AddUserToProject(t *testing.T) {
	// Mock request/response
	inputProjectUser := &ProjectUser{
		ProjectID: "proj-1",
		UserID:    "user-3",
		Role:      "viewer",
	}

	mockResponse := ProjectUser{
		ID:        "pu-3",
		ProjectID: "proj-1",
		UserID:    "user-3",
		Role:      "viewer",
		AddedAt:   &time.Time{},
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/projects/proj-1/users" {
			t.Errorf("Expected path /api/v1/projects/proj-1/users, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
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

	// Test AddUserToProject
	result, err := client.AddUserToProject(inputProjectUser)
	if err != nil {
		t.Fatalf("AddUserToProject failed: %v", err)
	}

	if result.UserID != "user-3" {
		t.Errorf("Expected user ID 'user-3', got '%s'", result.UserID)
	}
	if result.Role != "viewer" {
		t.Errorf("Expected role 'viewer', got '%s'", result.Role)
	}
}

func TestClient_RemoveUserFromProject(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/projects/proj-1/users/user-3" {
			t.Errorf("Expected path /api/v1/projects/proj-1/users/user-3, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
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

	// Test RemoveUserFromProject
	err = client.RemoveUserFromProject("proj-1", "user-3")
	if err != nil {
		t.Fatalf("RemoveUserFromProject failed: %v", err)
	}
}
