package handlers

import (
	"fmt"
	"net/http"

	"github.com/domalab/uma/daemon/services/api/openapi/schemas"
	"github.com/domalab/uma/daemon/services/api/utils"
	"github.com/domalab/uma/daemon/services/api/version"
)

// DocsHandler handles API documentation HTTP requests
type DocsHandler struct {
	version string
	baseURL string
}

// NewDocsHandler creates a new docs handler
func NewDocsHandler(version, baseURL string) *DocsHandler {
	return &DocsHandler{
		version: version,
		baseURL: baseURL,
	}
}

// SwaggerUIHandler serves the Swagger UI interface
func (h *DocsHandler) SwaggerUIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get the base URL from the request
	baseURL := h.getBaseURLFromRequest(r)

	// Serve Swagger UI HTML
	html := h.generateSwaggerHTML(baseURL)
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// OpenAPIHandler serves the OpenAPI specification
func (h *DocsHandler) OpenAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get the base URL from the request
	baseURL := h.getBaseURLFromRequest(r)

	spec := h.generateOpenAPISpec(baseURL)
	utils.WriteJSON(w, http.StatusOK, spec)
}

// getBaseURLFromRequest extracts the base URL from the HTTP request
func (h *DocsHandler) getBaseURLFromRequest(r *http.Request) string {
	// Use the configured baseURL if available
	if h.baseURL != "" {
		return h.baseURL
	}

	// Extract from request headers
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	host := r.Host
	if host == "" {
		host = r.Header.Get("Host")
	}

	return fmt.Sprintf("%s://%s", scheme, host)
}

// generateSwaggerHTML generates the Swagger UI HTML page
func (h *DocsHandler) generateSwaggerHTML(baseURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Unraid Management Agent REST API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.25.2/swagger-ui.css" />
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
    <script src="https://unpkg.com/swagger-ui-dist@5.25.2/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.25.2/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '%s/api/v1/openapi.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>`, baseURL)
}

// generateOpenAPISpec generates the OpenAPI 3.0.3 specification
func (h *DocsHandler) generateOpenAPISpec(baseURL string) map[string]interface{} {
	// Get enhanced version information
	apiVersion := h.getAPIVersion()
	buildInfo := h.getBuildInfo()

	return map[string]interface{}{
		"openapi": "3.0.3",
		"info": map[string]interface{}{
			"title":       "Unraid Management Agent REST API",
			"description": h.getEnhancedDescription(),
			"version":     apiVersion,
			"contact": map[string]interface{}{
				"name":  "UMA Project",
				"url":   "https://github.com/domalab/uma",
				"email": "support@uma-project.com",
			},
			"license": map[string]interface{}{
				"name": "MIT",
				"url":  "https://opensource.org/licenses/MIT",
			},
			"x-build-info": buildInfo,
			"x-api-features": []string{
				"Real-time monitoring",
				"Container management",
				"VM control",
				"Storage operations",
				"WebSocket streaming",
				"Async operations",
				"Prometheus metrics",
			},
		},
		"servers": h.generateServers(baseURL),
		"paths":   h.generatePaths(),
		"components": map[string]interface{}{
			"schemas": h.generateSchemas(),
		},
	}
}

// generatePaths generates the OpenAPI paths specification
func (h *DocsHandler) generatePaths() map[string]interface{} {
	paths := make(map[string]interface{})

	// Add all endpoint categories
	h.addHealthPaths(paths)
	h.addSystemPaths(paths)
	h.addStoragePaths(paths)
	h.addDockerPaths(paths)
	h.addVMPaths(paths)
	h.addMetricsPaths(paths)
	h.addWebSocketPaths(paths)
	h.addNotificationPaths(paths)
	h.addAsyncPaths(paths)
	h.addDiagnosticsPaths(paths)

	return paths
}

// addHealthPaths adds health and documentation endpoints
func (h *DocsHandler) addHealthPaths(paths map[string]interface{}) {
	paths["/health"] = map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []string{"Health"},
			"summary":     "Health check",
			"description": "Get system health status",
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Health status",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/HealthResponse",
							},
						},
					},
				},
			},
		},
	}
}

// addSystemPaths adds all system monitoring and control endpoints
func (h *DocsHandler) addSystemPaths(paths map[string]interface{}) {
	// System monitoring endpoints
	paths["/system/info"] = h.createGetEndpoint("System", "Get system information", "Retrieve general system information", "SystemInfo")
	paths["/system/cpu"] = h.createGetEndpoint("System", "Get CPU information", "Retrieve CPU usage and information", "CPUInfo")
	paths["/system/memory"] = h.createGetEndpoint("System", "Get memory information", "Retrieve memory usage and information", "MemoryInfo")
	paths["/system/temperature"] = h.createGetEndpoint("System", "Get temperature information", "Retrieve system temperature data", "TemperatureInfo")
	paths["/system/fans"] = h.createGetEndpoint("System", "Get fan information", "Retrieve fan speed and status", "FanInfo")
	paths["/system/gpu"] = h.createGetEndpoint("System", "Get GPU information", "Retrieve GPU usage and information", "GPUInfo")
	paths["/system/ups"] = h.createGetEndpoint("System", "Get UPS information", "Retrieve UPS status and battery information", "UPSInfo")
	paths["/system/network"] = h.createGetEndpoint("System", "Get network information", "Retrieve network interface information", "NetworkInfo")
	paths["/system/resources"] = h.createGetEndpoint("System", "Get system resources", "Retrieve comprehensive system resource information", "SystemResources")
	paths["/system/filesystems"] = h.createGetEndpoint("System", "Get filesystem information", "Retrieve filesystem usage information", "FilesystemInfo")
	paths["/system/logs"] = h.createGetEndpoint("System", "Get system logs", "Retrieve system log entries", "SystemLogs")

	// System control endpoints
	paths["/system/reboot"] = h.createPostEndpoint("System", "Reboot system", "Initiate system reboot", "SystemOperationResponse")
	paths["/system/shutdown"] = h.createPostEndpoint("System", "Shutdown system", "Initiate system shutdown", "SystemOperationResponse")

	// Parity endpoints
	paths["/system/parity/disk"] = h.createGetEndpoint("System", "Get parity disk information", "Retrieve parity disk status and information", "ParityDiskInfo")
	paths["/system/parity/check"] = map[string]interface{}{
		"get":  h.createMethodSpec("System", "Get parity check status", "Retrieve current parity check status", "ParityCheckStatus"),
		"post": h.createMethodSpec("System", "Start parity check", "Initiate a parity check operation", "ParityCheckResponse"),
	}
}

// generateSchemas generates the OpenAPI schemas using the full schema registry
func (h *DocsHandler) generateSchemas() map[string]interface{} {
	registry := schemas.NewRegistry()
	registry.RegisterAll()
	return registry.GetAllSchemas()
}

// addStoragePaths adds all storage management endpoints
func (h *DocsHandler) addStoragePaths(paths map[string]interface{}) {
	// Storage monitoring endpoints
	paths["/storage/array"] = h.createGetEndpoint("Storage", "Get array status", "Retrieve Unraid array status and information", "ArrayStatus")
	paths["/storage/disks"] = h.createGetEndpoint("Storage", "List storage disks", "Retrieve list of all storage disks", "DiskList")
	paths["/storage/cache"] = h.createGetEndpoint("Storage", "Get cache information", "Retrieve cache pool information", "CacheInfo")
	paths["/storage/boot"] = h.createGetEndpoint("Storage", "Get boot device information", "Retrieve boot device usage and information", "BootInfo")
	paths["/storage/zfs"] = h.createGetEndpoint("Storage", "Get ZFS information", "Retrieve ZFS pool and dataset information", "ZFSInfo")
	paths["/storage/general"] = h.createGetEndpoint("Storage", "Get general storage information", "Retrieve general storage usage information", "StorageGeneral")

	// Storage control endpoints
	paths["/storage/array/start"] = h.createPostEndpoint("Storage", "Start array", "Start the Unraid array", "ArrayOperationResponse")
	paths["/storage/array/stop"] = h.createPostEndpoint("Storage", "Stop array", "Stop the Unraid array", "ArrayOperationResponse")
}

// addDockerPaths adds all Docker management endpoints
func (h *DocsHandler) addDockerPaths(paths map[string]interface{}) {
	// Docker monitoring endpoints
	paths["/docker/containers"] = h.createGetEndpoint("Docker", "List Docker containers", "Retrieve list of Docker containers", "DockerContainerList")
	paths["/docker/images"] = h.createGetEndpoint("Docker", "List Docker images", "Retrieve list of Docker images", "DockerImageList")
	paths["/docker/networks"] = h.createGetEndpoint("Docker", "List Docker networks", "Retrieve list of Docker networks", "DockerNetworkList")
	paths["/docker/info"] = h.createGetEndpoint("Docker", "Get Docker information", "Retrieve Docker system information", "DockerInfo")

	// Docker container control endpoints
	paths["/docker/containers/{id}"] = map[string]interface{}{
		"get":  h.createMethodSpec("Docker", "Get container details", "Retrieve detailed information about a specific container", "DockerContainerInfo"),
		"post": h.createMethodSpec("Docker", "Control container", "Start, stop, restart, pause, or resume a container", "DockerOperationResponse"),
	}

	// Docker bulk operations
	paths["/docker/containers/bulk/start"] = h.createPostEndpoint("Docker", "Bulk start containers", "Start multiple containers", "BulkOperationResponse")
	paths["/docker/containers/bulk/stop"] = h.createPostEndpoint("Docker", "Bulk stop containers", "Stop multiple containers", "BulkOperationResponse")
	paths["/docker/containers/bulk/restart"] = h.createPostEndpoint("Docker", "Bulk restart containers", "Restart multiple containers", "BulkOperationResponse")
}

// addVMPaths adds all VM management endpoints
func (h *DocsHandler) addVMPaths(paths map[string]interface{}) {
	paths["/vms"] = h.createGetEndpoint("Virtual Machines", "List virtual machines", "Retrieve list of virtual machines", "VMList")
	paths["/vms/{name}"] = map[string]interface{}{
		"get":  h.createMethodSpec("Virtual Machines", "Get VM details", "Retrieve detailed information about a specific VM", "VMInfo"),
		"post": h.createMethodSpec("Virtual Machines", "Control VM", "Start, stop, pause, or resume a VM", "VMOperationResponse"),
	}
	paths["/vms/{name}/stats"] = h.createGetEndpoint("Virtual Machines", "Get VM statistics", "Retrieve performance statistics for a VM", "VMStats")
	paths["/vms/{name}/snapshots"] = map[string]interface{}{
		"get":  h.createMethodSpec("Virtual Machines", "List VM snapshots", "Retrieve list of VM snapshots", "VMSnapshotList"),
		"post": h.createMethodSpec("Virtual Machines", "Create VM snapshot", "Create a new VM snapshot", "VMSnapshotResponse"),
	}
}

// addMetricsPaths adds Prometheus metrics endpoints
func (h *DocsHandler) addMetricsPaths(paths map[string]interface{}) {
	paths["/metrics"] = map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []string{"Monitoring"},
			"summary":     "Get Prometheus metrics",
			"description": "Retrieve Prometheus-formatted metrics for monitoring",
			"produces":    []string{"text/plain"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Prometheus metrics",
					"content": map[string]interface{}{
						"text/plain": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
			},
		},
	}
}

// addWebSocketPaths adds WebSocket endpoints
func (h *DocsHandler) addWebSocketPaths(paths map[string]interface{}) {
	paths["/ws/system/stats"] = h.createGetEndpoint("WebSocket", "System stats WebSocket", "Real-time system statistics via WebSocket", "SystemStatsStream")
	paths["/ws/docker/events"] = h.createGetEndpoint("WebSocket", "Docker events WebSocket", "Real-time Docker events via WebSocket", "DockerEventsStream")
	paths["/ws/storage/status"] = h.createGetEndpoint("WebSocket", "Storage status WebSocket", "Real-time storage status via WebSocket", "StorageStatusStream")
}

// addNotificationPaths adds notification management endpoints
func (h *DocsHandler) addNotificationPaths(paths map[string]interface{}) {
	paths["/notifications"] = map[string]interface{}{
		"get":  h.createMethodSpec("Notifications", "List notifications", "Retrieve list of notifications", "NotificationList"),
		"post": h.createMethodSpec("Notifications", "Create notification", "Create a new notification", "NotificationResponse"),
	}
	paths["/notifications/{id}"] = map[string]interface{}{
		"get":    h.createMethodSpec("Notifications", "Get notification", "Retrieve a specific notification", "NotificationInfo"),
		"delete": h.createMethodSpec("Notifications", "Delete notification", "Delete a specific notification", "NotificationResponse"),
	}
	paths["/notifications/clear"] = h.createPostEndpoint("Notifications", "Clear all notifications", "Clear all notifications", "NotificationResponse")
	paths["/notifications/stats"] = h.createGetEndpoint("Notifications", "Get notification statistics", "Retrieve notification statistics", "NotificationStats")
	paths["/notifications/mark-all-read"] = h.createPostEndpoint("Notifications", "Mark all as read", "Mark all notifications as read", "NotificationResponse")
}

// addAsyncPaths adds async operation endpoints
func (h *DocsHandler) addAsyncPaths(paths map[string]interface{}) {
	paths["/operations"] = h.createGetEndpoint("Async Operations", "List operations", "Retrieve list of async operations", "OperationList")
	paths["/operations/{id}"] = h.createGetEndpoint("Async Operations", "Get operation", "Retrieve a specific async operation", "OperationInfo")
	paths["/operations/stats"] = h.createGetEndpoint("Async Operations", "Get operation statistics", "Retrieve async operation statistics", "OperationStats")
}

// addDiagnosticsPaths adds diagnostic endpoints
func (h *DocsHandler) addDiagnosticsPaths(paths map[string]interface{}) {
	paths["/diagnostics/health"] = h.createGetEndpoint("Diagnostics", "System health diagnostics", "Retrieve comprehensive system health diagnostics", "DiagnosticsHealth")
	paths["/diagnostics/info"] = h.createGetEndpoint("Diagnostics", "System diagnostic information", "Retrieve system diagnostic information", "DiagnosticsInfo")
	paths["/diagnostics/repair"] = h.createPostEndpoint("Diagnostics", "Run system repair", "Run system diagnostic repair operations", "DiagnosticsRepair")
}

// Helper methods for creating OpenAPI specifications

// createGetEndpoint creates a GET endpoint specification
func (h *DocsHandler) createGetEndpoint(tag, summary, description, schema string) map[string]interface{} {
	return map[string]interface{}{
		"get": h.createMethodSpec(tag, summary, description, schema),
	}
}

// createPostEndpoint creates a POST endpoint specification
func (h *DocsHandler) createPostEndpoint(tag, summary, description, schema string) map[string]interface{} {
	return map[string]interface{}{
		"post": h.createMethodSpec(tag, summary, description, schema),
	}
}

// createMethodSpec creates a method specification
func (h *DocsHandler) createMethodSpec(tag, summary, description, schema string) map[string]interface{} {
	return map[string]interface{}{
		"tags":        []string{tag},
		"summary":     summary,
		"description": description,
		"responses": map[string]interface{}{
			"200": map[string]interface{}{
				"description": "Success",
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/" + schema,
						},
					},
				},
			},
		},
	}
}

// getAPIVersion returns the API version for OpenAPI specification
func (h *DocsHandler) getAPIVersion() string {
	// Use the version package for consistent version handling
	return version.GetAPIVersion(h.version)
}

// getBuildInfo returns build information for the OpenAPI specification
func (h *DocsHandler) getBuildInfo() map[string]interface{} {
	// Get comprehensive build information
	buildInfo := version.GetBuildInfo(h.version)

	return map[string]interface{}{
		"version":    buildInfo.Version,
		"git_commit": buildInfo.GitCommit,
		"git_tag":    buildInfo.GitTag,
		"go_version": buildInfo.GoVersion,
		"platform":   buildInfo.Platform,
		"build_time": buildInfo.BuildTime.Format("2006-01-02T15:04:05Z"),
		"repository": "https://github.com/domalab/uma",
		"dirty":      buildInfo.Dirty,
		"user_agent": version.GetUserAgent(h.version),
		"is_dev":     version.IsDevVersion(h.version),
	}
}

// getEnhancedDescription returns an enhanced API description
func (h *DocsHandler) getEnhancedDescription() string {
	return `**Unraid Management Agent (UMA) REST API**

A comprehensive REST API for monitoring and controlling Unraid server infrastructure. UMA provides real-time access to system metrics, storage management, Docker container control, and virtual machine operations.

## Key Features

- **Real-time Monitoring**: Live system statistics, resource usage, and health metrics
- **Container Management**: Complete Docker container lifecycle control (start, stop, restart, pause, resume)
- **Virtual Machine Control**: VM management including snapshots, resource allocation, and state control
- **Storage Operations**: Array management, disk monitoring, parity operations, and SMART data
- **WebSocket Streaming**: Real-time event streams for system stats, Docker events, and storage status
- **Async Operations**: Long-running operation tracking with progress monitoring
- **Prometheus Integration**: Native metrics export for monitoring and alerting

## Security Model

This API is designed for **internal network use only** and does not require authentication. Security is handled at the network/firewall level. All operations are performed with appropriate system privileges.

## Response Format

All responses follow a standardized format with proper HTTP status codes, request IDs for tracing, and comprehensive error handling.

## Rate Limiting

The API implements intelligent rate limiting to prevent resource exhaustion while allowing normal operational use.`
}

// generateServers returns the servers configuration for OpenAPI
func (h *DocsHandler) generateServers(baseURL string) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"url":         fmt.Sprintf("%s/api/v1", baseURL),
			"description": "Unraid Management Agent API Server",
		},
	}
}
