.PHONY: all run build test build-image clean

all: build

run:
	docker build -f dockerfile -t wireguard_exporter_test .
	docker compose -f docker-compose.yml up 

test:
	docker build -f dockerfile.test -t wireguard_exporter_test .
	docker compose -f docker-compose-test.yml up --abort-on-container-exit 
build:
	go mod vendor
	go build -o bin/wireguard_exporter -mod=vendor ./cmd/main.go

build-image:
	docker build -t wireguard_exporter .

clean:
	rm -rf bin/*


