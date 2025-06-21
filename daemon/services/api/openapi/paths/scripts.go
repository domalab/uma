package paths

// GetScriptsPaths returns all scripts API paths
func GetScriptsPaths() map[string]interface{} {
	return map[string]interface{}{
		"/api/v1/scripts":  getScriptsPath(),
		"/api/v1/scripts/": getScriptsRootPath(),
	}
}

func getScriptsPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "List available scripts",
			"description": "Retrieve a list of available user scripts and their metadata",
			"operationId": "listScripts",
			"tags":        []string{"Scripts"},
			"parameters": []map[string]interface{}{
				{
					"name":        "category",
					"in":          "query",
					"description": "Filter scripts by category",
					"required":    false,
					"schema": map[string]interface{}{
						"type": "string",
						"enum": []string{"user", "system", "maintenance", "backup", "monitoring"},
					},
				},
				{
					"name":        "enabled_only",
					"in":          "query",
					"description": "Show only enabled scripts",
					"required":    false,
					"schema": map[string]interface{}{
						"type":    "boolean",
						"default": false,
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Scripts retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"type": "array",
												"items": map[string]interface{}{
													"$ref": "#/components/schemas/ScriptInfo",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"403": map[string]interface{}{"$ref": "#/components/responses/Forbidden"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
		"post": map[string]interface{}{
			"summary":     "Create or upload a script",
			"description": "Create a new script or upload a script file",
			"operationId": "createScript",
			"tags":        []string{"Scripts"},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/ScriptCreate",
						},
					},
					"multipart/form-data": map[string]interface{}{
						"schema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"file": map[string]interface{}{
									"type":        "string",
									"format":      "binary",
									"description": "Script file to upload",
								},
								"name": map[string]interface{}{
									"type":        "string",
									"description": "Script name",
								},
								"description": map[string]interface{}{
									"type":        "string",
									"description": "Script description",
								},
								"category": map[string]interface{}{
									"type": "string",
									"enum": []string{"user", "system", "maintenance", "backup", "monitoring"},
								},
							},
							"required": []string{"file", "name"},
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"201": map[string]interface{}{
					"description": "Script created successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/ScriptInfo",
											},
										},
									},
								},
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"403": map[string]interface{}{"$ref": "#/components/responses/Forbidden"},
				"409": map[string]interface{}{
					"description": "Script already exists",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/Error",
							},
						},
					},
				},
				"422": map[string]interface{}{"$ref": "#/components/responses/ValidationError"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getScriptsRootPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "List scripts (alternative endpoint)",
			"description": "Alternative endpoint for listing scripts with trailing slash",
			"operationId": "listScriptsAlt",
			"tags":        []string{"Scripts"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Scripts retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"type": "array",
												"items": map[string]interface{}{
													"$ref": "#/components/schemas/ScriptInfo",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"403": map[string]interface{}{"$ref": "#/components/responses/Forbidden"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}
