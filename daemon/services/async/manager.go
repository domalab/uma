package async

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/domalab/uma/daemon/logger"
	"github.com/google/uuid"
)

// AsyncManager manages asynchronous operations
type AsyncManager struct {
	operations map[string]*AsyncOperation
	executors  map[OperationType]OperationExecutor
	mutex      sync.RWMutex

	// Configuration
	maxOperations    int
	cleanupInterval  time.Duration
	operationTimeout time.Duration

	// Cleanup
	stopCleanup chan struct{}
	cleanupWG   sync.WaitGroup
}

// NewAsyncManager creates a new async manager
func NewAsyncManager() *AsyncManager {
	manager := &AsyncManager{
		operations:       make(map[string]*AsyncOperation),
		executors:        make(map[OperationType]OperationExecutor),
		maxOperations:    100, // Maximum concurrent operations
		cleanupInterval:  5 * time.Minute,
		operationTimeout: 30 * time.Minute,
		stopCleanup:      make(chan struct{}),
	}

	// Start cleanup goroutine
	manager.startCleanup()

	return manager
}

// RegisterExecutor registers an operation executor
func (am *AsyncManager) RegisterExecutor(executor OperationExecutor) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	am.executors[executor.GetType()] = executor
	logger.Blue("Registered async executor for operation type: %s", executor.GetType())
}

// StartOperation starts a new asynchronous operation
func (am *AsyncManager) StartOperation(req OperationRequest, createdBy string) (*AsyncOperation, error) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	// Check if we have an executor for this operation type
	executor, exists := am.executors[req.Type]
	if !exists {
		return nil, fmt.Errorf("no executor registered for operation type: %s", req.Type)
	}

	// Check operation limits
	if len(am.operations) >= am.maxOperations {
		return nil, fmt.Errorf("maximum number of operations (%d) reached", am.maxOperations)
	}

	// Check for conflicting operations (e.g., only one parity check at a time)
	if err := am.checkConflicts(req.Type); err != nil {
		return nil, err
	}

	// Generate operation ID
	operationID := uuid.New().String()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), am.operationTimeout)

	// Create operation
	operation := &AsyncOperation{
		ID:          operationID,
		Type:        req.Type,
		Status:      StatusPending,
		Progress:    0,
		Started:     time.Now(),
		Cancellable: req.Cancellable,
		Description: req.Description,
		CreatedBy:   createdBy,
		ctx:         ctx,
		cancel:      cancel,
	}

	// Store operation
	am.operations[operationID] = operation

	// Start operation in goroutine
	go am.executeOperation(operation, executor, req.Parameters)

	logger.Green("Started async operation %s (%s) by %s", operationID, req.Type, createdBy)

	return operation, nil
}

// GetOperation retrieves an operation by ID
func (am *AsyncManager) GetOperation(operationID string) (*AsyncOperation, error) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	operation, exists := am.operations[operationID]
	if !exists {
		return nil, fmt.Errorf("operation not found: %s", operationID)
	}

	return operation, nil
}

// ListOperations returns all operations with optional filtering
func (am *AsyncManager) ListOperations(status OperationStatus, operationType OperationType) *OperationListResponse {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	var filteredOps []SafeAsyncOperation
	var active, completed, failed int

	for _, op := range am.operations {
		// Apply filters
		if status != "" && op.Status != status {
			continue
		}
		if operationType != "" && op.Type != operationType {
			continue
		}

		// Create a safe copy to avoid race conditions
		safeCopy := op.ToSafeOperation()
		filteredOps = append(filteredOps, safeCopy)

		// Count by status
		switch op.Status {
		case StatusPending, StatusRunning:
			active++
		case StatusCompleted:
			completed++
		case StatusFailed:
			failed++
		}
	}

	return &OperationListResponse{
		Operations: filteredOps,
		Total:      len(filteredOps),
		Active:     active,
		Completed:  completed,
		Failed:     failed,
	}
}

// CancelOperation cancels an operation
func (am *AsyncManager) CancelOperation(operationID string) error {
	am.mutex.RLock()
	operation, exists := am.operations[operationID]
	am.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("operation not found: %s", operationID)
	}

	if !operation.Cancel() {
		return fmt.Errorf("operation cannot be cancelled (status: %s, cancellable: %v)",
			operation.Status, operation.Cancellable)
	}

	logger.Yellow("Cancelled async operation %s (%s)", operationID, operation.Type)
	return nil
}

// executeOperation executes an operation in a goroutine
func (am *AsyncManager) executeOperation(operation *AsyncOperation, executor OperationExecutor, params map[string]interface{}) {
	defer func() {
		if r := recover(); r != nil {
			logger.Red("Panic in async operation %s: %v", operation.ID, r)
			operation.SetError(fmt.Errorf("operation panicked: %v", r))
		}
	}()

	// Mark as running
	operation.SetRunning()

	// Execute the operation
	err := executor.Execute(operation.GetContext(), operation, params)

	// Handle result
	if err != nil {
		if err == context.Canceled {
			// Operation was cancelled, status already set
			logger.Yellow("Async operation %s was cancelled", operation.ID)
		} else {
			operation.SetError(err)
			logger.Red("Async operation %s failed: %v", operation.ID, err)
		}
	} else if operation.GetSafeStatus() == StatusRunning {
		// Only set completed if not already cancelled
		operation.SetCompleted(nil)
		logger.Green("Async operation %s completed successfully", operation.ID)
	}
}

// checkConflicts checks for conflicting operations
func (am *AsyncManager) checkConflicts(operationType OperationType) error {
	// Define conflicting operation types
	conflicts := map[OperationType][]OperationType{
		TypeParityCheck:    {TypeParityCheck, TypeParityCorrect, TypeArrayStart, TypeArrayStop},
		TypeParityCorrect:  {TypeParityCheck, TypeParityCorrect, TypeArrayStart, TypeArrayStop},
		TypeArrayStart:     {TypeParityCheck, TypeParityCorrect, TypeArrayStart, TypeArrayStop},
		TypeArrayStop:      {TypeParityCheck, TypeParityCorrect, TypeArrayStart, TypeArrayStop},
		TypeSystemReboot:   {TypeSystemReboot, TypeSystemShutdown},
		TypeSystemShutdown: {TypeSystemReboot, TypeSystemShutdown},
	}

	conflictTypes, hasConflicts := conflicts[operationType]
	if !hasConflicts {
		return nil // No conflicts defined for this operation type
	}

	// Check for active conflicting operations
	for _, op := range am.operations {
		if op.IsActive() {
			for _, conflictType := range conflictTypes {
				if op.Type == conflictType {
					return fmt.Errorf("conflicting operation already running: %s (%s)", op.ID, op.Type)
				}
			}
		}
	}

	return nil
}

// startCleanup starts the cleanup goroutine
func (am *AsyncManager) startCleanup() {
	am.cleanupWG.Add(1)
	go func() {
		defer am.cleanupWG.Done()

		ticker := time.NewTicker(am.cleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				am.cleanup()
			case <-am.stopCleanup:
				return
			}
		}
	}()
}

// cleanup removes old completed operations
func (am *AsyncManager) cleanup() {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour) // Keep operations for 24 hours
	var removed int

	for id, op := range am.operations {
		if !op.IsActive() && op.Started.Before(cutoff) {
			delete(am.operations, id)
			removed++
		}
	}

	if removed > 0 {
		logger.Blue("Cleaned up %d old async operations", removed)
	}
}

// Stop stops the async manager
func (am *AsyncManager) Stop() {
	close(am.stopCleanup)
	am.cleanupWG.Wait()

	// Cancel all active operations
	am.mutex.Lock()
	defer am.mutex.Unlock()

	for _, op := range am.operations {
		if op.IsActive() {
			op.Cancel()
		}
	}

	logger.Blue("Async manager stopped")
}

// GetStats returns statistics about operations
func (am *AsyncManager) GetStats() map[string]interface{} {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_operations": len(am.operations),
		"max_operations":   am.maxOperations,
		"by_status":        make(map[string]int),
		"by_type":          make(map[string]int),
	}

	statusCounts := make(map[string]int)
	typeCounts := make(map[string]int)

	for _, op := range am.operations {
		statusCounts[string(op.Status)]++
		typeCounts[string(op.Type)]++
	}

	stats["by_status"] = statusCounts
	stats["by_type"] = typeCounts

	return stats
}
