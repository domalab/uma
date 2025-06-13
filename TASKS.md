# OmniRaid Development Tasks

This document outlines all missing features identified in the OmniRaid project analysis, organized by priority levels. Each task includes implementation details, API specifications, and testing criteria.

## 🚨 CRITICAL PRIORITY TASKS

### TASK-C1: Complete Home Assistant Integration Data Format
**Status**: ❌ Not Started  
**Estimated Time**: 3-4 days  
**HA Compatibility**: ✅ Critical for HA integration

#### Task Definition
Implement complete system data endpoints to match Home Assistant Unraid integration expectations. The current `/api/v1/storage/ha-format` endpoint only provides storage data but HA integration expects comprehensive system information.

#### API Endpoints
```bash
# Enhanced HA format endpoint (modify existing)
GET /api/v1/storage/ha-format     # Complete HA-compatible data format

# New system data endpoints
GET /api/v1/system/cpu            # CPU usage, cores, temperature
GET /api/v1/system/memory         # RAM usage breakdown  
GET /api/v1/system/temperature    # All temperature sensors
GET /api/v1/system/network        # Network interface statistics
GET /api/v1/system/ups            # UPS status and metrics
GET /api/v1/system/gpu            # Intel GPU usage
GET /api/v1/system/filesystems    # Docker vDisk, log filesystem, boot usage
```

#### Data Structures
**Enhanced HA Format Response:**
```json
{
  "system_stats": {
    "cpu_usage": {
      "usage": 15.5,
      "cores": 8,
      "threads": 16,
      "temperature": 45
    },
    "memory_usage": {
      "total": 16777216,
      "used": 8388608,
      "free": 8388608,
      "percentage": 50.0
    },
    "temperature_data": {
      "sensors": [
        {"name": "CPU", "value": 45, "unit": "°C"},
        {"name": "Motherboard", "value": 35, "unit": "°C"}
      ]
    },
    "network_stats": {
      "interfaces": [
        {"name": "eth0", "rx_bytes": 1024000, "tx_bytes": 512000}
      ]
    },
    "ups_info": {
      "status": "online",
      "battery_charge": 100,
      "runtime": 3600
    },
    "intel_gpu": {
      "usage": 5.2,
      "temperature": 40
    },
    "docker_vdisk": {
      "total": 21474836480,
      "used": 10737418240,
      "free": 10737418240
    },
    "log_filesystem": {
      "total": 134217728,
      "used": 67108864,
      "free": 67108864
    },
    "boot_usage": {
      "total": 134217728,
      "used": 33554432,
      "free": 100663296
    },
    "array_usage": { /* existing storage data */ },
    "array_state": { /* existing array data */ },
    "individual_disks": [ /* existing disk data */ ]
  },
  "docker_containers": [
    {
      "name": "plex",
      "state": "running",
      "status": "Up 2 hours",
      "cpu_usage": 5.5,
      "memory_usage": 512000000
    }
  ],
  "vms": [
    {
      "name": "Windows10",
      "state": "running",
      "cpu_usage": 25.0,
      "memory_usage": 4294967296
    }
  ],
  "disk_mappings": { /* existing disk mappings */ }
}
```

#### Dependencies
- Existing system monitoring plugins (CPU, memory, sensors)
- Docker plugin for container data
- VM plugin for virtual machine data
- UPS plugin for power data

#### Testing Criteria
- [ ] `/api/v1/storage/ha-format` returns complete HA-compatible format
- [ ] All system data endpoints return valid JSON
- [ ] CPU usage matches `top` command output
- [ ] Memory usage matches `free -m` output
- [ ] Temperature data matches sensor readings
- [ ] Docker container data matches `docker ps` output
- [ ] VM data matches `virsh list` output
- [ ] Response time under 2 seconds

---

### TASK-C2: Array Control Operations
**Status**: ❌ Not Started  
**Estimated Time**: 2-3 days  
**HA Compatibility**: ✅ Required for HA array control services

#### Task Definition
Implement core Unraid array management operations including start, stop, and parity check management. These are essential for remote array control via Home Assistant.

#### API Endpoints
```bash
POST /api/v1/array/start          # Start Unraid array
POST /api/v1/array/stop           # Stop Unraid array
GET  /api/v1/array/parity-check   # Get parity check status
POST /api/v1/array/parity-check   # Start parity check
DELETE /api/v1/array/parity-check # Cancel parity check
POST /api/v1/array/disk/add       # Add disk to array
POST /api/v1/array/disk/remove    # Remove disk from array
```

#### Data Structures
**Array Start Request:**
```json
{
  "maintenance_mode": false,
  "check_filesystem": true
}
```

**Array Stop Request:**
```json
{
  "force": false,
  "unmount_shares": true
}
```

**Parity Check Request:**
```json
{
  "type": "check",  // "check" or "correct"
  "priority": "normal"  // "low", "normal", "high"
}
```

**Array Operation Response:**
```json
{
  "success": true,
  "message": "Array start initiated",
  "operation_id": "array_start_20241201_123456",
  "estimated_time": 30
}
```

#### Dependencies
- Array state monitoring (existing)
- Unraid mdcmd command interface
- File system operations

#### Testing Criteria
- [ ] Array start command successfully starts stopped array
- [ ] Array stop command safely stops running array
- [ ] Parity check starts and reports progress
- [ ] Parity check can be cancelled mid-operation
- [ ] Operations fail gracefully with proper error messages
- [ ] Array state updates reflect operation results

---

### TASK-C3: System Power Management
**Status**: ❌ Not Started  
**Estimated Time**: 1-2 days  
**HA Compatibility**: ✅ Required for HA system control services

#### Task Definition
Implement essential system power control operations for remote server management via Home Assistant integration.

#### API Endpoints
```bash
POST /api/v1/system/shutdown      # Shutdown server
POST /api/v1/system/reboot        # Reboot server
POST /api/v1/system/sleep         # Sleep/suspend server (if supported)
POST /api/v1/system/wake          # Wake-on-LAN functionality
```

#### Data Structures
**Shutdown Request:**
```json
{
  "delay": 0,  // seconds
  "message": "Shutdown initiated via OmniRaid API"
}
```

**Reboot Request:**
```json
{
  "delay": 0,  // seconds
  "force": false
}
```

**Wake Request:**
```json
{
  "mac_address": "00:11:22:33:44:55",
  "broadcast_ip": "192.168.1.255"
}
```

**Power Operation Response:**
```json
{
  "success": true,
  "message": "Shutdown initiated",
  "scheduled_time": "2024-12-01T12:35:00Z"
}
```

#### Dependencies
- System command execution capabilities
- Network interface access for WOL
- Proper shutdown procedures

#### Testing Criteria
- [ ] Shutdown command safely shuts down server
- [ ] Reboot command restarts server properly
- [ ] Operations respect delay parameters
- [ ] Wake-on-LAN successfully wakes target systems
- [ ] Proper error handling for invalid requests

## 🔥 HIGH PRIORITY TASKS

### TASK-H1: User Script Management
**Status**: ❌ Not Started  
**Estimated Time**: 2-3 days  
**HA Compatibility**: ✅ Required for HA script execution services

#### Task Definition
Implement user script management capabilities to allow execution of custom Unraid user scripts via API, supporting Home Assistant automation workflows.

#### API Endpoints
```bash
GET  /api/v1/scripts              # List available user scripts
POST /api/v1/scripts/{name}/execute # Execute user script
POST /api/v1/scripts/{name}/stop  # Stop running script
GET  /api/v1/scripts/{name}/status # Get script execution status
GET  /api/v1/scripts/{name}/logs  # Get script execution logs
```

#### Data Structures
**Script List Response:**
```json
{
  "scripts": [
    {
      "name": "backup_appdata",
      "description": "Backup application data",
      "path": "/boot/config/plugins/user.scripts/scripts/backup_appdata/script",
      "status": "idle",
      "last_run": "2024-12-01T10:30:00Z",
      "last_result": "success"
    }
  ]
}
```

**Script Execute Request:**
```json
{
  "background": true,
  "arguments": ["--verbose", "--dry-run"]
}
```

#### Dependencies
- User Scripts plugin detection
- Process management for script execution
- Log file monitoring

#### Testing Criteria
- [ ] Lists all available user scripts correctly
- [ ] Successfully executes scripts in background
- [ ] Provides real-time execution status
- [ ] Captures and returns script output/logs
- [ ] Handles script failures gracefully

---

### TASK-H2: Share Management
**Status**: ❌ Not Started  
**Estimated Time**: 3-4 days  
**HA Compatibility**: ⚠️ Useful for HA monitoring but not critical

#### Task Definition
Implement comprehensive Unraid share management operations for monitoring and controlling user shares.

#### API Endpoints
```bash
GET  /api/v1/shares               # List all shares
GET  /api/v1/shares/{name}        # Get share details
POST /api/v1/shares               # Create new share
PUT  /api/v1/shares/{name}        # Update share settings
DELETE /api/v1/shares/{name}      # Delete share
GET  /api/v1/shares/{name}/usage  # Get share usage statistics
```

#### Dependencies
- Share configuration file parsing
- SMB/NFS service management
- File system usage calculation

#### Testing Criteria
- [ ] Lists all configured shares
- [ ] Returns accurate usage statistics
- [ ] Share creation/modification works correctly
- [ ] Proper validation of share parameters

---

### TASK-H3: Enhanced Docker Operations
**Status**: ❌ Not Started  
**Estimated Time**: 1-2 days  
**HA Compatibility**: ✅ Required for complete HA Docker control

#### Task Definition
Extend existing Docker operations to support all container lifecycle management operations expected by Home Assistant integration.

#### API Endpoints
```bash
POST /api/v1/docker/container/{id}/pause    # Pause container
POST /api/v1/docker/container/{id}/resume   # Resume paused container
POST /api/v1/docker/container/{id}/restart  # Restart container
GET  /api/v1/docker/networks                # List Docker networks
GET  /api/v1/docker/container/{id}/logs     # Get container logs
```

#### Dependencies
- Existing Docker plugin
- Docker API client enhancements

#### Testing Criteria
- [ ] All container operations work correctly
- [ ] Container state changes are reflected immediately
- [ ] Log retrieval works for all containers
- [ ] Network information is accurate

## 🔧 MEDIUM PRIORITY TASKS

### TASK-M1: Enhanced VM Operations
**Status**: ❌ Not Started  
**Estimated Time**: 2-3 days  
**HA Compatibility**: ✅ Useful for complete HA VM control

#### Task Definition
Extend VM management capabilities to support advanced virtual machine operations.

#### API Endpoints
```bash
POST /api/v1/vm/{id}/pause        # Pause VM
POST /api/v1/vm/{id}/resume       # Resume paused VM
POST /api/v1/vm/{id}/restart      # Restart VM
POST /api/v1/vm/{id}/hibernate    # Hibernate VM
POST /api/v1/vm/{id}/force-stop   # Force stop VM
GET  /api/v1/vm/{id}/console      # Get VM console access
```

#### Dependencies
- Existing VM plugin
- libvirt/virsh command interface

#### Testing Criteria
- [ ] All VM operations work correctly
- [ ] VM state changes are reflected properly
- [ ] Operations handle errors gracefully

---

### TASK-M2: Notification System
**Status**: ❌ Not Started  
**Estimated Time**: 2-3 days  
**HA Compatibility**: ⚠️ Useful for HA notification integration

#### Task Definition
Implement notification management system for monitoring and managing Unraid system notifications.

#### API Endpoints
```bash
GET  /api/v1/notifications        # List notifications
POST /api/v1/notifications        # Create notification
PUT  /api/v1/notifications/{id}   # Update notification
DELETE /api/v1/notifications/{id} # Delete notification
POST /api/v1/notifications/clear  # Clear all notifications
```

#### Dependencies
- Unraid notification system integration
- Event monitoring capabilities

#### Testing Criteria
- [ ] Lists all system notifications
- [ ] Creates custom notifications successfully
- [ ] Notification management operations work correctly

---

### TASK-M3: Command Execution
**Status**: ❌ Not Started  
**Estimated Time**: 1-2 days  
**HA Compatibility**: ✅ Required for HA command execution services

#### Task Definition
Implement secure command execution capabilities for running arbitrary commands and container operations.

#### API Endpoints
```bash
POST /api/v1/execute/command      # Execute shell command
POST /api/v1/execute/container    # Execute command in container
```

#### Dependencies
- Secure command execution framework
- Container runtime access

#### Testing Criteria
- [ ] Commands execute safely with proper sandboxing
- [ ] Container command execution works correctly
- [ ] Proper error handling and output capture

---

## Implementation Priority Order

1. **TASK-C1**: Complete Home Assistant Integration Data Format
2. **TASK-C2**: Array Control Operations  
3. **TASK-C3**: System Power Management
4. **TASK-H1**: User Script Management
5. **TASK-H2**: Share Management
6. **TASK-H3**: Enhanced Docker Operations
7. **TASK-M1**: Enhanced VM Operations
8. **TASK-M2**: Notification System
9. **TASK-M3**: Command Execution

## Testing Strategy

Each task must be:
1. **Implemented** according to specification
2. **Tested** on real Unraid environment (192.168.20.21)
3. **Validated** for 100% working status
4. **Documented** with updated API specifications

No task should be considered complete until it passes all testing criteria and works correctly in the production Unraid environment.
