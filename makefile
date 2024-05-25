.PHONY: all run build test build-image clean

all: build

run:
	docker build -f dockerfile -t wireguard_exporter_test .
	docker compose -f docker-compose.yml up 

test:
	docker build -f dockerfile.test -t wireguard_exporter_test .
	docker build -f dockerfile.test -t wireguard_exporter_test .
	set -e; \
	docker-compose -f docker-compose-test.yml up -d; \
	# Wait for the test container to finish \
	docker-compose -f docker-compose-test.yml logs -f wireguard_exporter_test || true; \
	# Capture the exit code of the test container \
	EXIT_CODE=$$(docker inspect --format='{{.State.ExitCode}}' wireguard_exporter_test); \
	# Bring down the Docker Compose services \
	docker-compose -f docker-compose-test.yml down; \
	# Exit with the test container's exit code \
	if [ "$$EXIT_CODE" -ne 0 ]; then \
		echo "Tests failed, exiting with code $$EXIT_CODE"; \
		exit $$EXIT_CODE; \
	fi
build:
	go mod vendor
	go build -o bin/wireguard_exporter -mod=vendor ./cmd/main.go

build-image:
	docker build -t wireguard_exporter .

clean:
	rm -rf bin/*


