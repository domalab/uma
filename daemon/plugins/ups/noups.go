package ups

import "github.com/domalab/uma/daemon/dto"

// NoUps -
type NoUps struct {
	samples []dto.Sample
}

// NewNoUps -
func NewNoUps() *NoUps {
	noups := &NoUps{
		samples: make([]dto.Sample, 0),
	}
	return noups
}

// GetStatus -
func (n *NoUps) GetStatus() []dto.Sample {
	return n.samples
}
