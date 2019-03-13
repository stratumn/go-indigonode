# Stratumn Node

[![GoDoc](https://github.com/golang/gddo/blob/c782c79e0a3c3282dacdaaebeff9e6fd99cb2919/gddo-server/assets/status.svg)](http://godoc.org/github.com/stratumn/go-node)
[![Build Status](https://semaphoreci.com/api/v1/projects/7e0b5b26-d529-4d2b-a0a0-fabc120c414a/2050943/badge.svg)](https://semaphoreci.com/stratumn/go-node)
[![codecov](https://codecov.io/gh/stratumn/go-node/branch/master/graph/badge.svg?token=nVHWHcr5xQ)](https://codecov.io/gh/stratumn/go-node)
[![Go Report Card](https://goreportcard.com/badge/github.com/stratumn/go-node)](https://goreportcard.com/report/github.com/stratumn/go-node)

Stratumn Node is virtual infrastructure for interoperable P2P services.

## Project Status

The current focus is to build a solid architecture to develop P2P services and run Stratumn's products.

### Current features

- powered by IPFS's go-libp2p library
- core services (P2P, NAT, DHT routing, relay, etc...)
- P2P bootstrapping from seed nodes
- P2P bootstrapping for private networks (with a coordinator node)
- gRPC API
- CLI with gRPC command reflection
- neat and powerful inner-process service based architecture
- raft support for replicated state machine (non-BFT)
- system test framework
- monitoring

### Next

- private networks without coordinator
- layer 2 private networks
- BFT consensus mechanisms
- a simple, script-less, digital asset
- third-party applications/ecosystem

## Installation

Install Go. On macOS, you can install using `homebrew`:

```bash
brew install go
```

Compile and install `stratumn-node`:

```bash
make install
```

## Usage

Create a new directory for your node. Open a terminal in that directory
then create configuration files using `stratumn-node init`:

```bash
stratumn-node init
```

Now you can launch a node (from the same directory):

```bash
stratumn-node up
```

Open another terminal and connect to the node (from the same directory):

```bash
stratumn-node cli
```

The auto-completion should help you explore available APIs easily.

## Logs And Metrics

To view streaming logs (from the same directory):

```bash
stratumn-node log -f log.jsonld
```

To view metrics you need to install Prometheus. On macOS:

```bash
brew install prometheus
```

Then copy `monitoring/prometheus/prometheus.yml` and launch Prometheus:

```bash
prometheus
```

By default, Prometheus is available at `http://localhost:9090`.

## Dashboards and Graphs

To create dashboards, check out [grafana](https://grafana.com).

To install on macOS:

```bash
brew install grafana
```

We provide some useful graph definitions in `monitoring/grafana`.

Copy the `monitoring/grafana` folder and launch Grafana:

```bash
export GF_PATHS_PROVISIONING=./grafana/provisioning
grafana-server
```

## Traces

Distributed tracing is available in Stratumn Node but disabled by default.

You can configure it in the `monitoring` section in `stratumn_node.core.config`.

We recommend using [Jaeger](https://www.jaegertracing.io) to collect traces
locally (during development).

We recommend using [Stackdriver](https://cloud.google.com/stackdriver/)
on [AWS](https://aws.amazon.com/) or [GCP](https://cloud.google.com/).

To view traces locally, set `monitoring.jaeger.endpoint = "/ip4/127.0.0.1/tcp/14268"`
in `stratumn_node.core.config` and run:

```bash
docker run -p 14268:14268 -p 16686:16686 jaegertracing/all-in-one:latest
```

Then visit `http://localhost:16686/` to view your traces.

To use Stackdriver in a Cloud deployment, set `monitoring.stackdriver.project_id`
to your Stackdriver project ID, and traces and metrics should be collected
automatically.

## Documentation

For more information, read the [documentation](doc/README.md).
