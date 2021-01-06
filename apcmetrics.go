package main

import (
	"html/template"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Set by the build process: -ldflags="-X 'main.Version=xyz'"
var (
	Version  string
	Branch   string
	Revision string
)

const indexTpt = `
<!doctype html>
<html>
<head><title>APC UPS Metrics Exporter</title></head>
<body>
<h1>APC UPS Metrics Exporter</h1>
<p><a href="{{ . }}">Metrics</a></p>
</body>
</html>
`

var Log = setupLogger()

func setupLogger() *log.Logger {
	logger := log.New()
	logger.SetReportCaller(true)
	logger.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	return logger
}

func init() {
	// Set globals in the Prometheus version module based on our values
	// set by the build process to expose build information as a metric
	version.Version = Version
	version.Branch = Branch
	version.Revision = Revision
}

func main() {
	kp := kingpin.New(os.Args[0], "apcmetrics: APC UPS metrics exporter for Prometheus")
	metricsPath := kp.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	webAddr := kp.Flag("web.listen-address", "Address and port to expose Prometheus metrics on").Default(":9780").String()

	_, err := kp.Parse(os.Args[1:])
	if err != nil {
		Log.Fatal(err)
	}

	registry := prometheus.DefaultRegisterer
	versionInfo := version.NewCollector("apcmetrics")
	registry.MustRegister(versionInfo)

	index, err := template.New("index").Parse(indexTpt)
	if err != nil {
		Log.Fatal(err)
	}

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := index.Execute(w, *metricsPath); err != nil {
			Log.Errorf("Failed to render index: %s", err)
		}
	})
	Log.Error(http.ListenAndServe(*webAddr, nil))
}
