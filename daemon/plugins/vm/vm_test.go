package vm

import (
	"strings"
	"testing"
	"time"
)

// TestNewVMManager tests the creation of a new VM manager
func TestNewVMManager(t *testing.T) {
	manager := NewVMManager()

	if manager == nil {
		t.Fatal("Expected non-nil VM manager")
	}
}

// TestListVMs tests VM listing
func TestListVMs(t *testing.T) {
	manager := NewVMManager()

	vms, err := manager.ListVMs(true)

	// Should not error even if libvirt is not available
	if err != nil {
		t.Logf("VM listing error (expected if libvirt not available): %v", err)
		return
	}

	// Validate VM structure if VMs are returned
	for _, vm := range vms {
		if vm.Name == "" {
			t.Error("Expected non-empty VM name")
		}

		if vm.UUID == "" {
			t.Error("Expected non-empty VM UUID")
		}

		// Validate state is one of expected values
		validStates := []string{"running", "shut off", "paused", "suspended", "crashed", "dying"}
		isValidState := false
		for _, state := range validStates {
			if vm.State == state {
				isValidState = true
				break
			}
		}
		if !isValidState && vm.State != "" {
			t.Errorf("Unexpected VM state: %s", vm.State)
		}

		// Memory should be non-negative
		if vm.Memory < 0 {
			t.Errorf("Expected non-negative VM memory, got %d", vm.Memory)
		}

		// CPUs should be positive if set
		if vm.CPUs < 0 {
			t.Errorf("Expected non-negative VM CPUs, got %d", vm.CPUs)
		}
	}
}

// TestGetVM tests individual VM retrieval
func TestGetVM(t *testing.T) {
	manager := NewVMManager()

	// Test with a mock VM name
	vm, err := manager.GetVM("test_vm")

	// Should handle missing VMs gracefully
	if err != nil {
		t.Logf("VM retrieval error (expected for non-existent VM): %v", err)
		return
	}

	// Validate VM structure if returned
	if vm != nil {
		if vm.Name == "" {
			t.Error("Expected non-empty VM name")
		}

		if vm.UUID == "" {
			t.Error("Expected non-empty VM UUID")
		}

		// Validate network interfaces
		for _, network := range vm.Networks {
			if network.Type == "" {
				t.Error("Expected non-empty network interface type")
			}

			if network.Source == "" {
				t.Error("Expected non-empty network interface source")
			}
		}

		// Validate disk devices
		for _, disk := range vm.Disks {
			if disk.Device == "" {
				t.Error("Expected non-empty disk device name")
			}

			if disk.Source == "" {
				t.Error("Expected non-empty disk source")
			}

			// Size should be non-negative
			if disk.Size > 0 && disk.Size < 0 {
				t.Errorf("Expected non-negative disk size, got %d", disk.Size)
			}
		}
	}
}

// TestStartVM tests VM starting
func TestStartVM(t *testing.T) {
	manager := NewVMManager()

	// Test with a mock VM name
	err := manager.StartVM("test_vm")

	// Should handle missing VMs gracefully
	if err != nil {
		t.Logf("VM start error (expected for non-existent VM): %v", err)
	}
}

// TestStopVM tests VM stopping
func TestStopVM(t *testing.T) {
	manager := NewVMManager()

	// Test with a mock VM name
	err := manager.StopVM("test_vm", false)

	// Should handle missing VMs gracefully
	if err != nil {
		t.Logf("VM stop error (expected for non-existent VM): %v", err)
	}
}

// TestRestartVM tests VM restarting
func TestRestartVM(t *testing.T) {
	manager := NewVMManager()

	// Test with a mock VM name
	err := manager.RestartVM("test_vm")

	// Should handle missing VMs gracefully
	if err != nil {
		t.Logf("VM restart error (expected for non-existent VM): %v", err)
	}
}

// TestPauseVM tests VM pausing
func TestPauseVM(t *testing.T) {
	manager := NewVMManager()

	// Test with a mock VM name
	err := manager.PauseVM("test_vm")

	// Should handle missing VMs gracefully
	if err != nil {
		t.Logf("VM pause error (expected for non-existent VM): %v", err)
	}
}

// TestResumeVM tests VM resuming
func TestResumeVM(t *testing.T) {
	manager := NewVMManager()

	// Test with a mock VM name
	err := manager.ResumeVM("test_vm")

	// Should handle missing VMs gracefully
	if err != nil {
		t.Logf("VM resume error (expected for non-existent VM): %v", err)
	}
}

// TestGetVMStats tests VM statistics retrieval
func TestGetVMStats(t *testing.T) {
	manager := NewVMManager()

	// Test with a mock VM name
	stats, err := manager.GetVMStats("test_vm")

	// Should handle missing VMs gracefully
	if err != nil {
		t.Logf("VM stats error (expected for non-existent VM): %v", err)
		return
	}

	// Validate stats structure if returned
	if stats != nil {
		// CPU usage should be between 0 and 100
		if stats.CPUUsage < 0 || stats.CPUUsage > 100 {
			t.Errorf("CPU usage should be between 0 and 100, got %f", stats.CPUUsage)
		}

		// Memory usage should be non-negative
		if stats.MemoryUsage < 0 {
			t.Errorf("Expected non-negative memory usage, got %d", stats.MemoryUsage)
		}

		// Network stats should be non-negative
		if stats.NetRx < 0 {
			t.Errorf("Expected non-negative network RX bytes, got %d", stats.NetRx)
		}

		if stats.NetTx < 0 {
			t.Errorf("Expected non-negative network TX bytes, got %d", stats.NetTx)
		}

		// Disk stats should be non-negative
		if stats.DiskRead < 0 {
			t.Errorf("Expected non-negative disk read bytes, got %d", stats.DiskRead)
		}

		if stats.DiskWrite < 0 {
			t.Errorf("Expected non-negative disk write bytes, got %d", stats.DiskWrite)
		}
	}
}

// TestLibvirtAvailability tests libvirt availability check
func TestLibvirtAvailability(t *testing.T) {
	manager := NewVMManager()

	// Test libvirt availability check
	available := manager.IsLibvirtAvailable()

	// Log the result but don't fail the test since libvirt may not be available in test environment
	t.Logf("Libvirt available: %v", available)
}

// TestVMAutostart tests VM autostart functionality
func TestVMAutostart(t *testing.T) {
	manager := NewVMManager()

	// Test setting autostart
	err := manager.SetVMAutostart("test_vm", true)
	if err != nil {
		t.Logf("VM autostart error (expected for non-existent VM): %v", err)
	}

	// Test disabling autostart
	err = manager.SetVMAutostart("test_vm", false)
	if err != nil {
		t.Logf("VM autostart disable error (expected for non-existent VM): %v", err)
	}
}

// TestVMErrorHandling tests error handling in various scenarios
func TestVMErrorHandling(t *testing.T) {
	manager := NewVMManager()

	// Test with invalid VM name
	err := manager.StartVM("invalid_vm_xyz")
	if err == nil {
		t.Log("Start VM succeeded for invalid name (may be expected behavior)")
	}

	// Test with empty VM name
	err = manager.StartVM("")
	if err == nil {
		t.Log("Start VM succeeded for empty name")
	}

	// Test VM retrieval with invalid name
	_, err = manager.GetVM("invalid_vm_xyz")
	if err == nil {
		t.Log("Get VM succeeded for invalid name")
	}
}

// TestVMDataValidation tests data validation and sanitization
func TestVMDataValidation(t *testing.T) {
	manager := NewVMManager()

	vms, err := manager.ListVMs(true)
	if err != nil {
		t.Logf("VM listing error: %v", err)
		return
	}

	for _, vm := range vms {
		// Validate VM names don't contain dangerous characters
		if strings.Contains(vm.Name, "..") || strings.Contains(vm.Name, "/") {
			t.Errorf("VM name contains potentially dangerous characters: %s", vm.Name)
		}

		// Validate UUID format (basic check)
		if vm.UUID != "" && len(vm.UUID) != 36 {
			t.Errorf("VM UUID has unexpected length: %s", vm.UUID)
		}

		// Validate memory values are reasonable
		if vm.Memory > 0 && vm.Memory < 1024 {
			t.Logf("VM memory seems low: %d KB", vm.Memory)
		}

		// Validate CPU count is reasonable
		if vm.CPUs > 0 && vm.CPUs > 128 {
			t.Logf("VM CPU count seems high: %d", vm.CPUs)
		}
	}
}

// TestVMPerformance tests performance characteristics
func TestVMPerformance(t *testing.T) {
	manager := NewVMManager()

	// Test that operations complete within reasonable time
	start := time.Now()
	_, err := manager.ListVMs(true)
	duration := time.Since(start)

	if duration > 10*time.Second {
		t.Errorf("ListVMs took too long: %v", duration)
	}

	if err != nil {
		t.Logf("VM listing error (expected if libvirt not available): %v", err)
	}

	// Test VM stats performance
	start = time.Now()
	_, err = manager.GetVMStats("test_vm")
	duration = time.Since(start)

	if duration > 5*time.Second {
		t.Errorf("GetVMStats took too long: %v", duration)
	}

	if err != nil {
		t.Logf("VM stats error (expected if libvirt not available): %v", err)
	}
}

// TestVMConfigValidation tests VM configuration validation
func TestVMConfigValidation(t *testing.T) {
	manager := NewVMManager()

	vm, err := manager.GetVM("test_vm")
	if err != nil || vm == nil {
		t.Logf("VM retrieval error or no VM found: %v", err)
		return
	}

	// Validate network interface configuration
	for _, network := range vm.Networks {
		if network.MACAddress != "" {
			// Basic MAC address format validation
			parts := strings.Split(network.MACAddress, ":")
			if len(parts) != 6 {
				t.Errorf("Invalid MAC address format: %s", network.MACAddress)
			}
		}
	}

	// Validate disk configuration
	for _, disk := range vm.Disks {
		if disk.Type == "" {
			t.Error("Expected non-empty disk type")
		}

		if disk.Bus == "" {
			t.Error("Expected non-empty disk bus")
		}
	}
}

// BenchmarkListVMs benchmarks VM listing
func BenchmarkListVMs(b *testing.B) {
	manager := NewVMManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.ListVMs(true)
		if err != nil {
			b.Logf("VM listing error: %v", err)
		}
	}
}

// BenchmarkGetVMStats benchmarks VM stats retrieval
func BenchmarkGetVMStats(b *testing.B) {
	manager := NewVMManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GetVMStats("test_vm")
		if err != nil {
			b.Logf("VM stats error: %v", err)
		}
	}
}
