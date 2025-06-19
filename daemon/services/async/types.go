package async

import (
	"context"
	"sync"
	"time"
)

// OperationStatus represents the status of an async operation
type OperationStatus string

const (
	StatusPending   OperationStatus = "pending"
	StatusRunning   OperationStatus = "running"
	StatusCompleted OperationStatus = "completed"
	StatusFailed    OperationStatus = "failed"
	StatusCancelled OperationStatus = "cancelled"
)

// OperationType represents the type of async operation
type OperationType string

const (
	TypeParityCheck    OperationType = "parity_check"
	TypeParityCorrect  OperationType = "parity_correct"
	TypeArrayStart     OperationType = "array_start"
	TypeArrayStop      OperationType = "array_stop"
	TypeDiskScan       OperationType = "disk_scan"
	TypeSMARTScan      OperationType = "smart_scan"
	TypeSystemReboot   OperationType = "system_reboot"
	TypeSystemShutdown OperationType = "system_shutdown"
	TypeBulkContainer  OperationType = "bulk_container"
	TypeBulkVM         OperationType = "bulk_vm"
)

// AsyncOperation represents a long-running asynchronous operation
type AsyncOperation struct {
	ID          string                 `json:"id"`
	Type        OperationType          `json:"type"`
	Status      OperationStatus        `json:"status"`
	Progress    int                    `json:"progress"` // 0-100
	Started     time.Time              `json:"started"`
	Completed   *time.Time             `json:"completed,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Cancellable bool                   `json:"cancellable"`
	Description string                 `json:"description"`
	CreatedBy   string                 `json:"created_by,omitempty"`

	// Internal fields
	ctx        context.Context    `json:"-"`
	cancel     context.CancelFunc `json:"-"`
	mutex      sync.RWMutex       `json:"-"`
	onProgress func(int)          `json:"-"`
	onComplete func(error)        `json:"-"`
}

// OperationRequest represents a request to start an async operation
type OperationRequest struct {
	Type        OperationType          `json:"type" validate:"required"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Cancellable bool                   `json:"cancellable"`
}

// OperationResponse represents the response when starting an async operation
type OperationResponse struct {
	ID          string          `json:"id"`
	Type        OperationType   `json:"type"`
	Status      OperationStatus `json:"status"`
	Description string          `json:"description"`
	Cancellable bool            `json:"cancellable"`
	Started     time.Time       `json:"started"`
}

// OperationListResponse represents a list of operations
type OperationListResponse struct {
	Operations []SafeAsyncOperation `json:"operations"`
	Total      int                  `json:"total"`
	Active     int                  `json:"active"`
	Completed  int                  `json:"completed"`
	Failed     int                  `json:"failed"`
}

// UpdateProgress updates the operation progress (thread-safe)
func (op *AsyncOperation) UpdateProgress(progress int) {
	op.mutex.Lock()
	defer op.mutex.Unlock()

	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}

	op.Progress = progress

	if op.onProgress != nil {
		op.onProgress(progress)
	}
}

// SetError sets the operation error and status (thread-safe)
func (op *AsyncOperation) SetError(err error) {
	op.mutex.Lock()
	defer op.mutex.Unlock()

	op.Status = StatusFailed
	op.Error = err.Error()
	now := time.Now()
	op.Completed = &now

	if op.onComplete != nil {
		op.onComplete(err)
	}
}

// SetCompleted marks the operation as completed (thread-safe)
func (op *AsyncOperation) SetCompleted(result map[string]interface{}) {
	op.mutex.Lock()
	defer op.mutex.Unlock()

	op.Status = StatusCompleted
	op.Progress = 100
	op.Result = result
	now := time.Now()
	op.Completed = &now

	if op.onComplete != nil {
		op.onComplete(nil)
	}
}

// SetRunning marks the operation as running (thread-safe)
func (op *AsyncOperation) SetRunning() {
	op.mutex.Lock()
	defer op.mutex.Unlock()

	op.Status = StatusRunning
}

// Cancel cancels the operation if it's cancellable (thread-safe)
func (op *AsyncOperation) Cancel() bool {
	op.mutex.Lock()
	defer op.mutex.Unlock()

	if !op.Cancellable || op.Status == StatusCompleted || op.Status == StatusFailed || op.Status == StatusCancelled {
		return false
	}

	op.Status = StatusCancelled
	now := time.Now()
	op.Completed = &now

	if op.cancel != nil {
		op.cancel()
	}

	if op.onComplete != nil {
		op.onComplete(context.Canceled)
	}

	return true
}

// IsActive returns true if the operation is still active
func (op *AsyncOperation) IsActive() bool {
	op.mutex.RLock()
	defer op.mutex.RUnlock()

	return op.Status == StatusPending || op.Status == StatusRunning
}

// GetSafeStatus returns the current status (thread-safe)
func (op *AsyncOperation) GetSafeStatus() OperationStatus {
	op.mutex.RLock()
	defer op.mutex.RUnlock()

	return op.Status
}

// GetSafeProgress returns the current progress (thread-safe)
func (op *AsyncOperation) GetSafeProgress() int {
	op.mutex.RLock()
	defer op.mutex.RUnlock()

	return op.Progress
}

// SetProgressCallback sets a callback for progress updates
func (op *AsyncOperation) SetProgressCallback(callback func(int)) {
	op.mutex.Lock()
	defer op.mutex.Unlock()

	op.onProgress = callback
}

// SetCompletionCallback sets a callback for operation completion
func (op *AsyncOperation) SetCompletionCallback(callback func(error)) {
	op.mutex.Lock()
	defer op.mutex.Unlock()

	op.onComplete = callback
}

// GetContext returns the operation context
func (op *AsyncOperation) GetContext() context.Context {
	op.mutex.RLock()
	defer op.mutex.RUnlock()

	return op.ctx
}

// SafeAsyncOperation represents an operation without internal fields for safe copying
type SafeAsyncOperation struct {
	ID          string                 `json:"id"`
	Type        OperationType          `json:"type"`
	Status      OperationStatus        `json:"status"`
	Progress    int                    `json:"progress"`
	Started     time.Time              `json:"started"`
	Completed   *time.Time             `json:"completed,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Cancellable bool                   `json:"cancellable"`
	Description string                 `json:"description"`
	CreatedBy   string                 `json:"created_by,omitempty"`
}

// ToSafeOperation returns a copy of the operation without internal fields (thread-safe)
func (op *AsyncOperation) ToSafeOperation() SafeAsyncOperation {
	op.mutex.RLock()
	defer op.mutex.RUnlock()

	// Create a copy without the mutex and internal fields
	return SafeAsyncOperation{
		ID:          op.ID,
		Type:        op.Type,
		Status:      op.Status,
		Progress:    op.Progress,
		Started:     op.Started,
		Completed:   op.Completed,
		Error:       op.Error,
		Result:      op.Result,
		Cancellable: op.Cancellable,
		Description: op.Description,
		CreatedBy:   op.CreatedBy,
	}
}

// OperationExecutor defines the interface for executing operations
type OperationExecutor interface {
	Execute(ctx context.Context, op *AsyncOperation, params map[string]interface{}) error
	GetType() OperationType
	IsLongRunning() bool
}
