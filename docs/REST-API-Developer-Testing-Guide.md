# UMA REST API Developer Testing Guide

Complete curl command reference and integration guide for the UMA (Unraid Management Agent) REST API.

**Base URL:** `http://[UNRAID-SERVER-IP]:34600/api/v1/`  
**Example Server:** `http://192.168.20.21:34600/api/v1/`

---

## Table of Contents

1. [API Discovery Commands](#api-discovery-commands)
2. [Core System Information Commands](#core-system-information-commands)
3. [Container and VM Management Commands](#container-and-vm-management-commands)
4. [Real-time Monitoring Commands](#real-time-monitoring-commands)
5. [Authentication and Configuration](#authentication-and-configuration)
6. [Development Workflow Commands](#development-workflow-commands)
7. [Advanced Integration Examples](#advanced-integration-examples)
8. [Troubleshooting Commands](#troubleshooting-commands)
9. [Quick Reference](#quick-reference)

---

## API Discovery Commands

### Health Check & API Availability

```bash
# Basic health check
curl -X GET http://192.168.20.21:34600/api/v1/health

# Health check with detailed output
curl -X GET http://192.168.20.21:34600/api/v1/health | jq .

# Health check with timing
curl -w "Response Time: %{time_total}s\n" -X GET http://192.168.20.21:34600/api/v1/health
```

**Expected Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-06-24T12:00:00Z",
  "checks": {
    "auth": {"status": "healthy", "message": "Authentication disabled (internal-only API)"},
    "docker": {"status": "healthy", "message": "Docker daemon accessible"},
    "storage": {"status": "healthy", "message": "Storage systems accessible"},
    "system": {"status": "healthy", "message": "System information accessible"},
    "ups": {"status": "healthy", "message": "UPS monitoring available"},
    "vms": {"status": "healthy", "message": "VM management accessible"}
  }
}
```

### MCP Server Status & Capabilities

```bash
# MCP server status
curl -X GET http://192.168.20.21:34600/api/v1/mcp/status

# Available MCP tools (51+ tools)
curl -X GET http://192.168.20.21:34600/api/v1/mcp/tools

# MCP tools by category
curl -X GET http://192.168.20.21:34600/api/v1/mcp/tools/categories
```

**MCP Status Response:**
```json
{
  "status": "success",
  "data": {
    "enabled": true,
    "status": "running",
    "total_tools": 51,
    "active_connections": 0,
    "message": "MCP server integrated on HTTP port"
  }
}
```

---

## Core System Information Commands

### System Status & Hardware Info

```bash
# Complete system information
curl -X GET http://192.168.20.21:34600/api/v1/system/info

# System performance metrics
curl -X GET http://192.168.20.21:34600/api/v1/system/stats

# CPU information and usage
curl -X GET http://192.168.20.21:34600/api/v1/system/cpu

# Memory usage statistics
curl -X GET http://192.168.20.21:34600/api/v1/system/memory

# System load averages
curl -X GET http://192.168.20.21:34600/api/v1/system/load

# Network interface statistics
curl -X GET http://192.168.20.21:34600/api/v1/system/network

# Temperature monitoring
curl -X GET http://192.168.20.21:34600/api/v1/system/temperature
```

**System Info Response Example:**
```json
{
  "status": "success",
  "data": {
    "hostname": "Cube",
    "kernel": "6.12.24-Unraid",
    "uptime": 1234567,
    "load_average": [0.15, 0.25, 0.30],
    "cpu_cores": 6,
    "memory_total": 34359738368,
    "memory_available": 28991029248,
    "architecture": "x86_64"
  }
}
```

### Storage Array & Disk Information

```bash
# All storage disks
curl -X GET http://192.168.20.21:34600/api/v1/storage/disks

# Specific disk information
curl -X GET http://192.168.20.21:34600/api/v1/storage/disks/sda

# SMART data for all disks
curl -X GET http://192.168.20.21:34600/api/v1/storage/smart

# SMART data for specific disk
curl -X GET http://192.168.20.21:34600/api/v1/storage/smart/sda

# Array status
curl -X GET http://192.168.20.21:34600/api/v1/storage/array

# Parity status
curl -X GET http://192.168.20.21:34600/api/v1/storage/parity

# Cache status
curl -X GET http://192.168.20.21:34600/api/v1/storage/cache

# ZFS pools (if available)
curl -X GET http://192.168.20.21:34600/api/v1/storage/zfs
```

**Storage Disk Response Example:**
```json
{
  "status": "success",
  "data": [
    {
      "device": "/dev/sda",
      "name": "sda",
      "size": 240057409536,
      "model": "KINGSTON SA400S37240G",
      "serial": "50026B7682A2E8C5",
      "temperature": 42,
      "smart_status": "PASSED",
      "power_on_hours": 18859,
      "power_cycle_count": 241
    }
  ]
}
```

---

## Container and VM Management Commands

### Docker Container Management

```bash
# List all containers
curl -X GET http://192.168.20.21:34600/api/v1/docker/containers

# Running containers only
curl -X GET "http://192.168.20.21:34600/api/v1/docker/containers?status=running"

# Specific container info
curl -X GET http://192.168.20.21:34600/api/v1/docker/containers/homeassistant

# Container logs (last 100 lines)
curl -X GET "http://192.168.20.21:34600/api/v1/docker/containers/homeassistant/logs?lines=100"

# Container statistics
curl -X GET http://192.168.20.21:34600/api/v1/docker/containers/homeassistant/stats

# Docker images
curl -X GET http://192.168.20.21:34600/api/v1/docker/images

# Docker system info
curl -X GET http://192.168.20.21:34600/api/v1/docker/info
```

**Container Response Example:**
```json
{
  "status": "success",
  "data": [
    {
      "id": "a1b2c3d4e5f6",
      "name": "homeassistant",
      "image": "ghcr.io/home-assistant/home-assistant:stable",
      "status": "Up 4 hours",
      "state": "running",
      "ports": ["8123:8123/tcp"],
      "created": "2025-06-24T08:00:00Z",
      "environment": {
        "TZ": "Australia/Brisbane"
      }
    }
  ]
}
```

### Virtual Machine Management

```bash
# List all VMs
curl -X GET http://192.168.20.21:34600/api/v1/vms

# Specific VM information
curl -X GET http://192.168.20.21:34600/api/v1/vms/Bastion

# VM statistics
curl -X GET http://192.168.20.21:34600/api/v1/vms/Bastion/stats

# VM status
curl -X GET http://192.168.20.21:34600/api/v1/vms/Bastion/status
```

**VM Response Example:**
```json
{
  "status": "success",
  "data": [
    {
      "name": "Bastion",
      "state": "running",
      "id": 1,
      "uuid": "12345678-1234-1234-1234-123456789abc",
      "vcpus": 2,
      "memory": 4194304,
      "cpu_time": 4279.4,
      "memory_usage": 2147483648
    }
  ]
}
```

---

## Real-time Monitoring Commands

### WebSocket Connection Examples

#### Real-time System Monitoring WebSocket

```bash
# Test WebSocket upgrade (will hang - this is expected)
curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" \
     -H "Sec-WebSocket-Version: 13" -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
     http://192.168.20.21:34600/api/v1/ws

# Using websocat (install: cargo install websocat)
websocat ws://192.168.20.21:34600/api/v1/ws

# Subscribe to system events (send this JSON after connecting)
echo '{"type":"subscribe","channels":["system.stats","cpu.stats","memory.stats"]}' | websocat ws://192.168.20.21:34600/api/v1/ws
```

#### MCP JSON-RPC 2.0 WebSocket

```bash
# Test MCP WebSocket upgrade
curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" \
     -H "Sec-WebSocket-Version: 13" -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
     http://192.168.20.21:34600/api/v1/mcp

# MCP initialization (send after WebSocket connection)
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}' | websocat ws://192.168.20.21:34600/api/v1/mcp

# List available MCP tools
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | websocat ws://192.168.20.21:34600/api/v1/mcp
```

**Expected WebSocket Upgrade Response:**
```
HTTP/1.1 101 Switching Protocols
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=
```

---

## Authentication and Configuration

### Headers and Configuration

```bash
# Standard headers for JSON requests
curl -X GET -H "Accept: application/json" -H "Content-Type: application/json" \
     http://192.168.20.21:34600/api/v1/health

# CORS preflight request
curl -X OPTIONS -H "Origin: http://localhost:3000" \
     -H "Access-Control-Request-Method: GET" \
     -H "Access-Control-Request-Headers: Content-Type" \
     http://192.168.20.21:34600/api/v1/health

# Request with custom User-Agent
curl -X GET -H "User-Agent: MyApp/1.0" \
     http://192.168.20.21:34600/api/v1/system/info
```

### Error Handling Examples

```bash
# Test invalid endpoint (404)
curl -X GET http://192.168.20.21:34600/api/v1/invalid-endpoint

# Test invalid method (405)
curl -X DELETE http://192.168.20.21:34600/api/v1/health

# Test malformed request
curl -X POST -H "Content-Type: application/json" \
     -d '{"invalid": json}' \
     http://192.168.20.21:34600/api/v1/mcp/config
```

**Error Response Format:**
```json
{
  "status": "error",
  "error": {
    "code": 404,
    "message": "Endpoint not found",
    "details": "The requested endpoint does not exist"
  },
  "timestamp": "2025-06-24T12:00:00Z"
}
```

---

## Development Workflow Commands

### HTTP Method Testing

```bash
# GET requests (read operations)
curl -X GET http://192.168.20.21:34600/api/v1/system/info

# POST requests (create/action operations)
curl -X POST -H "Content-Type: application/json" \
     -d '{"action":"refresh"}' \
     http://192.168.20.21:34600/api/v1/mcp/tools/refresh

# PUT requests (update operations)
curl -X PUT -H "Content-Type: application/json" \
     -d '{"enabled":true,"max_connections":100}' \
     http://192.168.20.21:34600/api/v1/mcp/config

# HEAD requests (check existence)
curl -I http://192.168.20.21:34600/api/v1/health
```

### Rate Limiting and Connection Testing

```bash
# Test concurrent requests
for i in {1..5}; do
  curl -s http://192.168.20.21:34600/api/v1/health &
done
wait

# Test with timeout
curl --max-time 10 http://192.168.20.21:34600/api/v1/system/info

# Test connection persistence
curl --keepalive-time 60 http://192.168.20.21:34600/api/v1/health

# Verbose output for debugging
curl -v http://192.168.20.21:34600/api/v1/health
```

### Performance Testing

```bash
# Response time measurement
curl -w "Total: %{time_total}s, Connect: %{time_connect}s, Transfer: %{time_starttransfer}s\n" \
     -o /dev/null -s http://192.168.20.21:34600/api/v1/system/info

# Load testing with Apache Bench
ab -n 100 -c 10 http://192.168.20.21:34600/api/v1/health

# Simple load test with curl
for i in {1..10}; do
  time curl -s http://192.168.20.21:34600/api/v1/health > /dev/null
done
```

---

## Advanced Integration Examples

### Bash Script for System Monitoring

```bash
#!/bin/bash
# uma-monitor.sh - Simple UMA monitoring script

UMA_BASE="http://192.168.20.21:34600/api/v1"

# Check API health
echo "=== UMA Health Check ==="
curl -s "$UMA_BASE/health" | jq '.status'

# Get system stats
echo -e "\n=== System Information ==="
curl -s "$UMA_BASE/system/info" | jq '.data | {hostname, uptime, load_average, cpu_cores}'

# Get container status
echo -e "\n=== Running Containers ==="
curl -s "$UMA_BASE/docker/containers?status=running" | jq '.data[] | {name, status, image}'

# Get disk temperatures
echo -e "\n=== Disk Temperatures ==="
curl -s "$UMA_BASE/storage/disks" | jq '.data[] | {device, temperature, smart_status}'
```

### Python Integration Example

```python
import requests
import json

class UMAClient:
    def __init__(self, base_url="http://192.168.20.21:34600/api/v1"):
        self.base_url = base_url
        self.session = requests.Session()

    def health_check(self):
        response = self.session.get(f"{self.base_url}/health")
        return response.json()

    def get_system_info(self):
        response = self.session.get(f"{self.base_url}/system/info")
        return response.json()

    def get_containers(self, status=None):
        url = f"{self.base_url}/docker/containers"
        if status:
            url += f"?status={status}"
        response = self.session.get(url)
        return response.json()

# Usage
client = UMAClient()
health = client.health_check()
print(f"API Status: {health['status']}")
```

### JavaScript/Node.js Integration Example

```javascript
const axios = require('axios');

class UMAClient {
    constructor(baseUrl = 'http://192.168.20.21:34600/api/v1') {
        this.baseUrl = baseUrl;
        this.client = axios.create({
            baseURL: baseUrl,
            timeout: 10000,
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            }
        });
    }

    async healthCheck() {
        const response = await this.client.get('/health');
        return response.data;
    }

    async getSystemInfo() {
        const response = await this.client.get('/system/info');
        return response.data;
    }

    async getContainers(status = null) {
        const params = status ? { status } : {};
        const response = await this.client.get('/docker/containers', { params });
        return response.data;
    }
}

// Usage
const client = new UMAClient();
client.healthCheck().then(health => {
    console.log(`API Status: ${health.status}`);
});
```

---

## Troubleshooting Commands

### Connection Issues

```bash
# Test basic connectivity
ping 192.168.20.21

# Test port accessibility
telnet 192.168.20.21 34600

# Check if service is running
curl -I http://192.168.20.21:34600/api/v1/health

# Test with different IP (if using hostname)
curl http://localhost:34600/api/v1/health  # if running locally
```

### Debug Information

```bash
# Get detailed curl output
curl -v -X GET http://192.168.20.21:34600/api/v1/health

# Test CORS headers
curl -H "Origin: http://localhost:3000" -v http://192.168.20.21:34600/api/v1/health

# Check response headers
curl -I http://192.168.20.21:34600/api/v1/health
```

### Common Issues and Solutions

| Issue | Symptoms | Solution |
|-------|----------|----------|
| Connection Refused | `curl: (7) Failed to connect` | Check if UMA service is running |
| Timeout | `curl: (28) Operation timed out` | Verify network connectivity and firewall |
| 404 Not Found | `{"status":"error","error":{"code":404}}` | Check endpoint URL spelling |
| 405 Method Not Allowed | `{"status":"error","error":{"code":405}}` | Verify HTTP method is supported |
| CORS Error | Browser console CORS error | Add proper Origin header |

---

## Quick Reference

### Essential Endpoints

| Endpoint | Method | Description | Response |
|----------|--------|-------------|----------|
| `/health` | GET | API health check | Health status with subsystem checks |
| `/system/info` | GET | System information | Hostname, kernel, uptime, hardware |
| `/system/stats` | GET | Performance metrics | CPU, memory, load averages |
| `/docker/containers` | GET | Docker containers | Container list with status |
| `/docker/containers/{name}` | GET | Specific container | Detailed container information |
| `/storage/disks` | GET | Storage disks | Disk information with SMART data |
| `/storage/array` | GET | Array status | Unraid array status |
| `/vms` | GET | Virtual machines | VM list with status |
| `/mcp/status` | GET | MCP server status | MCP integration status |
| `/mcp/tools` | GET | Available MCP tools | List of 51+ MCP tools |

### WebSocket Endpoints

| Endpoint | Protocol | Description | Usage |
|----------|----------|-------------|-------|
| `/ws` | WebSocket | Real-time monitoring | System metrics streaming |
| `/mcp` | WebSocket | MCP JSON-RPC 2.0 | AI agent integration |

### Common Query Parameters

| Parameter | Endpoints | Description | Example |
|-----------|-----------|-------------|---------|
| `status` | `/docker/containers` | Filter by status | `?status=running` |
| `lines` | `/docker/containers/{name}/logs` | Log line limit | `?lines=100` |

### Response Status Codes

| Code | Status | Description |
|------|--------|-------------|
| `200` | OK | Request successful |
| `400` | Bad Request | Invalid request format |
| `404` | Not Found | Endpoint or resource not found |
| `405` | Method Not Allowed | HTTP method not supported |
| `500` | Internal Server Error | Server-side error |
| `503` | Service Unavailable | Service temporarily unavailable |

### Standard Response Format

All API responses follow this structure:

```json
{
  "status": "success|error",
  "data": { /* response data */ },
  "error": { /* error details if status is error */ },
  "timestamp": "2025-06-24T12:00:00Z"
}
```

### WebSocket Message Formats

#### Real-time Monitoring WebSocket

```json
// Subscribe to events
{
  "type": "subscribe",
  "channels": ["system.stats", "cpu.stats", "memory.stats"]
}

// Event message
{
  "type": "event",
  "channel": "system.stats",
  "data": { /* event data */ },
  "timestamp": "2025-06-24T12:00:00Z"
}
```

#### MCP JSON-RPC 2.0 WebSocket Messages

```json
// Initialize connection
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "capabilities": {},
    "clientInfo": {"name": "test-client", "version": "1.0.0"}
  }
}

// List tools
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/list",
  "params": {}
}
```

---

## Additional Resources

- **UMA GitHub Repository:** [https://github.com/domalab/uma](https://github.com/domalab/uma)
- **MCP Protocol Specification:** [https://modelcontextprotocol.io/](https://modelcontextprotocol.io/)
- **WebSocket Testing Tool:** [websocat](https://github.com/vi/websocat) - `cargo install websocat`
- **JSON Processing:** [jq](https://stedolan.github.io/jq/) - Command-line JSON processor

---

*This guide covers the UMA REST API as of version 2025.06.24. For the latest updates and changes, refer to the project repository.*
