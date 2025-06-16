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
	Schemas         map[string]interface{} `json:"schemas"`
	Responses       map[string]interface{} `json:"responses"`
	SecuritySchemes map[string]interface{} `json:"securitySchemes"`
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
			"BulkOperationRequest": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"container_ids": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
						"description": "Array of container IDs or names",
						"example":     []string{"plex", "nginx", "sonarr"},
						"minItems":    1,
						"maxItems":    50,
					},
				},
				"required": []string{"container_ids"},
			},
			"User": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Unique user identifier",
						"example":     "user_1234567890",
					},
					"username": map[string]interface{}{
						"type":        "string",
						"description": "Username",
						"example":     "admin",
					},
					"role": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"admin", "operator", "viewer"},
						"description": "User role with different permission levels",
						"example":     "admin",
					},
					"active": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether the user account is active",
						"example":     true,
					},
					"created_at": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "User creation timestamp",
						"example":     "2025-06-15T21:30:42Z",
					},
					"api_key": map[string]interface{}{
						"type":        "string",
						"description": "User API key (only shown during creation/regeneration)",
						"example":     "uma_1234567890abcdef",
					},
				},
			},
			"VMInfo": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "integer",
						"description": "VM ID (libvirt domain ID)",
						"example":     1,
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "VM name",
						"example":     "Windows10",
					},
					"uuid": map[string]interface{}{
						"type":        "string",
						"description": "VM UUID",
						"example":     "12345678-1234-1234-1234-123456789012",
					},
					"state": map[string]interface{}{
						"type":        "string",
						"description": "VM state",
						"enum":        []string{"running", "shut off", "paused", "suspended", "crashed"},
						"example":     "running",
					},
					"cpus": map[string]interface{}{
						"type":        "integer",
						"description": "Number of virtual CPUs",
						"example":     4,
					},
					"memory_kb": map[string]interface{}{
						"type":        "integer",
						"description": "Memory allocation in KB",
						"example":     8388608,
					},
					"autostart": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether VM starts automatically",
						"example":     true,
					},
					"persistent": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether VM configuration is persistent",
						"example":     true,
					},
					"os_type": map[string]interface{}{
						"type":        "string",
						"description": "Operating system type",
						"example":     "windows",
					},
					"architecture": map[string]interface{}{
						"type":        "string",
						"description": "VM architecture",
						"example":     "x86_64",
					},
				},
			},
			"VMStats": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"cpu_usage": map[string]interface{}{
						"type":        "number",
						"description": "CPU usage percentage",
						"example":     25.5,
					},
					"memory_usage": map[string]interface{}{
						"type":        "integer",
						"description": "Memory usage in bytes",
						"example":     4294967296,
					},
					"memory_available": map[string]interface{}{
						"type":        "integer",
						"description": "Available memory in bytes",
						"example":     8589934592,
					},
					"disk_read_bytes": map[string]interface{}{
						"type":        "integer",
						"description": "Total disk read bytes",
						"example":     1073741824,
					},
					"disk_write_bytes": map[string]interface{}{
						"type":        "integer",
						"description": "Total disk write bytes",
						"example":     536870912,
					},
					"network_rx_bytes": map[string]interface{}{
						"type":        "integer",
						"description": "Network received bytes",
						"example":     268435456,
					},
					"network_tx_bytes": map[string]interface{}{
						"type":        "integer",
						"description": "Network transmitted bytes",
						"example":     134217728,
					},
				},
			},
			"VMOperationResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"message": map[string]interface{}{
						"type":        "string",
						"description": "Operation result message",
						"example":     "VM started successfully",
					},
					"vm_name": map[string]interface{}{
						"type":        "string",
						"description": "VM name",
						"example":     "Windows10",
					},
					"action": map[string]interface{}{
						"type":        "string",
						"description": "Action performed",
						"enum":        []string{"start", "stop", "restart", "pause", "resume", "hibernate", "restore", "autostart"},
						"example":     "start",
					},
				},
			},
			"TemperatureData": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"sensors": map[string]interface{}{
						"type":        "object",
						"description": "Temperature sensors by chip",
						"additionalProperties": map[string]interface{}{
							"$ref": "#/components/schemas/SensorChip",
						},
					},
					"fans": map[string]interface{}{
						"type":        "object",
						"description": "Fan sensors",
						"additionalProperties": map[string]interface{}{
							"$ref": "#/components/schemas/FanInput",
						},
					},
				},
			},
			"SensorChip": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Chip name",
						"example":     "coretemp-isa-0000",
					},
					"temperatures": map[string]interface{}{
						"type":        "object",
						"description": "Temperature inputs",
						"additionalProperties": map[string]interface{}{
							"type":        "number",
							"description": "Temperature in Celsius",
							"example":     45.0,
						},
					},
				},
			},
			"FanInput": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"speed": map[string]interface{}{
						"type":        "integer",
						"description": "Fan speed in RPM",
						"example":     1200,
					},
					"label": map[string]interface{}{
						"type":        "string",
						"description": "Fan label",
						"example":     "CPU Fan",
					},
				},
			},
			"NetworkInfo": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"interface": map[string]interface{}{
						"type":        "string",
						"description": "Network interface name",
						"example":     "eth0",
					},
					"bytes_received": map[string]interface{}{
						"type":        "integer",
						"description": "Total bytes received",
						"example":     1073741824,
					},
					"bytes_sent": map[string]interface{}{
						"type":        "integer",
						"description": "Total bytes sent",
						"example":     536870912,
					},
					"packets_received": map[string]interface{}{
						"type":        "integer",
						"description": "Total packets received",
						"example":     1000000,
					},
					"packets_sent": map[string]interface{}{
						"type":        "integer",
						"description": "Total packets sent",
						"example":     500000,
					},
					"errors_received": map[string]interface{}{
						"type":        "integer",
						"description": "Receive errors",
						"example":     0,
					},
					"errors_sent": map[string]interface{}{
						"type":        "integer",
						"description": "Send errors",
						"example":     0,
					},
				},
			},
			"UPSStatus": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"status": map[string]interface{}{
						"type":        "string",
						"description": "UPS status",
						"enum":        []string{"Online", "On Battery", "Low Battery", "Charging"},
						"example":     "Online",
					},
					"charge": map[string]interface{}{
						"type":        "number",
						"description": "Battery charge percentage",
						"example":     95.0,
					},
					"runtime": map[string]interface{}{
						"type":        "number",
						"description": "Estimated runtime in minutes",
						"example":     45.5,
					},
					"load": map[string]interface{}{
						"type":        "number",
						"description": "Load percentage",
						"example":     25.0,
					},
					"power": map[string]interface{}{
						"type":        "number",
						"description": "Power consumption in watts",
						"example":     150.0,
					},
				},
			},
			"GPUInfo": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"index": map[string]interface{}{
						"type":        "integer",
						"description": "GPU index",
						"example":     0,
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "GPU name",
						"example":     "NVIDIA GeForce RTX 3080",
					},
					"driver": map[string]interface{}{
						"type":        "string",
						"description": "GPU driver",
						"enum":        []string{"nvidia", "amdgpu", "intel"},
						"example":     "nvidia",
					},
					"temperature": map[string]interface{}{
						"type":        "integer",
						"description": "GPU temperature in Celsius",
						"example":     65,
					},
					"utilization": map[string]interface{}{
						"type":        "integer",
						"description": "GPU utilization percentage",
						"example":     75,
					},
					"memory_total": map[string]interface{}{
						"type":        "integer",
						"description": "Total GPU memory in MB",
						"example":     10240,
					},
					"memory_used": map[string]interface{}{
						"type":        "integer",
						"description": "Used GPU memory in MB",
						"example":     5120,
					},
					"power_draw": map[string]interface{}{
						"type":        "number",
						"description": "Power draw in watts",
						"example":     220.5,
					},
					"fan_speed": map[string]interface{}{
						"type":        "integer",
						"description": "Fan speed percentage",
						"example":     60,
					},
				},
			},
			"ArrayInfo": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"status": map[string]interface{}{
						"type":        "string",
						"description": "Array status",
						"enum":        []string{"Started", "Stopped", "Starting", "Stopping"},
						"example":     "Started",
					},
					"total_size": map[string]interface{}{
						"type":        "integer",
						"description": "Total array size in bytes",
						"example":     10995116277760,
					},
					"used_size": map[string]interface{}{
						"type":        "integer",
						"description": "Used array size in bytes",
						"example":     5497558138880,
					},
					"free_size": map[string]interface{}{
						"type":        "integer",
						"description": "Free array size in bytes",
						"example":     5497558138880,
					},
					"usage_percent": map[string]interface{}{
						"type":        "number",
						"description": "Array usage percentage",
						"example":     50.0,
					},
					"parity_status": map[string]interface{}{
						"type":        "string",
						"description": "Parity check status",
						"enum":        []string{"Valid", "Invalid", "Checking", "Unknown"},
						"example":     "Valid",
					},
					"last_parity_check": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "Last parity check timestamp",
						"example":     "2025-06-15T12:00:00Z",
					},
					"disks": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"$ref": "#/components/schemas/DiskInfo",
						},
					},
				},
			},
			"CacheInfo": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"status": map[string]interface{}{
						"type":        "string",
						"description": "Cache pool status",
						"example":     "Online",
					},
					"total_size": map[string]interface{}{
						"type":        "integer",
						"description": "Total cache size in bytes",
						"example":     1099511627776,
					},
					"used_size": map[string]interface{}{
						"type":        "integer",
						"description": "Used cache size in bytes",
						"example":     549755813888,
					},
					"usage_percent": map[string]interface{}{
						"type":        "number",
						"description": "Cache usage percentage",
						"example":     50.0,
					},
					"disks": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"$ref": "#/components/schemas/DiskInfo",
						},
					},
				},
			},
			"BootInfo": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"device": map[string]interface{}{
						"type":        "string",
						"description": "Boot device path",
						"example":     "/dev/sda1",
					},
					"filesystem": map[string]interface{}{
						"type":        "string",
						"description": "Filesystem type",
						"example":     "vfat",
					},
					"total_size": map[string]interface{}{
						"type":        "integer",
						"description": "Total boot disk size in bytes",
						"example":     536870912,
					},
					"used_size": map[string]interface{}{
						"type":        "integer",
						"description": "Used boot disk size in bytes",
						"example":     268435456,
					},
					"usage_percent": map[string]interface{}{
						"type":        "number",
						"description": "Boot disk usage percentage",
						"example":     50.0,
					},
					"health": map[string]interface{}{
						"type":        "string",
						"description": "Boot disk health status",
						"enum":        []string{"Healthy", "Warning", "Critical"},
						"example":     "Healthy",
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
		SecuritySchemes: map[string]interface{}{
			"BearerAuth": map[string]interface{}{
				"type":         "http",
				"scheme":       "bearer",
				"bearerFormat": "JWT",
				"description":  "JWT token obtained from /api/v1/auth/login",
			},
			"ApiKeyAuth": map[string]interface{}{
				"type":        "apiKey",
				"in":          "header",
				"name":        "X-API-Key",
				"description": "API key for authentication",
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

	paths["/api/v1/docker/containers/bulk/stop"] = map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Stop multiple Docker containers",
			"description": "Stops multiple Docker containers in a single bulk operation",
			"tags":        []string{"Docker", "Bulk Operations"},
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

	paths["/api/v1/docker/containers/bulk/restart"] = map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Restart multiple Docker containers",
			"description": "Restarts multiple Docker containers in a single bulk operation",
			"tags":        []string{"Docker", "Bulk Operations"},
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

	// Authentication endpoints (Phase 3)
	paths["/api/v1/auth/login"] = map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Authenticate with API key",
			"description": "Authenticate using API key and receive JWT token for subsequent requests",
			"tags":        []string{"Authentication"},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"api_key": map[string]interface{}{
									"type":        "string",
									"description": "User API key",
									"example":     "uma_1234567890abcdef",
								},
							},
							"required": []string{"api_key"},
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Authentication successful",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"token": map[string]interface{}{
										"type":        "string",
										"description": "JWT token for API access",
									},
									"user": map[string]interface{}{
										"$ref": "#/components/schemas/User",
									},
								},
							},
						},
					},
				},
				"401": map[string]interface{}{
					"description": "Invalid API key",
				},
				"501": map[string]interface{}{
					"description": "Authentication is disabled",
				},
			},
		},
	}

	paths["/api/v1/auth/users"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "List all users",
			"description": "Get a list of all users (Admin only)",
			"tags":        []string{"Authentication"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Users retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"$ref": "#/components/schemas/User",
								},
							},
						},
					},
				},
				"403": map[string]interface{}{
					"description": "Insufficient permissions (Admin required)",
				},
				"501": map[string]interface{}{
					"description": "Authentication is disabled",
				},
			},
		},
		"post": map[string]interface{}{
			"summary":     "Create new user",
			"description": "Create a new user with specified role (Admin only)",
			"tags":        []string{"Authentication"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"username": map[string]interface{}{
									"type":        "string",
									"description": "Username for the new user",
									"example":     "operator1",
								},
								"role": map[string]interface{}{
									"type":        "string",
									"enum":        []string{"admin", "operator", "viewer"},
									"description": "User role",
									"example":     "operator",
								},
							},
							"required": []string{"username", "role"},
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"201": map[string]interface{}{
					"description": "User created successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/User",
							},
						},
					},
				},
				"400": map[string]interface{}{
					"description": "Invalid request data",
				},
				"403": map[string]interface{}{
					"description": "Insufficient permissions (Admin required)",
				},
			},
		},
	}

	paths["/api/v1/auth/stats"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get authentication statistics",
			"description": "Get authentication system statistics and status",
			"tags":        []string{"Authentication"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Authentication statistics retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"enabled": map[string]interface{}{
										"type":        "boolean",
										"description": "Whether authentication is enabled",
									},
									"total_users": map[string]interface{}{
										"type":        "integer",
										"description": "Total number of users",
									},
									"active_users": map[string]interface{}{
										"type":        "integer",
										"description": "Number of active users",
									},
									"roles": map[string]interface{}{
										"type":        "object",
										"description": "User count by role",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Virtual Machine endpoints
	paths["/api/v1/vms"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "List virtual machines",
			"description": "Get a list of all virtual machines with optional filtering",
			"tags":        []string{"Virtual Machines"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"parameters": []interface{}{
				map[string]interface{}{
					"name":        "all",
					"in":          "query",
					"description": "Include inactive VMs",
					"required":    false,
					"schema": map[string]interface{}{
						"type":    "boolean",
						"default": false,
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "VMs retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/StandardResponse",
							},
						},
					},
				},
				"403": map[string]interface{}{
					"description": "Insufficient permissions",
				},
				"500": map[string]interface{}{
					"description": "Libvirt not available or internal error",
				},
			},
		},
	}

	paths["/api/v1/vms/{name}"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get virtual machine details",
			"description": "Get detailed information about a specific virtual machine",
			"tags":        []string{"Virtual Machines"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"parameters": []interface{}{
				map[string]interface{}{
					"name":        "name",
					"in":          "path",
					"description": "VM name",
					"required":    true,
					"schema": map[string]interface{}{
						"type": "string",
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "VM details retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/VMInfo",
							},
						},
					},
				},
				"404": map[string]interface{}{
					"description": "VM not found",
				},
			},
		},
	}

	paths["/api/v1/vms/{name}/stats"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get VM performance statistics",
			"description": "Get real-time performance statistics for a virtual machine",
			"tags":        []string{"Virtual Machines"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"parameters": []interface{}{
				map[string]interface{}{
					"name":        "name",
					"in":          "path",
					"description": "VM name",
					"required":    true,
					"schema": map[string]interface{}{
						"type": "string",
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "VM statistics retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/VMStats",
							},
						},
					},
				},
			},
		},
	}

	paths["/api/v1/vms/{name}/start"] = map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Start virtual machine",
			"description": "Start a virtual machine",
			"tags":        []string{"Virtual Machines"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"parameters": []interface{}{
				map[string]interface{}{
					"name":        "name",
					"in":          "path",
					"description": "VM name",
					"required":    true,
					"schema": map[string]interface{}{
						"type": "string",
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "VM started successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/VMOperationResponse",
							},
						},
					},
				},
				"500": map[string]interface{}{
					"description": "Failed to start VM",
				},
			},
		},
	}

	paths["/api/v1/vms/{name}/stop"] = map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Stop virtual machine",
			"description": "Stop a virtual machine gracefully or forcefully",
			"tags":        []string{"Virtual Machines"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"parameters": []interface{}{
				map[string]interface{}{
					"name":        "name",
					"in":          "path",
					"description": "VM name",
					"required":    true,
					"schema": map[string]interface{}{
						"type": "string",
					},
				},
				map[string]interface{}{
					"name":        "force",
					"in":          "query",
					"description": "Force stop the VM (destroy instead of shutdown)",
					"required":    false,
					"schema": map[string]interface{}{
						"type":    "boolean",
						"default": false,
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "VM stopped successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/VMOperationResponse",
							},
						},
					},
				},
			},
		},
	}

	// Add missing VM control endpoints
	paths["/api/v1/vms/{name}/restart"] = map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Restart virtual machine",
			"description": "Restart a virtual machine",
			"tags":        []string{"Virtual Machines"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"parameters": []interface{}{
				map[string]interface{}{
					"name":        "name",
					"in":          "path",
					"description": "VM name",
					"required":    true,
					"schema": map[string]interface{}{
						"type": "string",
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "VM restarted successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/VMOperationResponse",
							},
						},
					},
				},
			},
		},
	}

	paths["/api/v1/vms/{name}/pause"] = map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Pause virtual machine",
			"description": "Pause a virtual machine",
			"tags":        []string{"Virtual Machines"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"parameters": []interface{}{
				map[string]interface{}{
					"name":        "name",
					"in":          "path",
					"description": "VM name",
					"required":    true,
					"schema": map[string]interface{}{
						"type": "string",
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "VM paused successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/VMOperationResponse",
							},
						},
					},
				},
			},
		},
	}

	paths["/api/v1/vms/{name}/resume"] = map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Resume virtual machine",
			"description": "Resume a paused virtual machine",
			"tags":        []string{"Virtual Machines"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"parameters": []interface{}{
				map[string]interface{}{
					"name":        "name",
					"in":          "path",
					"description": "VM name",
					"required":    true,
					"schema": map[string]interface{}{
						"type": "string",
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "VM resumed successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/VMOperationResponse",
							},
						},
					},
				},
			},
		},
	}

	paths["/api/v1/vms/{name}/hibernate"] = map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Hibernate virtual machine",
			"description": "Hibernate a virtual machine to disk",
			"tags":        []string{"Virtual Machines"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"parameters": []interface{}{
				map[string]interface{}{
					"name":        "name",
					"in":          "path",
					"description": "VM name",
					"required":    true,
					"schema": map[string]interface{}{
						"type": "string",
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "VM hibernated successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/VMOperationResponse",
							},
						},
					},
				},
			},
		},
	}

	// Enhanced System Monitoring endpoints
	paths["/api/v1/system/temperature"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get system temperature sensors",
			"description": "Get comprehensive temperature and fan monitoring data from all system sensors",
			"tags":        []string{"System", "Monitoring"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Temperature data retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/TemperatureData",
							},
						},
					},
				},
			},
		},
	}

	paths["/api/v1/system/network"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get network interface information",
			"description": "Get detailed network interface statistics and information",
			"tags":        []string{"System", "Monitoring"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Network information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"$ref": "#/components/schemas/NetworkInfo",
								},
							},
						},
					},
				},
			},
		},
	}

	paths["/api/v1/system/ups"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get UPS status and metrics",
			"description": "Get UPS power status, battery charge, load percentage, and power consumption",
			"tags":        []string{"System", "Monitoring"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "UPS status retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/UPSStatus",
							},
						},
					},
				},
				"404": map[string]interface{}{
					"description": "UPS not found or not configured",
				},
			},
		},
	}

	paths["/api/v1/system/gpu"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get GPU monitoring data",
			"description": "Get GPU usage, temperature, and status for Intel, Nvidia, and AMD graphics cards",
			"tags":        []string{"System", "Monitoring"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "GPU information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"$ref": "#/components/schemas/GPUInfo",
								},
							},
						},
					},
				},
			},
		},
	}

	paths["/api/v1/storage/array"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get array disk information",
			"description": "Get comprehensive array disk health, usage, and SMART data",
			"tags":        []string{"Storage", "Monitoring"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Array information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/ArrayInfo",
							},
						},
					},
				},
			},
		},
	}

	paths["/api/v1/storage/cache"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get cache disk status",
			"description": "Get cache disk monitoring, usage, and health status",
			"tags":        []string{"Storage", "Monitoring"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Cache disk information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/CacheInfo",
							},
						},
					},
				},
			},
		},
	}

	paths["/api/v1/storage/boot"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get boot disk information",
			"description": "Get boot disk health, usage, and filesystem information",
			"tags":        []string{"Storage", "Monitoring"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Boot disk information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/BootInfo",
							},
						},
					},
				},
			},
		},
	}

	// Metrics endpoint
	paths["/metrics"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Prometheus metrics",
			"description": "Get Prometheus-formatted metrics for monitoring and alerting",
			"tags":        []string{"Monitoring"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Metrics retrieved successfully",
					"content": map[string]interface{}{
						"text/plain": map[string]interface{}{
							"schema": map[string]interface{}{
								"type":        "string",
								"description": "Prometheus metrics format",
								"example":     "# HELP uma_api_requests_total Total API requests\n# TYPE uma_api_requests_total counter\numa_api_requests_total{endpoint=\"/api/v1/health\",method=\"GET\",status_code=\"200\"} 42",
							},
						},
					},
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
