variable "n8n_api_key" {
  description = "API key for n8n authentication"
  type        = string
  sensitive   = true
}

variable "n8n_base_url" {
  description = "Base URL of the n8n instance"
  type        = string
  default     = "http://localhost:5678"
}