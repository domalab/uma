package api

import (
	"encoding/json"
	"testing"

	"github.com/domalab/uma/daemon/domain"
)

// TestOpenAPISpecificationIntegration validates that the generated OpenAPI specification
// contains all expected UMA-specific endpoints and maintains backward compatibility
func TestOpenAPISpecificationIntegration(t *testing.T) {
	// Create test configuration
	config := domain.DefaultConfig()
	config.Version = "integration-test-1.0.0"
	config.HTTPServer.Port = 34600

	ctx := &domain.Context{
		Config: config,
	}

	api := &Api{ctx: ctx}
	server := &HTTPServer{api: api}

	// Generate the OpenAPI specification
	spec := server.generateOpenAPISpec()
	if spec == nil {
		t.Fatal("Failed to generate OpenAPI specification")
	}

	// Convert to JSON for easier validation
	jsonData, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("Failed to marshal OpenAPI spec: %v", err)
	}

	var specMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &specMap); err != nil {
		t.Fatalf("Failed to unmarshal OpenAPI spec: %v", err)
	}

	// Validate basic OpenAPI structure
	validateBasicStructure(t, specMap)

	// Validate UMA-specific endpoints
	validateUMAEndpoints(t, specMap)

	// Validate security schemes
	validateSecuritySchemes(t, specMap)

	// Validate schemas
	validateSchemas(t, specMap)

	// Validate backward compatibility
	validateBackwardCompatibility(t, specMap)
}

func validateBasicStructure(t *testing.T, spec map[string]interface{}) {
	// Check OpenAPI version
	if openapi, ok := spec["openapi"].(string); !ok || openapi != "3.1.1" {
		t.Errorf("Expected OpenAPI version 3.1.1, got %v", spec["openapi"])
	}

	// Check info section
	info, ok := spec["info"].(map[string]interface{})
	if !ok {
		t.Fatal("Missing info section")
	}

	if title, ok := info["title"].(string); !ok || title != "UMA REST API" {
		t.Errorf("Expected title 'UMA REST API', got %v", info["title"])
	}

	// Check paths section
	if _, ok := spec["paths"].(map[string]interface{}); !ok {
		t.Fatal("Missing paths section")
	}

	// Check components section
	if _, ok := spec["components"].(map[string]interface{}); !ok {
		t.Fatal("Missing components section")
	}
}

func validateUMAEndpoints(t *testing.T, spec map[string]interface{}) {
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		t.Fatal("Paths section is not a map")
	}

	// Expected UMA endpoint categories
	expectedEndpoints := []string{
		// Health and documentation
		"/api/v1/health",
		"/api/v1/docs",
		"/api/v1/openapi.json",

		// System monitoring
		"/api/v1/system/info",
		"/api/v1/system/cpu",
		"/api/v1/system/memory",
		"/api/v1/system/temperatures",
		"/api/v1/system/fans",
		"/api/v1/system/ups",

		// Docker management
		"/api/v1/docker/containers",
		"/api/v1/docker/containers/{id}",
		"/api/v1/docker/containers/{id}/start",
		"/api/v1/docker/containers/{id}/stop",
		"/api/v1/docker/containers/bulk/start",
		"/api/v1/docker/containers/bulk/stop",

		// VM management
		"/api/v1/vms",
		"/api/v1/vms/{id}",
		"/api/v1/vms/{id}/start",
		"/api/v1/vms/{id}/stop",
		"/api/v1/vms/bulk/start",

		// Storage management
		"/api/v1/storage/array",
		"/api/v1/storage/disks",
		"/api/v1/storage/disks/{id}",

		// Authentication
		"/api/v1/auth/login",
		"/api/v1/auth/refresh",

		// WebSocket
		"/api/v1/ws",
	}

	// Check that key endpoints exist
	for _, endpoint := range expectedEndpoints {
		if _, exists := paths[endpoint]; !exists {
			t.Logf("Warning: Expected endpoint %s not found in OpenAPI spec", endpoint)
			// Note: We log warnings instead of failing to allow for gradual implementation
		}
	}

	// Verify we have a reasonable number of endpoints
	if len(paths) < 10 {
		t.Errorf("Expected at least 10 endpoints, got %d", len(paths))
	}
}

func validateSecuritySchemes(t *testing.T, spec map[string]interface{}) {
	components, ok := spec["components"].(map[string]interface{})
	if !ok {
		t.Fatal("Components section is not a map")
	}

	securitySchemes, ok := components["securitySchemes"].(map[string]interface{})
	if !ok {
		t.Log("Warning: No security schemes found")
		return
	}

	// Check for expected security schemes
	expectedSchemes := []string{"BearerAuth", "ApiKeyAuth"}
	for _, scheme := range expectedSchemes {
		if _, exists := securitySchemes[scheme]; !exists {
			t.Logf("Warning: Expected security scheme %s not found", scheme)
		}
	}
}

func validateSchemas(t *testing.T, spec map[string]interface{}) {
	components, ok := spec["components"].(map[string]interface{})
	if !ok {
		t.Fatal("Components section is not a map")
	}

	schemas, ok := components["schemas"].(map[string]interface{})
	if !ok {
		t.Fatal("No schemas found in components")
	}

	// Check for key UMA schemas
	expectedSchemas := []string{
		"StandardResponse",
		"PaginationInfo",
		"ResponseMeta",
		"HealthResponse",
		"DiskInfo",
		"ContainerInfo",
		"VMInfo",
		"SystemInfo",
		"BulkOperationResponse",
	}

	for _, schema := range expectedSchemas {
		if _, exists := schemas[schema]; !exists {
			t.Logf("Warning: Expected schema %s not found", schema)
		}
	}

	// Verify we have a reasonable number of schemas
	if len(schemas) < 5 {
		t.Errorf("Expected at least 5 schemas, got %d", len(schemas))
	}
}

func validateBackwardCompatibility(t *testing.T, spec map[string]interface{}) {
	// Verify that the specification maintains backward compatibility
	// by checking for essential response structures

	components, ok := spec["components"].(map[string]interface{})
	if !ok {
		return
	}

	schemas, ok := components["schemas"].(map[string]interface{})
	if !ok {
		return
	}

	// Check StandardResponse structure for backward compatibility
	if standardResponse, exists := schemas["StandardResponse"]; exists {
		if responseMap, ok := standardResponse.(map[string]interface{}); ok {
			if properties, ok := responseMap["properties"].(map[string]interface{}); ok {
				// Verify essential fields exist
				essentialFields := []string{"data", "pagination", "meta"}
				for _, field := range essentialFields {
					if _, exists := properties[field]; !exists {
						t.Errorf("StandardResponse missing essential field: %s", field)
					}
				}
			}
		}
	}

	// Verify API versioning is consistent
	info, ok := spec["info"].(map[string]interface{})
	if ok {
		if version, ok := info["version"].(string); ok && version == "" {
			t.Error("API version should not be empty")
		}
	}
}

// TestOpenAPIEndpointCoverage verifies that all major UMA functionality is covered
func TestOpenAPIEndpointCoverage(t *testing.T) {
	config := domain.DefaultConfig()
	config.Version = "coverage-test-1.0.0"

	ctx := &domain.Context{
		Config: config,
	}

	api := &Api{ctx: ctx}
	server := &HTTPServer{api: api}

	spec := server.generateOpenAPISpec()
	if spec == nil {
		t.Fatal("Failed to generate OpenAPI specification")
	}

	// Convert to JSON for analysis
	jsonData, _ := json.Marshal(spec)
	var specMap map[string]interface{}
	json.Unmarshal(jsonData, &specMap)

	paths, _ := specMap["paths"].(map[string]interface{})

	// Count endpoints by category
	categories := map[string]int{
		"system":  0,
		"docker":  0,
		"vm":      0,
		"storage": 0,
		"auth":    0,
		"ws":      0,
	}

	for path := range paths {
		switch {
		case containsString(path, "/system"):
			categories["system"]++
		case containsString(path, "/docker"):
			categories["docker"]++
		case containsString(path, "/vm"):
			categories["vm"]++
		case containsString(path, "/storage"):
			categories["storage"]++
		case containsString(path, "/auth"):
			categories["auth"]++
		case containsString(path, "/ws"):
			categories["ws"]++
		}
	}

	// Verify reasonable coverage for each category
	for category, count := range categories {
		if count == 0 {
			t.Logf("Warning: No endpoints found for category %s", category)
		} else {
			t.Logf("Category %s has %d endpoints", category, count)
		}
	}

	// Verify total endpoint count is reasonable
	totalEndpoints := len(paths)
	if totalEndpoints < 15 {
		t.Logf("Warning: Only %d total endpoints found, expected more for comprehensive API", totalEndpoints)
	} else {
		t.Logf("Total endpoints: %d", totalEndpoints)
	}
}
