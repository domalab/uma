package paths

// GetAuthPaths returns all authentication API paths
func GetAuthPaths() map[string]interface{} {
	return map[string]interface{}{
		"/api/v1/auth/login":    getLoginPath(),
		"/api/v1/auth/logout":   getLogoutPath(),
		"/api/v1/auth/refresh":  getRefreshPath(),
		"/api/v1/auth/me":       getUserInfoPath(),
		"/api/v1/auth/sessions": getSessionsPath(),
		"/api/v1/auth/apikeys":  getAPIKeysPath(),
	}
}

func getLoginPath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "User login",
			"description": "Authenticate user and obtain JWT tokens",
			"operationId": "login",
			"tags":        []string{"Authentication"},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/LoginRequest",
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Login successful",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/LoginResponse",
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"429": map[string]interface{}{"$ref": "#/components/responses/TooManyRequests"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
		},
	}
}

func getRefreshPath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Refresh token",
			"description": "Refresh JWT access token using refresh token",
			"operationId": "refreshToken",
			"tags":        []string{"Authentication"},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/RefreshRequest",
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Token refreshed successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/TokenResponse",
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
		},
	}
}

func getLogoutPath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "User logout",
			"description": "Logout user and invalidate session",
			"operationId": "logout",
			"tags":        []string{"Authentication"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Logout successful",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/SuccessResponse",
							},
						},
					},
				},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
			},
		},
	}
}

func getUserInfoPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get current user information",
			"description": "Retrieve information about the currently authenticated user",
			"operationId": "getCurrentUser",
			"tags":        []string{"Authentication"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "User information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/UserInfo",
											},
										},
									},
								},
							},
						},
					},
				},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getSessionsPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "List user sessions",
			"description": "Retrieve a list of active user sessions",
			"operationId": "listSessions",
			"tags":        []string{"Authentication"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Sessions retrieved successfully",
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
													"$ref": "#/components/schemas/SessionInfo",
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
			},
		},
	}
}

func getAPIKeysPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "List API keys",
			"description": "Retrieve a list of API keys for the current user",
			"operationId": "listAPIKeys",
			"tags":        []string{"Authentication"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "API keys retrieved successfully",
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
													"$ref": "#/components/schemas/APIKeyInfo",
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
			},
		},
	}
}
