APP_NAME := rss-communicator
RSSD_BIN := bin/rssd
TGBOT_BIN := bin/tg-bot

DB_DSN ?= postgres://rss:rss@localhost:5432/rss?sslmode=disable
MIGRATIONS_DIR := ./migrations

.PHONY: help deps fmt lint test build build-rssd build-tg run-rssd run-tg db-up db-down migrate-up migrate-down compose-up compose-down

help:
	@echo "Targets:"
	@echo "  deps           - go mod tidy"
	@echo "  fmt            - gofmt"
	@echo "  test           - go test ./..."
	@echo "  build          - build both binaries"
	@echo "  run-rssd       - run rssd locally"
	@echo "  run-tg         - run tg-bot locally"
	@echo "  compose-up     - docker compose up -d"
	@echo "  compose-down   - docker compose down"
	@echo "  migrate-up     - goose up"
	@echo "  migrate-down   - goose down (1 step)"

deps:
	go mod tidy

fmt:
	gofmt -w .

test:
	go test ./...

build: build-rssd build-tg

build-rssd:
	mkdir -p bin
	CGO_ENABLED=0 go build -o $(RSSD_BIN) ./cmd/rssd

build-tg:
	mkdir -p bin
	CGO_ENABLED=0 go build -o $(TGBOT_BIN) ./cmd/tg-bot

run-rssd:
	DB_DSN=$(DB_DSN) go run ./cmd/rssd

run-tg:
	DB_DSN=$(DB_DSN) go run ./cmd/tg-bot

compose-up:
	docker compose up -d --build

compose-down:
	docker compose down

db-up: compose-up

db-down:
	docker compose down -v

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" down
