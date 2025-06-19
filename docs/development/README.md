# UMA Development Guide

This guide covers development setup, building from source, testing, and contribution guidelines for UMA (Unraid Management Agent).

## Development Environment Setup

### Prerequisites

- **Go 1.21 or later** - [Download Go](https://golang.org/dl/)
- **Git** - Version control
- **Make** - Build automation (optional)
- **Docker** - For testing container operations (optional)

### Getting Started

1. **Clone the repository**
   ```bash
   git clone https://github.com/domalab/uma.git
   cd uma
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Verify setup**
   ```bash
   go version
   go mod verify
   ```

## Building from Source

### Development Build

For local development and testing:

```bash
# Build for current platform
go build -o uma .

# Run with debug logging
./uma boot --http-port 34600
```

### Production Build

For deployment to Unraid servers:

```bash
# Build optimized binary for Linux x86_64
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o uma .

# Verify binary size (should be under 20MB)
ls -lh uma
```

### Build with Version Information

```bash
# Build with version and build info
VERSION=$(git describe --tags --always --dirty)
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse HEAD)

go build -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" -o uma .
```

### Cross-Platform Builds

```bash
# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o uma-linux-amd64 .
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o uma-linux-arm64 .
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o uma-darwin-amd64 .
```

## Project Structure

```
uma/
├── daemon/                 # Core application code
│   ├── cmd/               # Command-line interface
│   ├── common/            # Shared utilities
│   ├── domain/            # Domain models and types
│   ├── dto/               # Data transfer objects
│   ├── logger/            # Logging utilities
│   ├── plugins/           # Hardware monitoring plugins
│   │   ├── docker/        # Docker management
│   │   ├── gpu/           # GPU monitoring
│   │   ├── sensor/        # Hardware sensors
│   │   ├── storage/       # Storage monitoring
│   │   ├── system/        # System monitoring
│   │   ├── ups/           # UPS monitoring
│   │   └── vm/            # VM management
│   └── services/          # Core services
│       ├── api/           # HTTP API server
│       │   ├── handlers/  # Request handlers
│       │   ├── middleware/# HTTP middleware
│       │   ├── routes/    # Route definitions
│       │   └── schemas/   # OpenAPI schemas
│       ├── auth/          # Authentication
│       ├── cache/         # Caching system
│       ├── config/        # Configuration management
│       └── async/         # Async operations
├── docs/                  # Documentation
│   ├── api/              # API documentation
│   ├── deployment/       # Installation guides
│   └── development/      # Development guides
├── scripts/              # Build and deployment scripts
├── tests/                # Test files
├── uma.go                # Main application entry point
├── uma.plg               # Unraid plugin definition
└── go.mod                # Go module definition
```

## Code Style and Standards

### Go Style Guide

UMA follows the [Google Go Style Guide](https://google.github.io/styleguide/go/) with these specific conventions:

- **Package naming**: Use lowercase, single words when possible
- **Function naming**: Use camelCase for private, PascalCase for public
- **Error handling**: Always handle errors explicitly
- **Comments**: Use complete sentences for package and function comments
- **File organization**: Keep files under 500 lines when possible

### Code Quality Tools

Run these tools before submitting code:

```bash
# Format code
go fmt ./...

# Vet code for common issues
go vet ./...

# Run static analysis
staticcheck ./...

# Run linter (install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
golangci-lint run

# Security analysis (install: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest)
gosec ./...
```

### Pre-commit Hooks

Set up pre-commit hooks to ensure code quality:

```bash
# Create .git/hooks/pre-commit
#!/bin/bash
set -e

echo "Running pre-commit checks..."

# Format code
go fmt ./...

# Vet code
go vet ./...

# Run tests
go test ./...

# Run linter
golangci-lint run

echo "Pre-commit checks passed!"
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test package
go test ./daemon/services/api/handlers/

# Run tests with race detection
go test -race ./...

# Run tests with verbose output
go test -v ./...
```

### Test Structure

UMA uses table-driven tests and the testify framework:

```go
func TestSystemHandler_GetHealth(t *testing.T) {
    tests := []struct {
        name           string
        mockSetup      func(*mocks.MockAPIInterface)
        expectedStatus int
        expectedBody   string
    }{
        {
            name: "successful health check",
            mockSetup: func(m *mocks.MockAPIInterface) {
                m.EXPECT().GetHealth().Return(&dto.HealthResponse{
                    Status: "healthy",
                }, nil)
            },
            expectedStatus: 200,
            expectedBody:   `{"status":"healthy"}`,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Test Coverage Goals

- **Overall coverage**: >80%
- **Critical paths**: >95% (health checks, API handlers, storage operations)
- **New code**: 100% coverage required

### Testing on Unraid

For testing Unraid-specific functionality:

```bash
# Deploy to test Unraid server
make clean
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o uma .
scp uma root@192.168.20.21:/tmp/
ssh root@192.168.20.21 "killall uma 2>/dev/null; /tmp/uma boot --http-port 34600"
```

## API Development

### Adding New Endpoints

1. **Define the handler**
   ```go
   // daemon/services/api/handlers/new_handler.go
   func (h *NewHandler) HandleNewEndpoint(w http.ResponseWriter, r *http.Request) {
       // Implementation
   }
   ```

2. **Register the route**
   ```go
   // daemon/services/api/routes/new_routes.go
   func (r *Router) registerNewRoutes() {
       r.mux.HandleFunc("/api/v1/new/endpoint", r.newHandler.HandleNewEndpoint)
   }
   ```

3. **Add OpenAPI documentation**
   ```go
   // daemon/services/api/handlers/docs.go
   func (h *DocsHandler) addNewPaths(paths map[string]interface{}) {
       paths["/new/endpoint"] = h.createGetEndpoint("New", "Description", "Details", "ResponseSchema")
   }
   ```

4. **Write tests**
   ```go
   // daemon/services/api/handlers/new_handler_test.go
   func TestNewHandler_HandleNewEndpoint(t *testing.T) {
       // Test implementation
   }
   ```

### OpenAPI Schema Management

UMA uses a centralized schema registry:

```go
// daemon/services/api/schemas/registry.go
func (r *Registry) RegisterNewSchema() {
    r.schemas["NewResponse"] = map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "field": map[string]interface{}{
                "type": "string",
                "description": "Field description",
            },
        },
    }
}
```

## Plugin Development

### Creating New Monitoring Plugins

1. **Define the plugin interface**
   ```go
   // daemon/plugins/newplugin/interface.go
   type NewPlugin interface {
       GetData() (*NewData, error)
       Start() error
       Stop() error
   }
   ```

2. **Implement the plugin**
   ```go
   // daemon/plugins/newplugin/newplugin.go
   type newPlugin struct {
       // Plugin state
   }

   func NewPlugin() NewPlugin {
       return &newPlugin{}
   }
   ```

3. **Add to API integration**
   ```go
   // daemon/services/api/api.go
   a.newPlugin = newplugin.NewPlugin()
   ```

## Deployment and Release

### Local Development Deployment

```bash
# Build and deploy to local Unraid server
make deploy-dev
```

### Production Release Process

1. **Version tagging**
   ```bash
   git tag -a v2025.06.16 -m "Release v2025.06.16"
   git push origin v2025.06.16
   ```

2. **Build release binaries**
   ```bash
   make release
   ```

3. **Update plugin file**
   ```bash
   # Update version in uma.plg
   # Update MD5 hash for new bundle
   ```

### Continuous Integration

UMA uses GitHub Actions for CI/CD:

- **Pull Request Checks**: Tests, linting, security scans
- **Release Builds**: Automated binary builds and plugin packaging
- **Dependency Updates**: Automated dependency vulnerability scanning

## Debugging

### Local Debugging

```bash
# Run with debug logging
UMA_LOGGING_LEVEL=debug ./uma boot

# Run with race detection
go run -race . boot

# Use delve debugger
dlv debug . -- boot --http-port 34600
```

### Remote Debugging

```bash
# Deploy debug build to Unraid
GOOS=linux GOARCH=amd64 go build -gcflags="all=-N -l" -o uma .
scp uma root@192.168.20.21:/tmp/

# Run with remote debugging
ssh root@192.168.20.21 "/tmp/uma boot --http-port 34600"
```

### Performance Profiling

```bash
# Enable pprof endpoint
go run . boot --enable-pprof

# Analyze CPU profile
go tool pprof http://localhost:34600/debug/pprof/profile

# Analyze memory profile
go tool pprof http://localhost:34600/debug/pprof/heap
```

## Contributing Guidelines

### Pull Request Process

1. **Fork the repository**
2. **Create a feature branch** from `main`
3. **Make your changes** following code style guidelines
4. **Add tests** for new functionality
5. **Update documentation** as needed
6. **Run all quality checks**
7. **Submit a pull request** with clear description

### Commit Message Format

```
type(scope): brief description

Detailed explanation of changes if needed.

Fixes #issue-number
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

### Code Review Checklist

- [ ] Code follows style guidelines
- [ ] Tests are included and passing
- [ ] Documentation is updated
- [ ] No security vulnerabilities introduced
- [ ] Performance impact considered
- [ ] Backward compatibility maintained

## Resources

- **[Go Documentation](https://golang.org/doc/)**
- **[Unraid Plugin Development](https://forums.unraid.net/topic/38582-plug-in-development-documentation/)**
- **[OpenAPI Specification](https://swagger.io/specification/)**
- **[Prometheus Metrics](https://prometheus.io/docs/concepts/metric_types/)**
- **[WebSocket Protocol](https://tools.ietf.org/html/rfc6455)**

## Getting Help

- **GitHub Issues**: [Report bugs and request features](https://github.com/domalab/uma/issues)
- **GitHub Discussions**: [Ask questions and share ideas](https://github.com/domalab/uma/discussions)
- **Development Chat**: Join our development discussions
- **Code Review**: Request code review from maintainers
