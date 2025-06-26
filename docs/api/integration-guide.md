# UMA API Integration Guide

This guide provides real-world examples for integrating UMA's enhanced monitoring capabilities with Home Assistant, Prometheus, and other automation platforms. All examples use **actual response data** from production Unraid servers.

## Quick Integration Examples

### Home Assistant Integration

#### Storage Capacity Monitoring
Monitor your Unraid array capacity with real usage data:

```yaml
# configuration.yaml
sensor:
  - platform: rest
    name: "Unraid Array Usage"
    resource: "http://your-unraid-ip:34600/api/v1/storage/array"
    value_template: "{{ value_json.usage_percent | round(1) }}"
    unit_of_measurement: "%"
    json_attributes:
      - total_capacity_formatted
      - total_used_formatted
      - total_free_formatted
      - disk_count
```

**Real Response Data:**
```json
{
  "total_capacity": 41996310249472,
  "total_capacity_formatted": "38.2 TB",
  "total_used": 9099742822400,
  "total_used_formatted": "8.3 TB",
  "total_free": 32896567427072,
  "total_free_formatted": "29.9 TB",
  "usage_percent": 21.67,
  "disk_count": 8
}
```

#### UPS Power Monitoring
Monitor real UPS power consumption and battery status:

```yaml
# configuration.yaml
sensor:
  - platform: rest
    name: "UPS Power Consumption"
    resource: "http://your-unraid-ip:34600/api/v1/system/ups"
    value_template: "{{ value_json.power_consumption }}"
    unit_of_measurement: "W"
    json_attributes:
      - battery_charge
      - runtime
      - voltage
      - nominal_power
      - status

  - platform: rest
    name: "UPS Battery"
    resource: "http://your-unraid-ip:34600/api/v1/system/ups"
    value_template: "{{ value_json.battery_charge }}"
    unit_of_measurement: "%"
```

**Real Response Data:**
```json
{
  "available": true,
  "status": "online",
  "battery_charge": 100,
  "load": 0,
  "runtime": 220,
  "voltage": 236,
  "power_consumption": 0,
  "nominal_power": 480
}
```

### Prometheus Integration

#### Storage Metrics
```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'unraid-uma'
    static_configs:
      - targets: ['your-unraid-ip:34600']
    metrics_path: '/metrics'
    scrape_interval: 30s
```

#### Custom Metrics Collection
```bash
# Collect storage usage percentage
curl -s http://your-unraid-ip:34600/api/v1/storage/array | \
  jq '.usage_percent' | \
  awk '{print "unraid_array_usage_percent " $1}'

# Collect UPS power consumption
curl -s http://your-unraid-ip:34600/api/v1/system/ups | \
  jq '.power_consumption' | \
  awk '{print "unraid_ups_power_watts " $1}'
```

## Advanced Integration Patterns

### Real-time WebSocket Monitoring

#### JavaScript WebSocket Client
```javascript
const ws = new WebSocket('ws://your-unraid-ip:34600/api/v1/ws');

ws.onopen = function() {
    // Subscribe to storage events
    ws.send(JSON.stringify({
        action: 'subscribe',
        channel: 'storage'
    }));
    
    // Subscribe to UPS events
    ws.send(JSON.stringify({
        action: 'subscribe', 
        channel: 'ups'
    }));
};

ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    
    if (data.channel === 'storage') {
        console.log('Storage update:', data.usage_percent + '%');
    }
    
    if (data.channel === 'ups') {
        console.log('UPS power:', data.power_consumption + 'W');
    }
};
```

### Container Performance Monitoring

#### Docker Container Metrics
```bash
# Get all container performance metrics
curl -s http://your-unraid-ip:34600/api/v1/docker/containers | \
  jq '.[] | {name: .name, cpu_percent: .cpu_percent, memory_usage: .memory_usage}'
```

**Real Response Example:**
```json
[
  {
    "name": "plex",
    "cpu_percent": 2.5,
    "memory_usage": 1073741824,
    "memory_percent": 6.25,
    "network_rx": 1048576,
    "network_tx": 2097152
  }
]
```

### GPU Monitoring Integration

#### GPU Performance Tracking
```bash
# Get GPU utilization and temperature
curl -s http://your-unraid-ip:34600/api/v1/system/gpu | \
  jq '.[] | {name: .name, utilization: .utilization, temperature: .temperature}'
```

**Real Response Example:**
```json
[
  {
    "name": "Intel UHD Graphics 630",
    "utilization": {
      "gpu": 0,
      "memory": 0
    },
    "temperature": 45,
    "power": {
      "draw_watts": 2.5,
      "limit_watts": 15
    }
  }
]
```

## Error Handling and Best Practices

### Robust API Calls
```bash
# Storage monitoring with error handling
STORAGE_DATA=$(curl -s --max-time 10 http://your-unraid-ip:34600/api/v1/storage/array)
if [ $? -eq 0 ] && echo "$STORAGE_DATA" | jq -e '.usage_percent' > /dev/null; then
    USAGE=$(echo "$STORAGE_DATA" | jq -r '.usage_percent')
    echo "Array usage: ${USAGE}%"
else
    echo "Failed to get storage data"
fi

# UPS monitoring with fallback
UPS_DATA=$(curl -s --max-time 10 http://your-unraid-ip:34600/api/v1/system/ups)
if [ $? -eq 0 ] && echo "$UPS_DATA" | jq -e '.available' > /dev/null; then
    if [ "$(echo "$UPS_DATA" | jq -r '.available')" = "true" ]; then
        POWER=$(echo "$UPS_DATA" | jq -r '.power_consumption')
        BATTERY=$(echo "$UPS_DATA" | jq -r '.battery_charge')
        echo "UPS: ${POWER}W, Battery: ${BATTERY}%"
    else
        echo "UPS not available"
    fi
else
    echo "Failed to get UPS data"
fi
```

### Rate Limiting and Caching
```bash
# Cache responses to avoid excessive API calls
CACHE_DIR="/tmp/uma_cache"
CACHE_TTL=30  # seconds

mkdir -p "$CACHE_DIR"

get_cached_data() {
    local endpoint="$1"
    local cache_file="$CACHE_DIR/$(echo "$endpoint" | sed 's/[^a-zA-Z0-9]/_/g')"
    
    if [ -f "$cache_file" ] && [ $(($(date +%s) - $(stat -c %Y "$cache_file"))) -lt $CACHE_TTL ]; then
        cat "$cache_file"
    else
        curl -s "http://your-unraid-ip:34600/api/v1/$endpoint" | tee "$cache_file"
    fi
}

# Usage
STORAGE_DATA=$(get_cached_data "storage/array")
UPS_DATA=$(get_cached_data "system/ups")
```

## Data Quality Assurance

All UMA API endpoints provide:
- ✅ **Real System Data**: No placeholder or hardcoded values
- ✅ **Hardware Validation**: Tested on production Unraid servers
- ✅ **Consistent Formatting**: Standardized response structures
- ✅ **Error Handling**: Proper HTTP status codes and error messages
- ✅ **Real-time Updates**: Data refreshed from live system sources

For complete API documentation, visit: `http://your-unraid-ip:34600/api/v1/docs`
