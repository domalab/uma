package schemas

// GetStandardResponse returns the standard response schema
func GetStandardResponse() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"meta": GetStandardResponseMeta(),
		},
	}
}

// GetStandardResponseMeta returns the standard response metadata schema
func GetStandardResponseMeta() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"timestamp": map[string]interface{}{
				"type":        "integer",
				"format":      "int64",
				"description": "Unix timestamp of the response",
			},
			"request_id": map[string]interface{}{
				"type":        "string",
				"description": "Unique request identifier for tracing",
			},
			"api_version": map[string]interface{}{
				"type":        "string",
				"example":     "v1",
				"description": "API version",
			},
		},
		"required": []string{"timestamp", "api_version"},
	}
}

// GetAsyncOperationRequest returns the schema for async operation request
func GetAsyncOperationRequest() map[string]interface{} {
	return map[string]interface{}{
		"type":     "object",
		"required": []string{"type"},
		"properties": map[string]interface{}{
			"type": map[string]interface{}{
				"type": "string",
				"enum": []string{
					"parity_check", "parity_correct", "array_start", "array_stop",
					"disk_scan", "smart_scan", "system_reboot", "system_shutdown",
					"bulk_container", "bulk_vm",
				},
				"description": "Type of asynchronous operation",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"maxLength":   500,
				"description": "Human-readable description of the operation",
				"example":     "Comprehensive SMART data collection for all disks",
			},
			"cancellable": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Whether the operation can be cancelled",
			},
			"parameters": map[string]interface{}{
				"type":                 "object",
				"description":          "Operation-specific parameters",
				"additionalProperties": true,
				"examples": []map[string]interface{}{
					{
						"type":     "check",
						"priority": "normal",
					},
				},
			},
		},
	}
}

// GetAsyncOperationResponse returns the schema for async operation response
func GetAsyncOperationResponse() map[string]interface{} {
	return map[string]interface{}{
		"allOf": []map[string]interface{}{
			GetStandardResponse(),
			{
				"type": "object",
				"properties": map[string]interface{}{
					"data": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"id": map[string]interface{}{
								"type":        "string",
								"format":      "uuid",
								"description": "Unique identifier for the operation",
							},
							"type": map[string]interface{}{
								"type": "string",
								"enum": []string{
									"parity_check", "parity_correct", "array_start", "array_stop",
									"disk_scan", "smart_scan", "system_reboot", "system_shutdown",
									"bulk_container", "bulk_vm",
								},
								"description": "Type of asynchronous operation",
							},
							"status": map[string]interface{}{
								"type":        "string",
								"enum":        []string{"pending", "running", "completed", "failed", "cancelled"},
								"description": "Current status of the operation",
							},
							"description": map[string]interface{}{
								"type":        "string",
								"description": "Human-readable description",
							},
							"cancellable": map[string]interface{}{
								"type":        "boolean",
								"description": "Whether the operation can be cancelled",
							},
							"started": map[string]interface{}{
								"type":        "string",
								"format":      "date-time",
								"description": "When the operation was started",
							},
						},
					},
				},
			},
		},
	}
}

// GetAsyncOperationDetailResponse returns the schema for detailed async operation response
func GetAsyncOperationDetailResponse() map[string]interface{} {
	return map[string]interface{}{
		"allOf": []map[string]interface{}{
			GetStandardResponse(),
			{
				"type": "object",
				"properties": map[string]interface{}{
					"data": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"id": map[string]interface{}{
								"type":        "string",
								"format":      "uuid",
								"description": "Unique identifier for the operation",
							},
							"type": map[string]interface{}{
								"type": "string",
								"enum": []string{
									"parity_check", "parity_correct", "array_start", "array_stop",
									"disk_scan", "smart_scan", "system_reboot", "system_shutdown",
									"bulk_container", "bulk_vm",
								},
								"description": "Type of asynchronous operation",
							},
							"status": map[string]interface{}{
								"type":        "string",
								"enum":        []string{"pending", "running", "completed", "failed", "cancelled"},
								"description": "Current status of the operation",
							},
							"progress": map[string]interface{}{
								"type":        "integer",
								"minimum":     0,
								"maximum":     100,
								"description": "Progress percentage (0-100)",
							},
							"started": map[string]interface{}{
								"type":        "string",
								"format":      "date-time",
								"description": "When the operation was started",
							},
							"completed": map[string]interface{}{
								"type":        "string",
								"format":      "date-time",
								"nullable":    true,
								"description": "When the operation completed (if finished)",
							},
							"error": map[string]interface{}{
								"type":        "string",
								"nullable":    true,
								"description": "Error message if operation failed",
							},
							"result": map[string]interface{}{
								"type":                 "object",
								"nullable":             true,
								"description":          "Operation result data",
								"additionalProperties": true,
							},
							"cancellable": map[string]interface{}{
								"type":        "boolean",
								"description": "Whether the operation can be cancelled",
							},
							"description": map[string]interface{}{
								"type":        "string",
								"description": "Human-readable description",
							},
							"created_by": map[string]interface{}{
								"type":        "string",
								"description": "User or system that created the operation",
							},
						},
					},
				},
			},
		},
	}
}

// GetAsyncOperationListResponse returns the schema for async operation list response
func GetAsyncOperationListResponse() map[string]interface{} {
	return map[string]interface{}{
		"allOf": []map[string]interface{}{
			GetStandardResponse(),
			{
				"type": "object",
				"properties": map[string]interface{}{
					"data": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"operations": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"id": map[string]interface{}{
											"type":   "string",
											"format": "uuid",
										},
										"type": map[string]interface{}{
											"type": "string",
											"enum": []string{
												"parity_check", "parity_correct", "array_start", "array_stop",
												"disk_scan", "smart_scan", "system_reboot", "system_shutdown",
												"bulk_container", "bulk_vm",
											},
										},
										"status": map[string]interface{}{
											"type": "string",
											"enum": []string{"pending", "running", "completed", "failed", "cancelled"},
										},
										"progress": map[string]interface{}{
											"type":    "integer",
											"minimum": 0,
											"maximum": 100,
										},
										"started": map[string]interface{}{
											"type":   "string",
											"format": "date-time",
										},
										"description": map[string]interface{}{
											"type": "string",
										},
										"cancellable": map[string]interface{}{
											"type": "boolean",
										},
									},
								},
							},
							"total": map[string]interface{}{
								"type":        "integer",
								"description": "Total number of operations",
							},
							"active": map[string]interface{}{
								"type":        "integer",
								"description": "Number of active operations",
							},
							"completed": map[string]interface{}{
								"type":        "integer",
								"description": "Number of completed operations",
							},
							"failed": map[string]interface{}{
								"type":        "integer",
								"description": "Number of failed operations",
							},
						},
					},
				},
			},
		},
	}
}

// GetAsyncOperationCancelResponse returns the schema for operation cancellation response
func GetAsyncOperationCancelResponse() map[string]interface{} {
	return map[string]interface{}{
		"allOf": []map[string]interface{}{
			GetStandardResponse(),
			{
				"type": "object",
				"properties": map[string]interface{}{
					"data": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"success": map[string]interface{}{
								"type":    "boolean",
								"example": true,
							},
							"message": map[string]interface{}{
								"type":    "string",
								"example": "Operation cancelled successfully",
							},
						},
					},
				},
			},
		},
	}
}

// GetAsyncOperationStatsResponse returns the schema for operation statistics response
func GetAsyncOperationStatsResponse() map[string]interface{} {
	return map[string]interface{}{
		"allOf": []map[string]interface{}{
			GetStandardResponse(),
			{
				"type": "object",
				"properties": map[string]interface{}{
					"data": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"total_operations": map[string]interface{}{
								"type":        "integer",
								"description": "Total number of operations",
							},
							"max_operations": map[string]interface{}{
								"type":        "integer",
								"description": "Maximum concurrent operations allowed",
							},
							"by_status": map[string]interface{}{
								"type": "object",
								"additionalProperties": map[string]interface{}{
									"type": "integer",
								},
								"description": "Count of operations by status",
							},
							"by_type": map[string]interface{}{
								"type": "object",
								"additionalProperties": map[string]interface{}{
									"type": "integer",
								},
								"description": "Count of operations by type",
							},
						},
					},
				},
			},
		},
	}
}

// GetOperationSchemas returns operation-related schemas
func GetOperationSchemas() map[string]interface{} {
	return map[string]interface{}{
		"OperationList":  getOperationListSchema(),
		"OperationStats": getOperationStatsSchema(),
		"OperationInfo":  getOperationInfoSchema(),
	}
}

func getOperationListSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":        "array",
		"description": "List of async operations",
		"items": map[string]interface{}{
			"$ref": "#/components/schemas/OperationInfo",
		},
		"example": []interface{}{
			map[string]interface{}{
				"id":          "op-123",
				"type":        "parity_check",
				"status":      "running",
				"progress":    45,
				"description": "Parity check in progress",
			},
		},
	}
}

func getOperationStatsSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"total": map[string]interface{}{
				"type":        "integer",
				"description": "Total number of operations",
				"example":     15,
				"minimum":     0,
			},
			"active": map[string]interface{}{
				"type":        "integer",
				"description": "Number of active operations",
				"example":     3,
				"minimum":     0,
			},
			"completed": map[string]interface{}{
				"type":        "integer",
				"description": "Number of completed operations",
				"example":     10,
				"minimum":     0,
			},
			"failed": map[string]interface{}{
				"type":        "integer",
				"description": "Number of failed operations",
				"example":     2,
				"minimum":     0,
			},
			"by_type": map[string]interface{}{
				"type":        "object",
				"description": "Operation count by type",
				"additionalProperties": map[string]interface{}{
					"type": "integer",
				},
				"example": map[string]interface{}{
					"parity_check": 5,
					"array_start":  3,
					"disk_scan":    7,
				},
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "When statistics were last updated",
				"example":     "2024-01-01T12:00:00Z",
			},
		},
		"required": []string{"total", "active", "completed", "failed"},
	}
}

func getOperationInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Unique operation identifier",
				"example":     "op-123",
			},
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Type of operation",
				"enum":        []string{"parity_check", "parity_correct", "array_start", "array_stop", "disk_scan", "smart_scan", "system_reboot", "system_shutdown", "bulk_container", "bulk_vm"},
				"example":     "parity_check",
			},
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Current operation status",
				"enum":        []string{"pending", "running", "completed", "failed", "cancelled"},
				"example":     "running",
			},
			"progress": map[string]interface{}{
				"type":        "integer",
				"description": "Progress percentage (0-100)",
				"minimum":     0,
				"maximum":     100,
				"example":     45,
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Human-readable operation description",
				"example":     "Parity check in progress",
			},
			"cancellable": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the operation can be cancelled",
				"example":     true,
			},
			"started": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "When the operation was started",
				"example":     "2024-01-01T12:00:00Z",
			},
			"completed": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "When the operation completed (null if not finished)",
				"example":     "2024-01-01T13:00:00Z",
				"nullable":    true,
			},
			"error": map[string]interface{}{
				"type":        "string",
				"description": "Error message if operation failed",
				"example":     "Disk read error",
				"nullable":    true,
			},
			"result": map[string]interface{}{
				"type":                 "object",
				"description":          "Operation result data",
				"additionalProperties": true,
				"example": map[string]interface{}{
					"errors_found":    0,
					"sectors_checked": 1000000,
					"duration":        3600,
				},
				"nullable": true,
			},
			"created_by": map[string]interface{}{
				"type":        "string",
				"description": "User or system that created the operation",
				"example":     "admin",
			},
		},
		"required": []string{"id", "type", "status", "description", "started"},
	}
}
