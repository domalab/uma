package diagnostics

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/domalab/uma/daemon/lib"
	"github.com/domalab/uma/daemon/logger"
)

// DiagnosticsManager provides system diagnostics and health checks
type DiagnosticsManager struct{}

// HealthCheck represents a system health check
type HealthCheck struct {
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Status      string    `json:"status"`
	Message     string    `json:"message"`
	Details     string    `json:"details,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	Critical    bool      `json:"critical"`
	Remediation string    `json:"remediation,omitempty"`
}

// SystemHealth represents overall system health
type SystemHealth struct {
	OverallStatus string        `json:"overall_status"`
	Timestamp     time.Time     `json:"timestamp"`
	Checks        []HealthCheck `json:"checks"`
	Summary       HealthSummary `json:"summary"`
}

// HealthSummary provides a summary of health check results
type HealthSummary struct {
	Total    int `json:"total"`
	Passed   int `json:"passed"`
	Warning  int `json:"warning"`
	Critical int `json:"critical"`
	Failed   int `json:"failed"`
}

// DiagnosticInfo represents diagnostic information
type DiagnosticInfo struct {
	Category    string                 `json:"category"`
	Name        string                 `json:"name"`
	Value       interface{}            `json:"value"`
	Unit        string                 `json:"unit,omitempty"`
	Status      string                 `json:"status"`
	Threshold   map[string]interface{} `json:"threshold,omitempty"`
	Description string                 `json:"description,omitempty"`
}

// RepairAction represents an automated repair action
type RepairAction struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Commands    []string `json:"commands"`
	Dangerous   bool     `json:"dangerous"`
	Category    string   `json:"category"`
}

// NewDiagnosticsManager creates a new diagnostics manager
func NewDiagnosticsManager() *DiagnosticsManager {
	return &DiagnosticsManager{}
}

// RunHealthChecks performs comprehensive system health checks
func (d *DiagnosticsManager) RunHealthChecks() (*SystemHealth, error) {
	health := &SystemHealth{
		Timestamp: time.Now(),
		Checks:    make([]HealthCheck, 0),
		Summary:   HealthSummary{},
	}

	// Run all health checks
	checks := []func() []HealthCheck{
		d.checkArrayHealth,
		d.checkDiskHealth,
		d.checkMemoryHealth,
		d.checkCPUHealth,
		d.checkNetworkHealth,
		d.checkDockerHealth,
		d.checkSystemServices,
		d.checkFileSystemHealth,
		d.checkLogHealth,
		d.checkSecurityHealth,
	}

	for _, checkFunc := range checks {
		results := checkFunc()
		health.Checks = append(health.Checks, results...)
	}

	// Calculate summary
	d.calculateHealthSummary(health)

	// Determine overall status
	d.determineOverallStatus(health)

	return health, nil
}

// GetDiagnosticInfo returns detailed diagnostic information
func (d *DiagnosticsManager) GetDiagnosticInfo() ([]DiagnosticInfo, error) {
	diagnostics := make([]DiagnosticInfo, 0)

	// System information
	diagnostics = append(diagnostics, d.getSystemDiagnostics()...)

	// Storage diagnostics
	diagnostics = append(diagnostics, d.getStorageDiagnostics()...)

	// Network diagnostics
	diagnostics = append(diagnostics, d.getNetworkDiagnostics()...)

	// Performance diagnostics
	diagnostics = append(diagnostics, d.getPerformanceDiagnostics()...)

	return diagnostics, nil
}

// GetAvailableRepairs returns available automated repair actions
func (d *DiagnosticsManager) GetAvailableRepairs() []RepairAction {
	return []RepairAction{
		{
			Name:        "clear_system_logs",
			Description: "Clear old system logs to free up space",
			Commands:    []string{"journalctl --vacuum-time=7d", "find /var/log -name '*.log' -mtime +7 -delete"},
			Dangerous:   false,
			Category:    "maintenance",
		},
		{
			Name:        "restart_docker",
			Description: "Restart Docker service to resolve container issues",
			Commands:    []string{"systemctl restart docker"},
			Dangerous:   false,
			Category:    "docker",
		},
		{
			Name:        "clear_docker_cache",
			Description: "Clean up Docker images and containers",
			Commands:    []string{"docker system prune -f"},
			Dangerous:   false,
			Category:    "docker",
		},
		{
			Name:        "fix_array_permissions",
			Description: "Fix Unraid array permissions",
			Commands:    []string{"newperms /mnt/user"},
			Dangerous:   false,
			Category:    "array",
		},
		{
			Name:        "restart_network",
			Description: "Restart network services",
			Commands:    []string{"systemctl restart networking"},
			Dangerous:   true,
			Category:    "network",
		},
	}
}

// ExecuteRepair executes an automated repair action
func (d *DiagnosticsManager) ExecuteRepair(repairName string) error {
	repairs := d.GetAvailableRepairs()
	
	var repair *RepairAction
	for _, r := range repairs {
		if r.Name == repairName {
			repair = &r
			break
		}
	}

	if repair == nil {
		return fmt.Errorf("repair action not found: %s", repairName)
	}

	logger.Blue("Executing repair action: %s", repair.Name)

	for _, command := range repair.Commands {
		parts := strings.Fields(command)
		if len(parts) == 0 {
			continue
		}

		output := lib.GetCmdOutput(parts[0], parts[1:]...)
		
		// Check for errors in output
		for _, line := range output {
			if strings.Contains(strings.ToLower(line), "error") ||
			   strings.Contains(strings.ToLower(line), "failed") {
				logger.Yellow("Warning during repair execution: %s", line)
			}
		}
	}

	logger.Blue("Completed repair action: %s", repair.Name)
	return nil
}

// checkArrayHealth checks Unraid array health
func (d *DiagnosticsManager) checkArrayHealth() []HealthCheck {
	checks := make([]HealthCheck, 0)

	// Check array state
	check := HealthCheck{
		Name:      "Array State",
		Category:  "array",
		Timestamp: time.Now(),
	}

	if exists, _ := lib.Exists("/proc/mdstat"); exists {
		content, err := os.ReadFile("/proc/mdstat")
		if err != nil {
			check.Status = "failed"
			check.Message = "Cannot read array status"
			check.Critical = true
		} else {
			mdstat := string(content)
			if strings.Contains(mdstat, "active") {
				check.Status = "passed"
				check.Message = "Array is active and healthy"
			} else {
				check.Status = "warning"
				check.Message = "Array is not active"
				check.Remediation = "Start the array from Unraid web interface"
			}
		}
	} else {
		check.Status = "warning"
		check.Message = "Array status unavailable"
	}

	checks = append(checks, check)

	// Check for array errors
	errorCheck := HealthCheck{
		Name:      "Array Errors",
		Category:  "array",
		Timestamp: time.Now(),
	}

	if exists, _ := lib.Exists("/var/log/syslog"); exists {
		output := lib.GetCmdOutput("grep", "-i", "error", "/var/log/syslog")
		errorCount := 0
		for _, line := range output {
			if strings.Contains(line, "md") || strings.Contains(line, "disk") {
				errorCount++
			}
		}

		if errorCount == 0 {
			errorCheck.Status = "passed"
			errorCheck.Message = "No array errors detected"
		} else {
			errorCheck.Status = "critical"
			errorCheck.Message = fmt.Sprintf("Found %d array-related errors", errorCount)
			errorCheck.Critical = true
			errorCheck.Remediation = "Check system logs and disk health"
		}
	} else {
		errorCheck.Status = "warning"
		errorCheck.Message = "Cannot check for array errors"
	}

	checks = append(checks, errorCheck)

	return checks
}

// checkDiskHealth checks disk health using SMART
func (d *DiagnosticsManager) checkDiskHealth() []HealthCheck {
	checks := make([]HealthCheck, 0)

	// Find all disks
	output := lib.GetCmdOutput("lsblk", "-d", "-n", "-o", "NAME")
	
	for _, diskName := range output {
		diskName = strings.TrimSpace(diskName)
		if diskName == "" || strings.HasPrefix(diskName, "loop") {
			continue
		}

		check := HealthCheck{
			Name:      fmt.Sprintf("Disk Health (%s)", diskName),
			Category:  "storage",
			Timestamp: time.Now(),
		}

		device := "/dev/" + diskName
		smartOutput := lib.GetCmdOutput("smartctl", "-H", device)
		
		healthy := false
		for _, line := range smartOutput {
			if strings.Contains(line, "SMART overall-health") {
				if strings.Contains(line, "PASSED") {
					healthy = true
					break
				}
			}
		}

		if healthy {
			check.Status = "passed"
			check.Message = "Disk health is good"
		} else {
			check.Status = "critical"
			check.Message = "Disk health check failed"
			check.Critical = true
			check.Remediation = "Check disk SMART attributes and consider replacement"
		}

		checks = append(checks, check)
	}

	return checks
}

// checkMemoryHealth checks memory usage and health
func (d *DiagnosticsManager) checkMemoryHealth() []HealthCheck {
	checks := make([]HealthCheck, 0)

	check := HealthCheck{
		Name:      "Memory Usage",
		Category:  "system",
		Timestamp: time.Now(),
	}

	// Read memory info
	content, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		check.Status = "failed"
		check.Message = "Cannot read memory information"
		checks = append(checks, check)
		return checks
	}

	var memTotal, memAvailable uint64
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
					memTotal = val * 1024 // Convert KB to bytes
				}
			}
		} else if strings.HasPrefix(line, "MemAvailable:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
					memAvailable = val * 1024 // Convert KB to bytes
				}
			}
		}
	}

	if memTotal > 0 {
		usedPercent := float64(memTotal-memAvailable) / float64(memTotal) * 100
		
		if usedPercent < 80 {
			check.Status = "passed"
			check.Message = fmt.Sprintf("Memory usage is normal (%.1f%%)", usedPercent)
		} else if usedPercent < 90 {
			check.Status = "warning"
			check.Message = fmt.Sprintf("Memory usage is high (%.1f%%)", usedPercent)
			check.Remediation = "Consider stopping unnecessary services or adding more RAM"
		} else {
			check.Status = "critical"
			check.Message = fmt.Sprintf("Memory usage is critical (%.1f%%)", usedPercent)
			check.Critical = true
			check.Remediation = "Immediately stop unnecessary services or restart system"
		}
	} else {
		check.Status = "failed"
		check.Message = "Cannot determine memory usage"
	}

	checks = append(checks, check)
	return checks
}

// checkCPUHealth checks CPU usage and temperature
func (d *DiagnosticsManager) checkCPUHealth() []HealthCheck {
	checks := make([]HealthCheck, 0)

	// CPU temperature check
	tempCheck := HealthCheck{
		Name:      "CPU Temperature",
		Category:  "system",
		Timestamp: time.Now(),
	}

	// Try to get CPU temperature
	temp := d.getCPUTemperature()
	if temp > 0 {
		if temp < 70 {
			tempCheck.Status = "passed"
			tempCheck.Message = fmt.Sprintf("CPU temperature is normal (%d°C)", temp)
		} else if temp < 85 {
			tempCheck.Status = "warning"
			tempCheck.Message = fmt.Sprintf("CPU temperature is elevated (%d°C)", temp)
			tempCheck.Remediation = "Check system cooling and clean dust from fans"
		} else {
			tempCheck.Status = "critical"
			tempCheck.Message = fmt.Sprintf("CPU temperature is critical (%d°C)", temp)
			tempCheck.Critical = true
			tempCheck.Remediation = "Immediately check cooling system and reduce system load"
		}
	} else {
		tempCheck.Status = "warning"
		tempCheck.Message = "CPU temperature unavailable"
	}

	checks = append(checks, tempCheck)
	return checks
}

// checkNetworkHealth checks network connectivity and interfaces
func (d *DiagnosticsManager) checkNetworkHealth() []HealthCheck {
	checks := make([]HealthCheck, 0)

	// Network connectivity check
	connectCheck := HealthCheck{
		Name:      "Network Connectivity",
		Category:  "network",
		Timestamp: time.Now(),
	}

	// Test connectivity to common DNS servers
	output := lib.GetCmdOutput("ping", "-c", "1", "-W", "5", "8.8.8.8")
	if len(output) > 0 && strings.Contains(strings.Join(output, ""), "1 received") {
		connectCheck.Status = "passed"
		connectCheck.Message = "Network connectivity is working"
	} else {
		connectCheck.Status = "critical"
		connectCheck.Message = "Network connectivity failed"
		connectCheck.Critical = true
		connectCheck.Remediation = "Check network cables and router configuration"
	}

	checks = append(checks, connectCheck)

	// Interface status check
	ifaceCheck := HealthCheck{
		Name:      "Network Interfaces",
		Category:  "network",
		Timestamp: time.Now(),
	}

	output = lib.GetCmdOutput("ip", "link", "show")
	activeInterfaces := 0
	for _, line := range output {
		if strings.Contains(line, "state UP") {
			activeInterfaces++
		}
	}

	if activeInterfaces > 0 {
		ifaceCheck.Status = "passed"
		ifaceCheck.Message = fmt.Sprintf("%d network interfaces are active", activeInterfaces)
	} else {
		ifaceCheck.Status = "critical"
		ifaceCheck.Message = "No active network interfaces found"
		ifaceCheck.Critical = true
	}

	checks = append(checks, ifaceCheck)
	return checks
}

// checkDockerHealth checks Docker service health
func (d *DiagnosticsManager) checkDockerHealth() []HealthCheck {
	checks := make([]HealthCheck, 0)

	check := HealthCheck{
		Name:      "Docker Service",
		Category:  "docker",
		Timestamp: time.Now(),
	}

	// Check if Docker is running
	output := lib.GetCmdOutput("systemctl", "is-active", "docker")
	if len(output) > 0 && strings.TrimSpace(output[0]) == "active" {
		check.Status = "passed"
		check.Message = "Docker service is running"
	} else {
		check.Status = "warning"
		check.Message = "Docker service is not running"
		check.Remediation = "Start Docker service: systemctl start docker"
	}

	checks = append(checks, check)
	return checks
}

// checkSystemServices checks critical system services
func (d *DiagnosticsManager) checkSystemServices() []HealthCheck {
	checks := make([]HealthCheck, 0)

	services := []string{"sshd", "nginx", "smbd", "nfsd"}

	for _, service := range services {
		check := HealthCheck{
			Name:      fmt.Sprintf("Service: %s", service),
			Category:  "services",
			Timestamp: time.Now(),
		}

		output := lib.GetCmdOutput("systemctl", "is-active", service)
		if len(output) > 0 && strings.TrimSpace(output[0]) == "active" {
			check.Status = "passed"
			check.Message = fmt.Sprintf("%s service is running", service)
		} else {
			check.Status = "warning"
			check.Message = fmt.Sprintf("%s service is not running", service)
			check.Remediation = fmt.Sprintf("Start service: systemctl start %s", service)
		}

		checks = append(checks, check)
	}

	return checks
}

// checkFileSystemHealth checks file system health and disk space
func (d *DiagnosticsManager) checkFileSystemHealth() []HealthCheck {
	checks := make([]HealthCheck, 0)

	// Check root filesystem space
	rootCheck := HealthCheck{
		Name:      "Root Filesystem Space",
		Category:  "storage",
		Timestamp: time.Now(),
	}

	output := lib.GetCmdOutput("df", "-h", "/")
	if len(output) >= 2 {
		fields := strings.Fields(output[1])
		if len(fields) >= 5 {
			usageStr := strings.TrimSuffix(fields[4], "%")
			if usage, err := strconv.Atoi(usageStr); err == nil {
				if usage < 80 {
					rootCheck.Status = "passed"
					rootCheck.Message = fmt.Sprintf("Root filesystem usage is normal (%d%%)", usage)
				} else if usage < 90 {
					rootCheck.Status = "warning"
					rootCheck.Message = fmt.Sprintf("Root filesystem usage is high (%d%%)", usage)
					rootCheck.Remediation = "Clean up unnecessary files or expand storage"
				} else {
					rootCheck.Status = "critical"
					rootCheck.Message = fmt.Sprintf("Root filesystem usage is critical (%d%%)", usage)
					rootCheck.Critical = true
					rootCheck.Remediation = "Immediately free up disk space"
				}
			}
		}
	}

	if rootCheck.Status == "" {
		rootCheck.Status = "failed"
		rootCheck.Message = "Cannot determine root filesystem usage"
	}

	checks = append(checks, rootCheck)
	return checks
}

// checkLogHealth checks system log health
func (d *DiagnosticsManager) checkLogHealth() []HealthCheck {
	checks := make([]HealthCheck, 0)

	check := HealthCheck{
		Name:      "System Logs",
		Category:  "system",
		Timestamp: time.Now(),
	}

	// Check for recent critical errors
	output := lib.GetCmdOutput("journalctl", "--since", "1 hour ago", "--priority", "err", "--no-pager")
	errorCount := len(output)

	if errorCount == 0 {
		check.Status = "passed"
		check.Message = "No recent critical errors in logs"
	} else if errorCount < 10 {
		check.Status = "warning"
		check.Message = fmt.Sprintf("Found %d recent errors in logs", errorCount)
		check.Remediation = "Review system logs for details"
	} else {
		check.Status = "critical"
		check.Message = fmt.Sprintf("Found %d recent errors in logs", errorCount)
		check.Critical = true
		check.Remediation = "Investigate system logs immediately"
	}

	checks = append(checks, check)
	return checks
}

// checkSecurityHealth checks basic security settings
func (d *DiagnosticsManager) checkSecurityHealth() []HealthCheck {
	checks := make([]HealthCheck, 0)

	// Check SSH root login
	sshCheck := HealthCheck{
		Name:      "SSH Security",
		Category:  "security",
		Timestamp: time.Now(),
	}

	if exists, _ := lib.Exists("/etc/ssh/sshd_config"); exists {
		content, err := os.ReadFile("/etc/ssh/sshd_config")
		if err == nil {
			config := string(content)
			if strings.Contains(config, "PermitRootLogin no") {
				sshCheck.Status = "passed"
				sshCheck.Message = "SSH root login is disabled"
			} else {
				sshCheck.Status = "warning"
				sshCheck.Message = "SSH root login may be enabled"
				sshCheck.Remediation = "Consider disabling SSH root login for security"
			}
		}
	}

	if sshCheck.Status == "" {
		sshCheck.Status = "warning"
		sshCheck.Message = "Cannot check SSH configuration"
	}

	checks = append(checks, sshCheck)
	return checks
}

// Helper methods for diagnostics

// calculateHealthSummary calculates the health summary
func (d *DiagnosticsManager) calculateHealthSummary(health *SystemHealth) {
	health.Summary.Total = len(health.Checks)

	for _, check := range health.Checks {
		switch check.Status {
		case "passed":
			health.Summary.Passed++
		case "warning":
			health.Summary.Warning++
		case "critical":
			health.Summary.Critical++
		case "failed":
			health.Summary.Failed++
		}
	}
}

// determineOverallStatus determines the overall system health status
func (d *DiagnosticsManager) determineOverallStatus(health *SystemHealth) {
	if health.Summary.Critical > 0 || health.Summary.Failed > 0 {
		health.OverallStatus = "critical"
	} else if health.Summary.Warning > 0 {
		health.OverallStatus = "warning"
	} else {
		health.OverallStatus = "healthy"
	}
}

// getCPUTemperature gets CPU temperature
func (d *DiagnosticsManager) getCPUTemperature() int {
	// Try sensors command first
	output := lib.GetCmdOutput("sensors")
	for _, line := range output {
		if strings.Contains(line, "Core 0") || strings.Contains(line, "CPU Temperature") {
			if temp := d.parseTemperature(line); temp > 0 {
				return temp
			}
		}
	}

	// Try thermal zone files
	thermalFiles := []string{
		"/sys/class/thermal/thermal_zone0/temp",
		"/sys/class/thermal/thermal_zone1/temp",
	}

	for _, file := range thermalFiles {
		if content, err := os.ReadFile(file); err == nil {
			if temp, err := strconv.Atoi(strings.TrimSpace(string(content))); err == nil {
				return temp / 1000 // Convert from millidegrees
			}
		}
	}

	return 0
}

// parseTemperature parses temperature from sensor output
func (d *DiagnosticsManager) parseTemperature(line string) int {
	fields := strings.Fields(line)
	for _, field := range fields {
		if strings.Contains(field, "°C") {
			tempStr := strings.TrimSuffix(field, "°C")
			tempStr = strings.TrimPrefix(tempStr, "+")
			if temp, err := strconv.ParseFloat(tempStr, 64); err == nil {
				return int(temp)
			}
		}
	}
	return 0
}

// getSystemDiagnostics returns system diagnostic information
func (d *DiagnosticsManager) getSystemDiagnostics() []DiagnosticInfo {
	diagnostics := make([]DiagnosticInfo, 0)

	// Uptime
	if content, err := os.ReadFile("/proc/uptime"); err == nil {
		fields := strings.Fields(string(content))
		if len(fields) > 0 {
			if uptime, err := strconv.ParseFloat(fields[0], 64); err == nil {
				diagnostics = append(diagnostics, DiagnosticInfo{
					Category:    "system",
					Name:        "uptime",
					Value:       uptime,
					Unit:        "seconds",
					Status:      "normal",
					Description: "System uptime in seconds",
				})
			}
		}
	}

	// Load average
	if content, err := os.ReadFile("/proc/loadavg"); err == nil {
		fields := strings.Fields(string(content))
		if len(fields) >= 3 {
			if load1, err := strconv.ParseFloat(fields[0], 64); err == nil {
				status := "normal"
				if load1 > 4.0 {
					status = "critical"
				} else if load1 > 2.0 {
					status = "warning"
				}

				diagnostics = append(diagnostics, DiagnosticInfo{
					Category: "system",
					Name:     "load_average_1min",
					Value:    load1,
					Status:   status,
					Threshold: map[string]interface{}{
						"warning":  2.0,
						"critical": 4.0,
					},
					Description: "1-minute load average",
				})
			}
		}
	}

	return diagnostics
}

// getStorageDiagnostics returns storage diagnostic information
func (d *DiagnosticsManager) getStorageDiagnostics() []DiagnosticInfo {
	diagnostics := make([]DiagnosticInfo, 0)

	// Disk usage for important mount points
	mountPoints := []string{"/", "/boot", "/var/log"}

	for _, mountPoint := range mountPoints {
		output := lib.GetCmdOutput("df", "-h", mountPoint)
		if len(output) >= 2 {
			fields := strings.Fields(output[1])
			if len(fields) >= 5 {
				usageStr := strings.TrimSuffix(fields[4], "%")
				if usage, err := strconv.Atoi(usageStr); err == nil {
					status := "normal"
					if usage > 90 {
						status = "critical"
					} else if usage > 80 {
						status = "warning"
					}

					diagnostics = append(diagnostics, DiagnosticInfo{
						Category: "storage",
						Name:     fmt.Sprintf("disk_usage_%s", strings.ReplaceAll(mountPoint, "/", "_")),
						Value:    usage,
						Unit:     "%",
						Status:   status,
						Threshold: map[string]interface{}{
							"warning":  80,
							"critical": 90,
						},
						Description: fmt.Sprintf("Disk usage for %s", mountPoint),
					})
				}
			}
		}
	}

	return diagnostics
}

// getNetworkDiagnostics returns network diagnostic information
func (d *DiagnosticsManager) getNetworkDiagnostics() []DiagnosticInfo {
	diagnostics := make([]DiagnosticInfo, 0)

	// Network interface statistics
	if content, err := os.ReadFile("/proc/net/dev"); err == nil {
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if i < 2 { // Skip header lines
				continue
			}

			fields := strings.Fields(line)
			if len(fields) >= 17 {
				interfaceName := strings.TrimSuffix(fields[0], ":")
				if interfaceName == "lo" { // Skip loopback
					continue
				}

				if rxBytes, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
					diagnostics = append(diagnostics, DiagnosticInfo{
						Category:    "network",
						Name:        fmt.Sprintf("interface_%s_rx_bytes", interfaceName),
						Value:       rxBytes,
						Unit:        "bytes",
						Status:      "normal",
						Description: fmt.Sprintf("Bytes received on %s", interfaceName),
					})
				}

				if txBytes, err := strconv.ParseUint(fields[9], 10, 64); err == nil {
					diagnostics = append(diagnostics, DiagnosticInfo{
						Category:    "network",
						Name:        fmt.Sprintf("interface_%s_tx_bytes", interfaceName),
						Value:       txBytes,
						Unit:        "bytes",
						Status:      "normal",
						Description: fmt.Sprintf("Bytes transmitted on %s", interfaceName),
					})
				}
			}
		}
	}

	return diagnostics
}

// getPerformanceDiagnostics returns performance diagnostic information
func (d *DiagnosticsManager) getPerformanceDiagnostics() []DiagnosticInfo {
	diagnostics := make([]DiagnosticInfo, 0)

	// CPU temperature
	if temp := d.getCPUTemperature(); temp > 0 {
		status := "normal"
		if temp > 85 {
			status = "critical"
		} else if temp > 75 {
			status = "warning"
		}

		diagnostics = append(diagnostics, DiagnosticInfo{
			Category: "performance",
			Name:     "cpu_temperature",
			Value:    temp,
			Unit:     "°C",
			Status:   status,
			Threshold: map[string]interface{}{
				"warning":  75,
				"critical": 85,
			},
			Description: "CPU temperature",
		})
	}

	return diagnostics
}
