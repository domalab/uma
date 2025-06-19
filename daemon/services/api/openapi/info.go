package openapi

// GenerateInfo creates the OpenAPI info section
func GenerateInfo(config *Config) OpenAPIInfo {
	version := config.Version
	if version == "" || version == "unknown" {
		version = "2025.06.17"
	}

	return OpenAPIInfo{
		Title: "UMA REST API",
		Description: `Unraid Management Agent API providing 100% functionality coverage for comprehensive server management.

## Features

### System Monitoring
- **Hardware Monitoring**: CPU usage, RAM utilization, temperatures, fan speeds
- **Real-time Metrics**: Live system statistics with WebSocket updates
- **GPU Monitoring**: Graphics card status and utilization
- **Network Monitoring**: Interface statistics and connectivity status

### Storage Management
- **Array Management**: Unraid array start/stop with proper orchestration
- **Disk Monitoring**: Individual disk health, SMART data, temperatures
- **Cache Management**: Cache disk status and performance metrics
- **ZFS Support**: ZFS pool monitoring and management
- **Parity Operations**: Parity check status and scheduling

### Container & VM Control
- **Docker Management**: Individual and bulk container operations (start/stop/restart/pause)
- **VM Lifecycle**: Virtual machine control and monitoring
- **Resource Monitoring**: Container and VM resource usage

### UPS & Power Management
- **UPS Integration**: Real hardware integration with apcupsd daemon
- **Power Control**: System shutdown and reboot capabilities
- **Battery Monitoring**: UPS battery status, runtime, and load information

### System Control
- **User Scripts**: Execute custom Unraid user scripts
- **Log Management**: System log access and monitoring
- **Command Execution**: Secure command execution with proper validation

### Real-time Updates
- **WebSocket Support**: Live updates for system stats, Docker events, and storage status
- **Event Streaming**: Real-time notifications for system changes

## Architecture

Built with optimized HTTP mux architecture for production deployment on Unraid servers. Designed for reliability, performance, and comprehensive monitoring capabilities.

## Authentication

Supports JWT-based authentication and API key authentication for secure access control.`,
		Version: version,
		Contact: OpenAPIContact{
			Name:  "UMA Development Team",
			URL:   "https://github.com/domalab/uma",
			Email: "ruaan.deysel@gmail.com",
		},
	}
}
