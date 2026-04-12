FROM docker.io/golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o arbitrage ./cmd/arbitrage

FROM docker.io/alpine:latest

WORKDIR /app
COPY --from=builder /app/arbitrage .
# We don't necessarily copy .env here since it's injected by podman-compose, but we need the working directory set.

CMD ["./arbitrage"]
