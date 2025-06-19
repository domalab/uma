# UMA Troubleshooting Guide

This guide helps you diagnose and resolve common issues with UMA (Unraid Management Agent).

## Quick Diagnostics

### 1. Service Status Check

```bash
# Check if UMA is running
ps aux | grep uma

# Check port availability
netstat -tlnp | grep 34600

# Test basic connectivity
curl -I http://localhost:34600/api/v1/health
```

### 2. Log Analysis

```bash
# View recent logs
tail -f /var/log/uma.log

# Search for errors
grep -i error /var/log/uma.log | tail -20

# Check startup logs
grep "starting uma" /var/log/uma.log
```

### 3. Configuration Validation

```bash
# Check configuration
uma config show

# Validate API endpoints
curl -s http://localhost:34600/api/v1/health | jq '.'
```

## Common Issues

### UMA Service Won't Start

**Symptoms:**
- Plugin shows as installed but not running
- API endpoints return connection refused
- No process visible in `ps aux | grep uma`

**Causes and Solutions:**

1. **Port Conflict**
   ```bash
   # Check what's using port 34600
   netstat -tlnp | grep 34600
   lsof -i :34600
   
   # Solution: Change port in plugin settings or stop conflicting service
   ```

2. **Permission Issues**
   ```bash
   # Check file permissions
   ls -la /usr/local/emhttp/plugins/uma/
   
   # Fix permissions if needed
   chmod +x /usr/local/emhttp/plugins/uma/scripts/start
   ```

3. **Missing Dependencies**
   ```bash
   # Check for required binaries
   which docker
   which virsh
   
   # Verify Docker is running
   systemctl status docker
   ```

4. **Configuration Errors**
   ```bash
   # Check for syntax errors in config
   uma config show
   
   # Reset to defaults
   rm -f /boot/config/plugins/uma/uma.env
   ```

### API Returns 500 Internal Server Error

**Symptoms:**
- Health check fails with 500 status
- Specific endpoints return server errors
- Logs show panic or error messages

**Debugging Steps:**

1. **Check Logs for Stack Traces**
   ```bash
   grep -A 10 -B 5 "panic\|fatal\|error" /var/log/uma.log
   ```

2. **Test Individual Components**
   ```bash
   # Test Docker connectivity
   docker ps
   
   # Test storage access
   ls -la /mnt/disk*
   
   # Test VM management
   virsh list --all
   ```

3. **Restart with Debug Logging**
   ```bash
   # Stop UMA
   killall uma
   
   # Start with debug logging
   UMA_LOGGING_LEVEL=debug uma boot --http-port 34600
   ```

### WebSocket Connections Fail

**Symptoms:**
- WebSocket endpoints return 404 or connection errors
- Real-time monitoring doesn't work
- Browser console shows WebSocket errors

**Solutions:**

1. **Verify WebSocket Support**
   ```bash
   # Check if WebSocket endpoints are registered
   curl -I http://localhost:34600/api/v1/ws/system/stats
   ```

2. **Test with websocat**
   ```bash
   # Install websocat for testing
   websocat ws://localhost:34600/api/v1/ws/system/stats
   ```

3. **Check Proxy Configuration**
   - Ensure reverse proxies support WebSocket upgrades
   - Verify no firewall blocking WebSocket traffic

### Missing or Incorrect Data

**Symptoms:**
- API returns empty arrays or null values
- System stats show zeros or placeholder data
- Storage information is incomplete

**Diagnostic Steps:**

1. **Verify Data Sources**
   ```bash
   # Check Docker daemon
   docker info
   
   # Check storage mounts
   df -h
   mount | grep /mnt/
   
   # Check system sensors
   sensors
   ```

2. **Test Direct Access**
   ```bash
   # Test Docker API directly
   curl --unix-socket /var/run/docker.sock http://localhost/containers/json
   
   # Check storage files
   ls -la /proc/mdstat
   cat /proc/meminfo
   ```

3. **Check Permissions**
   ```bash
   # Verify UMA can access required files
   ls -la /var/run/docker.sock
   ls -la /proc/mdstat
   ```

### High CPU or Memory Usage

**Symptoms:**
- UMA process consuming excessive resources
- System becomes slow when UMA is running
- Memory usage continuously increasing

**Solutions:**

1. **Identify Resource Usage**
   ```bash
   # Monitor UMA process
   top -p $(pidof uma)
   
   # Check memory usage
   ps aux | grep uma
   ```

2. **Reduce Monitoring Frequency**
   ```bash
   # Increase monitoring intervals
   export UMA_MONITORING_INTERVAL=60s
   ```

3. **Disable Unnecessary Features**
   ```bash
   # Disable WebSocket if not needed
   export UMA_WEBSOCKET_ENABLED=false
   
   # Reduce log level
   export UMA_LOGGING_LEVEL=warn
   ```

### Plugin Installation Issues

**Symptoms:**
- Plugin fails to install from Community Applications
- Manual installation returns errors
- Plugin appears corrupted or incomplete

**Solutions:**

1. **Clear Plugin Cache**
   ```bash
   # Remove old plugin files
   rm -rf /usr/local/emhttp/plugins/uma/
   rm -f /boot/config/plugins/uma/uma*.tgz
   ```

2. **Manual Installation**
   ```bash
   # Download plugin directly
   wget https://github.com/domalab/uma/releases/latest/download/uma.plg
   
   # Install manually
   plugin install uma.plg
   ```

3. **Check Network Connectivity**
   ```bash
   # Test GitHub access
   curl -I https://github.com/domalab/uma/releases/latest
   
   # Check DNS resolution
   nslookup github.com
   ```

## Performance Optimization

### Reduce Resource Usage

1. **Optimize Monitoring Intervals**
   ```bash
   # Increase intervals for less frequent updates
   export UMA_MONITORING_INTERVAL=60s
   export UMA_CACHE_TTL=10m
   ```

2. **Limit WebSocket Connections**
   ```bash
   # Reduce maximum connections
   export UMA_WEBSOCKET_MAX_CONNECTIONS=10
   ```

3. **Adjust Log Settings**
   ```bash
   # Reduce log verbosity
   export UMA_LOGGING_LEVEL=warn
   export UMA_LOGGING_MAX_SIZE=10
   ```

### Optimize for Large Systems

1. **Pagination Settings**
   ```bash
   # Use smaller page sizes for large container lists
   curl "http://localhost:34600/api/v1/docker/containers?limit=10"
   ```

2. **Selective Monitoring**
   ```bash
   # Disable unused monitoring features
   export UMA_MONITORING_GPU_ENABLED=false
   export UMA_MONITORING_UPS_ENABLED=false
   ```

## Network Issues

### Firewall Configuration

1. **Check iptables Rules**
   ```bash
   # View current rules
   iptables -L -n
   
   # Allow UMA port
   iptables -A INPUT -p tcp --dport 34600 -j ACCEPT
   ```

2. **Test from Different Networks**
   ```bash
   # Test from local network
   curl http://192.168.1.100:34600/api/v1/health
   
   # Test from external network (if needed)
   curl http://external-ip:34600/api/v1/health
   ```

### Reverse Proxy Issues

**Common Nginx Configuration:**
```nginx
location /api/ {
    proxy_pass http://unraid-server:34600/api/;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}
```

**Common Apache Configuration:**
```apache
ProxyPass /api/ http://unraid-server:34600/api/
ProxyPassReverse /api/ http://unraid-server:34600/api/
ProxyPreserveHost On
```

## Data Integrity Issues

### Storage Monitoring Problems

1. **SMART Data Missing**
   ```bash
   # Check smartctl availability
   which smartctl
   
   # Test SMART access
   smartctl -a /dev/sda
   ```

2. **Array Status Incorrect**
   ```bash
   # Check mdstat directly
   cat /proc/mdstat
   
   # Verify array status
   mdcmd status
   ```

### Docker Integration Issues

1. **Container Information Missing**
   ```bash
   # Test Docker socket access
   docker ps
   
   # Check socket permissions
   ls -la /var/run/docker.sock
   ```

2. **Container Operations Fail**
   ```bash
   # Test container control
   docker start test-container
   docker stop test-container
   ```

## Advanced Debugging

### Enable Debug Mode

```bash
# Stop UMA
killall uma

# Start with maximum debugging
UMA_LOGGING_LEVEL=debug \
UMA_LOGGING_FORMAT=json \
uma boot --http-port 34600 2>&1 | tee debug.log
```

### Memory Profiling

```bash
# Enable pprof endpoint
UMA_ENABLE_PPROF=true uma boot

# Analyze memory usage
go tool pprof http://localhost:34600/debug/pprof/heap
```

### API Request Tracing

```bash
# Enable request tracing
export UMA_TRACE_REQUESTS=true

# Monitor requests
tail -f /var/log/uma.log | grep "request_id"
```

## Getting Help

### Information to Collect

When reporting issues, include:

1. **System Information**
   ```bash
   # UMA version
   curl -s http://localhost:34600/api/v1/health | jq '.version'
   
   # Unraid version
   cat /etc/unraid-version
   
   # System specs
   uname -a
   free -h
   df -h
   ```

2. **Error Logs**
   ```bash
   # Recent errors
   grep -i error /var/log/uma.log | tail -50
   
   # Startup logs
   grep "starting uma" /var/log/uma.log
   ```

3. **Configuration**
   ```bash
   # Current config
   uma config show
   
   # Environment variables
   env | grep UMA_
   ```

### Support Channels

- **GitHub Issues**: [Report bugs](https://github.com/domalab/uma/issues)
- **GitHub Discussions**: [Ask questions](https://github.com/domalab/uma/discussions)
- **Unraid Forums**: Community support
- **Documentation**: [Complete guides](../README.md)

### Before Reporting

1. **Search existing issues** on GitHub
2. **Try the latest version** of UMA
3. **Test with minimal configuration**
4. **Reproduce the issue** consistently
5. **Collect debug information** as outlined above

## Recovery Procedures

### Complete Reset

```bash
# Stop UMA
killall uma

# Remove all UMA data
rm -rf /usr/local/emhttp/plugins/uma/
rm -rf /var/log/uma/
rm -f /boot/config/plugins/uma/uma.env

# Reinstall plugin
# (Use Community Applications or manual method)
```

### Configuration Reset

```bash
# Reset to default configuration
rm -f /boot/config/plugins/uma/uma.env

# Restart UMA service
/usr/local/emhttp/plugins/uma/scripts/stop
/usr/local/emhttp/plugins/uma/scripts/start
```

### Backup and Restore

```bash
# Backup configuration
cp /boot/config/plugins/uma/uma.env uma-config-backup.env

# Restore configuration
cp uma-config-backup.env /boot/config/plugins/uma/uma.env
```
