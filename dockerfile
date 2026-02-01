# Test base: Ubuntu + WireGuard (needs apt for golang-go in test stage)
FROM ubuntu:24.04 AS test-base
COPY ./setup/wireguard/wireguard_setup.sh .
RUN chmod +x ./wireguard_setup.sh && ./wireguard_setup.sh

# Builder: compile the Go binary
FROM golang:1.25.6 AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY cmd/ cmd/
COPY internal/ internal/
ARG VERSION=dev
ARG COMMIT=unknown
RUN go mod vendor && \
    go build -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -linkmode external -extldflags '-static'" -o main -mod=vendor ./cmd/wireguard-exporter

# Test target: Ubuntu + WireGuard + Go 1.25.6 + source (for integration tests)
FROM test-base AS test
RUN apt-get install -y ca-certificates
COPY --from=builder /usr/local/go /usr/local/go
ENV PATH="/usr/local/go/bin:${PATH}"
WORKDIR /app
COPY go.mod go.sum ./
COPY cmd/ cmd/
COPY internal/ internal/

# Production target: Alpine + WireGuard + compiled binary (minimal image)
FROM alpine:3.20 AS prod
RUN apk add --no-cache wireguard-tools iproute2 bash
WORKDIR /opt/
COPY --from=builder /app/main .
