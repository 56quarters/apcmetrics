package apcmetrics

import (
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

	Status              string        `json:"status"`
	TimeLeft            time.Duration `json:"time_left"`
	LoadPercent         Percent       `json:"load_percent"`
	ChargePercent       Percent       `json:"charge_percent"`
	LineVoltage         Voltage       `json:"line_voltage"`
	LowTransferVoltage  Voltage       `json:"low_transfer_voltage"`
	HighTransferVoltage Voltage       `json:"high_transfer_voltage"`
	BatteryVoltage      Voltage       `json:"battery_voltage"`
	NominalVoltage      Voltage       `json:"nominal_voltage"`
	NominalWattage      Wattage       `json:"nominal_wattage"`

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

	if v, ok := kvs["STATUS"]; ok {
		status.Status = v
	}

	if v, ok := kvs["TIMELEFT"]; ok {
		if parsed, err := parseDuration(v); err != nil {
			return nil, fmt.Errorf("unable to parse TIMELEFT %s: %w", v, err)
		} else {
			status.TimeLeft = parsed
		}
	}

	if v, ok := kvs["LOADPCT"]; ok {
		if parsed, err := parsePercent(v); err != nil {
			return nil, fmt.Errorf("unable to parse LOADPCT %s: %w", v, err)
		} else {
			status.LoadPercent = parsed
		}
	}

	if v, ok := kvs["BCHARGE"]; ok {
		if parsed, err := parsePercent(v); err != nil {
			return nil, fmt.Errorf("unable to parse BCHARGE %s: %w", v, err)
		} else {
			status.ChargePercent = parsed
		}
	}

	if v, ok := kvs["LINEV"]; ok {
		if parsed, err := parseVoltage(v); err != nil {
			return nil, fmt.Errorf("unable to parse LINEV %s: %w", v, err)
		} else {
			status.LineVoltage = parsed
		}
	}

	if v, ok := kvs["LOTRANS"]; ok {
		if parsed, err := parseVoltage(v); err != nil {
			return nil, fmt.Errorf("unable to parse LOTRANS %s: %w", v, err)
		} else {
			status.LowTransferVoltage = parsed
		}
	}

	if v, ok := kvs["HITRANS"]; ok {
		if parsed, err := parseVoltage(v); err != nil {
			return nil, fmt.Errorf("unable to parse HITRANS %s: %w", v, err)
		} else {
			status.HighTransferVoltage = parsed
		}
	}

	if v, ok := kvs["BATTV"]; ok {
		if parsed, err := parseVoltage(v); err != nil {
			return nil, fmt.Errorf("unable to parse BATTV %s: %w", v, err)
		} else {
			status.BatteryVoltage = parsed
		}
	}

	if v, ok := kvs["NOMBATTV"]; ok {
		if parsed, err := parseVoltage(v); err != nil {
			return nil, fmt.Errorf("unable to parse NOMBATTV %s: %w", v, err)
		} else {
			status.NominalVoltage = parsed
		}
	}

	if v, ok := kvs["NOMPOWER"]; ok {
		if parsed, err := parseWatts(v); err != nil {
			return nil, fmt.Errorf("unable to parse NOMPOWER %s: %w", v, err)
		} else {
			status.NominalWattage = parsed
		}
	}

	if v, ok := kvs["BATTDATE"]; ok {
		if parsed, err := parseDate(v); err != nil {
			return nil, fmt.Errorf("unable to parse BATTDATE %s: %w", v, err)
		} else {
			status.BatteryDate = parsed
		}
	}

	if v, ok := kvs["XONBATT"]; ok {
		if parsed, err := parseDateTime(v); err != nil {
			return nil, fmt.Errorf("unable to parse XONBATT %s: %w", v, err)
		} else {
			status.LastTimeOnBattery = parsed
		}
	}

	if v, ok := kvs["XOFFBATT"]; ok {
		if parsed, err := parseDateTime(v); err != nil {
			return nil, fmt.Errorf("unable to parse XOFFBATT %s: %w", v, err)
		} else {
			status.LastTimeOffBattery = parsed
		}
	}

	if v, ok := kvs["LASTSTEST"]; ok {
		if parsed, err := parseDateTime(v); err != nil {
			return nil, fmt.Errorf("unable to parse LASTSTEST %s: %w", v, err)
		} else {
			status.LastSelfTest = parsed
		}
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

func parsePercent(raw string) (Percent, error) {
	parts := strings.Split(raw, " ")
	if len(parts) != 2 {
		return 0.0, fmt.Errorf("unable parse %s as percent", raw)
	}

	res, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0.0, fmt.Errorf("unable parse %s as percent: %w", parts[0], err)
	}

	return Percent(res), nil
}

func parseVoltage(raw string) (Voltage, error) {
	parts := strings.Split(raw, " ")
	if len(parts) != 2 {
		return 0.0, fmt.Errorf("unable parse %s as voltage", raw)
	}

	res, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0.0, fmt.Errorf("unable parse %s as voltage: %w", parts[0], err)
	}

	return Voltage(res), nil
}

func parseWatts(raw string) (Wattage, error) {
	parts := strings.Split(raw, " ")
	if len(parts) != 2 {
		return 0.0, fmt.Errorf("unable parse %s as wattage", raw)
	}

	res, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0.0, fmt.Errorf("unable parse %s as wattage: %w", parts[0], err)
	}

	return Wattage(res), nil
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
		return 0, fmt.Errorf("unable to parse %s as duration: unexpected unit", raw)
	}

	parts := strings.Split(raw, " ")
	if len(parts) != 2 {
		return 0.0, fmt.Errorf("unable parse %s as duration", raw)
	}

	res, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("unable to parse %s as duration: %w", raw, err)
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
			return nil, fmt.Errorf("unable to parse %s as event", line)
		}

		timestamp := strings.TrimSpace(parts[0])
		message := strings.TrimSpace(parts[1])

		ts, err := parseDateTime(timestamp)
		if err != nil {
			return nil, fmt.Errorf("unable to parse %s as timestamp: %w", timestamp, err)
		}

		out = append(out, ApcEvent{
			TimeStamp: ts,
			Message:   message,
		})
	}

	return out, nil
}
