# Basic n8n Terraform Provider Example

This example demonstrates the basic usage of the n8n Terraform provider to create a simple workflow.

## Prerequisites

1. A running n8n instance (local or remote)
2. An API key for your n8n instance
3. Terraform installed

## Setup

1. Copy the example variables file:
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

2. Edit `terraform.tfvars` with your n8n instance details:
   ```hcl
   n8n_base_url = "http://your-n8n-instance:5678"
   n8n_api_key = "your-api-key-here"
   ```

3. Initialize Terraform:
   ```bash
   terraform init
   ```

4. Plan the changes:
   ```bash
   terraform plan
   ```

5. Apply the configuration:
   ```bash
   terraform apply
   ```

## What this example creates

- A simple n8n workflow named "Example HTTP Workflow"
- The workflow contains two nodes:
  - A Start node (trigger)
  - An HTTP Request node that fetches data from httpbin.org
- The workflow is tagged with "terraform", "example", and "http"
- The workflow is set to active status

## Cleanup

To remove the created resources:

```bash
terraform destroy
```

## Next Steps

- Check out the [credentials example](../credentials/) to learn about managing API credentials
- See the [complete example](../complete/) for a more comprehensive setup
- Read the [provider documentation](../../docs/) for all available resources