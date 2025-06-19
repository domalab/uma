package schemas

// GetAPIError returns the schema for structured API errors
func GetAPIError() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"error": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"code": map[string]interface{}{
						"type": "string",
						"enum": []string{
							// General errors
							"INVALID_REQUEST", "UNAUTHORIZED", "FORBIDDEN", "NOT_FOUND",
							"CONFLICT", "INTERNAL_ERROR", "SERVICE_UNAVAILABLE", "RATE_LIMIT_EXCEEDED",
							// Validation errors
							"VALIDATION_FAILED", "MISSING_PARAMETER", "INVALID_PARAMETER", "PARAMETER_OUT_OF_RANGE",
							// Storage/Array errors
							"ARRAY_NOT_STOPPED", "ARRAY_NOT_STARTED", "ARRAY_INVALID_STATE",
							"DISK_NOT_FOUND", "DISK_OFFLINE", "DISK_READ_ONLY",
							"PARITY_CHECK_ACTIVE", "PARITY_CHECK_FAILED", "INSUFFICIENT_SPACE",
							// Docker errors
							"CONTAINER_NOT_FOUND", "CONTAINER_NOT_RUNNING", "CONTAINER_NOT_STOPPED",
							"DOCKER_DAEMON_ERROR", "IMAGE_NOT_FOUND", "NETWORK_NOT_FOUND",
							// VM errors
							"VM_NOT_FOUND", "VM_NOT_RUNNING", "VM_NOT_STOPPED",
							"VM_CONFIG_ERROR", "VIRT_MANAGER_ERROR",
							// System errors
							"SYSTEM_NOT_READY", "COMMAND_FAILED", "PERMISSION_DENIED",
							"RESOURCE_BUSY", "HARDWARE_ERROR",
							// Async operation errors
							"OPERATION_NOT_FOUND", "OPERATION_NOT_CANCELLABLE", "OPERATION_CONFLICT",
							"OPERATION_TIMEOUT", "MAX_OPERATIONS_REACHED",
							// Authentication errors
							"INVALID_CREDENTIALS", "TOKEN_EXPIRED", "TOKEN_INVALID", "SESSION_EXPIRED",
							// Configuration errors
							"CONFIG_NOT_FOUND", "CONFIG_INVALID", "CONFIG_READ_ONLY",
						},
						"description": "Standardized error code",
						"example":     "OPERATION_NOT_FOUND",
					},
					"message": map[string]interface{}{
						"type":        "string",
						"description": "Human-readable error message",
						"example":     "Operation not found",
					},
					"details": map[string]interface{}{
						"type":                 "object",
						"nullable":             true,
						"description":          "Additional error context and debugging information",
						"additionalProperties": true,
						"properties": map[string]interface{}{
							"resource_id": map[string]interface{}{
								"type":        "string",
								"description": "ID of the resource that caused the error",
							},
							"resource_type": map[string]interface{}{
								"type":        "string",
								"description": "Type of resource (disk, container, vm, operation, etc.)",
							},
							"operation_type": map[string]interface{}{
								"type":        "string",
								"description": "Type of operation that failed",
							},
							"client_ip": map[string]interface{}{
								"type":        "string",
								"description": "Client IP address for rate limiting errors",
							},
							"validation_errors": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"$ref": "#/components/schemas/ValidationError",
								},
								"description": "Detailed validation errors for each field",
							},
							"conflicting_operation": map[string]interface{}{
								"type":        "string",
								"description": "ID or type of conflicting operation",
							},
							"limit": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"requests": map[string]interface{}{
										"type": "integer",
									},
									"window": map[string]interface{}{
										"type": "string",
									},
								},
								"description": "Rate limit that was exceeded",
							},
						},
					},
				},
				"required": []string{"code", "message"},
			},
			"meta": GetStandardResponseMeta(),
		},
		"required": []string{"error", "meta"},
	}
}

// GetValidationError returns the schema for validation errors
func GetValidationError() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"field": map[string]interface{}{
				"type":        "string",
				"description": "Name of the field that failed validation",
				"example":     "container_id",
			},
			"value": map[string]interface{}{
				"description": "The invalid value that was provided",
				"example":     "invalid-container-id",
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Human-readable validation error message",
				"example":     "Invalid container ID format (expected: 12-64 hex characters)",
			},
			"code": map[string]interface{}{
				"type":        "string",
				"description": "Validation error code for programmatic handling",
				"example":     "INVALID_FORMAT",
			},
		},
		"required": []string{"field", "message"},
	}
}

// GetValidationErrorResponse returns the schema for validation error responses
func GetValidationErrorResponse() map[string]interface{} {
	return map[string]interface{}{
		"allOf": []map[string]interface{}{
			GetAPIError(),
			{
				"type": "object",
				"properties": map[string]interface{}{
					"error": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"code": map[string]interface{}{
								"type": "string",
								"enum": []string{"VALIDATION_FAILED"},
							},
							"details": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"validation_errors": map[string]interface{}{
										"type": "array",
										"items": map[string]interface{}{
											"$ref": "#/components/schemas/ValidationError",
										},
										"minItems": 1,
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

// GetResourceNotFoundError returns the schema for resource not found errors
func GetResourceNotFoundError() map[string]interface{} {
	return map[string]interface{}{
		"allOf": []map[string]interface{}{
			GetAPIError(),
			{
				"type": "object",
				"properties": map[string]interface{}{
					"error": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"code": map[string]interface{}{
								"type": "string",
								"enum": []string{
									"DISK_NOT_FOUND", "CONTAINER_NOT_FOUND",
									"VM_NOT_FOUND", "OPERATION_NOT_FOUND",
								},
							},
							"details": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"resource_id": map[string]interface{}{
										"type":        "string",
										"description": "ID of the resource that was not found",
									},
									"resource_type": map[string]interface{}{
										"type":        "string",
										"description": "Type of resource (disk, container, vm, operation)",
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

// GetConflictError returns the schema for conflict errors
func GetConflictError() map[string]interface{} {
	return map[string]interface{}{
		"allOf": []map[string]interface{}{
			GetAPIError(),
			{
				"type": "object",
				"properties": map[string]interface{}{
					"error": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"code": map[string]interface{}{
								"type": "string",
								"enum": []string{
									"ARRAY_NOT_STOPPED", "ARRAY_NOT_STARTED",
									"OPERATION_CONFLICT", "PARITY_CHECK_ACTIVE",
								},
							},
							"details": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"conflicting_operation": map[string]interface{}{
										"type":        "string",
										"description": "ID or description of the conflicting operation",
									},
									"required_state": map[string]interface{}{
										"type":        "string",
										"description": "Required state for the operation to proceed",
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

// GetRateLimitError returns the schema for rate limit errors
func GetRateLimitError() map[string]interface{} {
	return map[string]interface{}{
		"allOf": []map[string]interface{}{
			GetAPIError(),
			{
				"type": "object",
				"properties": map[string]interface{}{
					"error": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"code": map[string]interface{}{
								"type": "string",
								"enum": []string{"RATE_LIMIT_EXCEEDED"},
							},
							"details": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"operation_type": map[string]interface{}{
										"type":        "string",
										"description": "The operation type that was rate limited",
									},
									"client_ip": map[string]interface{}{
										"type":        "string",
										"description": "The client IP that was rate limited",
									},
									"limit": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"requests": map[string]interface{}{
												"type":        "integer",
												"description": "Number of requests allowed",
											},
											"window": map[string]interface{}{
												"type":        "string",
												"description": "Time window for the rate limit",
											},
										},
									},
									"retry_after": map[string]interface{}{
										"type":        "integer",
										"description": "Seconds to wait before retrying",
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
