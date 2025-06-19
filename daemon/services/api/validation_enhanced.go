package api

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/domalab/uma/daemon/dto"
	"github.com/domalab/uma/daemon/services/async"
)

// Enhanced validation patterns
var (
	// Docker container ID pattern (12-64 hex characters)
	containerIDPattern = regexp.MustCompile(`^[a-f0-9]{12,64}$`)

	// VM ID pattern (UUID or name)
	vmIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	// Disk ID pattern (e.g., disk1, parity, cache)
	diskIDPattern = regexp.MustCompile(`^(disk\d+|parity\d*|cache\d*)$`)

	// Operation ID pattern (UUID)
	operationIDPattern = regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)
)

// Custom validator functions

// Enhanced validation functions using custom error types

// ValidateAsyncOperationRequest validates async operation requests
func (h *HTTPServer) ValidateAsyncOperationRequest(req *async.OperationRequest) *dto.APIError {
	validationErrors := dto.NewValidationError()

	// Validate operation type
	if req.Type == "" {
		validationErrors.AddError("type", "Operation type is required")
	} else {
		// Validate operation type directly
		validTypes := []string{
			string(async.TypeParityCheck),
			string(async.TypeParityCorrect),
			string(async.TypeArrayStart),
			string(async.TypeArrayStop),
			string(async.TypeDiskScan),
			string(async.TypeSMARTScan),
			string(async.TypeSystemReboot),
			string(async.TypeSystemShutdown),
			string(async.TypeBulkContainer),
			string(async.TypeBulkVM),
		}

		valid := false
		for _, validType := range validTypes {
			if string(req.Type) == validType {
				valid = true
				break
			}
		}

		if !valid {
			validationErrors.AddErrorWithValue("type",
				fmt.Sprintf("Invalid operation type. Valid types: %s", strings.Join(validTypes, ", ")),
				req.Type)
		}
	}

	// Validate description length
	if len(req.Description) > 500 {
		validationErrors.AddErrorWithValue("description",
			"Description must be 500 characters or less",
			len(req.Description))
	}

	// Validate parameters based on operation type
	if err := h.validateOperationParameters(req.Type, req.Parameters); err != nil {
		validationErrors.AddError("parameters", err.Error())
	}

	if validationErrors.HasErrors() {
		return validationErrors.ToAPIError()
	}

	return nil
}

// validateOperationParameters validates operation-specific parameters
func (h *HTTPServer) validateOperationParameters(opType async.OperationType, params map[string]interface{}) error {
	switch opType {
	case async.TypeParityCheck, async.TypeParityCorrect:
		return h.validateParityCheckParams(params)
	case async.TypeBulkContainer:
		return h.validateBulkContainerParams(params)
	case async.TypeBulkVM:
		return h.validateBulkVMParams(params)
	case async.TypeArrayStart, async.TypeArrayStop:
		return h.validateArrayOperationParams(params)
	}
	return nil
}

// validateParityCheckParams validates parity check parameters
func (h *HTTPServer) validateParityCheckParams(params map[string]interface{}) error {
	if params == nil {
		return nil
	}

	// Validate check type
	if checkType, ok := params["type"].(string); ok {
		validTypes := []string{"check", "correct"}
		valid := false
		for _, validType := range validTypes {
			if checkType == validType {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid parity check type '%s', must be 'check' or 'correct'", checkType)
		}
	}

	// Validate priority
	if priority, ok := params["priority"].(string); ok {
		validPriorities := []string{"low", "normal", "high"}
		valid := false
		for _, validPriority := range validPriorities {
			if priority == validPriority {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid priority '%s', must be 'low', 'normal', or 'high'", priority)
		}
	}

	return nil
}

// validateBulkContainerParams validates bulk container operation parameters
func (h *HTTPServer) validateBulkContainerParams(params map[string]interface{}) error {
	if params == nil {
		return fmt.Errorf("container_ids parameter is required for bulk container operations")
	}

	// Validate container IDs
	containerIDs, ok := params["container_ids"].([]string)
	if !ok {
		return fmt.Errorf("container_ids must be an array of strings")
	}

	if len(containerIDs) == 0 {
		return fmt.Errorf("at least one container ID is required")
	}

	if len(containerIDs) > 50 {
		return fmt.Errorf("maximum 50 containers allowed per bulk operation")
	}

	// Validate each container ID
	seen := make(map[string]bool)
	for i, id := range containerIDs {
		if id == "" {
			return fmt.Errorf("container ID at index %d cannot be empty", i)
		}

		if seen[id] {
			return fmt.Errorf("duplicate container ID: %s", id)
		}
		seen[id] = true

		// Validate container ID format
		if len(id) < 12 || !containerIDPattern.MatchString(id) {
			return fmt.Errorf("invalid container ID format: %s", id)
		}
	}

	// Validate operation
	if operation, ok := params["operation"].(string); ok {
		validOperations := []string{"start", "stop", "restart", "pause", "resume"}
		valid := false
		for _, validOp := range validOperations {
			if operation == validOp {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid operation '%s', must be one of: %s",
				operation, strings.Join(validOperations, ", "))
		}
	} else {
		return fmt.Errorf("operation parameter is required")
	}

	return nil
}

// validateBulkVMParams validates bulk VM operation parameters
func (h *HTTPServer) validateBulkVMParams(params map[string]interface{}) error {
	if params == nil {
		return fmt.Errorf("vm_ids parameter is required for bulk VM operations")
	}

	// Validate VM IDs
	vmIDs, ok := params["vm_ids"].([]string)
	if !ok {
		return fmt.Errorf("vm_ids must be an array of strings")
	}

	if len(vmIDs) == 0 {
		return fmt.Errorf("at least one VM ID is required")
	}

	if len(vmIDs) > 20 {
		return fmt.Errorf("maximum 20 VMs allowed per bulk operation")
	}

	// Validate each VM ID
	seen := make(map[string]bool)
	for i, id := range vmIDs {
		if id == "" {
			return fmt.Errorf("VM ID at index %d cannot be empty", i)
		}

		if seen[id] {
			return fmt.Errorf("duplicate VM ID: %s", id)
		}
		seen[id] = true

		// Validate VM ID format
		if !vmIDPattern.MatchString(id) {
			return fmt.Errorf("invalid VM ID format: %s", id)
		}
	}

	return nil
}

// validateArrayOperationParams validates array operation parameters
func (h *HTTPServer) validateArrayOperationParams(params map[string]interface{}) error {
	if params == nil {
		return nil
	}

	// Validate maintenance mode flag
	if _, ok := params["maintenance_mode"]; ok {
		if _, isBool := params["maintenance_mode"].(bool); !isBool {
			return fmt.Errorf("maintenance_mode must be a boolean value")
		}
	}

	// Validate check filesystem flag
	if _, ok := params["check_filesystem"]; ok {
		if _, isBool := params["check_filesystem"].(bool); !isBool {
			return fmt.Errorf("check_filesystem must be a boolean value")
		}
	}

	return nil
}

// ValidateDiskID validates disk ID format and returns appropriate error
func (h *HTTPServer) ValidateDiskID(diskID string) *dto.APIError {
	if diskID == "" {
		return dto.NewParameterValidationError("disk_id", diskID, "disk ID cannot be empty")
	}

	if !diskIDPattern.MatchString(diskID) {
		return dto.NewParameterValidationError("disk_id", diskID,
			"invalid disk ID format (expected: disk1, parity, cache, etc.)")
	}

	return nil
}

// ValidateContainerID validates container ID format and returns appropriate error
func (h *HTTPServer) ValidateContainerID(containerID string) *dto.APIError {
	if containerID == "" {
		return dto.NewParameterValidationError("container_id", containerID, "container ID cannot be empty")
	}

	if len(containerID) < 12 || !containerIDPattern.MatchString(containerID) {
		return dto.NewParameterValidationError("container_id", containerID,
			"invalid container ID format (expected: 12-64 hex characters)")
	}

	return nil
}

// ValidateVMID validates VM ID format and returns appropriate error
func (h *HTTPServer) ValidateVMID(vmID string) *dto.APIError {
	if vmID == "" {
		return dto.NewParameterValidationError("vm_id", vmID, "VM ID cannot be empty")
	}

	if !vmIDPattern.MatchString(vmID) {
		return dto.NewParameterValidationError("vm_id", vmID,
			"invalid VM ID format (expected: alphanumeric, underscore, or dash)")
	}

	return nil
}

// ValidateOperationID validates operation ID format and returns appropriate error
func (h *HTTPServer) ValidateOperationID(operationID string) *dto.APIError {
	if operationID == "" {
		return dto.NewParameterValidationError("operation_id", operationID, "operation ID cannot be empty")
	}

	if !operationIDPattern.MatchString(operationID) {
		return dto.NewParameterValidationError("operation_id", operationID,
			"invalid operation ID format (expected: UUID)")
	}

	return nil
}
