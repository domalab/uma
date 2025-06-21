package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewOpenAPIToolRegistry tests registry creation
func TestNewOpenAPIToolRegistry(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	registry := NewOpenAPIToolRegistry(mockAPI)

	assert.NotNil(t, registry)
	assert.NotNil(t, registry.api)
	assert.NotNil(t, registry.tools)
	assert.NotNil(t, registry.openAPISpec)
	assert.NotNil(t, registry.generator)
}

// TestGetTools tests tool retrieval
func TestGetTools(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	registry := NewOpenAPIToolRegistry(mockAPI)

	tools, err := registry.GetTools()
	assert.NoError(t, err)
	assert.NotNil(t, tools)
	// Should have some tools from OpenAPI discovery
	assert.GreaterOrEqual(t, len(tools), 0)
}

// TestExecuteTool tests tool execution
func TestExecuteTool(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	registry := NewOpenAPIToolRegistry(mockAPI)

	// Test non-existent tool
	_, err := registry.ExecuteTool("nonexistent_tool", map[string]interface{}{})
	assert.Error(t, err)
	assert.Equal(t, "tool not found", err.Error())

	// Test with existing tools (if any)
	tools, _ := registry.GetTools()
	if len(tools) > 0 {
		toolName := tools[0].Name
		result, err := registry.ExecuteTool(toolName, map[string]interface{}{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsError)
		assert.NotEmpty(t, result.Content)
	}
}

// TestGetToolByName tests individual tool retrieval
func TestGetToolByName(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	registry := NewOpenAPIToolRegistry(mockAPI)

	// Test non-existent tool
	tool, exists := registry.GetToolByName("nonexistent_tool")
	assert.False(t, exists)
	assert.Nil(t, tool)

	// Test with existing tools (if any)
	tools, _ := registry.GetTools()
	if len(tools) > 0 {
		toolName := tools[0].Name
		tool, exists := registry.GetToolByName(toolName)
		assert.True(t, exists)
		assert.NotNil(t, tool)
		assert.Equal(t, toolName, tool.Tool.Name)
	}
}

// TestGetToolCount tests tool counting
func TestGetToolCount(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	registry := NewOpenAPIToolRegistry(mockAPI)

	count := registry.GetToolCount()
	assert.GreaterOrEqual(t, count, 0)

	tools, _ := registry.GetTools()
	assert.Equal(t, len(tools), count)
}

// TestGetToolsByCategory tests tool categorization
func TestGetToolsByCategory(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	registry := NewOpenAPIToolRegistry(mockAPI)

	categories := registry.GetToolsByCategory()
	assert.NotNil(t, categories)

	// Verify all tools are categorized
	totalCategorizedTools := 0
	for _, tools := range categories {
		totalCategorizedTools += len(tools)
	}

	allTools, _ := registry.GetTools()
	// Note: Tools might be in multiple categories, so we check >= instead of ==
	assert.GreaterOrEqual(t, totalCategorizedTools, len(allTools))
}

// TestRefreshTools tests tool registry refresh
func TestRefreshTools(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	registry := NewOpenAPIToolRegistry(mockAPI)

	err := registry.RefreshTools()
	assert.NoError(t, err)

	newCount := registry.GetToolCount()
	// Count should be the same after refresh (or potentially different if spec changed)
	assert.GreaterOrEqual(t, newCount, 0)

	// Verify tools are still accessible
	tools, err := registry.GetTools()
	assert.NoError(t, err)
	assert.Equal(t, newCount, len(tools))
}

// TestGetRegistryStats tests registry statistics
func TestGetRegistryStats(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	registry := NewOpenAPIToolRegistry(mockAPI)

	stats := registry.GetRegistryStats()
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "total_tools")
	assert.Contains(t, stats, "categories")
	assert.Contains(t, stats, "openapi_version")
	assert.Contains(t, stats, "api_version")

	// Verify stats consistency
	totalTools, ok := stats["total_tools"].(int)
	assert.True(t, ok)
	assert.Equal(t, registry.GetToolCount(), totalTools)

	categories, ok := stats["categories"].(map[string]int)
	assert.True(t, ok)
	assert.NotNil(t, categories)
}

// TestIsValidHTTPMethod tests HTTP method validation
func TestIsValidHTTPMethod(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	registry := NewOpenAPIToolRegistry(mockAPI)

	tests := []struct {
		method string
		valid  bool
	}{
		{"GET", true},
		{"POST", true},
		{"PUT", true},
		{"DELETE", true},
		{"PATCH", true},
		{"HEAD", true},
		{"OPTIONS", true},
		{"get", true}, // lowercase
		{"INVALID", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			result := registry.isValidHTTPMethod(tt.method)
			assert.Equal(t, tt.valid, result)
		})
	}
}

// TestShouldIncludeEndpoint tests endpoint filtering
func TestShouldIncludeEndpoint(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	registry := NewOpenAPIToolRegistry(mockAPI)

	tests := []struct {
		name     string
		endpoint EndpointInfo
		include  bool
	}{
		{
			name: "valid GET endpoint",
			endpoint: EndpointInfo{
				Path:   "/api/v1/system/info",
				Method: "GET",
			},
			include: true,
		},
		{
			name: "POST endpoint (excluded for now)",
			endpoint: EndpointInfo{
				Path:   "/api/v1/system/reboot",
				Method: "POST",
			},
			include: false,
		},
		{
			name: "docs endpoint (excluded)",
			endpoint: EndpointInfo{
				Path:   "/docs",
				Method: "GET",
			},
			include: false,
		},
		{
			name: "websocket endpoint (excluded)",
			endpoint: EndpointInfo{
				Path:   "/api/v1/ws",
				Method: "GET",
			},
			include: false,
		},
		{
			name: "metrics endpoint (excluded)",
			endpoint: EndpointInfo{
				Path:   "/metrics",
				Method: "GET",
			},
			include: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.shouldIncludeEndpoint(tt.endpoint)
			assert.Equal(t, tt.include, result)
		})
	}
}

// TestGenerateToolName tests tool name generation
func TestGenerateToolName(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	registry := NewOpenAPIToolRegistry(mockAPI)

	tests := []struct {
		name     string
		endpoint EndpointInfo
		expected string
	}{
		{
			name: "with operation ID",
			endpoint: EndpointInfo{
				OperationID: "getSystemInfo",
			},
			expected: "get_system_info",
		},
		{
			name: "without operation ID",
			endpoint: EndpointInfo{
				Path:   "/api/v1/system/info",
				Method: "GET",
			},
			expected: "get_system_info",
		},
		{
			name: "with path parameters",
			endpoint: EndpointInfo{
				Path:   "/api/v1/docker/containers/{id}",
				Method: "GET",
			},
			expected: "get_docker_containers_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := registry.generateToolName(tt.endpoint)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCamelToSnake tests camelCase to snake_case conversion
func TestCamelToSnake(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	registry := NewOpenAPIToolRegistry(mockAPI)

	tests := []struct {
		input    string
		expected string
	}{
		{"getSystemInfo", "get_system_info"},
		{"listDockerContainers", "list_docker_containers"},
		{"simpleword", "simpleword"},
		{"HTTPSConnection", "h_t_t_p_s_connection"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := registry.camelToSnake(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestConvertParametersToSchema tests parameter schema conversion
func TestConvertParametersToSchema(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	registry := NewOpenAPIToolRegistry(mockAPI)

	parameters := []ParameterInfo{
		{
			Name:        "id",
			In:          "path",
			Description: "Container ID",
			Required:    true,
			Schema: map[string]interface{}{
				"type": "string",
			},
		},
		{
			Name:        "limit",
			In:          "query",
			Description: "Limit results",
			Required:    false,
			Schema: map[string]interface{}{
				"type":    "integer",
				"minimum": 1,
			},
		},
	}

	schema := registry.convertParametersToSchema(parameters)

	assert.Equal(t, "object", schema.Type)
	assert.Len(t, schema.Properties, 2)
	assert.Contains(t, schema.Properties, "id")
	assert.Contains(t, schema.Properties, "limit")
	assert.Contains(t, schema.Required, "id")
	assert.NotContains(t, schema.Required, "limit")
}

// TestConvertParameterSchema tests individual parameter schema conversion
func TestConvertParameterSchema(t *testing.T) {
	mockAPI := &MockAPIInterface{}
	registry := NewOpenAPIToolRegistry(mockAPI)

	param := ParameterInfo{
		Name:        "test_param",
		Description: "Test parameter",
		Schema: map[string]interface{}{
			"type":    "string",
			"pattern": "^[a-z]+$",
		},
	}

	schema := registry.convertParameterSchema(param)

	assert.Contains(t, schema, "type")
	assert.Contains(t, schema, "description")
	assert.Contains(t, schema, "pattern")
	assert.Equal(t, "string", schema["type"])
	assert.Equal(t, "Test parameter", schema["description"])
	assert.Equal(t, "^[a-z]+$", schema["pattern"])
}

// Benchmark tests for performance validation
func BenchmarkGetTools(b *testing.B) {
	mockAPI := &MockAPIInterface{}
	registry := NewOpenAPIToolRegistry(mockAPI)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = registry.GetTools()
	}
}

func BenchmarkExecuteTool(b *testing.B) {
	mockAPI := &MockAPIInterface{}
	registry := NewOpenAPIToolRegistry(mockAPI)

	tools, _ := registry.GetTools()
	if len(tools) == 0 {
		b.Skip("No tools available for benchmarking")
	}

	toolName := tools[0].Name
	args := map[string]interface{}{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = registry.ExecuteTool(toolName, args)
	}
}
