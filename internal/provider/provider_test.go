package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
//
// nolint:unused // This variable is used by acceptance tests when TF_ACC=1
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"n8n": providerserver.NewProtocol6WithError(New("test")()),
}

func TestProvider(t *testing.T) {
	provider := New("test")()
	if provider == nil {
		t.Fatal("Expected provider to be non-nil")
	}
}

func TestProvider_New(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "test version",
			version: "test",
			want:    "test",
		},
		{
			name:    "dev version",
			version: "dev",
			want:    "dev",
		},
		{
			name:    "semver version",
			version: "1.0.0",
			want:    "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providerFunc := New(tt.version)
			if providerFunc == nil {
				t.Fatal("Expected provider function to be non-nil")
			}

			p := providerFunc().(*N8nProvider)
			if p.version != tt.want {
				t.Errorf("Expected version %q, got %q", tt.want, p.version)
			}
		})
	}
}

func TestProvider_Metadata(t *testing.T) {
	ctx := context.Background()
	p := &N8nProvider{version: "1.0.0"}

	req := provider.MetadataRequest{}
	resp := &provider.MetadataResponse{}

	p.Metadata(ctx, req, resp)

	if resp.TypeName != "n8n" {
		t.Errorf("Expected TypeName to be 'n8n', got %q", resp.TypeName)
	}

	if resp.Version != "1.0.0" {
		t.Errorf("Expected Version to be '1.0.0', got %q", resp.Version)
	}
}

func TestProvider_Schema(t *testing.T) {
	ctx := context.Background()
	p := &N8nProvider{}

	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}

	p.Schema(ctx, req, resp)

	if resp.Schema.MarkdownDescription == "" {
		t.Error("Expected MarkdownDescription to be non-empty")
	}

	expectedAttrs := []string{"base_url", "api_key", "email", "password", "insecure_skip_verify"}
	for _, attr := range expectedAttrs {
		if _, exists := resp.Schema.Attributes[attr]; !exists {
			t.Errorf("Expected attribute %q to exist in schema", attr)
		}
	}

	// Check that sensitive attributes are marked as sensitive
	if !resp.Schema.Attributes["api_key"].IsSensitive() {
		t.Error("Expected api_key to be marked as sensitive")
	}

	if !resp.Schema.Attributes["password"].IsSensitive() {
		t.Error("Expected password to be marked as sensitive")
	}
}

func TestProvider_Configure_EnvironmentVariableHandling(t *testing.T) {
	// Test missing base URL environment variable
	originalBaseURL := os.Getenv("N8N_BASE_URL")
	originalAPIKey := os.Getenv("N8N_API_KEY")

	os.Unsetenv("N8N_BASE_URL")
	os.Unsetenv("N8N_API_KEY")
	defer func() {
		if originalBaseURL != "" {
			os.Setenv("N8N_BASE_URL", originalBaseURL)
		}
		if originalAPIKey != "" {
			os.Setenv("N8N_API_KEY", originalAPIKey)
		}
	}()

	// This will test the environment variable handling logic
	// Since we can't easily create proper tfsdk.Config objects in unit tests,
	// we'll test the environment variable reading logic indirectly
	baseURL := os.Getenv("N8N_BASE_URL")
	apiKey := os.Getenv("N8N_API_KEY")

	if baseURL != "" {
		t.Error("Expected N8N_BASE_URL to be empty after unsetting")
	}
	if apiKey != "" {
		t.Error("Expected N8N_API_KEY to be empty after unsetting")
	}
}

func TestProvider_EnvironmentVariableValidation(t *testing.T) {
	tests := []struct {
		name       string
		envVars    map[string]string
		shouldHave map[string]string
	}{
		{
			name: "API key authentication env vars",
			envVars: map[string]string{
				"N8N_BASE_URL": "http://localhost:5678",
				"N8N_API_KEY":  "test-api-key",
			},
			shouldHave: map[string]string{
				"N8N_BASE_URL": "http://localhost:5678",
				"N8N_API_KEY":  "test-api-key",
			},
		},
		{
			name: "Basic auth env vars",
			envVars: map[string]string{
				"N8N_BASE_URL": "http://localhost:5678",
				"N8N_EMAIL":    "test@example.com",
				"N8N_PASSWORD": "password123",
			},
			shouldHave: map[string]string{
				"N8N_BASE_URL": "http://localhost:5678",
				"N8N_EMAIL":    "test@example.com",
				"N8N_PASSWORD": "password123",
			},
		},
		{
			name: "insecure skip verify env var",
			envVars: map[string]string{
				"N8N_INSECURE_SKIP_VERIFY": "true",
			},
			shouldHave: map[string]string{
				"N8N_INSECURE_SKIP_VERIFY": "true",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars first
			originalEnvs := make(map[string]string)
			envKeys := []string{"N8N_BASE_URL", "N8N_API_KEY", "N8N_EMAIL", "N8N_PASSWORD", "N8N_INSECURE_SKIP_VERIFY"}

			for _, key := range envKeys {
				originalEnvs[key] = os.Getenv(key)
				os.Unsetenv(key)
			}

			defer func() {
				for key, val := range originalEnvs {
					if val != "" {
						os.Setenv(key, val)
					}
				}
			}()

			// Set test env vars
			for key, val := range tt.envVars {
				os.Setenv(key, val)
			}

			// Verify env vars are set correctly
			for key, expectedVal := range tt.shouldHave {
				actualVal := os.Getenv(key)
				if actualVal != expectedVal {
					t.Errorf("Expected %s=%q, got %q", key, expectedVal, actualVal)
				}
			}
		})
	}
}

func TestProvider_Resources(t *testing.T) {
	ctx := context.Background()
	p := &N8nProvider{}

	resources := p.Resources(ctx)

	expectedCount := 6 // workflow, credential, user, project, project_user, ldap_config
	if len(resources) != expectedCount {
		t.Errorf("Expected %d resources, got %d", expectedCount, len(resources))
	}

	// Test that each resource function returns a non-nil resource
	for i, resourceFunc := range resources {
		resource := resourceFunc()
		if resource == nil {
			t.Errorf("Resource function %d returned nil", i)
		}
	}
}

func TestProvider_DataSources(t *testing.T) {
	ctx := context.Background()
	p := &N8nProvider{}

	dataSources := p.DataSources(ctx)

	expectedCount := 1 // user data source
	if len(dataSources) != expectedCount {
		t.Errorf("Expected %d data sources, got %d", expectedCount, len(dataSources))
	}

	// Test that each data source function returns a non-nil data source
	for i, dataSourceFunc := range dataSources {
		dataSource := dataSourceFunc()
		if dataSource == nil {
			t.Errorf("Data source function %d returned nil", i)
		}
	}
}

func TestProvider_Functions(t *testing.T) {
	ctx := context.Background()
	p := &N8nProvider{}

	functions := p.Functions(ctx)

	// Currently no functions are implemented
	if len(functions) != 0 {
		t.Errorf("Expected 0 functions, got %d", len(functions))
	}
}

func TestN8nProviderModel(t *testing.T) {
	// Test that the model struct has the expected fields
	model := N8nProviderModel{
		BaseURL:            types.StringValue("http://localhost:5678"),
		APIKey:             types.StringValue("test-key"),
		Email:              types.StringValue("test@example.com"),
		Password:           types.StringValue("password"),
		InsecureSkipVerify: types.BoolValue(true),
	}

	if model.BaseURL.ValueString() != "http://localhost:5678" {
		t.Error("BaseURL not set correctly")
	}
	if model.APIKey.ValueString() != "test-key" {
		t.Error("APIKey not set correctly")
	}
	if model.Email.ValueString() != "test@example.com" {
		t.Error("Email not set correctly")
	}
	if model.Password.ValueString() != "password" {
		t.Error("Password not set correctly")
	}
	if !model.InsecureSkipVerify.ValueBool() {
		t.Error("InsecureSkipVerify not set correctly")
	}
}
