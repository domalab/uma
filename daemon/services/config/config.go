package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/domalab/uma/daemon/domain"
	"github.com/domalab/uma/daemon/logger"
	"gopkg.in/ini.v1"
)

const (
	DefaultConfigPath = "/boot/config/plugins/uma/uma.json"
	LegacyConfigPath  = "/boot/config/plugins/uma/uma.cfg"
)

// Manager handles configuration loading and saving
type Manager struct {
	configPath string
	config     domain.Config
}

// NewManager creates a new configuration manager
func NewManager(configPath string) *Manager {
	if configPath == "" {
		configPath = DefaultConfigPath
	}

	return &Manager{
		configPath: configPath,
		config:     domain.DefaultConfig(),
	}
}

// Load loads configuration from file
func (m *Manager) Load() error {
	// Try to load JSON config first
	if err := m.loadJSON(); err == nil {
		logger.Blue("Loaded configuration from %s", m.configPath)
		return nil
	}

	// Fall back to legacy INI config
	if err := m.loadLegacyINI(); err == nil {
		logger.Blue("Loaded legacy configuration from %s", LegacyConfigPath)
		// Save as JSON for future use
		if saveErr := m.Save(); saveErr != nil {
			logger.Yellow("Failed to save migrated config: %v", saveErr)
		}
		return nil
	}

	// No config found, use defaults
	logger.Blue("No configuration found, using defaults")
	return m.Save()
}

// Save saves configuration to file
func (m *Manager) Save() error {
	// Ensure directory exists
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write JSON config
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	logger.Blue("Configuration saved to %s", m.configPath)
	return nil
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() domain.Config {
	return m.config
}

// UpdateConfig updates the configuration
func (m *Manager) UpdateConfig(config domain.Config) {
	m.config = config
}

// loadJSON loads configuration from JSON file
func (m *Manager) loadJSON() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &m.config); err != nil {
		return fmt.Errorf("failed to parse JSON config: %w", err)
	}

	// Validate and set defaults for missing fields
	m.validateAndSetDefaults()
	return nil
}

// loadLegacyINI loads configuration from legacy INI file
func (m *Manager) loadLegacyINI() error {
	cfg, err := ini.Load(LegacyConfigPath)
	if err != nil {
		return err
	}

	// Start with defaults
	m.config = domain.DefaultConfig()

	// Parse legacy settings with improved error handling and defaults
	service := cfg.Section("").Key("SERVICE").MustString("disable")
	m.config.HTTPServer.Enabled = (service == "enable")

	return nil
}

// validateAndSetDefaults ensures all config fields have valid values
func (m *Manager) validateAndSetDefaults() {
	defaults := domain.DefaultConfig()

	// Validate HTTP server config
	if m.config.HTTPServer.Port <= 0 || m.config.HTTPServer.Port > 65535 {
		m.config.HTTPServer.Port = defaults.HTTPServer.Port
	}
	if m.config.HTTPServer.Host == "" {
		m.config.HTTPServer.Host = defaults.HTTPServer.Host
	}

	// Validate logging config
	if m.config.Logging.Level == "" {
		m.config.Logging.Level = defaults.Logging.Level
	}
	if m.config.Logging.MaxSize <= 0 {
		m.config.Logging.MaxSize = defaults.Logging.MaxSize
	}
	if m.config.Logging.MaxBackups < 0 {
		m.config.Logging.MaxBackups = defaults.Logging.MaxBackups
	}
	if m.config.Logging.MaxAge < 0 {
		m.config.Logging.MaxAge = defaults.Logging.MaxAge
	}

	// Validate MCP config
	if m.config.MCP.Port <= 0 || m.config.MCP.Port > 65535 {
		m.config.MCP.Port = defaults.MCP.Port
	}
	if m.config.MCP.MaxConnections <= 0 {
		m.config.MCP.MaxConnections = defaults.MCP.MaxConnections
	}
}

// SetHTTPEnabled enables or disables the HTTP server
func (m *Manager) SetHTTPEnabled(enabled bool) error {
	m.config.HTTPServer.Enabled = enabled
	return m.Save()
}

// SetHTTPPort sets the HTTP server port
func (m *Manager) SetHTTPPort(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port number: %d", port)
	}
	m.config.HTTPServer.Port = port
	return m.Save()
}

// GetHTTPPort returns the configured HTTP port
func (m *Manager) GetHTTPPort() int {
	return m.config.HTTPServer.Port
}

// IsHTTPEnabled returns whether HTTP server is enabled
func (m *Manager) IsHTTPEnabled() bool {
	return m.config.HTTPServer.Enabled
}

// SetMCPEnabled enables or disables the MCP server
func (m *Manager) SetMCPEnabled(enabled bool) error {
	m.config.MCP.Enabled = enabled
	return m.Save()
}

// SetMCPPort sets the MCP server port
func (m *Manager) SetMCPPort(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid MCP port number: %d", port)
	}
	m.config.MCP.Port = port
	return m.Save()
}

// SetMCPMaxConnections sets the maximum number of MCP connections
func (m *Manager) SetMCPMaxConnections(maxConnections int) error {
	if maxConnections <= 0 {
		return fmt.Errorf("invalid MCP max connections: %d", maxConnections)
	}
	m.config.MCP.MaxConnections = maxConnections
	return m.Save()
}

// GetMCPPort returns the configured MCP port
func (m *Manager) GetMCPPort() int {
	return m.config.MCP.Port
}

// IsMCPEnabled returns whether MCP server is enabled
func (m *Manager) IsMCPEnabled() bool {
	return m.config.MCP.Enabled
}

// GetMCPMaxConnections returns the maximum number of MCP connections
func (m *Manager) GetMCPMaxConnections() int {
	return m.config.MCP.MaxConnections
}
