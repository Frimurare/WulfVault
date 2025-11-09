.PHONY: build run clean test docker-build docker-run

build:
	go build -o sharecare ./cmd/server

run:
	go run ./cmd/server

clean:
	rm -f sharecare
	rm -rf data/ uploads/

test:
	go test ./...

docker-build:
	docker build -t sharecare:latest .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

install-deps:
	go mod download
	go mod tidy

dev:
	go run ./cmd/server --setup

.DEFAULT_GOAL := build
