FROM docker.io/golang:alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o allele ./cmd/allele

FROM docker.io/alpine:latest

WORKDIR /app
COPY --from=builder /app/allele .
# All configuration is dynamic and loaded from SQLite

CMD ["./allele"]
