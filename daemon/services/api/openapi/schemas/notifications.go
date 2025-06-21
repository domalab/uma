package schemas

// GetNotificationSchemas returns notification-related schemas
func GetNotificationSchemas() map[string]interface{} {
	return map[string]interface{}{
		"NotificationList":  getNotificationListSchema(),
		"NotificationStats": getNotificationStatsSchema(),
		"NotificationInfo":  getNotificationInfoSchema(),
	}
}

func getNotificationListSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":        "array",
		"description": "List of notifications",
		"items": map[string]interface{}{
			"$ref": "#/components/schemas/NotificationInfo",
		},
		"example": []interface{}{
			map[string]interface{}{
				"id":       "notif-123",
				"title":    "System Alert",
				"message":  "High CPU usage detected",
				"severity": "warning",
				"read":     false,
			},
		},
	}
}

func getNotificationStatsSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"total": map[string]interface{}{
				"type":        "integer",
				"description": "Total number of notifications",
				"example":     25,
				"minimum":     0,
			},
			"unread": map[string]interface{}{
				"type":        "integer",
				"description": "Number of unread notifications",
				"example":     5,
				"minimum":     0,
			},
			"by_severity": map[string]interface{}{
				"type":        "object",
				"description": "Notification count by severity level",
				"properties": map[string]interface{}{
					"info": map[string]interface{}{
						"type":        "integer",
						"description": "Number of info notifications",
						"example":     10,
						"minimum":     0,
					},
					"warning": map[string]interface{}{
						"type":        "integer",
						"description": "Number of warning notifications",
						"example":     8,
						"minimum":     0,
					},
					"error": map[string]interface{}{
						"type":        "integer",
						"description": "Number of error notifications",
						"example":     5,
						"minimum":     0,
					},
					"critical": map[string]interface{}{
						"type":        "integer",
						"description": "Number of critical notifications",
						"example":     2,
						"minimum":     0,
					},
				},
			},
			"persistent": map[string]interface{}{
				"type":        "integer",
				"description": "Number of persistent notifications",
				"example":     0,
				"minimum":     0,
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "When notifications were last updated",
				"example":     "2024-01-01T12:00:00Z",
			},
		},
		"required": []string{"total", "unread", "by_severity", "persistent"},
	}
}

func getNotificationInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Unique notification identifier",
				"example":     "notif-123",
			},
			"title": map[string]interface{}{
				"type":        "string",
				"description": "Notification title",
				"example":     "System Alert",
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Notification message content",
				"example":     "High CPU usage detected on server",
			},
			"severity": map[string]interface{}{
				"type":        "string",
				"description": "Notification severity level",
				"enum":        []string{"info", "warning", "error", "critical"},
				"example":     "warning",
			},
			"category": map[string]interface{}{
				"type":        "string",
				"description": "Notification category",
				"enum":        []string{"system", "storage", "docker", "vm", "network", "security"},
				"example":     "system",
			},
			"source": map[string]interface{}{
				"type":        "string",
				"description": "Source component that generated the notification",
				"example":     "system_monitor",
			},
			"read": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the notification has been read",
				"example":     false,
			},
			"created_at": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "When the notification was created",
				"example":     "2024-01-01T12:00:00Z",
			},
			"read_at": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "When the notification was read (null if unread)",
				"example":     "2024-01-01T12:05:00Z",
				"nullable":    true,
			},
			"expires_at": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "When the notification expires (null if permanent)",
				"example":     "2024-01-02T12:00:00Z",
				"nullable":    true,
			},
			"actions": map[string]interface{}{
				"type":        "array",
				"description": "Available actions for this notification",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type":        "string",
							"description": "Action identifier",
							"example":     "acknowledge",
						},
						"label": map[string]interface{}{
							"type":        "string",
							"description": "Action display label",
							"example":     "Acknowledge",
						},
						"url": map[string]interface{}{
							"type":        "string",
							"description": "Action endpoint URL",
							"example":     "/api/v1/notifications/notif-123/acknowledge",
						},
					},
					"required": []string{"id", "label"},
				},
			},
			"metadata": map[string]interface{}{
				"type":                 "object",
				"description":          "Additional notification metadata",
				"additionalProperties": true,
				"example": map[string]interface{}{
					"cpu_usage":    "85%",
					"threshold":    "80%",
					"affected_vms": []string{"vm1", "vm2"},
				},
			},
		},
		"required": []string{"id", "title", "message", "severity", "category", "read", "created_at"},
	}
}
