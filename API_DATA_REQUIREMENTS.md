# Unraid REST API Specification

**Document Version**: 1.0  
**Date**: June 14, 2025  
**Purpose**: Complete REST API specification for Unraid server monitoring and management

## Overview

This document provides a comprehensive specification for a REST API that exposes Unraid server functionality for monitoring, management, and automation. The API enables external applications to monitor system health, control Docker containers and VMs, execute scripts, and manage server operations.

## API Coverage

The API provides **31 different data endpoints** across 4 categories:

- **16 System Monitoring APIs** (CPU, memory, storage, network, UPS)
- **10 Health Monitoring APIs** (disk health, array status, diagnostics)
- **Dynamic Control APIs** (Docker containers, VMs)
- **System Management APIs** (power control, script execution)

## Core Data Structure

The API returns data in this structure:

```json
{
  "system_stats": {
    "cpu_usage": 45.2,
    "memory_usage": { ... },
    "temperature_data": { ... },
    "network_stats": { ... },
    "individual_disks": [ ... ],
    "ups_info": { ... }
  },
  "docker_containers": [ ... ],
  "vms": [ ... ],
  "user_scripts": [ ... ],
  "disk_config": { ... },
  "parity_info": { ... }
}
```

---

## 1. System Monitoring APIs

### 1.1 CPU Usage API
**API Endpoint**: `GET /api/system/stats` (cpu_usage field)  
**Purpose**: Real-time CPU utilization monitoring

**Required Data**:

```json
{
  "system_stats": {
    "cpu_usage": 45.2,
    "cpu_cores": 8,
    "cpu_arch": "x86_64",
    "cpu_model": "Intel Core i7-9700K",
    "cpu_threads_per_core": 2,
    "cpu_sockets": 1,
    "cpu_max_freq": 3600,
    "cpu_min_freq": 800,
    "cpu_temp": 45,
    "cpu_temp_critical": false,
    "cpu_temp_warning": false,
    "cpu_load_averages": {
      "load_1m": 0.85,
      "load_5m": 0.92,
      "load_15m": 0.78
    }
  }
}
```

**Critical Fields**:
- `cpu_usage` (float): CPU percentage 0-100 **[REQUIRED]**
- `cpu_cores` (int): Number of CPU cores
- `cpu_model` (string): Processor model name

### 1.2 Memory Usage API

**API Endpoint**: `GET /api/system/stats` (memory_usage field)  
**Purpose**: Memory utilization with detailed breakdown

**Required Data**:

```json
{
  "system_stats": {
    "memory_usage": {
      "percentage": 67.8,
      "total": "32 GB",
      "used": "21.7 GB",
      "free_memory": "10.3 GB",
      "cached": "8.2 GB",
      "buffers": "1.1 GB",
      "available": "10.3 GB",
      "system_memory": "4.2 GB",
      "system_memory_percentage": 13.1,
      "vm_memory": "8.5 GB",
      "vm_memory_percentage": 26.6,
      "docker_memory": "2.8 GB",
      "docker_memory_percentage": 8.8,
      "zfs_memory": "6.2 GB",
      "zfs_memory_percentage": 19.4
    }
  }
}
```

**Critical Fields**:
- `percentage` (float): Memory usage percentage **[REQUIRED]**
- `total`, `used`, `available` (string): Memory amounts with units

### 1.3 Temperature APIs

**API Endpoints**: `GET /api/system/stats` (temperature_data field)  
**Purpose**: Hardware temperature monitoring

**Required Data**:

```json
{
  "system_stats": {
    "temperature_data": {
      "sensors": {
        "coretemp-isa-0000": {
          "Package id 0": {
            "temp1_input": 45000
          },
          "Core 0": {
            "temp2_input": 43000
          }
        },
        "nct6798-isa-0290": {
          "SYSTIN": {
            "temp1_input": 38000
          },
          "AUXTIN0": {
            "temp2_input": 42000
          }
        }
      },
      "fans": {
        "fan1": {
          "label": "CPU Fan",
          "input": 1200,
          "min": 300,
          "max": 2000
        }
      }
    }
  }
}
```

**Critical Fields**:
- Temperature values in millidegrees Celsius (45000 = 45°C)
- Support for multiple sensor chips and temperature inputs
- Fan data with RPM values

### 1.4 System Information Sensors

**Uptime Sensor**:

```json
{
  "system_stats": {
    "uptime": "7 days, 14:23:45"
  }
}
```

**Intel GPU Sensor** (optional):

```json
{
  "system_stats": {
    "intel_gpu": {
      "usage": 15.4,
      "model": "Intel UHD Graphics 630"
    }
  }
}
```

**Storage System Sensors**:

```json
{
  "system_stats": {
    "docker_vdisk": {
      "usage": 45.2,
      "size": "100 GB",
      "used": "45.2 GB"
    },
    "log_filesystem": {
      "usage": 23.7,
      "size": "128 MB",
      "used": "30.3 MB"
    },
    "boot_usage": {
      "usage": 12.4,
      "size": "32 GB",
      "used": "3.97 GB"
    }
  }
}
```

---

## 2. Storage Monitoring APIs

### 2.1 Array and Disk Sensors

**Array Sensor** (`sensor.{hostname}_array`):

```json
{
  "system_stats": {
    "array": {
      "status": "Started",
      "usage": 78.5,
      "total": "8 TB",
      "used": "6.28 TB",
      "free": "1.72 TB"
    }
  }
}
```

**Individual Disk Sensors** (`sensor.{hostname}_disk{N}_usage`):

```json
{
  "system_stats": {
    "individual_disks": [
      {
        "name": "disk1",
        "total": 4000000000000,
        "used": 2800000000000,
        "free": 1200000000000,
        "percentage": 70.0,
        "mount_point": "/mnt/disk1",
        "filesystem": "xfs",
        "state": "active",
        "temperature": 42,
        "health": "PASSED",
        "spin_down_delay": "15 minutes",
        "smart_data": {
          "smart_status": true,
          "ata_smart_attributes": {
            "table": [
              {
                "name": "Reallocated_Sector_Ct",
                "value": 100,
                "worst": 100,
                "thresh": 36,
                "raw": { "value": 0 },
                "when_failed": ""
              }
            ]
          }
        }
      }
    ]
  }
}
```

**Critical Fields**:
- `name` (string): Disk identifier (disk1, disk2, cache, etc.) **[REQUIRED]**
- `state` (string): "active" or "standby" **[REQUIRED]**
- `smart_data.smart_status` (boolean): Overall SMART health **[REQUIRED]**
- Storage amounts in bytes for calculations

### 2.2 SMART Health Monitoring

**Required SMART Attributes for Health Monitoring**:

```json
{
  "smart_data": {
    "smart_status": true,
    "ata_smart_attributes": {
      "table": [
        {
          "name": "Reallocated_Sector_Ct",
          "raw": { "value": 0 }
        },
        {
          "name": "Current_Pending_Sector",
          "raw": { "value": 0 }
        },
        {
          "name": "Offline_Uncorrectable",
          "raw": { "value": 0 }
        },
        {
          "name": "Reallocated_Event_Count",
          "raw": { "value": 0 }
        },
        {
          "name": "Temperature_Celsius",
          "raw": { "value": 42 }
        }
      ]
    }
  }
}
```

**Critical Requirements**:
- Raw values must be 0 for critical attributes (reallocated sectors, pending sectors)
- Temperature should be reasonable (15-80°C typical range)
- Handle standby disks by caching last known values

---

## 3. Network Monitoring APIs

### 3.1 Network Interface APIs

**API Endpoints**: `GET /api/system/stats` (network_stats field)

**Required Data**:

```json
{
  "system_stats": {
    "network_stats": {
      "eth0": {
        "connected": true,
        "rx_bytes": 1250000000,
        "tx_bytes": 850000000,
        "rx_packets": 1500000,
        "tx_packets": 1200000,
        "rx_errors": 0,
        "tx_errors": 0,
        "speed": "1000baseT/Full",
        "duplex": "full"
      },
      "bond0": {
        "connected": true,
        "rx_bytes": 2500000000,
        "tx_bytes": 1800000000
      }
    }
  }
}
```

**Critical Fields**:
- `connected` (boolean): Interface connection status **[REQUIRED]**
- `rx_bytes`, `tx_bytes` (int): Cumulative byte counts **[REQUIRED]**
- Only include active/connected interfaces

---

## 4. UPS Monitoring APIs (Optional)

### 4.1 UPS Power and Energy APIs

**API Endpoints**: `GET /api/system/stats` (ups_info field)

**Required Data**:

```json
{
  "system_stats": {
    "ups_info": {
      "NOMPOWER": "1500",
      "LOADPCT": "45.2",
      "MODEL": "CyberPower CP1500PFCLCD",
      "BCHARGE": "100",
      "TIMELEFT": "125.5",
      "STATUS": "ONLINE"
    }
  }
}
```

**Critical Fields**:
- `NOMPOWER` (string): Nominal power in watts **[REQUIRED]**
- `LOADPCT` (string): Load percentage **[REQUIRED]**
- Power calculation: `(NOMPOWER × LOADPCT) / 100`

---

## 5. Docker Container Management APIs

### 5.1 Container Control APIs

**API Endpoints**: `GET /api/docker/containers`, `POST /api/docker/containers/{name}/start`, `POST /api/docker/containers/{name}/stop`

**Required Data**:

```json
{
  "docker_containers": [
    {
      "name": "plex",
      "id": "abc123def456",
      "state": "running",
      "status": "Up 2 days",
      "image": "plexinc/pms-docker:latest"
    },
    {
      "name": "nginx",
      "id": "def456ghi789",
      "state": "stopped",
      "status": "Exited (0) 1 hour ago",
      "image": "nginx:latest"
    }
  ]
}
```

**Critical Fields**:
- `name` (string): Container name **[REQUIRED]**
- `state` (string): "running", "stopped", or "paused" **[REQUIRED]**
- `id` (string): Container ID for management operations

**Required API Commands**:
- `docker start {container_name}`
- `docker stop {container_name}`
- `docker exec {container_name} {command}`

---

## 6. Virtual Machine Management APIs

### 6.1 VM Control APIs

**API Endpoints**: `GET /api/vms`, `POST /api/vms/{name}/start`, `POST /api/vms/{name}/stop`

**Required Data**:

```json
{
  "vms": [
    {
      "name": "Windows 10",
      "state": "running",
      "uuid": "12345678-1234-1234-1234-123456789abc",
      "os_type": "windows",
      "memory": "8192",
      "vcpus": "4",
      "autostart": true
    },
    {
      "name": "Ubuntu Server",
      "state": "stopped",
      "uuid": "87654321-4321-4321-4321-cba987654321",
      "os_type": "linux",
      "memory": "4096",
      "vcpus": "2",
      "autostart": false
    }
  ]
}
```

**Critical Fields**:
- `name` (string): VM name **[REQUIRED]**
- `state` (string): "running", "stopped", or "paused" **[REQUIRED]**
- `uuid` (string): VM UUID for management

**Required API Commands**:
- `virsh start {vm_name}`
- `virsh shutdown {vm_name}`
- `virsh destroy {vm_name}` (force stop)

---

## 7. User Script Management APIs

### 7.1 Script Execution APIs

**API Endpoints**: `GET /api/scripts`, `POST /api/scripts/{name}/execute`

**Required Data**:

```json
{
  "user_scripts": [
    {
      "name": "backup_script",
      "background_only": false,
      "foreground_only": false,
      "description": "Daily backup script",
      "path": "/boot/config/plugins/user.scripts/scripts/backup_script/script"
    },
    {
      "name": "maintenance",
      "background_only": true,
      "foreground_only": false,
      "description": "System maintenance tasks"
    }
  ]
}
```

**Critical Fields**:
- `name` (string): Script identifier **[REQUIRED]**
- `background_only` (boolean): Can only run in background
- `foreground_only` (boolean): Can only run in foreground

**Required API Commands**:
- Foreground: `/boot/config/plugins/user.scripts/scripts/{script_name}/script`
- Background: `/boot/config/plugins/user.scripts/scripts/{script_name}/script &`

---

## 8. System Health Diagnostics APIs

### 8.1 Health Status APIs

**Connectivity Status**:
- API Available: `system_stats` data exists
- API Unavailable: No `system_stats` data

**Service Status**:
- Docker: Available if `docker_containers` array exists
- VM: Available if `vms` array exists

**Array Health**: `binary_sensor.{hostname}_array_status`

```json
{
  "system_stats": {
    "array": {
      "status": "Started",
      "health": "OK"
    }
  }
}
```

### 8.2 Parity Monitoring (Optional)

**Parity Check Sensor**: `binary_sensor.{hostname}_parity_check`

**Required Data from `mdcmd status`**:

```json
{
  "parity_info": {
    "diskNumber.0": "0",
    "diskName.0": "Parity",
    "diskState.0": "active",
    "rdevName.0": "/dev/sda1",
    "rdevStatus.0": "DISK_OK"
  }
}
```

**Parity Check Status**:

```json
{
  "system_stats": {
    "parity_check": {
      "status": "Success",
      "progress": 0,
      "speed": "N/A",
      "errors": 0,
      "last_check": "2024-06-01 02:00",
      "next_check": "2024-07-01 02:00",
      "duration": "4h 32m",
      "last_status": "Success"
    }
  }
}
```

---

## 9. Configuration Data

### 9.1 Disk Configuration

**Required for Spin-Down Management**:

```json
{
  "disk_config": {
    "spindownDelay": "15",
    "diskSpindownDelay.1": "30",
    "diskSpindownDelay.2": "-1"
  }
}
```

**Values**:
- Global: `spindownDelay` (minutes, "0" = never)
- Per-disk: `diskSpindownDelay.{N}` ("-1" = use global)

---

## 10. System Control Requirements

### 10.1 System Management Buttons
**Entities**: `button.{hostname}_reboot`, `button.{hostname}_shutdown`

**Required API Commands**:
- Shutdown: `shutdown -h +0`
- Reboot: `shutdown -r +0`

### 10.2 Additional API Commands

**General Command Execution**:
- `{any_shell_command}` - For general system commands

**Container Management**:
- `docker exec {container} {command}` - Execute in container

**Script Management**:
- Execute: `/path/to/script`
- Stop: `pkill -f script_name`

---

## 11. API Implementation Guidelines

### 11.1 Update Frequencies
- **System Stats**: Every 30 seconds
- **Disk Data**: Every 60 seconds (respect standby)
- **Network Stats**: Every 30 seconds
- **Docker/VM**: Every 30 seconds
- **SMART Data**: Every 5 minutes (active disks only)

### 11.2 Standby Disk Handling
- Detect standby via `state` field
- Cache last known values for standby disks
- Don't execute commands that wake sleeping disks
- Return cached SMART data for standby disks

### 11.3 Error Handling
- Return partial data when some components fail
- Use consistent error response format
- Handle SSH connection failures gracefully
- Provide meaningful error messages

### 11.4 Data Validation
- Validate numeric ranges (temperatures, percentages)
- Handle missing/null data gracefully
- Ensure consistent data types
- Validate SMART attribute raw values

### 11.5 Performance Considerations
- Cache expensive operations (SMART reads, temperature sensors)
- Implement rate limiting for disk operations
- Use background tasks for slow operations
- Minimize disk wake-ups

---

## 12. Testing Checklist

### 12.1 Required Test Scenarios
- [ ] All system sensors return valid data
- [ ] Disk standby state properly detected and handled
- [ ] SMART data parsing for various disk types
- [ ] Network interface detection (ignore virtual interfaces)
- [ ] Docker container start/stop operations
- [ ] VM start/stop operations
- [ ] User script execution (foreground/background)
- [ ] System shutdown/reboot commands
- [ ] UPS data parsing (if UPS present)
- [ ] Parity check status detection
- [ ] Temperature sensor detection across different hardware
- [ ] Error handling for offline/failed components

### 12.2 Edge Cases
- [ ] No disks in standby
- [ ] All disks in standby
- [ ] Missing SMART data
- [ ] Network interfaces with no traffic
- [ ] Stopped Docker service
- [ ] Stopped VM service
- [ ] No UPS connected
- [ ] No parity disk configured
- [ ] Temperature sensors not available

---

## Summary

This REST API specification provides a comprehensive monitoring and control system for Unraid servers with:

- **Real-time monitoring**: CPU, RAM, temperatures, network, storage
- **Health diagnostics**: SMART monitoring, array health, service status
- **Remote control**: Docker containers, VMs, system power, user scripts
- **Energy tracking**: UPS power monitoring for energy management systems

**Total API Endpoints**: 31+ base endpoints plus dynamic endpoints based on configuration
**API Implementation**: REST HTTP with JSON payloads
**Update Frequency**: 30-60 second intervals with intelligent caching

The API handles edge cases like disk standby states, missing hardware, and service failures gracefully while providing comprehensive monitoring capabilities for Unraid server management through external automation platforms.

---

## REST API Endpoints for GO Implementation

### Required HTTP Endpoints

The GO REST API plugin must expose these endpoints to provide all necessary data:

#### 1. System Statistics Endpoint
**GET** `/api/system/stats`

Returns comprehensive system information including CPU, memory, temperatures, and storage data.

**Response Structure**:
```json
{
  "system_stats": {
    "cpu_usage": 45.2,
    "memory_usage": { ... },
    "temperature_data": { ... },
    "network_stats": { ... },
    "individual_disks": [ ... ],
    "ups_info": { ... },
    "array": { ... },
    "docker_vdisk": { ... },
    "log_filesystem": { ... },
    "boot_usage": { ... },
    "uptime": "7 days, 14:23:45",
    "intel_gpu": { ... },
    "parity_check": { ... }
  }
}
```

#### 2. Docker Management Endpoints

**GET** `/api/docker/containers`
```json
{
  "docker_containers": [
    {
      "name": "plex",
      "id": "abc123def456",
      "state": "running",
      "status": "Up 2 days",
      "image": "plexinc/pms-docker:latest"
    }
  ]
}
```

**POST** `/api/docker/containers/{name}/start`
**POST** `/api/docker/containers/{name}/stop`
**POST** `/api/docker/containers/{name}/exec`

Request body for exec:
```json
{
  "command": "ls -la /app",
  "detached": false
}
```

#### 3. Virtual Machine Management Endpoints

**GET** `/api/vms`
```json
{
  "vms": [
    {
      "name": "Windows 10",
      "state": "running",
      "uuid": "12345678-1234-1234-1234-123456789abc",
      "os_type": "windows",
      "memory": "8192",
      "vcpus": "4",
      "autostart": true
    }
  ]
}
```

**POST** `/api/vms/{name}/start`
**POST** `/api/vms/{name}/stop`
**POST** `/api/vms/{name}/force-stop`

#### 4. User Scripts Endpoints

**GET** `/api/scripts`
```json
{
  "user_scripts": [
    {
      "name": "backup_script",
      "background_only": false,
      "foreground_only": false,
      "description": "Daily backup script",
      "path": "/boot/config/plugins/user.scripts/scripts/backup_script/script"
    }
  ]
}
```

**POST** `/api/scripts/{name}/execute`

Request body:
```json
{
  "background": false
}
```

**POST** `/api/scripts/{name}/stop`

#### 5. System Control Endpoints

**POST** `/api/system/shutdown`
**POST** `/api/system/reboot`

Optional request body for both:
```json
{
  "delay": 0
}
```

#### 6. Command Execution Endpoint

**POST** `/api/system/execute`

Request body:
```json
{
  "command": "df -h",
  "timeout": 30
}
```

Response:
```json
{
  "exit_code": 0,
  "stdout": "Filesystem      Size  Used Avail Use% Mounted on\n...",
  "stderr": "",
  "execution_time": 0.123
}
```

#### 7. Configuration Endpoints

**GET** `/api/config/disk`
```json
{
  "disk_config": {
    "spindownDelay": "15",
    "diskSpindownDelay.1": "30",
    "diskSpindownDelay.2": "-1"
  }
}
```

**GET** `/api/config/parity`
```json
{
  "parity_info": {
    "diskNumber.0": "0",
    "diskName.0": "Parity",
    "diskState.0": "active",
    "rdevName.0": "/dev/sda1",
    "rdevStatus.0": "DISK_OK"
  }
}
```

### Authentication & Security

#### API Key Authentication
All endpoints should support API key authentication:

**Header**: `X-API-Key: your-api-key-here`

#### Rate Limiting
Implement rate limiting to prevent API abuse:
- **System stats**: Max 2 requests/second
- **Docker/VM operations**: Max 10 requests/minute per container/VM
- **System control**: Max 5 requests/minute
- **Command execution**: Max 20 requests/minute

### Error Response Format

All endpoints should return consistent error responses:

```json
{
  "error": {
    "code": "DISK_NOT_FOUND",
    "message": "Disk 'disk99' not found",
    "details": {
      "available_disks": ["disk1", "disk2", "cache"]
    }
  },
  "timestamp": "2025-06-14T10:30:00Z"
}
```

### HTTP Status Codes

- **200**: Success
- **400**: Bad Request (invalid parameters)
- **401**: Unauthorized (invalid API key)
- **404**: Not Found (container/VM/script not found)
- **409**: Conflict (operation not allowed in current state)
- **429**: Too Many Requests (rate limit exceeded)
- **500**: Internal Server Error
- **503**: Service Unavailable (Unraid service down)

### Data Refresh Strategy

#### Endpoint Caching Recommendations

1. **System Stats** (`/api/system/stats`):
   - Cache for 30 seconds
   - Include `Last-Modified` header
   - Support `If-Modified-Since` requests

2. **Docker/VM Lists**:
   - Cache for 30 seconds
   - Invalidate cache on state changes

3. **Script Lists**:
   - Cache for 5 minutes
   - Rarely changes

4. **Configuration Data**:
   - Cache for 10 minutes
   - Invalidate on configuration changes

### WebSocket Support (Optional Enhancement)

For real-time updates, consider implementing WebSocket endpoints:

**WS** `/api/ws/system/stats` - Real-time system statistics
**WS** `/api/ws/docker/events` - Docker container state changes
**WS** `/api/ws/vm/events` - VM state changes

### Implementation Priority

**Phase 1 (Core Functionality)**:
1. `/api/system/stats` - **Critical for all sensors**
2. `/api/docker/containers` + container control
3. `/api/vms` + VM control
4. `/api/system/execute` - **Critical for system operations**

**Phase 2 (Enhanced Features)**:
5. `/api/scripts` + script execution
6. `/api/system/shutdown` + `/api/system/reboot`
7. `/api/config/disk` + `/api/config/parity`

**Phase 3 (Advanced Features)**:
8. WebSocket support
9. Advanced error handling
10. Comprehensive logging

### Testing Endpoints

Create a test endpoint to verify data completeness:

**GET** `/api/test/endpoints`

Returns a checklist of all supported API functionality:

```json
{
  "endpoints_supported": {
    "system_monitoring": {
      "cpu_usage": true,
      "ram_usage": true,
      "cpu_temperature": true,
      "individual_disks": ["disk1", "disk2", "cache"],
      "network_interfaces": ["eth0", "bond0"],
      "ups_power": false
    },
    "health_monitoring": {
      "disk_health": ["disk1", "disk2"],
      "array_status": true,
      "parity_check": true
    },
    "control_apis": {
      "docker_containers": ["plex", "nginx"],
      "vms": ["Windows 10"],
      "system_control": true,
      "user_scripts": ["backup_script"]
    }
  },
  "missing_functionality": [
    "ups_info - UPS not detected",
    "intel_gpu - No Intel GPU found"
  ]
}
```

### API Documentation

The GO plugin should auto-generate OpenAPI/Swagger documentation available at:
**GET** `/api/docs` - Interactive API documentation

This ensures external applications can discover and validate all available endpoints and their expected data formats.

---
