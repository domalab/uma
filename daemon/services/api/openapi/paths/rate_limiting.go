package paths

// GetRateLimitingPaths returns all rate limiting API paths
func GetRateLimitingPaths() map[string]interface{} {
	return map[string]interface{}{
		"/api/v1/rate-limits/config": getRateLimitConfigPath(),
		"/api/v1/rate-limits/stats":  getRateLimitStatsPath(),
	}
}

func getRateLimitConfigPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get rate limiting configuration",
			"description": "Retrieve current rate limiting configuration and rules",
			"operationId": "getRateLimitConfig",
			"tags":        []string{"Rate Limiting"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Rate limiting configuration retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/RateLimitConfig",
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
		"put": map[string]interface{}{
			"summary":     "Update rate limiting configuration",
			"description": "Update rate limiting configuration and rules",
			"operationId": "updateRateLimitConfig",
			"tags":        []string{"Rate Limiting"},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/RateLimitConfigUpdate",
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Rate limiting configuration updated successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/RateLimitConfig",
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
				"422": map[string]interface{}{"$ref": "#/components/responses/ValidationError"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
			},
		},
	}
}

func getRateLimitStatsPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get rate limiting statistics",
			"description": "Retrieve rate limiting statistics including current usage and blocked requests",
			"operationId": "getRateLimitStats",
			"tags":        []string{"Rate Limiting"},
			"parameters": []map[string]interface{}{
				{
					"name":        "time_range",
					"in":          "query",
					"description": "Time range for statistics",
					"required":    false,
					"schema": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"1h", "24h", "7d", "30d"},
						"default": "24h",
					},
				},
				{
					"name":        "client_ip",
					"in":          "query",
					"description": "Filter statistics by client IP address",
					"required":    false,
					"schema": map[string]interface{}{
						"type":   "string",
						"format": "ipv4",
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Rate limiting statistics retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/RateLimitStats",
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
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
			},
		},
		"delete": map[string]interface{}{
			"summary":     "Reset rate limiting statistics",
			"description": "Reset rate limiting statistics and counters",
			"operationId": "resetRateLimitStats",
			"tags":        []string{"Rate Limiting"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Rate limiting statistics reset successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/StandardResponse",
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
