// Validate incoming webhook request
const request = items[0].json;

// Define validation rules
const requiredFields = ['action', 'data'];
const allowedActions = ['create', 'update', 'delete', 'query'];

let isValid = true;
let errors = [];

// Check required fields
for (const field of requiredFields) {
  if (!request.hasOwnProperty(field) || request[field] === null || request[field] === undefined) {
    isValid = false;
    errors.push(`Missing required field: ${field}`);
  }
}

// Validate action field
if (request.action && !allowedActions.includes(request.action)) {
  isValid = false;
  errors.push(`Invalid action: ${request.action}. Allowed actions: ${allowedActions.join(', ')}`);
}

// Validate data field structure
if (request.data && typeof request.data !== 'object') {
  isValid = false;
  errors.push('Data field must be an object');
}

// Additional validation based on action type
if (isValid && request.action) {
  switch (request.action) {
    case 'create':
    case 'update':
      if (!request.data.id && request.action === 'update') {
        isValid = false;
        errors.push('ID is required for update operations');
      }
      break;
    case 'delete':
      if (!request.data.id) {
        isValid = false;
        errors.push('ID is required for delete operations');
      }
      break;
    case 'query':
      // Query validation can be added here
      break;
  }
}

if (isValid) {
  // Return valid request data with additional metadata
  return [{
    json: {
      ...request,
      validated_at: new Date().toISOString(),
      request_id: `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      is_valid: true
    }
  }];
} else {
  // Return error information (will be sent to error branch)
  return [{
    json: {
      error: errors.join('; '),
      request: request,
      validated_at: new Date().toISOString(),
      is_valid: false
    }
  }];
}