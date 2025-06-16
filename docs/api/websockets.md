# WebSocket Real-time Monitoring Guide

UMA provides WebSocket endpoints for real-time monitoring of system statistics, Docker events, and storage status. This enables live dashboards and immediate notifications of system changes.

## Available WebSocket Endpoints

### System Statistics
**Endpoint**: `ws://your-unraid-ip:34600/api/v1/ws/system/stats`

Real-time system performance metrics including CPU usage, memory consumption, and uptime.

### Docker Events
**Endpoint**: `ws://your-unraid-ip:34600/api/v1/ws/docker/events`

Live Docker container events including start, stop, restart, and status changes.

### Storage Status
**Endpoint**: `ws://your-unraid-ip:34600/api/v1/ws/storage/status`

Real-time storage information including disk usage, array status, and health updates.

## Connection Examples

### JavaScript (Browser)
```javascript
// Connect to system stats
const statsSocket = new WebSocket('ws://your-unraid-ip:34600/api/v1/ws/system/stats');

statsSocket.onopen = function(event) {
    console.log('Connected to system stats');
};

statsSocket.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('System stats:', data);
    updateDashboard(data);
};

statsSocket.onclose = function(event) {
    console.log('Disconnected from system stats');
};

statsSocket.onerror = function(error) {
    console.error('WebSocket error:', error);
};
```

### Python
```python
import asyncio
import websockets
import json

async def monitor_system_stats():
    uri = "ws://your-unraid-ip:34600/api/v1/ws/system/stats"
    
    async with websockets.connect(uri) as websocket:
        print("Connected to system stats")
        
        async for message in websocket:
            data = json.loads(message)
            print(f"CPU: {data['cpu_percent']}%, Memory: {data['memory_percent']}%")

# Run the monitor
asyncio.run(monitor_system_stats())
```

### Node.js
```javascript
const WebSocket = require('ws');

const ws = new WebSocket('ws://your-unraid-ip:34600/api/v1/ws/system/stats');

ws.on('open', function open() {
    console.log('Connected to system stats');
});

ws.on('message', function message(data) {
    const stats = JSON.parse(data);
    console.log('System stats received:', stats);
});

ws.on('close', function close() {
    console.log('Disconnected from system stats');
});
```

### curl (for testing)
```bash
# Test WebSocket connection
curl --include \
     --no-buffer \
     --header "Connection: Upgrade" \
     --header "Upgrade: websocket" \
     --header "Sec-WebSocket-Key: SGVsbG8sIHdvcmxkIQ==" \
     --header "Sec-WebSocket-Version: 13" \
     http://your-unraid-ip:34600/api/v1/ws/system/stats
```

## Message Formats

### System Statistics Message
```json
{
  "type": "stats",
  "timestamp": "2025-06-15T23:00:00Z",
  "data": {
    "cpu_percent": 15.2,
    "memory_percent": 45.8,
    "memory_used": 8589934592,
    "memory_total": 17179869184,
    "uptime": 86400,
    "load_average": [0.5, 0.7, 0.9],
    "disk_io": {
      "read_bytes": 1048576,
      "write_bytes": 2097152
    },
    "network_io": {
      "bytes_sent": 1048576,
      "bytes_recv": 2097152
    }
  }
}
```

### Docker Events Message
```json
{
  "type": "docker_event",
  "timestamp": "2025-06-15T23:00:00Z",
  "data": {
    "action": "start",
    "container_id": "abc123def456",
    "container_name": "plex",
    "image": "plexinc/pms-docker:latest",
    "status": "running",
    "attributes": {
      "exitCode": "0",
      "signal": ""
    }
  }
}
```

### Storage Status Message
```json
{
  "type": "storage_status",
  "timestamp": "2025-06-15T23:00:00Z",
  "data": {
    "array_status": "started",
    "array_protection": "protected",
    "disks": [
      {
        "name": "disk1",
        "device": "/dev/sdb1",
        "status": "active",
        "temperature": 35,
        "size": "8TB",
        "used": "4.2TB",
        "free": "3.8TB"
      }
    ],
    "cache_disks": [
      {
        "name": "cache",
        "device": "/dev/nvme0n1",
        "status": "active",
        "temperature": 42,
        "size": "1TB",
        "used": "256GB",
        "free": "768GB"
      }
    ]
  }
}
```

## Building a Real-time Dashboard

### HTML Dashboard Example
```html
<!DOCTYPE html>
<html>
<head>
    <title>UMA Real-time Dashboard</title>
    <style>
        .metric { margin: 10px; padding: 10px; border: 1px solid #ccc; }
        .value { font-size: 24px; font-weight: bold; }
    </style>
</head>
<body>
    <h1>UMA System Monitor</h1>
    
    <div class="metric">
        <div>CPU Usage</div>
        <div class="value" id="cpu">--</div>
    </div>
    
    <div class="metric">
        <div>Memory Usage</div>
        <div class="value" id="memory">--</div>
    </div>
    
    <div class="metric">
        <div>Uptime</div>
        <div class="value" id="uptime">--</div>
    </div>
    
    <div class="metric">
        <div>Docker Events</div>
        <div id="docker-events"></div>
    </div>

    <script>
        // System stats connection
        const statsWs = new WebSocket('ws://your-unraid-ip:34600/api/v1/ws/system/stats');
        
        statsWs.onmessage = function(event) {
            const data = JSON.parse(event.data);
            if (data.type === 'stats') {
                document.getElementById('cpu').textContent = data.data.cpu_percent.toFixed(1) + '%';
                document.getElementById('memory').textContent = data.data.memory_percent.toFixed(1) + '%';
                document.getElementById('uptime').textContent = formatUptime(data.data.uptime);
            }
        };
        
        // Docker events connection
        const dockerWs = new WebSocket('ws://your-unraid-ip:34600/api/v1/ws/docker/events');
        
        dockerWs.onmessage = function(event) {
            const data = JSON.parse(event.data);
            if (data.type === 'docker_event') {
                const eventsDiv = document.getElementById('docker-events');
                const eventElement = document.createElement('div');
                eventElement.textContent = `${data.data.action}: ${data.data.container_name}`;
                eventsDiv.insertBefore(eventElement, eventsDiv.firstChild);
                
                // Keep only last 5 events
                while (eventsDiv.children.length > 5) {
                    eventsDiv.removeChild(eventsDiv.lastChild);
                }
            }
        };
        
        function formatUptime(seconds) {
            const days = Math.floor(seconds / 86400);
            const hours = Math.floor((seconds % 86400) / 3600);
            const minutes = Math.floor((seconds % 3600) / 60);
            return `${days}d ${hours}h ${minutes}m`;
        }
    </script>
</body>
</html>
```

## Connection Management

### Automatic Reconnection
```javascript
class UMAWebSocket {
    constructor(endpoint) {
        this.endpoint = endpoint;
        this.reconnectInterval = 5000; // 5 seconds
        this.maxReconnectAttempts = 10;
        this.reconnectAttempts = 0;
        this.connect();
    }
    
    connect() {
        this.ws = new WebSocket(this.endpoint);
        
        this.ws.onopen = () => {
            console.log('Connected to', this.endpoint);
            this.reconnectAttempts = 0;
        };
        
        this.ws.onmessage = (event) => {
            this.handleMessage(JSON.parse(event.data));
        };
        
        this.ws.onclose = () => {
            console.log('Disconnected from', this.endpoint);
            this.reconnect();
        };
        
        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    }
    
    reconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            console.log(`Reconnecting... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
            setTimeout(() => this.connect(), this.reconnectInterval);
        } else {
            console.error('Max reconnection attempts reached');
        }
    }
    
    handleMessage(data) {
        // Override this method to handle messages
        console.log('Received:', data);
    }
}

// Usage
const systemMonitor = new UMAWebSocket('ws://your-unraid-ip:34600/api/v1/ws/system/stats');
systemMonitor.handleMessage = function(data) {
    if (data.type === 'stats') {
        updateSystemStats(data.data);
    }
};
```

## Performance Considerations

### Message Frequency
- **System stats**: Updated every 5 seconds
- **Docker events**: Real-time (as they occur)
- **Storage status**: Updated every 30 seconds

### Connection Limits
- Maximum 100 concurrent WebSocket connections per endpoint
- Connections are automatically cleaned up when clients disconnect

### Bandwidth Usage
- System stats: ~500 bytes per message
- Docker events: ~200-800 bytes per event
- Storage status: ~1-5KB per message

## Troubleshooting

### Common Issues

**Connection refused:**
```
Error: WebSocket connection failed
```
- Verify UMA is running
- Check the correct port (34600)
- Ensure WebSocket endpoint path is correct

**Connection drops frequently:**
```javascript
// Add connection monitoring
ws.addEventListener('close', function(event) {
    console.log('Close code:', event.code);
    console.log('Close reason:', event.reason);
});
```

**No messages received:**
- Check if the specific monitoring service is enabled
- Verify system has activity to generate events
- Test with curl to ensure endpoint responds

### Debugging WebSocket Connections

```javascript
// Enable verbose logging
const ws = new WebSocket('ws://your-unraid-ip:34600/api/v1/ws/system/stats');

ws.addEventListener('open', function(event) {
    console.log('WebSocket opened:', event);
});

ws.addEventListener('message', function(event) {
    console.log('Message received:', event.data);
});

ws.addEventListener('error', function(event) {
    console.error('WebSocket error:', event);
});

ws.addEventListener('close', function(event) {
    console.log('WebSocket closed:', event.code, event.reason);
});
```

## Integration Examples

### Home Assistant
```yaml
# configuration.yaml
sensor:
  - platform: websocket
    resource: ws://your-unraid-ip:34600/api/v1/ws/system/stats
    name: "Unraid CPU Usage"
    value_template: "{{ value_json.data.cpu_percent }}"
    unit_of_measurement: "%"
```

### Grafana Dashboard
Use WebSocket data with Grafana's WebSocket data source plugin to create real-time dashboards.

### Custom Monitoring Scripts
Create custom monitoring solutions that react to real-time events and send notifications or trigger automation.

## Next Steps

- **[Complete Endpoint Reference](endpoints.md)** - All available REST endpoints
- **[Bulk Operations](bulk-operations.md)** - Efficient container management
- **[Metrics Guide](../development/metrics.md)** - Prometheus monitoring setup
