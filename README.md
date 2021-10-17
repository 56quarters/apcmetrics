# Roger

[![build status](https://circleci.com/gh/56quarters/apcmetrics.svg?style=shield)](https://circleci.com/gh/56quarters/apcmetrics)

Prometheus exporter for APC UPSes controlled by [apcupsd](http://www.apcupsd.org/)

## Features

TBD

## Building

To build from source you'll need Go 1.16 installed.

```
git clone git@github.com:56quarters/apcmetrics.git && cd apcmetrics
make build
```

The `apcmetrics` binary will then be in the root of the checkout.

## Install

TBD

## Usage

TBD

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
