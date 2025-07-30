variable "n8n_api_key" {
  description = "API key for n8n authentication"
  type        = string
  sensitive   = true
}

variable "http_username" {
  description = "Username for HTTP Basic Authentication"
  type        = string
  sensitive   = true
}

variable "http_password" {
  description = "Password for HTTP Basic Authentication"
  type        = string
  sensitive   = true
}

variable "api_key_value" {
  description = "API key value for header authentication"
  type        = string
  sensitive   = true
}

variable "oauth_client_id" {
  description = "OAuth2 client ID"
  type        = string
  sensitive   = true
}

variable "oauth_client_secret" {
  description = "OAuth2 client secret"
  type        = string
  sensitive   = true
}

variable "db_host" {
  description = "Database host"
  type        = string
  default     = "localhost"
}

variable "db_name" {
  description = "Database name"
  type        = string
  default     = "n8n_db"
}

variable "db_username" {
  description = "Database username"
  type        = string
  sensitive   = true
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}

variable "db_port" {
  description = "Database port"
  type        = number
  default     = 5432
}

variable "aws_access_key_id" {
  description = "AWS Access Key ID"
  type        = string
  sensitive   = true
}

variable "aws_secret_access_key" {
  description = "AWS Secret Access Key"
  type        = string
  sensitive   = true
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}