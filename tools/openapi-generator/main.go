package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/domalab/uma/daemon/services/api/openapi"
)

func main() {
	// Generate OpenAPI specification
	generator := openapi.NewGenerator(nil) // Use default config
	spec := generator.Generate()

	// Validate the specification
	errors := generator.ValidateSpec()
	if len(errors) > 0 {
		fmt.Printf("⚠️  Validation errors found:\n")
		for i, err := range errors {
			fmt.Printf("  %d. %s\n", i+1, err)
		}
		fmt.Println()
	}

	// Get statistics
	stats := generator.GetStats()
	fmt.Printf("📊 OpenAPI Specification Statistics:\n")
	fmt.Printf("  OpenAPI Version: %v\n", stats["openapi_version"])
	fmt.Printf("  API Version: %v\n", stats["api_version"])
	fmt.Printf("  Total Paths: %v\n", stats["total_paths"])
	fmt.Printf("  Total Schemas: %v\n", stats["total_schemas"])
	fmt.Printf("  Total Responses: %v\n", stats["total_responses"])
	
	if schemasByCategory, ok := stats["schemas_by_category"].(map[string]interface{}); ok {
		fmt.Printf("  Schemas by Category:\n")
		for category, count := range schemasByCategory {
			fmt.Printf("    %s: %v\n", category, count)
		}
	}
	
	if featuresEnabled, ok := stats["features_enabled"].(map[string]interface{}); ok {
		fmt.Printf("  Features Enabled:\n")
		for feature, enabled := range featuresEnabled {
			fmt.Printf("    %s: %v\n", feature, enabled)
		}
	}
	fmt.Println()

	// Serialize to JSON
	jsonData, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		fmt.Printf("❌ Error serializing OpenAPI spec: %v\n", err)
		os.Exit(1)
	}

	// Write to file
	outputFile := "openapi-spec.json"
	if len(os.Args) > 1 {
		outputFile = os.Args[1]
	}

	err = os.WriteFile(outputFile, jsonData, 0644)
	if err != nil {
		fmt.Printf("❌ Error writing OpenAPI spec to file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ OpenAPI specification generated successfully: %s\n", outputFile)
	fmt.Printf("📄 File size: %d bytes\n", len(jsonData))
}
