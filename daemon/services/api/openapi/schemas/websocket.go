package schemas

// GetWebSocketSchemas returns WebSocket-related schemas
func GetWebSocketSchemas() map[string]interface{} {
	return map[string]interface{}{
		"WebSocketMessage":      getWebSocketMessageSchema(),
		"WebSocketEvent":        getWebSocketEventSchema(),
		"WebSocketSubscription": getWebSocketSubscriptionSchema(),
		"WebSocketError":        getWebSocketErrorSchema(),
		"WebSocketStats":        getWebSocketStatsSchema(),
		"WebSocketConnection":   getWebSocketConnectionSchema(),
		"DockerEventsStream":    getDockerEventsStreamSchema(),
		"SystemStatsStream":     getSystemStatsStreamSchema(),
		"StorageStatusStream":   getStorageStatusStreamSchema(),
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

func getWebSocketStatsSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"connections": map[string]interface{}{
				"type":        "integer",
				"description": "Number of active WebSocket connections",
				"example":     5,
				"minimum":     0,
			},
			"subscriptions": map[string]interface{}{
				"type":        "integer",
				"description": "Total number of active subscriptions",
				"example":     15,
				"minimum":     0,
			},
			"messages_sent": map[string]interface{}{
				"type":        "integer",
				"description": "Total messages sent since startup",
				"example":     1000,
				"minimum":     0,
			},
			"messages_received": map[string]interface{}{
				"type":        "integer",
				"description": "Total messages received since startup",
				"example":     250,
				"minimum":     0,
			},
			"events_published": map[string]interface{}{
				"type":        "integer",
				"description": "Total events published since startup",
				"example":     750,
				"minimum":     0,
			},
			"channels": map[string]interface{}{
				"type":        "object",
				"description": "Per-channel statistics",
				"additionalProperties": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"subscribers": map[string]interface{}{
							"type":    "integer",
							"minimum": 0,
						},
						"messages": map[string]interface{}{
							"type":    "integer",
							"minimum": 0,
						},
					},
				},
				"example": map[string]interface{}{
					"system.stats": map[string]interface{}{
						"subscribers": 3,
						"messages":    500,
					},
					"docker.events": map[string]interface{}{
						"subscribers": 2,
						"messages":    250,
					},
				},
			},
			"uptime": map[string]interface{}{
				"type":        "integer",
				"description": "WebSocket server uptime in seconds",
				"example":     3600,
				"minimum":     0,
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"connections", "subscriptions", "messages_sent", "messages_received", "uptime", "last_updated"},
	}
}

func getWebSocketConnectionSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Connection ID",
				"example":     "conn_1234567890",
			},
			"remote_addr": map[string]interface{}{
				"type":        "string",
				"description": "Client IP address",
				"example":     "192.168.1.100",
			},
			"user_agent": map[string]interface{}{
				"type":        "string",
				"description": "Client user agent",
				"example":     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			},
			"connected_at": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Connection timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
			"subscriptions": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "Active subscriptions",
				"example":     []string{"system.stats", "docker.events"},
			},
			"messages_sent": map[string]interface{}{
				"type":        "integer",
				"description": "Messages sent to this connection",
				"example":     100,
				"minimum":     0,
			},
			"messages_received": map[string]interface{}{
				"type":        "integer",
				"description": "Messages received from this connection",
				"example":     10,
				"minimum":     0,
			},
			"last_activity": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last activity timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"id", "remote_addr", "connected_at", "subscriptions", "last_activity"},
	}
}

func getDockerEventsStreamSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Event type",
				"example":     "container",
			},
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Event action",
				"example":     "start",
			},
			"actor": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Container/image ID",
						"example":     "1234567890ab",
					},
					"attributes": map[string]interface{}{
						"type":        "object",
						"description": "Event attributes",
						"additionalProperties": map[string]interface{}{
							"type": "string",
						},
						"example": map[string]interface{}{
							"name":  "plex",
							"image": "plexinc/pms-docker:latest",
						},
					},
				},
			},
			"time": map[string]interface{}{
				"type":        "integer",
				"description": "Event timestamp (Unix)",
				"example":     1640995200,
			},
			"timeNano": map[string]interface{}{
				"type":        "integer",
				"description": "Event timestamp (nanoseconds)",
				"example":     1640995200000000000,
			},
		},
		"required": []string{"type", "action", "actor", "time"},
	}
}

func getSystemStatsStreamSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"cpu": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"usage_percent": map[string]interface{}{
						"type":        "number",
						"description": "CPU usage percentage",
						"example":     25.5,
					},
					"load_average": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "number",
						},
						"description": "Load average (1, 5, 15 minutes)",
						"example":     []float64{0.5, 0.7, 0.8},
					},
				},
			},
			"memory": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"usage_percent": map[string]interface{}{
						"type":        "number",
						"description": "Memory usage percentage",
						"example":     50.0,
					},
					"used": map[string]interface{}{
						"type":        "integer",
						"description": "Used memory in bytes",
						"example":     17179869184,
					},
					"total": map[string]interface{}{
						"type":        "integer",
						"description": "Total memory in bytes",
						"example":     34359738368,
					},
				},
			},
			"network": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"bytes_sent": map[string]interface{}{
						"type":        "integer",
						"description": "Network bytes sent",
						"example":     1048576,
					},
					"bytes_recv": map[string]interface{}{
						"type":        "integer",
						"description": "Network bytes received",
						"example":     2097152,
					},
				},
			},
			"timestamp": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Stats timestamp",
				"example":     "2024-01-01T12:00:00Z",
			},
		},
		"required": []string{"cpu", "memory", "timestamp"},
	}
}

func getStorageStatusStreamSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"array": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"status": map[string]interface{}{
						"type":        "string",
						"description": "Array status",
						"enum":        []string{"started", "stopped", "starting", "stopping"},
						"example":     "started",
					},
					"state": map[string]interface{}{
						"type":        "string",
						"description": "Array state",
						"enum":        []string{"normal", "degraded", "invalid", "emulated"},
						"example":     "normal",
					},
					"usage_percent": map[string]interface{}{
						"type":        "number",
						"description": "Array usage percentage",
						"example":     50.0,
					},
				},
			},
			"disks": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Disk name",
							"example":     "disk1",
						},
						"status": map[string]interface{}{
							"type":        "string",
							"description": "Disk status",
							"enum":        []string{"active", "standby", "spun_down", "error", "missing"},
							"example":     "active",
						},
						"temperature": map[string]interface{}{
							"type":        "number",
							"description": "Disk temperature in Celsius",
							"example":     35.0,
						},
						"usage_percent": map[string]interface{}{
							"type":        "number",
							"description": "Disk usage percentage",
							"example":     50.0,
						},
					},
				},
			},
			"parity": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"check_status": map[string]interface{}{
						"type":        "string",
						"description": "Parity check status",
						"enum":        []string{"idle", "running", "paused", "cancelled"},
						"example":     "idle",
					},
					"progress": map[string]interface{}{
						"type":        "number",
						"description": "Check progress percentage",
						"example":     0.0,
					},
					"errors": map[string]interface{}{
						"type":        "integer",
						"description": "Number of errors found",
						"example":     0,
					},
				},
			},
			"timestamp": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Status timestamp",
				"example":     "2024-01-01T12:00:00Z",
			},
		},
		"required": []string{"array", "disks", "parity", "timestamp"},
	}
}
