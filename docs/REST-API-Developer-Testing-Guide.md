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
  "version": "unknown",
  "uptime": 12133,
  "timestamp": "2025-06-24T07:11:21.375653991Z",
  "checks": {
    "auth": {
      "status": "pass",
      "message": "Authentication not implemented - UMA operates without authentication",
      "timestamp": "2025-06-24T07:11:21.325461789Z",
      "duration": "0ms"
    },
    "docker": {
      "status": "pass",
      "message": "Docker API healthy",
      "timestamp": "2025-06-24T07:11:21.325461144Z",
      "duration": "67.272812ms"
    },
    "storage": {
      "status": "pass",
      "message": "Storage API healthy",
      "timestamp": "2025-06-24T07:11:21.25818645Z",
      "duration": "845.041655ms"
    },
    "system": {
      "status": "pass",
      "message": "System API healthy",
      "timestamp": "2025-06-24T07:11:20.413141499Z",
      "duration": "101.237261ms"
    },
    "ups": {
      "status": "pass",
      "message": "UPS API healthy",
      "timestamp": "2025-06-24T07:11:21.325463721Z",
      "duration": "1.735µs"
    },
    "vms": {
      "status": "pass",
      "message": "Virtual Machines API healthy",
      "timestamp": "2025-06-24T07:11:21.375652351Z",
      "duration": "50.188242ms"
    }
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

# Get MCP tools count
curl -s http://192.168.20.21:34600/api/v1/mcp/tools | jq '.data.count'

# Get first 3 MCP tools
curl -s http://192.168.20.21:34600/api/v1/mcp/tools | jq '.data.tools[0:3] | .[].name'
```

**MCP Status Response:**
```json
{
  "data": {
    "active_connections": 0,
    "enabled": true,
    "max_connections": 100,
    "message": "MCP server is integrated and available on HTTP port",
    "status": "running",
    "total_tools": 51
  },
  "success": true
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

# Fan monitoring
curl -X GET http://192.168.20.21:34600/api/v1/system/fans

# Extract hostname only
curl -s http://192.168.20.21:34600/api/v1/system/info | jq '.hostname'

# Extract specific system metrics
curl -s http://192.168.20.21:34600/api/v1/system/info | jq '{hostname, cpu_cores, memory_total, uptime}'
```

**System Info Response Example:**
```json
{
  "cpu_cores": 6,
  "cpu_usage": 8.47457627118644,
  "hostname": "Cube",
  "kernel": "6.12.24-Unraid",
  "last_updated": "2025-06-24T07:14:37Z",
  "load_average": [1.29, 1.27, 1.21],
  "memory_total": 32547304,
  "uptime": 120863
}
```

**CPU Information Response Example:**
```json
{
  "architecture": "x86_64",
  "cores": 6,
  "frequency": 4099.863,
  "last_updated": "2025-06-24T10:13:09Z",
  "load1": 1.14,
  "load15": 1.13,
  "load5": 1.16,
  "model": "Intel(R) Core(TM) i7-8700K CPU @ 3.70GHz",
  "temperature": 43,
  "usage": 15.2
}
```

**Memory Usage Response Example:**
```json
{
  "available": 26975227904,
  "buffers": 9392128,
  "cached": 26542600192,
  "free": 436535296,
  "last_updated": "2025-06-24T10:13:17Z",
  "total": 33328439296,
  "usage": 19.062432943754725,
  "used": 6353211392
}
```

**System Load Response Example:**
```json
{
  "load1": 1.14,
  "load5": 1.16,
  "load15": 1.13,
  "last_updated": "2025-06-24T10:13:09Z"
}
```

**Network Statistics Response Example:**
```json
[
  {
    "interface": "eth0",
    "rx_bytes": 1234567890123,
    "tx_bytes": 987654321098,
    "rx_packets": 12345678,
    "tx_packets": 9876543,
    "rx_errors": 0,
    "tx_errors": 0,
    "last_updated": "2025-06-24T10:13:17Z"
  },
  {
    "interface": "br0",
    "rx_bytes": 2345678901234,
    "tx_bytes": 1876543210987,
    "rx_packets": 23456789,
    "tx_packets": 18765432,
    "rx_errors": 0,
    "tx_errors": 0,
    "last_updated": "2025-06-24T10:13:17Z"
  }
]
```

**Temperature Monitoring Response Example:**
```json
{
  "fans": [
    {
      "name": "nct6793 - fan1",
      "source": "nct6793",
      "speed": 812,
      "status": "normal",
      "unit": "RPM"
    },
    {
      "name": "nct6793 - fan2",
      "source": "nct6793",
      "speed": 1045,
      "status": "normal",
      "unit": "RPM"
    }
  ],
  "temperatures": [
    {
      "name": "coretemp - Package id 0",
      "source": "coretemp",
      "temperature": 43.0,
      "unit": "°C",
      "critical": 100.0,
      "max": 85.0
    },
    {
      "name": "nct6793 - SYSTIN",
      "source": "nct6793",
      "temperature": 35.0,
      "unit": "°C"
    }
  ],
  "last_updated": "2025-06-24T10:21:17Z"
}
```

**Fan Monitoring Response Example:**
```json
{
  "fans": [
    {
      "name": "nct6793 - fan1",
      "source": "nct6793",
      "speed": 812,
      "status": "normal",
      "unit": "RPM"
    },
    {
      "name": "nct6793 - fan2",
      "source": "nct6793",
      "speed": 1045,
      "status": "normal",
      "unit": "RPM"
    },
    {
      "name": "nct6793 - fan3",
      "source": "nct6793",
      "speed": 0,
      "status": "normal",
      "unit": "RPM"
    }
  ],
  "last_updated": "2025-06-24T10:21:17Z"
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

# Cache pool status
curl -X GET http://192.168.20.21:34600/api/v1/storage/cache
```

**All Storage Disks Response Example:**
```json
[
  {
    "device": "/dev/sda",
    "health": "healthy",
    "last_updated": "2025-06-24T10:10:49Z",
    "name": "sda",
    "size": 240088671846,
    "smart_data": {
      "attributes": {
        "power_cycle_count": "241",
        "power_on_hours": "18867"
      },
      "available": true,
      "status": "passed"
    },
    "status": "active",
    "temperature": 43,
    "type": "disk"
  },
  {
    "device": "/dev/sdb",
    "health": "healthy",
    "last_updated": "2025-06-24T10:10:49Z",
    "name": "sdb",
    "size": 12000138625024,
    "smart_data": {
      "attributes": {
        "power_cycle_count": "12",
        "power_on_hours": "18867"
      },
      "available": true,
      "status": "passed"
    },
    "status": "active",
    "temperature": 35,
    "type": "disk"
  }
]
```

**SMART Data Response Example:**
```json
[
  {
    "device": "/dev/sda",
    "smart_data": {
      "attributes": {
        "power_cycle_count": "241",
        "power_on_hours": "18867",
        "temperature_celsius": "43",
        "reallocated_sector_count": "0",
        "current_pending_sector": "0"
      },
      "available": true,
      "status": "passed",
      "test_result": "completed without error"
    },
    "last_updated": "2025-06-24T10:10:49Z"
  }
]
```

**Array Status Response Example:**
```json
{
  "status": "started",
  "state": "normal",
  "protection": "dual-parity",
  "sync_action": "idle",
  "last_check": "2025-06-20T08:00:00Z",
  "total_disks": 8,
  "data_disks": 6,
  "parity_disks": 2,
  "cache_disks": 1
}
```

**Cache Pool Response Example:**
```json
{
  "disks": [
    {
      "checksum_errors": "0",
      "device": "/dev/sda1",
      "name": "sda1",
      "pool": "cache",
      "read_errors": "0",
      "size": 238370684928,
      "state": "ONLINE",
      "write_errors": "0"
    }
  ],
  "pool_name": "cache",
  "status": "ONLINE",
  "total_size": 238370684928,
  "used_size": 125829120,
  "available_size": 238244855808,
  "usage_percent": 0.05,
  "filesystem": "btrfs",
  "last_updated": "2025-06-24T10:21:36Z"
}
```

### UPS Monitoring Commands

```bash
# UPS status and power information
curl -X GET http://192.168.20.21:34600/api/v1/ups/status

# UPS power consumption
curl -s http://192.168.20.21:34600/api/v1/ups/status | jq '{load, battery_charge, estimated_runtime}'

# UPS availability check
curl -s http://192.168.20.21:34600/api/v1/ups/status | jq '.available'
```

**UPS Status Response Example:**
```json
{
  "available": true,
  "battery_charge": 100,
  "detection": {
    "available": true,
    "type": 1,
    "last_check": "2025-06-24T20:21:06.892545099+10:00"
  },
  "estimated_runtime": 3600,
  "input_voltage": 120.0,
  "last_updated": "2025-06-24T10:21:36Z",
  "load": 25,
  "model": "APC Smart-UPS 1500",
  "output_voltage": 120.0,
  "power_consumption": 150.5,
  "status": "Online",
  "temperature": 25.0
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

**All Containers Response Example:**
```json
[
  {
    "created": "2025-06-23T21:59:34.178891095Z",
    "environment": [
      "UMASK=022",
      "TZ=Australia/Brisbane",
      "HOST_OS=Unraid",
      "HOST_HOSTNAME=Cube",
      "HOST_CONTAINERNAME=homeassistant",
      "TCP_PORT_8123=8123",
      "PUID=99",
      "PGID=100"
    ],
    "id": "28f97602f263ed7e15340272dcb99f9329fa432bf8d1b162f40325e428b3924a",
    "image": "homeassistant/home-assistant",
    "labels": {
      "net.unraid.docker.managed": "dockerman",
      "net.unraid.docker.webui": "http://[IP]:[PORT:8123]/"
    },
    "mounts": [
      {
        "destination": "/config",
        "source": "/mnt/user/appdata/homeassistant",
        "type": "bind"
      }
    ],
    "name": "homeassistant",
    "networks": [
      {
        "name": "bridge",
        "ip_address": "172.17.0.2"
      }
    ],
    "ports": [
      {
        "private_port": 8123,
        "public_port": 8123,
        "type": "tcp"
      }
    ],
    "state": "running",
    "status": "Up 12 hours"
  }
]
```

**Docker Images Response Example:**
```json
[
  {
    "id": "sha256:abc123def456",
    "repository": "homeassistant/home-assistant",
    "tag": "latest",
    "created": "2025-06-20T10:30:00Z",
    "size": 1234567890,
    "virtual_size": 1234567890
  },
  {
    "id": "sha256:def456ghi789",
    "repository": "linuxserver/plex",
    "tag": "latest",
    "created": "2025-06-19T15:20:00Z",
    "size": 2345678901,
    "virtual_size": 2345678901
  }
]
```

**Container Logs Response Example:**
```json
{
  "logs": [
    "2025-06-24 10:15:30 INFO (MainThread) [homeassistant.core] Starting Home Assistant",
    "2025-06-24 10:15:31 INFO (MainThread) [homeassistant.loader] Loaded integration: default_config",
    "2025-06-24 10:15:32 INFO (MainThread) [homeassistant.setup] Setting up default_config",
    "2025-06-24 10:15:33 INFO (MainThread) [homeassistant.core] Home Assistant initialized"
  ],
  "container": "homeassistant",
  "lines": 100,
  "timestamp": "2025-06-24T10:15:35Z"
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

**All VMs Response Example:**
```json
[
  {
    "autostart": false,
    "cpu_time": "5430.9s",
    "created": "2025-06-24T10:11:16Z",
    "description": "",
    "detailed_state": "running",
    "id": "1",
    "last_updated": "2025-06-24T10:11:16Z",
    "max_memory": "4194304 KiB",
    "memory": 0,
    "name": "Bastion",
    "os_type": "other",
    "resources": {
      "cpu": 2,
      "memory": 4096
    },
    "state": "running",
    "stats": {
      "memory_actual": "2097152",
      "memory_available": "4194304",
      "memory_unused": "2097152"
    },
    "uuid": "12345678-1234-1234-1234-123456789abc"
  }
]
```

**VM Statistics Response Example:**
```json
{
  "name": "Bastion",
  "state": "running",
  "cpu_time": "5430.9s",
  "memory_stats": {
    "actual": "2097152 KiB",
    "available": "4194304 KiB",
    "unused": "2097152 KiB",
    "usage_percent": 50.0
  },
  "cpu_stats": {
    "vcpus": 2,
    "cpu_time": "5430.9s",
    "usage_percent": 15.2
  },
  "disk_stats": [
    {
      "device": "vda",
      "read_bytes": 1234567890,
      "write_bytes": 987654321,
      "read_requests": 12345,
      "write_requests": 6789
    }
  ],
  "network_stats": [
    {
      "interface": "vnet0",
      "rx_bytes": 2345678901,
      "tx_bytes": 1234567890,
      "rx_packets": 23456,
      "tx_packets": 12345
    }
  ],
  "last_updated": "2025-06-24T10:11:16Z"
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
curl -s "$UMA_BASE/system/info" | jq '{hostname, uptime, load_average, cpu_cores}'

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

system_info = client.get_system_info()
print(f"Hostname: {system_info['hostname']}")
print(f"CPU Cores: {system_info['cpu_cores']}")
print(f"Memory Total: {system_info['memory_total']}")
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
| `/system/info` | GET | System information | Hostname, kernel, uptime, hardware (flat JSON) |
| `/system/cpu` | GET | CPU information | CPU details, usage, temperature |
| `/system/memory` | GET | Memory usage | RAM usage statistics |
| `/system/network` | GET | Network statistics | Interface data, traffic stats |
| `/system/temperature` | GET | Temperature monitoring | CPU, motherboard, disk temperatures |
| `/system/fans` | GET | Fan monitoring | Fan speeds and status |
| `/docker/containers` | GET | Docker containers | Container list with status |
| `/docker/containers/{name}` | GET | Specific container | Detailed container information |
| `/docker/images` | GET | Docker images | Image list with details |
| `/storage/disks` | GET | Storage disks | Disk information with SMART data |
| `/storage/array` | GET | Array status | Unraid array status |
| `/storage/cache` | GET | Cache pool status | Cache/pool disk information |
| `/storage/smart` | GET | SMART data | Disk health and SMART attributes |
| `/vms` | GET | Virtual machines | VM list with status |
| `/vms/{name}/stats` | GET | VM statistics | Detailed VM performance data |
| `/ups/status` | GET | UPS monitoring | UPS power, battery, load status |
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

### Common Data Extraction Examples

| Task | Command | Output |
|------|---------|--------|
| Get hostname | `curl -s http://192.168.20.21:34600/api/v1/system/info \| jq '.hostname'` | `"Cube"` |
| Get CPU usage | `curl -s http://192.168.20.21:34600/api/v1/system/cpu \| jq '.usage'` | `15.2` |
| Get memory usage % | `curl -s http://192.168.20.21:34600/api/v1/system/memory \| jq '.usage'` | `19.06` |
| Get CPU temperature | `curl -s http://192.168.20.21:34600/api/v1/system/temperature \| jq '.temperatures[0].temperature'` | `43.0` |
| Get fan speeds | `curl -s http://192.168.20.21:34600/api/v1/system/fans \| jq '.fans[].speed'` | `812, 1045, 0` |
| Get array status | `curl -s http://192.168.20.21:34600/api/v1/storage/array \| jq '.status'` | `"started"` |
| Get UPS battery | `curl -s http://192.168.20.21:34600/api/v1/ups/status \| jq '.battery_charge'` | `100` |
| Get running containers | `curl -s http://192.168.20.21:34600/api/v1/docker/containers \| jq 'map(select(.state=="running")) \| length'` | `3` |
| Get VM count | `curl -s http://192.168.20.21:34600/api/v1/vms \| jq 'length'` | `1` |
| Check API health | `curl -s http://192.168.20.21:34600/api/v1/health \| jq '.status'` | `"healthy"` |
| Get MCP tools count | `curl -s http://192.168.20.21:34600/api/v1/mcp/tools \| jq '.data.count'` | `51` |

### Response Status Codes

| Code | Status | Description |
|------|--------|-------------|
| `200` | OK | Request successful |
| `400` | Bad Request | Invalid request format |
| `404` | Not Found | Endpoint or resource not found |
| `405` | Method Not Allowed | HTTP method not supported |
| `500` | Internal Server Error | Server-side error |
| `503` | Service Unavailable | Service temporarily unavailable |

### WebSocket Message Formats

**Note:** The UMA API uses different response formats for different endpoints. Some endpoints return flat JSON objects (like `/system/info`), while others use wrapper objects with `data` or `success` fields. Always refer to the specific endpoint examples above for the exact response format.

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
