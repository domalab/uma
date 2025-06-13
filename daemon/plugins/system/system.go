package system

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/domalab/omniraid/daemon/dto"
	"github.com/domalab/omniraid/daemon/lib"
	"github.com/domalab/omniraid/daemon/logger"
)

// SystemMonitor provides system resource monitoring
type SystemMonitor struct {
	lastCPUStats CPUStats
	lastTime     time.Time
}

// CPUStats represents CPU statistics
type CPUStats struct {
	User   uint64 `json:"user"`
	Nice   uint64 `json:"nice"`
	System uint64 `json:"system"`
	Idle   uint64 `json:"idle"`
	IOWait uint64 `json:"iowait"`
	IRQ    uint64 `json:"irq"`
	SoftIRQ uint64 `json:"softirq"`
	Steal  uint64 `json:"steal"`
	Guest  uint64 `json:"guest"`
	Total  uint64 `json:"total"`
}

// CPUInfo represents CPU information and usage
type CPUInfo struct {
	Model       string  `json:"model"`
	Cores       int     `json:"cores"`
	Threads     int     `json:"threads"`
	Usage       float64 `json:"usage_percent"`
	Temperature int     `json:"temperature,omitempty"`
	Frequency   int     `json:"frequency_mhz,omitempty"`
}

// MemoryInfo represents memory information and usage
type MemoryInfo struct {
	Total       uint64  `json:"total_bytes"`
	Available   uint64  `json:"available_bytes"`
	Used        uint64  `json:"used_bytes"`
	Free        uint64  `json:"free_bytes"`
	Buffers     uint64  `json:"buffers_bytes"`
	Cached      uint64  `json:"cached_bytes"`
	UsedPercent float64 `json:"used_percent"`
}

// LoadInfo represents system load information
type LoadInfo struct {
	Load1  float64 `json:"load_1min"`
	Load5  float64 `json:"load_5min"`
	Load15 float64 `json:"load_15min"`
}

// UptimeInfo represents system uptime information
type UptimeInfo struct {
	Uptime   float64 `json:"uptime_seconds"`
	IdleTime float64 `json:"idle_seconds"`
}

// NetworkInfo represents network interface information
type NetworkInfo struct {
	Interface   string `json:"interface"`
	BytesRecv   uint64 `json:"bytes_received"`
	BytesSent   uint64 `json:"bytes_sent"`
	PacketsRecv uint64 `json:"packets_received"`
	PacketsSent uint64 `json:"packets_sent"`
	ErrorsRecv  uint64 `json:"errors_received"`
	ErrorsSent  uint64 `json:"errors_sent"`
}

// NewSystemMonitor creates a new system monitor
func NewSystemMonitor() *SystemMonitor {
	return &SystemMonitor{}
}

// GetCPUInfo returns CPU information and current usage
func (s *SystemMonitor) GetCPUInfo() (*CPUInfo, error) {
	cpuInfo := &CPUInfo{}

	// Get CPU model and core count
	if err := s.getCPUDetails(cpuInfo); err != nil {
		logger.Yellow("Failed to get CPU details: %v", err)
	}

	// Get CPU usage
	usage, err := s.getCPUUsage()
	if err != nil {
		logger.Yellow("Failed to get CPU usage: %v", err)
	} else {
		cpuInfo.Usage = usage
	}

	// Get CPU temperature
	if temp := s.getCPUTemperature(); temp > 0 {
		cpuInfo.Temperature = temp
	}

	// Get CPU frequency
	if freq := s.getCPUFrequency(); freq > 0 {
		cpuInfo.Frequency = freq
	}

	return cpuInfo, nil
}

// GetMemoryInfo returns memory information and usage
func (s *SystemMonitor) GetMemoryInfo() (*MemoryInfo, error) {
	memInfo := &MemoryInfo{}

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
			memInfo.Total = value
		case "MemAvailable":
			memInfo.Available = value
		case "MemFree":
			memInfo.Free = value
		case "Buffers":
			memInfo.Buffers = value
		case "Cached":
			memInfo.Cached = value
		}
	}

	// Calculate used memory
	memInfo.Used = memInfo.Total - memInfo.Available

	// Calculate usage percentage
	if memInfo.Total > 0 {
		memInfo.UsedPercent = float64(memInfo.Used) / float64(memInfo.Total) * 100
	}

	return memInfo, nil
}

// GetLoadInfo returns system load information
func (s *SystemMonitor) GetLoadInfo() (*LoadInfo, error) {
	loadInfo := &LoadInfo{}

	content, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return loadInfo, err
	}

	fields := strings.Fields(string(content))
	if len(fields) < 3 {
		return loadInfo, fmt.Errorf("invalid loadavg format")
	}

	if loadInfo.Load1, err = strconv.ParseFloat(fields[0], 64); err != nil {
		return loadInfo, err
	}
	if loadInfo.Load5, err = strconv.ParseFloat(fields[1], 64); err != nil {
		return loadInfo, err
	}
	if loadInfo.Load15, err = strconv.ParseFloat(fields[2], 64); err != nil {
		return loadInfo, err
	}

	return loadInfo, nil
}

// GetUptimeInfo returns system uptime information
func (s *SystemMonitor) GetUptimeInfo() (*UptimeInfo, error) {
	uptimeInfo := &UptimeInfo{}

	content, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return uptimeInfo, err
	}

	fields := strings.Fields(string(content))
	if len(fields) < 2 {
		return uptimeInfo, fmt.Errorf("invalid uptime format")
	}

	if uptimeInfo.Uptime, err = strconv.ParseFloat(fields[0], 64); err != nil {
		return uptimeInfo, err
	}
	if uptimeInfo.IdleTime, err = strconv.ParseFloat(fields[1], 64); err != nil {
		return uptimeInfo, err
	}

	return uptimeInfo, nil
}

// GetNetworkInfo returns network interface information
func (s *SystemMonitor) GetNetworkInfo() ([]NetworkInfo, error) {
	networks := make([]NetworkInfo, 0)

	file, err := os.Open("/proc/net/dev")
	if err != nil {
		return networks, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Skip header lines
	scanner.Scan()
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 17 {
			continue
		}

		interfaceName := strings.TrimSuffix(fields[0], ":")
		
		// Skip loopback interface
		if interfaceName == "lo" {
			continue
		}

		netInfo := NetworkInfo{
			Interface: interfaceName,
		}

		// Parse network statistics
		if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
			netInfo.BytesRecv = val
		}
		if val, err := strconv.ParseUint(fields[2], 10, 64); err == nil {
			netInfo.PacketsRecv = val
		}
		if val, err := strconv.ParseUint(fields[3], 10, 64); err == nil {
			netInfo.ErrorsRecv = val
		}
		if val, err := strconv.ParseUint(fields[9], 10, 64); err == nil {
			netInfo.BytesSent = val
		}
		if val, err := strconv.ParseUint(fields[10], 10, 64); err == nil {
			netInfo.PacketsSent = val
		}
		if val, err := strconv.ParseUint(fields[11], 10, 64); err == nil {
			netInfo.ErrorsSent = val
		}

		networks = append(networks, netInfo)
	}

	return networks, nil
}

// GetSystemSamples returns system information as DTO samples for compatibility
func (s *SystemMonitor) GetSystemSamples() []dto.Sample {
	samples := make([]dto.Sample, 0)

	// CPU information
	if cpuInfo, err := s.GetCPUInfo(); err == nil {
		samples = append(samples, dto.Sample{
			Key:       "CPU_USAGE",
			Value:     fmt.Sprintf("%.1f", cpuInfo.Usage),
			Unit:      "%",
			Condition: s.getCondition(cpuInfo.Usage, 80, 90),
		})

		if cpuInfo.Temperature > 0 {
			samples = append(samples, dto.Sample{
				Key:       "CPU_TEMP",
				Value:     fmt.Sprintf("%d", cpuInfo.Temperature),
				Unit:      "°C",
				Condition: s.getTempCondition(cpuInfo.Temperature, 70, 85),
			})
		}
	}

	// Memory information
	if memInfo, err := s.GetMemoryInfo(); err == nil {
		samples = append(samples, dto.Sample{
			Key:       "MEMORY_USAGE",
			Value:     fmt.Sprintf("%.1f", memInfo.UsedPercent),
			Unit:      "%",
			Condition: s.getCondition(memInfo.UsedPercent, 80, 90),
		})

		samples = append(samples, dto.Sample{
			Key:       "MEMORY_USED",
			Value:     fmt.Sprintf("%.1f", float64(memInfo.Used)/(1024*1024*1024)),
			Unit:      "GB",
			Condition: "neutral",
		})
	}

	// Load information
	if loadInfo, err := s.GetLoadInfo(); err == nil {
		samples = append(samples, dto.Sample{
			Key:       "LOAD_1MIN",
			Value:     fmt.Sprintf("%.2f", loadInfo.Load1),
			Unit:      "",
			Condition: s.getLoadCondition(loadInfo.Load1),
		})
	}

	return samples
}

// getCondition returns condition based on percentage thresholds
func (s *SystemMonitor) getCondition(value, warning, critical float64) string {
	if value >= critical {
		return "critical"
	} else if value >= warning {
		return "warning"
	}
	return "normal"
}

// getTempCondition returns condition based on temperature thresholds
func (s *SystemMonitor) getTempCondition(temp, warning, critical int) string {
	if temp >= critical {
		return "critical"
	} else if temp >= warning {
		return "warning"
	}
	return "normal"
}

// getLoadCondition returns condition based on load average
func (s *SystemMonitor) getLoadCondition(load float64) string {
	if load >= 4.0 {
		return "critical"
	} else if load >= 2.0 {
		return "warning"
	}
	return "normal"
}

// getCPUDetails gets CPU model and core information
func (s *SystemMonitor) getCPUDetails(cpuInfo *CPUInfo) error {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	coreCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "model name") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				cpuInfo.Model = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "processor") {
			coreCount++
		} else if strings.HasPrefix(line, "cpu cores") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				if cores, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
					cpuInfo.Cores = cores
				}
			}
		}
	}

	cpuInfo.Threads = coreCount
	if cpuInfo.Cores == 0 {
		cpuInfo.Cores = coreCount
	}

	return nil
}

// getCPUUsage calculates CPU usage percentage
func (s *SystemMonitor) getCPUUsage() (float64, error) {
	stats, err := s.readCPUStats()
	if err != nil {
		return 0, err
	}

	now := time.Now()

	// If this is the first reading, store it and return 0
	if s.lastTime.IsZero() {
		s.lastCPUStats = stats
		s.lastTime = now
		return 0, nil
	}

	// Calculate differences
	totalDiff := stats.Total - s.lastCPUStats.Total
	idleDiff := stats.Idle - s.lastCPUStats.Idle

	// Store current stats for next calculation
	s.lastCPUStats = stats
	s.lastTime = now

	if totalDiff == 0 {
		return 0, nil
	}

	// Calculate usage percentage
	usage := float64(totalDiff-idleDiff) / float64(totalDiff) * 100
	return usage, nil
}

// readCPUStats reads CPU statistics from /proc/stat
func (s *SystemMonitor) readCPUStats() (CPUStats, error) {
	var stats CPUStats

	file, err := os.Open("/proc/stat")
	if err != nil {
		return stats, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return stats, fmt.Errorf("failed to read CPU stats")
	}

	line := scanner.Text()
	fields := strings.Fields(line)
	if len(fields) < 8 || fields[0] != "cpu" {
		return stats, fmt.Errorf("invalid CPU stats format")
	}

	// Parse CPU time values
	values := make([]uint64, len(fields)-1)
	for i, field := range fields[1:] {
		if val, err := strconv.ParseUint(field, 10, 64); err == nil {
			values[i] = val
		}
	}

	stats.User = values[0]
	stats.Nice = values[1]
	stats.System = values[2]
	stats.Idle = values[3]
	if len(values) > 4 {
		stats.IOWait = values[4]
	}
	if len(values) > 5 {
		stats.IRQ = values[5]
	}
	if len(values) > 6 {
		stats.SoftIRQ = values[6]
	}
	if len(values) > 7 {
		stats.Steal = values[7]
	}
	if len(values) > 8 {
		stats.Guest = values[8]
	}

	// Calculate total
	for _, val := range values {
		stats.Total += val
	}

	return stats, nil
}

// getCPUTemperature gets CPU temperature from sensors
func (s *SystemMonitor) getCPUTemperature() int {
	// Try different methods to get CPU temperature

	// Method 1: Use sensors command
	output := lib.GetCmdOutput("sensors")
	for _, line := range output {
		if strings.Contains(line, "Core 0") || strings.Contains(line, "CPU Temperature") {
			if temp := s.parseTemperature(line); temp > 0 {
				return temp
			}
		}
	}

	// Method 2: Read from thermal zone
	thermalFiles := []string{
		"/sys/class/thermal/thermal_zone0/temp",
		"/sys/class/thermal/thermal_zone1/temp",
	}

	for _, file := range thermalFiles {
		if content, err := os.ReadFile(file); err == nil {
			if temp, err := strconv.Atoi(strings.TrimSpace(string(content))); err == nil {
				// Convert from millidegrees to degrees
				return temp / 1000
			}
		}
	}

	return 0
}

// getCPUFrequency gets current CPU frequency
func (s *SystemMonitor) getCPUFrequency() int {
	// Read CPU frequency from /proc/cpuinfo
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu MHz") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				if freq, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
					return int(freq)
				}
			}
			break
		}
	}

	return 0
}

// parseTemperature parses temperature from sensor output
func (s *SystemMonitor) parseTemperature(line string) int {
	// Look for temperature patterns like "+45.0°C" or "45°C"
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
