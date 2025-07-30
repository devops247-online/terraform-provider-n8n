# Testing Guide

This document describes the testing infrastructure and how to run tests for the n8n Terraform provider.

## Table of Contents

- [Overview](#overview)
- [Test Types](#test-types)
- [Running Tests](#running-tests)
- [Test Environment Setup](#test-environment-setup)
- [Continuous Integration](#continuous-integration)
- [Writing Tests](#writing-tests)
- [Troubleshooting](#troubleshooting)

## Overview

The n8n Terraform provider includes comprehensive testing infrastructure with:

- **Unit Tests**: Fast, isolated tests for individual components
- **Acceptance Tests**: End-to-end tests with real n8n instances
- **Integration Tests**: Tests that verify component interactions
- **Security Scanning**: Automated security vulnerability checks
- **Code Quality**: Formatting, linting, and static analysis

All tests are run automatically in CI/CD pipelines and can be executed locally.

## Test Types

### Unit Tests

Unit tests validate individual functions and components in isolation using mocked dependencies.

**Location**: `*_test.go` files alongside source code
**Coverage**: Aims for >80% code coverage
**Dependencies**: No external services required

### Acceptance Tests

Acceptance tests verify the provider works correctly with real n8n instances.

**Location**: `internal/provider/*_test.go` files with `TestAcc*` prefix
**Dependencies**: Running n8n instance
**Duration**: Longer running tests (up to 120 minutes timeout)

### Integration Tests

Integration tests verify the HTTP client works correctly with various n8n API scenarios.

**Location**: `internal/client/*_test.go` files
**Dependencies**: Mock HTTP servers
**Coverage**: API client functionality, error handling, retries

## Running Tests

### Prerequisites

- Go 1.23.4 or later
- Docker and docker-compose (for acceptance tests)
- Make

### Unit Tests

```bash
# Run all unit tests
make test

# Run tests with coverage report
make coverage

# Run tests for specific package
go test -v ./internal/client/
```

### Acceptance Tests

#### Option 1: With Docker (Recommended)

```bash
# Run acceptance tests with automatic n8n setup
make testacc-docker

# Run with specific n8n version
./scripts/test-acceptance.sh --n8n-version 1.0.5

# Run with custom timeout
./scripts/test-acceptance.sh --timeout 180
```

#### Option 2: With External n8n Instance

```bash
# Set environment variables
export TF_ACC=1
export N8N_BASE_URL="http://localhost:5678"
export N8N_API_KEY="your-api-key"

# Run acceptance tests
make testacc
```

### Code Quality Checks

```bash
# Run all quality checks
make pre-release

# Individual checks
make fmt      # Format code
make vet      # Static analysis
make lint     # Linting
make security # Security scanning
```

### All Tests

```bash
# Run everything (unit, acceptance, quality checks)
make pre-release
```

## Test Environment Setup

### Local n8n Instance with Docker

The easiest way to run acceptance tests is using the provided Docker setup:

```bash
# Start n8n test environment
docker-compose -f docker-compose.test.yml up -d

# Wait for n8n to be ready
curl -f http://localhost:5678/healthz

# Run tests
export TF_ACC=1
export N8N_BASE_URL="http://localhost:5678"
export N8N_API_KEY="test-api-key-12345"
make testacc

# Cleanup
docker-compose -f docker-compose.test.yml down -v
```

### Manual n8n Setup

If you prefer to run n8n manually:

1. Install and start n8n
2. Enable API access and create an API key
3. Set environment variables:
   ```bash
   export N8N_BASE_URL="http://localhost:5678"
   export N8N_API_KEY="your-api-key"
   ```

### Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `TF_ACC` | Enable acceptance tests | Yes (for acc tests) | - |
| `N8N_BASE_URL` | n8n instance URL | Yes (for acc tests) | - |
| `N8N_API_KEY` | n8n API key | Yes* | - |
| `N8N_EMAIL` | n8n user email | Yes* | - |
| `N8N_PASSWORD` | n8n user password | Yes* | - |
| `N8N_INSECURE_SKIP_VERIFY` | Skip TLS verification | No | false |

*Either `N8N_API_KEY` or `N8N_EMAIL`/`N8N_PASSWORD` is required.

## Continuous Integration

### GitHub Actions Workflows

The project includes several GitHub Actions workflows:

#### Test Workflow (`.github/workflows/test.yml`)

Runs on every push and PR:

- **Unit Tests**: Multiple Go versions (1.22.x, 1.23.4)
- **Code Quality**: Formatting, linting, security scanning
- **Acceptance Tests**: Multiple n8n versions (1.0.5, 1.1.0, latest)
- **Documentation**: Validates generated docs
- **Build**: Multi-platform builds

#### Release Workflow (`.github/workflows/release.yml`)

Runs on version tags:

- **Pre-release Checks**: All quality gates
- **Build & Release**: Multi-platform binaries with GoReleaser
- **Post-release Test**: Validates published provider

### Test Matrix

| Component | Go Versions | n8n Versions | Platforms |
|-----------|-------------|--------------|-----------|
| Unit Tests | 1.22.x, 1.23.4 | N/A | linux |
| Acceptance Tests | 1.23.4 | 1.0.5, 1.1.0, latest | linux |
| Builds | 1.23.4 | N/A | linux, windows, darwin |

## Writing Tests

### Unit Test Guidelines

```go
func TestClient_GetWorkflow(t *testing.T) {
    // Setup
    mockWorkflow := Workflow{
        ID:   "test-id",
        Name: "Test Workflow",
    }

    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify request
        if r.URL.Path != "/api/v1/workflows/test-id" {
            t.Errorf("Expected path '/api/v1/workflows/test-id', got %s", r.URL.Path)
        }
        
        // Return mock response
        json.NewEncoder(w).Encode(mockWorkflow)
    }))
    defer server.Close()

    // Test
    client := createTestClient(server.URL)
    result, err := client.GetWorkflow("test-id")

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "test-id", result.ID)
}
```

### Acceptance Test Guidelines

```go
func TestAccWorkflowResource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccWorkflowResourceConfig("test-workflow"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("n8n_workflow.test", "name", "test-workflow"),
                    resource.TestCheckResourceAttrSet("n8n_workflow.test", "id"),
                ),
            },
        },
    })
}
```

### Test Data Management

- Use deterministic test data
- Clean up resources after tests
- Avoid hardcoded IDs or timestamps
- Use table-driven tests for multiple scenarios

### Mock Guidelines

- Mock external dependencies
- Use httptest.NewServer for HTTP client tests
- Verify request parameters and headers
- Return realistic responses

## Troubleshooting

### Common Issues

#### Acceptance Tests Failing

1. **n8n not ready**: Wait longer for n8n startup
   ```bash
   curl -f http://localhost:5678/healthz
   ```

2. **API key issues**: Verify the API key is correct
   ```bash
   curl -H "X-N8N-API-KEY: test-api-key-12345" http://localhost:5678/api/v1/workflows
   ```

3. **Port conflicts**: Change ports in docker-compose.test.yml

#### Unit Tests Failing

1. **Race conditions**: Use `go test -race`
2. **Environment issues**: Check Go version and dependencies
3. **Mock setup**: Verify mock server configurations

### Debug Mode

Enable verbose output:

```bash
# Verbose test output
go test -v ./...

# Race detection
go test -race ./...

# Build with debug info
go build -gcflags="-N -l" .
```

### Test Coverage

Check coverage reports:

```bash
make coverage
open coverage.html
```

### CI Debug

For GitHub Actions debugging:

1. Check workflow logs
2. Verify environment variables
3. Test locally with same conditions
4. Use `act` to run GitHub Actions locally

## Performance Testing

### Benchmarks

Run performance benchmarks:

```bash
go test -bench=. -benchmem ./...
```

### Load Testing

For acceptance tests with load:

```bash
# Parallel acceptance tests
TF_ACC=1 go test -parallel 4 ./internal/provider/
```

## Security Testing

### Vulnerability Scanning

```bash
make security
```

This runs:
- `gosec`: Security analyzer
- `govulncheck`: Vulnerability database check

### Dependency Scanning

```bash
go list -m -u all
go mod audit  # If available
```

## Maintenance

### Updating Test Dependencies

```bash
go get -u github.com/hashicorp/terraform-plugin-testing
go mod tidy
```

### Test Data Cleanup

Regularly review and clean up:
- Test fixtures
- Mock data
- Temporary files
- Docker volumes

### Performance Monitoring

Monitor test performance:
- Test execution time
- Coverage percentage
- CI pipeline duration
- Resource usage

---

For more information, see:
- [Contributing Guidelines](CONTRIBUTING.md)
- [Development Setup](README.md#development)
- [Issue Templates](.github/ISSUE_TEMPLATE/)