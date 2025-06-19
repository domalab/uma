package utils

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/domalab/uma/daemon/services/api/types/requests"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom validators
	validate.RegisterValidation("share_name", validateShareName)
	validate.RegisterValidation("mac_address", validateMACAddress)
	validate.RegisterValidation("ip_address", validateIPAddress)
	validate.RegisterValidation("container_name", validateContainerName)
	validate.RegisterValidation("vm_name", validateVMName)
}

// ValidateStruct validates a struct using the validator package
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

// ValidateShareCreateRequest validates a share creation request
func ValidateShareCreateRequest(req *requests.ShareCreateRequest) error {
	if req.Name == "" {
		return fmt.Errorf("share name is required")
	}

	if !IsValidShareName(req.Name) {
		return fmt.Errorf("invalid share name: must be 1-40 characters, alphanumeric with hyphens and underscores")
	}

	// Validate allocator method
	if req.AllocatorMethod != "" {
		validMethods := []string{"high-water", "most-free", "fill-up"}
		if !contains(validMethods, req.AllocatorMethod) {
			return fmt.Errorf("invalid allocator method: must be one of %v", validMethods)
		}
	}

	// Validate cache usage
	if req.UseCache != "" {
		validCacheOptions := []string{"yes", "no", "only", "prefer"}
		if !contains(validCacheOptions, req.UseCache) {
			return fmt.Errorf("invalid cache usage: must be one of %v", validCacheOptions)
		}
	}

	// Validate SMB security
	if req.SMBSecurity != "" {
		validSecurityOptions := []string{"public", "secure", "private"}
		if !contains(validSecurityOptions, req.SMBSecurity) {
			return fmt.Errorf("invalid SMB security: must be one of %v", validSecurityOptions)
		}
	}

	return nil
}

// ValidateCommandExecuteRequest validates a command execution request
func ValidateCommandExecuteRequest(req *requests.CommandExecuteRequest) error {
	if strings.TrimSpace(req.Command) == "" {
		return fmt.Errorf("command is required")
	}

	// Check for dangerous commands
	if IsCommandBlacklisted(req.Command) {
		return fmt.Errorf("command not allowed for security reasons")
	}

	// Validate timeout
	if req.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}

	if req.Timeout > 300 {
		return fmt.Errorf("timeout cannot exceed 300 seconds")
	}

	return nil
}

// ValidateDockerContainerCreateRequest validates a Docker container creation request
func ValidateDockerContainerCreateRequest(req *requests.DockerContainerCreateRequest) error {
	if req.Name == "" {
		return fmt.Errorf("container name is required")
	}

	if req.Image == "" {
		return fmt.Errorf("container image is required")
	}

	if !IsValidContainerName(req.Name) {
		return fmt.Errorf("invalid container name")
	}

	// Validate restart policy
	if req.RestartPolicy != "" {
		validPolicies := []string{"no", "always", "unless-stopped", "on-failure"}
		if !contains(validPolicies, req.RestartPolicy) {
			return fmt.Errorf("invalid restart policy: must be one of %v", validPolicies)
		}
	}

	return nil
}

// ValidateVMCreateRequest validates a VM creation request
func ValidateVMCreateRequest(req *requests.VMCreateRequest) error {
	if req.Name == "" {
		return fmt.Errorf("VM name is required")
	}

	if !IsValidVMName(req.Name) {
		return fmt.Errorf("invalid VM name")
	}

	if req.CPUs <= 0 {
		return fmt.Errorf("CPU count must be greater than 0")
	}

	if req.Memory <= 0 {
		return fmt.Errorf("memory must be greater than 0")
	}

	// Validate storage configurations
	for i, storage := range req.Storage {
		if err := validateVMStorageConfig(&storage); err != nil {
			return fmt.Errorf("storage config %d: %v", i, err)
		}
	}

	// Validate network configurations
	for i, network := range req.Networks {
		if err := validateVMNetworkConfig(&network); err != nil {
			return fmt.Errorf("network config %d: %v", i, err)
		}
	}

	return nil
}

// IsValidShareName checks if a share name is valid
func IsValidShareName(name string) bool {
	if len(name) < 1 || len(name) > 40 {
		return false
	}

	// Allow alphanumeric characters, hyphens, and underscores
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name)
	return matched
}

// IsValidContainerName checks if a container name is valid
func IsValidContainerName(name string) bool {
	if len(name) < 1 || len(name) > 63 {
		return false
	}

	// Docker container name rules
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`, name)
	return matched
}

// IsValidVMName checks if a VM name is valid
func IsValidVMName(name string) bool {
	if len(name) < 1 || len(name) > 63 {
		return false
	}

	// VM name rules (similar to hostname rules)
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]$|^[a-zA-Z0-9]$`, name)
	return matched
}

// IsCommandBlacklisted checks if a command is blacklisted for security
func IsCommandBlacklisted(command string) bool {
	lowerCommand := strings.ToLower(command)

	// Dangerous commands that should be blocked
	blacklistedCommands := []string{
		"rm -rf /",
		"dd if=",
		"mkfs",
		"fdisk",
		"parted",
		"shutdown",
		"reboot",
		"halt",
		"poweroff",
		"init 0",
		"init 6",
		"systemctl poweroff",
		"systemctl reboot",
		"systemctl halt",
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

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func validateShareName(fl validator.FieldLevel) bool {
	return IsValidShareName(fl.Field().String())
}

func validateMACAddress(fl validator.FieldLevel) bool {
	_, err := net.ParseMAC(fl.Field().String())
	return err == nil
}

func validateIPAddress(fl validator.FieldLevel) bool {
	ip := net.ParseIP(fl.Field().String())
	return ip != nil
}

func validateContainerName(fl validator.FieldLevel) bool {
	return IsValidContainerName(fl.Field().String())
}

func validateVMName(fl validator.FieldLevel) bool {
	return IsValidVMName(fl.Field().String())
}

func validateVMStorageConfig(config *requests.VMStorageConfig) error {
	if config.Type == "" {
		return fmt.Errorf("storage type is required")
	}

	validTypes := []string{"disk", "cdrom", "floppy"}
	if !contains(validTypes, config.Type) {
		return fmt.Errorf("invalid storage type: must be one of %v", validTypes)
	}

	if config.Source == "" {
		return fmt.Errorf("storage source is required")
	}

	if config.Target == "" {
		return fmt.Errorf("storage target is required")
	}

	return nil
}

func validateVMNetworkConfig(config *requests.VMNetworkConfig) error {
	if config.Type == "" {
		return fmt.Errorf("network type is required")
	}

	validTypes := []string{"bridge", "nat", "host"}
	if !contains(validTypes, config.Type) {
		return fmt.Errorf("invalid network type: must be one of %v", validTypes)
	}

	if config.MAC != "" {
		if _, err := net.ParseMAC(config.MAC); err != nil {
			return fmt.Errorf("invalid MAC address: %v", err)
		}
	}

	return nil
}
