package schemas

// GetResponseSchemas returns response-related schemas
func GetResponseSchemas() map[string]interface{} {
	return map[string]interface{}{
		"NotificationResponse":    getNotificationResponseSchema(),
		"ParityCheckResponse":     getParityCheckResponseSchema(),
		"SystemOperationResponse": getSystemOperationResponseSchema(),
		"ArrayOperationResponse":  getArrayOperationResponseSchema(),
		"DockerOperationResponse": getDockerOperationResponseSchema(),
	}
}

func getNotificationResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the notification operation was successful",
				"example":     true,
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Operation result message",
				"example":     "Notification created successfully",
			},
			"notification": map[string]interface{}{
				"$ref": "#/components/schemas/NotificationInfo",
			},
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "Notification operation performed",
				"enum":        []string{"create", "update", "delete", "mark_read", "clear_all"},
				"example":     "create",
			},
			"count": map[string]interface{}{
				"type":        "integer",
				"description": "Number of notifications affected (for bulk operations)",
				"example":     1,
				"minimum":     0,
			},
		},
		"required": []string{"success", "message", "operation"},
	}
}

func getParityCheckResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the parity check operation was successful",
				"example":     true,
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Operation result message",
				"example":     "Parity check started successfully",
			},
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "Parity operation performed",
				"enum":        []string{"start", "stop", "pause", "resume"},
				"example":     "start",
			},
			"check_type": map[string]interface{}{
				"type":        "string",
				"description": "Type of parity check",
				"enum":        []string{"check", "correct"},
				"example":     "check",
			},
			"estimated_duration": map[string]interface{}{
				"type":        "integer",
				"description": "Estimated duration in seconds",
				"example":     28800,
				"minimum":     0,
			},
			"operation_id": map[string]interface{}{
				"type":        "string",
				"description": "Async operation ID for tracking",
				"example":     "op-parity-123",
			},
		},
		"required": []string{"success", "message", "operation"},
	}
}

func getSystemOperationResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the system operation was successful",
				"example":     true,
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Operation result message",
				"example":     "System reboot initiated successfully",
			},
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "System operation performed",
				"enum":        []string{"reboot", "shutdown", "restart_service", "stop_service"},
				"example":     "reboot",
			},
			"scheduled_time": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "When the operation is scheduled to execute",
				"example":     "2024-01-01T12:05:00Z",
			},
			"delay_seconds": map[string]interface{}{
				"type":        "integer",
				"description": "Delay before operation executes",
				"example":     60,
				"minimum":     0,
			},
			"operation_id": map[string]interface{}{
				"type":        "string",
				"description": "Async operation ID for tracking",
				"example":     "op-reboot-456",
			},
		},
		"required": []string{"success", "message", "operation"},
	}
}

func getArrayOperationResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the array operation was successful",
				"example":     true,
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Operation result message",
				"example":     "Array started successfully",
			},
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "Array operation performed",
				"enum":        []string{"start", "stop"},
				"example":     "start",
			},
			"array_status": map[string]interface{}{
				"type":        "string",
				"description": "Current array status after operation",
				"enum":        []string{"started", "stopped", "starting", "stopping"},
				"example":     "starting",
			},
			"warnings": map[string]interface{}{
				"type":        "array",
				"description": "Any warnings from the operation",
				"items": map[string]interface{}{
					"type": "string",
				},
				"example": []string{"Disk temperature high"},
			},
			"operation_id": map[string]interface{}{
				"type":        "string",
				"description": "Async operation ID for tracking",
				"example":     "op-array-789",
			},
		},
		"required": []string{"success", "message", "operation", "array_status"},
	}
}

func getDockerOperationResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the Docker operation was successful",
				"example":     true,
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Operation result message",
				"example":     "Container started successfully",
			},
			"container_id": map[string]interface{}{
				"type":        "string",
				"description": "Container ID or name",
				"example":     "plex",
			},
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "Docker operation performed",
				"enum":        []string{"start", "stop", "restart", "pause", "resume"},
				"example":     "start",
			},
			"container_status": map[string]interface{}{
				"type":        "string",
				"description": "Current container status after operation",
				"enum":        []string{"created", "running", "paused", "restarting", "removing", "exited", "dead"},
				"example":     "running",
			},
			"duration": map[string]interface{}{
				"type":        "number",
				"description": "Operation duration in seconds",
				"example":     2.5,
				"minimum":     0,
			},
		},
		"required": []string{"success", "message", "container_id", "operation"},
	}
}
