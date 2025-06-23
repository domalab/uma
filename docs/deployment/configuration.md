# UMA Configuration Guide

This guide covers all configuration options for UMA (Unraid Management Agent), including plugin settings, environment variables, configuration files, and command-line options.

## Configuration Priority

UMA uses the following configuration priority order (highest to lowest):

1. **Command Line Flags** (highest priority)
2. **Environment Variables** (prefix: `UMA_`)
3. **Configuration Files** (YAML, JSON, INI)
4. **Default Values** (lowest priority)

## Plugin Configuration

### Unraid Web Interface

Configure UMA through the Unraid web interface:

**Settings > Utilities > UMA**
- **Enable/Disable**: Toggle UMA service on/off
- **HTTP Port**: Set API port (1024-65535, default: 34600)
- **Open Web UI**: Direct link to Swagger API documentation

**Plugins > Installed Plugins > UMA**
- Plugin management and version information
- Start/stop service controls
- Configuration file access

### Plugin Settings

The plugin provides these configurable options:

- **Service Enable**: Enable/disable the UMA service
- **HTTP Port**: Configure the API server port
- **Log Level**: Set logging verbosity (debug, info, warn, error)
- **Monitoring Interval**: System monitoring frequency (default: 30s)

## Environment Variables

All configuration options can be set via environment variables with the `UMA_` prefix:

### HTTP Server Configuration

```bash
# HTTP Server
export UMA_HTTP_PORT=34600
export UMA_HTTP_HOST="0.0.0.0"
export UMA_HTTP_TIMEOUT="60s"
export UMA_HTTP_READ_TIMEOUT="30s"
export UMA_HTTP_WRITE_TIMEOUT="30s"
```



### Logging Configuration

```bash
# Logging
export UMA_LOGGING_LEVEL=info
export UMA_LOGGING_FORMAT=console
export UMA_LOGGING_FILE=""
export UMA_LOGGING_MAX_SIZE=100
export UMA_LOGGING_MAX_BACKUPS=3
export UMA_LOGGING_MAX_AGE=28
```

### Monitoring Configuration

```bash
# Monitoring
export UMA_MONITORING_INTERVAL=30s
export UMA_MONITORING_DOCKER_ENABLED=true
export UMA_MONITORING_STORAGE_ENABLED=true
export UMA_MONITORING_SYSTEM_ENABLED=true
export UMA_MONITORING_UPS_ENABLED=true
export UMA_MONITORING_GPU_ENABLED=true
```

### Metrics and WebSocket Configuration

```bash
# Metrics
export UMA_METRICS_ENABLED=true
export UMA_METRICS_PATH="/metrics"

# WebSocket
export UMA_WEBSOCKET_ENABLED=true
export UMA_WEBSOCKET_MAX_CONNECTIONS=100
export UMA_WEBSOCKET_PING_INTERVAL="30s"
export UMA_WEBSOCKET_PONG_TIMEOUT="60s"
```

### Cache and Rate Limiting

```bash
# Cache
export UMA_CACHE_ENABLED=true
export UMA_CACHE_TTL="5m"
export UMA_CACHE_CLEANUP_INTERVAL="10m"

# Rate Limiting
export UMA_RATE_LIMIT_ENABLED=false
export UMA_RATE_LIMIT_REQUESTS_PER_MINUTE=60
export UMA_RATE_LIMIT_BURST=10
```

### CORS Configuration

```bash
# CORS
export UMA_CORS_ENABLED=true
export UMA_CORS_ALLOWED_ORIGINS="*"
export UMA_CORS_ALLOWED_METHODS="GET,POST,PUT,DELETE,OPTIONS"
export UMA_CORS_ALLOWED_HEADERS="*"
```

## Configuration Files

UMA supports configuration files in YAML, JSON, or INI format. Place the configuration file in the working directory or specify the path using the `--config` flag.

### YAML Configuration Example

Create `uma.yaml` in the working directory:

```yaml
# UMA Configuration File
# Complete example with all available options

# HTTP Server Configuration
http:
  port: 34600
  host: "0.0.0.0"
  timeout: "60s"
  read_timeout: "30s"
  write_timeout: "30s"



# Logging Configuration
logging:
  level: "info"
  format: "console"
  file: ""
  max_size: 100
  max_backups: 3
  max_age: 28

# Metrics Configuration
metrics:
  enabled: true
  path: "/metrics"

# WebSocket Configuration
websocket:
  enabled: true
  max_connections: 100
  ping_interval: "30s"
  pong_timeout: "60s"

# Monitoring Configuration
monitoring:
  interval: "30s"
  docker:
    enabled: true
  storage:
    enabled: true
  system:
    enabled: true
  ups:
    enabled: true  # Auto-detects APC/NUT daemons
  gpu:
    enabled: true

# Cache Configuration
cache:
  enabled: true
  ttl: "5m"
  cleanup_interval: "10m"

# Rate Limiting Configuration
rate_limit:
  enabled: false
  requests_per_minute: 60
  burst: 10

# CORS Configuration
cors:
  enabled: true
  allowed_origins: ["*"]
  allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
  allowed_headers: ["*"]
```

### JSON Configuration Example

```json
{
  "http": {
    "port": 34600,
    "host": "0.0.0.0",
    "timeout": "60s"
  },
  "logging": {
    "level": "info",
    "format": "console"
  },
  "metrics": {
    "enabled": true,
    "path": "/metrics"
  },
  "monitoring": {
    "interval": "30s",
    "docker": {"enabled": true},
    "storage": {"enabled": true},
    "system": {"enabled": true},
    "ups": {"enabled": true},
    "gpu": {"enabled": true}
  }
}
```

## Command Line Options

UMA supports various command-line flags for advanced usage:

```bash
# Basic usage
uma boot --http-port 34600

# Configuration management
uma config show                    # Show current configuration
uma config set --port 8080        # Set HTTP port
uma config set --log-level debug  # Set log level
uma config generate --api-key     # Generate new API key

# Help and version
uma --help                        # Show help
uma --version                     # Show version information
```

### Available Flags

- `--http-port`: HTTP server port (default: 34600)
- `--logs-dir`: Directory for log files (default: /var/log)
- `--config-path`: Path to configuration file
- `--help`: Show help information
- `--version`: Show version information

## Hardware Auto-Detection

UMA automatically detects and configures monitoring for available hardware:

### UPS Detection
- **apcupsd daemon**: Automatic detection and monitoring
- **NUT (Network UPS Tools)**: Support for various UPS models
- **Fallback**: Graceful handling when UPS is unavailable

### Sensor Detection
- **IPMI**: Hardware sensor monitoring via IPMI interface
- **System Temperature**: Unraid system temperature plugin integration
- **Network Sensors**: Remote sensor monitoring capabilities

### GPU Detection
- **NVIDIA**: GPU monitoring via nvidia-smi
- **AMD**: Basic GPU detection and monitoring
- **Intel**: Integrated graphics monitoring

## Security Considerations

UMA is designed for internal network use with the following security model:

- **No Authentication**: Disabled by default for trusted network environments
- **Network Security**: Security handled at network/firewall level
- **Internal API**: Designed for local network access only
- **CORS**: Configurable for web application integration

## Troubleshooting

### Common Configuration Issues

1. **Port Conflicts**: Ensure port 34600 is not used by other services
2. **Permission Issues**: Verify UMA has access to system resources
3. **Hardware Detection**: Check system logs for sensor/UPS detection issues
4. **Configuration Syntax**: Validate YAML/JSON syntax in configuration files

### Log Analysis

Enable debug logging for troubleshooting:

```bash
export UMA_LOGGING_LEVEL=debug
```

Check logs at `/var/log/uma.log` for detailed information.

### Configuration Validation

Use the configuration management commands to validate settings:

```bash
uma config show    # Display current configuration
uma config set --help  # Show available configuration options
```

## Performance Tuning

### Monitoring Intervals

Adjust monitoring frequency based on your needs:

```yaml
monitoring:
  interval: "15s"  # More frequent updates
  # or
  interval: "60s"  # Less frequent for lower resource usage
```

### Cache Settings

Optimize cache settings for your environment:

```yaml
cache:
  enabled: true
  ttl: "2m"        # Shorter TTL for more current data
  cleanup_interval: "5m"
```

### WebSocket Connections

Limit WebSocket connections based on client needs:

```yaml
websocket:
  max_connections: 50  # Reduce for lower resource usage
  ping_interval: "60s" # Longer intervals for stable connections
```

## Integration Examples

### Home Assistant

Configure UMA for Home Assistant integration:

```yaml
# Optimized for Home Assistant
monitoring:
  interval: "30s"
metrics:
  enabled: true
websocket:
  enabled: true
  max_connections: 10
```

### Prometheus

Configure UMA for Prometheus monitoring:

```yaml
# Optimized for Prometheus
metrics:
  enabled: true
  path: "/metrics"
monitoring:
  interval: "15s"  # Frequent updates for metrics
```

### Custom Applications

For custom integrations, consider:

- Using WebSocket endpoints for real-time data
- Implementing proper error handling for API calls
- Respecting rate limits and monitoring intervals
- Using request IDs for debugging and tracing
