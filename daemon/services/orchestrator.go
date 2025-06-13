package services

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/domalab/omniraid/daemon/domain"
	"github.com/domalab/omniraid/daemon/logger"
	"github.com/domalab/omniraid/daemon/services/api"
)

type Orchestrator struct {
	ctx *domain.Context
}

func CreateOrchestrator(ctx *domain.Context) *Orchestrator {
	return &Orchestrator{
		ctx: ctx,
	}
}

func (o *Orchestrator) Run() error {
	logger.Blue("starting omniraid %s ...", o.ctx.Config.Version)

	apiService := api.Create(o.ctx)

	err := apiService.Run()
	if err != nil {
		return err
	}

	// Wait for shutdown signal
	w := make(chan os.Signal, 1)
	signal.Notify(w, syscall.SIGTERM, syscall.SIGINT)
	sig := <-w
	logger.Blue("received %s signal. shutting down the app ...", sig)

	// Graceful shutdown
	if err := apiService.Stop(); err != nil {
		logger.Yellow("Error during shutdown: %v", err)
	}

	logger.Blue("omniraid shutdown complete")
	return nil
}
