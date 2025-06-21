package schemas

// GetDocumentationSchemas returns all documentation-related schemas
func GetDocumentationSchemas() map[string]interface{} {
	return map[string]interface{}{
		"OpenAPISpec": getOpenAPISpecSchema(),
	}
}

func getOpenAPISpecSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"openapi": map[string]interface{}{
				"type":        "string",
				"description": "OpenAPI specification version",
				"example":     "3.0.3",
			},
			"info": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type":        "string",
						"description": "API title",
						"example":     "UMA API",
					},
					"version": map[string]interface{}{
						"type":        "string",
						"description": "API version",
						"example":     "1.0.0",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "API description",
						"example":     "Unraid Management API for system monitoring and control",
					},
				},
				"required": []string{"title", "version"},
			},
			"servers": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"url": map[string]interface{}{
							"type":        "string",
							"description": "Server URL",
							"example":     "http://192.168.20.21:34600",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Server description",
							"example":     "UMA API Server",
						},
					},
					"required": []string{"url"},
				},
			},
			"paths": map[string]interface{}{
				"type":        "object",
				"description": "API paths and operations",
				"additionalProperties": map[string]interface{}{
					"type": "object",
					"description": "Path item with HTTP operations",
				},
			},
			"components": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"schemas": map[string]interface{}{
						"type":        "object",
						"description": "Reusable schema definitions",
						"additionalProperties": map[string]interface{}{
							"type": "object",
							"description": "Schema definition",
						},
					},
					"responses": map[string]interface{}{
						"type":        "object",
						"description": "Reusable response definitions",
						"additionalProperties": map[string]interface{}{
							"type": "object",
							"description": "Response definition",
						},
					},
					"parameters": map[string]interface{}{
						"type":        "object",
						"description": "Reusable parameter definitions",
						"additionalProperties": map[string]interface{}{
							"type": "object",
							"description": "Parameter definition",
						},
					},
					"securitySchemes": map[string]interface{}{
						"type":        "object",
						"description": "Security scheme definitions",
						"additionalProperties": map[string]interface{}{
							"type": "object",
							"description": "Security scheme definition",
						},
					},
				},
			},
			"tags": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Tag name",
							"example":     "System",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Tag description",
							"example":     "System monitoring and control operations",
						},
					},
					"required": []string{"name"},
				},
			},
		},
		"required": []string{"openapi", "info", "paths"},
		"description": "Complete OpenAPI 3.0 specification for the UMA API",
	}
}
