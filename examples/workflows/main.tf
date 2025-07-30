terraform {
  required_providers {
    n8n = {
      source  = "devops247-online/n8n"
      version = "~> 1.0"
    }
  }
}

# Configure the n8n Provider
provider "n8n" {
  base_url = "http://localhost:5678" # Your n8n instance URL
  api_key  = var.n8n_api_key         # Set via environment variable or terraform.tfvars
}

# Example: Simple HTTP Request Workflow
resource "n8n_workflow" "simple_http" {
  name   = "Simple HTTP Request"
  active = false # Start inactive for testing

  nodes = jsonencode({
    "Start" = {
      "parameters"  = {},
      "type"        = "n8n-nodes-base.start",
      "typeVersion" = 1,
      "position"    = [240, 300]
    },
    "HTTP Request" = {
      "parameters" = {
        "url"            = "https://httpbin.org/json",
        "responseFormat" = "json"
      },
      "type"        = "n8n-nodes-base.httpRequest",
      "typeVersion" = 1,
      "position"    = [440, 300]
    }
  })

  connections = jsonencode({
    "Start" = {
      "main" = [
        [
          {
            "node"  = "HTTP Request",
            "type"  = "main",
            "index" = 0
          }
        ]
      ]
    }
  })

  tags = ["terraform", "http", "simple"]
}

# Example: Webhook Triggered Workflow
resource "n8n_workflow" "webhook_workflow" {
  name   = "Webhook Processing Workflow"
  active = true

  nodes = jsonencode({
    "Webhook" = {
      "parameters" = {
        "httpMethod"   = "POST",
        "path"         = "webhook-endpoint",
        "responseMode" = "responseNode"
      },
      "type"        = "n8n-nodes-base.webhook",
      "typeVersion" = 1,
      "position"    = [240, 300]
    },
    "Set Variables" = {
      "parameters" = {
        "assignments" = {
          "assignments" = [
            {
              "id"    = "processed_at",
              "name"  = "processed_at",
              "type"  = "string",
              "value" = "={{ new Date().toISOString() }}"
            },
            {
              "id"    = "status",
              "name"  = "status",
              "type"  = "string",
              "value" = "processed"
            }
          ]
        }
      },
      "type"        = "n8n-nodes-base.set",
      "typeVersion" = 3,
      "position"    = [440, 300]
    },
    "Respond to Webhook" = {
      "parameters" = {
        "responseBody" = "={{ { \"status\": \"success\", \"processed_at\": $json.processed_at } }}",
        "options"      = {}
      },
      "type"        = "n8n-nodes-base.respondToWebhook",
      "typeVersion" = 1,
      "position"    = [640, 300]
    }
  })

  connections = jsonencode({
    "Webhook" = {
      "main" = [
        [
          {
            "node"  = "Set Variables",
            "type"  = "main",
            "index" = 0
          }
        ]
      ]
    },
    "Set Variables" = {
      "main" = [
        [
          {
            "node"  = "Respond to Webhook",
            "type"  = "main",
            "index" = 0
          }
        ]
      ]
    }
  })

  settings = jsonencode({
    "saveManualExecutions" = true,
    "callerPolicy"         = "workflowsFromSameOwner"
  })

  tags = ["terraform", "webhook", "api"]
}

# Example: Scheduled Data Processing Workflow
resource "n8n_workflow" "scheduled_processing" {
  name   = "Daily Data Processing"
  active = false # Manual activation recommended

  nodes = jsonencode({
    "Schedule Trigger" = {
      "parameters" = {
        "rule" = {
          "interval" = [
            {
              "field" = "cronExpression",
              "value" = "0 9 * * *" # Daily at 9 AM
            }
          ]
        }
      },
      "type"        = "n8n-nodes-base.scheduleTrigger",
      "typeVersion" = 1,
      "position"    = [240, 300]
    },
    "Fetch Data" = {
      "parameters" = {
        "url"                = var.data_api_url,
        "authentication"     = "predefinedCredentialType",
        "nodeCredentialType" = "httpHeaderAuth",
        "responseFormat"     = "json"
      },
      "type"        = "n8n-nodes-base.httpRequest",
      "typeVersion" = 1,
      "position"    = [440, 300],
      "credentials" = {
        "httpHeaderAuth" = {
          "id"   = var.api_credential_id,
          "name" = "API Credentials"
        }
      }
    },
    "Process Data" = {
      "parameters" = {
        "functionCode" = "// Process the fetched data\nconst data = items[0].json;\nconst processedData = {\n  timestamp: new Date().toISOString(),\n  recordCount: Array.isArray(data) ? data.length : 1,\n  processed: true\n};\n\nreturn [{ json: processedData }];"
      },
      "type"        = "n8n-nodes-base.function",
      "typeVersion" = 1,
      "position"    = [640, 300]
    },
    "Save Results" = {
      "parameters" = {
        "url"                = var.results_api_url,
        "httpMethod"         = "POST",
        "authentication"     = "predefinedCredentialType",
        "nodeCredentialType" = "httpHeaderAuth",
        "responseFormat"     = "json",
        "sendBody"           = true,
        "bodyContentType"    = "json"
      },
      "type"        = "n8n-nodes-base.httpRequest",
      "typeVersion" = 1,
      "position"    = [840, 300],
      "credentials" = {
        "httpHeaderAuth" = {
          "id"   = var.api_credential_id,
          "name" = "API Credentials"
        }
      }
    }
  })

  connections = jsonencode({
    "Schedule Trigger" = {
      "main" = [
        [
          {
            "node"  = "Fetch Data",
            "type"  = "main",
            "index" = 0
          }
        ]
      ]
    },
    "Fetch Data" = {
      "main" = [
        [
          {
            "node"  = "Process Data",
            "type"  = "main",
            "index" = 0
          }
        ]
      ]
    },
    "Process Data" = {
      "main" = [
        [
          {
            "node"  = "Save Results",
            "type"  = "main",
            "index" = 0
          }
        ]
      ]
    }
  })

  settings = jsonencode({
    "executionOrder"       = "v1",
    "saveManualExecutions" = true,
    "callerPolicy"         = "workflowsFromSameOwner"
  })

  static_data = jsonencode({
    "node:Schedule Trigger" = {
      "recurrenceRules" = []
    }
  })

  tags = ["terraform", "scheduled", "data-processing"]
}