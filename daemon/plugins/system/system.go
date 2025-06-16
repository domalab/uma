package system

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/domalab/uma/daemon/dto"
	"github.com/domalab/uma/daemon/lib"
	"github.com/domalab/uma/daemon/logger"
)

// SystemMonitor provides system resource monitoring
type SystemMonitor struct {
	lastCPUStats CPUStats
	lastTime     time.Time
}

// CPUStats represents CPU statistics
type CPUStats struct {
	User    uint64 `json:"user"`
	Nice    uint64 `json:"nice"`
	System  uint64 `json:"system"`
	Idle    uint64 `json:"idle"`
	IOWait  uint64 `json:"iowait"`
	IRQ     uint64 `json:"irq"`
	SoftIRQ uint64 `json:"softirq"`
	Steal   uint64 `json:"steal"`
	Guest   uint64 `json:"guest"`
	Total   uint64 `json:"total"`
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

// MemoryUsageBreakdown represents memory usage by category
type MemoryUsageBreakdown struct {
	System   uint64 `json:"system_bytes"`
	VM       uint64 `json:"vm_bytes"`
	Docker   uint64 `json:"docker_bytes"`
	ZFSCache uint64 `json:"zfs_cache_bytes"`
	Other    uint64 `json:"other_bytes"`
	// Percentages
	SystemPercent   float64 `json:"system_percent"`
	VMPercent       float64 `json:"vm_percent"`
	DockerPercent   float64 `json:"docker_percent"`
	ZFSCachePercent float64 `json:"zfs_cache_percent"`
	OtherPercent    float64 `json:"other_percent"`
	// Human-readable formatted fields
	SystemFormatted   string `json:"system_formatted"`    // "4 GB"
	VMFormatted       string `json:"vm_formatted"`        // "8 GB"
	DockerFormatted   string `json:"docker_formatted"`    // "2 GB"
	ZFSCacheFormatted string `json:"zfs_cache_formatted"` // "16 GB"
	OtherFormatted    string `json:"other_formatted"`     // "2 GB"
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
	// Human-readable formatted fields
	TotalFormatted     string `json:"total_formatted"`     // "32 GB"
	AvailableFormatted string `json:"available_formatted"` // "24 GB"
	UsedFormatted      string `json:"used_formatted"`      // "8 GB"
	FreeFormatted      string `json:"free_formatted"`      // "16 GB"
	BuffersFormatted   string `json:"buffers_formatted"`   // "512 MB"
	CachedFormatted    string `json:"cached_formatted"`    // "4 GB"
	// Memory usage breakdown
	Breakdown *MemoryUsageBreakdown `json:"breakdown,omitempty"`
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

	// Populate human-readable formatted fields
	memInfo.TotalFormatted = s.formatBytes(memInfo.Total)
	memInfo.AvailableFormatted = s.formatBytes(memInfo.Available)
	memInfo.UsedFormatted = s.formatBytes(memInfo.Used)
	memInfo.FreeFormatted = s.formatBytes(memInfo.Free)
	memInfo.BuffersFormatted = s.formatBytes(memInfo.Buffers)
	memInfo.CachedFormatted = s.formatBytes(memInfo.Cached)

	// Calculate memory usage breakdown
	s.calculateMemoryBreakdown(memInfo)

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

// GetEnhancedTemperatureData returns comprehensive temperature and fan monitoring data
func (s *SystemMonitor) GetEnhancedTemperatureData() (*EnhancedTemperatureData, error) {
	data := &EnhancedTemperatureData{
		Sensors: make(map[string]SensorChip),
		Fans:    make(map[string]FanInput),
	}

	// Get sensor data using sensors command
	output := lib.GetCmdOutput("sensors", "-A", "-u")
	s.parseEnhancedSensorOutput(output, data)

	return data, nil
}

// parseEnhancedSensorOutput parses detailed sensor output to extract chip-specific data
func (s *SystemMonitor) parseEnhancedSensorOutput(output []string, data *EnhancedTemperatureData) {
	var currentChip *SensorChip
	var currentChipName string

	for _, line := range output {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Detect chip headers (e.g., "coretemp-isa-0000", "nct6798-isa-0290")
		if !strings.HasPrefix(line, " ") && strings.Contains(line, "-") && !strings.Contains(line, ":") {
			currentChipName = line
			currentChip = &SensorChip{
				Name:         currentChipName,
				Temperatures: make(map[string]TemperatureInput),
				Fans:         make(map[string]FanInput),
			}
			data.Sensors[currentChipName] = *currentChip
			continue
		}

		// Parse adapter line
		if strings.HasPrefix(line, "Adapter:") && currentChip != nil {
			currentChip.Adapter = strings.TrimSpace(strings.TrimPrefix(line, "Adapter:"))
			data.Sensors[currentChipName] = *currentChip
			continue
		}

		// Parse temperature and fan data
		if currentChip != nil && strings.Contains(line, ":") {
			s.parseTemperatureOrFanLine(line, currentChip, currentChipName, data)
		}
	}
}

// parseTemperatureOrFanLine parses individual temperature or fan sensor lines
func (s *SystemMonitor) parseTemperatureOrFanLine(line string, currentChip *SensorChip, chipName string, data *EnhancedTemperatureData) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return
	}

	key := strings.TrimSpace(parts[0])
	valueStr := strings.TrimSpace(parts[1])

	// Parse temperature inputs
	if strings.Contains(key, "temp") && strings.Contains(key, "_input") {
		tempName := strings.TrimSuffix(key, "_input")
		if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
			// Convert to millidegrees Celsius
			millidegrees := int(value * 1000)

			tempInput := currentChip.Temperatures[tempName]
			tempInput.Value = millidegrees
			tempInput.Label = s.getTemperatureLabel(tempName, chipName)
			currentChip.Temperatures[tempName] = tempInput
			data.Sensors[chipName] = *currentChip
		}
	}

	// Parse temperature critical/max/min values
	if strings.Contains(key, "temp") && (strings.Contains(key, "_crit") || strings.Contains(key, "_max") || strings.Contains(key, "_min")) {
		tempName := strings.Split(key, "_")[0]
		if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
			millidegrees := int(value * 1000)

			tempInput := currentChip.Temperatures[tempName]
			if strings.Contains(key, "_crit") {
				tempInput.Critical = millidegrees
			} else if strings.Contains(key, "_max") {
				tempInput.Max = millidegrees
			} else if strings.Contains(key, "_min") {
				tempInput.Min = millidegrees
			}
			currentChip.Temperatures[tempName] = tempInput
			data.Sensors[chipName] = *currentChip
		}
	}

	// Parse fan inputs
	if strings.Contains(key, "fan") && strings.Contains(key, "_input") {
		fanName := strings.TrimSuffix(key, "_input")
		if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
			rpm := int(value)

			fanInput := FanInput{
				Label: s.getFanLabel(fanName, chipName),
				Input: rpm,
			}

			currentChip.Fans[fanName] = fanInput
			data.Fans[fanName] = fanInput
			data.Sensors[chipName] = *currentChip
		}
	}

	// Parse fan min/max values
	if strings.Contains(key, "fan") && (strings.Contains(key, "_min") || strings.Contains(key, "_max")) {
		fanName := strings.Split(key, "_")[0]
		if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
			rpm := int(value)

			fanInput := currentChip.Fans[fanName]
			if strings.Contains(key, "_min") {
				fanInput.Min = rpm
			} else if strings.Contains(key, "_max") {
				fanInput.Max = rpm
			}
			currentChip.Fans[fanName] = fanInput
			data.Fans[fanName] = fanInput
			data.Sensors[chipName] = *currentChip
		}
	}
}

// getTemperatureLabel returns a human-readable label for temperature sensors
func (s *SystemMonitor) getTemperatureLabel(tempName, chipName string) string {
	// Map common temperature sensor names to readable labels
	labelMap := map[string]string{
		"temp1": "Package id 0",
		"temp2": "Core 0",
		"temp3": "Core 1",
		"temp4": "Core 2",
		"temp5": "Core 3",
	}

	if strings.Contains(chipName, "coretemp") {
		if label, exists := labelMap[tempName]; exists {
			return label
		}
		return fmt.Sprintf("Core Temperature %s", strings.TrimPrefix(tempName, "temp"))
	}

	if strings.Contains(chipName, "nct") {
		switch tempName {
		case "temp1":
			return "SYSTIN"
		case "temp2":
			return "AUXTIN0"
		case "temp3":
			return "AUXTIN1"
		case "temp4":
			return "AUXTIN2"
		default:
			return fmt.Sprintf("Temperature %s", strings.TrimPrefix(tempName, "temp"))
		}
	}

	return fmt.Sprintf("Temperature %s", strings.TrimPrefix(tempName, "temp"))
}

// getFanLabel returns a human-readable label for fan sensors
func (s *SystemMonitor) getFanLabel(fanName, chipName string) string {
	fanMap := map[string]string{
		"fan1": "CPU Fan",
		"fan2": "System Fan 1",
		"fan3": "System Fan 2",
		"fan4": "System Fan 3",
		"fan5": "System Fan 4",
	}

	if label, exists := fanMap[fanName]; exists {
		return label
	}

	return fmt.Sprintf("Fan %s", strings.TrimPrefix(fanName, "fan"))
}

// calculateMemoryBreakdown calculates memory usage breakdown by category
func (s *SystemMonitor) calculateMemoryBreakdown(memInfo *MemoryInfo) {
	breakdown := &MemoryUsageBreakdown{}

	// Get VM memory usage
	breakdown.VM = s.getVMMemoryUsage()

	// Get Docker memory usage
	breakdown.Docker = s.getDockerMemoryUsage()

	// Get ZFS cache usage
	breakdown.ZFSCache = s.getZFSCacheUsage()

	// Calculate system memory (everything else)
	totalUsedByCategories := breakdown.VM + breakdown.Docker + breakdown.ZFSCache
	if memInfo.Used > totalUsedByCategories {
		breakdown.System = memInfo.Used - totalUsedByCategories
	} else {
		breakdown.System = memInfo.Used
		breakdown.Other = 0
	}

	// Calculate any remaining "other" usage
	if totalUsedByCategories > memInfo.Used {
		breakdown.Other = totalUsedByCategories - memInfo.Used
	}

	// Calculate percentages
	if memInfo.Total > 0 {
		breakdown.SystemPercent = float64(breakdown.System) / float64(memInfo.Total) * 100
		breakdown.VMPercent = float64(breakdown.VM) / float64(memInfo.Total) * 100
		breakdown.DockerPercent = float64(breakdown.Docker) / float64(memInfo.Total) * 100
		breakdown.ZFSCachePercent = float64(breakdown.ZFSCache) / float64(memInfo.Total) * 100
		breakdown.OtherPercent = float64(breakdown.Other) / float64(memInfo.Total) * 100
	}

	// Format human-readable values
	breakdown.SystemFormatted = s.formatBytes(breakdown.System)
	breakdown.VMFormatted = s.formatBytes(breakdown.VM)
	breakdown.DockerFormatted = s.formatBytes(breakdown.Docker)
	breakdown.ZFSCacheFormatted = s.formatBytes(breakdown.ZFSCache)
	breakdown.OtherFormatted = s.formatBytes(breakdown.Other)

	memInfo.Breakdown = breakdown
}

// getVMMemoryUsage returns total memory usage by VMs
func (s *SystemMonitor) getVMMemoryUsage() uint64 {
	var totalVMMemory uint64

	// Parse /proc/meminfo for KVM/QEMU memory usage
	output := lib.GetCmdOutput("ps", "-eo", "comm,rss", "--no-headers")
	for _, line := range output {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			comm := fields[0]
			if strings.Contains(comm, "qemu") || strings.Contains(comm, "kvm") {
				if rss, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
					totalVMMemory += rss * 1024 // Convert KB to bytes
				}
			}
		}
	}

	return totalVMMemory
}

// getDockerMemoryUsage returns total memory usage by Docker containers
func (s *SystemMonitor) getDockerMemoryUsage() uint64 {
	var totalDockerMemory uint64

	// Get Docker memory usage from cgroup
	cgroupPaths := []string{
		"/sys/fs/cgroup/memory/docker/memory.usage_in_bytes",
		"/sys/fs/cgroup/memory/system.slice/docker.service/memory.usage_in_bytes",
	}

	for _, path := range cgroupPaths {
		if data, err := os.ReadFile(path); err == nil {
			if usage, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64); err == nil {
				totalDockerMemory = usage
				break
			}
		}
	}

	// Fallback: parse Docker processes
	if totalDockerMemory == 0 {
		output := lib.GetCmdOutput("ps", "-eo", "comm,rss", "--no-headers")
		for _, line := range output {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				comm := fields[0]
				if strings.Contains(comm, "docker") || strings.Contains(comm, "containerd") {
					if rss, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
						totalDockerMemory += rss * 1024 // Convert KB to bytes
					}
				}
			}
		}
	}

	return totalDockerMemory
}

// getZFSCacheUsage returns ZFS ARC cache usage
func (s *SystemMonitor) getZFSCacheUsage() uint64 {
	// Check if ZFS is available
	if _, err := os.Stat("/proc/spl/kstat/zfs/arcstats"); err != nil {
		return 0
	}

	// Read ZFS ARC stats
	data, err := os.ReadFile("/proc/spl/kstat/zfs/arcstats")
	if err != nil {
		return 0
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "size ") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				if size, err := strconv.ParseUint(fields[2], 10, 64); err == nil {
					return size
				}
			}
		}
	}

	return 0
}

// formatBytes converts bytes to human-readable format
func (s *SystemMonitor) formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	if exp >= len(units) {
		exp = len(units) - 1
	}

	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// TemperatureInput represents a temperature sensor input
type TemperatureInput struct {
	Label    string `json:"label"`
	Value    int    `json:"temp_input"` // Temperature in millidegrees Celsius
	Critical int    `json:"temp_crit,omitempty"`
	Max      int    `json:"temp_max,omitempty"`
	Min      int    `json:"temp_min,omitempty"`
}

// FanInput represents a fan sensor input
type FanInput struct {
	Label string `json:"label"`
	Input int    `json:"input"` // RPM
	Min   int    `json:"min,omitempty"`
	Max   int    `json:"max,omitempty"`
}

// SensorChip represents a hardware sensor chip
type SensorChip struct {
	Name         string                      `json:"name"`
	Adapter      string                      `json:"adapter,omitempty"`
	Temperatures map[string]TemperatureInput `json:"temperatures,omitempty"`
	Fans         map[string]FanInput         `json:"fans,omitempty"`
}

// EnhancedTemperatureData represents comprehensive temperature and fan monitoring data
type EnhancedTemperatureData struct {
	Sensors map[string]SensorChip `json:"sensors"`
	Fans    map[string]FanInput   `json:"fans"`
}

// FilesystemInfo represents filesystem usage information
type FilesystemInfo struct {
	Path        string  `json:"path"`
	Type        string  `json:"type"` // boot, log, docker_vdisk, system
	Total       uint64  `json:"total_bytes"`
	Used        uint64  `json:"used_bytes"`
	Free        uint64  `json:"free_bytes"`
	UsedPercent float64 `json:"used_percent"`
	// Human-readable formatted fields
	TotalFormatted string `json:"total_formatted"` // "16 GB"
	UsedFormatted  string `json:"used_formatted"`  // "8 GB"
	FreeFormatted  string `json:"free_formatted"`  // "8 GB"
}

// FilesystemData represents comprehensive filesystem monitoring data
type FilesystemData struct {
	Boot        *FilesystemInfo `json:"boot,omitempty"`
	Log         *FilesystemInfo `json:"log,omitempty"`
	DockerVDisk *FilesystemInfo `json:"docker_vdisk,omitempty"`
	System      *FilesystemInfo `json:"system,omitempty"`
}

// GetFilesystemData returns comprehensive filesystem usage information
func (s *SystemMonitor) GetFilesystemData() (*FilesystemData, error) {
	data := &FilesystemData{}

	// Get boot filesystem usage
	if bootInfo := s.getFilesystemInfo("/boot", "boot"); bootInfo != nil {
		data.Boot = bootInfo
	}

	// Get log filesystem usage
	if logInfo := s.getFilesystemInfo("/var/log", "log"); logInfo != nil {
		data.Log = logInfo
	}

	// Get Docker vDisk usage
	if dockerInfo := s.getDockerVDiskInfo(); dockerInfo != nil {
		data.DockerVDisk = dockerInfo
	}

	// Get system root filesystem usage
	if systemInfo := s.getFilesystemInfo("/", "system"); systemInfo != nil {
		data.System = systemInfo
	}

	return data, nil
}

// getFilesystemInfo returns filesystem usage information for a given path
func (s *SystemMonitor) getFilesystemInfo(path, fsType string) *FilesystemInfo {
	// Use df command to get filesystem usage
	output := lib.GetCmdOutput("df", "-B1", path)
	if len(output) < 2 {
		return nil
	}

	// Parse df output (skip header line)
	fields := strings.Fields(output[1])
	if len(fields) < 6 {
		return nil
	}

	// Parse values from df output
	total, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return nil
	}

	used, err := strconv.ParseUint(fields[2], 10, 64)
	if err != nil {
		return nil
	}

	free, err := strconv.ParseUint(fields[3], 10, 64)
	if err != nil {
		return nil
	}

	var usedPercent float64
	if total > 0 {
		usedPercent = float64(used) / float64(total) * 100
	}

	info := &FilesystemInfo{
		Path:        path,
		Type:        fsType,
		Total:       total,
		Used:        used,
		Free:        free,
		UsedPercent: usedPercent,
	}

	// Populate formatted fields
	info.TotalFormatted = s.formatBytes(total)
	info.UsedFormatted = s.formatBytes(used)
	info.FreeFormatted = s.formatBytes(free)

	return info
}

// getDockerVDiskInfo returns Docker vDisk usage information
func (s *SystemMonitor) getDockerVDiskInfo() *FilesystemInfo {
	// Check common Docker vDisk locations
	dockerPaths := []string{
		"/var/lib/docker",
		"/mnt/user/system/docker/docker.img",
		"/mnt/cache/system/docker/docker.img",
	}

	for _, path := range dockerPaths {
		if info := s.getFilesystemInfo(path, "docker_vdisk"); info != nil {
			return info
		}
	}

	// If no Docker vDisk found, check if Docker is using overlay2
	if info := s.getFilesystemInfo("/var/lib/docker", "docker_vdisk"); info != nil {
		return info
	}

	return nil
}
