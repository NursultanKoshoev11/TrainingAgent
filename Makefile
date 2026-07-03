SHELL := /bin/sh

.PHONY: test build run-compose stop-compose

test:
	go test ./...

build:
	go build ./cmd/news
	go build ./cmd/market
	go build ./cmd/engine
	go build ./cmd/evaluator
	go build ./cmd/api-gateway

run-compose:
	docker compose -f compose.yaml up --build

stop-compose:
	docker compose -f compose.yaml down
