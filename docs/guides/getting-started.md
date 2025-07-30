# Getting Started with the n8n Terraform Provider

This guide will help you get started with managing your n8n automation infrastructure using Terraform.

## Prerequisites

Before you begin, ensure you have:

- **n8n instance**: A running n8n instance (self-hosted or cloud)
- **Terraform**: Version 1.0 or later installed
- **API Access**: n8n API key or email/password credentials
- **Network Access**: Terraform can reach your n8n instance

## Quick Start

### 1. Create a New Terraform Configuration

Create a new directory for your Terraform configuration:

```bash
mkdir n8n-terraform
cd n8n-terraform
```

### 2. Create Your First Configuration

Create a `main.tf` file with the following content:

```hcl
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
  base_url = "http://localhost:5678"  # Change to your n8n URL
  api_key  = var.n8n_api_key
}

# Create your first workflow
resource "n8n_workflow" "hello_world" {
  name   = "Hello World Workflow"
  active = false  # Start inactive for testing
  
  nodes = jsonencode({
    "Start" = {
      "parameters" = {},
      "type"       = "n8n-nodes-base.start",
      "typeVersion" = 1,
      "position"   = [240, 300]
    },
    "Set Message" = {
      "parameters" = {
        "assignments" = {
          "assignments" = [
            {
              "id"    = "message",
              "name"  = "message",
              "type"  = "string",
              "value" = "Hello from Terraform!"
            }
          ]
        }
      },
      "type"       = "n8n-nodes-base.set",
      "typeVersion" = 3,
      "position"   = [440, 300]
    }
  })
  
  connections = jsonencode({
    "Start" = {
      "main" = [
        [
          {
            "node"  = "Set Message",
            "type"  = "main",
            "index" = 0
          }
        ]
      ]
    }
  })
  
  tags = ["terraform", "getting-started"]
}
```

### 3. Create Variables File

Create a `variables.tf` file:

```hcl
variable "n8n_api_key" {
  description = "API key for n8n authentication"
  type        = string
  sensitive   = true
}
```

### 4. Set Your API Key

Create a `terraform.tfvars` file (don't commit this to version control):

```hcl
n8n_api_key = "your-actual-api-key-here"
```

Or set it as an environment variable:

```bash
export TF_VAR_n8n_api_key="your-actual-api-key-here"
```

### 5. Initialize and Apply

```bash
# Initialize Terraform
terraform init

# Preview changes
terraform plan

# Apply changes
terraform apply
```

### 6. Verify Your Workflow

1. Open your n8n instance in a web browser
2. Navigate to the workflows section
3. You should see "Hello World Workflow" in the list
4. Click on it to view the workflow
5. Test it by clicking the "Execute Workflow" button

## Authentication Methods

The n8n provider supports two authentication methods:

### API Key Authentication (Recommended)

```hcl
provider "n8n" {
  base_url = "https://your-n8n-instance.com"
  api_key  = var.n8n_api_key
}
```

To get your API key:
1. Open n8n settings
2. Go to "API" section
3. Generate a new API key
4. Copy the key for use in Terraform

### Basic Authentication

```hcl
provider "n8n" {
  base_url = "https://your-n8n-instance.com"
  email    = var.n8n_email
  password = var.n8n_password
}
```

## Environment Variables

You can configure the provider using environment variables:

```bash
export N8N_BASE_URL="https://your-n8n-instance.com"
export N8N_API_KEY="your-api-key"
# Or for basic auth:
export N8N_EMAIL="your-email@example.com"
export N8N_PASSWORD="your-password"
```

## Best Practices

### 1. Start with Inactive Workflows

Always create workflows with `active = false` initially:

```hcl
resource "n8n_workflow" "example" {
  name   = "My Workflow"
  active = false  # Test first, then activate
  # ... rest of configuration
}
```

### 2. Use Version Control

- Keep your Terraform configurations in Git
- Use `.gitignore` to exclude sensitive files:

```gitignore
# Terraform
*.tfstate
*.tfstate.*
.terraform/
terraform.tfvars

# Sensitive files
.env
secrets.tf
```

### 3. Organize Your Code

Structure your Terraform files logically:

```
├── main.tf              # Provider and main resources
├── variables.tf         # Variable definitions
├── outputs.tf          # Output definitions
├── terraform.tfvars    # Variable values (don't commit)
├── workflows/          # Complex workflow JSON files
│   ├── data-processor.json
│   └── webhook-handler.json
└── README.md           # Documentation
```

### 4. Use External Files for Complex Workflows

Store complex workflow configurations in separate JSON files:

```hcl
resource "n8n_workflow" "complex" {
  name        = "Complex Workflow"
  nodes       = file("${path.module}/workflows/complex-nodes.json")
  connections = file("${path.module}/workflows/complex-connections.json")
  settings    = file("${path.module}/workflows/complex-settings.json")
}
```

### 5. Tag Your Resources

Use consistent tagging:

```hcl
resource "n8n_workflow" "example" {
  name = "Example Workflow"
  tags = ["environment:prod", "team:data", "project:automation"]
}
```

## Common Patterns

### Workflow with Credentials

```hcl
# Create credential first
resource "n8n_credential" "api_key" {
  name = "External API Key"
  type = "httpHeaderAuth"
  
  data = jsonencode({
    name  = "Authorization"
    value = "Bearer ${var.external_api_token}"
  })
}

# Reference credential in workflow
resource "n8n_workflow" "api_consumer" {
  depends_on = [n8n_credential.api_key]
  
  name = "API Consumer Workflow"
  
  nodes = jsonencode({
    # ... other nodes ...
    "API Request" = {
      "parameters" = {
        "url" = "https://api.example.com/data"
        "authentication" = "predefinedCredentialType"
        "nodeCredentialType" = "httpHeaderAuth"
      },
      "type" = "n8n-nodes-base.httpRequest",
      "credentials" = {
        "httpHeaderAuth" = {
          "id"   = n8n_credential.api_key.id,
          "name" = n8n_credential.api_key.name
        }
      }
    }
  })
}
```

### Multiple Environment Support

```hcl
# variables.tf
variable "environment" {
  description = "Environment (dev, staging, prod)"
  type        = string
  default     = "dev"
}

# main.tf
resource "n8n_workflow" "app_workflow" {
  name   = "${var.environment}-app-workflow"
  active = var.environment == "prod"  # Only activate in production
  
  tags = ["environment:${var.environment}", "managed-by:terraform"]
}
```

## Next Steps

Now that you have a basic setup running:

1. **Explore Examples**: Check out the [examples directory](../../examples/) for more complex scenarios
2. **Learn About Credentials**: Read the [credential management guide](./credential-management.md)
3. **Set Up CI/CD**: Learn about [CI/CD integration patterns](./cicd-integration.md)
4. **Monitor Your Workflows**: Set up [monitoring and alerting](./monitoring.md)

## Troubleshooting

### Common Issues

**Connection refused or timeout errors:**
- Verify your n8n instance is running and accessible
- Check the `base_url` configuration
- Ensure firewall rules allow access

**Authentication errors:**
- Verify your API key is correct and not expired
- Check that the API key has necessary permissions
- Try basic authentication if API key doesn't work

**Workflow not appearing in n8n:**
- Check Terraform apply output for errors
- Verify the workflow was created: `terraform show`
- Refresh your n8n browser page

**JSON syntax errors:**
- Validate your JSON using an online validator
- Use `terraform validate` to check syntax
- Check for proper escaping of quotes and special characters

### Getting Help

- **Provider Issues**: Report on [GitHub Issues](https://github.com/devops247-online/terraform-provider-n8n/issues)
- **n8n Questions**: Visit the [n8n Community Forum](https://community.n8n.io/)
- **Terraform Help**: Check the [Terraform Documentation](https://www.terraform.io/docs/)

Happy automating! 🚀