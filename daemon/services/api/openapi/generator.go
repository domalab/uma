package openapi

import (
	"github.com/domalab/uma/daemon/services/api/openapi/paths"
	"github.com/domalab/uma/daemon/services/api/openapi/responses"
	"github.com/domalab/uma/daemon/services/api/openapi/schemas"
)

// Generator coordinates all OpenAPI modules and generates the complete specification
type Generator struct {
	config         *Config
	schemaRegistry *schemas.Registry
}

// NewGenerator creates a new OpenAPI generator with the given configuration
func NewGenerator(config *Config) *Generator {
	if config == nil {
		config = DefaultConfig()
	}

	registry := schemas.NewRegistry()
	registry.RegisterAll()

	return &Generator{
		config:         config,
		schemaRegistry: registry,
	}
}

// Generate creates the complete OpenAPI specification
func (g *Generator) Generate() *OpenAPISpec {
	spec := &OpenAPISpec{
		OpenAPI:    "3.1.1",
		Info:       GenerateInfo(g.config),
		Servers:    g.config.GetServers(),
		Paths:      g.generatePaths(),
		Components: g.generateComponents(),
	}

	return spec
}

// generatePaths combines all path definitions from different modules
func (g *Generator) generatePaths() map[string]interface{} {
	allPaths := make(map[string]interface{})

	// Add Docker paths
	for path, definition := range paths.GetDockerPaths() {
		allPaths[path] = definition
	}

	// Add System paths
	for path, definition := range paths.GetSystemPaths() {
		allPaths[path] = definition
	}

	// Add Storage paths (if feature enabled)
	if g.config.Features.ArrayControl {
		for path, definition := range paths.GetStoragePaths() {
			allPaths[path] = definition
		}
	}

	// Add VM paths (if feature enabled)
	if g.config.Features.VMManagement {
		for path, definition := range paths.GetVMPaths() {
			allPaths[path] = definition
		}
	}

	// Add Authentication paths (if feature enabled)
	if g.config.Features.Authentication {
		for path, definition := range paths.GetAuthPaths() {
			allPaths[path] = definition
		}
	}

	// Add WebSocket paths (if feature enabled)
	if g.config.Features.WebSockets {
		for path, definition := range paths.GetWebSocketPaths() {
			allPaths[path] = definition
		}
	}

	// Add async operations paths
	for path, definition := range paths.AsyncOperationsPaths() {
		allPaths[path] = definition
	}

	// Add rate limiting paths
	for path, definition := range paths.RateLimitingPaths() {
		allPaths[path] = definition
	}

	// Add health check and documentation paths
	allPaths["/api/v1/health"] = g.getHealthPath()
	allPaths["/api/v1/docs"] = g.getDocsPath()
	allPaths["/api/v1/openapi.json"] = g.getOpenAPIPath()

	return allPaths
}

// generateComponents combines all component definitions
func (g *Generator) generateComponents() OpenAPIComponents {
	return OpenAPIComponents{
		Schemas:         g.schemaRegistry.GetAllSchemas(),
		Responses:       responses.GetCommonResponses(),
		SecuritySchemes: g.config.GetSecuritySchemes(),
	}
}

// getHealthPath returns the health check endpoint definition
func (g *Generator) getHealthPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Health check",
			"description": "Check the health status of the UMA API service and its dependencies",
			"operationId": "healthCheck",
			"tags":        []string{"Monitoring"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Service is healthy",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/HealthResponse",
							},
						},
					},
				},
				"503": map[string]interface{}{
					"description": "Service is unhealthy",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/HealthResponse",
							},
						},
					},
				},
			},
		},
	}
}

// getDocsPath returns the Swagger UI documentation endpoint
func (g *Generator) getDocsPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "API Documentation",
			"description": "Interactive Swagger UI documentation for the UMA REST API",
			"operationId": "getDocumentation",
			"tags":        []string{"Documentation"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Swagger UI HTML page",
					"content": map[string]interface{}{
						"text/html": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
			},
		},
	}
}

// getOpenAPIPath returns the OpenAPI specification endpoint
func (g *Generator) getOpenAPIPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "OpenAPI Specification",
			"description": "Get the complete OpenAPI 3.1.1 specification for the UMA REST API",
			"operationId": "getOpenAPISpec",
			"tags":        []string{"Documentation"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "OpenAPI specification in JSON format",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "object",
							},
						},
					},
				},
			},
		},
	}
}

// GetSchemaRegistry returns the schema registry for external access
func (g *Generator) GetSchemaRegistry() *schemas.Registry {
	return g.schemaRegistry
}

// GetConfig returns the generator configuration
func (g *Generator) GetConfig() *Config {
	return g.config
}

// UpdateConfig updates the generator configuration
func (g *Generator) UpdateConfig(config *Config) {
	g.config = config
}

// ValidateSpec performs basic validation on the generated specification
func (g *Generator) ValidateSpec() []string {
	var errors []string
	spec := g.Generate()

	// Check required fields
	if spec.OpenAPI == "" {
		errors = append(errors, "OpenAPI version is required")
	}

	if spec.Info.Title == "" {
		errors = append(errors, "API title is required")
	}

	if spec.Info.Version == "" {
		errors = append(errors, "API version is required")
	}

	if len(spec.Paths) == 0 {
		errors = append(errors, "At least one path is required")
	}

	// Check for required schemas
	requiredSchemas := []string{"Error", "SuccessResponse", "StandardResponse"}
	for _, schemaName := range requiredSchemas {
		if !g.schemaRegistry.HasSchema(schemaName) {
			errors = append(errors, "Required schema missing: "+schemaName)
		}
	}

	return errors
}

// GetStats returns statistics about the generated specification
func (g *Generator) GetStats() map[string]interface{} {
	spec := g.Generate()
	schemasByCategory := g.schemaRegistry.GetSchemasByCategory()

	stats := map[string]interface{}{
		"openapi_version": spec.OpenAPI,
		"api_version":     spec.Info.Version,
		"total_paths":     len(spec.Paths),
		"total_schemas":   len(spec.Components.Schemas),
		"total_responses": len(spec.Components.Responses),
		"schemas_by_category": map[string]interface{}{
			"common":    len(schemasByCategory["Common"]),
			"docker":    len(schemasByCategory["Docker"]),
			"system":    len(schemasByCategory["System"]),
			"storage":   len(schemasByCategory["Storage"]),
			"vm":        len(schemasByCategory["VM"]),
			"websocket": len(schemasByCategory["WebSocket"]),
			"auth":      len(schemasByCategory["Auth"]),
		},
		"features_enabled": map[string]interface{}{
			"authentication":  g.config.Features.Authentication,
			"bulk_operations": g.config.Features.BulkOperations,
			"websockets":      g.config.Features.WebSockets,
			"metrics":         g.config.Features.Metrics,
			"zfs":             g.config.Features.ZFS,
			"array_control":   g.config.Features.ArrayControl,
			"vm_management":   g.config.Features.VMManagement,
		},
	}

	return stats
}
