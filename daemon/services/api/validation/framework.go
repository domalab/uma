package validation

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

// ValidationFramework provides comprehensive validation capabilities
type ValidationFramework struct {
	validator *validator.Validate
	rules     map[string]ValidationRule
}

// ValidationRule represents a custom validation rule
type ValidationRule struct {
	Name        string
	Description string
	Validator   func(interface{}) error
}

// ValidationError represents a validation error with context
type ValidationError struct {
	Field   string      `json:"field"`
	Value   interface{} `json:"value"`
	Tag     string      `json:"tag"`
	Message string      `json:"message"`
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

// NewValidationFramework creates a new validation framework
func NewValidationFramework() *ValidationFramework {
	vf := &ValidationFramework{
		validator: validator.New(),
		rules:     make(map[string]ValidationRule),
	}

	// Register custom validators
	vf.registerBuiltinValidators()

	return vf
}

// registerBuiltinValidators registers built-in custom validators
func (vf *ValidationFramework) registerBuiltinValidators() {
	// Share name validation
	vf.validator.RegisterValidation("share_name", func(fl validator.FieldLevel) bool {
		return vf.isValidShareName(fl.Field().String())
	})

	// Container name validation
	vf.validator.RegisterValidation("container_name", func(fl validator.FieldLevel) bool {
		return vf.isValidContainerName(fl.Field().String())
	})

	// VM name validation
	vf.validator.RegisterValidation("vm_name", func(fl validator.FieldLevel) bool {
		return vf.isValidVMName(fl.Field().String())
	})

	// MAC address validation
	vf.validator.RegisterValidation("mac_address", func(fl validator.FieldLevel) bool {
		_, err := net.ParseMAC(fl.Field().String())
		return err == nil
	})

	// IP address validation
	vf.validator.RegisterValidation("ip_address", func(fl validator.FieldLevel) bool {
		ip := net.ParseIP(fl.Field().String())
		return ip != nil
	})

	// Port validation
	vf.validator.RegisterValidation("port", func(fl validator.FieldLevel) bool {
		port, err := strconv.Atoi(fl.Field().String())
		return err == nil && port > 0 && port <= 65535
	})

	// Path validation (no path traversal)
	vf.validator.RegisterValidation("safe_path", func(fl validator.FieldLevel) bool {
		path := fl.Field().String()
		return !strings.Contains(path, "..") && !strings.Contains(path, "//")
	})

	// Command validation (not blacklisted)
	vf.validator.RegisterValidation("safe_command", func(fl validator.FieldLevel) bool {
		return !vf.isCommandBlacklisted(fl.Field().String())
	})

	// Duration validation
	vf.validator.RegisterValidation("duration", func(fl validator.FieldLevel) bool {
		_, err := time.ParseDuration(fl.Field().String())
		return err == nil
	})

	// Alphanumeric with specific characters
	vf.validator.RegisterValidation("alphanum_dash", func(fl validator.FieldLevel) bool {
		matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, fl.Field().String())
		return matched
	})
}

// ValidateStruct validates a struct and returns detailed results
func (vf *ValidationFramework) ValidateStruct(s interface{}) ValidationResult {
	err := vf.validator.Struct(s)
	if err == nil {
		return ValidationResult{Valid: true}
	}

	var validationErrors []ValidationError

	if validatorErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validatorErrors {
			validationErrors = append(validationErrors, ValidationError{
				Field:   fieldError.Field(),
				Value:   fieldError.Value(),
				Tag:     fieldError.Tag(),
				Message: vf.getErrorMessage(fieldError),
			})
		}
	} else {
		// Handle other types of errors
		validationErrors = append(validationErrors, ValidationError{
			Field:   "unknown",
			Value:   nil,
			Tag:     "error",
			Message: err.Error(),
		})
	}

	return ValidationResult{
		Valid:  false,
		Errors: validationErrors,
	}
}

// ValidateField validates a single field value
func (vf *ValidationFramework) ValidateField(value interface{}, tag string) ValidationResult {
	err := vf.validator.Var(value, tag)
	if err == nil {
		return ValidationResult{Valid: true}
	}

	var validationErrors []ValidationError

	if validatorErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validatorErrors {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "value",
				Value:   fieldError.Value(),
				Tag:     fieldError.Tag(),
				Message: vf.getErrorMessage(fieldError),
			})
		}
	}

	return ValidationResult{
		Valid:  false,
		Errors: validationErrors,
	}
}

// RegisterCustomRule registers a custom validation rule
func (vf *ValidationFramework) RegisterCustomRule(rule ValidationRule) error {
	if rule.Name == "" {
		return fmt.Errorf("validation rule name cannot be empty")
	}

	if rule.Validator == nil {
		return fmt.Errorf("validation rule validator function cannot be nil")
	}

	vf.rules[rule.Name] = rule

	// Register with the validator
	return vf.validator.RegisterValidation(rule.Name, func(fl validator.FieldLevel) bool {
		err := rule.Validator(fl.Field().Interface())
		return err == nil
	})
}

// ValidateWithCustomRule validates a value using a custom rule
func (vf *ValidationFramework) ValidateWithCustomRule(value interface{}, ruleName string) error {
	rule, exists := vf.rules[ruleName]
	if !exists {
		return fmt.Errorf("validation rule '%s' not found", ruleName)
	}

	return rule.Validator(value)
}

// Built-in validation functions

func (vf *ValidationFramework) isValidShareName(name string) bool {
	if len(name) < 1 || len(name) > 40 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name)
	return matched
}

func (vf *ValidationFramework) isValidContainerName(name string) bool {
	if len(name) < 1 || len(name) > 63 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`, name)
	return matched
}

func (vf *ValidationFramework) isValidVMName(name string) bool {
	if len(name) < 1 || len(name) > 63 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]$|^[a-zA-Z0-9]$`, name)
	return matched
}

func (vf *ValidationFramework) isCommandBlacklisted(command string) bool {
	lowerCommand := strings.ToLower(command)

	blacklistedCommands := []string{
		"rm -rf /", "dd if=", "mkfs", "fdisk", "parted",
		"shutdown", "reboot", "halt", "poweroff",
		"init 0", "init 6", "systemctl poweroff",
		"systemctl reboot", "systemctl halt",
	}

	for _, blocked := range blacklistedCommands {
		if strings.Contains(lowerCommand, blocked) {
			return true
		}
	}

	// Additional security checks
	if strings.Contains(lowerCommand, ">/dev/") ||
		strings.Contains(lowerCommand, "rm -rf") ||
		strings.Contains(lowerCommand, "chmod 777") ||
		strings.Contains(lowerCommand, "chown root") ||
		strings.Contains(lowerCommand, "sudo su") ||
		strings.Contains(lowerCommand, "su -") {
		return true
	}

	return false
}

// getErrorMessage returns a human-readable error message for validation errors
func (vf *ValidationFramework) getErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", fe.Field(), fe.Param())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fe.Field())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", fe.Field())
	case "share_name":
		return fmt.Sprintf("%s must be a valid share name (1-40 characters, alphanumeric with hyphens and underscores)", fe.Field())
	case "container_name":
		return fmt.Sprintf("%s must be a valid container name", fe.Field())
	case "vm_name":
		return fmt.Sprintf("%s must be a valid VM name", fe.Field())
	case "mac_address":
		return fmt.Sprintf("%s must be a valid MAC address", fe.Field())
	case "ip_address":
		return fmt.Sprintf("%s must be a valid IP address", fe.Field())
	case "port":
		return fmt.Sprintf("%s must be a valid port number (1-65535)", fe.Field())
	case "safe_path":
		return fmt.Sprintf("%s contains invalid path characters", fe.Field())
	case "safe_command":
		return fmt.Sprintf("%s contains blacklisted command", fe.Field())
	case "duration":
		return fmt.Sprintf("%s must be a valid duration", fe.Field())
	case "alphanum_dash":
		return fmt.Sprintf("%s must contain only alphanumeric characters, hyphens, and underscores", fe.Field())
	default:
		return fmt.Sprintf("%s failed validation for tag '%s'", fe.Field(), fe.Tag())
	}
}

// Validation helper functions

// ValidateStringLength validates string length
func (vf *ValidationFramework) ValidateStringLength(value string, min, max int) error {
	length := len(value)
	if length < min {
		return fmt.Errorf("string must be at least %d characters long", min)
	}
	if length > max {
		return fmt.Errorf("string must be at most %d characters long", max)
	}
	return nil
}

// ValidateIntRange validates integer range
func (vf *ValidationFramework) ValidateIntRange(value, min, max int) error {
	if value < min {
		return fmt.Errorf("value must be at least %d", min)
	}
	if value > max {
		return fmt.Errorf("value must be at most %d", max)
	}
	return nil
}

// ValidateEnum validates that a value is in a list of allowed values
func (vf *ValidationFramework) ValidateEnum(value string, allowedValues []string) error {
	for _, allowed := range allowedValues {
		if value == allowed {
			return nil
		}
	}
	return fmt.Errorf("value must be one of: %s", strings.Join(allowedValues, ", "))
}

// ValidateRegex validates that a string matches a regex pattern
func (vf *ValidationFramework) ValidateRegex(value, pattern, description string) error {
	matched, err := regexp.MatchString(pattern, value)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %v", err)
	}
	if !matched {
		return fmt.Errorf("value does not match required pattern: %s", description)
	}
	return nil
}

// Global validation framework instance
var globalValidationFramework = NewValidationFramework()

// GetValidationFramework returns the global validation framework
func GetValidationFramework() *ValidationFramework {
	return globalValidationFramework
}

// Convenience functions

// ValidateStruct validates a struct using the global framework
func ValidateStruct(s interface{}) ValidationResult {
	return globalValidationFramework.ValidateStruct(s)
}

// ValidateField validates a field using the global framework
func ValidateField(value interface{}, tag string) ValidationResult {
	return globalValidationFramework.ValidateField(value, tag)
}

// IsValidShareName checks if a share name is valid
func IsValidShareName(name string) bool {
	return globalValidationFramework.isValidShareName(name)
}

// IsValidContainerName checks if a container name is valid
func IsValidContainerName(name string) bool {
	return globalValidationFramework.isValidContainerName(name)
}

// IsValidVMName checks if a VM name is valid
func IsValidVMName(name string) bool {
	return globalValidationFramework.isValidVMName(name)
}

// IsCommandBlacklisted checks if a command is blacklisted
func IsCommandBlacklisted(command string) bool {
	return globalValidationFramework.isCommandBlacklisted(command)
}
