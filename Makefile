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

.PHONY: all build clean test install uninstall fmt vet lint docs testacc

all: build

build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

test:
	$(GOTEST) -v ./...

testacc:
	TF_ACC=1 $(GOTEST) -v ./... -timeout 120m

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

# Development helpers
dev-setup: tools tidy

# Release preparation
pre-release: fmt vet lint test docs

# Local testing
local-install: build
	mkdir -p ~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)/
	cp $(BINARY_NAME) ~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)/$(BINARY_NAME)_v$(VERSION)

# Help
help:
	@echo "Available commands:"
	@echo "  build        - Build the provider binary"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run unit tests"
	@echo "  testacc      - Run acceptance tests"
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