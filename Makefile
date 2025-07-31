# Build variables
BINARY_NAME=terraform-provider-n8n
VERSION?=dev
LDFLAGS=-ldflags "-X main.version=${VERSION}"

# Go related variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# Terraform related variables
TERRAFORM_PLUGINS_DIR=~/.terraform.d/plugins
HOSTNAME=registry.terraform.io
NAMESPACE=devops247-online
NAME=n8n
OS_ARCH=linux_amd64

.PHONY: all build clean test install uninstall fmt vet lint docs testacc pre-commit-install pre-commit-run

all: build

build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

test:
	$(GOTEST) -v ./...

testacc:
	@echo "Note: Acceptance tests require a running n8n instance with proper authentication."
	@echo "Set N8N_BASE_URL and N8N_API_KEY (or N8N_EMAIL/N8N_PASSWORD) environment variables."
	@echo "Set TF_ACC_SKIP=1 to skip acceptance tests."
	TF_ACC=1 $(GOTEST) -v ./... -timeout 120m

# Run acceptance tests with Docker n8n instance
testacc-docker:
	./scripts/test-acceptance.sh

# Run only unit tests (no acceptance tests)
test-unit:
	$(GOTEST) -v ./... -timeout 30m

# Run tests with acceptance tests skipped
test-skip-acc:
	TF_ACC=1 TF_ACC_SKIP=1 $(GOTEST) -v ./... -timeout 30m

install: build
	mkdir -p $(TERRAFORM_PLUGINS_DIR)/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)
	cp $(BINARY_NAME) $(TERRAFORM_PLUGINS_DIR)/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)/

uninstall:
	rm -rf $(TERRAFORM_PLUGINS_DIR)/$(HOSTNAME)/$(NAMESPACE)/$(NAME)

fmt:
	$(GOFMT) -s -w .

vet:
	$(GOCMD) vet ./...

lint:
	golangci-lint run

tidy:
	$(GOMOD) tidy

deps:
	$(GOMOD) download

docs:
	go generate

tools:
	go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

# Development helpers
dev-setup: tools tidy

# Security scanning
security: gosec vulncheck

gosec:
	gosec -quiet ./...

vulncheck:
	govulncheck ./...

# Coverage reporting
coverage:
	$(GOTEST) -v -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Release preparation
pre-release: fmt vet lint test security docs coverage

# Local testing
local-install: build
	mkdir -p ~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)/
	cp $(BINARY_NAME) ~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)/$(BINARY_NAME)_v$(VERSION)

# Pre-commit hooks
pre-commit-install:
	./scripts/install-pre-commit-hooks.sh

pre-commit-run:
	pre-commit run --all-files

# Help
help:
	@echo "Available commands:"
	@echo "  build        - Build the provider binary"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run unit tests"
	@echo "  test-unit    - Run only unit tests (no acceptance tests)"
	@echo "  test-skip-acc- Run tests with acceptance tests skipped"
	@echo "  testacc      - Run acceptance tests (requires n8n setup)"
	@echo "  testacc-docker- Run acceptance tests with Docker n8n"
	@echo "  install      - Install provider locally for testing"
	@echo "  uninstall    - Remove locally installed provider"
	@echo "  fmt          - Format Go code"
	@echo "  vet          - Run go vet"
	@echo "  lint         - Run golangci-lint"
	@echo "  tidy         - Clean up go.mod"
	@echo "  deps         - Download dependencies"
	@echo "  docs         - Generate documentation"
	@echo "  tools        - Install development tools"
	@echo "  dev-setup    - Set up development environment"
	@echo "  pre-release  - Run all checks before release"
	@echo "  local-install- Install provider for local testing"
	@echo "  pre-commit-install - Install pre-commit hooks"
	@echo "  pre-commit-run - Run pre-commit on all files"
