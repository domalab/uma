package models

import "time"

// Script represents a user script
type Script struct {
	Name        string            `json:"name"`
	Path        string            `json:"path"`
	Description string            `json:"description"`
	Category    string            `json:"category"` // "system", "maintenance", "backup", "custom"
	Executable  bool              `json:"executable"`
	Size        int64             `json:"size"`
	Modified    time.Time         `json:"modified"`
	Permissions string            `json:"permissions"` // File permissions (e.g., "755")
	Owner       string            `json:"owner"`
	Group       string            `json:"group"`
	Hash        string            `json:"hash"` // File hash for integrity checking
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ScriptExecution represents a script execution record
type ScriptExecution struct {
	ID          string            `json:"id"`
	ScriptName  string            `json:"script_name"`
	ScriptPath  string            `json:"script_path"`
	Arguments   []string          `json:"arguments,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	StartTime   time.Time         `json:"start_time"`
	EndTime     *time.Time        `json:"end_time,omitempty"`
	Duration    int64             `json:"duration,omitempty"` // Duration in milliseconds
	ExitCode    int               `json:"exit_code"`
	Status      string            `json:"status"`           // "running", "completed", "failed", "timeout", "killed"
	Output      string            `json:"output,omitempty"` // Combined stdout/stderr
	Stdout      string            `json:"stdout,omitempty"`
	Stderr      string            `json:"stderr,omitempty"`
	PID         int               `json:"pid,omitempty"`     // Process ID
	User        string            `json:"user,omitempty"`    // User who executed the script
	Background  bool              `json:"background"`        // Was executed in background
	Timeout     int               `json:"timeout,omitempty"` // Timeout in seconds
}

// ScriptSchedule represents a scheduled script execution
type ScriptSchedule struct {
	ID           string            `json:"id"`
	ScriptName   string            `json:"script_name"`
	ScriptPath   string            `json:"script_path"`
	Schedule     string            `json:"schedule"` // Cron expression
	Enabled      bool              `json:"enabled"`
	Arguments    []string          `json:"arguments,omitempty"`
	Environment  map[string]string `json:"environment,omitempty"`
	Timeout      int               `json:"timeout,omitempty"` // Timeout in seconds
	MaxRetries   int               `json:"max_retries,omitempty"`
	RetryDelay   int               `json:"retry_delay,omitempty"` // Delay between retries in seconds
	LastRun      *time.Time        `json:"last_run,omitempty"`
	NextRun      *time.Time        `json:"next_run,omitempty"`
	LastStatus   string            `json:"last_status,omitempty"` // Status of last execution
	RunCount     int               `json:"run_count"`             // Total number of executions
	SuccessCount int               `json:"success_count"`         // Number of successful executions
	FailureCount int               `json:"failure_count"`         // Number of failed executions
	Created      time.Time         `json:"created"`
	Updated      time.Time         `json:"updated"`
}

// ScriptTemplate represents a script template
type ScriptTemplate struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Category    string          `json:"category"`
	Language    string          `json:"language"`             // "bash", "python", "perl", etc.
	Template    string          `json:"template"`             // Script template content
	Parameters  []TemplateParam `json:"parameters,omitempty"` // Template parameters
	Tags        []string        `json:"tags,omitempty"`
	Author      string          `json:"author,omitempty"`
	Version     string          `json:"version,omitempty"`
	Created     time.Time       `json:"created"`
	Updated     time.Time       `json:"updated"`
}

// TemplateParam represents a template parameter
type TemplateParam struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"` // "string", "number", "boolean", "select"
	Description  string      `json:"description"`
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	Options      []string    `json:"options,omitempty"`    // For select type
	Validation   string      `json:"validation,omitempty"` // Validation regex
}

// ScriptStats represents script execution statistics
type ScriptStats struct {
	TotalScripts      int        `json:"total_scripts"`
	ExecutableScripts int        `json:"executable_scripts"`
	ScheduledScripts  int        `json:"scheduled_scripts"`
	TotalExecutions   int        `json:"total_executions"`
	SuccessfulRuns    int        `json:"successful_runs"`
	FailedRuns        int        `json:"failed_runs"`
	AverageRuntime    float64    `json:"average_runtime"` // Average runtime in seconds
	LastExecution     *time.Time `json:"last_execution,omitempty"`
	LastUpdated       time.Time  `json:"last_updated"`
}

// UserScript represents a user script with its metadata and status (legacy format)
type UserScript struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
	Status      string `json:"status"`      // "idle", "running", "completed", "failed"
	LastRun     string `json:"last_run"`    // ISO 8601 format
	LastResult  string `json:"last_result"` // "success", "failed", "unknown"
	PID         int    `json:"pid,omitempty"`
}
