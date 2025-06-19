package adapters

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/domalab/uma/daemon/logger"
)

// SystemMonitor provides real system data collection
type SystemMonitor struct{}

// NewSystemMonitor creates a new system monitor
func NewSystemMonitor() *SystemMonitor {
	return &SystemMonitor{}
}

// GetRealCPUInfo retrieves actual CPU information from the system
func (s *SystemMonitor) GetRealCPUInfo() (interface{}, error) {
	cpuInfo := map[string]interface{}{
		"usage":        0.0,
		"temperature":  0.0,
		"cores":        0,
		"threads":      0,
		"model":        "Unknown",
		"architecture": "Unknown",
		"frequency":    0.0,
		"load1":        0.0,
		"load5":        0.0,
		"load15":       0.0,
	}

	// Get CPU model and core information from /proc/cpuinfo
	if err := s.parseCPUInfo(cpuInfo); err != nil {
		logger.Yellow("Failed to parse CPU info: %v", err)
	}

	// Get load averages from /proc/loadavg
	if err := s.parseLoadAverage(cpuInfo); err != nil {
		logger.Yellow("Failed to parse load average: %v", err)
	}

	// Get CPU usage from /proc/stat
	if usage, err := s.calculateCPUUsage(); err == nil {
		cpuInfo["usage"] = usage
	}

	// Get CPU temperature
	if temp := s.getCPUTemperature(); temp > 0 {
		cpuInfo["temperature"] = temp
	}

	return cpuInfo, nil
}

// parseCPUInfo parses /proc/cpuinfo for CPU details
func (s *SystemMonitor) parseCPUInfo(cpuInfo map[string]interface{}) error {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return err
	}
	defer file.Close()

	var cores, threads int
	var model, architecture string
	var frequency float64

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "model name":
				if model == "" {
					model = value
				}
			case "cpu cores":
				if c, err := strconv.Atoi(value); err == nil && cores == 0 {
					cores = c
				}
			case "siblings":
				if t, err := strconv.Atoi(value); err == nil && threads == 0 {
					threads = t
				}
			case "cpu MHz":
				if f, err := strconv.ParseFloat(value, 64); err == nil && frequency == 0 {
					frequency = f
				}
			case "flags":
				if strings.Contains(value, "lm") {
					architecture = "x86_64"
				} else {
					architecture = "x86"
				}
			}
		}
	}

	if model != "" {
		cpuInfo["model"] = model
	}
	if cores > 0 {
		cpuInfo["cores"] = cores
	}
	if threads > 0 {
		cpuInfo["threads"] = threads
	}
	if frequency > 0 {
		cpuInfo["frequency"] = frequency
	}
	if architecture != "" {
		cpuInfo["architecture"] = architecture
	}

	return scanner.Err()
}

// parseLoadAverage parses /proc/loadavg for load averages
func (s *SystemMonitor) parseLoadAverage(cpuInfo map[string]interface{}) error {
	content, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return err
	}

	fields := strings.Fields(string(content))
	if len(fields) >= 3 {
		if load1, err := strconv.ParseFloat(fields[0], 64); err == nil {
			cpuInfo["load1"] = load1
		}
		if load5, err := strconv.ParseFloat(fields[1], 64); err == nil {
			cpuInfo["load5"] = load5
		}
		if load15, err := strconv.ParseFloat(fields[2], 64); err == nil {
			cpuInfo["load15"] = load15
		}
	}

	return nil
}

// calculateCPUUsage calculates current CPU usage from /proc/stat
func (s *SystemMonitor) calculateCPUUsage() (float64, error) {
	// Read /proc/stat twice with a small interval to calculate usage
	stat1, err := s.readCPUStat()
	if err != nil {
		return 0, err
	}

	time.Sleep(100 * time.Millisecond)

	stat2, err := s.readCPUStat()
	if err != nil {
		return 0, err
	}

	// Calculate usage percentage
	totalDiff := stat2.total - stat1.total
	idleDiff := stat2.idle - stat1.idle

	if totalDiff == 0 {
		return 0, nil
	}

	usage := 100.0 * (1.0 - float64(idleDiff)/float64(totalDiff))
	return usage, nil
}

// cpuStat represents CPU statistics from /proc/stat
type cpuStat struct {
	total uint64
	idle  uint64
}

// readCPUStat reads CPU statistics from /proc/stat
func (s *SystemMonitor) readCPUStat() (*cpuStat, error) {
	content, err := os.ReadFile("/proc/stat")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) >= 5 {
				var total, idle uint64
				for i := 1; i < len(fields) && i <= 10; i++ {
					val, err := strconv.ParseUint(fields[i], 10, 64)
					if err != nil {
						continue
					}
					total += val
					if i == 4 { // idle time is the 4th field
						idle = val
					}
				}
				return &cpuStat{total: total, idle: idle}, nil
			}
		}
	}

	return nil, fmt.Errorf("cpu line not found in /proc/stat")
}

// getCPUTemperature gets CPU temperature from thermal zones
func (s *SystemMonitor) getCPUTemperature() float64 {
	// Try different thermal zone files
	thermalPaths := []string{
		"/sys/class/thermal/thermal_zone0/temp",
		"/sys/class/thermal/thermal_zone1/temp",
		"/sys/class/hwmon/hwmon0/temp1_input",
		"/sys/class/hwmon/hwmon1/temp1_input",
		"/sys/class/hwmon/hwmon2/temp1_input",
	}

	for _, path := range thermalPaths {
		if content, err := os.ReadFile(path); err == nil {
			if temp, err := strconv.ParseFloat(strings.TrimSpace(string(content)), 64); err == nil {
				// Convert millidegrees to degrees if needed
				if temp > 1000 {
					temp = temp / 1000
				}
				if temp > 0 && temp < 150 { // Reasonable temperature range
					return temp
				}
			}
		}
	}

	return 0
}

// GetRealMemoryInfo retrieves actual memory information from the system
func (s *SystemMonitor) GetRealMemoryInfo() (interface{}, error) {
	memInfo := map[string]interface{}{
		"total":     0,
		"used":      0,
		"free":      0,
		"available": 0,
		"cached":    0,
		"buffers":   0,
		"usage":     0.0,
	}

	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return memInfo, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}

		// Convert from KB to bytes
		value *= 1024

		switch key {
		case "MemTotal":
			memInfo["total"] = value
		case "MemAvailable":
			memInfo["available"] = value
		case "MemFree":
			memInfo["free"] = value
		case "Buffers":
			memInfo["buffers"] = value
		case "Cached":
			memInfo["cached"] = value
		}
	}

	// Calculate used memory and usage percentage
	if total, ok := memInfo["total"].(uint64); ok && total > 0 {
		if available, ok := memInfo["available"].(uint64); ok {
			used := total - available
			memInfo["used"] = used
			memInfo["usage"] = float64(used) / float64(total) * 100.0
		}
	}

	return memInfo, scanner.Err()
}

// GetRealNetworkInfo retrieves actual network interface information
func (s *SystemMonitor) GetRealNetworkInfo() (interface{}, error) {
	interfaces := make([]map[string]interface{}, 0)

	// Parse /proc/net/dev for interface statistics
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		return map[string]interface{}{"interfaces": interfaces}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Skip header lines
	scanner.Scan()
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		// Interface name is the first field, remove colon
		ifaceName := strings.TrimSuffix(fields[0], ":")

		// Skip loopback interface
		if ifaceName == "lo" {
			continue
		}

		// Parse statistics
		rxBytes, _ := strconv.ParseUint(fields[1], 10, 64)
		txBytes, _ := strconv.ParseUint(fields[9], 10, 64)

		// Get interface details using ip command
		ipAddr, status := s.getInterfaceDetails(ifaceName)

		iface := map[string]interface{}{
			"name":       ifaceName,
			"rx_bytes":   rxBytes,
			"tx_bytes":   txBytes,
			"ip_address": ipAddr,
			"status":     status,
		}

		interfaces = append(interfaces, iface)
	}

	return map[string]interface{}{"interfaces": interfaces}, scanner.Err()
}

// getInterfaceDetails gets IP address and status for an interface
func (s *SystemMonitor) getInterfaceDetails(ifaceName string) (string, string) {
	// Use ip command to get interface details
	cmd := exec.Command("ip", "addr", "show", ifaceName)
	output, err := cmd.Output()
	if err != nil {
		return "", "unknown"
	}

	lines := strings.Split(string(output), "\n")
	var ipAddr string
	status := "down"

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check interface status
		if strings.Contains(line, "state UP") {
			status = "up"
		}

		// Extract IP address
		if strings.Contains(line, "inet ") && !strings.Contains(line, "inet6") {
			fields := strings.Fields(line)
			for i, field := range fields {
				if field == "inet" && i+1 < len(fields) {
					ipAddr = strings.Split(fields[i+1], "/")[0]
					break
				}
			}
		}
	}

	return ipAddr, status
}

// GetRealTemperatureData retrieves actual temperature sensor information
func (s *SystemMonitor) GetRealTemperatureData() (interface{}, error) {
	sensors := make([]map[string]interface{}, 0)
	fans := make([]map[string]interface{}, 0)

	// Try to use sensors command first
	if sensorData := s.parseSensorsCommand(); len(sensorData) > 0 {
		sensors = append(sensors, sensorData...)
	} else {
		// Fallback to reading hwmon directly
		if hwmonData := s.parseHwmonSensors(); len(hwmonData) > 0 {
			sensors = append(sensors, hwmonData...)
		}
	}

	// Get fan data
	if fanData := s.parseFanData(); len(fanData) > 0 {
		fans = append(fans, fanData...)
	}

	return map[string]interface{}{
		"sensors": sensors,
		"fans":    fans,
	}, nil
}

// parseSensorsCommand parses output from the sensors command
func (s *SystemMonitor) parseSensorsCommand() []map[string]interface{} {
	sensors := make([]map[string]interface{}, 0)

	cmd := exec.Command("sensors")
	output, err := cmd.Output()
	if err != nil {
		return sensors
	}

	lines := strings.Split(string(output), "\n")
	var currentChip string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Detect chip name
		if strings.Contains(line, "-") && !strings.Contains(line, ":") && !strings.Contains(line, "°C") {
			currentChip = line
			continue
		}

		// Parse temperature lines
		if strings.Contains(line, "°C") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[0])
				valueStr := strings.TrimSpace(parts[1])

				// Extract temperature value
				if tempStr := s.extractTemperature(valueStr); tempStr != "" {
					if temp, err := strconv.ParseFloat(tempStr, 64); err == nil {
						sensor := map[string]interface{}{
							"name":        fmt.Sprintf("%s - %s", currentChip, name),
							"temperature": temp,
							"unit":        "°C",
							"status":      s.getTemperatureStatus(temp),
							"source":      currentChip,
						}
						sensors = append(sensors, sensor)
					}
				}
			}
		}
	}

	return sensors
}

// extractTemperature extracts temperature value from sensors output
func (s *SystemMonitor) extractTemperature(valueStr string) string {
	// Look for pattern like "+45.0°C"
	fields := strings.Fields(valueStr)
	for _, field := range fields {
		if strings.Contains(field, "°C") {
			tempStr := strings.TrimPrefix(field, "+")
			tempStr = strings.TrimSuffix(tempStr, "°C")
			return tempStr
		}
	}
	return ""
}

// parseHwmonSensors parses hwmon sensors directly
func (s *SystemMonitor) parseHwmonSensors() []map[string]interface{} {
	sensors := make([]map[string]interface{}, 0)

	// Check common hwmon paths
	hwmonPaths := []string{
		"/sys/class/hwmon/hwmon0",
		"/sys/class/hwmon/hwmon1",
		"/sys/class/hwmon/hwmon2",
		"/sys/class/hwmon/hwmon3",
	}

	for _, hwmonPath := range hwmonPaths {
		if _, err := os.Stat(hwmonPath); os.IsNotExist(err) {
			continue
		}

		// Get chip name
		chipName := s.getHwmonChipName(hwmonPath)

		// Look for temperature inputs
		for i := 1; i <= 10; i++ {
			tempPath := fmt.Sprintf("%s/temp%d_input", hwmonPath, i)
			labelPath := fmt.Sprintf("%s/temp%d_label", hwmonPath, i)

			if content, err := os.ReadFile(tempPath); err == nil {
				if temp, err := strconv.ParseFloat(strings.TrimSpace(string(content)), 64); err == nil {
					temp = temp / 1000 // Convert millidegrees to degrees

					// Get label if available
					label := fmt.Sprintf("temp%d", i)
					if labelContent, err := os.ReadFile(labelPath); err == nil {
						label = strings.TrimSpace(string(labelContent))
					}

					sensor := map[string]interface{}{
						"name":        fmt.Sprintf("%s - %s", chipName, label),
						"temperature": temp,
						"unit":        "°C",
						"status":      s.getTemperatureStatus(temp),
						"source":      chipName,
					}
					sensors = append(sensors, sensor)
				}
			}
		}
	}

	return sensors
}

// getHwmonChipName gets the chip name for a hwmon device
func (s *SystemMonitor) getHwmonChipName(hwmonPath string) string {
	if content, err := os.ReadFile(hwmonPath + "/name"); err == nil {
		return strings.TrimSpace(string(content))
	}
	return "Unknown"
}

// parseFanData parses fan speed data
func (s *SystemMonitor) parseFanData() []map[string]interface{} {
	fans := make([]map[string]interface{}, 0)

	// Check hwmon paths for fan data
	hwmonPaths := []string{
		"/sys/class/hwmon/hwmon0",
		"/sys/class/hwmon/hwmon1",
		"/sys/class/hwmon/hwmon2",
		"/sys/class/hwmon/hwmon3",
	}

	for _, hwmonPath := range hwmonPaths {
		chipName := s.getHwmonChipName(hwmonPath)

		// Look for fan inputs
		for i := 1; i <= 10; i++ {
			fanPath := fmt.Sprintf("%s/fan%d_input", hwmonPath, i)
			labelPath := fmt.Sprintf("%s/fan%d_label", hwmonPath, i)

			if content, err := os.ReadFile(fanPath); err == nil {
				if rpm, err := strconv.ParseFloat(strings.TrimSpace(string(content)), 64); err == nil {
					// Get label if available
					label := fmt.Sprintf("fan%d", i)
					if labelContent, err := os.ReadFile(labelPath); err == nil {
						label = strings.TrimSpace(string(labelContent))
					}

					fan := map[string]interface{}{
						"name":   fmt.Sprintf("%s - %s", chipName, label),
						"speed":  rpm,
						"unit":   "RPM",
						"status": s.getFanStatus(rpm),
						"source": chipName,
					}
					fans = append(fans, fan)
				}
			}
		}
	}

	return fans
}

// getTemperatureStatus determines temperature status
func (s *SystemMonitor) getTemperatureStatus(temp float64) string {
	if temp > 80 {
		return "critical"
	} else if temp > 70 {
		return "warning"
	} else if temp > 60 {
		return "warm"
	}
	return "normal"
}

// getFanStatus determines fan status based on RPM
func (s *SystemMonitor) getFanStatus(rpm float64) string {
	if rpm == 0 {
		return "stopped"
	} else if rpm < 500 {
		return "low"
	} else if rpm > 3000 {
		return "high"
	}
	return "normal"
}

// GetRealUptimeInfo retrieves actual system uptime
func (s *SystemMonitor) GetRealUptimeInfo() (interface{}, error) {
	content, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return map[string]interface{}{"uptime": "0d 0h 0m 0s"}, err
	}

	fields := strings.Fields(string(content))
	if len(fields) == 0 {
		return map[string]interface{}{"uptime": "0d 0h 0m 0s"}, fmt.Errorf("invalid uptime format")
	}

	uptimeSeconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return map[string]interface{}{"uptime": "0d 0h 0m 0s"}, err
	}

	// Convert to days, hours, minutes, seconds
	days := int(uptimeSeconds) / 86400
	hours := (int(uptimeSeconds) % 86400) / 3600
	minutes := (int(uptimeSeconds) % 3600) / 60
	seconds := int(uptimeSeconds) % 60

	uptimeStr := fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)

	return map[string]interface{}{
		"uptime":         uptimeStr,
		"uptime_seconds": int(uptimeSeconds),
		"days":           days,
		"hours":          hours,
		"minutes":        minutes,
		"seconds":        seconds,
	}, nil
}

// GetRealGPUInfo retrieves actual GPU information
func (s *SystemMonitor) GetRealGPUInfo() (interface{}, error) {
	gpus := make([]interface{}, 0)

	// Use lspci to detect GPUs
	cmd := exec.Command("lspci", "-v")
	output, err := cmd.Output()
	if err != nil {
		return map[string]interface{}{"gpus": gpus}, nil
	}

	lines := strings.Split(string(output), "\n")
	var currentGPU map[string]interface{}

	for _, line := range lines {
		// Look for VGA controller lines
		if strings.Contains(line, "VGA compatible controller") {
			if currentGPU != nil {
				gpus = append(gpus, currentGPU)
			}

			parts := strings.Split(line, ": ")
			if len(parts) >= 2 {
				currentGPU = map[string]interface{}{
					"name":        parts[1],
					"type":        "integrated",
					"vendor":      s.extractGPUVendor(parts[1]),
					"driver":      "unknown",
					"memory":      "unknown",
					"temperature": 0.0,
					"usage":       0.0,
				}

				// Determine GPU type
				if strings.Contains(strings.ToLower(parts[1]), "intel") {
					currentGPU["type"] = "integrated"
					currentGPU["vendor"] = "Intel"
				} else if strings.Contains(strings.ToLower(parts[1]), "nvidia") {
					currentGPU["type"] = "discrete"
					currentGPU["vendor"] = "NVIDIA"
				} else if strings.Contains(strings.ToLower(parts[1]), "amd") || strings.Contains(strings.ToLower(parts[1]), "radeon") {
					currentGPU["type"] = "discrete"
					currentGPU["vendor"] = "AMD"
				}
			}
		}

		// Look for driver information
		if currentGPU != nil && strings.Contains(line, "Kernel driver in use:") {
			parts := strings.Split(line, ": ")
			if len(parts) >= 2 {
				currentGPU["driver"] = strings.TrimSpace(parts[1])
			}
		}
	}

	// Add the last GPU if any
	if currentGPU != nil {
		gpus = append(gpus, currentGPU)
	}

	// Try to get Intel GPU temperature
	for _, gpu := range gpus {
		if gpuMap, ok := gpu.(map[string]interface{}); ok {
			if gpuMap["vendor"] == "Intel" {
				if temp := s.getIntelGPUTemperature(); temp > 0 {
					gpuMap["temperature"] = temp
				}
			}
		}
	}

	return map[string]interface{}{"gpus": gpus}, nil
}

// extractGPUVendor extracts vendor from GPU name
func (s *SystemMonitor) extractGPUVendor(name string) string {
	nameLower := strings.ToLower(name)
	if strings.Contains(nameLower, "intel") {
		return "Intel"
	} else if strings.Contains(nameLower, "nvidia") {
		return "NVIDIA"
	} else if strings.Contains(nameLower, "amd") || strings.Contains(nameLower, "radeon") {
		return "AMD"
	}
	return "Unknown"
}

// getIntelGPUTemperature gets Intel GPU temperature
func (s *SystemMonitor) getIntelGPUTemperature() float64 {
	// Try different paths for Intel GPU temperature
	tempPaths := []string{
		"/sys/class/drm/card0/device/hwmon/hwmon*/temp1_input",
		"/sys/class/hwmon/hwmon*/temp1_input",
	}

	for _, pathPattern := range tempPaths {
		// Use glob to find matching paths
		if matches, err := filepath.Glob(pathPattern); err == nil {
			for _, path := range matches {
				if content, err := os.ReadFile(path); err == nil {
					if temp, err := strconv.ParseFloat(strings.TrimSpace(string(content)), 64); err == nil {
						// Convert millidegrees to degrees
						temp = temp / 1000
						if temp > 0 && temp < 150 {
							return temp
						}
					}
				}
			}
		}
	}

	return 0
}

// GetRealSystemLogs retrieves system log information
func (s *SystemMonitor) GetRealSystemLogs() (interface{}, error) {
	logs := make([]interface{}, 0)

	// Get recent syslog entries
	if syslogEntries := s.getLogEntries("/var/log/syslog", 50); len(syslogEntries) > 0 {
		logs = append(logs, map[string]interface{}{
			"name":    "syslog",
			"path":    "/var/log/syslog",
			"entries": syslogEntries,
		})
	}

	// Get Unraid specific logs
	unraidLogPaths := []string{
		"/var/log/unraid.log",
		"/var/log/docker.log",
		"/var/log/libvirt.log",
	}

	for _, logPath := range unraidLogPaths {
		if _, err := os.Stat(logPath); err == nil {
			if entries := s.getLogEntries(logPath, 20); len(entries) > 0 {
				logName := strings.TrimSuffix(filepath.Base(logPath), ".log")
				logs = append(logs, map[string]interface{}{
					"name":    logName,
					"path":    logPath,
					"entries": entries,
				})
			}
		}
	}

	return map[string]interface{}{"logs": logs}, nil
}

// getLogEntries gets recent log entries from a file
func (s *SystemMonitor) getLogEntries(logPath string, maxEntries int) []interface{} {
	entries := make([]interface{}, 0)

	// Use tail command to get recent entries
	cmd := exec.Command("tail", "-n", fmt.Sprintf("%d", maxEntries), logPath)
	output, err := cmd.Output()
	if err != nil {
		return entries
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			entry := map[string]interface{}{
				"message":   line,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			}

			// Try to parse timestamp from log line
			if timestamp := s.parseLogTimestamp(line); !timestamp.IsZero() {
				entry["timestamp"] = timestamp.Format(time.RFC3339)
			}

			entries = append(entries, entry)
		}
	}

	return entries
}

// parseLogTimestamp attempts to parse timestamp from log line
func (s *SystemMonitor) parseLogTimestamp(line string) time.Time {
	// Common log timestamp formats
	formats := []string{
		"Jan 2 15:04:05",
		"2006-01-02T15:04:05",
		"2006/01/02 15:04:05",
	}

	for _, format := range formats {
		if strings.Contains(line, " ") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				timeStr := strings.Join(parts[:3], " ")
				if t, err := time.Parse(format, timeStr); err == nil {
					// Add current year if not present
					if t.Year() == 0 {
						t = t.AddDate(time.Now().Year(), 0, 0)
					}
					return t
				}
			}
		}
	}

	return time.Time{}
}

// VMMonitor provides VM management and monitoring
type VMMonitor struct{}

// NewVMMonitor creates a new VM monitor
func NewVMMonitor() *VMMonitor {
	return &VMMonitor{}
}

// GetRealVMs retrieves actual VM information using libvirt
func (v *VMMonitor) GetRealVMs() (interface{}, error) {
	vms := make([]interface{}, 0)

	// Use virsh to list all VMs
	cmd := exec.Command("virsh", "list", "--all")
	output, err := cmd.Output()
	if err != nil {
		return vms, nil // libvirt might not be available
	}

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		// Skip header lines
		if i < 2 {
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 3 {
			vmId := fields[0]
			vmName := fields[1]
			vmState := strings.Join(fields[2:], " ")

			vm := map[string]interface{}{
				"id":    vmId,
				"name":  vmName,
				"state": vmState,
				"type":  "kvm",
			}

			// Get detailed VM information
			if vmDetails := v.getVMDetails(vmName); vmDetails != nil {
				for k, v := range vmDetails {
					vm[k] = v
				}
			}

			vms = append(vms, vm)
		}
	}

	return vms, nil
}

// getVMDetails gets detailed information about a VM
func (v *VMMonitor) getVMDetails(vmName string) map[string]interface{} {
	details := make(map[string]interface{})

	// Get VM domain info
	cmd := exec.Command("virsh", "dominfo", vmName)
	output, err := cmd.Output()
	if err != nil {
		return details
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "Max memory":
					details["max_memory"] = value
				case "Used memory":
					details["used_memory"] = value
				case "CPU(s)":
					details["vcpus"] = value
				case "OS Type":
					details["os_type"] = value
				case "State":
					details["detailed_state"] = value
				case "CPU time":
					details["cpu_time"] = value
				}
			}
		}
	}

	// Get VM statistics if running
	if details["detailed_state"] == "running" {
		if stats := v.getVMStats(vmName); stats != nil {
			details["stats"] = stats
		}
	}

	return details
}

// getVMStats gets runtime statistics for a VM
func (v *VMMonitor) getVMStats(vmName string) map[string]interface{} {
	stats := make(map[string]interface{})

	// Get CPU stats
	cmd := exec.Command("virsh", "cpu-stats", vmName)
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "cpu_time") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					stats["cpu_time_ns"] = parts[1]
				}
			}
		}
	}

	// Get memory stats
	cmd = exec.Command("virsh", "dommemstat", vmName)
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "actual") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					stats["memory_actual"] = parts[1]
				}
			}
			if strings.Contains(line, "rss") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					stats["memory_rss"] = parts[1]
				}
			}
		}
	}

	return stats
}

// ControlVM controls VM operations (start/stop/restart)
func (v *VMMonitor) ControlVM(vmName, action string) (interface{}, error) {
	result := map[string]interface{}{
		"vm_name":   vmName,
		"action":    action,
		"success":   false,
		"message":   "",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	var cmd *exec.Cmd
	switch action {
	case "start":
		cmd = exec.Command("virsh", "start", vmName)
	case "stop":
		cmd = exec.Command("virsh", "shutdown", vmName)
	case "force-stop":
		cmd = exec.Command("virsh", "destroy", vmName)
	case "restart":
		// First shutdown, then start
		shutdownCmd := exec.Command("virsh", "shutdown", vmName)
		if err := shutdownCmd.Run(); err != nil {
			result["message"] = fmt.Sprintf("Failed to shutdown VM: %v", err)
			return result, err
		}
		// Wait a moment for shutdown
		time.Sleep(3 * time.Second)
		cmd = exec.Command("virsh", "start", vmName)
	default:
		result["message"] = fmt.Sprintf("Unknown action: %s", action)
		return result, fmt.Errorf("unknown action: %s", action)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		result["message"] = fmt.Sprintf("Command failed: %v - %s", err, string(output))
		return result, err
	}

	result["success"] = true
	result["message"] = fmt.Sprintf("VM %s %s successfully", vmName, action)
	return result, nil
}

// StorageMonitor provides storage health monitoring
type StorageMonitor struct{}

// NewStorageMonitor creates a new storage monitor
func NewStorageMonitor() *StorageMonitor {
	return &StorageMonitor{}
}

// GetRealArrayInfo retrieves actual Unraid array information
func (s *StorageMonitor) GetRealArrayInfo() (interface{}, error) {
	arrayInfo := map[string]interface{}{
		"state":      "unknown",
		"protection": "unknown",
		"disks":      []interface{}{},
		"parity":     []interface{}{},
	}

	// Parse Unraid array status from /proc/mdstat
	if unraidData := s.parseUnraidStatus(); unraidData != nil {
		arrayInfo["state"] = unraidData["state"]
		arrayInfo["protection"] = unraidData["protection"]
		arrayInfo["disks"] = unraidData["disks"]
		arrayInfo["parity"] = unraidData["parity"]
		arrayInfo["sync_action"] = unraidData["sync_action"]
		arrayInfo["sync_progress"] = unraidData["sync_progress"]
	}

	return arrayInfo, nil
}

// parseUnraidStatus parses Unraid status from /var/local/emhttp/var.ini
func (s *StorageMonitor) parseUnraidStatus() map[string]interface{} {
	// Read from Unraid's real-time status file
	content, err := os.ReadFile("/var/local/emhttp/var.ini")
	if err != nil {
		// Fallback to /proc/mdstat if var.ini is not available
		return s.parseUnraidStatusFromMdstat()
	}

	unraidData := map[string]interface{}{
		"state":         "unknown",
		"protection":    "unknown",
		"disks":         []interface{}{},
		"parity":        []interface{}{},
		"sync_action":   "none",
		"sync_progress": 0.0,
	}

	lines := strings.Split(string(content), "\n")

	// Parse var.ini key=value format
	var mdResyncAction, mdResyncPos, mdResyncSize, mdState string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"'") // Remove quotes

		switch key {
		case "mdState":
			mdState = value
			if value == "STARTED" {
				unraidData["state"] = "started"
			} else {
				unraidData["state"] = strings.ToLower(value)
			}
		case "mdResyncAction":
			mdResyncAction = value
		case "mdResyncPos":
			mdResyncPos = value
		case "mdResyncSize":
			mdResyncSize = value
		}
	}

	// Determine sync action and progress
	if mdState == "STARTED" && mdResyncAction != "" && mdResyncAction != "IDLE" {
		unraidData["sync_action"] = mdResyncAction

		// Calculate progress percentage
		if pos, err := strconv.ParseInt(mdResyncPos, 10, 64); err == nil {
			if size, err := strconv.ParseInt(mdResyncSize, 10, 64); err == nil && size > 0 {
				progress := float64(pos) / float64(size) * 100.0
				unraidData["sync_progress"] = progress
			}
		}
	} else {
		unraidData["sync_action"] = "none"
	}

	// Parse disk information from /proc/mdstat for disk details
	s.parseUnraidDisks(unraidData)

	// Set protection based on parity disks
	if len(unraidData["parity"].([]interface{})) > 0 {
		unraidData["protection"] = "parity"
	}

	// Add parity check history
	history := s.getParityCheckHistory()
	unraidData["parity_history"] = history

	return unraidData
}

// parseUnraidStatusFromMdstat fallback method to parse from /proc/mdstat
func (s *StorageMonitor) parseUnraidStatusFromMdstat() map[string]interface{} {
	content, err := os.ReadFile("/proc/mdstat")
	if err != nil {
		return nil
	}

	unraidData := map[string]interface{}{
		"state":         "unknown",
		"protection":    "unknown",
		"disks":         []interface{}{},
		"parity":        []interface{}{},
		"sync_action":   "none",
		"sync_progress": 0.0,
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		// Parse array state
		if strings.Contains(line, "mdState=") {
			if strings.Contains(line, "STARTED") {
				unraidData["state"] = "started"
			} else if strings.Contains(line, "STOPPED") {
				unraidData["state"] = "stopped"
			}
		}

		// Parse sync action
		if strings.Contains(line, "mdResyncAction=") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				action := strings.TrimSpace(parts[1])
				unraidData["sync_action"] = action
			}
		}
	}

	s.parseUnraidDisks(unraidData)
	return unraidData
}

// parseUnraidDisks parses disk information from /proc/mdstat
func (s *StorageMonitor) parseUnraidDisks(unraidData map[string]interface{}) {
	content, err := os.ReadFile("/proc/mdstat")
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		// Parse disk information
		if strings.Contains(line, "rdevName.") && strings.Contains(line, "=sd") {
			diskInfo := s.parseUnraidDisk(lines, line)
			if diskInfo != nil {
				if diskInfo["type"] == "parity" {
					parity := unraidData["parity"].([]interface{})
					unraidData["parity"] = append(parity, diskInfo)
				} else {
					disks := unraidData["disks"].([]interface{})
					unraidData["disks"] = append(disks, diskInfo)
				}
			}
		}
	}
}

// getParityCheckHistory retrieves parity check history from logs
func (s *StorageMonitor) getParityCheckHistory() map[string]interface{} {
	history := map[string]interface{}{
		"last_check":     nil,
		"last_duration":  nil,
		"last_speed":     nil,
		"last_errors":    0,
		"last_action":    "unknown",
		"next_scheduled": nil,
		"checks":         []interface{}{},
	}

	// Try to read parity check log
	logPath := "/boot/config/parity-checks.log"
	content, err := os.ReadFile(logPath)
	if err != nil {
		// Try alternative location
		logPath = "/var/log/parity-checks.log"
		content, err = os.ReadFile(logPath)
		if err != nil {
			// Return empty history but still include the structure
			return history
		}
	}

	lines := strings.Split(string(content), "\n")
	checks := make([]interface{}, 0)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Parse parity check log entries
		// Format: timestamp|duration|speed|errors|bytes|action|total_bytes
		parts := strings.Split(line, "|")
		if len(parts) >= 6 {
			check := map[string]interface{}{
				"timestamp":   parts[0],
				"duration":    parts[1],
				"speed":       parts[2],
				"errors":      parts[3],
				"bytes":       parts[4],
				"action":      parts[5],
				"total_bytes": "",
			}
			if len(parts) >= 7 {
				check["total_bytes"] = parts[6]
			}
			checks = append(checks, check)
		}
	}

	// Set last check information if available
	if len(checks) > 0 {
		lastCheck := checks[len(checks)-1].(map[string]interface{})
		history["last_check"] = lastCheck["timestamp"]
		history["last_duration"] = lastCheck["duration"]
		history["last_speed"] = lastCheck["speed"]
		history["last_errors"] = lastCheck["errors"]
		history["last_action"] = lastCheck["action"]
		history["last_bytes"] = lastCheck["bytes"]
	}

	history["checks"] = checks
	return history
}

// parseUnraidDisk parses individual disk information from Unraid status
func (s *StorageMonitor) parseUnraidDisk(lines []string, diskLine string) map[string]interface{} {
	// Extract disk number from line like "rdevName.1=sdd"
	parts := strings.Split(diskLine, ".")
	if len(parts) < 2 {
		return nil
	}

	diskNumPart := strings.Split(parts[1], "=")
	if len(diskNumPart) < 2 {
		return nil
	}

	diskNum := diskNumPart[0]
	deviceName := diskNumPart[1]

	disk := map[string]interface{}{
		"name":        fmt.Sprintf("disk%s", diskNum),
		"device":      fmt.Sprintf("/dev/%s", deviceName),
		"status":      "unknown",
		"health":      "unknown",
		"temperature": 0.0,
		"size":        "unknown",
		"type":        "data",
	}

	// Parse additional disk information
	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf("rdevStatus.%s=", diskNum)) {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				status := strings.TrimSpace(parts[1])
				if status == "DISK_OK" {
					disk["status"] = "active"
				} else {
					disk["status"] = "inactive"
				}
			}
		}

		if strings.Contains(line, fmt.Sprintf("rdevSize.%s=", diskNum)) {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				if size, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64); err == nil {
					// Convert sectors to GB (assuming 512 byte sectors)
					sizeGB := float64(size) * 512 / (1024 * 1024 * 1024)
					disk["size"] = fmt.Sprintf("%.1fGB", sizeGB)
				}
			}
		}

		if strings.Contains(line, fmt.Sprintf("rdevId.%s=", diskNum)) {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				diskId := strings.TrimSpace(parts[1])
				disk["serial"] = diskId
			}
		}
	}

	// Determine if this is a parity disk (disk 0 is typically parity)
	if diskNum == "0" {
		disk["type"] = "parity"
		disk["name"] = "parity"
	}

	// Get SMART health data
	disk["health"] = s.getDiskHealth(disk["device"].(string))
	disk["temperature"] = s.getDiskTemperature(disk["device"].(string))
	disk["smart_data"] = s.getSMARTData(disk["device"].(string))

	return disk
}

// getArrayDisks gets information about array disks
func (s *StorageMonitor) getArrayDisks() []interface{} {
	disks := make([]interface{}, 0)

	// Use lsblk to get disk information
	cmd := exec.Command("lsblk", "-J", "-o", "NAME,SIZE,TYPE,MOUNTPOINT,MODEL")
	output, err := cmd.Output()
	if err != nil {
		return disks
	}

	// Parse JSON output (simplified - would need proper JSON parsing)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "disk") && !strings.Contains(line, "loop") {
			// Extract disk information (simplified parsing)
			if strings.Contains(line, "sd") {
				disk := map[string]interface{}{
					"name":   "disk1",    // Would extract actual name
					"device": "/dev/sda", // Would extract actual device
					"size":   "1TB",      // Would extract actual size
					"status": "active",
					"health": s.getDiskHealth("/dev/sda"),
				}
				disks = append(disks, disk)
			}
		}
	}

	return disks
}

// getParityDisks gets information about parity disks
func (s *StorageMonitor) getParityDisks() []interface{} {
	parity := make([]interface{}, 0)

	// Check for parity disks in /proc/mdstat
	if content, err := os.ReadFile("/proc/mdstat"); err == nil {
		if strings.Contains(string(content), "parity") {
			parityDisk := map[string]interface{}{
				"name":   "parity1",
				"device": "/dev/sdb", // Would extract actual device
				"size":   "1TB",      // Would extract actual size
				"status": "active",
				"health": s.getDiskHealth("/dev/sdb"),
			}
			parity = append(parity, parityDisk)
		}
	}

	return parity
}

// getDiskHealth gets SMART health status for a disk
func (s *StorageMonitor) getDiskHealth(device string) string {
	// Try to get SMART status using smartctl
	cmd := exec.Command("smartctl", "-H", device)
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "PASSED") {
		return "healthy"
	} else if strings.Contains(outputStr, "FAILED") {
		return "failed"
	}

	return "unknown"
}

// GetRealDisks retrieves actual disk information with SMART data
func (s *StorageMonitor) GetRealDisks() (interface{}, error) {
	disks := make([]interface{}, 0)

	// Get all block devices
	cmd := exec.Command("lsblk", "-d", "-n", "-o", "NAME,SIZE,TYPE")
	output, err := cmd.Output()
	if err != nil {
		return disks, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[2] == "disk" {
			deviceName := fields[0]
			devicePath := "/dev/" + deviceName

			// Skip loop devices and other virtual devices
			if strings.HasPrefix(deviceName, "loop") ||
				strings.HasPrefix(deviceName, "ram") ||
				strings.HasPrefix(deviceName, "dm-") {
				continue
			}

			disk := map[string]interface{}{
				"name":        deviceName,
				"device":      devicePath,
				"size":        fields[1],
				"type":        "disk",
				"health":      s.getDiskHealth(devicePath),
				"temperature": s.getDiskTemperature(devicePath),
				"smart_data":  s.getSMARTData(devicePath),
			}

			disks = append(disks, disk)
		}
	}

	return disks, nil
}

// getDiskTemperature gets disk temperature from SMART data
func (s *StorageMonitor) getDiskTemperature(device string) float64 {
	cmd := exec.Command("smartctl", "-A", device)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Temperature") || strings.Contains(line, "194") {
			fields := strings.Fields(line)
			if len(fields) >= 10 {
				if temp, err := strconv.ParseFloat(fields[9], 64); err == nil {
					return temp
				}
			}
		}
	}

	return 0
}

// getSMARTData gets basic SMART data for a disk
func (s *StorageMonitor) getSMARTData(device string) map[string]interface{} {
	smartData := map[string]interface{}{
		"available":  false,
		"status":     "unknown",
		"attributes": map[string]interface{}{},
	}

	cmd := exec.Command("smartctl", "-a", device)
	output, err := cmd.Output()
	if err != nil {
		return smartData
	}

	outputStr := string(output)
	smartData["available"] = true

	// Parse SMART status
	if strings.Contains(outputStr, "SMART overall-health self-assessment test result: PASSED") {
		smartData["status"] = "passed"
	} else if strings.Contains(outputStr, "FAILED") {
		smartData["status"] = "failed"
	}

	// Parse key attributes (simplified)
	attributes := map[string]interface{}{}
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Power_On_Hours") {
			fields := strings.Fields(line)
			if len(fields) >= 10 {
				attributes["power_on_hours"] = fields[9]
			}
		}
		if strings.Contains(line, "Power_Cycle_Count") {
			fields := strings.Fields(line)
			if len(fields) >= 10 {
				attributes["power_cycle_count"] = fields[9]
			}
		}
	}

	smartData["attributes"] = attributes
	return smartData
}

// GetRealZFSPools retrieves actual ZFS pool information
func (s *StorageMonitor) GetRealZFSPools() (interface{}, error) {
	pools := make([]interface{}, 0)

	// Execute zpool list command
	cmd := exec.Command("zpool", "list", "-H", "-o", "name,size,alloc,free,cap,health")
	output, err := cmd.Output()
	if err != nil {
		return pools, nil // ZFS might not be available
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 6 {
			pool := map[string]interface{}{
				"name":   fields[0],
				"size":   fields[1],
				"alloc":  fields[2],
				"free":   fields[3],
				"cap":    fields[4],
				"health": fields[5],
			}

			// Get detailed pool information
			if poolDetails := s.getZFSPoolDetails(fields[0]); poolDetails != nil {
				for k, v := range poolDetails {
					pool[k] = v
				}
			}

			pools = append(pools, pool)
		}
	}

	return pools, nil
}

// getZFSPoolDetails gets detailed information about a ZFS pool
func (s *StorageMonitor) getZFSPoolDetails(poolName string) map[string]interface{} {
	details := make(map[string]interface{})

	// Get pool status
	cmd := exec.Command("zpool", "status", poolName)
	output, err := cmd.Output()
	if err != nil {
		return details
	}

	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	devices := make([]interface{}, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for device lines (typically start with /dev/ or have disk names)
		if strings.HasPrefix(line, "/dev/") ||
			(strings.Contains(line, "sd") && !strings.Contains(line, "pool:")) {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				device := map[string]interface{}{
					"name":  fields[0],
					"state": fields[1],
				}
				if len(fields) >= 3 {
					device["read_errors"] = fields[2]
				}
				if len(fields) >= 4 {
					device["write_errors"] = fields[3]
				}
				if len(fields) >= 5 {
					device["checksum_errors"] = fields[4]
				}
				devices = append(devices, device)
			}
		}
	}

	details["devices"] = devices
	return details
}

// GetRealCacheInfo retrieves actual cache pool information
func (s *StorageMonitor) GetRealCacheInfo() (interface{}, error) {
	cacheInfo := map[string]interface{}{
		"pools": []interface{}{},
	}

	pools := make([]interface{}, 0)

	// Check for cache mount point
	if _, err := os.Stat("/mnt/cache"); err == nil {
		// Get cache filesystem information
		cmd := exec.Command("df", "-h", "/mnt/cache")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			if len(lines) >= 2 {
				fields := strings.Fields(lines[1])
				if len(fields) >= 6 {
					cachePool := map[string]interface{}{
						"name":       "cache",
						"device":     fields[0],
						"size":       fields[1],
						"used":       fields[2],
						"available":  fields[3],
						"usage":      fields[4],
						"mountpoint": fields[5],
						"type":       "cache",
					}

					// Get device health if it's a real device
					if strings.HasPrefix(fields[0], "/dev/") {
						cachePool["health"] = s.getDiskHealth(fields[0])
						cachePool["temperature"] = s.getDiskTemperature(fields[0])
						cachePool["smart_data"] = s.getSMARTData(fields[0])
					}

					pools = append(pools, cachePool)
				}
			}
		}
	}

	// Also check for ZFS cache pools
	if zfsPools, err := s.GetRealZFSPools(); err == nil {
		if poolList, ok := zfsPools.([]interface{}); ok {
			for _, pool := range poolList {
				if poolMap, ok := pool.(map[string]interface{}); ok {
					poolMap["type"] = "zfs_cache"
					pools = append(pools, poolMap)
				}
			}
		}
	}

	cacheInfo["pools"] = pools
	return cacheInfo, nil
}
