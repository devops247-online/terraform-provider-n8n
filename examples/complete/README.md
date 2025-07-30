# Complete n8n Infrastructure Example

This directory contains a comprehensive example of managing a complete n8n automation infrastructure using Terraform. It demonstrates advanced patterns including credential management, workflow orchestration, error handling, and external integrations.

## Architecture Overview

This example creates:

- **Credentials**: API keys, database connections, and webhook credentials
- **Data Processing Pipeline**: Scheduled workflow that fetches, transforms, and stores data
- **Error Handling**: Centralized error workflow with notifications
- **API Webhook**: External integration endpoint with validation and processing
- **Monitoring**: Slack notifications for success and error states

## Components

### Credentials
- `external_api`: API credentials for external data source
- `slack_webhook`: Slack integration for notifications
- `database`: PostgreSQL database connection

### Workflows

#### Data Processing Pipeline
- **Trigger**: Cron-scheduled (configurable)
- **Flow**: Fetch → Transform → Store → Notify
- **Features**: Error handling, status notifications, database storage

#### Error Handler
- **Trigger**: Workflow errors
- **Flow**: Log → Notify team
- **Features**: Centralized error handling, detailed error reporting

#### API Webhook
- **Trigger**: HTTP POST requests
- **Flow**: Validate → Process → Respond
- **Features**: Input validation, CRUD operations, error responses

## Prerequisites

- n8n instance (v0.200.0+)
- PostgreSQL database
- Slack webhook URL
- External API with bearer token authentication
- Terraform >= 1.0

## Setup Instructions

### 1. Database Setup

Create the required database table:

```sql
CREATE TABLE processed_data (
    id SERIAL PRIMARY KEY,
    data JSONB NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    source VARCHAR(100) NOT NULL,
    record_id VARCHAR(255),
    record_type VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_processed_data_source ON processed_data(source);
CREATE INDEX idx_processed_data_processed_at ON processed_data(processed_at);
```

### 2. Configuration

Copy and customize the configuration:

```bash
cp terraform.tfvars.example terraform.tfvars
```

Edit `terraform.tfvars` with your values:

```hcl
# n8n Configuration
n8n_base_url = "https://your-n8n-instance.com"
n8n_api_key  = "your-api-key"

# External API
external_api_url   = "https://api.example.com/data"
external_api_token = "your-bearer-token"

# Database
db_host     = "localhost"
db_name     = "your_database"
db_username = "db_user"
db_password = "secure_password"

# Notifications
slack_webhook_url = "https://hooks.slack.com/services/..."

# Scheduling
processing_schedule = "0 9 * * *"  # Daily at 9 AM
activate_workflows  = false        # Start inactive for testing
```

### 3. Deployment

```bash
terraform init
terraform plan
terraform apply
```

### 4. Testing

Before activating workflows, test them manually:

1. **Test Data Processing**: Execute the workflow manually in n8n UI
2. **Test Webhook**: Send a test request to the webhook endpoint
3. **Test Error Handling**: Trigger an error to verify notifications

### 5. Activation

Once tested, activate workflows:

```bash
terraform apply -var="activate_workflows=true"
```

## Usage Examples

### API Webhook Usage

#### Create Resource
```bash
curl -X POST "https://your-n8n-instance.com/webhook/api/webhook" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "create",
    "data": {
      "name": "New Resource",
      "type": "example"
    }
  }'
```

#### Update Resource
```bash
curl -X POST "https://your-n8n-instance.com/webhook/api/webhook" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "update",
    "data": {
      "id": "resource_123",
      "name": "Updated Resource"
    }
  }'
```

#### Query Resources
```bash
curl -X POST "https://your-n8n-instance.com/webhook/api/webhook" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "query",
    "data": {
      "filter": "active",
      "limit": 10
    }
  }'
```

## Monitoring and Maintenance

### Workflow Execution History
Check execution history in n8n UI:
- Navigate to "Executions" tab
- Filter by workflow name
- Review success/failure rates

### Database Monitoring
Monitor data processing:
```sql
-- Check recent processing activity
SELECT 
    source,
    COUNT(*) as record_count,
    MAX(processed_at) as last_processed
FROM processed_data 
WHERE processed_at > NOW() - INTERVAL '24 hours'
GROUP BY source;

-- Check for processing errors (gaps in schedule)
SELECT DATE(processed_at) as date, COUNT(*) as records
FROM processed_data 
WHERE source = 'external_api'
AND processed_at > NOW() - INTERVAL '7 days'
GROUP BY DATE(processed_at)
ORDER BY date;
```

### Slack Notifications
Notifications are sent for:
- Successful data processing completions
- Workflow errors and failures
- API webhook errors

## Customization

### Adding New Data Sources
1. Create new credentials for the data source
2. Add new nodes to the data processing workflow
3. Update the transform script to handle new data formats

### Modifying Processing Schedule
Update the `processing_schedule` variable with a new cron expression:
- `"0 */6 * * *"` - Every 6 hours
- `"0 0 * * 0"` - Weekly on Sunday
- `"0 9 * * 1-5"` - Weekdays at 9 AM

### Custom Validation Rules
Modify `scripts/validate-input.js` to add:
- Additional required fields
- Custom validation logic
- Business rule validation

### Error Handling Enhancement
Extend the error handler workflow to:
- Send emails in addition to Slack
- Create support tickets
- Implement retry logic

## Security Best Practices

1. **Credentials**: Never commit sensitive values to version control
2. **Database**: Use connection pooling and read replicas for high load
3. **API Keys**: Rotate API keys regularly
4. **Webhooks**: Implement authentication for production webhooks
5. **Network**: Use HTTPS and restrict network access where possible

## Troubleshooting

### Common Issues

#### Workflow Not Triggering
- Check schedule trigger configuration
- Verify workflow is active
- Review n8n logs for errors

#### Database Connection Errors
- Verify database credentials
- Check network connectivity
- Ensure SSL configuration matches database settings

#### API Authentication Failures
- Verify API token validity
- Check token expiration
- Review API endpoint URLs

#### Slack Notifications Not Working
- Verify webhook URL format
- Check Slack workspace permissions
- Review message payload format

### Getting Help

- Check n8n execution logs
- Review Terraform state for resource IDs
- Monitor Slack notifications for error details
- Use n8n community forum for workflow-specific issues

## Production Considerations

- **Scaling**: Consider horizontal scaling for high-volume processing
- **Backup**: Implement regular database backups
- **Monitoring**: Add application performance monitoring
- **Logging**: Centralize logs for better observability
- **CI/CD**: Implement automated testing and deployment
- **Disaster Recovery**: Plan for infrastructure recovery scenarios