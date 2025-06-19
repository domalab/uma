package responses

import (
	"time"

	"github.com/domalab/uma/daemon/services/api/types/models"
)

// System-related response types

// SystemInfo represents general system information
type SystemInfo struct {
	Hostname     string    `json:"hostname"`
	Kernel       string    `json:"kernel"`
	Architecture string    `json:"architecture"`
	Uptime       string    `json:"uptime"`
	LoadAverage  []float64 `json:"load_average"`
	Timezone     string    `json:"timezone"`
	LastUpdated  time.Time `json:"last_updated"`
}

// CPUInfo represents CPU information
type CPUInfo struct {
	Model       string    `json:"model"`
	Cores       int       `json:"cores"`
	Threads     int       `json:"threads"`
	Usage       float64   `json:"usage"`       // Percentage
	Temperature float64   `json:"temperature"` // Celsius
	Frequency   float64   `json:"frequency"`   // MHz
	LastUpdated time.Time `json:"last_updated"`
}

// MemoryInfo represents memory information
type MemoryInfo struct {
	Total       int64     `json:"total"`      // Bytes
	Used        int64     `json:"used"`       // Bytes
	Free        int64     `json:"free"`       // Bytes
	Available   int64     `json:"available"`  // Bytes
	Cached      int64     `json:"cached"`     // Bytes
	Buffers     int64     `json:"buffers"`    // Bytes
	SwapTotal   int64     `json:"swap_total"` // Bytes
	SwapUsed    int64     `json:"swap_used"`  // Bytes
	SwapFree    int64     `json:"swap_free"`  // Bytes
	Usage       float64   `json:"usage"`      // Percentage
	LastUpdated time.Time `json:"last_updated"`
}

// NetworkInfo represents network interface information
type NetworkInfo struct {
	Interface   string    `json:"interface"`
	Status      string    `json:"status"` // "up", "down"
	IPAddress   string    `json:"ip_address"`
	MACAddress  string    `json:"mac_address"`
	Speed       string    `json:"speed"`
	BytesRx     int64     `json:"bytes_rx"`
	BytesTx     int64     `json:"bytes_tx"`
	PacketsRx   int64     `json:"packets_rx"`
	PacketsTx   int64     `json:"packets_tx"`
	ErrorsRx    int64     `json:"errors_rx"`
	ErrorsTx    int64     `json:"errors_tx"`
	LastUpdated time.Time `json:"last_updated"`
}

// TemperatureInfo represents temperature sensor information
type TemperatureInfo struct {
	Sensor      string    `json:"sensor"`
	Temperature float64   `json:"temperature"` // Celsius
	Critical    float64   `json:"critical"`    // Critical temperature threshold
	Max         float64   `json:"max"`         // Maximum temperature threshold
	Status      string    `json:"status"`      // "normal", "warning", "critical"
	LastUpdated time.Time `json:"last_updated"`
}

// FanInfo represents fan information
type FanInfo struct {
	Fan         string    `json:"fan"`
	Speed       int       `json:"speed"`  // RPM
	Target      int       `json:"target"` // Target RPM
	Status      string    `json:"status"` // "normal", "warning", "critical"
	LastUpdated time.Time `json:"last_updated"`
}

// GPUInfo represents GPU information
type GPUInfo struct {
	Name        string    `json:"name"`
	Driver      string    `json:"driver"`
	Memory      int64     `json:"memory"`      // Bytes
	MemoryUsed  int64     `json:"memory_used"` // Bytes
	Usage       float64   `json:"usage"`       // Percentage
	Temperature float64   `json:"temperature"` // Celsius
	PowerDraw   float64   `json:"power_draw"`  // Watts
	LastUpdated time.Time `json:"last_updated"`
}

// UPSInfo represents UPS information
type UPSInfo struct {
	Status         string    `json:"status"`          // "online", "on_battery", "low_battery", "charging"
	BatteryCharge  float64   `json:"battery_charge"`  // Percentage
	BatteryRuntime int       `json:"battery_runtime"` // Minutes
	Load           float64   `json:"load"`            // Percentage
	Voltage        float64   `json:"voltage"`         // Volts
	LastUpdated    time.Time `json:"last_updated"`
}

// PowerOperationResponse represents the response from power operations
type PowerOperationResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	OperationID string `json:"operation_id,omitempty"`
}

// ScriptExecuteResponse represents a response from script execution
type ScriptExecuteResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	ExecutionID string `json:"execution_id,omitempty"`
	PID         int    `json:"pid,omitempty"`
}

// ScriptStatusResponse represents the status of a script execution
type ScriptStatusResponse struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	PID       int    `json:"pid,omitempty"`
	StartTime string `json:"start_time,omitempty"`
	Duration  string `json:"duration,omitempty"`
	ExitCode  int    `json:"exit_code,omitempty"`
}

// ScriptLogsResponse represents the logs from a script execution
type ScriptLogsResponse struct {
	Name string   `json:"name"`
	Logs []string `json:"logs"`
}

// ScriptListResponse represents a list of available user scripts
type ScriptListResponse struct {
	Scripts []models.UserScript `json:"scripts"`
}

// CommandExecuteResponse represents a command execution response
type CommandExecuteResponse struct {
	ExitCode        int    `json:"exit_code"`
	Stdout          string `json:"stdout"`
	Stderr          string `json:"stderr"`
	ExecutionTimeMs int64  `json:"execution_time_ms"`
	Command         string `json:"command"`
	WorkingDir      string `json:"working_directory,omitempty"`
}

// LogsResponse represents system logs response
type LogsResponse struct {
	LogType    string    `json:"log_type"`
	Lines      []string  `json:"lines"`
	TotalLines int       `json:"total_lines"`
	Timestamp  time.Time `json:"timestamp"`
	Source     string    `json:"source,omitempty"`
}
