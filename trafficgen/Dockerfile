FROM docker.io/golang:1.25-alpine AS builder

WORKDIR /var/app

COPY . .

RUN go build -o trafficgen .

FROM docker.io/alpine AS runner

WORKDIR /var/app

COPY --from=builder /var/app/trafficgen .

CMD ./trafficgen
