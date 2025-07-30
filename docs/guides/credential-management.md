# Credential Management with n8n Terraform Provider

This guide covers how to securely manage credentials for your n8n workflows using Terraform.

## Overview

Credentials in n8n store authentication information for external services and APIs. The Terraform provider allows you to manage these credentials as Infrastructure as Code while maintaining security best practices.

## Credential Types

n8n supports many credential types. Here are the most common ones:

### HTTP Authentication

#### Basic Authentication
```hcl
resource "n8n_credential" "http_basic" {
  name = "HTTP Basic Auth"
  type = "httpBasicAuth"
  
  data = jsonencode({
    user     = var.http_username
    password = var.http_password
  })
}
```

#### Header Authentication (API Keys)
```hcl
resource "n8n_credential" "api_key" {
  name = "API Key Credential"
  type = "httpHeaderAuth"
  
  data = jsonencode({
    name  = "X-API-Key"
    value = var.api_key_value
  })
}
```

#### Bearer Token
```hcl
resource "n8n_credential" "bearer_token" {
  name = "Bearer Token"
  type = "httpHeaderAuth"
  
  data = jsonencode({
    name  = "Authorization"
    value = "Bearer ${var.bearer_token}"
  })
}
```

### OAuth2 Authentication
```hcl
resource "n8n_credential" "oauth2" {
  name = "OAuth2 Service"
  type = "oAuth2Api"
  
  data = jsonencode({
    clientId          = var.oauth_client_id
    clientSecret      = var.oauth_client_secret
    accessTokenUrl    = "https://api.example.com/oauth/token"
    authUrl           = "https://api.example.com/oauth/authorize"
    scope             = "read write"
    authentication    = "body"
    grantType         = "authorizationCode"
  })
}
```

### Database Credentials

#### PostgreSQL
```hcl
resource "n8n_credential" "postgres" {
  name = "PostgreSQL Database"
  type = "postgres"
  
  data = jsonencode({
    host                       = var.db_host
    database                   = var.db_name
    user                       = var.db_username
    password                   = var.db_password
    port                       = var.db_port
    ssl                        = "require"
    allowUnauthorizedCerts     = false
    connectionTimeout          = 20000
  })
}
```

#### MySQL
```hcl
resource "n8n_credential" "mysql" {
  name = "MySQL Database"
  type = "mysql"
  
  data = jsonencode({
    host         = var.mysql_host
    database     = var.mysql_database
    user         = var.mysql_username
    password     = var.mysql_password
    port         = var.mysql_port
    ssl          = true
  })
}
```

### Cloud Service Credentials

#### AWS
```hcl
resource "n8n_credential" "aws" {
  name = "AWS Credentials"
  type = "aws"
  
  data = jsonencode({
    accessKeyId     = var.aws_access_key_id
    secretAccessKey = var.aws_secret_access_key
    region          = var.aws_region
    customEndpoints = {
      s3 = var.aws_s3_endpoint
    }
  })
}
```

#### Google Cloud
```hcl
resource "n8n_credential" "gcp" {
  name = "Google Cloud Credentials"
  type = "googleApi"
  
  data = jsonencode({
    serviceAccountEmail = var.gcp_service_account_email
    privateKey         = var.gcp_private_key
    projectId          = var.gcp_project_id
  })
}
```

## Security Best Practices

### 1. Never Hardcode Sensitive Values

❌ **Don't do this:**
```hcl
resource "n8n_credential" "bad_example" {
  name = "API Key"
  type = "httpHeaderAuth"
  
  data = jsonencode({
    name  = "X-API-Key"
    value = "sk_live_abc123_this_is_hardcoded"  # ❌ Never do this!
  })
}
```

✅ **Do this instead:**
```hcl
resource "n8n_credential" "good_example" {
  name = "API Key"
  type = "httpHeaderAuth"
  
  data = jsonencode({
    name  = "X-API-Key"
    value = var.api_key_value  # ✅ Use variables
  })
}
```

### 2. Use Terraform Variables

Define sensitive variables properly:

```hcl
# variables.tf
variable "api_key_value" {
  description = "API key for external service"
  type        = string
  sensitive   = true  # Mark as sensitive
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
  validation {
    condition     = length(var.db_password) >= 8
    error_message = "Database password must be at least 8 characters long."
  }
}
```

### 3. Environment Variables

Set sensitive values via environment variables:

```bash
export TF_VAR_api_key_value="your-secret-api-key"
export TF_VAR_db_password="secure-database-password"
```

### 4. Use terraform.tfvars (with .gitignore)

Create `terraform.tfvars` for local development:

```hcl
# terraform.tfvars (add to .gitignore!)
api_key_value = "your-development-api-key"
db_password   = "dev-database-password"
```

**Always add terraform.tfvars to .gitignore:**
```gitignore
# .gitignore
terraform.tfvars
*.tfvars
.terraform/
*.tfstate
*.tfstate.*
```

### 5. Remote State with Encryption

For production, use encrypted remote state:

```hcl
terraform {
  backend "s3" {
    bucket         = "my-terraform-state"
    key            = "n8n/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    kms_key_id     = "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
  }
}
```

## Using Credentials in Workflows

### Basic Usage

```hcl
resource "n8n_workflow" "api_consumer" {
  depends_on = [n8n_credential.api_key]
  
  name = "API Consumer"
  
  nodes = jsonencode({
    "HTTP Request" = {
      "parameters" = {
        "url" = "https://api.example.com/data"
        "authentication" = "predefinedCredentialType"
        "nodeCredentialType" = "httpHeaderAuth"
      },
      "type" = "n8n-nodes-base.httpRequest",
      "credentials" = {
        "httpHeaderAuth" = {
          "id"   = n8n_credential.api_key.id
          "name" = n8n_credential.api_key.name
        }
      }
    }
  })
}
```

### Multiple Credentials in One Workflow

```hcl
resource "n8n_workflow" "multi_service" {
  depends_on = [
    n8n_credential.api_key,
    n8n_credential.database
  ]
  
  name = "Multi-Service Workflow"
  
  nodes = jsonencode({
    "Fetch Data" = {
      "parameters" = {
        "url" = "https://api.example.com/data"
        "authentication" = "predefinedCredentialType"
        "nodeCredentialType" = "httpHeaderAuth"
      },
      "type" = "n8n-nodes-base.httpRequest",
      "credentials" = {
        "httpHeaderAuth" = {
          "id"   = n8n_credential.api_key.id
          "name" = n8n_credential.api_key.name
        }
      }
    },
    "Store Data" = {
      "parameters" = {
        "operation" = "insert"
        "table" = "api_data"
      },
      "type" = "n8n-nodes-base.postgres",
      "credentials" = {
        "postgres" = {
          "id"   = n8n_credential.database.id
          "name" = n8n_credential.database.name
        }
      }
    }
  })
}
```

## Advanced Patterns

### Conditional Credentials by Environment

```hcl
locals {
  api_credentials = {
    dev  = "dev-api-credentials"
    prod = "prod-api-credentials"
  }
}

resource "n8n_credential" "api_key" {
  name = local.api_credentials[var.environment]
  type = "httpHeaderAuth"
  
  data = jsonencode({
    name  = "Authorization"
    value = "Bearer ${var.api_tokens[var.environment]}"
  })
}
```

### Shared Credentials with Node Access Control

```hcl
resource "n8n_credential" "shared_api" {
  name = "Shared API Access"
  type = "httpHeaderAuth"
  
  data = jsonencode({
    name  = "X-API-Key"
    value = var.shared_api_key
  })
  
  # Restrict access to specific node types (if supported)
  node_access = [
    "httpRequest",
    "webhook"
  ]
}
```

### Credential Factory Pattern

```hcl
# Create multiple similar credentials
locals {
  api_services = {
    service_a = {
      name = "Service A API"
      url  = "https://api.service-a.com"
      key  = var.service_a_key
    }
    service_b = {
      name = "Service B API"
      url  = "https://api.service-b.com"
      key  = var.service_b_key
    }
  }
}

resource "n8n_credential" "api_services" {
  for_each = local.api_services
  
  name = each.value.name
  type = "httpHeaderAuth"
  
  data = jsonencode({
    name  = "Authorization"
    value = "Bearer ${each.value.key}"
  })
  
  tags = ["api", "external-service", each.key]
}
```

## Credential Rotation

### Manual Rotation

```hcl
resource "n8n_credential" "rotatable_key" {
  name = "Rotatable API Key"
  type = "httpHeaderAuth"
  
  data = jsonencode({
    name  = "X-API-Key"
    value = var.current_api_key
  })
  
  # Use lifecycle rules to prevent accidental deletion
  lifecycle {
    create_before_destroy = true
  }
}
```

### Automated Rotation with External Tools

```hcl
# Use external data source for dynamic credentials
data "external" "api_key" {
  program = ["/path/to/get-current-api-key.sh"]
}

resource "n8n_credential" "dynamic_key" {
  name = "Auto-Rotated API Key"
  type = "httpHeaderAuth"
  
  data = jsonencode({
    name  = "X-API-Key"
    value = data.external.api_key.result.key
  })
}
```

## Testing and Validation

### Test Credential Creation

```bash
# Plan to see what will be created
terraform plan -target=n8n_credential.api_key

# Apply only the credential
terraform apply -target=n8n_credential.api_key

# Verify in n8n UI
# Navigate to Settings > Credentials to see the created credential
```

### Validate Credential Configuration

Create a simple test workflow to validate credentials:

```hcl
resource "n8n_workflow" "credential_test" {
  name   = "Credential Test - ${n8n_credential.api_key.name}"
  active = false  # Keep inactive
  
  nodes = jsonencode({
    "Test Request" = {
      "parameters" = {
        "url" = "https://httpbin.org/headers"
        "authentication" = "predefinedCredentialType"
        "nodeCredentialType" = "httpHeaderAuth"
      },
      "type" = "n8n-nodes-base.httpRequest",
      "credentials" = {
        "httpHeaderAuth" = {
          "id"   = n8n_credential.api_key.id
          "name" = n8n_credential.api_key.name
        }
      }
    }
  })
  
  tags = ["test", "credential-validation"]
}
```

## Importing Existing Credentials

If you have existing credentials in n8n, you can import them:

```bash
# Get the credential ID from n8n UI or API
terraform import n8n_credential.existing_credential <credential-id>
```

After importing, create the corresponding Terraform configuration:

```hcl
resource "n8n_credential" "existing_credential" {
  name = "Imported Credential"
  type = "httpHeaderAuth"
  
  data = jsonencode({
    name  = "Authorization"
    value = var.imported_credential_value
  })
}
```

## Troubleshooting

### Common Issues

**Credential not found in workflow:**
- Ensure the credential is created before the workflow
- Use `depends_on` to enforce dependency order
- Check that the credential ID matches in the workflow configuration

**Authentication failures:**
- Verify the credential type matches the node requirements
- Check that sensitive values are properly set
- Test the credential manually in n8n UI first

**Permission errors:**
- Ensure your n8n API key has permission to manage credentials
- Check that the credential sharing settings allow access

### Debugging Tips

```bash
# Show current state of credential
terraform show n8n_credential.api_key

# Get credential ID for debugging
terraform output credential_ids

# Force refresh credential from n8n
terraform refresh
```

## Best Practices Summary

1. ✅ **Use variables** for all sensitive values
2. ✅ **Mark variables as sensitive** in Terraform
3. ✅ **Use environment variables** or secure variable files
4. ✅ **Add terraform.tfvars to .gitignore**
5. ✅ **Use encrypted remote state** for production
6. ✅ **Test credentials** before using in workflows
7. ✅ **Plan credential rotation** strategy
8. ✅ **Use depends_on** for proper resource ordering
9. ✅ **Tag credentials** for organization
10. ✅ **Monitor credential usage** and expiration

Following these practices will help you maintain secure, manageable credential infrastructure for your n8n automation workflows.