package adapters

import (
	"bufio"
	"os"
	"strings"
	"testing"
	"time"
)

// MockFileSystem provides mock filesystem access for testing
type MockFileSystem struct {
	files map[string]string
	stats map[string]bool
}

func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		files: make(map[string]string),
		stats: make(map[string]bool),
	}
}

func (m *MockFileSystem) SetFileContent(path, content string) {
	m.files[path] = content
}

func (m *MockFileSystem) SetFileExists(path string, exists bool) {
	m.stats[path] = exists
}

func (m *MockFileSystem) Open(name string) (FileInterface, error) {
	if content, exists := m.files[name]; exists {
		return &MockFile{content: content}, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) ReadFile(filename string) ([]byte, error) {
	if content, exists := m.files[filename]; exists {
		return []byte(content), nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	if exists, found := m.stats[name]; found && exists {
		return &MockFileInfo{name: name}, nil
	}
	return nil, os.ErrNotExist
}

// MockFile implements FileInterface for testing
type MockFile struct {
	content string
}

func (m *MockFile) Close() error {
	return nil
}

func (m *MockFile) Read(b []byte) (int, error) {
	copy(b, []byte(m.content))
	return len(m.content), nil
}

func (m *MockFile) Scan() *bufio.Scanner {
	return bufio.NewScanner(strings.NewReader(m.content))
}

// MockFileInfo implements os.FileInfo for testing
type MockFileInfo struct {
	name string
}

func (m *MockFileInfo) Name() string       { return m.name }
func (m *MockFileInfo) Size() int64        { return 0 }
func (m *MockFileInfo) Mode() os.FileMode  { return 0 }
func (m *MockFileInfo) ModTime() time.Time { return time.Time{} }
func (m *MockFileInfo) IsDir() bool        { return false }
func (m *MockFileInfo) Sys() interface{}   { return nil }

// MockCommandExecutor provides mock command execution for testing
type MockCommandExecutor struct {
	responses map[string][]byte
}

func NewMockCommandExecutor() *MockCommandExecutor {
	return &MockCommandExecutor{
		responses: make(map[string][]byte),
	}
}

func (m *MockCommandExecutor) SetResponse(name string, args []string, response []byte) {
	key := name + " " + strings.Join(args, " ")
	m.responses[key] = response
}

func (m *MockCommandExecutor) Command(name string, args ...string) CommandInterface {
	key := name + " " + strings.Join(args, " ")
	if response, exists := m.responses[key]; exists {
		return &MockCommand{output: response}
	}
	return &MockCommand{output: []byte{}, err: os.ErrNotExist}
}

// MockCommand implements CommandInterface for testing
type MockCommand struct {
	output []byte
	err    error
}

func (m *MockCommand) Output() ([]byte, error) {
	return m.output, m.err
}

// setupMockSystemAdapter creates a SystemAdapter with mocked dependencies
func setupMockSystemAdapter() (*SystemAdapter, *MockFileSystem, *MockCommandExecutor) {
	mockFS := NewMockFileSystem()
	mockCmd := NewMockCommandExecutor()

	// Mock /proc/cpuinfo
	mockFS.SetFileContent("/proc/cpuinfo", `processor	: 0
vendor_id	: GenuineIntel
cpu family	: 6
model		: 142
model name	: Intel(R) Core(TM) i7-8565U CPU @ 1.80GHz
stepping	: 12
microcode	: 0xf0
cpu MHz		: 1800.000
cache size	: 8192 KB
physical id	: 0
siblings	: 8
core id		: 0
cpu cores	: 4
apicid		: 0
initial apicid	: 0
fpu		: yes
fpu_exception	: yes
cpuid level	: 22
wp		: yes
flags		: fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush dts acpi mmx fxsr sse sse2 ss ht tm pbe syscall nx pdpe1gb rdtscp lm constant_tsc art arch_perfmon pebs bts rep_good nopl xtopology nonstop_tsc cpuid aperfmperf pni pclmulqdq dtes64 monitor ds_cpl vmx est tm2 ssse3 sdbg fma cx16 xtpr pdcm pcid sse4_1 sse4_2 x2apic movbe popcnt tsc_deadline_timer aes xsave avx f16c rdrand lahf_lm abm 3dnowprefetch cpuid_fault epb invpcid_single pti ssbd ibrs ibpb stibp tpr_shadow vnmi flexpriority ept vpid ept_ad fsgsbase tsc_adjust bmi1 avx2 smep bmi2 erms invpcid mpx rdseed adx smap clflushopt intel_pt xsaveopt xsavec xgetbv1 xsaves dtherm ida arat pln pts hwp hwp_notify hwp_act_window hwp_epp md_clear flush_l1d arch_capabilities
bugs		: cpu_meltdown spectre_v1 spectre_v2 spec_store_bypass l1tf mds swapgs taa itlb_multihit srbds mmio_stale_data retbleed
bogomips	: 3999.93
clflush size	: 64
cache_alignment	: 64
address sizes	: 39 bits physical, 48 bits virtual
power management:`)

	// Mock /proc/loadavg
	mockFS.SetFileContent("/proc/loadavg", "0.25 0.30 0.35 1/234 12345")

	// Mock /proc/stat
	mockFS.SetFileContent("/proc/stat", `cpu  123456 0 234567 890123 0 0 0 0 0 0
cpu0 30864 0 58641 222530 0 0 0 0 0 0
cpu1 30864 0 58641 222530 0 0 0 0 0 0
cpu2 30864 0 58641 222530 0 0 0 0 0 0
cpu3 30864 0 58641 222530 0 0 0 0 0 0
intr 1234567 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
ctxt 12345678
btime 1640995200
processes 12345
procs_running 2
procs_blocked 0
softirq 1234567 0 0 0 0 0 0 0 0 0 0`)

	// Mock /proc/meminfo
	mockFS.SetFileContent("/proc/meminfo", `MemTotal:       16384000 kB
MemFree:         8192000 kB
MemAvailable:   12288000 kB
Buffers:          512000 kB
Cached:          2048000 kB
SwapCached:            0 kB
Active:          4096000 kB
Inactive:        2048000 kB
Active(anon):    2048000 kB
Inactive(anon):   512000 kB
Active(file):    2048000 kB
Inactive(file):  1536000 kB
Unevictable:           0 kB
Mlocked:               0 kB
SwapTotal:       2097152 kB
SwapFree:        2097152 kB
Dirty:              1024 kB
Writeback:             0 kB
AnonPages:       2048000 kB
Mapped:           512000 kB
Shmem:            256000 kB
Slab:             256000 kB
SReclaimable:     128000 kB
SUnreclaim:       128000 kB
KernelStack:       16384 kB
PageTables:        32768 kB
NFS_Unstable:          0 kB
Bounce:                0 kB
WritebackTmp:          0 kB
CommitLimit:    10289152 kB
Committed_AS:    4194304 kB
VmallocTotal:   34359738367 kB
VmallocUsed:      262144 kB
VmallocChunk:   34359476223 kB
HardwareCorrupted:     0 kB
AnonHugePages:         0 kB
CmaTotal:              0 kB
CmaFree:               0 kB
HugePages_Total:       0
HugePages_Free:        0
HugePages_Rsvd:        0
HugePages_Surp:        0
Hugepagesize:       2048 kB
DirectMap4k:      262144 kB
DirectMap2M:     8126464 kB
DirectMap1G:     8388608 kB`)

	// Mock /proc/uptime
	mockFS.SetFileContent("/proc/uptime", "123456.78 98765.43")

	// Mock /proc/net/dev
	mockFS.SetFileContent("/proc/net/dev", `Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
    lo: 1234567    1234    0    0    0     0          0         0  1234567    1234    0    0    0     0       0          0
  eth0: 9876543210  987654    0    0    0     0          0         0 1234567890  123456    0    0    0     0       0          0
  wlan0: 5555555555  555555    0    0    0     0          0         0 3333333333  333333    0    0    0     0       0          0`)

	// Mock ip command responses
	mockCmd.SetResponse("ip", []string{"addr", "show", "eth0"}, []byte(`2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP group default qlen 1000
    link/ether 00:11:22:33:44:55 brd ff:ff:ff:ff:ff:ff
    inet 192.168.1.100/24 brd 192.168.1.255 scope global dynamic eth0
       valid_lft 86400sec preferred_lft 86400sec
    inet6 fe80::211:22ff:fe33:4455/64 scope link
       valid_lft forever preferred_lft forever`))

	mockCmd.SetResponse("ip", []string{"addr", "show", "wlan0"}, []byte(`3: wlan0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc mq state UP group default qlen 1000
    link/ether aa:bb:cc:dd:ee:ff brd ff:ff:ff:ff:ff:ff
    inet 192.168.1.101/24 brd 192.168.1.255 scope global dynamic wlan0
       valid_lft 86400sec preferred_lft 86400sec`))

	monitor := NewSystemMonitorWithDeps(mockFS, mockCmd)
	adapter := &SystemAdapter{
		api:     &MockAPI{},
		monitor: monitor,
	}

	return adapter, mockFS, mockCmd
}

// setupMockStorageAdapter creates a StorageAdapter with mocked dependencies
func setupMockStorageAdapter() (*StorageAdapter, *MockFileSystem, *MockCommandExecutor) {
	mockFS := NewMockFileSystem()
	mockCmd := NewMockCommandExecutor()

	// Mock /var/local/emhttp/var.ini (Unraid status file)
	mockFS.SetFileContent("/var/local/emhttp/var.ini", `mdState=STARTED
mdResync=0
mdResyncPos=0
mdResyncDt=0
sbState=STARTED
sbSynced=1
sbSyncExit=0
parity=1
parity2=1
diskCount=8
diskSpindownDelay=30
diskSpindownDelay2=30
diskSpindownDelay3=30
diskSpindownDelay4=30
diskSpindownDelay5=30
diskSpindownDelay6=30
diskSpindownDelay7=30
diskSpindownDelay8=30`)

	// Mock /proc/mdstat
	mockFS.SetFileContent("/proc/mdstat", `Personalities : [raid6] [raid5] [raid4] [linear] [multipath] [raid0] [raid1] [raid10]
md1 : active raid6 sdf1[5] sde1[4] sdd1[3] sdc1[2] sdb1[1] sda1[0]
      7813769216 blocks super 1.2 level 6, 512k chunk, algorithm 2 [6/6] [UUUUUU]

md2 : active raid1 sdf2[1] sda2[0]
      1048576 blocks super 1.2 [2/2] [UU]

unused devices: <none>`)

	// Mock lsblk command responses
	mockCmd.SetResponse("lsblk", []string{"-J", "-o", "NAME,SIZE,TYPE,MOUNTPOINT,MODEL"}, []byte(`{
   "blockdevices": [
      {"name": "sda", "size": "8T", "type": "disk", "mountpoint": null, "model": "WDC WD80EFAX-68LHPN0"},
      {"name": "sda1", "size": "7.3T", "type": "part", "mountpoint": "/mnt/disk1", "model": null},
      {"name": "sda2", "size": "1M", "type": "part", "mountpoint": null, "model": null},
      {"name": "sdb", "size": "8T", "type": "disk", "mountpoint": null, "model": "WDC WD80EFAX-68LHPN0"},
      {"name": "sdb1", "size": "7.3T", "type": "part", "mountpoint": "/mnt/disk2", "model": null},
      {"name": "sdc", "size": "8T", "type": "disk", "mountpoint": null, "model": "WDC WD80EFAX-68LHPN0"},
      {"name": "sdc1", "size": "7.3T", "type": "part", "mountpoint": "/mnt/disk3", "model": null}
   ]
}`))

	// Mock lsblk for real disks enumeration
	mockCmd.SetResponse("lsblk", []string{"-d", "-n", "-o", "NAME,SIZE,TYPE"}, []byte(`sda 8T disk
sdb 8T disk
sdc 8T disk
sdd 8T disk
sde 8T disk
sdf 8T disk`))

	// Mock smartctl commands
	mockCmd.SetResponse("smartctl", []string{"-H", "/dev/sda"}, []byte(`smartctl 7.2 2020-12-30 r5155 [x86_64-linux-6.1.0-unraid] (local build)
Copyright (C) 2002-20, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF READ SMART DATA SECTION ===
SMART overall-health self-assessment test result: PASSED`))

	mockCmd.SetResponse("smartctl", []string{"-A", "/dev/sda"}, []byte(`smartctl 7.2 2020-12-30 r5155 [x86_64-linux-6.1.0-unraid] (local build)
Copyright (C) 2002-20, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF READ SMART DATA SECTION ===
SMART Attributes Data Structure revision number: 16
Vendor Specific SMART Attributes with Thresholds:
ID# ATTRIBUTE_NAME          FLAGS    VALUE WORST THRESH FAIL RAW_VALUE
194 Temperature_Celsius     0x0022   032   032   000    -    32 (Min/Max 18/45)`))

	mockCmd.SetResponse("smartctl", []string{"-a", "/dev/sda"}, []byte(`smartctl 7.2 2020-12-30 r5155 [x86_64-linux-6.1.0-unraid] (local build)
Copyright (C) 2002-20, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF INFORMATION SECTION ===
Model Family:     Western Digital Red
Device Model:     WDC WD80EFAX-68LHPN0
Serial Number:    WD-WXA2A83XXXXX
LU WWN Device Id: 5 0014ee 2b5a5a5a5
Firmware Version: 83.H0A83
User Capacity:    8,001,563,222,016 bytes [8.00 TB]
Sector Sizes:     512 bytes logical, 4096 bytes physical
Rotation Rate:    5400 rpm
Form Factor:      3.5 inches
Device is:        In smartctl database [for details use: -P show]
ATA Version is:   ACS-2, ATA8-ACS T13/1699-D revision 4
SATA Version is:  SATA 3.0, 6.0 Gb/s (current: 6.0 Gb/s)
Local Time is:    Fri Jun 21 10:00:00 2024 UTC
SMART support is: Available - device has SMART capability.
SMART support is: Enabled

=== START OF READ SMART DATA SECTION ===
SMART overall-health self-assessment test result: PASSED`))

	// Mock ZFS commands
	mockCmd.SetResponse("zpool", []string{"list", "-H", "-o", "name,size,alloc,free,cap,health"}, []byte(`cache	1.82T	1.45T	384G	79%	ONLINE
backup	7.27T	5.12T	2.15T	70%	ONLINE`))

	mockCmd.SetResponse("zpool", []string{"status", "cache"}, []byte(`  pool: cache
 state: ONLINE
  scan: scrub repaired 0B in 02:15:30 with 0 errors on Sun Jun 16 02:39:31 2024
config:

	NAME        STATE     READ WRITE CKSUM
	cache       ONLINE       0     0     0
	  nvme0n1   ONLINE       0     0     0

errors: No known data errors`))

	// Mock df command for cache filesystem
	mockFS.SetFileExists("/mnt/cache", true)
	mockCmd.SetResponse("df", []string{"-h", "/mnt/cache"}, []byte(`Filesystem      Size  Used Avail Use% Mounted on
/dev/nvme0n1p1  1.8T  1.5T  384G  79% /mnt/cache`))

	monitor := NewStorageMonitorWithDeps(mockCmd, mockFS)
	adapter := &StorageAdapter{
		api:     &MockAPI{},
		monitor: monitor,
	}

	return adapter, mockFS, mockCmd
}

// setupMockVMAdapter creates a VMAdapter with mocked dependencies
func setupMockVMAdapter() (*VMAdapter, *MockCommandExecutor) {
	mockCmd := NewMockCommandExecutor()

	// Mock virsh list command
	mockCmd.SetResponse("virsh", []string{"list", "--all"}, []byte(` Id   Name               State
------------------------------------
 1    ubuntu-server      running
 -    windows-10         shut off
 -    test-vm            shut off`))

	// Mock virsh dominfo command
	mockCmd.SetResponse("virsh", []string{"dominfo", "ubuntu-server"}, []byte(`Id:             1
Name:           ubuntu-server
UUID:           550e8400-e29b-41d4-a716-446655440000
OS Type:        hvm
State:          running
CPU(s):         4
CPU time:       123.4s
Max memory:     4194304 KiB
Used memory:    4194304 KiB
Persistent:     yes
Autostart:      disable
Managed save:   no
Security model: none
Security DOI:   0`))

	// Mock virsh cpu-stats command
	mockCmd.SetResponse("virsh", []string{"cpu-stats", "ubuntu-server"}, []byte(`cpu_time          123456789000
vcpu_time         123456789000
user_time         45678901000
system_time       12345678000`))

	// Mock virsh dommemstat command
	mockCmd.SetResponse("virsh", []string{"dommemstat", "ubuntu-server"}, []byte(`actual 4194304
swap_in 0
swap_out 0
major_fault 1234
minor_fault 56789
unused 1048576
available 4194304
usable 3145728
last_update 1624275600
rss 3145728`))

	// Mock VM control commands
	mockCmd.SetResponse("virsh", []string{"start", "test-vm"}, []byte(`Domain test-vm started`))
	mockCmd.SetResponse("virsh", []string{"shutdown", "test-vm"}, []byte(`Domain test-vm is being shutdown`))
	mockCmd.SetResponse("virsh", []string{"destroy", "test-vm"}, []byte(`Domain test-vm destroyed`))

	monitor := NewVMMonitorWithDeps(mockCmd)
	adapter := &VMAdapter{
		api:     &MockAPI{},
		monitor: monitor,
	}

	return adapter, mockCmd
}

// TestAPIAdapter tests the APIAdapter functionality
func TestAPIAdapter(t *testing.T) {
	t.Run("NewAPIAdapter", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := NewAPIAdapter(mockAPI)

		if adapter == nil {
			t.Error("Expected non-nil APIAdapter")
			return
		}
		if adapter.api != mockAPI {
			t.Error("Expected adapter to store the provided API")
		}
	})

	t.Run("GetInfo", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := NewAPIAdapter(mockAPI)

		info := adapter.GetInfo()
		if info == nil {
			t.Error("Expected non-nil info")
		}

		infoMap, ok := info.(map[string]interface{})
		if !ok {
			t.Error("Expected info to be a map")
		}

		if infoMap["service"] != "UMA REST API" {
			t.Errorf("Expected service 'UMA REST API', got '%v'", infoMap["service"])
		}
		if infoMap["version"] != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got '%v'", infoMap["version"])
		}
		if infoMap["status"] != "running" {
			t.Errorf("Expected status 'running', got '%v'", infoMap["status"])
		}
	})

	t.Run("GetSystem", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := NewAPIAdapter(mockAPI)

		system := adapter.GetSystem()
		if system == nil {
			t.Error("Expected non-nil SystemInterface")
		}

		// Test that it returns a SystemAdapter
		systemAdapter, ok := system.(*SystemAdapter)
		if !ok {
			t.Error("Expected SystemAdapter")
		}
		if systemAdapter.api != mockAPI {
			t.Error("Expected SystemAdapter to have the correct API reference")
		}
	})

	t.Run("GetStorage", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := NewAPIAdapter(mockAPI)

		storage := adapter.GetStorage()
		if storage == nil {
			t.Error("Expected non-nil StorageInterface")
		}

		// Test that it returns a StorageAdapter
		storageAdapter, ok := storage.(*StorageAdapter)
		if !ok {
			t.Error("Expected StorageAdapter")
		}
		if storageAdapter.api != mockAPI {
			t.Error("Expected StorageAdapter to have the correct API reference")
		}
	})

	t.Run("GetDocker", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := NewAPIAdapter(mockAPI)

		docker := adapter.GetDocker()
		if docker == nil {
			t.Error("Expected non-nil DockerInterface")
		}

		// Test that it returns a DockerAdapter
		dockerAdapter, ok := docker.(*DockerAdapter)
		if !ok {
			t.Error("Expected DockerAdapter")
		}
		if dockerAdapter.api != mockAPI {
			t.Error("Expected DockerAdapter to have the correct API reference")
		}
	})

	t.Run("GetVM", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := NewAPIAdapter(mockAPI)

		vm := adapter.GetVM()
		if vm == nil {
			t.Error("Expected non-nil VMInterface")
		}

		// Test that it returns a VMAdapter
		vmAdapter, ok := vm.(*VMAdapter)
		if !ok {
			t.Error("Expected VMAdapter")
		}
		if vmAdapter.api != mockAPI {
			t.Error("Expected VMAdapter to have the correct API reference")
		}
	})

	t.Run("GetAuth", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := NewAPIAdapter(mockAPI)

		auth := adapter.GetAuth()
		if auth == nil {
			t.Error("Expected non-nil AuthInterface")
		}

		// Test that it returns an AuthAdapter
		authAdapter, ok := auth.(*AuthAdapter)
		if !ok {
			t.Error("Expected AuthAdapter")
		}
		if authAdapter.api != mockAPI {
			t.Error("Expected AuthAdapter to have the correct API reference")
		}
	})
}

// TestSystemAdapter tests the SystemAdapter functionality
func TestSystemAdapter(t *testing.T) {
	t.Run("GetCPUInfo", func(t *testing.T) {
		adapter, _, _ := setupMockSystemAdapter()

		info, err := adapter.GetCPUInfo()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if info == nil {
			t.Error("Expected non-nil CPU info")
		}

		// Validate CPU info structure
		if cpuMap, ok := info.(map[string]interface{}); ok {
			if cpuMap["model"] == "Unknown" {
				t.Error("Expected CPU model to be parsed from mock data")
			}
			if cpuMap["cores"] == 0 {
				t.Error("Expected CPU cores to be parsed from mock data")
			}
		}
	})

	t.Run("GetMemoryInfo", func(t *testing.T) {
		adapter, _, _ := setupMockSystemAdapter()

		info, err := adapter.GetMemoryInfo()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if info == nil {
			t.Error("Expected non-nil memory info")
		}

		// Validate memory info structure
		if memMap, ok := info.(map[string]interface{}); ok {
			if memMap["total"] == 0 {
				t.Error("Expected memory total to be parsed from mock data")
			}
		}
	})

	t.Run("GetLoadInfo", func(t *testing.T) {
		adapter, _, _ := setupMockSystemAdapter()

		info, err := adapter.GetLoadInfo()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if info == nil {
			t.Error("Expected non-nil load info")
		}
	})

	t.Run("GetUptimeInfo", func(t *testing.T) {
		adapter, _, _ := setupMockSystemAdapter()

		info, err := adapter.GetUptimeInfo()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if info == nil {
			t.Error("Expected non-nil uptime info")
		}
	})

	t.Run("GetNetworkInfo", func(t *testing.T) {
		adapter, _, _ := setupMockSystemAdapter()

		info, err := adapter.GetNetworkInfo()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if info == nil {
			t.Error("Expected non-nil network info")
		}
	})

	t.Run("GetEnhancedTemperatureData", func(t *testing.T) {
		adapter, _, _ := setupMockSystemAdapter()

		info, err := adapter.GetEnhancedTemperatureData()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if info == nil {
			t.Error("Expected non-nil temperature data")
		}
	})
}

// TestStorageAdapter tests the StorageAdapter functionality
func TestStorageAdapter(t *testing.T) {
	t.Run("GetArrayInfo", func(t *testing.T) {
		adapter, _, _ := setupMockStorageAdapter()

		info, err := adapter.GetArrayInfo()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if info == nil {
			t.Error("Expected non-nil array info")
		}

		// Validate array info structure
		if arrayMap, ok := info.(map[string]interface{}); ok {
			if arrayMap["state"] == "Unknown" {
				t.Error("Expected array state to be parsed from mock data")
			}
		}
	})

	t.Run("GetDisks", func(t *testing.T) {
		adapter, _, _ := setupMockStorageAdapter()

		disks, err := adapter.GetDisks()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if disks == nil {
			t.Error("Expected non-nil disks info")
		}

		// Validate disks structure
		if diskSlice, ok := disks.([]interface{}); ok {
			if len(diskSlice) == 0 {
				t.Error("Expected disks to be parsed from mock data")
			}
		}
	})

	t.Run("GetZFSPools", func(t *testing.T) {
		adapter, _, _ := setupMockStorageAdapter()

		pools, err := adapter.GetZFSPools()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if pools == nil {
			t.Error("Expected non-nil ZFS pools info")
		}
	})

	t.Run("GetCacheInfo", func(t *testing.T) {
		adapter, _, _ := setupMockStorageAdapter()

		cache, err := adapter.GetCacheInfo()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if cache == nil {
			t.Error("Expected non-nil cache info")
		}
	})

	t.Run("StartArray", func(t *testing.T) {
		adapter, _, _ := setupMockStorageAdapter()

		err := adapter.StartArray(map[string]interface{}{"force": false})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("StopArray", func(t *testing.T) {
		adapter, _, _ := setupMockStorageAdapter()

		err := adapter.StopArray(map[string]interface{}{"force": true})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

// TestDockerAdapter tests the DockerAdapter functionality
func TestDockerAdapter(t *testing.T) {
	t.Run("GetContainers", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &DockerAdapter{api: mockAPI}

		containers, err := adapter.GetContainers()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if containers == nil {
			t.Error("Expected non-nil containers")
		}
	})

	t.Run("GetContainer", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &DockerAdapter{api: mockAPI}

		container, err := adapter.GetContainer("test-id")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if container == nil {
			t.Error("Expected non-nil container")
		}

		containerMap, ok := container.(map[string]interface{})
		if !ok {
			t.Error("Expected container to be a map")
		}
		if containerMap["id"] != "test-id" {
			t.Errorf("Expected id 'test-id', got '%v'", containerMap["id"])
		}
	})

	t.Run("StartContainer", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &DockerAdapter{api: mockAPI}

		err := adapter.StartContainer("test-id")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("StopContainer", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &DockerAdapter{api: mockAPI}

		err := adapter.StopContainer("test-id")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("RestartContainer", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &DockerAdapter{api: mockAPI}

		err := adapter.RestartContainer("test-id")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("GetImages", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &DockerAdapter{api: mockAPI}

		images, err := adapter.GetImages()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if images == nil {
			t.Error("Expected non-nil images")
		}
	})

	t.Run("GetNetworks", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &DockerAdapter{api: mockAPI}

		networks, err := adapter.GetNetworks()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if networks == nil {
			t.Error("Expected non-nil networks")
		}
	})

	t.Run("GetSystemInfo", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &DockerAdapter{api: mockAPI}

		info, err := adapter.GetSystemInfo()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if info == nil {
			t.Error("Expected non-nil system info")
		}
	})
}

// TestVMAdapter tests the VMAdapter functionality
func TestVMAdapter(t *testing.T) {
	t.Run("GetVMs", func(t *testing.T) {
		adapter, _ := setupMockVMAdapter()

		vms, err := adapter.GetVMs()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if vms == nil {
			t.Error("Expected non-nil VMs")
		}

		// Validate VMs structure
		if vmSlice, ok := vms.([]interface{}); ok {
			if len(vmSlice) == 0 {
				t.Error("Expected VMs to be parsed from mock data")
			}
		}
	})

	t.Run("GetVM", func(t *testing.T) {
		adapter, _ := setupMockVMAdapter()

		vm, err := adapter.GetVM("ubuntu-server")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if vm == nil {
			t.Error("Expected non-nil VM")
		}

		vmMap, ok := vm.(map[string]interface{})
		if !ok {
			t.Error("Expected VM to be a map")
		}
		if vmMap["name"] != "ubuntu-server" {
			t.Errorf("Expected name 'ubuntu-server', got '%v'", vmMap["name"])
		}
	})

	t.Run("StartVM", func(t *testing.T) {
		adapter, _ := setupMockVMAdapter()

		err := adapter.StartVM("test-vm")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("StopVM", func(t *testing.T) {
		adapter, _ := setupMockVMAdapter()

		err := adapter.StopVM("test-vm")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("RestartVM", func(t *testing.T) {
		adapter, _ := setupMockVMAdapter()

		err := adapter.RestartVM("test-vm")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("GetVMStats", func(t *testing.T) {
		adapter, _ := setupMockVMAdapter()

		stats, err := adapter.GetVMStats("ubuntu-server")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if stats == nil {
			t.Error("Expected non-nil VM stats")
		}
	})

	t.Run("GetVMConsole", func(t *testing.T) {
		adapter, _ := setupMockVMAdapter()

		console, err := adapter.GetVMConsole("ubuntu-server")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if console == nil {
			t.Error("Expected non-nil VM console")
		}
	})

	t.Run("SetVMAutostart", func(t *testing.T) {
		adapter, _ := setupMockVMAdapter()

		err := adapter.SetVMAutostart("ubuntu-server", true)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

// TestAuthAdapter tests the AuthAdapter functionality
func TestAuthAdapter(t *testing.T) {
	t.Run("Login", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &AuthAdapter{api: mockAPI}

		result, err := adapter.Login("testuser", "testpass")
		// Expect error since authentication is not implemented in UMA
		if err == nil {
			t.Error("Expected error for unimplemented authentication")
		}
		if !strings.Contains(err.Error(), "authentication is not implemented") {
			t.Errorf("Expected 'authentication is not implemented' error, got %v", err)
		}
		if result != nil {
			t.Error("Expected nil result when authentication fails")
		}
	})

	t.Run("GetUsers", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &AuthAdapter{api: mockAPI}

		users, err := adapter.GetUsers()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if users == nil {
			t.Error("Expected non-nil users")
		}
	})

	t.Run("GetStats", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &AuthAdapter{api: mockAPI}

		stats, err := adapter.GetStats()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if stats == nil {
			t.Error("Expected non-nil auth stats")
		}
	})

	t.Run("IsEnabled", func(t *testing.T) {
		mockAPI := &MockAPI{}
		adapter := &AuthAdapter{api: mockAPI}

		enabled := adapter.IsEnabled()
		// Should return false for placeholder implementation
		if enabled {
			t.Error("Expected auth to be disabled in placeholder implementation")
		}
	})
}

// MockAPI provides a mock implementation for testing
type MockAPI struct{}

// GetDockerManager returns nil to simulate Docker manager unavailable
func (m *MockAPI) GetDockerManager() interface{} {
	return nil
}

// GetVMManager returns nil to simulate VM manager unavailable
func (m *MockAPI) GetVMManager() interface{} {
	return nil
}
