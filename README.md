# Prometheus WireGuard Exporter

A lightweight WireGuard connection stats exporter for Prometheus.

## Usage

```bash
wireguard_exporter -p 9011 -i wg0,wg1
```

| Flag | Description | Default |
| :--- | :---------- | :------ |
| `-p` | Exporter listening port | `9011` |
| `-i` | Comma-separated list of interfaces to monitor | All interfaces |

Flags can also be set via environment variables:

| Environment Variable | Equivalent Flag |
| :------------------- | :-------------- |
| `WIREGUARD_EXPORTER_PORT` | `-p` |
| `WIREGUARD_EXPORTER_INTERFACES` | `-i` |

CLI flags take precedence over environment variables.

## Exported Metrics

| Metric | Type | Description |
| :----- | :--- | :---------- |
| `wireguard_latest_handshake_seconds` | Gauge | Unix timestamp of the latest handshake for a peer |
| `wireguard_transmitted_bytes` | Gauge | Total bytes transmitted to a peer |
| `wireguard_received_bytes` | Gauge | Total bytes received from a peer |
| `wireguard_peer_up` | Gauge | Whether a peer has had a handshake within 5 minutes (1 = up, 0 = down) |
| `wireguard_interface_info` | Gauge | Info metric for a WireGuard interface (labels: interface, public_key, listen_port) |
| `wireguard_scrape_success` | Gauge | Whether the last scrape succeeded (1 = success, 0 = failure) |
| `wireguard_scrape_duration_seconds` | Gauge | Duration of the last scrape in seconds |

Peer metrics use the labels: `interface`, `public_key`, `allowed_ips`.

## Endpoints

| Path | Description |
| :--- | :---------- |
| `/metrics` | Prometheus metrics |
| `/health` | Health check (returns `200 ok`) |

## Build

Build the binary locally:

```bash
make build
```

The version and commit are baked in automatically from git tags:

```bash
git tag v1.0.0
make build
# produces bin/wireguard_exporter with version=v1.0.0
```

Override manually if needed:

```bash
make build VERSION=1.0.0 COMMIT=abc123
```

Build the Docker image:

```bash
make build-image
```

## Test

Run integration tests inside a Docker container with a real WireGuard interface:

```bash
make test
```

Run unit tests locally (no WireGuard required):

```bash
go test ./...
```

## Run Locally

A docker-compose setup simulates two WireGuard nodes with Prometheus and Grafana.

Make sure Docker, docker compose, and make are installed.

```bash
make run
```

This starts:
- Two WireGuard exporter containers connected via WireGuard VPN
- Prometheus scraping both exporters
- Grafana with a pre-configured dashboard

Access Grafana at [http://localhost:3000](http://localhost:3000) (default credentials: `admin` / `admin`).

## Releases

Pushing a tag triggers the CI pipeline to build linux/amd64 and linux/arm64 binaries and publish them to GitHub Releases:

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Project Structure

```
cmd/wireguard-exporter/   # Application entrypoint and CLI
internal/wgprometheus/    # Prometheus collector implementation
setup/                    # WireGuard configs, Prometheus, Grafana provisioning
```

<img width="2346" height="1167" alt="Screenshot 2026-02-01 at 6 09 42 PM" src="https://github.com/user-attachments/assets/25b3133e-9120-4412-b321-39d6920babf7" />

