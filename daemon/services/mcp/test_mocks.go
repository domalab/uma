package mcp

import (
	"github.com/domalab/uma/daemon/services/api/utils"
)

// MockAPIInterface provides a mock implementation for testing
type MockAPIInterface struct{}

func (m *MockAPIInterface) GetInfo() interface{} {
	return map[string]interface{}{
		"version": "test-version",
		"status":  "healthy",
	}
}

func (m *MockAPIInterface) GetSystem() utils.SystemInterface   { return nil }
func (m *MockAPIInterface) GetStorage() utils.StorageInterface { return nil }
func (m *MockAPIInterface) GetDocker() utils.DockerInterface   { return nil }
func (m *MockAPIInterface) GetVM() utils.VMInterface           { return nil }
func (m *MockAPIInterface) GetAuth() utils.AuthInterface       { return nil }

func (m *MockAPIInterface) GetNotifications() utils.NotificationInterface { return nil }
func (m *MockAPIInterface) GetUPSDetector() utils.UPSDetectorInterface    { return nil }
