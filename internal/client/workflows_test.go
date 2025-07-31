package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_GetWorkflows(t *testing.T) {
	mockWorkflows := WorkflowListResponse{
		Data: []Workflow{
			{
				ID:     "1",
				Name:   "Test Workflow 1",
				Active: true,
				Tags:   []string{"tag1", "tag2"},
			},
			{
				ID:     "2",
				Name:   "Test Workflow 2",
				Active: false,
				Tags:   []string{"tag3"},
			},
		},
		NextCursor: "next-cursor-123",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/workflows" {
			t.Errorf("Expected path '/api/v1/workflows', got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockWorkflows)
	}))
	defer server.Close()

	config := &Config{
		BaseURL: server.URL,
		Auth:    &APIKeyAuth{APIKey: "test-key"},
		Timeout: time.Second * 5,
	}
	client, _ := NewClient(config)
	client.httpClient = server.Client()

	result, err := client.GetWorkflows(nil)
	if err != nil {
		t.Fatalf("GetWorkflows failed: %v", err)
	}

	if len(result.Data) != 2 {
		t.Errorf("Expected 2 workflows, got %d", len(result.Data))
	}

	if result.Data[0].Name != "Test Workflow 1" {
		t.Errorf("Expected first workflow name 'Test Workflow 1', got %s", result.Data[0].Name)
	}

	if result.NextCursor != "next-cursor-123" {
		t.Errorf("Expected NextCursor 'next-cursor-123', got %s", result.NextCursor)
	}
}

func TestClient_GetWorkflowsWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/v1/workflows"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path '%s', got %s", expectedPath, r.URL.Path)
		}

		query := r.URL.Query()
		if query.Get("active") != "true" {
			t.Errorf("Expected active=true, got %s", query.Get("active"))
		}
		if query.Get("limit") != "10" {
			t.Errorf("Expected limit=10, got %s", query.Get("limit"))
		}
		if query.Get("offset") != "5" {
			t.Errorf("Expected offset=5, got %s", query.Get("offset"))
		}
		if query.Get("projectId") != "project-123" {
			t.Errorf("Expected projectId=project-123, got %s", query.Get("projectId"))
		}

		tags := query["tags"]
		if len(tags) != 2 || tags[0] != "tag1" || tags[1] != "tag2" {
			t.Errorf("Expected tags [tag1, tag2], got %v", tags)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(WorkflowListResponse{Data: []Workflow{}})
	}))
	defer server.Close()

	config := &Config{
		BaseURL: server.URL,
		Auth:    &APIKeyAuth{APIKey: "test-key"},
		Timeout: time.Second * 5,
	}
	client, _ := NewClient(config)
	client.httpClient = server.Client()

	active := true
	options := &WorkflowListOptions{
		Active:    &active,
		Tags:      []string{"tag1", "tag2"},
		ProjectID: "project-123",
		Limit:     10,
		Offset:    5,
	}

	_, err := client.GetWorkflows(options)
	if err != nil {
		t.Fatalf("GetWorkflows with options failed: %v", err)
	}
}

func TestClient_GetWorkflow(t *testing.T) {
	mockWorkflow := Workflow{
		ID:        "test-id",
		Name:      "Test Workflow",
		Active:    true,
		Tags:      []string{"test"},
		CreatedAt: &time.Time{},
		UpdatedAt: &time.Time{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/v1/workflows/test-id"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path '%s', got %s", expectedPath, r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockWorkflow)
	}))
	defer server.Close()

	config := &Config{
		BaseURL: server.URL,
		Auth:    &APIKeyAuth{APIKey: "test-key"},
		Timeout: time.Second * 5,
	}
	client, _ := NewClient(config)
	client.httpClient = server.Client()

	result, err := client.GetWorkflow("test-id")
	if err != nil {
		t.Fatalf("GetWorkflow failed: %v", err)
	}

	if result.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got %s", result.ID)
	}
	if result.Name != "Test Workflow" {
		t.Errorf("Expected name 'Test Workflow', got %s", result.Name)
	}
}

func TestClient_GetWorkflowEmptyID(t *testing.T) {
	client := &Client{}

	_, err := client.GetWorkflow("")
	if err == nil {
		t.Error("Expected error for empty workflow ID")
	}
	if err.Error() != "workflow ID is required" {
		t.Errorf("Expected 'workflow ID is required', got %s", err.Error())
	}
}

func TestClient_CreateWorkflow(t *testing.T) {
	inputWorkflow := &Workflow{
		Name:   "New Workflow",
		Active: false,
		Nodes:  []interface{}{map[string]interface{}{"id": "node1", "type": "trigger"}},
	}

	mockResponse := Workflow{
		ID:     "new-id",
		Name:   "New Workflow",
		Active: false,
		Nodes:  []interface{}{map[string]interface{}{"id": "node1", "type": "trigger"}},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/workflows" {
			t.Errorf("Expected path '/api/v1/workflows', got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		var receivedWorkflow Workflow
		_ = json.NewDecoder(r.Body).Decode(&receivedWorkflow)
		if receivedWorkflow.Name != "New Workflow" {
			t.Errorf("Expected name 'New Workflow', got %s", receivedWorkflow.Name)
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	config := &Config{
		BaseURL: server.URL,
		Auth:    &APIKeyAuth{APIKey: "test-key"},
		Timeout: time.Second * 5,
	}
	client, _ := NewClient(config)
	client.httpClient = server.Client()

	result, err := client.CreateWorkflow(inputWorkflow)
	if err != nil {
		t.Fatalf("CreateWorkflow failed: %v", err)
	}

	if result.ID != "new-id" {
		t.Errorf("Expected ID 'new-id', got %s", result.ID)
	}
	if result.Name != "New Workflow" {
		t.Errorf("Expected name 'New Workflow', got %s", result.Name)
	}
}

func TestClient_CreateWorkflowValidation(t *testing.T) {
	client := &Client{}

	tests := []struct {
		name     string
		workflow *Workflow
		wantErr  string
	}{
		{
			name:     "nil workflow",
			workflow: nil,
			wantErr:  "workflow is required",
		},
		{
			name:     "empty name",
			workflow: &Workflow{Name: ""},
			wantErr:  "workflow name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.CreateWorkflow(tt.workflow)
			if err == nil {
				t.Error("Expected error for invalid workflow")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("Expected error '%s', got '%s'", tt.wantErr, err.Error())
			}
		})
	}
}

func TestClient_UpdateWorkflow(t *testing.T) {
	inputWorkflow := &Workflow{
		Name:   "Updated Workflow",
		Active: true,
	}

	mockResponse := *inputWorkflow
	mockResponse.ID = "test-id"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/v1/workflows/test-id"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path '%s', got %s", expectedPath, r.URL.Path)
		}
		if r.Method != "PUT" {
			t.Errorf("Expected PUT method, got %s", r.Method)
		}

		var receivedWorkflow Workflow
		_ = json.NewDecoder(r.Body).Decode(&receivedWorkflow)
		if receivedWorkflow.Name != "Updated Workflow" {
			t.Errorf("Expected name 'Updated Workflow', got %s", receivedWorkflow.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	config := &Config{
		BaseURL: server.URL,
		Auth:    &APIKeyAuth{APIKey: "test-key"},
		Timeout: time.Second * 5,
	}
	client, _ := NewClient(config)
	client.httpClient = server.Client()

	result, err := client.UpdateWorkflow("test-id", inputWorkflow)
	if err != nil {
		t.Fatalf("UpdateWorkflow failed: %v", err)
	}

	if result.Name != "Updated Workflow" {
		t.Errorf("Expected name 'Updated Workflow', got %s", result.Name)
	}
}

func TestClient_UpdateWorkflowValidation(t *testing.T) {
	client := &Client{}

	tests := []struct {
		name     string
		id       string
		workflow *Workflow
		wantErr  string
	}{
		{
			name:     "empty ID",
			id:       "",
			workflow: &Workflow{Name: "Test"},
			wantErr:  "workflow ID is required",
		},
		{
			name:     "nil workflow",
			id:       "test-id",
			workflow: nil,
			wantErr:  "workflow is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.UpdateWorkflow(tt.id, tt.workflow)
			if err == nil {
				t.Error("Expected error for invalid input")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("Expected error '%s', got '%s'", tt.wantErr, err.Error())
			}
		})
	}
}

func TestClient_DeleteWorkflow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/v1/workflows/test-id"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path '%s', got %s", expectedPath, r.URL.Path)
		}
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	config := &Config{
		BaseURL: server.URL,
		Auth:    &APIKeyAuth{APIKey: "test-key"},
		Timeout: time.Second * 5,
	}
	client, _ := NewClient(config)
	client.httpClient = server.Client()

	err := client.DeleteWorkflow("test-id")
	if err != nil {
		t.Fatalf("DeleteWorkflow failed: %v", err)
	}
}

func TestClient_DeleteWorkflowEmptyID(t *testing.T) {
	client := &Client{}

	err := client.DeleteWorkflow("")
	if err == nil {
		t.Error("Expected error for empty workflow ID")
	}
	if err.Error() != "workflow ID is required" {
		t.Errorf("Expected 'workflow ID is required', got %s", err.Error())
	}
}

func TestClient_ActivateWorkflow(t *testing.T) {
	mockResponse := Workflow{
		ID:     "test-id",
		Name:   "Test Workflow",
		Active: true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/v1/workflows/test-id/activate"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path '%s', got %s", expectedPath, r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	config := &Config{
		BaseURL: server.URL,
		Auth:    &APIKeyAuth{APIKey: "test-key"},
		Timeout: time.Second * 5,
	}
	client, _ := NewClient(config)
	client.httpClient = server.Client()

	result, err := client.ActivateWorkflow("test-id")
	if err != nil {
		t.Fatalf("ActivateWorkflow failed: %v", err)
	}

	if result.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got %s", result.ID)
	}
	if !result.Active {
		t.Error("Expected workflow to be active")
	}
}

func TestClient_ActivateWorkflowEmptyID(t *testing.T) {
	client := &Client{}

	_, err := client.ActivateWorkflow("")
	if err == nil {
		t.Error("Expected error for empty workflow ID")
	}
	if err.Error() != "workflow ID is required" {
		t.Errorf("Expected 'workflow ID is required', got %s", err.Error())
	}
}

func TestClient_DeactivateWorkflow(t *testing.T) {
	mockResponse := Workflow{
		ID:     "test-id",
		Name:   "Test Workflow",
		Active: false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/v1/workflows/test-id/deactivate"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path '%s', got %s", expectedPath, r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	config := &Config{
		BaseURL: server.URL,
		Auth:    &APIKeyAuth{APIKey: "test-key"},
		Timeout: time.Second * 5,
	}
	client, _ := NewClient(config)
	client.httpClient = server.Client()

	result, err := client.DeactivateWorkflow("test-id")
	if err != nil {
		t.Fatalf("DeactivateWorkflow failed: %v", err)
	}

	if result.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got %s", result.ID)
	}
	if result.Active {
		t.Error("Expected workflow to be inactive")
	}
}

func TestClient_DeactivateWorkflowEmptyID(t *testing.T) {
	client := &Client{}

	_, err := client.DeactivateWorkflow("")
	if err == nil {
		t.Error("Expected error for empty workflow ID")
	}
	if err.Error() != "workflow ID is required" {
		t.Errorf("Expected 'workflow ID is required', got %s", err.Error())
	}
}
