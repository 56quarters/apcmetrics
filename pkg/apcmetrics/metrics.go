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
	"context"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

func NewApcCollector(client *ApcClient, timeout time.Duration, logger log.Logger) prometheus.Collector {
	return &apcCollector{
		client:  client,
		timeout: timeout,
		logger:  logger,

		// These descriptions mostly come from the apcupsd manual.
		// http://www.apcupsd.org/manual/manual.html#status-report-fields
		info: prometheus.NewDesc(
			"apc_info",
			"Info about the UPS",
			[]string{"hostname", "version", "ups_name", "model", "driver", "ups_mode"},
			nil,
		),
		status: prometheus.NewDesc(
			"apc_status",
			"Current status of the UPS",
			[]string{"status"},
			nil,
		),
		timeLeft: prometheus.NewDesc(
			"apc_time_left",
			"Remaining runtime left on the batteries in seconds",
			nil,
			nil,
		),
		loadPercent: prometheus.NewDesc(
			"apc_load_percent",
			"Percentage of load capacity",
			nil,
			nil,
		),
		chargePercent: prometheus.NewDesc(
			"apc_charge_percent",
			"Percentage of charge of the batteries",
			nil,
			nil,
		),
		lineVoltage: prometheus.NewDesc(
			"apc_line_voltage",
			"Current line voltage",
			nil,
			nil,
		),
		lowTransferVoltage: prometheus.NewDesc(
			"apc_low_transfer_voltage",
			"Line voltage below which the UPS will switch to batteries",
			nil,
			nil,
		),
		highTransferVoltage: prometheus.NewDesc(
			"apc_high_transfer_voltage",
			"Line voltage above which the UPS will switch to batteries",
			nil,
			nil,
		),
		batteryVoltage: prometheus.NewDesc(
			"apc_battery_voltage",
			"Battery voltage",
			nil,
			nil,
		),
		nominalBatteryVoltage: prometheus.NewDesc(
			"apc_nominal_battery_voltage",
			"Nominal battery voltage",
			nil,
			nil,
		),
		nominalInputVoltage: prometheus.NewDesc(
			"apc_nominal_input_voltage",
			"Nominal input voltage",
			nil,
			nil,
		),
		nominalWattage: prometheus.NewDesc(
			"apc_nominal_wattage",
			"Max power the UPS is designed to supply",
			nil,
			nil,
		),
		batteryDate: prometheus.NewDesc(
			"apc_battery_date",
			"Date the batteries were last replaced as a UNIX timestamp",
			nil,
			nil,
		),
		lastTimeOnBattery: prometheus.NewDesc(
			"apc_last_time_on_battery",
			"Last transfer on to batteries as a UNIX timestamp",
			nil,
			nil,
		),
		lastTimeOffBattery: prometheus.NewDesc(
			"apc_last_time_off_battery",
			"Last transfer off of batteries as a UNIX timestamp",
			nil,
			nil,
		),
		lastSelfTest: prometheus.NewDesc(
			"apc_last_self_test",
			"Last self test as a UNIX timestamp",
			nil,
			nil,
		),
	}
}

type apcCollector struct {
	client  *ApcClient
	timeout time.Duration
	logger  log.Logger

	info                  *prometheus.Desc
	status                *prometheus.Desc
	timeLeft              *prometheus.Desc
	loadPercent           *prometheus.Desc
	chargePercent         *prometheus.Desc
	lineVoltage           *prometheus.Desc
	lowTransferVoltage    *prometheus.Desc
	highTransferVoltage   *prometheus.Desc
	batteryVoltage        *prometheus.Desc
	nominalBatteryVoltage *prometheus.Desc
	nominalInputVoltage   *prometheus.Desc
	nominalWattage        *prometheus.Desc
	batteryDate           *prometheus.Desc
	lastTimeOnBattery     *prometheus.Desc
	lastTimeOffBattery    *prometheus.Desc
	lastSelfTest          *prometheus.Desc
}

func (a *apcCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- a.info
	ch <- a.status
	ch <- a.timeLeft
	ch <- a.loadPercent
	ch <- a.chargePercent
	ch <- a.lineVoltage
	ch <- a.lowTransferVoltage
	ch <- a.highTransferVoltage
	ch <- a.batteryVoltage
	ch <- a.nominalBatteryVoltage
	ch <- a.nominalInputVoltage
	ch <- a.nominalWattage
	ch <- a.batteryDate
	ch <- a.lastTimeOnBattery
	ch <- a.lastTimeOffBattery
	ch <- a.lastSelfTest
}

func (a *apcCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	status, err := a.client.Status(ctx)
	if err != nil {
		level.Error(a.logger).Log("msg", "unable to determine UPS status", "err", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		a.info,
		prometheus.GaugeValue,
		1,
		status.Hostname,
		status.Version,
		status.UpsName,
		status.Model,
		status.Driver,
		status.UpsMode,
	)

	ch <- prometheus.MustNewConstMetric(a.status, prometheus.GaugeValue, 1, status.Status)
	ch <- prometheus.MustNewConstMetric(a.timeLeft, prometheus.GaugeValue, status.TimeLeft.Seconds())
	ch <- prometheus.MustNewConstMetric(a.loadPercent, prometheus.GaugeValue, float64(status.LoadPercent))
	ch <- prometheus.MustNewConstMetric(a.chargePercent, prometheus.GaugeValue, float64(status.ChargePercent))
	ch <- prometheus.MustNewConstMetric(a.lineVoltage, prometheus.GaugeValue, float64(status.LineVoltage))
	ch <- prometheus.MustNewConstMetric(a.lowTransferVoltage, prometheus.GaugeValue, float64(status.LowTransferVoltage))
	ch <- prometheus.MustNewConstMetric(a.highTransferVoltage, prometheus.GaugeValue, float64(status.HighTransferVoltage))
	ch <- prometheus.MustNewConstMetric(a.batteryVoltage, prometheus.GaugeValue, float64(status.BatteryVoltage))
	ch <- prometheus.MustNewConstMetric(a.nominalBatteryVoltage, prometheus.GaugeValue, float64(status.NominalBatteryVoltage))
	ch <- prometheus.MustNewConstMetric(a.nominalInputVoltage, prometheus.GaugeValue, float64(status.NominalInputVoltage))
	ch <- prometheus.MustNewConstMetric(a.nominalWattage, prometheus.GaugeValue, float64(status.NominalWattage))
	ch <- prometheus.MustNewConstMetric(a.batteryDate, prometheus.GaugeValue, float64(status.BatteryDate.Unix()))
	ch <- prometheus.MustNewConstMetric(a.lastTimeOnBattery, prometheus.GaugeValue, float64(status.LastTimeOnBattery.Unix()))
	ch <- prometheus.MustNewConstMetric(a.lastTimeOffBattery, prometheus.GaugeValue, float64(status.LastTimeOffBattery.Unix()))
	ch <- prometheus.MustNewConstMetric(a.lastSelfTest, prometheus.GaugeValue, float64(status.LastSelfTest.Unix()))
}
