package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/domalab/uma/daemon/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// ViperConfigService provides enhanced configuration management with Viper
type ViperConfigService struct {
	viper        *viper.Viper
	configPaths  []string
	configName   string
	configType   string
	watchEnabled bool
}

// NewViperConfigService creates a new Viper-based configuration service
func NewViperConfigService() *ViperConfigService {
	v := viper.New()

	service := &ViperConfigService{
		viper:        v,
		configPaths:  []string{".", "/etc/uma", "/usr/local/etc/uma", "$HOME/.uma"},
		configName:   "uma",
		configType:   "yaml", // Default to YAML, but will auto-detect
		watchEnabled: true,
	}

	service.setupViper()
	return service
}

// setupViper configures Viper with UMA-specific settings
func (c *ViperConfigService) setupViper() {
	// Set config name and type
	c.viper.SetConfigName(c.configName)
	c.viper.SetConfigType(c.configType)

	// Add config paths
	for _, path := range c.configPaths {
		c.viper.AddConfigPath(path)
	}

	// Environment variable configuration
	c.viper.SetEnvPrefix("UMA")
	c.viper.AutomaticEnv()
	c.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Set default values
	c.setDefaults()

	// Try to read config file
	if err := c.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Info("No config file found, using defaults and environment variables")
		} else {
			logger.Warn("Error reading config file: %v", err)
		}
	} else {
		logger.Info("Using config file: %s", c.viper.ConfigFileUsed())
	}

	// Setup config file watching
	if c.watchEnabled {
		c.viper.WatchConfig()
		c.viper.OnConfigChange(func(e fsnotify.Event) {
			logger.Info("Config file changed: %s", e.Name)
			c.onConfigChange()
		})
	}
}

// setDefaults sets default configuration values
func (c *ViperConfigService) setDefaults() {
	// HTTP Server defaults
	c.viper.SetDefault("http.port", 34600)
	c.viper.SetDefault("http.host", "0.0.0.0")
	c.viper.SetDefault("http.timeout", "60s")
	c.viper.SetDefault("http.read_timeout", "30s")
	c.viper.SetDefault("http.write_timeout", "30s")

	// Logging defaults - optimized for minimal disk usage
	c.viper.SetDefault("logging.level", "info")
	c.viper.SetDefault("logging.format", "console")
	c.viper.SetDefault("logging.file", "")
	c.viper.SetDefault("logging.max_size", 10)   // 10MB limit as requested
	c.viper.SetDefault("logging.max_backups", 0) // DISABLED - no backup files
	c.viper.SetDefault("logging.max_age", 0)     // DISABLED - no age-based retention

	// Metrics defaults
	c.viper.SetDefault("metrics.enabled", true)
	c.viper.SetDefault("metrics.path", "/metrics")

	// WebSocket defaults
	c.viper.SetDefault("websocket.enabled", true)
	c.viper.SetDefault("websocket.max_connections", 100)
	c.viper.SetDefault("websocket.ping_interval", "30s")
	c.viper.SetDefault("websocket.pong_timeout", "60s")

	// Monitoring defaults
	c.viper.SetDefault("monitoring.interval", "30s")
	c.viper.SetDefault("monitoring.docker.enabled", true)
	c.viper.SetDefault("monitoring.storage.enabled", true)
	c.viper.SetDefault("monitoring.system.enabled", true)
	c.viper.SetDefault("monitoring.ups.enabled", true)
	c.viper.SetDefault("monitoring.gpu.enabled", true)

	// Cache defaults
	c.viper.SetDefault("cache.enabled", true)
	c.viper.SetDefault("cache.ttl", "5m")
	c.viper.SetDefault("cache.cleanup_interval", "10m")

	// Rate limiting defaults
	c.viper.SetDefault("rate_limit.enabled", false)
	c.viper.SetDefault("rate_limit.requests_per_minute", 60)
	c.viper.SetDefault("rate_limit.burst", 10)

	// CORS defaults
	c.viper.SetDefault("cors.enabled", true)
	c.viper.SetDefault("cors.allowed_origins", []string{"*"})
	c.viper.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	c.viper.SetDefault("cors.allowed_headers", []string{"*"})

	// MCP (Model Context Protocol) defaults
	c.viper.SetDefault("mcp.enabled", false)
	c.viper.SetDefault("mcp.port", 34800)
	c.viper.SetDefault("mcp.max_connections", 100)
}

// onConfigChange handles configuration file changes
func (c *ViperConfigService) onConfigChange() {
	logger.Info("Configuration reloaded")
	// Emit config change event if needed
	// This could trigger service restarts or reconfigurations
}

// GetString returns a string configuration value
func (c *ViperConfigService) GetString(key string) string {
	return c.viper.GetString(key)
}

// GetInt returns an integer configuration value
func (c *ViperConfigService) GetInt(key string) int {
	return c.viper.GetInt(key)
}

// GetBool returns a boolean configuration value
func (c *ViperConfigService) GetBool(key string) bool {
	return c.viper.GetBool(key)
}

// GetDuration returns a duration configuration value
func (c *ViperConfigService) GetDuration(key string) time.Duration {
	return c.viper.GetDuration(key)
}

// GetStringSlice returns a string slice configuration value
func (c *ViperConfigService) GetStringSlice(key string) []string {
	return c.viper.GetStringSlice(key)
}

// GetStringMap returns a string map configuration value
func (c *ViperConfigService) GetStringMap(key string) map[string]interface{} {
	return c.viper.GetStringMap(key)
}

// Set sets a configuration value
func (c *ViperConfigService) Set(key string, value interface{}) {
	c.viper.Set(key, value)
}

// IsSet checks if a configuration key is set
func (c *ViperConfigService) IsSet(key string) bool {
	return c.viper.IsSet(key)
}

// GetAllSettings returns all configuration settings
func (c *ViperConfigService) GetAllSettings() map[string]interface{} {
	return c.viper.AllSettings()
}

// WriteConfig writes the current configuration to file
func (c *ViperConfigService) WriteConfig() error {
	return c.viper.WriteConfig()
}

// WriteConfigAs writes the current configuration to a specific file
func (c *ViperConfigService) WriteConfigAs(filename string) error {
	return c.viper.WriteConfigAs(filename)
}

// GetConfigFile returns the path to the config file being used
func (c *ViperConfigService) GetConfigFile() string {
	return c.viper.ConfigFileUsed()
}

// LoadFromFile loads configuration from a specific file
func (c *ViperConfigService) LoadFromFile(filename string) error {
	c.viper.SetConfigFile(filename)
	return c.viper.ReadInConfig()
}

// LoadFromEnv loads configuration from environment variables
func (c *ViperConfigService) LoadFromEnv() {
	// Environment variables are automatically loaded due to AutomaticEnv()
	logger.Info("Environment variables loaded automatically")
}

// GetHTTPConfig returns HTTP server configuration
func (c *ViperConfigService) GetHTTPConfig() HTTPConfig {
	return HTTPConfig{
		Port:         c.GetInt("http.port"),
		Host:         c.GetString("http.host"),
		Timeout:      c.GetDuration("http.timeout"),
		ReadTimeout:  c.GetDuration("http.read_timeout"),
		WriteTimeout: c.GetDuration("http.write_timeout"),
	}
}

// GetLoggingConfig returns logging configuration
func (c *ViperConfigService) GetLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level:      c.GetString("logging.level"),
		Format:     c.GetString("logging.format"),
		File:       c.GetString("logging.file"),
		MaxSize:    c.GetInt("logging.max_size"),
		MaxBackups: c.GetInt("logging.max_backups"),
		MaxAge:     c.GetInt("logging.max_age"),
	}
}

// GetMonitoringConfig returns monitoring configuration
func (c *ViperConfigService) GetMonitoringConfig() MonitoringConfig {
	return MonitoringConfig{
		Interval: c.GetDuration("monitoring.interval"),
		Docker:   c.GetBool("monitoring.docker.enabled"),
		Storage:  c.GetBool("monitoring.storage.enabled"),
		System:   c.GetBool("monitoring.system.enabled"),
		UPS:      c.GetBool("monitoring.ups.enabled"),
		GPU:      c.GetBool("monitoring.gpu.enabled"),
	}
}

// GetMCPConfig returns MCP (Model Context Protocol) configuration
func (c *ViperConfigService) GetMCPConfig() MCPConfig {
	return MCPConfig{
		Enabled:        c.GetBool("mcp.enabled"),
		Port:           c.GetInt("mcp.port"),
		MaxConnections: c.GetInt("mcp.max_connections"),
	}
}

// Configuration structs
type HTTPConfig struct {
	Port         int
	Host         string
	Timeout      time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type LoggingConfig struct {
	Level      string
	Format     string
	File       string
	MaxSize    int
	MaxBackups int
	MaxAge     int
}

type MonitoringConfig struct {
	Interval time.Duration
	Docker   bool
	Storage  bool
	System   bool
	UPS      bool
	GPU      bool
}

type MCPConfig struct {
	Enabled        bool
	Port           int
	MaxConnections int
}

// ValidateConfig validates the configuration
func (c *ViperConfigService) ValidateConfig() error {
	// Validate HTTP port
	port := c.GetInt("http.port")
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid HTTP port: %d", port)
	}

	// Validate log level
	level := c.GetString("logging.level")
	validLevels := []string{"debug", "info", "warn", "error", "fatal"}
	valid := false
	for _, validLevel := range validLevels {
		if level == validLevel {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid log level: %s", level)
	}

	// Validate monitoring interval
	interval := c.GetDuration("monitoring.interval")
	if interval < time.Second {
		return fmt.Errorf("monitoring interval too short: %v", interval)
	}

	// Validate MCP port
	mcpPort := c.GetInt("mcp.port")
	if mcpPort < 1024 || mcpPort > 65535 {
		return fmt.Errorf("invalid MCP port: %d", mcpPort)
	}

	// Validate MCP max connections
	maxConnections := c.GetInt("mcp.max_connections")
	if maxConnections < 1 || maxConnections > 10000 {
		return fmt.Errorf("invalid MCP max connections: %d", maxConnections)
	}

	return nil
}

// CreateSampleConfig creates a sample configuration file
func (c *ViperConfigService) CreateSampleConfig(filename string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Sample configuration content
	sampleConfig := `# UMA Configuration File
# This file contains configuration options for the Unraid Management Agent

# HTTP Server Configuration
http:
  port: 34600
  host: "0.0.0.0"
  timeout: "60s"
  read_timeout: "30s"
  write_timeout: "30s"

# Authentication Configuration
auth:
  enabled: false
  api_key: ""
  jwt_secret: ""
  token_expiry: "24h"

# Logging Configuration - Optimized for minimal disk usage
logging:
  level: "info"
  format: "console"
  file: ""
  max_size: 10      # 10MB limit to prevent disk space issues
  max_backups: 0    # DISABLED - no backup files
  max_age: 0        # DISABLED - no age-based retention

# Metrics Configuration
metrics:
  enabled: true
  path: "/metrics"

# WebSocket Configuration
websocket:
  enabled: true
  max_connections: 100
  ping_interval: "30s"
  pong_timeout: "60s"

# Monitoring Configuration
monitoring:
  interval: "30s"
  docker:
    enabled: true
  storage:
    enabled: true
  system:
    enabled: true
  ups:
    enabled: true
  gpu:
    enabled: true

# Cache Configuration
cache:
  enabled: true
  ttl: "5m"
  cleanup_interval: "10m"

# Rate Limiting Configuration
rate_limit:
  enabled: false
  requests_per_minute: 60
  burst: 10

# CORS Configuration
cors:
  enabled: true
  allowed_origins: ["*"]
  allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
  allowed_headers: ["*"]

# MCP (Model Context Protocol) Configuration
mcp:
  enabled: false
  port: 34800
  max_connections: 100
`

	return os.WriteFile(filename, []byte(sampleConfig), 0644)
}
