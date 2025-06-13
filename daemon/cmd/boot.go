package cmd

import (
	"github.com/domalab/omniraid/daemon/domain"
	"github.com/domalab/omniraid/daemon/services"
)

type Boot struct{}

func (b *Boot) Run(ctx *domain.Context) error {
	return services.CreateOrchestrator(ctx).Run()
}
