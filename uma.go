package main

import (
	"log"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/cskr/pubsub"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/domalab/uma/daemon/cmd"
	"github.com/domalab/uma/daemon/domain"
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

	log.SetOutput(&lumberjack.Logger{
		Filename:   filepath.Join(cli.LogsDir, "uma.log"),
		MaxSize:    1,    // megabytes (reduced from 10MB to 1MB)
		MaxBackups: 3,    // reduced from 10 to 3 backups
		MaxAge:     7,    // days (reduced from 28 to 7 days)
		Compress:   true, // enabled to save space
	})

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
