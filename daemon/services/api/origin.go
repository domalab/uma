package api

import (
	"github.com/domalab/omniraid/daemon/dto"
	"github.com/domalab/omniraid/daemon/lib"
	"github.com/domalab/omniraid/daemon/logger"
)

func (a *Api) getOrigin() *dto.Origin {
	if a.origin == nil {
		origin, err := lib.GetOrigin()
		if err != nil {
			logger.Yellow(" unable to get origin: %s", err)
		}
		a.origin = origin
	}

	return a.origin
}
