package async

import (
	"context"
	"strings"
	"testing"
	"time"
)

// MockExecutor for testing
type MockExecutor struct {
	operationType OperationType
	shouldFail    bool
	duration      time.Duration
}

func (m *MockExecutor) Execute(ctx context.Context, op *AsyncOperation, params map[string]interface{}) error {
	if m.shouldFail {
		return context.Canceled
	}

	// Simulate work
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(m.duration):
		op.UpdateProgress(100)
		return nil
	}
}

func (m *MockExecutor) GetType() OperationType {
	return m.operationType
}

func (m *MockExecutor) IsLongRunning() bool {
	return m.duration > time.Second
}

func TestAsyncManager_StartOperation(t *testing.T) {
	manager := NewAsyncManager()
	defer manager.Stop()

	// Register mock executor
	mockExecutor := &MockExecutor{
		operationType: TypeSMARTScan,
		shouldFail:    false,
		duration:      100 * time.Millisecond,
	}
	manager.RegisterExecutor(mockExecutor)

	// Test starting an operation
	req := OperationRequest{
		Type:        TypeSMARTScan,
		Description: "Test SMART scan",
		Cancellable: true,
	}

	operation, err := manager.StartOperation(req, "test-user")
	if err != nil {
		t.Fatalf("Failed to start operation: %v", err)
	}

	if operation.Type != TypeSMARTScan {
		t.Errorf("Expected operation type %s, got %s", TypeSMARTScan, operation.Type)
	}

	// Operation might start immediately, so check for pending or running
	if operation.Status != StatusPending && operation.Status != StatusRunning {
		t.Errorf("Expected status %s or %s, got %s", StatusPending, StatusRunning, operation.Status)
	}

	// Wait for operation to complete
	time.Sleep(200 * time.Millisecond)

	// Check final status
	finalOp, err := manager.GetOperation(operation.ID)
	if err != nil {
		t.Fatalf("Failed to get operation: %v", err)
	}

	if finalOp.Status != StatusCompleted {
		t.Errorf("Expected final status %s, got %s", StatusCompleted, finalOp.Status)
	}

	if finalOp.Progress != 100 {
		t.Errorf("Expected progress 100, got %d", finalOp.Progress)
	}
}

func TestAsyncManager_CancelOperation(t *testing.T) {
	manager := NewAsyncManager()
	defer manager.Stop()

	// Register mock executor with longer duration
	mockExecutor := &MockExecutor{
		operationType: TypeParityCheck,
		shouldFail:    false,
		duration:      5 * time.Second, // Long running
	}
	manager.RegisterExecutor(mockExecutor)

	// Start operation
	req := OperationRequest{
		Type:        TypeParityCheck,
		Description: "Test parity check",
		Cancellable: true,
	}

	operation, err := manager.StartOperation(req, "test-user")
	if err != nil {
		t.Fatalf("Failed to start operation: %v", err)
	}

	// Wait a bit for operation to start
	time.Sleep(50 * time.Millisecond)

	// Cancel the operation
	err = manager.CancelOperation(operation.ID)
	if err != nil {
		t.Fatalf("Failed to cancel operation: %v", err)
	}

	// Check status
	cancelledOp, err := manager.GetOperation(operation.ID)
	if err != nil {
		t.Fatalf("Failed to get cancelled operation: %v", err)
	}

	if cancelledOp.Status != StatusCancelled {
		t.Errorf("Expected status %s, got %s", StatusCancelled, cancelledOp.Status)
	}
}

func TestAsyncManager_ConflictingOperations(t *testing.T) {
	manager := NewAsyncManager()
	defer manager.Stop()

	// Register mock executor
	mockExecutor := &MockExecutor{
		operationType: TypeParityCheck,
		shouldFail:    false,
		duration:      2 * time.Second,
	}
	manager.RegisterExecutor(mockExecutor)

	// Start first operation
	req1 := OperationRequest{
		Type:        TypeParityCheck,
		Description: "First parity check",
		Cancellable: true,
	}

	_, err := manager.StartOperation(req1, "test-user")
	if err != nil {
		t.Fatalf("Failed to start first operation: %v", err)
	}

	// Try to start conflicting operation
	req2 := OperationRequest{
		Type:        TypeParityCheck,
		Description: "Second parity check",
		Cancellable: true,
	}

	_, err = manager.StartOperation(req2, "test-user")
	if err == nil {
		t.Error("Expected error for conflicting operation, but got none")
	}

	// Check that error mentions conflicting operation (exact format may vary)
	if !strings.Contains(err.Error(), "conflicting operation") {
		t.Errorf("Expected error to mention conflicting operation, got: %v", err)
	}
}

func TestAsyncManager_ListOperations(t *testing.T) {
	manager := NewAsyncManager()
	defer manager.Stop()

	// Register mock executor
	mockExecutor := &MockExecutor{
		operationType: TypeSMARTScan,
		shouldFail:    false,
		duration:      100 * time.Millisecond,
	}
	manager.RegisterExecutor(mockExecutor)

	// Start multiple operations
	for i := 0; i < 3; i++ {
		req := OperationRequest{
			Type:        TypeSMARTScan,
			Description: "Test operation",
			Cancellable: true,
		}

		_, err := manager.StartOperation(req, "test-user")
		if err != nil {
			t.Fatalf("Failed to start operation %d: %v", i, err)
		}
	}

	// Give operations a moment to start
	time.Sleep(10 * time.Millisecond)

	// List all operations
	response := manager.ListOperations("", "")
	if response.Total != 3 {
		t.Errorf("Expected 3 total operations, got %d", response.Total)
	}

	// List by type
	response = manager.ListOperations("", TypeSMARTScan)
	if response.Total != 3 {
		t.Errorf("Expected 3 SMART scan operations, got %d", response.Total)
	}

	// List by status (should have some running/pending or completed)
	response = manager.ListOperations(StatusRunning, "")
	runningCount := response.Total

	response = manager.ListOperations(StatusCompleted, "")
	completedCount := response.Total

	if runningCount == 0 && completedCount == 0 {
		t.Error("Expected some operations to be running or completed")
	}
}

func TestAsyncOperation_ProgressUpdates(t *testing.T) {
	operation := &AsyncOperation{
		ID:          "test-op",
		Type:        TypeSMARTScan,
		Status:      StatusRunning,
		Progress:    0,
		Cancellable: true,
	}

	// Test progress updates
	operation.UpdateProgress(25)
	if operation.GetSafeProgress() != 25 {
		t.Errorf("Expected progress 25, got %d", operation.GetSafeProgress())
	}

	operation.UpdateProgress(50)
	if operation.GetSafeProgress() != 50 {
		t.Errorf("Expected progress 50, got %d", operation.GetSafeProgress())
	}

	// Test progress bounds
	operation.UpdateProgress(-10)
	if operation.GetSafeProgress() != 0 {
		t.Errorf("Expected progress 0 (bounded), got %d", operation.GetSafeProgress())
	}

	operation.UpdateProgress(150)
	if operation.GetSafeProgress() != 100 {
		t.Errorf("Expected progress 100 (bounded), got %d", operation.GetSafeProgress())
	}
}

func TestAsyncOperation_StatusTransitions(t *testing.T) {
	operation := &AsyncOperation{
		ID:          "test-op",
		Type:        TypeSMARTScan,
		Status:      StatusPending,
		Progress:    0,
		Cancellable: true,
	}

	// Test status transitions
	operation.SetRunning()
	if operation.GetSafeStatus() != StatusRunning {
		t.Errorf("Expected status %s, got %s", StatusRunning, operation.GetSafeStatus())
	}

	// Test completion
	result := map[string]interface{}{"test": "result"}
	operation.SetCompleted(result)
	if operation.GetSafeStatus() != StatusCompleted {
		t.Errorf("Expected status %s, got %s", StatusCompleted, operation.GetSafeStatus())
	}

	if operation.GetSafeProgress() != 100 {
		t.Errorf("Expected progress 100 after completion, got %d", operation.GetSafeProgress())
	}
}

func TestAsyncManager_Stats(t *testing.T) {
	manager := NewAsyncManager()
	defer manager.Stop()

	// Register mock executor
	mockExecutor := &MockExecutor{
		operationType: TypeSMARTScan,
		shouldFail:    false,
		duration:      100 * time.Millisecond,
	}
	manager.RegisterExecutor(mockExecutor)

	// Start some operations
	for i := 0; i < 2; i++ {
		req := OperationRequest{
			Type:        TypeSMARTScan,
			Description: "Test operation",
			Cancellable: true,
		}

		_, err := manager.StartOperation(req, "test-user")
		if err != nil {
			t.Fatalf("Failed to start operation %d: %v", i, err)
		}
	}

	// Get stats
	stats := manager.GetStats()

	totalOps, ok := stats["total_operations"].(int)
	if !ok || totalOps != 2 {
		t.Errorf("Expected 2 total operations in stats, got %v", stats["total_operations"])
	}

	maxOps, ok := stats["max_operations"].(int)
	if !ok || maxOps != 100 {
		t.Errorf("Expected max_operations 100 in stats, got %v", stats["max_operations"])
	}

	// Check by_status and by_type maps exist
	if _, ok := stats["by_status"]; !ok {
		t.Error("Expected by_status in stats")
	}

	if _, ok := stats["by_type"]; !ok {
		t.Error("Expected by_type in stats")
	}
}
