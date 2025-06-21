package paths

// GetSystemPaths returns all system monitoring and control API paths
func GetSystemPaths() map[string]interface{} {
	return map[string]interface{}{
		"/api/v1/system/info":                   getSystemInfoPath(),
		"/api/v1/system/cpu":                    getCPUInfoPath(),
		"/api/v1/system/memory":                 getMemoryInfoPath(),
		"/api/v1/system/temperature":            getTemperaturePath(),
		"/api/v1/system/temperatures":           getTemperaturesPath(),
		"/api/v1/system/logs/all":               getSystemLogsAllPath(),
		"/api/v1/system/fans":                   getFansPath(),
		"/api/v1/system/gpu":                    getGPUInfoPath(),
		"/api/v1/system/ups":                    getUPSInfoPath(),
		"/api/v1/system/network":                getNetworkInfoPath(),
		"/api/v1/system/resources":              getSystemResourcesPath(),
		"/api/v1/system/filesystems":            getFilesystemsPath(),
		"/api/v1/system/scripts":                getSystemScriptsPath(),
		"/api/v1/system/scripts/{id}":           getSystemScriptPath(),
		"/api/v1/system/execute":                getExecuteCommandPath(),
		"/api/v1/system/logs":                   getSystemLogsPath(),
		"/api/v1/system/shutdown":               getSystemShutdownPath(),
		"/api/v1/system/reboot":                 getSystemRebootPath(),
		"/api/v1/system/temperature/thresholds": getTemperatureThresholdsPath(),
		"/api/v1/system/temperature/alerts":     getTemperatureAlertsPath(),
	}
}

func getSystemInfoPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get system information",
			"description": "Retrieve general system information including hostname, kernel, uptime, and load average",
			"operationId": "getSystemInfo",
			"tags":        []string{"System"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "System information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/SystemInfo",
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

func getCPUInfoPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get CPU information",
			"description": "Retrieve CPU usage, core count, model, frequency, and temperature information",
			"operationId": "getCPUInfo",
			"tags":        []string{"System"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "CPU information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/CPUInfo",
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

func getMemoryInfoPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get memory information",
			"description": "Retrieve memory usage, total, available, buffers, cached, and swap information",
			"operationId": "getMemoryInfo",
			"tags":        []string{"System"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Memory information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/MemoryInfo",
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

func getTemperaturesPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get temperature sensors",
			"description": "Retrieve temperature data from all available sensors including CPU, motherboard, and other components",
			"operationId": "getTemperatures",
			"tags":        []string{"System"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Temperature data retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/TemperatureInfo",
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

func getTemperaturePath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get temperature data",
			"description": "Retrieve current temperature data from all available sensors (alias for /api/v1/system/temperatures)",
			"operationId": "getTemperature",
			"tags":        []string{"System"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Temperature data retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/TemperatureInfo",
											},
										},
									},
								},
							},
						},
					},
				},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
				"503": map[string]interface{}{
					"description": "Temperature sensors unavailable",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/Error",
							},
						},
					},
				},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getFansPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get fan information",
			"description": "Retrieve fan speed data from all available fan sensors",
			"operationId": "getFans",
			"tags":        []string{"System"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Fan data retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/FanInfo",
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

func getUPSInfoPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get UPS information",
			"description": "Retrieve UPS status, battery charge, runtime, load, and voltage information from apcupsd daemon",
			"operationId": "getUPSInfo",
			"tags":        []string{"System"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "UPS information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/UPSInfo",
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
				"503": map[string]interface{}{"$ref": "#/components/responses/ServiceUnavailable"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getSystemShutdownPath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Shutdown system",
			"description": "Initiate system shutdown with optional delay and message",
			"operationId": "shutdownSystem",
			"tags":        []string{"System"},
			"parameters": []interface{}{
				map[string]interface{}{"$ref": "#/components/parameters/ForceParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/TimeoutParameter"},
			},
			"responses": map[string]interface{}{
				"202": map[string]interface{}{"$ref": "#/components/responses/Accepted"},
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

func getGPUInfoPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get comprehensive GPU monitoring data",
			"description": "Retrieve detailed GPU information including utilization, memory usage, power consumption, clock frequencies, and vendor-specific metrics. Supports Intel (intel_gpu_top), NVIDIA (nvidia-smi), and AMD (rocm-smi) GPUs with real-time monitoring data.",
			"operationId": "getGPUInfo",
			"tags":        []string{"System"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "GPU information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/GPUInfo",
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

func getNetworkInfoPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get network information",
			"description": "Retrieve network interface information and statistics",
			"operationId": "getNetworkInfo",
			"tags":        []string{"System"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Network information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/NetworkInfo",
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

func getSystemResourcesPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get system resources",
			"description": "Retrieve comprehensive system resource information",
			"operationId": "getSystemResources",
			"tags":        []string{"System"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "System resources retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/SystemResources",
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

func getFilesystemsPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get filesystem information",
			"description": "Retrieve filesystem mount points and usage information",
			"operationId": "getFilesystems",
			"tags":        []string{"System"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Filesystem information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/FilesystemInfo",
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

func getSystemScriptsPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "List system scripts",
			"description": "Retrieve a list of available system scripts",
			"operationId": "listSystemScripts",
			"tags":        []string{"System"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "System scripts retrieved successfully",
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
													"$ref": "#/components/schemas/SystemScript",
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

func getSystemScriptPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get system script details",
			"description": "Retrieve details about a specific system script",
			"operationId": "getSystemScript",
			"tags":        []string{"System"},
			"parameters": []interface{}{
				map[string]interface{}{"$ref": "#/components/parameters/ScriptIDParameter"},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "System script details retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/SystemScript",
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
		"post": map[string]interface{}{
			"summary":     "Execute system script",
			"description": "Execute a specific system script with optional parameters",
			"operationId": "executeSystemScript",
			"tags":        []string{"System"},
			"parameters": []interface{}{
				map[string]interface{}{"$ref": "#/components/parameters/ScriptIDParameter"},
			},
			"requestBody": map[string]interface{}{
				"required": false,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/ScriptExecutionRequest",
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Script executed successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/ScriptExecutionResponse",
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

func getExecuteCommandPath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Execute system command",
			"description": "Execute a system command with optional parameters and timeout",
			"operationId": "executeCommand",
			"tags":        []string{"System"},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/CommandExecutionRequest",
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Command executed successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/CommandExecutionResponse",
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

func getSystemLogsPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get system logs",
			"description": "Retrieve system logs with optional filtering and pagination",
			"operationId": "getSystemLogs",
			"tags":        []string{"System"},
			"parameters": []interface{}{
				map[string]interface{}{"$ref": "#/components/parameters/PageParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/LimitParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/LogLevelParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/LogSourceParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/SinceParameter"},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "System logs retrieved successfully",
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
													"$ref": "#/components/schemas/LogEntry",
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

func getSystemRebootPath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Reboot system",
			"description": "Initiate system reboot with optional delay and message",
			"operationId": "rebootSystem",
			"tags":        []string{"System"},
			"parameters": []interface{}{
				map[string]interface{}{"$ref": "#/components/parameters/ForceParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/TimeoutParameter"},
			},
			"responses": map[string]interface{}{
				"202": map[string]interface{}{"$ref": "#/components/responses/Accepted"},
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

func getTemperatureThresholdsPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get temperature thresholds",
			"description": "Retrieve current temperature monitoring thresholds for different sensor types",
			"operationId": "getTemperatureThresholds",
			"tags":        []string{"System", "Monitoring"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Temperature thresholds retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/TemperatureThresholds",
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
		"put": map[string]interface{}{
			"summary":     "Update temperature thresholds",
			"description": "Update temperature monitoring thresholds for different sensor types",
			"operationId": "updateTemperatureThresholds",
			"tags":        []string{"System", "Monitoring"},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/TemperatureThresholds",
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Temperature thresholds updated successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/StandardResponse",
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

func getTemperatureAlertsPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get temperature alerts",
			"description": "Retrieve recent temperature alerts and threshold violations",
			"operationId": "getTemperatureAlerts",
			"tags":        []string{"System", "Monitoring", "Alerts"},
			"parameters": []interface{}{
				map[string]interface{}{
					"name":        "limit",
					"in":          "query",
					"description": "Maximum number of alerts to return",
					"required":    false,
					"schema": map[string]interface{}{
						"type":    "integer",
						"minimum": 1,
						"maximum": 100,
						"default": 50,
					},
				},
				map[string]interface{}{
					"name":        "level",
					"in":          "query",
					"description": "Filter alerts by severity level",
					"required":    false,
					"schema": map[string]interface{}{
						"type": "string",
						"enum": []string{"warning", "critical", "emergency"},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Temperature alerts retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"type": "object",
												"properties": map[string]interface{}{
													"alerts": map[string]interface{}{
														"type": "array",
														"items": map[string]interface{}{
															"$ref": "#/components/schemas/TemperatureAlert",
														},
													},
													"count": map[string]interface{}{
														"type":        "integer",
														"description": "Total number of alerts returned",
													},
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

func getSystemLogsAllPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get all system logs",
			"description": "Retrieve comprehensive system logs from all available sources including syslog, kernel, and application logs",
			"operationId": "getAllSystemLogs",
			"tags":        []string{"System"},
			"parameters": []map[string]interface{}{
				{
					"name":        "lines",
					"in":          "query",
					"description": "Number of log lines to retrieve",
					"required":    false,
					"schema": map[string]interface{}{
						"type":    "integer",
						"minimum": 1,
						"maximum": 10000,
						"default": 1000,
					},
				},
				{
					"name":        "since",
					"in":          "query",
					"description": "Retrieve logs since this timestamp",
					"required":    false,
					"schema": map[string]interface{}{
						"type":   "string",
						"format": "date-time",
					},
				},
				{
					"name":        "level",
					"in":          "query",
					"description": "Filter logs by severity level",
					"required":    false,
					"schema": map[string]interface{}{
						"type": "string",
						"enum": []string{"emergency", "alert", "critical", "error", "warning", "notice", "info", "debug"},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "System logs retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/SystemLogsAll",
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
