# Troubleshooting Guide for n8n Terraform Provider

This guide helps you diagnose and resolve common issues when using the n8n Terraform provider.

## Common Issues and Solutions

### Connection and Authentication Issues

#### Error: Connection refused or timeout

**Symptoms:**
```
Error: failed to create n8n client: connection refused
Error: timeout waiting for n8n response
```

**Possible Causes and Solutions:**

1. **n8n instance not running or accessible**
   ```bash
   # Check if n8n is running
   curl -I http://localhost:5678/healthz
   
   # Check if n8n is listening on the correct port
   netstat -tuln | grep 5678
   ```

2. **Incorrect base_url configuration**
   ```hcl
   provider "n8n" {
     base_url = "http://localhost:5678"  # Ensure this matches your n8n instance
     api_key  = var.n8n_api_key
   }
   ```

3. **Firewall or network issues**
   ```bash
   # Test network connectivity
   telnet your-n8n-host 5678
   
   # Check if running in Docker with correct port mapping
   docker ps | grep n8n
   ```

4. **HTTPS/SSL configuration issues**
   ```hcl
   provider "n8n" {
     base_url            = "https://your-n8n-instance.com"
     insecure_skip_verify = true  # Only for testing with self-signed certificates
   }
   ```

#### Error: Authentication failed

**Symptoms:**
```
Error: authentication failed: invalid API key
Error: authentication failed: unauthorized
```

**Solutions:**

1. **Verify API key is correct**
   ```bash
   # Test API key manually
   curl -H "X-N8N-API-KEY: your-api-key" http://localhost:5678/api/v1/workflows
   ```

2. **Check API key permissions**
   - Ensure the API key has necessary permissions in n8n
   - Try with owner/admin account API key

3. **Try basic authentication as fallback**
   ```hcl
   provider "n8n" {
     base_url = "http://localhost:5678"
     email    = var.n8n_email
     password = var.n8n_password
   }
   ```

4. **Verify environment variables**
   ```bash
   echo $N8N_API_KEY
   echo $TF_VAR_n8n_api_key
   ```

### Resource Creation Issues

#### Error: Workflow creation failed

**Symptoms:**
```
Error: failed to create workflow: invalid JSON in nodes field
Error: workflow validation failed
```

**Solutions:**

1. **Validate JSON syntax**
   ```bash
   # Test your JSON with jq
   echo '{"nodes": {...}}' | jq .
   
   # Validate specific JSON files
   jq empty workflows/my-workflow-nodes.json
   ```

2. **Check node type compatibility**
   ```hcl
   # Ensure node types exist in your n8n version
   nodes = jsonencode({
     "HTTP Request" = {
       "type"        = "n8n-nodes-base.httpRequest",  # Correct node type
       "typeVersion" = 1,                             # Compatible version
       # ...
     }
   })
   ```

3. **Verify node parameters**
   ```hcl
   # Check required parameters for each node type
   "HTTP Request" = {
     "parameters" = {
       "url" = "https://api.example.com",  # Required parameter
       "method" = "GET"                     # Check parameter names
     }
   }
   ```

4. **Test workflow in n8n UI first**
   - Create the workflow manually in n8n
   - Export it to see the correct JSON structure
   - Use the exported JSON as a reference

#### Error: Credential creation failed

**Symptoms:**
```
Error: failed to create credential: invalid credential type
Error: credential validation failed: missing required field
```

**Solutions:**

1. **Verify credential type**
   ```bash
   # Check available credential types in n8n
   curl -H "X-N8N-API-KEY: your-key" http://localhost:5678/api/v1/credential-types
   ```

2. **Check required fields for credential type**
   ```hcl
   resource "n8n_credential" "example" {
     name = "Test Credential"
     type = "httpHeaderAuth"  # Correct type name
     
     data = jsonencode({
       name  = "Authorization",    # Required field
       value = "Bearer token123"  # Required field
     })
   }
   ```

3. **Validate credential data structure**
   ```bash
   # Test credential data JSON
   echo '{"name":"Authorization","value":"Bearer token"}' | jq .
   ```

### State Management Issues

#### Error: Resource already exists

**Symptoms:**
```
Error: workflow already exists with name "My Workflow"
```

**Solutions:**

1. **Import existing resource**
   ```bash
   # Get workflow ID from n8n
   terraform import n8n_workflow.existing_workflow <workflow-id>
   ```

2. **Use unique names**
   ```hcl
   resource "n8n_workflow" "example" {
     name = "${var.environment}-my-workflow-${random_id.suffix.hex}"
     # ...
   }
   
   resource "random_id" "suffix" {
     byte_length = 4
   }
   ```

3. **Check for naming conflicts**
   ```bash
   # List existing workflows
   curl -H "X-N8N-API-KEY: your-key" http://localhost:5678/api/v1/workflows
   ```

#### Error: Resource not found in state

**Symptoms:**
```
Error: resource not found in Terraform state
Error: workflow has been deleted outside of Terraform
```

**Solutions:**

1. **Refresh Terraform state**
   ```bash
   terraform refresh
   ```

2. **Re-import missing resources**
   ```bash
   terraform import n8n_workflow.missing_workflow <workflow-id>
   ```

3. **Recreate resources if necessary**
   ```bash
   # Remove from state and recreate
   terraform state rm n8n_workflow.missing_workflow
   terraform apply
   ```

### Provider Configuration Issues

#### Error: Provider initialization failed

**Symptoms:**
```
Error: failed to configure provider
Error: required provider not found
```

**Solutions:**

1. **Verify provider configuration**
   ```hcl
   terraform {
     required_providers {
       n8n = {
         source  = "devops247-online/n8n"
         version = "~> 1.0"
       }
     }
   }
   ```

2. **Update provider registry**
   ```bash
   terraform init -upgrade
   ```

3. **Clear provider cache**
   ```bash
   rm -rf .terraform/
   terraform init
   ```

### JSON and Template Issues

#### Error: Invalid JSON in workflow configuration

**Symptoms:**
```
Error: invalid character '}' looking for beginning of object key string
Error: unexpected end of JSON input
```

**Solutions:**

1. **Use proper JSON escaping**
   ```hcl
   # ❌ Incorrect
   nodes = jsonencode({
     "node": {
       "parameters": {
         "message": "Hello "world""  # Incorrect escaping
       }
     }
   })
   
   # ✅ Correct
   nodes = jsonencode({
     "node": {
       "parameters": {
         "message": "Hello \"world\""  # Proper escaping
       }
     }
   })
   ```

2. **Use templatefile for complex configurations**
   ```hcl
   nodes = templatefile("${path.module}/templates/workflow-nodes.json.tpl", {
     api_url        = var.api_url
     credential_id  = n8n_credential.api.id
   })
   ```

3. **Validate templates separately**
   ```bash
   # Test template rendering
   terraform console
   > templatefile("templates/workflow.json.tpl", {api_url = "https://api.example.com"})
   ```

### Performance and Timeout Issues

#### Error: Operation timed out

**Symptoms:**
```
Error: timeout while waiting for workflow creation
Error: context deadline exceeded
```

**Solutions:**

1. **Increase timeout settings**
   ```bash
   export TF_CLI_CONFIG_FILE=terraform.rc
   ```
   
   ```hcl
   # terraform.rc
   provider_installation {
     dev_overrides {
       "devops247-online/n8n" = "/path/to/provider"
     }
   }
   ```

2. **Check n8n instance performance**
   ```bash
   # Monitor n8n logs
   docker logs -f n8n-container
   
   # Check system resources
   top
   free -h
   df -h
   ```

3. **Optimize workflow complexity**
   - Reduce number of nodes in single workflow
   - Split complex workflows into smaller ones
   - Use simpler node configurations for testing

## Debugging Techniques

### Enable Detailed Logging

1. **Terraform debug logging**
   ```bash
   export TF_LOG=DEBUG
   export TF_LOG_PATH=terraform-debug.log
   terraform apply
   ```

2. **Provider-specific logging**
   ```bash
   export TF_LOG_PROVIDER=DEBUG
   ```

### Manual API Testing

Test the n8n API directly to isolate issues:

```bash
# Test authentication
curl -v -H "X-N8N-API-KEY: your-key" \
  http://localhost:5678/api/v1/workflows

# Test workflow creation
curl -v -X POST \
  -H "X-N8N-API-KEY: your-key" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Workflow","nodes":{},"connections":{}}' \
  http://localhost:5678/api/v1/workflows

# Test credential creation
curl -v -X POST \
  -H "X-N8N-API-KEY: your-key" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Cred","type":"httpHeaderAuth","data":{"name":"Auth","value":"test"}}' \
  http://localhost:5678/api/v1/credentials
```

### State Inspection

```bash
# Show current state
terraform show

# List resources in state
terraform state list

# Show specific resource
terraform state show n8n_workflow.example

# Check for drift
terraform plan -detailed-exitcode
```

### Workflow Testing

Create test workflows to validate configurations:

```hcl
resource "n8n_workflow" "connectivity_test" {
  name   = "Connectivity Test"
  active = false
  
  nodes = jsonencode({
    "Start" = {
      "parameters" = {},
      "type"       = "n8n-nodes-base.start",
      "typeVersion" = 1,
      "position"   = [240, 300]
    }
  })
  
  tags = ["test", "connectivity"]
}
```

## Version Compatibility Issues

### n8n Version Compatibility

Different n8n versions may have different:
- Node types and versions
- API endpoints
- Required parameters

**Solutions:**

1. **Check n8n version**
   ```bash
   curl -H "X-N8N-API-KEY: your-key" http://localhost:5678/api/v1/
   ```

2. **Use version-specific configurations**
   ```hcl
   locals {
     node_version = var.n8n_version >= "0.200.0" ? 2 : 1
   }
   
   resource "n8n_workflow" "version_aware" {
     nodes = jsonencode({
       "HTTP Request" = {
         "typeVersion" = local.node_version
         # ...
       }
     })
   }
   ```

### Provider Version Issues

**Symptoms:**
```
Error: incompatible provider version
Error: provider version constraint not satisfied
```

**Solutions:**

1. **Check provider version constraints**
   ```hcl
   terraform {
     required_providers {
       n8n = {
         source  = "devops247-online/n8n"
         version = "~> 1.0"  # Allow patch updates only
       }
     }
   }
   ```

2. **Upgrade provider**
   ```bash
   terraform init -upgrade
   ```

3. **Lock to specific version**
   ```hcl
   terraform {
     required_providers {
       n8n = {
         source  = "devops247-online/n8n"
         version = "= 1.0.0"  # Exact version
       }
     }
   }
   ```

## Recovery Procedures

### Recovering from Failed State

1. **Backup current state**
   ```bash
   cp terraform.tfstate terraform.tfstate.backup
   ```

2. **Manual state manipulation**
   ```bash
   # Remove problematic resource from state
   terraform state rm n8n_workflow.problematic
   
   # Import existing resource
   terraform import n8n_workflow.problematic <workflow-id>
   ```

3. **Force replacement**
   ```bash
   terraform apply -replace=n8n_workflow.example
   ```

### Disaster Recovery

1. **Export all workflows from n8n**
   ```bash
   # Export workflows via API
   curl -H "X-N8N-API-KEY: your-key" \
     http://localhost:5678/api/v1/workflows > workflows-backup.json
   ```

2. **Recreate from backup**
   ```bash
   # Remove all resources from state
   terraform destroy
   
   # Import from backup
   # (manual process or custom script)
   ```

## Getting Help

### Information to Collect

When reporting issues, include:

1. **Terraform version**: `terraform version`
2. **Provider version**: Check `terraform.lock.hcl`
3. **n8n version**: Check n8n UI or API
4. **Error messages**: Complete error output
5. **Configuration**: Sanitized Terraform files
6. **Debug logs**: With `TF_LOG=DEBUG`

### Resources

- **Provider Issues**: [GitHub Issues](https://github.com/devops247-online/terraform-provider-n8n/issues)
- **n8n Questions**: [n8n Community](https://community.n8n.io/)
- **Terraform Help**: [Terraform Documentation](https://www.terraform.io/docs/)

### Creating a Minimal Reproduction

```hcl
# minimal-repro.tf
terraform {
  required_providers {
    n8n = {
      source  = "devops247-online/n8n"
      version = "~> 1.0"
    }
  }
}

provider "n8n" {
  base_url = "http://localhost:5678"
  api_key  = "test-api-key"
}

resource "n8n_workflow" "minimal" {
  name = "Minimal Test Workflow"
  
  nodes = jsonencode({
    "Start" = {
      "type"       = "n8n-nodes-base.start",
      "typeVersion" = 1,
      "position"   = [240, 300]
    }
  })
}
```

This minimal configuration helps isolate the issue from complex setups.

## Prevention Strategies

1. **Use validation**: `terraform validate` before apply
2. **Test in development**: Never test directly in production
3. **Version control**: Keep all configurations in Git
4. **Backup state**: Regular state backups
5. **Monitor workflows**: Set up alerting for workflow failures
6. **Update gradually**: Test provider updates in development first
7. **Document changes**: Keep changelog of infrastructure changes

Following these troubleshooting steps and prevention strategies will help you maintain stable n8n infrastructure with Terraform.