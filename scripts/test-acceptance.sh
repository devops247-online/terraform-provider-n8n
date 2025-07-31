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
N8N_PORT="5678"
TIMEOUT=120
SETUP_TIMEOUT=60

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
    docker compose -f "$COMPOSE_FILE" down -v --remove-orphans
}

# Function to setup n8n authentication
setup_n8n_authentication() {
    local setup_payload='{
        "email": "admin@example.com",
        "firstName": "Admin",
        "lastName": "User",
        "password": "Admin123!"
    }'
    
    print_info "Attempting n8n owner setup via various API endpoints..."
    
    # List of possible setup endpoints to try (prioritize known working ones)
    local endpoints=(
        "/rest/owner/setup"
        "/api/v1/owner/setup"
        "/api/v1/setup" 
        "/rest/setup"
        "/api/v1/owner"
        "/rest/owner"
        "/api/v1/users"
    )
    
    local setup_successful=false
    for endpoint in "${endpoints[@]}"; do
        print_info "Trying setup endpoint: $endpoint"
        
        setup_response=$(curl -s -w "HTTPSTATUS:%{http_code}" \
            -X POST "http://localhost:${N8N_PORT}$endpoint" \
            -H "Content-Type: application/json" \
            -H "Accept: application/json" \
            -d "$setup_payload" 2>/dev/null)
        
        setup_http_code=$(echo "$setup_response" | grep -o "HTTPSTATUS:[0-9]*" | cut -d: -f2)
        setup_body=$(echo "$setup_response" | sed -E 's/HTTPSTATUS:[0-9]*$//')
        
        print_info "Endpoint $endpoint response code: $setup_http_code"
        print_info "Endpoint $endpoint response body: ${setup_body:0:200}..."
        
        if [[ "$setup_http_code" = "200" ]] || [[ "$setup_http_code" = "201" ]]; then
            print_info "✅ n8n owner setup successful via $endpoint"
            setup_successful=true
            break
        elif [[ "$setup_http_code" = "400" ]] && echo "$setup_body" | grep -q "owner.*already.*setup\|already.*initialized\|already.*exists\|user.*already.*exists"; then
            print_info "ℹ️ n8n owner already exists (via $endpoint), continuing..."
            setup_successful=true
            break
        elif [[ "$setup_http_code" != "404" ]] && [[ "$setup_http_code" != "405" ]]; then
            print_warning "Endpoint $endpoint returned $setup_http_code, might be usable"
        fi
    done
    
    if [[ "$setup_successful" != "true" ]]; then
        print_warning "All setup endpoints failed, trying alternative approaches..."
        
        # Check web interface
        web_response=$(curl -s -w "HTTPSTATUS:%{http_code}" http://localhost:${N8N_PORT}/ 2>/dev/null)
        web_http_code=$(echo "$web_response" | grep -o "HTTPSTATUS:[0-9]*" | cut -d: -f2)
        
        if [[ "$web_http_code" = "200" ]]; then
            print_info "Web interface accessible, trying /rest/owner/setup via web..."
            
            web_setup_response=$(curl -s -w "HTTPSTATUS:%{http_code}" \
                -X POST http://localhost:${N8N_PORT}/rest/owner/setup \
                -H "Content-Type: application/json" \
                -H "Accept: application/json" \
                -d "$setup_payload" 2>/dev/null)
            
            web_setup_http_code=$(echo "$web_setup_response" | grep -o "HTTPSTATUS:[0-9]*" | cut -d: -f2)
            web_setup_body=$(echo "$web_setup_response" | sed -E 's/HTTPSTATUS:[0-9]*$//')
            
            if [[ "$web_setup_http_code" = "200" ]] || [[ "$web_setup_http_code" = "201" ]]; then
                print_info "✅ Web-based setup successful"
                setup_successful=true
            fi
        fi
    fi
    
    if [[ "$setup_successful" != "true" ]]; then
        print_error "Failed to setup n8n owner"
        return 1
    fi
    
    # Now try to generate API key
    print_info "Attempting to generate API key..."
    
    # Try multiple login endpoints first
    local login_endpoints=(
        "/api/v1/auth/login"
        "/api/v1/login"
        "/rest/login"
        "/rest/auth/login"
    )
    
    local login_successful=false
    for login_endpoint in "${login_endpoints[@]}"; do
        print_info "Trying login endpoint: $login_endpoint"
        
        # Try different login payload formats based on endpoint
        local login_payload
        if [[ "$login_endpoint" == "/rest/login" ]]; then
            login_payload='{"emailOrLdapLoginId": "admin@example.com", "password": "Admin123!"}'
        else
            login_payload='{"email": "admin@example.com", "password": "Admin123!"}'
        fi
        
        login_response=$(curl -s -w "HTTPSTATUS:%{http_code}" \
            -X POST "http://localhost:${N8N_PORT}$login_endpoint" \
            -H "Content-Type: application/json" \
            -c /tmp/n8n_cookies.txt \
            -d "$login_payload" 2>/dev/null)
        
        login_http_code=$(echo "$login_response" | grep -o "HTTPSTATUS:[0-9]*" | cut -d: -f2)
        login_body=$(echo "$login_response" | sed -E 's/HTTPSTATUS:[0-9]*$//')
        
        print_info "Login endpoint $login_endpoint response code: $login_http_code"
        print_info "Login endpoint $login_endpoint response body: ${login_body:0:200}..."
        
        if [[ "$login_http_code" = "200" ]]; then
            print_info "✅ Login successful via $login_endpoint"
            login_successful=true
            break
        elif [[ "$login_http_code" != "404" ]] && [[ "$login_http_code" != "405" ]]; then
            print_warning "Login endpoint $login_endpoint returned $login_http_code, might need different payload"
        fi
    done
    
    if [[ "$login_successful" = "true" ]]; then
        # Try to generate API key
        local api_key_endpoints=(
            "/api/v1/me/api-key"
            "/api/v1/me/api-keys"
            "/rest/me/api-key"
            "/rest/me/api-keys"
            "/api/v1/auth/api-key"
            "/rest/auth/api-key"
        )
        
        for api_endpoint in "${api_key_endpoints[@]}"; do
            print_info "Trying API key endpoint: $api_endpoint"
            
            api_key_response=$(curl -s -w "HTTPSTATUS:%{http_code}" \
                -X POST "http://localhost:${N8N_PORT}$api_endpoint" \
                -b /tmp/n8n_cookies.txt \
                -H "Content-Type: application/json" \
                -d '{"name": "terraform-provider"}' 2>/dev/null)
            
            api_key_http_code=$(echo "$api_key_response" | grep -o "HTTPSTATUS:[0-9]*" | cut -d: -f2)
            api_key_body=$(echo "$api_key_response" | sed -E 's/HTTPSTATUS:[0-9]*$//')
            
            print_info "API key endpoint $api_endpoint response code: $api_key_http_code"
            print_info "API key endpoint $api_endpoint response body: ${api_key_body:0:200}..."
            
            if [[ "$api_key_http_code" = "200" ]] || [[ "$api_key_http_code" = "201" ]]; then
                print_info "✅ API key generation successful via $api_endpoint"
                
                # Extract API key from response
                if command -v jq &> /dev/null; then
                    api_key=$(echo "$api_key_body" | jq -r '.apiKey // .key // .token // empty' 2>/dev/null)
                else
                    # Fallback without jq - try to extract common patterns
                    api_key=$(echo "$api_key_body" | grep -o '"apiKey":"[^"]*"' | cut -d'"' -f4)
                    if [[ -z "$api_key" ]]; then
                        api_key=$(echo "$api_key_body" | grep -o '"key":"[^"]*"' | cut -d'"' -f4)
                    fi
                    if [[ -z "$api_key" ]]; then
                        api_key=$(echo "$api_key_body" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
                    fi
                fi
                
                if [[ -n "$api_key" ]] && [[ "$api_key" != "null" ]]; then
                    print_info "✅ API key extracted successfully: ${api_key:0:8}..."
                    export N8N_API_KEY="$api_key"
                    return 0
                else
                    print_warning "Could not extract API key from response"
                fi
            elif [[ "$api_key_http_code" != "404" ]] && [[ "$api_key_http_code" != "405" ]]; then
                print_warning "API key endpoint $api_endpoint returned $api_key_http_code, might need different payload"
            fi
        done
        
        # If API key generation failed, try creating API key via correct authenticated endpoint
        print_warning "Direct API key generation failed, trying authenticated /api-keys endpoint..."
        
        # Test if n8n public API is enabled first
        api_enabled_response=$(curl -s -w "HTTPSTATUS:%{http_code}" \
            -X GET "http://localhost:${N8N_PORT}/api-keys" \
            -b /tmp/n8n_cookies.txt \
            -H "Accept: application/json" 2>/dev/null)
        
        api_enabled_http_code=$(echo "$api_enabled_response" | grep -o "HTTPSTATUS:[0-9]*" | cut -d: -f2)
        print_info "API enablement check response code: $api_enabled_http_code"
        
        if [[ "$api_enabled_http_code" = "404" ]]; then
            print_warning "n8n public API appears to be disabled (404 on GET /api-keys)"
            print_warning "Trying to enable API or use alternative approach..."
        fi
        
        # Try creating API key with correct endpoint and payload
        api_keys_response=$(curl -s -w "HTTPSTATUS:%{http_code}" \
            -X POST "http://localhost:${N8N_PORT}/api-keys" \
            -b /tmp/n8n_cookies.txt \
            -H "Content-Type: application/json" \
            -d '{"label": "terraform-provider-test", "scopes": ["workflow:read", "workflow:create", "workflow:update", "workflow:delete"]}' 2>/dev/null)
        
        api_keys_http_code=$(echo "$api_keys_response" | grep -o "HTTPSTATUS:[0-9]*" | cut -d: -f2)
        api_keys_body=$(echo "$api_keys_response" | sed -E 's/HTTPSTATUS:[0-9]*$//')
        
        print_info "API keys endpoint response code: $api_keys_http_code"
        print_info "API keys endpoint response body: ${api_keys_body:0:200}..."
        
        if [[ "$api_keys_http_code" = "200" ]] || [[ "$api_keys_http_code" = "201" ]]; then
            print_info "✅ API key created via authenticated endpoint"
            
            # Extract API key from response (n8n returns JWT tokens)
            if command -v jq &> /dev/null; then
                # Try multiple possible response formats
                api_key=$(echo "$api_keys_body" | jq -r '.rawApiKey // .apiKey // .key // .token // empty' 2>/dev/null)
            else
                # Fallback without jq - try to extract JWT patterns and common fields
                api_key=$(echo "$api_keys_body" | grep -o '"rawApiKey":"[^"]*"' | cut -d'"' -f4)
                if [[ -z "$api_key" ]]; then
                    api_key=$(echo "$api_keys_body" | grep -o '"apiKey":"[^"]*"' | cut -d'"' -f4)
                fi
                if [[ -z "$api_key" ]]; then
                    api_key=$(echo "$api_keys_body" | grep -o '"key":"[^"]*"' | cut -d'"' -f4)
                fi
                if [[ -z "$api_key" ]]; then
                    # Look for JWT pattern (starts with eyJ)
                    api_key=$(echo "$api_keys_body" | grep -o '"[^"]*":"eyJ[^"]*"' | cut -d'"' -f4)
                fi
            fi
            
            if [[ -n "$api_key" ]] && [[ "$api_key" != "null" ]]; then
                print_info "✅ API key extracted successfully: ${api_key:0:8}..."
                export N8N_API_KEY="$api_key"
                return 0
            else
                print_warning "Could not extract API key from response"
            fi
        fi
        
        # If API key generation still failed, try session-based authentication
        print_warning "API key generation failed, testing session-based authentication..."
        
        session_test_response=$(curl -s -w "HTTPSTATUS:%{http_code}" \
            -b /tmp/n8n_cookies.txt \
            http://localhost:${N8N_PORT}/api/v1/workflows 2>/dev/null)
        
        session_test_http_code=$(echo "$session_test_response" | grep -o "HTTPSTATUS:[0-9]*" | cut -d: -f2)
        
        if [[ "$session_test_http_code" = "200" ]]; then
            print_info "✅ Session-based API access works, using cookie authentication"
            export N8N_USE_SESSION_AUTH="true"
            export N8N_COOKIE_FILE="/tmp/n8n_cookies.txt"
            # Copy cookies for later use
            cp /tmp/n8n_cookies.txt /tmp/n8n_session_cookies.txt 2>/dev/null || true
            return 0
        fi
    fi
    
    print_warning "Authentication setup completed but API key generation failed"
    print_warning "Will attempt to run tests with basic configuration"
    return 0
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --n8n-version)
            N8N_VERSION="$2"
            # Use version-specific compose file if it exists
            if [[ -f "docker-compose.n8n-${N8N_VERSION}.yml" ]]; then
                COMPOSE_FILE="docker-compose.n8n-${N8N_VERSION}.yml"
                # Set version-specific ports to avoid conflicts
                if [[ "$N8N_VERSION" == "latest" ]]; then
                    N8N_PORT="5679"
                elif [[ "$N8N_VERSION" == "1.1.0" ]]; then
                    N8N_PORT="5680"
                else
                    N8N_PORT="5678"
                fi
            fi
            shift 2
            ;;
        --timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        --setup-timeout)
            SETUP_TIMEOUT="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --n8n-version VERSION   n8n Docker image version (default: latest)"
            echo "  --timeout SECONDS       Timeout for n8n startup (default: 120)"
            echo "  --setup-timeout SECONDS Timeout for n8n setup (default: 60)"
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

# Check if docker compose (v2) or docker-compose is available
if ! docker compose version &> /dev/null && ! command -v docker-compose &> /dev/null; then
    print_error "Docker Compose is not installed or not in PATH"
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
docker compose -f "$COMPOSE_FILE" up -d

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
        docker compose -f "$COMPOSE_FILE" logs n8n
        exit 1
    fi
    
    # Try multiple health check endpoints
    if curl -f -s http://localhost:${N8N_PORT}/healthz > /dev/null 2>&1 || \
       curl -f -s http://localhost:${N8N_PORT}/healthz/readiness > /dev/null 2>&1 || \
       curl -f -s http://localhost:${N8N_PORT}/ > /dev/null 2>&1; then
        print_info "n8n is ready!"
        break
    fi
    
    print_info "Still waiting for n8n... (${elapsed}s elapsed)"
    sleep 2
done

# Additional delay to ensure n8n is fully initialized
sleep 5

# Setup n8n owner and get API key
print_info "Setting up n8n owner and generating API key..."
if ! setup_n8n_authentication; then
    print_error "Failed to set up n8n authentication"
    print_info "Checking n8n logs for debugging..."
    docker compose -f "$COMPOSE_FILE" logs n8n | tail -50
    exit 1
fi

print_info "Running acceptance tests..."

# Set environment variables for acceptance tests
export TF_ACC=1
export N8N_BASE_URL="http://localhost:${N8N_PORT}"

# Use the API key from authentication setup or fall back to email/password
if [[ -z "$N8N_API_KEY" ]]; then
    if [[ "$N8N_USE_SESSION_AUTH" == "true" ]]; then
        print_info "Using session-based authentication with cookies"
        export N8N_COOKIE_FILE="/tmp/n8n_session_cookies.txt"
    else
        print_info "No API key available, using email/password authentication"
        export N8N_EMAIL="admin@example.com"
        export N8N_PASSWORD="Admin123!"
    fi
else
    print_info "Using API key authentication: ${N8N_API_KEY:0:8}..."
fi

# Run the tests
if make testacc; then
    print_info "All acceptance tests passed!"
else
    print_error "Some acceptance tests failed"
    print_info "Checking n8n logs for debugging..."
    docker compose -f "$COMPOSE_FILE" logs n8n | tail -50
    exit 1
fi

print_info "Acceptance tests completed successfully!"