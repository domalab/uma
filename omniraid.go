package main

import (
	"log"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/cskr/pubsub"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/domalab/omniraid/daemon/cmd"
	"github.com/domalab/omniraid/daemon/domain"
)

var Version string

var cli struct {
	LogsDir    string `default:"/var/log" help:"directory to store logs"`
	ConfigPath string `default:"" help:"path to configuration file"`
	HTTPPort   int    `default:"34600" help:"HTTP API server port"`
	ShowUps    bool   `env:"SHOW_UPS" default:"false" help:"whether to provide ups status or not"`

	Boot   cmd.Boot      `cmd:"" default:"1" help:"start processing"`
	Config cmd.ConfigCmd `cmd:"" help:"manage configuration"`
}

func main() {
	ctx := kong.Parse(&cli)

	log.SetOutput(&lumberjack.Logger{
		Filename:   filepath.Join(cli.LogsDir, "omniraid.log"),
		MaxSize:    10, // megabytes
		MaxBackups: 10,
		MaxAge:     28, // days
		// Compress:   true, // disabled by default
	})

	// Create base configuration
	config := domain.DefaultConfig()
	config.Version = Version
	config.ShowUps = cli.ShowUps

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
