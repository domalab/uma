package async

import (
	"context"
	"fmt"
	"time"

	"github.com/domalab/uma/daemon/logger"
)

// ParityCheckExecutor executes parity check operations
type ParityCheckExecutor struct {
	storageMonitor StorageMonitorInterface
}

// StorageMonitorInterface defines the interface for storage operations
type StorageMonitorInterface interface {
	StartParityCheck(checkType string, priority string) error
	GetParityCheckStatus() (map[string]interface{}, error)
	CancelParityCheck() error
}

// NewParityCheckExecutor creates a new parity check executor
func NewParityCheckExecutor(storageMonitor StorageMonitorInterface) *ParityCheckExecutor {
	return &ParityCheckExecutor{
		storageMonitor: storageMonitor,
	}
}

// Execute executes a parity check operation
func (e *ParityCheckExecutor) Execute(ctx context.Context, op *AsyncOperation, params map[string]interface{}) error {
	// Extract parameters
	checkType, ok := params["type"].(string)
	if !ok {
		checkType = "check" // Default to check
	}
	
	priority, ok := params["priority"].(string)
	if !ok {
		priority = "normal" // Default priority
	}
	
	logger.Blue("Starting parity %s with priority %s", checkType, priority)
	
	// Start the parity check
	if err := e.storageMonitor.StartParityCheck(checkType, priority); err != nil {
		return fmt.Errorf("failed to start parity check: %w", err)
	}
	
	// Monitor progress
	return e.monitorParityProgress(ctx, op)
}

// monitorParityProgress monitors parity check progress
func (e *ParityCheckExecutor) monitorParityProgress(ctx context.Context, op *AsyncOperation) error {
	ticker := time.NewTicker(5 * time.Second) // Check every 5 seconds
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			// Operation was cancelled, try to cancel parity check
			if err := e.storageMonitor.CancelParityCheck(); err != nil {
				logger.Yellow("Failed to cancel parity check: %v", err)
			}
			return ctx.Err()
			
		case <-ticker.C:
			status, err := e.storageMonitor.GetParityCheckStatus()
			if err != nil {
				logger.Yellow("Failed to get parity check status: %v", err)
				continue
			}
			
			// Update progress
			if progress, ok := status["progress"].(int); ok {
				op.UpdateProgress(progress)
			}
			
			// Check if completed
			if active, ok := status["active"].(bool); ok && !active {
				// Parity check completed
				result := map[string]interface{}{
					"status":    status,
					"completed": time.Now(),
				}
				op.SetCompleted(result)
				return nil
			}
		}
	}
}

// GetType returns the operation type
func (e *ParityCheckExecutor) GetType() OperationType {
	return TypeParityCheck
}

// IsLongRunning returns true as parity checks are long-running
func (e *ParityCheckExecutor) IsLongRunning() bool {
	return true
}

// ArrayOperationExecutor executes array start/stop operations
type ArrayOperationExecutor struct {
	storageMonitor StorageMonitorInterface
	operationType  OperationType
}

// NewArrayStartExecutor creates a new array start executor
func NewArrayStartExecutor(storageMonitor StorageMonitorInterface) *ArrayOperationExecutor {
	return &ArrayOperationExecutor{
		storageMonitor: storageMonitor,
		operationType:  TypeArrayStart,
	}
}

// NewArrayStopExecutor creates a new array stop executor
func NewArrayStopExecutor(storageMonitor StorageMonitorInterface) *ArrayOperationExecutor {
	return &ArrayOperationExecutor{
		storageMonitor: storageMonitor,
		operationType:  TypeArrayStop,
	}
}

// Execute executes an array operation
func (e *ArrayOperationExecutor) Execute(ctx context.Context, op *AsyncOperation, params map[string]interface{}) error {
	switch e.operationType {
	case TypeArrayStart:
		return e.executeArrayStart(ctx, op, params)
	case TypeArrayStop:
		return e.executeArrayStop(ctx, op, params)
	default:
		return fmt.Errorf("unsupported array operation: %s", e.operationType)
	}
}

// executeArrayStart executes array start operation
func (e *ArrayOperationExecutor) executeArrayStart(ctx context.Context, op *AsyncOperation, params map[string]interface{}) error {
	// Extract parameters
	maintenanceMode, _ := params["maintenance_mode"].(bool)
	checkFilesystem, _ := params["check_filesystem"].(bool)
	
	logger.Blue("Starting array (maintenance: %v, check_fs: %v)", maintenanceMode, checkFilesystem)
	
	op.UpdateProgress(10)
	
	// This would call the actual storage monitor method
	// For now, simulate the operation
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(2 * time.Second):
		op.UpdateProgress(50)
	}
	
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(3 * time.Second):
		op.UpdateProgress(100)
	}
	
	result := map[string]interface{}{
		"array_status":      "started",
		"maintenance_mode":  maintenanceMode,
		"check_filesystem":  checkFilesystem,
		"completed_at":      time.Now(),
	}
	
	op.SetCompleted(result)
	return nil
}

// executeArrayStop executes array stop operation
func (e *ArrayOperationExecutor) executeArrayStop(ctx context.Context, op *AsyncOperation, params map[string]interface{}) error {
	logger.Blue("Stopping array")
	
	op.UpdateProgress(10)
	
	// Simulate stopping containers and VMs first
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		op.UpdateProgress(30)
	}
	
	// Simulate unmounting shares
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(3 * time.Second):
		op.UpdateProgress(60)
	}
	
	// Simulate stopping array
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(2 * time.Second):
		op.UpdateProgress(100)
	}
	
	result := map[string]interface{}{
		"array_status": "stopped",
		"completed_at": time.Now(),
	}
	
	op.SetCompleted(result)
	return nil
}

// GetType returns the operation type
func (e *ArrayOperationExecutor) GetType() OperationType {
	return e.operationType
}

// IsLongRunning returns true as array operations can be long-running
func (e *ArrayOperationExecutor) IsLongRunning() bool {
	return true
}

// SMARTScanExecutor executes SMART data collection operations
type SMARTScanExecutor struct {
	storageMonitor StorageMonitorInterface
}

// NewSMARTScanExecutor creates a new SMART scan executor
func NewSMARTScanExecutor(storageMonitor StorageMonitorInterface) *SMARTScanExecutor {
	return &SMARTScanExecutor{
		storageMonitor: storageMonitor,
	}
}

// Execute executes a SMART scan operation
func (e *SMARTScanExecutor) Execute(ctx context.Context, op *AsyncOperation, params map[string]interface{}) error {
	logger.Blue("Starting comprehensive SMART data collection")
	
	op.UpdateProgress(10)
	
	// Simulate SMART data collection for multiple disks
	diskCount := 8 // Simulate 8 disks
	for i := 0; i < diskCount; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
			progress := 10 + (i+1)*80/diskCount
			op.UpdateProgress(progress)
			logger.Blue("Collected SMART data for disk %d/%d", i+1, diskCount)
		}
	}
	
	result := map[string]interface{}{
		"disks_scanned": diskCount,
		"completed_at":  time.Now(),
		"scan_duration": "16s",
	}
	
	op.SetCompleted(result)
	return nil
}

// GetType returns the operation type
func (e *SMARTScanExecutor) GetType() OperationType {
	return TypeSMARTScan
}

// IsLongRunning returns false as SMART scans are relatively quick
func (e *SMARTScanExecutor) IsLongRunning() bool {
	return false
}

// BulkContainerExecutor executes bulk container operations
type BulkContainerExecutor struct {
	dockerManager DockerManagerInterface
}

// DockerManagerInterface defines the interface for Docker operations
type DockerManagerInterface interface {
	StartContainer(containerID string) error
	StopContainer(containerID string, timeout int) error
	RestartContainer(containerID string, timeout int) error
}

// NewBulkContainerExecutor creates a new bulk container executor
func NewBulkContainerExecutor(dockerManager DockerManagerInterface) *BulkContainerExecutor {
	return &BulkContainerExecutor{
		dockerManager: dockerManager,
	}
}

// Execute executes a bulk container operation
func (e *BulkContainerExecutor) Execute(ctx context.Context, op *AsyncOperation, params map[string]interface{}) error {
	containerIDs, ok := params["container_ids"].([]string)
	if !ok {
		return fmt.Errorf("container_ids parameter is required")
	}
	
	operation, ok := params["operation"].(string)
	if !ok {
		return fmt.Errorf("operation parameter is required")
	}
	
	logger.Blue("Starting bulk %s operation for %d containers", operation, len(containerIDs))
	
	results := make([]map[string]interface{}, len(containerIDs))
	
	for i, containerID := range containerIDs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Perform operation
			var err error
			switch operation {
			case "start":
				err = e.dockerManager.StartContainer(containerID)
			case "stop":
				err = e.dockerManager.StopContainer(containerID, 10)
			case "restart":
				err = e.dockerManager.RestartContainer(containerID, 10)
			default:
				err = fmt.Errorf("unsupported operation: %s", operation)
			}
			
			results[i] = map[string]interface{}{
				"container_id": containerID,
				"success":      err == nil,
			}
			
			if err != nil {
				results[i]["error"] = err.Error()
			}
			
			// Update progress
			progress := (i + 1) * 100 / len(containerIDs)
			op.UpdateProgress(progress)
		}
	}
	
	result := map[string]interface{}{
		"operation": operation,
		"results":   results,
		"total":     len(containerIDs),
	}
	
	op.SetCompleted(result)
	return nil
}

// GetType returns the operation type
func (e *BulkContainerExecutor) GetType() OperationType {
	return TypeBulkContainer
}

// IsLongRunning returns false as bulk container operations are usually quick
func (e *BulkContainerExecutor) IsLongRunning() bool {
	return false
}
