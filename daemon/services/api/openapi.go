package api

import (
	"fmt"
	"net/http"

	"github.com/domalab/uma/daemon/services/api/openapi"
	"github.com/domalab/uma/daemon/services/api/utils"
)

// Legacy type aliases for backward compatibility
type OpenAPISpec = openapi.OpenAPISpec
type OpenAPIInfo = openapi.OpenAPIInfo
type OpenAPIContact = openapi.OpenAPIContact
type OpenAPIServer = openapi.OpenAPIServer
type OpenAPIComponents = openapi.OpenAPIComponents

// generateOpenAPISpec creates the complete OpenAPI specification using the new modular structure
func (h *HTTPServer) generateOpenAPISpec() *OpenAPISpec {
	// Create configuration for the OpenAPI generator
	config := &openapi.Config{
		Version:     h.getAPIVersion(),
		Port:        h.api.ctx.Config.HTTPServer.Port,
		BaseURL:     "",
		Environment: "prod",
		Features: openapi.FeatureFlags{
			Authentication: true,
			BulkOperations: true,
			WebSockets:     true,
			Metrics:        true,
			ZFS:            true,
			ArrayControl:   true,
			VMManagement:   true,
		},
	}

	// Create generator and generate specification
	generator := openapi.NewGenerator(config)
	return generator.Generate()
}

// getAPIVersion returns the API version with fallback
func (h *HTTPServer) getAPIVersion() string {
	version := h.api.ctx.Config.Version
	if version == "" || version == "unknown" {
		version = "2025.06.16" // Current plugin version
	}
	return version
}

// OpenAPIHandler serves the OpenAPI specification
func (h *HTTPServer) OpenAPIHandler(w http.ResponseWriter, r *http.Request) {
	spec := h.generateOpenAPISpec()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	utils.WriteJSON(w, http.StatusOK, spec)
}

// SwaggerUIHandler serves the Swagger UI documentation
func (h *HTTPServer) SwaggerUIHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>UMA REST API Documentation</title>
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
        .swagger-ui .topbar {
            background-color: #2c3e50;
        }
        .swagger-ui .topbar .download-url-wrapper .select-label {
            color: #ffffff;
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
                tryItOutEnabled: true,
                requestInterceptor: function(request) {
                    // Add any custom headers or authentication here
                    return request;
                },
                responseInterceptor: function(response) {
                    // Handle responses here if needed
                    return response;
                }
            });
        };
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// HealthHandler provides a health check endpoint
func (h *HTTPServer) HealthHandler(w http.ResponseWriter, r *http.Request) {
	// Create a simple health response
	health := map[string]interface{}{
		"status":    "healthy",
		"version":   h.getAPIVersion(),
		"service":   "uma",
		"timestamp": fmt.Sprintf("%d", r.Context().Value("timestamp")),
		"dependencies": map[string]interface{}{
			"docker":     "available", // This would be checked dynamically
			"unraid_api": "available",
			"apcupsd":    "available",
		},
		"uptime": 0, // This would be calculated dynamically
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")

	utils.WriteJSON(w, http.StatusOK, health)
}

// GetOpenAPIStats returns statistics about the OpenAPI specification
func (h *HTTPServer) GetOpenAPIStats() map[string]interface{} {
	config := &openapi.Config{
		Version:     h.getAPIVersion(),
		Port:        h.api.ctx.Config.HTTPServer.Port,
		BaseURL:     "",
		Environment: "prod",
		Features: openapi.FeatureFlags{
			Authentication: true,
			BulkOperations: true,
			WebSockets:     true,
			Metrics:        true,
			ZFS:            true,
			ArrayControl:   true,
			VMManagement:   true,
		},
	}

	generator := openapi.NewGenerator(config)
	return generator.GetStats()
}

// ValidateOpenAPISpec validates the generated OpenAPI specification
func (h *HTTPServer) ValidateOpenAPISpec() []string {
	config := &openapi.Config{
		Version:     h.getAPIVersion(),
		Port:        h.api.ctx.Config.HTTPServer.Port,
		BaseURL:     "",
		Environment: "prod",
		Features: openapi.FeatureFlags{
			Authentication: true,
			BulkOperations: true,
			WebSockets:     true,
			Metrics:        true,
			ZFS:            true,
			ArrayControl:   true,
			VMManagement:   true,
		},
	}

	generator := openapi.NewGenerator(config)
	return generator.ValidateSpec()
}
