FROM docker.io/golang:1.25-alpine AS builder

WORKDIR /var/app

COPY . .

RUN go build -o gen-traffic .

FROM docker.io/alpine AS runner

WORKDIR /var/app

COPY --from=builder /var/app/gen-traffic .

CMD ./gen-traffic
