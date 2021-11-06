# APC Metrics

[![build status](https://circleci.com/gh/56quarters/apcmetrics.svg?style=shield)](https://circleci.com/gh/56quarters/apcmetrics)

Prometheus exporter for APC UPSes controlled by [apcupsd](http://www.apcupsd.org/)

## Features

* Export metrics about your APC UPS such as runtime remaining, battery charge, current load, etc.
* Inspect the current status of your APC UPS using `apcmetrics status`
* Inspect recent events for your APC UPS using `apcmetrics events`

## Building

To build from source you'll need Go 1.16 installed.

```
git clone git@github.com:56quarters/apcmetrics.git && cd apcmetrics
make build
```

The `apcmetrics` binary will then be in the root of the checkout.

## Install

At the moment, `apcmetrics` is GNU/Linux specific. As such, these instructions assume a
GNU/Linux system.

To install `apcmetrics` after building as described above:

* Copy the binary to `/usr/local/bin`

```
sudo cp apcmetrics /usr/local/bin/
```

* Copy and enable the Systemd unit

```
sudo cp ext/apcmetrics.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable apcmetrics.service
```

* Start the daemon

```
sudo systemctl start apcmetrics.service
```

## Usage

The primary purpose of `apcmetrics` is to export metrics about an APC UPS to Prometheus. However,
it can also be used to inspect the current status of your UPS or recent events (such as power outages
and self-tests). Examples of how to do each are given below

### Metrics export

`apcmetrics` is must connect to `apcupsd` to export metrics. It defaults to collecting metrics from
a locally running `apcupsd` daemon using the default port of `3551`. If `apcupsd` is not running
locally, you can supply the address using the `--ups.address` CLI flag.

Prometheus metrics are exposed on port `9780` at `/metrics` by default. Once `apcmetrics`
is running, configure scrapes of it by your Prometheus server. Add the host running
`apcmetrocs` as a target under the Prometheus `scrape_configs` section as described by
the example below.

```yaml
# Sample config for Prometheus.

global:
  scrape_interval:     15s
  evaluation_interval: 15s
  external_labels:
      monitor: 'my_prom'

scrape_configs:
  - job_name: apcmetrics
    static_configs:
      - targets: ['example:9780']
```

### `apcmetrics status`

Running `apcmetrics status` will display the current status of the APC UPS as JSON. It defaults to
collecting  metrics from  a locally running `apcupsd` daemon using the default port of `3551`. If
`apcupsd` is not running  locally, you can supply the address using the `--ups.address` CLI flag. 
An example is given below.

```
$ apcmetrics status
{
  "hostname": "example",
  "version": "3.14.14 (31 May 2016) debian",
  "ups_name": "example",
  "model": "Back-UPS XS 1500 LCD",
  "driver": "USB UPS Driver",
  "ups_mode": "Stand Alone",
  "status": "ONLINE",
  "time_left": 4320000000000,
  "load_percent": 6,
  "charge_percent": 82,
  "line_voltage": 120,
  "low_transfer_voltage": 88,
  "high_transfer_voltage": 139,
  "battery_voltage": 25.2,
  "nominal_battery_voltage": 24,
  "nominal_input_voltage": 120,
  "nominal_wattage": 865,
  "battery_date": "2013-07-15T00:00:00Z",
  "last_time_on_battery": "2021-11-06T15:39:29-04:00",
  "last_time_off_battery": "2021-11-06T15:40:23-04:00",
  "last_self_test": "0001-01-01T00:00:00Z"
}
```

### `apcmetrics events`

Running `apcmetrics events` will display the last few events recorded by the APC UPS as JSON.
It defaults to collecting  metrics from  a locally running `apcupsd` daemon using the default
port of `3551`. If `apcupsd` is not running  locally, you can supply the address using the 
`--ups.address` CLI flag. An example is given below.

```
$ apcmetrics events
[
  {
    "timestamp": "2021-10-31T19:28:19-04:00",
    "message": "UPS Self Test switch to battery."
  },
  {
    "timestamp": "2021-10-31T19:28:28-04:00",
    "message": "UPS Self Test completed: Battery OK"
  },
  {
    "timestamp": "2021-11-06T15:39:29-04:00",
    "message": "Power failure."
  },
  {
    "timestamp": "2021-11-06T15:39:35-04:00",
    "message": "Running on UPS batteries."
  },
  {
    "timestamp": "2021-11-06T15:40:23-04:00",
    "message": "Mains returned. No longer on UPS batteries."
  },
  {
    "timestamp": "2021-11-06T15:40:23-04:00",
    "message": "Power is back. UPS running on mains."
  }
]
```

## Development

To build a binary:

```
make build
```

To build a tagged release binary:

```
make build-dist
```

To build a Docker image

```
make image
```

To build a tagged release Docker image

```
make image-dist
```

To run tests

```
make test
```

To run lints

```
make lint
```

## License

Licensed under either of
* Apache License, Version 2.0 ([LICENSE-APACHE](LICENSE-APACHE) or http://www.apache.org/licenses/LICENSE-2.0)
* MIT license ([LICENSE-MIT](LICENSE-MIT) or http://opensource.org/licenses/MIT)

at your option.
