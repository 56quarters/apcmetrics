package main

import "fmt"

// Set by the build process: -ldflags="-X 'main.Version=xyz'"
var (
	Version  string
	Branch   string
	Revision string
)

const indexTpt = `
<!doctype html>
<html>
<head><title>APC Metrics Exporter</title></head>
<body>
<h1>APC Metrics Exporter</h1>
<p><a href="{{ . }}">Metrics</a></p>
</body>
</html>
`

func init() {
	// Set globals in the Prometheus version module based on our values
	// set by the build process to expose build information as a metric
	version.Version = Version
	version.Branch = Branch
	version.Revision = Revision
}

func main() {
	fmt.Println("HELLO!")
}
