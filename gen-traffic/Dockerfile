FROM golang:1.25-alpine AS builder

WORKDIR /var/app

COPY . .

RUN go build -o gen-traffic .

FROM alpine AS runner

WORKDIR /var/app

COPY --from=builder /var/app/gen-traffic .

CMD ./gen-traffic
