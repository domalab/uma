package responses

// GetCommonResponses returns all common reusable responses
func GetCommonResponses() map[string]interface{} {
	return map[string]interface{}{
		"BadRequest":          getBadRequestResponse(),
		"Unauthorized":        getUnauthorizedResponse(),
		"Forbidden":           getForbiddenResponse(),
		"NotFound":            getNotFoundResponse(),
		"Conflict":            getConflictResponse(),
		"UnprocessableEntity": getUnprocessableEntityResponse(),
		"TooManyRequests":     getTooManyRequestsResponse(),
		"InternalServerError": getInternalServerErrorResponse(),
		"ServiceUnavailable":  getServiceUnavailableResponse(),
		"Success":             getSuccessResponse(),
		"Created":             getCreatedResponse(),
		"Accepted":            getAcceptedResponse(),
		"NoContent":           getNoContentResponse(),
	}
}

func getBadRequestResponse() map[string]interface{} {
	return map[string]interface{}{
		"description": "Bad Request - Invalid request parameters or malformed request",
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": map[string]interface{}{
					"$ref": "#/components/schemas/Error",
				},
				"examples": map[string]interface{}{
					"validation_error": map[string]interface{}{
						"summary": "Validation Error",
						"value": map[string]interface{}{
							"error": "Invalid request parameters",
							"code":  "VALIDATION_ERROR",
							"details": map[string]interface{}{
								"field":   "container_ids",
								"message": "must contain at least 1 item",
							},
							"request_id": "req_1234567890_5678",
						},
					},
					"malformed_json": map[string]interface{}{
						"summary": "Malformed JSON",
						"value": map[string]interface{}{
							"error": "Invalid JSON in request body",
							"code":  "MALFORMED_JSON",
							"details": map[string]interface{}{
								"line":   5,
								"column": 12,
							},
							"request_id": "req_1234567890_5678",
						},
					},
					"missing_parameter": map[string]interface{}{
						"summary": "Missing Required Parameter",
						"value": map[string]interface{}{
							"error": "Missing required parameter",
							"code":  "MISSING_PARAMETER",
							"details": map[string]interface{}{
								"parameter": "operation",
								"location":  "request body",
							},
							"request_id": "req_1234567890_5678",
						},
					},
				},
			},
		},
	}
}

func getUnauthorizedResponse() map[string]interface{} {
	return map[string]interface{}{
		"description": "Unauthorized - Authentication required or invalid credentials",
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": map[string]interface{}{
					"$ref": "#/components/schemas/AuthError",
				},
				"examples": map[string]interface{}{
					"missing_token": map[string]interface{}{
						"summary": "Missing Authentication Token",
						"value": map[string]interface{}{
							"error":      "Authentication required",
							"error_code": "TOKEN_MISSING",
							"details": map[string]interface{}{
								"message": "Authorization header is required",
							},
							"timestamp": "2025-06-16T14:30:00Z",
						},
					},
					"invalid_token": map[string]interface{}{
						"summary": "Invalid Token",
						"value": map[string]interface{}{
							"error":      "Invalid authentication token",
							"error_code": "TOKEN_INVALID",
							"details": map[string]interface{}{
								"message": "Token signature verification failed",
							},
							"timestamp": "2025-06-16T14:30:00Z",
						},
					},
					"expired_token": map[string]interface{}{
						"summary": "Expired Token",
						"value": map[string]interface{}{
							"error":      "Authentication token has expired",
							"error_code": "TOKEN_EXPIRED",
							"details": map[string]interface{}{
								"expired_at": "2025-06-16T13:30:00Z",
							},
							"timestamp": "2025-06-16T14:30:00Z",
						},
					},
				},
			},
		},
	}
}

func getForbiddenResponse() map[string]interface{} {
	return map[string]interface{}{
		"description": "Forbidden - Insufficient permissions for the requested operation",
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": map[string]interface{}{
					"$ref": "#/components/schemas/AuthError",
				},
				"examples": map[string]interface{}{
					"insufficient_permissions": map[string]interface{}{
						"summary": "Insufficient Permissions",
						"value": map[string]interface{}{
							"error":      "Insufficient permissions",
							"error_code": "INSUFFICIENT_PERMISSIONS",
							"details": map[string]interface{}{
								"required_permission": "docker.containers.manage",
								"user_permissions":    []string{"docker.containers.read"},
							},
							"timestamp": "2025-06-16T14:30:00Z",
						},
					},
					"readonly_user": map[string]interface{}{
						"summary": "Read-only User",
						"value": map[string]interface{}{
							"error":      "Read-only user cannot perform write operations",
							"error_code": "READONLY_USER",
							"details": map[string]interface{}{
								"operation": "container_start",
								"user_role": "readonly",
							},
							"timestamp": "2025-06-16T14:30:00Z",
						},
					},
				},
			},
		},
	}
}

func getNotFoundResponse() map[string]interface{} {
	return map[string]interface{}{
		"description": "Not Found - The requested resource does not exist",
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": map[string]interface{}{
					"$ref": "#/components/schemas/Error",
				},
				"examples": map[string]interface{}{
					"container_not_found": map[string]interface{}{
						"summary": "Container Not Found",
						"value": map[string]interface{}{
							"error": "Container 'nonexistent' not found",
							"code":  "CONTAINER_NOT_FOUND",
							"details": map[string]interface{}{
								"container_id": "nonexistent",
								"suggestion":   "Check container name or ID",
							},
							"request_id": "req_1234567890_5678",
						},
					},
					"vm_not_found": map[string]interface{}{
						"summary": "VM Not Found",
						"value": map[string]interface{}{
							"error": "Virtual machine 'missing-vm' not found",
							"code":  "VM_NOT_FOUND",
							"details": map[string]interface{}{
								"vm_id":      "missing-vm",
								"suggestion": "Check VM name or ID",
							},
							"request_id": "req_1234567890_5678",
						},
					},
					"endpoint_not_found": map[string]interface{}{
						"summary": "Endpoint Not Found",
						"value": map[string]interface{}{
							"error": "The requested endpoint does not exist",
							"code":  "ENDPOINT_NOT_FOUND",
							"details": map[string]interface{}{
								"path":       "/api/v1/nonexistent",
								"suggestion": "Check API documentation for valid endpoints",
							},
							"request_id": "req_1234567890_5678",
						},
					},
				},
			},
		},
	}
}

func getConflictResponse() map[string]interface{} {
	return map[string]interface{}{
		"description": "Conflict - The request conflicts with the current state of the resource",
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": map[string]interface{}{
					"$ref": "#/components/schemas/Error",
				},
				"examples": map[string]interface{}{
					"container_already_running": map[string]interface{}{
						"summary": "Container Already Running",
						"value": map[string]interface{}{
							"error": "Container 'plex' is already running",
							"code":  "CONTAINER_ALREADY_RUNNING",
							"details": map[string]interface{}{
								"container_id":     "plex",
								"current_status":   "running",
								"requested_action": "start",
							},
							"request_id": "req_1234567890_5678",
						},
					},
					"array_already_started": map[string]interface{}{
						"summary": "Array Already Started",
						"value": map[string]interface{}{
							"error": "Unraid array is already started",
							"code":  "ARRAY_ALREADY_STARTED",
							"details": map[string]interface{}{
								"current_status":   "started",
								"requested_action": "start",
							},
							"request_id": "req_1234567890_5678",
						},
					},
				},
			},
		},
	}
}

func getUnprocessableEntityResponse() map[string]interface{} {
	return map[string]interface{}{
		"description": "Unprocessable Entity - The request is well-formed but contains semantic errors",
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": map[string]interface{}{
					"$ref": "#/components/schemas/Error",
				},
				"examples": map[string]interface{}{
					"invalid_operation": map[string]interface{}{
						"summary": "Invalid Operation",
						"value": map[string]interface{}{
							"error": "Cannot perform operation on container in current state",
							"code":  "INVALID_OPERATION",
							"details": map[string]interface{}{
								"container_id":     "plex",
								"current_status":   "exited",
								"requested_action": "pause",
								"valid_actions":    []string{"start", "remove"},
							},
							"request_id": "req_1234567890_5678",
						},
					},
				},
			},
		},
	}
}

func getTooManyRequestsResponse() map[string]interface{} {
	return map[string]interface{}{
		"description": "Too Many Requests - Rate limit exceeded",
		"headers": map[string]interface{}{
			"X-RateLimit-Limit": map[string]interface{}{
				"description": "Request limit per time window",
				"schema": map[string]interface{}{
					"type": "integer",
				},
			},
			"X-RateLimit-Remaining": map[string]interface{}{
				"description": "Remaining requests in current window",
				"schema": map[string]interface{}{
					"type": "integer",
				},
			},
			"X-RateLimit-Reset": map[string]interface{}{
				"description": "Time when rate limit resets (Unix timestamp)",
				"schema": map[string]interface{}{
					"type": "integer",
				},
			},
		},
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": map[string]interface{}{
					"$ref": "#/components/schemas/Error",
				},
				"examples": map[string]interface{}{
					"rate_limited": map[string]interface{}{
						"summary": "Rate Limit Exceeded",
						"value": map[string]interface{}{
							"error": "Rate limit exceeded",
							"code":  "RATE_LIMITED",
							"details": map[string]interface{}{
								"limit":       100,
								"window":      "1 hour",
								"reset_at":    "2025-06-16T15:30:00Z",
								"retry_after": 3600,
							},
							"request_id": "req_1234567890_5678",
						},
					},
				},
			},
		},
	}
}

func getInternalServerErrorResponse() map[string]interface{} {
	return map[string]interface{}{
		"description": "Internal Server Error - An unexpected error occurred",
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": map[string]interface{}{
					"$ref": "#/components/schemas/Error",
				},
				"examples": map[string]interface{}{
					"service_unavailable": map[string]interface{}{
						"summary": "Service Unavailable",
						"value": map[string]interface{}{
							"error": "Docker daemon is not available",
							"code":  "SERVICE_UNAVAILABLE",
							"details": map[string]interface{}{
								"service": "docker",
								"status":  "unreachable",
							},
							"request_id": "req_1234567890_5678",
						},
					},
					"unexpected_error": map[string]interface{}{
						"summary": "Unexpected Error",
						"value": map[string]interface{}{
							"error": "An unexpected error occurred",
							"code":  "INTERNAL_ERROR",
							"details": map[string]interface{}{
								"error_id": "err_1234567890",
							},
							"request_id": "req_1234567890_5678",
						},
					},
				},
			},
		},
	}
}

func getServiceUnavailableResponse() map[string]interface{} {
	return map[string]interface{}{
		"description": "Service Unavailable - The service is temporarily unavailable",
		"headers": map[string]interface{}{
			"Retry-After": map[string]interface{}{
				"description": "Seconds to wait before retrying",
				"schema": map[string]interface{}{
					"type": "integer",
				},
			},
		},
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": map[string]interface{}{
					"$ref": "#/components/schemas/Error",
				},
				"examples": map[string]interface{}{
					"maintenance_mode": map[string]interface{}{
						"summary": "Maintenance Mode",
						"value": map[string]interface{}{
							"error": "Service is in maintenance mode",
							"code":  "MAINTENANCE_MODE",
							"details": map[string]interface{}{
								"estimated_duration": "30 minutes",
								"retry_after":        1800,
							},
							"request_id": "req_1234567890_5678",
						},
					},
				},
			},
		},
	}
}

func getSuccessResponse() map[string]interface{} {
	return map[string]interface{}{
		"description": "Success - Operation completed successfully",
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": map[string]interface{}{
					"$ref": "#/components/schemas/SuccessResponse",
				},
				"examples": map[string]interface{}{
					"operation_success": map[string]interface{}{
						"summary": "Operation Success",
						"value": map[string]interface{}{
							"success": true,
							"message": "Operation completed successfully",
							"data": map[string]interface{}{
								"operation_id": "op_1234567890",
								"timestamp":    "2025-06-16T14:30:00Z",
							},
						},
					},
				},
			},
		},
	}
}

func getCreatedResponse() map[string]interface{} {
	return map[string]interface{}{
		"description": "Created - Resource created successfully",
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": map[string]interface{}{
					"$ref": "#/components/schemas/SuccessResponse",
				},
				"examples": map[string]interface{}{
					"resource_created": map[string]interface{}{
						"summary": "Resource Created",
						"value": map[string]interface{}{
							"success": true,
							"message": "Resource created successfully",
							"data": map[string]interface{}{
								"id":         "resource_1234567890",
								"created_at": "2025-06-16T14:30:00Z",
							},
						},
					},
				},
			},
		},
	}
}

func getAcceptedResponse() map[string]interface{} {
	return map[string]interface{}{
		"description": "Accepted - Request accepted for processing",
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": map[string]interface{}{
					"$ref": "#/components/schemas/SuccessResponse",
				},
				"examples": map[string]interface{}{
					"async_operation": map[string]interface{}{
						"summary": "Asynchronous Operation",
						"value": map[string]interface{}{
							"success": true,
							"message": "Request accepted for processing",
							"data": map[string]interface{}{
								"operation_id":         "op_1234567890",
								"status":               "pending",
								"estimated_completion": "2025-06-16T14:35:00Z",
							},
						},
					},
				},
			},
		},
	}
}

func getNoContentResponse() map[string]interface{} {
	return map[string]interface{}{
		"description": "No Content - Operation completed successfully with no response body",
	}
}

// Convenience functions for new error types

// BadRequest returns a 400 Bad Request response
func BadRequest() map[string]interface{} {
	return getBadRequestResponse()
}

// Unauthorized returns a 401 Unauthorized response
func Unauthorized() map[string]interface{} {
	return getUnauthorizedResponse()
}

// Forbidden returns a 403 Forbidden response
func Forbidden() map[string]interface{} {
	return getForbiddenResponse()
}

// NotFound returns a 404 Not Found response
func NotFound() map[string]interface{} {
	return getNotFoundResponse()
}

// Conflict returns a 409 Conflict response
func Conflict() map[string]interface{} {
	return getConflictResponse()
}

// RateLimitExceeded returns a 429 Too Many Requests response
func RateLimitExceeded() map[string]interface{} {
	return getTooManyRequestsResponse()
}

// InternalServerError returns a 500 Internal Server Error response
func InternalServerError() map[string]interface{} {
	return getInternalServerErrorResponse()
}

// NotImplemented returns a 501 Not Implemented response
func NotImplemented() map[string]interface{} {
	return map[string]interface{}{
		"description": "Not Implemented - The requested functionality is not yet implemented",
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": map[string]interface{}{
					"$ref": "#/components/schemas/APIError",
				},
				"examples": map[string]interface{}{
					"not_implemented": map[string]interface{}{
						"summary": "Feature Not Implemented",
						"value": map[string]interface{}{
							"error": map[string]interface{}{
								"code":    "NOT_IMPLEMENTED",
								"message": "This feature is not yet implemented",
								"details": map[string]interface{}{
									"feature": "rate_limit_config_update",
									"note":    "This endpoint would require admin authentication",
								},
							},
							"meta": map[string]interface{}{
								"timestamp":   1640995200,
								"api_version": "v1",
							},
						},
					},
				},
			},
		},
	}
}
