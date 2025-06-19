# UMA API Quick Start Guide

Get started with the UMA (Unraid Management Agent) REST API in minutes. This guide covers the essential endpoints and usage patterns for monitoring and managing your Unraid server.

## Prerequisites

- UMA plugin installed and running on Unraid
- API accessible at `http://your-unraid-ip:34600/api/v1`
- Basic understanding of REST APIs and HTTP methods

## API Basics

### Base URL
```
http://your-unraid-ip:34600/api/v1
```

### Authentication
No authentication required - UMA is designed for trusted network environments.

### Content Type
All requests and responses use JSON format:
```
Content-Type: application/json
Accept: application/vnd.uma.v1+json
```

### Request IDs
Include a unique request ID for debugging and tracing:
```
X-Request-ID: unique-identifier-123
```

## Essential Endpoints

### 1. Health Check

Check if UMA is running and healthy:

```bash
curl -H "X-Request-ID: health-check" \
     http://your-unraid-ip:34600/api/v1/health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-06-19T10:30:00Z",
  "dependencies": {
    "docker": "healthy",
    "storage": "healthy",
    "system": "healthy"
  },
  "uptime": "2h30m15s"
}
```

### 2. System Statistics

Get real-time system metrics:

```bash
curl -H "X-Request-ID: system-stats" \
     http://your-unraid-ip:34600/api/v1/system/resources
```

**Response:**
```json
{
  "cpu": {
    "usage_percent": 15.2,
    "cores": 8,
    "load_average": [0.5, 0.7, 0.8]
  },
  "memory": {
    "total_bytes": 34359738368,
    "used_bytes": 8589934592,
    "available_bytes": 25769803776,
    "usage_percent": 25.0
  },
  "uptime": "2h30m15s",
  "timestamp": "2025-06-19T10:30:00Z"
}
```

### 3. Storage Information

Monitor array and disk status:

```bash
curl -H "X-Request-ID: storage-info" \
     http://your-unraid-ip:34600/api/v1/storage/array
```

**Response:**
```json
{
  "status": "started",
  "total_size_bytes": 12000000000000,
  "used_bytes": 8000000000000,
  "free_bytes": 4000000000000,
  "parity_disks": 2,
  "data_disks": 8,
  "cache_disks": 2,
  "last_parity_check": "2025-06-15T02:00:00Z",
  "parity_errors": 0
}
```

### 4. Docker Containers

List all Docker containers:

```bash
curl -H "X-Request-ID: docker-list" \
     http://your-unraid-ip:34600/api/v1/docker/containers
```

**Response:**
```json
{
  "containers": [
    {
      "id": "abc123",
      "name": "plex",
      "image": "plexinc/pms-docker:latest",
      "status": "running",
      "state": "Up 2 hours",
      "ports": ["32400:32400/tcp"],
      "cpu_percent": 5.2,
      "memory_usage": 1073741824
    }
  ],
  "total": 15,
  "running": 12,
  "stopped": 3
}
```

### 5. Temperature Monitoring

Get system temperature readings:

```bash
curl -H "X-Request-ID: temp-check" \
     http://your-unraid-ip:34600/api/v1/system/temperature
```

**Response:**
```json
{
  "sensors": [
    {
      "name": "CPU",
      "temperature_celsius": 45.0,
      "critical_temp": 85.0,
      "status": "normal"
    },
    {
      "name": "Motherboard",
      "temperature_celsius": 38.0,
      "critical_temp": 70.0,
      "status": "normal"
    }
  ],
  "timestamp": "2025-06-19T10:30:00Z"
}
```

## Common Operations

### Start Docker Container

```bash
curl -X POST \
     -H "X-Request-ID: start-container" \
     http://your-unraid-ip:34600/api/v1/docker/containers/abc123/start
```

### Stop Docker Container

```bash
curl -X POST \
     -H "X-Request-ID: stop-container" \
     http://your-unraid-ip:34600/api/v1/docker/containers/abc123/stop
```

### Bulk Container Operations

Start multiple containers at once:

```bash
curl -X POST \
     -H "Content-Type: application/json" \
     -H "X-Request-ID: bulk-start" \
     -d '{"container_ids": ["abc123", "def456", "ghi789"]}' \
     http://your-unraid-ip:34600/api/v1/docker/containers/bulk/start
```

## Real-time Monitoring with WebSockets

### System Statistics Stream

Connect to real-time system statistics:

```javascript
const ws = new WebSocket('ws://your-unraid-ip:34600/api/v1/ws/system/stats');

ws.onmessage = function(event) {
    const stats = JSON.parse(event.data);
    console.log('CPU Usage:', stats.cpu.usage_percent + '%');
    console.log('Memory Usage:', stats.memory.usage_percent + '%');
};
```

### Docker Events Stream

Monitor Docker container events:

```javascript
const ws = new WebSocket('ws://your-unraid-ip:34600/api/v1/ws/docker/events');

ws.onmessage = function(event) {
    const event_data = JSON.parse(event.data);
    console.log('Container:', event_data.container_name);
    console.log('Action:', event_data.action);
    console.log('Status:', event_data.status);
};
```

### Storage Status Stream

Monitor storage changes:

```javascript
const ws = new WebSocket('ws://your-unraid-ip:34600/api/v1/ws/storage/status');

ws.onmessage = function(event) {
    const storage = JSON.parse(event.data);
    console.log('Array Status:', storage.array_status);
    console.log('Parity Check:', storage.parity_check_progress);
};
```

## Error Handling

### Standard Error Response

All errors follow a consistent format:

```json
{
  "error": {
    "code": "CONTAINER_NOT_FOUND",
    "message": "Container with ID 'abc123' not found",
    "details": {
      "container_id": "abc123",
      "timestamp": "2025-06-19T10:30:00Z"
    }
  },
  "request_id": "unique-identifier-123"
}
```

### Common HTTP Status Codes

- **200 OK**: Successful request
- **201 Created**: Resource created successfully
- **400 Bad Request**: Invalid request parameters
- **404 Not Found**: Resource not found
- **500 Internal Server Error**: Server error
- **503 Service Unavailable**: Service temporarily unavailable

### Error Handling Best Practices

```bash
# Check HTTP status code
response=$(curl -s -w "%{http_code}" -o response.json \
               -H "X-Request-ID: error-check" \
               http://your-unraid-ip:34600/api/v1/health)

if [ "$response" -eq 200 ]; then
    echo "Success"
    cat response.json
else
    echo "Error: HTTP $response"
    cat response.json
fi
```

## Pagination

For endpoints that return large datasets:

```bash
# Get first page (default: 20 items)
curl "http://your-unraid-ip:34600/api/v1/docker/containers?page=1&limit=20"

# Get specific page with custom limit
curl "http://your-unraid-ip:34600/api/v1/storage/disks?page=2&limit=10"
```

**Paginated Response:**
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 45,
    "total_pages": 3,
    "has_next": true,
    "has_prev": false
  }
}
```

## Integration Examples

### Home Assistant

```yaml
# configuration.yaml
sensor:
  - platform: rest
    name: "Unraid CPU Usage"
    resource: "http://192.168.1.100:34600/api/v1/system/resources"
    value_template: "{{ value_json.cpu.usage_percent }}"
    unit_of_measurement: "%"
    headers:
      X-Request-ID: "homeassistant-cpu"
```

### Prometheus

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'unraid-uma'
    static_configs:
      - targets: ['192.168.1.100:34600']
    metrics_path: '/metrics'
    scrape_interval: 30s
```

### Python Script

```python
import requests
import json

# UMA API client
class UMAClient:
    def __init__(self, base_url):
        self.base_url = base_url
        self.session = requests.Session()
        self.session.headers.update({
            'Content-Type': 'application/json',
            'Accept': 'application/vnd.uma.v1+json'
        })
    
    def get_health(self):
        response = self.session.get(
            f"{self.base_url}/health",
            headers={'X-Request-ID': 'python-health-check'}
        )
        return response.json()
    
    def get_system_stats(self):
        response = self.session.get(
            f"{self.base_url}/system/resources",
            headers={'X-Request-ID': 'python-stats'}
        )
        return response.json()

# Usage
client = UMAClient('http://192.168.1.100:34600/api/v1')
health = client.get_health()
print(f"System Status: {health['status']}")
```

## Next Steps

- **[Complete API Reference](endpoints.md)** - Detailed documentation for all endpoints
- **[WebSocket Guide](websockets.md)** - Advanced real-time monitoring setup
- **[OpenAPI Guide](openapi-guide.md)** - Using the interactive documentation
- **[Integration Examples](../deployment/integrations.md)** - Home Assistant, Prometheus, and custom integrations

## Interactive Documentation

For hands-on API exploration, visit the Swagger UI:
```
http://your-unraid-ip:34600/api/v1/docs
```

This interactive interface allows you to:
- Browse all available endpoints
- Test API calls directly from your browser
- View request/response schemas
- Generate code examples in multiple languages
