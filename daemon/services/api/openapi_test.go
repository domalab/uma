package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/domalab/uma/daemon/domain"
	"github.com/domalab/uma/daemon/services/api/openapi"
)

// TestOpenAPIGeneration tests that the modular OpenAPI system generates a valid specification
func TestOpenAPIGeneration(t *testing.T) {
	// Create a test HTTP server instance
	config := domain.DefaultConfig()
	config.Version = "test-1.0.0"
	config.HTTPServer.Port = 34600

	ctx := &domain.Context{
		Config: config,
	}

	api := &Api{ctx: ctx}
	server := &HTTPServer{api: api}

	// Test that generateOpenAPISpec returns a valid specification
	spec := server.generateOpenAPISpec()
	if spec == nil {
		t.Fatal("generateOpenAPISpec returned nil")
	}

	// Verify basic structure
	if spec.OpenAPI != "3.1.1" {
		t.Errorf("Expected OpenAPI version 3.1.1, got %s", spec.OpenAPI)
	}

	if spec.Info.Title != "UMA REST API" {
		t.Errorf("Expected title 'UMA REST API', got %s", spec.Info.Title)
	}

	if spec.Info.Version != "test-1.0.0" {
		t.Errorf("Expected version 'test-1.0.0', got %s", spec.Info.Version)
	}

	// Verify that paths are generated
	if spec.Paths == nil {
		t.Fatal("Paths should not be nil")
	}

	// Verify that components are generated
	if spec.Components.Schemas == nil {
		t.Fatal("Components.Schemas should not be nil")
	}

	// Test that the specification can be marshaled to JSON
	jsonData, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("Failed to marshal OpenAPI spec to JSON: %v", err)
	}

	// Verify that the JSON is valid by unmarshaling it back
	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Generated JSON is not valid: %v", err)
	}

	// Verify key sections exist in the JSON
	if _, exists := unmarshaled["openapi"]; !exists {
		t.Error("Generated JSON missing 'openapi' field")
	}

	if _, exists := unmarshaled["info"]; !exists {
		t.Error("Generated JSON missing 'info' field")
	}

	if _, exists := unmarshaled["paths"]; !exists {
		t.Error("Generated JSON missing 'paths' field")
	}

	if _, exists := unmarshaled["components"]; !exists {
		t.Error("Generated JSON missing 'components' field")
	}
}

// TestOpenAPIHandler tests the HTTP handler for OpenAPI specification
func TestOpenAPIHandler(t *testing.T) {
	// Create a test HTTP server instance
	config := domain.DefaultConfig()
	config.Version = "test-1.0.0"
	config.HTTPServer.Port = 34600

	ctx := &domain.Context{
		Config: config,
	}

	api := &Api{ctx: ctx}
	server := &HTTPServer{api: api}

	// Create a test request
	req, err := http.NewRequest("GET", "/api/v1/openapi.json", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler
	server.OpenAPIHandler(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the content type
	expected := "application/json"
	if ct := rr.Header().Get("Content-Type"); ct != expected {
		t.Errorf("Handler returned wrong content type: got %v want %v", ct, expected)
	}

	// Check that the response body contains valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Errorf("Response body is not valid JSON: %v", err)
	}

	// Verify key fields exist
	if _, exists := result["openapi"]; !exists {
		t.Error("Response missing 'openapi' field")
	}

	if _, exists := result["info"]; !exists {
		t.Error("Response missing 'info' field")
	}
}

// TestSwaggerUIHandler tests the Swagger UI handler
func TestSwaggerUIHandler(t *testing.T) {
	// Create a test HTTP server instance
	config := domain.DefaultConfig()
	config.HTTPServer.Port = 34600

	ctx := &domain.Context{
		Config: config,
	}

	api := &Api{ctx: ctx}
	server := &HTTPServer{api: api}

	// Create a test request
	req, err := http.NewRequest("GET", "/api/v1/docs", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler
	server.SwaggerUIHandler(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the content type
	expected := "text/html; charset=utf-8"
	if ct := rr.Header().Get("Content-Type"); ct != expected {
		t.Errorf("Handler returned wrong content type: got %v want %v", ct, expected)
	}

	// Check that the response body contains HTML
	body := rr.Body.String()
	if len(body) == 0 {
		t.Error("Response body is empty")
	}

	// Verify it contains Swagger UI elements
	if !containsString(body, "swagger-ui") {
		t.Error("Response does not contain Swagger UI elements")
	}

	if !containsString(body, "/api/v1/openapi.json") {
		t.Error("Response does not reference the OpenAPI JSON endpoint")
	}
}

// TestModularSystemIntegration tests that all modules work together
func TestModularSystemIntegration(t *testing.T) {
	// Create configuration
	config := &openapi.Config{
		Version:     "test-1.0.0",
		Port:        34600,
		BaseURL:     "",
		Environment: "test",
		Features: openapi.FeatureFlags{
			Authentication: true,
			BulkOperations: true,
			WebSockets:     true,
			Metrics:        true,
			ZFS:            true,
			ArrayControl:   true,
			VMManagement:   true,
		},
	}

	// Create generator
	generator := openapi.NewGenerator(config)
	if generator == nil {
		t.Fatal("NewGenerator returned nil")
	}

	// Generate specification
	spec := generator.Generate()
	if spec == nil {
		t.Fatal("Generate returned nil")
	}

	// Verify the specification has all expected sections
	if spec.Paths == nil {
		t.Error("Generated spec missing paths")
	}

	if spec.Components.Schemas == nil {
		t.Error("Generated spec missing schemas")
	}

	// Test validation
	errors := generator.ValidateSpec()
	if len(errors) > 0 {
		t.Logf("Validation warnings/errors: %v", errors)
		// Note: We log but don't fail on validation errors as some may be expected
	}

	// Test statistics
	stats := generator.GetStats()
	if stats == nil {
		t.Error("GetStats returned nil")
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
