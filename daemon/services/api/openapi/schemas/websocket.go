package schemas

// GetWebSocketSchemas returns WebSocket-related schemas
func GetWebSocketSchemas() map[string]interface{} {
	return map[string]interface{}{
		"WebSocketMessage":       getWebSocketMessageSchema(),
		"WebSocketEvent":         getWebSocketEventSchema(),
		"WebSocketSubscription":  getWebSocketSubscriptionSchema(),
		"WebSocketError":         getWebSocketErrorSchema(),
		"UnifiedWebSocketStream": getUnifiedWebSocketStreamSchema(),
	}
}

func getWebSocketMessageSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Message type",
				"enum":        []string{"event", "data", "error", "ping", "pong", "subscribe", "unsubscribe"},
				"example":     "event",
			},
			"event": map[string]interface{}{
				"type":        "string",
				"description": "Event name",
				"example":     "system.stats",
			},
			"data": map[string]interface{}{
				"description": "Message data payload",
				"example": map[string]interface{}{
					"cpu_usage":    25.5,
					"memory_usage": 45.2,
					"timestamp":    "2025-06-16T14:30:00Z",
				},
			},
			"timestamp": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Message timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Message ID for tracking",
				"example":     "msg_1234567890",
			},
			"channel": map[string]interface{}{
				"type":        "string",
				"description": "WebSocket channel",
				"example":     "system.stats",
			},
		},
		"required": []string{"type", "timestamp"},
	}
}

func getWebSocketEventSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"event": map[string]interface{}{
				"type":        "string",
				"description": "Event name",
				"enum": []string{
					"system.stats", "docker.container.start", "docker.container.stop",
					"storage.array.status", "vm.state.change", "ups.status.change",
					"temperature.alert", "disk.smart.warning",
				},
				"example": "docker.container.start",
			},
			"source": map[string]interface{}{
				"type":        "string",
				"description": "Event source",
				"enum":        []string{"system", "docker", "storage", "vm", "ups", "monitoring"},
				"example":     "docker",
			},
			"severity": map[string]interface{}{
				"type":        "string",
				"description": "Event severity",
				"enum":        []string{"info", "warning", "error", "critical"},
				"example":     "info",
			},
			"data": map[string]interface{}{
				"description": "Event-specific data",
				"example": map[string]interface{}{
					"container_id":   "plex",
					"container_name": "plex",
					"status":         "running",
				},
			},
			"timestamp": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Event timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
			"correlation_id": map[string]interface{}{
				"type":        "string",
				"description": "Correlation ID for tracking related events",
				"example":     "corr_1234567890",
			},
		},
		"required": []string{"event", "source", "severity", "timestamp"},
	}
}

func getWebSocketSubscriptionSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Subscription action",
				"enum":        []string{"subscribe", "unsubscribe"},
				"example":     "subscribe",
			},
			"channels": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
					"enum": []string{
						"system.stats", "docker.events", "storage.status",
						"vm.events", "ups.status", "temperature.alerts",
						"disk.smart", "network.stats",
					},
				},
				"description": "Channels to subscribe/unsubscribe",
				"example":     []string{"system.stats", "docker.events"},
				"minItems":    1,
				"maxItems":    10,
				"uniqueItems": true,
			},
			"filters": map[string]interface{}{
				"type":        "object",
				"description": "Optional filters for events",
				"properties": map[string]interface{}{
					"severity": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
							"enum": []string{"info", "warning", "error", "critical"},
						},
						"description": "Filter by event severity",
					},
					"source": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
							"enum": []string{"system", "docker", "storage", "vm", "ups"},
						},
						"description": "Filter by event source",
					},
					"container_ids": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
						"description": "Filter Docker events by container IDs",
					},
				},
			},
			"rate_limit": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum events per second (0 = no limit)",
				"example":     10,
				"minimum":     0,
				"maximum":     100,
				"default":     0,
			},
		},
		"required": []string{"action", "channels"},
	}
}

func getWebSocketErrorSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"error": map[string]interface{}{
				"type":        "string",
				"description": "Error message",
				"example":     "Invalid subscription channel",
			},
			"code": map[string]interface{}{
				"type":        "string",
				"description": "Error code",
				"enum": []string{
					"INVALID_MESSAGE", "INVALID_CHANNEL", "SUBSCRIPTION_FAILED",
					"RATE_LIMITED", "AUTHENTICATION_REQUIRED", "PERMISSION_DENIED",
				},
				"example": "INVALID_CHANNEL",
			},
			"details": map[string]interface{}{
				"type":                 "object",
				"description":          "Additional error details",
				"additionalProperties": true,
				"example": map[string]interface{}{
					"channel": "invalid.channel",
					"reason":  "Channel does not exist",
				},
			},
			"timestamp": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Error timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"error", "code", "timestamp"},
	}
}

func getUnifiedWebSocketStreamSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":        "object",
		"description": "Unified WebSocket stream supporting multiple event types with subscription management",
		"properties": map[string]interface{}{
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Event type",
				"enum":        []string{"system.stats", "docker.events", "storage.status", "temperature.alert", "resource.alert", "infrastructure.status"},
				"example":     "system.stats",
			},
			"channel": map[string]interface{}{
				"type":        "string",
				"description": "Event channel for subscription management",
				"example":     "system.stats",
			},
			"data": map[string]interface{}{
				"type":                 "object",
				"description":          "Event data (varies by type)",
				"additionalProperties": true,
				"example": map[string]interface{}{
					"cpu_percent":    25.5,
					"memory_percent": 50.0,
					"timestamp":      "2025-06-19T14:30:00Z",
				},
			},
			"timestamp": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Event timestamp",
				"example":     "2025-06-19T14:30:00Z",
			},
			"subscription_id": map[string]interface{}{
				"type":        "string",
				"description": "Subscription ID for tracking",
				"example":     "sub_1234567890",
			},
		},
		"required": []string{"type", "channel", "data", "timestamp"},
	}
}
