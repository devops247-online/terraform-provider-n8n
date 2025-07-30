# Migration Guide: From Manual n8n Management to Terraform

This guide helps you migrate your existing n8n workflows and credentials from manual management to Infrastructure as Code with Terraform.

## Overview

Migrating to Terraform management provides several benefits:
- **Version Control**: Track changes to your automation infrastructure
- **Reproducibility**: Easily replicate environments
- **Collaboration**: Team-based workflow management
- **Backup & Recovery**: Infrastructure as code serves as documentation and backup
- **CI/CD Integration**: Automated deployment and testing

## Pre-Migration Assessment

### 1. Inventory Your Current n8n Setup

First, understand what you currently have:

```bash
# Get all workflows
curl -H "X-N8N-API-KEY: your-api-key" \
  http://localhost:5678/api/v1/workflows > current-workflows.json

# Get all credentials
curl -H "X-N8N-API-KEY: your-api-key" \
  http://localhost:5678/api/v1/credentials > current-credentials.json

# Get user information
curl -H "X-N8N-API-KEY: your-api-key" \
  http://localhost:5678/api/v1/users > current-users.json
```

### 2. Categorize Your Resources

Organize your resources by:
- **Environment** (dev, staging, prod)
- **Team ownership**
- **Business function** (data processing, notifications, integrations)
- **Criticality** (mission-critical, important, experimental)

### 3. Identify Dependencies

Map out dependencies between:
- Workflows that trigger other workflows
- Shared credentials
- External systems and APIs

## Migration Strategy

### Phase 1: Read-Only Import (Recommended)

Start by importing existing resources without making changes:

1. **Set up Terraform configuration**
2. **Import existing resources**
3. **Verify state matches reality**
4. **Test planned changes**

### Phase 2: Gradual Management

Gradually take over management:

1. **Start with non-critical workflows**
2. **Test changes in development environment**
3. **Migrate credentials**
4. **Move production workflows last**

### Phase 3: Full Management

Complete the transition:

1. **All resources managed by Terraform**
2. **Manual creation disabled/discouraged**
3. **CI/CD pipelines established**
4. **Team training completed**

## Step-by-Step Migration Process

### Step 1: Set Up Terraform Project

Create your Terraform project structure:

```bash
mkdir n8n-terraform-migration
cd n8n-terraform-migration

# Create basic structure
mkdir -p {environments/{dev,staging,prod},modules/{workflows,credentials},scripts}
```

### Step 2: Create Basic Provider Configuration

```hcl
# main.tf
terraform {
  required_providers {
    n8n = {
      source  = "devops247-online/n8n"
      version = "~> 1.0"
    }
  }
}

provider "n8n" {
  base_url = var.n8n_base_url
  api_key  = var.n8n_api_key
}
```

### Step 3: Generate Import Scripts

Create a script to help with imports:

```bash
#!/bin/bash
# scripts/generate-import-commands.sh

API_KEY="your-api-key"
BASE_URL="http://localhost:5678"

echo "# Workflow Import Commands"
curl -s -H "X-N8N-API-KEY: $API_KEY" "$BASE_URL/api/v1/workflows" | \
  jq -r '.data[] | "terraform import n8n_workflow." + (.name | gsub("[^a-zA-Z0-9]"; "_") | ascii_downcase) + " " + .id'

echo ""
echo "# Credential Import Commands" 
curl -s -H "X-N8N-API-KEY: $API_KEY" "$BASE_URL/api/v1/credentials" | \
  jq -r '.data[] | "terraform import n8n_credential." + (.name | gsub("[^a-zA-Z0-9]"; "_") | ascii_downcase) + " " + .id'
```

### Step 4: Start with Non-Critical Resources

Begin with test or non-critical workflows:

```hcl
# Import a simple workflow first
resource "n8n_workflow" "test_workflow" {
  name   = "Test Workflow"
  active = false  # Keep inactive during migration
  
  # Import will populate these fields
  nodes       = ""
  connections = ""
  settings    = ""
}
```

Import it:
```bash
terraform import n8n_workflow.test_workflow workflow-id-here
```

### Step 5: Extract Configuration

After importing, extract the actual configuration:

```bash
# Get the current state
terraform show n8n_workflow.test_workflow

# Or use the n8n API to get full configuration
curl -H "X-N8N-API-KEY: your-api-key" \
  http://localhost:5678/api/v1/workflows/workflow-id-here > workflow-config.json
```

Update your Terraform configuration with the real values:

```hcl
resource "n8n_workflow" "test_workflow" {
  name   = "Test Workflow"
  active = false
  
  nodes = jsonencode({
    # Copy from the API response or terraform show
    "Start" = {
      "parameters" = {},
      "type"       = "n8n-nodes-base.start",
      "typeVersion" = 1,
      "position"   = [240, 300]
    }
    # ... rest of nodes
  })
  
  connections = jsonencode({
    # Copy connections configuration
  })
  
  settings = jsonencode({
    # Copy settings if any
  })
  
  tags = ["migrated", "test"]
}
```

### Step 6: Validate the Migration

Test that Terraform matches the current state:

```bash
# This should show no changes
terraform plan

# If there are differences, adjust the configuration
# until terraform plan shows no changes
```

## Migration Utilities

### Workflow Export Script

```bash
#!/bin/bash
# scripts/export-workflow.sh

WORKFLOW_ID=$1
API_KEY=$2
BASE_URL=${3:-"http://localhost:5678"}

if [ -z "$WORKFLOW_ID" ] || [ -z "$API_KEY" ]; then
  echo "Usage: $0 <workflow-id> <api-key> [base-url]"
  exit 1
fi

# Get workflow details
WORKFLOW=$(curl -s -H "X-N8N-API-KEY: $API_KEY" \
  "$BASE_URL/api/v1/workflows/$WORKFLOW_ID")

# Extract name for file naming
NAME=$(echo "$WORKFLOW" | jq -r '.name' | tr ' ' '_' | tr '[:upper:]' '[:lower:]')

# Save to file
echo "$WORKFLOW" > "exported_workflows/${NAME}.json"

# Generate Terraform resource
cat > "exported_workflows/${NAME}.tf" << EOF
resource "n8n_workflow" "${NAME}" {
  name   = $(echo "$WORKFLOW" | jq '.name')
  active = $(echo "$WORKFLOW" | jq '.active')
  
  nodes = jsonencode($(echo "$WORKFLOW" | jq '.nodes'))
  
  connections = jsonencode($(echo "$WORKFLOW" | jq '.connections'))
  
  $(if [ "$(echo "$WORKFLOW" | jq '.settings')" != "null" ]; then
    echo "settings = jsonencode($(echo "$WORKFLOW" | jq '.settings'))"
  fi)
  
  $(if [ "$(echo "$WORKFLOW" | jq '.staticData')" != "null" ]; then
    echo "static_data = jsonencode($(echo "$WORKFLOW" | jq '.staticData'))"
  fi)
  
  tags = $(echo "$WORKFLOW" | jq '.tags // []')
}
EOF

echo "Exported workflow to exported_workflows/${NAME}.tf"
echo "Import command: terraform import n8n_workflow.${NAME} $WORKFLOW_ID"
```

### Credential Migration Script

```bash
#!/bin/bash
# scripts/migrate-credential.sh

CREDENTIAL_ID=$1
API_KEY=$2
NEW_NAME=$3

# Note: You cannot retrieve credential data via API for security reasons
# You'll need to recreate credentials with Terraform

echo "Creating Terraform configuration for credential $CREDENTIAL_ID"

# Get credential metadata (not sensitive data)
CRED_META=$(curl -s -H "X-N8N-API-KEY: $API_KEY" \
  "http://localhost:5678/api/v1/credentials/$CREDENTIAL_ID")

NAME=$(echo "$CRED_META" | jq -r '.name' | tr ' ' '_' | tr '[:upper:]' '[:lower:]')
TYPE=$(echo "$CRED_META" | jq -r '.type')

cat > "credentials/${NAME}.tf" << EOF
# Migrated credential: $(echo "$CRED_META" | jq -r '.name')
# Original ID: $CREDENTIAL_ID
# Type: $TYPE

resource "n8n_credential" "$NAME" {
  name = $(echo "$CRED_META" | jq '.name')
  type = "$TYPE"
  
  # TODO: Set the credential data
  # You'll need to manually configure the data field
  # based on the credential type requirements
  data = jsonencode({
    # Add required fields for credential type: $TYPE
    # Example for httpHeaderAuth:
    # name  = "Authorization"
    # value = var.${NAME}_token
  })
}

# TODO: Add corresponding variable
variable "${NAME}_token" {
  description = "Token for $(echo "$CRED_META" | jq -r '.name')"
  type        = string
  sensitive   = true
}
EOF

echo "Generated template at credentials/${NAME}.tf"
echo "You'll need to:"
echo "1. Configure the data field with actual credential data"
echo "2. Set up variables for sensitive values"
echo "3. Run: terraform import n8n_credential.$NAME $CREDENTIAL_ID"
```

## Common Migration Challenges

### Challenge 1: Credential Data Not Accessible

**Problem**: n8n API doesn't return sensitive credential data for security reasons.

**Solution**: Recreate credentials with Terraform:

```hcl
# Create new credential with Terraform
resource "n8n_credential" "migrated_api_key" {
  name = "Migrated API Key"
  type = "httpHeaderAuth"
  
  data = jsonencode({
    name  = "Authorization"
    value = var.api_key_value  # Set this variable
  })
}

# Update workflows to use new credential ID
resource "n8n_workflow" "updated_workflow" {
  # ... workflow config ...
  
  depends_on = [n8n_credential.migrated_api_key]
  
  nodes = jsonencode({
    "API Request" = {
      # ... node config ...
      "credentials" = {
        "httpHeaderAuth" = {
          "id"   = n8n_credential.migrated_api_key.id
          "name" = n8n_credential.migrated_api_key.name
        }
      }
    }
  })
}
```

### Challenge 2: Complex Workflow JSON

**Problem**: Large, complex workflows are hard to manage as JSON strings.

**Solution**: Use external files and templates:

```hcl
resource "n8n_workflow" "complex_workflow" {
  name = "Complex Data Processing"
  
  # Store complex JSON in separate files
  nodes       = file("${path.module}/workflows/complex-nodes.json")
  connections = file("${path.module}/workflows/complex-connections.json")
  settings    = templatefile("${path.module}/workflows/complex-settings.json.tpl", {
    timeout = var.workflow_timeout
    retries = var.max_retries
  })
}
```

### Challenge 3: Workflow Dependencies

**Problem**: Workflows reference other workflows by ID.

**Solution**: Use Terraform references:

```hcl
resource "n8n_workflow" "error_handler" {
  name = "Global Error Handler"
  # ... configuration ...
}

resource "n8n_workflow" "main_workflow" {
  name = "Main Workflow"
  
  settings = jsonencode({
    errorWorkflow = n8n_workflow.error_handler.id  # Reference other workflow
  })
  
  depends_on = [n8n_workflow.error_handler]
}
```

### Challenge 4: Environment Differences

**Problem**: Different configurations per environment.

**Solution**: Use workspace or variable-based configuration:

```hcl
# variables.tf
variable "environment" {
  description = "Environment name"
  type        = string
}

# main.tf
locals {
  environment_config = {
    dev = {
      api_url = "https://dev-api.example.com"
      retries = 1
    }
    prod = {
      api_url = "https://api.example.com"
      retries = 3
    }
  }
  
  config = local.environment_config[var.environment]
}

resource "n8n_workflow" "api_processor" {
  name = "${var.environment}-api-processor"
  
  nodes = templatefile("${path.module}/templates/api-workflow.json.tpl", {
    api_url = local.config.api_url
    retries = local.config.retries
  })
}
```

## Post-Migration Best Practices

### 1. Establish Processes

- **Code Review**: All workflow changes go through PR review
- **Testing**: Test workflows in development before production
- **Documentation**: Keep README files updated
- **Training**: Train team on Terraform workflows

### 2. Set Up CI/CD

```yaml
# .github/workflows/n8n-deploy.yml
name: Deploy n8n Infrastructure

on:
  push:
    branches: [main]
    paths: ['n8n/**']

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup Terraform
      uses: hashicorp/setup-terraform@v2
      
    - name: Terraform Init
      run: terraform init
      working-directory: ./n8n
      
    - name: Terraform Plan
      run: terraform plan -no-color
      working-directory: ./n8n
      env:
        TF_VAR_n8n_api_key: ${{ secrets.N8N_API_KEY }}
        
    - name: Terraform Apply
      run: terraform apply -auto-approve
      working-directory: ./n8n
      env:
        TF_VAR_n8n_api_key: ${{ secrets.N8N_API_KEY }}
```

### 3. Monitoring and Maintenance

```hcl
# Create monitoring workflows
resource "n8n_workflow" "terraform_drift_detection" {
  name = "Terraform Drift Detection"
  
  nodes = jsonencode({
    "Schedule" = {
      "parameters" = {
        "rule" = {
          "interval" = [{"field" = "cronExpression", "value" = "0 9 * * *"}]
        }
      },
      "type" = "n8n-nodes-base.scheduleTrigger"
    },
    "Check Drift" = {
      "parameters" = {
        "command" = "cd /terraform && terraform plan -detailed-exitcode"
      },
      "type" = "n8n-nodes-base.executeCommand"
    },
    "Alert on Drift" = {
      # Send alert if plan shows changes
    }
  })
}
```

## Migration Checklist

### Pre-Migration
- [ ] Inventory all workflows and credentials
- [ ] Categorize by criticality and ownership
- [ ] Set up development/staging environment
- [ ] Create backup of current n8n instance
- [ ] Install and configure Terraform

### During Migration
- [ ] Start with non-critical resources
- [ ] Import resources one by one
- [ ] Verify each import with `terraform plan`
- [ ] Test workflow functionality after migration
- [ ] Update documentation

### Post-Migration
- [ ] Establish code review process
- [ ] Set up CI/CD pipeline
- [ ] Train team on new processes
- [ ] Document migration lessons learned
- [ ] Set up monitoring for drift detection
- [ ] Plan for credential rotation

## Rollback Plan

Always have a rollback plan:

1. **Keep backups** of original n8n configuration
2. **Export workflows** before making changes
3. **Test rollback** in development environment
4. **Document rollback** procedures

```bash
# Emergency rollback script
#!/bin/bash
# scripts/rollback.sh

echo "Rolling back n8n infrastructure..."

# Destroy all Terraform-managed resources
terraform destroy -auto-approve

# Restore from backup
curl -X POST -H "Content-Type: application/json" \
  -H "X-N8N-API-KEY: $API_KEY" \
  -d @backup/workflows.json \
  http://localhost:5678/api/v1/workflows/import

echo "Rollback completed"
```

## Success Metrics

Track your migration success:

- **Coverage**: % of workflows managed by Terraform
- **Reliability**: Reduced deployment errors
- **Speed**: Faster environment provisioning
- **Collaboration**: Team participation in workflow changes
- **Recovery**: Mean time to recover from issues

A successful migration to Terraform management will provide better control, visibility, and reliability for your n8n automation infrastructure.