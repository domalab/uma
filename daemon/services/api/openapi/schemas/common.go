package schemas

// GetCommonSchemas returns common schemas used across all API endpoints
func GetCommonSchemas() map[string]interface{} {
	return map[string]interface{}{
		"StandardResponse": getStandardResponseSchema(),
		"PaginationInfo":   getPaginationSchema(),
		"ResponseMeta":     getResponseMetaSchema(),
		"HealthResponse":   getHealthResponseSchema(),
		"Error":            getErrorSchema(),
		"SuccessResponse":  getSuccessResponseSchema(),
	}
}

func getStandardResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"data": map[string]interface{}{
				"description": "The response data",
			},
			"pagination": map[string]interface{}{
				"$ref": "#/components/schemas/PaginationInfo",
			},
			"meta": map[string]interface{}{
				"$ref": "#/components/schemas/ResponseMeta",
			},
		},
		"required": []string{"data"},
	}
}

func getPaginationSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"page": map[string]interface{}{
				"type":        "integer",
				"description": "Current page number",
				"example":     1,
				"minimum":     1,
			},
			"per_page": map[string]interface{}{
				"type":        "integer",
				"description": "Number of items per page",
				"example":     50,
				"minimum":     1,
				"maximum":     1000,
			},
			"total": map[string]interface{}{
				"type":        "integer",
				"description": "Total number of items",
				"example":     150,
				"minimum":     0,
			},
			"has_more": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether there are more pages available",
				"example":     true,
			},
			"total_pages": map[string]interface{}{
				"type":        "integer",
				"description": "Total number of pages",
				"example":     3,
				"minimum":     0,
			},
		},
		"required": []string{"page", "per_page", "total", "has_more", "total_pages"},
	}
}

func getResponseMetaSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"request_id": map[string]interface{}{
				"type":        "string",
				"description": "Unique request identifier for tracing",
				"example":     "req_1234567890_5678",
			},
			"timestamp": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Response timestamp in ISO 8601 format",
				"example":     "2025-06-16T14:30:00Z",
			},
			"version": map[string]interface{}{
				"type":        "string",
				"description": "API version",
				"example":     "v1",
			},
			"server": map[string]interface{}{
				"type":        "string",
				"description": "Server identifier",
				"example":     "uma-server-01",
			},
		},
	}
}

func getHealthResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Overall health status",
				"enum":        []string{"healthy", "degraded", "unhealthy"},
				"example":     "healthy",
			},
			"version": map[string]interface{}{
				"type":        "string",
				"description": "UMA version",
				"example":     "2025.06.16",
			},
			"uptime": map[string]interface{}{
				"type":        "integer",
				"description": "Server uptime in seconds",
				"example":     86400,
				"minimum":     0,
			},
			"timestamp": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Health check timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
			"checks": map[string]interface{}{
				"type":        "object",
				"description": "Status of service health checks",
				"properties": map[string]interface{}{
					"auth": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"healthy", "unhealthy"},
						"example": "healthy",
					},
					"docker": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"healthy", "unhealthy"},
						"example": "healthy",
					},
					"storage": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"healthy", "unhealthy"},
						"example": "healthy",
					},
					"system": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"healthy", "unhealthy"},
						"example": "healthy",
					},
				},
			},
		},
		"required": []string{"status", "version", "uptime", "timestamp"},
	}
}

func getErrorSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"error": map[string]interface{}{
				"type":        "string",
				"description": "Human-readable error message",
				"example":     "Invalid request parameters",
			},
			"code": map[string]interface{}{
				"type":        "string",
				"description": "Machine-readable error code for programmatic handling",
				"example":     "INVALID_REQUEST",
			},
			"details": map[string]interface{}{
				"type":                 "object",
				"description":          "Additional error details and context",
				"additionalProperties": true,
				"example": map[string]interface{}{
					"field":   "container_ids",
					"message": "must contain at least 1 item",
				},
			},
			"request_id": map[string]interface{}{
				"type":        "string",
				"description": "Request ID for error tracking",
				"example":     "req_1234567890_5678",
			},
		},
		"required": []string{"error"},
	}
}

func getSuccessResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type":        "boolean",
				"description": "Operation success status",
				"example":     true,
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Success message",
				"example":     "Operation completed successfully",
			},
			"data": map[string]interface{}{
				"description": "Optional response data",
			},
		},
		"required": []string{"success"},
	}
}
