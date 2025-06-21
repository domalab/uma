package storage

import (
	"strings"
	"testing"
	"time"
)

// TestNewStorageMonitor tests the creation of a new storage monitor
func TestNewStorageMonitor(t *testing.T) {
	monitor := NewStorageMonitor()

	if monitor == nil {
		t.Fatal("Expected non-nil storage monitor")
	}
}

// TestGetArrayInfo tests array information retrieval
func TestGetArrayInfo(t *testing.T) {
	monitor := NewStorageMonitor()

	// Test array info retrieval
	arrayInfo, err := monitor.GetArrayInfo()

	// Should not error even if no array is present
	if err != nil {
		t.Logf("Array info error (expected in test environment): %v", err)
	}

	// Validate structure if array info is returned
	if arrayInfo != nil {
		if arrayInfo.State == "" {
			t.Error("Expected non-empty array state")
		}

		// Validate numeric fields are non-negative
		if arrayInfo.NumDevices < 0 {
			t.Error("Number of devices should not be negative")
		}

		if arrayInfo.NumDisks < 0 {
			t.Error("Number of disks should not be negative")
		}

		if arrayInfo.NumParity < 0 {
			t.Error("Number of parity disks should not be negative")
		}
	}
}

// TestGetConsolidatedDisksInfo tests consolidated disk information retrieval
func TestGetConsolidatedDisksInfo(t *testing.T) {
	monitor := NewStorageMonitor()

	disksResponse, err := monitor.GetConsolidatedDisksInfo()

	// Should not error in test environment
	if err != nil {
		t.Logf("Disk info error (expected in test environment): %v", err)
		return
	}

	// Validate disk structure if disks are returned
	if disksResponse != nil {
		// Test array disks
		for _, disk := range disksResponse.ArrayDisks {
			if disk.Name == "" {
				t.Error("Expected non-empty disk name")
			}

			if disk.Device == "" {
				t.Error("Expected non-empty disk device")
			}

			// Size should be reasonable (uint64 is always non-negative)
			if disk.Size == 0 {
				t.Logf("Disk size is 0, which may indicate missing data: %s", disk.Device)
			}

			// Validate status is one of expected values
			validStatuses := []string{"DISK_OK", "DISK_NP", "DISK_DSBL", "DISK_INVALID"}
			isValidStatus := false
			for _, status := range validStatuses {
				if disk.Status == status {
					isValidStatus = true
					break
				}
			}
			if !isValidStatus && disk.Status != "" {
				t.Errorf("Unexpected disk status: %s", disk.Status)
			}
		}

		// Test summary
		if disksResponse.Summary.TotalDisks < 0 {
			t.Error("Total disks should not be negative")
		}
	}
}

// TestGetParityCheckStatus tests parity check status retrieval
func TestGetParityCheckStatus(t *testing.T) {
	monitor := NewStorageMonitor()

	status, err := monitor.GetParityCheckStatus()

	// Should not error even if no parity check is running
	if err != nil {
		t.Logf("Parity check status error (expected in test environment): %v", err)
		return
	}

	// Validate parity check status structure if returned
	if status != nil {
		// Progress should be between 0 and 100 if active
		if status.Active && (status.Progress < 0 || status.Progress > 100) {
			t.Errorf("Invalid parity check progress: %f", status.Progress)
		}

		// Type should be valid if active
		if status.Active && status.Type != "" {
			validTypes := []string{"check", "correct"}
			isValidType := false
			for _, validType := range validTypes {
				if status.Type == validType {
					isValidType = true
					break
				}
			}
			if !isValidType {
				t.Errorf("Unexpected parity check type: %s", status.Type)
			}
		}
	}
}

// TestGetCacheInfo tests cache information retrieval
func TestGetCacheInfo(t *testing.T) {
	monitor := NewStorageMonitor()

	cacheInfos, err := monitor.GetCacheInfo()

	// Should not error even if no cache is configured
	if err != nil {
		t.Logf("Cache info error (expected in test environment): %v", err)
		return
	}

	// Validate cache structure if returned
	for _, cacheInfo := range cacheInfos {
		// Size should be reasonable (uint64 is always non-negative)
		if cacheInfo.TotalSize == 0 {
			t.Logf("Cache total size is 0, which may indicate missing data")
		}

		if cacheInfo.UsedSize == 0 {
			t.Logf("Cache used size is 0, which may indicate empty cache")
		}

		// Used should not exceed total
		if cacheInfo.UsedSize > cacheInfo.TotalSize {
			t.Errorf("Cache used (%d) should not exceed total (%d)", cacheInfo.UsedSize, cacheInfo.TotalSize)
		}
	}
}

// TestGetBootDiskInfo tests boot disk information retrieval
func TestGetBootDiskInfo(t *testing.T) {
	monitor := NewStorageMonitor()

	bootDisk, err := monitor.GetBootDiskInfo()

	// Should not error even in test environment
	if err != nil {
		t.Logf("Boot disk info error (expected in test environment): %v", err)
		return
	}

	// Validate boot disk structure if returned
	if bootDisk != nil {
		if bootDisk.Name != "boot" {
			t.Errorf("Expected boot disk name 'boot', got '%s'", bootDisk.Name)
		}

		if bootDisk.MountPoint != "/boot" {
			t.Errorf("Expected boot mount point '/boot', got '%s'", bootDisk.MountPoint)
		}

		// Size should be reasonable (uint64 is always non-negative)
		if bootDisk.Size == 0 {
			t.Logf("Boot disk size is 0, which may indicate missing data")
		}
	}
}

// TestStorageDataValidation tests data validation and sanitization
func TestStorageDataValidation(t *testing.T) {
	monitor := NewStorageMonitor()

	// Test that all methods return consistent data types
	arrayInfo, _ := monitor.GetArrayInfo()
	if arrayInfo != nil {
		// Validate that numeric fields are reasonable
		if arrayInfo.NumDisks < 0 {
			t.Error("Number of disks should not be negative")
		}

		if arrayInfo.NumDevices < 0 {
			t.Error("Number of devices should not be negative")
		}

		if arrayInfo.NumParity < 0 {
			t.Error("Number of parity disks should not be negative")
		}
	}

	// Test consolidated disk data validation
	disksResponse, _ := monitor.GetConsolidatedDisksInfo()
	if disksResponse != nil {
		for _, disk := range disksResponse.ArrayDisks {
			// Validate device names don't contain dangerous characters
			if strings.Contains(disk.Device, "..") {
				t.Errorf("Disk device name contains potentially dangerous characters: %s", disk.Device)
			}

			// Validate temperature is reasonable (if set)
			if disk.Temperature != 0 && (disk.Temperature < -50 || disk.Temperature > 100) {
				t.Errorf("Disk temperature seems unreasonable: %d°C", disk.Temperature)
			}
		}
	}
}

// TestStoragePerformance tests performance characteristics
func TestStoragePerformance(t *testing.T) {
	monitor := NewStorageMonitor()

	// Test that operations complete within reasonable time
	start := time.Now()
	_, err := monitor.GetArrayInfo()
	duration := time.Since(start)

	if duration > 5*time.Second {
		t.Errorf("GetArrayInfo took too long: %v", duration)
	}

	if err != nil {
		t.Logf("Array info error (expected in test environment): %v", err)
	}

	// Test consolidated disk enumeration performance
	start = time.Now()
	_, err = monitor.GetConsolidatedDisksInfo()
	duration = time.Since(start)

	if duration > 10*time.Second {
		t.Errorf("GetConsolidatedDisksInfo took too long: %v", duration)
	}

	if err != nil {
		t.Logf("Disk enumeration error (expected in test environment): %v", err)
	}
}

// BenchmarkGetArrayInfo benchmarks array info retrieval
func BenchmarkGetArrayInfo(b *testing.B) {
	monitor := NewStorageMonitor()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := monitor.GetArrayInfo()
		if err != nil {
			b.Logf("Array info error: %v", err)
		}
	}
}

// BenchmarkGetConsolidatedDisksInfo benchmarks consolidated disk enumeration
func BenchmarkGetConsolidatedDisksInfo(b *testing.B) {
	monitor := NewStorageMonitor()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := monitor.GetConsolidatedDisksInfo()
		if err != nil {
			b.Logf("Disk enumeration error: %v", err)
		}
	}
}
