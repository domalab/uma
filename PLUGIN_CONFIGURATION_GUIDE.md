# UMA Plugin Configuration Guide

## Overview

The UMA (Unraid Management Agent) plugin has been enhanced with Phase 3 features and an improved configuration interface that provides users with comprehensive control through the Unraid web UI.

## Plugin Configuration Options

### 1. Enable/Disable Toggle
- **Setting**: "Enable UMA"
- **Options**: Yes/No
- **Default**: Yes
- **Description**: Controls whether the UMA service starts automatically

### 2. Port Configuration
- **Setting**: "HTTP Port"
- **Range**: 1024-65535
- **Default**: 34600
- **Validation**: 
  - Prevents use of ports below 1024 (system reserved)
  - Warns about commonly used ports (22, 80, 443, etc.)
  - Validates numeric input

### 3. UPS Status
- **Setting**: "Provide UPS status"
- **Options**: Yes/No
- **Default**: Yes
- **Description**: Enables UPS monitoring via NUT integration

### 4. Quick Access Link
- **Feature**: "Open Web UI" button
- **Target**: Swagger API documentation interface
- **URL**: `http://[UNRAID-IP]:[CONFIGURED-PORT]/api/v1/docs`
- **Availability**: Only shown when service is running

## Plugin Access Methods

### Method 1 - Settings Access
1. Navigate to **Settings > Utilities**
2. Locate and click on **"UMA"**
3. Configure settings as needed
4. Click **"Open Web UI"** button to launch Swagger interface

### Method 2 - Plugin Management Access
1. Navigate to **Plugins > Installed Plugins**
2. Locate and click on **"UMA"**
3. Configure settings as needed
4. Click **"Open Web UI"** button to launch Swagger interface

## Technical Implementation

### Configuration Files
- **Plugin Config**: `/boot/config/plugins/uma/uma.cfg`
- **Environment**: `/boot/config/plugins/uma/uma.env`
- **Plugin Definition**: `uma.plg`

### Configuration Parameters
```bash
SERVICE="enable"     # Enable/disable service
PORT="34600"         # HTTP port number
UPS="enable"         # UPS monitoring
```

### Start Script Enhancement
The start script now supports dynamic port configuration:
```bash
nohup sudo -H bash -c "$prog boot --http-port=$PORT $SHOWUPS" >/dev/null 2>&1 &
```

### Port Validation
- Client-side JavaScript validation
- Server-side validation in start script
- Reserved port warnings
- Fallback to default port (34600) if invalid

## User Interface Enhancements

### Status Display
- **Running**: Green status with port information and Web UI link
- **Stopped**: Orange status with configuration prompt
- **Version**: Displays current UMA version (2025.06.16)

### Form Styling
- Professional CSS styling
- Responsive layout
- Clear visual hierarchy
- Intuitive button placement

### Interactive Elements
- Real-time form validation
- Dynamic Web UI link generation
- Contextual help text
- Confirmation dialogs for reserved ports

## Phase 3 Features Integration

### Enhanced Binary
- **Size**: 14.4MB (optimized with `-ldflags="-s -w"`)
- **Features**: JWT Authentication, Chi Router, Viper Configuration
- **Compatibility**: 100% backward compatible

### Configuration Management
- **Viper Integration**: Multi-format configuration support
- **Environment Variables**: `UMA_*` prefix support
- **Hot Reload**: Configuration file watching
- **Validation**: Comprehensive input validation

### API Documentation
- **Swagger UI**: Interactive API documentation
- **OpenAPI 3.1.1**: Complete API specification
- **Real-time Access**: Direct link from plugin interface

## Installation and Upgrade

### New Installations
1. Install plugin via Community Applications or manual URL
2. Configure desired port and settings
3. Enable service
4. Access Web UI via provided link

### Upgrades from Previous Versions
- **Automatic Migration**: Existing configurations preserved
- **Default Port**: Maintains 34600 if not configured
- **Service State**: Preserves enable/disable state
- **Zero Downtime**: Seamless upgrade process

## Troubleshooting

### Common Issues
1. **Port Conflicts**: Use port validation to avoid conflicts
2. **Service Not Starting**: Check port availability and permissions
3. **Web UI Not Accessible**: Verify port configuration and firewall

### Validation Errors
- **Port Range**: Must be between 1024-65535
- **Reserved Ports**: Warnings for commonly used ports
- **Numeric Input**: Only accepts valid numbers

## Security Considerations

### Port Selection
- Avoid well-known ports (22, 80, 443, etc.)
- Use high-numbered ports (above 10000) for better security
- Consider firewall rules for external access

### Authentication
- JWT authentication available (disabled by default)
- API key-based access control
- Role-based permissions (Admin, Operator, Viewer)

## Support and Documentation

### Resources
- **API Documentation**: Available via Web UI link
- **GitHub Repository**: https://github.com/domalab/uma
- **Plugin Bundle**: 7.0MB compressed package
- **Binary Size**: 14.4MB optimized executable

### Version Information
- **Plugin Version**: 2025.06.16
- **UMA Version**: Phase 3 Enhanced Features
- **Compatibility**: Unraid 6.8+ recommended

## Conclusion

The enhanced UMA plugin configuration interface provides users with comprehensive control over the monitoring agent while maintaining simplicity and professional presentation. The integration of Phase 3 features ensures enterprise-grade functionality with an intuitive user experience.
