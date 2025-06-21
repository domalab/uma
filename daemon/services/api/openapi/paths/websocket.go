package paths

// GetWebSocketPaths returns all WebSocket API paths
func GetWebSocketPaths() map[string]interface{} {
	return map[string]interface{}{
		"/api/v1/ws": getUnifiedWebSocketPath(),
	}
}

func getUnifiedWebSocketPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Unified WebSocket connection with subscription management",
			"description": "Establish WebSocket connection for real-time updates with subscription management. Supports subscribing to specific event channels including system stats, Docker events, storage status, temperature alerts, and more.",
			"operationId": "connectUnifiedWebSocket",
			"tags":        []string{"WebSocket", "Real-time"},
			"parameters": []interface{}{
				map[string]interface{}{
					"name":        "Upgrade",
					"in":          "header",
					"description": "WebSocket upgrade header",
					"required":    true,
					"schema": map[string]interface{}{
						"type": "string",
						"enum": []string{"websocket"},
					},
				},
				map[string]interface{}{
					"name":        "Connection",
					"in":          "header",
					"description": "Connection upgrade header",
					"required":    true,
					"schema": map[string]interface{}{
						"type": "string",
						"enum": []string{"Upgrade"},
					},
				},
				map[string]interface{}{
					"name":        "Sec-WebSocket-Key",
					"in":          "header",
					"description": "WebSocket key for handshake",
					"required":    true,
					"schema": map[string]interface{}{
						"type": "string",
					},
				},
			},
			"responses": map[string]interface{}{
				"101": map[string]interface{}{
					"description": "Switching Protocols - WebSocket connection established",
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"426": map[string]interface{}{
					"description": "Upgrade Required - Invalid WebSocket upgrade request",
				},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}
