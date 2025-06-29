/* Copyright 2017, Juan B. Rodriguez
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
	"strconv"
	"strings"

	"github.com/domalab/uma/daemon/dto"
	"github.com/domalab/uma/daemon/lib"
)

const apcBinary string = "/sbin/apcaccess"

// Apc - apcupsd
type Apc struct {
	legend map[string]string
}

// NewApc -
func NewApc() *Apc {
	apc := &Apc{}

	apc.legend = map[string]string{
		"TRIM ONLINE":  "Online (trim)",
		"BOOST ONLINE": "Online (boost)",
		"ONLINE":       "Online",
		"ONBATT":       "On battery",
		"COMMLOST":     "Lost comm",
		"NOBATT":       "No battery",
		"Minutes":      "m",
		"Hours":        "h",
		"Seconds":      "s",
	}

	return apc
}

// GetStatus -
func (a *Apc) GetStatus() []dto.Sample {
	return a.Parse(lib.GetCmdOutput(apcBinary))
}

// Parse -
func (a *Apc) Parse(lines []string) []dto.Sample {
	samples := make([]dto.Sample, 0)

	var power float64
	var load float64

	for _, line := range lines {
		tokens := strings.Split(line, ":")

		key := strings.Trim(tokens[0], " ")
		val := strings.Trim(tokens[1], " ")

		switch key {
		case "STATUS":
			value, ok := a.legend[val]

			sample := dto.Sample{Key: "UPS STATUS"}

			if ok {
				sample.Value = value
				if strings.Index(strings.ToLower(val), "online") == 0 {
					sample.Condition = Green
				} else {
					sample.Condition = Red
				}
			} else {
				sample.Value = "Refreshing ..."
				sample.Condition = Orange
			}

			samples = append(samples, sample)

		case "BCHARGE":
			text := strings.Fields(val)
			charge, _ := strconv.ParseFloat(text[0], 64)

			sample := dto.Sample{Key: "UPS CHARGE", Value: text[0], Unit: "%"}

			if charge <= 10 {
				sample.Condition = Red
			} else {
				sample.Condition = Green
			}

			samples = append(samples, sample)

		case "TIMELEFT":
			text := strings.Fields(val)
			left, _ := strconv.ParseFloat(text[0], 64)
			unit, ok := a.legend[strings.Trim(text[1], " ")]
			if !ok {
				unit = "m"
			}

			sample := dto.Sample{Key: "UPS LEFT", Value: text[0], Unit: unit}

			if left <= 5 {
				sample.Condition = Red
			} else {
				sample.Condition = Green
			}

			samples = append(samples, sample)

		case "NOMPOWER":
			text := strings.Fields(val)
			power, _ = strconv.ParseFloat(text[0], 64)

		case "LOADPCT":
			text := strings.Fields(val)
			load, _ = strconv.ParseFloat(text[0], 64)

			sample := dto.Sample{Key: "UPS LOAD", Value: text[0], Unit: "%"}

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
