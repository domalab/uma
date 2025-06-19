package openapi

import "fmt"

// Config holds configuration for OpenAPI spec generation
type Config struct {
	Version     string
	Port        int
	BaseURL     string
	Environment string // dev, staging, prod
	Features    FeatureFlags
}

// FeatureFlags controls which API features are included in the spec
type FeatureFlags struct {
	Authentication bool
	BulkOperations bool
	WebSockets     bool
	Metrics        bool
	ZFS            bool
	ArrayControl   bool
	VMManagement   bool
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Version:     "2025.06.16",
		Port:        34600,
		BaseURL:     "",
		Environment: "prod",
		Features: FeatureFlags{
			Authentication: false, // Disabled for internal network API
			BulkOperations: true,
			WebSockets:     true,
			Metrics:        true,
			ZFS:            true,
			ArrayControl:   true,
			VMManagement:   true,
		},
	}
}

// GetServers returns the server configurations for the OpenAPI spec
func (c *Config) GetServers() []OpenAPIServer {
	servers := []OpenAPIServer{
		{
			URL:         fmt.Sprintf("http://localhost:%d", c.Port),
			Description: "Local UMA API server",
		},
	}

	if c.Environment == "prod" {
		servers = append(servers, OpenAPIServer{
			URL:         "http://your-unraid-server:34600",
			Description: "Remote UMA API server (replace with your server IP)",
		})
	}

	if c.BaseURL != "" {
		servers = append(servers, OpenAPIServer{
			URL:         c.BaseURL,
			Description: "Custom UMA API server",
		})
	}

	return servers
}

// GetSecuritySchemes returns the security schemes for the OpenAPI spec
func (c *Config) GetSecuritySchemes() map[string]interface{} {
	schemes := make(map[string]interface{})

	if c.Features.Authentication {
		schemes["BearerAuth"] = map[string]interface{}{
			"type":         "http",
			"scheme":       "bearer",
			"bearerFormat": "JWT",
			"description":  "JWT token obtained from /api/v1/auth/login",
		}
		schemes["ApiKeyAuth"] = map[string]interface{}{
			"type":        "apiKey",
			"in":          "header",
			"name":        "X-API-Key",
			"description": "API key for authentication",
		}
	}

	return schemes
}
