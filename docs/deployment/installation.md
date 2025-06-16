# UMA Installation Guide

This guide covers the installation and initial setup of UMA (Unraid Management Agent) on your Unraid server.

## Prerequisites

- Unraid 6.8.0 or later
- Internet connection for downloading the plugin
- Basic familiarity with Unraid web interface

## Installation Methods

### Method 1: Community Applications (Recommended)

1. **Open Unraid Web Interface**
   - Navigate to your Unraid server: `http://your-unraid-ip`
   - Log in with your credentials

2. **Access Community Applications**
   - Go to the **Apps** tab
   - Click on the **Plugins** button
   - Search for "UMA" in the search box

3. **Install UMA**
   - Find "UMA (Unraid Management Agent)" in the results
   - Click **Install**
   - Wait for the installation to complete

### Method 2: Manual Plugin Installation

1. **Navigate to Plugins**
   - Go to the **Plugins** tab in Unraid
   - Click on **Install Plugin**

2. **Install from URL**
   - Paste the following URL in the input field:
     ```
     https://github.com/domalab/uma/releases/latest/download/uma.plg
     ```
   - Click **Install**

3. **Wait for Installation**
   - The plugin will download and install automatically
   - You'll see progress messages during installation

## Post-Installation Setup

### 1. Access UMA Settings

After installation, access UMA settings via:

**Option A: Settings Menu**
- Go to **Settings** > **Utilities**
- Click on **UMA**
- Click **Open Web UI**

**Option B: Plugins Menu**
- Go to **Plugins** > **Installed Plugins**
- Find **UMA** in the list
- Click **Open Web UI**

### 2. Verify Installation

1. **Check Service Status**
   - The UMA service should start automatically
   - Default port: `34600`
   - Access: `http://your-unraid-ip:34600`

2. **Test API Access**
   ```bash
   curl http://your-unraid-ip:34600/api/v1/health
   ```
   
   Expected response:
   ```json
   {
     "status": "healthy",
     "service": "uma",
     "dependencies": {
       "docker": "healthy",
       "libvirt": "healthy",
       "storage": "healthy",
       "notifications": "healthy"
     }
   }
   ```

3. **Access Documentation**
   - Swagger UI: `http://your-unraid-ip:34600/api/v1/docs`
   - OpenAPI Spec: `http://your-unraid-ip:34600/api/v1/openapi.json`
   - Metrics: `http://your-unraid-ip:34600/metrics`

## Configuration

### Basic Configuration

UMA works out of the box with default settings. No additional configuration is required for basic functionality.

### Advanced Configuration

For advanced users, configuration options include:

1. **Port Configuration**
   - Default: `34600`
   - Can be changed in plugin settings if needed

2. **Log Level**
   - Default: `info`
   - Available levels: `debug`, `info`, `warn`, `error`

3. **Monitoring Features**
   - All monitoring features are enabled by default
   - Includes: Docker, storage, system stats, WebSocket endpoints

### Environment Variables

UMA supports the following environment variables:

```bash
# Log level configuration
UMA_LOG_LEVEL=info

# HTTP server configuration
UMA_HTTP_PORT=34600

# Feature toggles
UMA_ENABLE_METRICS=true
UMA_ENABLE_WEBSOCKETS=true
```

## Verification Steps

### 1. Service Health Check

```bash
# Check if UMA is running
curl -s http://your-unraid-ip:34600/api/v1/health | jq '.'
```

### 2. API Functionality Test

```bash
# Test system stats
curl -s http://your-unraid-ip:34600/api/v1/system/stats | jq '.data'

# Test Docker containers
curl -s http://your-unraid-ip:34600/api/v1/docker/containers | jq '.data[0]'

# Test storage information
curl -s http://your-unraid-ip:34600/api/v1/storage/disks | jq '.data[0]'
```

### 3. WebSocket Connectivity

Test WebSocket endpoints:

```javascript
// Test in browser console
const ws = new WebSocket('ws://your-unraid-ip:34600/api/v1/ws/system/stats');
ws.onmessage = function(event) {
    console.log('Received:', JSON.parse(event.data));
};
```

### 4. Metrics Collection

```bash
# Check Prometheus metrics
curl -s http://your-unraid-ip:34600/metrics | grep uma_
```

## Integration Setup

### Home Assistant Integration

Add to your `configuration.yaml`:

```yaml
# REST sensors
sensor:
  - platform: rest
    resource: http://your-unraid-ip:34600/api/v1/system/stats
    name: "Unraid CPU Usage"
    value_template: "{{ value_json.data.cpu_percent }}"
    unit_of_measurement: "%"
    scan_interval: 30

  - platform: rest
    resource: http://your-unraid-ip:34600/api/v1/system/stats
    name: "Unraid Memory Usage"
    value_template: "{{ value_json.data.memory_percent }}"
    unit_of_measurement: "%"
    scan_interval: 30

# Binary sensors for health checks
binary_sensor:
  - platform: rest
    resource: http://your-unraid-ip:34600/api/v1/health
    name: "Unraid Docker Health"
    value_template: "{{ value_json.dependencies.docker == 'healthy' }}"
    scan_interval: 60
```

### Prometheus Monitoring

Add to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'uma'
    static_configs:
      - targets: ['your-unraid-ip:34600']
    metrics_path: '/metrics'
    scrape_interval: 30s
    scrape_timeout: 10s
```

### Grafana Dashboard

1. Import Prometheus data source
2. Create dashboard with UMA metrics
3. Use queries from the [metrics guide](../development/metrics.md)

## Troubleshooting

### Common Issues

**UMA not starting:**
1. Check Unraid system logs
2. Verify port 34600 is not in use
3. Restart the UMA plugin

**API not responding:**
1. Verify UMA service is running
2. Check firewall settings
3. Test with curl from Unraid terminal

**Missing data:**
1. Ensure Docker service is running
2. Check storage array is started
3. Verify libvirt is enabled (for VM monitoring)

### Log Analysis

Check UMA logs for issues:

```bash
# View recent logs
tail -f /var/log/uma/uma.log

# Search for errors
grep -i error /var/log/uma/uma.log

# Check structured logs
grep "component=api" /var/log/uma/uma.log | tail -10
```

### Performance Issues

If experiencing performance issues:

1. **Reduce log level** to `warn` or `error`
2. **Increase scrape intervals** in monitoring tools
3. **Limit concurrent connections** to WebSocket endpoints

### Getting Help

1. **Check Documentation**
   - [API Reference](../api/endpoints.md)
   - [WebSocket Guide](../api/websockets.md)
   - [Troubleshooting Guide](troubleshooting.md)

2. **Community Support**
   - [GitHub Issues](https://github.com/domalab/uma/issues)
   - [GitHub Discussions](https://github.com/domalab/uma/discussions)
   - Unraid Community Forums

3. **Debug Information**
   When reporting issues, include:
   - UMA version
   - Unraid version
   - Error logs
   - API response examples

## Uninstallation

To remove UMA:

1. **Stop the Service**
   - Go to **Plugins** > **Installed Plugins**
   - Find **UMA** and click **Remove**

2. **Clean Up (Optional)**
   ```bash
   # Remove logs
   rm -rf /var/log/uma/
   
   # Remove any custom configuration
   rm -f /etc/uma/config.ini
   ```

## Next Steps

After successful installation:

1. **Explore the API** - Visit the [API documentation](../api/README.md)
2. **Set up monitoring** - Follow the [monitoring guide](monitoring.md)
3. **Configure integrations** - Set up Home Assistant or Grafana
4. **Test WebSocket endpoints** - Try real-time monitoring

## Security Considerations

- UMA is designed for trusted network environments
- No authentication is currently implemented
- Consider firewall rules if exposing outside your network
- Monitor access logs for unusual activity

## Updates

UMA updates are delivered through the Unraid plugin system:

1. **Automatic Updates** - Enable in plugin settings
2. **Manual Updates** - Check for updates in the Plugins tab
3. **Version Information** - Available in the API at `/api/v1/health`

Stay updated with the latest features and security improvements!
