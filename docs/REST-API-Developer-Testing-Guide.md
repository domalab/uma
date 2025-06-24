# UMA REST API Developer Testing Guide

Complete curl command reference and integration guide for the UMA (Unraid Management Agent) REST API.

**Base URL:** `http://[UNRAID-SERVER-IP]:34600/api/v1/`  
**Example Server:** `http://192.168.20.21:34600/api/v1/`

---

## Table of Contents

1. [API Discovery Commands](#api-discovery-commands)
2. [Core System Information Commands](#core-system-information-commands)
3. [Container and VM Management Commands](#container-and-vm-management-commands)
4. [Real-time Monitoring and WebSocket Integration](#real-time-monitoring-and-websocket-integration)
5. [Authentication and Configuration](#authentication-and-configuration)
6. [Development Workflow Commands](#development-workflow-commands)
7. [Advanced Integration Examples](#advanced-integration-examples)
8. [WebSocket Testing and Debugging](#websocket-testing-and-debugging)
9. [Troubleshooting Commands](#troubleshooting-commands)
10. [Quick Reference](#quick-reference)

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

## Real-time Monitoring and WebSocket Integration

UMA provides comprehensive WebSocket support for real-time monitoring and MCP (Model Context Protocol) integration. This section covers both endpoints with complete examples and real data from the "Cube" Unraid server.

### WebSocket Endpoints Overview

| Endpoint | Protocol | Purpose | Max Connections |
|----------|----------|---------|-----------------|
| `/api/v1/ws` | WebSocket | Real-time monitoring with subscription management | 50 |
| `/api/v1/mcp` | WebSocket | MCP JSON-RPC 2.0 for AI agent integration | 100 |

---

## Real-time Monitoring WebSocket (`/api/v1/ws`)

### Connection and Basic Testing

```bash
# Test WebSocket upgrade (will hang - this is expected)
curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" \
     -H "Sec-WebSocket-Version: 13" -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
     http://192.168.20.21:34600/api/v1/ws

# Using websocat (install: cargo install websocat)
websocat ws://192.168.20.21:34600/api/v1/ws

# Interactive WebSocket testing
websocat -t ws://192.168.20.21:34600/api/v1/ws
```

**Expected WebSocket Upgrade Response:**
```http
HTTP/1.1 101 Switching Protocols
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=
```

### Available Subscription Channels

#### System Monitoring Channels
- `system.stats` - Real-time system performance metrics
- `system.health` - System health status updates
- `system.load` - System load averages
- `cpu.stats` - CPU usage and temperature data
- `memory.stats` - Memory usage statistics
- `network.stats` - Network interface statistics

#### Storage Monitoring Channels
- `storage.status` - Storage array status updates
- `disk.stats` - Individual disk statistics
- `array.status` - Array start/stop status changes
- `parity.status` - Parity check progress and status
- `disk.smart.warning` - SMART warning alerts
- `cache.status` - Cache pool status updates

#### Container and VM Channels
- `docker.events` - Docker container lifecycle events
- `container.stats` - Container resource usage
- `container.health` - Container health checks
- `image.events` - Docker image events
- `vm.events` - Virtual machine lifecycle events
- `vm.stats` - VM resource usage statistics
- `vm.health` - VM health monitoring

#### Alert and Infrastructure Channels
- `temperature.alert` - Temperature threshold alerts
- `resource.alert` - CPU/memory/disk usage alerts
- `security.alert` - Security-related alerts
- `system.alert` - General system alerts
- `ups.status` - UPS status and power events
- `fan.status` - Fan speed and status updates
- `power.status` - Power management events

#### Operational Channels
- `task.progress` - Long-running task progress
- `backup.status` - Backup operation status
- `update.status` - System update progress

### WebSocket Subscription Examples

#### Subscribe to Multiple Channels

```bash
# Subscribe to system monitoring channels
echo '{"type":"subscribe","channels":["system.stats","cpu.stats","memory.stats"]}' | websocat ws://192.168.20.21:34600/api/v1/ws

# Subscribe to storage monitoring
echo '{"type":"subscribe","channels":["storage.status","disk.stats","array.status"]}' | websocat ws://192.168.20.21:34600/api/v1/ws

# Subscribe to Docker events
echo '{"type":"subscribe","channels":["docker.events","container.stats"]}' | websocat ws://192.168.20.21:34600/api/v1/ws

# Subscribe to alerts and infrastructure
echo '{"type":"subscribe","channels":["temperature.alert","ups.status","fan.status"]}' | websocat ws://192.168.20.21:34600/api/v1/ws
```

#### Subscription Management

```bash
# List active subscriptions
echo '{"type":"list_subscriptions"}' | websocat ws://192.168.20.21:34600/api/v1/ws

# Unsubscribe from specific channels
echo '{"type":"unsubscribe","channels":["system.stats"]}' | websocat ws://192.168.20.21:34600/api/v1/ws

# Ping to keep connection alive
echo '{"type":"ping"}' | websocat ws://192.168.20.21:34600/api/v1/ws
```

### Real-time Message Formats

#### System Statistics Event

```json
{
  "event_type": "system.stats",
  "data": {
    "cpu_percent": 15.2,
    "memory_percent": 19.06,
    "memory_used": 6353211392,
    "memory_total": 33328439296,
    "uptime": 120863,
    "load_average": [1.29, 1.27, 1.21],
    "hostname": "Cube",
    "last_updated": "2025-06-24T10:21:36Z"
  },
  "timestamp": "2025-06-24T10:21:36Z"
}
```

#### CPU Statistics Event

```json
{
  "event_type": "cpu.stats",
  "data": {
    "usage": 15.2,
    "cores": 6,
    "frequency": 4099.863,
    "temperature": 43.0,
    "architecture": "x86_64",
    "model": "Intel(R) Core(TM) i7-8700K CPU @ 3.70GHz",
    "load1": 1.14,
    "load5": 1.16,
    "load15": 1.13
  },
  "timestamp": "2025-06-24T10:21:36Z"
}
```

#### Memory Statistics Event

```json
{
  "event_type": "memory.stats",
  "data": {
    "total": 33328439296,
    "available": 26975227904,
    "used": 6353211392,
    "free": 436535296,
    "buffers": 9392128,
    "cached": 26542600192,
    "usage": 19.062432943754725
  },
  "timestamp": "2025-06-24T10:21:36Z"
}
```

#### Docker Events

```json
{
  "event_type": "docker.events",
  "data": {
    "action": "start",
    "container_id": "28f97602f263ed7e15340272dcb99f9329fa432bf8d1b162f40325e428b3924a",
    "container_name": "homeassistant",
    "image": "homeassistant/home-assistant",
    "status": "running",
    "timestamp": "2025-06-24T10:21:36Z"
  },
  "timestamp": "2025-06-24T10:21:36Z"
}
```

#### Temperature Alert

```json
{
  "event_type": "temperature.alert",
  "data": {
    "sensor_name": "coretemp - Package id 0",
    "sensor_type": "cpu",
    "temperature": 75.2,
    "threshold": 70.0,
    "level": "warning",
    "message": "CPU Package temperature warning: 75.2°C (threshold: 70.0°C)",
    "source": "coretemp"
  },
  "timestamp": "2025-06-24T10:21:36Z"
}
```

#### UPS Status Event

```json
{
  "event_type": "ups.status",
  "data": {
    "available": true,
    "battery_charge": 100,
    "estimated_runtime": 3600,
    "input_voltage": 120.0,
    "load": 25,
    "model": "APC Smart-UPS 1500",
    "output_voltage": 120.0,
    "power_consumption": 150.5,
    "status": "Online",
    "temperature": 25.0
  },
  "timestamp": "2025-06-24T10:21:36Z"
}
```

#### Storage Status Event

```json
{
  "event_type": "storage.status",
  "data": {
    "array_status": "started",
    "array_state": "normal",
    "protection": "dual-parity",
    "sync_action": "idle",
    "total_disks": 8,
    "data_disks": 6,
    "parity_disks": 2,
    "cache_disks": 1,
    "last_check": "2025-06-20T02:00:00Z"
  },
  "timestamp": "2025-06-24T10:21:36Z"
}
```

## MCP JSON-RPC 2.0 WebSocket (`/api/v1/mcp`)

### Connection and Initialization

```bash
# Test MCP WebSocket upgrade
curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" \
     -H "Sec-WebSocket-Version: 13" -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
     http://192.168.20.21:34600/api/v1/mcp

# Connect with websocat
websocat ws://192.168.20.21:34600/api/v1/mcp

# Interactive MCP testing
websocat -t ws://192.168.20.21:34600/api/v1/mcp
```

### MCP Protocol Initialization

#### Initialize MCP Connection

```bash
# Send initialization message (required first step)
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}' | websocat ws://192.168.20.21:34600/api/v1/mcp
```

**Expected Initialization Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "tools": {},
      "resources": {},
      "prompts": {}
    },
    "serverInfo": {
      "name": "uma-mcp-server",
      "version": "1.0.0"
    }
  }
}
```

### Available MCP Methods

#### Core MCP Methods
- `initialize` - Initialize MCP connection and capabilities
- `tools/list` - List all available tools (51 tools)
- `tools/call` - Execute a specific tool
- `resources/list` - List available resources
- `prompts/list` - List available prompts

### MCP Tools Documentation (51 Tools Available)

#### System Information Tools

```bash
# List all available tools
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | websocat ws://192.168.20.21:34600/api/v1/mcp

# Get system information
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_system_info","arguments":{}}}' | websocat ws://192.168.20.21:34600/api/v1/mcp

# Get disk usage
echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_disk_usage","arguments":{}}}' | websocat ws://192.168.20.21:34600/api/v1/mcp
```

#### Key MCP Tools (First 10 of 51)

| Tool Name | Description | Parameters |
|-----------|-------------|------------|
| `get_system_info` | Get comprehensive system information | None |
| `get_disk_usage` | Get disk space usage for all filesystems | None |
| `get_vm_details` | Get detailed VM information | `vm_name` (string) |
| `get_ups_status` | Get UPS status and battery information | None |
| `get_security_status` | Get security status and recommendations | None |
| `get_firewall_status` | Get firewall configuration and rules | None |
| `get_share_config` | Get user share configuration | None |
| `list_plugins` | List all installed Unraid plugins | None |
| `get_filesystem_info` | Get filesystem information and mount points | None |
| `get_cache_status` | Get cache drive status and usage | None |

### MCP Tool Execution Examples

#### Execute System Information Tool

```bash
echo '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"get_system_info","arguments":{}}}' | websocat ws://192.168.20.21:34600/api/v1/mcp
```

**Expected Tool Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "System Information for Cube:\n- CPU: Intel(R) Core(TM) i7-8700K CPU @ 3.70GHz (6 cores)\n- Memory: 31.0 GB total, 25.1 GB available\n- Uptime: 1 day, 9 hours, 34 minutes\n- Kernel: 6.12.24-Unraid\n- Load Average: 1.29, 1.27, 1.21"
      }
    ],
    "isError": false
  }
}
```

#### Execute UPS Status Tool

```bash
echo '{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"get_ups_status","arguments":{}}}' | websocat ws://192.168.20.21:34600/api/v1/mcp
```

**Expected UPS Tool Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 6,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "UPS Status:\n- Model: APC Smart-UPS 1500\n- Status: Online\n- Battery Charge: 100%\n- Load: 25%\n- Estimated Runtime: 60 minutes\n- Input Voltage: 120.0V\n- Output Voltage: 120.0V"
      }
    ],
    "isError": false
  }
}
```

#### Execute VM Details Tool

```bash
echo '{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"get_vm_details","arguments":{"vm_name":"Bastion"}}}' | websocat ws://192.168.20.21:34600/api/v1/mcp
```

**Expected VM Tool Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 7,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "VM Details for Bastion:\n- State: running\n- CPU: 2 vCPUs\n- Memory: 4096 MB allocated\n- CPU Time: 5430.9s\n- Memory Usage: 50%\n- Autostart: disabled\n- OS Type: other"
      }
    ],
    "isError": false
  }
}
```

### MCP Error Handling

#### Tool Not Found Error

```json
{
  "jsonrpc": "2.0",
  "id": 8,
  "error": {
    "code": -32601,
    "message": "Method not found",
    "data": {
      "method": "invalid_tool"
    }
  }
}
```

#### Invalid Parameters Error

```json
{
  "jsonrpc": "2.0",
  "id": 9,
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": {
      "tool": "get_vm_details",
      "missing": ["vm_name"]
    }
  }
}
```

## WebSocket Connection Management

### Connection Lifecycle

#### 1. Connection Establishment
```bash
# Real-time monitoring WebSocket
websocat ws://192.168.20.21:34600/api/v1/ws

# MCP WebSocket (requires initialization)
websocat ws://192.168.20.21:34600/api/v1/mcp
```

#### 2. Authentication
- **No authentication required** - UMA operates as internal-only API
- Connections are limited by IP origin validation
- Rate limiting: 100 messages per minute per connection

#### 3. Subscription/Initialization
```bash
# Real-time: Subscribe to channels
{"type":"subscribe","channels":["system.stats","docker.events"]}

# MCP: Initialize protocol
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"client","version":"1.0.0"}}}
```

#### 4. Data Exchange
- Real-time: Receive subscribed events automatically
- MCP: Send JSON-RPC 2.0 method calls and receive responses

#### 5. Connection Cleanup
```bash
# Graceful disconnect (websocat: Ctrl+C)
# Automatic cleanup after 5 minutes of inactivity
```

### Connection Limits and Rate Limiting

| Parameter | Real-time WebSocket | MCP WebSocket |
|-----------|-------------------|---------------|
| **Max Connections** | 50 | 100 |
| **Message Size Limit** | 1MB | 1MB |
| **Rate Limit** | 100 msg/min | 100 msg/min |
| **Idle Timeout** | 5 minutes | 5 minutes |
| **Reconnect Strategy** | Exponential backoff | Exponential backoff |

### Error Handling and Troubleshooting

#### Common WebSocket Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `Connection refused` | UMA service not running | Check service status |
| `Rate limit exceeded` | Too many messages | Reduce message frequency |
| `Invalid message format` | Malformed JSON | Validate JSON syntax |
| `Unknown channel` | Invalid subscription | Use valid channel names |
| `Connection timeout` | Network issues | Check network connectivity |

#### WebSocket Error Response Format

```json
{
  "type": "error",
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded: 100 messages per minute",
    "timestamp": "2025-06-24T10:21:36Z"
  }
}
```

### Reconnection Strategies

#### Exponential Backoff Pattern
```javascript
let reconnectDelay = 1000; // Start with 1 second
const maxDelay = 30000;    // Max 30 seconds

function reconnect() {
    setTimeout(() => {
        connect();
        reconnectDelay = Math.min(reconnectDelay * 2, maxDelay);
    }, reconnectDelay);
}
```

---

## Developer Integration Examples

### JavaScript/Node.js WebSocket Client

#### Real-time Monitoring Client

```javascript
const WebSocket = require('ws');

class UMAMonitor {
    constructor(baseUrl = 'ws://192.168.20.21:34600') {
        this.baseUrl = baseUrl;
        this.ws = null;
        this.reconnectDelay = 1000;
        this.maxReconnectDelay = 30000;
    }

    connect() {
        this.ws = new WebSocket(`${this.baseUrl}/api/v1/ws`);

        this.ws.on('open', () => {
            console.log('Connected to UMA WebSocket');
            this.reconnectDelay = 1000; // Reset delay on successful connection

            // Subscribe to system monitoring
            this.subscribe(['system.stats', 'cpu.stats', 'memory.stats', 'docker.events']);
        });

        this.ws.on('message', (data) => {
            const event = JSON.parse(data);
            this.handleEvent(event);
        });

        this.ws.on('close', () => {
            console.log('WebSocket connection closed, reconnecting...');
            this.reconnect();
        });

        this.ws.on('error', (error) => {
            console.error('WebSocket error:', error);
        });
    }

    subscribe(channels) {
        const message = {
            type: 'subscribe',
            channels: channels
        };
        this.ws.send(JSON.stringify(message));
    }

    handleEvent(event) {
        switch(event.event_type) {
            case 'system.stats':
                console.log(`CPU: ${event.data.cpu_percent}%, Memory: ${event.data.memory_percent}%`);
                break;
            case 'docker.events':
                console.log(`Docker: ${event.data.action} - ${event.data.container_name}`);
                break;
            case 'temperature.alert':
                console.log(`Temperature Alert: ${event.data.message}`);
                break;
        }
    }

    reconnect() {
        setTimeout(() => {
            this.connect();
            this.reconnectDelay = Math.min(this.reconnectDelay * 2, this.maxReconnectDelay);
        }, this.reconnectDelay);
    }
}

// Usage
const monitor = new UMAMonitor();
monitor.connect();
```

#### MCP Client Example

```javascript
const WebSocket = require('ws');

class UMAMCPClient {
    constructor(baseUrl = 'ws://192.168.20.21:34600') {
        this.baseUrl = baseUrl;
        this.ws = null;
        this.requestId = 1;
        this.pendingRequests = new Map();
    }

    async connect() {
        return new Promise((resolve, reject) => {
            this.ws = new WebSocket(`${this.baseUrl}/api/v1/mcp`);

            this.ws.on('open', async () => {
                console.log('Connected to UMA MCP WebSocket');
                await this.initialize();
                resolve();
            });

            this.ws.on('message', (data) => {
                const response = JSON.parse(data);
                this.handleResponse(response);
            });

            this.ws.on('error', reject);
        });
    }

    async initialize() {
        const response = await this.sendRequest('initialize', {
            protocolVersion: '2024-11-05',
            capabilities: {},
            clientInfo: {
                name: 'uma-client',
                version: '1.0.0'
            }
        });
        console.log('MCP initialized:', response.result.serverInfo);
    }

    async sendRequest(method, params = {}) {
        return new Promise((resolve, reject) => {
            const id = this.requestId++;
            const message = {
                jsonrpc: '2.0',
                id: id,
                method: method,
                params: params
            };

            this.pendingRequests.set(id, { resolve, reject });
            this.ws.send(JSON.stringify(message));
        });
    }

    handleResponse(response) {
        if (response.id && this.pendingRequests.has(response.id)) {
            const { resolve, reject } = this.pendingRequests.get(response.id);
            this.pendingRequests.delete(response.id);

            if (response.error) {
                reject(new Error(response.error.message));
            } else {
                resolve(response);
            }
        }
    }

    async getSystemInfo() {
        const response = await this.sendRequest('tools/call', {
            name: 'get_system_info',
            arguments: {}
        });
        return response.result.content[0].text;
    }

    async getUPSStatus() {
        const response = await this.sendRequest('tools/call', {
            name: 'get_ups_status',
            arguments: {}
        });
        return response.result.content[0].text;
    }

    async listTools() {
        const response = await this.sendRequest('tools/list', {});
        return response.result.tools;
    }
}

// Usage
async function main() {
    const client = new UMAMCPClient();
    await client.connect();

    const systemInfo = await client.getSystemInfo();
    console.log('System Info:', systemInfo);

    const tools = await client.listTools();
    console.log(`Available tools: ${tools.length}`);
}

main().catch(console.error);
```

### Python WebSocket Integration

#### Real-time Monitoring with asyncio

```python
import asyncio
import json
import websockets
import logging

class UMAMonitor:
    def __init__(self, base_url="ws://192.168.20.21:34600"):
        self.base_url = base_url
        self.websocket = None
        self.reconnect_delay = 1
        self.max_reconnect_delay = 30

    async def connect(self):
        """Connect to UMA WebSocket with automatic reconnection"""
        while True:
            try:
                uri = f"{self.base_url}/api/v1/ws"
                self.websocket = await websockets.connect(uri)
                logging.info("Connected to UMA WebSocket")

                # Subscribe to channels
                await self.subscribe(['system.stats', 'cpu.stats', 'docker.events'])

                # Reset reconnect delay on successful connection
                self.reconnect_delay = 1

                # Listen for messages
                await self.listen()

            except Exception as e:
                logging.error(f"Connection failed: {e}")
                await asyncio.sleep(self.reconnect_delay)
                self.reconnect_delay = min(self.reconnect_delay * 2, self.max_reconnect_delay)

    async def subscribe(self, channels):
        """Subscribe to WebSocket channels"""
        message = {
            "type": "subscribe",
            "channels": channels
        }
        await self.websocket.send(json.dumps(message))
        logging.info(f"Subscribed to channels: {channels}")

    async def listen(self):
        """Listen for WebSocket messages"""
        async for message in self.websocket:
            try:
                event = json.loads(message)
                await self.handle_event(event)
            except json.JSONDecodeError as e:
                logging.error(f"Failed to parse message: {e}")

    async def handle_event(self, event):
        """Handle incoming WebSocket events"""
        event_type = event.get('event_type')
        data = event.get('data', {})

        if event_type == 'system.stats':
            print(f"System: CPU {data.get('cpu_percent')}%, Memory {data.get('memory_percent')}%")
        elif event_type == 'docker.events':
            print(f"Docker: {data.get('action')} - {data.get('container_name')}")
        elif event_type == 'temperature.alert':
            print(f"Temperature Alert: {data.get('message')}")

# Usage
async def main():
    monitor = UMAMonitor()
    await monitor.connect()

if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    asyncio.run(main())
```

#### MCP Client with Python

```python
import asyncio
import json
import websockets
import uuid

class UMAMCPClient:
    def __init__(self, base_url="ws://192.168.20.21:34600"):
        self.base_url = base_url
        self.websocket = None
        self.pending_requests = {}

    async def connect(self):
        """Connect to MCP WebSocket"""
        uri = f"{self.base_url}/api/v1/mcp"
        self.websocket = await websockets.connect(uri)
        print("Connected to UMA MCP WebSocket")

        # Start message handler
        asyncio.create_task(self.message_handler())

        # Initialize MCP protocol
        await self.initialize()

    async def initialize(self):
        """Initialize MCP protocol"""
        response = await self.send_request("initialize", {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {
                "name": "uma-python-client",
                "version": "1.0.0"
            }
        })
        print(f"MCP initialized: {response['result']['serverInfo']}")

    async def send_request(self, method, params=None):
        """Send JSON-RPC 2.0 request"""
        request_id = str(uuid.uuid4())
        message = {
            "jsonrpc": "2.0",
            "id": request_id,
            "method": method,
            "params": params or {}
        }

        # Create future for response
        future = asyncio.Future()
        self.pending_requests[request_id] = future

        # Send request
        await self.websocket.send(json.dumps(message))

        # Wait for response
        return await future

    async def message_handler(self):
        """Handle incoming messages"""
        async for message in self.websocket:
            try:
                response = json.loads(message)
                request_id = response.get('id')

                if request_id and request_id in self.pending_requests:
                    future = self.pending_requests.pop(request_id)
                    if 'error' in response:
                        future.set_exception(Exception(response['error']['message']))
                    else:
                        future.set_result(response)

            except json.JSONDecodeError as e:
                print(f"Failed to parse message: {e}")

    async def get_system_info(self):
        """Get system information using MCP tool"""
        response = await self.send_request("tools/call", {
            "name": "get_system_info",
            "arguments": {}
        })
        return response['result']['content'][0]['text']

    async def get_ups_status(self):
        """Get UPS status using MCP tool"""
        response = await self.send_request("tools/call", {
            "name": "get_ups_status",
            "arguments": {}
        })
        return response['result']['content'][0]['text']

    async def list_tools(self):
        """List all available MCP tools"""
        response = await self.send_request("tools/list", {})
        return response['result']['tools']

# Usage
async def main():
    client = UMAMCPClient()
    await client.connect()

    # Get system information
    system_info = await client.get_system_info()
    print(f"System Info: {system_info}")

    # List available tools
    tools = await client.list_tools()
    print(f"Available tools: {len(tools)}")

    # Get UPS status
    ups_status = await client.get_ups_status()
    print(f"UPS Status: {ups_status}")

if __name__ == "__main__":
    asyncio.run(main())
```

### Home Assistant WebSocket Integration

#### Custom Component Configuration

```yaml
# configuration.yaml
sensor:
  - platform: websocket
    resource: ws://192.168.20.21:34600/api/v1/ws
    name: "Unraid CPU Usage"
    value_template: "{{ value_json.data.cpu_percent }}"
    unit_of_measurement: "%"
    device_class: "power_factor"
    state_class: "measurement"
    json_attributes_path: "$.data"
    json_attributes:
      - memory_percent
      - uptime
      - load_average
    subscription_message: |
      {
        "type": "subscribe",
        "channels": ["system.stats"]
      }

  - platform: websocket
    resource: ws://192.168.20.21:34600/api/v1/ws
    name: "Unraid Memory Usage"
    value_template: "{{ value_json.data.memory_percent }}"
    unit_of_measurement: "%"
    device_class: "power_factor"
    state_class: "measurement"
    subscription_message: |
      {
        "type": "subscribe",
        "channels": ["memory.stats"]
      }

binary_sensor:
  - platform: websocket
    resource: ws://192.168.20.21:34600/api/v1/ws
    name: "Unraid Array Status"
    value_template: "{{ value_json.data.array_status == 'started' }}"
    device_class: "running"
    subscription_message: |
      {
        "type": "subscribe",
        "channels": ["storage.status"]
      }
```

#### Home Assistant Custom Integration Pattern

```python
# custom_components/unraid_uma/sensor.py
import asyncio
import json
import logging
import websockets
from homeassistant.components.sensor import SensorEntity
from homeassistant.core import HomeAssistant
from homeassistant.helpers.entity_platform import AddEntitiesCallback
from homeassistant.helpers.typing import ConfigType, DiscoveryInfoType

_LOGGER = logging.getLogger(__name__)

class UnraidWebSocketSensor(SensorEntity):
    """Unraid WebSocket sensor."""

    def __init__(self, name, websocket_url, channel, value_template):
        self._name = name
        self._websocket_url = websocket_url
        self._channel = channel
        self._value_template = value_template
        self._state = None
        self._websocket = None

    @property
    def name(self):
        return self._name

    @property
    def state(self):
        return self._state

    async def async_added_to_hass(self):
        """Connect to WebSocket when added to hass."""
        await self.connect_websocket()

    async def connect_websocket(self):
        """Connect to UMA WebSocket."""
        try:
            self._websocket = await websockets.connect(self._websocket_url)

            # Subscribe to channel
            subscribe_message = {
                "type": "subscribe",
                "channels": [self._channel]
            }
            await self._websocket.send(json.dumps(subscribe_message))

            # Start listening
            asyncio.create_task(self.listen_for_updates())

        except Exception as e:
            _LOGGER.error(f"Failed to connect to WebSocket: {e}")

    async def listen_for_updates(self):
        """Listen for WebSocket updates."""
        try:
            async for message in self._websocket:
                data = json.loads(message)
                if data.get('event_type') == self._channel:
                    # Apply value template
                    self._state = self.extract_value(data)
                    self.async_write_ha_state()

        except Exception as e:
            _LOGGER.error(f"WebSocket error: {e}")
            # Implement reconnection logic here

    def extract_value(self, data):
        """Extract value using template."""
        # Implement template parsing logic
        return data.get('data', {}).get('cpu_percent', 0)
```

---

## WebSocket Testing and Debugging

### Step-by-Step WebSocket Testing

#### 1. Test WebSocket Connectivity

```bash
# Test basic WebSocket upgrade
curl -i -H "Connection: Upgrade" -H "Upgrade: websocket" \
     -H "Sec-WebSocket-Version: 13" -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
     http://192.168.20.21:34600/api/v1/ws

# Expected: HTTP/1.1 101 Switching Protocols
```

#### 2. Test Real-time Monitoring

```bash
# Connect and subscribe to system stats
echo '{"type":"subscribe","channels":["system.stats"]}' | websocat ws://192.168.20.21:34600/api/v1/ws

# Expected: Subscription acknowledgment followed by periodic system.stats events
```

#### 3. Test MCP Protocol

```bash
# Initialize MCP connection
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | websocat ws://192.168.20.21:34600/api/v1/mcp

# Expected: Initialization response with server info
```

#### 4. Test Tool Execution

```bash
# Execute system info tool
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_system_info","arguments":{}}}' | websocat ws://192.168.20.21:34600/api/v1/mcp

# Expected: Tool execution result with system information
```

### Common WebSocket Issues and Solutions

#### Connection Issues

| Issue | Symptoms | Diagnosis | Solution |
|-------|----------|-----------|----------|
| **Connection Refused** | `curl: (7) Failed to connect` | UMA service not running | Check service: `ps aux \| grep uma` |
| **Upgrade Failed** | HTTP 400/404 response | Wrong endpoint or headers | Verify URL and WebSocket headers |
| **Timeout** | Connection hangs | Network/firewall issues | Check network connectivity |
| **Rate Limited** | Connection drops quickly | Too many messages | Reduce message frequency |

#### Message Format Issues

```bash
# Test invalid JSON (should return error)
echo 'invalid json' | websocat ws://192.168.20.21:34600/api/v1/ws

# Test invalid channel (should return error)
echo '{"type":"subscribe","channels":["invalid.channel"]}' | websocat ws://192.168.20.21:34600/api/v1/ws

# Test missing parameters (should return error)
echo '{"type":"subscribe"}' | websocat ws://192.168.20.21:34600/api/v1/ws
```

### WebSocket Debugging Tools

#### Using websocat for Interactive Testing

```bash
# Interactive mode with verbose output
websocat -v ws://192.168.20.21:34600/api/v1/ws

# Save WebSocket traffic to file
websocat ws://192.168.20.21:34600/api/v1/ws --dump-traffic

# Test with custom headers
websocat ws://192.168.20.21:34600/api/v1/ws -H "User-Agent: TestClient/1.0"
```

#### Using curl for Protocol Testing

```bash
# Test WebSocket upgrade with verbose output
curl -v -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" \
     -H "Sec-WebSocket-Version: 13" -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
     http://192.168.20.21:34600/api/v1/ws

# Test with different WebSocket versions
curl -i -H "Connection: Upgrade" -H "Upgrade: websocket" \
     -H "Sec-WebSocket-Version: 8" -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
     http://192.168.20.21:34600/api/v1/ws
```

### Performance Testing

#### Connection Load Testing

```bash
# Test multiple concurrent connections
for i in {1..5}; do
  websocat ws://192.168.20.21:34600/api/v1/ws &
done

# Monitor connection count
curl -s http://192.168.20.21:34600/api/v1/mcp/status | jq '.data.active_connections'
```

#### Message Rate Testing

```bash
# Test message rate limits
for i in {1..150}; do
  echo '{"type":"ping"}' | websocat ws://192.168.20.21:34600/api/v1/ws &
done

# Expected: Rate limit errors after 100 messages/minute
```

### Monitoring WebSocket Connections

#### Check Active Connections

```bash
# Check MCP connection count
curl -s http://192.168.20.21:34600/api/v1/mcp/status | jq '{active_connections, max_connections}'

# Check UMA service status
curl -s http://192.168.20.21:34600/api/v1/health | jq '.checks'
```

#### Log Analysis

```bash
# Monitor UMA logs for WebSocket activity
ssh root@192.168.20.21 'tail -f /tmp/uma.log | grep -i websocket'

# Check for connection errors
ssh root@192.168.20.21 'grep -i "websocket\|error" /tmp/uma.log | tail -20'
```

### WebSocket Security Testing

#### Origin Validation Testing

```bash
# Test with valid origin (should work)
curl -i -H "Connection: Upgrade" -H "Upgrade: websocket" \
     -H "Origin: http://192.168.20.21" \
     -H "Sec-WebSocket-Version: 13" -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
     http://192.168.20.21:34600/api/v1/ws

# Test with external origin (should be rejected)
curl -i -H "Connection: Upgrade" -H "Upgrade: websocket" \
     -H "Origin: http://malicious-site.com" \
     -H "Sec-WebSocket-Version: 13" -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
     http://192.168.20.21:34600/api/v1/ws
```

#### Message Size Testing

```bash
# Test large message (should be rejected if > 1MB)
python3 -c "
import json
large_data = 'x' * (1024 * 1024 + 1)  # 1MB + 1 byte
message = json.dumps({'type': 'subscribe', 'channels': [large_data]})
print(message)
" | websocat ws://192.168.20.21:34600/api/v1/ws
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

| Endpoint | Protocol | Description | Max Connections | Rate Limit |
|----------|----------|-------------|-----------------|------------|
| `/api/v1/ws` | WebSocket | Real-time monitoring with 25+ channels | 50 | 100 msg/min |
| `/api/v1/mcp` | WebSocket | MCP JSON-RPC 2.0 with 51+ tools | 100 | 100 msg/min |

### WebSocket Channels (Real-time Monitoring)

| Channel Category | Available Channels | Description |
|------------------|-------------------|-------------|
| **System** | `system.stats`, `system.health`, `system.load`, `cpu.stats`, `memory.stats`, `network.stats` | System performance metrics |
| **Storage** | `storage.status`, `disk.stats`, `array.status`, `parity.status`, `cache.status` | Storage monitoring |
| **Containers** | `docker.events`, `container.stats`, `container.health`, `image.events` | Docker monitoring |
| **VMs** | `vm.events`, `vm.stats`, `vm.health` | Virtual machine monitoring |
| **Alerts** | `temperature.alert`, `resource.alert`, `security.alert`, `system.alert` | Alert notifications |
| **Infrastructure** | `ups.status`, `fan.status`, `power.status` | Hardware monitoring |
| **Operations** | `task.progress`, `backup.status`, `update.status` | Operational events |

### MCP Tools (AI Agent Integration)

| Tool Category | Example Tools | Description |
|---------------|---------------|-------------|
| **System** | `get_system_info`, `get_disk_usage`, `get_filesystem_info` | System information retrieval |
| **Storage** | `get_cache_status`, `get_array_status`, `get_disk_health` | Storage management |
| **Services** | `get_vm_details`, `get_docker_status`, `get_service_status` | Service monitoring |
| **Security** | `get_security_status`, `get_firewall_status`, `get_user_access` | Security analysis |
| **Configuration** | `get_share_config`, `list_plugins`, `get_network_config` | Configuration management |

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
