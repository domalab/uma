package paths

// GetSharesPaths returns all shares API paths
func GetSharesPaths() map[string]interface{} {
	return map[string]interface{}{
		"/api/v1/shares":  getSharesPath(),
		"/api/v1/shares/": getSharesRootPath(),
	}
}

func getSharesPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "List network shares",
			"description": "Retrieve a list of configured network shares including SMB, NFS, and AFP shares",
			"operationId": "listShares",
			"tags":        []string{"Shares"},
			"parameters": []map[string]interface{}{
				{
					"name":        "protocol",
					"in":          "query",
					"description": "Filter shares by protocol",
					"required":    false,
					"schema": map[string]interface{}{
						"type": "string",
						"enum": []string{"smb", "nfs", "afp", "ftp"},
					},
				},
				{
					"name":        "enabled_only",
					"in":          "query",
					"description": "Show only enabled shares",
					"required":    false,
					"schema": map[string]interface{}{
						"type":    "boolean",
						"default": false,
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Shares retrieved successfully",
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
													"$ref": "#/components/schemas/ShareInfo",
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
			"summary":     "Create a new share",
			"description": "Create a new network share configuration",
			"operationId": "createShare",
			"tags":        []string{"Shares"},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/ShareCreate",
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"201": map[string]interface{}{
					"description": "Share created successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/ShareInfo",
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
					"description": "Share already exists",
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

func getSharesRootPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "List shares (alternative endpoint)",
			"description": "Alternative endpoint for listing shares with trailing slash",
			"operationId": "listSharesAlt",
			"tags":        []string{"Shares"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Shares retrieved successfully",
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
													"$ref": "#/components/schemas/ShareInfo",
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
