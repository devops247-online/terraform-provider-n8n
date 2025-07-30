# n8n_workflow (Resource)

Manages an n8n workflow. Workflows are the core automation units in n8n that define a series of nodes and their connections.

## Example Usage

### Basic Workflow

```terraform
resource "n8n_workflow" "example" {
  name   = "My Terraform Workflow"
  active = true
  
  nodes = jsonencode({
    "Start": {
      "parameters": {},
      "type": "n8n-nodes-base.start",
      "typeVersion": 1,
      "position": [240, 300]
    },
    "HTTP Request": {
      "parameters": {
        "url": "https://api.example.com/data",
        "responseFormat": "json"
      },
      "type": "n8n-nodes-base.httpRequest",
      "typeVersion": 1,
      "position": [440, 300]
    }
  })
  
  connections = jsonencode({
    "Start": {
      "main": [
        [
          {
            "node": "HTTP Request",
            "type": "main",
            "index": 0
          }
        ]
      ]
    }
  })
  
  tags = ["terraform", "api", "automation"]
}
```

### Advanced Workflow with Settings

```terraform
resource "n8n_workflow" "advanced" {
  name   = "Advanced Workflow"
  active = false
  
  nodes = jsonencode({
    # Complex node configuration
  })
  
  connections = jsonencode({
    # Node connections
  })
  
  settings = jsonencode({
    "executionOrder": "v1",
    "saveManualExecutions": true,
    "callerPolicy": "workflowsFromSameOwner"
  })
  
  static_data = jsonencode({
    "node:Schedule Trigger": {
      "recurrenceRules": []
    }
  })
  
  tags = ["advanced", "scheduled"]
}
```

## Schema

### Required

- `name` (String) The name of the workflow

### Optional

- `active` (Boolean) Whether the workflow is active and can be triggered. Defaults to `false`.
- `nodes` (String) JSON string containing the workflow nodes configuration
- `connections` (String) JSON string containing the workflow connections between nodes
- `settings` (String) JSON string containing workflow settings
- `static_data` (String) JSON string containing static data for the workflow
- `pinned_data` (String) JSON string containing pinned data for testing purposes
- `tags` (List of String) List of tags associated with the workflow

### Read-Only

- `id` (String) Workflow identifier
- `version_id` (String) Version identifier of the workflow
- `created_at` (String) Timestamp when the workflow was created
- `updated_at` (String) Timestamp when the workflow was last updated

## Import

Import is supported using the workflow ID:

```bash
terraform import n8n_workflow.example 1234567890abcdef
```

## JSON Field Guidelines

The `nodes`, `connections`, `settings`, `static_data`, and `pinned_data` fields expect JSON strings. Here are some guidelines:

### Nodes Configuration

The `nodes` field should contain a JSON object where each key is the node name and the value contains the node configuration:

```json
{
  "Node Name": {
    "parameters": {
      // Node-specific parameters
    },
    "type": "n8n-nodes-base.nodetype",
    "typeVersion": 1,
    "position": [x, y],
    "credentials": {
      // Credential references if needed
    }
  }
}
```

### Connections Configuration

The `connections` field defines how nodes are connected:

```json
{
  "Source Node": {
    "main": [
      [
        {
          "node": "Target Node",
          "type": "main",
          "index": 0
        }
      ]
    ]
  }
}
```

### Settings Configuration

Common workflow settings:

```json
{
  "executionOrder": "v1",
  "saveManualExecutions": true,
  "callerPolicy": "workflowsFromSameOwner",
  "errorWorkflow": "workflow-id-for-errors"
}
```

## Best Practices

1. **Version Control**: Store complex workflow JSON in separate files and use `file()` function:
   ```terraform
   nodes = file("${path.module}/workflows/my-workflow-nodes.json")
   ```

2. **Validation**: Use `terraform validate` to check JSON syntax before applying

3. **Testing**: Use `pinned_data` for testing workflows with static data

4. **Organization**: Use tags to categorize and organize workflows

5. **Activation**: Start with `active = false` and activate after testing