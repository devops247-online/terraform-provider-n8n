package provider

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestWorkflowResourceModel_JSONValidation(t *testing.T) {
	tests := []struct {
		name      string
		jsonField string
		jsonValue string
		wantError bool
	}{
		{
			name:      "valid nodes JSON",
			jsonField: "nodes",
			jsonValue: `[{"id": "node1", "type": "trigger", "position": [100, 200]}]`,
			wantError: false,
		},
		{
			name:      "invalid nodes JSON",
			jsonField: "nodes",
			jsonValue: `invalid json`,
			wantError: true,
		},
		{
			name:      "valid connections JSON",
			jsonField: "connections",
			jsonValue: `{"node1": {"main": [[{"node": "node2", "type": "main", "index": 0}]]}}`,
			wantError: false,
		},
		{
			name:      "invalid connections JSON",
			jsonField: "connections",
			jsonValue: `{invalid}`,
			wantError: true,
		},
		{
			name:      "valid settings JSON",
			jsonField: "settings",
			jsonValue: `{"executionOrder": "v1", "saveDataErrorExecution": "all"}`,
			wantError: false,
		},
		{
			name:      "empty JSON object",
			jsonField: "settings",
			jsonValue: `{}`,
			wantError: false,
		},
		{
			name:      "empty string",
			jsonField: "settings",
			jsonValue: ``,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jsonData interface{}
			err := json.Unmarshal([]byte(tt.jsonValue), &jsonData)

			if tt.wantError && err == nil {
				t.Error("Expected JSON unmarshaling error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected JSON unmarshaling error: %v", err)
			}
		})
	}
}

func TestWorkflowResourceModel_FieldTypes(t *testing.T) {
	model := WorkflowResourceModel{
		ID:          types.StringValue("workflow-123"),
		Name:        types.StringValue("Test Workflow"),
		Active:      types.BoolValue(true),
		Nodes:       types.StringValue(`[{"id": "node1"}]`),
		Connections: types.StringValue(`{}`),
		Settings:    types.StringValue(`{"executionOrder": "v1"}`),
		Tags:        types.ListValueMust(types.StringType, []attr.Value{types.StringValue("tag1")}),
	}

	// Test that all fields are properly typed
	if model.ID.IsNull() || model.ID.IsUnknown() {
		t.Error("ID should not be null or unknown")
	}
	if model.Name.ValueString() != "Test Workflow" {
		t.Error("Name not set correctly")
	}
	if !model.Active.ValueBool() {
		t.Error("Active should be true")
	}

	// Test JSON fields
	if model.Nodes.IsNull() {
		t.Error("Nodes should not be null")
	}
	if model.Connections.IsNull() {
		t.Error("Connections should not be null")
	}
	if model.Settings.IsNull() {
		t.Error("Settings should not be null")
	}

	// Test tags list
	if model.Tags.IsNull() {
		t.Error("Tags should not be null")
	}

	tagElements := model.Tags.Elements()
	if len(tagElements) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(tagElements))
	}
}

func TestCredentialResourceModel_JSONValidation(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		wantError bool
	}{
		{
			name:      "valid credential data",
			data:      `{"clientId": "test-id", "clientSecret": "test-secret"}`,
			wantError: false,
		},
		{
			name:      "empty credential data",
			data:      `{}`,
			wantError: false,
		},
		{
			name:      "invalid credential data",
			data:      `{invalid json}`,
			wantError: true,
		},
		{
			name:      "null values in data",
			data:      `{"clientId": null, "clientSecret": "secret"}`,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jsonData interface{}
			err := json.Unmarshal([]byte(tt.data), &jsonData)

			if tt.wantError && err == nil {
				t.Error("Expected JSON unmarshaling error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected JSON unmarshaling error: %v", err)
			}
		})
	}
}

func TestUserResourceModel_SettingsValidation(t *testing.T) {
	tests := []struct {
		name      string
		settings  string
		wantError bool
	}{
		{
			name:      "valid user settings",
			settings:  `{"timezone": "UTC", "dateFormat": "MM/dd/yyyy"}`,
			wantError: false,
		},
		{
			name:      "empty user settings",
			settings:  `{}`,
			wantError: false,
		},
		{
			name:      "invalid user settings JSON",
			settings:  `{invalid}`,
			wantError: true,
		},
		{
			name:      "complex nested settings",
			settings:  `{"ui": {"theme": "dark", "compact": true}, "notifications": {"email": false}}`,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jsonData interface{}
			err := json.Unmarshal([]byte(tt.settings), &jsonData)

			if tt.wantError && err == nil {
				t.Error("Expected JSON unmarshaling error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected JSON unmarshaling error: %v", err)
			}
		})
	}
}

func TestJSONFieldMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "simple object",
			input:    map[string]interface{}{"key": "value"},
			expected: `{"key":"value"}`,
		},
		{
			name:     "array",
			input:    []interface{}{"item1", "item2"},
			expected: `["item1","item2"]`,
		},
		{
			name:     "nested object",
			input:    map[string]interface{}{"outer": map[string]interface{}{"inner": "value"}},
			expected: `{"outer":{"inner":"value"}}`,
		},
		{
			name:     "empty object",
			input:    map[string]interface{}{},
			expected: `{}`,
		},
		{
			name:     "null value",
			input:    nil,
			expected: `null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			jsonBytes, err := json.Marshal(tt.input)
			if err != nil {
				t.Errorf("Failed to marshal JSON: %v", err)
				return
			}

			jsonString := string(jsonBytes)
			if jsonString != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, jsonString)
			}

			// Unmarshal back
			var result interface{}
			err = json.Unmarshal(jsonBytes, &result)
			if err != nil {
				t.Errorf("Failed to unmarshal JSON: %v", err)
			}
		})
	}
}
