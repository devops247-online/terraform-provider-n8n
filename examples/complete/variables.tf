variable "n8n_base_url" {
  description = "Base URL of the n8n instance"
  type        = string
  default     = "http://localhost:5678"
}

variable "n8n_api_key" {
  description = "API key for n8n authentication"
  type        = string
  sensitive   = true
}

variable "external_api_token" {
  description = "Bearer token for external API"
  type        = string
  sensitive   = true
}

variable "external_api_url" {
  description = "URL of the external API to fetch data from"
  type        = string
  default     = "https://api.example.com/data"
}

variable "slack_webhook_url" {
  description = "Slack webhook URL for notifications"
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
  default     = "n8n_data"
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

variable "db_table" {
  description = "Database table for storing processed data"
  type        = string
  default     = "processed_data"
}

variable "processing_schedule" {
  description = "Cron expression for data processing schedule"
  type        = string
  default     = "0 9 * * *" # Daily at 9 AM
}

variable "activate_workflows" {
  description = "Whether to activate workflows immediately"
  type        = bool
  default     = false
}

variable "allowed_origins" {
  description = "Allowed origins for webhook CORS"
  type        = string
  default     = "*"
}