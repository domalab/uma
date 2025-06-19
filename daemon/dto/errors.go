package dto

import (
	"fmt"
	"net/http"
)

// ErrorCode represents standardized error codes for the UMA API
type ErrorCode string

const (
	// General errors
	ErrCodeInvalidRequest     ErrorCode = "INVALID_REQUEST"
	ErrCodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden          ErrorCode = "FORBIDDEN"
	ErrCodeNotFound           ErrorCode = "NOT_FOUND"
	ErrCodeConflict           ErrorCode = "CONFLICT"
	ErrCodeInternalError      ErrorCode = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeRateLimitExceeded  ErrorCode = "RATE_LIMIT_EXCEEDED"

	// Validation errors
	ErrCodeValidationFailed   ErrorCode = "VALIDATION_FAILED"
	ErrCodeMissingParameter   ErrorCode = "MISSING_PARAMETER"
	ErrCodeInvalidParameter   ErrorCode = "INVALID_PARAMETER"
	ErrCodeParameterOutOfRange ErrorCode = "PARAMETER_OUT_OF_RANGE"

	// Storage/Array errors
	ErrCodeArrayNotStopped    ErrorCode = "ARRAY_NOT_STOPPED"
	ErrCodeArrayNotStarted    ErrorCode = "ARRAY_NOT_STARTED"
	ErrCodeArrayInvalidState  ErrorCode = "ARRAY_INVALID_STATE"
	ErrCodeDiskNotFound       ErrorCode = "DISK_NOT_FOUND"
	ErrCodeDiskOffline        ErrorCode = "DISK_OFFLINE"
	ErrCodeDiskReadOnly       ErrorCode = "DISK_READ_ONLY"
	ErrCodeParityCheckActive  ErrorCode = "PARITY_CHECK_ACTIVE"
	ErrCodeParityCheckFailed  ErrorCode = "PARITY_CHECK_FAILED"
	ErrCodeInsufficientSpace  ErrorCode = "INSUFFICIENT_SPACE"

	// Docker errors
	ErrCodeContainerNotFound    ErrorCode = "CONTAINER_NOT_FOUND"
	ErrCodeContainerNotRunning  ErrorCode = "CONTAINER_NOT_RUNNING"
	ErrCodeContainerNotStopped  ErrorCode = "CONTAINER_NOT_STOPPED"
	ErrCodeDockerDaemonError    ErrorCode = "DOCKER_DAEMON_ERROR"
	ErrCodeImageNotFound        ErrorCode = "IMAGE_NOT_FOUND"
	ErrCodeNetworkNotFound      ErrorCode = "NETWORK_NOT_FOUND"

	// VM errors
	ErrCodeVMNotFound      ErrorCode = "VM_NOT_FOUND"
	ErrCodeVMNotRunning    ErrorCode = "VM_NOT_RUNNING"
	ErrCodeVMNotStopped    ErrorCode = "VM_NOT_STOPPED"
	ErrCodeVMConfigError   ErrorCode = "VM_CONFIG_ERROR"
	ErrCodeVirtManagerError ErrorCode = "VIRT_MANAGER_ERROR"

	// System errors
	ErrCodeSystemNotReady     ErrorCode = "SYSTEM_NOT_READY"
	ErrCodeCommandFailed      ErrorCode = "COMMAND_FAILED"
	ErrCodePermissionDenied   ErrorCode = "PERMISSION_DENIED"
	ErrCodeResourceBusy       ErrorCode = "RESOURCE_BUSY"
	ErrCodeHardwareError      ErrorCode = "HARDWARE_ERROR"

	// Async operation errors
	ErrCodeOperationNotFound     ErrorCode = "OPERATION_NOT_FOUND"
	ErrCodeOperationNotCancellable ErrorCode = "OPERATION_NOT_CANCELLABLE"
	ErrCodeOperationConflict     ErrorCode = "OPERATION_CONFLICT"
	ErrCodeOperationTimeout      ErrorCode = "OPERATION_TIMEOUT"
	ErrCodeMaxOperationsReached  ErrorCode = "MAX_OPERATIONS_REACHED"

	// Authentication errors
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	ErrCodeTokenExpired       ErrorCode = "TOKEN_EXPIRED"
	ErrCodeTokenInvalid       ErrorCode = "TOKEN_INVALID"
	ErrCodeSessionExpired     ErrorCode = "SESSION_EXPIRED"

	// Configuration errors
	ErrCodeConfigNotFound     ErrorCode = "CONFIG_NOT_FOUND"
	ErrCodeConfigInvalid      ErrorCode = "CONFIG_INVALID"
	ErrCodeConfigReadOnly     ErrorCode = "CONFIG_READ_ONLY"
)

// APIError represents a structured API error with code, message, and optional details
type APIError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	HTTPStatus int                    `json:"-"`
	Cause      error                  `json:"-"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause error
func (e *APIError) Unwrap() error {
	return e.Cause
}

// WithDetails adds details to the error
func (e *APIError) WithDetails(details map[string]interface{}) *APIError {
	e.Details = details
	return e
}

// WithCause adds a cause error
func (e *APIError) WithCause(cause error) *APIError {
	e.Cause = cause
	return e
}

// NewAPIError creates a new API error
func NewAPIError(code ErrorCode, message string, httpStatus int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// ValidationError represents a field-specific validation error
type ValidationError struct {
	Field   string      `json:"field"`
	Value   interface{} `json:"value,omitempty"`
	Message string      `json:"message"`
	Code    string      `json:"code,omitempty"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// Error implements the error interface
func (v *ValidationErrors) Error() string {
	if len(v.Errors) == 1 {
		return fmt.Sprintf("validation failed for field '%s': %s", v.Errors[0].Field, v.Errors[0].Message)
	}
	return fmt.Sprintf("validation failed for %d fields", len(v.Errors))
}

// AddError adds a validation error
func (v *ValidationErrors) AddError(field, message string) {
	v.Errors = append(v.Errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// AddErrorWithValue adds a validation error with the invalid value
func (v *ValidationErrors) AddErrorWithValue(field, message string, value interface{}) {
	v.Errors = append(v.Errors, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// HasErrors returns true if there are validation errors
func (v *ValidationErrors) HasErrors() bool {
	return len(v.Errors) > 0
}

// ToAPIError converts validation errors to an API error
func (v *ValidationErrors) ToAPIError() *APIError {
	details := map[string]interface{}{
		"validation_errors": v.Errors,
	}
	
	return NewAPIError(
		ErrCodeValidationFailed,
		v.Error(),
		http.StatusBadRequest,
	).WithDetails(details)
}

// Predefined common errors
var (
	ErrArrayNotStopped = NewAPIError(
		ErrCodeArrayNotStopped,
		"Array must be stopped before performing this operation",
		http.StatusConflict,
	)

	ErrArrayNotStarted = NewAPIError(
		ErrCodeArrayNotStarted,
		"Array must be started before performing this operation",
		http.StatusConflict,
	)

	ErrDiskNotFound = NewAPIError(
		ErrCodeDiskNotFound,
		"Disk not found",
		http.StatusNotFound,
	)

	ErrContainerNotFound = NewAPIError(
		ErrCodeContainerNotFound,
		"Container not found",
		http.StatusNotFound,
	)

	ErrVMNotFound = NewAPIError(
		ErrCodeVMNotFound,
		"Virtual machine not found",
		http.StatusNotFound,
	)

	ErrOperationNotFound = NewAPIError(
		ErrCodeOperationNotFound,
		"Operation not found",
		http.StatusNotFound,
	)

	ErrUnauthorized = NewAPIError(
		ErrCodeUnauthorized,
		"Authentication required",
		http.StatusUnauthorized,
	)

	ErrForbidden = NewAPIError(
		ErrCodeForbidden,
		"Access denied",
		http.StatusForbidden,
	)

	ErrRateLimitExceeded = NewAPIError(
		ErrCodeRateLimitExceeded,
		"Rate limit exceeded",
		http.StatusTooManyRequests,
	)

	ErrInternalError = NewAPIError(
		ErrCodeInternalError,
		"Internal server error",
		http.StatusInternalServerError,
	)
)

// Helper functions for creating specific errors

// NewValidationError creates a new validation error
func NewValidationError() *ValidationErrors {
	return &ValidationErrors{
		Errors: make([]ValidationError, 0),
	}
}

// NewDiskNotFoundError creates a disk not found error with disk ID
func NewDiskNotFoundError(diskID string) *APIError {
	return NewAPIError(
		ErrCodeDiskNotFound,
		fmt.Sprintf("Disk '%s' not found", diskID),
		http.StatusNotFound,
	).WithDetails(map[string]interface{}{
		"disk_id": diskID,
	})
}

// NewContainerNotFoundError creates a container not found error with container ID
func NewContainerNotFoundError(containerID string) *APIError {
	return NewAPIError(
		ErrCodeContainerNotFound,
		fmt.Sprintf("Container '%s' not found", containerID),
		http.StatusNotFound,
	).WithDetails(map[string]interface{}{
		"container_id": containerID,
	})
}

// NewVMNotFoundError creates a VM not found error with VM ID
func NewVMNotFoundError(vmID string) *APIError {
	return NewAPIError(
		ErrCodeVMNotFound,
		fmt.Sprintf("Virtual machine '%s' not found", vmID),
		http.StatusNotFound,
	).WithDetails(map[string]interface{}{
		"vm_id": vmID,
	})
}

// NewOperationConflictError creates an operation conflict error
func NewOperationConflictError(operationType string, conflictingOperation string) *APIError {
	return NewAPIError(
		ErrCodeOperationConflict,
		fmt.Sprintf("Cannot start %s operation: conflicting operation %s is already running", operationType, conflictingOperation),
		http.StatusConflict,
	).WithDetails(map[string]interface{}{
		"operation_type":        operationType,
		"conflicting_operation": conflictingOperation,
	})
}

// NewParameterValidationError creates a parameter validation error
func NewParameterValidationError(parameter string, value interface{}, reason string) *APIError {
	return NewAPIError(
		ErrCodeInvalidParameter,
		fmt.Sprintf("Invalid parameter '%s': %s", parameter, reason),
		http.StatusBadRequest,
	).WithDetails(map[string]interface{}{
		"parameter": parameter,
		"value":     value,
		"reason":    reason,
	})
}
