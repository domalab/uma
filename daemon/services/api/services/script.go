package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/services/api/utils"
)

// ScriptService handles user script-related business logic
type ScriptService struct {
	api utils.APIInterface
}

// NewScriptService creates a new script service
func NewScriptService(api utils.APIInterface) *ScriptService {
	return &ScriptService{
		api: api,
	}
}

// Script represents a user script
type Script struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	LastRun     string `json:"last_run,omitempty"`
	ExitCode    int    `json:"exit_code,omitempty"`
	Enabled     bool   `json:"enabled"`
}

// ScriptExecuteRequest represents a script execution request
type ScriptExecuteRequest struct {
	Background bool              `json:"background,omitempty"`
	Arguments  map[string]string `json:"arguments,omitempty"`
}

// ScriptExecuteResponse represents a script execution response
type ScriptExecuteResponse struct {
	Success     bool   `json:"success"`
	ScriptName  string `json:"script_name"`
	ExecutionID string `json:"execution_id"`
	StartedAt   string `json:"started_at"`
	Status      string `json:"status"`
	Message     string `json:"message"`
	PID         int    `json:"pid,omitempty"`
}

// ScriptStatusResponse represents a script status response
type ScriptStatusResponse struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	PID       int    `json:"pid,omitempty"`
	StartTime string `json:"start_time,omitempty"`
	Runtime   string `json:"runtime,omitempty"`
	ExitCode  int    `json:"exit_code,omitempty"`
}

// GetUserScripts returns a list of available user scripts
func (s *ScriptService) GetUserScripts() ([]Script, error) {
	scriptsDir := "/boot/config/plugins/user.scripts/scripts"

	// Check if scripts directory exists
	if _, err := os.Stat(scriptsDir); os.IsNotExist(err) {
		return []Script{}, nil
	}

	entries, err := os.ReadDir(scriptsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read scripts directory: %v", err)
	}

	var scripts []Script
	for _, entry := range entries {
		if entry.IsDir() {
			scriptName := entry.Name()
			script := Script{
				Name:        scriptName,
				Description: s.getScriptDescription(scriptName),
				Status:      s.getScriptCurrentStatus(scriptName),
				Enabled:     true, // Assume enabled if directory exists
			}

			// Get last run information
			lastRun, exitCode := s.getScriptLastRun(scriptName)
			script.LastRun = lastRun
			if exitCode != -1 {
				script.ExitCode = exitCode
			}

			scripts = append(scripts, script)
		}
	}

	return scripts, nil
}

// GetScriptStatus returns the current status of a specific script
func (s *ScriptService) GetScriptStatus(scriptName string) (*ScriptStatusResponse, error) {
	status := &ScriptStatusResponse{
		Name:   scriptName,
		Status: "idle",
	}

	// Check if script is running
	pidPath := fmt.Sprintf("/tmp/user.scripts.%s.pid", scriptName)
	if pidData, err := os.ReadFile(pidPath); err == nil {
		pidStr := strings.TrimSpace(string(pidData))
		if pid, err := strconv.Atoi(pidStr); err == nil {
			if s.isProcessRunning(pid) {
				status.Status = "running"
				status.PID = pid

				// Get process start time and calculate runtime
				if startTime, err := s.getProcessStartTime(pid); err == nil {
					status.StartTime = startTime.Format(time.RFC3339)
					status.Runtime = time.Since(startTime).String()
				}
			}
		}
	}

	// If not running, try to get exit code from last run
	if status.Status == "idle" {
		if exitCode, err := s.getScriptExitCode(scriptName); err == nil {
			status.ExitCode = exitCode
		}
	}

	return status, nil
}

// GetScriptLogs returns the logs for a specific script
func (s *ScriptService) GetScriptLogs(scriptName string) ([]string, error) {
	logPath := fmt.Sprintf("/tmp/user.scripts/tmpScripts/%s.log", scriptName)

	content, err := os.ReadFile(logPath)
	if err != nil {
		// Return empty logs if file doesn't exist
		return []string{}, nil
	}

	// Split content into lines and filter out empty lines
	lines := strings.Split(string(content), "\n")
	var logs []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			logs = append(logs, line)
		}
	}

	return logs, nil
}

// ExecuteScript executes a user script
func (s *ScriptService) ExecuteScript(scriptName string, req ScriptExecuteRequest) (*ScriptExecuteResponse, error) {
	// Validate script exists
	scriptPath := fmt.Sprintf("/boot/config/plugins/user.scripts/scripts/%s/script", scriptName)
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("script '%s' not found", scriptName)
	}

	// Check if script is already running
	if status, _ := s.GetScriptStatus(scriptName); status.Status == "running" {
		return nil, fmt.Errorf("script '%s' is already running (PID: %d)", scriptName, status.PID)
	}

	// Generate execution ID
	executionID := fmt.Sprintf("exec_%s_%d", scriptName, time.Now().Unix())

	// Prepare command
	cmd := exec.Command("/bin/bash", scriptPath)
	cmd.Dir = filepath.Dir(scriptPath)

	// Set environment variables if provided
	if len(req.Arguments) > 0 {
		env := os.Environ()
		for key, value := range req.Arguments {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = env
	}

	// Create log directory if it doesn't exist
	logDir := "/tmp/user.scripts/tmpScripts"
	os.MkdirAll(logDir, 0755)

	// Set up logging
	logPath := fmt.Sprintf("%s/%s.log", logDir, scriptName)
	logFile, err := os.Create(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %v", err)
	}

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// Start the script
	if err := cmd.Start(); err != nil {
		logFile.Close()
		return nil, fmt.Errorf("failed to start script: %v", err)
	}

	// Write PID file
	pidPath := fmt.Sprintf("/tmp/user.scripts.%s.pid", scriptName)
	pidFile, err := os.Create(pidPath)
	if err != nil {
		logger.Yellow("Failed to create PID file for script %s: %v", scriptName, err)
	} else {
		fmt.Fprintf(pidFile, "%d", cmd.Process.Pid)
		pidFile.Close()
	}

	response := &ScriptExecuteResponse{
		Success:     true,
		ScriptName:  scriptName,
		ExecutionID: executionID,
		StartedAt:   time.Now().Format(time.RFC3339),
		Status:      "running",
		Message:     "Script execution started",
		PID:         cmd.Process.Pid,
	}

	// If background execution, don't wait for completion
	if req.Background {
		// Start a goroutine to clean up when the process finishes
		go func() {
			defer logFile.Close()
			cmd.Wait()

			// Write exit code to status file
			statusPath := fmt.Sprintf("/tmp/user.scripts.%s.status", scriptName)
			if statusFile, err := os.Create(statusPath); err == nil {
				fmt.Fprintf(statusFile, "%d", cmd.ProcessState.ExitCode())
				statusFile.Close()
			}

			// Remove PID file
			os.Remove(pidPath)
		}()
	} else {
		// Wait for completion
		defer logFile.Close()
		err := cmd.Wait()

		// Write exit code to status file
		statusPath := fmt.Sprintf("/tmp/user.scripts.%s.status", scriptName)
		if statusFile, err := os.Create(statusPath); err == nil {
			fmt.Fprintf(statusFile, "%d", cmd.ProcessState.ExitCode())
			statusFile.Close()
		}

		// Remove PID file
		os.Remove(pidPath)

		if err != nil {
			response.Success = false
			response.Status = "failed"
			response.Message = fmt.Sprintf("Script execution failed: %v", err)
		} else {
			response.Status = "completed"
			response.Message = "Script execution completed"
		}
	}

	logger.Blue("Script %s executed with execution ID: %s", scriptName, executionID)
	return response, nil
}

// StopScript stops a running script
func (s *ScriptService) StopScript(scriptName string) error {
	// Check if script is running
	pidPath := fmt.Sprintf("/tmp/user.scripts.%s.pid", scriptName)
	pidData, err := os.ReadFile(pidPath)
	if err != nil {
		return fmt.Errorf("script '%s' is not running", scriptName)
	}

	pidStr := strings.TrimSpace(string(pidData))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return fmt.Errorf("invalid PID file for script '%s'", scriptName)
	}

	// Check if process is actually running
	if !s.isProcessRunning(pid) {
		// Clean up stale PID file
		os.Remove(pidPath)
		return fmt.Errorf("script '%s' is not running (stale PID file)", scriptName)
	}

	// Try to terminate the process gracefully first
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %v", pid, err)
	}

	// Send SIGTERM
	if err := process.Signal(os.Interrupt); err != nil {
		// If SIGTERM fails, try SIGKILL
		if err := process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process %d: %v", pid, err)
		}
	}

	// Wait a moment for the process to exit
	time.Sleep(1 * time.Second)

	// Clean up PID file
	os.Remove(pidPath)

	logger.Blue("Script %s stopped (PID: %d)", scriptName, pid)
	return nil
}

// Helper methods

// getScriptDescription gets the description of a script
func (s *ScriptService) getScriptDescription(scriptName string) string {
	descPath := fmt.Sprintf("/boot/config/plugins/user.scripts/scripts/%s/description", scriptName)
	content, err := os.ReadFile(descPath)
	if err != nil {
		return "No description available"
	}
	return strings.TrimSpace(string(content))
}

// getScriptCurrentStatus gets the current status of a script
func (s *ScriptService) getScriptCurrentStatus(scriptName string) string {
	// Check if script is currently running by looking for PID file
	pidPath := fmt.Sprintf("/tmp/user.scripts.%s.pid", scriptName)
	if pidData, err := os.ReadFile(pidPath); err == nil {
		pidStr := strings.TrimSpace(string(pidData))
		if pid, err := strconv.Atoi(pidStr); err == nil {
			if s.isProcessRunning(pid) {
				return "running"
			}
		}
		// Clean up stale PID file
		os.Remove(pidPath)
	}
	return "idle"
}

// getScriptLastRun gets the last run time and exit code of a script
func (s *ScriptService) getScriptLastRun(scriptName string) (string, int) {
	// Check for log file to determine last run
	logPath := fmt.Sprintf("/tmp/user.scripts/tmpScripts/%s.log", scriptName)
	if stat, err := os.Stat(logPath); err == nil {
		lastRun := stat.ModTime().Format(time.RFC3339)

		// Try to get exit code
		if exitCode, err := s.getScriptExitCode(scriptName); err == nil {
			return lastRun, exitCode
		}

		return lastRun, -1
	}
	return "", -1
}

// isProcessRunning checks if a process with the given PID is running
func (s *ScriptService) isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix systems, Signal(0) can be used to check if process exists
	err = process.Signal(os.Signal(nil))
	return err == nil
}

// getProcessStartTime gets the start time of a process
func (s *ScriptService) getProcessStartTime(pid int) (time.Time, error) {
	// Read process stat file
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	content, err := os.ReadFile(statPath)
	if err != nil {
		return time.Time{}, err
	}

	// Parse stat file to get start time
	fields := strings.Fields(string(content))
	if len(fields) < 22 {
		return time.Time{}, fmt.Errorf("invalid stat file format")
	}

	// Field 22 is start time in clock ticks since boot
	startTicks, err := strconv.ParseInt(fields[21], 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	// Get system boot time
	bootTime, err := s.getSystemBootTime()
	if err != nil {
		return time.Time{}, err
	}

	// Calculate process start time
	// Clock ticks per second is typically 100 (USER_HZ)
	clockTicks := int64(100)
	startTime := bootTime.Add(time.Duration(startTicks/clockTicks) * time.Second)

	return startTime, nil
}

// getSystemBootTime gets the system boot time
func (s *ScriptService) getSystemBootTime() (time.Time, error) {
	content, err := os.ReadFile("/proc/stat")
	if err != nil {
		return time.Time{}, err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "btime ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				bootTimestamp, err := strconv.ParseInt(fields[1], 10, 64)
				if err != nil {
					return time.Time{}, err
				}
				return time.Unix(bootTimestamp, 0), nil
			}
		}
	}

	return time.Time{}, fmt.Errorf("boot time not found in /proc/stat")
}

// getScriptExitCode gets the exit code of the last script execution
func (s *ScriptService) getScriptExitCode(scriptName string) (int, error) {
	// Try to read exit code from a status file (if User Scripts plugin creates one)
	statusPath := fmt.Sprintf("/tmp/user.scripts.%s.status", scriptName)
	if content, err := os.ReadFile(statusPath); err == nil {
		return strconv.Atoi(strings.TrimSpace(string(content)))
	}

	// Fallback: assume success if no status file
	return 0, fmt.Errorf("no exit code available")
}
