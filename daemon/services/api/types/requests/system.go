package requests

// System-related request types

// SystemShutdownRequest represents a request to shutdown the system
type SystemShutdownRequest struct {
	DelaySeconds int    `json:"delay_seconds"` // Delay before shutdown (0-300 seconds)
	Message      string `json:"message"`       // Message to display to users
	Force        bool   `json:"force"`         // Force shutdown even if users are logged in
}

// SystemRebootRequest represents a request to reboot the system
type SystemRebootRequest struct {
	DelaySeconds int    `json:"delay_seconds"` // Delay before reboot (0-300 seconds)
	Message      string `json:"message"`       // Message to display to users
	Force        bool   `json:"force"`         // Force reboot even if users are logged in
}

// SystemSleepRequest represents a request to put the system to sleep
type SystemSleepRequest struct {
	Type string `json:"type"` // "suspend", "hibernate", or "hybrid"
}

// SystemWakeRequest represents a request to wake a system via Wake-on-LAN
type SystemWakeRequest struct {
	TargetMAC   string `json:"target_mac"`   // MAC address to wake
	BroadcastIP string `json:"broadcast_ip"` // Broadcast IP (optional, defaults to 255.255.255.255)
	Port        int    `json:"port"`         // Port for WOL packet (optional, defaults to 9)
	RepeatCount int    `json:"repeat_count"` // Number of packets to send (optional, defaults to 3)
}

// CommandExecuteRequest represents a command execution request
type CommandExecuteRequest struct {
	Command          string `json:"command"`
	Timeout          int    `json:"timeout,omitempty"`           // Timeout in seconds, default 30
	WorkingDirectory string `json:"working_directory,omitempty"` // Optional working directory
}

// ScriptExecuteRequest represents a user script execution request
type ScriptExecuteRequest struct {
	Arguments   []string          `json:"arguments,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Background  bool              `json:"background,omitempty"`
	Timeout     int               `json:"timeout,omitempty"` // Timeout in seconds
}

// LogsRequest represents a request for system logs
type LogsRequest struct {
	LogType     string `json:"log_type,omitempty"`     // "syslog", "kernel", "auth", etc.
	Lines       int    `json:"lines,omitempty"`        // Number of lines to return
	Follow      bool   `json:"follow,omitempty"`       // Follow log output
	Since       string `json:"since,omitempty"`        // ISO 8601 timestamp
	GrepFilter  string `json:"grep_filter,omitempty"`  // Filter pattern
	CustomPath  string `json:"custom_path,omitempty"`  // Custom log file path
	Directory   string `json:"directory,omitempty"`    // Directory to search
	Recursive   bool   `json:"recursive,omitempty"`    // Recursive search
	FilePattern string `json:"file_pattern,omitempty"` // File pattern to match
	MaxFiles    int    `json:"max_files,omitempty"`    // Maximum files to return
}
