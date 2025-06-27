# Home Assistant UMA Integration Developer Guide

This comprehensive guide provides Home Assistant developers with everything needed to build a complete UMA integration. All examples use **real response data** from production Unraid servers, ensuring accurate integration development.

## Overview

UMA provides comprehensive monitoring and control capabilities for Unraid servers through a REST API with real-time WebSocket events. This guide maps UMA's API endpoints to specific Home Assistant entity types with complete configuration examples.

## API Foundation

### Base Configuration
```yaml
# configuration.yaml
rest:
  - resource: "http://your-unraid-ip:34600/api/v1"
    scan_interval: 30
    headers:
      X-Request-ID: "home-assistant"
      Accept: "application/json"
```

### Data Quality Guarantee
- ✅ **100% Real Data**: All UMA endpoints return actual system measurements
- ✅ **No Placeholders**: Eliminated hardcoded estimates and mock values
- ✅ **Hardware Validated**: Tested on production Unraid servers
- ✅ **Real-time Updates**: Live data from system sources

## Storage Monitoring Integration

### Array Capacity Monitoring

**API Endpoint:** `/api/v1/storage/array`

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
  "disk_count": 8,
  "state": "started",
  "protection": "parity"
}
```

**Home Assistant Configuration:**
```yaml
# Storage Array Sensors
sensor:
  - platform: rest
    name: "Unraid Array Usage"
    resource: "http://your-unraid-ip:34600/api/v1/storage/array"
    value_template: "{{ value_json.usage_percent | round(1) }}"
    unit_of_measurement: "%"
    device_class: "data_size"
    state_class: "measurement"
    json_attributes:
      - total_capacity_formatted
      - total_used_formatted
      - total_free_formatted
      - disk_count
      - state
      - protection

  - platform: rest
    name: "Unraid Array Capacity"
    resource: "http://your-unraid-ip:34600/api/v1/storage/array"
    value_template: "{{ (value_json.total_capacity / 1099511627776) | round(1) }}"
    unit_of_measurement: "TB"
    device_class: "data_size"
    state_class: "total"

  - platform: rest
    name: "Unraid Array Used"
    resource: "http://your-unraid-ip:34600/api/v1/storage/array"
    value_template: "{{ (value_json.total_used / 1099511627776) | round(1) }}"
    unit_of_measurement: "TB"
    device_class: "data_size"
    state_class: "total_increasing"

# Array Status Binary Sensor
binary_sensor:
  - platform: rest
    name: "Unraid Array Online"
    resource: "http://your-unraid-ip:34600/api/v1/storage/array"
    value_template: "{{ value_json.state == 'started' }}"
    device_class: "connectivity"
```

### Individual Disk Monitoring

**API Endpoint:** `/api/v1/storage/disks`

**Home Assistant Configuration:**
```yaml
# Disk Temperature Sensors (Template sensors from REST data)
sensor:
  - platform: rest
    name: "Unraid Disks Data"
    resource: "http://your-unraid-ip:34600/api/v1/storage/disks"
    json_attributes:
      - disks
    value_template: "{{ value_json.disks | length }}"

template:
  - sensor:
      - name: "Disk1 Temperature"
        unit_of_measurement: "°C"
        device_class: "temperature"
        state_class: "measurement"
        state: >
          {% set disks = state_attr('sensor.unraid_disks_data', 'disks') %}
          {% if disks %}
            {% for disk in disks %}
              {% if disk.name == 'disk1' %}
                {{ disk.temperature }}
              {% endif %}
            {% endfor %}
          {% endif %}

      - name: "Disk1 Usage"
        unit_of_measurement: "%"
        device_class: "data_size"
        state_class: "measurement"
        state: >
          {% set disks = state_attr('sensor.unraid_disks_data', 'disks') %}
          {% if disks %}
            {% for disk in disks %}
              {% if disk.name == 'disk1' %}
                {{ disk.usage_percent | round(1) }}
              {% endif %}
            {% endfor %}
          {% endif %}
```

## UPS Power Management Integration

### UPS Monitoring

**API Endpoint:** `/api/v1/system/ups`

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

**Home Assistant Configuration:**
```yaml
# UPS Power Monitoring
sensor:
  - platform: rest
    name: "UPS Power Consumption"
    resource: "http://your-unraid-ip:34600/api/v1/system/ups"
    value_template: "{{ value_json.power_consumption | round(1) }}"
    unit_of_measurement: "W"
    device_class: "power"
    state_class: "measurement"
    json_attributes:
      - nominal_power
      - voltage
      - load

  - platform: rest
    name: "UPS Battery Level"
    resource: "http://your-unraid-ip:34600/api/v1/system/ups"
    value_template: "{{ value_json.battery_charge }}"
    unit_of_measurement: "%"
    device_class: "battery"
    state_class: "measurement"

  - platform: rest
    name: "UPS Runtime"
    resource: "http://your-unraid-ip:34600/api/v1/system/ups"
    value_template: "{{ value_json.runtime }}"
    unit_of_measurement: "min"
    device_class: "duration"
    state_class: "measurement"

  - platform: rest
    name: "UPS Line Voltage"
    resource: "http://your-unraid-ip:34600/api/v1/system/ups"
    value_template: "{{ value_json.voltage }}"
    unit_of_measurement: "V"
    device_class: "voltage"
    state_class: "measurement"

# UPS Status Binary Sensors
binary_sensor:
  - platform: rest
    name: "UPS Available"
    resource: "http://your-unraid-ip:34600/api/v1/system/ups"
    value_template: "{{ value_json.available }}"
    device_class: "connectivity"

  - platform: rest
    name: "UPS Online"
    resource: "http://your-unraid-ip:34600/api/v1/system/ups"
    value_template: "{{ value_json.status == 'online' }}"
    device_class: "power"
```

## Container Management Integration

### Docker Container Control

**API Endpoint:** `/api/v1/docker/containers`

**Home Assistant Configuration:**
```yaml
# Container Control Switches
switch:
  - platform: rest
    name: "Plex Container"
    resource: "http://your-unraid-ip:34600/api/v1/docker/containers/plex"
    body_on: '{"action": "start"}'
    body_off: '{"action": "stop"}'
    is_on_template: >
      {{ value_json.status == 'running' }}
    headers:
      Content-Type: "application/json"
      X-Request-ID: "home-assistant-plex"

# Container Performance Sensors
sensor:
  - platform: rest
    name: "Plex CPU Usage"
    resource: "http://your-unraid-ip:34600/api/v1/docker/containers/plex"
    value_template: "{{ value_json.cpu_percent | round(1) }}"
    unit_of_measurement: "%"
    device_class: "power_factor"
    state_class: "measurement"

  - platform: rest
    name: "Plex Memory Usage"
    resource: "http://your-unraid-ip:34600/api/v1/docker/containers/plex"
    value_template: "{{ (value_json.memory_usage / 1073741824) | round(1) }}"
    unit_of_measurement: "GB"
    device_class: "data_size"
    state_class: "measurement"
```

## System Health Monitoring

### Health Status

**API Endpoint:** `/api/v1/health`

**Home Assistant Configuration:**
```yaml
# System Health Binary Sensor
binary_sensor:
  - platform: rest
    name: "Unraid System Health"
    resource: "http://your-unraid-ip:34600/api/v1/health"
    value_template: "{{ value_json.status == 'healthy' }}"
    device_class: "connectivity"
    json_attributes:
      - dependencies

# Individual Service Health
template:
  - binary_sensor:
      - name: "Docker Service Health"
        device_class: "connectivity"
        state: >
          {% set deps = state_attr('binary_sensor.unraid_system_health', 'dependencies') %}
          {{ deps.docker == 'healthy' if deps else false }}

      - name: "Storage Service Health"
        device_class: "connectivity"
        state: >
          {% set deps = state_attr('binary_sensor.unraid_system_health', 'dependencies') %}
          {{ deps.storage == 'healthy' if deps else false }}

## GPU Monitoring Integration

### GPU Performance

**API Endpoint:** `/api/v1/system/gpu`

**Real Response Data:**
```json
[
  {
    "name": "Intel UHD Graphics 630",
    "vendor": "Intel",
    "utilization": {
      "gpu": 0,
      "memory": 0
    },
    "temperature": 45,
    "power": {
      "draw_watts": 2.5,
      "limit_watts": 15,
      "usage_percent": 16.67
    },
    "clocks": {
      "core_mhz": 350,
      "memory_mhz": 0,
      "shader_mhz": 350
    }
  }
]
```

**Home Assistant Configuration:**
```yaml
# GPU Monitoring Sensors
sensor:
  - platform: rest
    name: "GPU Data"
    resource: "http://your-unraid-ip:34600/api/v1/system/gpu"
    json_attributes:
      - gpus
    value_template: "{{ value_json | length }}"

template:
  - sensor:
      - name: "GPU Temperature"
        unit_of_measurement: "°C"
        device_class: "temperature"
        state_class: "measurement"
        state: >
          {% set gpus = state_attr('sensor.gpu_data', 'gpus') %}
          {% if gpus and gpus|length > 0 %}
            {{ gpus[0].temperature }}
          {% endif %}

      - name: "GPU Power Draw"
        unit_of_measurement: "W"
        device_class: "power"
        state_class: "measurement"
        state: >
          {% set gpus = state_attr('sensor.gpu_data', 'gpus') %}
          {% if gpus and gpus|length > 0 %}
            {{ gpus[0].power.draw_watts }}
          {% endif %}

      - name: "GPU Utilization"
        unit_of_measurement: "%"
        device_class: "power_factor"
        state_class: "measurement"
        state: >
          {% set gpus = state_attr('sensor.gpu_data', 'gpus') %}
          {% if gpus and gpus|length > 0 %}
            {{ gpus[0].utilization.gpu }}
          {% endif %}
```

## Network Monitoring Integration

### Network Interface Statistics

**API Endpoint:** `/api/v1/network/interfaces`

**Home Assistant Configuration:**
```yaml
# Network Interface Monitoring
sensor:
  - platform: rest
    name: "Network Interfaces Data"
    resource: "http://your-unraid-ip:34600/api/v1/network/interfaces"
    json_attributes:
      - interfaces
    value_template: "{{ value_json.interfaces | length }}"

template:
  - sensor:
      - name: "Ethernet Speed"
        unit_of_measurement: "Mbps"
        device_class: "data_rate"
        state_class: "measurement"
        state: >
          {% set interfaces = state_attr('sensor.network_interfaces_data', 'interfaces') %}
          {% if interfaces %}
            {% for interface in interfaces %}
              {% if interface.name == 'eth0' %}
                {{ interface.speed }}
              {% endif %}
            {% endfor %}
          {% endif %}

      - name: "Network RX Rate"
        unit_of_measurement: "MB/s"
        device_class: "data_rate"
        state_class: "measurement"
        state: >
          {% set interfaces = state_attr('sensor.network_interfaces_data', 'interfaces') %}
          {% if interfaces %}
            {% for interface in interfaces %}
              {% if interface.name == 'eth0' %}
                {{ (interface.rx_bytes_per_sec / 1048576) | round(2) }}
              {% endif %}
            {% endfor %}
          {% endif %}

      - name: "Network TX Rate"
        unit_of_measurement: "MB/s"
        device_class: "data_rate"
        state_class: "measurement"
        state: >
          {% set interfaces = state_attr('sensor.network_interfaces_data', 'interfaces') %}
          {% if interfaces %}
            {% for interface in interfaces %}
              {% if interface.name == 'eth0' %}
                {{ (interface.tx_bytes_per_sec / 1048576) | round(2) }}
              {% endif %}
            {% endfor %}
          {% endif %}
```

## VM Management Integration

### Virtual Machine Control

**API Endpoint:** `/api/v1/vms`

**Home Assistant Configuration:**
```yaml
# VM Data Collection
sensor:
  - platform: rest
    name: "VMs Data"
    resource: "http://your-unraid-ip:34600/api/v1/vms"
    json_attributes:
      - vms
    value_template: "{{ value_json | length }}"

# VM Status and Performance (Template sensors from REST data)
template:
  - binary_sensor:
      - name: "VM1 Running"
        device_class: "connectivity"
        state: >
          {% set vms = state_attr('sensor.vms_data', 'vms') %}
          {% if vms and vms|length > 0 %}
            {{ vms[0].state == 'running' }}
          {% else %}
            false
          {% endif %}

  - sensor:
      - name: "VM1 Memory Usage"
        unit_of_measurement: "GB"
        device_class: "data_size"
        state_class: "measurement"
        state: >
          {% set vms = state_attr('sensor.vms_data', 'vms') %}
          {% if vms and vms|length > 0 %}
            {{ (vms[0].used_memory / 1073741824) | round(1) }}
          {% endif %}

      - name: "VM1 CPU Time"
        unit_of_measurement: "s"
        device_class: "duration"
        state_class: "total_increasing"
        state: >
          {% set vms = state_attr('sensor.vms_data', 'vms') %}
          {% if vms and vms|length > 0 %}
            {{ vms[0].cpu_time }}
          {% endif %}

      - name: "VM1 Name"
        state: >
          {% set vms = state_attr('sensor.vms_data', 'vms') %}
          {% if vms and vms|length > 0 %}
            {{ vms[0].name }}
          {% endif %}
```

## WebSocket Real-time Integration

### WebSocket Event Streaming

**WebSocket Endpoint:** `ws://your-unraid-ip:34600/api/v1/ws`

**Python Integration Example:**
```python
# custom_components/unraid_uma/websocket_client.py
import asyncio
import json
import websockets
from homeassistant.core import HomeAssistant
from homeassistant.helpers.event import async_track_time_interval

class UMAWebSocketClient:
    def __init__(self, hass: HomeAssistant, host: str, port: int):
        self.hass = hass
        self.url = f"ws://{host}:{port}/api/v1/ws"
        self.websocket = None

    async def connect(self):
        try:
            self.websocket = await websockets.connect(self.url)

            # Subscribe to storage events
            await self.websocket.send(json.dumps({
                "action": "subscribe",
                "channel": "storage"
            }))

            # Subscribe to UPS events
            await self.websocket.send(json.dumps({
                "action": "subscribe",
                "channel": "ups"
            }))

            # Listen for events
            async for message in self.websocket:
                data = json.loads(message)
                await self.handle_event(data)

        except Exception as e:
            print(f"WebSocket error: {e}")

    async def handle_event(self, event_data):
        if event_data.get("channel") == "storage":
            # Update storage sensors
            self.hass.bus.async_fire("unraid_storage_update", event_data)

        elif event_data.get("channel") == "ups":
            # Update UPS sensors
            self.hass.bus.async_fire("unraid_ups_update", event_data)
```

## MCP Integration for AI Assistants

### MCP Server Status

**API Endpoint:** `/api/v1/mcp/status`

**Home Assistant Configuration:**
```yaml
# MCP Server Monitoring
sensor:
  - platform: rest
    name: "MCP Server Status"
    resource: "http://your-unraid-ip:34600/api/v1/mcp/status"
    value_template: "{{ value_json.data.status }}"
    json_attributes:
      - active_connections
      - total_tools
      - max_connections

binary_sensor:
  - platform: rest
    name: "MCP Server Online"
    resource: "http://your-unraid-ip:34600/api/v1/mcp/status"
    value_template: "{{ value_json.data.enabled and value_json.data.status == 'running' }}"
    device_class: "connectivity"
```

## Complete Integration Example

### Full configuration.yaml Example

```yaml
# Complete UMA Integration for Home Assistant
# Place this in your configuration.yaml file

# REST Platform Configuration
rest:
  - resource: "http://192.168.20.21:34600/api/v1"
    scan_interval: 30
    headers:
      X-Request-ID: "home-assistant"
      Accept: "application/json"

# Storage Monitoring
sensor:
  # Array Usage
  - platform: rest
    name: "Unraid Array Usage"
    resource: "http://192.168.20.21:34600/api/v1/storage/array"
    value_template: "{{ value_json.usage_percent | round(1) }}"
    unit_of_measurement: "%"
    device_class: "data_size"
    state_class: "measurement"

  # UPS Power Monitoring
  - platform: rest
    name: "UPS Power Consumption"
    resource: "http://192.168.20.21:34600/api/v1/system/ups"
    value_template: "{{ value_json.power_consumption | round(1) }}"
    unit_of_measurement: "W"
    device_class: "power"
    state_class: "measurement"

  - platform: rest
    name: "UPS Battery Level"
    resource: "http://192.168.20.21:34600/api/v1/system/ups"
    value_template: "{{ value_json.battery_charge }}"
    unit_of_measurement: "%"
    device_class: "battery"
    state_class: "measurement"

# System Health
binary_sensor:
  - platform: rest
    name: "Unraid System Health"
    resource: "http://192.168.20.21:34600/api/v1/health"
    value_template: "{{ value_json.status == 'healthy' }}"
    device_class: "connectivity"

  - platform: rest
    name: "UPS Online"
    resource: "http://192.168.20.21:34600/api/v1/system/ups"
    value_template: "{{ value_json.status == 'online' }}"
    device_class: "power"

# Container Control
switch:
  - platform: rest
    name: "Plex Container"
    resource: "http://192.168.20.21:34600/api/v1/docker/containers/plex"
    body_on: '{"action": "start"}'
    body_off: '{"action": "stop"}'
    is_on_template: "{{ value_json.status == 'running' }}"
    headers:
      Content-Type: "application/json"
```

## Best Practices

### Update Intervals
- **Storage/UPS**: 30-60 seconds (data changes slowly)
- **Container Performance**: 15-30 seconds (moderate changes)
- **Network Statistics**: 10-15 seconds (frequent changes)
- **System Health**: 60-120 seconds (infrequent changes)

### Error Handling
```yaml
# Add availability templates for robust integration
sensor:
  - platform: rest
    name: "Unraid Array Usage"
    resource: "http://your-unraid-ip:34600/api/v1/storage/array"
    value_template: >
      {% if value_json.usage_percent is defined %}
        {{ value_json.usage_percent | round(1) }}
      {% else %}
        unavailable
      {% endif %}
    availability: >
      {{ value_json.usage_percent is defined }}
```

### Device Organization
```yaml
# Group entities by device for better organization
device_tracker:
  - platform: unraid_uma
    devices:
      unraid_server:
        name: "Unraid Server"
        entities:
          - sensor.unraid_array_usage
          - sensor.ups_power_consumption
          - binary_sensor.unraid_system_health
```

This comprehensive guide provides everything needed to build a complete Home Assistant integration with UMA, using real production data and proven configuration patterns.
```
