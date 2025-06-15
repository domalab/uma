package command

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// CommandExecutor handles secure command execution
type CommandExecutor struct {
	allowedCommands map[string]bool
	timeout         time.Duration
}

// CommandRequest represents a command execution request
type CommandRequest struct {
	Command     string            `json:"command"`
	Arguments   []string          `json:"arguments,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	WorkingDir  string            `json:"working_dir,omitempty"`
	Timeout     int               `json:"timeout,omitempty"` // seconds
	Background  bool              `json:"background,omitempty"`
}

// ContainerCommandRequest represents a container command execution request
type ContainerCommandRequest struct {
	ContainerID string            `json:"container_id"`
	Command     string            `json:"command"`
	Arguments   []string          `json:"arguments,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	WorkingDir  string            `json:"working_dir,omitempty"`
	User        string            `json:"user,omitempty"`
	Interactive bool              `json:"interactive,omitempty"`
	TTY         bool              `json:"tty,omitempty"`
	Timeout     int               `json:"timeout,omitempty"` // seconds
}

// CommandResponse represents the result of command execution
type CommandResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	Output      string `json:"output,omitempty"`
	Error       string `json:"error,omitempty"`
	ExitCode    int    `json:"exit_code"`
	ExecutionID string `json:"execution_id,omitempty"`
	PID         int    `json:"pid,omitempty"`
	Duration    string `json:"duration,omitempty"`
}

// NewCommandExecutor creates a new command executor with security settings
func NewCommandExecutor() *CommandExecutor {
	// Define allowed commands for security
	allowedCommands := map[string]bool{
		// System information
		"ps":       true,
		"top":      true,
		"htop":     true,
		"free":     true,
		"df":       true,
		"du":       true,
		"lsblk":    true,
		"mount":    true,
		"uptime":   true,
		"whoami":   true,
		"id":       true,
		"uname":    true,
		"hostname": true,
		"date":     true,

		// File operations (read-only)
		"ls":     true,
		"cat":    true,
		"head":   true,
		"tail":   true,
		"grep":   true,
		"find":   true,
		"locate": true,
		"which":  true,
		"file":   true,
		"stat":   true,

		// Network
		"ping":     true,
		"curl":     true,
		"wget":     true,
		"netstat":  true,
		"ss":       true,
		"nslookup": true,
		"dig":      true,

		// Docker
		"docker": true,

		// Unraid specific
		"mdcmd":    true,
		"emcmd":    true,
		"sensors":  true,
		"smartctl": true,

		// Text processing
		"awk":  true,
		"sed":  true,
		"sort": true,
		"uniq": true,
		"wc":   true,
		"cut":  true,

		// Archive
		"tar":    true,
		"gzip":   true,
		"gunzip": true,
		"zip":    true,
		"unzip":  true,
	}

	return &CommandExecutor{
		allowedCommands: allowedCommands,
		timeout:         30 * time.Second, // Default 30 second timeout
	}
}

// ExecuteCommand executes a shell command with security checks
func (ce *CommandExecutor) ExecuteCommand(req CommandRequest) (*CommandResponse, error) {
	startTime := time.Now()

	// Security validation
	if err := ce.validateCommand(req.Command); err != nil {
		return &CommandResponse{
			Success:  false,
			Message:  "Command validation failed",
			Error:    err.Error(),
			ExitCode: -1,
			Duration: time.Since(startTime).String(),
		}, nil
	}

	// Set timeout
	timeout := ce.timeout
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	// Create execution context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Build command
	var cmd *exec.Cmd
	if len(req.Arguments) > 0 {
		cmd = exec.CommandContext(ctx, req.Command, req.Arguments...)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", req.Command)
	}

	// Set working directory
	if req.WorkingDir != "" {
		cmd.Dir = req.WorkingDir
	}

	// Set environment variables
	if len(req.Environment) > 0 {
		env := os.Environ()
		for key, value := range req.Environment {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = env
	}

	executionID := fmt.Sprintf("cmd_%d", time.Now().Unix())

	if req.Background {
		// Execute in background
		if err := cmd.Start(); err != nil {
			return &CommandResponse{
				Success:     false,
				Message:     "Failed to start command",
				Error:       err.Error(),
				ExitCode:    -1,
				ExecutionID: executionID,
				Duration:    time.Since(startTime).String(),
			}, nil
		}

		return &CommandResponse{
			Success:     true,
			Message:     "Command started in background",
			ExecutionID: executionID,
			PID:         cmd.Process.Pid,
			Duration:    time.Since(startTime).String(),
		}, nil
	} else {
		// Execute synchronously
		output, err := cmd.CombinedOutput()
		duration := time.Since(startTime)

		response := &CommandResponse{
			Success:     err == nil,
			Output:      string(output),
			ExecutionID: executionID,
			Duration:    duration.String(),
		}

		if err != nil {
			response.Message = "Command execution failed"
			response.Error = err.Error()
			if exitError, ok := err.(*exec.ExitError); ok {
				response.ExitCode = exitError.ExitCode()
			} else {
				response.ExitCode = -1
			}
		} else {
			response.Message = "Command executed successfully"
			response.ExitCode = 0
		}

		return response, nil
	}
}

// ExecuteContainerCommand executes a command inside a Docker container
func (ce *CommandExecutor) ExecuteContainerCommand(req ContainerCommandRequest) (*CommandResponse, error) {
	startTime := time.Now()

	// Build docker exec command
	args := []string{"exec"}

	if req.Interactive {
		args = append(args, "-i")
	}

	if req.TTY {
		args = append(args, "-t")
	}

	if req.User != "" {
		args = append(args, "--user", req.User)
	}

	if req.WorkingDir != "" {
		args = append(args, "--workdir", req.WorkingDir)
	}

	// Add environment variables
	for key, value := range req.Environment {
		args = append(args, "--env", fmt.Sprintf("%s=%s", key, value))
	}

	// Add container ID
	args = append(args, req.ContainerID)

	// Add command and arguments
	args = append(args, req.Command)
	args = append(args, req.Arguments...)

	// Set timeout
	timeout := ce.timeout
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	// Create execution context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Execute docker command
	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	executionID := fmt.Sprintf("container_cmd_%d", time.Now().Unix())

	response := &CommandResponse{
		Success:     err == nil,
		Output:      string(output),
		ExecutionID: executionID,
		Duration:    duration.String(),
	}

	if err != nil {
		response.Message = "Container command execution failed"
		response.Error = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			response.ExitCode = exitError.ExitCode()
		} else {
			response.ExitCode = -1
		}
	} else {
		response.Message = "Container command executed successfully"
		response.ExitCode = 0
	}

	return response, nil
}

// validateCommand performs security validation on commands
func (ce *CommandExecutor) validateCommand(command string) error {
	// Extract the base command (first word)
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	baseCommand := parts[0]

	// Check if command is in allowed list
	if !ce.allowedCommands[baseCommand] {
		return fmt.Errorf("command '%s' is not allowed", baseCommand)
	}

	// Check for dangerous patterns
	dangerousPatterns := []string{
		"rm -rf",
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
		"kill -9",
		"killall",
		"pkill",
		"chmod 777",
		"chown root",
		"su -",
		"sudo",
		"passwd",
		"userdel",
		"groupdel",
		"crontab",
		"at ",
		"batch",
		"nohup",
		"&",
		"|",
		";",
		"&&",
		"||",
		"`",
		"$(",
		"${",
	}

	commandLower := strings.ToLower(command)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(commandLower, pattern) {
			return fmt.Errorf("command contains dangerous pattern: %s", pattern)
		}
	}

	// Check for path traversal attempts
	if strings.Contains(command, "..") {
		return fmt.Errorf("path traversal detected")
	}

	// Check for suspicious file paths
	suspiciousPaths := []string{
		"/etc/passwd",
		"/etc/shadow",
		"/etc/sudoers",
		"/root/",
		"/home/",
		"/var/log/",
		"/proc/",
		"/sys/",
		"/dev/",
	}

	for _, path := range suspiciousPaths {
		if strings.Contains(commandLower, path) {
			return fmt.Errorf("access to sensitive path detected: %s", path)
		}
	}

	return nil
}

// GetAllowedCommands returns the list of allowed commands
func (ce *CommandExecutor) GetAllowedCommands() []string {
	commands := make([]string, 0, len(ce.allowedCommands))
	for cmd := range ce.allowedCommands {
		commands = append(commands, cmd)
	}
	return commands
}
