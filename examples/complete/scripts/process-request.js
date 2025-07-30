// Process validated webhook request
const request = items[0].json;

let result = {};

switch (request.action) {
  case 'create':
    result = {
      id: `created_${Date.now()}`,
      action: 'create',
      status: 'success',
      message: 'Resource created successfully',
      data: {
        ...request.data,
        created_at: new Date().toISOString(),
        status: 'active'
      }
    };
    break;

  case 'update':
    result = {
      id: request.data.id,
      action: 'update',
      status: 'success',
      message: 'Resource updated successfully',
      data: {
        ...request.data,
        updated_at: new Date().toISOString(),
        modified_fields: Object.keys(request.data).filter(key => key !== 'id')
      }
    };
    break;

  case 'delete':
    result = {
      id: request.data.id,
      action: 'delete',
      status: 'success',
      message: 'Resource deleted successfully',
      data: {
        deleted_at: new Date().toISOString(),
        soft_delete: true
      }
    };
    break;

  case 'query':
    result = {
      action: 'query',
      status: 'success',
      message: 'Query executed successfully',
      data: {
        query_params: request.data,
        executed_at: new Date().toISOString(),
        results: [
          // Mock results - replace with actual query logic
          { id: 1, name: 'Sample Item 1', status: 'active' },
          { id: 2, name: 'Sample Item 2', status: 'inactive' }
        ],
        total_count: 2
      }
    };
    break;

  default:
    result = {
      action: request.action,
      status: 'error',
      message: 'Unsupported action',
      data: null
    };
}

// Add processing metadata
result.processed_at = new Date().toISOString();
result.request_id = request.request_id;
result.processing_time_ms = Date.now() - new Date(request.validated_at).getTime();

return [{ json: result }];