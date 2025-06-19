package mocks

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// SystemMockManager manages all system-level mocks for cross-platform testing
type SystemMockManager struct {
	tempDir     string
	originalEnv map[string]string
	mockFiles   map[string]string
}

// NewSystemMockManager creates a new system mock manager
func NewSystemMockManager() *SystemMockManager {
	return &SystemMockManager{
		originalEnv: make(map[string]string),
		mockFiles:   make(map[string]string),
	}
}

// Setup initializes all system mocks for cross-platform testing
func (m *SystemMockManager) Setup() error {
	// Create temporary directory for mock files
	tempDir, err := os.MkdirTemp("", "uma_test_mocks_")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	m.tempDir = tempDir

	// Setup mock files based on the operating system
	if runtime.GOOS != "linux" {
		if err := m.setupLinuxMocks(); err != nil {
			return fmt.Errorf("failed to setup Linux mocks: %v", err)
		}
	}

	return nil
}

// Cleanup removes all mock files and restores environment
func (m *SystemMockManager) Cleanup() error {
	// Restore original environment variables
	for key, value := range m.originalEnv {
		if value == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}

	// Remove temporary directory
	if m.tempDir != "" {
		return os.RemoveAll(m.tempDir)
	}

	return nil
}

// setupLinuxMocks creates mock Linux-specific files and directories
func (m *SystemMockManager) setupLinuxMocks() error {
	// Create mock /proc directory structure
	procDir := filepath.Join(m.tempDir, "proc")
	if err := os.MkdirAll(procDir, 0755); err != nil {
		return err
	}

	// Create mock /boot directory structure
	bootDir := filepath.Join(m.tempDir, "boot", "config")
	if err := os.MkdirAll(bootDir, 0755); err != nil {
		return err
	}

	// Setup mock /proc files
	if err := m.createMockProcFiles(procDir); err != nil {
		return err
	}

	// Setup mock /boot/config files
	if err := m.createMockBootFiles(bootDir); err != nil {
		return err
	}

	// Setup environment variables to point to mock directories
	m.setEnv("UMA_MOCK_PROC_DIR", procDir)
	m.setEnv("UMA_MOCK_BOOT_DIR", filepath.Join(m.tempDir, "boot"))
	m.setEnv("UMA_TESTING_MODE", "true")

	return nil
}

// createMockProcFiles creates mock /proc filesystem files
func (m *SystemMockManager) createMockProcFiles(procDir string) error {
	mockFiles := map[string]string{
		"cpuinfo": m.getMockCPUInfo(),
		"stat":    m.getMockStat(),
		"meminfo": m.getMockMemInfo(),
		"loadavg": m.getMockLoadAvg(),
		"uptime":  m.getMockUptime(),
		"mdstat":  m.getMockMDStat(),
		"version": m.getMockVersion(),
	}

	for filename, content := range mockFiles {
		filePath := filepath.Join(procDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create mock %s: %v", filename, err)
		}
		m.mockFiles[filename] = filePath
	}

	return nil
}

// createMockBootFiles creates mock /boot/config files
func (m *SystemMockManager) createMockBootFiles(bootDir string) error {
	mockFiles := map[string]string{
		"parity-checks.log":           m.getMockParityChecksLog(),
		"plugins/dynamix/dynamix.cfg": m.getMockDynamixConfig(),
		"network.cfg":                 m.getMockNetworkConfig(),
		"disk.cfg":                    m.getMockDiskConfig(),
	}

	for filename, content := range mockFiles {
		filePath := filepath.Join(bootDir, filename)

		// Create directory if it doesn't exist
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}

		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create mock %s: %v", filename, err)
		}
		m.mockFiles[filename] = filePath
	}

	return nil
}

// setEnv sets an environment variable and stores the original value
func (m *SystemMockManager) setEnv(key, value string) {
	m.originalEnv[key] = os.Getenv(key)
	os.Setenv(key, value)
}

// GetMockPath returns the path to a mock file
func (m *SystemMockManager) GetMockPath(filename string) string {
	if path, exists := m.mockFiles[filename]; exists {
		return path
	}
	return ""
}

// GetTempDir returns the temporary directory used for mocks
func (m *SystemMockManager) GetTempDir() string {
	return m.tempDir
}

// Mock file content generators
func (m *SystemMockManager) getMockCPUInfo() string {
	return `processor	: 0
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
power management:

processor	: 1
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
core id		: 1
cpu cores	: 4
apicid		: 2
initial apicid	: 2
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
power management:
`
}

func (m *SystemMockManager) getMockStat() string {
	return `cpu  123456 0 234567 8901234 12345 0 6789 0 0 0
cpu0 30864 0 58641 2225308 3086 0 1697 0 0 0
cpu1 30864 0 58642 2225309 3087 0 1698 0 0 0
cpu2 30864 0 58643 2225310 3088 0 1699 0 0 0
cpu3 30864 0 58641 2225307 3084 0 1695 0 0 0
intr 12345678 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
ctxt 23456789
btime 1671234567
processes 12345
procs_running 2
procs_blocked 0
softirq 3456789 0 1234567 0 234567 345678 0 456789 567890 0 678901
`
}

func (m *SystemMockManager) getMockMemInfo() string {
	return `MemTotal:       16384000 kB
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
SwapTotal:       4194304 kB
SwapFree:        4194304 kB
Dirty:             64000 kB
Writeback:             0 kB
AnonPages:       2048000 kB
Mapped:           512000 kB
Shmem:            256000 kB
KReclaimable:     256000 kB
Slab:             512000 kB
SReclaimable:     256000 kB
SUnreclaim:       256000 kB
KernelStack:       32000 kB
PageTables:        64000 kB
NFS_Unstable:          0 kB
Bounce:                0 kB
WritebackTmp:          0 kB
CommitLimit:    12386304 kB
Committed_AS:    4096000 kB
VmallocTotal:   34359738367 kB
VmallocUsed:      128000 kB
VmallocChunk:          0 kB
Percpu:            16384 kB
HardwareCorrupted:     0 kB
AnonHugePages:         0 kB
ShmemHugePages:        0 kB
ShmemPmdMapped:        0 kB
FileHugePages:         0 kB
FilePmdMapped:         0 kB
HugePages_Total:       0
HugePages_Free:        0
HugePages_Rsvd:        0
HugePages_Surp:        0
Hugepagesize:       2048 kB
Hugetlb:               0 kB
DirectMap4k:      524288 kB
DirectMap2M:    15728640 kB
DirectMap1G:           0 kB
`
}

func (m *SystemMockManager) getMockLoadAvg() string {
	return "1.25 1.50 1.75 2/345 12345\n"
}

func (m *SystemMockManager) getMockUptime() string {
	return "123456.78 987654.32\n"
}

func (m *SystemMockManager) getMockMDStat() string {
	return `Personalities : [raid1] [raid6] [raid5] [raid4] [linear] [multipath] [raid0] [raid10]
md1 : active raid1 sdb1[1] sda1[0]
      1048576 blocks super 1.2 [2/2] [UU]
      
md2 : active raid5 sdd1[3] sdc1[2] sdb2[1] sda2[0]
      8388608 blocks super 1.2 level 5, 512k chunk, algorithm 2 [4/4] [UUUU]
      
unused devices: <none>
`
}

func (m *SystemMockManager) getMockVersion() string {
	return "Linux version 5.19.17-Unraid (root@Develop) (gcc (GCC) 11.2.0, GNU ld (GNU Binutils) 2.37) #1 SMP PREEMPT_DYNAMIC Mon Oct 10 14:36:09 PDT 2022\n"
}

func (m *SystemMockManager) getMockParityChecksLog() string {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	lastWeek := now.AddDate(0, 0, -7)

	return fmt.Sprintf(`%d|0|28800|89478485|0
%d|0|28900|89123456|0
%d|0|29000|88765432|0
`, lastWeek.Unix(), yesterday.Unix(), now.Unix())
}

func (m *SystemMockManager) getMockDynamixConfig() string {
	return `[parity]
mode="1"
schedule="40 3 1 * *"
priority="normal"

[system]
timezone="America/New_York"
language="en_US"

[network]
hostname="tower"
domain="local"
`
}

func (m *SystemMockManager) getMockNetworkConfig() string {
	return `HOSTNAME="tower"
DOMAIN="local"
DHCP_KEEPRESOLV="yes"
USE_DHCP="yes"
IPADDR="192.168.1.100"
NETMASK="255.255.255.0"
GATEWAY="192.168.1.1"
DNS_SERVER1="8.8.8.8"
DNS_SERVER2="8.8.4.4"
`
}

func (m *SystemMockManager) getMockDiskConfig() string {
	return `diskP1="/dev/sda"
diskP2="/dev/sdb"
disk1="/dev/sdc"
disk2="/dev/sdd"
disk3="/dev/sde"
cache="/dev/nvme0n1"
`
}

// MockSensorsCommand provides mock output for the sensors command
func MockSensorsCommand() string {
	return `coretemp-isa-0000
Adapter: ISA adapter
Package id 0:  +45.0°C  (high = +100.0°C, crit = +100.0°C)
Core 0:        +42.0°C  (high = +100.0°C, crit = +100.0°C)
Core 1:        +43.0°C  (high = +100.0°C, crit = +100.0°C)
Core 2:        +44.0°C  (high = +100.0°C, crit = +100.0°C)
Core 3:        +45.0°C  (high = +100.0°C, crit = +100.0°C)

acpi-0
Adapter: ACPI interface
temp1:        +42.0°C  (crit = +100.0°C)

it8772-isa-0a30
Adapter: ISA adapter
in0:          +1.02 V  (min =  +0.00 V, max =  +4.08 V)
in1:          +1.02 V  (min =  +0.00 V, max =  +4.08 V)
in2:          +3.34 V  (min =  +0.00 V, max =  +4.08 V)
+5V:          +5.02 V  (min =  +0.00 V, max =  +6.85 V)
in4:          +3.02 V  (min =  +0.00 V, max =  +4.08 V)
in5:          +1.67 V  (min =  +0.00 V, max =  +4.08 V)
in6:          +1.34 V  (min =  +0.00 V, max =  +4.08 V)
3VSB:         +3.34 V  (min =  +0.00 V, max =  +4.08 V)
Vbat:         +3.18 V  (min =  +0.00 V, max =  +4.08 V)
fan1:        1200 RPM  (min =    0 RPM)
fan2:        1150 RPM  (min =    0 RPM)
fan3:        1100 RPM  (min =    0 RPM)
temp1:        +42.0°C  (low  = +127.0°C, high = +127.0°C)  sensor = thermistor
temp2:        +45.0°C  (low  = +127.0°C, high = +127.0°C)  sensor = thermistor
temp3:        +43.0°C  (low  = +127.0°C, high = +127.0°C)  sensor = thermistor
`
}

// IsMockingEnabled returns true if we're in testing mode with mocks enabled
func IsMockingEnabled() bool {
	return os.Getenv("UMA_TESTING_MODE") == "true"
}

// GetMockProcPath returns the mock /proc directory path if mocking is enabled
func GetMockProcPath() string {
	if IsMockingEnabled() {
		return os.Getenv("UMA_MOCK_PROC_DIR")
	}
	return "/proc"
}

// GetMockBootPath returns the mock /boot directory path if mocking is enabled
func GetMockBootPath() string {
	if IsMockingEnabled() {
		return os.Getenv("UMA_MOCK_BOOT_DIR")
	}
	return "/boot"
}
