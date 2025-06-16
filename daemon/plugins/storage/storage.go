package storage

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/domalab/uma/daemon/lib"
	"github.com/domalab/uma/daemon/logger"
)

// StorageMonitor provides storage monitoring capabilities
type StorageMonitor struct {
	arrayDisks []DiskInfo
	cacheDisks []DiskInfo
	bootDisk   DiskInfo
}

// SMARTAttribute represents a SMART attribute
type SMARTAttribute struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Value      int    `json:"value"`
	Worst      int    `json:"worst"`
	Threshold  int    `json:"threshold"`
	RawValue   uint64 `json:"raw_value"`
	WhenFailed string `json:"when_failed,omitempty"`
	Flags      string `json:"flags,omitempty"`
}

// SMARTData represents comprehensive SMART health data
type SMARTData struct {
	OverallHealth   string           `json:"overall_health"` // PASSED, FAILED, UNKNOWN
	SmartSupported  bool             `json:"smart_supported"`
	SmartEnabled    bool             `json:"smart_enabled"`
	Temperature     int              `json:"temperature,omitempty"`
	PowerOnHours    uint64           `json:"power_on_hours,omitempty"`
	PowerCycleCount uint64           `json:"power_cycle_count,omitempty"`
	Attributes      []SMARTAttribute `json:"attributes,omitempty"`
	// Critical health indicators
	ReallocatedSectors    uint64 `json:"reallocated_sectors"`
	CurrentPendingSectors uint64 `json:"current_pending_sectors"`
	OfflineUncorrectable  uint64 `json:"offline_uncorrectable"`
	ReallocatedEvents     uint64 `json:"reallocated_events"`
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
	// SMART health data
	SmartData *SMARTData `json:"smart_data,omitempty"`
}

// ConsolidatedDiskInfo represents enhanced disk information for the new /api/v1/storage/disks endpoint
type ConsolidatedDiskInfo struct {
	Device             string  `json:"device"`
	Name               string  `json:"name"`
	Role               string  `json:"role"` // array, parity, cache, boot
	Size               uint64  `json:"size"`
	SizeFormatted      string  `json:"size_formatted"` // "8 TB"
	Used               uint64  `json:"used"`
	UsedFormatted      string  `json:"used_formatted"` // "6.28 TB"
	Available          uint64  `json:"available"`
	AvailableFormatted string  `json:"available_formatted"` // "1.72 TB"
	UsedPercent        float64 `json:"used_percent"`
	FileSystem         string  `json:"filesystem"`
	MountPoint         string  `json:"mount_point"`
	Status             string  `json:"status"`
	Health             string  `json:"health"`

	// Hardware Information
	Model        string `json:"model,omitempty"`
	SerialNumber string `json:"serial_number,omitempty"`
	DiskType     string `json:"disk_type"` // HDD, SSD, NVMe
	Interface    string `json:"interface"` // SATA, NVMe, USB

	// Power and Temperature
	PowerState    string `json:"power_state"` // active, standby, sleeping, unknown
	Temperature   int    `json:"temperature,omitempty"`
	SpinDownDelay int    `json:"spin_down_delay,omitempty"` // minutes, -1 for disabled

	// Comprehensive SMART Data
	SmartData *SMARTData `json:"smart_data,omitempty"`
}

// DisksResponse represents the response for the /api/v1/storage/disks endpoint
type DisksResponse struct {
	ArrayDisks  []ConsolidatedDiskInfo `json:"array_disks"`
	ParityDisks []ConsolidatedDiskInfo `json:"parity_disks"`
	CacheDisks  []ConsolidatedDiskInfo `json:"cache_disks"`
	BootDisk    *ConsolidatedDiskInfo  `json:"boot_disk,omitempty"`
	Summary     DisksSummary           `json:"summary"`
}

// DisksSummary provides summary statistics for all disks
type DisksSummary struct {
	TotalDisks   int `json:"total_disks"`
	HealthyDisks int `json:"healthy_disks"`
	WarningDisks int `json:"warning_disks"`
	FailingDisks int `json:"failing_disks"`
	ActiveDisks  int `json:"active_disks"`
	StandbyDisks int `json:"standby_disks"`
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
	// Human-readable formatted fields
	TotalSizeFormatted string `json:"total_size_formatted"` // "8 TB"
	UsedSizeFormatted  string `json:"used_size_formatted"`  // "6.28 TB"
	FreeSizeFormatted  string `json:"free_size_formatted"`  // "1.72 TB"
}

// ZFSVdev represents a ZFS virtual device
type ZFSVdev struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`   // disk, mirror, raidz1, raidz2, raidz3, spare, cache, log
	State       string    `json:"state"`  // ONLINE, DEGRADED, FAULTED, OFFLINE, UNAVAIL, REMOVED
	Health      string    `json:"health"` // ONLINE, DEGRADED, FAULTED, OFFLINE, UNAVAIL, REMOVED
	ReadErrors  uint64    `json:"read_errors"`
	WriteErrors uint64    `json:"write_errors"`
	CksumErrors uint64    `json:"cksum_errors"`
	Children    []ZFSVdev `json:"children,omitempty"` // For mirror/raidz groups
}

// ZFSPool represents a ZFS storage pool
type ZFSPool struct {
	Name           string    `json:"name"`
	State          string    `json:"state"`           // ONLINE, DEGRADED, FAULTED, OFFLINE, UNAVAIL, REMOVED
	Health         string    `json:"health"`          // ONLINE, DEGRADED, FAULTED, OFFLINE, UNAVAIL, REMOVED
	Size           uint64    `json:"size"`            // Total pool size in bytes
	Allocated      uint64    `json:"allocated"`       // Allocated space in bytes
	Free           uint64    `json:"free"`            // Free space in bytes
	UsedPercent    float64   `json:"used_percent"`    // Used percentage
	SizeFormatted  string    `json:"size_formatted"`  // "8.0 TB"
	AllocFormatted string    `json:"alloc_formatted"` // "6.28 TB"
	FreeFormatted  string    `json:"free_formatted"`  // "1.72 TB"
	Fragmentation  float64   `json:"fragmentation"`   // Fragmentation percentage
	Deduplication  float64   `json:"deduplication"`   // Dedup ratio
	Compression    float64   `json:"compression"`     // Compression ratio
	ReadOps        uint64    `json:"read_ops"`        // Read operations
	WriteOps       uint64    `json:"write_ops"`       // Write operations
	ReadBandwidth  uint64    `json:"read_bandwidth"`  // Read bandwidth in bytes/sec
	WriteBandwidth uint64    `json:"write_bandwidth"` // Write bandwidth in bytes/sec
	Vdevs          []ZFSVdev `json:"vdevs"`           // Virtual devices in the pool
	LastScrub      string    `json:"last_scrub"`      // Last scrub date/time
	ScrubStatus    string    `json:"scrub_status"`    // none, scrub in progress, scrub completed
	ErrorCount     uint64    `json:"error_count"`     // Total error count
	Version        string    `json:"version"`         // ZFS version
	Features       []string  `json:"features"`        // Enabled features
	LastUpdated    string    `json:"last_updated"`    // ISO 8601 timestamp
}

// ZFSInfo represents comprehensive ZFS information
type ZFSInfo struct {
	Available   bool      `json:"available"`     // Whether ZFS is available on the system
	Version     string    `json:"version"`       // ZFS version
	Pools       []ZFSPool `json:"pools"`         // All ZFS pools
	ARCSize     uint64    `json:"arc_size"`      // ARC cache size in bytes
	ARCMax      uint64    `json:"arc_max"`       // ARC max size in bytes
	ARCHitRatio float64   `json:"arc_hit_ratio"` // ARC hit ratio percentage
	LastUpdated string    `json:"last_updated"`  // ISO 8601 timestamp
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
	// Human-readable formatted fields
	TotalSizeFormatted string `json:"total_size_formatted"` // "1 TB"
	UsedSizeFormatted  string `json:"used_size_formatted"`  // "512 GB"
	FreeSizeFormatted  string `json:"free_size_formatted"`  // "512 GB"
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
			// Get comprehensive SMART data
			s.getComprehensiveSMARTData(&diskInfo)

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
				// Get comprehensive SMART data
				s.getComprehensiveSMARTData(&diskInfo)
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
		// Get comprehensive SMART data
		s.getComprehensiveSMARTData(&diskInfo)
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

	// Populate human-readable formatted fields
	cache.TotalSizeFormatted = s.formatBytes(cache.TotalSize)
	cache.UsedSizeFormatted = s.formatBytes(cache.UsedSize)
	cache.FreeSizeFormatted = s.formatBytes(cache.FreeSize)

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
	output := lib.GetCmdOutput("hdparm", "-C", actualDevice)
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

// getComprehensiveSMARTData collects detailed SMART information for a disk
func (s *StorageMonitor) getComprehensiveSMARTData(disk *DiskInfo) {
	if disk.Device == "" {
		return
	}

	actualDevice := s.resolveDevicePath(disk.Device)
	if actualDevice == "" {
		return
	}

	smartData := &SMARTData{
		OverallHealth:  "UNKNOWN",
		SmartSupported: false,
		SmartEnabled:   false,
		Attributes:     make([]SMARTAttribute, 0),
	}

	// Get SMART information and attributes
	output := lib.GetCmdOutput("smartctl", "-A", "-i", "-H", actualDevice)
	s.parseSMARTOutput(output, smartData)

	// Only set SMART data if we got useful information
	if smartData.SmartSupported || len(smartData.Attributes) > 0 {
		disk.SmartData = smartData
	}
}

// parseSMARTOutput parses smartctl output to extract comprehensive SMART data
func (s *StorageMonitor) parseSMARTOutput(output []string, smartData *SMARTData) {
	inAttributeSection := false

	for _, line := range output {
		line = strings.TrimSpace(line)

		// Parse SMART support and status
		if strings.Contains(line, "SMART support is:") {
			if strings.Contains(line, "Available") {
				smartData.SmartSupported = true
			}
		} else if strings.Contains(line, "SMART support is: Enabled") {
			smartData.SmartEnabled = true
		} else if strings.Contains(line, "SMART overall-health self-assessment test result:") {
			if strings.Contains(line, "PASSED") {
				smartData.OverallHealth = "PASSED"
			} else if strings.Contains(line, "FAILED") {
				smartData.OverallHealth = "FAILED"
			}
		} else if strings.Contains(line, "SMART Health Status:") {
			// NVMe drives
			if strings.Contains(line, "OK") {
				smartData.OverallHealth = "PASSED"
			} else {
				smartData.OverallHealth = "FAILED"
			}
		}

		// Parse temperature
		if strings.Contains(line, "Temperature_Celsius") || strings.Contains(line, "Airflow_Temperature_Cel") {
			fields := strings.Fields(line)
			if len(fields) >= 10 {
				if temp, err := strconv.Atoi(fields[9]); err == nil {
					smartData.Temperature = temp
				}
			}
		}

		// Parse power on hours
		if strings.Contains(line, "Power_On_Hours") {
			fields := strings.Fields(line)
			if len(fields) >= 10 {
				if hours, err := strconv.ParseUint(fields[9], 10, 64); err == nil {
					smartData.PowerOnHours = hours
				}
			}
		}

		// Parse power cycle count
		if strings.Contains(line, "Power_Cycle_Count") {
			fields := strings.Fields(line)
			if len(fields) >= 10 {
				if cycles, err := strconv.ParseUint(fields[9], 10, 64); err == nil {
					smartData.PowerCycleCount = cycles
				}
			}
		}

		// Detect start of attribute table
		if strings.Contains(line, "ID# ATTRIBUTE_NAME") {
			inAttributeSection = true
			continue
		}

		// Parse SMART attributes
		if inAttributeSection && len(line) > 0 && !strings.HasPrefix(line, "=") {
			if attr := s.parseSMARTAttribute(line); attr != nil {
				smartData.Attributes = append(smartData.Attributes, *attr)

				// Extract critical health indicators
				switch attr.Name {
				case "Reallocated_Sector_Ct":
					smartData.ReallocatedSectors = attr.RawValue
				case "Current_Pending_Sector":
					smartData.CurrentPendingSectors = attr.RawValue
				case "Offline_Uncorrectable":
					smartData.OfflineUncorrectable = attr.RawValue
				case "Reallocated_Event_Count":
					smartData.ReallocatedEvents = attr.RawValue
				}
			}
		}

		// End of attribute section
		if inAttributeSection && (strings.HasPrefix(line, "=") || line == "") {
			inAttributeSection = false
		}
	}
}

// parseSMARTAttribute parses a single SMART attribute line
func (s *StorageMonitor) parseSMARTAttribute(line string) *SMARTAttribute {
	fields := strings.Fields(line)
	if len(fields) < 10 {
		return nil
	}

	// Parse ID
	id, err := strconv.Atoi(fields[0])
	if err != nil {
		return nil
	}

	// Parse values
	value, err := strconv.Atoi(fields[3])
	if err != nil {
		return nil
	}

	worst, err := strconv.Atoi(fields[4])
	if err != nil {
		return nil
	}

	threshold, err := strconv.Atoi(fields[5])
	if err != nil {
		return nil
	}

	// Parse raw value (can be complex, take the last field)
	rawValue, err := strconv.ParseUint(fields[9], 10, 64)
	if err != nil {
		// Try to parse hex values or complex formats
		if strings.HasPrefix(fields[9], "0x") {
			if val, err := strconv.ParseUint(fields[9][2:], 16, 64); err == nil {
				rawValue = val
			}
		}
	}

	return &SMARTAttribute{
		ID:         id,
		Name:       fields[1],
		Value:      value,
		Worst:      worst,
		Threshold:  threshold,
		RawValue:   rawValue,
		WhenFailed: fields[8],
		Flags:      fields[2],
	}
}

// formatBytes converts bytes to human-readable format
func (s *StorageMonitor) formatBytes(bytes uint64) string {
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

// convertDiskInfoToConsolidated converts DiskInfo to ConsolidatedDiskInfo with enhanced formatting
func (s *StorageMonitor) convertDiskInfoToConsolidated(disk DiskInfo, role string) ConsolidatedDiskInfo {
	return ConsolidatedDiskInfo{
		Device:             disk.Device,
		Name:               disk.Name,
		Role:               role,
		Size:               disk.Size,
		SizeFormatted:      s.formatBytes(disk.Size),
		Used:               disk.Used,
		UsedFormatted:      s.formatBytes(disk.Used),
		Available:          disk.Available,
		AvailableFormatted: s.formatBytes(disk.Available),
		UsedPercent:        disk.UsedPercent,
		FileSystem:         disk.FileSystem,
		MountPoint:         disk.MountPoint,
		Status:             disk.Status,
		Health:             disk.Health,
		Model:              disk.Model,
		SerialNumber:       disk.SerialNumber,
		DiskType:           disk.DiskType,
		Interface:          disk.Interface,
		PowerState:         disk.PowerState,
		Temperature:        disk.Temperature,
		SpinDownDelay:      disk.SpinDownDelay,
		SmartData:          disk.SmartData,
	}
}

// GetConsolidatedDisksInfo returns consolidated disk information for the new /api/v1/storage/disks endpoint
func (s *StorageMonitor) GetConsolidatedDisksInfo() (*DisksResponse, error) {
	response := &DisksResponse{
		ArrayDisks:  make([]ConsolidatedDiskInfo, 0),
		ParityDisks: make([]ConsolidatedDiskInfo, 0),
		CacheDisks:  make([]ConsolidatedDiskInfo, 0),
		Summary:     DisksSummary{},
	}

	// Get array information
	arrayInfo, err := s.GetArrayInfo()
	if err != nil {
		logger.Yellow("Failed to get array info for consolidated disks: %v", err)
	} else {
		// Process array disks
		for _, disk := range arrayInfo.Disks {
			var role string
			if strings.HasPrefix(disk.Name, "parity") {
				role = "parity"
				consolidated := s.convertDiskInfoToConsolidated(disk, role)
				response.ParityDisks = append(response.ParityDisks, consolidated)
			} else if strings.HasPrefix(disk.Name, "disk") {
				role = "array"
				consolidated := s.convertDiskInfoToConsolidated(disk, role)
				response.ArrayDisks = append(response.ArrayDisks, consolidated)
			}
		}
	}

	// Get cache information
	cacheInfos, err := s.GetCacheInfo()
	if err != nil {
		logger.Yellow("Failed to get cache info for consolidated disks: %v", err)
	} else {
		for _, cacheInfo := range cacheInfos {
			for _, disk := range cacheInfo.Disks {
				consolidated := s.convertDiskInfoToConsolidated(disk, "cache")
				response.CacheDisks = append(response.CacheDisks, consolidated)
			}
		}
	}

	// Get boot disk information
	bootDisk, err := s.GetBootDiskInfo()
	if err != nil {
		logger.Yellow("Failed to get boot disk info for consolidated disks: %v", err)
	} else {
		consolidated := s.convertDiskInfoToConsolidated(*bootDisk, "boot")
		response.BootDisk = &consolidated
	}

	// Calculate summary statistics
	s.calculateDisksSummary(response)

	return response, nil
}

// calculateDisksSummary calculates summary statistics for all disks
func (s *StorageMonitor) calculateDisksSummary(response *DisksResponse) {
	allDisks := make([]ConsolidatedDiskInfo, 0)

	// Collect all disks
	allDisks = append(allDisks, response.ArrayDisks...)
	allDisks = append(allDisks, response.ParityDisks...)
	allDisks = append(allDisks, response.CacheDisks...)
	if response.BootDisk != nil {
		allDisks = append(allDisks, *response.BootDisk)
	}

	response.Summary.TotalDisks = len(allDisks)

	// Count health and power states
	for _, disk := range allDisks {
		switch disk.Health {
		case "healthy":
			response.Summary.HealthyDisks++
		case "failing":
			response.Summary.FailingDisks++
		default:
			response.Summary.WarningDisks++
		}

		switch disk.PowerState {
		case "active":
			response.Summary.ActiveDisks++
		case "standby", "sleeping":
			response.Summary.StandbyDisks++
		}
	}
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

	// Populate human-readable formatted fields
	arrayInfo.TotalSizeFormatted = s.formatBytes(arrayInfo.TotalSize)
	arrayInfo.UsedSizeFormatted = s.formatBytes(arrayInfo.UsedSize)
	arrayInfo.FreeSizeFormatted = s.formatBytes(arrayInfo.FreeSize)
}

// calculateCacheTotals calculates total sizes for cache pools
func (s *StorageMonitor) calculateCacheTotals(cache *CacheInfo) {
	// Cache totals are already calculated in getCacheUsage
	// This method is here for consistency and future enhancements
}

// Array Control Operations

// StartArray starts the Unraid array with proper orchestration sequence
func (s *StorageMonitor) StartArray(maintenanceMode bool, checkFilesystem bool) error {
	logger.Blue("Starting Unraid array with orchestration (maintenance: %v, check_fs: %v)", maintenanceMode, checkFilesystem)

	// Step 1: Validate array configuration
	if err := s.validateArrayConfiguration(); err != nil {
		return fmt.Errorf("array configuration validation failed: %v", err)
	}

	// Step 2: Check for any running parity operations
	if err := s.checkParityOperations(); err != nil {
		return fmt.Errorf("parity operation check failed: %v", err)
	}

	// Step 3: Start array via mdcmd with proper parameters
	cmd := "mdcmd start"
	if maintenanceMode {
		cmd += " MAINTENANCE=1"
	}
	if checkFilesystem {
		cmd += " CHECK=1"
	}

	logger.Blue("Executing array start command: %s", cmd)
	output := lib.GetCmdOutput("sh", "-c", cmd)

	// Check for errors in output
	for _, line := range output {
		if strings.Contains(strings.ToLower(line), "error") ||
			strings.Contains(strings.ToLower(line), "failed") {
			return fmt.Errorf("array start failed: %s", line)
		}
	}

	// Step 4: Wait for array to become available
	if err := s.waitForArrayState("started", 60); err != nil {
		return fmt.Errorf("array failed to start within timeout: %v", err)
	}

	// Step 5: Verify filesystem mounts
	if err := s.verifyFilesystemMounts(); err != nil {
		logger.Yellow("Warning: Some filesystems may not have mounted properly: %v", err)
	}

	logger.Blue("Array start orchestration completed successfully")
	return nil
}

// StopArray stops the Unraid array with proper orchestration sequence
func (s *StorageMonitor) StopArray(force bool, unmountShares bool, stopContainers bool, stopVMs bool) error {
	logger.Blue("Stopping Unraid array with orchestration (force: %v, unmount_shares: %v, stop_containers: %v, stop_vms: %v)",
		force, unmountShares, stopContainers, stopVMs)

	// Step 1: Stop Docker containers if requested
	if stopContainers {
		logger.Blue("Step 1: Stopping Docker containers...")
		if err := s.stopDockerContainers(); err != nil && !force {
			return fmt.Errorf("failed to stop Docker containers: %v", err)
		}
	}

	// Step 2: Stop VMs if requested
	if stopVMs {
		logger.Blue("Step 2: Stopping virtual machines...")
		if err := s.stopVirtualMachines(); err != nil && !force {
			return fmt.Errorf("failed to stop virtual machines: %v", err)
		}
	}

	// Step 3: Handle running parity operations
	if !force {
		logger.Blue("Step 3: Checking for running parity operations...")
		if err := s.handleParityOperations(); err != nil {
			return fmt.Errorf("failed to handle parity operations: %v", err)
		}
	}

	// Step 4: Unmount user shares (FUSE mounts)
	if unmountShares {
		logger.Blue("Step 4: Unmounting user shares...")
		if err := s.unmountUserShares(); err != nil && !force {
			return fmt.Errorf("failed to unmount user shares: %v", err)
		}
	}

	// Step 5: Unmount array disks in reverse dependency order
	logger.Blue("Step 5: Unmounting array disks...")
	if err := s.unmountArrayDisks(); err != nil && !force {
		return fmt.Errorf("failed to unmount array disks: %v", err)
	}

	// Step 6: Stop MD devices using mdcmd
	logger.Blue("Step 6: Stopping MD devices...")
	cmd := "mdcmd stop"
	if force {
		cmd += " FORCE=1"
	}

	logger.Blue("Executing array stop command: %s", cmd)
	output := lib.GetCmdOutput("sh", "-c", cmd)

	// Check for errors in output
	for _, line := range output {
		if strings.Contains(strings.ToLower(line), "error") ||
			strings.Contains(strings.ToLower(line), "failed") {
			return fmt.Errorf("array stop failed: %s", line)
		}
	}

	// Step 7: Wait for array to stop
	if err := s.waitForArrayState("stopped", 120); err != nil {
		return fmt.Errorf("array failed to stop within timeout: %v", err)
	}

	logger.Blue("Array stop orchestration completed successfully")
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

// Array Orchestration Helper Methods

// validateArrayConfiguration validates the array configuration before starting
func (s *StorageMonitor) validateArrayConfiguration() error {
	logger.Blue("Validating array configuration...")

	// Check if array configuration exists
	if exists, _ := lib.Exists("/boot/config/disk.cfg"); !exists {
		return fmt.Errorf("array configuration file not found")
	}

	// Read disk assignments to ensure we have valid configuration
	diskAssignments, err := s.readDiskAssignments()
	if err != nil {
		return fmt.Errorf("failed to read disk assignments: %v", err)
	}

	if len(diskAssignments) == 0 {
		return fmt.Errorf("no disks assigned to array")
	}

	// Validate that assigned devices exist
	for device, assignment := range diskAssignments {
		if exists, _ := lib.Exists(device); !exists {
			logger.Yellow("Warning: Assigned device %s (%s) not found", device, assignment.Name)
		}
	}

	logger.Blue("Array configuration validation completed")
	return nil
}

// checkParityOperations checks for any running parity operations
func (s *StorageMonitor) checkParityOperations() error {
	logger.Blue("Checking for running parity operations...")

	// Check mdstat for any active sync operations
	output := lib.GetCmdOutput("cat", "/proc/mdstat")
	for _, line := range output {
		if strings.Contains(line, "resync") || strings.Contains(line, "recovery") || strings.Contains(line, "check") {
			return fmt.Errorf("parity operation in progress: %s", strings.TrimSpace(line))
		}
	}

	logger.Blue("No active parity operations found")
	return nil
}

// waitForArrayState waits for the array to reach the specified state
func (s *StorageMonitor) waitForArrayState(expectedState string, timeoutSeconds int) error {
	logger.Blue("Waiting for array state: %s (timeout: %ds)", expectedState, timeoutSeconds)

	for i := 0; i < timeoutSeconds; i++ {
		currentState := s.getArrayState()
		if currentState == expectedState {
			logger.Blue("Array reached expected state: %s", expectedState)
			return nil
		}

		if i%10 == 0 { // Log every 10 seconds
			logger.Blue("Array state: %s, waiting for: %s (%d/%ds)", currentState, expectedState, i, timeoutSeconds)
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("timeout waiting for array state %s", expectedState)
}

// verifyFilesystemMounts verifies that array filesystems are properly mounted
func (s *StorageMonitor) verifyFilesystemMounts() error {
	logger.Blue("Verifying filesystem mounts...")

	// Check /proc/mounts for array disk mounts
	output := lib.GetCmdOutput("cat", "/proc/mounts")
	mountedDisks := 0

	for _, line := range output {
		if strings.Contains(line, "/mnt/disk") || strings.Contains(line, "/mnt/cache") {
			mountedDisks++
		}
	}

	if mountedDisks == 0 {
		return fmt.Errorf("no array disks appear to be mounted")
	}

	logger.Blue("Verified %d array disk mounts", mountedDisks)
	return nil
}

// stopDockerContainers stops all running Docker containers
func (s *StorageMonitor) stopDockerContainers() error {
	logger.Blue("Stopping all Docker containers...")

	// Get list of running containers
	output := lib.GetCmdOutput("docker", "ps", "-q")
	if len(output) == 0 {
		logger.Blue("No running Docker containers found")
		return nil
	}

	containerIDs := make([]string, 0)
	for _, line := range output {
		if strings.TrimSpace(line) != "" {
			containerIDs = append(containerIDs, strings.TrimSpace(line))
		}
	}

	if len(containerIDs) == 0 {
		logger.Blue("No running Docker containers found")
		return nil
	}

	// Stop all containers with timeout
	logger.Blue("Stopping %d Docker containers...", len(containerIDs))
	args := append([]string{"stop", "-t", "30"}, containerIDs...)
	output = lib.GetCmdOutput("docker", args...)

	// Check for errors
	for _, line := range output {
		if strings.Contains(strings.ToLower(line), "error") {
			return fmt.Errorf("failed to stop containers: %s", line)
		}
	}

	logger.Blue("Successfully stopped %d Docker containers", len(containerIDs))
	return nil
}

// stopVirtualMachines stops all running virtual machines
func (s *StorageMonitor) stopVirtualMachines() error {
	logger.Blue("Stopping all virtual machines...")

	// Get list of running VMs via virsh
	output := lib.GetCmdOutput("virsh", "list", "--state-running", "--name")
	if len(output) == 0 {
		logger.Blue("No running virtual machines found")
		return nil
	}

	vmNames := make([]string, 0)
	for _, line := range output {
		vmName := strings.TrimSpace(line)
		if vmName != "" {
			vmNames = append(vmNames, vmName)
		}
	}

	if len(vmNames) == 0 {
		logger.Blue("No running virtual machines found")
		return nil
	}

	// Shutdown VMs gracefully
	logger.Blue("Shutting down %d virtual machines...", len(vmNames))
	for _, vmName := range vmNames {
		logger.Blue("Shutting down VM: %s", vmName)
		output := lib.GetCmdOutput("virsh", "shutdown", vmName)

		// Check for errors
		for _, line := range output {
			if strings.Contains(strings.ToLower(line), "error") {
				logger.Yellow("Warning: Failed to shutdown VM %s: %s", vmName, line)
			}
		}
	}

	// Wait for VMs to shutdown (up to 60 seconds)
	logger.Blue("Waiting for VMs to shutdown...")
	for i := 0; i < 60; i++ {
		output := lib.GetCmdOutput("virsh", "list", "--state-running", "--name")
		runningVMs := 0
		for _, line := range output {
			if strings.TrimSpace(line) != "" {
				runningVMs++
			}
		}

		if runningVMs == 0 {
			logger.Blue("All VMs have shutdown successfully")
			return nil
		}

		if i%10 == 0 {
			logger.Blue("Waiting for %d VMs to shutdown... (%d/60s)", runningVMs, i)
		}

		time.Sleep(1 * time.Second)
	}

	logger.Yellow("Warning: Some VMs may still be running after timeout")
	return nil
}

// handleParityOperations handles any running parity operations
func (s *StorageMonitor) handleParityOperations() error {
	logger.Blue("Handling running parity operations...")

	// Check for active parity operations
	output := lib.GetCmdOutput("cat", "/proc/mdstat")
	for _, line := range output {
		if strings.Contains(line, "resync") || strings.Contains(line, "recovery") || strings.Contains(line, "check") {
			logger.Blue("Found active parity operation: %s", strings.TrimSpace(line))

			// Cancel the parity operation
			logger.Blue("Cancelling parity operation...")
			if err := s.CancelParityCheck(); err != nil {
				return fmt.Errorf("failed to cancel parity operation: %v", err)
			}

			// Wait a moment for cancellation to take effect
			time.Sleep(5 * time.Second)
			break
		}
	}

	logger.Blue("Parity operations handled successfully")
	return nil
}

// unmountUserShares unmounts all user shares (FUSE mounts)
func (s *StorageMonitor) unmountUserShares() error {
	logger.Blue("Unmounting user shares...")

	// Get list of mounted user shares
	output := lib.GetCmdOutput("mount")
	userShares := make([]string, 0)

	for _, line := range output {
		if strings.Contains(line, "shfs") && strings.Contains(line, "/mnt/user") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				mountPoint := parts[2]
				userShares = append(userShares, mountPoint)
			}
		}
	}

	if len(userShares) == 0 {
		logger.Blue("No user shares found to unmount")
		return nil
	}

	// Unmount each user share
	logger.Blue("Unmounting %d user shares...", len(userShares))
	for _, mountPoint := range userShares {
		logger.Blue("Unmounting user share: %s", mountPoint)
		output := lib.GetCmdOutput("umount", mountPoint)

		// Check for errors
		for _, line := range output {
			if strings.Contains(strings.ToLower(line), "error") || strings.Contains(strings.ToLower(line), "busy") {
				return fmt.Errorf("failed to unmount user share %s: %s", mountPoint, line)
			}
		}
	}

	logger.Blue("Successfully unmounted %d user shares", len(userShares))
	return nil
}

// unmountArrayDisks unmounts array disks in reverse dependency order
func (s *StorageMonitor) unmountArrayDisks() error {
	logger.Blue("Unmounting array disks...")

	// Get list of mounted array disks
	output := lib.GetCmdOutput("mount")
	arrayMounts := make([]string, 0)

	for _, line := range output {
		if strings.Contains(line, "/mnt/disk") || strings.Contains(line, "/mnt/cache") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				mountPoint := parts[2]
				arrayMounts = append(arrayMounts, mountPoint)
			}
		}
	}

	if len(arrayMounts) == 0 {
		logger.Blue("No array disks found to unmount")
		return nil
	}

	// Sort mount points in reverse order for proper unmounting
	// This ensures dependencies are handled correctly
	sort.Sort(sort.Reverse(sort.StringSlice(arrayMounts)))

	// Unmount each array disk
	logger.Blue("Unmounting %d array disks...", len(arrayMounts))
	for _, mountPoint := range arrayMounts {
		logger.Blue("Unmounting array disk: %s", mountPoint)

		// Try lazy unmount first, then force if needed
		_ = lib.GetCmdOutput("umount", "-l", mountPoint)

		// Check if unmount was successful
		stillMounted := false
		checkOutput := lib.GetCmdOutput("mount")
		for _, line := range checkOutput {
			if strings.Contains(line, mountPoint) {
				stillMounted = true
				break
			}
		}

		if stillMounted {
			logger.Yellow("Lazy unmount failed for %s, trying force unmount...", mountPoint)
			output = lib.GetCmdOutput("umount", "-f", mountPoint)

			// Check for errors
			for _, line := range output {
				if strings.Contains(strings.ToLower(line), "error") {
					return fmt.Errorf("failed to force unmount %s: %s", mountPoint, line)
				}
			}
		}
	}

	logger.Blue("Successfully unmounted %d array disks", len(arrayMounts))
	return nil
}

// GetZFSInfo returns comprehensive ZFS information
func (s *StorageMonitor) GetZFSInfo() (*ZFSInfo, error) {
	timestamp := time.Now().UTC().Format(time.RFC3339)

	zfsInfo := &ZFSInfo{
		Available:   false,
		Pools:       make([]ZFSPool, 0),
		LastUpdated: timestamp,
	}

	// Check if ZFS is available
	if !s.isZFSAvailable() {
		return zfsInfo, nil
	}

	zfsInfo.Available = true

	// Get ZFS version
	if version := s.getZFSVersion(); version != "" {
		zfsInfo.Version = version
	}

	// Get ZFS pools
	pools, err := s.getZFSPools()
	if err != nil {
		logger.Yellow("Failed to get ZFS pools: %v", err)
	} else {
		zfsInfo.Pools = pools
	}

	// Get ARC statistics
	arcSize, arcMax, hitRatio := s.getZFSARCStats()
	zfsInfo.ARCSize = arcSize
	zfsInfo.ARCMax = arcMax
	zfsInfo.ARCHitRatio = hitRatio

	return zfsInfo, nil
}

// isZFSAvailable checks if ZFS is available on the system
func (s *StorageMonitor) isZFSAvailable() bool {
	// Check if zpool command exists
	if _, err := exec.LookPath("zpool"); err != nil {
		return false
	}

	// Check if zfs command exists
	if _, err := exec.LookPath("zfs"); err != nil {
		return false
	}

	// Check if ZFS kernel module is loaded
	if exists, _ := lib.Exists("/proc/spl/kstat/zfs"); !exists {
		return false
	}

	return true
}

// getZFSVersion gets the ZFS version
func (s *StorageMonitor) getZFSVersion() string {
	output := lib.GetCmdOutput("zfs", "version")
	if len(output) > 0 {
		// Parse version from output like "zfs-2.1.5-1"
		for _, line := range output {
			if strings.Contains(line, "zfs-") {
				parts := strings.Fields(line)
				if len(parts) > 0 {
					return strings.TrimPrefix(parts[0], "zfs-")
				}
			}
		}
	}
	return "unknown"
}

// getZFSPools gets information about all ZFS pools
func (s *StorageMonitor) getZFSPools() ([]ZFSPool, error) {
	pools := make([]ZFSPool, 0)

	// Get list of pools
	poolNames := s.getZFSPoolNames()
	if len(poolNames) == 0 {
		return pools, nil
	}

	for _, poolName := range poolNames {
		pool, err := s.getZFSPoolInfo(poolName)
		if err != nil {
			logger.Yellow("Failed to get info for pool %s: %v", poolName, err)
			continue
		}
		pools = append(pools, pool)
	}

	return pools, nil
}

// getZFSPoolNames gets the names of all ZFS pools
func (s *StorageMonitor) getZFSPoolNames() []string {
	output := lib.GetCmdOutput("zpool", "list", "-H", "-o", "name")
	poolNames := make([]string, 0)

	for _, line := range output {
		line = strings.TrimSpace(line)
		if line != "" {
			poolNames = append(poolNames, line)
		}
	}

	return poolNames
}

// getZFSPoolInfo gets detailed information about a specific ZFS pool
func (s *StorageMonitor) getZFSPoolInfo(poolName string) (ZFSPool, error) {
	timestamp := time.Now().UTC().Format(time.RFC3339)

	pool := ZFSPool{
		Name:        poolName,
		Vdevs:       make([]ZFSVdev, 0),
		Features:    make([]string, 0),
		LastUpdated: timestamp,
	}

	// Get pool status and properties
	if err := s.parseZFSPoolStatus(&pool); err != nil {
		return pool, err
	}

	// Get pool usage statistics
	if err := s.parseZFSPoolUsage(&pool); err != nil {
		return pool, err
	}

	// Get pool I/O statistics
	s.parseZFSPoolIOStats(&pool)

	// Get pool features
	s.parseZFSPoolFeatures(&pool)

	return pool, nil
}

// getZFSARCStats gets ZFS ARC cache statistics
func (s *StorageMonitor) getZFSARCStats() (uint64, uint64, float64) {
	var arcSize, arcMax uint64
	var hitRatio float64

	// Read ARC stats from /proc/spl/kstat/zfs/arcstats
	content, err := os.ReadFile("/proc/spl/kstat/zfs/arcstats")
	if err != nil {
		return 0, 0, 0
	}

	lines := strings.Split(string(content), "\n")
	var hits, misses uint64

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		switch fields[0] {
		case "size":
			if val, err := strconv.ParseUint(fields[2], 10, 64); err == nil {
				arcSize = val
			}
		case "c_max":
			if val, err := strconv.ParseUint(fields[2], 10, 64); err == nil {
				arcMax = val
			}
		case "hits":
			if val, err := strconv.ParseUint(fields[2], 10, 64); err == nil {
				hits = val
			}
		case "misses":
			if val, err := strconv.ParseUint(fields[2], 10, 64); err == nil {
				misses = val
			}
		}
	}

	// Calculate hit ratio
	if hits+misses > 0 {
		hitRatio = float64(hits) / float64(hits+misses) * 100
	}

	return arcSize, arcMax, hitRatio
}

// parseZFSPoolStatus parses zpool status output for a pool
func (s *StorageMonitor) parseZFSPoolStatus(pool *ZFSPool) error {
	output := lib.GetCmdOutput("zpool", "status", pool.Name)
	if len(output) == 0 {
		return fmt.Errorf("no status output for pool %s", pool.Name)
	}

	inVdevSection := false

	for _, line := range output {
		line = strings.TrimSpace(line)

		// Parse pool state and health
		if strings.HasPrefix(line, "state:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				pool.State = parts[1]
				pool.Health = parts[1] // State and health are often the same
			}
		} else if strings.HasPrefix(line, "status:") {
			// Additional status information
			continue
		} else if strings.HasPrefix(line, "action:") {
			// Action recommendations
			continue
		} else if strings.HasPrefix(line, "scan:") {
			// Parse scrub information
			s.parseZFSScrubInfo(line, pool)
		} else if strings.Contains(line, "errors:") {
			// Parse error count
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "errors:" && i+1 < len(parts) {
					if errors, err := strconv.ParseUint(parts[i+1], 10, 64); err == nil {
						pool.ErrorCount = errors
					}
					break
				}
			}
		} else if strings.HasPrefix(line, pool.Name) {
			// Start of vdev section
			inVdevSection = true
			continue
		} else if inVdevSection && line != "" {
			// Parse vdev information
			vdev := s.parseZFSVdevLine(line)
			if vdev != nil {
				pool.Vdevs = append(pool.Vdevs, *vdev)
			}
		}
	}

	return nil
}

// parseZFSPoolUsage parses zpool list output for usage statistics
func (s *StorageMonitor) parseZFSPoolUsage(pool *ZFSPool) error {
	output := lib.GetCmdOutput("zpool", "list", "-H", "-o", "size,alloc,free,fragmentation,dedup,compress", pool.Name)
	if len(output) == 0 {
		return fmt.Errorf("no usage output for pool %s", pool.Name)
	}

	fields := strings.Fields(output[0])
	if len(fields) < 6 {
		return fmt.Errorf("invalid usage output for pool %s", pool.Name)
	}

	// Parse size (field 0)
	pool.Size = s.parseZFSSize(fields[0])
	pool.SizeFormatted = fields[0]

	// Parse allocated (field 1)
	pool.Allocated = s.parseZFSSize(fields[1])
	pool.AllocFormatted = fields[1]

	// Parse free (field 2)
	pool.Free = s.parseZFSSize(fields[2])
	pool.FreeFormatted = fields[2]

	// Calculate used percentage
	if pool.Size > 0 {
		pool.UsedPercent = float64(pool.Allocated) / float64(pool.Size) * 100
	}

	// Parse fragmentation (field 3)
	if fragStr := strings.TrimSuffix(fields[3], "%"); fragStr != "-" {
		if frag, err := strconv.ParseFloat(fragStr, 64); err == nil {
			pool.Fragmentation = frag
		}
	}

	// Parse deduplication ratio (field 4)
	if dedupStr := strings.TrimSuffix(fields[4], "x"); dedupStr != "-" {
		if dedup, err := strconv.ParseFloat(dedupStr, 64); err == nil {
			pool.Deduplication = dedup
		}
	}

	// Parse compression ratio (field 5)
	if compStr := strings.TrimSuffix(fields[5], "x"); compStr != "-" {
		if comp, err := strconv.ParseFloat(compStr, 64); err == nil {
			pool.Compression = comp
		}
	}

	return nil
}

// parseZFSPoolIOStats parses zpool iostat output for I/O statistics
func (s *StorageMonitor) parseZFSPoolIOStats(pool *ZFSPool) {
	output := lib.GetCmdOutput("zpool", "iostat", "-H", pool.Name, "1", "1")
	if len(output) < 2 {
		return
	}

	// Skip header and get the data line
	fields := strings.Fields(output[1])
	if len(fields) < 7 {
		return
	}

	// Parse read/write operations and bandwidth
	if readOps, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
		pool.ReadOps = readOps
	}
	if writeOps, err := strconv.ParseUint(fields[2], 10, 64); err == nil {
		pool.WriteOps = writeOps
	}
	if readBW, err := strconv.ParseUint(fields[3], 10, 64); err == nil {
		pool.ReadBandwidth = readBW
	}
	if writeBW, err := strconv.ParseUint(fields[4], 10, 64); err == nil {
		pool.WriteBandwidth = writeBW
	}
}

// parseZFSPoolFeatures parses zpool get output for enabled features
func (s *StorageMonitor) parseZFSPoolFeatures(pool *ZFSPool) {
	output := lib.GetCmdOutput("zpool", "get", "-H", "-o", "property,value", "all", pool.Name)

	for _, line := range output {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			property := fields[0]
			value := fields[1]

			// Collect enabled features
			if strings.HasPrefix(property, "feature@") && (value == "enabled" || value == "active") {
				featureName := strings.TrimPrefix(property, "feature@")
				pool.Features = append(pool.Features, featureName)
			} else if property == "version" {
				pool.Version = value
			}
		}
	}
}

// parseZFSScrubInfo parses scrub information from zpool status
func (s *StorageMonitor) parseZFSScrubInfo(line string, pool *ZFSPool) {
	if strings.Contains(line, "scrub repaired") {
		pool.ScrubStatus = "completed"
		// Extract last scrub date if available
		if strings.Contains(line, "on") {
			parts := strings.Split(line, "on")
			if len(parts) > 1 {
				pool.LastScrub = strings.TrimSpace(parts[len(parts)-1])
			}
		}
	} else if strings.Contains(line, "scrub in progress") {
		pool.ScrubStatus = "in_progress"
	} else if strings.Contains(line, "none requested") {
		pool.ScrubStatus = "none"
	}
}

// parseZFSVdevLine parses a vdev line from zpool status
func (s *StorageMonitor) parseZFSVdevLine(line string) *ZFSVdev {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return nil
	}

	vdev := &ZFSVdev{
		Name:   fields[0],
		State:  fields[1],
		Health: fields[1],
	}

	// Parse error counts if available
	if len(fields) >= 5 {
		if readErr, err := strconv.ParseUint(fields[2], 10, 64); err == nil {
			vdev.ReadErrors = readErr
		}
		if writeErr, err := strconv.ParseUint(fields[3], 10, 64); err == nil {
			vdev.WriteErrors = writeErr
		}
		if cksumErr, err := strconv.ParseUint(fields[4], 10, 64); err == nil {
			vdev.CksumErrors = cksumErr
		}
	}

	// Determine vdev type based on name
	if strings.Contains(vdev.Name, "mirror") {
		vdev.Type = "mirror"
	} else if strings.Contains(vdev.Name, "raidz") {
		if strings.Contains(vdev.Name, "raidz3") {
			vdev.Type = "raidz3"
		} else if strings.Contains(vdev.Name, "raidz2") {
			vdev.Type = "raidz2"
		} else {
			vdev.Type = "raidz1"
		}
	} else if strings.HasPrefix(vdev.Name, "/dev/") || strings.Contains(vdev.Name, "sd") || strings.Contains(vdev.Name, "nvme") {
		vdev.Type = "disk"
	} else {
		vdev.Type = "unknown"
	}

	return vdev
}

// parseZFSSize converts ZFS size strings to bytes
func (s *StorageMonitor) parseZFSSize(sizeStr string) uint64 {
	if sizeStr == "-" || sizeStr == "" {
		return 0
	}

	// Remove any trailing characters and convert to uppercase
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	// Extract numeric part and unit
	var numStr string
	var unit string

	for i, char := range sizeStr {
		if char >= '0' && char <= '9' || char == '.' {
			numStr += string(char)
		} else {
			unit = sizeStr[i:]
			break
		}
	}

	// Parse the numeric value
	value, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0
	}

	// Convert based on unit
	switch unit {
	case "K", "KB":
		return uint64(value * 1024)
	case "M", "MB":
		return uint64(value * 1024 * 1024)
	case "G", "GB":
		return uint64(value * 1024 * 1024 * 1024)
	case "T", "TB":
		return uint64(value * 1024 * 1024 * 1024 * 1024)
	case "P", "PB":
		return uint64(value * 1024 * 1024 * 1024 * 1024 * 1024)
	default:
		// Assume bytes if no unit
		return uint64(value)
	}
}
