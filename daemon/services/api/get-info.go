package api

import (
	"net"
	"strings"

	"github.com/domalab/uma/daemon/common"
	"github.com/domalab/uma/daemon/dto"
	"github.com/domalab/uma/daemon/lib"
	"github.com/domalab/uma/daemon/logger"
	"gopkg.in/ini.v1"
)

func (a *Api) getInfo() *dto.Info {
	prefs := getPrefs()

	sensorReadings := a.sensor.GetReadings(prefs)
	upsReadings := a.ups.GetStatus()

	samples := append(sensorReadings, upsReadings...)

	return &dto.Info{
		Version:  2,
		Wake:     getMac(),
		Prefs:    prefs,
		Samples:  samples,
		Features: getFeatures(),
	}
}

func getMac() dto.Wake {
	wake := dto.Wake{
		Mac:       "",
		Broadcast: "255.255.255.255",
	}

	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		if iface.Name == "eth0" {
			wake.Mac = iface.HardwareAddr.String()
			break
		}
	}

	return wake
}

func getPrefs() dto.Prefs {
	prefs := dto.Prefs{
		Number: ".,",
		Unit:   "C",
	}

	cfg, err := ini.Load(common.Prefs)
	if err != nil {
		logger.Yellow("unable to load/parse prefs file (%s): %s", common.Prefs, err)
		return prefs
	}

	// Use improved parsing with automatic quote removal and defaults
	displaySection := cfg.Section("display")
	if displaySection != nil {
		number := displaySection.Key("number").MustString(".,")
		prefs.Number = strings.Replace(number, "\"", "", -1)

		unit := displaySection.Key("unit").MustString("C")
		prefs.Unit = strings.Replace(unit, "\"", "", -1)
	}

	return prefs
}

func getFeatures() map[string]bool {
	features := make(map[string]bool)

	// is sleep available ?
	exists, err := lib.Exists(common.Sleep)
	if err != nil {
		logger.Yellow("getfeatures:sleep:(%s)", err)
	}

	features["sleep"] = exists

	return features
}
