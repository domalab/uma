# Model Context Protocol (MCP) Server

UMA includes a built-in Model Context Protocol (MCP) server that automatically exposes all REST API endpoints as MCP tools, enabling seamless integration with AI assistants and automation platforms.

## Overview

The MCP server provides:
- **50+ Auto-discovered Tools**: All REST API endpoints automatically converted to MCP tools
- **JSON-RPC 2.0 Protocol**: Standard MCP communication over WebSocket
- **Real-time Configuration**: Enable/disable and configure via REST API
- **Tool Registry**: Dynamic tool discovery from OpenAPI specification

## Configuration

### Enable MCP Server

```bash
# Via REST API
curl -X PUT http://your-unraid-ip:34600/api/v1/mcp/config \
  -H "Content-Type: application/json" \
  -d '{"enabled": true}'

# Via configuration file
{
  "mcp": {
    "enabled": true,
    "port": 34800,
    "max_connections": 100
  }
}
```

### Configuration Options

| Setting | Default | Description |
|---------|---------|-------------|
| `enabled` | `false` | Enable/disable MCP server |
| `port` | `34800` | WebSocket port (1024-65535) |
| `max_connections` | `100` | Maximum concurrent connections |

## API Endpoints

### Get MCP Status
```http
GET /api/v1/mcp/status
```

### Get MCP Configuration
```http
GET /api/v1/mcp/config
```

### Update MCP Configuration
```http
PUT /api/v1/mcp/config
Content-Type: application/json

{
  "enabled": true,
  "port": 34800,
  "max_connections": 100
}
```

### List Available Tools
```http
GET /api/v1/mcp/tools
```

### Refresh Tool Registry
```http
POST /api/v1/mcp/tools/refresh
```

## WebSocket Connection

Connect to the MCP server via WebSocket:

```
ws://your-unraid-ip:34800/mcp
```

### Example Connection (JavaScript)
```javascript
const ws = new WebSocket('ws://your-unraid-ip:34800/mcp');

ws.onopen = function() {
    // Send MCP initialize request
    ws.send(JSON.stringify({
        jsonrpc: "2.0",
        id: 1,
        method: "initialize",
        params: {
            protocolVersion: "2024-11-05",
            capabilities: {},
            clientInfo: {
                name: "UMA-Client",
                version: "1.0.0"
            }
        }
    }));
};

ws.onmessage = function(event) {
    const response = JSON.parse(event.data);
    console.log('MCP Response:', response);
};
```

## Available Tools

The MCP server automatically discovers and registers 50+ tools from the UMA REST API:

### System Monitoring Tools
- `health_check` - System health status
- `get_system_info` - System information
- `get_c_p_u_info` - CPU information
- `get_memory_info` - Memory usage
- `get_temperatures` - Temperature sensors
- `get_fans` - Fan speeds
- `get_u_p_s_info` - UPS status

### Storage Management Tools
- `get_array_info` - Unraid array status
- `list_disks` - Disk information
- `get_parity_info` - Parity disk status
- `get_parity_check` - Parity check progress
- `get_cache_info` - Cache pool information
- `list_z_f_s_pools` - ZFS pool status

### Container Management Tools
- `list_containers` - Docker containers
- `get_container` - Container details
- `get_docker_info` - Docker system info
- `list_images` - Docker images
- `list_networks` - Docker networks

### VM Management Tools
- `list_v_ms` - Virtual machines
- `get_v_m` - VM details
- `get_v_m_stats` - VM statistics

### And Many More...
All REST API endpoints are automatically converted to MCP tools with appropriate parameter mapping and response handling.

## Integration Examples

### Claude Desktop Integration
Add to your Claude Desktop configuration:

```json
{
  "mcpServers": {
    "uma": {
      "command": "npx",
      "args": ["@modelcontextprotocol/server-websocket", "ws://your-unraid-ip:34800/mcp"]
    }
  }
}
```

### Custom MCP Client
```python
import asyncio
import websockets
import json

async def mcp_client():
    uri = "ws://your-unraid-ip:34800/mcp"
    
    async with websockets.connect(uri) as websocket:
        # Initialize MCP session
        init_request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "initialize",
            "params": {
                "protocolVersion": "2024-11-05",
                "capabilities": {},
                "clientInfo": {"name": "Python-Client", "version": "1.0.0"}
            }
        }
        
        await websocket.send(json.dumps(init_request))
        response = await websocket.recv()
        print("Initialize:", json.loads(response))
        
        # List available tools
        tools_request = {
            "jsonrpc": "2.0",
            "id": 2,
            "method": "tools/list"
        }
        
        await websocket.send(json.dumps(tools_request))
        response = await websocket.recv()
        tools = json.loads(response)
        print(f"Available tools: {len(tools['result']['tools'])}")

# Run the client
asyncio.run(mcp_client())
```

## Security Considerations

- **Internal Network Only**: MCP server is designed for internal network use
- **No Authentication**: Uses network-level security (firewall/VPN)
- **WebSocket Only**: Secure WebSocket (WSS) recommended for external access
- **Tool Permissions**: All tools have read-only access by default

## Troubleshooting

### Check MCP Server Status
```bash
curl http://your-unraid-ip:34600/api/v1/mcp/status
```

### Verify WebSocket Connection
```bash
# Test WebSocket upgrade
curl -i -N \
  -H "Connection: Upgrade" \
  -H "Upgrade: websocket" \
  -H "Sec-WebSocket-Version: 13" \
  -H "Sec-WebSocket-Key: x3JJHMbDL1EzLkh9GBhXDw==" \
  http://your-unraid-ip:34800/mcp
```

### Check Tool Discovery
```bash
curl http://your-unraid-ip:34600/api/v1/mcp/tools | jq '.data | length'
```

## Support

For MCP-related issues:
1. Check UMA logs for MCP server startup messages
2. Verify WebSocket connectivity on port 34800
3. Ensure MCP is enabled in configuration
4. Review tool discovery logs for OpenAPI parsing

For more information, see the [UMA API Documentation](../api/) and [WebSocket Guide](websockets.md).
