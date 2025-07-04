/* Copyright 2017, 2024 Juan B. Rodriguez
 * Copyright 2017, Derek Macias.
 * Copyright 2005-2016, Lime Technology
 * Copyright 2015, Dan Landon.
 * Copyright 2015, Bergware International.
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License version 2,
 * as published by the Free Software Foundation.
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 */
package ups

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/domalab/uma/daemon/dto"
	"github.com/domalab/uma/daemon/lib"
	"github.com/domalab/uma/daemon/logger"

	"gopkg.in/ini.v1"
)

const (
	nutBinary string = "/usr/bin/upsc"
	nutConfig string = "/boot/config/plugins/nut-dw/nut-dw.cfg"
)

// Nut -
type Nut struct {
	legend  map[string]string
	address string
}

// NewNut -
func NewNut() *Nut {
	nut := &Nut{}

	nut.legend = map[string]string{
		"OL":    "Online",
		"OB":    "On batt",
		"OL LB": "Online low batt",
		"OB LB": "Low batt",
	}

	var name, ip string

	cfg, err := ini.Load(nutConfig)
	if err != nil {
		name = "ups"
		ip = "127.0.0.1"
		logger.Yellow("nuts: unable to parse config: (%s). Using defaults ...", nutConfig)
	} else {
		// Use improved parsing with automatic defaults and quote removal
		name = cfg.Section("").Key("NAME").MustString("ups")
		name = strings.ReplaceAll(name, "\"", "")

		ip = cfg.Section("").Key("IPADDR").MustString("127.0.0.1")
		ip = strings.ReplaceAll(ip, "\"", "")

		if name == "ups" {
			logger.Yellow("nuts: name not found. Defaulting to ups ...")
		}
		if ip == "127.0.0.1" {
			logger.Yellow("nuts: ip not found. Defaulting to 127.0.0.1 ...")
		}
	}

	nut.address = fmt.Sprintf("%s@%s", name, ip)
	logger.Blue("nuts: using address: %s", nut.address)

	return nut
}

// GetStatus -
func (n *Nut) GetStatus() []dto.Sample {
	lines := lib.GetCmdOutput(nutBinary, n.address)
	for _, line := range lines {
		logger.Blue("nuts: line-> %s", line)
	}
	return n.Parse(lib.GetCmdOutput(nutBinary, n.address))
}

// Parse -
func (n *Nut) Parse(lines []string) []dto.Sample {
	samples := make([]dto.Sample, 0)

	var power float64
	var load float64

	for _, line := range lines {
		tokens := strings.Split(line, ":")

		key := strings.Trim(tokens[0], " ")
		val := strings.Trim(tokens[1], " ")

		switch key {
		case "ups.status":
			value, ok := n.legend[val]

			sample := dto.Sample{Key: "UPS STATUS"}

			if ok {
				sample.Value = value
				if strings.Index(strings.ToLower(value), "online") == 0 {
					sample.Condition = Green
				} else {
					sample.Condition = Red
				}
			} else {
				sample.Value = "Refreshing ..."
				sample.Condition = Orange
			}

			samples = append(samples, sample)

		case "battery.charge":
			sample := dto.Sample{Key: "UPS CHARGE", Value: val, Unit: "%"}

			charge, _ := strconv.ParseFloat(val, 64)
			if charge <= 10 {
				sample.Condition = Red
			} else {
				sample.Condition = Green
			}

			samples = append(samples, sample)

		case "battery.runtime":
			left, _ := strconv.ParseFloat(val, 64)
			runtime := strconv.FormatFloat(left/60, 'f', 1, 64)

			sample := dto.Sample{Key: "UPS LEFT", Value: runtime, Unit: "m"}

			if left <= 5 {
				sample.Condition = Red
			} else {
				sample.Condition = Green
			}

			samples = append(samples, sample)

		case "ups.realpower.nominal":
			power, _ = strconv.ParseFloat(val, 64)

		case "ups.load":
			sample := dto.Sample{Key: "UPS LOAD", Value: val, Unit: "%"}

			load, _ = strconv.ParseFloat(val, 64)
			if load >= 90 {
				sample.Condition = Red
			} else {
				sample.Condition = Green
			}

			samples = append(samples, sample)
		}
	}

	if power != 0 && load != 0 {
		value := power * load / 100
		watts := strconv.FormatFloat(value, 'f', 1, 64)

		sample := dto.Sample{Key: "UPS POWER", Value: watts, Unit: "w"}

		if load >= 90 {
			sample.Condition = Red
		} else {
			sample.Condition = Green
		}

		samples = append(samples, sample)
	}

	return samples
}
