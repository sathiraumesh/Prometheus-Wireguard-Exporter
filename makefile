.PHONY: run build build-image

run:
	docker compose up

build:
	go mod vendor
	go build -o bin/wireguard_exporter -mod=vendor ./cmd/main.go

build-static:
	go mod vendor
	go build -ldflags "-linkmode external -extldflags '-static'" -o wireguard_exporter  -mod=vendor ./cmd/main.go

build-image:
	docker build -t wireguard_exporter .

clean:
	rm -rf bin/*


