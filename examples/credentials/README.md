# Credential Management Examples

This directory contains examples of managing various types of credentials with the n8n Terraform provider.

## Overview

Credentials in n8n store authentication information for external services and APIs. They can be securely managed through Terraform while keeping sensitive data encrypted.

## Prerequisites

- n8n instance running and accessible
- Valid n8n API key or email/password credentials
- Terraform >= 1.0

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

## Credential Types

### HTTP Basic Authentication
Used for services that require username/password authentication via HTTP Basic Auth.

### API Key Authentication  
Used for services that require API key authentication via HTTP headers.

### OAuth2 Authentication
Used for services that implement OAuth2 authorization flows.

### Database Credentials
Used for connecting to databases like PostgreSQL, MySQL, etc.

### AWS Credentials
Used for AWS services integration.

## Security Best Practices

1. **Never commit sensitive values**: Always use `terraform.tfvars` (which should be in `.gitignore`)
2. **Use environment variables**: You can set variables via `TF_VAR_` environment variables
3. **Remote state**: Use remote state with encryption for production deployments
4. **Least privilege**: Only grant credentials access to nodes that need them

## Environment Variables

You can also set credentials via environment variables:

```bash
export TF_VAR_n8n_api_key="your-api-key"
export TF_VAR_http_username="username"
export TF_VAR_http_password="password"
# ... etc
```

## Importing Existing Credentials

You can import existing credentials from your n8n instance:

```bash
terraform import n8n_credential.existing_credential <credential-id>
```