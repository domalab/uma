# UMA Testing Guide

## Prerequisites

### Development Environment
- Go 1.19+ toolchain
- Access to Unraid test server (192.168.20.21)
- SSH access to test server

### Test Server Details
- **IP**: 192.168.20.21
- **Credentials**: root/tasvyh-4Gehju-ridxic
- **Unraid Version**: 6.12.24-Unraid
- **Array**: 5 disks (16TB parity + 4 data disks)
- **Docker**: 14 containers running
- **VMs**: 1 VM (Bastion) running

## Build Testing

### 1. Compilation Test
```bash
# Clean previous builds
make clean

# Build for local testing
make local

# Verify binary creation
ls -la uma
./uma --version
```

### 2. Configuration Test
```bash
# Test configuration loading
./uma config show

# Test configuration validation
./uma config validate
```

### 3. Plugin Loading Test
```bash
# Test plugin initialization
./uma daemon --dry-run

# Check plugin registration
./uma plugins list
```

## API Testing

### 1. Start the Daemon
```bash
# Start daemon in foreground for testing
./uma daemon --config ./config.yaml --log-level debug

# Or start in background
./uma daemon --config ./config.yaml &
```

### 2. Health Check
```bash
# Basic health check
curl -s http://localhost:8080/api/v1/health | jq

# Expected response:
# {
#   "status": "ok",
#   "timestamp": "2025-01-XX...",
#   "version": "2.0.0"
# }
```

### 3. System Monitoring Tests
```bash
# System resources
curl -s http://localhost:8080/api/v1/system/resources | jq

# System information
curl -s http://localhost:8080/api/v1/system/info | jq

# System logs
curl -s http://localhost:8080/api/v1/system/logs | jq
```

### 4. Storage Monitoring Tests
```bash
# Array information
curl -s http://localhost:8080/api/v1/storage/array | jq

# Cache information
curl -s http://localhost:8080/api/v1/storage/cache | jq

# Boot disk information
curl -s http://localhost:8080/api/v1/storage/boot | jq
```

### 5. GPU Monitoring Tests
```bash
# GPU information
curl -s http://localhost:8080/api/v1/gpu | jq

# Expected: Intel UHD Graphics 630 information
```

### 6. Docker Management Tests
```bash
# List all containers
curl -s http://localhost:8080/api/v1/docker/containers | jq

# List all containers (including stopped)
curl -s http://localhost:8080/api/v1/docker/containers?all=true | jq

# Get specific container info (use actual container name from test server)
curl -s http://localhost:8080/api/v1/docker/container/plex | jq

# Get container logs
curl -s http://localhost:8080/api/v1/docker/container/plex/logs?lines=50 | jq

# Get container stats
curl -s http://localhost:8080/api/v1/docker/container/plex/stats | jq

# Container operations (BE CAREFUL - these affect running containers)
curl -X POST http://localhost:8080/api/v1/docker/container/test-container/restart
curl -X POST http://localhost:8080/api/v1/docker/container/test-container/pause
curl -X POST http://localhost:8080/api/v1/docker/container/test-container/unpause
```

### 7. VM Management Tests
```bash
# List VMs
curl -s http://localhost:8080/api/v1/vm/list | jq

# Get VM information (Bastion VM is running on test server)
curl -s http://localhost:8080/api/v1/vm/Bastion | jq

# Get VM stats
curl -s http://localhost:8080/api/v1/vm/Bastion/stats | jq

# Get VM console info
curl -s http://localhost:8080/api/v1/vm/Bastion/console | jq

# VM operations (BE CAREFUL - these affect running VMs)
# curl -X POST http://localhost:8080/api/v1/vm/Bastion/pause
# curl -X POST http://localhost:8080/api/v1/vm/Bastion/resume
```

### 8. Diagnostics Tests
```bash
# Health checks
curl -s http://localhost:8080/api/v1/diagnostics/health | jq

# Diagnostic information
curl -s http://localhost:8080/api/v1/diagnostics/info | jq

# Available repairs
curl -s http://localhost:8080/api/v1/diagnostics/repair | jq

# Execute repair (BE CAREFUL)
# curl -X POST http://localhost:8080/api/v1/diagnostics/repair?action=clear_logs
```

## Integration Testing

### 1. Home Assistant Integration Test
Create a test configuration file `ha_test.yaml`:

```yaml
sensor:
  - platform: rest
    resource: "http://192.168.20.21:8080/api/v1/system/resources"
    name: "Unraid CPU Usage"
    value_template: "{{ value_json.cpu.usage }}"
    unit_of_measurement: "%"
    
  - platform: rest
    resource: "http://192.168.20.21:8080/api/v1/storage/array"
    name: "Unraid Array State"
    value_template: "{{ value_json.state }}"
    
  - platform: rest
    resource: "http://192.168.20.21:8080/api/v1/docker/containers"
    name: "Unraid Container Count"
    value_template: "{{ value_json | length }}"

switch:
  - platform: rest
    resource: "http://192.168.20.21:8080/api/v1/docker/container/plex"
    name: "Plex Container"
    body_on: '{"action": "start"}'
    body_off: '{"action": "stop"}'
    is_on_template: "{{ value_json.state == 'running' }}"
```

### 2. Performance Testing
```bash
# Load testing with multiple concurrent requests
for i in {1..10}; do
  curl -s http://localhost:8080/api/v1/system/resources &
done
wait

# Memory usage monitoring
ps aux | grep uma

# Response time testing
time curl -s http://localhost:8080/api/v1/storage/array > /dev/null
```

## Validation Checklist

### ✅ Build Validation
- [ ] Code compiles without errors
- [ ] All plugins load successfully
- [ ] Configuration validates correctly
- [ ] Binary starts without crashes

### ✅ API Validation
- [ ] Health endpoint responds correctly
- [ ] All system endpoints return valid JSON
- [ ] Storage endpoints return Unraid-specific data
- [ ] Docker endpoints interact with real containers
- [ ] VM endpoints work with running VMs
- [ ] Diagnostics endpoints provide useful information

### ✅ Data Validation
- [ ] Array state correctly reflects Unraid status
- [ ] Disk information matches actual hardware
- [ ] Container data matches `docker ps` output
- [ ] VM data matches `virsh list` output
- [ ] Temperature readings are reasonable
- [ ] Usage statistics are accurate

### ✅ Error Handling
- [ ] Invalid endpoints return proper HTTP errors
- [ ] Missing resources return 404
- [ ] Invalid operations return appropriate errors
- [ ] Service unavailable scenarios handled gracefully

### ✅ Security Validation
- [ ] Authentication works correctly
- [ ] Rate limiting functions properly
- [ ] CORS headers are set appropriately
- [ ] No sensitive information leaked in responses

## Troubleshooting

### Common Issues

1. **Permission Errors**
   ```bash
   # Ensure proper permissions for system files
   sudo chown root:root uma
   sudo chmod +x uma
   ```

2. **Port Conflicts**
   ```bash
   # Check if port 8080 is in use
   netstat -tlnp | grep 8080

   # Use different port if needed
   ./uma daemon --port 8081
   ```

3. **Plugin Loading Errors**
   ```bash
   # Check plugin directory permissions
   ls -la daemon/plugins/
   
   # Verify all imports are correct
   go mod tidy
   ```

4. **Docker API Errors**
   ```bash
   # Verify Docker socket permissions
   ls -la /var/run/docker.sock
   
   # Test Docker API directly
   curl --unix-socket /var/run/docker.sock http://localhost/containers/json
   ```

5. **Libvirt Connection Issues**
   ```bash
   # Check libvirt service
   systemctl status libvirtd
   
   # Test virsh connection
   virsh list --all
   ```

### Debug Mode
```bash
# Run with maximum debugging
./uma daemon --log-level trace --debug

# Enable API request logging
./uma daemon --log-requests
```

### Log Analysis
```bash
# Monitor logs in real-time
tail -f /var/log/uma.log

# Search for specific errors
grep -i error /var/log/uma.log
```

## Expected Test Results

### System Resources Response
```json
{
  "cpu": {
    "usage": 15.2,
    "cores": 6,
    "temperature": 29
  },
  "memory": {
    "total": 17179869184,
    "used": 8589934592,
    "available": 8589934592,
    "usage_percent": 50.0
  },
  "load": {
    "load1": 0.5,
    "load5": 0.3,
    "load15": 0.2
  }
}
```

### Storage Array Response
```json
{
  "state": "started",
  "total_size": 42949672960000,
  "used_size": 8589934592000,
  "free_size": 34359738368000,
  "usage_percent": 20.0,
  "num_disks": 4,
  "num_parity": 1,
  "disks": [...]
}
```

### Docker Containers Response
```json
[
  {
    "id": "ce08181e7766",
    "name": "jackett",
    "image": "lscr.io/linuxserver/jackett",
    "state": "running",
    "status": "Up 11 hours",
    "ports": ["9117:9117"]
  }
]
```

This testing guide provides comprehensive validation of all UMA Phase 2 features and ensures production readiness.
