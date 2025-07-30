#!/bin/bash

# Script to run acceptance tests with Docker n8n instance

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
COMPOSE_FILE="docker-compose.test.yml"
N8N_VERSION="latest"
TIMEOUT=120

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to cleanup
cleanup() {
    print_info "Cleaning up test environment..."
    docker-compose -f "$COMPOSE_FILE" down -v --remove-orphans
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --n8n-version)
            N8N_VERSION="$2"
            shift 2
            ;;
        --timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --n8n-version VERSION   n8n Docker image version (default: latest)"
            echo "  --timeout SECONDS       Timeout for n8n startup (default: 120)"
            echo "  --help                  Show this help message"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Check if Docker and docker-compose are available
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed or not in PATH"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    print_error "docker-compose is not installed or not in PATH"
    exit 1
fi

# Check if compose file exists
if [[ ! -f "$COMPOSE_FILE" ]]; then
    print_error "Docker compose file not found: $COMPOSE_FILE"
    exit 1
fi

print_info "Starting test environment with n8n version: $N8N_VERSION"

# Set n8n version in environment
export N8N_VERSION

# Start services
docker-compose -f "$COMPOSE_FILE" up -d

print_info "Waiting for services to be healthy..."

# Wait for n8n to be ready
print_info "Waiting for n8n to be ready (timeout: ${TIMEOUT}s)..."
start_time=$(date +%s)

while true; do
    current_time=$(date +%s)
    elapsed=$((current_time - start_time))
    
    if [[ $elapsed -gt $TIMEOUT ]]; then
        print_error "Timeout waiting for n8n to be ready"
        print_info "Checking service logs..."
        docker-compose -f "$COMPOSE_FILE" logs n8n
        exit 1
    fi
    
    if curl -f -s http://localhost:5678/healthz > /dev/null 2>&1; then
        print_info "n8n is ready!"
        break
    fi
    
    print_info "Still waiting for n8n... (${elapsed}s elapsed)"
    sleep 2
done

# Additional delay to ensure n8n is fully initialized
sleep 5

print_info "Running acceptance tests..."

# Set environment variables for acceptance tests
export TF_ACC=1
export N8N_BASE_URL="http://localhost:5678"
export N8N_API_KEY="test-api-key-12345"

# Run the tests
if make testacc; then
    print_info "All acceptance tests passed!"
else
    print_error "Some acceptance tests failed"
    print_info "Checking n8n logs for debugging..."
    docker-compose -f "$COMPOSE_FILE" logs n8n | tail -50
    exit 1
fi

print_info "Acceptance tests completed successfully!"