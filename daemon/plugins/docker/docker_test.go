package docker

import (
	"strings"
	"testing"
	"time"
)

// MockCommandExecutor provides mock Docker command responses for testing
type MockCommandExecutor struct {
	responses map[string][]string
}

func NewMockCommandExecutor() *MockCommandExecutor {
	return &MockCommandExecutor{
		responses: make(map[string][]string),
	}
}

func (m *MockCommandExecutor) GetCmdOutput(command string, args ...string) []string {
	key := command + " " + strings.Join(args, " ")
	if response, exists := m.responses[key]; exists {
		return response
	}
	// Return empty slice for unknown commands (simulates command not found)
	return []string{}
}

func (m *MockCommandExecutor) SetResponse(command string, args []string, response []string) {
	key := command + " " + strings.Join(args, " ")
	m.responses[key] = response
}

// setupMockDockerManager creates a Docker manager with mocked responses for testing
func setupMockDockerManager() (*DockerManager, *MockCommandExecutor) {
	mockExecutor := NewMockCommandExecutor()

	// Mock Docker availability check
	mockExecutor.SetResponse("docker", []string{"version", "--format", "{{.Server.Version}}"}, []string{"20.10.8"})

	// Mock Docker info
	mockExecutor.SetResponse("docker", []string{"info", "--format", "json"}, []string{
		`{"ServerVersion":"20.10.8","Containers":5,"ContainersRunning":3,"ContainersPaused":0,"ContainersStopped":2,"Images":10}`,
	})

	// Mock container listing
	mockExecutor.SetResponse("docker", []string{"ps", "--format", "json", "--no-trunc", "--all"}, []string{
		`{"ID":"abc123","Names":"test-container","Image":"nginx:latest","Status":"Up 2 hours","State":"running","CreatedAt":"2024-06-21 10:00:00 +0000 UTC"}`,
		`{"ID":"def456","Names":"test-container-2","Image":"redis:latest","Status":"Exited (0) 1 hour ago","State":"exited","CreatedAt":"2024-06-21 09:00:00 +0000 UTC"}`,
	})

	// Mock container inspect
	mockExecutor.SetResponse("docker", []string{"inspect", "test_container"}, []string{
		`[{"Id":"abc123","Name":"/test-container","State":{"Status":"running","Running":true},"Config":{"Image":"nginx:latest"}}]`,
	})

	// Mock inspect for container IDs returned by ps command
	mockExecutor.SetResponse("docker", []string{"inspect", "abc123"}, []string{
		`[{"Id":"abc123","Name":"/test-container","State":{"Status":"running","Running":true},"Config":{"Image":"nginx:latest"},"Created":"2024-06-21T10:00:00.000000000Z"}]`,
	})
	mockExecutor.SetResponse("docker", []string{"inspect", "def456"}, []string{
		`[{"Id":"def456","Name":"/test-container-2","State":{"Status":"exited","Running":false},"Config":{"Image":"redis:latest"},"Created":"2024-06-21T09:00:00.000000000Z"}]`,
	})

	// Mock container operations
	mockExecutor.SetResponse("docker", []string{"start", "test_container"}, []string{"test_container"})
	mockExecutor.SetResponse("docker", []string{"stop", "--time", "10", "test_container"}, []string{"test_container"})
	mockExecutor.SetResponse("docker", []string{"restart", "--time", "10", "test_container"}, []string{"test_container"})

	// Mock image listing
	mockExecutor.SetResponse("docker", []string{"images", "--format", "json", "--no-trunc"}, []string{
		`{"ID":"sha256:abc123","Repository":"nginx","Tag":"latest","Size":133000000,"CreatedAt":"2024-06-20 10:00:00 +0000 UTC"}`,
	})

	// Mock image inspect
	mockExecutor.SetResponse("docker", []string{"image", "inspect", "sha256:abc123"}, []string{
		`[{"Id":"sha256:abc123","RepoTags":["nginx:latest"],"Size":133000000,"Created":"2024-06-20T10:00:00.000000000Z"}]`,
	})

	// Mock network listing
	mockExecutor.SetResponse("docker", []string{"network", "ls", "--format", "json", "--no-trunc"}, []string{
		`{"ID":"net123","Name":"bridge","Driver":"bridge","Scope":"local"}`,
	})

	// Mock network inspect
	mockExecutor.SetResponse("docker", []string{"network", "inspect", "net123"}, []string{
		`[{"Id":"net123","Name":"bridge","Driver":"bridge","Scope":"local","Created":"2024-06-20T10:00:00.000000000Z"}]`,
	})

	manager := NewDockerManagerWithExecutor(mockExecutor)
	return manager, mockExecutor
}

// TestContainerOperations tests pause, unpause, and remove operations
func TestContainerOperations(t *testing.T) {
	mockExecutor := NewMockCommandExecutor()

	// Mock Docker availability check
	mockExecutor.SetResponse("docker", []string{"version", "--format", "{{.Server.Version}}"}, []string{"20.10.8"})

	// Mock pause operation
	mockExecutor.SetResponse("docker", []string{"pause", "test_container"}, []string{"test_container"})

	// Mock unpause operation
	mockExecutor.SetResponse("docker", []string{"unpause", "test_container"}, []string{"test_container"})

	// Mock remove operation
	mockExecutor.SetResponse("docker", []string{"rm", "--force", "test_container"}, []string{"test_container"})

	manager := NewDockerManagerWithExecutor(mockExecutor)

	// Test pause
	err := manager.PauseContainer("test_container")
	if err != nil {
		t.Fatalf("Unexpected error pausing container: %v", err)
	}

	// Test unpause
	err = manager.UnpauseContainer("test_container")
	if err != nil {
		t.Fatalf("Unexpected error unpausing container: %v", err)
	}

	// Test remove
	err = manager.RemoveContainer("test_container", true)
	if err != nil {
		t.Fatalf("Unexpected error removing container: %v", err)
	}
}

// TestContainerLogs tests container log retrieval
func TestContainerLogs(t *testing.T) {
	mockExecutor := NewMockCommandExecutor()

	// Mock Docker availability check
	mockExecutor.SetResponse("docker", []string{"version", "--format", "{{.Server.Version}}"}, []string{"20.10.8"})

	// Mock logs operation
	mockExecutor.SetResponse("docker", []string{"logs", "--tail", "100", "test_container"}, []string{
		"2024-06-21 10:00:00 Starting application",
		"2024-06-21 10:00:01 Application ready",
	})

	manager := NewDockerManagerWithExecutor(mockExecutor)

	logs, err := manager.GetContainerLogs("test_container", 100, false)
	if err != nil {
		t.Fatalf("Unexpected error getting container logs: %v", err)
	}

	if len(logs) != 2 {
		t.Errorf("Expected 2 log lines, got %d", len(logs))
	}
}

// TestContainerStats tests container statistics retrieval
func TestContainerStats(t *testing.T) {
	mockExecutor := NewMockCommandExecutor()

	// Mock Docker availability check
	mockExecutor.SetResponse("docker", []string{"version", "--format", "{{.Server.Version}}"}, []string{"20.10.8"})

	// Mock stats operation
	mockExecutor.SetResponse("docker", []string{"stats", "--no-stream", "--format", "json", "test_container"}, []string{
		`{"ID":"abc123","Name":"test_container","CPUPerc":"5.25%","MemUsage":"256MiB / 2GiB","MemPerc":"12.5%","NetIO":"1.2kB / 2.3kB","BlockIO":"10MB / 5MB","PIDs":"15"}`,
	})

	manager := NewDockerManagerWithExecutor(mockExecutor)

	stats, err := manager.GetContainerStats("test_container")
	if err != nil {
		t.Fatalf("Unexpected error getting container stats: %v", err)
	}

	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}

	if stats.CPUPercent == 0 {
		t.Error("Expected non-zero CPU percentage")
	}
}

// TestNewDockerManager tests the creation of a new Docker manager
func TestNewDockerManager(t *testing.T) {
	manager := NewDockerManager()

	if manager == nil {
		t.Fatal("Expected non-nil Docker manager")
	}

	if manager.cmdExecutor == nil {
		t.Fatal("Expected non-nil command executor")
	}
}

// TestListContainers tests container listing
func TestListContainers(t *testing.T) {
	manager, _ := setupMockDockerManager()

	containers, err := manager.ListContainers(true)

	// Should not error with mock data
	if err != nil {
		t.Fatalf("Unexpected error with mock data: %v", err)
	}

	if len(containers) != 2 {
		t.Errorf("Expected 2 containers, got %d", len(containers))
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
	manager, _ := setupMockDockerManager()

	// Test with a mock container ID
	container, err := manager.GetContainer("test_container")

	// Should work with mock data
	if err != nil {
		t.Fatalf("Unexpected error with mock data: %v", err)
	}

	if container == nil {
		t.Fatal("Expected non-nil container")
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
	manager, _ := setupMockDockerManager()

	// Test with a mock container ID
	err := manager.StartContainer("test_container")

	// Should work with mock data
	if err != nil {
		t.Fatalf("Unexpected error with mock data: %v", err)
	}
}

// TestStopContainer tests container stopping
func TestStopContainer(t *testing.T) {
	manager, _ := setupMockDockerManager()

	// Test with a mock container ID
	err := manager.StopContainer("test_container", 10)

	// Should work with mock data
	if err != nil {
		t.Fatalf("Unexpected error with mock data: %v", err)
	}
}

// TestRestartContainer tests container restarting
func TestRestartContainer(t *testing.T) {
	manager, _ := setupMockDockerManager()

	// Test with a mock container ID
	err := manager.RestartContainer("test_container", 10)

	// Should work with mock data
	if err != nil {
		t.Fatalf("Unexpected error with mock data: %v", err)
	}
}

// TestListImages tests image listing
func TestListImages(t *testing.T) {
	manager, _ := setupMockDockerManager()

	images, err := manager.ListImages()

	// Should work with mock data
	if err != nil {
		t.Fatalf("Unexpected error with mock data: %v", err)
	}

	if len(images) != 1 {
		t.Errorf("Expected 1 image, got %d", len(images))
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
	manager, _ := setupMockDockerManager()

	networks, err := manager.ListNetworks()

	// Should work with mock data
	if err != nil {
		t.Fatalf("Unexpected error with mock data: %v", err)
	}

	if len(networks) != 1 {
		t.Errorf("Expected 1 network, got %d", len(networks))
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
	manager, _ := setupMockDockerManager()

	info, err := manager.GetDockerInfo()

	// Should work with mock data
	if err != nil {
		t.Fatalf("Unexpected error with mock data: %v", err)
	}

	// Validate info structure
	if info == nil {
		t.Fatal("Expected non-nil Docker info")
	}

	// Check for version information
	if version, ok := info["ServerVersion"].(string); !ok || version == "" {
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

// TestDockerAvailability tests Docker availability check
func TestDockerAvailability(t *testing.T) {
	manager, _ := setupMockDockerManager()

	// Test Docker availability check with mock
	available := manager.IsDockerAvailable()

	// Should return true with mock data
	if !available {
		t.Error("Expected Docker to be available with mock data")
	}

	// Test unavailable Docker scenario
	mockExecutor := NewMockCommandExecutor()
	mockExecutor.SetResponse("docker", []string{"version", "--format", "{{.Server.Version}}"}, []string{})
	unavailableManager := NewDockerManagerWithExecutor(mockExecutor)

	unavailable := unavailableManager.IsDockerAvailable()
	if unavailable {
		t.Error("Expected Docker to be unavailable with empty mock response")
	}
}

// TestDockerErrorHandling tests error handling in various scenarios
func TestDockerErrorHandling(t *testing.T) {
	// Create mock executor for error scenarios
	mockExecutor := NewMockCommandExecutor()

	// Mock error responses for invalid operations
	mockExecutor.SetResponse("docker", []string{"start", "invalid_container_xyz"}, []string{"Error: No such container: invalid_container_xyz"})
	mockExecutor.SetResponse("docker", []string{"start", ""}, []string{"Error: container name cannot be empty"})
	mockExecutor.SetResponse("docker", []string{"inspect", "invalid_container_xyz"}, []string{})

	manager := NewDockerManagerWithExecutor(mockExecutor)

	// Test with invalid container ID
	err := manager.StartContainer("invalid_container_xyz")
	if err == nil {
		t.Error("Expected error for invalid container ID")
	}

	// Test with empty container ID
	err = manager.StartContainer("")
	if err == nil {
		t.Error("Expected error for empty container ID")
	}

	// Test container retrieval with invalid ID
	_, err = manager.GetContainer("invalid_container_xyz")
	if err == nil {
		t.Error("Expected error for invalid container ID")
	}
}

// TestDockerDataValidation tests data validation and sanitization
func TestDockerDataValidation(t *testing.T) {
	manager, _ := setupMockDockerManager()

	containers, err := manager.ListContainers(true)
	if err != nil {
		t.Fatalf("Unexpected error with mock data: %v", err)
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
	manager, _ := setupMockDockerManager()

	// Test that operations complete within reasonable time
	start := time.Now()
	_, err := manager.ListContainers(true)
	duration := time.Since(start)

	if duration > 10*time.Second {
		t.Errorf("ListContainers took too long: %v", duration)
	}

	if err != nil {
		t.Fatalf("Unexpected error with mock data: %v", err)
	}

	// Test Docker info performance
	start = time.Now()
	_, err = manager.GetDockerInfo()
	duration = time.Since(start)

	if duration > 5*time.Second {
		t.Errorf("GetDockerInfo took too long: %v", duration)
	}

	if err != nil {
		t.Fatalf("Unexpected error with mock data: %v", err)
	}
}

// BenchmarkListContainers benchmarks container listing
func BenchmarkListContainers(b *testing.B) {
	manager, _ := setupMockDockerManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.ListContainers(true)
		if err != nil {
			b.Fatalf("Unexpected error with mock data: %v", err)
		}
	}
}

// BenchmarkGetDockerInfo benchmarks Docker info retrieval
func BenchmarkGetDockerInfo(b *testing.B) {
	manager, _ := setupMockDockerManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GetDockerInfo()
		if err != nil {
			b.Fatalf("Unexpected error with mock data: %v", err)
		}
	}
}
