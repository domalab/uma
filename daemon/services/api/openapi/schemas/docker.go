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
				"example":     "jackett",
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
				"example":     "lscr.io/linuxserver/jackett",
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
			"environment": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "Environment variables",
				"example":     []string{"PATH=/usr/local/sbin:/usr/local/bin", "HOME=/root"},
			},
			"networks": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type": "string",
						},
						"ip_address": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"description": "Network configurations",
			},
			"restart_policy": map[string]interface{}{
				"type":        "string",
				"description": "Container restart policy",
				"enum":        []string{"no", "always", "unless-stopped", "on-failure"},
				"example":     "unless-stopped",
			},
			"started_at": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Container start timestamp",
				"example":     "2025-06-16T14:30:00Z",
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
				"example":     []string{"jackett", "homeassistant", "qbittorrent"},
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
				"example":     "container1",
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
				"example":     "container1",
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
				"example":     []string{"lscr.io/linuxserver/jackett:latest"},
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
				"example":     13,
				"minimum":     0,
			},
			"containers_running": map[string]interface{}{
				"type":        "integer",
				"description": "Number of running containers",
				"example":     13,
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
				"example":     0,
				"minimum":     0,
			},
			"images": map[string]interface{}{
				"type":        "integer",
				"description": "Total number of images",
				"example":     13,
				"minimum":     0,
			},
			"server_version": map[string]interface{}{
				"type":        "string",
				"description": "Docker server version",
				"example":     "27.5.1",
			},
			"CPUSet": map[string]interface{}{
				"type":        "boolean",
				"description": "CPU set support",
				"example":     true,
			},
			"CgroupVersion": map[string]interface{}{
				"type":        "string",
				"description": "Cgroup version",
				"example":     "2",
			},
			"LiveRestoreEnabled": map[string]interface{}{
				"type":        "boolean",
				"description": "Live restore enabled",
				"example":     false,
			},
			"RuncCommit": map[string]interface{}{
				"type":        "object",
				"description": "Runc commit information",
				"properties": map[string]interface{}{
					"Expected": map[string]interface{}{
						"type":        "string",
						"description": "Expected runc version",
						"example":     "v1.2.4-0-g6c52b3f",
					},
					"ID": map[string]interface{}{
						"type":        "string",
						"description": "Runc commit ID",
						"example":     "v1.2.4-0-g6c52b3f",
					},
				},
			},
			"SystemTime": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "System time",
				"example":     "2025-06-20T10:57:01Z",
			},
			"LoggingDriver": map[string]interface{}{
				"type":        "string",
				"description": "Logging driver",
				"example":     "json-file",
			},
			"NGoroutines": map[string]interface{}{
				"type":        "integer",
				"description": "Number of goroutines",
				"example":     153,
				"minimum":     0,
			},
			"Name": map[string]interface{}{
				"type":        "string",
				"description": "Docker daemon name",
				"example":     "Cube",
			},
			"RegistryConfig": map[string]interface{}{
				"type":                 "object",
				"description":          "Registry configuration",
				"additionalProperties": true,
			},
			"CPUShares": map[string]interface{}{
				"type":        "boolean",
				"description": "CPU shares support",
				"example":     true,
			},
			"Debug": map[string]interface{}{
				"type":        "boolean",
				"description": "Debug mode enabled",
				"example":     false,
			},
			"HttpProxy": map[string]interface{}{
				"type":        "string",
				"description": "HTTP proxy setting",
				"example":     "",
			},
			"IPv4Forwarding": map[string]interface{}{
				"type":        "boolean",
				"description": "IPv4 forwarding enabled",
				"example":     true,
			},
			"Isolation": map[string]interface{}{
				"type":        "string",
				"description": "Container isolation",
				"example":     "",
			},
			"KernelVersion": map[string]interface{}{
				"type":        "string",
				"description": "Kernel version",
				"example":     "6.12.24-Unraid",
			},
			"OperatingSystem": map[string]interface{}{
				"type":        "string",
				"description": "Operating system",
				"example":     "Unraid OS 7.1 x86_64",
			},
			"SecurityOptions": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "Security options",
				"example":     []string{"name=seccomp,profile=builtin", "name=cgroupns"},
			},
			"CpuCfsPeriod": map[string]interface{}{
				"type":        "boolean",
				"description": "CPU CFS period support",
				"example":     true,
			},
			"DriverStatus": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"description": "Driver status",
				"example":     [][]string{{"Btrfs", ""}},
			},
			"NEventsListener": map[string]interface{}{
				"type":        "integer",
				"description": "Number of events listeners",
				"example":     1,
				"minimum":     0,
			},
			"PidsLimit": map[string]interface{}{
				"type":        "boolean",
				"description": "PIDs limit support",
				"example":     true,
			},
			"Containers": map[string]interface{}{
				"type":        "integer",
				"description": "Total containers (alternative field)",
				"example":     13,
				"minimum":     0,
			},
			"ContainersRunning": map[string]interface{}{
				"type":        "integer",
				"description": "Running containers (alternative field)",
				"example":     13,
				"minimum":     0,
			},
			"ID": map[string]interface{}{
				"type":        "string",
				"description": "Docker daemon ID",
				"example":     "dcf99289-02ab-4aaf-b8f6-562f8ca37734",
			},
			"MemoryLimit": map[string]interface{}{
				"type":        "boolean",
				"description": "Memory limit support",
				"example":     true,
			},
			"BridgeNfIptables": map[string]interface{}{
				"type":        "boolean",
				"description": "Bridge netfilter iptables support",
				"example":     false,
			},
			"ClientInfo": map[string]interface{}{
				"type":                 "object",
				"description":          "Docker client information",
				"additionalProperties": true,
			},
			"Images": map[string]interface{}{
				"type":        "integer",
				"description": "Total images (alternative field)",
				"example":     13,
				"minimum":     0,
			},
			"NCPU": map[string]interface{}{
				"type":        "integer",
				"description": "Number of CPUs",
				"example":     12,
				"minimum":     1,
			},
			"SwapLimit": map[string]interface{}{
				"type":        "boolean",
				"description": "Swap limit support",
				"example":     false,
			},
			"ContainersStopped": map[string]interface{}{
				"type":        "integer",
				"description": "Stopped containers (alternative field)",
				"example":     0,
				"minimum":     0,
			},
			"CpuCfsQuota": map[string]interface{}{
				"type":        "boolean",
				"description": "CPU CFS quota support",
				"example":     true,
			},
			"DefaultRuntime": map[string]interface{}{
				"type":        "string",
				"description": "Default container runtime",
				"example":     "runc",
			},
			"NFd": map[string]interface{}{
				"type":        "integer",
				"description": "Number of file descriptors",
				"example":     125,
				"minimum":     0,
			},
			"Runtimes": map[string]interface{}{
				"type":                 "object",
				"description":          "Available container runtimes",
				"additionalProperties": true,
			},
			"BridgeNfIp6tables": map[string]interface{}{
				"type":        "boolean",
				"description": "Bridge netfilter ip6tables support",
				"example":     false,
			},
			"InitBinary": map[string]interface{}{
				"type":        "string",
				"description": "Init binary path",
				"example":     "docker-init",
			},
			"Labels": map[string]interface{}{
				"type":        "array",
				"description": "Docker daemon labels",
				"items": map[string]interface{}{
					"type": "string",
				},
				"example": []string{},
			},
			"ServerVersion": map[string]interface{}{
				"type":        "string",
				"description": "Server version (alternative field)",
				"example":     "27.5.1",
			},
			"HttpsProxy": map[string]interface{}{
				"type":        "string",
				"description": "HTTPS proxy setting",
				"example":     "",
			},
			"NoProxy": map[string]interface{}{
				"type":        "string",
				"description": "No proxy setting",
				"example":     "",
			},
			"Containerd": map[string]interface{}{
				"type":                 "object",
				"description":          "Containerd information",
				"additionalProperties": true,
			},
			"GenericResources": map[string]interface{}{
				"anyOf": []interface{}{
					map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
						},
					},
					map[string]interface{}{"type": "null"},
				},
				"description": "Generic resources (null if none)",
				"example":     []interface{}{},
			},
			"CDISpecDirs": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "CDI specification directories",
				"example":     []string{"/etc/cdi", "/var/run/cdi"},
			},
			"DockerRootDir": map[string]interface{}{
				"type":        "string",
				"description": "Docker root directory",
				"example":     "/var/lib/docker",
			},
			"Driver": map[string]interface{}{
				"type":        "string",
				"description": "Storage driver",
				"example":     "btrfs",
			},
			"Plugins": map[string]interface{}{
				"type":                 "object",
				"description":          "Docker plugins",
				"additionalProperties": true,
			},
			"Architecture": map[string]interface{}{
				"type":        "string",
				"description": "System architecture",
				"example":     "x86_64",
			},
			"OomKillDisable": map[string]interface{}{
				"type":        "boolean",
				"description": "OOM kill disable support",
				"example":     true,
			},
			"ContainerdCommit": map[string]interface{}{
				"type":        "object",
				"description": "Containerd commit information",
				"properties": map[string]interface{}{
					"Expected": map[string]interface{}{
						"type":        "string",
						"description": "Expected containerd version",
						"example":     "v1.7.24-0-g61f9fd88",
					},
					"ID": map[string]interface{}{
						"type":        "string",
						"description": "Containerd commit ID",
						"example":     "v1.7.24-0-g61f9fd88",
					},
				},
			},
			"Warnings": map[string]interface{}{
				"type":        "array",
				"description": "Docker daemon warnings",
				"items": map[string]interface{}{
					"type": "string",
				},
				"example": []string{},
			},
			"CgroupDriver": map[string]interface{}{
				"type":        "string",
				"description": "Cgroup driver",
				"example":     "systemd",
			},
			"Swarm": map[string]interface{}{
				"type":                 "object",
				"description":          "Docker swarm information",
				"additionalProperties": true,
			},
			"ExperimentalBuild": map[string]interface{}{
				"type":        "boolean",
				"description": "Experimental build flag",
				"example":     false,
			},
			"MemTotal": map[string]interface{}{
				"type":        "integer",
				"description": "Total memory in bytes",
				"example":     67645440000,
				"minimum":     0,
			},
			"IndexServerAddress": map[string]interface{}{
				"type":        "string",
				"description": "Index server address",
				"example":     "https://index.docker.io/v1/",
			},
			"InitCommit": map[string]interface{}{
				"type":        "object",
				"description": "Init commit information",
				"properties": map[string]interface{}{
					"Expected": map[string]interface{}{
						"type":        "string",
						"description": "Expected init version",
						"example":     "de40ad0",
					},
					"ID": map[string]interface{}{
						"type":        "string",
						"description": "Init commit ID",
						"example":     "de40ad0",
					},
				},
			},
			"OSType": map[string]interface{}{
				"type":        "string",
				"description": "Operating system type",
				"example":     "linux",
			},
			"ContainersPaused": map[string]interface{}{
				"type":        "integer",
				"description": "Paused containers (alternative field)",
				"example":     0,
				"minimum":     0,
			},
			"OSVersion": map[string]interface{}{
				"type":        "string",
				"description": "OS version",
				"example":     "7.1",
			},
			"ProductLicense": map[string]interface{}{
				"type":        "string",
				"description": "Product license",
				"example":     "Community Engine",
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-20T01:06:25Z",
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
				"name":   "jackett",
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
				"tags": []string{"organization/application:latest"},
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
