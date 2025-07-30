# n8n Provider

The n8n provider allows Terraform to manage n8n workflow automation resources.

## Example Usage

```terraform
terraform {
  required_providers {
    n8n = {
      source  = "devops247-online/n8n"
      version = "~> 1.0"
    }
  }
}

provider "n8n" {
  base_url = "https://your-n8n-instance.com"
  api_key  = var.n8n_api_key
}
```

## Authentication

The n8n provider supports two authentication methods:

### API Key Authentication (Recommended)

```terraform
provider "n8n" {
  base_url = "https://your-n8n-instance.com"
  api_key  = var.n8n_api_key
}
```

### Basic Authentication

```terraform
provider "n8n" {
  base_url = "https://your-n8n-instance.com"
  email    = var.n8n_email
  password = var.n8n_password
}
```

## Schema

### Required

- `base_url` (String) The base URL of your n8n instance

### Optional

- `api_key` (String, Sensitive) API key for authentication with n8n. Can be set via the `N8N_API_KEY` environment variable.
- `email` (String) Email for basic authentication with n8n. Can be set via the `N8N_EMAIL` environment variable. Alternative to api_key.
- `password` (String, Sensitive) Password for basic authentication with n8n. Can be set via the `N8N_PASSWORD` environment variable. Alternative to api_key.
- `insecure_skip_verify` (Boolean) Skip TLS certificate verification. Can be set via the `N8N_INSECURE_SKIP_VERIFY` environment variable. Defaults to false.

## Environment Variables

You can configure the provider using environment variables:

- `N8N_BASE_URL` - The base URL of your n8n instance
- `N8N_API_KEY` - API key for authentication
- `N8N_EMAIL` - Email for basic authentication
- `N8N_PASSWORD` - Password for basic authentication
- `N8N_INSECURE_SKIP_VERIFY` - Skip TLS certificate verification (default: false)