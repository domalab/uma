# UMA API Validation Workflow

This document describes the enhanced API validation workflow for UMA development, including automated schema generation, route discovery, and comprehensive validation tools.

## Overview

The UMA API validation system provides:

- **Automated Route Discovery**: Scans codebase for all HTTP endpoints
- **Live Schema Generation**: Creates OpenAPI schemas from real API responses
- **WebSocket Documentation**: Documents real-time endpoints and channels
- **Comprehensive Validation**: Validates schemas against live API responses
- **CI/CD Integration**: Automated validation in GitHub Actions

## Quick Start

### Prerequisites

1. **Go 1.21+** installed
2. **jq** for JSON processing: `brew install jq` (macOS) or `apt install jq` (Linux)
3. **Node.js 18+** for OpenAPI tools (optional): `npm install -g @redocly/cli`

### Initial Setup

```bash
# 1. Build all validation tools
make build-tools

# 2. Run comprehensive validation
./tools/validate-schemas.sh

# 3. Review generated reports
ls -la reports/
```

## Development Workflow

### 1. Before Making API Changes

```bash
# Check current API state
./tools/validate-schemas.sh

# Review current route coverage
./tools/route-scanner/route-scanner daemon/services/api/routes --json
```

### 2. During Development

#### Adding New Endpoints

1. **Implement the handler** in `daemon/services/api/handlers/`
2. **Register the route** in `daemon/services/api/routes/`
3. **Add OpenAPI documentation** in `daemon/services/api/openapi/paths/`
4. **Add schemas** in `daemon/services/api/openapi/schemas/`

#### Modifying Existing Endpoints

1. **Update the handler** implementation
2. **Update OpenAPI documentation** if response structure changed
3. **Run validation** to ensure consistency

### 3. Before Committing

```bash
# Run pre-commit validation (automatic if hook installed)
./tools/pre-commit-api-check.sh

# Or run full validation manually
./tools/validate-schemas.sh
```

### 4. Pre-commit Hook Setup (Recommended)

```bash
# Install the pre-commit hook
cp tools/pre-commit-api-check.sh .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

# Test the hook
./tools/pre-commit-api-check.sh --force
```

## Validation Tools

### 1. Route Scanner (`tools/route-scanner/`)

**Purpose**: Discovers all HTTP routes in the codebase

```bash
# Scan all routes
./tools/route-scanner/route-scanner daemon/services/api/routes

# Output as JSON
./tools/route-scanner/route-scanner daemon/services/api/routes --json --output routes.json

# Analyze specific patterns
./tools/route-scanner/route-scanner daemon/services/api/routes --pattern "auth|docker"
```

**Output**: Comprehensive list of discovered routes with methods and handlers

### 2. Schema Generator (`tools/schema-generator/`)

**Purpose**: Generates OpenAPI schemas from live API responses

```bash
# Generate from live API (requires UMA running)
./tools/schema-generator/schema-generator http://192.168.20.21:34600

# Include route discovery
./tools/schema-generator/schema-generator http://192.168.20.21:34600 --routes

# Output to file
./tools/schema-generator/schema-generator http://192.168.20.21:34600 --output generated.json
```

**Output**: Complete OpenAPI specification based on real API responses

### 3. WebSocket Documenter (`tools/websocket-documenter/`)

**Purpose**: Documents WebSocket endpoints and channels

```bash
# Document WebSocket endpoints
./tools/websocket-documenter/websocket-documenter daemon/services/api

# Output to file
./tools/websocket-documenter/websocket-documenter daemon/services/api --output websocket.json
```

**Output**: Comprehensive WebSocket documentation with channels and message types

### 4. Schema Validator (`tools/schema-validator/`)

**Purpose**: Validates OpenAPI schemas against live API responses

```bash
# Validate against live API
./tools/schema-validator/schema-validator http://192.168.20.21:34600

# Detailed validation report
./tools/schema-validator/schema-validator http://192.168.20.21:34600 --detailed
```

**Output**: Validation report with schema compliance results

### 5. Comprehensive Validation (`tools/validate-schemas.sh`)

**Purpose**: Runs all validation tools in sequence

```bash
# Full validation with default settings
./tools/validate-schemas.sh

# Against different server
UMA_HOST=localhost UMA_PORT=8080 ./tools/validate-schemas.sh

# Clean old reports
./tools/validate-schemas.sh --clean
```

**Output**: Complete validation report with all tool results

## Understanding Reports

### Route Discovery Report

```json
{
  "total_routes": 63,
  "categories": {
    "auth": 8,
    "docker": 12,
    "system": 15,
    "storage": 10
  },
  "routes": [
    {
      "path": "/api/v1/auth/login",
      "methods": ["GET", "POST"],
      "handler": "AuthHandler.HandleLogin"
    }
  ]
}
```

### Schema Validation Report

```json
{
  "summary": {
    "total_endpoints": 60,
    "validated": 59,
    "failed": 1,
    "critical_issues": 0
  },
  "results": [
    {
      "endpoint": "/api/v1/health",
      "status": "PASS",
      "response_time": "45ms"
    }
  ]
}
```

### WebSocket Documentation

```json
{
  "channels": {
    "system.stats": {
      "description": "Real-time system performance metrics",
      "message_types": ["subscribe", "unsubscribe", "event"]
    }
  },
  "event_types": [
    "cpu_usage", "memory_usage", "temperature_alert"
  ]
}
```

## Best Practices

### 1. Documentation Standards

- **Always document new endpoints** in OpenAPI specification
- **Use consistent naming** for operations and schemas
- **Include examples** in schema definitions
- **Add proper descriptions** for all fields

### 2. Validation Workflow

- **Run validation locally** before committing
- **Check route discovery** for undocumented endpoints
- **Compare generated vs manual schemas** for accuracy
- **Review WebSocket changes** if real-time features modified

### 3. Error Handling

- **Fix critical validation issues** immediately
- **Address schema mismatches** before release
- **Update documentation** when API structure changes
- **Test endpoints manually** if validation fails

### 4. Performance Considerations

- **Route discovery** is fast (< 5 seconds)
- **Schema generation** requires live API (10-30 seconds)
- **Full validation** takes 1-2 minutes with live API
- **Use caching** for repeated validations

## Troubleshooting

### Common Issues

1. **"Route scanner not found"**
   ```bash
   cd tools/route-scanner && go build -o route-scanner main.go
   ```

2. **"UMA API not responding"**
   ```bash
   # Check if UMA is running
   curl http://192.168.20.21:34600/api/v1/health
   
   # Use different host/port
   UMA_HOST=localhost UMA_PORT=8080 ./tools/validate-schemas.sh
   ```

3. **"OpenAPI compilation failed"**
   ```bash
   # Check syntax in OpenAPI files
   cd daemon/services/api/openapi && go build -v .
   ```

4. **"Schema validation errors"**
   ```bash
   # Run detailed validation
   ./tools/schema-validator/schema-validator http://192.168.20.21:34600 --detailed
   ```

### Getting Help

- **Check logs** in `reports/validation_output_*.txt`
- **Review generated reports** in `reports/` directory
- **Run tools individually** to isolate issues
- **Use `--help` flag** on any tool for usage information

## Integration with IDEs

### VS Code

Add to `.vscode/tasks.json`:

```json
{
  "label": "Validate UMA API",
  "type": "shell",
  "command": "./tools/validate-schemas.sh",
  "group": "build",
  "presentation": {
    "echo": true,
    "reveal": "always",
    "focus": false,
    "panel": "shared"
  }
}
```

### Git Hooks

The pre-commit hook automatically runs validation for API changes. To customize:

```bash
# Edit the hook
vim .git/hooks/pre-commit

# Test manually
./tools/pre-commit-api-check.sh --force
```

## Next Steps

1. **Set up pre-commit hooks** for automatic validation
2. **Integrate with CI/CD** using provided GitHub Actions
3. **Establish documentation standards** for your team
4. **Create custom validation rules** as needed
5. **Monitor API evolution** using generated reports
