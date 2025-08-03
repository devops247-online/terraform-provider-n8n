# Terraform Provider for n8n

<div align="center">

[![Build Status](https://github.com/devops247-online/terraform-provider-n8n/actions/workflows/test.yml/badge.svg)](https://github.com/devops247-online/terraform-provider-n8n/actions/workflows/test.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/devops247-online/terraform-provider-n8n)](https://golang.org/dl/)
[![Terraform](https://img.shields.io/badge/terraform-%3E%3D1.0-blueviolet)](https://www.terraform.io/)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)
[![Latest Release](https://img.shields.io/github/release/devops247-online/terraform-provider-n8n.svg)](https://github.com/devops247-online/terraform-provider-n8n/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/devops247-online/terraform-provider-n8n)](https://goreportcard.com/report/github.com/devops247-online/terraform-provider-n8n)
[![Registry](https://img.shields.io/badge/registry-terraform-purple)](https://registry.terraform.io/providers/devops247-online/n8n/latest)
[![Documentation](https://img.shields.io/badge/docs-terraform--provider--n8n-blue)](https://registry.terraform.io/providers/devops247-online/n8n/latest/docs)
[![GitHub Issues](https://img.shields.io/github/issues/devops247-online/terraform-provider-n8n)](https://github.com/devops247-online/terraform-provider-n8n/issues)
[![GitHub Stars](https://img.shields.io/github/stars/devops247-online/terraform-provider-n8n)](https://github.com/devops247-online/terraform-provider-n8n/stargazers)

</div>

**The official Terraform provider for [n8n](https://n8n.io/) workflow automation platform**

Manage your n8n workflows, credentials, users, and automation infrastructure as code with Terraform. This provider enables Infrastructure as Code (IaC) best practices for n8n deployments, allowing you to version control, automate, and scale your workflow automation setup.

## 🚀 Why Use This Provider?

- **Infrastructure as Code**: Manage n8n resources with the same rigor as your infrastructure
- **Version Control**: Track changes to workflows and credentials in Git
- **Automation**: Deploy and manage n8n resources through CI/CD pipelines
- **Consistency**: Ensure consistent n8n setups across environments
- **Scale**: Manage large numbers of workflows and credentials efficiently
- **Integration**: Seamlessly integrate with existing Terraform infrastructure

## 📋 Table of Contents

- [Why Use This Provider?](#-why-use-this-provider)
- [Features](#-features)
- [Quick Start](#-quick-start)  
- [Installation](#-installation)
- [Usage](#-usage)
- [Authentication](#-authentication)
- [Examples](#-examples)
- [Development](#️-development)
- [Documentation](#-documentation)
- [Contributing](#-contributing)
- [Support](#-support)
- [License](#license)

## ✨ Features

- **Workflows**: Create, manage, and organize n8n workflows
- **Credentials**: Securely manage authentication credentials for external services
- **Users**: Manage user accounts and permissions
- **Projects**: Organize workflows in projects (Enterprise feature)
- **Tags**: Categorize and organize workflows with tags
- **Folders**: Hierarchical organization of workflows

## 🚀 Quick Start

Get started with the n8n Terraform provider in 3 simple steps:

```hcl
# 1. Configure the provider
terraform {
  required_providers {
    n8n = {
      source  = "devops247-online/n8n"
      version = "~> 1.0"
    }
  }
}

# 2. Set up authentication
provider "n8n" {
  base_url = "https://your-n8n-instance.com"
  api_key  = var.n8n_api_key
}

# 3. Create your first workflow
resource "n8n_workflow" "hello_world" {
  name   = "Hello World Workflow"
  active = true
  
  nodes = jsonencode({
    "Start": {
      "parameters": {},
      "type": "n8n-nodes-base.start",
      "typeVersion": 1,
      "position": [240, 300]
    }
  })
}
```

## 📦 Installation

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

## 🔧 Usage

### Provider Configuration

```hcl
provider "n8n" {
  base_url = "https://your-n8n-instance.com"
  api_key  = var.n8n_api_key
}
```

## 🔐 Authentication

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

## 📝 Examples

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

## 🛠️ Development

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

## 📚 Documentation

- [Provider Documentation](./docs/index.md)
- [Resource Reference](./docs/resources/)
- [Data Source Reference](./docs/data-sources/)
- [Examples](./examples/)

## 🤝 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

### Reporting Issues

Please use the [GitHub issue tracker](https://github.com/devops247-online/terraform-provider-n8n/issues) to report bugs or request features.

## 📄 License

This project is licensed under the **Mozilla Public License 2.0** - see the [LICENSE](LICENSE) file for details.

[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](LICENSE)

## 💬 Support

- [Documentation](https://registry.terraform.io/providers/devops247-online/n8n/latest/docs)
- [GitHub Issues](https://github.com/devops247-online/terraform-provider-n8n/issues)
- [n8n Community](https://community.n8n.io/)

## 🙏 Acknowledgments

- [n8n](https://n8n.io/) team for creating an amazing workflow automation tool
- [HashiCorp](https://www.hashicorp.com/) for Terraform and the Plugin Framework  
- The open-source community for inspiration and contributions

---

<div align="center">

**⭐ Star this repository if it helped you! ⭐**

[Report Bug](https://github.com/devops247-online/terraform-provider-n8n/issues) • [Request Feature](https://github.com/devops247-online/terraform-provider-n8n/issues) • [Contribute](CONTRIBUTING.md)

Made with ❤️ by [DevOps247](https://github.com/devops247-online) | Powered by [n8n](https://n8n.io/) & [Terraform](https://terraform.io/)

</div>