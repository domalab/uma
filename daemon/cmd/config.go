package cmd

import (
	"fmt"

	"github.com/domalab/uma/daemon/domain"
	"github.com/domalab/uma/daemon/services/auth"
	"github.com/domalab/uma/daemon/services/config"
)

// ConfigCmd handles configuration management commands
type ConfigCmd struct {
	Show     ConfigShowCmd     `cmd:"" help:"Show current configuration"`
	Set      ConfigSetCmd      `cmd:"" help:"Set configuration values"`
	Generate ConfigGenerateCmd `cmd:"" help:"Generate configuration values"`
}

// ConfigShowCmd shows the current configuration
type ConfigShowCmd struct{}

func (c *ConfigShowCmd) Run(ctx *domain.Context) error {
	manager := config.NewManager("")
	if err := manager.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := manager.GetConfig()

	fmt.Printf("UMA Configuration:\n")
	fmt.Printf("  Version: %s\n", cfg.Version)
	fmt.Printf("\n")
	fmt.Printf("HTTP Server:\n")
	fmt.Printf("  Enabled: %t\n", cfg.HTTPServer.Enabled)
	fmt.Printf("  Host: %s\n", cfg.HTTPServer.Host)
	fmt.Printf("  Port: %d\n", cfg.HTTPServer.Port)
	fmt.Printf("\n")
	fmt.Printf("Authentication:\n")
	fmt.Printf("  Enabled: %t\n", cfg.Auth.Enabled)
	if cfg.Auth.APIKey != "" {
		fmt.Printf("  API Key: %s...\n", cfg.Auth.APIKey[:8])
	} else {
		fmt.Printf("  API Key: (not set)\n")
	}
	fmt.Printf("\n")
	fmt.Printf("Logging:\n")
	fmt.Printf("  Level: %s\n", cfg.Logging.Level)
	fmt.Printf("  Max Size: %d MB\n", cfg.Logging.MaxSize)
	fmt.Printf("  Max Backups: %d\n", cfg.Logging.MaxBackups)
	fmt.Printf("  Max Age: %d days\n", cfg.Logging.MaxAge)

	return nil
}

// ConfigSetCmd sets configuration values
type ConfigSetCmd struct {
	HTTPEnabled *bool   `help:"Enable/disable HTTP server"`
	Port        *int    `name:"port" help:"Set HTTP server port"`
	AuthEnabled *bool   `help:"Enable/disable authentication"`
	APIKey      *string `help:"Set API key"`
	LogLevel    *string `help:"Set log level"`
}

func (c *ConfigSetCmd) Run(ctx *domain.Context) error {
	manager := config.NewManager("")
	if err := manager.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := manager.GetConfig()
	changed := false

	if c.HTTPEnabled != nil {
		cfg.HTTPServer.Enabled = *c.HTTPEnabled
		changed = true
		fmt.Printf("HTTP server enabled: %t\n", *c.HTTPEnabled)
	}

	if c.Port != nil {
		if *c.Port <= 0 || *c.Port > 65535 {
			return fmt.Errorf("invalid port number: %d", *c.Port)
		}
		cfg.HTTPServer.Port = *c.Port
		changed = true
		fmt.Printf("HTTP server port: %d\n", *c.Port)
	}

	if c.AuthEnabled != nil {
		cfg.Auth.Enabled = *c.AuthEnabled
		changed = true
		fmt.Printf("Authentication enabled: %t\n", *c.AuthEnabled)
	}

	if c.APIKey != nil {
		cfg.Auth.APIKey = *c.APIKey
		changed = true
		fmt.Printf("API key updated\n")
	}

	if c.LogLevel != nil {
		cfg.Logging.Level = *c.LogLevel
		changed = true
		fmt.Printf("Log level: %s\n", *c.LogLevel)
	}

	if !changed {
		return fmt.Errorf("no configuration changes specified")
	}

	manager.UpdateConfig(cfg)
	if err := manager.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Configuration saved successfully\n")
	return nil
}

// ConfigGenerateCmd generates configuration values
type ConfigGenerateCmd struct {
	APIKey bool `help:"Generate a new API key"`
}

func (c *ConfigGenerateCmd) Run(ctx *domain.Context) error {
	if c.APIKey {
		authService := auth.NewAuthService(domain.AuthConfig{})
		apiKey, err := authService.GenerateAPIKey()
		if err != nil {
			return fmt.Errorf("failed to generate API key: %w", err)
		}

		fmt.Printf("Generated API key: %s\n", apiKey)
		fmt.Printf("\nTo set this as the active API key, run:\n")
		fmt.Printf("  uma config set --api-key=%s\n", apiKey)
		return nil
	}

	return fmt.Errorf("no generation option specified")
}
