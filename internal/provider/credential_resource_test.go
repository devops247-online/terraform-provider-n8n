package provider

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCredentialResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCredentials(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCredentialResourceConfig("test-credential", "httpBasicAuth"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_credential.test", "name", "test-credential"),
					resource.TestCheckResourceAttr("n8n_credential.test", "type", "httpBasicAuth"),
					resource.TestCheckResourceAttrSet("n8n_credential.test", "id"),
					resource.TestCheckResourceAttrSet("n8n_credential.test", "created_at"),
					resource.TestCheckResourceAttrSet("n8n_credential.test", "updated_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "n8n_credential.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Skip verifying sensitive data field
				ImportStateVerifyIgnore: []string{"data"},
			},
			// Update and Read testing
			{
				Config: testAccCredentialResourceConfig("test-credential-updated", "httpBasicAuth"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_credential.test", "name", "test-credential-updated"),
					resource.TestCheckResourceAttr("n8n_credential.test", "type", "httpBasicAuth"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccCredentialResourceWithData(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCredentials(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with credential data
			{
				Config: testAccCredentialResourceConfigWithData("test-credential-data", "httpBasicAuth"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_credential.test", "name", "test-credential-data"),
					resource.TestCheckResourceAttr("n8n_credential.test", "type", "httpBasicAuth"),
					resource.TestCheckResourceAttrSet("n8n_credential.test", "data"),
				),
			},
			// Update credential data
			{
				Config: testAccCredentialResourceConfigWithUpdatedData("test-credential-data", "httpBasicAuth"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_credential.test", "name", "test-credential-data"),
					resource.TestCheckResourceAttr("n8n_credential.test", "type", "httpBasicAuth"),
					resource.TestCheckResourceAttrSet("n8n_credential.test", "data"),
				),
			},
		},
	})
}

func TestAccCredentialResourceWithNodeAccess(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCredentials(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with node access
			{
				Config: testAccCredentialResourceConfigWithNodeAccess("test-credential-nodes", "apiKey"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_credential.test", "name", "test-credential-nodes"),
					resource.TestCheckResourceAttr("n8n_credential.test", "type", "apiKey"),
					resource.TestCheckResourceAttr("n8n_credential.test", "node_access.#", "2"),
					resource.TestCheckTypeSetElemAttr("n8n_credential.test", "node_access.*", "httpRequest"),
					resource.TestCheckTypeSetElemAttr("n8n_credential.test", "node_access.*", "webhook"),
				),
			},
		},
	})
}

func TestAccCredentialResourceOAuth2(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCredentials(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create OAuth2 credential
			{
				Config: testAccCredentialResourceConfigOAuth2("test-oauth2-credential"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_credential.test", "name", "test-oauth2-credential"),
					resource.TestCheckResourceAttr("n8n_credential.test", "type", "oAuth2Api"),
					resource.TestCheckResourceAttrSet("n8n_credential.test", "data"),
				),
			},
		},
	})
}

func TestAccCredentialResourceAPIKey(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCredentials(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create API key credential
			{
				Config: testAccCredentialResourceConfigAPIKey("test-apikey-credential"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_credential.test", "name", "test-apikey-credential"),
					resource.TestCheckResourceAttr("n8n_credential.test", "type", "apiKey"),
					resource.TestCheckResourceAttrSet("n8n_credential.test", "data"),
				),
			},
		},
	})
}

func TestAccCredentialResourceBearerToken(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCredentials(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create bearer token credential
			{
				Config: testAccCredentialResourceConfigBearerToken("test-bearer-credential"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_credential.test", "name", "test-bearer-credential"),
					resource.TestCheckResourceAttr("n8n_credential.test", "type", "bearerTokenAuth"),
					resource.TestCheckResourceAttrSet("n8n_credential.test", "data"),
				),
			},
		},
	})
}

func TestAccCredentialResourceAWS(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCredentials(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create AWS credential
			{
				Config: testAccCredentialResourceConfigAWS("test-aws-credential"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_credential.test", "name", "test-aws-credential"),
					resource.TestCheckResourceAttr("n8n_credential.test", "type", "awsApi"),
					resource.TestCheckResourceAttrSet("n8n_credential.test", "data"),
				),
			},
		},
	})
}

func TestAccCredentialResourceInvalidType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCredentials(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test invalid credential type
			{
				Config:      testAccCredentialResourceConfig("test-invalid", "invalidType"),
				ExpectError: regexp.MustCompile("Invalid Credential Type"),
			},
		},
	})
}

func TestAccCredentialResourceInvalidData(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCredentials(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test invalid JSON data
			{
				Config:      testAccCredentialResourceConfigInvalidJSON("test-invalid-json"),
				ExpectError: regexp.MustCompile("Invalid JSON"),
			},
			// Test missing required fields for httpBasicAuth
			{
				Config:      testAccCredentialResourceConfigMissingUser("test-missing-user"),
				ExpectError: regexp.MustCompile("httpBasicAuth credential requires 'user' field"),
			},
			// Test missing required fields for apiKey
			{
				Config:      testAccCredentialResourceConfigMissingAPIKey("test-missing-apikey"),
				ExpectError: regexp.MustCompile("apiKey credential requires 'apiKey' field"),
			},
			// Test missing required fields for oAuth2Api
			{
				Config:      testAccCredentialResourceConfigMissingOAuth2Fields("test-missing-oauth2"),
				ExpectError: regexp.MustCompile("oAuth2Api credential requires 'clientId' field"),
			},
		},
	})
}

func TestAccCredentialResourceTypeRequiresReplace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCredentials(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with httpBasicAuth
			{
				Config: testAccCredentialResourceConfig("test-type-change", "httpBasicAuth"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_credential.test", "type", "httpBasicAuth"),
				),
			},
			// Change type (should require replace)
			{
				Config: testAccCredentialResourceConfig("test-type-change", "apiKey"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_credential.test", "type", "apiKey"),
				),
			},
		},
	})
}

// Helper functions for test configurations

func testAccCredentialResourceConfig(name, credType string) string {
	return fmt.Sprintf(`
resource "n8n_credential" "test" {
  name = "%s"
  type = "%s"
}
`, name, credType)
}

func testAccCredentialResourceConfigWithData(name, credType string) string {
	return fmt.Sprintf(`
resource "n8n_credential" "test" {
  name = "%s"
  type = "%s"
  data = jsonencode({
    user     = "testuser"
    password = "testpass"
  })
}
`, name, credType)
}

func testAccCredentialResourceConfigWithUpdatedData(name, credType string) string {
	return fmt.Sprintf(`
resource "n8n_credential" "test" {
  name = "%s"
  type = "%s"
  data = jsonencode({
    user     = "updateduser"
    password = "updatedpass"
  })
}
`, name, credType)
}

func testAccCredentialResourceConfigWithNodeAccess(name, credType string) string {
	return fmt.Sprintf(`
resource "n8n_credential" "test" {
  name = "%s"
  type = "%s"
  data = jsonencode({
    apiKey = "test-api-key-123"
  })
  node_access = ["httpRequest", "webhook"]
}
`, name, credType)
}

func testAccCredentialResourceConfigOAuth2(name string) string {
	return fmt.Sprintf(`
resource "n8n_credential" "test" {
  name = "%s"
  type = "oAuth2Api"
  data = jsonencode({
    clientId         = "test-client-id"
    clientSecret     = "test-client-secret"
    accessTokenUrl   = "https://example.com/oauth/token"
    authUrl          = "https://example.com/oauth/authorize"
    scope            = "read write"
    authentication  = "body"
  })
}
`, name)
}

func testAccCredentialResourceConfigAPIKey(name string) string {
	return fmt.Sprintf(`
resource "n8n_credential" "test" {
  name = "%s"
  type = "apiKey"
  data = jsonencode({
    apiKey = "sk-test-api-key-12345678901234567890"
  })
}
`, name)
}

func testAccCredentialResourceConfigBearerToken(name string) string {
	return fmt.Sprintf(`
resource "n8n_credential" "test" {
  name = "%s"
  type = "bearerTokenAuth"
  data = jsonencode({
    token = "bearer-token-abcdef123456"
  })
}
`, name)
}

func testAccCredentialResourceConfigAWS(name string) string {
	return fmt.Sprintf(`
resource "n8n_credential" "test" {
  name = "%s"
  type = "awsApi"
  data = jsonencode({
    accessKeyId     = "AKIAIOSFODNN7EXAMPLE"
    secretAccessKey = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
    region          = "us-east-1"
  })
}
`, name)
}

func testAccCredentialResourceConfigInvalidJSON(name string) string {
	return fmt.Sprintf(`
resource "n8n_credential" "test" {
  name = "%s"
  type = "httpBasicAuth"
  data = "invalid json"
}
`, name)
}

func testAccCredentialResourceConfigMissingUser(name string) string {
	return fmt.Sprintf(`
resource "n8n_credential" "test" {
  name = "%s"
  type = "httpBasicAuth"
  data = jsonencode({
    password = "testpass"
  })
}
`, name)
}

func testAccCredentialResourceConfigMissingAPIKey(name string) string {
	return fmt.Sprintf(`
resource "n8n_credential" "test" {
  name = "%s"
  type = "apiKey"
  data = jsonencode({
    name = "Authorization"
  })
}
`, name)
}

func testAccCredentialResourceConfigMissingOAuth2Fields(name string) string {
	return fmt.Sprintf(`
resource "n8n_credential" "test" {
  name = "%s"
  type = "oAuth2Api"
  data = jsonencode({
    clientSecret = "test-client-secret"
  })
}
`, name)
}

// testAccPreCheckCredentials validates that credentials API is available
// Skips the test if credentials API is not accessible
func testAccPreCheckCredentials(t *testing.T) {
	testAccPreCheck(t) // First check basic requirements

	// Check if credentials API is available by testing list endpoint
	baseURL := os.Getenv("N8N_BASE_URL")
	apiKey := os.Getenv("N8N_API_KEY")

	if baseURL == "" || apiKey == "" {
		t.Skip("Skipping credential test: N8N_BASE_URL or N8N_API_KEY not set")
		return
	}

	// Test if credentials list endpoint is accessible
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", baseURL+"/api/v1/credentials", nil)
	if err != nil {
		t.Skip("Skipping credential test: Unable to create request")
		return
	}

	req.Header.Set("X-N8N-API-KEY", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		t.Skip("Skipping credential test: Unable to connect to n8n API")
		return
	}
	defer resp.Body.Close()

	// If we get 404 or 405, credentials API is not available
	if resp.StatusCode == 404 || resp.StatusCode == 405 {
		t.Skip("Skipping credential test: Credentials API not available (possibly not supported in this n8n version)")
		return
	}

	// If we get other errors (500, etc.), still skip but log it
	if resp.StatusCode >= 400 {
		t.Skipf("Skipping credential test: API returned status %d", resp.StatusCode)
		return
	}

	// If we reach here, credentials API appears to be available
}
