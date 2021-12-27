// apcmetrics - APC UPS metrics exporter for Prometheus
//
// Copyright 2021 Nick Pillitteri
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/56quarters/apcmetrics/pkg/apcmetrics"
)

// Set by the build process: -ldflags="-X 'main.Version=xyz'"
var (
	Version  string
	Branch   string
	Revision string
)

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
	versionInfo := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "apcmetrics",
		Name:      "build_info",
		Help:      "APC Metrics version information",
		ConstLabels: prometheus.Labels{
			"version":   Version,
			"revision":  Revision,
			"branch":    Branch,
			"goversion": runtime.Version(),
		},
	}, func() float64 { return 1 })
	prometheus.MustRegister(versionInfo)
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
