package domain

// Config holds the application configuration
type Config struct {
	Version    string     `json:"version"`
	HTTPServer HTTPConfig `json:"http_server"`
	Auth       AuthConfig `json:"auth"`
	Logging    LogConfig  `json:"logging"`
	MCP        MCPConfig  `json:"mcp"`
}

// HTTPConfig holds HTTP server configuration
type HTTPConfig struct {
	Enabled bool   `json:"enabled"`
	Port    int    `json:"port"`
	Host    string `json:"host"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled   bool   `json:"enabled"`
	APIKey    string `json:"api_key,omitempty"`
	JWTSecret string `json:"jwt_secret,omitempty"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level      string `json:"level"`
	MaxSize    int    `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`
}

// MCPConfig holds MCP (Model Context Protocol) server configuration
type MCPConfig struct {
	Enabled        bool `json:"enabled"`
	Port           int  `json:"port"`
	MaxConnections int  `json:"max_connections"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() Config {
	return Config{
		Version: "unknown",
		HTTPServer: HTTPConfig{
			Enabled: true,
			Port:    34600,
			Host:    "0.0.0.0",
		},
		Auth: AuthConfig{
			Enabled: false,
		},
		Logging: LogConfig{
			Level:      "info",
			MaxSize:    10, // 10MB limit as requested
			MaxBackups: 0,  // DISABLED - no backup files to prevent disk space issues
			MaxAge:     0,  // DISABLED - no age-based retention
		},
		MCP: MCPConfig{
			Enabled:        false,
			Port:           34800,
			MaxConnections: 100,
		},
	}
}
