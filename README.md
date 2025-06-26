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
- **MCP Server**: Model Context Protocol support with 50+ auto-discovered tools
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

## API Documentation

UMA provides a comprehensive REST API with OpenAPI 3.0.3 specification for integration with Home Assistant, Prometheus, and other automation platforms. **All endpoints return real system data** collected from actual hardware and services.

### 🚀 Enhanced Monitoring Capabilities

#### 🗄️ **Storage Monitoring with Real Usage Calculations**
- **Array Capacity Planning**: Real total capacity (38.2 TB), used space (8.3 TB), usage percentage (21.67%)
- **Individual Disk Metrics**: Per-disk usage, temperature, SMART data, health status
- **Cache & Boot Monitoring**: Complete filesystem usage tracking with real percentages
- **ZFS Pool Support**: Comprehensive ZFS storage monitoring and health checks

#### ⚡ **UPS Power Monitoring with Real Consumption Data**
- **Power Consumption Tracking**: Real watts calculated from UPS load × nominal power
- **Battery Management**: Live charge level (100%), runtime estimates (220 min), voltage monitoring
- **UPS Health Monitoring**: Line voltage, load percentage, operational status
- **Multi-UPS Support**: APC (apcupsd) and NUT (Network UPS Tools) integration

#### 🖥️ **Performance Monitoring Across All Components**
- **Container Performance**: CPU, memory, network I/O metrics for all Docker containers
- **VM Performance**: CPU usage, disk I/O, network stats via libvirt integration
- **GPU Monitoring**: Intel/NVIDIA/AMD GPU utilization, memory, temperatures, power draw
- **Network Interfaces**: Speeds, duplex settings, traffic statistics, connectivity testing

#### 📡 **Real-time Event Streaming & Integration**
- **WebSocket Events**: Live system changes, container events, performance updates
- **MCP Integration**: Model Context Protocol for AI assistant integration
- **Prometheus Export**: Complete metrics for monitoring and alerting systems
- **Home Assistant Ready**: Optimized for home automation platform integration

### Key API Endpoints

- **Enhanced Storage**: `/api/v1/storage/*` - Real capacity calculations and disk metrics
- **UPS Power Monitoring**: `/api/v1/system/ups` - Real power consumption and battery data
- **Container Performance**: `/api/v1/docker/*` - CPU, memory, network metrics
- **VM Monitoring**: `/api/v1/vm/*` - Complete virtual machine performance tracking
- **GPU Metrics**: `/api/v1/system/gpu` - Multi-vendor GPU monitoring
- **Network Monitoring**: `/api/v1/network/*` - Interface speeds, traffic, connectivity
- **Real-time Streams**: `/api/v1/ws/*` - WebSocket event broadcasting
- **System Health**: `/api/v1/health` - Comprehensive dependency monitoring
- **Prometheus Metrics**: `/metrics` - Complete metrics export

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
- **MCP Server**: Model Context Protocol on port 34800 (configurable, disabled by default)
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
