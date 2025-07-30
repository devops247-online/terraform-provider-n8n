terraform {
  required_providers {
    n8n = {
      source  = "devops247-online/n8n"
      version = "~> 1.0"
    }
  }
}

# Configure the n8n Provider
provider "n8n" {
  base_url = "http://localhost:5678" # Your n8n instance URL
  api_key  = var.n8n_api_key         # Set via environment variable or terraform.tfvars
}

# Example: HTTP Basic Authentication credential
resource "n8n_credential" "http_basic" {
  name = "HTTP Basic Auth Example"
  type = "httpBasicAuth"

  data = jsonencode({
    user     = var.http_username
    password = var.http_password
  })
}

# Example: API Key credential
resource "n8n_credential" "api_key" {
  name = "API Key Example"
  type = "httpHeaderAuth"

  data = jsonencode({
    name  = "X-API-Key"
    value = var.api_key_value
  })
}

# Example: OAuth2 credential
resource "n8n_credential" "oauth2" {
  name = "OAuth2 Example"
  type = "oAuth2Api"

  data = jsonencode({
    clientId       = var.oauth_client_id
    clientSecret   = var.oauth_client_secret
    accessTokenUrl = "https://api.example.com/oauth/token"
    authUrl        = "https://api.example.com/oauth/authorize"
    scope          = "read write"
    authentication = "body"
    grantType      = "authorizationCode"
  })
}

# Example: Database credential
resource "n8n_credential" "postgres" {
  name = "PostgreSQL Database"
  type = "postgres"

  data = jsonencode({
    host                   = var.db_host
    database               = var.db_name
    user                   = var.db_username
    password               = var.db_password
    port                   = var.db_port
    ssl                    = "require"
    allowUnauthorizedCerts = false
  })
}

# Example: AWS credential
resource "n8n_credential" "aws" {
  name = "AWS Credentials"
  type = "aws"

  data = jsonencode({
    accessKeyId     = var.aws_access_key_id
    secretAccessKey = var.aws_secret_access_key
    region          = var.aws_region
  })
}