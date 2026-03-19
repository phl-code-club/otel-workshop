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

This is a simple service to service flow to demonstrate the capabilites of OpenTelemetry. We have an `Auth` service to create, validate and get users and a `Profile` service to create, update and get profile data. We are using the _LGTM_ stack, Loki for logs, Grafana for our UI, Tempo for traces, and Mimir (Promethius) for merics data. That sure looks good to me...

Below we have a diagram explaining how data flows between these services

![data flow diagram](./system-diagram.svg)

### Auth Service

An authentication service with the following end points:

`POST /signUp`

```
curl -X POST http://localhost:8000/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "Password123"}'
```

`POST /signIn`

```
curl -X POST http://localhost:8000/signin \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "Password123"}'
```

`GET /user`
_requires token from signin/signup_

```
curl http://localhost:8000/user \
  -H "Authorization: bearer <JWT_TOKEN>"
```

### Profile Service

A user profile service with the following end points:

`GET /`

```
curl http://localhost:8080/ \
  -H "Authorization: bearer <JWT_TOKEN>"
```

`PUT /`
**NOTE:** _this will create a user if one does not exist._

```
curl -X PUT http://localhost:8080/ \
  -H "Authorization: bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"username": "johndoe", "bio": "Hello world", "location": "New York, NY, 10001, 123 Main St"}'
```

OTHER NOTE: The Profile serivce sends a request with the header back to the Auth service to authenticate users.

### Otel Collector

Proxy that receives, processes and exports data to Loki, Tempo, and Prometheus backends

Runs on the following endpoints:

| Transport | Use | Port |
| --------------- | --------------- | --------------- |
| GRPC | OTLP | `4317` |
| HTTP | OTLP | `4318` |
| HTTP | Health Check | `131311` |
| HTTP | Debug | `55679` |
| HTTP | Metrics | `8888` |

### Tempo

Data store for traces

Runs on <http://localhost:3100/>

- receives OTLP data over GRPC

#### Red Panda

Tempo utilizes a Kafka compatible queue for handling workloads.

Runs on <http://localhost:9092/>

### Loki

Data store for logs

Runs on <http://localhost:4317/>

- receives OTLP data over HTTP

### Prometheus

Data store for metrics

- Scrapes data from the Collector on port `9090`

## Grafana

Grafana is our visualization tool. It pulls in data from our sources to let us create graphs and dashboards.

The `Grafana` container exposes port `3000`. Just load up `http//localhost:3000`.

## Starting the Services

To bring up the services defined here you can run `make up`. This will start the services in the current terminal. If you want to start them detached you can use `make ARGS="-d" up`.

## Gen Traffic

Once all the services are up and running you can run our `gentraffic` script. This is a pretty simple script that just generates dummy data for services so we can actually visualize something. It will create a bunch of fake users, then create profiles for them as well. Run the following:

`make trafficgen`

This will spin up a docker container and create 100 fake users. If you want to change the number of users you can use the `USER_COUNT` environment variable:

`make trafficgen USER_COUNT=1000` would generate 1000 fake users.

## Generating Signals

Once you have the collector running you can run [telemetrygen](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/cmd/telemetrygen), a helpful CLI tool from OpenTelemetry to generate dummy signals to test our collector setup. You can run this inside of docker using `make telemetrygen`. This will default to sending 3 traces to the collector. If you want to test other signals in other quantities you can use these make variables:

- `SIGNAL`
  - The signal to send. Needs to be one of `traces`, `logs`, or `metrics`
- `COUNT`
  - The number of signals to send

For example we could use the following to send 5 logs: `make SIGNAL=logs COUNT=5 telemetrygen`.
