# UMA API Documentation

The UMA API provides comprehensive access to Unraid system monitoring, management, and control capabilities through a RESTful interface with OpenAPI 3.0.3 specification.

## API Overview

- **Base URL**: `http://your-unraid-ip:34600/api/v1`
- **Format**: JSON
- **Authentication**: None (runs on trusted network)
- **Versioning**: Accept header-based (`application/vnd.uma.v1+json`)
- **Compression**: Gzip supported
- **Documentation**: Interactive Swagger UI available

## Quick Start

### 1. Access API Documentation
Visit the interactive Swagger UI:
```
http://your-unraid-ip:34600/api/v1/docs
```

### 2. Get System Health
```bash
curl -H "X-Request-ID: health-check" \
     http://your-unraid-ip:34600/api/v1/health
```

### 3. List Docker Containers
```bash
curl -H "Accept: application/vnd.uma.v1+json" \
     "http://your-unraid-ip:34600/api/v1/docker/containers?page=1&limit=10"
```

### 4. Bulk Container Management
```bash
curl -X POST \
     -H "Content-Type: application/json" \
     -H "X-Request-ID: bulk-start" \
     -d '{"container_ids": ["plex", "nginx"]}' \
     http://your-unraid-ip:34600/api/v1/docker/containers/bulk/start
```

## API Features

### Request Tracking
Every request can include a custom request ID for tracking:
```bash
curl -H "X-Request-ID: my-unique-id" http://your-unraid-ip:34600/api/v1/health
```

### Response Compression
Enable gzip compression for large responses:
```bash
curl -H "Accept-Encoding: gzip" http://your-unraid-ip:34600/api/v1/storage/disks
```

### API Versioning
Specify API version in Accept header:
```bash
curl -H "Accept: application/vnd.uma.v1+json" http://your-unraid-ip:34600/api/v1/health
```

### Pagination
Most list endpoints support pagination:
```bash
curl "http://your-unraid-ip:34600/api/v1/docker/containers?page=2&limit=5"
```

## Response Format

All API responses follow a standardized format:

### Success Response
```json
{
  "data": {
    // Response data here
  },
  "meta": {
    "request_id": "unique-request-id",
    "version": "v1",
    "timestamp": "2025-06-15T23:00:00Z"
  },
  "pagination": {
    "page": 1,
    "per_page": 10,
    "total": 25,
    "has_more": true,
    "total_pages": 3
  }
}
```

### Error Response
```json
{
  "error": "Validation failed: container ID cannot be empty",
  "meta": {
    "request_id": "unique-request-id",
    "version": "v1",
    "timestamp": "2025-06-15T23:00:00Z"
  }
}
```

## Endpoint Categories

### System Monitoring
- Health checks and system status
- CPU, memory, and performance metrics
- Temperature and hardware sensors
- Network interface information
- UPS status and power management
- GPU monitoring and statistics

### Storage Management
- Array disk information and health
- Cache disk status and usage
- Boot disk information
- SMART data and disk health
- Storage usage statistics

### Docker Management
- Container listing and status
- Bulk operations (start, stop, restart)
- Container logs and information
- Docker events and monitoring

### Real-time Monitoring
- WebSocket endpoints for live data
- System statistics streaming
- Docker event streaming
- Storage status updates

## Error Handling

The API uses standard HTTP status codes:

- `200 OK` - Request successful
- `400 Bad Request` - Invalid request parameters
- `404 Not Found` - Resource not found
- `405 Method Not Allowed` - HTTP method not supported
- `500 Internal Server Error` - Server error

Error responses include descriptive messages:
```json
{
  "error": "validation failed: at least 1 container ID is required"
}
```

## Rate Limiting

Currently, no rate limiting is implemented. The API is designed for trusted network environments.

## Next Steps

- **[Complete Endpoint Reference](endpoints.md)** - All available endpoints
- **[WebSocket Guide](websockets.md)** - Real-time monitoring
- **[Bulk Operations](bulk-operations.md)** - Efficient container management
- **[OpenAPI Guide](openapi-guide.md)** - Using Swagger UI

## Examples

See the [endpoints documentation](endpoints.md) for comprehensive examples of all API endpoints.
