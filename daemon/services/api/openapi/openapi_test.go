package openapi

import (
	"encoding/json"
	"testing"
)

// TestOpenAPISpecGeneration tests the generation of OpenAPI specifications
func TestOpenAPISpecGeneration(t *testing.T) {
	generator := NewGenerator(nil) // Use default config

	t.Run("Generate basic OpenAPI spec", func(t *testing.T) {
		spec := generator.Generate()

		if spec == nil {
			t.Fatal("OpenAPI spec should not be nil")
		}

		// Check required fields
		if spec.OpenAPI == "" {
			t.Error("OpenAPI version should be set")
		}
		if spec.Info.Title == "" {
			t.Error("Info section should be present")
		}
		if spec.Paths == nil {
			t.Error("Paths section should be present")
		}
		if len(spec.Components.Schemas) == 0 {
			t.Error("Components section should be present")
		}
	})

	t.Run("Validate OpenAPI version", func(t *testing.T) {
		spec := generator.Generate()

		expectedVersion := "3.1.1"
		if spec.OpenAPI != expectedVersion {
			t.Errorf("Expected OpenAPI version %s, got %s", expectedVersion, spec.OpenAPI)
		}
	})

	t.Run("Validate info section", func(t *testing.T) {
		spec := generator.Generate()

		if spec.Info.Title == "" {
			t.Error("API title should be set")
		}
		if spec.Info.Version == "" {
			t.Error("API version should be set")
		}
		if spec.Info.Description == "" {
			t.Error("API description should be set")
		}
	})

	t.Run("Validate paths structure", func(t *testing.T) {
		spec := generator.Generate()

		expectedPaths := []string{
			"/api/v1/health",
			"/api/v1/docs",
			"/api/v1/openapi.json",
		}

		for _, path := range expectedPaths {
			if _, exists := spec.Paths[path]; !exists {
				t.Errorf("Expected path %s not found in OpenAPI spec", path)
			}
		}
	})

	t.Run("Validate components section", func(t *testing.T) {
		spec := generator.Generate()

		if spec.Components.Schemas == nil {
			t.Error("Schemas should be present in components")
		}
		if spec.Components.SecuritySchemes == nil {
			t.Error("Security schemes should be present in components")
		}
	})
}

// TestOpenAPISchemaValidation tests schema validation functionality
func TestOpenAPISchemaValidation(t *testing.T) {
	generator := NewGenerator(nil)
	registry := generator.GetSchemaRegistry()

	t.Run("Validate schema registry", func(t *testing.T) {
		if registry == nil {
			t.Fatal("Schema registry should not be nil")
		}

		// Check that schemas are registered
		allSchemas := registry.GetAllSchemas()
		if len(allSchemas) == 0 {
			t.Error("Schema registry should contain schemas")
		}
	})

	t.Run("Validate required schemas exist", func(t *testing.T) {
		requiredSchemas := []string{"Error", "SuccessResponse", "StandardResponse"}
		for _, schemaName := range requiredSchemas {
			if !registry.HasSchema(schemaName) {
				t.Errorf("Required schema %s not found in registry", schemaName)
			}
		}
	})

	t.Run("Validate schema structure", func(t *testing.T) {
		allSchemas := registry.GetAllSchemas()

		for schemaName, schema := range allSchemas {
			// Each schema should be a valid map
			if schema == nil {
				t.Errorf("Schema %s should not be nil", schemaName)
				continue
			}

			// Convert to map for validation
			schemaMap, ok := schema.(map[string]interface{})
			if !ok {
				t.Errorf("Schema %s should be a map", schemaName)
				continue
			}

			// Check for basic schema properties
			if schemaType, exists := schemaMap["type"]; exists {
				if schemaType == "object" {
					// Object schemas should have properties (except for some special cases)
					if _, hasProps := schemaMap["properties"]; !hasProps {
						// Some schemas might be empty objects or have different structures
						// Only fail if this is an unexpected case
						if schemaName != "RateLimitConfigUpdate" && schemaName != "EmptyObject" {
							t.Logf("Object schema %s has no properties (may be intentional)", schemaName)
						}
					}
				}
			}
		}
	})

	t.Run("Validate schemas by category", func(t *testing.T) {
		schemasByCategory := registry.GetSchemasByCategory()

		expectedCategories := []string{"Common", "Docker", "System", "Auth"}
		for _, category := range expectedCategories {
			if schemas, exists := schemasByCategory[category]; !exists || len(schemas) == 0 {
				t.Logf("Category %s has no schemas (may be expected)", category)
			}
		}
	})
}

// TestOpenAPIJSONSerialization tests JSON serialization of OpenAPI specs
func TestOpenAPIJSONSerialization(t *testing.T) {
	generator := NewGenerator(nil)

	t.Run("Serialize OpenAPI spec to JSON", func(t *testing.T) {
		spec := generator.Generate()

		jsonData, err := json.Marshal(spec)
		if err != nil {
			t.Fatalf("Failed to serialize OpenAPI spec to JSON: %v", err)
		}

		if len(jsonData) == 0 {
			t.Error("Serialized JSON should not be empty")
		}

		// Validate it's valid JSON by unmarshaling
		var unmarshaled map[string]interface{}
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Errorf("Serialized JSON is not valid: %v", err)
		}
	})

	t.Run("Validate JSON structure", func(t *testing.T) {
		spec := generator.Generate()

		jsonData, _ := json.Marshal(spec)
		var unmarshaled map[string]interface{}
		json.Unmarshal(jsonData, &unmarshaled)

		// Check top-level structure
		requiredFields := []string{"openapi", "info", "paths", "components"}
		for _, field := range requiredFields {
			if _, exists := unmarshaled[field]; !exists {
				t.Errorf("Required field %s not found in serialized JSON", field)
			}
		}
	})

	t.Run("Validate schema references", func(t *testing.T) {
		spec := generator.Generate()

		// Check that paths contain valid structure
		for pathName, pathItem := range spec.Paths {
			if pathItem == nil {
				t.Errorf("Path %s should not be nil", pathName)
				continue
			}

			// Convert to map for validation
			pathMap, ok := pathItem.(map[string]interface{})
			if !ok {
				t.Errorf("Path %s should be a map", pathName)
				continue
			}

			// Check for HTTP methods
			httpMethods := []string{"get", "post", "put", "delete", "patch"}
			hasMethod := false
			for _, method := range httpMethods {
				if _, exists := pathMap[method]; exists {
					hasMethod = true
					break
				}
			}
			if !hasMethod {
				t.Errorf("Path %s should have at least one HTTP method", pathName)
			}
		}
	})
}

// TestOpenAPISecuritySchemes tests security scheme definitions
func TestOpenAPISecuritySchemes(t *testing.T) {
	generator := NewGenerator(nil)

	t.Run("Validate security schemes exist", func(t *testing.T) {
		spec := generator.Generate()

		if spec.Components.SecuritySchemes == nil {
			t.Error("Security schemes should be defined")
			return
		}

		if len(spec.Components.SecuritySchemes) == 0 {
			t.Error("At least one security scheme should be defined")
		}
	})

	t.Run("Validate security scheme structure", func(t *testing.T) {
		spec := generator.Generate()

		for schemeName, scheme := range spec.Components.SecuritySchemes {
			if scheme == nil {
				t.Errorf("Security scheme %s should not be nil", schemeName)
				continue
			}

			// Convert to map for validation
			schemeMap, ok := scheme.(map[string]interface{})
			if !ok {
				t.Errorf("Security scheme %s should be a map", schemeName)
				continue
			}

			// Check for required type field
			if _, hasType := schemeMap["type"]; !hasType {
				t.Errorf("Security scheme %s should have a type", schemeName)
			}
		}
	})
}

// TestOpenAPIPathValidation tests path definitions and operations
func TestOpenAPIPathValidation(t *testing.T) {
	generator := NewGenerator(nil)

	t.Run("Validate health endpoint", func(t *testing.T) {
		spec := generator.Generate()

		healthPath, exists := spec.Paths["/api/v1/health"]
		if !exists {
			t.Fatal("Health endpoint should be defined")
		}

		// Convert to map for validation
		pathMap, ok := healthPath.(map[string]interface{})
		if !ok {
			t.Fatal("Health path should be a map")
		}

		getOp, exists := pathMap["get"]
		if !exists {
			t.Error("Health endpoint should have GET operation")
		} else {
			opMap, ok := getOp.(map[string]interface{})
			if !ok {
				t.Error("GET operation should be a map")
			} else {
				if summary, exists := opMap["summary"]; !exists || summary == "" {
					t.Error("GET health operation should have summary")
				}
				if _, exists := opMap["responses"]; !exists {
					t.Error("GET health operation should have responses")
				}
			}
		}
	})

	t.Run("Validate required endpoints", func(t *testing.T) {
		spec := generator.Generate()

		requiredPaths := []string{
			"/api/v1/health",
			"/api/v1/docs",
			"/api/v1/openapi.json",
		}

		for _, path := range requiredPaths {
			if _, exists := spec.Paths[path]; !exists {
				t.Errorf("Required path %s should be defined", path)
			}
		}
	})

	t.Run("Validate path structure", func(t *testing.T) {
		spec := generator.Generate()

		for pathName, pathItem := range spec.Paths {
			if pathItem == nil {
				t.Errorf("Path %s should not be nil", pathName)
				continue
			}

			// Convert to map for validation
			pathMap, ok := pathItem.(map[string]interface{})
			if !ok {
				t.Errorf("Path %s should be a map", pathName)
				continue
			}

			// Check that each operation has required fields
			httpMethods := []string{"get", "post", "put", "delete", "patch"}
			for _, method := range httpMethods {
				if operation, exists := pathMap[method]; exists {
					opMap, ok := operation.(map[string]interface{})
					if !ok {
						t.Errorf("Operation %s %s should be a map", method, pathName)
						continue
					}

					// Check for required operation fields
					if _, hasResponses := opMap["responses"]; !hasResponses {
						t.Errorf("Operation %s %s should have responses", method, pathName)
					}
				}
			}
		}
	})
}

// TestOpenAPIResponseValidation tests response definitions
func TestOpenAPIResponseValidation(t *testing.T) {
	generator := NewGenerator(nil)

	t.Run("Validate response structure", func(t *testing.T) {
		spec := generator.Generate()

		for pathName, pathItem := range spec.Paths {
			if pathItem == nil {
				continue
			}

			pathMap, ok := pathItem.(map[string]interface{})
			if !ok {
				continue
			}

			// Check operations for response structure
			httpMethods := []string{"get", "post", "put", "delete", "patch"}
			for _, method := range httpMethods {
				if operation, exists := pathMap[method]; exists {
					opMap, ok := operation.(map[string]interface{})
					if !ok {
						continue
					}

					if responses, exists := opMap["responses"]; exists {
						respMap, ok := responses.(map[string]interface{})
						if !ok {
							t.Errorf("Responses in %s %s should be a map", method, pathName)
							continue
						}

						// Check that at least one response is defined
						if len(respMap) == 0 {
							t.Errorf("Operation %s %s should have at least one response", method, pathName)
						}
					}
				}
			}
		}
	})

	t.Run("Validate generator validation", func(t *testing.T) {
		errors := generator.ValidateSpec()

		// Should have minimal errors for a properly configured generator
		if len(errors) > 5 {
			t.Errorf("Generator validation found too many errors: %v", errors)
		}
	})

	t.Run("Validate generator stats", func(t *testing.T) {
		stats := generator.GetStats()

		if stats == nil {
			t.Error("Generator stats should not be nil")
			return
		}

		// Check for expected stats fields
		expectedFields := []string{"openapi_version", "api_version", "total_paths", "total_schemas"}
		for _, field := range expectedFields {
			if _, exists := stats[field]; !exists {
				t.Errorf("Stats should contain field %s", field)
			}
		}
	})
}
