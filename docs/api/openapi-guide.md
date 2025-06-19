# OpenAPI/Swagger Documentation Guide

UMA provides comprehensive API documentation through OpenAPI 3.0.3 specification with an interactive Swagger UI interface.

## Accessing the Documentation

### Swagger UI (Interactive)
The interactive Swagger UI allows you to explore and test the API directly from your browser:

```
http://your-unraid-ip:34600/api/v1/docs
```

**Features:**
- Interactive API exploration
- Try-it-out functionality for all endpoints
- Request/response examples
- Schema documentation
- Authentication testing (when applicable)

### OpenAPI Specification (JSON)
The machine-readable OpenAPI specification is available at:

```
http://your-unraid-ip:34600/api/v1/openapi.json
```

**Use cases:**
- Code generation for client libraries
- API testing tool integration
- Documentation generation
- API validation and linting

## Using Swagger UI

### 1. Navigate to the Interface
Open your browser and go to:
```
http://your-unraid-ip:34600/api/v1/docs
```

### 2. Explore Endpoints
The interface organizes endpoints by category:
- **System** - Health checks, stats, sensors
- **Storage** - Disk information and management
- **Docker** - Container management and bulk operations
- **WebSocket** - Real-time monitoring endpoints

### 3. Test Endpoints
For each endpoint, you can:

1. **Click "Try it out"** to enable the testing interface
2. **Fill in parameters** (path, query, or body parameters)
3. **Add headers** like `X-Request-ID` for tracking
4. **Execute** the request to see live results

### 4. View Responses
The interface shows:
- **Response body** with actual data
- **Response headers** including custom headers
- **HTTP status code**
- **Response time**

## Example: Testing the Health Endpoint

1. Navigate to the **System** section
2. Find the `GET /api/v1/health` endpoint
3. Click **"Try it out"**
4. Optionally add a custom request ID in headers:
   ```
   X-Request-ID: swagger-test-123
   ```
5. Click **"Execute"**
6. View the response:
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

## Example: Testing Bulk Operations

1. Navigate to the **Docker** section
2. Find the `POST /api/v1/docker/containers/bulk/start` endpoint
3. Click **"Try it out"**
4. Fill in the request body:
   ```json
   {
     "container_ids": ["plex", "nginx"]
   }
   ```
5. Add headers:
   ```
   X-Request-ID: bulk-test-456
   Content-Type: application/json
   ```
6. Click **"Execute"**
7. View the bulk operation results

## Schema Documentation

The Swagger UI provides detailed schema documentation for:

### Request Bodies
- Required fields
- Field types and formats
- Validation rules
- Example values

### Response Objects
- Response structure
- Field descriptions
- Data types
- Nested object schemas

### Error Responses
- Error message formats
- Status code meanings
- Error handling examples

## Advanced Features

### Custom Headers
Test custom headers like:
- `X-Request-ID` - Request tracking
- `Accept` - API versioning (`application/vnd.uma.v1+json`)
- `Accept-Encoding` - Response compression (`gzip`)

### Pagination Testing
For paginated endpoints, test different parameters:
- `page` - Page number (default: 1)
- `limit` - Items per page (default: 10, max: 1000)

Example:
```
GET /api/v1/docker/containers?page=2&limit=5
```

### WebSocket Documentation
WebSocket endpoints are documented with:
- Connection URLs
- Message formats
- Event types
- Usage examples

## Integration with Development Tools

### Postman Integration
Import the OpenAPI spec into Postman:
1. Open Postman
2. Click **Import**
3. Enter the OpenAPI URL: `http://your-unraid-ip:34600/api/v1/openapi.json`
4. Postman will create a collection with all endpoints

### Code Generation
Generate client libraries using the OpenAPI spec:

```bash
# Generate Python client
openapi-generator generate -i http://your-unraid-ip:34600/api/v1/openapi.json \
                          -g python \
                          -o ./uma-python-client

# Generate JavaScript client
openapi-generator generate -i http://your-unraid-ip:34600/api/v1/openapi.json \
                          -g javascript \
                          -o ./uma-js-client
```

### API Testing Tools
Use the spec with testing tools:
- **Insomnia** - Import OpenAPI spec
- **Thunder Client** (VS Code) - Import collection
- **curl** - Generate curl commands from Swagger UI

## Troubleshooting

### Common Issues

**Swagger UI not loading:**
- Verify UMA is running: `curl http://your-unraid-ip:34600/api/v1/health`
- Check the correct port (default: 34600)
- Ensure no firewall blocking access

**CORS errors in browser:**
- UMA includes CORS headers for browser access
- Try accessing from the same network as the Unraid server

**OpenAPI spec not accessible:**
- Verify the endpoint: `curl http://your-unraid-ip:34600/api/v1/openapi.json`
- Check UMA logs for any errors

### Getting Help

If you encounter issues:
1. Check the [troubleshooting guide](../deployment/troubleshooting.md)
2. Review UMA logs for errors
3. Open an issue on [GitHub](https://github.com/domalab/uma/issues)

## Next Steps

- **[Complete Endpoint Reference](endpoints.md)** - Detailed endpoint documentation
- **[WebSocket Guide](websockets.md)** - Real-time monitoring setup
- **[Bulk Operations](bulk-operations.md)** - Efficient container management
