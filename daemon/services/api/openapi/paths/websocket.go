package paths

// GetWebSocketPaths returns all WebSocket API paths
func GetWebSocketPaths() map[string]interface{} {
	return map[string]interface{}{
		"/api/v1/ws":          getWebSocketPath(),
		"/api/v1/ws/stats":    getWebSocketStatsPath(),
		"/api/v1/ws/channels": getWebSocketChannelsPath(),
	}
}

func getWebSocketPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "WebSocket connection",
			"description": "Establish WebSocket connection for real-time updates",
			"operationId": "connectWebSocket",
			"tags":        []string{"WebSocket"},
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

func getWebSocketStatsPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get WebSocket statistics",
			"description": "Retrieve WebSocket server statistics including connections, subscriptions, and message counts",
			"operationId": "getWebSocketStats",
			"tags":        []string{"WebSocket"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "WebSocket statistics retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/WebSocketStats",
											},
										},
									},
								},
							},
						},
					},
				},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getWebSocketChannelsPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "List WebSocket channels",
			"description": "Retrieve a list of available WebSocket channels for real-time updates",
			"operationId": "listWebSocketChannels",
			"tags":        []string{"WebSocket"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "WebSocket channels retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"type": "array",
												"items": map[string]interface{}{
													"$ref": "#/components/schemas/WebSocketChannel",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}
