package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetWithPagination(t *testing.T) {
	tests := []struct {
		name           string
		response       map[string]interface{}
		expectedPaging *PaginationInfo
		wantErr        bool
	}{
		{
			name: "response with nextCursor and total",
			response: map[string]interface{}{
				"data": []map[string]interface{}{
					{"id": "1", "name": "item1"},
					{"id": "2", "name": "item2"},
				},
				"nextCursor": "cursor_123",
				"total":      10.0, // JSON numbers are float64
			},
			expectedPaging: &PaginationInfo{
				NextCursor: "cursor_123",
				HasNext:    true,
				Total:      10,
			},
			wantErr: false,
		},
		{
			name: "response with empty nextCursor",
			response: map[string]interface{}{
				"data": []map[string]interface{}{
					{"id": "1", "name": "item1"},
				},
				"nextCursor": "",
				"total":      1.0,
			},
			expectedPaging: &PaginationInfo{
				NextCursor: "",
				HasNext:    false,
				Total:      1,
			},
			wantErr: false,
		},
		{
			name: "response without pagination fields",
			response: map[string]interface{}{
				"data": []map[string]interface{}{
					{"id": "1", "name": "item1"},
				},
			},
			expectedPaging: &PaginationInfo{
				NextCursor: "",
				HasNext:    false,
				Total:      0,
			},
			wantErr: false,
		},
		{
			name: "response with null nextCursor",
			response: map[string]interface{}{
				"data":       []map[string]interface{}{{"id": "1"}},
				"nextCursor": nil,
				"total":      5.0,
			},
			expectedPaging: &PaginationInfo{
				NextCursor: "",
				HasNext:    false,
				Total:      5,
			},
			wantErr: false,
		},
		{
			name: "response with invalid total type",
			response: map[string]interface{}{
				"data":       []map[string]interface{}{{"id": "1"}},
				"nextCursor": "cursor_abc",
				"total":      "invalid_number",
			},
			expectedPaging: &PaginationInfo{
				NextCursor: "cursor_abc",
				HasNext:    true,
				Total:      0, // Should default to 0 for invalid type
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := CreateTestClient(t, server.URL)

			var result map[string]interface{}
			pagination, err := client.GetWithPagination("test", &result)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if pagination == nil {
				t.Fatal("Expected pagination info but got nil")
			}

			if pagination.NextCursor != tt.expectedPaging.NextCursor {
				t.Errorf("Expected NextCursor %q, got %q", tt.expectedPaging.NextCursor, pagination.NextCursor)
			}

			if pagination.HasNext != tt.expectedPaging.HasNext {
				t.Errorf("Expected HasNext %v, got %v", tt.expectedPaging.HasNext, pagination.HasNext)
			}

			if pagination.Total != tt.expectedPaging.Total {
				t.Errorf("Expected Total %d, got %d", tt.expectedPaging.Total, pagination.Total)
			}
		})
	}
}

func TestClient_GetWithPagination_ErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		wantErr    bool
	}{
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			response:   `{"code": 500, "message": "Internal Server Error"}`,
			wantErr:    true,
		},
		{
			name:       "unauthorized error",
			statusCode: http.StatusUnauthorized,
			response:   `{"code": 401, "message": "Unauthorized"}`,
			wantErr:    true,
		},
		{
			name:       "invalid JSON response",
			statusCode: http.StatusOK,
			response:   `invalid json`,
			wantErr:    true,
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

			var result map[string]interface{}
			pagination, err := client.GetWithPagination("test", &result)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if pagination != nil {
					t.Error("Expected nil pagination on error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if pagination == nil {
					t.Error("Expected pagination info but got nil")
				}
			}
		})
	}
}

func TestClient_GetWithPagination_RealWorldScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real-world scenarios test in short mode")
	}
	t.Run("workflows pagination", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate workflow list API response
			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":     "workflow_1",
						"name":   "My First Workflow",
						"active": true,
						"tags":   []string{"production"},
					},
					{
						"id":     "workflow_2",
						"name":   "My Second Workflow",
						"active": false,
						"tags":   []string{"development"},
					},
				},
				"nextCursor": "eyJpZCI6IndvcmtmbG93XzIifQ==", // Base64 encoded cursor
				"total":      25.0,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := CreateTestClient(t, server.URL)

		var result map[string]interface{}
		pagination, err := client.GetWithPagination("workflows", &result)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify pagination info
		if pagination.NextCursor != "eyJpZCI6IndvcmtmbG93XzIifQ==" {
			t.Errorf("Expected specific cursor, got %q", pagination.NextCursor)
		}
		if !pagination.HasNext {
			t.Error("Expected HasNext to be true")
		}
		if pagination.Total != 25 {
			t.Errorf("Expected total 25, got %d", pagination.Total)
		}

		// Verify response data
		data, ok := result["data"].([]interface{})
		if !ok {
			t.Fatal("Expected data to be an array")
		}
		if len(data) != 2 {
			t.Errorf("Expected 2 workflows, got %d", len(data))
		}
	})

	t.Run("users pagination", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate users list API response
			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":        "user_1",
						"email":     "admin@example.com",
						"firstName": "Admin",
						"lastName":  "User",
						"isOwner":   true,
					},
				},
				"nextCursor": "", // Last page
				"total":      1.0,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := CreateTestClient(t, server.URL)

		var result map[string]interface{}
		pagination, err := client.GetWithPagination("users", &result)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify this is the last page
		if pagination.NextCursor != "" {
			t.Errorf("Expected empty cursor for last page, got %q", pagination.NextCursor)
		}
		if pagination.HasNext {
			t.Error("Expected HasNext to be false for last page")
		}
		if pagination.Total != 1 {
			t.Errorf("Expected total 1, got %d", pagination.Total)
		}
	})

	t.Run("empty result set", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := map[string]interface{}{
				"data":       []map[string]interface{}{},
				"nextCursor": "",
				"total":      0.0,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := CreateTestClient(t, server.URL)

		var result map[string]interface{}
		pagination, err := client.GetWithPagination("empty-collection", &result)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if pagination.NextCursor != "" {
			t.Errorf("Expected empty cursor, got %q", pagination.NextCursor)
		}
		if pagination.HasNext {
			t.Error("Expected HasNext to be false for empty result")
		}
		if pagination.Total != 0 {
			t.Errorf("Expected total 0, got %d", pagination.Total)
		}

		// Verify empty data array
		data, ok := result["data"].([]interface{})
		if !ok {
			t.Fatal("Expected data to be an array")
		}
		if len(data) != 0 {
			t.Errorf("Expected empty data array, got %d items", len(data))
		}
	})
}

func TestPaginationInfo_DefaultValues(t *testing.T) {
	// Test that PaginationInfo fields have expected zero values
	pagination := &PaginationInfo{}

	if pagination.Limit != 0 {
		t.Errorf("Expected Limit to be 0, got %d", pagination.Limit)
	}
	if pagination.Offset != 0 {
		t.Errorf("Expected Offset to be 0, got %d", pagination.Offset)
	}
	if pagination.Total != 0 {
		t.Errorf("Expected Total to be 0, got %d", pagination.Total)
	}
	if pagination.NextCursor != "" {
		t.Errorf("Expected NextCursor to be empty, got %q", pagination.NextCursor)
	}
	if pagination.HasNext {
		t.Error("Expected HasNext to be false")
	}
}

func TestClient_GetWithPagination_TypeAssertions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping type assertions test in short mode")
	}
	tests := []struct {
		name     string
		response map[string]interface{}
		desc     string
	}{
		{
			name: "nextCursor as non-string",
			response: map[string]interface{}{
				"data":       []interface{}{},
				"nextCursor": 12345, // Should be ignored
				"total":      5.0,
			},
			desc: "non-string nextCursor should be ignored",
		},
		{
			name: "total as string number",
			response: map[string]interface{}{
				"data":       []interface{}{},
				"nextCursor": "cursor123",
				"total":      "10", // String instead of number
			},
			desc: "string total should default to 0",
		},
		{
			name: "total as integer",
			response: map[string]interface{}{
				"data":       []interface{}{},
				"nextCursor": "cursor123",
				"total":      int64(15), // Integer instead of float64
			},
			desc: "integer total should default to 0 (only float64 handled)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := CreateTestClient(t, server.URL)

			var result map[string]interface{}
			pagination, err := client.GetWithPagination("test", &result)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Test should not crash and should handle type mismatches gracefully
			if pagination == nil {
				t.Fatal("Expected pagination info but got nil")
			}

			t.Logf("Test case: %s - Pagination: %+v", tt.desc, pagination)
		})
	}
}

func TestClient_GetWithPagination_NonMapResult(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping non-map result test in short mode")
	}
	// Test when result is not a map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := []map[string]interface{}{
			{"id": "1", "name": "item1"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := CreateTestClient(t, server.URL)

	// Use a slice as result instead of map
	var result []map[string]interface{}
	pagination, err := client.GetWithPagination("test", &result)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should create default pagination info when result is not a map
	if pagination == nil {
		t.Fatal("Expected pagination info but got nil")
	}

	// Should have default values since we can't parse pagination from a slice result
	if pagination.NextCursor != "" {
		t.Errorf("Expected empty NextCursor, got %q", pagination.NextCursor)
	}
	if pagination.HasNext {
		t.Error("Expected HasNext to be false")
	}
	if pagination.Total != 0 {
		t.Errorf("Expected Total to be 0, got %d", pagination.Total)
	}
}
