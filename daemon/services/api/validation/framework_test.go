package validation

import (
	"fmt"
	"testing"
)

func TestValidationFramework_BasicValidation(t *testing.T) {
	vf := NewValidationFramework()

	t.Run("ValidateStruct_Success", func(t *testing.T) {
		type TestStruct struct {
			Name  string `validate:"required,min=3,max=50"`
			Email string `validate:"required,email"`
			Age   int    `validate:"min=0,max=120"`
		}

		validData := TestStruct{
			Name:  "John Doe",
			Email: "john@example.com",
			Age:   30,
		}

		result := vf.ValidateStruct(validData)
		if !result.Valid {
			t.Errorf("Expected validation to pass, but got errors: %v", result.Errors)
		}
	})

	t.Run("ValidateStruct_Failure", func(t *testing.T) {
		type TestStruct struct {
			Name  string `validate:"required,min=3,max=50"`
			Email string `validate:"required,email"`
			Age   int    `validate:"min=0,max=120"`
		}

		invalidData := TestStruct{
			Name:  "Jo", // Too short
			Email: "invalid-email",
			Age:   150, // Too old
		}

		result := vf.ValidateStruct(invalidData)
		if result.Valid {
			t.Error("Expected validation to fail")
		}

		if len(result.Errors) == 0 {
			t.Error("Expected validation errors")
		}

		// Check that we have errors for all invalid fields
		errorFields := make(map[string]bool)
		for _, err := range result.Errors {
			errorFields[err.Field] = true
		}

		expectedFields := []string{"Name", "Email", "Age"}
		for _, field := range expectedFields {
			if !errorFields[field] {
				t.Errorf("Expected validation error for field: %s", field)
			}
		}
	})
}

func TestValidationFramework_CustomValidators(t *testing.T) {
	vf := NewValidationFramework()

	t.Run("ShareName_Valid", func(t *testing.T) {
		validNames := []string{
			"valid-share",
			"share_name",
			"ShareName123",
			"a",
			"1234567890123456789012345678901234567890", // 40 chars
		}

		for _, name := range validNames {
			if !vf.isValidShareName(name) {
				t.Errorf("Expected share name to be valid: %s", name)
			}
		}
	})

	t.Run("ShareName_Invalid", func(t *testing.T) {
		invalidNames := []string{
			"",           // Empty
			"share name", // Space
			"share@name", // Special character
			"12345678901234567890123456789012345678901", // 41 chars (too long)
			"share.name", // Dot
		}

		for _, name := range invalidNames {
			if vf.isValidShareName(name) {
				t.Errorf("Expected share name to be invalid: %s", name)
			}
		}
	})

	t.Run("ContainerName_Valid", func(t *testing.T) {
		validNames := []string{
			"container",
			"my-container",
			"container_name",
			"container.name",
			"a1",
			"container123",
		}

		for _, name := range validNames {
			if !vf.isValidContainerName(name) {
				t.Errorf("Expected container name to be valid: %s", name)
			}
		}
	})

	t.Run("ContainerName_Invalid", func(t *testing.T) {
		invalidNames := []string{
			"",               // Empty
			"-container",     // Starts with dash
			"container name", // Space
			"container@name", // Special character
			"1234567890123456789012345678901234567890123456789012345678901234", // 64 chars (too long)
		}

		for _, name := range invalidNames {
			if vf.isValidContainerName(name) {
				t.Errorf("Expected container name to be invalid: %s", name)
			}
		}
	})

	t.Run("VMName_Valid", func(t *testing.T) {
		validNames := []string{
			"vm",
			"my-vm",
			"vm123",
			"a",
			"vm-name-123",
		}

		for _, name := range validNames {
			if !vf.isValidVMName(name) {
				t.Errorf("Expected VM name to be valid: %s", name)
			}
		}
	})

	t.Run("VMName_Invalid", func(t *testing.T) {
		invalidNames := []string{
			"",        // Empty
			"-vm",     // Starts with dash
			"vm-",     // Ends with dash
			"vm_name", // Underscore
			"vm name", // Space
			"vm.name", // Dot
			"1234567890123456789012345678901234567890123456789012345678901234", // 64 chars (too long)
		}

		for _, name := range invalidNames {
			if vf.isValidVMName(name) {
				t.Errorf("Expected VM name to be invalid: %s", name)
			}
		}
	})
}

func TestValidationFramework_CommandBlacklist(t *testing.T) {
	vf := NewValidationFramework()

	t.Run("BlacklistedCommands", func(t *testing.T) {
		blacklistedCommands := []string{
			"rm -rf /",
			"dd if=/dev/zero of=/dev/sda",
			"mkfs.ext4 /dev/sda1",
			"fdisk /dev/sda",
			"shutdown now",
			"reboot",
			"halt",
			"poweroff",
			"init 0",
			"init 6",
			"systemctl poweroff",
			"systemctl reboot",
			"chmod 777 /etc/passwd",
			"chown root:root /etc/passwd",
			"sudo su -",
			"su -",
			"echo test >/dev/sda",
		}

		for _, cmd := range blacklistedCommands {
			if !vf.isCommandBlacklisted(cmd) {
				t.Errorf("Expected command to be blacklisted: %s", cmd)
			}
		}
	})

	t.Run("SafeCommands", func(t *testing.T) {
		safeCommands := []string{
			"ls -la",
			"cat /etc/hostname",
			"ps aux",
			"df -h",
			"free -m",
			"uptime",
			"whoami",
			"date",
			"echo hello",
			"grep pattern file.txt",
		}

		for _, cmd := range safeCommands {
			if vf.isCommandBlacklisted(cmd) {
				t.Errorf("Expected command to be safe: %s", cmd)
			}
		}
	})
}

func TestValidationFramework_FieldValidation(t *testing.T) {
	vf := NewValidationFramework()

	t.Run("ValidateField_Success", func(t *testing.T) {
		result := vf.ValidateField("test@example.com", "email")
		if !result.Valid {
			t.Errorf("Expected field validation to pass, but got errors: %v", result.Errors)
		}
	})

	t.Run("ValidateField_Failure", func(t *testing.T) {
		result := vf.ValidateField("invalid-email", "email")
		if result.Valid {
			t.Error("Expected field validation to fail")
		}

		if len(result.Errors) == 0 {
			t.Error("Expected validation errors")
		}
	})
}

func TestValidationFramework_CustomRules(t *testing.T) {
	vf := NewValidationFramework()

	t.Run("RegisterCustomRule", func(t *testing.T) {
		rule := ValidationRule{
			Name:        "even_number",
			Description: "Value must be an even number",
			Validator: func(value interface{}) error {
				if num, ok := value.(int); ok {
					if num%2 != 0 {
						return fmt.Errorf("number must be even")
					}
					return nil
				}
				return fmt.Errorf("value must be an integer")
			},
		}

		err := vf.RegisterCustomRule(rule)
		if err != nil {
			t.Errorf("Failed to register custom rule: %v", err)
		}

		// Test the custom rule
		err = vf.ValidateWithCustomRule(4, "even_number")
		if err != nil {
			t.Errorf("Expected even number validation to pass: %v", err)
		}

		err = vf.ValidateWithCustomRule(3, "even_number")
		if err == nil {
			t.Error("Expected odd number validation to fail")
		}
	})

	t.Run("RegisterCustomRule_EmptyName", func(t *testing.T) {
		rule := ValidationRule{
			Name: "",
			Validator: func(value interface{}) error {
				return nil
			},
		}

		err := vf.RegisterCustomRule(rule)
		if err == nil {
			t.Error("Expected error when registering rule with empty name")
		}
	})

	t.Run("RegisterCustomRule_NilValidator", func(t *testing.T) {
		rule := ValidationRule{
			Name:      "test_rule",
			Validator: nil,
		}

		err := vf.RegisterCustomRule(rule)
		if err == nil {
			t.Error("Expected error when registering rule with nil validator")
		}
	})
}

func TestValidationFramework_HelperFunctions(t *testing.T) {
	vf := NewValidationFramework()

	t.Run("ValidateStringLength", func(t *testing.T) {
		err := vf.ValidateStringLength("hello", 3, 10)
		if err != nil {
			t.Errorf("Expected string length validation to pass: %v", err)
		}

		err = vf.ValidateStringLength("hi", 3, 10)
		if err == nil {
			t.Error("Expected string length validation to fail for too short string")
		}

		err = vf.ValidateStringLength("this is too long", 3, 10)
		if err == nil {
			t.Error("Expected string length validation to fail for too long string")
		}
	})

	t.Run("ValidateIntRange", func(t *testing.T) {
		err := vf.ValidateIntRange(5, 1, 10)
		if err != nil {
			t.Errorf("Expected int range validation to pass: %v", err)
		}

		err = vf.ValidateIntRange(0, 1, 10)
		if err == nil {
			t.Error("Expected int range validation to fail for too small value")
		}

		err = vf.ValidateIntRange(15, 1, 10)
		if err == nil {
			t.Error("Expected int range validation to fail for too large value")
		}
	})

	t.Run("ValidateEnum", func(t *testing.T) {
		allowedValues := []string{"red", "green", "blue"}

		err := vf.ValidateEnum("red", allowedValues)
		if err != nil {
			t.Errorf("Expected enum validation to pass: %v", err)
		}

		err = vf.ValidateEnum("yellow", allowedValues)
		if err == nil {
			t.Error("Expected enum validation to fail for invalid value")
		}
	})

	t.Run("ValidateRegex", func(t *testing.T) {
		err := vf.ValidateRegex("hello123", `^[a-z]+[0-9]+$`, "lowercase letters followed by numbers")
		if err != nil {
			t.Errorf("Expected regex validation to pass: %v", err)
		}

		err = vf.ValidateRegex("Hello123", `^[a-z]+[0-9]+$`, "lowercase letters followed by numbers")
		if err == nil {
			t.Error("Expected regex validation to fail for invalid pattern")
		}

		err = vf.ValidateRegex("hello", `[`, "invalid regex")
		if err == nil {
			t.Error("Expected regex validation to fail for invalid regex pattern")
		}
	})
}

// Test the global convenience functions
func TestGlobalValidationFunctions(t *testing.T) {
	t.Run("IsValidShareName", func(t *testing.T) {
		if !IsValidShareName("valid-share") {
			t.Error("Expected valid share name to pass")
		}

		if IsValidShareName("invalid share") {
			t.Error("Expected invalid share name to fail")
		}
	})

	t.Run("IsValidContainerName", func(t *testing.T) {
		if !IsValidContainerName("valid-container") {
			t.Error("Expected valid container name to pass")
		}

		if IsValidContainerName("-invalid") {
			t.Error("Expected invalid container name to fail")
		}
	})

	t.Run("IsValidVMName", func(t *testing.T) {
		if !IsValidVMName("valid-vm") {
			t.Error("Expected valid VM name to pass")
		}

		if IsValidVMName("invalid_vm") {
			t.Error("Expected invalid VM name to fail")
		}
	})

	t.Run("IsCommandBlacklisted", func(t *testing.T) {
		if !IsCommandBlacklisted("rm -rf /") {
			t.Error("Expected dangerous command to be blacklisted")
		}

		if IsCommandBlacklisted("ls -la") {
			t.Error("Expected safe command to not be blacklisted")
		}
	})
}
