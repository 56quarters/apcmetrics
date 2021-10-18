// apcmetrics - APC UPS metrics exporter for Prometheus
//
// Copyright 2021 Nick Pillitteri
//
// Licensed under the Apache License, Version 2.0 <LICENSE-APACHE or
// http://www.apache.org/licenses/LICENSE-2.0> or the MIT license
// <LICENSE-MIT or http://opensource.org/licenses/MIT>, at your
// option. This file may not be copied, modified, or distributed
// except according to those terms.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/56quarters/apcmetrics/pkg/apcmetrics"
)

// Set by the build process: -ldflags="-X 'main.Version=xyz'"
var (
	Version  string
	Branch   string
	Revision string
)

func init() {
	// Set globals in the Prometheus version module based on our values
	// set by the build process to expose build information as a metric
	version.Version = Version
	version.Branch = Branch
	version.Revision = Revision
}

func setupLogger(l level.Option) log.Logger {
	logger := log.NewSyncLogger(log.NewLogfmtLogger(os.Stderr))
	logger = level.NewFilter(logger, l)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	return logger
}

func main() {
	logger := setupLogger(level.AllowInfo())

	kp := kingpin.New(os.Args[0], "apcmetrics: APC UPS metrics exporter for Prometheus")
	upsAddress := kp.Flag("ups.address", "Address and port of the apcupsd daemon to connect to").Default("localhost:3551").String()
	upsTimeout := kp.Flag("ups.timeout", "Max time reads from the apcupsd daemon may take").Default("5s").Duration()

	metrics := kp.Command("metrics", "Export Prometheus metrics via HTTP")
	metricsPath := metrics.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	metricsAddress := metrics.Flag("web.listen-address", "Address and port to expose Prometheus metrics on").Default(":9780").String()

	status := kp.Command("status", "Display the current status of the UPS as JSON")
	statusRaw := status.Flag("raw", "Output the unparsed status response from apcupsd").Default("false").Bool()

	events := kp.Command("events", "Display recent UPS events as JSON")
	eventsRaw := events.Flag("raw", "Output the unparsed events response from apcupsd").Default("false").Bool()

	command, err := kp.Parse(os.Args[1:])
	if err != nil {
		level.Error(logger).Log("msg", "failed to parse CLI options", "err", err)
		os.Exit(1)
	}

	client := apcmetrics.NewApcClient(*upsAddress, logger)

	switch command {
	case metrics.FullCommand():
		if err := serveMetrics(client, logger, *upsTimeout, *metricsPath, *metricsAddress); err != nil {
			level.Error(logger).Log("msg", "unable to serve UPS metrics", "err", err)
			os.Exit(1)
		}
	case status.FullCommand():
		if err := showStatus(client, *upsTimeout, *statusRaw); err != nil {
			level.Error(logger).Log("msg", "unable to get UPS status", "err", err)
			os.Exit(1)
		}
	case events.FullCommand():
		if err := showEvents(client, *upsTimeout, *eventsRaw); err != nil {
			level.Error(logger).Log("msg", "unable to get UPS events", "err", err)
			os.Exit(1)
		}
	}
}

func serveMetrics(client *apcmetrics.ApcClient, logger log.Logger, upsTimeout time.Duration, metricsPath string, metricsAddress string) error {
	prometheus.MustRegister(version.NewCollector("apcmetrics"))
	prometheus.MustRegister(apcmetrics.NewApcCollector(client, upsTimeout, logger))

	http.Handle(metricsPath, promhttp.Handler())
	level.Info(logger).Log("msg", "serving Prometheus metrics", "path", metricsPath, "address", metricsAddress)
	if err := http.ListenAndServe(metricsAddress, nil); err != nil {
		return err
	}

	return nil
}

func showStatus(client *apcmetrics.ApcClient, upsTimeout time.Duration, raw bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), upsTimeout)
	defer cancel()

	var output string
	if raw {
		lines, err := client.StatusRaw(ctx)
		if err != nil {
			return err
		}

		output = strings.Join(lines, "\n")
	} else {
		status, err := client.Status(ctx)
		if err != nil {
			return err
		}

		bytes, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			return err
		}

		output = string(bytes)
	}

	fmt.Println(output)
	return nil
}

func showEvents(client *apcmetrics.ApcClient, upsTimeout time.Duration, raw bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), upsTimeout)
	defer cancel()

	var output string
	if raw {
		lines, err := client.EventsRaw(ctx)
		if err != nil {
			return err
		}

		output = strings.Join(lines, "\n")
	} else {
		events, err := client.Events(ctx)
		if err != nil {
			return err
		}

		bytes, err := json.MarshalIndent(events, "", "  ")
		if err != nil {
			return err
		}

		output = string(bytes)
	}

	fmt.Println(output)
	return nil
}
