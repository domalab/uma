package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// SchemaGenerator generates OpenAPI schemas from live API responses
type SchemaGenerator struct {
	baseURL   string
	client    *http.Client
	schemas   map[string]interface{}
	endpoints []EndpointInfo
}

// EndpointInfo represents an API endpoint with its generated schema
type EndpointInfo struct {
	Path        string                 `json:"path"`
	Method      string                 `json:"method"`
	StatusCode  int                    `json:"status_code"`
	Schema      map[string]interface{} `json:"schema"`
	Example     interface{}            `json:"example"`
	ContentType string                 `json:"content_type"`
	Error       string                 `json:"error,omitempty"`
}

// GeneratedSpec represents the complete generated OpenAPI specification
type GeneratedSpec struct {
	OpenAPI    string                            `json:"openapi"`
	Info       map[string]interface{}            `json:"info"`
	Servers    []map[string]interface{}          `json:"servers"`
	Paths      map[string]map[string]interface{} `json:"paths"`
	Components map[string]interface{}            `json:"components"`
	Generated  map[string]interface{}            `json:"x-generated"`
}

// NewSchemaGenerator creates a new schema generator
func NewSchemaGenerator(baseURL string) *SchemaGenerator {
	return &SchemaGenerator{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		schemas:   make(map[string]interface{}),
		endpoints: make([]EndpointInfo, 0),
	}
}

// GenerateFromEndpoints generates schemas by calling live API endpoints
func (sg *SchemaGenerator) GenerateFromEndpoints(endpoints []string) error {
	fmt.Printf("🔄 Generating schemas from %d live endpoints...\n", len(endpoints))

	for i, endpoint := range endpoints {
		fmt.Printf("[%d/%d] Analyzing %s...\n", i+1, len(endpoints), endpoint)

		if err := sg.analyzeEndpoint(endpoint); err != nil {
			fmt.Printf("  ⚠️  Error: %v\n", err)
		} else {
			fmt.Printf("  ✅ Schema generated\n")
		}

		// Small delay to avoid overwhelming the server
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// analyzeEndpoint analyzes a single endpoint and generates its schema
func (sg *SchemaGenerator) analyzeEndpoint(path string) error {
	url := sg.baseURL + path

	resp, err := sg.client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to call endpoint: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	contentType := resp.Header.Get("Content-Type")

	// Parse JSON response
	var data interface{}
	if strings.Contains(contentType, "application/json") {
		if err := json.Unmarshal(body, &data); err != nil {
			return fmt.Errorf("failed to parse JSON: %v", err)
		}
	} else {
		// Handle non-JSON responses (like metrics endpoint)
		data = string(body)
	}

	// Generate schema from the response data
	schema := sg.generateSchemaFromData(data)

	// Store endpoint info
	endpointInfo := EndpointInfo{
		Path:        path,
		Method:      "GET",
		StatusCode:  resp.StatusCode,
		Schema:      schema,
		Example:     data,
		ContentType: contentType,
	}

	sg.endpoints = append(sg.endpoints, endpointInfo)

	return nil
}

// generateSchemaFromData generates a JSON schema from actual data
func (sg *SchemaGenerator) generateSchemaFromData(data interface{}) map[string]interface{} {
	return sg.inferSchema(data)
}

// inferSchema infers JSON schema from Go data structures
func (sg *SchemaGenerator) inferSchema(data interface{}) map[string]interface{} {
	if data == nil {
		return map[string]interface{}{
			"type":     "null",
			"nullable": true,
		}
	}

	switch v := data.(type) {
	case bool:
		return map[string]interface{}{
			"type": "boolean",
		}
	case float64:
		// Check if it's actually an integer
		if v == float64(int64(v)) {
			return map[string]interface{}{
				"type": "integer",
			}
		}
		return map[string]interface{}{
			"type": "number",
		}
	case string:
		schema := map[string]interface{}{
			"type": "string",
		}

		// Try to detect common string formats
		if sg.isDateTimeString(v) {
			schema["format"] = "date-time"
		} else if sg.isEmailString(v) {
			schema["format"] = "email"
		} else if sg.isUUIDString(v) {
			schema["format"] = "uuid"
		}

		return schema
	case []interface{}:
		if len(v) == 0 {
			return map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "object"},
			}
		}

		// Infer schema from first item (assuming homogeneous array)
		itemSchema := sg.inferSchema(v[0])

		return map[string]interface{}{
			"type":  "array",
			"items": itemSchema,
		}
	case map[string]interface{}:
		properties := make(map[string]interface{})
		required := make([]string, 0)

		// Sort keys for consistent output
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			value := v[key]
			properties[key] = sg.inferSchema(value)

			// Consider non-null values as required
			if value != nil {
				required = append(required, key)
			}
		}

		schema := map[string]interface{}{
			"type":       "object",
			"properties": properties,
		}

		if len(required) > 0 {
			schema["required"] = required
		}

		return schema
	default:
		// Fallback for unknown types
		return map[string]interface{}{
			"type":        "object",
			"description": fmt.Sprintf("Unknown type: %T", data),
		}
	}
}

// Helper functions for string format detection
func (sg *SchemaGenerator) isDateTimeString(s string) bool {
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if _, err := time.Parse(format, s); err == nil {
			return true
		}
	}
	return false
}

func (sg *SchemaGenerator) isEmailString(s string) bool {
	return strings.Contains(s, "@") && strings.Contains(s, ".")
}

func (sg *SchemaGenerator) isUUIDString(s string) bool {
	return len(s) == 36 && strings.Count(s, "-") == 4
}

// GenerateOpenAPISpec generates a complete OpenAPI specification
func (sg *SchemaGenerator) GenerateOpenAPISpec() *GeneratedSpec {
	paths := make(map[string]map[string]interface{})
	components := map[string]interface{}{
		"schemas": sg.schemas,
	}

	// Generate paths from analyzed endpoints
	for _, endpoint := range sg.endpoints {
		if paths[endpoint.Path] == nil {
			paths[endpoint.Path] = make(map[string]interface{})
		}

		operation := map[string]interface{}{
			"summary":     fmt.Sprintf("Generated from live API response"),
			"description": fmt.Sprintf("Auto-generated schema for %s", endpoint.Path),
			"responses": map[string]interface{}{
				fmt.Sprintf("%d", endpoint.StatusCode): map[string]interface{}{
					"description": "Successful response",
					"content": map[string]interface{}{
						endpoint.ContentType: map[string]interface{}{
							"schema":  endpoint.Schema,
							"example": endpoint.Example,
						},
					},
				},
			},
		}

		if endpoint.Error != "" {
			operation["x-error"] = endpoint.Error
		}

		paths[endpoint.Path][strings.ToLower(endpoint.Method)] = operation
	}

	return &GeneratedSpec{
		OpenAPI: "3.0.3",
		Info: map[string]interface{}{
			"title":       "UMA API - Generated",
			"description": "Auto-generated OpenAPI specification from live API responses",
			"version":     "1.0.0-generated",
		},
		Servers: []map[string]interface{}{
			{
				"url":         sg.baseURL,
				"description": "Live API server",
			},
		},
		Paths:      paths,
		Components: components,
		Generated: map[string]interface{}{
			"timestamp":          time.Now().UTC().Format(time.RFC3339),
			"generator":          "UMA Schema Generator",
			"endpoints_analyzed": len(sg.endpoints),
			"base_url":           sg.baseURL,
		},
	}
}

// SaveSpec saves the generated specification to a file
func (sg *SchemaGenerator) SaveSpec(filename string) error {
	spec := sg.GenerateOpenAPISpec()

	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal spec: %v", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// PrintSummary prints a summary of the generation process
func (sg *SchemaGenerator) PrintSummary() {
	fmt.Printf("\n📊 Schema Generation Summary:\n")
	fmt.Printf("  Total endpoints analyzed: %d\n", len(sg.endpoints))

	statusCodes := make(map[int]int)
	contentTypes := make(map[string]int)

	for _, endpoint := range sg.endpoints {
		statusCodes[endpoint.StatusCode]++
		contentTypes[endpoint.ContentType]++
	}

	fmt.Printf("  Status codes:\n")
	for code, count := range statusCodes {
		fmt.Printf("    %d: %d endpoints\n", code, count)
	}

	fmt.Printf("  Content types:\n")
	for contentType, count := range contentTypes {
		fmt.Printf("    %s: %d endpoints\n", contentType, count)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: schema-generator <base-url> [--output <file>] [--endpoints <file>]")
		fmt.Println("  base-url: Base URL of the API (e.g., http://192.168.20.21:34600)")
		fmt.Println("  --output: Output file for generated OpenAPI spec (default: generated-openapi.json)")
		fmt.Println("  --endpoints: JSON file containing list of endpoints to analyze")
		fmt.Println("  --routes: Use discovered routes from route-scanner output")
		os.Exit(1)
	}

	baseURL := os.Args[1]
	outputFile := "generated-openapi.json"
	endpointsFile := ""
	useRoutes := false

	// Parse command line arguments
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--output":
			if i+1 < len(os.Args) {
				outputFile = os.Args[i+1]
				i++
			}
		case "--endpoints":
			if i+1 < len(os.Args) {
				endpointsFile = os.Args[i+1]
				i++
			}
		case "--routes":
			useRoutes = true
		}
	}

	generator := NewSchemaGenerator(baseURL)

	var endpoints []string

	if useRoutes {
		// Load endpoints from route scanner output
		endpoints = loadEndpointsFromRoutes()
	} else if endpointsFile != "" {
		// Load endpoints from file
		var err error
		endpoints, err = loadEndpointsFromFile(endpointsFile)
		if err != nil {
			fmt.Printf("Error loading endpoints: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Default set of endpoints for testing
		endpoints = getDefaultEndpoints()
	}

	fmt.Printf("🚀 Starting schema generation for %s\n", baseURL)
	fmt.Printf("📋 Analyzing %d endpoints\n", len(endpoints))

	if err := generator.GenerateFromEndpoints(endpoints); err != nil {
		fmt.Printf("Error generating schemas: %v\n", err)
		os.Exit(1)
	}

	// Save the generated specification
	if err := generator.SaveSpec(outputFile); err != nil {
		fmt.Printf("Error saving specification: %v\n", err)
		os.Exit(1)
	}

	generator.PrintSummary()
	fmt.Printf("\n✅ Generated OpenAPI specification saved to: %s\n", outputFile)
}

// loadEndpointsFromRoutes loads endpoints from route scanner output
func loadEndpointsFromRoutes() []string {
	// Look for the most recent route scanner output
	files, err := os.ReadDir("reports")
	if err != nil {
		fmt.Printf("Warning: Could not read reports directory: %v\n", err)
		return getDefaultEndpoints()
	}

	var latestFile string
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "discovered_routes_") && strings.HasSuffix(file.Name(), ".json") {
			if latestFile == "" || file.Name() > latestFile {
				latestFile = file.Name()
			}
		}
	}

	if latestFile == "" {
		fmt.Printf("Warning: No route scanner output found, using default endpoints\n")
		return getDefaultEndpoints()
	}

	data, err := os.ReadFile("reports/" + latestFile)
	if err != nil {
		fmt.Printf("Warning: Could not read route file: %v\n", err)
		return getDefaultEndpoints()
	}

	var routeData struct {
		Routes []struct {
			Path string `json:"path"`
		} `json:"routes"`
	}

	if err := json.Unmarshal(data, &routeData); err != nil {
		fmt.Printf("Warning: Could not parse route file: %v\n", err)
		return getDefaultEndpoints()
	}

	endpoints := make([]string, 0, len(routeData.Routes))
	for _, route := range routeData.Routes {
		// Skip endpoints that require parameters or are not suitable for GET requests
		if !strings.Contains(route.Path, "/") ||
			strings.Contains(route.Path, "execute") ||
			strings.Contains(route.Path, "reboot") ||
			strings.Contains(route.Path, "shutdown") {
			continue
		}
		endpoints = append(endpoints, route.Path)
	}

	fmt.Printf("📁 Loaded %d endpoints from %s\n", len(endpoints), latestFile)
	return endpoints
}

// loadEndpointsFromFile loads endpoints from a JSON file
func loadEndpointsFromFile(filename string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var endpoints []string
	if err := json.Unmarshal(data, &endpoints); err != nil {
		return nil, err
	}

	return endpoints, nil
}

// getDefaultEndpoints returns a default set of endpoints for testing
func getDefaultEndpoints() []string {
	return []string{
		"/api/v1/health",
		"/api/v1/system/info",
		"/api/v1/system/cpu",
		"/api/v1/system/memory",
		"/api/v1/storage/array",
		"/api/v1/docker/info",
		"/api/v1/vms",
		"/api/v1/notifications",
		"/api/v1/operations",
	}
}
