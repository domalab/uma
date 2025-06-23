package ups

import (
	"os/exec"
	"sync"
	"time"

	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/plugins/ups"
)

// DetectionStatus represents the current UPS detection status
type DetectionStatus struct {
	Available bool      `json:"available"`
	Type      ups.Kind  `json:"type"`
	LastCheck time.Time `json:"last_check"`
	Error     string    `json:"error,omitempty"`
}

// StatusChangeCallback is called when UPS detection status changes
type StatusChangeCallback func(available bool, upsType ups.Kind)

// Detector handles automatic UPS detection and monitoring
type Detector struct {
	mu            sync.RWMutex
	status        DetectionStatus
	stopCh        chan struct{}
	checkInterval time.Duration
	callbacks     []StatusChangeCallback
}

// NewDetector creates a new UPS detector
func NewDetector() *Detector {
	return &Detector{
		status: DetectionStatus{
			Available: false,
			Type:      ups.DNE,
			LastCheck: time.Now(),
		},
		stopCh:        make(chan struct{}),
		checkInterval: 30 * time.Second, // Check every 30 seconds
		callbacks:     make([]StatusChangeCallback, 0),
	}
}

// Start begins periodic UPS detection
func (d *Detector) Start() {
	logger.Blue("Starting UPS auto-detection service...")

	// Perform initial detection
	d.detectUPS()

	// Start periodic detection
	go d.periodicDetection()
}

// Stop stops the UPS detection service
func (d *Detector) Stop() {
	logger.Blue("Stopping UPS auto-detection service...")
	close(d.stopCh)
}

// GetStatus returns the current UPS detection status
func (d *Detector) GetStatus() DetectionStatus {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.status
}

// IsAvailable returns true if a UPS is currently detected and available
func (d *Detector) IsAvailable() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.status.Available
}

// GetUPSType returns the detected UPS type
func (d *Detector) GetUPSType() ups.Kind {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.status.Type
}

// AddStatusChangeCallback adds a callback to be called when UPS status changes
func (d *Detector) AddStatusChangeCallback(callback StatusChangeCallback) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.callbacks = append(d.callbacks, callback)
}

// periodicDetection runs UPS detection checks periodically
func (d *Detector) periodicDetection() {
	ticker := time.NewTicker(d.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.detectUPS()
		case <-d.stopCh:
			return
		}
	}
}

// detectUPS performs UPS detection and updates status
func (d *Detector) detectUPS() {
	d.mu.Lock()

	// Store previous status for comparison
	prevAvailable := d.status.Available
	prevType := d.status.Type

	d.status.LastCheck = time.Now()
	d.status.Error = ""

	// Try to identify UPS using existing logic
	upsType, err := ups.IdentifyUps()
	if err != nil {
		d.status.Available = false
		d.status.Type = ups.DNE
		d.status.Error = err.Error()
		logger.Yellow("UPS detection failed: %v", err)
		d.mu.Unlock()
		d.notifyStatusChange(prevAvailable, prevType)
		return
	}

	if upsType == ups.DNE {
		d.status.Available = false
		d.status.Type = ups.DNE
		logger.Blue("No UPS detected")
		d.mu.Unlock()
		d.notifyStatusChange(prevAvailable, prevType)
		return
	}

	// Additional validation - try to communicate with the UPS
	if d.validateUPSCommunication(upsType) {
		d.status.Available = true
		d.status.Type = upsType
		logger.Green("UPS detected and validated: %s", d.getUPSTypeName(upsType))
	} else {
		d.status.Available = false
		d.status.Type = ups.DNE
		d.status.Error = "UPS detected but communication failed"
		logger.Yellow("UPS detected but communication validation failed")
	}

	d.mu.Unlock()
	d.notifyStatusChange(prevAvailable, prevType)
}

// notifyStatusChange calls all registered callbacks if status changed
func (d *Detector) notifyStatusChange(prevAvailable bool, prevType ups.Kind) {
	d.mu.RLock()
	currentAvailable := d.status.Available
	currentType := d.status.Type
	callbacks := make([]StatusChangeCallback, len(d.callbacks))
	copy(callbacks, d.callbacks)
	d.mu.RUnlock()

	// Only notify if status actually changed
	if prevAvailable != currentAvailable || prevType != currentType {
		logger.Blue("UPS status changed: available=%t, type=%s", currentAvailable, d.getUPSTypeName(currentType))
		for _, callback := range callbacks {
			go callback(currentAvailable, currentType)
		}
	}
}

// validateUPSCommunication validates that we can actually communicate with the detected UPS
func (d *Detector) validateUPSCommunication(upsType ups.Kind) bool {
	switch upsType {
	case ups.APC:
		return d.validateAPCCommunication()
	case ups.NUT:
		return d.validateNUTCommunication()
	case ups.USB_HID:
		return d.validateUSBHIDCommunication()
	case ups.SNMP:
		return d.validateSNMPCommunication()
	default:
		return false
	}
}

// validateAPCCommunication checks if apcaccess command works
func (d *Detector) validateAPCCommunication() bool {
	cmd := exec.Command("apcaccess")
	output, err := cmd.Output()
	if err != nil {
		logger.Yellow("apcaccess command failed: %v", err)
		return false
	}

	// Check if output contains expected UPS data
	outputStr := string(output)
	return len(outputStr) > 0 &&
		(contains(outputStr, "STATUS") ||
			contains(outputStr, "BCHARGE") ||
			contains(outputStr, "TIMELEFT"))
}

// validateNUTCommunication checks if upsc command works
func (d *Detector) validateNUTCommunication() bool {
	// Try default UPS name first
	cmd := exec.Command("upsc", "ups")
	output, err := cmd.Output()
	if err != nil {
		// Try to list available UPS devices
		listCmd := exec.Command("upsc", "-l")
		listOutput, listErr := listCmd.Output()
		if listErr != nil {
			logger.Yellow("upsc command failed: %v", err)
			return false
		}

		// If we can list devices, try the first one
		if len(listOutput) > 0 {
			return true
		}
		return false
	}

	// Check if output contains expected UPS data
	outputStr := string(output)
	return len(outputStr) > 0 &&
		(contains(outputStr, "ups.status") ||
			contains(outputStr, "battery.charge") ||
			contains(outputStr, "ups.load"))
}

// validateUSBHIDCommunication checks if USB HID UPS communication works
func (d *Detector) validateUSBHIDCommunication() bool {
	// Check if we can read from USB HID device
	cmd := exec.Command("cat", "/sys/class/power_supply/UPS/status")
	output, err := cmd.Output()
	if err != nil {
		// Try alternative method - check for hiddev devices
		cmd = exec.Command("ls", "/dev/usb/hiddev*")
		_, err = cmd.Output()
		return err == nil
	}

	// Check if output contains expected UPS status
	outputStr := string(output)
	return len(outputStr) > 0 &&
		(contains(outputStr, "Discharging") ||
			contains(outputStr, "Charging") ||
			contains(outputStr, "Full") ||
			contains(outputStr, "Not charging"))
}

// validateSNMPCommunication checks if SNMP UPS communication works
func (d *Detector) validateSNMPCommunication() bool {
	// This would require SNMP tools and UPS configuration
	// For now, return false as it requires specific setup
	return false
}

// getUPSTypeName returns a human-readable name for the UPS type
func (d *Detector) getUPSTypeName(upsType ups.Kind) string {
	switch upsType {
	case ups.APC:
		return "APC UPS (apcupsd)"
	case ups.NUT:
		return "Network UPS Tools (NUT)"
	case ups.USB_HID:
		return "USB HID UPS"
	case ups.SNMP:
		return "SNMP UPS"
	default:
		return "Unknown"
	}
}

// contains is a simple string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					indexOfSubstring(s, substr) >= 0)))
}

// indexOfSubstring finds the index of a substring
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
