package vm

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/domalab/uma/daemon/lib"
	"github.com/domalab/uma/daemon/logger"
)

// VMManager provides virtual machine management capabilities
type VMManager struct{}

// VMInfo represents information about a virtual machine
type VMInfo struct {
	ID           int           `json:"id"`
	Name         string        `json:"name"`
	UUID         string        `json:"uuid"`
	State        string        `json:"state"`
	CPUs         int           `json:"cpus"`
	Memory       uint64        `json:"memory_kb"`
	Autostart    bool          `json:"autostart"`
	Persistent   bool          `json:"persistent"`
	OSType       string        `json:"os_type"`
	Architecture string        `json:"architecture"`
	Disks        []VMDisk      `json:"disks"`
	Networks     []VMNetwork   `json:"networks"`
	Graphics     []VMGraphics  `json:"graphics"`
	USBDevices   []VMUSBDevice `json:"usb_devices"`
	PCIDevices   []VMPCIDevice `json:"pci_devices"`
	CPUUsage     float64       `json:"cpu_usage_percent,omitempty"`
	MemoryUsage  uint64        `json:"memory_usage_kb,omitempty"`
}

// VMDisk represents a virtual machine disk
type VMDisk struct {
	Device string `json:"device"`
	Source string `json:"source"`
	Target string `json:"target"`
	Bus    string `json:"bus"`
	Type   string `json:"type"`
	Size   uint64 `json:"size_bytes,omitempty"`
}

// VMNetwork represents a virtual machine network interface
type VMNetwork struct {
	Type       string `json:"type"`
	Source     string `json:"source"`
	Model      string `json:"model"`
	MACAddress string `json:"mac_address"`
	Bridge     string `json:"bridge,omitempty"`
}

// VMGraphics represents virtual machine graphics configuration
type VMGraphics struct {
	Type     string `json:"type"`
	Port     int    `json:"port,omitempty"`
	Listen   string `json:"listen,omitempty"`
	Password string `json:"password,omitempty"`
}

// VMUSBDevice represents a USB device passed through to VM
type VMUSBDevice struct {
	Vendor  string `json:"vendor"`
	Product string `json:"product"`
	Bus     string `json:"bus,omitempty"`
	Device  string `json:"device,omitempty"`
}

// VMPCIDevice represents a PCI device passed through to VM
type VMPCIDevice struct {
	Domain   string `json:"domain"`
	Bus      string `json:"bus"`
	Slot     string `json:"slot"`
	Function string `json:"function"`
	Vendor   string `json:"vendor,omitempty"`
	Product  string `json:"product,omitempty"`
}

// VMStats represents VM statistics
type VMStats struct {
	VMID        int     `json:"vm_id"`
	Name        string  `json:"name"`
	CPUTime     uint64  `json:"cpu_time_ns"`
	CPUUsage    float64 `json:"cpu_usage_percent"`
	MemoryUsage uint64  `json:"memory_usage_kb"`
	MemoryTotal uint64  `json:"memory_total_kb"`
	DiskRead    uint64  `json:"disk_read_bytes"`
	DiskWrite   uint64  `json:"disk_write_bytes"`
	NetRx       uint64  `json:"network_rx_bytes"`
	NetTx       uint64  `json:"network_tx_bytes"`
}

// NewVMManager creates a new VM manager
func NewVMManager() *VMManager {
	return &VMManager{}
}

// IsLibvirtAvailable checks if libvirt is available
func (v *VMManager) IsLibvirtAvailable() bool {
	output := lib.GetCmdOutput("which", "virsh")
	if len(output) == 0 {
		return false
	}

	// Test connection to libvirt
	output = lib.GetCmdOutput("virsh", "version")
	return len(output) > 0 && !strings.Contains(strings.Join(output, ""), "failed to connect")
}

// ListVMs returns a list of all virtual machines
func (v *VMManager) ListVMs(all bool) ([]VMInfo, error) {
	vms := make([]VMInfo, 0)

	if !v.IsLibvirtAvailable() {
		return vms, fmt.Errorf("libvirt is not available")
	}

	args := []string{"list"}
	if all {
		args = append(args, "--all")
	}

	output := lib.GetCmdOutput("virsh", args...)

	// Parse virsh list output
	for i, line := range output {
		// Skip header lines
		if i < 2 || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		vm := VMInfo{}

		// Parse ID (might be "-" for inactive VMs)
		if fields[0] != "-" {
			if id, err := strconv.Atoi(fields[0]); err == nil {
				vm.ID = id
			}
		}

		// Parse name
		vm.Name = fields[1]

		// Parse state
		vm.State = strings.Join(fields[2:], " ")

		// Get detailed information
		if err := v.getVMDetails(&vm); err != nil {
			logger.Yellow("Failed to get VM details for %s: %v", vm.Name, err)
		}

		vms = append(vms, vm)
	}

	return vms, nil
}

// GetVM returns information about a specific virtual machine
func (v *VMManager) GetVM(name string) (*VMInfo, error) {
	if !v.IsLibvirtAvailable() {
		return nil, fmt.Errorf("libvirt is not available")
	}

	vm := &VMInfo{Name: name}
	if err := v.getVMDetails(vm); err != nil {
		return nil, err
	}

	return vm, nil
}

// StartVM starts a virtual machine
func (v *VMManager) StartVM(name string) error {
	if !v.IsLibvirtAvailable() {
		return fmt.Errorf("libvirt is not available")
	}

	output := lib.GetCmdOutput("virsh", "start", name)

	// Check for errors
	for _, line := range output {
		if strings.Contains(line, "error:") || strings.Contains(line, "failed") {
			return fmt.Errorf("failed to start VM: %s", line)
		}
	}

	logger.Blue("Started VM: %s", name)
	return nil
}

// StopVM stops a virtual machine
func (v *VMManager) StopVM(name string, force bool) error {
	if !v.IsLibvirtAvailable() {
		return fmt.Errorf("libvirt is not available")
	}

	command := "shutdown"
	if force {
		command = "destroy"
	}

	output := lib.GetCmdOutput("virsh", command, name)

	// Check for errors
	for _, line := range output {
		if strings.Contains(line, "error:") || strings.Contains(line, "failed") {
			return fmt.Errorf("failed to stop VM: %s", line)
		}
	}

	logger.Blue("Stopped VM: %s", name)
	return nil
}

// RestartVM restarts a virtual machine
func (v *VMManager) RestartVM(name string) error {
	if !v.IsLibvirtAvailable() {
		return fmt.Errorf("libvirt is not available")
	}

	output := lib.GetCmdOutput("virsh", "reboot", name)

	// Check for errors
	for _, line := range output {
		if strings.Contains(line, "error:") || strings.Contains(line, "failed") {
			return fmt.Errorf("failed to restart VM: %s", line)
		}
	}

	logger.Blue("Restarted VM: %s", name)
	return nil
}

// PauseVM pauses a virtual machine
func (v *VMManager) PauseVM(name string) error {
	if !v.IsLibvirtAvailable() {
		return fmt.Errorf("libvirt is not available")
	}

	output := lib.GetCmdOutput("virsh", "suspend", name)

	// Check for errors
	for _, line := range output {
		if strings.Contains(line, "error:") || strings.Contains(line, "failed") {
			return fmt.Errorf("failed to pause VM: %s", line)
		}
	}

	logger.Blue("Paused VM: %s", name)
	return nil
}

// ResumeVM resumes a paused virtual machine
func (v *VMManager) ResumeVM(name string) error {
	if !v.IsLibvirtAvailable() {
		return fmt.Errorf("libvirt is not available")
	}

	output := lib.GetCmdOutput("virsh", "resume", name)

	// Check for errors
	for _, line := range output {
		if strings.Contains(line, "error:") || strings.Contains(line, "failed") {
			return fmt.Errorf("failed to resume VM: %s", line)
		}
	}

	logger.Blue("Resumed VM: %s", name)
	return nil
}

// HibernateVM hibernates a virtual machine (save state to disk)
func (v *VMManager) HibernateVM(name string) error {
	if !v.IsLibvirtAvailable() {
		return fmt.Errorf("libvirt is not available")
	}

	// Use virsh save to hibernate the VM (save state to disk)
	saveFile := fmt.Sprintf("/tmp/%s.save", name)
	output := lib.GetCmdOutput("virsh", "save", name, saveFile)

	// Check for errors
	for _, line := range output {
		if strings.Contains(line, "error:") || strings.Contains(line, "failed") {
			return fmt.Errorf("failed to hibernate VM: %s", line)
		}
	}

	logger.Blue("Hibernated VM: %s (saved to %s)", name, saveFile)
	return nil
}

// RestoreVM restores a hibernated virtual machine
func (v *VMManager) RestoreVM(name string) error {
	if !v.IsLibvirtAvailable() {
		return fmt.Errorf("libvirt is not available")
	}

	// Use virsh restore to restore the VM from hibernation
	saveFile := fmt.Sprintf("/tmp/%s.save", name)
	output := lib.GetCmdOutput("virsh", "restore", saveFile)

	// Check for errors
	for _, line := range output {
		if strings.Contains(line, "error:") || strings.Contains(line, "failed") {
			return fmt.Errorf("failed to restore VM: %s", line)
		}
	}

	logger.Blue("Restored VM: %s (from %s)", name, saveFile)
	return nil
}

// GetVMStats returns statistics for a virtual machine
func (v *VMManager) GetVMStats(name string) (*VMStats, error) {
	if !v.IsLibvirtAvailable() {
		return nil, fmt.Errorf("libvirt is not available")
	}

	stats := &VMStats{Name: name}

	// Get CPU and memory stats
	output := lib.GetCmdOutput("virsh", "domstats", name)
	for _, line := range output {
		if strings.Contains(line, "cpu.time=") {
			if parts := strings.Split(line, "="); len(parts) == 2 {
				if cpuTime, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
					stats.CPUTime = cpuTime
				}
			}
		} else if strings.Contains(line, "balloon.current=") {
			if parts := strings.Split(line, "="); len(parts) == 2 {
				if memUsage, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
					stats.MemoryUsage = memUsage
				}
			}
		} else if strings.Contains(line, "balloon.maximum=") {
			if parts := strings.Split(line, "="); len(parts) == 2 {
				if memTotal, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
					stats.MemoryTotal = memTotal
				}
			}
		}
	}

	return stats, nil
}

// SetVMAutostart sets autostart for a virtual machine
func (v *VMManager) SetVMAutostart(name string, autostart bool) error {
	if !v.IsLibvirtAvailable() {
		return fmt.Errorf("libvirt is not available")
	}

	command := "autostart"
	if !autostart {
		command = "autostart --disable"
	}

	output := lib.GetCmdOutput("virsh", strings.Fields(command+" "+name)...)

	// Check for errors
	for _, line := range output {
		if strings.Contains(line, "error:") || strings.Contains(line, "failed") {
			return fmt.Errorf("failed to set autostart: %s", line)
		}
	}

	logger.Blue("Set autostart for VM %s: %t", name, autostart)
	return nil
}

// GetVMConsole returns console access information for a VM
func (v *VMManager) GetVMConsole(name string) (string, error) {
	if !v.IsLibvirtAvailable() {
		return "", fmt.Errorf("libvirt is not available")
	}

	// Get VNC display information
	output := lib.GetCmdOutput("virsh", "vncdisplay", name)
	if len(output) > 0 && !strings.Contains(output[0], "error") {
		return output[0], nil
	}

	return "", fmt.Errorf("no console available for VM: %s", name)
}

// getVMDetails gets detailed information about a VM
func (v *VMManager) getVMDetails(vm *VMInfo) error {
	// Get VM info
	output := lib.GetCmdOutput("virsh", "dominfo", vm.Name)
	for _, line := range output {
		if strings.Contains(line, "Id:") {
			if parts := strings.Fields(line); len(parts) >= 2 {
				if id, err := strconv.Atoi(parts[1]); err == nil {
					vm.ID = id
				}
			}
		} else if strings.Contains(line, "UUID:") {
			if parts := strings.Fields(line); len(parts) >= 2 {
				vm.UUID = parts[1]
			}
		} else if strings.Contains(line, "State:") {
			if parts := strings.Fields(line); len(parts) >= 2 {
				vm.State = strings.Join(parts[1:], " ")
			}
		} else if strings.Contains(line, "CPU(s):") {
			if parts := strings.Fields(line); len(parts) >= 2 {
				if cpus, err := strconv.Atoi(parts[1]); err == nil {
					vm.CPUs = cpus
				}
			}
		} else if strings.Contains(line, "Max memory:") {
			if parts := strings.Fields(line); len(parts) >= 3 {
				if memory, err := strconv.ParseUint(parts[2], 10, 64); err == nil {
					vm.Memory = memory
				}
			}
		} else if strings.Contains(line, "Autostart:") {
			if parts := strings.Fields(line); len(parts) >= 2 {
				vm.Autostart = (parts[1] == "enable")
			}
		} else if strings.Contains(line, "Persistent:") {
			if parts := strings.Fields(line); len(parts) >= 2 {
				vm.Persistent = (parts[1] == "yes")
			}
		}
	}

	// Get XML configuration for detailed parsing
	if err := v.parseVMXML(vm); err != nil {
		logger.Yellow("Failed to parse VM XML for %s: %v", vm.Name, err)
	}

	return nil
}

// parseVMXML parses VM XML configuration
func (v *VMManager) parseVMXML(vm *VMInfo) error {
	output := lib.GetCmdOutput("virsh", "dumpxml", vm.Name)
	if len(output) == 0 {
		return fmt.Errorf("failed to get VM XML")
	}

	xmlContent := strings.Join(output, "\n")

	// Parse basic domain information
	type Domain struct {
		Type string `xml:"type,attr"`
		OS   struct {
			Type struct {
				Arch    string `xml:"arch,attr"`
				Machine string `xml:"machine,attr"`
				Text    string `xml:",chardata"`
			} `xml:"type"`
		} `xml:"os"`
		Devices struct {
			Disks []struct {
				Type   string `xml:"type,attr"`
				Device string `xml:"device,attr"`
				Source struct {
					File string `xml:"file,attr"`
					Dev  string `xml:"dev,attr"`
				} `xml:"source"`
				Target struct {
					Dev string `xml:"dev,attr"`
					Bus string `xml:"bus,attr"`
				} `xml:"target"`
			} `xml:"disk"`
			Interfaces []struct {
				Type   string `xml:"type,attr"`
				Source struct {
					Bridge  string `xml:"bridge,attr"`
					Network string `xml:"network,attr"`
				} `xml:"source"`
				Model struct {
					Type string `xml:"type,attr"`
				} `xml:"model"`
				MAC struct {
					Address string `xml:"address,attr"`
				} `xml:"mac"`
			} `xml:"interface"`
			Graphics []struct {
				Type     string `xml:"type,attr"`
				Port     string `xml:"port,attr"`
				Listen   string `xml:"listen,attr"`
				Password string `xml:"passwd,attr"`
			} `xml:"graphics"`
			Hostdevs []struct {
				Mode   string `xml:"mode,attr"`
				Type   string `xml:"type,attr"`
				Source struct {
					Vendor struct {
						ID string `xml:"id,attr"`
					} `xml:"vendor"`
					Product struct {
						ID string `xml:"id,attr"`
					} `xml:"product"`
					Address struct {
						Domain   string `xml:"domain,attr"`
						Bus      string `xml:"bus,attr"`
						Slot     string `xml:"slot,attr"`
						Function string `xml:"function,attr"`
					} `xml:"address"`
				} `xml:"source"`
			} `xml:"hostdev"`
		} `xml:"devices"`
	}

	var domain Domain
	if err := xml.Unmarshal([]byte(xmlContent), &domain); err != nil {
		return err
	}

	// Extract information
	vm.OSType = domain.OS.Type.Text
	vm.Architecture = domain.OS.Type.Arch

	// Parse disks
	vm.Disks = make([]VMDisk, 0)
	for _, disk := range domain.Devices.Disks {
		vmDisk := VMDisk{
			Device: disk.Device,
			Target: disk.Target.Dev,
			Bus:    disk.Target.Bus,
			Type:   disk.Type,
		}

		if disk.Source.File != "" {
			vmDisk.Source = disk.Source.File
		} else if disk.Source.Dev != "" {
			vmDisk.Source = disk.Source.Dev
		}

		vm.Disks = append(vm.Disks, vmDisk)
	}

	// Parse networks
	vm.Networks = make([]VMNetwork, 0)
	for _, iface := range domain.Devices.Interfaces {
		vmNet := VMNetwork{
			Type:       iface.Type,
			Model:      iface.Model.Type,
			MACAddress: iface.MAC.Address,
		}

		if iface.Source.Bridge != "" {
			vmNet.Source = iface.Source.Bridge
			vmNet.Bridge = iface.Source.Bridge
		} else if iface.Source.Network != "" {
			vmNet.Source = iface.Source.Network
		}

		vm.Networks = append(vm.Networks, vmNet)
	}

	// Parse graphics
	vm.Graphics = make([]VMGraphics, 0)
	for _, graphics := range domain.Devices.Graphics {
		vmGraphics := VMGraphics{
			Type:     graphics.Type,
			Listen:   graphics.Listen,
			Password: graphics.Password,
		}

		if graphics.Port != "" {
			if port, err := strconv.Atoi(graphics.Port); err == nil {
				vmGraphics.Port = port
			}
		}

		vm.Graphics = append(vm.Graphics, vmGraphics)
	}

	// Parse hostdevs (USB and PCI passthrough)
	vm.USBDevices = make([]VMUSBDevice, 0)
	vm.PCIDevices = make([]VMPCIDevice, 0)

	for _, hostdev := range domain.Devices.Hostdevs {
		if hostdev.Type == "usb" {
			usbDev := VMUSBDevice{
				Vendor:  hostdev.Source.Vendor.ID,
				Product: hostdev.Source.Product.ID,
			}
			vm.USBDevices = append(vm.USBDevices, usbDev)
		} else if hostdev.Type == "pci" {
			pciDev := VMPCIDevice{
				Domain:   hostdev.Source.Address.Domain,
				Bus:      hostdev.Source.Address.Bus,
				Slot:     hostdev.Source.Address.Slot,
				Function: hostdev.Source.Address.Function,
			}
			vm.PCIDevices = append(vm.PCIDevices, pciDev)
		}
	}

	return nil
}
