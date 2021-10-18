// apcmetrics - APC UPS metrics exporter for Prometheus
//
// Copyright 2021 Nick Pillitteri
//
// Licensed under the Apache License, Version 2.0 <LICENSE-APACHE or
// http://www.apache.org/licenses/LICENSE-2.0> or the MIT license
// <LICENSE-MIT or http://opensource.org/licenses/MIT>, at your
// option. This file may not be copied, modified, or distributed
// except according to those terms.

package apcmetrics

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Percent float64
type Voltage float64
type Wattage float64

type ApcStatus struct {
	Hostname string `json:"hostname"`
	Version  string `json:"version"`
	UpsName  string `json:"ups_name"`
	Model    string `json:"model"`
	Driver   string `json:"driver"`
	UpsMode  string `json:"ups_mode"`

	Status                string        `json:"status"`
	TimeLeft              time.Duration `json:"time_left"`
	LoadPercent           Percent       `json:"load_percent"`
	ChargePercent         Percent       `json:"charge_percent"`
	LineVoltage           Voltage       `json:"line_voltage"`
	LowTransferVoltage    Voltage       `json:"low_transfer_voltage"`
	HighTransferVoltage   Voltage       `json:"high_transfer_voltage"`
	BatteryVoltage        Voltage       `json:"battery_voltage"`
	NominalBatteryVoltage Voltage       `json:"nominal_battery_voltage"`
	NominalInputVoltage   Voltage       `json:"nominal_input_voltage"`
	NominalWattage        Wattage       `json:"nominal_wattage"`

	BatteryDate        time.Time `json:"battery_date"`
	LastTimeOnBattery  time.Time `json:"last_time_on_battery"`
	LastTimeOffBattery time.Time `json:"last_time_off_battery"`
	LastSelfTest       time.Time `json:"last_self_test"`
}

func ParseStatusFromLines(lines []string) (*ApcStatus, error) {
	kvs := parseLines(lines)
	status := &ApcStatus{}

	if v, ok := kvs["HOSTNAME"]; ok {
		status.Hostname = v
	}

	if v, ok := kvs["VERSION"]; ok {
		status.Version = v
	}

	if v, ok := kvs["UPSNAME"]; ok {
		status.UpsName = v
	}

	if v, ok := kvs["MODEL"]; ok {
		status.Model = v
	}

	if v, ok := kvs["DRIVER"]; ok {
		status.Driver = v
	}

	if v, ok := kvs["UPSMODE"]; ok {
		status.UpsMode = v
	}

	if v, ok := kvs["STATUS"]; ok {
		status.Status = v
	}

	if v, ok := kvs["TIMELEFT"]; ok {
		parsed, err := parseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("unable to parse TIMELEFT %s: %w", v, err)
		}

		status.TimeLeft = parsed
	}

	if v, ok := kvs["LOADPCT"]; ok {
		parsed, err := parseFloatAndUnit(v)
		if err != nil {
			return nil, fmt.Errorf("unable to parse LOADPCT %s: %w", v, err)
		}

		status.LoadPercent = Percent(parsed)
	}

	if v, ok := kvs["BCHARGE"]; ok {
		parsed, err := parseFloatAndUnit(v)
		if err != nil {
			return nil, fmt.Errorf("unable to parse BCHARGE %s: %w", v, err)
		}

		status.ChargePercent = Percent(parsed)
	}

	if v, ok := kvs["LINEV"]; ok {
		parsed, err := parseFloatAndUnit(v)
		if err != nil {
			return nil, fmt.Errorf("unable to parse LINEV %s: %w", v, err)
		}

		status.LineVoltage = Voltage(parsed)
	}

	if v, ok := kvs["LOTRANS"]; ok {
		parsed, err := parseFloatAndUnit(v)
		if err != nil {
			return nil, fmt.Errorf("unable to parse LOTRANS %s: %w", v, err)
		}

		status.LowTransferVoltage = Voltage(parsed)
	}

	if v, ok := kvs["HITRANS"]; ok {
		parsed, err := parseFloatAndUnit(v)
		if err != nil {
			return nil, fmt.Errorf("unable to parse HITRANS %s: %w", v, err)
		}

		status.HighTransferVoltage = Voltage(parsed)
	}

	if v, ok := kvs["BATTV"]; ok {
		parsed, err := parseFloatAndUnit(v)
		if err != nil {
			return nil, fmt.Errorf("unable to parse BATTV %s: %w", v, err)
		}

		status.BatteryVoltage = Voltage(parsed)
	}

	if v, ok := kvs["NOMBATTV"]; ok {
		parsed, err := parseFloatAndUnit(v)
		if err != nil {
			return nil, fmt.Errorf("unable to parse NOMBATTV %s: %w", v, err)
		}

		status.NominalBatteryVoltage = Voltage(parsed)
	}

	if v, ok := kvs["NOMINV"]; ok {
		parsed, err := parseFloatAndUnit(v)
		if err != nil {
			return nil, fmt.Errorf("unable to parse NOMINV %s: %w", v, err)
		}

		status.NominalInputVoltage = Voltage(parsed)
	}

	if v, ok := kvs["NOMPOWER"]; ok {
		parsed, err := parseFloatAndUnit(v)
		if err != nil {
			return nil, fmt.Errorf("unable to parse NOMPOWER %s: %w", v, err)
		}

		status.NominalWattage = Wattage(parsed)
	}

	if v, ok := kvs["BATTDATE"]; ok {
		parsed, err := parseDate(v)
		if err != nil {
			return nil, fmt.Errorf("unable to parse BATTDATE %s: %w", v, err)
		}

		status.BatteryDate = parsed
	}

	if v, ok := kvs["XONBATT"]; ok {
		parsed, err := parseDateTime(v)
		if err != nil {
			return nil, fmt.Errorf("unable to parse XONBATT %s: %w", v, err)
		}

		status.LastTimeOnBattery = parsed
	}

	if v, ok := kvs["XOFFBATT"]; ok {
		parsed, err := parseDateTime(v)
		if err != nil {
			return nil, fmt.Errorf("unable to parse XOFFBATT %s: %w", v, err)
		}

		status.LastTimeOffBattery = parsed
	}

	if v, ok := kvs["LASTSTEST"]; ok {
		parsed, err := parseDateTime(v)
		if err != nil {
			return nil, fmt.Errorf("unable to parse LASTSTEST %s: %w", v, err)
		}

		status.LastSelfTest = parsed
	}

	return status, nil
}

func parseLines(lines []string) map[string]string {
	out := make(map[string]string)

	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		out[key] = val
	}

	return out
}

func parseFloatAndUnit(raw string) (float64, error) {
	parts := strings.Split(raw, " ")
	if len(parts) != 2 {
		return 0.0, errors.New("expected two parts")
	}

	res, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0.0, err
	}

	return res, nil
}

func parseDuration(raw string) (time.Duration, error) {
	raw = strings.ToLower(raw)

	// use a float for scale since any values will also be floats, thus
	// we want to convert to an int64 (what time.Duration is) after we've
	// done any required conversions to avoid truncation.
	var scale float64
	if strings.Contains(raw, "minute") {
		scale = float64(time.Minute)
	} else if strings.Contains(raw, "second") {
		scale = float64(time.Second)
	} else {
		return 0, errors.New("unexpected unit")
	}

	parts := strings.Split(raw, " ")
	if len(parts) != 2 {
		return 0.0, errors.New("expected two parts")
	}

	res, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, err
	}

	return time.Duration(res * scale), nil
}

func parseDate(raw string) (time.Time, error) {
	return time.Parse("2006-01-02", raw)
}

func parseDateTime(raw string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05 -0700", raw)
}

type ApcEvent struct {
	TimeStamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

func ParseEventsFromLines(lines []string) ([]ApcEvent, error) {
	out := make([]ApcEvent, 0, len(lines))

	for _, line := range lines {
		// timestamp is separated from message with two spaces
		parts := strings.Split(line, "  ")
		if len(parts) != 2 {
			return nil, errors.New("expected two parts")
		}

		timestamp := strings.TrimSpace(parts[0])
		message := strings.TrimSpace(parts[1])

		ts, err := parseDateTime(timestamp)
		if err != nil {
			return nil, err
		}

		out = append(out, ApcEvent{
			TimeStamp: ts,
			Message:   message,
		})
	}

	return out, nil
}
