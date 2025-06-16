package api

import (
	"fmt"
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
	// Get version with fallback
	version := h.api.ctx.Config.Version
	if version == "" || version == "unknown" {
		version = "2025.06.16" // Current plugin version
	}

	return &OpenAPISpec{
		OpenAPI: "3.1.1",
		Info: OpenAPIInfo{
			Title:       "UMA REST API",
			Description: "Unraid Management Agent API providing 100% functionality coverage for comprehensive server management. Features include system monitoring (CPU, RAM, temperatures, fans), storage management (array, disks, cache, SMART data), Docker container control (individual and bulk operations), VM lifecycle management, UPS monitoring with real hardware integration, system control (scripts, logs, power management), and WebSocket real-time updates. Built with optimized HTTP mux architecture for production deployment.",
			Version:     version,
			Contact: OpenAPIContact{
				Name:  "UMA Development Team",
				URL:   "https://github.com/domalab/uma",
				Email: "support@domalab.net",
			},
		},
		Servers: []OpenAPIServer{
			{
				URL:         fmt.Sprintf("http://localhost:%d", h.api.ctx.Config.HTTPServer.Port),
				Description: "Local UMA API server",
			},
			{
				URL:         "http://your-unraid-server:34600",
				Description: "Remote UMA API server (replace with your server IP)",
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
			"EnhancedUPSStatus": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"status": map[string]interface{}{
						"type":        "string",
						"description": "UPS status",
						"enum":        []string{"online", "offline", "on_battery", "low_battery", "unknown"},
						"example":     "online",
					},
					"battery_charge": map[string]interface{}{
						"type":        "number",
						"description": "Battery charge percentage",
						"minimum":     0,
						"maximum":     100,
						"example":     100.0,
					},
					"runtime": map[string]interface{}{
						"type":        "number",
						"description": "Estimated runtime in minutes",
						"minimum":     0,
						"example":     220.0,
					},
					"load_percent": map[string]interface{}{
						"type":        "number",
						"description": "Load percentage",
						"minimum":     0,
						"maximum":     100,
						"example":     0.0,
					},
					"input_voltage": map[string]interface{}{
						"type":        "number",
						"description": "Input voltage in volts",
						"minimum":     0,
						"example":     246.0,
					},
					"output_voltage": map[string]interface{}{
						"type":        "number",
						"description": "Output voltage in volts",
						"minimum":     0,
						"example":     246.0,
					},
					"model": map[string]interface{}{
						"type":        "string",
						"description": "UPS model name",
						"example":     "Back-UPS XS 950U",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "UPS configured name",
						"example":     "Cube",
					},
					"serial_number": map[string]interface{}{
						"type":        "string",
						"description": "UPS serial number",
						"example":     "4B1920P16814",
					},
					"nominal_power": map[string]interface{}{
						"type":        "number",
						"description": "UPS nominal power rating in watts",
						"minimum":     0,
						"example":     480.0,
					},
					"connected": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether UPS hardware is connected and detected",
						"example":     true,
					},
					"ups_type": map[string]interface{}{
						"type":        "string",
						"description": "UPS system type",
						"enum":        []string{"apc", "nut", "unknown"},
						"example":     "apc",
					},
				},
				"required": []string{"status", "battery_charge", "runtime", "connected"},
			},
			"ContainerOperationResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
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
					"timestamp": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "Operation timestamp",
						"example":     "2025-06-16T14:30:00Z",
					},
				},
				"required": []string{"message", "container_id", "timestamp"},
			},
			"UserScript": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Script name",
						"example":     "backup-script",
					},
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Script file path",
						"example":     "/boot/config/plugins/user.scripts/scripts/backup-script/script",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Script description",
						"example":     "Daily backup script",
					},
					"executable": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether script is executable",
						"example":     true,
					},
					"last_run": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "Last execution timestamp",
						"example":     "2025-06-16T02:00:00Z",
					},
				},
				"required": []string{"name", "path", "executable"},
			},
			"ScriptExecutionRequest": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"script_name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the script to execute",
						"example":     "backup-script",
					},
					"parameters": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
						"description": "Optional script parameters",
						"example":     []string{"--full", "--compress"},
					},
					"background": map[string]interface{}{
						"type":        "boolean",
						"description": "Execute script in background",
						"default":     false,
						"example":     true,
					},
				},
				"required": []string{"script_name"},
			},
			"ScriptExecutionResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"message": map[string]interface{}{
						"type":        "string",
						"description": "Execution result message",
						"example":     "Script executed successfully",
					},
					"script_name": map[string]interface{}{
						"type":        "string",
						"description": "Name of executed script",
						"example":     "backup-script",
					},
					"exit_code": map[string]interface{}{
						"type":        "integer",
						"description": "Script exit code (if not background)",
						"example":     0,
					},
					"output": map[string]interface{}{
						"type":        "string",
						"description": "Script output (if not background)",
						"example":     "Backup completed successfully",
					},
					"background": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether script was executed in background",
						"example":     false,
					},
					"timestamp": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "Execution timestamp",
						"example":     "2025-06-16T14:30:00Z",
					},
				},
				"required": []string{"message", "script_name", "background", "timestamp"},
			},
			"SystemPowerRequest": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"delay": map[string]interface{}{
						"type":        "integer",
						"description": "Delay in seconds before executing power operation",
						"minimum":     0,
						"maximum":     3600,
						"default":     0,
						"example":     30,
					},
					"message": map[string]interface{}{
						"type":        "string",
						"description": "Custom message to display during power operation",
						"maxLength":   200,
						"example":     "System maintenance reboot",
					},
				},
			},
			"SystemPowerResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"message": map[string]interface{}{
						"type":        "string",
						"description": "Power operation result message",
						"example":     "System reboot initiated successfully",
					},
					"operation": map[string]interface{}{
						"type":        "string",
						"description": "Power operation type",
						"enum":        []string{"reboot", "shutdown"},
						"example":     "reboot",
					},
					"delay": map[string]interface{}{
						"type":        "integer",
						"description": "Delay in seconds before execution",
						"example":     30,
					},
					"timestamp": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "Operation initiation timestamp",
						"example":     "2025-06-16T14:30:00Z",
					},
					"scheduled_time": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "Scheduled execution time",
						"example":     "2025-06-16T14:30:30Z",
					},
				},
				"required": []string{"message", "operation", "delay", "timestamp"},
			},
			"SystemLogsResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"log_type": map[string]interface{}{
						"type":        "string",
						"description": "Type of logs retrieved",
						"enum":        []string{"system", "kernel", "docker", "nginx", "unraid"},
						"example":     "system",
					},
					"lines_requested": map[string]interface{}{
						"type":        "integer",
						"description": "Number of lines requested",
						"example":     100,
					},
					"lines_returned": map[string]interface{}{
						"type":        "integer",
						"description": "Actual number of lines returned",
						"example":     95,
					},
					"since": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "Timestamp filter applied",
						"example":     "2025-06-16T12:00:00Z",
					},
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Custom log file path (if specified)",
						"example":     "/var/log/custom.log",
					},
					"grep_filter": map[string]interface{}{
						"type":        "string",
						"description": "Applied grep filter (if specified)",
						"example":     "error",
					},
					"logs": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"timestamp": map[string]interface{}{
									"type":        "string",
									"format":      "date-time",
									"description": "Log entry timestamp",
									"example":     "2025-06-16T14:30:00Z",
								},
								"level": map[string]interface{}{
									"type":        "string",
									"description": "Log level",
									"enum":        []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"},
									"example":     "INFO",
								},
								"source": map[string]interface{}{
									"type":        "string",
									"description": "Log source/component",
									"example":     "kernel",
								},
								"message": map[string]interface{}{
									"type":        "string",
									"description": "Log message content",
									"example":     "System startup completed",
								},
							},
							"required": []string{"timestamp", "message"},
						},
						"description": "Array of log entries",
					},
					"timestamp": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "Response generation timestamp",
						"example":     "2025-06-16T14:30:00Z",
					},
				},
				"required": []string{"log_type", "lines_requested", "lines_returned", "logs", "timestamp"},
			},
			"ArrayStartRequest": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"maintenance_mode": map[string]interface{}{
						"type":        "boolean",
						"description": "Start array in maintenance mode",
						"default":     false,
						"example":     false,
					},
					"check_filesystem": map[string]interface{}{
						"type":        "boolean",
						"description": "Perform filesystem check during start",
						"default":     false,
						"example":     false,
					},
				},
			},
			"ArrayStopRequest": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"force": map[string]interface{}{
						"type":        "boolean",
						"description": "Force array stop even if dependencies fail",
						"default":     false,
						"example":     false,
					},
					"unmount_shares": map[string]interface{}{
						"type":        "boolean",
						"description": "Unmount user shares before stopping array",
						"default":     true,
						"example":     true,
					},
					"stop_containers": map[string]interface{}{
						"type":        "boolean",
						"description": "Stop all Docker containers before stopping array",
						"default":     true,
						"example":     true,
					},
					"stop_vms": map[string]interface{}{
						"type":        "boolean",
						"description": "Stop all virtual machines before stopping array",
						"default":     true,
						"example":     true,
					},
				},
			},
			"ArrayOperationResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"message": map[string]interface{}{
						"type":        "string",
						"description": "Operation result message",
						"example":     "Array started successfully with orchestration",
					},
					"operation": map[string]interface{}{
						"type":        "string",
						"description": "Array operation type",
						"enum":        []string{"start", "stop"},
						"example":     "start",
					},
					"orchestration_steps": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"step": map[string]interface{}{
									"type":        "string",
									"description": "Orchestration step name",
									"example":     "Disk detection and validation",
								},
								"status": map[string]interface{}{
									"type":        "string",
									"description": "Step completion status",
									"enum":        []string{"completed", "failed", "skipped"},
									"example":     "completed",
								},
								"duration_ms": map[string]interface{}{
									"type":        "integer",
									"description": "Step duration in milliseconds",
									"example":     1250,
								},
							},
							"required": []string{"step", "status"},
						},
						"description": "Array of orchestration steps performed",
					},
					"array_state": map[string]interface{}{
						"type":        "string",
						"description": "Final array state after operation",
						"enum":        []string{"started", "stopped", "invalid", "unknown"},
						"example":     "started",
					},
					"timestamp": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "Operation completion timestamp",
						"example":     "2025-06-16T14:30:00Z",
					},
				},
				"required": []string{"message", "operation", "array_state", "timestamp"},
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
			"CommandExecuteRequest": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{
						"type":        "string",
						"description": "Command to execute",
						"example":     "ls -la /var/log",
					},
					"timeout": map[string]interface{}{
						"type":        "integer",
						"description": "Timeout in seconds (default: 30, max: 300)",
						"minimum":     1,
						"maximum":     300,
						"default":     30,
						"example":     60,
					},
					"working_directory": map[string]interface{}{
						"type":        "string",
						"description": "Optional working directory for command execution",
						"example":     "/tmp",
					},
				},
				"required": []string{"command"},
			},
			"CommandExecuteResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"exit_code": map[string]interface{}{
						"type":        "integer",
						"description": "Command exit code (0 = success)",
						"example":     0,
					},
					"stdout": map[string]interface{}{
						"type":        "string",
						"description": "Command standard output",
						"example":     "total 48\ndrwxr-xr-x 8 root root 4096 Jun 16 14:30 .\n",
					},
					"stderr": map[string]interface{}{
						"type":        "string",
						"description": "Command standard error output",
						"example":     "",
					},
					"execution_time_ms": map[string]interface{}{
						"type":        "integer",
						"description": "Command execution time in milliseconds",
						"example":     125,
					},
					"command": map[string]interface{}{
						"type":        "string",
						"description": "The executed command",
						"example":     "ls -la /var/log",
					},
					"working_directory": map[string]interface{}{
						"type":        "string",
						"description": "Working directory used for execution",
						"example":     "/tmp",
					},
				},
				"required": []string{"exit_code", "stdout", "stderr", "execution_time_ms", "command"},
			},
			"LogFilesResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"directory": map[string]interface{}{
						"type":        "string",
						"description": "Scanned directory path",
						"example":     "/var/log",
					},
					"recursive": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether subdirectories were included",
						"example":     true,
					},
					"file_pattern": map[string]interface{}{
						"type":        "string",
						"description": "File pattern filter applied",
						"example":     "*.log",
					},
					"max_files": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum files limit",
						"example":     50,
					},
					"total_found": map[string]interface{}{
						"type":        "integer",
						"description": "Total number of log files found",
						"example":     23,
					},
					"files": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"$ref": "#/components/schemas/LogFileInfo",
						},
						"description": "Array of log file metadata",
					},
				},
				"required": []string{"directory", "recursive", "max_files", "total_found", "files"},
			},
			"LogFileInfo": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Full file path",
						"example":     "/var/log/syslog",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "File name",
						"example":     "syslog",
					},
					"size": map[string]interface{}{
						"type":        "integer",
						"description": "File size in bytes",
						"example":     1048576,
					},
					"modified_time": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "Last modification time",
						"example":     "2025-06-16T14:30:00Z",
					},
					"readable": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether the file is readable",
						"example":     true,
					},
				},
				"required": []string{"path", "name", "size", "modified_time", "readable"},
			},
			"CPUResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"usage_percent": map[string]interface{}{
						"type":        "number",
						"description": "CPU usage percentage",
						"example":     5.12,
					},
					"cores": map[string]interface{}{
						"type":        "integer",
						"description": "Number of CPU cores",
						"example":     8,
					},
					"threads": map[string]interface{}{
						"type":        "integer",
						"description": "Number of CPU threads",
						"example":     16,
					},
					"threads_per_core": map[string]interface{}{
						"type":        "integer",
						"description": "Threads per core",
						"example":     2,
					},
					"sockets": map[string]interface{}{
						"type":        "integer",
						"description": "Number of CPU sockets",
						"example":     1,
					},
					"model": map[string]interface{}{
						"type":        "string",
						"description": "CPU model name",
						"example":     "Intel(R) Core(TM) i7-8700K CPU @ 3.70GHz",
					},
					"frequency_mhz": map[string]interface{}{
						"type":        "number",
						"description": "Current CPU frequency in MHz",
						"example":     3700.0,
					},
					"max_frequency_mhz": map[string]interface{}{
						"type":        "number",
						"description": "Maximum CPU frequency in MHz",
						"example":     4700.0,
					},
					"min_frequency_mhz": map[string]interface{}{
						"type":        "number",
						"description": "Minimum CPU frequency in MHz",
						"example":     800.0,
					},
					"temperature": map[string]interface{}{
						"type":        "number",
						"description": "CPU temperature in Celsius",
						"example":     45.5,
					},
					"load_1min": map[string]interface{}{
						"type":        "number",
						"description": "1-minute load average",
						"example":     0.25,
					},
					"load_5min": map[string]interface{}{
						"type":        "number",
						"description": "5-minute load average",
						"example":     0.30,
					},
					"load_15min": map[string]interface{}{
						"type":        "number",
						"description": "15-minute load average",
						"example":     0.28,
					},
					"last_updated": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "Last update timestamp",
						"example":     "2025-06-16T05:56:51Z",
					},
				},
				"required": []string{"usage_percent", "cores", "model", "last_updated"},
			},
			"MemoryResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"usage_percent": map[string]interface{}{
						"type":        "number",
						"description": "Memory usage percentage",
						"example":     16.39,
					},
					"total_bytes": map[string]interface{}{
						"type":        "integer",
						"description": "Total memory in bytes",
						"example":     17179869184,
					},
					"used_bytes": map[string]interface{}{
						"type":        "integer",
						"description": "Used memory in bytes",
						"example":     2814377984,
					},
					"free_bytes": map[string]interface{}{
						"type":        "integer",
						"description": "Free memory in bytes",
						"example":     14365491200,
					},
					"available_bytes": map[string]interface{}{
						"type":        "integer",
						"description": "Available memory in bytes",
						"example":     14365491200,
					},
					"buffers_bytes": map[string]interface{}{
						"type":        "integer",
						"description": "Buffer memory in bytes",
						"example":     134217728,
					},
					"cached_bytes": map[string]interface{}{
						"type":        "integer",
						"description": "Cached memory in bytes",
						"example":     1073741824,
					},
					"total_formatted": map[string]interface{}{
						"type":        "string",
						"description": "Total memory formatted",
						"example":     "16.0 GB",
					},
					"used_formatted": map[string]interface{}{
						"type":        "string",
						"description": "Used memory formatted",
						"example":     "2.6 GB",
					},
					"breakdown": map[string]interface{}{
						"type":        "object",
						"description": "Memory breakdown by category",
						"properties": map[string]interface{}{
							"system_bytes": map[string]interface{}{
								"type": "integer",
							},
							"vm_bytes": map[string]interface{}{
								"type": "integer",
							},
							"docker_bytes": map[string]interface{}{
								"type": "integer",
							},
							"zfs_cache_bytes": map[string]interface{}{
								"type": "integer",
							},
						},
					},
					"last_updated": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "Last update timestamp",
						"example":     "2025-06-16T05:57:05Z",
					},
				},
				"required": []string{"usage_percent", "total_bytes", "used_bytes", "last_updated"},
			},
			"FansResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"fans": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"label": map[string]interface{}{
									"type":        "string",
									"description": "Fan label/name",
									"example":     "CPU Fan",
								},
								"speed_rpm": map[string]interface{}{
									"type":        "integer",
									"description": "Fan speed in RPM",
									"example":     1200,
								},
								"speed_percent": map[string]interface{}{
									"type":        "number",
									"description": "Fan speed percentage",
									"example":     45.5,
								},
								"status": map[string]interface{}{
									"type":        "string",
									"description": "Fan status",
									"example":     "OK",
								},
							},
						},
					},
					"last_updated": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "Last update timestamp",
						"example":     "2025-06-16T05:58:00Z",
					},
				},
				"required": []string{"fans", "last_updated"},
			},
			"ParityDiskResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"device": map[string]interface{}{
						"type":        "string",
						"description": "Device path",
						"example":     "/dev/sdc",
					},
					"serial_number": map[string]interface{}{
						"type":        "string",
						"description": "Disk serial number",
						"example":     "WUH721816ALE6L4_2CGV0URP",
					},
					"capacity": map[string]interface{}{
						"type":        "string",
						"description": "Disk capacity from SMART data",
						"example":     "16.0 TB",
					},
					"temperature": map[string]interface{}{
						"type":        "string",
						"description": "Disk temperature from SMART sensors",
						"example":     "35°C",
					},
					"smart_status": map[string]interface{}{
						"type":        "string",
						"description": "SMART health status (PASSED/FAILED)",
						"example":     "PASSED",
					},
					"power_state": map[string]interface{}{
						"type":        "string",
						"description": "Current power state (Active/Standby)",
						"example":     "Active",
					},
					"spin_down_delay": map[string]interface{}{
						"type":        "string",
						"description": "Configured spin down delay",
						"example":     "Never",
					},
					"health_assessment": map[string]interface{}{
						"type":        "string",
						"description": "Overall health assessment based on SMART data",
						"example":     "Healthy",
					},
					"state": map[string]interface{}{
						"type":        "integer",
						"description": "Unraid disk state code (7=active, 0=missing)",
						"example":     7,
					},
					"device_name": map[string]interface{}{
						"type":        "string",
						"description": "Device name from mdcmd",
						"example":     "sdc",
					},
					"last_updated": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "Last update timestamp",
						"example":     "2025-06-16T06:30:00Z",
					},
				},
				"required": []string{"device", "serial_number", "power_state", "health_assessment", "state", "last_updated"},
			},
			"ParityCheckResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"status": map[string]interface{}{
						"type":        "string",
						"description": "Current parity check status (Running/Idle)",
						"example":     "Running",
					},
					"progress": map[string]interface{}{
						"type":        "integer",
						"description": "Progress percentage (0-100, when running)",
						"example":     75,
					},
					"speed": map[string]interface{}{
						"type":        "string",
						"description": "Current check speed (when running)",
						"example":     "45.2 MB/s",
					},
					"errors": map[string]interface{}{
						"type":        "integer",
						"description": "Current error count during check",
						"example":     0,
					},
					"last_check": map[string]interface{}{
						"type":        "string",
						"description": "Last completed check date/time",
						"example":     "Dec 15 14:30:25",
					},
					"duration": map[string]interface{}{
						"type":        "string",
						"description": "Duration of last completed check",
						"example":     "2h 15m",
					},
					"last_status": map[string]interface{}{
						"type":        "string",
						"description": "Status of last completed check",
						"example":     "Success",
					},
					"last_speed": map[string]interface{}{
						"type":        "string",
						"description": "Average speed of last completed check",
						"example":     "42.1 MB/s",
					},
					"next_check": map[string]interface{}{
						"type":        "string",
						"description": "Next scheduled check (if configured)",
						"example":     "Scheduled (check cron)",
					},
					"is_running": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether parity check is currently running",
						"example":     true,
					},
					"action": map[string]interface{}{
						"type":        "string",
						"description": "Current mdcmd resync action",
						"example":     "check P",
					},
					"resync_active": map[string]interface{}{
						"type":        "integer",
						"description": "Resync active indicator from mdcmd",
						"example":     1,
					},
					"last_updated": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "Last update timestamp",
						"example":     "2025-06-16T06:30:00Z",
					},
				},
				"required": []string{"status", "is_running", "action", "resync_active", "last_updated"},
			},
			"ZFSResponse": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"available": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether ZFS is available on the system",
						"example":     true,
					},
					"version": map[string]interface{}{
						"type":        "string",
						"description": "ZFS version",
						"example":     "2.1.5-1",
					},
					"pools": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"$ref": "#/components/schemas/ZFSPool",
						},
						"description": "Array of ZFS pools",
					},
					"arc_size": map[string]interface{}{
						"type":        "integer",
						"description": "ARC cache size in bytes",
						"example":     1073741824,
					},
					"arc_max": map[string]interface{}{
						"type":        "integer",
						"description": "ARC maximum size in bytes",
						"example":     2147483648,
					},
					"arc_hit_ratio": map[string]interface{}{
						"type":        "number",
						"description": "ARC hit ratio percentage",
						"example":     95.5,
					},
					"last_updated": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "Last update timestamp",
						"example":     "2025-06-16T06:00:00Z",
					},
				},
				"required": []string{"available", "pools", "last_updated"},
			},
			"ZFSPool": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Pool name",
						"example":     "tank",
					},
					"state": map[string]interface{}{
						"type":        "string",
						"description": "Pool state",
						"example":     "ONLINE",
					},
					"health": map[string]interface{}{
						"type":        "string",
						"description": "Pool health status",
						"example":     "ONLINE",
					},
					"size": map[string]interface{}{
						"type":        "integer",
						"description": "Total pool size in bytes",
						"example":     8796093022208,
					},
					"allocated": map[string]interface{}{
						"type":        "integer",
						"description": "Allocated space in bytes",
						"example":     6597069766656,
					},
					"free": map[string]interface{}{
						"type":        "integer",
						"description": "Free space in bytes",
						"example":     2199023255552,
					},
					"used_percent": map[string]interface{}{
						"type":        "number",
						"description": "Used percentage",
						"example":     75.0,
					},
					"size_formatted": map[string]interface{}{
						"type":        "string",
						"description": "Formatted total size",
						"example":     "8.0T",
					},
					"alloc_formatted": map[string]interface{}{
						"type":        "string",
						"description": "Formatted allocated space",
						"example":     "6.0T",
					},
					"free_formatted": map[string]interface{}{
						"type":        "string",
						"description": "Formatted free space",
						"example":     "2.0T",
					},
					"fragmentation": map[string]interface{}{
						"type":        "number",
						"description": "Fragmentation percentage",
						"example":     15.5,
					},
					"deduplication": map[string]interface{}{
						"type":        "number",
						"description": "Deduplication ratio",
						"example":     1.2,
					},
					"compression": map[string]interface{}{
						"type":        "number",
						"description": "Compression ratio",
						"example":     1.8,
					},
					"read_ops": map[string]interface{}{
						"type":        "integer",
						"description": "Read operations",
						"example":     1000,
					},
					"write_ops": map[string]interface{}{
						"type":        "integer",
						"description": "Write operations",
						"example":     500,
					},
					"read_bandwidth": map[string]interface{}{
						"type":        "integer",
						"description": "Read bandwidth in bytes/sec",
						"example":     104857600,
					},
					"write_bandwidth": map[string]interface{}{
						"type":        "integer",
						"description": "Write bandwidth in bytes/sec",
						"example":     52428800,
					},
					"vdevs": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"$ref": "#/components/schemas/ZFSVdev",
						},
						"description": "Virtual devices in the pool",
					},
					"last_scrub": map[string]interface{}{
						"type":        "string",
						"description": "Last scrub date/time",
						"example":     "Sun Dec 15 14:30:00 2024",
					},
					"scrub_status": map[string]interface{}{
						"type":        "string",
						"description": "Scrub status",
						"example":     "completed",
					},
					"error_count": map[string]interface{}{
						"type":        "integer",
						"description": "Total error count",
						"example":     0,
					},
					"version": map[string]interface{}{
						"type":        "string",
						"description": "Pool version",
						"example":     "5000",
					},
					"features": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
						"description": "Enabled features",
						"example":     []string{"async_destroy", "empty_bpobj", "lz4_compress"},
					},
					"last_updated": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "Last update timestamp",
						"example":     "2025-06-16T06:00:00Z",
					},
				},
				"required": []string{"name", "state", "health", "size", "allocated", "free", "last_updated"},
			},
			"ZFSVdev": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Vdev name",
						"example":     "/dev/sda1",
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "Vdev type",
						"example":     "disk",
					},
					"state": map[string]interface{}{
						"type":        "string",
						"description": "Vdev state",
						"example":     "ONLINE",
					},
					"health": map[string]interface{}{
						"type":        "string",
						"description": "Vdev health",
						"example":     "ONLINE",
					},
					"read_errors": map[string]interface{}{
						"type":        "integer",
						"description": "Read error count",
						"example":     0,
					},
					"write_errors": map[string]interface{}{
						"type":        "integer",
						"description": "Write error count",
						"example":     0,
					},
					"cksum_errors": map[string]interface{}{
						"type":        "integer",
						"description": "Checksum error count",
						"example":     0,
					},
					"children": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"$ref": "#/components/schemas/ZFSVdev",
						},
						"description": "Child vdevs for mirror/raidz groups",
					},
				},
				"required": []string{"name", "type", "state", "health"},
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

	// Docker individual container control endpoints
	paths["/api/v1/docker/containers/{id}/start"] = map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Start individual Docker container",
			"description": "Start a specific Docker container by ID or name",
			"tags":        []string{"Docker"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"parameters": []interface{}{
				map[string]interface{}{
					"name":        "id",
					"in":          "path",
					"description": "Container ID or name",
					"required":    true,
					"schema": map[string]interface{}{
						"type":    "string",
						"example": "plex",
					},
				},
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
				"404": map[string]interface{}{
					"description": "Container not found",
				},
				"500": map[string]interface{}{
					"description": "Failed to start container",
				},
			},
		},
	}

	paths["/api/v1/docker/containers/{id}/stop"] = map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Stop individual Docker container",
			"description": "Stop a specific Docker container by ID or name with optional timeout",
			"tags":        []string{"Docker"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"parameters": []interface{}{
				map[string]interface{}{
					"name":        "id",
					"in":          "path",
					"description": "Container ID or name",
					"required":    true,
					"schema": map[string]interface{}{
						"type":    "string",
						"example": "plex",
					},
				},
				map[string]interface{}{
					"name":        "timeout",
					"in":          "query",
					"description": "Timeout in seconds before force kill",
					"required":    false,
					"schema": map[string]interface{}{
						"type":    "integer",
						"minimum": 1,
						"maximum": 300,
						"default": 10,
					},
				},
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
				"404": map[string]interface{}{
					"description": "Container not found",
				},
				"500": map[string]interface{}{
					"description": "Failed to stop container",
				},
			},
		},
	}

	paths["/api/v1/docker/containers/{id}/restart"] = map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Restart individual Docker container",
			"description": "Restart a specific Docker container by ID or name with optional timeout",
			"tags":        []string{"Docker"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"parameters": []interface{}{
				map[string]interface{}{
					"name":        "id",
					"in":          "path",
					"description": "Container ID or name",
					"required":    true,
					"schema": map[string]interface{}{
						"type":    "string",
						"example": "plex",
					},
				},
				map[string]interface{}{
					"name":        "timeout",
					"in":          "query",
					"description": "Timeout in seconds before force kill",
					"required":    false,
					"schema": map[string]interface{}{
						"type":    "integer",
						"minimum": 1,
						"maximum": 300,
						"default": 10,
					},
				},
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
				"404": map[string]interface{}{
					"description": "Container not found",
				},
				"500": map[string]interface{}{
					"description": "Failed to restart container",
				},
			},
		},
	}

	paths["/api/v1/docker/containers/{id}/pause"] = map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Pause individual Docker container",
			"description": "Pause a specific Docker container by ID or name",
			"tags":        []string{"Docker"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"parameters": []interface{}{
				map[string]interface{}{
					"name":        "id",
					"in":          "path",
					"description": "Container ID or name",
					"required":    true,
					"schema": map[string]interface{}{
						"type":    "string",
						"example": "plex",
					},
				},
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
				"404": map[string]interface{}{
					"description": "Container not found",
				},
				"500": map[string]interface{}{
					"description": "Failed to pause container",
				},
			},
		},
	}

	paths["/api/v1/docker/containers/{id}/resume"] = map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Resume individual Docker container",
			"description": "Resume a paused Docker container by ID or name",
			"tags":        []string{"Docker"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"parameters": []interface{}{
				map[string]interface{}{
					"name":        "id",
					"in":          "path",
					"description": "Container ID or name",
					"required":    true,
					"schema": map[string]interface{}{
						"type":    "string",
						"example": "plex",
					},
				},
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
				"404": map[string]interface{}{
					"description": "Container not found",
				},
				"500": map[string]interface{}{
					"description": "Failed to resume container",
				},
			},
		},
	}

	// Array Control endpoints with enhanced orchestration
	paths["/api/v1/storage/array/start"] = map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Start Unraid array with orchestration",
			"description": "Start the Unraid array with proper orchestration sequence including disk detection, MD device assembly, filesystem mounting, and service initialization",
			"tags":        []string{"Storage", "Array Control"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"requestBody": map[string]interface{}{
				"required": false,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/ArrayStartRequest",
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Array started successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/ArrayOperationResponse",
							},
						},
					},
				},
				"400": map[string]interface{}{
					"description": "Invalid request parameters",
				},
				"409": map[string]interface{}{
					"description": "Array already started or configuration invalid",
				},
				"500": map[string]interface{}{
					"description": "Array start failed",
				},
			},
		},
	}

	paths["/api/v1/storage/array/stop"] = map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Stop Unraid array with orchestration",
			"description": "Stop the Unraid array with proper orchestration sequence including Docker container shutdown, VM shutdown, user share unmounting, disk unmounting, and MD device deactivation",
			"tags":        []string{"Storage", "Array Control"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"requestBody": map[string]interface{}{
				"required": false,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/ArrayStopRequest",
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Array stopped successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/ArrayOperationResponse",
							},
						},
					},
				},
				"400": map[string]interface{}{
					"description": "Invalid request parameters",
				},
				"409": map[string]interface{}{
					"description": "Array already stopped or dependencies prevent stop",
				},
				"500": map[string]interface{}{
					"description": "Array stop failed",
				},
			},
		},
	}

	// ZFS storage endpoint
	paths["/api/v1/storage/zfs"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get ZFS storage information",
			"description": "Retrieve comprehensive ZFS pool information including health, usage, configuration, and ARC cache statistics",
			"tags":        []string{"Storage", "Monitoring"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "ZFS information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/ZFSResponse",
							},
						},
					},
				},
				"500": map[string]interface{}{
					"description": "Failed to get ZFS information",
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
			"description": "Get comprehensive UPS monitoring data including battery status, power metrics, and hardware information from connected APC or NUT UPS systems",
			"tags":        []string{"System", "Monitoring"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "UPS status retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/EnhancedUPSStatus",
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

	// System Control endpoints
	paths["/api/v1/system/scripts"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "List user scripts",
			"description": "Get a list of all available user scripts on the system",
			"tags":        []string{"System", "Management"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "User scripts retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"$ref": "#/components/schemas/UserScript",
								},
							},
						},
					},
				},
			},
		},
		"post": map[string]interface{}{
			"summary":     "Execute user script",
			"description": "Execute a user script with optional parameters and background execution",
			"tags":        []string{"System", "Management"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"requestBody": map[string]interface{}{
				"required": true,
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
				"404": map[string]interface{}{
					"description": "Script not found",
				},
				"500": map[string]interface{}{
					"description": "Script execution failed",
				},
			},
		},
	}

	paths["/api/v1/system/reboot"] = map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Reboot system",
			"description": "Safely reboot the Unraid system with optional delay and custom message",
			"tags":        []string{"System", "Management"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"requestBody": map[string]interface{}{
				"required": false,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/SystemPowerRequest",
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Reboot initiated successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/SystemPowerResponse",
							},
						},
					},
				},
				"500": map[string]interface{}{
					"description": "Failed to initiate reboot",
				},
			},
		},
	}

	paths["/api/v1/system/shutdown"] = map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Shutdown system",
			"description": "Safely shutdown the Unraid system with optional delay and custom message",
			"tags":        []string{"System", "Management"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"requestBody": map[string]interface{}{
				"required": false,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/SystemPowerRequest",
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Shutdown initiated successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/SystemPowerResponse",
							},
						},
					},
				},
				"500": map[string]interface{}{
					"description": "Failed to initiate shutdown",
				},
			},
		},
	}

	paths["/api/v1/system/logs"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get system logs",
			"description": "Retrieve system logs with filtering options for type, lines, and time range",
			"tags":        []string{"System", "Monitoring"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"parameters": []interface{}{
				map[string]interface{}{
					"name":        "type",
					"in":          "query",
					"description": "Log type to retrieve",
					"required":    false,
					"schema": map[string]interface{}{
						"type":    "string",
						"enum":    []string{"system", "kernel", "docker", "nginx", "unraid"},
						"default": "system",
					},
				},
				map[string]interface{}{
					"name":        "lines",
					"in":          "query",
					"description": "Number of log lines to retrieve",
					"required":    false,
					"schema": map[string]interface{}{
						"type":    "integer",
						"minimum": 1,
						"maximum": 10000,
						"default": 100,
					},
				},
				map[string]interface{}{
					"name":        "since",
					"in":          "query",
					"description": "Retrieve logs since this timestamp (ISO 8601 format)",
					"required":    false,
					"schema": map[string]interface{}{
						"type":   "string",
						"format": "date-time",
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "System logs retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/SystemLogsResponse",
							},
						},
					},
				},
				"400": map[string]interface{}{
					"description": "Invalid parameters",
				},
				"500": map[string]interface{}{
					"description": "Failed to retrieve logs",
				},
			},
		},
	}

	// Enhanced system logs - all log files endpoint
	paths["/api/v1/system/logs/all"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get all available log files",
			"description": "Scan and return metadata for all log files in the system. Restricted to /var/log directory for security.",
			"tags":        []string{"System", "Monitoring"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"parameters": []interface{}{
				map[string]interface{}{
					"name":        "directory",
					"in":          "query",
					"description": "Base directory to scan (default: /var/log)",
					"required":    false,
					"schema": map[string]interface{}{
						"type":    "string",
						"default": "/var/log",
					},
				},
				map[string]interface{}{
					"name":        "recursive",
					"in":          "query",
					"description": "Include subdirectories in scan",
					"required":    false,
					"schema": map[string]interface{}{
						"type":    "boolean",
						"default": true,
					},
				},
				map[string]interface{}{
					"name":        "file_pattern",
					"in":          "query",
					"description": "Regex pattern for file matching",
					"required":    false,
					"schema": map[string]interface{}{
						"type": "string",
					},
				},
				map[string]interface{}{
					"name":        "max_files",
					"in":          "query",
					"description": "Maximum number of files to return",
					"required":    false,
					"schema": map[string]interface{}{
						"type":    "integer",
						"minimum": 1,
						"maximum": 1000,
						"default": 50,
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Log files metadata retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/LogFilesResponse",
							},
						},
					},
				},
				"403": map[string]interface{}{
					"description": "Access denied - directory outside /var/log",
				},
				"500": map[string]interface{}{
					"description": "Failed to scan log files",
				},
			},
		},
	}

	// Command execution endpoint
	paths["/api/v1/system/execute"] = map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Execute system command",
			"description": "Execute arbitrary system commands with timeout and security restrictions. Commands are executed with proper sanitization and blacklist filtering.",
			"tags":        []string{"System"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/CommandExecuteRequest",
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
								"$ref": "#/components/schemas/CommandExecuteResponse",
							},
						},
					},
				},
				"400": map[string]interface{}{
					"description": "Invalid request body or missing command",
				},
				"403": map[string]interface{}{
					"description": "Command not allowed (blacklisted for security)",
				},
				"500": map[string]interface{}{
					"description": "Command execution failed",
				},
			},
		},
	}

	// Dedicated CPU endpoint
	paths["/api/v1/system/cpu"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get CPU information",
			"description": "Retrieve comprehensive CPU metrics including usage, cores, frequency, temperature, and load averages",
			"tags":        []string{"System", "Monitoring"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "CPU information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/CPUResponse",
							},
						},
					},
				},
				"500": map[string]interface{}{
					"description": "Failed to get CPU information",
				},
			},
		},
	}

	// Dedicated Memory endpoint
	paths["/api/v1/system/memory"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get memory information",
			"description": "Retrieve comprehensive memory metrics including usage, breakdown by category, and formatted values",
			"tags":        []string{"System", "Monitoring"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Memory information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/MemoryResponse",
							},
						},
					},
				},
				"500": map[string]interface{}{
					"description": "Failed to get memory information",
				},
			},
		},
	}

	// Dedicated Fans endpoint
	paths["/api/v1/system/fans"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get fan information",
			"description": "Retrieve fan speed, labels, and status information from system sensors",
			"tags":        []string{"System", "Monitoring"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Fan information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/FansResponse",
							},
						},
					},
				},
				"500": map[string]interface{}{
					"description": "Failed to get fan information",
				},
			},
		},
	}

	// Parity disk endpoint
	paths["/api/v1/system/parity/disk"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get parity disk information",
			"description": "Retrieve parity disk status, health, and configuration information from mdcmd",
			"tags":        []string{"System", "Storage"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Parity disk information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/ParityDiskResponse",
							},
						},
					},
				},
				"404": map[string]interface{}{
					"description": "No parity disk found",
				},
				"500": map[string]interface{}{
					"description": "Failed to get parity disk information",
				},
			},
		},
	}

	// Parity check endpoint
	paths["/api/v1/system/parity/check"] = map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get parity check status",
			"description": "Retrieve parity check status, progress, and history information from mdcmd",
			"tags":        []string{"System", "Storage"},
			"security": []map[string]interface{}{
				{"BearerAuth": []string{}},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Parity check information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/ParityCheckResponse",
							},
						},
					},
				},
				"404": map[string]interface{}{
					"description": "No parity check information available",
				},
				"500": map[string]interface{}{
					"description": "Failed to get parity check information",
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
