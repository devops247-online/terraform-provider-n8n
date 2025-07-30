terraform {
  required_providers {
    n8n = {
      source = "devops247-online/n8n"
      version = "~> 1.0"
    }
  }
}

# Configure the n8n Provider
provider "n8n" {
  base_url = "http://localhost:5678"  # Your n8n instance URL
  api_key  = var.n8n_api_key         # Set via environment variable or terraform.tfvars
}

# Example: Create a simple HTTP request workflow
resource "n8n_workflow" "example" {
  name   = "Example HTTP Workflow"
  active = true
  
  nodes = jsonencode({
    "Start": {
      "parameters": {},
      "type": "n8n-nodes-base.start",
      "typeVersion": 1,
      "position": [240, 300]
    },
    "HTTP Request": {
      "parameters": {
        "url": "https://httpbin.org/json",
        "responseFormat": "json"
      },
      "type": "n8n-nodes-base.httpRequest",
      "typeVersion": 1,
      "position": [440, 300]
    }
  })
  
  connections = jsonencode({
    "Start": {
      "main": [
        [
          {
            "node": "HTTP Request",
            "type": "main",
            "index": 0
          }
        ]
      ]
    }
  })
  
  tags = ["terraform", "example", "http"]
}