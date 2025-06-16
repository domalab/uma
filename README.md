# UMA (Unraid Management Agent)

_tl;dr_ **UMA** is a comprehensive monitoring and control system for Unraid servers, providing real-time system insights, hardware monitoring, and remote management capabilities with enterprise-grade observability.

## Features

### Core Monitoring & Management

- **System Monitoring**: CPU, RAM, storage, and boot disk status with real-time metrics
- **Hardware Sensors**: Temperature monitoring, fan speeds, and system health via IPMI
- **Storage Management**: Array disk health, cache disk status, SMART data, and usage statistics
- **Power Management**: UPS status and metrics (NUT integration)
- **GPU Monitoring**: GPU usage, temperature, and status
- **Docker Management**: Container control with bulk operations (start, stop, restart multiple containers)
- **VM Control**: Virtual machine management and status monitoring

### Advanced API & Integration

- **REST API**: Full OpenAPI 3.1.1 documented API with pagination, compression, and versioning
- **WebSocket Support**: Real-time monitoring endpoints for system stats, Docker events, and storage status
- **Bulk Operations**: Efficient management of multiple Docker containers simultaneously
- **Prometheus Metrics**: Comprehensive metrics collection for monitoring and alerting
- **Structured Logging**: Production-grade logging with contextual fields and JSON output

### Security & Production Features

- **JWT Authentication**: Role-based access control with API key and JWT token support
- **Enhanced Routing**: Chi router with organized route groups and advanced middleware
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

UMA provides a comprehensive REST API with OpenAPI 3.1.1 specification for integration with Home Assistant, Prometheus, and other automation systems.

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

#### Authentication Management (Phase 3)
- `POST /api/v1/auth/login` - Authenticate with API key and get JWT token
- `GET /api/v1/auth/users` - List all users (Admin only)
- `POST /api/v1/auth/users` - Create new user (Admin only)
- `GET /api/v1/auth/users/{id}` - Get user details
- `PUT /api/v1/auth/users/{id}` - Update user (Admin only)
- `DELETE /api/v1/auth/users/{id}` - Delete user (Admin only)
- `POST /api/v1/auth/users/{id}/regenerate-key` - Regenerate API key
- `GET /api/v1/auth/stats` - Authentication statistics

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

# Build optimized binary (14.4MB)
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

# Authentication
export UMA_AUTH_ENABLED=false
export UMA_AUTH_JWT_SECRET="your-secret-key"

# Logging
export UMA_LOGGING_LEVEL=info
export UMA_LOGGING_FORMAT=console

# Monitoring
export UMA_MONITORING_INTERVAL=30s
```

#### Configuration File Example

Create `uma.yaml` in the working directory:

```yaml
# UMA Configuration
http:
  port: 34600
  host: "0.0.0.0"
  timeout: "60s"

auth:
  enabled: false
  jwt_secret: "your-secret-key"
  token_expiry: "24h"

logging:
  level: "info"
  format: "console"

metrics:
  enabled: true
  path: "/metrics"

monitoring:
  interval: "30s"
  docker:
    enabled: true
  storage:
    enabled: true
```

## Recent Updates

### Phase 3: Enhanced Features (Latest)

- ✅ **JWT Authentication**: Role-based access control with Admin, Operator, and Viewer roles
- ✅ **Enhanced Routing**: Chi router with organized route groups and improved performance
- ✅ **Configuration Management**: Viper integration with hot reload and environment variables
- ✅ **Optimized Binary**: 14.4MB optimized binary (28% under 20MB target) with all features
- ✅ **Authentication API**: Complete user management with API key generation and JWT tokens
- ✅ **Backward Compatibility**: 100% API compatibility maintained with existing integrations

### Phase 2: Production Readiness

- ✅ **Structured Logging**: Implemented Zerolog with contextual fields and JSON output
- ✅ **Prometheus Metrics**: Comprehensive metrics collection for monitoring and alerting
- ✅ **Testing Framework**: Added Testify with comprehensive unit tests and benchmarks
- ✅ **Enhanced Observability**: Full production-grade monitoring and logging stack
- ✅ **Performance Tracking**: Request duration, success rates, and operation metrics

### Phase 1: Security Updates

- ✅ **Dependency Security**: Replaced 11-year-old INI library with modern alternative
- ✅ **Enhanced Validation**: Comprehensive input validation with user-friendly errors
- ✅ **Updated Dependencies**: Kong CLI framework updated to latest stable version
- ✅ **Backward Compatibility**: All existing functionality preserved during updates

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
- [Chi](https://github.com/go-chi/chi) - HTTP router and middleware (Phase 3)
- [Viper](https://github.com/spf13/viper) - Configuration management (Phase 3)
- [JWT-Go](https://github.com/golang-jwt/jwt) - JWT authentication (Phase 3)
- [Validator](https://github.com/go-playground/validator) - Input validation
- [Zerolog](https://github.com/rs/zerolog) - Structured logging
- [Prometheus](https://github.com/prometheus/client_golang) - Metrics collection
- [Testify](https://github.com/stretchr/testify) - Testing framework
- [Kong](https://github.com/alecthomas/kong) - CLI framework
- [Gorilla WebSocket](https://github.com/gorilla/websocket) - WebSocket implementation
- [All dependencies](./go.mod) - Complete dependency list
