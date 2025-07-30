variable "n8n_api_key" {
  description = "API key for n8n authentication"
  type        = string
  sensitive   = true
}

variable "data_api_url" {
  description = "URL for fetching data in scheduled workflow"
  type        = string
  default     = "https://api.example.com/data"
}

variable "results_api_url" {
  description = "URL for saving processed results"
  type        = string
  default     = "https://api.example.com/results"
}

variable "api_credential_id" {
  description = "ID of the API credential to use for HTTP requests"
  type        = string
  default     = ""
}