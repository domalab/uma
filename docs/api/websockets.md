# WebSocket Real-time Monitoring Guide

UMA provides a unified WebSocket endpoint for real-time monitoring with subscription management. This enables live dashboards and immediate notifications of system changes across multiple event types.

## Unified WebSocket Endpoint

### Main Endpoint
**Endpoint**: `ws://your-unraid-ip:34600/api/v1/ws`

The unified WebSocket endpoint supports subscription management for multiple event types:

- **System Statistics**: CPU usage, memory consumption, and uptime
- **Docker Events**: Container start, stop, restart, and status changes  
- **Storage Status**: Array status, disk health, and parity information
- **Temperature Alerts**: Critical temperature monitoring and alerts
- **Resource Alerts**: CPU, memory, and disk usage threshold alerts
- **Infrastructure Status**: UPS, fans, and power monitoring

## Connection Examples

### JavaScript (Browser)

```javascript
// Connect to unified WebSocket endpoint
const socket = new WebSocket('ws://your-unraid-ip:34600/api/v1/ws');

socket.onopen = function(event) {
    console.log('Connected to unified WebSocket');
    
    // Subscribe to system stats
    socket.send(JSON.stringify({
        type: 'subscribe',
        channel: 'system.stats'
    }));
    
    // Subscribe to Docker events
    socket.send(JSON.stringify({
        type: 'subscribe', 
        channel: 'docker.events'
    }));
    
    // Subscribe to temperature alerts
    socket.send(JSON.stringify({
        type: 'subscribe',
        channel: 'temperature.alert'
    }));
};

socket.onmessage = function(event) {
    const message = JSON.parse(event.data);
    console.log('Received event:', message);
    
    switch(message.type) {
        case 'system.stats':
            updateSystemStats(message.data);
            break;
        case 'docker.events':
            updateDockerEvents(message.data);
            break;
        case 'storage.status':
            updateStorageStatus(message.data);
            break;
        case 'temperature.alert':
            showTemperatureAlert(message.data);
            break;
        case 'resource.alert':
            showResourceAlert(message.data);
            break;
    }
};

socket.onclose = function(event) {
    console.log('Disconnected from WebSocket');
};

socket.onerror = function(error) {
    console.error('WebSocket error:', error);
};
```

### Python

```python
import asyncio
import websockets
import json

async def monitor_unraid():
    uri = "ws://your-unraid-ip:34600/api/v1/ws"
    
    async with websockets.connect(uri) as websocket:
        print("Connected to unified WebSocket")
        
        # Subscribe to multiple channels
        await websocket.send(json.dumps({
            "type": "subscribe",
            "channel": "system.stats"
        }))
        
        await websocket.send(json.dumps({
            "type": "subscribe", 
            "channel": "docker.events"
        }))
        
        async for message in websocket:
            data = json.loads(message)
            
            if data['type'] == 'system.stats':
                print(f"CPU: {data['data']['cpu_percent']}%, Memory: {data['data']['memory_percent']}%")
            elif data['type'] == 'docker.events':
                print(f"Docker: {data['data']['action']} - {data['data']['container_name']}")
            elif data['type'] == 'temperature.alert':
                print(f"Temperature Alert: {data['data']['message']}")

# Run the monitor
asyncio.run(monitor_unraid())
```

### Node.js

```javascript
const WebSocket = require('ws');

const ws = new WebSocket('ws://your-unraid-ip:34600/api/v1/ws');

ws.on('open', function open() {
    console.log('Connected to unified WebSocket');
    
    // Subscribe to system stats
    ws.send(JSON.stringify({
        type: 'subscribe',
        channel: 'system.stats'
    }));
    
    // Subscribe to storage status
    ws.send(JSON.stringify({
        type: 'subscribe',
        channel: 'storage.status'
    }));
});

ws.on('message', function message(data) {
    const event = JSON.parse(data);
    console.log('Event received:', event);
    
    switch(event.type) {
        case 'system.stats':
            console.log(`System: CPU ${event.data.cpu_percent}%, Memory ${event.data.memory_percent}%`);
            break;
        case 'storage.status':
            console.log(`Storage: Array ${event.data.array_status}, Disks: ${event.data.disks.length}`);
            break;
    }
});

ws.on('close', function close() {
    console.log('Disconnected from WebSocket');
});
```

## Subscription Management

### Available Channels

- `system.stats` - Real-time system performance metrics
- `docker.events` - Docker container lifecycle events
- `storage.status` - Storage array and disk status updates
- `temperature.alert` - Temperature threshold alerts
- `resource.alert` - CPU/memory/disk usage alerts
- `infrastructure.status` - UPS, fans, power monitoring

### Subscribe to Channel

```javascript
socket.send(JSON.stringify({
    type: 'subscribe',
    channel: 'system.stats'
}));
```

### Unsubscribe from Channel

```javascript
socket.send(JSON.stringify({
    type: 'unsubscribe',
    channel: 'system.stats'
}));
```

### List Active Subscriptions

```javascript
socket.send(JSON.stringify({
    type: 'list_subscriptions'
}));
```

## Message Formats

### System Statistics Event

```json
{
  "type": "system.stats",
  "channel": "system.stats",
  "timestamp": "2025-06-19T14:30:00Z",
  "data": {
    "cpu_percent": 25.5,
    "memory_percent": 50.0,
    "memory_used": 8589934592,
    "memory_total": 17179869184,
    "uptime": 86400,
    "load_average": [0.5, 0.7, 0.9],
    "network_io": {
      "bytes_sent": 1048576,
      "bytes_recv": 2097152
    }
  }
}
```

### Docker Events

```json
{
  "type": "docker.events",
  "channel": "docker.events", 
  "timestamp": "2025-06-19T14:30:00Z",
  "data": {
    "action": "start",
    "container_id": "abc123def456",
    "container_name": "plex",
    "image": "plexinc/pms-docker:latest",
    "status": "running"
  }
}
```

### Temperature Alert

```json
{
  "type": "temperature.alert",
  "channel": "temperature.alert",
  "timestamp": "2025-06-19T14:30:00Z", 
  "data": {
    "sensor_name": "CPU Package",
    "sensor_type": "cpu",
    "temperature": 75.2,
    "threshold": 70.0,
    "level": "warning",
    "message": "CPU Package temperature warning: 75.2°C (threshold: 70.0°C)"
  }
}
```

## Integration Examples

### Home Assistant

```yaml
# configuration.yaml
sensor:
  - platform: websocket
    resource: ws://your-unraid-ip:34600/api/v1/ws
    name: "Unraid CPU Usage"
    value_template: "{{ value_json.data.cpu_percent }}"
    unit_of_measurement: "%"
    json_attributes_path: "$.data"
    json_attributes:
      - memory_percent
      - uptime
      - load_average
```

### Real-time Dashboard

Create responsive dashboards that react to live system events and provide immediate feedback on system health and performance.

## Next Steps

- **[Complete API Reference](endpoints.md)** - All available REST endpoints
- **[Temperature Monitoring](temperature-monitoring.md)** - Advanced temperature alerts
- **[Metrics Guide](../development/metrics.md)** - Prometheus monitoring setup
