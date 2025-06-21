# UMA Development Setup Guide

This guide covers setting up the development environment for the UMA (Unraid Management Agent) project, including schema validation tools and development workflows.

## Prerequisites

- **Go 1.24+** - For building the UMA binary and schema validator
- **Node.js 18+** - For OpenAPI validation tools
- **npm** - For installing OpenAPI validation dependencies
- **Git** - For version control
- **Make** - For build automation

## Quick Start

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd uma
   ```

2. **Install OpenAPI validation tools**:
   ```bash
   make install-schema-tools
   ```

3. **Build the schema validator**:
   ```bash
   make build-validator
   ```

4. **Run schema validation**:
   ```bash
   make validate-schemas
   ```

## OpenAPI Validation Tools

### Modern Tool Stack

The project uses modern, actively maintained OpenAPI validation tools:

- **@redocly/cli** - Modern OpenAPI linting and validation
- **@pb33f/openapi-changes** - Modern OpenAPI specification comparison (zero deprecated direct dependencies)

### Installation

```bash
# Install all required tools
make install-schema-tools

# Manual installation (if needed)
npm install -g @redocly/cli @pb33f/openapi-changes
```

### Deprecation Warnings

During installation, you may see a few deprecation warnings from transitive dependencies (dependencies of our dependencies). These are **not from packages we directly control** and are acceptable:

- **inflight@1.0.6** - From @redocly/cli → glob@7.2.3 → inflight@1.0.6
- **glob@7.2.3** - From @redocly/cli → glob@7.2.3
- **node-domexception@1.0.0** - From @pb33f/openapi-changes → node-fetch → fetch-blob

These warnings don't affect functionality and will be resolved when the upstream packages update their dependencies. We've eliminated all deprecation warnings from our **direct dependencies**.

### Available Commands

After installation, these commands are available:

```bash
# Validate and lint OpenAPI specification
openapi lint <spec-file>

# Alternative command
redocly lint <spec-file>

# Compare two OpenAPI specifications
openapi-changes summary <old-spec> <new-spec>

# Check version
openapi --version
openapi-changes --version
```

## Development Workflow

### Schema Validation Targets

```bash
# Show help for schema validation
make schema-help

# Install OpenAPI validation tools
make install-schema-tools

# Build the schema validator
make build-validator

# Run validation against live Unraid server (default: 192.168.20.21:34600)
make validate-schemas

# Run validation against custom server
UMA_HOST=localhost UMA_PORT=8080 make validate-schemas-remote
```

### Local Development

1. **Start UMA locally**:
   ```bash
   make serve
   ```

2. **Run validation against local instance**:
   ```bash
   UMA_HOST=localhost make validate-schemas-remote
   ```

3. **Build and test**:
   ```bash
   make build
   make test
   ```

## CI/CD Integration

### GitHub Actions

The project includes automated schema validation in GitHub Actions:

- **Pull Request Validation** - Validates schemas on every PR
- **Push Validation** - Validates on pushes to main/develop
- **Manual Dispatch** - Allows manual validation runs
- **Scheduled Drift Detection** - Daily schema drift monitoring

### Workflow Files

- `.github/workflows/schema-validation.yml` - Main validation workflow
- `.github/workflows/code-quality.yml` - Code quality checks

## Troubleshooting

### Common Issues

1. **Command not found errors**:
   ```bash
   # Ensure tools are installed globally
   npm install -g @redocly/cli @pb33f/openapi-changes
   
   # Check PATH includes npm global bin
   npm config get prefix
   ```

2. **Permission errors**:
   ```bash
   # Use sudo if needed (not recommended)
   sudo npm install -g @redocly/cli @pb33f/openapi-changes
   
   # Or configure npm to use a different directory
   npm config set prefix ~/.npm-global
   export PATH=~/.npm-global/bin:$PATH
   ```

3. **Validation failures**:
   - OpenAPI linting failures are expected during development
   - Focus on critical schema mismatches between API and documentation
   - Use `openapi lint --max-problems 50` to see more issues

### PATH Configuration

If globally installed npm packages aren't available:

```bash
# Check npm global bin directory
npm config get prefix

# Add to your shell profile (.bashrc, .zshrc, etc.)
export PATH="$(npm config get prefix)/bin:$PATH"

# Reload shell configuration
source ~/.bashrc  # or ~/.zshrc
```

## Schema Validation Process

### Validation Steps

1. **OpenAPI Syntax Validation** - Validates spec syntax and structure
2. **Endpoint Availability Testing** - Tests all documented endpoints
3. **Schema-Response Validation** - Compares API responses to schemas
4. **Coverage Analysis** - Reports validation coverage statistics
5. **Drift Detection** - Compares against baseline specifications

### Reports

Validation generates detailed reports in the `reports/` directory:

- `openapi_spec_<timestamp>.json` - Downloaded OpenAPI specification
- `schema_validation_<timestamp>.json` - Detailed validation results
- `validation_output_<timestamp>.txt` - Console output
- `validation_summary_<timestamp>.md` - Human-readable summary

### Understanding Results

- **✅ Validated** - Endpoint passes all schema checks
- **❌ Validation errors** - Schema mismatches or endpoint failures
- **⚠️ Warnings** - Non-critical issues or missing optional fields
- **Coverage %** - Percentage of documented endpoints successfully validated

## Best Practices

### Development

1. **Run validation before commits**:
   ```bash
   make validate-schemas
   ```

2. **Update schemas when changing API responses**
3. **Test against live Unraid server for accuracy**
4. **Review validation reports for schema drift**

### Schema Maintenance

1. **Keep schemas synchronized with API implementations**
2. **Use real hardware data in examples**
3. **Document all required fields accurately**
4. **Validate breaking changes before release**

## Additional Resources

- [OpenAPI Specification](https://spec.openapis.org/oas/v3.1.0)
- [Redocly CLI Documentation](https://redocly.com/docs/cli/)
- [UMA API Documentation](../README.md)
- [Testing Guide](./testing.md)
