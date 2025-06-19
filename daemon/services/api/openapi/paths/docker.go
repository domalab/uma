package paths

// GetDockerPaths returns all Docker-related API paths
func GetDockerPaths() map[string]interface{} {
	return map[string]interface{}{
		"/api/v1/docker/containers":              getContainersListPath(),
		"/api/v1/docker/containers/{id}":         getContainerByIDPath(),
		"/api/v1/docker/containers/{id}/start":   getContainerStartPath(),
		"/api/v1/docker/containers/{id}/stop":    getContainerStopPath(),
		"/api/v1/docker/containers/{id}/restart": getContainerRestartPath(),
		"/api/v1/docker/containers/{id}/pause":   getContainerPausePath(),
		"/api/v1/docker/containers/{id}/resume":  getContainerResumePath(),
		"/api/v1/docker/containers/bulk/start":   getBulkContainerStartPath(),
		"/api/v1/docker/containers/bulk/stop":    getBulkContainerStopPath(),
		"/api/v1/docker/containers/bulk/restart": getBulkContainerRestartPath(),
		"/api/v1/docker/containers/bulk/pause":   getBulkContainerPausePath(),
		"/api/v1/docker/containers/bulk/resume":  getBulkContainerResumePath(),
		"/api/v1/docker/images":                  getImagesListPath(),
		"/api/v1/docker/networks":                getNetworksListPath(),
		"/api/v1/docker/info":                    getDockerInfoPath(),
	}
}

func getContainersListPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "List Docker containers",
			"description": "Retrieve a list of Docker containers with optional filtering and pagination",
			"operationId": "listContainers",
			"tags":        []string{"Docker"},
			"parameters": []interface{}{
				map[string]interface{}{"$ref": "#/components/parameters/PageParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/LimitParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/AllContainersParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/StatusFilterParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/VerboseParameter"},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "List of containers retrieved successfully",
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
													"$ref": "#/components/schemas/ContainerInfo",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
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

func getContainerByIDPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get container information",
			"description": "Retrieve detailed information about a specific Docker container",
			"operationId": "getContainer",
			"tags":        []string{"Docker"},
			"parameters": []interface{}{
				map[string]interface{}{"$ref": "#/components/parameters/ContainerIDParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/VerboseParameter"},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Container information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/ContainerInfo",
											},
										},
									},
								},
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"404": map[string]interface{}{"$ref": "#/components/responses/NotFound"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getContainerStartPath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Start container",
			"description": "Start a Docker container",
			"operationId": "startContainer",
			"tags":        []string{"Docker"},
			"parameters": []interface{}{
				map[string]interface{}{"$ref": "#/components/parameters/ContainerIDParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/TimeoutParameter"},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Container started successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/ContainerOperationResponse",
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"403": map[string]interface{}{"$ref": "#/components/responses/Forbidden"},
				"404": map[string]interface{}{"$ref": "#/components/responses/NotFound"},
				"409": map[string]interface{}{"$ref": "#/components/responses/Conflict"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getContainerStopPath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Stop container",
			"description": "Stop a Docker container gracefully or forcefully",
			"operationId": "stopContainer",
			"tags":        []string{"Docker"},
			"parameters": []interface{}{
				map[string]interface{}{"$ref": "#/components/parameters/ContainerIDParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/ForceParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/TimeoutParameter"},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Container stopped successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/ContainerOperationResponse",
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"403": map[string]interface{}{"$ref": "#/components/responses/Forbidden"},
				"404": map[string]interface{}{"$ref": "#/components/responses/NotFound"},
				"409": map[string]interface{}{"$ref": "#/components/responses/Conflict"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getContainerRestartPath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Restart container",
			"description": "Restart a Docker container",
			"operationId": "restartContainer",
			"tags":        []string{"Docker"},
			"parameters": []interface{}{
				map[string]interface{}{"$ref": "#/components/parameters/ContainerIDParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/TimeoutParameter"},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Container restarted successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/ContainerOperationResponse",
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"403": map[string]interface{}{"$ref": "#/components/responses/Forbidden"},
				"404": map[string]interface{}{"$ref": "#/components/responses/NotFound"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getContainerPausePath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Pause container",
			"description": "Pause a running Docker container",
			"operationId": "pauseContainer",
			"tags":        []string{"Docker"},
			"parameters": []interface{}{
				map[string]interface{}{"$ref": "#/components/parameters/ContainerIDParameter"},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Container paused successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/ContainerOperationResponse",
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"403": map[string]interface{}{"$ref": "#/components/responses/Forbidden"},
				"404": map[string]interface{}{"$ref": "#/components/responses/NotFound"},
				"409": map[string]interface{}{"$ref": "#/components/responses/Conflict"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getContainerResumePath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Resume container",
			"description": "Resume a paused Docker container",
			"operationId": "resumeContainer",
			"tags":        []string{"Docker"},
			"parameters": []interface{}{
				map[string]interface{}{"$ref": "#/components/parameters/ContainerIDParameter"},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Container resumed successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/ContainerOperationResponse",
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"403": map[string]interface{}{"$ref": "#/components/responses/Forbidden"},
				"404": map[string]interface{}{"$ref": "#/components/responses/NotFound"},
				"409": map[string]interface{}{"$ref": "#/components/responses/Conflict"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getBulkContainerStartPath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Start multiple containers",
			"description": "Start multiple Docker containers in a single operation",
			"operationId": "bulkStartContainers",
			"tags":        []string{"Docker"},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/BulkOperationRequest",
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Bulk start operation completed",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/BulkOperationResponse",
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"403": map[string]interface{}{"$ref": "#/components/responses/Forbidden"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getBulkContainerStopPath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Stop multiple containers",
			"description": "Stop multiple Docker containers in a single operation",
			"operationId": "bulkStopContainers",
			"tags":        []string{"Docker"},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/BulkOperationRequest",
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Bulk stop operation completed",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/BulkOperationResponse",
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"403": map[string]interface{}{"$ref": "#/components/responses/Forbidden"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getBulkContainerRestartPath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Restart multiple containers",
			"description": "Restart multiple Docker containers in a single operation",
			"operationId": "bulkRestartContainers",
			"tags":        []string{"Docker"},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/BulkOperationRequest",
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Bulk restart operation completed",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/BulkOperationResponse",
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"403": map[string]interface{}{"$ref": "#/components/responses/Forbidden"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getBulkContainerPausePath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Pause multiple containers",
			"description": "Pause multiple Docker containers in a single operation",
			"operationId": "bulkPauseContainers",
			"tags":        []string{"Docker"},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/BulkOperationRequest",
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Bulk pause operation completed",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/BulkOperationResponse",
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"403": map[string]interface{}{"$ref": "#/components/responses/Forbidden"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getBulkContainerResumePath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Resume multiple containers",
			"description": "Resume multiple paused Docker containers in a single operation",
			"operationId": "bulkResumeContainers",
			"tags":        []string{"Docker"},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/BulkOperationRequest",
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Bulk resume operation completed",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/BulkOperationResponse",
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"403": map[string]interface{}{"$ref": "#/components/responses/Forbidden"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getImagesListPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "List Docker images",
			"description": "Retrieve a list of Docker images",
			"operationId": "listImages",
			"tags":        []string{"Docker"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "List of images retrieved successfully",
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
													"$ref": "#/components/schemas/DockerImage",
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

func getNetworksListPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "List Docker networks",
			"description": "Retrieve a list of Docker networks",
			"operationId": "listNetworks",
			"tags":        []string{"Docker"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "List of networks retrieved successfully",
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
													"$ref": "#/components/schemas/DockerNetwork",
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

func getDockerInfoPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get Docker information",
			"description": "Retrieve Docker daemon information and statistics",
			"operationId": "getDockerInfo",
			"tags":        []string{"Docker"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Docker information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/DockerInfo",
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
