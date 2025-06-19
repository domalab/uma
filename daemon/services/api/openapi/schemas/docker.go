package schemas

// GetDockerSchemas returns Docker-related schemas
func GetDockerSchemas() map[string]interface{} {
	return map[string]interface{}{
		"ContainerInfo":              getContainerInfoSchema(),
		"ContainerState":             getContainerStateSchema(),
		"ContainerOperationResult":   getContainerOperationResultSchema(),
		"ContainerOperationResponse": getContainerOperationResponseSchema(),
		"BulkOperationRequest":       getBulkOperationRequestSchema(),
		"BulkOperationResponse":      getBulkOperationResponseSchema(),
		"BulkOperationSummary":       getBulkOperationSummarySchema(),
		"DockerImage":                getDockerImageSchema(),
		"DockerNetwork":              getDockerNetworkSchema(),
		"DockerInfo":                 getDockerInfoSchema(),
		"ContainerPort":              getContainerPortSchema(),
		"DockerContainerList":        getDockerContainerListSchema(),
		"DockerContainerInfo":        getDockerContainerInfoSchema(),
		"DockerImageList":            getDockerImageListSchema(),
		"DockerNetworkList":          getDockerNetworkListSchema(),
	}
}

func getContainerInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Container ID",
				"example":     "1234567890ab",
				"pattern":     "^[a-f0-9]{12}$",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Container name",
				"example":     "plex",
				"pattern":     "^[a-zA-Z0-9][a-zA-Z0-9_.-]+$",
			},
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Container status",
				"enum":        []string{"created", "running", "paused", "restarting", "removing", "exited", "dead"},
				"example":     "running",
			},
			"state": map[string]interface{}{
				"$ref": "#/components/schemas/ContainerState",
			},
			"created": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Container creation timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
			"image": map[string]interface{}{
				"type":        "string",
				"description": "Container image",
				"example":     "linuxserver/plex:latest",
			},
			"ports": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"$ref": "#/components/schemas/ContainerPort",
				},
				"description": "Container port mappings",
			},
			"labels": map[string]interface{}{
				"type":        "object",
				"description": "Container labels",
				"additionalProperties": map[string]interface{}{
					"type": "string",
				},
			},
			"mounts": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"source": map[string]interface{}{
							"type": "string",
						},
						"destination": map[string]interface{}{
							"type": "string",
						},
						"mode": map[string]interface{}{
							"type": "string",
						},
					},
				},
			},
		},
		"required": []string{"id", "name", "status", "created"},
	}
}

func getContainerStateSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Container state status",
				"example":     "running",
			},
			"running": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the container is running",
				"example":     true,
			},
			"paused": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the container is paused",
				"example":     false,
			},
			"restarting": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the container is restarting",
				"example":     false,
			},
			"pid": map[string]interface{}{
				"type":        "integer",
				"description": "Container process ID",
				"example":     1234,
			},
			"exit_code": map[string]interface{}{
				"type":        "integer",
				"description": "Container exit code",
				"example":     0,
			},
			"started_at": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Container start timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
			"finished_at": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Container finish timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
	}
}

func getContainerPortSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"private_port": map[string]interface{}{
				"type":        "integer",
				"description": "Container internal port",
				"example":     32400,
				"minimum":     1,
				"maximum":     65535,
			},
			"public_port": map[string]interface{}{
				"type":        "integer",
				"description": "Host external port",
				"example":     32400,
				"minimum":     1,
				"maximum":     65535,
			},
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Port protocol",
				"enum":        []string{"tcp", "udp"},
				"example":     "tcp",
			},
			"ip": map[string]interface{}{
				"type":        "string",
				"description": "Bind IP address",
				"example":     "0.0.0.0",
			},
		},
		"required": []string{"private_port", "type"},
	}
}

func getBulkOperationRequestSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"container_ids": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type":    "string",
					"pattern": "^[a-zA-Z0-9][a-zA-Z0-9_.-]+$",
				},
				"description": "Array of container IDs or names",
				"example":     []string{"plex", "nginx", "sonarr"},
				"minItems":    1,
				"maxItems":    50,
				"uniqueItems": true,
			},
		},
		"required": []string{"container_ids"},
	}
}

func getBulkOperationResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"summary": map[string]interface{}{
				"$ref": "#/components/schemas/BulkOperationSummary",
			},
			"results": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"$ref": "#/components/schemas/ContainerOperationResult",
				},
				"description": "Individual operation results",
			},
		},
		"required": []string{"summary", "results"},
	}
}

func getBulkOperationSummarySchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"total": map[string]interface{}{
				"type":        "integer",
				"description": "Total number of operations attempted",
				"example":     3,
				"minimum":     0,
			},
			"successful": map[string]interface{}{
				"type":        "integer",
				"description": "Number of successful operations",
				"example":     2,
				"minimum":     0,
			},
			"failed": map[string]interface{}{
				"type":        "integer",
				"description": "Number of failed operations",
				"example":     1,
				"minimum":     0,
			},
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "Type of operation performed",
				"enum":        []string{"start", "stop", "restart", "pause", "resume"},
				"example":     "start",
			},
		},
		"required": []string{"total", "successful", "failed", "operation"},
	}
}

func getContainerOperationResultSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"container_id": map[string]interface{}{
				"type":        "string",
				"description": "Container ID or name",
				"example":     "plex",
			},
			"success": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the operation was successful",
				"example":     true,
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Operation result message",
				"example":     "Container started successfully",
			},
			"error": map[string]interface{}{
				"type":        "string",
				"description": "Error message if operation failed",
				"example":     "Container not found",
			},
		},
		"required": []string{"container_id", "success"},
	}
}

func getContainerOperationResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the operation was successful",
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
		},
		"required": []string{"success", "message", "container_id"},
	}
}

func getDockerImageSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Image ID",
				"example":     "sha256:1234567890ab",
			},
			"repo_tags": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "Repository tags",
				"example":     []string{"linuxserver/plex:latest"},
			},
			"size": map[string]interface{}{
				"type":        "integer",
				"description": "Image size in bytes",
				"example":     1073741824,
				"minimum":     0,
			},
			"created": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Image creation timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"id", "repo_tags", "size", "created"},
	}
}

func getDockerNetworkSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Network ID",
				"example":     "1234567890ab",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Network name",
				"example":     "bridge",
			},
			"driver": map[string]interface{}{
				"type":        "string",
				"description": "Network driver",
				"example":     "bridge",
			},
			"scope": map[string]interface{}{
				"type":        "string",
				"description": "Network scope",
				"example":     "local",
			},
		},
		"required": []string{"id", "name", "driver"},
	}
}

func getDockerInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"containers": map[string]interface{}{
				"type":        "integer",
				"description": "Total number of containers",
				"example":     15,
				"minimum":     0,
			},
			"containers_running": map[string]interface{}{
				"type":        "integer",
				"description": "Number of running containers",
				"example":     12,
				"minimum":     0,
			},
			"containers_paused": map[string]interface{}{
				"type":        "integer",
				"description": "Number of paused containers",
				"example":     0,
				"minimum":     0,
			},
			"containers_stopped": map[string]interface{}{
				"type":        "integer",
				"description": "Number of stopped containers",
				"example":     3,
				"minimum":     0,
			},
			"images": map[string]interface{}{
				"type":        "integer",
				"description": "Total number of images",
				"example":     25,
				"minimum":     0,
			},
			"server_version": map[string]interface{}{
				"type":        "string",
				"description": "Docker server version",
				"example":     "20.10.21",
			},
		},
		"required": []string{"containers", "containers_running", "containers_paused", "containers_stopped", "images"},
	}
}

func getDockerContainerListSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":        "array",
		"description": "List of Docker containers",
		"items": map[string]interface{}{
			"$ref": "#/components/schemas/ContainerInfo",
		},
		"example": []interface{}{
			map[string]interface{}{
				"id":     "1234567890ab",
				"name":   "plex",
				"status": "running",
			},
		},
	}
}

func getDockerContainerInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"allOf": []interface{}{
			map[string]interface{}{"$ref": "#/components/schemas/ContainerInfo"},
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"logs": map[string]interface{}{
						"type":        "array",
						"description": "Recent container logs",
						"items": map[string]interface{}{
							"type": "string",
						},
						"example": []string{"Container started", "Service initialized"},
					},
					"stats": map[string]interface{}{
						"type":        "object",
						"description": "Container resource statistics",
						"properties": map[string]interface{}{
							"cpu_percent": map[string]interface{}{
								"type":        "number",
								"description": "CPU usage percentage",
								"example":     15.5,
							},
							"memory_usage": map[string]interface{}{
								"type":        "integer",
								"description": "Memory usage in bytes",
								"example":     134217728,
							},
							"memory_limit": map[string]interface{}{
								"type":        "integer",
								"description": "Memory limit in bytes",
								"example":     1073741824,
							},
						},
					},
				},
			},
		},
	}
}

func getDockerImageListSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":        "array",
		"description": "List of Docker images",
		"items": map[string]interface{}{
			"$ref": "#/components/schemas/DockerImage",
		},
		"example": []interface{}{
			map[string]interface{}{
				"id":   "sha256:abc123",
				"tags": []string{"nginx:latest"},
				"size": 142000000,
			},
		},
	}
}

func getDockerNetworkListSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":        "array",
		"description": "List of Docker networks",
		"items": map[string]interface{}{
			"$ref": "#/components/schemas/DockerNetwork",
		},
		"example": []interface{}{
			map[string]interface{}{
				"id":     "network123",
				"name":   "bridge",
				"driver": "bridge",
			},
		},
	}
}
