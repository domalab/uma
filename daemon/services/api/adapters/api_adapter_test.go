package adapters

import (
	"testing"
)

// TestAPIAdapter tests the APIAdapter functionality
func TestAPIAdapter(t *testing.T) {
	t.Run("NewAPIAdapter", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := NewAPIAdapter(mockAPI)

		if adapter == nil {
			t.Error("Expected non-nil APIAdapter")
		}
		if adapter.api != mockAPI {
			t.Error("Expected adapter to store the provided API")
		}
	})

	t.Run("GetInfo", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := NewAPIAdapter(mockAPI)

		info := adapter.GetInfo()
		if info == nil {
			t.Error("Expected non-nil info")
		}

		infoMap, ok := info.(map[string]interface{})
		if !ok {
			t.Error("Expected info to be a map")
		}

		if infoMap["service"] != "UMA API" {
			t.Errorf("Expected service 'UMA API', got '%v'", infoMap["service"])
		}
		if infoMap["version"] != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got '%v'", infoMap["version"])
		}
		if infoMap["status"] != "running" {
			t.Errorf("Expected status 'running', got '%v'", infoMap["status"])
		}
	})

	t.Run("GetSystem", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := NewAPIAdapter(mockAPI)

		system := adapter.GetSystem()
		if system == nil {
			t.Error("Expected non-nil SystemInterface")
		}

		// Test that it returns a SystemAdapter
		systemAdapter, ok := system.(*SystemAdapter)
		if !ok {
			t.Error("Expected SystemAdapter")
		}
		if systemAdapter.api != mockAPI {
			t.Error("Expected SystemAdapter to have the correct API reference")
		}
	})

	t.Run("GetStorage", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := NewAPIAdapter(mockAPI)

		storage := adapter.GetStorage()
		if storage == nil {
			t.Error("Expected non-nil StorageInterface")
		}

		// Test that it returns a StorageAdapter
		storageAdapter, ok := storage.(*StorageAdapter)
		if !ok {
			t.Error("Expected StorageAdapter")
		}
		if storageAdapter.api != mockAPI {
			t.Error("Expected StorageAdapter to have the correct API reference")
		}
	})

	t.Run("GetDocker", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := NewAPIAdapter(mockAPI)

		docker := adapter.GetDocker()
		if docker == nil {
			t.Error("Expected non-nil DockerInterface")
		}

		// Test that it returns a DockerAdapter
		dockerAdapter, ok := docker.(*DockerAdapter)
		if !ok {
			t.Error("Expected DockerAdapter")
		}
		if dockerAdapter.api != mockAPI {
			t.Error("Expected DockerAdapter to have the correct API reference")
		}
	})

	t.Run("GetVM", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := NewAPIAdapter(mockAPI)

		vm := adapter.GetVM()
		if vm == nil {
			t.Error("Expected non-nil VMInterface")
		}

		// Test that it returns a VMAdapter
		vmAdapter, ok := vm.(*VMAdapter)
		if !ok {
			t.Error("Expected VMAdapter")
		}
		if vmAdapter.api != mockAPI {
			t.Error("Expected VMAdapter to have the correct API reference")
		}
	})

	t.Run("GetAuth", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := NewAPIAdapter(mockAPI)

		auth := adapter.GetAuth()
		if auth == nil {
			t.Error("Expected non-nil AuthInterface")
		}

		// Test that it returns an AuthAdapter
		authAdapter, ok := auth.(*AuthAdapter)
		if !ok {
			t.Error("Expected AuthAdapter")
		}
		if authAdapter.api != mockAPI {
			t.Error("Expected AuthAdapter to have the correct API reference")
		}
	})
}

// TestSystemAdapter tests the SystemAdapter functionality
func TestSystemAdapter(t *testing.T) {
	t.Run("GetCPUInfo", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &SystemAdapter{api: mockAPI}

		info, err := adapter.GetCPUInfo()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if info == nil {
			t.Error("Expected non-nil CPU info")
		}
	})

	t.Run("GetMemoryInfo", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &SystemAdapter{api: mockAPI}

		info, err := adapter.GetMemoryInfo()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if info == nil {
			t.Error("Expected non-nil memory info")
		}
	})

	t.Run("GetLoadInfo", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &SystemAdapter{api: mockAPI}

		info, err := adapter.GetLoadInfo()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if info == nil {
			t.Error("Expected non-nil load info")
		}
	})

	t.Run("GetUptimeInfo", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &SystemAdapter{api: mockAPI}

		info, err := adapter.GetUptimeInfo()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if info == nil {
			t.Error("Expected non-nil uptime info")
		}
	})

	t.Run("GetNetworkInfo", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &SystemAdapter{api: mockAPI}

		info, err := adapter.GetNetworkInfo()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if info == nil {
			t.Error("Expected non-nil network info")
		}
	})

	t.Run("GetEnhancedTemperatureData", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &SystemAdapter{api: mockAPI}

		info, err := adapter.GetEnhancedTemperatureData()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if info == nil {
			t.Error("Expected non-nil temperature data")
		}
	})
}

// TestStorageAdapter tests the StorageAdapter functionality
func TestStorageAdapter(t *testing.T) {
	t.Run("GetArrayInfo", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &StorageAdapter{api: mockAPI}

		info, err := adapter.GetArrayInfo()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if info == nil {
			t.Error("Expected non-nil array info")
		}
	})

	t.Run("GetDisks", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &StorageAdapter{api: mockAPI}

		disks, err := adapter.GetDisks()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if disks == nil {
			t.Error("Expected non-nil disks info")
		}
	})

	t.Run("GetZFSPools", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &StorageAdapter{api: mockAPI}

		pools, err := adapter.GetZFSPools()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if pools == nil {
			t.Error("Expected non-nil ZFS pools info")
		}
	})

	t.Run("GetCacheInfo", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &StorageAdapter{api: mockAPI}

		cache, err := adapter.GetCacheInfo()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if cache == nil {
			t.Error("Expected non-nil cache info")
		}
	})

	t.Run("StartArray", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &StorageAdapter{api: mockAPI}

		err := adapter.StartArray(map[string]interface{}{"force": false})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("StopArray", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &StorageAdapter{api: mockAPI}

		err := adapter.StopArray(map[string]interface{}{"force": true})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

// TestDockerAdapter tests the DockerAdapter functionality
func TestDockerAdapter(t *testing.T) {
	t.Run("GetContainers", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &DockerAdapter{api: mockAPI}

		containers, err := adapter.GetContainers()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if containers == nil {
			t.Error("Expected non-nil containers")
		}
	})

	t.Run("GetContainer", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &DockerAdapter{api: mockAPI}

		container, err := adapter.GetContainer("test-id")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if container == nil {
			t.Error("Expected non-nil container")
		}

		containerMap, ok := container.(map[string]interface{})
		if !ok {
			t.Error("Expected container to be a map")
		}
		if containerMap["id"] != "test-id" {
			t.Errorf("Expected id 'test-id', got '%v'", containerMap["id"])
		}
	})

	t.Run("StartContainer", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &DockerAdapter{api: mockAPI}

		err := adapter.StartContainer("test-id")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("StopContainer", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &DockerAdapter{api: mockAPI}

		err := adapter.StopContainer("test-id")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("RestartContainer", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &DockerAdapter{api: mockAPI}

		err := adapter.RestartContainer("test-id")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("GetImages", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &DockerAdapter{api: mockAPI}

		images, err := adapter.GetImages()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if images == nil {
			t.Error("Expected non-nil images")
		}
	})

	t.Run("GetNetworks", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &DockerAdapter{api: mockAPI}

		networks, err := adapter.GetNetworks()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if networks == nil {
			t.Error("Expected non-nil networks")
		}
	})

	t.Run("GetSystemInfo", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &DockerAdapter{api: mockAPI}

		info, err := adapter.GetSystemInfo()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if info == nil {
			t.Error("Expected non-nil system info")
		}
	})
}

// TestVMAdapter tests the VMAdapter functionality
func TestVMAdapter(t *testing.T) {
	t.Run("GetVMs", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &VMAdapter{api: mockAPI}

		vms, err := adapter.GetVMs()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if vms == nil {
			t.Error("Expected non-nil VMs")
		}
	})

	t.Run("GetVM", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &VMAdapter{api: mockAPI}

		vm, err := adapter.GetVM("test-vm")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if vm == nil {
			t.Error("Expected non-nil VM")
		}

		vmMap, ok := vm.(map[string]interface{})
		if !ok {
			t.Error("Expected VM to be a map")
		}
		if vmMap["name"] != "test-vm" {
			t.Errorf("Expected name 'test-vm', got '%v'", vmMap["name"])
		}
	})

	t.Run("StartVM", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &VMAdapter{api: mockAPI}

		err := adapter.StartVM("test-vm")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("StopVM", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &VMAdapter{api: mockAPI}

		err := adapter.StopVM("test-vm")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("RestartVM", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &VMAdapter{api: mockAPI}

		err := adapter.RestartVM("test-vm")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("GetVMStats", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &VMAdapter{api: mockAPI}

		stats, err := adapter.GetVMStats("test-vm")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if stats == nil {
			t.Error("Expected non-nil VM stats")
		}
	})

	t.Run("GetVMConsole", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &VMAdapter{api: mockAPI}

		console, err := adapter.GetVMConsole("test-vm")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if console == nil {
			t.Error("Expected non-nil VM console")
		}
	})

	t.Run("SetVMAutostart", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &VMAdapter{api: mockAPI}

		err := adapter.SetVMAutostart("test-vm", true)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

// TestAuthAdapter tests the AuthAdapter functionality
func TestAuthAdapter(t *testing.T) {
	t.Run("Login", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &AuthAdapter{api: mockAPI}

		result, err := adapter.Login("testuser", "testpass")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Error("Expected non-nil login result")
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Error("Expected login result to be a map")
		}
		if resultMap["access_token"] == nil {
			t.Error("Expected access_token in login result")
		}
	})

	t.Run("GetUsers", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &AuthAdapter{api: mockAPI}

		users, err := adapter.GetUsers()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if users == nil {
			t.Error("Expected non-nil users")
		}
	})

	t.Run("GetStats", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &AuthAdapter{api: mockAPI}

		stats, err := adapter.GetStats()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if stats == nil {
			t.Error("Expected non-nil auth stats")
		}
	})

	t.Run("IsEnabled", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &AuthAdapter{api: mockAPI}

		enabled := adapter.IsEnabled()
		// Should return false for placeholder implementation
		if enabled {
			t.Error("Expected auth to be disabled in placeholder implementation")
		}
	})
}

// MockAPI provides a mock implementation for testing
type MockAPI struct{}

// Add any methods that the adapters might call on the original API
// For now, this is just a placeholder since the adapters return mock data
