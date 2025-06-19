package services

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/services/api/utils"
)

// CommandService handles system command execution business logic
type CommandService struct {
	api utils.APIInterface
}

// NewCommandService creates a new command service
func NewCommandService(api utils.APIInterface) *CommandService {
	return &CommandService{
		api: api,
	}
}

// CommandExecuteRequest represents a command execution request (local type for transition)
type CommandExecuteRequest struct {
	Command     string            `json:"command"`
	Arguments   []string          `json:"arguments,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	WorkingDir  string            `json:"working_dir,omitempty"`
	Timeout     int               `json:"timeout,omitempty"` // seconds
}

// CommandExecuteResponse represents a command execution response (local type for transition)
type CommandExecuteResponse struct {
	Success    bool   `json:"success"`
	Output     string `json:"output"`
	Error      string `json:"error,omitempty"`
	ExitCode   int    `json:"exit_code"`
	Duration   string `json:"duration"`
	ExecutedAt string `json:"executed_at"`
}

// ExecuteCommand executes a system command with safety checks
func (s *CommandService) ExecuteCommand(request CommandExecuteRequest) CommandExecuteResponse {
	startTime := time.Now()

	// Prepare command
	fullCommand := request.Command
	if len(request.Arguments) > 0 {
		fullCommand += " " + strings.Join(request.Arguments, " ")
	}

	// Security check - prevent dangerous commands
	if s.isCommandBlacklisted(request.Command) {
		return CommandExecuteResponse{
			Success:    false,
			Error:      "Command is blacklisted for security reasons",
			ExitCode:   1,
			Duration:   time.Since(startTime).String(),
			ExecutedAt: startTime.Format(time.RFC3339),
		}
	}

	// Create command
	var cmd *exec.Cmd
	if len(request.Arguments) > 0 {
		cmd = exec.Command(request.Command, request.Arguments...)
	} else {
		cmd = exec.Command("sh", "-c", request.Command)
	}

	// Set working directory if specified
	if request.WorkingDir != "" {
		cmd.Dir = request.WorkingDir
	}

	// Set environment variables
	if len(request.Environment) > 0 {
		env := os.Environ()
		for key, value := range request.Environment {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = env
	}

	// Execute command with timeout
	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	response := CommandExecuteResponse{
		Success:    err == nil,
		Output:     string(output),
		ExitCode:   exitCode,
		Duration:   time.Since(startTime).String(),
		ExecutedAt: startTime.Format(time.RFC3339),
	}

	if err != nil {
		response.Error = err.Error()
	}

	logger.Blue("Command executed: %s (exit code: %d, duration: %s)", fullCommand, exitCode, response.Duration)

	return response
}

// isCommandBlacklisted checks if a command is blacklisted for security
func (s *CommandService) isCommandBlacklisted(command string) bool {
	// List of dangerous commands that should not be allowed
	blacklistedCommands := []string{
		"rm -rf /",
		"rm -rf /*",
		"dd if=/dev/zero",
		"mkfs",
		"fdisk",
		"parted",
		"wipefs",
		"shred",
		":(){ :|:& };:",
		"chmod -R 777 /",
		"chown -R root:root /",
		"mv / /dev/null",
		"cat /dev/urandom > /dev/sda",
		"echo 1 > /proc/sys/kernel/sysrq",
		"echo c > /proc/sysrq-trigger",
	}

	// Check exact matches
	for _, blacklisted := range blacklistedCommands {
		if strings.Contains(strings.ToLower(command), strings.ToLower(blacklisted)) {
			return true
		}
	}

	// Check for dangerous patterns
	dangerousPatterns := []string{
		"rm -rf",
		"rm -r /",
		"rm -f /",
		"> /dev/sd",
		"dd if=",
		"dd of=/dev/",
		"mkfs.",
		"format ",
		"fdisk /dev/",
		"parted /dev/",
		"wipefs /dev/",
		"shred /dev/",
		"chmod 777 /",
		"chmod -R 777",
		"chown root /",
		"chown -R root",
		"mv / ",
		"mv /* ",
		"cat /dev/urandom >",
		"cat /dev/zero >",
		"echo c > /proc/sysrq-trigger",
		"reboot",
		"shutdown",
		"halt",
		"poweroff",
		"init 0",
		"init 6",
		"systemctl reboot",
		"systemctl poweroff",
		"systemctl halt",
	}

	commandLower := strings.ToLower(command)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(commandLower, pattern) {
			return true
		}
	}

	return false
}

// ExecuteSystemShutdown executes a system shutdown command
func (s *CommandService) ExecuteSystemShutdown(delaySeconds int, message string, force bool) error {
	// Build shutdown command
	cmd := "shutdown"

	var args []string
	if delaySeconds > 0 {
		args = append(args, fmt.Sprintf("+%d", delaySeconds/60)) // Convert to minutes
	} else {
		args = append(args, "now")
	}

	if message != "" {
		args = append(args, fmt.Sprintf("\"%s\"", message))
	}

	if force {
		args = append(args, "-f") // Force shutdown
	}

	// Execute shutdown command
	request := CommandExecuteRequest{
		Command:   cmd,
		Arguments: args,
		Timeout:   30,
	}

	response := s.ExecuteCommand(request)
	if !response.Success {
		return fmt.Errorf("shutdown failed: %s", response.Error)
	}

	logger.Blue("System shutdown initiated with %d second delay", delaySeconds)
	return nil
}

// ExecuteSystemReboot executes a system reboot command
func (s *CommandService) ExecuteSystemReboot(delaySeconds int, message string, force bool) error {
	// Build reboot command
	cmd := "reboot"

	var args []string
	if delaySeconds > 0 {
		// Use shutdown command with reboot flag for delayed reboot
		cmd = "shutdown"
		args = append(args, "-r", fmt.Sprintf("+%d", delaySeconds/60)) // Convert to minutes
	}

	if message != "" {
		args = append(args, fmt.Sprintf("\"%s\"", message))
	}

	if force {
		args = append(args, "-f") // Force reboot
	}

	// Execute reboot command
	request := CommandExecuteRequest{
		Command:   cmd,
		Arguments: args,
		Timeout:   30,
	}

	response := s.ExecuteCommand(request)
	if !response.Success {
		return fmt.Errorf("reboot failed: %s", response.Error)
	}

	logger.Blue("System reboot initiated with %d second delay", delaySeconds)
	return nil
}

// ExecuteSystemSleep executes a system sleep command
func (s *CommandService) ExecuteSystemSleep(sleepType string) error {
	var cmd string

	switch sleepType {
	case "suspend":
		cmd = "systemctl suspend"
	case "hibernate":
		cmd = "systemctl hibernate"
	case "hybrid":
		cmd = "systemctl hybrid-sleep"
	default:
		return fmt.Errorf("invalid sleep type: %s", sleepType)
	}

	// Execute sleep command
	request := CommandExecuteRequest{
		Command: cmd,
		Timeout: 30,
	}

	response := s.ExecuteCommand(request)
	if !response.Success {
		return fmt.Errorf("sleep failed: %s", response.Error)
	}

	logger.Blue("System sleep (%s) initiated", sleepType)
	return nil
}

// GetAllowedCommands returns a list of commands that are safe to execute
func (s *CommandService) GetAllowedCommands() []string {
	return []string{
		"ls",
		"cat",
		"grep",
		"find",
		"ps",
		"top",
		"htop",
		"df",
		"du",
		"free",
		"uptime",
		"whoami",
		"id",
		"date",
		"uname",
		"lscpu",
		"lsblk",
		"lsusb",
		"lspci",
		"ip",
		"netstat",
		"ss",
		"ping",
		"curl",
		"wget",
		"systemctl status",
		"journalctl",
		"dmesg",
		"sensors",
		"nvidia-smi",
		"docker ps",
		"docker images",
		"docker stats",
		"virsh list",
		"virsh dominfo",
		"smartctl",
		"mdadm --detail",
		"zpool status",
		"zfs list",
	}
}
