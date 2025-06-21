package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/services/api/utils"
)

// VMService handles virtual machine-related business logic
type VMService struct {
	api utils.APIInterface
}

// NewVMService creates a new VM service
func NewVMService(api utils.APIInterface) *VMService {
	return &VMService{
		api: api,
	}
}

// VM represents a virtual machine
type VM struct {
	Name        string `json:"name"`
	UUID        string `json:"uuid"`
	State       string `json:"state"`
	CPUs        int    `json:"cpus"`
	Memory      int64  `json:"memory"`
	Description string `json:"description"`
	Autostart   bool   `json:"autostart"`
}

// VMStats represents VM resource usage statistics
type VMStats struct {
	Name          string  `json:"name"`
	State         string  `json:"state"`
	CPUTime       uint64  `json:"cpu_time"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryUsed    uint64  `json:"memory_used"`
	MemoryTotal   uint64  `json:"memory_total"`
	MemoryPercent float64 `json:"memory_percent"`
	NetworkRx     uint64  `json:"network_rx"`
	NetworkTx     uint64  `json:"network_tx"`
	DiskRead      uint64  `json:"disk_read"`
	DiskWrite     uint64  `json:"disk_write"`
}

// VMActionRequest represents a VM action request
type VMActionRequest struct {
	Action string `json:"action"` // start, stop, restart, pause, resume, shutdown
	Force  bool   `json:"force"`  // force action
}

// GetVMs retrieves all virtual machines
func (v *VMService) GetVMs() ([]VM, error) {
	vmsInfo, err := v.api.GetVM().GetVMs()
	if err != nil {
		return nil, fmt.Errorf("failed to get VMs: %v", err)
	}

	// Try to convert to VM slice
	if vmSlice, ok := vmsInfo.([]VM); ok {
		return vmSlice, nil
	}

	// Fallback: parse from interface{}
	vms := []VM{}
	if vmData, ok := vmsInfo.([]interface{}); ok {
		for _, item := range vmData {
			if vmMap, ok := item.(map[string]interface{}); ok {
				vm := VM{
					Name:        v.getStringValue(vmMap, "name"),
					UUID:        v.getStringValue(vmMap, "uuid"),
					State:       v.getStringValue(vmMap, "state"),
					CPUs:        v.getIntValue(vmMap, "cpus"),
					Memory:      v.getInt64Value(vmMap, "memory"),
					Description: v.getStringValue(vmMap, "description"),
					Autostart:   v.getBoolValue(vmMap, "autostart"),
				}
				vms = append(vms, vm)
			}
		}
	}

	return vms, nil
}

// GetVM retrieves a specific virtual machine
func (v *VMService) GetVM(vmName string) (*VM, error) {
	vmInfo, err := v.api.GetVM().GetVM(vmName)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM: %v", err)
	}

	// Try to convert to VM
	if vm, ok := vmInfo.(*VM); ok {
		return vm, nil
	}

	// Fallback: parse from interface{}
	if vmMap, ok := vmInfo.(map[string]interface{}); ok {
		vm := &VM{
			Name:        v.getStringValue(vmMap, "name"),
			UUID:        v.getStringValue(vmMap, "uuid"),
			State:       v.getStringValue(vmMap, "state"),
			CPUs:        v.getIntValue(vmMap, "cpus"),
			Memory:      v.getInt64Value(vmMap, "memory"),
			Description: v.getStringValue(vmMap, "description"),
			Autostart:   v.getBoolValue(vmMap, "autostart"),
		}
		return vm, nil
	}

	return nil, fmt.Errorf("invalid VM data format")
}

// GetVMData retrieves VM data in optimized format
func (v *VMService) GetVMData() []map[string]interface{} {
	vms, err := v.GetVMs()
	if err != nil {
		logger.Red("Failed to get VM data: %v", err)
		return []map[string]interface{}{}
	}

	// Convert VMs to map format
	vmData := make([]map[string]interface{}, len(vms))
	for i, vm := range vms {
		vmData[i] = map[string]interface{}{
			"name":        vm.Name,
			"uuid":        vm.UUID,
			"state":       vm.State,
			"cpus":        vm.CPUs,
			"memory":      vm.Memory,
			"description": vm.Description,
			"autostart":   vm.Autostart,
		}
	}

	return vmData
}

// GetVMDataOptimized retrieves VM data in optimized format with enhanced information
func (v *VMService) GetVMDataOptimized() interface{} {
	// Check if VM manager is available
	if v.api.GetVM() == nil {
		return map[string]interface{}{
			"available": false,
			"message":   "VM manager not available",
			"vms":       []interface{}{},
		}
	}

	vms := v.GetVMData()

	return map[string]interface{}{
		"available":    true,
		"total_vms":    len(vms),
		"vms":          vms,
		"last_updated": time.Now().UTC().Format(time.RFC3339),
	}
}

// StartVM starts a virtual machine
func (v *VMService) StartVM(vmName string) error {
	if err := v.api.GetVM().StartVM(vmName); err != nil {
		return fmt.Errorf("failed to start VM: %v", err)
	}

	logger.Blue("Started VM: %s", vmName)
	return nil
}

// StopVM stops a virtual machine
func (v *VMService) StopVM(vmName string, force bool) error {
	if err := v.api.GetVM().StopVM(vmName); err != nil {
		return fmt.Errorf("failed to stop VM: %v", err)
	}

	logger.Blue("Stopped VM: %s", vmName)
	return nil
}

// RestartVM restarts a virtual machine
func (v *VMService) RestartVM(vmName string) error {
	if err := v.api.GetVM().RestartVM(vmName); err != nil {
		return fmt.Errorf("failed to restart VM: %v", err)
	}

	logger.Blue("Restarted VM: %s", vmName)
	return nil
}

// PauseVM pauses a virtual machine
func (v *VMService) PauseVM(vmName string) error {
	// Pause functionality might not be available in current interface
	// Would need to be implemented in the VM plugin
	logger.Blue("Pause VM requested: %s (checking implementation)", vmName)
	return fmt.Errorf("pause VM operation may not be implemented")
}

// ResumeVM resumes a paused virtual machine
func (v *VMService) ResumeVM(vmName string) error {
	// Resume functionality might not be available in current interface
	// Would need to be implemented in the VM plugin
	logger.Blue("Resume VM requested: %s (checking implementation)", vmName)
	return fmt.Errorf("resume VM operation may not be implemented")
}

// ShutdownVM gracefully shuts down a virtual machine
func (v *VMService) ShutdownVM(vmName string) error {
	// Graceful shutdown might not be available in current interface
	// Would need to be implemented in the VM plugin
	logger.Blue("Shutdown VM requested: %s (checking implementation)", vmName)
	return fmt.Errorf("shutdown VM operation may not be implemented")
}

// GetVMStats retrieves VM resource usage statistics
func (v *VMService) GetVMStats(vmName string) (*VMStats, error) {
	// Call the actual VM interface to get stats
	stats, err := v.api.GetVM().GetVMStats(vmName)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM stats: %v", err)
	}

	// Convert interface{} to VMStats
	if statsMap, ok := stats.(map[string]interface{}); ok {
		vmStats := &VMStats{
			Name: vmName,
		}

		// Extract values from the stats map
		if cpuPercent, ok := statsMap["cpu_percent"].(float64); ok {
			vmStats.CPUPercent = cpuPercent
		}
		if memoryUsed, ok := statsMap["memory_used"].(uint64); ok {
			vmStats.MemoryUsed = memoryUsed
		}
		if memoryTotal, ok := statsMap["memory_total"].(uint64); ok {
			vmStats.MemoryTotal = memoryTotal
		}
		if memoryPercent, ok := statsMap["memory_percent"].(float64); ok {
			vmStats.MemoryPercent = memoryPercent
		}
		if networkRx, ok := statsMap["network_rx"].(uint64); ok {
			vmStats.NetworkRx = networkRx
		}
		if networkTx, ok := statsMap["network_tx"].(uint64); ok {
			vmStats.NetworkTx = networkTx
		}
		if diskRead, ok := statsMap["disk_read"].(uint64); ok {
			vmStats.DiskRead = diskRead
		}
		if diskWrite, ok := statsMap["disk_write"].(uint64); ok {
			vmStats.DiskWrite = diskWrite
		}

		return vmStats, nil
	}

	// Fallback if conversion fails
	return &VMStats{
		Name:          vmName,
		State:         "unknown",
		CPUTime:       0,
		CPUPercent:    0.0,
		MemoryUsed:    0,
		MemoryTotal:   0,
		MemoryPercent: 0.0,
		NetworkRx:     0,
		NetworkTx:     0,
		DiskRead:      0,
		DiskWrite:     0,
	}, nil
}

// CheckLibvirtHealth checks the health of the libvirt connection
func (v *VMService) CheckLibvirtHealth() string {
	if v.api.GetVM() == nil {
		return "unavailable"
	}

	// Try to get VMs to test connection
	_, err := v.api.GetVM().GetVMs()
	if err != nil {
		if strings.Contains(err.Error(), "connection") {
			return "disconnected"
		}
		return "error"
	}

	return "healthy"
}

// Helper methods

// getStringValue safely gets a string value from a map
func (v *VMService) getStringValue(m map[string]interface{}, key string) string {
	if value, ok := m[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// getIntValue safely gets an int value from a map
func (v *VMService) getIntValue(m map[string]interface{}, key string) int {
	if value, ok := m[key]; ok {
		if intVal, ok := value.(int); ok {
			return intVal
		}
		if floatVal, ok := value.(float64); ok {
			return int(floatVal)
		}
	}
	return 0
}

// getInt64Value safely gets an int64 value from a map
func (v *VMService) getInt64Value(m map[string]interface{}, key string) int64 {
	if value, ok := m[key]; ok {
		if intVal, ok := value.(int64); ok {
			return intVal
		}
		if intVal, ok := value.(int); ok {
			return int64(intVal)
		}
		if floatVal, ok := value.(float64); ok {
			return int64(floatVal)
		}
	}
	return 0
}

// getBoolValue safely gets a bool value from a map
func (v *VMService) getBoolValue(m map[string]interface{}, key string) bool {
	if value, ok := m[key]; ok {
		if boolVal, ok := value.(bool); ok {
			return boolVal
		}
	}
	return false
}
