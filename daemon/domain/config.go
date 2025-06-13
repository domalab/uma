package domain

// Config holds the application configuration
type Config struct {
	Version    string     `json:"version"`
	ShowUps    bool       `json:"showups"`
	HTTPServer HTTPConfig `json:"http_server"`
	Auth       AuthConfig `json:"auth"`
	Logging    LogConfig  `json:"logging"`
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

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() Config {
	return Config{
		Version: "unknown",
		ShowUps: false,
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
			MaxSize:    10,
			MaxBackups: 10,
			MaxAge:     28,
		},
	}
}
