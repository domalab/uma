package storage

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/domalab/uma/daemon/lib"
	"github.com/domalab/uma/daemon/logger"
)

// StorageMonitor provides storage monitoring capabilities
type StorageMonitor struct {
	arrayDisks []DiskInfo
	cacheDisks []DiskInfo
	bootDisk   DiskInfo
}

// DiskInfo represents information about a disk
type DiskInfo struct {
	Device      string  `json:"device"`
	Name        string  `json:"name"`
	Size        uint64  `json:"size"`
	Used        uint64  `json:"used"`
	Available   uint64  `json:"available"`
	UsedPercent float64 `json:"used_percent"`
	FileSystem  string  `json:"filesystem"`
	MountPoint  string  `json:"mount_point"`
	Status      string  `json:"status"`
	Temperature int     `json:"temperature,omitempty"`
	Health      string  `json:"health"`
	// Enhanced disk information
	PowerState    string `json:"power_state"` // active, standby, sleeping, unknown
	DiskType      string `json:"disk_type"`   // HDD, SSD, NVMe
	Interface     string `json:"interface"`   // SATA, NVMe, USB
	Model         string `json:"model,omitempty"`
	SerialNumber  string `json:"serial_number,omitempty"`
	SpinDownDelay int    `json:"spin_down_delay,omitempty"` // minutes, -1 for disabled
}

// ArrayInfo represents Unraid array information
type ArrayInfo struct {
	State       string     `json:"state"`
	NumDevices  int        `json:"num_devices"`
	NumDisks    int        `json:"num_disks"`
	NumParity   int        `json:"num_parity"`
	TotalSize   uint64     `json:"total_size"`
	UsedSize    uint64     `json:"used_size"`
	FreeSize    uint64     `json:"free_size"`
	UsedPercent float64    `json:"used_percent"`
	Disks       []DiskInfo `json:"disks"`
	// Enhanced array information
	SpinDownDelay int `json:"spin_down_delay"` // Default spin-down delay in minutes
}

// ParityCheckStatus represents the status of a parity check operation
type ParityCheckStatus struct {
	Active        bool    `json:"active"`
	Type          string  `json:"type,omitempty"`           // "check" or "correct"
	Progress      float64 `json:"progress,omitempty"`       // 0-100
	Speed         string  `json:"speed,omitempty"`          // e.g., "45.2 MB/s"
	TimeRemaining string  `json:"time_remaining,omitempty"` // e.g., "2h 15m"
	Errors        int     `json:"errors,omitempty"`
}

// CacheInfo represents cache pool information
type CacheInfo struct {
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	FileSystem  string     `json:"filesystem"`
	TotalSize   uint64     `json:"total_size"`
	UsedSize    uint64     `json:"used_size"`
	FreeSize    uint64     `json:"free_size"`
	UsedPercent float64    `json:"used_percent"`
	Disks       []DiskInfo `json:"disks"`
}

// NewStorageMonitor creates a new storage monitor
func NewStorageMonitor() *StorageMonitor {
	return &StorageMonitor{}
}

// GetArrayInfo returns information about the Unraid array
func (s *StorageMonitor) GetArrayInfo() (*ArrayInfo, error) {
	arrayInfo := &ArrayInfo{
		State: s.getArrayState(),
		Disks: make([]DiskInfo, 0),
	}

	// Get Unraid spin-down configuration
	arrayInfo.SpinDownDelay = s.getSpinDownDelay()

	// Read array configuration
	if err := s.loadArrayDisks(arrayInfo); err != nil {
		logger.Yellow("Failed to load array disk information: %v", err)
	}

	// Calculate totals
	s.calculateArrayTotals(arrayInfo)

	return arrayInfo, nil
}

// GetCacheInfo returns information about cache pools
func (s *StorageMonitor) GetCacheInfo() ([]CacheInfo, error) {
	caches := make([]CacheInfo, 0)

	// Check for cache pools
	cachePools, err := s.findCachePools()
	if err != nil {
		return caches, err
	}

	for _, pool := range cachePools {
		cacheInfo := CacheInfo{
			Name:   pool,
			Status: "online",
			Disks:  make([]DiskInfo, 0),
		}

		if err := s.loadCacheDisks(&cacheInfo); err != nil {
			logger.Yellow("Failed to load cache disk information for %s: %v", pool, err)
			continue
		}

		s.calculateCacheTotals(&cacheInfo)
		caches = append(caches, cacheInfo)
	}

	return caches, nil
}

// GetBootDiskInfo returns information about the boot disk
func (s *StorageMonitor) GetBootDiskInfo() (*DiskInfo, error) {
	bootDisk := &DiskInfo{
		Name:       "boot",
		MountPoint: "/boot",
		Status:     "online",
	}

	// Get boot disk usage
	if err := s.getDiskUsage(bootDisk); err != nil {
		return bootDisk, err
	}

	// Get boot disk device information
	if err := s.getBootDiskDevice(bootDisk); err != nil {
		logger.Yellow("Failed to get boot disk device info: %v", err)
	}

	return bootDisk, nil
}

// getArrayState reads the array state from Unraid
func (s *StorageMonitor) getArrayState() string {
	// Check Unraid's custom /proc/mdstat format first
	if exists, _ := lib.Exists("/proc/mdstat"); exists {
		content, err := os.ReadFile("/proc/mdstat")
		if err != nil {
			return "unknown"
		}

		// Check if this is Unraid format (contains key=value pairs)
		if strings.Contains(string(content), "mdState=") {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "mdState=") {
					state := strings.TrimPrefix(line, "mdState=")
					switch state {
					case "STARTED":
						return "started"
					case "STOPPED":
						return "stopped"
					case "INVALID":
						return "invalid"
					default:
						return strings.ToLower(state)
					}
				}
			}
		} else {
			// Fall back to standard Linux mdstat format
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				if strings.Contains(line, "md") && strings.Contains(line, "active") {
					return "started"
				}
			}
		}
	}

	// Check array state file
	if exists, _ := lib.Exists("/var/local/emhttp/array_state"); exists {
		content, err := os.ReadFile("/var/local/emhttp/array_state")
		if err != nil {
			return "unknown"
		}
		return strings.TrimSpace(string(content))
	}

	return "stopped"
}

// loadArrayDisks loads information about array disks
func (s *StorageMonitor) loadArrayDisks(arrayInfo *ArrayInfo) error {
	// Try to read from Unraid's mdstat format first
	if err := s.loadArrayDisksFromMdstat(arrayInfo); err != nil {
		logger.Yellow("Failed to load from mdstat, trying disk assignments: %v", err)

		// Fallback to reading disk assignments from Unraid configuration
		diskAssignments, err := s.readDiskAssignments()
		if err != nil {
			return err
		}

		for device, assignment := range diskAssignments {
			diskInfo := DiskInfo{
				Device: device,
				Name:   assignment.Name,
				Status: assignment.Status,
			}

			// Get disk usage if mounted
			if assignment.MountPoint != "" {
				diskInfo.MountPoint = assignment.MountPoint
				s.getDiskUsage(&diskInfo)
			}

			// Get disk health and temperature
			s.getDiskHealth(&diskInfo)
			s.getDiskTemperature(&diskInfo)

			arrayInfo.Disks = append(arrayInfo.Disks, diskInfo)

			// Count disk types
			if strings.HasPrefix(assignment.Name, "parity") {
				arrayInfo.NumParity++
			} else if strings.HasPrefix(assignment.Name, "disk") {
				arrayInfo.NumDisks++
			}
		}
	}

	arrayInfo.NumDevices = len(arrayInfo.Disks)
	return nil
}

// loadArrayDisksFromMdstat loads disk information from Unraid's mdstat format
func (s *StorageMonitor) loadArrayDisksFromMdstat(arrayInfo *ArrayInfo) error {
	content, err := os.ReadFile("/proc/mdstat")
	if err != nil {
		return err
	}

	// Parse Unraid's key=value format
	mdstatData := make(map[string]string)
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				mdstatData[key] = value
			}
		}
	}

	logger.Blue("Parsed mdstat data: found %d entries", len(mdstatData))

	// Extract array information
	if numDisks, ok := mdstatData["mdNumDisks"]; ok {
		if n, err := strconv.Atoi(numDisks); err == nil {
			arrayInfo.NumDisks = n
		}
	}

	// Process each disk
	for i := 0; i < 30; i++ { // Unraid supports up to 30 disks
		diskPrefix := fmt.Sprintf("diskNumber.%d", i)

		// Check if this disk exists
		if _, exists := mdstatData[diskPrefix]; !exists {
			continue
		}

		diskInfo := DiskInfo{}

		// Get disk information
		if diskName, ok := mdstatData[fmt.Sprintf("diskName.%d", i)]; ok && diskName != "" {
			diskInfo.Name = diskName
		} else {
			diskInfo.Name = fmt.Sprintf("disk%d", i)
		}

		if diskSize, ok := mdstatData[fmt.Sprintf("diskSize.%d", i)]; ok {
			if size, err := strconv.ParseUint(diskSize, 10, 64); err == nil {
				diskInfo.Size = size * 1024 // Convert from KB to bytes
			}
		}

		if diskId, ok := mdstatData[fmt.Sprintf("diskId.%d", i)]; ok {
			diskInfo.Device = "/dev/disk/by-id/" + diskId
		}

		// Get device information
		if rdevName, ok := mdstatData[fmt.Sprintf("rdevName.%d", i)]; ok && rdevName != "" {
			diskInfo.Device = "/dev/" + rdevName
		}

		if rdevStatus, ok := mdstatData[fmt.Sprintf("rdevStatus.%d", i)]; ok {
			switch rdevStatus {
			case "DISK_OK":
				diskInfo.Status = "online"
				diskInfo.Health = "healthy"
			case "DISK_NP":
				diskInfo.Status = "not_present"
				diskInfo.Health = "unknown"
			case "DISK_NP_DSBL":
				diskInfo.Status = "disabled"
				diskInfo.Health = "unknown"
			default:
				diskInfo.Status = "unknown"
				diskInfo.Health = "unknown"
			}
		}

		// Set mount point for data disks
		if strings.HasPrefix(diskInfo.Name, "md") && diskInfo.Status == "online" {
			diskNum := strings.TrimPrefix(diskInfo.Name, "md")
			diskNum = strings.TrimSuffix(diskNum, "p1")
			diskInfo.MountPoint = "/mnt/disk" + diskNum

			// Get disk usage if mounted
			s.getDiskUsage(&diskInfo)
		}

		// Get enhanced disk information
		if diskInfo.Status == "online" {
			// Use proper device path resolution like HA integration
			if diskInfo.Device == "" {
				diskInfo.Device = s.getDevicePathFromMount(diskInfo.Name)
			}

			if diskInfo.Device != "" {
				s.getDiskTemperature(&diskInfo)
				s.getDiskHealth(&diskInfo)
				s.getDiskPowerState(&diskInfo)
				s.getDiskTypeAndInterface(&diskInfo)
				s.getDiskModel(&diskInfo)
				s.getDiskSpinDownDelay(&diskInfo, i)
			}
		}

		// Only add disks that actually exist
		if diskInfo.Status != "not_present" {
			arrayInfo.Disks = append(arrayInfo.Disks, diskInfo)

			// Count disk types
			if i == 0 {
				arrayInfo.NumParity++
			} else {
				arrayInfo.NumDisks++
			}
		}
	}

	logger.Blue("Loaded %d disks from mdstat", len(arrayInfo.Disks))
	return nil
}

// DiskAssignment represents a disk assignment in Unraid
type DiskAssignment struct {
	Name       string
	Status     string
	MountPoint string
}

// readDiskAssignments reads disk assignments from Unraid configuration
func (s *StorageMonitor) readDiskAssignments() (map[string]DiskAssignment, error) {
	assignments := make(map[string]DiskAssignment)

	// Read from /var/local/emhttp/disks.ini if it exists
	diskIniPath := "/var/local/emhttp/disks.ini"
	if exists, _ := lib.Exists(diskIniPath); exists {
		if err := s.parseDisksIni(diskIniPath, assignments); err != nil {
			// If parsing fails, fall back to scanning mounted disks
			s.scanMountedDisks(assignments)
		}
	} else {
		// Scan mounted disks as fallback
		s.scanMountedDisks(assignments)
	}

	return assignments, nil
}

// parseDisksIni parses Unraid's disks.ini file
func (s *StorageMonitor) parseDisksIni(filePath string, assignments map[string]DiskAssignment) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentSection string
	var currentDisk DiskAssignment

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for section headers like ["parity"] or ["disk1"]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			// Save previous disk if we have one
			if currentSection != "" && currentDisk.Name != "" {
				device := "/dev/" + currentDisk.Name
				if currentDisk.Name != "" {
					assignments[device] = currentDisk
				}
			}

			// Start new section
			currentSection = strings.Trim(line, "[]\"")
			currentDisk = DiskAssignment{
				Name:   currentSection,
				Status: "unknown",
			}
			continue
		}

		// Parse key=value pairs
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), "\"")

			switch key {
			case "device":
				currentDisk.Name = value
			case "status":
				switch value {
				case "DISK_OK":
					currentDisk.Status = "online"
				case "DISK_NP":
					currentDisk.Status = "not_present"
				case "DISK_DSBL":
					currentDisk.Status = "disabled"
				default:
					currentDisk.Status = "unknown"
				}
			case "name":
				if value != "" {
					currentDisk.Name = value
				}
			}
		}
	}

	// Save the last disk
	if currentSection != "" && currentDisk.Name != "" {
		device := "/dev/" + currentDisk.Name
		assignments[device] = currentDisk

		// Set mount point based on disk type
		if strings.HasPrefix(currentSection, "disk") {
			currentDisk.MountPoint = "/mnt/" + currentSection
		} else if currentSection == "cache" {
			currentDisk.MountPoint = "/mnt/cache"
		}
		assignments[device] = currentDisk
	}

	return scanner.Err()
}

// scanMountedDisks scans for mounted Unraid disks
func (s *StorageMonitor) scanMountedDisks(assignments map[string]DiskAssignment) {
	// Scan /mnt/disk* for array disks
	diskDirs, _ := filepath.Glob("/mnt/disk*")
	for _, dir := range diskDirs {
		if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
			diskName := filepath.Base(dir)
			device := s.findDeviceForMountPoint(dir)
			if device != "" {
				assignments[device] = DiskAssignment{
					Name:       diskName,
					Status:     "online",
					MountPoint: dir,
				}
			}
		}
	}

	// Scan /mnt/cache* for cache disks
	cacheDirs, _ := filepath.Glob("/mnt/cache*")
	for _, dir := range cacheDirs {
		if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
			cacheName := filepath.Base(dir)
			device := s.findDeviceForMountPoint(dir)
			if device != "" {
				assignments[device] = DiskAssignment{
					Name:       cacheName,
					Status:     "online",
					MountPoint: dir,
				}
			}
		}
	}
}

// findDeviceForMountPoint finds the device for a given mount point
func (s *StorageMonitor) findDeviceForMountPoint(mountPoint string) string {
	// Read /proc/mounts to find the device
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && fields[1] == mountPoint {
			return fields[0]
		}
	}

	return ""
}

// getDiskUsage gets disk usage statistics
func (s *StorageMonitor) getDiskUsage(disk *DiskInfo) error {
	if disk.MountPoint == "" {
		return fmt.Errorf("no mount point specified")
	}

	// Use df command to get disk usage
	output := lib.GetCmdOutput("df", "-B1", disk.MountPoint)
	if len(output) < 2 {
		return fmt.Errorf("failed to get disk usage")
	}

	// Parse df output
	fields := strings.Fields(output[1])
	if len(fields) < 4 {
		return fmt.Errorf("invalid df output")
	}

	var err error
	if disk.Size, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
		return err
	}
	if disk.Used, err = strconv.ParseUint(fields[2], 10, 64); err != nil {
		return err
	}
	if disk.Available, err = strconv.ParseUint(fields[3], 10, 64); err != nil {
		return err
	}

	if disk.Size > 0 {
		disk.UsedPercent = float64(disk.Used) / float64(disk.Size) * 100
	}

	return nil
}

// getDiskTemperature gets disk temperature using smartctl
func (s *StorageMonitor) getDiskTemperature(disk *DiskInfo) {
	if disk.Device == "" {
		return
	}

	// Use smartctl to get disk temperature
	output := lib.GetCmdOutput("smartctl", "-A", disk.Device)
	for _, line := range output {
		if strings.Contains(line, "Temperature_Celsius") || strings.Contains(line, "Airflow_Temperature_Cel") {
			fields := strings.Fields(line)
			if len(fields) >= 10 {
				if temp, err := strconv.Atoi(fields[9]); err == nil {
					disk.Temperature = temp
					break
				}
			}
		}
	}
}

// getBootDiskDevice gets boot disk device information
func (s *StorageMonitor) getBootDiskDevice(disk *DiskInfo) error {
	// Find boot device from /proc/mounts
	disk.Device = s.findDeviceForMountPoint("/boot")
	if disk.Device == "" {
		return fmt.Errorf("boot device not found")
	}

	// Get filesystem type
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 3 && fields[1] == "/boot" {
			disk.FileSystem = fields[2]
			break
		}
	}

	return nil
}

// findCachePools finds available cache pools
func (s *StorageMonitor) findCachePools() ([]string, error) {
	pools := make([]string, 0)

	// Look for cache directories
	cacheDirs, err := filepath.Glob("/mnt/cache*")
	if err != nil {
		return pools, err
	}

	for _, dir := range cacheDirs {
		if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
			poolName := filepath.Base(dir)
			pools = append(pools, poolName)
		}
	}

	return pools, nil
}

// loadCacheDisks loads information about cache pool disks
func (s *StorageMonitor) loadCacheDisks(cache *CacheInfo) error {
	mountPoint := "/mnt/" + cache.Name

	// Get cache pool usage
	if err := s.getCacheUsage(cache, mountPoint); err != nil {
		return err
	}

	// Find devices in the cache pool
	devices := s.findCacheDevices(cache.Name)
	for _, device := range devices {
		diskInfo := DiskInfo{
			Device:     device,
			Name:       cache.Name,
			MountPoint: mountPoint,
			Status:     "online",
		}

		s.getDiskHealth(&diskInfo)
		s.getDiskTemperature(&diskInfo)
		cache.Disks = append(cache.Disks, diskInfo)
	}

	return nil
}

// getCacheUsage gets cache pool usage statistics
func (s *StorageMonitor) getCacheUsage(cache *CacheInfo, mountPoint string) error {
	// Use df command to get cache usage
	output := lib.GetCmdOutput("df", "-B1", mountPoint)
	if len(output) < 2 {
		return fmt.Errorf("failed to get cache usage")
	}

	// Parse df output
	fields := strings.Fields(output[1])
	if len(fields) < 4 {
		return fmt.Errorf("invalid df output")
	}

	var err error
	if cache.TotalSize, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
		return err
	}
	if cache.UsedSize, err = strconv.ParseUint(fields[2], 10, 64); err != nil {
		return err
	}
	if cache.FreeSize, err = strconv.ParseUint(fields[3], 10, 64); err != nil {
		return err
	}

	if cache.TotalSize > 0 {
		cache.UsedPercent = float64(cache.UsedSize) / float64(cache.TotalSize) * 100
	}

	// Get filesystem type
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 3 && fields[1] == mountPoint {
			cache.FileSystem = fields[2]
			break
		}
	}

	return nil
}

// findCacheDevices finds devices that belong to a cache pool
func (s *StorageMonitor) findCacheDevices(poolName string) []string {
	devices := make([]string, 0)

	// This is a simplified implementation
	// In reality, you'd need to parse Unraid's configuration to find
	// which devices belong to which cache pool

	// For now, we'll assume cache pools use /dev/nvme* or /dev/sd* devices
	// and try to match them based on mount information
	mountPoint := "/mnt/" + poolName
	device := s.findDeviceForMountPoint(mountPoint)
	if device != "" {
		devices = append(devices, device)
	}

	return devices
}

// getSpinDownDelay reads the default spin-down delay from Unraid configuration
func (s *StorageMonitor) getSpinDownDelay() int {
	// Read from /boot/config/disk.cfg
	content, err := os.ReadFile("/boot/config/disk.cfg")
	if err != nil {
		logger.Yellow("Failed to read disk.cfg: %v", err)
		return 30 // Default to 30 minutes
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "spindownDelay=") {
			delayStr := strings.TrimPrefix(line, "spindownDelay=")
			delayStr = strings.Trim(delayStr, "\"")
			if delay, err := strconv.Atoi(delayStr); err == nil {
				return delay
			}
		}
	}

	return 30 // Default to 30 minutes
}

// getDiskPowerState gets the current power state of a disk using HA integration logic
func (s *StorageMonitor) getDiskPowerState(disk *DiskInfo) {
	if disk.Device == "" {
		disk.PowerState = "unknown"
		return
	}

	// Convert device path to actual device
	actualDevice := s.resolveDevicePath(disk.Device)
	if actualDevice == "" {
		disk.PowerState = "unknown"
		return
	}

	// Skip NVMe devices - they're always active
	if strings.Contains(strings.ToLower(actualDevice), "nvme") {
		disk.PowerState = "active"
		return
	}

	// Method 1: Use SMART with standby detection (primary method from HA integration)
	// smartctl -n standby returns:
	// - Exit code 0: Device is active
	// - Exit code 2: Device is in standby
	// - Other codes: Error or unknown state
	output := lib.GetCmdOutput("smartctl", "-n", "standby", actualDevice)

	// Check the exit code by running the command and capturing both output and exit status
	cmd := exec.Command("smartctl", "-n", "standby", actualDevice)
	err := cmd.Run()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode := exitError.ExitCode()
			if exitCode == 2 {
				// Exit code 2 means device is in standby
				disk.PowerState = "standby"
				return
			}
			// Other non-zero exit codes fall through to hdparm fallback
		}
	} else {
		// Exit code 0 means device is active
		disk.PowerState = "active"
		return
	}

	// Method 2: Fall back to hdparm if SMART fails
	output = lib.GetCmdOutput("hdparm", "-C", actualDevice)
	for _, line := range output {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "drive state is:") {
			if strings.Contains(line, "active/idle") {
				disk.PowerState = "active"
			} else if strings.Contains(line, "standby") {
				disk.PowerState = "standby"
			} else if strings.Contains(line, "sleeping") {
				disk.PowerState = "sleeping"
			} else {
				disk.PowerState = "unknown"
			}
			return
		}
	}

	// Default to unknown if all methods fail
	disk.PowerState = "unknown"
}

// getDiskHealth gets the SMART health status of a disk
func (s *StorageMonitor) getDiskHealth(disk *DiskInfo) {
	if disk.Device == "" {
		disk.Health = "unknown"
		return
	}

	actualDevice := s.resolveDevicePath(disk.Device)
	if actualDevice == "" {
		disk.Health = "unknown"
		return
	}

	// Use smartctl to check health
	output := lib.GetCmdOutput("smartctl", "-H", actualDevice)
	for _, line := range output {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "SMART overall-health self-assessment test result:") {
			if strings.Contains(line, "PASSED") {
				disk.Health = "healthy"
			} else {
				disk.Health = "failing"
			}
			return
		}
		// For NVMe drives
		if strings.Contains(line, "SMART Health Status:") {
			if strings.Contains(line, "OK") {
				disk.Health = "healthy"
			} else {
				disk.Health = "failing"
			}
			return
		}
	}

	disk.Health = "unknown"
}

// getDiskTypeAndInterface determines disk type (HDD/SSD/NVMe) and interface
func (s *StorageMonitor) getDiskTypeAndInterface(disk *DiskInfo) {
	if disk.Device == "" {
		disk.DiskType = "unknown"
		disk.Interface = "unknown"
		return
	}

	actualDevice := s.resolveDevicePath(disk.Device)
	if actualDevice == "" {
		disk.DiskType = "unknown"
		disk.Interface = "unknown"
		return
	}

	// Check if it's NVMe
	if strings.Contains(actualDevice, "nvme") {
		disk.DiskType = "NVMe"
		disk.Interface = "NVMe"
		return
	}

	// Use smartctl to determine disk type
	output := lib.GetCmdOutput("smartctl", "-i", actualDevice)
	for _, line := range output {
		line = strings.ToLower(strings.TrimSpace(line))

		// Check for SSD indicators
		if strings.Contains(line, "solid state") ||
			strings.Contains(line, "ssd") ||
			strings.Contains(line, "flash") {
			disk.DiskType = "SSD"
		}

		// Check for rotation rate (HDD indicator)
		if strings.Contains(line, "rotation rate") &&
			!strings.Contains(line, "solid state device") {
			disk.DiskType = "HDD"
		}

		// Determine interface
		if strings.Contains(line, "sata") {
			disk.Interface = "SATA"
		} else if strings.Contains(line, "usb") {
			disk.Interface = "USB"
		}
	}

	// Default values if not detected
	if disk.DiskType == "" {
		if strings.Contains(actualDevice, "sd") {
			disk.DiskType = "HDD" // Assume SATA drives are HDDs unless proven otherwise
		} else {
			disk.DiskType = "unknown"
		}
	}

	if disk.Interface == "" {
		if strings.Contains(actualDevice, "sd") {
			disk.Interface = "SATA"
		} else {
			disk.Interface = "unknown"
		}
	}
}

// getDiskModel gets disk model and serial number
func (s *StorageMonitor) getDiskModel(disk *DiskInfo) {
	if disk.Device == "" {
		return
	}

	actualDevice := s.resolveDevicePath(disk.Device)
	if actualDevice == "" {
		return
	}

	// Use smartctl to get disk model and serial
	output := lib.GetCmdOutput("smartctl", "-i", actualDevice)
	for _, line := range output {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Device Model:") || strings.HasPrefix(line, "Model Number:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				disk.Model = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "Serial Number:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				disk.SerialNumber = strings.TrimSpace(parts[1])
			}
		}
	}
}

// getDiskSpinDownDelay gets the effective disk spin-down delay (always resolved, never -1)
func (s *StorageMonitor) getDiskSpinDownDelay(disk *DiskInfo, diskIndex int) {
	// Get the global default first
	globalDefault := s.getSpinDownDelay()

	// Read from /boot/config/disk.cfg
	content, err := os.ReadFile("/boot/config/disk.cfg")
	if err != nil {
		disk.SpinDownDelay = globalDefault
		return
	}

	lines := strings.Split(string(content), "\n")
	diskKey := fmt.Sprintf("diskSpindownDelay.%d=", diskIndex)

	// Look for individual disk setting
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, diskKey) {
			delayStr := strings.TrimPrefix(line, diskKey)
			delayStr = strings.Trim(delayStr, "\"")
			if delay, err := strconv.Atoi(delayStr); err == nil {
				if delay == -1 {
					// -1 means use global default
					disk.SpinDownDelay = globalDefault
				} else if delay == 0 {
					// 0 means never spin down
					disk.SpinDownDelay = 0
				} else {
					// Use the specific value for this disk
					disk.SpinDownDelay = delay
				}
				return
			}
		}
	}

	// No individual setting found, use global default
	disk.SpinDownDelay = globalDefault
}

// resolveDevicePath converts device paths to actual device nodes using HA integration logic
func (s *StorageMonitor) resolveDevicePath(devicePath string) string {
	// If it's already a direct device path, return it
	if strings.HasPrefix(devicePath, "/dev/sd") || strings.HasPrefix(devicePath, "/dev/nvme") {
		return devicePath
	}

	// If it's a by-id path, resolve it
	if strings.HasPrefix(devicePath, "/dev/disk/by-id/") {
		if target, err := os.Readlink(devicePath); err == nil {
			// Resolve relative path
			if !strings.HasPrefix(target, "/") {
				target = filepath.Join(filepath.Dir(devicePath), target)
			}
			return filepath.Clean(target)
		}
	}

	// Try to find the device by scanning /dev/disk/by-id/
	if !strings.HasPrefix(devicePath, "/dev/") {
		// Assume it's just the device ID
		byIdPath := "/dev/disk/by-id/" + devicePath
		if target, err := os.Readlink(byIdPath); err == nil {
			if !strings.HasPrefix(target, "/") {
				target = filepath.Join("/dev/disk/by-id", target)
			}
			return filepath.Clean(target)
		}
	}

	return devicePath
}

// getDevicePathFromMount gets the actual device path from mount point (HA integration method)
func (s *StorageMonitor) getDevicePathFromMount(diskName string) string {
	// Use findmnt to get the source device for the mount point
	output := lib.GetCmdOutput("findmnt", "-n", "-o", "SOURCE", "/mnt/"+diskName)
	if len(output) == 0 {
		return ""
	}

	devicePath := strings.TrimSpace(output[0])
	if devicePath == "" {
		return ""
	}

	// If it's an MD device with partition (e.g., /dev/md1p1), resolve to physical device
	if strings.HasPrefix(devicePath, "/dev/md") && strings.Contains(devicePath, "p1") {
		// Extract MD number (e.g., md1 from /dev/md1p1)
		mdDevice := strings.Split(devicePath, "p1")[0] // /dev/md1
		physicalDevice := s.resolveMDToPhysical(mdDevice)
		if physicalDevice != "" {
			return physicalDevice
		}
	}

	return devicePath
}

// resolveMDToPhysical resolves MD device to underlying physical device using mdcmd status
func (s *StorageMonitor) resolveMDToPhysical(mdDevice string) string {
	// Extract MD number (e.g., md1 from /dev/md1)
	mdNum := strings.TrimPrefix(mdDevice, "/dev/md")

	// Use mdcmd status to get the mapping
	output := lib.GetCmdOutput("mdcmd", "status")
	if len(output) == 0 {
		return ""
	}

	// Parse the output to find the physical device for this MD device
	diskNumber := ""
	for _, line := range output {
		line = strings.TrimSpace(line)
		// Look for diskName.X=mdYp1 where Y matches our MD number
		if strings.HasPrefix(line, "diskName.") && strings.HasSuffix(line, "=md"+mdNum+"p1") {
			// Extract the disk number (X from diskName.X)
			parts := strings.Split(line, ".")
			if len(parts) >= 2 {
				diskNumber = strings.Split(parts[1], "=")[0]
				break
			}
		}
	}

	if diskNumber == "" {
		return ""
	}

	// Now find the corresponding rdevName.X line
	for _, line := range output {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "rdevName."+diskNumber+"=") {
			physicalDevice := "/dev/" + strings.Split(line, "=")[1]
			return physicalDevice
		}
	}

	return ""
}

// hasRecentIOActivity checks if a disk has recent I/O activity by examining /proc/diskstats
func (s *StorageMonitor) hasRecentIOActivity(devicePath string) bool {
	// Extract device name from path (e.g., /dev/sdc -> sdc)
	deviceName := strings.TrimPrefix(devicePath, "/dev/")

	// Read /proc/diskstats
	content, err := os.ReadFile("/proc/diskstats")
	if err != nil {
		return false
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 14 {
			continue
		}

		// Field 2 is the device name
		if fields[2] == deviceName {
			// Fields 3 and 7 are read and write I/O counts
			// If either is > 0, the disk has had I/O activity
			readIOs, _ := strconv.ParseUint(fields[3], 10, 64)
			writeIOs, _ := strconv.ParseUint(fields[7], 10, 64)

			// Consider disk active if it has significant I/O activity
			// This is a simple heuristic - in practice, you might want to
			// track I/O over time to detect recent activity
			return readIOs > 100 || writeIOs > 10
		}
	}

	return false
}

// isArrayDisk checks if a disk is part of the Unraid array
func (s *StorageMonitor) isArrayDisk(diskName string) bool {
	// Array disks follow the pattern: disk0, disk1, disk2, etc.
	// Parity disks are also part of the array
	return strings.HasPrefix(diskName, "disk") || diskName == "parity" || diskName == "parity2"
}

// calculateArrayTotals calculates total sizes for the array
func (s *StorageMonitor) calculateArrayTotals(arrayInfo *ArrayInfo) {
	for _, disk := range arrayInfo.Disks {
		if !strings.HasPrefix(disk.Name, "parity") {
			arrayInfo.TotalSize += disk.Size
			arrayInfo.UsedSize += disk.Used
			arrayInfo.FreeSize += disk.Available
		}
	}

	if arrayInfo.TotalSize > 0 {
		arrayInfo.UsedPercent = float64(arrayInfo.UsedSize) / float64(arrayInfo.TotalSize) * 100
	}
}

// calculateCacheTotals calculates total sizes for cache pools
func (s *StorageMonitor) calculateCacheTotals(cache *CacheInfo) {
	// Cache totals are already calculated in getCacheUsage
	// This method is here for consistency and future enhancements
}

// Array Control Operations

// StartArray starts the Unraid array
func (s *StorageMonitor) StartArray(maintenanceMode bool, checkFilesystem bool) error {
	logger.Blue("Starting Unraid array (maintenance: %v, check_fs: %v)", maintenanceMode, checkFilesystem)

	// Build mdcmd command
	cmd := "mdcmd start"
	if maintenanceMode {
		cmd += " MAINTENANCE=1"
	}
	if checkFilesystem {
		cmd += " CHECK=1"
	}

	// Execute the command
	output := lib.GetCmdOutput("sh", "-c", cmd)

	// Check for errors in output
	for _, line := range output {
		if strings.Contains(strings.ToLower(line), "error") ||
			strings.Contains(strings.ToLower(line), "failed") {
			return fmt.Errorf("array start failed: %s", line)
		}
	}

	logger.Blue("Array start command executed successfully")
	return nil
}

// StopArray stops the Unraid array
func (s *StorageMonitor) StopArray(force bool, unmountShares bool) error {
	logger.Blue("Stopping Unraid array (force: %v, unmount_shares: %v)", force, unmountShares)

	// Build mdcmd command
	cmd := "mdcmd stop"
	if force {
		cmd += " FORCE=1"
	}

	// Execute the command
	output := lib.GetCmdOutput("sh", "-c", cmd)

	// Check for errors in output
	for _, line := range output {
		if strings.Contains(strings.ToLower(line), "error") ||
			strings.Contains(strings.ToLower(line), "failed") {
			return fmt.Errorf("array stop failed: %s", line)
		}
	}

	logger.Blue("Array stop command executed successfully")
	return nil
}

// GetParityCheckStatus returns the current parity check status
func (s *StorageMonitor) GetParityCheckStatus() (*ParityCheckStatus, error) {
	status := &ParityCheckStatus{
		Active: false,
	}

	// Check /proc/mdstat for parity check information
	if exists, _ := lib.Exists("/proc/mdstat"); exists {
		content, err := os.ReadFile("/proc/mdstat")
		if err != nil {
			return status, fmt.Errorf("failed to read mdstat: %v", err)
		}

		mdstatContent := string(content)

		// Look for parity check indicators
		if strings.Contains(mdstatContent, "check") || strings.Contains(mdstatContent, "repair") {
			status.Active = true

			// Parse the type
			if strings.Contains(mdstatContent, "check") {
				status.Type = "check"
			} else if strings.Contains(mdstatContent, "repair") {
				status.Type = "correct"
			}

			// Parse progress if available
			// Example: [==>..................]  recovery = 12.5% (1234567/9876543) finish=123.4min speed=45678K/sec
			lines := strings.Split(mdstatContent, "\n")
			for _, line := range lines {
				if strings.Contains(line, "%") && (strings.Contains(line, "recovery") || strings.Contains(line, "check")) {
					// Extract progress percentage
					if idx := strings.Index(line, "%"); idx > 0 {
						start := idx - 1
						for start > 0 && (line[start] >= '0' && line[start] <= '9' || line[start] == '.') {
							start--
						}
						if start < idx {
							if progress, err := strconv.ParseFloat(line[start+1:idx], 64); err == nil {
								status.Progress = progress
							}
						}
					}

					// Extract speed
					if speedIdx := strings.Index(line, "speed="); speedIdx >= 0 {
						speedEnd := strings.Index(line[speedIdx:], " ")
						if speedEnd > 0 {
							status.Speed = line[speedIdx+6 : speedIdx+speedEnd]
						}
					}

					// Extract time remaining
					if finishIdx := strings.Index(line, "finish="); finishIdx >= 0 {
						finishEnd := strings.Index(line[finishIdx:], " ")
						if finishEnd > 0 {
							status.TimeRemaining = line[finishIdx+7 : finishIdx+finishEnd]
						}
					}
				}
			}
		}
	}

	return status, nil
}

// StartParityCheck starts a parity check operation
func (s *StorageMonitor) StartParityCheck(checkType string, priority string) error {
	logger.Blue("Starting parity %s with priority %s", checkType, priority)

	// Build mdcmd command
	cmd := "mdcmd check"
	if checkType == "correct" {
		cmd = "mdcmd check CORRECT=1"
	}

	// Set priority (Unraid uses nice values: low=19, normal=0, high=-10)
	switch priority {
	case "low":
		cmd = "nice -n 19 " + cmd
	case "high":
		cmd = "nice -n -10 " + cmd
		// normal priority uses default nice value (0)
	}

	// Execute the command
	output := lib.GetCmdOutput("sh", "-c", cmd)

	// Check for errors in output
	for _, line := range output {
		if strings.Contains(strings.ToLower(line), "error") ||
			strings.Contains(strings.ToLower(line), "failed") {
			return fmt.Errorf("parity check start failed: %s", line)
		}
	}

	logger.Blue("Parity %s started successfully", checkType)
	return nil
}

// CancelParityCheck cancels an active parity check operation
func (s *StorageMonitor) CancelParityCheck() error {
	logger.Blue("Cancelling parity check")

	// Execute the command to cancel parity check
	output := lib.GetCmdOutput("mdcmd", "nocheck")

	// Check for errors in output
	for _, line := range output {
		if strings.Contains(strings.ToLower(line), "error") ||
			strings.Contains(strings.ToLower(line), "failed") {
			return fmt.Errorf("parity check cancel failed: %s", line)
		}
	}

	logger.Blue("Parity check cancelled successfully")
	return nil
}

// AddDisk adds a disk to the array at the specified position
func (s *StorageMonitor) AddDisk(device string, position string) error {
	logger.Blue("Adding disk %s to position %s", device, position)

	// Safety check: ensure array is stopped
	arrayInfo, err := s.GetArrayInfo()
	if err != nil {
		return fmt.Errorf("failed to get array state: %v", err)
	}

	if arrayInfo.State != "stopped" {
		return fmt.Errorf("array must be stopped to add disks")
	}

	// Validate device exists
	if exists, _ := lib.Exists(device); !exists {
		return fmt.Errorf("device %s does not exist", device)
	}

	// Build mdcmd command to assign disk
	cmd := fmt.Sprintf("mdcmd set %s %s", position, device)

	// Execute the command
	output := lib.GetCmdOutput("sh", "-c", cmd)

	// Check for errors in output
	for _, line := range output {
		if strings.Contains(strings.ToLower(line), "error") ||
			strings.Contains(strings.ToLower(line), "failed") {
			return fmt.Errorf("disk add failed: %s", line)
		}
	}

	logger.Blue("Disk %s added to position %s successfully", device, position)
	return nil
}

// RemoveDisk removes a disk from the specified array position
func (s *StorageMonitor) RemoveDisk(position string) error {
	logger.Blue("Removing disk from position %s", position)

	// Safety check: ensure array is stopped
	arrayInfo, err := s.GetArrayInfo()
	if err != nil {
		return fmt.Errorf("failed to get array state: %v", err)
	}

	if arrayInfo.State != "stopped" {
		return fmt.Errorf("array must be stopped to remove disks")
	}

	// Build mdcmd command to unassign disk
	cmd := fmt.Sprintf("mdcmd unassign %s", position)

	// Execute the command
	output := lib.GetCmdOutput("sh", "-c", cmd)

	// Check for errors in output
	for _, line := range output {
		if strings.Contains(strings.ToLower(line), "error") ||
			strings.Contains(strings.ToLower(line), "failed") {
			return fmt.Errorf("disk remove failed: %s", line)
		}
	}

	logger.Blue("Disk removed from position %s successfully", position)
	return nil
}
