# Terraform Provider for n8n

A Terraform provider for managing [n8n](https://n8n.io/) workflow automation resources.

## Features

- **Workflows**: Create, manage, and organize n8n workflows
- **Credentials**: Securely manage authentication credentials for external services
- **Users**: Manage user accounts and permissions
- **Projects**: Organize workflows in projects (Enterprise feature)
- **Tags**: Categorize and organize workflows with tags
- **Folders**: Hierarchical organization of workflows

## Installation

### Terraform Registry (Recommended)

```hcl
terraform {
  required_providers {
    n8n = {
      source  = "devops247-online/n8n"
      version = "~> 1.0"
    }
  }
}
```

### Manual Installation

1. Download the latest release from the [releases page](https://github.com/devops247-online/terraform-provider-n8n/releases)
2. Extract the binary to your Terraform plugins directory:
   ```bash
   mkdir -p ~/.terraform.d/plugins/registry.terraform.io/devops247-online/n8n/1.0.0/linux_amd64/
   cp terraform-provider-n8n ~/.terraform.d/plugins/registry.terraform.io/devops247-online/n8n/1.0.0/linux_amd64/
   ```

## Usage

### Provider Configuration

```hcl
provider "n8n" {
  base_url = "https://your-n8n-instance.com"
  api_key  = var.n8n_api_key
}
```

### Authentication Methods

#### API Key Authentication (Recommended)
```hcl
provider "n8n" {
  base_url = "https://your-n8n-instance.com"
  api_key  = var.n8n_api_key
}
```

#### Basic Authentication
```hcl
provider "n8n" {
  base_url = "https://your-n8n-instance.com"
  email    = var.n8n_email
  password = var.n8n_password
}
```

### Environment Variables

You can configure the provider using environment variables:

- `N8N_BASE_URL` - The base URL of your n8n instance
- `N8N_API_KEY` - API key for authentication
- `N8N_EMAIL` - Email for basic authentication
- `N8N_PASSWORD` - Password for basic authentication
- `N8N_INSECURE_SKIP_VERIFY` - Skip TLS certificate verification (default: false)

### Example Usage

#### Basic Workflow Management

```hcl
# Create a simple workflow
resource "n8n_workflow" "example" {
  name   = "My Terraform Workflow"
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
        "url": "https://api.example.com/data",
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
  
  tags = ["terraform", "api"]
}
```

#### Credential Management

```hcl
# Create API credentials
resource "n8n_credential" "api_key" {
  name = "External API Key"
  type = "httpHeaderAuth"
  
  data = {
    name  = "X-API-Key"
    value = var.external_api_key
  }
}

# Use credentials in workflow
resource "n8n_workflow" "api_workflow" {
  name   = "API Integration Workflow"
  active = true
  
  # Reference the credential in workflow nodes
  nodes = jsonencode({
    "API Call": {
      "parameters": {
        "url": "https://api.example.com/data",
        "authentication": "predefinedCredentialType",
        "nodeCredentialType": "httpHeaderAuth"
      },
      "credentials": {
        "httpHeaderAuth": {
          "id": n8n_credential.api_key.id,
          "name": n8n_credential.api_key.name
        }
      },
      "type": "n8n-nodes-base.httpRequest",
      "typeVersion": 1
    }
  })
}
```

## Development

### Prerequisites

- [Go](https://golang.org/doc/install) >= 1.21
- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [n8n instance](https://docs.n8n.io/hosting/) for testing

### Building the Provider

```bash
# Clone the repository
git clone https://github.com/devops247-online/terraform-provider-n8n.git
cd terraform-provider-n8n

# Install dependencies
make deps

# Build the provider
make build

# Install locally for testing
make local-install
```

### Running Tests

```bash
# Unit tests
make test

# Acceptance tests (requires n8n instance)
export N8N_BASE_URL="http://localhost:5678"
export N8N_API_KEY="your-api-key"
make testacc
```

### Development Setup

```bash
# Set up development environment
make dev-setup

# Format code
make fmt

# Run linting
make lint

# Generate documentation
make docs
```

## Documentation

- [Provider Documentation](./docs/index.md)
- [Resource Reference](./docs/resources/)
- [Data Source Reference](./docs/data-sources/)
- [Examples](./examples/)

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

### Reporting Issues

Please use the [GitHub issue tracker](https://github.com/devops247-online/terraform-provider-n8n/issues) to report bugs or request features.

## License

This project is licensed under the Mozilla Public License 2.0 - see the [LICENSE](LICENSE) file for details.

## Support

- [Documentation](https://registry.terraform.io/providers/devops247-online/n8n/latest/docs)
- [GitHub Issues](https://github.com/devops247-online/terraform-provider-n8n/issues)
- [n8n Community](https://community.n8n.io/)

## Acknowledgments

- [n8n](https://n8n.io/) team for creating an amazing workflow automation tool
- [HashiCorp](https://www.hashicorp.com/) for Terraform and the Plugin Framework
- The open-source community for inspiration and contributions