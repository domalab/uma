package dto

import (
	"net/http"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	// Test basic error
	err := NewAPIError(ErrCodeDiskNotFound, "Disk not found", http.StatusNotFound)
	expected := "DISK_NOT_FOUND: Disk not found"
	if err.Error() != expected {
		t.Errorf("Expected %s, got %s", expected, err.Error())
	}
	
	// Test error with cause
	causeErr := &APIError{Code: ErrCodeInternalError, Message: "Internal error"}
	err = err.WithCause(causeErr)
	expectedWithCause := "DISK_NOT_FOUND: Disk not found (caused by: INTERNAL_ERROR: Internal error)"
	if err.Error() != expectedWithCause {
		t.Errorf("Expected %s, got %s", expectedWithCause, err.Error())
	}
}

func TestAPIError_WithDetails(t *testing.T) {
	err := NewAPIError(ErrCodeDiskNotFound, "Disk not found", http.StatusNotFound)
	details := map[string]interface{}{
		"disk_id": "disk1",
		"reason":  "not mounted",
	}
	
	err = err.WithDetails(details)
	
	if err.Details["disk_id"] != "disk1" {
		t.Errorf("Expected disk_id 'disk1', got %v", err.Details["disk_id"])
	}
	
	if err.Details["reason"] != "not mounted" {
		t.Errorf("Expected reason 'not mounted', got %v", err.Details["reason"])
	}
}

func TestValidationErrors_AddError(t *testing.T) {
	validationErrors := NewValidationError()
	
	validationErrors.AddError("field1", "Field1 is required")
	validationErrors.AddError("field2", "Field2 must be positive")
	
	if len(validationErrors.Errors) != 2 {
		t.Errorf("Expected 2 validation errors, got %d", len(validationErrors.Errors))
	}
	
	if validationErrors.Errors[0].Field != "field1" {
		t.Errorf("Expected field 'field1', got %s", validationErrors.Errors[0].Field)
	}
	
	if validationErrors.Errors[0].Message != "Field1 is required" {
		t.Errorf("Expected message 'Field1 is required', got %s", validationErrors.Errors[0].Message)
	}
}

func TestValidationErrors_AddErrorWithValue(t *testing.T) {
	validationErrors := NewValidationError()
	
	validationErrors.AddErrorWithValue("age", "Age must be between 0 and 120", -5)
	
	if len(validationErrors.Errors) != 1 {
		t.Errorf("Expected 1 validation error, got %d", len(validationErrors.Errors))
	}
	
	if validationErrors.Errors[0].Value != -5 {
		t.Errorf("Expected value -5, got %v", validationErrors.Errors[0].Value)
	}
}

func TestValidationErrors_HasErrors(t *testing.T) {
	validationErrors := NewValidationError()
	
	if validationErrors.HasErrors() {
		t.Error("Expected no errors initially")
	}
	
	validationErrors.AddError("field1", "Error message")
	
	if !validationErrors.HasErrors() {
		t.Error("Expected to have errors after adding one")
	}
}

func TestValidationErrors_Error(t *testing.T) {
	validationErrors := NewValidationError()
	
	// Test single error
	validationErrors.AddError("field1", "Field1 is required")
	expected := "validation failed for field 'field1': Field1 is required"
	if validationErrors.Error() != expected {
		t.Errorf("Expected %s, got %s", expected, validationErrors.Error())
	}
	
	// Test multiple errors
	validationErrors.AddError("field2", "Field2 is invalid")
	expected = "validation failed for 2 fields"
	if validationErrors.Error() != expected {
		t.Errorf("Expected %s, got %s", expected, validationErrors.Error())
	}
}

func TestValidationErrors_ToAPIError(t *testing.T) {
	validationErrors := NewValidationError()
	validationErrors.AddError("field1", "Field1 is required")
	validationErrors.AddError("field2", "Field2 is invalid")
	
	apiError := validationErrors.ToAPIError()
	
	if apiError.Code != ErrCodeValidationFailed {
		t.Errorf("Expected code %s, got %s", ErrCodeValidationFailed, apiError.Code)
	}
	
	if apiError.HTTPStatus != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, apiError.HTTPStatus)
	}
	
	if apiError.Details == nil {
		t.Error("Expected details to be set")
	}
	
	validationErrorsDetail, ok := apiError.Details["validation_errors"]
	if !ok {
		t.Error("Expected validation_errors in details")
	}
	
	errors, ok := validationErrorsDetail.([]ValidationError)
	if !ok {
		t.Error("Expected validation_errors to be []ValidationError")
	}
	
	if len(errors) != 2 {
		t.Errorf("Expected 2 validation errors in details, got %d", len(errors))
	}
}

func TestPredefinedErrors(t *testing.T) {
	// Test that predefined errors have correct properties
	if ErrArrayNotStopped.Code != ErrCodeArrayNotStopped {
		t.Errorf("Expected code %s, got %s", ErrCodeArrayNotStopped, ErrArrayNotStopped.Code)
	}
	
	if ErrArrayNotStopped.HTTPStatus != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, ErrArrayNotStopped.HTTPStatus)
	}
	
	if ErrDiskNotFound.Code != ErrCodeDiskNotFound {
		t.Errorf("Expected code %s, got %s", ErrCodeDiskNotFound, ErrDiskNotFound.Code)
	}
	
	if ErrDiskNotFound.HTTPStatus != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, ErrDiskNotFound.HTTPStatus)
	}
}

func TestNewDiskNotFoundError(t *testing.T) {
	diskID := "disk1"
	err := NewDiskNotFoundError(diskID)
	
	if err.Code != ErrCodeDiskNotFound {
		t.Errorf("Expected code %s, got %s", ErrCodeDiskNotFound, err.Code)
	}
	
	expectedMessage := "Disk 'disk1' not found"
	if err.Message != expectedMessage {
		t.Errorf("Expected message %s, got %s", expectedMessage, err.Message)
	}
	
	if err.Details["disk_id"] != diskID {
		t.Errorf("Expected disk_id %s in details, got %v", diskID, err.Details["disk_id"])
	}
}

func TestNewContainerNotFoundError(t *testing.T) {
	containerID := "container123"
	err := NewContainerNotFoundError(containerID)
	
	if err.Code != ErrCodeContainerNotFound {
		t.Errorf("Expected code %s, got %s", ErrCodeContainerNotFound, err.Code)
	}
	
	expectedMessage := "Container 'container123' not found"
	if err.Message != expectedMessage {
		t.Errorf("Expected message %s, got %s", expectedMessage, err.Message)
	}
	
	if err.Details["container_id"] != containerID {
		t.Errorf("Expected container_id %s in details, got %v", containerID, err.Details["container_id"])
	}
}

func TestNewVMNotFoundError(t *testing.T) {
	vmID := "vm-test"
	err := NewVMNotFoundError(vmID)
	
	if err.Code != ErrCodeVMNotFound {
		t.Errorf("Expected code %s, got %s", ErrCodeVMNotFound, err.Code)
	}
	
	expectedMessage := "Virtual machine 'vm-test' not found"
	if err.Message != expectedMessage {
		t.Errorf("Expected message %s, got %s", expectedMessage, err.Message)
	}
	
	if err.Details["vm_id"] != vmID {
		t.Errorf("Expected vm_id %s in details, got %v", vmID, err.Details["vm_id"])
	}
}

func TestNewOperationConflictError(t *testing.T) {
	operationType := "parity_check"
	conflictingOperation := "existing_parity_check"
	err := NewOperationConflictError(operationType, conflictingOperation)
	
	if err.Code != ErrCodeOperationConflict {
		t.Errorf("Expected code %s, got %s", ErrCodeOperationConflict, err.Code)
	}
	
	expectedMessage := "Cannot start parity_check operation: conflicting operation existing_parity_check is already running"
	if err.Message != expectedMessage {
		t.Errorf("Expected message %s, got %s", expectedMessage, err.Message)
	}
	
	if err.Details["operation_type"] != operationType {
		t.Errorf("Expected operation_type %s in details, got %v", operationType, err.Details["operation_type"])
	}
	
	if err.Details["conflicting_operation"] != conflictingOperation {
		t.Errorf("Expected conflicting_operation %s in details, got %v", conflictingOperation, err.Details["conflicting_operation"])
	}
}

func TestNewParameterValidationError(t *testing.T) {
	parameter := "timeout"
	value := -1
	reason := "must be positive"
	err := NewParameterValidationError(parameter, value, reason)
	
	if err.Code != ErrCodeInvalidParameter {
		t.Errorf("Expected code %s, got %s", ErrCodeInvalidParameter, err.Code)
	}
	
	expectedMessage := "Invalid parameter 'timeout': must be positive"
	if err.Message != expectedMessage {
		t.Errorf("Expected message %s, got %s", expectedMessage, err.Message)
	}
	
	if err.Details["parameter"] != parameter {
		t.Errorf("Expected parameter %s in details, got %v", parameter, err.Details["parameter"])
	}
	
	if err.Details["value"] != value {
		t.Errorf("Expected value %v in details, got %v", value, err.Details["value"])
	}
	
	if err.Details["reason"] != reason {
		t.Errorf("Expected reason %s in details, got %v", reason, err.Details["reason"])
	}
}
