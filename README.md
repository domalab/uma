# UMA (Unraid Management Agent)

_tl;dr_ **UMA** is a comprehensive monitoring and control system for Unraid servers, providing real-time system insights, hardware monitoring, and remote management capabilities.

## Features

- **System Monitoring**: CPU, RAM, storage, and boot disk status
- **Hardware Sensors**: Temperature monitoring, fan speeds, and system health
- **Storage Management**: Array disk health, cache disk status, and usage statistics
- **Power Management**: UPS status and metrics (if connected)
- **GPU Monitoring**: GPU usage, temperature, and status
- **Docker Management**: Container control (start, stop, restart, logs)
- **VM Control**: Virtual machine management and status monitoring
- **REST API**: Full API access for Home Assistant and third-party integrations
- **Security**: Built-in authentication and secure command execution
- **Diagnostics**: System health checks and automated repair workflows

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

UMA provides a comprehensive REST API for integration with Home Assistant and other automation systems. The API includes endpoints for:

- System monitoring and status
- Storage and array management
- Docker container control
- Virtual machine management
- Hardware sensor readings
- System diagnostics and health checks

API documentation is available at `http://your-unraid-ip:8080/api/docs` when the plugin is running.

## Credits

- [Go](https://golang.org/)
- [other packages](./go.mod)
- Original ControlR codebase by Juan B. Rodriguez
