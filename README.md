# UMA (Unraid Management Agent)

_tl;dr_ **UMA** is a comprehensive monitoring and control system for Unraid servers, providing real-time system insights, hardware monitoring, and remote management capabilities with enterprise-grade observability.

## Features

### Core Monitoring & Management

- **System Monitoring**: CPU, RAM, storage, and boot disk status with real-time metrics
- **Hardware Sensors**: Temperature monitoring, fan speeds, and system health via IPMI
- **Storage Management**: Array disk health, cache disk status, SMART data, and usage statistics
- **Power Management**: Automatic UPS detection and monitoring with APC and NUT integration, system power control (reboot/shutdown)
- **GPU Monitoring**: GPU usage, temperature, and status
- **Docker Management**: Complete container control with individual and bulk operations (start, stop, restart, pause, resume)
- **VM Control**: Virtual machine management and status monitoring

### Advanced API & Integration

- **REST API**: Full OpenAPI 3.1.1 documented API with pagination, compression, and versioning
- **WebSocket Support**: Real-time monitoring endpoints for system stats, Docker events, and storage status
- **Bulk Operations**: Efficient management of multiple Docker containers simultaneously
- **Prometheus Metrics**: Comprehensive metrics collection for monitoring and alerting
- **Structured Logging**: Production-grade logging with contextual fields and JSON output

### Security & Production Features

- **Optimized HTTP Mux**: Clean HTTP multiplexer with organized route groups and efficient middleware
- **Configuration Management**: Viper-based configuration with hot reload and environment variables
- **Enhanced Security**: Modern dependency management with security vulnerability elimination
- **Input Validation**: Comprehensive request validation with user-friendly error messages
- **Request Tracking**: Full request lifecycle tracking with unique request IDs
- **Health Checks**: Detailed dependency health monitoring and system diagnostics
- **Performance Monitoring**: Response time tracking and operation duration metrics

## Installation

There are 2 ways to install this application

- Community Applications<br/>
  Go to the Apps tab<br/>
  Click on the Plugins button<br/>
  Look for UMA<br/>
  Click Install

- Plugins Tab (manual)<br/>
  Go to the Plugins tab<br/>
  Click on Install Plugin<br/>
  Paste the following address in the input field: <https://github.com/domalab/uma/releases/latest/download/uma.plg><br/>
  Click Install

## Running the app

After installing the plugin, you can access the web UI, via the following methods:

- Method 1<br/>
  Go to Settings > Utilities<br/>
  Click on UMA<br/>
  Click on Open Web UI<br/>

- Method 2<br/>
  Go to Plugins > Installed Plugins<br/>
  Click on UMA<br/>
  Click on Open Web UI<br/>

## API Access

UMA provides a comprehensive REST API with OpenAPI 3.0.3 specification for integration with Home Assistant, Prometheus, and other automation systems.

### API Documentation

- **Swagger UI**: `http://your-unraid-ip:34600/api/v1/docs` - Interactive API documentation
- **OpenAPI Spec**: `http://your-unraid-ip:34600/api/v1/openapi.json` - Machine-readable API specification
- **Prometheus Metrics**: `http://your-unraid-ip:34600/metrics` - Metrics endpoint for monitoring

### API Features

- **RESTful Design**: Standard HTTP methods with JSON responses
- **Request ID Tracking**: Unique request IDs for debugging and monitoring
- **Response Compression**: Gzip compression for large responses
- **API Versioning**: Accept header-based versioning (`application/vnd.uma.v1+json`)
- **Pagination**: Configurable pagination for large datasets
- **Input Validation**: Comprehensive validation with user-friendly error messages
- **Bulk Operations**: Efficient management of multiple resources

### Core API Endpoints

#### System Monitoring
- `GET /api/v1/health` - System health and dependency status
- `GET /api/v1/system/stats` - CPU, memory, and uptime statistics
- `GET /api/v1/system/temperature` - System temperature sensors
- `GET /api/v1/system/network` - Network interface information
- `GET /api/v1/system/ups` - UPS status and metrics
- `GET /api/v1/system/gpu` - GPU usage and temperature

#### Storage Management
- `GET /api/v1/storage/array` - Array disk information
- `GET /api/v1/storage/cache` - Cache disk status
- `GET /api/v1/storage/boot` - Boot disk information
- `GET /api/v1/storage/disks` - All disk information with pagination

#### Docker Management
- `GET /api/v1/docker/containers` - List all containers with pagination
- `POST /api/v1/docker/containers/bulk/start` - Start multiple containers
- `POST /api/v1/docker/containers/bulk/stop` - Stop multiple containers
- `POST /api/v1/docker/containers/bulk/restart` - Restart multiple containers



### WebSocket Endpoints

Real-time monitoring via WebSocket connections:

- `ws://your-unraid-ip:34600/api/v1/ws/system/stats` - Real-time system statistics
- `ws://your-unraid-ip:34600/api/v1/ws/docker/events` - Docker container events
- `ws://your-unraid-ip:34600/api/v1/ws/storage/status` - Storage status updates

### Prometheus Integration

UMA exposes comprehensive metrics for monitoring and alerting:

```bash
# API request metrics
uma_api_requests_total{endpoint="/api/v1/health",method="GET",status_code="200"}
uma_api_request_duration_seconds{endpoint="/api/v1/health",method="GET"}

# Health check metrics
uma_health_check_status{dependency="docker"}
uma_health_check_duration_seconds

# Bulk operation metrics
uma_bulk_operation_duration_seconds{operation="start"}
uma_bulk_operation_success_rate{operation="start"}

# WebSocket metrics
uma_websocket_connections{endpoint="/api/v1/ws/system/stats"}
uma_websocket_messages_total{endpoint="/api/v1/ws/system/stats",message_type="stats"}
```

## Building from Source

### Prerequisites

- Go 1.21 or later
- Git

### Build Instructions

```bash
# Clone the repository
git clone https://github.com/domalab/uma.git
cd uma

# Install dependencies
go mod tidy

# Build optimized binary (1.9MB)
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o uma .

# Run tests
go test ./...
```

### Configuration

UMA supports multiple configuration methods with the following priority order:

1. **Command Line Flags** (highest priority)
2. **Environment Variables** (prefix: `UMA_`)
3. **Configuration Files** (YAML, JSON, INI)
4. **Default Values** (lowest priority)

#### Environment Variables

```bash
# HTTP Server
export UMA_HTTP_PORT=34600
export UMA_HTTP_HOST="0.0.0.0"
export UMA_HTTP_TIMEOUT="60s"

# Logging
export UMA_LOGGING_LEVEL=info
export UMA_LOGGING_FORMAT=console

# Monitoring
export UMA_MONITORING_INTERVAL=30s
export UMA_MONITORING_UPS_ENABLED=true

# Metrics
export UMA_METRICS_ENABLED=true
```

#### Configuration File Example

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

# Authentication Configuration (disabled by default for internal use)
auth:
  enabled: false
  api_key: ""
  jwt_secret: ""
  token_expiry: "24h"

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

## Documentation

Comprehensive documentation is available in the `docs/` folder:

- **[API Documentation](docs/api/)** - Complete API reference and usage guides
- **[Development Guide](docs/development/)** - Developer documentation and contribution guidelines
- **[Deployment Guide](docs/deployment/)** - Installation and configuration instructions

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Run tests (`go test ./...`)
4. Commit your changes (`git commit -m 'Add amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Credits

- [Go](https://golang.org/) - Programming language
- [Go HTTP Mux](https://pkg.go.dev/net/http#ServeMux) - Optimized HTTP multiplexer with custom middleware
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Validator](https://github.com/go-playground/validator) - Input validation
- [Zerolog](https://github.com/rs/zerolog) - Structured logging
- [Prometheus](https://github.com/prometheus/client_golang) - Metrics collection
- [Testify](https://github.com/stretchr/testify) - Testing framework
- [Kong](https://github.com/alecthomas/kong) - CLI framework
- [Gorilla WebSocket](https://github.com/gorilla/websocket) - WebSocket implementation
- [All dependencies](./go.mod) - Complete dependency list
