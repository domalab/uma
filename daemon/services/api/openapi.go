package api

import (
	"net/http"
)

// OpenAPISpec represents the OpenAPI 3.0 specification
type OpenAPISpec struct {
	OpenAPI    string                 `json:"openapi"`
	Info       OpenAPIInfo            `json:"info"`
	Servers    []OpenAPIServer        `json:"servers"`
	Paths      map[string]interface{} `json:"paths"`
	Components OpenAPIComponents      `json:"components"`
}

// OpenAPIInfo contains API information
type OpenAPIInfo struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Version     string         `json:"version"`
	Contact     OpenAPIContact `json:"contact"`
}

// OpenAPIContact contains contact information
type OpenAPIContact struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Email string `json:"email"`
}

// OpenAPIServer represents a server
type OpenAPIServer struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

// OpenAPIComponents contains reusable components
type OpenAPIComponents struct {
	Schemas   map[string]interface{} `json:"schemas"`
	Responses map[string]interface{} `json:"responses"`
}

// generateOpenAPISpec creates the complete OpenAPI specification
func (h *HTTPServer) generateOpenAPISpec() *OpenAPISpec {
	return &OpenAPISpec{
		OpenAPI: "3.1.1",
		Info: OpenAPIInfo{
			Title:       "UMA API",
			Description: "Unraid Management Agent API for system monitoring, Docker container management, VM control, and storage management",
			Version:     h.api.ctx.Config.Version,
			Contact: OpenAPIContact{
				Name:  "UMA Development Team",
				URL:   "https://github.com/domalab/uma",
				Email: "support@domalab.net",
			},
		},
		Servers: []OpenAPIServer{
			{
				URL:         "http://your-unraid-server:34600",
				Description: "Local UMA API server",
			},
			{
				URL:         "unix:/var/run/uma-api.sock",
				Description: "Unix socket API server",
			},
		},
		Paths:      h.generatePaths(),
		Components: h.generateComponents(),
	}
}

// generateComponents creates reusable OpenAPI components
func (h *HTTPServer) generateComponents() OpenAPIComponents {
	return OpenAPIComponents{
		Schemas: map[string]interface{}{
			"StandardResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"data": map[string]interface{}{
						"description": "The response data",
					},
					"pagination": map[string]interface{}{
						"$ref": "#/components/schemas/PaginationInfo",
					},
					"meta": map[string]interface{}{
						"$ref": "#/components/schemas/ResponseMeta",
					},
				},
			},
			"PaginationInfo": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page": map[string]interface{}{
						"type":        "integer",
						"description": "Current page number",
						"example":     1,
					},
					"per_page": map[string]interface{}{
						"type":        "integer",
						"description": "Number of items per page",
						"example":     50,
					},
					"total": map[string]interface{}{
						"type":        "integer",
						"description": "Total number of items",
						"example":     150,
					},
					"has_more": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether there are more pages available",
						"example":     true,
					},
					"total_pages": map[string]interface{}{
						"type":        "integer",
						"description": "Total number of pages",
						"example":     3,
					},
				},
			},
			"ResponseMeta": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"request_id": map[string]interface{}{
						"type":        "string",
						"description": "Unique request identifier for tracing",
						"example":     "req_1234567890_5678",
					},
					"version": map[string]interface{}{
						"type":        "string",
						"description": "API version",
						"example":     "1.0.0",
					},
					"timestamp": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "Response timestamp in ISO 8601 format",
						"example":     "2025-06-15T21:30:42Z",
					},
				},
			},
			"HealthResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"status": map[string]interface{}{
						"type":        "string",
						"description": "Overall health status",
						"enum":        []string{"healthy", "degraded", "unhealthy"},
						"examples":    []string{"healthy", "degraded"},
					},
					"version": map[string]interface{}{
						"type":        "string",
						"description": "UMA version",
						"examples":    []string{"1.0.0", "1.1.0"},
					},
					"service": map[string]interface{}{
						"type":        "string",
						"description": "Service name",
						"const":       "uma",
					},
					"timestamp": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "Health check timestamp",
					},
					"dependencies": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"docker": map[string]interface{}{
								"type":        "string",
								"description": "Docker daemon health status",
								"enum":        []string{"healthy", "unhealthy", "unavailable"},
								"examples":    []string{"healthy", "unhealthy"},
							},
							"libvirt": map[string]interface{}{
								"type":        "string",
								"description": "Libvirt service health status",
								"enum":        []string{"healthy", "unhealthy", "unavailable"},
								"examples":    []string{"healthy", "unavailable"},
							},
							"storage": map[string]interface{}{
								"type":        "string",
								"description": "Storage system health status",
								"enum":        []string{"healthy", "unhealthy", "unavailable"},
								"examples":    []string{"healthy"},
							},
							"notifications": map[string]interface{}{
								"type":        "string",
								"description": "Notification system health status",
								"enum":        []string{"healthy", "unhealthy", "unavailable"},
								"examples":    []string{"healthy"},
							},
						},
					},
					"metrics": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"uptime_seconds": map[string]interface{}{
								"type":        "integer",
								"description": "System uptime in seconds",
								"example":     86400,
							},
							"memory_usage_percent": map[string]interface{}{
								"type":        "number",
								"description": "Memory usage percentage",
								"example":     45.2,
							},
							"cpu_usage_percent": map[string]interface{}{
								"type":        "number",
								"description": "CPU usage percentage",
								"example":     23.5,
							},
							"api_calls_total": map[string]interface{}{
								"type":        "integer",
								"description": "Total API calls processed",
								"example":     1234,
							},
						},
					},
					"checks": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"response_time_ms": map[string]interface{}{
								"type":        "integer",
								"description": "Health check response time in milliseconds",
								"example":     150,
							},
						},
					},
				},
			},
			"DiskInfo": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"device": map[string]interface{}{
						"type":        "string",
						"description": "Device path",
						"example":     "/dev/sdc",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Disk name",
						"example":     "disk0",
					},
					"role": map[string]interface{}{
						"type":        "string",
						"description": "Disk role in the array",
						"enum":        []string{"array", "parity", "cache", "boot"},
						"examples":    []string{"array", "parity", "cache"},
					},
					"size": map[string]interface{}{
						"type":        "integer",
						"description": "Disk size in bytes",
						"example":     16000900608000,
					},
					"size_formatted": map[string]interface{}{
						"type":        "string",
						"description": "Human-readable disk size",
						"example":     "14.6 TB",
					},
					"status": map[string]interface{}{
						"type":        "string",
						"description": "Disk operational status",
						"enum":        []string{"online", "offline", "disabled", "error"},
						"examples":    []string{"online", "offline"},
					},
					"health": map[string]interface{}{
						"type":        "string",
						"description": "Disk health status",
						"enum":        []string{"healthy", "warning", "critical", "unknown"},
						"examples":    []string{"healthy", "warning"},
					},
					"temperature": map[string]interface{}{
						"type":        "integer",
						"description": "Disk temperature in Celsius",
						"example":     33,
					},
					"smart_data": map[string]interface{}{
						"type":        "object",
						"description": "SMART health data",
						"properties": map[string]interface{}{
							"overall_health": map[string]interface{}{
								"type":        "string",
								"description": "Overall SMART health status",
								"example":     "PASSED",
							},
							"temperature": map[string]interface{}{
								"type":        "integer",
								"description": "SMART temperature reading",
								"example":     33,
							},
							"power_on_hours": map[string]interface{}{
								"type":        "integer",
								"description": "Total power-on hours",
								"example":     37556,
							},
							"attributes": map[string]interface{}{
								"type":        "array",
								"description": "SMART attribute details",
								"items": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"id": map[string]interface{}{
											"type": "integer",
										},
										"name": map[string]interface{}{
											"type": "string",
										},
										"value": map[string]interface{}{
											"type": "integer",
										},
										"raw_value": map[string]interface{}{
											"type": "integer",
										},
									},
								},
							},
						},
					},
				},
			},
			"BulkOperationResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"operation": map[string]interface{}{
						"type":        "string",
						"description": "The bulk operation performed",
						"enum":        []string{"start", "stop", "restart"},
						"examples":    []string{"start", "stop", "restart"},
					},
					"results": map[string]interface{}{
						"type":        "array",
						"description": "Results for each container operation",
						"items": map[string]interface{}{
							"$ref": "#/components/schemas/ContainerOperationResult",
						},
					},
					"summary": map[string]interface{}{
						"$ref": "#/components/schemas/BulkOperationSummary",
					},
				},
			},
			"ContainerOperationResult": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"container_id": map[string]interface{}{
						"type":        "string",
						"description": "Container ID or name",
						"examples":    []string{"plex", "nginx", "sonarr"},
					},
					"container_name": map[string]interface{}{
						"type":        "string",
						"description": "Container name",
						"examples":    []string{"plex", "nginx", "sonarr"},
					},
					"success": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether the operation succeeded",
						"examples":    []interface{}{true, false},
					},
					"error": map[string]interface{}{
						"type":        "string",
						"description": "Error message if operation failed",
						"examples":    []string{"Container not found", "Failed to start container", "Operation timeout"},
					},
					"duration": map[string]interface{}{
						"type":        "string",
						"description": "Operation duration",
						"examples":    []string{"43.658896ms", "337.794446ms", "1.234s"},
					},
				},
			},
			"BulkOperationSummary": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"total": map[string]interface{}{
						"type":        "integer",
						"description": "Total number of containers processed",
						"minimum":     0,
						"examples":    []interface{}{1, 2, 3, 5},
					},
					"succeeded": map[string]interface{}{
						"type":        "integer",
						"description": "Number of successful operations",
						"minimum":     0,
						"examples":    []interface{}{0, 1, 2, 3},
					},
					"failed": map[string]interface{}{
						"type":        "integer",
						"description": "Number of failed operations",
						"minimum":     0,
						"examples":    []interface{}{0, 1, 2},
					},
				},
			},
		},
		Responses: map[string]interface{}{
			"BadRequest": map[string]interface{}{
				"description": "Bad Request",
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"error": map[string]interface{}{
									"type":        "string",
									"description": "Error message",
								},
							},
						},
					},
				},
			},
			"InternalServerError": map[string]interface{}{
				"description": "Internal Server Error",
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"error": map[string]interface{}{
									"type":        "string",
									"description": "Error message",
								},
							},
						},
					},
				},
			},
		},
	}
}

// generatePaths creates the OpenAPI paths specification
func (h *HTTPServer) generatePaths() map[string]interface{} {
	paths := make(map[string]interface{})

	// Health endpoint
	paths["/api/v1/health"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get system health status",
			"description": "Returns comprehensive health information including dependency status, system metrics, and performance data",
			"tags":        []string{"Health"},
			"parameters": []interface{}{
				map[string]interface{}{
					"name":        "X-Request-ID",
					"in":          "header",
					"description": "Optional request ID for tracing",
					"required":    false,
					"schema": map[string]interface{}{
						"type": "string",
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Health status retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/HealthResponse",
							},
						},
					},
				},
				"500": map[string]interface{}{
					"$ref": "#/components/responses/InternalServerError",
				},
			},
		},
	}

	// Docker containers endpoint with pagination
	paths["/api/v1/docker/containers"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "List Docker containers",
			"description": "Returns a list of Docker containers with optional pagination support",
			"tags":        []string{"Docker"},
			"parameters": []interface{}{
				map[string]interface{}{
					"name":        "all",
					"in":          "query",
					"description": "Include stopped containers",
					"required":    false,
					"schema": map[string]interface{}{
						"type":    "boolean",
						"default": false,
					},
				},
				map[string]interface{}{
					"name":        "page",
					"in":          "query",
					"description": "Page number for pagination (enables paginated response)",
					"required":    false,
					"schema": map[string]interface{}{
						"type":    "integer",
						"minimum": 1,
						"default": 1,
					},
				},
				map[string]interface{}{
					"name":        "limit",
					"in":          "query",
					"description": "Number of items per page (enables paginated response)",
					"required":    false,
					"schema": map[string]interface{}{
						"type":    "integer",
						"minimum": 1,
						"maximum": 1000,
						"default": 50,
					},
				},
				map[string]interface{}{
					"name":        "X-Request-ID",
					"in":          "header",
					"description": "Optional request ID for tracing",
					"required":    false,
					"schema": map[string]interface{}{
						"type": "string",
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Containers retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"oneOf": []interface{}{
									map[string]interface{}{
										"description": "Paginated response (when page or limit parameters are provided)",
										"allOf": []interface{}{
											map[string]interface{}{
												"$ref": "#/components/schemas/StandardResponse",
											},
											map[string]interface{}{
												"properties": map[string]interface{}{
													"data": map[string]interface{}{
														"type": "array",
														"items": map[string]interface{}{
															"type":        "object",
															"description": "Docker container information",
														},
													},
												},
											},
										},
									},
									map[string]interface{}{
										"description": "Legacy response format (when no pagination parameters)",
										"type":        "array",
										"items": map[string]interface{}{
											"type":        "object",
											"description": "Docker container information",
										},
									},
								},
							},
						},
					},
				},
				"500": map[string]interface{}{
					"$ref": "#/components/responses/InternalServerError",
				},
			},
		},
	}

	// Storage disks endpoint with consolidated SMART data
	paths["/api/v1/storage/disks"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get consolidated disk information",
			"description": "Returns comprehensive disk information including SMART health data, with optional pagination support",
			"tags":        []string{"Storage"},
			"parameters": []interface{}{
				map[string]interface{}{
					"name":        "page",
					"in":          "query",
					"description": "Page number for pagination (enables paginated response with flattened disk list)",
					"required":    false,
					"schema": map[string]interface{}{
						"type":    "integer",
						"minimum": 1,
						"default": 1,
					},
				},
				map[string]interface{}{
					"name":        "limit",
					"in":          "query",
					"description": "Number of disks per page (enables paginated response)",
					"required":    false,
					"schema": map[string]interface{}{
						"type":    "integer",
						"minimum": 1,
						"maximum": 100,
						"default": 10,
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Disk information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"oneOf": []interface{}{
									map[string]interface{}{
										"description": "Paginated response (when page or limit parameters are provided)",
										"$ref":        "#/components/schemas/StandardResponse",
									},
									map[string]interface{}{
										"description": "Structured response format (when no pagination parameters)",
										"type":        "object",
										"properties": map[string]interface{}{
											"array_disks": map[string]interface{}{
												"type":        "array",
												"description": "Array member disks",
											},
											"cache_disks": map[string]interface{}{
												"type":        "array",
												"description": "Cache pool disks",
											},
											"boot_disk": map[string]interface{}{
												"type":        "object",
												"description": "Boot disk information",
											},
											"summary": map[string]interface{}{
												"type":        "object",
												"description": "Disk summary statistics",
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
	}

	// Docker bulk operations endpoints
	paths["/api/v1/docker/containers/bulk/start"] = map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Start multiple Docker containers",
			"description": "Starts multiple Docker containers in a single bulk operation",
			"tags":        []string{"Docker", "Bulk Operations"},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"container_ids": map[string]interface{}{
									"type": "array",
									"items": map[string]interface{}{
										"type": "string",
									},
									"description": "Array of container IDs or names to start",
									"example":     []string{"plex", "nginx", "sonarr"},
									"minItems":    1,
									"maxItems":    50,
								},
							},
							"required": []string{"container_ids"},
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Bulk operation completed (may include partial failures)",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/BulkOperationResponse",
							},
						},
					},
				},
			},
		},
	}

	// WebSocket endpoints documentation
	paths["/api/v1/ws/system/stats"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Real-time system statistics WebSocket",
			"description": "WebSocket endpoint for real-time system statistics including CPU, memory, and uptime",
			"tags":        []string{"WebSocket", "System"},
			"responses": map[string]interface{}{
				"101": map[string]interface{}{
					"description": "WebSocket connection established",
				},
			},
		},
	}

	paths["/api/v1/ws/docker/events"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Real-time Docker events WebSocket",
			"description": "WebSocket endpoint for real-time Docker container events (start, stop, restart, health changes)",
			"tags":        []string{"WebSocket", "Docker"},
			"responses": map[string]interface{}{
				"101": map[string]interface{}{
					"description": "WebSocket connection established",
				},
			},
		},
	}

	paths["/api/v1/ws/storage/status"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Real-time storage status WebSocket",
			"description": "WebSocket endpoint for real-time storage array status updates (disk health, SMART alerts, array status changes)",
			"tags":        []string{"WebSocket", "Storage"},
			"responses": map[string]interface{}{
				"101": map[string]interface{}{
					"description": "WebSocket connection established",
				},
			},
		},
	}

	return paths
}

// handleOpenAPISpec handles GET /api/v1/openapi.json
func (h *HTTPServer) handleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	spec := h.generateOpenAPISpec()
	h.writeJSON(w, http.StatusOK, spec)
}

// handleSwaggerUI handles GET /api/v1/docs
func (h *HTTPServer) handleSwaggerUI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Generate Swagger UI HTML
	html := h.generateSwaggerUIHTML()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// generateSwaggerUIHTML creates the Swagger UI HTML page
func (h *HTTPServer) generateSwaggerUIHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>UMA API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/api/v1/openapi.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                validatorUrl: null,
                tryItOutEnabled: true,
                supportedSubmitMethods: ['get', 'post', 'put', 'delete', 'patch'],
                onComplete: function() {
                    console.log('UMA API Documentation loaded');
                }
            });
        };
    </script>
</body>
</html>`
}
