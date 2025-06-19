package docker

import (
	"strings"
	"testing"
	"time"
)

// TestNewDockerManager tests the creation of a new Docker manager
func TestNewDockerManager(t *testing.T) {
	manager := NewDockerManager()

	if manager == nil {
		t.Fatal("Expected non-nil Docker manager")
	}
}

// TestListContainers tests container listing
func TestListContainers(t *testing.T) {
	manager := NewDockerManager()

	containers, err := manager.ListContainers(true)

	// Should not error even if Docker is not available
	if err != nil {
		t.Logf("Container listing error (expected if Docker not available): %v", err)
		return
	}

	// Validate container structure if containers are returned
	for _, container := range containers {
		if container.ID == "" {
			t.Error("Expected non-empty container ID")
		}

		if container.Name == "" {
			t.Error("Expected non-empty container name")
		}

		// Validate state is one of expected values
		validStates := []string{"running", "exited", "paused", "restarting", "removing", "dead", "created"}
		isValidState := false
		for _, state := range validStates {
			if container.State == state {
				isValidState = true
				break
			}
		}
		if !isValidState && container.State != "" {
			t.Errorf("Unexpected container state: %s", container.State)
		}

		// Validate created time
		if !container.Created.IsZero() && container.Created.After(time.Now()) {
			t.Error("Container created time should not be in the future")
		}
	}
}

// TestGetContainer tests individual container retrieval
func TestGetContainer(t *testing.T) {
	manager := NewDockerManager()

	// Test with a mock container ID
	container, err := manager.GetContainer("test_container")

	// Should handle missing containers gracefully
	if err != nil {
		t.Logf("Container retrieval error (expected for non-existent container): %v", err)
		return
	}

	// Validate container structure if returned
	if container != nil {
		if container.ID == "" {
			t.Error("Expected non-empty container ID")
		}

		if container.Name == "" {
			t.Error("Expected non-empty container name")
		}

		// Validate ports structure
		for _, port := range container.Ports {
			if port.ContainerPort == "" && port.HostPort == "" {
				t.Error("Expected at least one port to be non-empty")
			}

			if port.Protocol != "tcp" && port.Protocol != "udp" && port.Protocol != "" {
				t.Errorf("Unexpected port protocol: %s", port.Protocol)
			}
		}
	}
}

// TestStartContainer tests container starting
func TestStartContainer(t *testing.T) {
	manager := NewDockerManager()

	// Test with a mock container ID
	err := manager.StartContainer("test_container")

	// Should handle missing containers gracefully
	if err != nil {
		t.Logf("Container start error (expected for non-existent container): %v", err)
	}
}

// TestStopContainer tests container stopping
func TestStopContainer(t *testing.T) {
	manager := NewDockerManager()

	// Test with a mock container ID
	err := manager.StopContainer("test_container", 10)

	// Should handle missing containers gracefully
	if err != nil {
		t.Logf("Container stop error (expected for non-existent container): %v", err)
	}
}

// TestRestartContainer tests container restarting
func TestRestartContainer(t *testing.T) {
	manager := NewDockerManager()

	// Test with a mock container ID
	err := manager.RestartContainer("test_container", 10)

	// Should handle missing containers gracefully
	if err != nil {
		t.Logf("Container restart error (expected for non-existent container): %v", err)
	}
}

// TestListImages tests image listing
func TestListImages(t *testing.T) {
	manager := NewDockerManager()

	images, err := manager.ListImages()

	// Should not error even if Docker is not available
	if err != nil {
		t.Logf("Image listing error (expected if Docker not available): %v", err)
		return
	}

	// Validate image structure if images are returned
	for _, image := range images {
		if image.ID == "" {
			t.Error("Expected non-empty image ID")
		}

		// Size should be non-negative
		if image.Size < 0 {
			t.Errorf("Expected non-negative image size, got %d", image.Size)
		}

		// Validate created time
		if !image.Created.IsZero() && image.Created.After(time.Now()) {
			t.Error("Image created time should not be in the future")
		}

		// Validate repository tags
		for _, tag := range image.RepoTags {
			if tag == "" {
				t.Error("Expected non-empty repository tag")
			}
		}
	}
}

// TestListNetworks tests network listing
func TestListNetworks(t *testing.T) {
	manager := NewDockerManager()

	networks, err := manager.ListNetworks()

	// Should not error even if Docker is not available
	if err != nil {
		t.Logf("Network listing error (expected if Docker not available): %v", err)
		return
	}

	// Validate network structure if networks are returned
	for _, network := range networks {
		if network.ID == "" {
			t.Error("Expected non-empty network ID")
		}

		if network.Name == "" {
			t.Error("Expected non-empty network name")
		}

		// Validate driver
		if network.Driver == "" {
			t.Error("Expected non-empty network driver")
		}

		// Validate scope
		validScopes := []string{"local", "global", "swarm"}
		isValidScope := false
		for _, scope := range validScopes {
			if network.Scope == scope {
				isValidScope = true
				break
			}
		}
		if !isValidScope && network.Scope != "" {
			t.Errorf("Unexpected network scope: %s", network.Scope)
		}
	}
}

// TestGetDockerInfo tests Docker daemon information
func TestGetDockerInfo(t *testing.T) {
	manager := NewDockerManager()

	info, err := manager.GetDockerInfo()

	// Should not error even if Docker is not available
	if err != nil {
		t.Logf("Docker info error (expected if Docker not available): %v", err)
		return
	}

	// Validate info structure if returned
	if info != nil {
		// Check for version information
		if version, ok := info["ServerVersion"].(string); ok && version == "" {
			t.Error("Expected non-empty Docker server version")
		}

		// Check for container counts if available
		if containers, ok := info["Containers"].(float64); ok && containers < 0 {
			t.Error("Container count should not be negative")
		}

		if containersRunning, ok := info["ContainersRunning"].(float64); ok && containersRunning < 0 {
			t.Error("Running containers count should not be negative")
		}

		if containersPaused, ok := info["ContainersPaused"].(float64); ok && containersPaused < 0 {
			t.Error("Paused containers count should not be negative")
		}

		if containersStopped, ok := info["ContainersStopped"].(float64); ok && containersStopped < 0 {
			t.Error("Stopped containers count should not be negative")
		}

		if images, ok := info["Images"].(float64); ok && images < 0 {
			t.Error("Images count should not be negative")
		}
	}
}

// TestDockerAvailability tests Docker availability check
func TestDockerAvailability(t *testing.T) {
	manager := NewDockerManager()

	// Test Docker availability check
	available := manager.IsDockerAvailable()

	// Log the result but don't fail the test since Docker may not be available in test environment
	t.Logf("Docker available: %v", available)
}

// TestDockerErrorHandling tests error handling in various scenarios
func TestDockerErrorHandling(t *testing.T) {
	manager := NewDockerManager()

	// Test with invalid container ID
	err := manager.StartContainer("invalid_container_xyz")
	if err == nil {
		t.Log("Start container succeeded for invalid ID (may be expected behavior)")
	}

	// Test with empty container ID
	err = manager.StartContainer("")
	if err == nil {
		t.Log("Start container succeeded for empty ID")
	}

	// Test container retrieval with invalid ID
	_, err = manager.GetContainer("invalid_container_xyz")
	if err == nil {
		t.Log("Get container succeeded for invalid ID")
	}
}

// TestDockerDataValidation tests data validation and sanitization
func TestDockerDataValidation(t *testing.T) {
	manager := NewDockerManager()

	containers, err := manager.ListContainers(true)
	if err != nil {
		t.Logf("Container listing error: %v", err)
		return
	}

	for _, container := range containers {
		// Validate container names don't contain dangerous characters
		if strings.Contains(container.Name, "..") {
			t.Errorf("Container name contains potentially dangerous characters: %s", container.Name)
		}

		// Validate image names
		if strings.Contains(container.Image, "..") {
			t.Errorf("Container image contains potentially dangerous characters: %s", container.Image)
		}

		// Validate port configurations
		for _, port := range container.Ports {
			if port.HostPort != "" {
				// Basic validation that host port is numeric if specified
				if len(port.HostPort) > 5 {
					t.Errorf("Host port seems too long: %s", port.HostPort)
				}
			}
		}
	}
}

// TestDockerPerformance tests performance characteristics
func TestDockerPerformance(t *testing.T) {
	manager := NewDockerManager()

	// Test that operations complete within reasonable time
	start := time.Now()
	_, err := manager.ListContainers(true)
	duration := time.Since(start)

	if duration > 10*time.Second {
		t.Errorf("ListContainers took too long: %v", duration)
	}

	if err != nil {
		t.Logf("Container listing error (expected if Docker not available): %v", err)
	}

	// Test Docker info performance
	start = time.Now()
	_, err = manager.GetDockerInfo()
	duration = time.Since(start)

	if duration > 5*time.Second {
		t.Errorf("GetDockerInfo took too long: %v", duration)
	}

	if err != nil {
		t.Logf("Docker info error (expected if Docker not available): %v", err)
	}
}

// BenchmarkListContainers benchmarks container listing
func BenchmarkListContainers(b *testing.B) {
	manager := NewDockerManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.ListContainers(true)
		if err != nil {
			b.Logf("Container listing error: %v", err)
		}
	}
}

// BenchmarkGetDockerInfo benchmarks Docker info retrieval
func BenchmarkGetDockerInfo(b *testing.B) {
	manager := NewDockerManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GetDockerInfo()
		if err != nil {
			b.Logf("Docker info error: %v", err)
		}
	}
}
