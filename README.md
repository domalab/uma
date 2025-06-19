# UMA (Unraid Management Agent)

**UMA** is a comprehensive REST API service for Unraid server management, providing real-time monitoring, hardware control, and automation capabilities for Home Assistant, Prometheus, and other integration platforms.

## Key Features

- **System Monitoring**: Real-time CPU, memory, storage, and temperature monitoring
- **Docker Management**: Complete container lifecycle control with bulk operations
- **VM Control**: Virtual machine management and status monitoring
- **Storage Operations**: Array management, SMART data, and parity operations
- **Hardware Sensors**: Temperature, fan speeds, and UPS monitoring via IPMI/apcupsd
- **WebSocket Streaming**: Real-time event streams for live monitoring
- **REST API**: OpenAPI 3.0.3 documented API with comprehensive endpoint coverage
- **Prometheus Integration**: Native metrics export for monitoring and alerting

## Quick Start

### Installation

**Option 1: Community Applications**
1. Go to the Apps tab
2. Click on the Plugins button
3. Search for "UMA"
4. Click Install

**Option 2: Manual Plugin Installation**
1. Go to the Plugins tab
2. Click on Install Plugin
3. Paste: `https://github.com/domalab/uma/releases/latest/download/uma.plg`
4. Click Install

### Configuration

After installation, configure UMA via:
- **Settings > Utilities > UMA** - Main configuration interface
- **Plugins > Installed Plugins > UMA** - Plugin management

### API Access

Once installed, the REST API is available at:
```
http://your-unraid-ip:34600/api/v1
```

**Interactive Documentation**: `http://your-unraid-ip:34600/api/v1/docs`

**Interactive Documentation**: `http://your-unraid-ip:34600/api/v1/docs`

## API Documentation

UMA provides a comprehensive REST API with OpenAPI 3.0.3 specification for integration with Home Assistant, Prometheus, and other automation platforms.

### Key Endpoints

- **System Monitoring**: `/api/v1/health`, `/api/v1/system/*` - Health checks and system metrics
- **Storage Management**: `/api/v1/storage/*` - Array, cache, and disk information
- **Docker Control**: `/api/v1/docker/*` - Container management and bulk operations
- **VM Management**: `/api/v1/vm/*` - Virtual machine control and monitoring
- **WebSocket Streams**: `/api/v1/ws/*` - Real-time monitoring endpoints
- **Prometheus Metrics**: `/metrics` - Metrics export for monitoring systems

### Documentation Resources

- **[Complete API Reference](docs/api/)** - Detailed endpoint documentation
- **[WebSocket Guide](docs/api/websockets.md)** - Real-time monitoring setup
- **[OpenAPI Guide](docs/api/openapi-guide.md)** - API specification usage

## Configuration

UMA supports flexible configuration via:

- **Plugin Settings**: Configure via Unraid web interface (Settings > Utilities > UMA)
- **Environment Variables**: Use `UMA_` prefix (e.g., `UMA_HTTP_PORT=34600`)
- **Configuration Files**: YAML, JSON, or INI format
- **Command Line**: Flags for advanced usage

### Key Settings

- **HTTP Port**: Default 34600 (configurable 1024-65535)
- **Monitoring**: Auto-detection of UPS, GPU, and sensor hardware
- **WebSocket**: Real-time streaming (enabled by default)
- **Metrics**: Prometheus export (enabled by default)

For detailed configuration options, see **[Configuration Guide](docs/deployment/configuration.md)**

## Documentation

Comprehensive documentation is available in the `docs/` folder:

- **[API Documentation](docs/api/)** - Complete API reference and usage guides
- **[Development Guide](docs/development/)** - Developer documentation and contribution guidelines
- **[Deployment Guide](docs/deployment/)** - Installation and configuration instructions

## Contributing

We welcome contributions! Please see our **[Development Guide](docs/development/)** for details on:

- Setting up the development environment
- Running tests and quality checks
- Code style and contribution guidelines
- Building and deployment procedures

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
