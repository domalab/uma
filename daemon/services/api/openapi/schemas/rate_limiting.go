package schemas

// GetRateLimitingSchemas returns all rate limiting schemas
func GetRateLimitingSchemas() map[string]interface{} {
	return map[string]interface{}{
		"RateLimitConfig":       GetRateLimitConfig(),
		"RateLimitStats":        GetRateLimitStats(),
		"RateLimitConfigUpdate": GetRateLimitConfigUpdate(),
	}
}

// GetRateLimitConfig returns the schema for rate limiting configuration
func GetRateLimitConfig() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"additionalProperties": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"requests": map[string]interface{}{
					"type":        "integer",
					"minimum":     1,
					"maximum":     1000,
					"description": "Number of requests allowed",
					"example":     60,
				},
				"window": map[string]interface{}{
					"type":        "string",
					"pattern":     "^(\\d+(\\.\\d+)?(ns|us|µs|ms|s|m|h))+$",
					"description": "Time window for the rate limit (Go duration format)",
					"example":     "1m0s",
				},
			},
			"required": []string{"requests", "window"},
		},
		"description": "Rate limiting configuration by operation type",
		"example": map[string]interface{}{
			"smart_data": map[string]interface{}{
				"requests": 2,
				"window":   "2m0s",
			},
			"docker_bulk": map[string]interface{}{
				"requests": 10,
				"window":   "1m0s",
			},
		},
	}
}

// GetRateLimitStats returns the schema for rate limiting statistics
func GetRateLimitStats() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"general_rate_limiter": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"type": map[string]interface{}{
						"type":    "string",
						"example": "general",
					},
					"total_requests": map[string]interface{}{
						"type":        "integer",
						"description": "Total number of requests processed",
						"example":     1250,
						"minimum":     0,
					},
					"blocked_requests": map[string]interface{}{
						"type":        "integer",
						"description": "Number of requests blocked by rate limiting",
						"example":     15,
						"minimum":     0,
					},
				},
				"description": "General rate limiter statistics",
			},
			"operation_rate_limiter": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"type": map[string]interface{}{
						"type":    "string",
						"example": "operation_specific",
					},
					"total_clients": map[string]interface{}{
						"type":        "integer",
						"description": "Total number of tracked clients",
						"example":     5,
						"minimum":     0,
					},
					"operation_limits": map[string]interface{}{
						"type": "object",
						"additionalProperties": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"requests": map[string]interface{}{
									"type":        "integer",
									"description": "Number of requests allowed",
									"example":     60,
								},
								"window": map[string]interface{}{
									"type":        "string",
									"description": "Time window for the rate limit",
									"example":     "1m0s",
								},
							},
						},
						"description": "Current rate limits by operation type",
					},
					"client_stats": map[string]interface{}{
						"type": "object",
						"additionalProperties": map[string]interface{}{
							"type": "object",
							"additionalProperties": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"tokens_remaining": map[string]interface{}{
										"type":        "integer",
										"minimum":     0,
										"description": "Number of tokens remaining for this client",
										"example":     45,
									},
									"max_tokens": map[string]interface{}{
										"type":        "integer",
										"minimum":     1,
										"description": "Maximum number of tokens for this operation type",
										"example":     60,
									},
									"requests_made": map[string]interface{}{
										"type":        "integer",
										"minimum":     0,
										"description": "Number of requests made in current window",
										"example":     15,
									},
									"last_request": map[string]interface{}{
										"type":        "string",
										"format":      "date-time",
										"description": "Timestamp of last request",
										"example":     "2025-06-20T14:30:00Z",
									},
								},
							},
						},
						"description": "Per-client rate limiting statistics",
					},
				},
			},
			"time_range": map[string]interface{}{
				"type":        "string",
				"description": "Time range for these statistics",
				"example":     "24h",
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-20T14:30:00Z",
			},
		},
		"required": []string{"general_rate_limiter", "operation_rate_limiter", "last_updated"},
	}
}

// GetRateLimitStatsResponse returns the schema for rate limiting statistics response
func GetRateLimitStatsResponse() map[string]interface{} {
	return map[string]interface{}{
		"allOf": []map[string]interface{}{
			GetStandardResponse(),
			{
				"type": "object",
				"properties": map[string]interface{}{
					"data": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"general_rate_limiter": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"type": map[string]interface{}{
										"type":    "string",
										"example": "general",
									},
								},
								"description": "General rate limiter statistics",
							},
							"operation_rate_limiter": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"type": map[string]interface{}{
										"type":    "string",
										"example": "operation_specific",
									},
									"total_clients": map[string]interface{}{
										"type":        "integer",
										"description": "Total number of tracked clients",
									},
									"operation_limits": map[string]interface{}{
										"type": "object",
										"additionalProperties": map[string]interface{}{
											"type": "object",
											"properties": map[string]interface{}{
												"requests": map[string]interface{}{
													"type":        "integer",
													"description": "Number of requests allowed",
												},
												"window": map[string]interface{}{
													"type":        "string",
													"description": "Time window for the rate limit",
													"example":     "1m0s",
												},
											},
										},
										"description": "Current rate limits by operation type",
									},
									"client_stats": map[string]interface{}{
										"type": "object",
										"additionalProperties": map[string]interface{}{
											"type": "object",
											"additionalProperties": map[string]interface{}{
												"type": "object",
												"properties": map[string]interface{}{
													"tokens_remaining": map[string]interface{}{
														"type":        "integer",
														"minimum":     0,
														"description": "Number of tokens remaining for this client",
													},
													"max_tokens": map[string]interface{}{
														"type":        "integer",
														"minimum":     1,
														"description": "Maximum number of tokens for this operation type",
													},
												},
											},
										},
										"description": "Per-client rate limiting statistics",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// GetRateLimitConfigResponse returns the schema for rate limiting configuration response
func GetRateLimitConfigResponse() map[string]interface{} {
	return map[string]interface{}{
		"allOf": []map[string]interface{}{
			GetStandardResponse(),
			{
				"type": "object",
				"properties": map[string]interface{}{
					"data": map[string]interface{}{
						"type": "object",
						"additionalProperties": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"requests": map[string]interface{}{
									"type":        "integer",
									"minimum":     1,
									"maximum":     1000,
									"description": "Number of requests allowed",
									"example":     60,
								},
								"window": map[string]interface{}{
									"type":        "string",
									"pattern":     "^(\\d+(\\.\\d+)?(ns|us|µs|ms|s|m|h))+$",
									"description": "Time window for the rate limit (Go duration format)",
									"example":     "1m0s",
								},
							},
							"required": []string{"requests", "window"},
						},
						"description": "Rate limiting configuration by operation type",
					},
				},
			},
		},
	}
}

// GetRateLimitConfigUpdate returns the schema for rate limiting configuration update request
func GetRateLimitConfigUpdate() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"additionalProperties": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"requests": map[string]interface{}{
					"type":        "integer",
					"minimum":     1,
					"maximum":     1000,
					"description": "Number of requests allowed",
					"example":     60,
				},
				"window": map[string]interface{}{
					"type":        "string",
					"pattern":     "^(\\d+(\\.\\d+)?(ns|us|µs|ms|s|m|h))+$",
					"description": "Time window for the rate limit (Go duration format)",
					"example":     "1m0s",
				},
			},
			"required": []string{"requests", "window"},
		},
		"description":   "Rate limiting configuration updates by operation type",
		"minProperties": 1,
		"example": map[string]interface{}{
			"smart_data": map[string]interface{}{
				"requests": 2,
				"window":   "2m0s",
			},
			"docker_bulk": map[string]interface{}{
				"requests": 10,
				"window":   "1m0s",
			},
		},
	}
}

// GetRateLimitConfigUpdateResponse returns the schema for rate limiting configuration update response
func GetRateLimitConfigUpdateResponse() map[string]interface{} {
	return map[string]interface{}{
		"allOf": []map[string]interface{}{
			GetStandardResponse(),
			{
				"type": "object",
				"properties": map[string]interface{}{
					"data": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"message": map[string]interface{}{
								"type":    "string",
								"example": "Rate limit configuration update not implemented yet",
							},
							"note": map[string]interface{}{
								"type":    "string",
								"example": "This endpoint would require admin authentication",
							},
							"updated_operations": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"type": "string",
									"enum": []string{
										"general", "health_check", "smart_data", "parity_check",
										"array_control", "disk_info", "docker_list", "docker_control",
										"docker_bulk", "vm_list", "vm_control", "vm_bulk",
										"system_info", "system_control", "sensor_data",
										"async_create", "async_list", "async_cancel",
									},
								},
								"description": "List of operation types that were updated",
							},
						},
					},
				},
			},
		},
	}
}
