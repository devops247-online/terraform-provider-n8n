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

func TestAccWorkflowResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccWorkflowResourceConfig("test-workflow"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_workflow.test", "name", "test-workflow"),
					resource.TestCheckResourceAttr("n8n_workflow.test", "active", "false"),
					resource.TestCheckResourceAttrSet("n8n_workflow.test", "id"),
					resource.TestCheckResourceAttrSet("n8n_workflow.test", "created_at"),
					resource.TestCheckResourceAttrSet("n8n_workflow.test", "updated_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "n8n_workflow.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccWorkflowResourceConfig("test-workflow-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_workflow.test", "name", "test-workflow-updated"),
					resource.TestCheckResourceAttr("n8n_workflow.test", "active", "false"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccWorkflowResourceWithNodes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with nodes
			{
				Config: testAccWorkflowResourceConfigWithNodes("test-workflow-nodes"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_workflow.test", "name", "test-workflow-nodes"),
					resource.TestCheckResourceAttr("n8n_workflow.test", "active", "false"),
					resource.TestCheckResourceAttrSet("n8n_workflow.test", "nodes"),
					resource.TestCheckResourceAttrSet("n8n_workflow.test", "connections"),
				),
			},
			// Update nodes
			{
				Config: testAccWorkflowResourceConfigWithUpdatedNodes("test-workflow-nodes"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_workflow.test", "name", "test-workflow-nodes"),
					resource.TestCheckResourceAttr("n8n_workflow.test", "active", "false"),
					resource.TestCheckResourceAttrSet("n8n_workflow.test", "nodes"),
				),
			},
		},
	})
}

func TestAccWorkflowResourceWithTags(t *testing.T) {
	t.Skip("Tags are read-only in n8n API - skipping until proper tag management is implemented")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with tags
			{
				Config: testAccWorkflowResourceConfigWithTags("test-workflow-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_workflow.test", "name", "test-workflow-tags"),
					resource.TestCheckResourceAttr("n8n_workflow.test", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr("n8n_workflow.test", "tags.*", "automation"),
					resource.TestCheckTypeSetElemAttr("n8n_workflow.test", "tags.*", "test"),
				),
			},
		},
	})
}

func TestAccWorkflowResourceInvalidJSON(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test invalid nodes JSON
			{
				Config:      testAccWorkflowResourceConfigInvalidNodesJSON("test-workflow-invalid"),
				ExpectError: regexp.MustCompile("Invalid Nodes JSON"),
			},
			// Test invalid connections JSON
			{
				Config:      testAccWorkflowResourceConfigInvalidConnectionsJSON("test-workflow-invalid"),
				ExpectError: regexp.MustCompile("Invalid Connections JSON"),
			},
		},
	})
}

func TestAccWorkflowResourceLargeWorkflow(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create large workflow
			{
				Config: testAccWorkflowResourceConfigLarge("test-workflow-large"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("n8n_workflow.test", "name", "test-workflow-large"),
					resource.TestCheckResourceAttrSet("n8n_workflow.test", "nodes"),
					resource.TestCheckResourceAttrSet("n8n_workflow.test", "connections"),
					resource.TestCheckResourceAttrSet("n8n_workflow.test", "settings"),
				),
			},
		},
	})
}

// testAccPreCheck validates the necessary test API credentials exist
func testAccPreCheck(t *testing.T) {
	// Check for required environment variables
	if v := os.Getenv("N8N_BASE_URL"); v == "" {
		t.Fatal("N8N_BASE_URL must be set for acceptance tests")
	}
	if v := os.Getenv("N8N_API_KEY"); v == "" {
		if email := os.Getenv("N8N_EMAIL"); email == "" {
			t.Fatal("N8N_API_KEY or N8N_EMAIL/N8N_PASSWORD must be set for acceptance tests")
		}
		if password := os.Getenv("N8N_PASSWORD"); password == "" {
			t.Fatal("N8N_PASSWORD must be set when using N8N_EMAIL for acceptance tests")
		}
	}
}

// testAccPreCheckEnterprise validates that Enterprise features are available
// Skips the test if Enterprise APIs are not accessible (e.g., community n8n version)
func testAccPreCheckEnterprise(t *testing.T) {
	testAccPreCheck(t) // First check basic requirements

	// Check if Enterprise features are available by testing a lightweight endpoint
	// If not available, skip the test gracefully
	baseURL := os.Getenv("N8N_BASE_URL")
	apiKey := os.Getenv("N8N_API_KEY")

	if baseURL == "" || apiKey == "" {
		t.Skip("Skipping Enterprise test: N8N_BASE_URL or N8N_API_KEY not set")
		return
	}

	// Test if projects endpoint is accessible (lightweight check for Enterprise features)
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", baseURL+"/api/v1/projects", nil)
	if err != nil {
		t.Skip("Skipping Enterprise test: Unable to create request")
		return
	}

	req.Header.Set("X-N8N-API-KEY", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		t.Skip("Skipping Enterprise test: Unable to connect to n8n API")
		return
	}
	defer resp.Body.Close()

	// If we get 404 or 403, Enterprise features are not available
	if resp.StatusCode == 404 || resp.StatusCode == 403 {
		t.Skip("Skipping Enterprise test: Projects API not available (likely community n8n version)")
		return
	}

	// If we get other errors (500, etc.), still skip but log it
	if resp.StatusCode >= 400 {
		t.Skipf("Skipping Enterprise test: API returned status %d", resp.StatusCode)
		return
	}

	// If we reach here, Enterprise features appear to be available
}

func testAccWorkflowResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "n8n_workflow" "test" {
  name   = "%s"
  active = false
  
  nodes = jsonencode({
    "start": {
      "type": "n8n-nodes-base.start",
      "position": [240, 300],
      "parameters": {}
    }
  })
  
  connections = jsonencode({})
  
  settings = jsonencode({
    "executionOrder": "v1"
  })
}
`, name)
}

func testAccWorkflowResourceConfigWithNodes(name string) string {
	return fmt.Sprintf(`
resource "n8n_workflow" "test" {
  name   = "%s"
  active = false
  
  nodes = jsonencode({
    "start": {
      "type": "n8n-nodes-base.start",
      "position": [240, 300],
      "parameters": {}
    },
    "webhook": {
      "type": "n8n-nodes-base.webhook",
      "position": [460, 300],
      "parameters": {
        "path": "test-webhook",
        "httpMethod": "GET"
      }
    }
  })
  
  connections = jsonencode({
    "start": {
      "main": [
        [
          {
            "node": "webhook",
            "type": "main",
            "index": 0
          }
        ]
      ]
    }
  })
  
  settings = jsonencode({
    "executionOrder": "v1"
  })
}
`, name)
}

func testAccWorkflowResourceConfigWithUpdatedNodes(name string) string {
	return fmt.Sprintf(`
resource "n8n_workflow" "test" {
  name   = "%s"
  active = false
  
  nodes = jsonencode({
    "start": {
      "type": "n8n-nodes-base.start",
      "position": [240, 300],
      "parameters": {}
    },
    "webhook": {
      "type": "n8n-nodes-base.webhook",
      "position": [460, 300],
      "parameters": {
        "path": "updated-webhook",
        "httpMethod": "POST"
      }
    },
    "http": {
      "type": "n8n-nodes-base.httpRequest",
      "position": [680, 300],
      "parameters": {
        "url": "https://httpbin.org/post",
        "method": "POST"
      }
    }
  })
  
  connections = jsonencode({
    "start": {
      "main": [
        [
          {
            "node": "webhook",
            "type": "main",
            "index": 0
          }
        ]
      ]
    },
    "webhook": {
      "main": [
        [
          {
            "node": "http", 
            "type": "main",
            "index": 0
          }
        ]
      ]
    }
  })
  
  settings = jsonencode({
    "executionOrder": "v1"
  })
}
`, name)
}

func testAccWorkflowResourceConfigWithTags(name string) string {
	return fmt.Sprintf(`
resource "n8n_workflow" "test" {
  name   = "%s"
  active = false
  tags   = ["automation", "test"]
  
  nodes = jsonencode({
    "start": {
      "type": "n8n-nodes-base.start",
      "position": [240, 300],
      "parameters": {}
    }
  })
  
  connections = jsonencode({})
  
  settings = jsonencode({
    "executionOrder": "v1"
  })
}
`, name)
}

func testAccWorkflowResourceConfigInvalidNodesJSON(name string) string {
	return fmt.Sprintf(`
resource "n8n_workflow" "test" {
  name  = "%s"
  nodes = "invalid json"
}
`, name)
}

func testAccWorkflowResourceConfigInvalidConnectionsJSON(name string) string {
	return fmt.Sprintf(`
resource "n8n_workflow" "test" {
  name        = "%s"
  connections = "invalid json"
}
`, name)
}

func testAccWorkflowResourceConfigLarge(name string) string {
	return fmt.Sprintf(`
resource "n8n_workflow" "test" {
  name   = "%s"
  active = false
  
  nodes = jsonencode({
    "start": {
      "type": "n8n-nodes-base.start",
      "position": [240, 300],
      "parameters": {}
    },
    "webhook1": {
      "type": "n8n-nodes-base.webhook",
      "position": [460, 200],
      "parameters": {
        "path": "webhook1",
        "httpMethod": "GET"
      }
    },
    "webhook2": {
      "type": "n8n-nodes-base.webhook", 
      "position": [460, 400],
      "parameters": {
        "path": "webhook2",
        "httpMethod": "POST"
      }
    },
    "merge": {
      "type": "n8n-nodes-base.merge",
      "position": [680, 300],
      "parameters": {
        "mode": "append"
      }
    },
    "http": {
      "type": "n8n-nodes-base.httpRequest",
      "position": [900, 300],
      "parameters": {
        "url": "https://httpbin.org/post",
        "method": "POST"
      }
    }
  })
  
  connections = jsonencode({
    "start": {
      "main": [
        [
          {
            "node": "webhook1",
            "type": "main",
            "index": 0
          },
          {
            "node": "webhook2",
            "type": "main", 
            "index": 0
          }
        ]
      ]
    },
    "webhook1": {
      "main": [
        [
          {
            "node": "merge",
            "type": "main",
            "index": 0
          }
        ]
      ]
    },
    "webhook2": {
      "main": [
        [
          {
            "node": "merge",
            "type": "main",
            "index": 1
          }
        ]
      ]
    },
    "merge": {
      "main": [
        [
          {
            "node": "http",
            "type": "main",
            "index": 0
          }
        ]
      ]
    }
  })
  
  settings = jsonencode({
    "executionOrder": "v1"
  })
}
`, name)
}
