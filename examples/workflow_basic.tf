# Example: Basic n8n workflow
terraform {
  required_providers {
    n8n = {
      source  = "devops247-online/n8n"
      version = "~> 1.0"
    }
  }
}

provider "n8n" {
  base_url = "http://localhost:5678" # or use N8N_BASE_URL env var
  api_key  = "your-api-key"          # or use N8N_API_KEY env var
}

# Simple workflow with just a name
resource "n8n_workflow" "simple" {
  name   = "Simple Workflow"
  active = false
}

# Workflow with nodes and connections
resource "n8n_workflow" "webhook_example" {
  name   = "Webhook to HTTP Request"
  active = true
  tags   = ["automation", "webhook"]

  nodes = jsonencode({
    "webhook" : {
      "type" : "n8n-nodes-base.webhook",
      "position" : [240, 300],
      "parameters" : {
        "path" : "example-webhook",
        "httpMethod" : "POST"
      }
    },
    "http" : {
      "type" : "n8n-nodes-base.httpRequest",
      "position" : [460, 300],
      "parameters" : {
        "url" : "https://httpbin.org/post",
        "method" : "POST"
      }
    }
  })

  connections = jsonencode({
    "webhook" : {
      "main" : [
        [
          {
            "node" : "http",
            "type" : "main",
            "index" : 0
          }
        ]
      ]
    }
  })

  settings = jsonencode({
    "executionOrder" : "v1",
    "saveManualExecutions" : true
  })
}

# Output the workflow ID
output "simple_workflow_id" {
  value = n8n_workflow.simple.id
}

output "webhook_workflow_id" {
  value = n8n_workflow.webhook_example.id
}