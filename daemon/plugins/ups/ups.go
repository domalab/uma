package ups

import (
	"github.com/domalab/uma/daemon/dto"
	"github.com/domalab/uma/daemon/lib"
)

// Kind -
type Kind int

// DNE -
const (
	DNE Kind = iota // does not exist
	APC
	NUT
	USB_HID // USB HID UPS devices
	SNMP    // SNMP-based UPS devices
)

// Green -
const (
	Green  = "green"
	Red    = "red"
	Orange = "orange"
)

// IdentifyUps - Enhanced UPS detection with support for multiple UPS types
func IdentifyUps() (Kind, error) {
	// Check for NUT (Network UPS Tools) - highest priority
	exists, err := lib.Exists("/var/run/nut/upsmon.pid")
	if err != nil {
		return DNE, err
	}
	if exists {
		return NUT, nil
	}

	// Check for alternative NUT locations
	exists, err = lib.Exists("/var/run/upsmon.pid")
	if err != nil {
		return DNE, err
	}
	if exists {
		return NUT, nil
	}

	// Check for APC UPS daemon
	exists, err = lib.Exists("/var/run/apcupsd.pid")
	if err != nil {
		return DNE, err
	}
	if exists {
		return APC, nil
	}

	// Check for USB HID UPS devices
	exists, err = lib.Exists("/dev/usb/hiddev0")
	if err != nil {
		return DNE, err
	}
	if exists {
		return USB_HID, nil
	}

	// Check for additional USB UPS devices
	exists, err = lib.Exists("/sys/class/power_supply/UPS")
	if err != nil {
		return DNE, err
	}
	if exists {
		return USB_HID, nil
	}

	return DNE, nil
}

// Ups -
type Ups interface {
	GetStatus() []dto.Sample
}
