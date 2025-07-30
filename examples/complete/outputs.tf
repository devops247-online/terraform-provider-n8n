output "data_processor_workflow_id" {
  description = "ID of the data processing workflow"
  value       = n8n_workflow.data_processor.id
}

output "error_handler_workflow_id" {
  description = "ID of the error handler workflow"
  value       = n8n_workflow.error_handler.id
}

output "api_webhook_workflow_id" {
  description = "ID of the API webhook workflow"
  value       = n8n_workflow.api_webhook.id
}

output "webhook_url" {
  description = "URL for the API webhook endpoint"
  value       = "${var.n8n_base_url}/webhook/api/webhook"
}

output "external_api_credential_id" {
  description = "ID of the external API credential"
  value       = n8n_credential.external_api.id
  sensitive   = true
}

output "database_credential_id" {
  description = "ID of the database credential"
  value       = n8n_credential.database.id
  sensitive   = true
}

output "slack_credential_id" {
  description = "ID of the Slack webhook credential"
  value       = n8n_credential.slack_webhook.id
  sensitive   = true
}