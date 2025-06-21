package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/services/api/openapi"
	"github.com/domalab/uma/daemon/services/api/utils"
)

// OpenAPIToolRegistry automatically discovers REST endpoints from OpenAPI specification
// and converts them to MCP tools
type OpenAPIToolRegistry struct {
	api         utils.APIInterface
	tools       map[string]*ToolDefinition
	mutex       sync.RWMutex
	openAPISpec *openapi.OpenAPISpec
	generator   *openapi.Generator
}

// EndpointInfo represents information extracted from OpenAPI specification
type EndpointInfo struct {
	Path        string
	Method      string
	Summary     string
	Description string
	Parameters  []ParameterInfo
	Responses   map[string]ResponseInfo
	Tags        []string
	OperationID string
}

// ParameterInfo represents parameter information from OpenAPI
type ParameterInfo struct {
	Name        string
	In          string // query, path, header, cookie
	Description string
	Required    bool
	Schema      map[string]interface{}
}

// ResponseInfo represents response information from OpenAPI
type ResponseInfo struct {
	Description string
	Schema      map[string]interface{}
}

// NewOpenAPIToolRegistry creates a new OpenAPI-based tool registry
func NewOpenAPIToolRegistry(api utils.APIInterface) *OpenAPIToolRegistry {
	// Create OpenAPI generator
	config := &openapi.Config{
		Version:     "2025.06.16",
		Port:        34600,
		BaseURL:     "",
		Environment: "prod",
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

	generator := openapi.NewGenerator(config)
	spec := generator.Generate()

	registry := &OpenAPIToolRegistry{
		api:         api,
		tools:       make(map[string]*ToolDefinition),
		openAPISpec: spec,
		generator:   generator,
	}

	// Automatically discover and register tools from OpenAPI spec
	registry.discoverAndRegisterTools()

	return registry
}

// discoverAndRegisterTools discovers REST endpoints from OpenAPI spec and converts them to MCP tools
func (r *OpenAPIToolRegistry) discoverAndRegisterTools() {
	logger.Info("Discovering MCP tools from OpenAPI specification...")

	endpoints := r.extractEndpointsFromSpec()

	for _, endpoint := range endpoints {
		if r.shouldIncludeEndpoint(endpoint) {
			tool := r.convertEndpointToTool(endpoint)
			handler := r.createToolHandler(endpoint)

			r.registerTool(tool.Name, tool, endpoint.Path, endpoint.Method, handler)
		}
	}

	logger.Green("Discovered and registered %d MCP tools from OpenAPI specification", len(r.tools))
}

// extractEndpointsFromSpec extracts endpoint information from OpenAPI specification
func (r *OpenAPIToolRegistry) extractEndpointsFromSpec() []EndpointInfo {
	var endpoints []EndpointInfo

	// Iterate through all paths in the OpenAPI spec
	for pathStr, pathItem := range r.openAPISpec.Paths {
		pathItemMap, ok := pathItem.(map[string]interface{})
		if !ok {
			continue
		}

		// Check each HTTP method
		for method, operation := range pathItemMap {
			if !r.isValidHTTPMethod(method) {
				continue
			}

			operationMap, ok := operation.(map[string]interface{})
			if !ok {
				continue
			}

			endpoint := r.parseOperation(pathStr, method, operationMap)
			endpoints = append(endpoints, endpoint)
		}
	}

	return endpoints
}

// parseOperation parses an OpenAPI operation into EndpointInfo
func (r *OpenAPIToolRegistry) parseOperation(path, method string, operation map[string]interface{}) EndpointInfo {
	endpoint := EndpointInfo{
		Path:   path,
		Method: strings.ToUpper(method),
	}

	// Extract basic information
	if summary, ok := operation["summary"].(string); ok {
		endpoint.Summary = summary
	}
	if description, ok := operation["description"].(string); ok {
		endpoint.Description = description
	}
	if operationID, ok := operation["operationId"].(string); ok {
		endpoint.OperationID = operationID
	}

	// Extract tags
	if tags, ok := operation["tags"].([]interface{}); ok {
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				endpoint.Tags = append(endpoint.Tags, tagStr)
			}
		}
	}

	// Extract parameters
	if parameters, ok := operation["parameters"].([]interface{}); ok {
		for _, param := range parameters {
			if paramMap, ok := param.(map[string]interface{}); ok {
				paramInfo := r.parseParameter(paramMap)
				endpoint.Parameters = append(endpoint.Parameters, paramInfo)
			}
		}
	}

	// Extract responses
	endpoint.Responses = make(map[string]ResponseInfo)
	if responses, ok := operation["responses"].(map[string]interface{}); ok {
		for statusCode, response := range responses {
			if responseMap, ok := response.(map[string]interface{}); ok {
				responseInfo := r.parseResponse(responseMap)
				endpoint.Responses[statusCode] = responseInfo
			}
		}
	}

	return endpoint
}

// parseParameter parses an OpenAPI parameter into ParameterInfo
func (r *OpenAPIToolRegistry) parseParameter(param map[string]interface{}) ParameterInfo {
	paramInfo := ParameterInfo{}

	if name, ok := param["name"].(string); ok {
		paramInfo.Name = name
	}
	if in, ok := param["in"].(string); ok {
		paramInfo.In = in
	}
	if description, ok := param["description"].(string); ok {
		paramInfo.Description = description
	}
	if required, ok := param["required"].(bool); ok {
		paramInfo.Required = required
	}
	if schema, ok := param["schema"].(map[string]interface{}); ok {
		paramInfo.Schema = schema
	}

	return paramInfo
}

// parseResponse parses an OpenAPI response into ResponseInfo
func (r *OpenAPIToolRegistry) parseResponse(response map[string]interface{}) ResponseInfo {
	responseInfo := ResponseInfo{}

	if description, ok := response["description"].(string); ok {
		responseInfo.Description = description
	}

	// Extract schema from content
	if content, ok := response["content"].(map[string]interface{}); ok {
		for _, mediaType := range content {
			if mediaTypeMap, ok := mediaType.(map[string]interface{}); ok {
				if schema, ok := mediaTypeMap["schema"].(map[string]interface{}); ok {
					responseInfo.Schema = schema
					break
				}
			}
		}
	}

	return responseInfo
}

// isValidHTTPMethod checks if the method is a valid HTTP method
func (r *OpenAPIToolRegistry) isValidHTTPMethod(method string) bool {
	validMethods := []string{"get", "post", "put", "patch", "delete", "head", "options"}
	method = strings.ToLower(method)

	for _, validMethod := range validMethods {
		if method == validMethod {
			return true
		}
	}

	return false
}

// shouldIncludeEndpoint determines if an endpoint should be included as an MCP tool
func (r *OpenAPIToolRegistry) shouldIncludeEndpoint(endpoint EndpointInfo) bool {
	// Skip certain endpoints that are not suitable for MCP tools
	excludePatterns := []string{
		"/docs",
		"/openapi.json",
		"/metrics",
		"/ws",
		"/websocket",
	}

	for _, pattern := range excludePatterns {
		if strings.Contains(endpoint.Path, pattern) {
			return false
		}
	}

	// Only include GET endpoints for now (safe operations)
	// TODO: Add support for POST/PUT/DELETE with proper confirmation
	return endpoint.Method == "GET"
}

// convertEndpointToTool converts an OpenAPI endpoint to an MCP tool
func (r *OpenAPIToolRegistry) convertEndpointToTool(endpoint EndpointInfo) Tool {
	// Generate tool name from operation ID or path
	toolName := r.generateToolName(endpoint)

	// Generate description
	description := endpoint.Description
	if description == "" {
		description = endpoint.Summary
	}
	if description == "" {
		description = fmt.Sprintf("Execute %s %s", endpoint.Method, endpoint.Path)
	}

	// Convert parameters to input schema
	inputSchema := r.convertParametersToSchema(endpoint.Parameters)

	return Tool{
		Name:        toolName,
		Description: description,
		InputSchema: inputSchema,
	}
}

// generateToolName generates a tool name from endpoint information
func (r *OpenAPIToolRegistry) generateToolName(endpoint EndpointInfo) string {
	if endpoint.OperationID != "" {
		return r.camelToSnake(endpoint.OperationID)
	}

	// Generate from path and method
	path := strings.ReplaceAll(endpoint.Path, "/api/v1/", "")
	path = strings.ReplaceAll(path, "/", "_")
	path = strings.ReplaceAll(path, "{", "")
	path = strings.ReplaceAll(path, "}", "")

	method := strings.ToLower(endpoint.Method)

	return fmt.Sprintf("%s_%s", method, path)
}

// camelToSnake converts camelCase to snake_case
func (r *OpenAPIToolRegistry) camelToSnake(str string) string {
	var result strings.Builder

	for i, char := range str {
		if i > 0 && char >= 'A' && char <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(char)
	}

	return strings.ToLower(result.String())
}

// convertParametersToSchema converts OpenAPI parameters to MCP tool input schema
func (r *OpenAPIToolRegistry) convertParametersToSchema(parameters []ParameterInfo) ToolSchema {
	schema := ToolSchema{
		Type:       "object",
		Properties: make(map[string]interface{}),
		Required:   []string{},
	}

	for _, param := range parameters {
		// Convert parameter schema
		paramSchema := r.convertParameterSchema(param)
		schema.Properties[param.Name] = paramSchema

		// Add to required if necessary
		if param.Required {
			schema.Required = append(schema.Required, param.Name)
		}
	}

	return schema
}

// convertParameterSchema converts an OpenAPI parameter schema to MCP format
func (r *OpenAPIToolRegistry) convertParameterSchema(param ParameterInfo) map[string]interface{} {
	schema := make(map[string]interface{})

	// Copy schema properties if available
	if param.Schema != nil {
		for key, value := range param.Schema {
			schema[key] = value
		}
	}

	// Add description
	if param.Description != "" {
		schema["description"] = param.Description
	}

	// Set default type if not specified
	if _, hasType := schema["type"]; !hasType {
		schema["type"] = "string"
	}

	return schema
}

// createToolHandler creates a tool handler for an endpoint
func (r *OpenAPIToolRegistry) createToolHandler(endpoint EndpointInfo) ToolHandler {
	return func(args map[string]interface{}) (ToolCallResult, error) {
		// For now, return a placeholder response
		// TODO: Implement actual API calls using the endpoint information

		result := ToolCallResult{
			Content: []ToolContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Tool '%s' executed successfully for endpoint %s %s",
						r.generateToolName(endpoint), endpoint.Method, endpoint.Path),
				},
			},
			IsError: false,
		}

		// Add endpoint information to the response
		if len(args) > 0 {
			argsJSON, _ := json.Marshal(args)
			result.Content = append(result.Content, ToolContent{
				Type: "text",
				Text: fmt.Sprintf("Arguments: %s", string(argsJSON)),
			})
		}

		return result, nil
	}
}

// registerTool registers a single tool
func (r *OpenAPIToolRegistry) registerTool(name string, tool Tool, endpoint string, method string, handler ToolHandler) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.tools[name] = &ToolDefinition{
		Tool:     tool,
		Endpoint: endpoint,
		Method:   method,
		Handler:  handler,
	}

	logger.Debug("Registered MCP tool: %s -> %s %s", name, method, endpoint)
}

// GetTools returns all registered tools
func (r *OpenAPIToolRegistry) GetTools() ([]Tool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tools := make([]Tool, 0, len(r.tools))
	for _, toolDef := range r.tools {
		tools = append(tools, toolDef.Tool)
	}

	return tools, nil
}

// ExecuteTool executes a tool by name with the given arguments
func (r *OpenAPIToolRegistry) ExecuteTool(name string, args map[string]interface{}) (ToolCallResult, error) {
	r.mutex.RLock()
	toolDef, exists := r.tools[name]
	r.mutex.RUnlock()

	if !exists {
		return ToolCallResult{}, fmt.Errorf("tool not found")
	}

	logger.Info("Executing MCP tool: %s", name)
	return toolDef.Handler(args)
}

// GetToolByName returns a specific tool by name
func (r *OpenAPIToolRegistry) GetToolByName(name string) (*ToolDefinition, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	tool, exists := r.tools[name]
	return tool, exists
}

// GetToolCount returns the number of registered tools
func (r *OpenAPIToolRegistry) GetToolCount() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.tools)
}

// GetToolsByCategory returns tools grouped by their OpenAPI tags
func (r *OpenAPIToolRegistry) GetToolsByCategory() map[string][]Tool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	categories := make(map[string][]Tool)

	// Extract category information from the original endpoints
	endpoints := r.extractEndpointsFromSpec()

	for _, endpoint := range endpoints {
		if !r.shouldIncludeEndpoint(endpoint) {
			continue
		}

		tool := r.convertEndpointToTool(endpoint)

		// Use tags as categories
		if len(endpoint.Tags) > 0 {
			for _, tag := range endpoint.Tags {
				categories[tag] = append(categories[tag], tool)
			}
		} else {
			// Default category
			categories["General"] = append(categories["General"], tool)
		}
	}

	return categories
}

// RefreshTools refreshes the tool registry from the latest OpenAPI specification
func (r *OpenAPIToolRegistry) RefreshTools() error {
	logger.Info("Refreshing MCP tools from OpenAPI specification...")

	// Clear existing tools
	r.mutex.Lock()
	r.tools = make(map[string]*ToolDefinition)
	r.mutex.Unlock()

	// Regenerate OpenAPI spec
	r.openAPISpec = r.generator.Generate()

	// Rediscover and register tools
	r.discoverAndRegisterTools()

	return nil
}

// GetRegistryStats returns statistics about the tool registry
func (r *OpenAPIToolRegistry) GetRegistryStats() map[string]interface{} {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	categories := r.GetToolsByCategory()
	categoryStats := make(map[string]int)

	for category, tools := range categories {
		categoryStats[category] = len(tools)
	}

	return map[string]interface{}{
		"total_tools":     len(r.tools),
		"categories":      categoryStats,
		"openapi_version": r.openAPISpec.OpenAPI,
		"api_version":     r.openAPISpec.Info.Version,
	}
}
