# Otel Workshop

## Requirements

- A container runtime
  - [Docker Desktop](https://docs.docker.com/desktop/)
  - [Podman Desktop](https://podman-desktop.io/)
    - If you use Podman you will need to set the `CONTAINER_RUNTIME` environment variable to `podman`
- Make
  - Generally preinstalled on UNIX like systems
  - [Make for Windows](https://gnuwin32.sourceforge.net/packages/make.htm)
    - I don't have a Windows machine so YMMV

## Overview

***TODO***

## Starting the Services

To bring up the services defined here you can run `make up`. This will start the services in the current terminal. If you want to start them detached you can use `make ARGS="-d" up`.

## Testing Collector

Once you have the collector running you can run [telemetrygen](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/cmd/telemetrygen), a helpful CLI tool from OpenTelemetry to generate dummy signals to test our collector setup. You can run this inside of docker using `make telemetrygen`. This will default to sending 3 traces to the collector. If you want to test other signals in other quantities you can use these make variables:

- `SIGNAL`
  - The signal to send. Needs to be one of `traces`, `logs`, or `metrics`
- `COUNT`
  - The number of signals to send

For example we could use the following to send 5 logs: `make SIGNAL=logs COUNT=5 telemetrygen`.

## Testing Tempo

Once Tempo is running, if you send some dummy traces in using `telemetrygen` (*See [Testing Collector](#testing-collector)*), you can query Tempo for those same traces. You can use [TraceQL](https://grafana.com/docs/tempo/latest/traceql/construct-traceql-queries/) to search for traces, like so:

`curl http://localhost:3200/api/search -d '{resource.service.name = "telemetrygen"}'`

Then you can take the trace id's from that output and get additional information using:

`curl http://localhost:3200/api/traces/$TRACE_ID`
