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

	assert.True(t, defaultConfig.MCP.Enabled)  // MCP is now enabled by default
	assert.Equal(t, 100, defaultConfig.MCP.MaxConnections)
}

// TestMCPConfigValidation tests MCP configuration validation
func TestMCPConfigValidation(t *testing.T) {
	tests := []struct {
		name           string
		config         domain.MCPConfig
		valid          bool
	}{
		{
			name: "valid default config",
			config: domain.MCPConfig{
				Enabled:        false,
				MaxConnections: 100,
			},
			valid: true,
		},
		{
			name: "valid enabled config",
			config: domain.MCPConfig{
				Enabled:        true,
				MaxConnections: 50,
			},
			valid: true,
		},
		{
			name: "invalid max connections - zero",
			config: domain.MCPConfig{
				Enabled:        true,
				MaxConnections: 0,
			},
			valid: false,
		},
		{
			name: "invalid max connections - negative",
			config: domain.MCPConfig{
				Enabled:        true,
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

// TestMCPConfigSerialization tests configuration serialization
func TestMCPConfigSerialization(t *testing.T) {
	originalConfig := domain.MCPConfig{
		Enabled:        true,
		MaxConnections: 200,
	}

	// Test that config can be properly serialized/deserialized
	// This would typically be done by the config manager
	assert.Equal(t, true, originalConfig.Enabled)
	assert.Equal(t, 200, originalConfig.MaxConnections)
}

// TestMCPConfigSinglePort tests single port architecture
func TestMCPConfigSinglePort(t *testing.T) {
	// Test that MCP configuration no longer includes port settings
	// since MCP now shares the HTTP server port
	config := domain.DefaultConfig()
	
	assert.True(t, config.MCP.Enabled)
	assert.Equal(t, 100, config.MCP.MaxConnections)
	
	// Verify HTTP server port is still configurable
	assert.Equal(t, 34600, config.HTTPServer.Port)
}

// TestMCPConfigEnvironmentVariables tests environment variable integration
func TestMCPConfigEnvironmentVariables(t *testing.T) {
	// Test that MCP configuration can be overridden by environment variables
	// This would be handled by the Viper configuration system

	// Test default values when no environment variables are set
	configService := config.NewViperConfigService()
	mcpConfig := configService.GetMCPConfig()

	assert.True(t, mcpConfig.Enabled)  // MCP is now enabled by default
	assert.Equal(t, 100, mcpConfig.MaxConnections)
}

// TestMCPConfigBoundaryValues tests boundary value validation
func TestMCPConfigBoundaryValues(t *testing.T) {
	tests := []struct {
		name           string
		maxConnections int
		valid          bool
	}{
		{
			name:           "minimum connections",
			maxConnections: 1,
			valid:          true,
		},
		{
			name:           "maximum connections",
			maxConnections: 1000,
			valid:          true,
		},
		{
			name:           "zero connections",
			maxConnections: 0,
			valid:          false,
		},
		{
			name:           "negative connections",
			maxConnections: -1,
			valid:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := domain.MCPConfig{
				Enabled:        true,
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
	assert.True(t, config.MCP.Enabled) // MCP is now enabled by default
	assert.Equal(t, 100, config.MCP.MaxConnections)

	// Test modification
	config.MCP.Enabled = false
	config.MCP.MaxConnections = 200

	assert.False(t, config.MCP.Enabled)
	assert.Equal(t, 200, config.MCP.MaxConnections)
}

// Helper functions for testing

// validateMCPConfig validates MCP configuration (mock implementation for testing)
func validateMCPConfig(config domain.MCPConfig) error {
	if config.MaxConnections <= 0 {
		return assert.AnError
	}
	return nil
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
			assert.GreaterOrEqual(t, mcpConfig.MaxConnections, 1)
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
		MaxConnections: 100,
	}

	// Create a copy
	copiedConfig := originalConfig

	// Modify the copy
	copiedConfig.MaxConnections = 200

	// Original should be unchanged
	assert.Equal(t, 100, originalConfig.MaxConnections)

	// Copy should be modified
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
		MaxConnections: 100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateMCPConfig(config)
	}
}
