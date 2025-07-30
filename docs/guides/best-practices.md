# Best Practices for n8n Terraform Provider

This guide outlines best practices for managing n8n infrastructure with Terraform, covering code organization, security, testing, and operational considerations.

## Code Organization

### Project Structure

Organize your Terraform code for maintainability and clarity:

```
n8n-infrastructure/
├── environments/
│   ├── dev/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── terraform.tfvars
│   │   └── outputs.tf
│   ├── staging/
│   └── prod/
├── modules/
│   ├── workflows/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   └── credentials/
│       ├── main.tf
│       ├── variables.tf
│       └── outputs.tf
├── shared/
│   ├── data-sources.tf
│   └── locals.tf
└── docs/
    └── README.md
```

### Use Modules for Reusability

Create reusable modules for common patterns:

```hcl
# modules/api-workflow/main.tf
resource "n8n_credential" "api_key" {
  name = "${var.service_name} API Key"
  type = "httpHeaderAuth"
  
  data = jsonencode({
    name  = var.auth_header_name
    value = var.api_key_value
  })
}

resource "n8n_workflow" "api_processor" {
  depends_on = [n8n_credential.api_key]
  
  name   = "${var.service_name} Processor"
  active = var.activate_workflow
  
  nodes       = templatefile("${path.module}/templates/api-workflow.json", {
    credential_id   = n8n_credential.api_key.id
    credential_name = n8n_credential.api_key.name
    api_url        = var.api_url
  })
  
  connections = file("${path.module}/templates/api-connections.json")
  
  tags = concat(var.common_tags, ["api", var.service_name])
}
```

Usage:
```hcl
module "github_integration" {
  source = "./modules/api-workflow"
  
  service_name      = "github"
  api_url          = "https://api.github.com"
  auth_header_name = "Authorization"
  api_key_value    = "token ${var.github_token}"
  activate_workflow = var.environment == "prod"
  
  common_tags = local.common_tags
}
```

### Environment-Specific Configuration

Use locals and variables for environment differences:

```hcl
# environments/prod/main.tf
locals {
  environment = "prod"
  
  common_tags = [
    "environment:${local.environment}",
    "managed-by:terraform",
    "team:automation"
  ]
  
  # Environment-specific settings
  workflow_settings = {
    save_manual_executions = true
    execution_timeout      = 300
    retry_on_fail         = 3
  }
  
  # Resource naming
  name_prefix = "${local.environment}-"
}
```

## Security Best Practices

### Credential Management

1. **Never commit secrets to version control:**

```hcl
# ❌ Bad
resource "n8n_credential" "bad" {
  name = "API Key"
  type = "httpHeaderAuth"
  data = jsonencode({
    name  = "X-API-Key"
    value = "sk_live_abc123"  # Never hardcode!
  })
}

# ✅ Good
resource "n8n_credential" "good" {
  name = "API Key"
  type = "httpHeaderAuth"
  data = jsonencode({
    name  = "X-API-Key"
    value = var.api_key  # Use variables
  })
}
```

2. **Use proper variable declarations:**

```hcl
variable "api_key" {
  description = "API key for external service"
  type        = string
  sensitive   = true
  
  validation {
    condition     = length(var.api_key) > 0
    error_message = "API key cannot be empty."
  }
}
```

3. **Use external secret management:**

```hcl
# Use AWS Secrets Manager
data "aws_secretsmanager_secret_version" "api_key" {
  secret_id = "n8n/api-keys/external-service"
}

locals {
  api_key = jsondecode(data.aws_secretsmanager_secret_version.api_key.secret_string)["api_key"]
}
```

### Access Control

```hcl
# Restrict credential access to specific workflows/nodes
resource "n8n_credential" "restricted" {
  name = "Restricted API Access"
  type = "httpHeaderAuth"
  
  data = jsonencode({
    name  = "Authorization"
    value = "Bearer ${var.restricted_token}"
  })
  
  # Only allow specific node types to use this credential
  node_access = ["httpRequest", "webhook"]
}
```

## Workflow Design Patterns

### Modular Workflow Design

Break complex workflows into smaller, manageable pieces:

```hcl
# Data ingestion workflow
resource "n8n_workflow" "data_ingestion" {
  name = "Data Ingestion"
  
  # Focus on one responsibility
  nodes = jsonencode({
    "Schedule Trigger" = { /* ... */ },
    "Fetch Data"      = { /* ... */ },
    "Validate Data"   = { /* ... */ },
    "Store Raw Data"  = { /* ... */ }
  })
  
  tags = ["data", "ingestion", "etl-step-1"]
}

# Data processing workflow (triggered by ingestion)
resource "n8n_workflow" "data_processing" {
  name = "Data Processing"
  
  nodes = jsonencode({
    "Webhook Trigger"  = { /* ... */ },
    "Transform Data"   = { /* ... */ },
    "Apply Rules"      = { /* ... */ },
    "Store Processed"  = { /* ... */ }
  })
  
  tags = ["data", "processing", "etl-step-2"]
}
```

### Error Handling Patterns

Implement comprehensive error handling:

```hcl
resource "n8n_workflow" "robust_workflow" {
  name = "Robust Workflow with Error Handling"
  
  settings = jsonencode({
    errorWorkflow         = n8n_workflow.error_handler.id
    saveManualExecutions = true
    executionTimeout     = var.execution_timeout
  })
  
  # Include error paths in your workflow design
  nodes = jsonencode({
    "Main Process" = {
      # Main processing logic
      "onError" = "continueErrorOutput"
    },
    "Error Handler" = {
      # Handle specific errors
      "parameters" = {
        "functionCode" = file("${path.module}/scripts/error-handler.js")
      }
    },
    "Notify Team" = {
      # Send notifications on error
    }
  })
}

# Dedicated error handling workflow
resource "n8n_workflow" "error_handler" {
  name = "Global Error Handler"
  
  nodes = jsonencode({
    "Error Trigger" = {
      "type" = "n8n-nodes-base.errorTrigger"
    },
    "Log Error" = {
      # Structured error logging
    },
    "Create Incident" = {
      # Automated incident creation
    }
  })
}
```

## Testing Strategies

### Testing Pyramid for n8n Workflows

1. **Unit Tests**: Test individual node logic
2. **Integration Tests**: Test workflow execution
3. **End-to-End Tests**: Test complete automation flows

### Test Workflow Pattern

```hcl
resource "n8n_workflow" "test_workflow" {
  count = var.create_test_workflows ? 1 : 0
  
  name   = "TEST - ${var.workflow_name}"
  active = false  # Never activate test workflows
  
  # Use pinned data for consistent testing
  pinned_data = jsonencode({
    "Start" = [
      {
        "json" = {
          "test_data" = "sample input",
          "timestamp" = "2024-01-01T00:00:00Z"
        }
      }
    ]
  })
  
  # Copy of production workflow with test modifications
  nodes = local.test_workflow_nodes
  
  tags = ["test", "validation", var.workflow_name]
}
```

### Validation Scripts

Create validation scripts for your Terraform configurations:

```bash
#!/bin/bash
# scripts/validate.sh

set -e

echo "Validating Terraform configuration..."
terraform validate

echo "Checking formatting..."
terraform fmt -check -recursive

echo "Running security scan..."
tfsec .

echo "Validating workflow JSON..."
for file in workflows/*.json; do
  if ! jq empty "$file" 2>/dev/null; then
    echo "Invalid JSON in $file"
    exit 1
  fi
done

echo "All validations passed!"
```

## Operational Excellence

### Monitoring and Observability

```hcl
# Create monitoring workflows
resource "n8n_workflow" "health_check" {
  name   = "System Health Check"
  active = true
  
  nodes = jsonencode({
    "Schedule" = {
      "parameters" = {
        "rule" = {
          "interval" = [{"field" = "cronExpression", "value" = "*/5 * * * *"}]  # Every 5 minutes
        }
      },
      "type" = "n8n-nodes-base.scheduleTrigger"
    },
    "Check APIs" = {
      # Health check logic
    },
    "Alert on Failure" = {
      # Alerting logic
    }
  })
  
  tags = ["monitoring", "health-check", "ops"]
}
```

### Backup and Disaster Recovery

```hcl
# Backup workflow configurations
resource "n8n_workflow" "backup_workflows" {
  name = "Workflow Backup"
  
  nodes = jsonencode({
    "Daily Trigger" = {
      "parameters" = {
        "rule" = {
          "interval" = [{"field" = "cronExpression", "value" = "0 2 * * *"}]  # Daily at 2 AM
        }
      }
    },
    "Export Workflows" = {
      # Export all workflows to external storage
    },
    "Upload to S3" = {
      # Store backups in S3 with versioning
    }
  })
  
  tags = ["backup", "disaster-recovery", "ops"]
}
```

### Resource Tagging Strategy

Implement consistent tagging:

```hcl
locals {
  common_tags = [
    "environment:${var.environment}",
    "team:${var.team_name}",
    "project:${var.project_name}",
    "cost-center:${var.cost_center}",
    "managed-by:terraform",
    "created-date:${formatdate("YYYY-MM-DD", timestamp())}"
  ]
}

resource "n8n_workflow" "example" {
  name = "Example Workflow"
  tags = concat(
    local.common_tags,
    [
      "workflow-type:data-processing",
      "criticality:high",
      "schedule:daily"
    ]
  )
}
```

## Performance Optimization

### Workflow Efficiency

```hcl
resource "n8n_workflow" "optimized_workflow" {
  name = "Optimized Data Processing"
  
  settings = jsonencode({
    executionOrder        = "v1"           # Use latest execution order
    saveManualExecutions = false          # Disable if not needed
    callerPolicy         = "workflowsFromSameOwner"  # Restrict execution
    executionTimeout     = 300            # Set appropriate timeout
  })
  
  # Optimize node configuration
  nodes = jsonencode({
    "Batch Processor" = {
      "parameters" = {
        "batchSize" = 100                 # Process in batches
        "options" = {
          "continueOnFail" = true         # Don't stop on single item failure
        }
      }
    }
  })
}
```

### Resource Management

```hcl
# Use lifecycle rules to prevent accidental deletion of critical workflows
resource "n8n_workflow" "critical_workflow" {
  name = "Critical Business Process"
  
  lifecycle {
    prevent_destroy = true
    create_before_destroy = true
  }
}

# Use data sources to reference existing resources
data "n8n_workflow" "existing_workflow" {
  id = var.existing_workflow_id
}

resource "n8n_workflow" "dependent_workflow" {
  name = "Dependent Workflow"
  
  settings = jsonencode({
    errorWorkflow = data.n8n_workflow.existing_workflow.id
  })
}
```

## CI/CD Integration

### GitHub Actions Example

```yaml
# .github/workflows/terraform.yml
name: Terraform n8n Infrastructure

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  terraform:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup Terraform
      uses: hashicorp/setup-terraform@v2
      
    - name: Terraform Format Check
      run: terraform fmt -check -recursive
      
    - name: Terraform Init
      run: terraform init
      
    - name: Terraform Validate
      run: terraform validate
      
    - name: Terraform Plan
      run: terraform plan -no-color
      env:
        TF_VAR_n8n_api_key: ${{ secrets.N8N_API_KEY }}
        
    - name: Terraform Apply
      if: github.ref == 'refs/heads/main'
      run: terraform apply -auto-approve
      env:
        TF_VAR_n8n_api_key: ${{ secrets.N8N_API_KEY }}
```

### Pre-commit Hooks

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.4.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-json
      - id: check-yaml

  - repo: https://github.com/antonbabenko/pre-commit-terraform
    rev: v1.77.0
    hooks:
      - id: terraform_fmt
      - id: terraform_validate
      - id: terraform_docs
      - id: terraform_tflint
```

## Documentation Standards

### Inline Documentation

```hcl
# Main data processing workflow
# Runs daily at 9 AM UTC to fetch and process external data
# Dependencies: external_api_credential, database_credential
# Triggers: error_handler_workflow on failure
resource "n8n_workflow" "data_processor" {
  name        = "Daily Data Processor"
  description = "Fetches data from external API, transforms it, and stores in database"
  active      = var.activate_workflows
  
  # Workflow configuration with detailed comments
  nodes = jsonencode({
    # Cron trigger - runs daily at 9 AM UTC
    "Daily Trigger" = {
      "parameters" = {
        "rule" = {
          "interval" = [
            {
              "field" = "cronExpression",
              "value" = "0 9 * * *"  # 9 AM UTC daily
            }
          ]
        }
      },
      "type"       = "n8n-nodes-base.scheduleTrigger",
      "typeVersion" = 1,
      "position"   = [240, 300]
    }
  })
  
  tags = [
    "data-processing",
    "scheduled",
    "production"
  ]
}
```

### README Templates

Create comprehensive README files:

```markdown
# n8n Automation Infrastructure

## Overview
This repository manages our n8n workflow automation infrastructure using Terraform.

## Architecture
- **Environment**: Production
- **Workflows**: 15 active workflows
- **Credentials**: 8 managed credentials
- **Integrations**: GitHub, Slack, PostgreSQL, AWS S3

## Quick Start
1. Clone this repository
2. Copy `terraform.tfvars.example` to `terraform.tfvars`
3. Set your credentials in the tfvars file
4. Run `terraform init && terraform apply`

## Workflows
| Name | Purpose | Schedule | Dependencies |
|------|---------|----------|--------------|
| Data Processor | ETL pipeline | Daily 9 AM | external_api, database |
| Error Handler | Global error handling | On-demand | slack_webhook |

## Maintenance
- **Credential Rotation**: Monthly
- **Backup Schedule**: Daily 2 AM
- **Health Checks**: Every 5 minutes
```

## Common Anti-Patterns to Avoid

### ❌ Don't Do This

```hcl
# Hard-coded sensitive values
resource "n8n_credential" "bad" {
  data = jsonencode({
    password = "hardcoded_password"  # Never!
  })
}

# Monolithic workflows
resource "n8n_workflow" "everything" {
  name = "Do Everything Workflow"  # Too broad
  # 500 lines of node configuration...
}

# No error handling
resource "n8n_workflow" "fragile" {
  # No error workflow configuration
  # No retry logic
  # No monitoring
}

# Poor resource naming
resource "n8n_workflow" "wf1" {  # Unclear naming
  name = "workflow"  # Generic name
}
```

### ✅ Do This Instead

```hcl
# Use variables for sensitive data
resource "n8n_credential" "secure" {
  data = jsonencode({
    password = var.database_password
  })
}

# Focused, single-purpose workflows
resource "n8n_workflow" "data_ingestion" {
  name = "Customer Data Ingestion"
  # Clear purpose and scope
}

# Comprehensive error handling
resource "n8n_workflow" "robust" {
  settings = jsonencode({
    errorWorkflow = n8n_workflow.error_handler.id
  })
}

# Clear, descriptive naming
resource "n8n_workflow" "customer_data_sync" {
  name = "Customer Data Synchronization"
}
```

## Summary Checklist

Before deploying to production, ensure:

- [ ] All sensitive values use variables
- [ ] Workflows have appropriate error handling
- [ ] Resource names are descriptive
- [ ] Code is properly formatted (`terraform fmt`)
- [ ] Configuration is validated (`terraform validate`)
- [ ] Security scan passes (`tfsec`)
- [ ] Documentation is up to date
- [ ] Testing workflows are created
- [ ] Monitoring is configured
- [ ] Backup strategy is implemented
- [ ] CI/CD pipeline is working
- [ ] Team has access to required secrets

Following these best practices will help you build maintainable, secure, and reliable n8n automation infrastructure with Terraform.