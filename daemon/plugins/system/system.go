package system

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	Model          string  `json:"model"`
	Cores          int     `json:"cores"`
	Threads        int     `json:"threads"`
	ThreadsPerCore int     `json:"threads_per_core,omitempty"`
	Sockets        int     `json:"sockets,omitempty"`
	Usage          float64 `json:"usage_percent"`
	Temperature    int     `json:"temperature,omitempty"`
	Frequency      int     `json:"frequency_mhz,omitempty"`
	MaxFrequency   int     `json:"max_frequency_mhz,omitempty"`
	MinFrequency   int     `json:"min_frequency_mhz,omitempty"`
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
	Connected   bool   `json:"connected"`
	SpeedMbps   int    `json:"speed_mbps,omitempty"`
	Duplex      string `json:"duplex,omitempty"`
}

// ParityDiskInfo represents parity disk information
type ParityDiskInfo struct {
	Device           string `json:"device"`
	SerialNumber     string `json:"serial_number"`
	Capacity         string `json:"capacity"`
	Temperature      string `json:"temperature"`
	SmartStatus      string `json:"smart_status"`
	PowerState       string `json:"power_state"`
	SpinDownDelay    string `json:"spin_down_delay"`
	HealthAssessment string `json:"health_assessment"`
	LastUpdated      string `json:"last_updated"`
	State            int    `json:"state"`
	DeviceName       string `json:"device_name"`
}

// ParityCheckInfo represents parity check status information
type ParityCheckInfo struct {
	Status       string `json:"status"`
	Progress     int    `json:"progress,omitempty"`
	Speed        string `json:"speed,omitempty"`
	Errors       int    `json:"errors,omitempty"`
	LastCheck    string `json:"last_check,omitempty"`
	Duration     string `json:"duration,omitempty"`
	LastStatus   string `json:"last_status,omitempty"`
	LastSpeed    string `json:"last_speed,omitempty"`
	NextCheck    string `json:"next_check,omitempty"`
	LastUpdated  string `json:"last_updated"`
	IsRunning    bool   `json:"is_running"`
	Action       string `json:"action"`
	ResyncActive int    `json:"resync_active"`
}

// MDCmdStatus represents mdcmd status information
type MDCmdStatus struct {
	State        string            `json:"state"`
	ResyncAction string            `json:"resync_action"`
	ResyncSize   uint64            `json:"resync_size"`
	ResyncCorr   int               `json:"resync_corr"`
	Resync       int               `json:"resync"`
	ResyncPos    uint64            `json:"resync_pos"`
	ResyncDt     int               `json:"resync_dt"`
	ResyncDb     int               `json:"resync_db"`
	DiskStates   map[string]int    `json:"disk_states"`
	DiskIds      map[string]string `json:"disk_ids"`
	DeviceNames  map[string]string `json:"device_names"`
	ParityDisk   *ParityDiskInfo   `json:"parity_disk,omitempty"`
	ParityCheck  *ParityCheckInfo  `json:"parity_check,omitempty"`
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

		// Get additional network interface information
		s.getNetworkInterfaceDetails(&netInfo)

		networks = append(networks, netInfo)
	}

	return networks, nil
}

// getNetworkInterfaceDetails gets additional network interface information
func (s *SystemMonitor) getNetworkInterfaceDetails(netInfo *NetworkInfo) {
	interfacePath := fmt.Sprintf("/sys/class/net/%s", netInfo.Interface)

	// Check if interface is connected (carrier status)
	if content, err := os.ReadFile(fmt.Sprintf("%s/carrier", interfacePath)); err == nil {
		if strings.TrimSpace(string(content)) == "1" {
			netInfo.Connected = true
		}
	}

	// Get interface speed
	if content, err := os.ReadFile(fmt.Sprintf("%s/speed", interfacePath)); err == nil {
		if speed, err := strconv.Atoi(strings.TrimSpace(string(content))); err == nil && speed > 0 {
			netInfo.SpeedMbps = speed
		}
	}

	// Get duplex mode
	if content, err := os.ReadFile(fmt.Sprintf("%s/duplex", interfacePath)); err == nil {
		duplex := strings.TrimSpace(string(content))
		if duplex == "full" || duplex == "half" {
			netInfo.Duplex = duplex
		} else {
			netInfo.Duplex = "unknown"
		}
	} else {
		netInfo.Duplex = "unknown"
	}

	// If interface is not connected, set speed and duplex as unavailable
	if !netInfo.Connected {
		netInfo.SpeedMbps = 0
		netInfo.Duplex = "unknown"
	}
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
	physicalIDs := make(map[string]bool)
	siblingsPerCore := 0

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
		} else if strings.HasPrefix(line, "physical id") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				physicalIDs[strings.TrimSpace(parts[1])] = true
			}
		} else if strings.HasPrefix(line, "siblings") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				if siblings, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
					siblingsPerCore = siblings
				}
			}
		}
	}

	cpuInfo.Threads = coreCount
	if cpuInfo.Cores == 0 {
		cpuInfo.Cores = coreCount
	}

	// Calculate sockets
	cpuInfo.Sockets = len(physicalIDs)
	if cpuInfo.Sockets == 0 {
		cpuInfo.Sockets = 1 // Default to 1 socket if not detected
	}

	// Calculate threads per core
	if cpuInfo.Cores > 0 {
		cpuInfo.ThreadsPerCore = cpuInfo.Threads / cpuInfo.Cores
	} else if siblingsPerCore > 0 && cpuInfo.Cores > 0 {
		cpuInfo.ThreadsPerCore = siblingsPerCore / cpuInfo.Cores
	}

	// Get CPU frequency information
	s.getCPUFrequencyDetails(cpuInfo)

	return nil
}

// getCPUFrequencyDetails gets CPU frequency information from sysfs
func (s *SystemMonitor) getCPUFrequencyDetails(cpuInfo *CPUInfo) {
	// Try to get max frequency from cpuinfo_max_freq
	if content, err := os.ReadFile("/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_max_freq"); err == nil {
		if freq, err := strconv.Atoi(strings.TrimSpace(string(content))); err == nil {
			cpuInfo.MaxFrequency = freq / 1000 // Convert from kHz to MHz
		}
	}

	// Try to get min frequency from cpuinfo_min_freq
	if content, err := os.ReadFile("/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_min_freq"); err == nil {
		if freq, err := strconv.Atoi(strings.TrimSpace(string(content))); err == nil {
			cpuInfo.MinFrequency = freq / 1000 // Convert from kHz to MHz
		}
	}

	// If sysfs files are not available, try to parse from /proc/cpuinfo
	if cpuInfo.MaxFrequency == 0 || cpuInfo.MinFrequency == 0 {
		s.getCPUFrequencyFromCpuinfo(cpuInfo)
	}
}

// getCPUFrequencyFromCpuinfo gets frequency info from /proc/cpuinfo as fallback
func (s *SystemMonitor) getCPUFrequencyFromCpuinfo(cpuInfo *CPUInfo) {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu MHz") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				if freq, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
					// Use current frequency as a reference if max/min not available
					currentFreq := int(freq)
					if cpuInfo.MaxFrequency == 0 {
						cpuInfo.MaxFrequency = currentFreq
					}
					if cpuInfo.MinFrequency == 0 {
						cpuInfo.MinFrequency = currentFreq
					}
				}
			}
		}
	}
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

// GetMDCmdStatus returns mdcmd status information including parity disk and check data
func (s *SystemMonitor) GetMDCmdStatus() (*MDCmdStatus, error) {
	cmd := exec.Command("mdcmd", "status")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute mdcmd: %v", err)
	}

	status := &MDCmdStatus{
		DiskStates:  make(map[string]int),
		DiskIds:     make(map[string]string),
		DeviceNames: make(map[string]string),
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch {
		case key == "mdState":
			status.State = value
		case key == "mdResyncAction":
			status.ResyncAction = value
		case key == "mdResyncSize":
			if val, err := strconv.ParseUint(value, 10, 64); err == nil {
				status.ResyncSize = val
			}
		case key == "mdResyncCorr":
			if val, err := strconv.Atoi(value); err == nil {
				status.ResyncCorr = val
			}
		case key == "mdResync":
			if val, err := strconv.Atoi(value); err == nil {
				status.Resync = val
			}
		case key == "mdResyncPos":
			if val, err := strconv.ParseUint(value, 10, 64); err == nil {
				status.ResyncPos = val
			}
		case key == "mdResyncDt":
			if val, err := strconv.Atoi(value); err == nil {
				status.ResyncDt = val
			}
		case key == "mdResyncDb":
			if val, err := strconv.Atoi(value); err == nil {
				status.ResyncDb = val
			}
		case strings.HasPrefix(key, "diskState."):
			diskNum := strings.TrimPrefix(key, "diskState.")
			if val, err := strconv.Atoi(value); err == nil {
				status.DiskStates[diskNum] = val
			}
		case strings.HasPrefix(key, "diskId."):
			diskNum := strings.TrimPrefix(key, "diskId.")
			status.DiskIds[diskNum] = value
		case strings.HasPrefix(key, "rdevName."):
			diskNum := strings.TrimPrefix(key, "rdevName.")
			status.DeviceNames[diskNum] = value
		}
	}

	// Process parity disk information (disk 0 is typically parity)
	if diskState, exists := status.DiskStates["0"]; exists && diskState > 0 {
		status.ParityDisk = s.buildParityDiskInfo(status, "0")
	}

	// Process parity check information
	status.ParityCheck = s.buildParityCheckInfo(status)

	return status, nil
}

// buildParityDiskInfo builds parity disk information from mdcmd data with SMART integration
func (s *SystemMonitor) buildParityDiskInfo(status *MDCmdStatus, diskNum string) *ParityDiskInfo {
	timestamp := time.Now().UTC().Format(time.RFC3339)

	parityDisk := &ParityDiskInfo{
		LastUpdated:      timestamp,
		State:            status.DiskStates[diskNum],
		SerialNumber:     status.DiskIds[diskNum],
		DeviceName:       status.DeviceNames[diskNum],
		Device:           fmt.Sprintf("/dev/%s", status.DeviceNames[diskNum]),
		PowerState:       "Active",  // Will be enhanced with actual detection
		SmartStatus:      "Unknown", // Will be enhanced with SMART data
		HealthAssessment: "Unknown", // Will be enhanced with health logic
		SpinDownDelay:    "Never",   // Will be enhanced with actual settings
		Capacity:         "Unknown", // Will be enhanced with disk size
		Temperature:      "N/A",     // Will be enhanced with SMART temperature
	}

	// Enhance with real SMART data if device is available
	if parityDisk.DeviceName != "" && status.DiskStates[diskNum] > 0 {
		s.enhanceParityDiskWithSMARTData(parityDisk)
	}

	// Determine basic state information
	switch status.DiskStates[diskNum] {
	case 7: // Active/healthy state in Unraid
		if parityDisk.PowerState == "Unknown" {
			parityDisk.PowerState = "Active"
		}
		if parityDisk.HealthAssessment == "Unknown" {
			parityDisk.HealthAssessment = "Healthy"
		}
	case 0: // Disabled/missing
		parityDisk.PowerState = "Standby"
		parityDisk.HealthAssessment = "Missing"
		parityDisk.Temperature = "N/A (Standby)"
	default:
		if parityDisk.PowerState == "Unknown" {
			parityDisk.PowerState = "Unknown"
		}
		if parityDisk.HealthAssessment == "Unknown" {
			parityDisk.HealthAssessment = "Unknown"
		}
	}

	return parityDisk
}

// enhanceParityDiskWithSMARTData enhances parity disk info with real SMART data
func (s *SystemMonitor) enhanceParityDiskWithSMARTData(parityDisk *ParityDiskInfo) {
	devicePath := parityDisk.Device
	if devicePath == "" {
		return
	}

	// Get SMART health status
	if health := s.getSMARTHealth(devicePath); health != "" {
		parityDisk.SmartStatus = health
		// Update health assessment based on SMART status
		if health == "PASSED" {
			parityDisk.HealthAssessment = "Healthy"
		} else if health == "FAILED" {
			parityDisk.HealthAssessment = "Failing"
		}
	}

	// Get disk temperature from SMART data
	if temp := s.getSMARTTemperature(devicePath); temp > 0 {
		parityDisk.Temperature = fmt.Sprintf("%d°C", temp)
	}

	// Get disk capacity
	if capacity := s.getDiskCapacity(devicePath); capacity > 0 {
		parityDisk.Capacity = s.formatBytes(uint64(capacity))
	}

	// Get power state using existing logic from storage plugin
	if powerState := s.getDiskPowerState(devicePath); powerState != "" {
		parityDisk.PowerState = powerState
	}

	// Get spin down delay
	if spinDown := s.getDiskSpinDownDelay(devicePath); spinDown != "" {
		parityDisk.SpinDownDelay = spinDown
	}
}

// getSMARTHealth gets SMART health status for a device
func (s *SystemMonitor) getSMARTHealth(devicePath string) string {
	actualDevice := s.resolveDevicePath(devicePath)
	if actualDevice == "" {
		return "Unknown"
	}

	// Use smartctl to check health
	output := lib.GetCmdOutput("smartctl", "-H", actualDevice)
	for _, line := range output {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "SMART overall-health self-assessment test result:") {
			if strings.Contains(line, "PASSED") {
				return "PASSED"
			} else {
				return "FAILED"
			}
		}
		// For NVMe drives
		if strings.Contains(line, "SMART Health Status:") {
			if strings.Contains(line, "OK") {
				return "PASSED"
			} else {
				return "FAILED"
			}
		}
	}

	return "Unknown"
}

// getSMARTTemperature gets disk temperature from SMART data
func (s *SystemMonitor) getSMARTTemperature(devicePath string) int {
	actualDevice := s.resolveDevicePath(devicePath)
	if actualDevice == "" {
		return 0
	}

	// Use smartctl to get disk temperature
	output := lib.GetCmdOutput("smartctl", "-A", actualDevice)
	for _, line := range output {
		if strings.Contains(line, "Temperature_Celsius") || strings.Contains(line, "Airflow_Temperature_Cel") {
			fields := strings.Fields(line)
			if len(fields) >= 10 {
				if temp, err := strconv.Atoi(fields[9]); err == nil {
					return temp
				}
			}
		}
	}

	return 0
}

// getDiskCapacity gets disk capacity in bytes
func (s *SystemMonitor) getDiskCapacity(devicePath string) int64 {
	actualDevice := s.resolveDevicePath(devicePath)
	if actualDevice == "" {
		return 0
	}

	// Use smartctl to get disk capacity
	output := lib.GetCmdOutput("smartctl", "-i", actualDevice)
	for _, line := range output {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "User Capacity:") {
			// Parse capacity from line like "User Capacity:    8,001,563,222,016 bytes [8.00 TB]"
			if idx := strings.Index(line, "["); idx != -1 {
				if endIdx := strings.Index(line[idx:], "]"); endIdx != -1 {
					capacityStr := strings.TrimSpace(line[idx+1 : idx+endIdx])
					return s.parseCapacityString(capacityStr)
				}
			}
		}
	}

	// Fallback: try to get size from /sys/block
	deviceName := strings.TrimPrefix(actualDevice, "/dev/")
	if content, err := os.ReadFile(fmt.Sprintf("/sys/block/%s/size", deviceName)); err == nil {
		if sectors, err := strconv.ParseInt(strings.TrimSpace(string(content)), 10, 64); err == nil {
			return sectors * 512 // Convert sectors to bytes (512 bytes per sector)
		}
	}

	return 0
}

// parseCapacityString parses capacity strings like "8.00 TB" to bytes
func (s *SystemMonitor) parseCapacityString(capacityStr string) int64 {
	parts := strings.Fields(capacityStr)
	if len(parts) != 2 {
		return 0
	}

	value, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0
	}

	unit := strings.ToUpper(parts[1])
	switch unit {
	case "TB":
		return int64(value * 1024 * 1024 * 1024 * 1024)
	case "GB":
		return int64(value * 1024 * 1024 * 1024)
	case "MB":
		return int64(value * 1024 * 1024)
	case "KB":
		return int64(value * 1024)
	default:
		return int64(value)
	}
}

// getDiskPowerState gets disk power state
func (s *SystemMonitor) getDiskPowerState(devicePath string) string {
	actualDevice := s.resolveDevicePath(devicePath)
	if actualDevice == "" {
		return "unknown"
	}

	// Skip NVMe devices - they're always active
	if strings.Contains(strings.ToLower(actualDevice), "nvme") {
		return "Active"
	}

	// Use smartctl with standby detection
	cmd := exec.Command("smartctl", "-n", "standby", actualDevice)
	err := cmd.Run()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			switch exitError.ExitCode() {
			case 2:
				return "Standby"
			case 0:
				return "Active"
			default:
				return "Unknown"
			}
		}
		return "Unknown"
	}

	return "Active"
}

// getDiskSpinDownDelay gets disk spin down delay setting
func (s *SystemMonitor) getDiskSpinDownDelay(devicePath string) string {
	// This would need to read from Unraid configuration files
	// For now, return a default value
	return "Never"
}

// resolveDevicePath resolves device path to actual device
func (s *SystemMonitor) resolveDevicePath(devicePath string) string {
	// If it's already a direct device path, return it
	if strings.HasPrefix(devicePath, "/dev/sd") || strings.HasPrefix(devicePath, "/dev/nvme") {
		return devicePath
	}

	// Try to resolve symlinks
	if resolved, err := filepath.EvalSymlinks(devicePath); err == nil {
		return resolved
	}

	return devicePath
}

// buildParityCheckInfo builds parity check information from mdcmd data with enhanced progress tracking
func (s *SystemMonitor) buildParityCheckInfo(status *MDCmdStatus) *ParityCheckInfo {
	timestamp := time.Now().UTC().Format(time.RFC3339)

	parityCheck := &ParityCheckInfo{
		LastUpdated:  timestamp,
		Action:       status.ResyncAction,
		ResyncActive: status.Resync,
		IsRunning:    status.Resync > 0,
		Errors:       status.ResyncCorr,
	}

	// Determine status and calculate progress based on resync action and state
	switch status.ResyncAction {
	case "check P":
		if status.Resync > 0 {
			parityCheck.Status = "Running"
			// Calculate progress percentage
			if status.ResyncSize > 0 && status.ResyncPos > 0 {
				parityCheck.Progress = int((status.ResyncPos * 100) / status.ResyncSize)
			}
			// Calculate current speed
			if speed := s.calculateParityCheckSpeed(status); speed != "" {
				parityCheck.Speed = speed
			} else {
				parityCheck.Speed = "Calculating..."
			}
		} else {
			parityCheck.Status = "Idle"
			parityCheck.LastStatus = "Success" // Default assumption
		}
	case "IDLE":
		parityCheck.Status = "Idle"
		parityCheck.IsRunning = false
	default:
		if status.Resync > 0 {
			parityCheck.Status = "Running"
			// Calculate progress for other operations
			if status.ResyncSize > 0 && status.ResyncPos > 0 {
				parityCheck.Progress = int((status.ResyncPos * 100) / status.ResyncSize)
			}
		} else {
			parityCheck.Status = "Idle"
		}
	}

	// Get historical parity check data
	s.enhanceParityCheckWithHistoricalData(parityCheck)

	return parityCheck
}

// calculateParityCheckSpeed calculates current parity check speed
func (s *SystemMonitor) calculateParityCheckSpeed(status *MDCmdStatus) string {
	// Speed calculation based on mdResyncDt (time delta) and mdResyncDb (data delta)
	if status.ResyncDt > 0 && status.ResyncDb > 0 {
		// Calculate speed in MB/s
		// ResyncDb is typically in KB, ResyncDt is in deciseconds (0.1s)
		speedKBps := float64(status.ResyncDb) / (float64(status.ResyncDt) / 10.0)
		speedMBps := speedKBps / 1024.0

		if speedMBps >= 1.0 {
			return fmt.Sprintf("%.1f MB/s", speedMBps)
		} else {
			return fmt.Sprintf("%.0f KB/s", speedKBps)
		}
	}

	return ""
}

// enhanceParityCheckWithHistoricalData adds historical parity check information
func (s *SystemMonitor) enhanceParityCheckWithHistoricalData(parityCheck *ParityCheckInfo) {
	// Try to read parity check history from Unraid logs
	if lastCheck := s.getLastParityCheckInfo(); lastCheck != nil {
		parityCheck.LastCheck = lastCheck.Date
		parityCheck.Duration = lastCheck.Duration
		parityCheck.LastSpeed = lastCheck.Speed
		parityCheck.LastStatus = lastCheck.Status
	} else {
		// Set default values if no historical data available
		parityCheck.LastCheck = "Unknown"
		parityCheck.Duration = "Unknown"
		parityCheck.LastSpeed = "Unknown"
		parityCheck.LastStatus = "Unknown"
	}

	// Try to get next scheduled check
	if nextCheck := s.getNextScheduledParityCheck(); nextCheck != "" {
		parityCheck.NextCheck = nextCheck
	} else {
		parityCheck.NextCheck = "Unknown"
	}
}

// ParityCheckHistoryInfo represents historical parity check information
type ParityCheckHistoryInfo struct {
	Date     string
	Duration string
	Speed    string
	Status   string
}

// getLastParityCheckInfo retrieves the last parity check information from logs
func (s *SystemMonitor) getLastParityCheckInfo() *ParityCheckHistoryInfo {
	// Try to read from syslog for parity check completion messages
	output := lib.GetCmdOutput("grep", "-i", "parity", "/var/log/syslog")

	var lastCheck *ParityCheckHistoryInfo

	for _, line := range output {
		// Look for parity check completion messages
		if strings.Contains(line, "parity check") && (strings.Contains(line, "completed") || strings.Contains(line, "finished")) {
			if check := s.parseParityCheckLogLine(line); check != nil {
				lastCheck = check
			}
		}
	}

	// If no recent data in syslog, try to read from Unraid's parity history file
	if lastCheck == nil {
		lastCheck = s.readParityHistoryFile()
	}

	return lastCheck
}

// parseParityCheckLogLine parses a parity check log line
func (s *SystemMonitor) parseParityCheckLogLine(line string) *ParityCheckHistoryInfo {
	// Example log line: "Dec 15 14:30:25 Tower kernel: md: parity check completed in 2h 15m at 42.1 MB/s"

	// Extract date
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return nil
	}

	date := fmt.Sprintf("%s %s %s", parts[0], parts[1], parts[2])

	check := &ParityCheckHistoryInfo{
		Date:   date,
		Status: "Success",
	}

	// Extract duration and speed
	if strings.Contains(line, "completed in") {
		// Look for duration pattern like "2h 15m"
		if idx := strings.Index(line, "completed in "); idx != -1 {
			remaining := line[idx+13:]
			if atIdx := strings.Index(remaining, " at "); atIdx != -1 {
				check.Duration = strings.TrimSpace(remaining[:atIdx])
				// Extract speed
				speedPart := remaining[atIdx+4:]
				if spaceIdx := strings.Index(speedPart, " "); spaceIdx != -1 {
					check.Speed = strings.TrimSpace(speedPart[:spaceIdx+5]) // Include "MB/s"
				}
			}
		}
	}

	return check
}

// readParityHistoryFile reads parity history from Unraid's history file
func (s *SystemMonitor) readParityHistoryFile() *ParityCheckHistoryInfo {
	// Unraid stores parity check history in /boot/config/parity-checks.log
	content, err := os.ReadFile("/boot/config/parity-checks.log")
	if err != nil {
		return nil
	}

	lines := strings.Split(string(content), "\n")
	// Get the last non-empty line
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			return s.parseParityHistoryLine(line)
		}
	}

	return nil
}

// parseParityHistoryLine parses a line from parity history file
func (s *SystemMonitor) parseParityHistoryLine(line string) *ParityCheckHistoryInfo {
	// Actual Unraid format: "2025 Jun 16 15:38:06|3230|0|-4|0|check P|15625879500"
	// Fields: timestamp|duration_seconds|speed_bytes_per_sec|status_code|total_bytes|operation|array_size
	parts := strings.Split(line, "|")
	if len(parts) < 7 {
		return nil
	}

	check := &ParityCheckHistoryInfo{}

	// Parse timestamp and format it nicely
	check.Date = s.formatParityCheckDate(parts[0])

	// Parse duration from seconds
	if durationSec, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
		check.Duration = s.formatDuration(durationSec)
	} else {
		check.Duration = "Unknown"
	}

	// Parse speed from bytes per second
	if speedBps, err := strconv.ParseInt(parts[2], 10, 64); err == nil {
		check.Speed = s.formatSpeed(speedBps)
	} else {
		check.Speed = "Unknown"
	}

	// Parse status code
	if statusCode, err := strconv.Atoi(parts[3]); err == nil {
		check.Status = s.formatParityCheckStatus(statusCode, parts[4])
	} else {
		check.Status = "Unknown"
	}

	return check
}

// formatParityCheckDate formats the timestamp into a readable format
func (s *SystemMonitor) formatParityCheckDate(timestamp string) string {
	// Parse "2025 Jun 16 15:38:06" format
	t, err := time.Parse("2006 Jan 2 15:04:05", timestamp)
	if err != nil {
		return timestamp // Return original if parsing fails
	}

	// Format as "2025-06-16, 15:38:06 (Monday)"
	return t.Format("2006-01-02, 15:04:05 (Monday)")
}

// formatDuration converts seconds to human-readable duration
func (s *SystemMonitor) formatDuration(seconds int64) string {
	if seconds <= 0 {
		return "0 sec"
	}

	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	var parts []string

	if days > 0 {
		if days == 1 {
			parts = append(parts, "1 day")
		} else {
			parts = append(parts, fmt.Sprintf("%d day", days))
		}
	}

	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d hr", hours))
	}

	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%d min", minutes))
	}

	if secs > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%d sec", secs))
	}

	return strings.Join(parts, ", ")
}

// formatSpeed converts bytes per second to MB/s
func (s *SystemMonitor) formatSpeed(bytesPerSec int64) string {
	if bytesPerSec <= 0 {
		return "0.0 MB/s"
	}

	mbPerSec := float64(bytesPerSec) / (1024 * 1024)
	return fmt.Sprintf("%.1f MB/s", mbPerSec)
}

// formatParityCheckStatus converts status code to descriptive text
func (s *SystemMonitor) formatParityCheckStatus(statusCode int, totalBytesStr string) string {
	switch statusCode {
	case 0:
		return "OK"
	case -4:
		return "Canceled"
	default:
		if statusCode > 0 {
			// Parse total bytes to determine if this was an error count
			if totalBytes, err := strconv.ParseInt(totalBytesStr, 10, 64); err == nil && totalBytes > 0 {
				return fmt.Sprintf("Failed (%d errors)", statusCode)
			}
			return "Canceled"
		}
		return "Unknown"
	}
}

// getNextScheduledParityCheck gets the next scheduled parity check
func (s *SystemMonitor) getNextScheduledParityCheck() string {
	// Try to read from Unraid's parity check cron configuration
	if nextCheck := s.parseUnraidParityCron(); nextCheck != "" {
		return nextCheck
	}

	// Fallback: Check system crontab for parity check schedule
	output := lib.GetCmdOutput("crontab", "-l")
	for _, line := range output {
		if strings.Contains(line, "parity") && (strings.Contains(line, "check") || strings.Contains(line, "mdcmd")) {
			if nextCheck := s.parseNextCronRun(line); nextCheck != "" {
				return nextCheck
			}
		}
	}

	// Final fallback
	return "Not scheduled"
}

// parseUnraidParityCron parses Unraid's parity check cron configuration
func (s *SystemMonitor) parseUnraidParityCron() string {
	// Read Unraid's parity check cron file
	content, err := os.ReadFile("/boot/config/plugins/dynamix/parity-check.cron")
	if err != nil {
		return ""
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Look for parity check commands (mdcmd check)
		if strings.Contains(line, "mdcmd check") {
			return s.parseNextCronRun(line)
		}
	}

	return ""
}

// parseNextCronRun parses cron schedule and calculates next run time
func (s *SystemMonitor) parseNextCronRun(cronLine string) string {
	// Extract cron expression from the line
	// Format: "minute hour day month weekday command"
	fields := strings.Fields(cronLine)
	if len(fields) < 5 {
		return ""
	}

	// Parse cron fields
	minute := fields[0]
	hour := fields[1]
	day := fields[2]
	month := fields[3]
	weekday := fields[4]

	// Get current time
	now := time.Now()

	// Calculate next execution time
	nextTime := s.calculateNextCronTime(now, minute, hour, day, month, weekday)
	if nextTime.IsZero() {
		return "Invalid schedule"
	}

	// Format human-readable output
	return s.formatNextCheckTime(nextTime, now)
}

// calculateNextCronTime calculates the next execution time for a cron expression
func (s *SystemMonitor) calculateNextCronTime(now time.Time, minute, hour, day, month, weekday string) time.Time {
	// Parse minute
	targetMinute, err := s.parseCronField(minute, 0, 59)
	if err != nil {
		return time.Time{}
	}

	// Parse hour
	targetHour, err := s.parseCronField(hour, 0, 23)
	if err != nil {
		return time.Time{}
	}

	// Parse day
	targetDay, err := s.parseCronField(day, 1, 31)
	if err != nil {
		return time.Time{}
	}

	// Parse month
	targetMonth, err := s.parseCronField(month, 1, 12)
	if err != nil {
		return time.Time{}
	}

	// Parse weekday (0 = Sunday, 6 = Saturday)
	targetWeekday, err := s.parseCronField(weekday, 0, 6)
	if err != nil {
		return time.Time{}
	}

	// Start from next minute to avoid immediate execution
	next := now.Add(time.Minute).Truncate(time.Minute)

	// Find next matching time (limit search to avoid infinite loops)
	for i := 0; i < 366*24*60; i++ { // Search up to 1 year
		if s.cronTimeMatches(next, targetMinute, targetHour, targetDay, targetMonth, targetWeekday) {
			return next
		}
		next = next.Add(time.Minute)
	}

	return time.Time{} // Not found within reasonable time
}

// parseCronField parses a single cron field (supports *, numbers, and basic ranges)
func (s *SystemMonitor) parseCronField(field string, min, max int) (int, error) {
	if field == "*" {
		return -1, nil // Wildcard
	}

	// Handle simple numbers
	if value, err := strconv.Atoi(field); err == nil {
		if value >= min && value <= max {
			return value, nil
		}
		return 0, fmt.Errorf("value %d out of range [%d-%d]", value, min, max)
	}

	// For now, return error for complex expressions (ranges, lists, steps)
	return 0, fmt.Errorf("complex cron expressions not supported: %s", field)
}

// cronTimeMatches checks if a time matches the cron expression
func (s *SystemMonitor) cronTimeMatches(t time.Time, minute, hour, day, month, weekday int) bool {
	// Check minute
	if minute != -1 && t.Minute() != minute {
		return false
	}

	// Check hour
	if hour != -1 && t.Hour() != hour {
		return false
	}

	// Check month
	if month != -1 && int(t.Month()) != month {
		return false
	}

	// Check day and weekday (both conditions must be satisfied if both are specified)
	dayMatches := (day == -1 || t.Day() == day)
	weekdayMatches := (weekday == -1 || int(t.Weekday()) == weekday)

	// If both day and weekday are specified, either can match (OR logic)
	if day != -1 && weekday != -1 {
		return dayMatches || weekdayMatches
	}

	// If only one is specified, it must match
	return dayMatches && weekdayMatches
}

// formatNextCheckTime formats the next check time in human-readable format
func (s *SystemMonitor) formatNextCheckTime(nextTime, now time.Time) string {
	duration := nextTime.Sub(now)

	// If it's within the next 7 days, show relative time
	if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		hours := int(duration.Hours()) % 24

		if days == 0 {
			if hours == 0 {
				return fmt.Sprintf("Today at %s", nextTime.Format("15:04"))
			}
			return fmt.Sprintf("Tomorrow at %s", nextTime.Format("15:04"))
		} else if days == 1 {
			return fmt.Sprintf("Tomorrow at %s", nextTime.Format("15:04"))
		} else {
			weekday := nextTime.Format("Monday")
			return fmt.Sprintf("Next %s at %s", weekday, nextTime.Format("15:04"))
		}
	}

	// For longer periods, show absolute date
	if nextTime.Year() == now.Year() {
		return nextTime.Format("Jan 2 at 15:04")
	}

	return nextTime.Format("Jan 2, 2006 at 15:04")
}

// GetParityDiskInfo returns parity disk information
func (s *SystemMonitor) GetParityDiskInfo() (*ParityDiskInfo, error) {
	status, err := s.GetMDCmdStatus()
	if err != nil {
		return nil, err
	}

	if status.ParityDisk == nil {
		return nil, fmt.Errorf("no parity disk found")
	}

	return status.ParityDisk, nil
}

// GetParityCheckInfo returns parity check status information
func (s *SystemMonitor) GetParityCheckInfo() (*ParityCheckInfo, error) {
	status, err := s.GetMDCmdStatus()
	if err != nil {
		return nil, err
	}

	if status.ParityCheck == nil {
		return nil, fmt.Errorf("no parity check information available")
	}

	return status.ParityCheck, nil
}
