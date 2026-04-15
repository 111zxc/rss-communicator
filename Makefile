APP_NAME := rss-communicator
RSSD_BIN := bin/rssd
TGBOT_BIN := bin/tg-bot

DB_DRIVER ?= postgres
DB_DSN ?= postgres://rss:rss@localhost:5432/rss?sslmode=disable
MIGRATIONS_DIR := ./migrations/$(DB_DRIVER)
GOOSE_DRIVER := $(DB_DRIVER)

ifeq ($(DB_DRIVER),sqlite)
GOOSE_DRIVER := sqlite3
endif

.PHONY: help deps fmt lint test build build-rssd build-tg run-rssd run-tg db-up db-down migrate-up migrate-down migrate-up-postgres migrate-down-postgres migrate-up-sqlite migrate-down-sqlite compose-up compose-down compose-up-sqlite-memory compose-up-postgres-memory compose-up-sqlite-nats compose-up-postgres-nats

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
	@echo "  compose-up-sqlite-memory   - docker compose profile sqlite-memory"
	@echo "  compose-up-postgres-memory - docker compose profile postgres-memory"
	@echo "  compose-up-sqlite-nats     - docker compose profile sqlite-nats"
	@echo "  compose-up-postgres-nats   - docker compose profile postgres-nats"
	@echo "  migrate-up     - goose up for DB_DRIVER=$(DB_DRIVER)"
	@echo "  migrate-down   - goose down (1 step) for DB_DRIVER=$(DB_DRIVER)"
	@echo "  migrate-up-postgres   - goose up for postgres"
	@echo "  migrate-down-postgres - goose down for postgres"
	@echo "  migrate-up-sqlite     - goose up for sqlite"
	@echo "  migrate-down-sqlite   - goose down for sqlite"

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
	DB_DRIVER=$(DB_DRIVER) DB_DSN=$(DB_DSN) go run ./cmd/rssd

run-tg:
	DB_DRIVER=$(DB_DRIVER) DB_DSN=$(DB_DSN) go run ./cmd/tg-bot

compose-up:
	docker compose up -d --build

compose-up-sqlite-memory:
	docker compose --profile sqlite-memory up -d --build

compose-up-postgres-memory:
	docker compose --profile postgres-memory up -d --build

compose-up-sqlite-nats:
	docker compose --profile sqlite-nats up -d --build

compose-up-postgres-nats:
	docker compose --profile postgres-nats up -d --build

compose-down:
	docker compose down

db-up: compose-up

db-down:
	docker compose down -v

migrate-up:
	goose -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(DB_DSN)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(DB_DSN)" down

migrate-up-postgres:
	$(MAKE) migrate-up DB_DRIVER=postgres DB_DSN=$(DB_DSN)

migrate-down-postgres:
	$(MAKE) migrate-down DB_DRIVER=postgres DB_DSN=$(DB_DSN)

migrate-up-sqlite:
	$(MAKE) migrate-up DB_DRIVER=sqlite DB_DSN=$(DB_DSN)

migrate-down-sqlite:
	$(MAKE) migrate-down DB_DRIVER=sqlite DB_DSN=$(DB_DSN)
