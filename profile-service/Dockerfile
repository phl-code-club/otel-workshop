FROM docker.io/golang:1.25-alpine AS builder

WORKDIR /var/app

COPY . .

RUN go build -o profile-service .

FROM docker.io/alpine AS runner

WORKDIR /var/app

COPY --from=builder /var/app/profile-service .

EXPOSE 8000

CMD ./profile-service
