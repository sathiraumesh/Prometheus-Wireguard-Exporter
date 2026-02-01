.PHONY: all run build test build-image clean

VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

all: build

run:
	docker build --target prod --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) -t wireguard_exporter .
	docker compose --profile run up

test:
	docker compose --profile test up --build --abort-on-container-exit

build:
	go mod vendor && go build -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT)" -o bin/wireguard_exporter -mod=vendor ./cmd/wireguard-exporter

build-image:
	docker build --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) -t wireguard_exporter .

clean:
	rm -rf bin/
