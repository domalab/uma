package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
)

// SchemaValidator validates OpenAPI schemas against live API responses
type SchemaValidator struct {
	baseURL    string
	client     *http.Client
	violations []ValidationViolation
}

// ValidationViolation represents a schema-API mismatch
type ValidationViolation struct {
	Endpoint    string `json:"endpoint"`
	Field       string `json:"field"`
	SchemaType  string `json:"schema_type"`
	ActualType  string `json:"actual_type"`
	SchemaValue string `json:"schema_value,omitempty"`
	ActualValue string `json:"actual_value,omitempty"`
	Severity    string `json:"severity"` // CRITICAL, WARNING, INFO
	Message     string `json:"message"`
}

// NewSchemaValidator creates a new schema validator
func NewSchemaValidator(baseURL string) *SchemaValidator {
	return &SchemaValidator{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		violations: []ValidationViolation{},
	}
}

// ValidateEndpoint validates a single endpoint against its schema
func (sv *SchemaValidator) ValidateEndpoint(endpoint string, schemaName string) error {
	// Get OpenAPI spec
	spec, err := sv.getOpenAPISpec()
	if err != nil {
		return fmt.Errorf("failed to get OpenAPI spec: %v", err)
	}

	// Get live API response
	response, err := sv.getAPIResponse(endpoint)
	if err != nil {
		return fmt.Errorf("failed to get API response for %s: %v", endpoint, err)
	}

	// Get schema definition
	schema, err := sv.getSchemaDefinition(spec, schemaName)
	if err != nil {
		return fmt.Errorf("failed to get schema for %s: %v", schemaName, err)
	}

	// Validate response against schema based on schema type
	schemaType, ok := schema["type"].(string)
	if !ok {
		sv.addViolation(endpoint, "", "", "", "", "", "CRITICAL", "Schema has no type defined")
		return nil
	}

	switch schemaType {
	case "array":
		// Handle array response validation
		if err := sv.validateArrayResponse(endpoint, response, schema); err != nil {
			sv.addViolation(endpoint, "", "array", "", "", "", "CRITICAL", err.Error())
		}
	case "object":
		// Handle object response validation
		if responseObj, ok := response.(map[string]interface{}); ok {
			sv.validateResponseAgainstSchema(endpoint, responseObj, schema)
		} else {
			sv.addViolation(endpoint, "", "object", sv.getJSONType(response), "", "", "CRITICAL",
				fmt.Sprintf("Expected object response but got %s", sv.getJSONType(response)))
		}
	default:
		sv.addViolation(endpoint, "", schemaType, sv.getJSONType(response), "", "", "WARNING",
			fmt.Sprintf("Unsupported schema type: %s", schemaType))
	}

	return nil
}

// getOpenAPISpec fetches the OpenAPI specification
func (sv *SchemaValidator) getOpenAPISpec() (map[string]interface{}, error) {
	resp, err := sv.client.Get(sv.baseURL + "/api/v1/openapi.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(body, &spec); err != nil {
		return nil, err
	}

	return spec, nil
}

// getAPIResponse fetches live API response and returns it as interface{}
func (sv *SchemaValidator) getAPIResponse(endpoint string) (interface{}, error) {
	resp, err := sv.client.Get(sv.baseURL + endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response, nil
}

// getSchemaDefinition extracts schema definition from OpenAPI spec
func (sv *SchemaValidator) getSchemaDefinition(spec map[string]interface{}, schemaName string) (map[string]interface{}, error) {
	components, ok := spec["components"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no components found in spec")
	}

	schemas, ok := components["schemas"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no schemas found in components")
	}

	schema, ok := schemas[schemaName].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("schema %s not found", schemaName)
	}

	return schema, nil
}

// validateArrayResponse validates array API response against array schema
func (sv *SchemaValidator) validateArrayResponse(endpoint string, response interface{}, schema map[string]interface{}) error {
	// Check if response is actually an array
	responseArray, ok := response.([]interface{})
	if !ok {
		return fmt.Errorf("expected array response but got %s", sv.getJSONType(response))
	}

	// Get the items schema from the array schema
	items, ok := schema["items"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("array schema missing items definition")
	}

	// Check if items schema has a $ref
	if ref, hasRef := items["$ref"].(string); hasRef {
		// Extract schema name from $ref (e.g., "#/components/schemas/ContainerInfo" -> "ContainerInfo")
		parts := strings.Split(ref, "/")
		if len(parts) > 0 {
			itemSchemaName := parts[len(parts)-1]

			// Get the referenced schema from the OpenAPI spec
			spec, err := sv.getOpenAPISpec()
			if err != nil {
				return fmt.Errorf("failed to get OpenAPI spec for item validation: %v", err)
			}

			itemSchema, err := sv.getSchemaDefinition(spec, itemSchemaName)
			if err != nil {
				return fmt.Errorf("failed to get item schema %s: %v", itemSchemaName, err)
			}

			// Validate each item in the array against the item schema
			for i, item := range responseArray {
				if itemObj, ok := item.(map[string]interface{}); ok {
					sv.validateResponseAgainstSchema(fmt.Sprintf("%s[%d]", endpoint, i), itemObj, itemSchema)
				} else {
					sv.addViolation(endpoint, fmt.Sprintf("[%d]", i), "object", sv.getJSONType(item), "", "", "CRITICAL",
						fmt.Sprintf("Array item %d: expected object but got %s", i, sv.getJSONType(item)))
				}
			}
		}
	} else {
		// Direct items schema (not a reference)
		for i, item := range responseArray {
			if itemObj, ok := item.(map[string]interface{}); ok {
				sv.validateResponseAgainstSchema(fmt.Sprintf("%s[%d]", endpoint, i), itemObj, items)
			} else {
				sv.addViolation(endpoint, fmt.Sprintf("[%d]", i), "object", sv.getJSONType(item), "", "", "CRITICAL",
					fmt.Sprintf("Array item %d: expected object but got %s", i, sv.getJSONType(item)))
			}
		}
	}

	// Validate array-level constraints
	if minItems, ok := schema["minItems"].(float64); ok {
		if float64(len(responseArray)) < minItems {
			sv.addViolation(endpoint, "", "minItems", "constraint", fmt.Sprintf("%.0f", minItems),
				fmt.Sprintf("%d", len(responseArray)), "CRITICAL",
				fmt.Sprintf("Array has %d items, minimum required: %.0f", len(responseArray), minItems))
		}
	}

	if maxItems, ok := schema["maxItems"].(float64); ok {
		if float64(len(responseArray)) > maxItems {
			sv.addViolation(endpoint, "", "maxItems", "constraint", fmt.Sprintf("%.0f", maxItems),
				fmt.Sprintf("%d", len(responseArray)), "CRITICAL",
				fmt.Sprintf("Array has %d items, maximum allowed: %.0f", len(responseArray), maxItems))
		}
	}

	return nil
}

// validateResponseAgainstSchema validates API response against schema
func (sv *SchemaValidator) validateResponseAgainstSchema(endpoint string, response, schema map[string]interface{}) {
	// Get schema properties
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		sv.addViolation(endpoint, "", "", "", "", "", "CRITICAL", "Schema has no properties defined")
		return
	}

	// Get required fields
	requiredFields := []string{}
	if required, ok := schema["required"].([]interface{}); ok {
		for _, field := range required {
			if fieldStr, ok := field.(string); ok {
				requiredFields = append(requiredFields, fieldStr)
			}
		}
	}

	// Check required fields exist in response
	for _, field := range requiredFields {
		if _, exists := response[field]; !exists {
			sv.addViolation(endpoint, field, "required", "missing", "", "", "CRITICAL",
				fmt.Sprintf("Required field '%s' missing from API response", field))
		}
	}

	// Check each response field against schema
	for field, value := range response {
		if fieldSchema, exists := properties[field]; exists {
			sv.validateField(endpoint, field, value, fieldSchema.(map[string]interface{}))
		} else {
			sv.addViolation(endpoint, field, "", "", "", fmt.Sprintf("%v", value), "WARNING",
				fmt.Sprintf("Field '%s' exists in API response but not in schema", field))
		}
	}

	// Check for schema fields not in response
	for field := range properties {
		if _, exists := response[field]; !exists {
			// Only report if it's a required field (already handled above) or if it should have a default
			if !sv.isRequired(field, requiredFields) {
				sv.addViolation(endpoint, field, "optional", "missing", "", "", "INFO",
					fmt.Sprintf("Optional field '%s' defined in schema but not in API response", field))
			}
		}
	}
}

// validateField validates a single field against its schema definition
func (sv *SchemaValidator) validateField(endpoint, field string, value interface{}, fieldSchema map[string]interface{}) {
	// Get expected type
	expectedType, ok := fieldSchema["type"].(string)
	if !ok {
		return // Skip validation if no type defined
	}

	// Get actual type
	actualType := sv.getJSONType(value)

	// Check type match
	if !sv.typesMatch(expectedType, actualType) {
		sv.addViolation(endpoint, field, expectedType, actualType, "", fmt.Sprintf("%v", value), "CRITICAL",
			fmt.Sprintf("Type mismatch: expected %s, got %s", expectedType, actualType))
	}

	// Validate examples if present
	if example, exists := fieldSchema["example"]; exists {
		sv.validateExample(endpoint, field, value, example)
	}

	// Validate enums if present
	if enumValues, exists := fieldSchema["enum"]; exists {
		sv.validateEnum(endpoint, field, value, enumValues)
	}

	// Validate numeric constraints
	if expectedType == "number" || expectedType == "integer" {
		sv.validateNumericConstraints(endpoint, field, value, fieldSchema)
	}
}

// getJSONType returns the JSON type of a Go value
func (sv *SchemaValidator) getJSONType(value interface{}) string {
	if value == nil {
		return "null"
	}

	switch value := value.(type) {
	case bool:
		return "boolean"
	case float64:
		// Check if it's actually an integer
		if value == float64(int64(value)) {
			return "integer"
		}
		return "number"
	case string:
		return "string"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "unknown"
	}
}

// typesMatch checks if schema type matches actual type
func (sv *SchemaValidator) typesMatch(schemaType, actualType string) bool {
	if schemaType == actualType {
		return true
	}

	// Number can be integer
	if schemaType == "number" && actualType == "integer" {
		return true
	}

	return false
}

// validateExample validates field value against schema example
func (sv *SchemaValidator) validateExample(endpoint, field string, value, example interface{}) {
	// Type check
	if reflect.TypeOf(value) != reflect.TypeOf(example) {
		sv.addViolation(endpoint, field, fmt.Sprintf("%T", example), fmt.Sprintf("%T", value),
			fmt.Sprintf("%v", example), fmt.Sprintf("%v", value), "WARNING",
			"Value type doesn't match example type")
	}
}

// validateEnum validates field value against enum constraints
func (sv *SchemaValidator) validateEnum(endpoint, field string, value interface{}, enumValues interface{}) {
	if enumSlice, ok := enumValues.([]interface{}); ok {
		for _, enumValue := range enumSlice {
			if value == enumValue {
				return // Valid enum value
			}
		}
		sv.addViolation(endpoint, field, "enum", fmt.Sprintf("%T", value),
			fmt.Sprintf("%v", enumValues), fmt.Sprintf("%v", value), "CRITICAL",
			fmt.Sprintf("Value '%v' not in allowed enum values", value))
	}
}

// validateNumericConstraints validates numeric field constraints
func (sv *SchemaValidator) validateNumericConstraints(endpoint, field string, value interface{}, fieldSchema map[string]interface{}) {
	numValue, ok := value.(float64)
	if !ok {
		return
	}

	if minimum, exists := fieldSchema["minimum"]; exists {
		if min, ok := minimum.(float64); ok && numValue < min {
			sv.addViolation(endpoint, field, "minimum", "constraint", fmt.Sprintf("%.2f", min),
				fmt.Sprintf("%.2f", numValue), "CRITICAL",
				fmt.Sprintf("Value %.2f below minimum %.2f", numValue, min))
		}
	}

	if maximum, exists := fieldSchema["maximum"]; exists {
		if max, ok := maximum.(float64); ok && numValue > max {
			sv.addViolation(endpoint, field, "maximum", "constraint", fmt.Sprintf("%.2f", max),
				fmt.Sprintf("%.2f", numValue), "CRITICAL",
				fmt.Sprintf("Value %.2f above maximum %.2f", numValue, max))
		}
	}
}

// isRequired checks if a field is in the required list
func (sv *SchemaValidator) isRequired(field string, requiredFields []string) bool {
	for _, required := range requiredFields {
		if field == required {
			return true
		}
	}
	return false
}

// addViolation adds a validation violation
func (sv *SchemaValidator) addViolation(endpoint, field, schemaType, actualType, schemaValue, actualValue, severity, message string) {
	sv.violations = append(sv.violations, ValidationViolation{
		Endpoint:    endpoint,
		Field:       field,
		SchemaType:  schemaType,
		ActualType:  actualType,
		SchemaValue: schemaValue,
		ActualValue: actualValue,
		Severity:    severity,
		Message:     message,
	})
}

// checkWebSocketEndpoint checks if a WebSocket endpoint is available
func (sv *SchemaValidator) checkWebSocketEndpoint(endpoint string) bool {
	// For WebSocket endpoints, we can check if they're documented in the OpenAPI spec
	// and if the base HTTP endpoint responds (WebSocket upgrade happens at runtime)

	// Try to connect to the WebSocket endpoint with regular HTTP request
	resp, err := sv.client.Get(sv.baseURL + endpoint)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// WebSocket endpoints typically return 400 or 426 for non-WebSocket requests
	// This indicates the endpoint exists but requires WebSocket upgrade
	return resp.StatusCode == 400 || resp.StatusCode == 426 || resp.StatusCode == 200
}

// checkPrometheusEndpoint checks if a Prometheus metrics endpoint is available
func (sv *SchemaValidator) checkPrometheusEndpoint(endpoint string) bool {
	// Try to connect to the metrics endpoint
	resp, err := sv.client.Get(sv.baseURL + endpoint)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Prometheus endpoints should return 200 with text/plain content type
	if resp.StatusCode != 200 {
		return false
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	return strings.Contains(contentType, "text/plain")
}

// GetViolations returns all validation violations
func (sv *SchemaValidator) GetViolations() []ValidationViolation {
	return sv.violations
}

// GetViolationsBySeverity returns violations filtered by severity
func (sv *SchemaValidator) GetViolationsBySeverity(severity string) []ValidationViolation {
	var filtered []ValidationViolation
	for _, violation := range sv.violations {
		if violation.Severity == severity {
			filtered = append(filtered, violation)
		}
	}
	return filtered
}

// PrintReport prints a validation report
func (sv *SchemaValidator) PrintReport() {
	fmt.Printf("\n🔍 API-Schema Validation Report\n")
	fmt.Printf("=====================================\n")

	critical := sv.GetViolationsBySeverity("CRITICAL")
	warnings := sv.GetViolationsBySeverity("WARNING")
	info := sv.GetViolationsBySeverity("INFO")

	fmt.Printf("🚨 Critical Issues: %d\n", len(critical))
	fmt.Printf("⚠️  Warnings: %d\n", len(warnings))
	fmt.Printf("ℹ️  Info: %d\n", len(info))
	fmt.Printf("\n")

	if len(critical) > 0 {
		fmt.Printf("🚨 CRITICAL ISSUES:\n")
		for _, violation := range critical {
			fmt.Printf("  • %s [%s]: %s\n", violation.Endpoint, violation.Field, violation.Message)
		}
		fmt.Printf("\n")
	}

	if len(warnings) > 0 {
		fmt.Printf("⚠️  WARNINGS:\n")
		for _, violation := range warnings {
			fmt.Printf("  • %s [%s]: %s\n", violation.Endpoint, violation.Field, violation.Message)
		}
		fmt.Printf("\n")
	}
}

// SaveReport saves validation report to JSON file
func (sv *SchemaValidator) SaveReport(filename string) error {
	report := map[string]interface{}{
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"violations": sv.violations,
		"summary": map[string]int{
			"total":    len(sv.violations),
			"critical": len(sv.GetViolationsBySeverity("CRITICAL")),
			"warning":  len(sv.GetViolationsBySeverity("WARNING")),
			"info":     len(sv.GetViolationsBySeverity("INFO")),
		},
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// generateCoverageReport generates a comprehensive coverage report
func generateCoverageReport(baseURL string, endpoints []struct {
	path   string
	schema string
	note   string
}) error {
	fmt.Printf("\n📋 Generating coverage report...\n")

	// Get OpenAPI spec to compare against
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(baseURL + "/api/v1/openapi.json")
	if err != nil {
		return fmt.Errorf("failed to fetch OpenAPI spec: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read OpenAPI spec: %v", err)
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(body, &spec); err != nil {
		return fmt.Errorf("failed to parse OpenAPI spec: %v", err)
	}

	// Extract all documented paths
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("no paths found in OpenAPI spec")
	}

	// Create map of validated endpoints
	validatedPaths := make(map[string]bool)
	for _, endpoint := range endpoints {
		validatedPaths[endpoint.path] = true
	}

	// Find missing endpoints
	var missingEndpoints []string
	var documentedEndpoints []string

	for path := range paths {
		documentedEndpoints = append(documentedEndpoints, path)
		if !validatedPaths[path] {
			missingEndpoints = append(missingEndpoints, path)
		}
	}

	// Calculate coverage
	totalDocumented := len(documentedEndpoints)
	totalValidated := len(endpoints)
	coveragePercent := float64(totalValidated) / float64(totalDocumented) * 100

	// Generate report
	report := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"coverage": map[string]interface{}{
			"total_documented_endpoints": totalDocumented,
			"total_validated_endpoints":  totalValidated,
			"coverage_percentage":        coveragePercent,
			"missing_endpoints":          missingEndpoints,
		},
		"documented_endpoints": documentedEndpoints,
		"validated_endpoints": func() []string {
			var validated []string
			for _, ep := range endpoints {
				validated = append(validated, ep.path)
			}
			return validated
		}(),
	}

	// Save coverage report
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal coverage report: %v", err)
	}

	if err := os.WriteFile("schema-coverage-report.json", data, 0644); err != nil {
		return fmt.Errorf("failed to save coverage report: %v", err)
	}

	// Print coverage summary
	fmt.Printf("📊 Coverage Report:\n")
	fmt.Printf("  📚 Total documented endpoints: %d\n", totalDocumented)
	fmt.Printf("  ✅ Total validated endpoints: %d\n", totalValidated)
	fmt.Printf("  📈 Coverage percentage: %.1f%%\n", coveragePercent)

	if len(missingEndpoints) > 0 {
		fmt.Printf("  ⚠️  Missing endpoints (%d):\n", len(missingEndpoints))
		for _, missing := range missingEndpoints {
			fmt.Printf("    - %s\n", missing)
		}
	} else {
		fmt.Printf("  🎉 All documented endpoints are validated!\n")
	}

	fmt.Printf("📄 Detailed coverage report saved to schema-coverage-report.json\n")

	return nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: schema-validator <base-url>")
	}

	baseURL := strings.TrimSuffix(os.Args[1], "/")
	validator := NewSchemaValidator(baseURL)

	// Define endpoints to validate - comprehensive coverage of all UMA API endpoints
	endpoints := []struct {
		path   string
		schema string
		note   string // Optional note about the endpoint
	}{
		// System endpoints
		{"/api/v1/system/cpu", "CPUInfo", "CPU information and performance metrics"},
		{"/api/v1/system/memory", "MemoryInfo", "System memory information"},
		{"/api/v1/system/temperature", "TemperatureInfo", "Temperature sensors data"},
		{"/api/v1/system/network", "NetworkInfo", "Network interface information"},
		{"/api/v1/system/ups", "UPSInfo", "UPS status and information"},
		{"/api/v1/system/gpu", "GPUInfo", "GPU information and statistics"},
		{"/api/v1/system/fans", "FanInfo", "Fan speed and status data"},
		{"/api/v1/system/resources", "SystemResources", "Combined system resources"},
		{"/api/v1/system/filesystems", "FilesystemInfo", "Filesystem information"},
		{"/api/v1/system/info", "SystemInfo", "General system information"},
		{"/api/v1/system/logs", "SystemLogs", "System log entries"},
		{"/api/v1/system/parity/disk", "ParityDiskInfo", "Parity disk information"},
		{"/api/v1/system/parity/check", "ParityCheckStatus", "Parity check status"},

		// Storage endpoints
		{"/api/v1/storage/array", "ArrayInfo", "Storage array information"},
		{"/api/v1/storage/disks", "DiskList", "List of storage disks"},
		{"/api/v1/storage/boot", "BootInfo", "Boot device information"},
		{"/api/v1/storage/cache", "CacheInfo", "Cache storage information"},
		{"/api/v1/storage/general", "StorageGeneral", "General storage information"},
		{"/api/v1/storage/zfs", "ZFSInfo", "ZFS storage information"},

		// Docker endpoints
		{"/api/v1/docker/containers", "DockerContainerList", "List of Docker containers"},
		{"/api/v1/docker/info", "DockerInfo", "Docker system information"},
		{"/api/v1/docker/images", "DockerImageList", "List of Docker images"},
		{"/api/v1/docker/networks", "DockerNetworkList", "List of Docker networks"},

		// VM endpoints
		{"/api/v1/vms", "VMList", "List of virtual machines"},

		// Health and diagnostics
		{"/api/v1/health", "HealthResponse", "API health status"},
		{"/api/v1/diagnostics/health", "DiagnosticsHealth", "System diagnostics health"},
		{"/api/v1/diagnostics/info", "DiagnosticsInfo", "Diagnostics information"},

		// Operations and async
		{"/api/v1/operations", "OperationList", "List of async operations"},
		{"/api/v1/operations/stats", "OperationStats", "Operation statistics"},

		// Notifications
		{"/api/v1/notifications", "NotificationList", "List of notifications"},
		{"/api/v1/notifications/stats", "NotificationStats", "Notification statistics"},

		// Metrics (Prometheus endpoint)
		{"/metrics", "PrometheusMetrics", "Prometheus metrics endpoint"},

		// WebSocket endpoints (documentation only - cannot validate response schemas)
		{"/api/v1/ws", "UnifiedWebSocketStream", "Unified WebSocket with subscription management"},

		// Note: Parameterized endpoints like /api/v1/vms/{name} and action endpoints
		// like /api/v1/system/reboot require special handling and are not included
		// in basic schema validation as they need specific test data or are POST operations
	}

	fmt.Printf("🔍 Validating API schemas against live endpoints...\n")
	fmt.Printf("Base URL: %s\n", baseURL)
	fmt.Printf("Total endpoints to validate: %d\n\n", len(endpoints))

	// Track validation statistics
	var validatedCount, errorCount, skippedCount int

	// Validate each endpoint
	for i, endpoint := range endpoints {
		fmt.Printf("[%d/%d] Validating %s...\n", i+1, len(endpoints), endpoint.path)
		if endpoint.note != "" {
			fmt.Printf("  📝 %s\n", endpoint.note)
		}

		// Special handling for WebSocket endpoints
		if strings.HasPrefix(endpoint.path, "/api/v1/ws") {
			fmt.Printf("  🔌 WebSocket endpoint - checking availability only\n")
			if validator.checkWebSocketEndpoint(endpoint.path) {
				fmt.Printf("  ✅ WebSocket endpoint available\n")
				validatedCount++
			} else {
				fmt.Printf("  ❌ WebSocket endpoint not available\n")
				errorCount++
			}
		} else if endpoint.path == "/metrics" {
			// Special handling for Prometheus metrics endpoint (text/plain response)
			fmt.Printf("  📊 Prometheus metrics endpoint - checking availability only\n")
			if validator.checkPrometheusEndpoint(endpoint.path) {
				fmt.Printf("  ✅ Prometheus metrics endpoint available\n")
				validatedCount++
			} else {
				fmt.Printf("  ❌ Prometheus metrics endpoint not available\n")
				errorCount++
			}
		} else {
			// Regular REST endpoint validation
			if err := validator.ValidateEndpoint(endpoint.path, endpoint.schema); err != nil {
				fmt.Printf("  ❌ Error: %v\n", err)
				errorCount++
			} else {
				fmt.Printf("  ✅ Validated\n")
				validatedCount++
			}
		}

		// Add small delay to avoid overwhelming the server
		if i < len(endpoints)-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Print validation summary
	fmt.Printf("\n📊 Validation Summary:\n")
	fmt.Printf("✅ Successfully validated: %d\n", validatedCount)
	fmt.Printf("❌ Validation errors: %d\n", errorCount)
	fmt.Printf("⏭️  Skipped: %d\n", skippedCount)
	fmt.Printf("📈 Coverage: %.1f%% (%d/%d)\n",
		float64(validatedCount)/float64(len(endpoints))*100,
		validatedCount, len(endpoints))

	// Print report
	validator.PrintReport()

	// Save detailed report
	if err := validator.SaveReport("schema-validation-report.json"); err != nil {
		log.Printf("Failed to save report: %v", err)
	} else {
		fmt.Printf("📄 Detailed report saved to schema-validation-report.json\n")
	}

	// Generate coverage report
	if err := generateCoverageReport(baseURL, endpoints); err != nil {
		log.Printf("Failed to generate coverage report: %v", err)
	}

	// Exit with error code if critical issues found
	if len(validator.GetViolationsBySeverity("CRITICAL")) > 0 {
		os.Exit(1)
	}
}
