package cmd

import (
	"github.com/domalab/uma/daemon/domain"
	"github.com/domalab/uma/daemon/services"
)

type Boot struct{}

func (b *Boot) Run(ctx *domain.Context) error {
	return services.CreateOrchestrator(ctx).Run()
}
