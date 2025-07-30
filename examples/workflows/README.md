# Workflow Examples

This directory contains examples of managing various types of n8n workflows with the Terraform provider.

## Overview

These examples demonstrate different workflow patterns and use cases that can be managed through Terraform.

## Prerequisites

- n8n instance running and accessible
- Valid n8n API key or email/password credentials
- Terraform >= 1.0

## Workflows Included

### Simple HTTP Request
A basic workflow that makes an HTTP request to httpbin.org. Perfect for testing connectivity and basic functionality.

### Webhook Processing Workflow
A webhook-triggered workflow that processes incoming HTTP requests, adds metadata, and returns a response. Useful for API integrations.

### Scheduled Data Processing
A cron-scheduled workflow that fetches data from an API, processes it, and saves the results. Demonstrates credential usage and complex node chains.

## Usage

1. Copy the example tfvars file:
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

2. Edit `terraform.tfvars` with your actual values

3. Initialize and apply:
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

## Best Practices

### Workflow Activation
- Start workflows as inactive (`active = false`) for testing
- Activate manually or through Terraform after validation
- Use `terraform apply -target=resource_name` for selective activation

### Node Configuration
- Store complex node configurations in separate JSON files using `file()`
- Use variables for URLs, credentials, and configurable parameters
- Include position coordinates for proper canvas layout

### Credentials Integration
- Reference credentials by ID in node configurations
- Create credentials first, then reference them in workflows
- Use credential sharing appropriately based on security requirements

### Testing Workflows
- Use `pinned_data` for testing workflows with static data
- Test workflows manually before activating them
- Monitor execution logs during initial deployment

## JSON File Organization

For complex workflows, consider storing node and connection configurations in separate files:

```hcl
resource "n8n_workflow" "complex" {
  name        = "Complex Workflow"
  nodes       = file("${path.module}/workflows/complex-nodes.json")
  connections = file("${path.module}/workflows/complex-connections.json")
  settings    = file("${path.module}/workflows/complex-settings.json")
}
```

## Importing Existing Workflows

You can import existing workflows from your n8n instance:

```bash
terraform import n8n_workflow.existing_workflow <workflow-id>
```

## Common Patterns

### Error Handling
Configure error workflows for production systems:

```hcl
settings = jsonencode({
  "errorWorkflow" = "error-handler-workflow-id"
})
```

### Webhook Security
For webhook workflows, consider:
- Authentication requirements
- Rate limiting
- Input validation

### Scheduled Workflows
For scheduled workflows:
- Use appropriate cron expressions
- Consider timezone settings
- Plan for failure handling and retries