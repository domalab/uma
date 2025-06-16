# Complete API Endpoints Reference

This document provides a comprehensive reference for all UMA API endpoints with examples and response formats.

## Base URL
```
http://your-unraid-ip:34600/api/v1
```

## Authentication
Currently, no authentication is required. UMA is designed for trusted network environments.

## Common Headers
```
X-Request-ID: unique-request-identifier
Accept: application/vnd.uma.v1+json
Accept-Encoding: gzip
Content-Type: application/json
```

## Health & System Endpoints

### GET /health
Get comprehensive system health status.

**Example Request:**
```bash
curl -H "X-Request-ID: health-check-123" \
     http://your-unraid-ip:34600/api/v1/health
```

**Response:**
```json
{
  "status": "healthy",
  "service": "uma",
  "dependencies": {
    "docker": "healthy",
    "libvirt": "healthy",
    "storage": "healthy",
    "notifications": "healthy"
  }
}
```

### GET /system/stats
Get real-time system statistics.

**Example Request:**
```bash
curl http://your-unraid-ip:34600/api/v1/system/stats
```

**Response:**
```json
{
  "data": {
    "cpu_percent": 15.2,
    "memory_percent": 45.8,
    "memory_used": 8589934592,
    "memory_total": 17179869184,
    "uptime": 86400,
    "load_average": [0.5, 0.7, 0.9]
  },
  "meta": {
    "request_id": "stats-request-456",
    "version": "v1",
    "timestamp": "2025-06-15T23:00:00Z"
  }
}
```

### GET /system/temperature
Get system temperature sensors.

**Example Request:**
```bash
curl http://your-unraid-ip:34600/api/v1/system/temperature
```

### GET /system/network
Get network interface information.

**Example Request:**
```bash
curl http://your-unraid-ip:34600/api/v1/system/network
```

### GET /system/ups
Get UPS status and metrics.

**Example Request:**
```bash
curl http://your-unraid-ip:34600/api/v1/system/ups
```

### GET /system/gpu
Get GPU usage and temperature.

**Example Request:**
```bash
curl http://your-unraid-ip:34600/api/v1/system/gpu
```

## Storage Endpoints

### GET /storage/disks
Get all disk information with pagination support.

**Query Parameters:**
- `page` (integer, default: 1) - Page number
- `limit` (integer, default: 10, max: 1000) - Items per page

**Example Request:**
```bash
curl "http://your-unraid-ip:34600/api/v1/storage/disks?page=1&limit=5"
```

**Response:**
```json
{
  "data": [
    {
      "name": "disk1",
      "device": "/dev/sdb1",
      "size": "8TB",
      "used": "4.2TB",
      "free": "3.8TB",
      "temperature": 35,
      "status": "active"
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 5,
    "total": 12,
    "has_more": true,
    "total_pages": 3
  },
  "meta": {
    "request_id": "storage-query-789",
    "version": "v1",
    "timestamp": "2025-06-15T23:00:00Z"
  }
}
```

### GET /storage/array
Get array disk information.

**Example Request:**
```bash
curl http://your-unraid-ip:34600/api/v1/storage/array
```

### GET /storage/cache
Get cache disk status.

**Example Request:**
```bash
curl http://your-unraid-ip:34600/api/v1/storage/cache
```

### GET /storage/boot
Get boot disk information.

**Example Request:**
```bash
curl http://your-unraid-ip:34600/api/v1/storage/boot
```

## Docker Management Endpoints

### GET /docker/containers
List all Docker containers with pagination.

**Query Parameters:**
- `page` (integer, default: 1) - Page number
- `limit` (integer, default: 10, max: 1000) - Items per page

**Example Request:**
```bash
curl -H "Accept: application/vnd.uma.v1+json" \
     "http://your-unraid-ip:34600/api/v1/docker/containers?page=1&limit=10"
```

**Response:**
```json
{
  "data": [
    {
      "id": "abc123def456",
      "name": "plex",
      "image": "plexinc/pms-docker:latest",
      "status": "running",
      "state": "Up 2 hours",
      "ports": ["32400:32400/tcp"]
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 10,
    "total": 25,
    "has_more": true,
    "total_pages": 3
  },
  "meta": {
    "request_id": "containers-list-101",
    "version": "v1",
    "timestamp": "2025-06-15T23:00:00Z"
  }
}
```

### POST /docker/containers/bulk/start
Start multiple Docker containers.

**Request Body:**
```json
{
  "container_ids": ["plex", "nginx", "sonarr"]
}
```

**Example Request:**
```bash
curl -X POST \
     -H "Content-Type: application/json" \
     -H "X-Request-ID: bulk-start-202" \
     -d '{"container_ids": ["plex", "nginx"]}' \
     http://your-unraid-ip:34600/api/v1/docker/containers/bulk/start
```

**Response:**
```json
{
  "data": {
    "results": [
      {
        "container_id": "plex",
        "success": true,
        "message": "Container started successfully"
      },
      {
        "container_id": "nginx",
        "success": true,
        "message": "Container started successfully"
      }
    ],
    "summary": {
      "total": 2,
      "succeeded": 2,
      "failed": 0
    }
  },
  "meta": {
    "request_id": "bulk-start-202",
    "version": "v1",
    "timestamp": "2025-06-15T23:00:00Z"
  }
}
```

### POST /docker/containers/bulk/stop
Stop multiple Docker containers.

**Request Body:**
```json
{
  "container_ids": ["plex", "nginx", "sonarr"]
}
```

**Example Request:**
```bash
curl -X POST \
     -H "Content-Type: application/json" \
     -H "X-Request-ID: bulk-stop-303" \
     -d '{"container_ids": ["plex", "nginx"]}' \
     http://your-unraid-ip:34600/api/v1/docker/containers/bulk/stop
```

### POST /docker/containers/bulk/restart
Restart multiple Docker containers.

**Request Body:**
```json
{
  "container_ids": ["plex", "nginx", "sonarr"]
}
```

**Example Request:**
```bash
curl -X POST \
     -H "Content-Type: application/json" \
     -H "X-Request-ID: bulk-restart-404" \
     -d '{"container_ids": ["plex", "nginx"]}' \
     http://your-unraid-ip:34600/api/v1/docker/containers/bulk/restart
```

## Documentation Endpoints

### GET /docs
Interactive Swagger UI documentation.

**Example Request:**
```bash
curl http://your-unraid-ip:34600/api/v1/docs
```

**Response:** HTML page with Swagger UI interface

### GET /openapi.json
OpenAPI 3.1.1 specification in JSON format.

**Example Request:**
```bash
curl http://your-unraid-ip:34600/api/v1/openapi.json
```

**Response:**
```json
{
  "openapi": "3.1.1",
  "info": {
    "title": "UMA API",
    "version": "1.0.0",
    "description": "Unraid Management Agent API"
  },
  "paths": {
    "/api/v1/health": {
      "get": {
        "summary": "Get system health status",
        "responses": {
          "200": {
            "description": "Health status"
          }
        }
      }
    }
  }
}
```

## Metrics Endpoint

### GET /metrics
Prometheus metrics endpoint.

**Example Request:**
```bash
curl http://your-unraid-ip:34600/metrics
```

**Response:**
```prometheus
# HELP uma_api_requests_total Total number of API requests
# TYPE uma_api_requests_total counter
uma_api_requests_total{endpoint="/api/v1/health",method="GET",status_code="200"} 42

# HELP uma_api_request_duration_seconds Duration of API requests in seconds
# TYPE uma_api_request_duration_seconds histogram
uma_api_request_duration_seconds_bucket{endpoint="/api/v1/health",method="GET",le="0.005"} 0
uma_api_request_duration_seconds_bucket{endpoint="/api/v1/health",method="GET",le="0.01"} 0
uma_api_request_duration_seconds_bucket{endpoint="/api/v1/health",method="GET",le="+Inf"} 42
uma_api_request_duration_seconds_sum{endpoint="/api/v1/health",method="GET"} 84.5
uma_api_request_duration_seconds_count{endpoint="/api/v1/health",method="GET"} 42
```

## Error Responses

All endpoints may return error responses in the following format:

### 400 Bad Request
```json
{
  "error": "validation failed: at least 1 container ID is required",
  "meta": {
    "request_id": "bulk-start-505",
    "version": "v1",
    "timestamp": "2025-06-15T23:00:00Z"
  }
}
```

### 404 Not Found
```json
{
  "error": "endpoint not found",
  "meta": {
    "request_id": "invalid-endpoint-606",
    "version": "v1",
    "timestamp": "2025-06-15T23:00:00Z"
  }
}
```

### 405 Method Not Allowed
```json
{
  "error": "method not allowed",
  "meta": {
    "request_id": "wrong-method-707",
    "version": "v1",
    "timestamp": "2025-06-15T23:00:00Z"
  }
}
```

### 500 Internal Server Error
```json
{
  "error": "internal server error",
  "meta": {
    "request_id": "server-error-808",
    "version": "v1",
    "timestamp": "2025-06-15T23:00:00Z"
  }
}
```

## Rate Limiting

Currently, no rate limiting is implemented. The API is designed for trusted network environments.

## Next Steps

- **[WebSocket Guide](websockets.md)** - Real-time monitoring endpoints
- **[Bulk Operations](bulk-operations.md)** - Detailed bulk operation examples
- **[OpenAPI Guide](openapi-guide.md)** - Using the interactive documentation
