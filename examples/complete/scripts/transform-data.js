// Transform external API data for database storage
const data = items[0].json;

// Handle different data structures
let processedItems = [];

if (Array.isArray(data)) {
  // Data is an array of objects
  processedItems = data.map((item, index) => ({
    json: {
      data: JSON.stringify(item),
      processed_at: new Date().toISOString(),
      source: 'external_api',
      record_id: item.id || `record_${index}`,
      record_type: item.type || 'unknown'
    }
  }));
} else if (data && typeof data === 'object') {
  // Data is a single object
  processedItems = [{
    json: {
      data: JSON.stringify(data),
      processed_at: new Date().toISOString(),
      source: 'external_api',
      record_id: data.id || 'single_record',
      record_type: data.type || 'single'
    }
  }];
} else {
  // Fallback for other data types
  processedItems = [{
    json: {
      data: JSON.stringify(data),
      processed_at: new Date().toISOString(),
      source: 'external_api',
      record_id: 'raw_data',
      record_type: 'raw'
    }
  }];
}

// Add summary information for notifications
const summary = {
  json: {
    count: processedItems.length,
    processed_at: new Date().toISOString(),
    source: 'external_api'
  }
};

return [...processedItems, summary];