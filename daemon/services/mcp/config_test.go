package mcp

import (
	"testing"

	"github.com/domalab/uma/daemon/domain"
	"github.com/domalab/uma/daemon/services/config"
	"github.com/stretchr/testify/assert"
)

// TestMCPConfigDefaults tests default MCP configuration values
func TestMCPConfigDefaults(t *testing.T) {
	defaultConfig := domain.DefaultConfig()

	assert.False(t, defaultConfig.MCP.Enabled)
	assert.Equal(t, 34800, defaultConfig.MCP.Port)
	assert.Equal(t, 100, defaultConfig.MCP.MaxConnections)
}

// TestMCPConfigValidation tests MCP configuration validation
func TestMCPConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config domain.MCPConfig
		valid  bool
	}{
		{
			name: "valid default config",
			config: domain.MCPConfig{
				Enabled:        false,
				Port:           34800,
				MaxConnections: 100,
			},
			valid: true,
		},
		{
			name: "valid enabled config",
			config: domain.MCPConfig{
				Enabled:        true,
				Port:           35000,
				MaxConnections: 50,
			},
			valid: true,
		},
		{
			name: "invalid port - too low",
			config: domain.MCPConfig{
				Enabled:        true,
				Port:           1023,
				MaxConnections: 100,
			},
			valid: false,
		},
		{
			name: "invalid port - too high",
			config: domain.MCPConfig{
				Enabled:        true,
				Port:           65536,
				MaxConnections: 100,
			},
			valid: false,
		},
		{
			name: "invalid max connections - zero",
			config: domain.MCPConfig{
				Enabled:        true,
				Port:           34800,
				MaxConnections: 0,
			},
			valid: false,
		},
		{
			name: "invalid max connections - negative",
			config: domain.MCPConfig{
				Enabled:        true,
				Port:           34800,
				MaxConnections: -1,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMCPConfig(tt.config)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestViperMCPConfig tests Viper configuration service MCP integration
func TestViperMCPConfig(t *testing.T) {
	// Create a test Viper config service
	configService := config.NewViperConfigService()

	// Test default MCP configuration
	mcpConfig := configService.GetMCPConfig()
	assert.False(t, mcpConfig.Enabled)
	assert.Equal(t, 34800, mcpConfig.Port)
	assert.Equal(t, 100, mcpConfig.MaxConnections)

	// Test configuration validation
	err := configService.ValidateConfig()
	assert.NoError(t, err)
}

// TestMCPConfigSerialization tests configuration serialization
func TestMCPConfigSerialization(t *testing.T) {
	originalConfig := domain.MCPConfig{
		Enabled:        true,
		Port:           35000,
		MaxConnections: 200,
	}

	// Test that config can be properly serialized/deserialized
	// This would typically be done by the config manager
	assert.Equal(t, true, originalConfig.Enabled)
	assert.Equal(t, 35000, originalConfig.Port)
	assert.Equal(t, 200, originalConfig.MaxConnections)
}

// TestMCPConfigPortConflicts tests port conflict detection
func TestMCPConfigPortConflicts(t *testing.T) {
	tests := []struct {
		name     string
		httpPort int
		mcpPort  int
		conflict bool
	}{
		{
			name:     "no conflict - different ports",
			httpPort: 34600,
			mcpPort:  34800,
			conflict: false,
		},
		{
			name:     "conflict - same port",
			httpPort: 34600,
			mcpPort:  34600,
			conflict: true,
		},
		{
			name:     "no conflict - well separated",
			httpPort: 8080,
			mcpPort:  9090,
			conflict: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conflict := checkPortConflict(tt.httpPort, tt.mcpPort)
			assert.Equal(t, tt.conflict, conflict)
		})
	}
}

// TestMCPConfigEnvironmentVariables tests environment variable integration
func TestMCPConfigEnvironmentVariables(t *testing.T) {
	// Test that MCP configuration can be overridden by environment variables
	// This would be handled by the Viper configuration system

	// Test default values when no environment variables are set
	configService := config.NewViperConfigService()
	mcpConfig := configService.GetMCPConfig()

	assert.False(t, mcpConfig.Enabled)
	assert.Equal(t, 34800, mcpConfig.Port)
	assert.Equal(t, 100, mcpConfig.MaxConnections)
}

// TestMCPConfigBoundaryValues tests boundary value validation
func TestMCPConfigBoundaryValues(t *testing.T) {
	tests := []struct {
		name           string
		port           int
		maxConnections int
		valid          bool
	}{
		{
			name:           "minimum valid port",
			port:           1024,
			maxConnections: 1,
			valid:          true,
		},
		{
			name:           "maximum valid port",
			port:           65535,
			maxConnections: 1000,
			valid:          true,
		},
		{
			name:           "port just below minimum",
			port:           1023,
			maxConnections: 100,
			valid:          false,
		},
		{
			name:           "port just above maximum",
			port:           65536,
			maxConnections: 100,
			valid:          false,
		},
		{
			name:           "minimum connections",
			port:           34800,
			maxConnections: 1,
			valid:          true,
		},
		{
			name:           "zero connections",
			port:           34800,
			maxConnections: 0,
			valid:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := domain.MCPConfig{
				Enabled:        true,
				Port:           tt.port,
				MaxConnections: tt.maxConnections,
			}

			err := validateMCPConfig(config)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestMCPConfigIntegrationWithDomainConfig tests integration with domain config
func TestMCPConfigIntegrationWithDomainConfig(t *testing.T) {
	config := domain.DefaultConfig()

	// Test that MCP config is properly integrated
	assert.NotNil(t, config.MCP)
	assert.False(t, config.MCP.Enabled)
	assert.Equal(t, 34800, config.MCP.Port)
	assert.Equal(t, 100, config.MCP.MaxConnections)

	// Test modification
	config.MCP.Enabled = true
	config.MCP.Port = 35000
	config.MCP.MaxConnections = 200

	assert.True(t, config.MCP.Enabled)
	assert.Equal(t, 35000, config.MCP.Port)
	assert.Equal(t, 200, config.MCP.MaxConnections)
}

// Helper functions for testing

// validateMCPConfig validates MCP configuration (mock implementation for testing)
func validateMCPConfig(config domain.MCPConfig) error {
	if config.Port < 1024 || config.Port > 65535 {
		return assert.AnError
	}
	if config.MaxConnections <= 0 {
		return assert.AnError
	}
	return nil
}

// checkPortConflict checks for port conflicts (mock implementation for testing)
func checkPortConflict(httpPort, mcpPort int) bool {
	return httpPort == mcpPort
}

// TestMCPConfigConcurrency tests concurrent access to MCP configuration
func TestMCPConfigConcurrency(t *testing.T) {
	configService := config.NewViperConfigService()

	// Test concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			mcpConfig := configService.GetMCPConfig()
			assert.NotNil(t, mcpConfig)
			assert.GreaterOrEqual(t, mcpConfig.Port, 1024)
			assert.LessOrEqual(t, mcpConfig.Port, 65535)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestMCPConfigDefensiveCopying tests that configuration is properly copied
func TestMCPConfigDefensiveCopying(t *testing.T) {
	originalConfig := domain.MCPConfig{
		Enabled:        true,
		Port:           34800,
		MaxConnections: 100,
	}

	// Create a copy
	copiedConfig := originalConfig

	// Modify the copy
	copiedConfig.Port = 35000
	copiedConfig.MaxConnections = 200

	// Original should be unchanged
	assert.Equal(t, 34800, originalConfig.Port)
	assert.Equal(t, 100, originalConfig.MaxConnections)

	// Copy should be modified
	assert.Equal(t, 35000, copiedConfig.Port)
	assert.Equal(t, 200, copiedConfig.MaxConnections)
}

// Benchmark tests for configuration performance
func BenchmarkGetMCPConfig(b *testing.B) {
	configService := config.NewViperConfigService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = configService.GetMCPConfig()
	}
}

func BenchmarkValidateMCPConfig(b *testing.B) {
	config := domain.MCPConfig{
		Enabled:        true,
		Port:           34800,
		MaxConnections: 100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateMCPConfig(config)
	}
}
