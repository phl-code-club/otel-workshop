# Logs, Traces, Metrics, Oh My

## Introduction

Hello and welcome to our [OpenTelemetry](https://opentelemetry.io) workshop.
Today you will be setting up an [OTel Collector](https://opentelemetry.io/docs/collector),
a small proxy service for moving telemetry data, sometimes called _signals_,
from place to place. This workshop is mostly self-guided, outside of the short
tutorial in this README and the short slideshow found [here](./slides.pdf). We
want to encourage you to experiment, break things, and ask questions! Don't be shy
about talking to your neighbors as well. Learning is better when it's done together :)

### What Is Telemetry

> Telemetry is derived from the Greek roots _tele_, meaning "far off" and _metron_,
> "measure".

Telemetry is all the information about a _remote system_. Some examples of Telemetry data:

- A reading from a mass spectrometer
- The speed of a car being measured by its speedometer
- That text you sent your boss to tell them you are running late for the 3rd time
this week
  - But only in the most abstract sense :p

In our case, it is our information about **running software**. This can be things
like memory usage, network packets sent, information about HTTP requests, etc.

### Why Is Telemetry

Telemetry allows you to peer into remote systems and retrieve data about that system.
There are tons of great uses for telemetry data, including:

- Optimizing a web-server to reduce request latency
- Reading logs from a piece of software to see why it's behavior doesn't match
your expectation
- Tracing an error in a distributed system to find the specific service that
failed and why

### When Is Telemetry

Ideally any time your software is running it should be keeping information about
itself.

### Who Is Telemetry

What ever is having information about it measured. For our use case, this is software.

### Where Is Telemetry

This is the focus of this workshop!

## What We Are Not Covering

- How Docker and similar container runtimes work
- How to scale these systems in production
  - We are happy to chat about that though, it's just out of scope for this tutorial
- The specifics of the OTLP spec
- The internals of the storage backends

## Requirements

- A container runtime
  - [Docker Desktop](https://docs.docker.com/desktop/)
  - [Podman Desktop](https://podman-desktop.io/)
    - If you use Podman you will need to set the `CONTAINER_RUNTIME` environment
    variable to `podman`
- Make
  - Generally preinstalled on UNIX like systems
  - [Make for Windows](https://gnuwin32.sourceforge.net/packages/make.htm)
    - I don't have a Windows machine so YMMV

## Tutorial

Head on over [here](./tutorial.md) to start setting up the OTel Collector!

## Overview

This is a simple service to service flow to demonstrate the capabilities of OpenTelemetry.
We have an [Auth service](./auth-service/) to create, validate and get users and
a [Profile service](./profile-service/) to create, update and get profile data.
This flows into our [OpenTelemetry Collector](./collector/), which takes all our
signals and forwards them to their respective storage systems. Speaking of
telemetry storage systems, we are using the _LGTM_ stack:

- `Loki` for logs
- `Grafana` for our UI
- `Tempo` for traces
- `Mimir` (We are using an equivalent called `Prometheus`) for metrics.

That sure looks good to me...

### Diagram

Below we have a diagram explaining how data flows between these services

![data flow diagram](./system-diagram.png)

### Auth Service

An authentication service with the following endpoints:

`POST /signUp`

```sh
curl -X POST http://localhost:8000/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "Password123"}'
```

`POST /signIn`

```sh
curl -X POST http://localhost:8000/signin \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "Password123"}'
```

`GET /user`

_requires token from signin/signup_

```sh
curl http://localhost:8000/user \
  -H "Authorization: bearer $TOKEN"
```

### Profile Service

A user profile service that authenticates against the `Auth` service. It has the
following endpoints:

`GET /`

```sh
curl http://localhost:8080/ \
  -H "Authorization: bearer $TOKEN"
```

`PUT /`

**NOTE:** _this will create a profile if one does not exist._

```sh
curl -X PUT http://localhost:8080/ \
  -H "Authorization: bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"username": "johndoe", "bio": "Hello world", "location": "New York, NY, 10001, 123 Main St"}'
```

### Otel Collector

Proxy that receives, processes, and exports data to Loki, Tempo, and Prometheus
backends.

Runs on the following ports:

| Transport | Use | Port |
| --------------- | --------------- | --------------- |
| gRPC | OTLP | `4317` |
| HTTP | OTLP | `4318` |
| HTTP | Health Check | `131311` |
| HTTP | Debug | `55679` |
| HTTP | Metrics | `8888` |

### Tempo

Data store for traces

Runs on <http://localhost:3200/>

- Receives OTLP data over gRPC on port `4317` (This port isn't exposed as it
would conflict with our collector)

#### Red Panda

Tempo utilizes a Kafka compatible queue for handling workloads.

### Loki

Data store for logs

Runs on <http://localhost:3100/>

- Receives OTLP data over HTTP

### Prometheus

Data store for metrics

- Runs on <http://localhost:9090>
- Scrapes data from the Collector on port `9090` (This port isn't exposed on the
collector container because it would conflict with the Prometheus container)

### Grafana

Grafana is our visualization tool. It pulls in data from our sources to let us
create graphs and dashboards.

The `Grafana` container exposes port `3000`. Just load up `http//localhost:3000`.

## Starting the Services

First we want to build our local images by running: `make`

Then to bring up the services defined here you can run `make up`. This will start
the services in the current terminal. If you want to start them detached you can
use `make ARGS="-d" up`.

## Gen Traffic

Once all the services are up and running you can run our `gentraffic` script. This
is a pretty simple script that just generates dummy data for services so we can
actually visualize something. It will create a bunch of fake users, then create
profiles for them as well. Run the following:

`make trafficgen`

This will spin up a Docker container and create 100 fake users. If you want to
change the number of users you can use the `USER_COUNT` environment variable:

`make trafficgen USER_COUNT=1000` would generate 1000 fake users.

## Generating Signals

Once you have the collector running you can run
[telemetrygen](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/cmd/telemetrygen),
a helpful CLI tool from OpenTelemetry to generate dummy signals to test our
collector setup. You can run this inside of Docker using `make telemetrygen`.
This will default to sending 3 traces to the collector. If you want to test other
signals in other quantities you can use these make variables:

- `SIGNAL`
  - The signal to send. Needs to be one of `traces`, `logs`, or `metrics`
- `COUNT`
  - The number of signals to send

For example, we could use the following to send 5 logs: `make SIGNAL=logs COUNT=5 telemetrygen`.

## Links

Here are links to all the technologies we used:

- [OTel Collector](https://github.com/open-telemetry/opentelemetry-collector-contrib)
- [Loki](https://github.com/grafana/loki)
- [Grafana](https://github.com/grafana/grafana)
- [Tempo](https://github.com/grafana/tempo)
- [Prometheus](https://github.com/prometheus/prometheus)
- [Red Panda](https://github.com/redpanda-data/redpanda)
- [Postgres](https://www.postgresql.org/list/pgsql-hackers/)
- [Postgres Exporter](https://github.com/prometheus-community/postgres_exporter)
- [NodeJS](https://github.com/nodejs/node)
- [Go](https://github.com/golang/go)
- [Tons of Otel libraries and packages](https://opentelemetry.io/docs/languages/)
