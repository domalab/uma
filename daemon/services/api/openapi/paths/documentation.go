package paths

// GetDocumentationPaths returns all documentation API paths
func GetDocumentationPaths() map[string]interface{} {
	return map[string]interface{}{
		"/api/v1/docs":        getDocsPath(),
		"/api/v1/openapi.json": getOpenAPISpecPath(),
	}
}

func getDocsPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get API documentation",
			"description": "Retrieve interactive API documentation (Swagger UI)",
			"operationId": "getAPIDocs",
			"tags":        []string{"Documentation"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "API documentation page",
					"content": map[string]interface{}{
						"text/html": map[string]interface{}{
							"schema": map[string]interface{}{
								"type":        "string",
								"description": "HTML page with interactive API documentation",
								"example":     "<!DOCTYPE html><html>...</html>",
							},
						},
					},
				},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
		},
	}
}

func getOpenAPISpecPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get OpenAPI specification",
			"description": "Retrieve the complete OpenAPI 3.0 specification for the UMA API",
			"operationId": "getOpenAPISpec",
			"tags":        []string{"Documentation"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "OpenAPI specification",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/OpenAPISpec",
							},
						},
					},
				},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
		},
	}
}
