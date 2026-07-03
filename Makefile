SHELL := /bin/sh
SERVICES := news-ingestor binance-ingestor signal-engine api-gateway

.PHONY: test build docker-up docker-down

test:
	go test ./...

build:
	mkdir -p bin
	for svc in $(SERVICES); do \
		go build -o bin/$$svc ./cmd/$$svc; \
	done

docker-up:
	docker compose up --build

docker-down:
	docker compose down
