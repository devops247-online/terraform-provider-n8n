# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Terraform provider for n8n workflow automation platform. The provider enables Infrastructure as Code management of n8n workflows, credentials, users, and other resources through Terraform.

## Development Commands

### Building and Testing
- `make build` - Build the provider binary
- `make test` - Run unit tests
- `make testacc` - Run acceptance tests (requires running n8n instance)
- `make fmt` - Format Go code using gofmt
- `make vet` - Run go vet for static analysis
- `make lint` - Run golangci-lint (requires golangci-lint installed)
- `make tidy` - Clean up go.mod dependencies

### Development Setup
- `make dev-setup` - Install development tools and set up environment
- `make tools` - Install required development tools (tfplugindocs, golangci-lint)
- `make deps` - Download Go module dependencies
- `make docs` - Generate provider documentation using terraform-plugin-docs

### Local Installation and Testing
- `make local-install` - Install provider locally for testing with Terraform
- `make install` - Install to standard Terraform plugins directory
- `make uninstall` - Remove locally installed provider

### Release Preparation
- `make pre-release` - Run all checks (fmt, vet, lint, test, docs) before release

## Code Architecture

### Core Components

#### Provider Layer (`internal/provider/`)
- **`provider.go`** - Main provider implementation with configuration schema and authentication
- **`workflow_resource.go`** - Terraform resource for managing n8n workflows

#### Client Layer (`internal/client/`)
- **`client.go`** - HTTP client with authentication handling (API key and basic auth)
- **`workflows.go`** - n8n API client methods for workflow operations (CRUD, activate/deactivate)

#### Main Entry Point
- **`main.go`** - Provider server initialization and Terraform plugin framework integration

### Authentication Methods
The provider supports dual authentication:
1. **API Key Authentication** (recommended) - Uses `X-N8N-API-KEY` header
2. **Basic Authentication** - Uses email/password credentials

### Key Patterns

#### Resource Implementation
- Resources implement Terraform Plugin Framework interfaces
- JSON fields (nodes, connections, settings) are stored as strings and marshaled/unmarshaled as needed
- State management follows Terraform best practices with proper error handling

#### Client Design
- Generic HTTP client with pluggable authentication methods
- Error handling with structured API error responses
- Automatic API path construction (`/api/v1/` prefix)

## Environment Variables

### Provider Configuration
- `N8N_BASE_URL` - Base URL of n8n instance
- `N8N_API_KEY` - API key for authentication  
- `N8N_EMAIL` - Email for basic authentication
- `N8N_PASSWORD` - Password for basic authentication
- `N8N_INSECURE_SKIP_VERIFY` - Skip TLS verification (default: false)

### Testing
- `TF_ACC=1` - Enable acceptance tests (set when running `make testacc`)

## Testing Requirements

### Prerequisites for Acceptance Tests
- Running n8n instance (typically http://localhost:5678)
- Valid API key or email/password credentials
- Set environment variables: `N8N_BASE_URL` and `N8N_API_KEY`

### Test Structure
- Unit tests in `*_test.go` files alongside source code
- Acceptance tests require `TF_ACC=1` environment variable
- Tests use Terraform Plugin Framework testing utilities

## Documentation Generation

The provider uses `terraform-plugin-docs` for automatic documentation generation:
- Provider docs: `docs/index.md`
- Resource docs: `docs/resources/`
- Generated from schema definitions and examples in code

## JSON Configuration Patterns

Workflow resources use JSON strings for complex configuration:
- **nodes** - Node definitions with type, parameters, positions
- **connections** - Inter-node connection definitions  
- **settings** - Workflow execution settings
- **static_data** - Persistent workflow data
- **pinned_data** - Test data for development

## Build Configuration

### Go Module
- Module: `github.com/devops247-online/terraform-provider-n8n`
- Go version: 1.23.4
- Key dependencies: Terraform Plugin Framework, HashiCorp Plugin SDK

### Terraform Provider Registry
- Registry path: `registry.terraform.io/devops247-online/n8n`
- Version constraint: `~> 1.0`

## Common Development Workflows

### Adding a New Resource
1. Create resource struct implementing `resource.Resource` interface
2. Define resource model with terraform-plugin-framework types
3. Implement Schema, Create, Read, Update, Delete methods
4. Add corresponding client methods in `internal/client/`
5. Register resource in provider's `Resources()` method
6. Add tests and documentation

### Adding Client Methods
1. Define data structures in client package
2. Implement API methods using the generic `doRequest` method
3. Handle error responses and JSON marshaling/unmarshaling
4. Add unit tests with mock HTTP responses