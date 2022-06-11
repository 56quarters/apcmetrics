# APC Metrics

![build status](https://github.com/56quarters/apcmetrics/actions/workflows/go.yml/badge.svg)

Prometheus exporter for APC UPSes controlled by [apcupsd](http://www.apcupsd.org/)

## Features

* Export metrics about your APC UPS such as runtime remaining, battery charge, current load, etc.
* Inspect the current status of your APC UPS using `apcmetrics status`
* Inspect recent events for your APC UPS using `apcmetrics events`

The following metrics are exported:

* `apc_info` - Info about the UPS
* `apc_status` - Current status of the UPS
* `apc_time_left` - Remaining runtime left on the batteries in seconds
* `apc_load_percent` - Percentage of load capacity
* `apc_charge_percent` - Percentage of charge of the batteries
* `apc_line_voltage` - Current line voltage
* `apc_low_transfer_voltage` - Line voltage below which the UPS will switch to batteries
* `apc_high_transfer_voltage` - Line voltage above which the UPS will switch to batteries
* `apc_battery_voltage` - Battery voltage
* `apc_nominal_battery_voltage` - Nominal battery voltage
* `apc_nominal_input_voltage` - Nominal input voltage
* `apc_nominal_wattage` - Max power the UPS is designed to supply
* `apc_battery_date` - Date the batteries were last replaced as a UNIX timestamp
* `apc_last_time_on_battery` - Last transfer on to batteries as a UNIX timestamp
* `apc_last_time_off_battery` - Last transfer off of batteries as a UNIX timestamp
* `apc_last_self_test` - Last self test as a UNIX timestamp

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

### Dependencies

`apcmetrics` connects to [`apcupsd`](http://www.apcupsd.org/) to read metrics about an APC UPS. As
such you will need an instance of `apcupsd` running that `apcmetrics` can connect to. `apcmetrics`
uses the [NIS Server](http://www.apcupsd.org/manual/manual.html#nis-server-client-configuration-using-the-net-driver)
feature of `apcupsd` which is usually enabled by default.

`apcupsd` can be installed on Debian or Ubuntu systems with `apt-get install apcupsd`. `apcmetrics`
must be able to connect to the server run by `apcupsd`. The easiest way to do this is to run
`apcmetrics` on the same host that `apcupsd` is running on.

Make sure the following settings are in place for `apcupsd`:

* `NETSERVER on`
* `NISIP <interface address>`
* `NISPORT 3551`

## Usage

The primary purpose of `apcmetrics` is to export metrics about an APC UPS to Prometheus. However,
it can also be used to inspect the current status of your UPS or recent events (such as power outages
and self-tests). Examples of how to do each are given below

### Metrics export

`apcmetrics` must connect to `apcupsd` to export metrics. It defaults to collecting metrics from
a locally running `apcupsd` daemon using the default port of `3551`. If `apcupsd` is not running
locally, you can supply the address using the `--ups.address` CLI flag.

An example of running `apcmetrics` this way:

```
./apcmetrics --ups.address=example:3551 metrics
```

In another terminal:

```
curl -s 'http://localhost:9780/metrics'
```

Prometheus metrics are exposed on port `9780` at `/metrics` by default. Once `apcmetrics`
is running, configure scrapes of it by your Prometheus server. Add the host running
`apcmetrics` as a target under the Prometheus `scrape_configs` section as described by
the example below.

```yaml
# Sample config for Prometheus.

global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    monitor: 'my_prom'

scrape_configs:
  - job_name: apcmetrics
    static_configs:
      - targets: [ 'example:9780' ]
```

### `apcmetrics status`

Running `apcmetrics status` will display the current status of the APC UPS as JSON. It defaults to
collecting metrics from a locally running `apcupsd` daemon using the default port of `3551`. If
`apcupsd` is not running locally, you can supply the address using the `--ups.address` CLI flag.
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
It defaults to collecting metrics from a locally running `apcupsd` daemon using the default
port of `3551`. If `apcupsd` is not running locally, you can supply the address using the
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

apcmetrics is available under the terms of the [GPL, version 3](LICENSE).

### Contribution

Unless you explicitly state otherwise, any contribution intentionally submitted
for inclusion in the work by you shall be licensed as above, without any
additional terms or conditions.
