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
  base_url = var.n8n_base_url
  api_key  = var.n8n_api_key
}

# Create API credentials for external services
resource "n8n_credential" "external_api" {
  name = "External API Credentials"
  type = "httpHeaderAuth"

  data = jsonencode({
    name  = "Authorization"
    value = "Bearer ${var.external_api_token}"
  })
}

resource "n8n_credential" "slack_webhook" {
  name = "Slack Webhook"
  type = "httpHeaderAuth"

  data = jsonencode({
    name  = "Content-Type"
    value = "application/json"
  })
}

resource "n8n_credential" "database" {
  name = "Application Database"
  type = "postgres"

  data = jsonencode({
    host     = var.db_host
    database = var.db_name
    user     = var.db_username
    password = var.db_password
    port     = var.db_port
    ssl      = "require"
  })
}

# Main data processing workflow
resource "n8n_workflow" "data_processor" {
  depends_on = [n8n_credential.external_api, n8n_credential.database]

  name   = "Data Processing Pipeline"
  active = var.activate_workflows

  nodes = jsonencode({
    "Schedule Trigger" = {
      "parameters" = {
        "rule" = {
          "interval" = [
            {
              "field" = "cronExpression",
              "value" = var.processing_schedule
            }
          ]
        }
      },
      "type"        = "n8n-nodes-base.scheduleTrigger",
      "typeVersion" = 1,
      "position"    = [240, 300]
    },
    "Fetch External Data" = {
      "parameters" = {
        "url"                = var.external_api_url,
        "authentication"     = "predefinedCredentialType",
        "nodeCredentialType" = "httpHeaderAuth",
        "responseFormat"     = "json"
      },
      "type"        = "n8n-nodes-base.httpRequest",
      "typeVersion" = 1,
      "position"    = [440, 300],
      "credentials" = {
        "httpHeaderAuth" = {
          "id"   = n8n_credential.external_api.id,
          "name" = n8n_credential.external_api.name
        }
      }
    },
    "Transform Data" = {
      "parameters" = {
        "functionCode" = file("${path.module}/scripts/transform-data.js")
      },
      "type"        = "n8n-nodes-base.function",
      "typeVersion" = 1,
      "position"    = [640, 300]
    },
    "Store in Database" = {
      "parameters" = {
        "operation" = "insert",
        "table"     = var.db_table,
        "columns"   = "data, processed_at, source",
        "additionalFields" = {
          "mode" = "independently"
        }
      },
      "type"        = "n8n-nodes-base.postgres",
      "typeVersion" = 2,
      "position"    = [840, 200],
      "credentials" = {
        "postgres" = {
          "id"   = n8n_credential.database.id,
          "name" = n8n_credential.database.name
        }
      }
    },
    "Send Success Notification" = {
      "parameters" = {
        "url"             = var.slack_webhook_url,
        "httpMethod"      = "POST",
        "sendBody"        = true,
        "bodyContentType" = "json",
        "jsonBody"        = "={ \"text\": \"Data processing completed successfully. Processed \" + $json.count + \" records.\" }"
      },
      "type"        = "n8n-nodes-base.httpRequest",
      "typeVersion" = 1,
      "position"    = [1040, 200],
      "credentials" = {
        "httpHeaderAuth" = {
          "id"   = n8n_credential.slack_webhook.id,
          "name" = n8n_credential.slack_webhook.name
        }
      }
    },
    "Handle Errors" = {
      "parameters" = {
        "url"             = var.slack_webhook_url,
        "httpMethod"      = "POST",
        "sendBody"        = true,
        "bodyContentType" = "json",
        "jsonBody"        = "={ \"text\": \"❌ Data processing failed: \" + $json.error }"
      },
      "type"        = "n8n-nodes-base.httpRequest",
      "typeVersion" = 1,
      "position"    = [840, 400],
      "credentials" = {
        "httpHeaderAuth" = {
          "id"   = n8n_credential.slack_webhook.id,
          "name" = n8n_credential.slack_webhook.name
        }
      }
    }
  })

  connections = jsonencode({
    "Schedule Trigger" = {
      "main" = [
        [
          {
            "node"  = "Fetch External Data",
            "type"  = "main",
            "index" = 0
          }
        ]
      ]
    },
    "Fetch External Data" = {
      "main" = [
        [
          {
            "node"  = "Transform Data",
            "type"  = "main",
            "index" = 0
          }
        ]
      ]
    },
    "Transform Data" = {
      "main" = [
        [
          {
            "node"  = "Store in Database",
            "type"  = "main",
            "index" = 0
          }
        ]
      ]
    },
    "Store in Database" = {
      "main" = [
        [
          {
            "node"  = "Send Success Notification",
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
    "callerPolicy"         = "workflowsFromSameOwner",
    "errorWorkflow"        = n8n_workflow.error_handler.id
  })

  tags = ["terraform", "production", "data-processing", "automated"]
}

# Error handling workflow
resource "n8n_workflow" "error_handler" {
  depends_on = [n8n_credential.slack_webhook]

  name   = "Error Handler"
  active = var.activate_workflows

  nodes = jsonencode({
    "Error Trigger" = {
      "parameters"  = {},
      "type"        = "n8n-nodes-base.errorTrigger",
      "typeVersion" = 1,
      "position"    = [240, 300]
    },
    "Log Error" = {
      "parameters" = {
        "functionCode" = "console.error('Workflow Error:', JSON.stringify($json, null, 2));\nreturn items;"
      },
      "type"        = "n8n-nodes-base.function",
      "typeVersion" = 1,
      "position"    = [440, 300]
    },
    "Notify Team" = {
      "parameters" = {
        "url"             = var.slack_webhook_url,
        "httpMethod"      = "POST",
        "sendBody"        = true,
        "bodyContentType" = "json",
        "jsonBody"        = "={ \"text\": \"🚨 Workflow Error in \" + $json.workflow.name + \": \" + $json.error.message }"
      },
      "type"        = "n8n-nodes-base.httpRequest",
      "typeVersion" = 1,
      "position"    = [640, 300],
      "credentials" = {
        "httpHeaderAuth" = {
          "id"   = n8n_credential.slack_webhook.id,
          "name" = n8n_credential.slack_webhook.name
        }
      }
    }
  })

  connections = jsonencode({
    "Error Trigger" = {
      "main" = [
        [
          {
            "node"  = "Log Error",
            "type"  = "main",
            "index" = 0
          }
        ]
      ]
    },
    "Log Error" = {
      "main" = [
        [
          {
            "node"  = "Notify Team",
            "type"  = "main",
            "index" = 0
          }
        ]
      ]
    }
  })

  tags = ["terraform", "error-handling", "notifications"]
}

# API webhook for external integrations
resource "n8n_workflow" "api_webhook" {
  name   = "External API Webhook"
  active = var.activate_workflows

  nodes = jsonencode({
    "Webhook" = {
      "parameters" = {
        "httpMethod"     = "POST",
        "path"           = "api/webhook",
        "responseMode"   = "responseNode",
        "authentication" = "headerAuth",
        "options" = {
          "allowedOrigins" = var.allowed_origins
        }
      },
      "type"        = "n8n-nodes-base.webhook",
      "typeVersion" = 1,
      "position"    = [240, 300]
    },
    "Validate Input" = {
      "parameters" = {
        "functionCode" = file("${path.module}/scripts/validate-input.js")
      },
      "type"        = "n8n-nodes-base.function",
      "typeVersion" = 1,
      "position"    = [440, 300]
    },
    "Process Request" = {
      "parameters" = {
        "functionCode" = file("${path.module}/scripts/process-request.js")
      },
      "type"        = "n8n-nodes-base.function",
      "typeVersion" = 1,
      "position"    = [640, 200]
    },
    "Success Response" = {
      "parameters" = {
        "responseBody" = "={ { \"status\": \"success\", \"id\": $json.id, \"timestamp\": new Date().toISOString() } }",
        "responseCode" = 200,
        "options"      = {}
      },
      "type"        = "n8n-nodes-base.respondToWebhook",
      "typeVersion" = 1,
      "position"    = [840, 200]
    },
    "Error Response" = {
      "parameters" = {
        "responseBody" = "={ { \"status\": \"error\", \"message\": $json.error || \"Invalid request\" } }",
        "responseCode" = 400,
        "options"      = {}
      },
      "type"        = "n8n-nodes-base.respondToWebhook",
      "typeVersion" = 1,
      "position"    = [640, 400]
    }
  })

  connections = jsonencode({
    "Webhook" = {
      "main" = [
        [
          {
            "node"  = "Validate Input",
            "type"  = "main",
            "index" = 0
          }
        ]
      ]
    },
    "Validate Input" = {
      "main" = [
        [
          {
            "node"  = "Process Request",
            "type"  = "main",
            "index" = 0
          }
        ],
        [
          {
            "node"  = "Error Response",
            "type"  = "main",
            "index" = 0
          }
        ]
      ]
    },
    "Process Request" = {
      "main" = [
        [
          {
            "node"  = "Success Response",
            "type"  = "main",
            "index" = 0
          }
        ]
      ]
    }
  })

  tags = ["terraform", "api", "webhook", "integration"]
}