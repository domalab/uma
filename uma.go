package main

import (
	"log"

	"github.com/alecthomas/kong"
	"github.com/cskr/pubsub"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/domalab/uma/daemon/cmd"
	"github.com/domalab/uma/daemon/domain"
	"github.com/domalab/uma/daemon/logger"
)

var Version string

// Sentry functions temporarily disabled for testing

var cli struct {
	LogsDir    string `default:"/var/log" help:"directory to store logs"`
	ConfigPath string `default:"" help:"path to configuration file"`
	HTTPPort   int    `default:"34600" help:"HTTP API server port"`

	Boot   cmd.Boot      `cmd:"" default:"1" help:"start processing"`
	Config cmd.ConfigCmd `cmd:"" help:"manage configuration"`
}

func main() {
	ctx := kong.Parse(&cli)

	// Initialize Sentry for production error monitoring
	// initializeSentry()
	// defer sentry.Flush(2 * time.Second)

	// Clean up any existing backup log files first
	if err := logger.CleanupOldLogFiles(cli.LogsDir); err != nil {
		log.Printf("Warning: failed to cleanup old log files: %v", err)
	}

	// Setup optimized file logging for Unraid systems
	logConfig := logger.UnraidOptimizedConfig(cli.LogsDir)
	if err := logger.ValidateLogConfiguration(logConfig); err != nil {
		log.Fatalf("Invalid log configuration: %v", err)
	}

	log.SetOutput(&lumberjack.Logger{
		Filename:   logConfig.Filename,
		MaxSize:    logConfig.MaxSize,    // 10MB limit as requested
		MaxBackups: logConfig.MaxBackups, // 0 - DISABLED to prevent disk space issues
		MaxAge:     logConfig.MaxAge,     // 0 - DISABLED for minimal disk usage
		Compress:   logConfig.Compress,   // false - DISABLED to avoid backup files
	})

	// Log disk usage information for monitoring
	logger.LogDiskUsageInfo(cli.LogsDir)

	// Create base configuration
	config := domain.DefaultConfig()
	config.Version = Version

	// Override HTTP port if specified
	if cli.HTTPPort != 34600 {
		config.HTTPServer.Port = cli.HTTPPort
	}

	err := ctx.Run(&domain.Context{
		Config: config,
		Hub:    pubsub.New(623),
	})
	ctx.FatalIfErrorf(err)
}
