.PHONY: run build build-image

run:
	docker compose up

build:
	go build -o bin/wireguard_exporter -mod=vendor ./cmd/main.go

build-image:
	docker build -t wireguard_exporter .

clean:
	rm -rf bin/*


