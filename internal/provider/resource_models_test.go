package provider

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestWorkflowResourceModel_Validation(t *testing.T) {
	tests := []struct {
		name          string
		model         WorkflowResourceModel
		expectValid   bool
		expectedError string
		desc          string
	}{
		{
			name: "valid workflow model",
			model: WorkflowResourceModel{
				ID:          types.StringValue("workflow_123"),
				Name:        types.StringValue("Test Workflow"),
				Active:      types.BoolValue(true),
				Nodes:       types.StringValue(`[{"id": "node1", "type": "trigger"}]`),
				Connections: types.StringValue(`{"node1": {"main": [[]]}}`),
				Settings:    types.StringValue(`{"executionOrder": "v1"}`),
				Tags:        types.ListValueMust(types.StringType, []attr.Value{types.StringValue("test")}),
			},
			expectValid: true,
			desc:        "should validate valid workflow model",
		},
		{
			name: "workflow with empty name",
			model: WorkflowResourceModel{
				ID:          types.StringValue("workflow_123"),
				Name:        types.StringValue(""),
				Active:      types.BoolValue(true),
				Nodes:       types.StringValue(`[]`),
				Connections: types.StringValue(`{}`),
			},
			expectValid: true, // Name validation happens at API level
			desc:        "should handle empty name (validated by API)",
		},
		{
			name: "workflow with null optional fields",
			model: WorkflowResourceModel{
				ID:          types.StringValue("workflow_123"),
				Name:        types.StringValue("Test Workflow"),
				Active:      types.BoolNull(),
				Nodes:       types.StringNull(),
				Connections: types.StringValue(`{}`),
				Settings:    types.StringNull(),
				StaticData:  types.StringNull(),
				PinnedData:  types.StringNull(),
				Tags:        types.ListNull(types.StringType),
			},
			expectValid: true,
			desc:        "should handle null optional fields",
		},
		{
			name: "workflow with unknown values",
			model: WorkflowResourceModel{
				ID:          types.StringUnknown(),
				Name:        types.StringValue("Test Workflow"),
				Active:      types.BoolUnknown(),
				Nodes:       types.StringValue(`[]`),
				Connections: types.StringValue(`{}`),
			},
			expectValid: true,
			desc:        "should handle unknown values during planning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test basic field access and validation
			if !tt.model.Name.IsNull() && !tt.model.Name.IsUnknown() {
				name := tt.model.Name.ValueString()
				if len(name) > 255 { // Reasonable length check
					t.Error("Workflow name should have reasonable length limits")
				}
			}

			// Test JSON field validation
			jsonFields := map[string]types.String{
				"nodes":       tt.model.Nodes,
				"connections": tt.model.Connections,
				"settings":    tt.model.Settings,
				"staticData":  tt.model.StaticData,
				"pinnedData":  tt.model.PinnedData,
			}

			for fieldName, fieldValue := range jsonFields {
				if !fieldValue.IsNull() && !fieldValue.IsUnknown() {
					jsonStr := fieldValue.ValueString()
					if jsonStr != "" {
						var jsonData interface{}
						if err := json.Unmarshal([]byte(jsonStr), &jsonData); err != nil {
							t.Errorf("Field %s contains invalid JSON: %v", fieldName, err)
						}
					}
				}
			}

			// Test tags list validation
			if !tt.model.Tags.IsNull() && !tt.model.Tags.IsUnknown() {
				tags := tt.model.Tags.Elements()
				for i, tag := range tags {
					if stringTag, ok := tag.(types.String); ok {
						if !stringTag.IsNull() && !stringTag.IsUnknown() {
							tagValue := stringTag.ValueString()
							if len(tagValue) == 0 {
								t.Errorf("Tag at index %d should not be empty", i)
							}
						}
					}
				}
			}

			t.Logf("Test case: %s", tt.desc)
		})
	}
}

func TestCredentialResourceModel_Validation(t *testing.T) {
	tests := []struct {
		name        string
		model       CredentialResourceModel
		expectValid bool
		desc        string
	}{
		{
			name: "valid credential model",
			model: CredentialResourceModel{
				ID:         types.StringValue("cred_123"),
				Name:       types.StringValue("Test Credential"),
				Type:       types.StringValue("httpBasicAuth"),
				Data:       types.StringValue(`{"username": "user", "password": "pass"}`),
				NodeAccess: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("httpRequest")}),
			},
			expectValid: true,
			desc:        "should validate valid credential model",
		},
		{
			name: "credential with null optional fields",
			model: CredentialResourceModel{
				ID:         types.StringValue("cred_123"),
				Name:       types.StringValue("Test Credential"),
				Type:       types.StringValue("httpBasicAuth"),
				Data:       types.StringValue(`{}`),
				NodeAccess: types.ListNull(types.StringType),
			},
			expectValid: true,
			desc:        "should handle null optional fields",
		},
		{
			name: "credential with sensitive data",
			model: CredentialResourceModel{
				ID:   types.StringValue("cred_123"),
				Name: types.StringValue("API Key Credential"),
				Type: types.StringValue("apiKey"),
				Data: types.StringValue(`{"apiKey": "secret_key_123", "header": "X-API-KEY"}`),
			},
			expectValid: true,
			desc:        "should handle sensitive credential data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test credential type validation
			if !tt.model.Type.IsNull() && !tt.model.Type.IsUnknown() {
				credType := tt.model.Type.ValueString()
				if credType == "" {
					t.Error("Credential type should not be empty")
				}
			}

			// Test credential data JSON validation
			if !tt.model.Data.IsNull() && !tt.model.Data.IsUnknown() {
				dataStr := tt.model.Data.ValueString()
				if dataStr != "" {
					var jsonData interface{}
					if err := json.Unmarshal([]byte(dataStr), &jsonData); err != nil {
						t.Errorf("Credential data contains invalid JSON: %v", err)
					}
				}
			}

			// Test node access validation
			if !tt.model.NodeAccess.IsNull() && !tt.model.NodeAccess.IsUnknown() {
				nodeAccess := tt.model.NodeAccess.Elements()
				for i, nodeType := range nodeAccess {
					if stringType, ok := nodeType.(types.String); ok {
						if !stringType.IsNull() && !stringType.IsUnknown() {
							nodeTypeValue := stringType.ValueString()
							if nodeTypeValue == "" {
								t.Errorf("Node type at index %d should not be empty", i)
							}
						}
					}
				}
			}

			t.Logf("Test case: %s", tt.desc)
		})
	}
}

func TestUserResourceModel_Validation(t *testing.T) {
	tests := []struct {
		name        string
		model       UserResourceModel
		expectValid bool
		desc        string
	}{
		{
			name: "valid user model",
			model: UserResourceModel{
				ID:        types.StringValue("user_123"),
				Email:     types.StringValue("test@example.com"),
				FirstName: types.StringValue("John"),
				LastName:  types.StringValue("Doe"),
				Password:  types.StringValue("secure_password"),
				Role:      types.StringValue("member"),
				Settings:  types.ObjectNull(map[string]attr.Type{}),
			},
			expectValid: true,
			desc:        "should validate valid user model",
		},
		{
			name: "user with null optional fields",
			model: UserResourceModel{
				ID:        types.StringValue("user_123"),
				Email:     types.StringValue("test@example.com"),
				FirstName: types.StringNull(),
				LastName:  types.StringNull(),
				Password:  types.StringValue("password"),
				Settings:  types.ObjectNull(map[string]attr.Type{}),
			},
			expectValid: true,
			desc:        "should handle null optional fields",
		},
		{
			name: "user with role and settings",
			model: UserResourceModel{
				ID:       types.StringValue("user_123"),
				Email:    types.StringValue("admin@example.com"),
				Password: types.StringValue("admin_password"),
				Role:     types.StringValue("admin"),
				IsOwner:  types.BoolValue(true),
				Settings: types.ObjectNull(map[string]attr.Type{}),
			},
			expectValid: true,
			desc:        "should handle admin user configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test email validation (basic format check)
			if !tt.model.Email.IsNull() && !tt.model.Email.IsUnknown() {
				email := tt.model.Email.ValueString()
				if email != "" && !isValidEmailFormat(email) {
					t.Errorf("Invalid email format: %s", email)
				}
			}

			// Test password validation (non-empty check)
			if !tt.model.Password.IsNull() && !tt.model.Password.IsUnknown() {
				password := tt.model.Password.ValueString()
				if password == "" {
					t.Error("Password should not be empty")
				}
			}

			// Test settings object validation
			if !tt.model.Settings.IsNull() && !tt.model.Settings.IsUnknown() {
				// Settings is a types.Object, not a JSON string
				// This would be validated through the Terraform framework
				t.Logf("Settings object present for user: %s", tt.model.Email.ValueString())
			}

			t.Logf("Test case: %s", tt.desc)
		})
	}
}

func TestProjectResourceModel_Validation(t *testing.T) {
	tests := []struct {
		name        string
		model       ProjectResourceModel
		expectValid bool
		desc        string
	}{
		{
			name: "valid project model",
			model: ProjectResourceModel{
				ID:          types.StringValue("proj_123"),
				Name:        types.StringValue("Test Project"),
				Description: types.StringValue("A test project"),
				Settings:    types.StringValue(`{"homeProject": false}`),
			},
			expectValid: true,
			desc:        "should validate valid project model",
		},
		{
			name: "project with null optional fields",
			model: ProjectResourceModel{
				ID:          types.StringValue("proj_123"),
				Name:        types.StringValue("Simple Project"),
				Description: types.StringNull(),
				Settings:    types.StringNull(),
			},
			expectValid: true,
			desc:        "should handle null optional fields",
		},
		{
			name: "project with additional metadata",
			model: ProjectResourceModel{
				ID:          types.StringValue("home_proj"),
				Name:        types.StringValue("Home Project"),
				Description: types.StringValue("Main project for workflows"),
				Settings:    types.StringValue(`{"homeProject": true, "defaultWorkflow": true}`),
				Icon:        types.StringValue("home"),
				Color:       types.StringValue("#3f82f6"),
				OwnerID:     types.StringValue("user_123"),
			},
			expectValid: true,
			desc:        "should handle project with metadata",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test project name validation
			if !tt.model.Name.IsNull() && !tt.model.Name.IsUnknown() {
				name := tt.model.Name.ValueString()
				if name == "" {
					t.Error("Project name should not be empty")
				}
			}

			// Test project description validation
			if !tt.model.Description.IsNull() && !tt.model.Description.IsUnknown() {
				description := tt.model.Description.ValueString()
				if len(description) > 500 { // Reasonable description length
					t.Error("Project description should have reasonable length limits")
				}
			}

			// Test settings JSON validation
			if !tt.model.Settings.IsNull() && !tt.model.Settings.IsUnknown() {
				settingsStr := tt.model.Settings.ValueString()
				if settingsStr != "" {
					var jsonData interface{}
					if err := json.Unmarshal([]byte(settingsStr), &jsonData); err != nil {
						t.Errorf("Project settings contain invalid JSON: %v", err)
					}
				}
			}

			t.Logf("Test case: %s", tt.desc)
		})
	}
}

func TestLDAPConfigResourceModel_Validation(t *testing.T) {
	tests := []struct {
		name        string
		model       LDAPConfigResourceModel
		expectValid bool
		desc        string
	}{
		{
			name: "valid LDAP config model",
			model: LDAPConfigResourceModel{
				ID:                     types.StringValue("ldap_123"),
				ServerURL:              types.StringValue("ldap://ldap.example.com:389"),
				BindDN:                 types.StringValue("cn=admin,dc=example,dc=com"),
				BindPassword:           types.StringValue("admin_password"),
				SearchBase:             types.StringValue("dc=example,dc=com"),
				SearchFilter:           types.StringValue("(uid={username})"),
				UserIDAttribute:        types.StringValue("uid"),
				UserEmailAttribute:     types.StringValue("mail"),
				UserFirstNameAttribute: types.StringValue("givenName"),
				UserLastNameAttribute:  types.StringValue("sn"),
			},
			expectValid: true,
			desc:        "should validate valid LDAP config model",
		},
		{
			name: "LDAP config with null optional fields",
			model: LDAPConfigResourceModel{
				ID:                     types.StringValue("ldap_123"),
				ServerURL:              types.StringValue("ldap://ldap.example.com:389"),
				BindDN:                 types.StringValue("cn=admin,dc=example,dc=com"),
				BindPassword:           types.StringValue("password"),
				SearchBase:             types.StringValue("dc=example,dc=com"),
				SearchFilter:           types.StringNull(),
				UserIDAttribute:        types.StringNull(),
				UserEmailAttribute:     types.StringNull(),
				UserFirstNameAttribute: types.StringNull(),
				UserLastNameAttribute:  types.StringNull(),
			},
			expectValid: true,
			desc:        "should handle null optional fields",
		},
		{
			name: "LDAP config with LDAPS",
			model: LDAPConfigResourceModel{
				ID:           types.StringValue("ldaps_123"),
				ServerURL:    types.StringValue("ldaps://secure-ldap.example.com:636"),
				BindDN:       types.StringValue("cn=readonly,dc=secure,dc=example,dc=com"),
				BindPassword: types.StringValue("readonly_password"),
				SearchBase:   types.StringValue("dc=secure,dc=example,dc=com"),
			},
			expectValid: true,
			desc:        "should handle LDAPS configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test server URL validation
			if !tt.model.ServerURL.IsNull() && !tt.model.ServerURL.IsUnknown() {
				serverURL := tt.model.ServerURL.ValueString()
				if serverURL == "" {
					t.Error("LDAP server URL should not be empty")
				}
				if !strings.HasPrefix(serverURL, "ldap://") && !strings.HasPrefix(serverURL, "ldaps://") {
					t.Errorf("Invalid LDAP server URL format: %s", serverURL)
				}
			}

			// Test search base validation
			if !tt.model.SearchBase.IsNull() && !tt.model.SearchBase.IsUnknown() {
				searchBase := tt.model.SearchBase.ValueString()
				if searchBase == "" {
					t.Error("LDAP search base should not be empty")
				}
			}

			// Test bind DN validation
			if !tt.model.BindDN.IsNull() && !tt.model.BindDN.IsUnknown() {
				bindDN := tt.model.BindDN.ValueString()
				if bindDN == "" {
					t.Error("LDAP bind DN should not be empty")
				}
			}

			t.Logf("Test case: %s", tt.desc)
		})
	}
}

func TestResourceModel_TypeSafety(t *testing.T) {
	t.Run("workflow resource model types", func(t *testing.T) {
		model := WorkflowResourceModel{}

		// Test that all fields have expected types
		_ = model.ID.ValueString()          // types.String
		_ = model.Name.ValueString()        // types.String
		_ = model.Active.ValueBool()        // types.Bool
		_ = model.Nodes.ValueString()       // types.String
		_ = model.Connections.ValueString() // types.String
		_ = model.Settings.ValueString()    // types.String
		_ = model.StaticData.ValueString()  // types.String
		_ = model.PinnedData.ValueString()  // types.String
		_ = model.Tags.Elements()           // types.List
	})

	t.Run("credential resource model types", func(t *testing.T) {
		model := CredentialResourceModel{}

		// Test that all fields have expected types
		_ = model.ID.ValueString()      // types.String
		_ = model.Name.ValueString()    // types.String
		_ = model.Type.ValueString()    // types.String
		_ = model.Data.ValueString()    // types.String
		_ = model.NodeAccess.Elements() // types.List
	})

	t.Run("user resource model types", func(t *testing.T) {
		model := UserResourceModel{}

		// Test that all fields have expected types
		_ = model.ID.ValueString()        // types.String
		_ = model.Email.ValueString()     // types.String
		_ = model.FirstName.ValueString() // types.String
		_ = model.LastName.ValueString()  // types.String
		_ = model.Password.ValueString()  // types.String
		_ = model.Settings.Attributes()   // types.Object
	})
}

func TestResourceModel_NullAndUnknownHandling(t *testing.T) {
	tests := []struct {
		name  string
		field interface{}
		desc  string
	}{
		{
			name:  "string null handling",
			field: types.StringNull(),
			desc:  "should handle null string values",
		},
		{
			name:  "string unknown handling",
			field: types.StringUnknown(),
			desc:  "should handle unknown string values",
		},
		{
			name:  "bool null handling",
			field: types.BoolNull(),
			desc:  "should handle null bool values",
		},
		{
			name:  "bool unknown handling",
			field: types.BoolUnknown(),
			desc:  "should handle unknown bool values",
		},
		{
			name:  "list null handling",
			field: types.ListNull(types.StringType),
			desc:  "should handle null list values",
		},
		{
			name:  "list unknown handling",
			field: types.ListUnknown(types.StringType),
			desc:  "should handle unknown list values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.field.(type) {
			case types.String:
				if !v.IsNull() && !v.IsUnknown() {
					_ = v.ValueString()
				}
			case types.Bool:
				if !v.IsNull() && !v.IsUnknown() {
					_ = v.ValueBool()
				}
			case types.List:
				if !v.IsNull() && !v.IsUnknown() {
					_ = v.Elements()
				}
			}

			t.Logf("Test case: %s", tt.desc)
		})
	}
}

// Helper functions

func isValidEmailFormat(email string) bool {
	// Simple email validation for testing
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// contains is a helper function for tests
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Use the contains function to avoid unused warning
var _ = contains
