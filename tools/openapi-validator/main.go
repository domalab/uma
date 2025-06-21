package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type OpenAPISpec struct {
	OpenAPI    string                 `json:"openapi"`
	Info       map[string]interface{} `json:"info"`
	Paths      map[string]interface{} `json:"paths"`
	Components map[string]interface{} `json:"components"`
}

type ValidationResult struct {
	TotalPaths        int
	DocumentedPaths   int
	UndocumentedPaths []string
	Issues            []string
	Warnings          []string
	Coverage          float64
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <openapi-spec.json>")
		os.Exit(1)
	}

	specFile := os.Args[1]

	// Read OpenAPI spec
	data, err := os.ReadFile(specFile)
	if err != nil {
		fmt.Printf("❌ Error reading OpenAPI spec: %v\n", err)
		os.Exit(1)
	}

	var spec OpenAPISpec
	if err := json.Unmarshal(data, &spec); err != nil {
		fmt.Printf("❌ Error parsing OpenAPI spec: %v\n", err)
		os.Exit(1)
	}

	// Validate the specification
	result := validateOpenAPISpec(spec)

	// Print results
	printValidationResults(result)
}

func validateOpenAPISpec(spec OpenAPISpec) ValidationResult {
	result := ValidationResult{
		TotalPaths:        len(spec.Paths),
		DocumentedPaths:   0,
		UndocumentedPaths: []string{},
		Issues:            []string{},
		Warnings:          []string{},
	}

	// Check basic structure
	if spec.OpenAPI == "" {
		result.Issues = append(result.Issues, "OpenAPI version is missing")
	}

	if spec.Info == nil {
		result.Issues = append(result.Issues, "Info section is missing")
	} else {
		if title, ok := spec.Info["title"].(string); !ok || title == "" {
			result.Issues = append(result.Issues, "API title is missing")
		}
		if version, ok := spec.Info["version"].(string); !ok || version == "" {
			result.Issues = append(result.Issues, "API version is missing")
		}
	}

	// Validate paths
	for pathName, pathItem := range spec.Paths {
		if pathItem == nil {
			result.Issues = append(result.Issues, fmt.Sprintf("Path %s is null", pathName))
			continue
		}

		pathMap, ok := pathItem.(map[string]interface{})
		if !ok {
			result.Issues = append(result.Issues, fmt.Sprintf("Path %s is not a valid object", pathName))
			continue
		}

		hasValidOperation := false
		httpMethods := []string{"get", "post", "put", "delete", "patch", "options", "head"}

		for _, method := range httpMethods {
			if operation, exists := pathMap[method]; exists {
				hasValidOperation = true

				opMap, ok := operation.(map[string]interface{})
				if !ok {
					result.Issues = append(result.Issues, fmt.Sprintf("Operation %s %s is not a valid object", method, pathName))
					continue
				}

				// Check required operation fields
				if summary, exists := opMap["summary"]; !exists || summary == "" {
					result.Warnings = append(result.Warnings, fmt.Sprintf("Operation %s %s missing summary", method, pathName))
				}

				if description, exists := opMap["description"]; !exists || description == "" {
					result.Warnings = append(result.Warnings, fmt.Sprintf("Operation %s %s missing description", method, pathName))
				}

				if responses, exists := opMap["responses"]; !exists {
					result.Issues = append(result.Issues, fmt.Sprintf("Operation %s %s missing responses", method, pathName))
				} else {
					respMap, ok := responses.(map[string]interface{})
					if !ok || len(respMap) == 0 {
						result.Issues = append(result.Issues, fmt.Sprintf("Operation %s %s has no response definitions", method, pathName))
					}
				}

				if operationId, exists := opMap["operationId"]; !exists || operationId == "" {
					result.Warnings = append(result.Warnings, fmt.Sprintf("Operation %s %s missing operationId", method, pathName))
				}

				if tags, exists := opMap["tags"]; !exists {
					result.Warnings = append(result.Warnings, fmt.Sprintf("Operation %s %s missing tags", method, pathName))
				} else {
					tagArray, ok := tags.([]interface{})
					if !ok || len(tagArray) == 0 {
						result.Warnings = append(result.Warnings, fmt.Sprintf("Operation %s %s has empty tags", method, pathName))
					}
				}
			}
		}

		if hasValidOperation {
			result.DocumentedPaths++
		} else {
			result.UndocumentedPaths = append(result.UndocumentedPaths, pathName)
		}
	}

	// Calculate coverage
	if result.TotalPaths > 0 {
		result.Coverage = float64(result.DocumentedPaths) / float64(result.TotalPaths) * 100
	}

	return result
}

func printValidationResults(result ValidationResult) {
	fmt.Printf("🔍 OpenAPI Specification Validation Report\n")
	fmt.Printf("==========================================\n\n")

	fmt.Printf("📊 Coverage Statistics:\n")
	fmt.Printf("  Total Paths: %d\n", result.TotalPaths)
	fmt.Printf("  Documented Paths: %d\n", result.DocumentedPaths)
	fmt.Printf("  Coverage: %.1f%%\n\n", result.Coverage)

	if len(result.Issues) > 0 {
		fmt.Printf("🚨 Critical Issues (%d):\n", len(result.Issues))
		for i, issue := range result.Issues {
			fmt.Printf("  %d. %s\n", i+1, issue)
		}
		fmt.Println()
	}

	if len(result.Warnings) > 0 {
		fmt.Printf("⚠️  Warnings (%d):\n", len(result.Warnings))
		for i, warning := range result.Warnings {
			fmt.Printf("  %d. %s\n", i+1, warning)
		}
		fmt.Println()
	}

	if len(result.UndocumentedPaths) > 0 {
		fmt.Printf("📝 Undocumented Paths (%d):\n", len(result.UndocumentedPaths))
		for i, path := range result.UndocumentedPaths {
			fmt.Printf("  %d. %s\n", i+1, path)
		}
		fmt.Println()
	}

	// Summary
	if len(result.Issues) == 0 && len(result.Warnings) == 0 {
		fmt.Printf("✅ OpenAPI specification is valid with %.1f%% coverage!\n", result.Coverage)
	} else if len(result.Issues) == 0 {
		fmt.Printf("✅ OpenAPI specification is valid with %d warnings (%.1f%% coverage)\n", len(result.Warnings), result.Coverage)
	} else {
		fmt.Printf("❌ OpenAPI specification has %d critical issues and %d warnings\n", len(result.Issues), len(result.Warnings))
	}

	// Coverage assessment
	if result.Coverage >= 90 {
		fmt.Printf("🎯 Excellent coverage! Target achieved.\n")
	} else if result.Coverage >= 80 {
		fmt.Printf("🎯 Good coverage, approaching target of 90%%.\n")
	} else {
		fmt.Printf("🎯 Coverage below target. Need to document %d more paths to reach 90%%.\n",
			int(float64(result.TotalPaths)*0.9)-result.DocumentedPaths)
	}
}
