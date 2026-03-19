.PHONY: build run test test-ci test-integration test-e2e test-coverage lint swagger migrate-up migrate-down docker-up docker-down docker-dev-up docker-dev-down monitoring-up monitoring-down monitoring-restart monitoring-logs stack-up stack-down

# Versions
GOLANGCI_LINT_VERSION := v2.1.5
BIN_DIR := $(PWD)/bin
GOLANGCI_LINT := $(BIN_DIR)/golangci-lint
GOFUMPT := $(BIN_DIR)/gofumpt
GCI := $(BIN_DIR)/gci
GOFUMPT_VERSION := v0.8.0
GCI_VERSION := v0.13.6
LINT_BASE ?= master

# Build
build:
	go build -ldflags="-s -w" -o bin/app ./cmd/app

run:
	go run ./cmd/app -config config/config.yml

# Testing
test:
	go test -v -count=1 ./internal/...

test-ci:
	go test -v -count=1 -coverprofile=coverage.out ./internal/...
	go tool cover -func=coverage.out | tail -1
	go tool cover -html=coverage.out -o coverage.html

test-integration:
	go test -v -count=1 -tags=integration ./tests/...

test-e2e:
	go test -v -count=1 -tags=integration -timeout=120s ./tests/e2e/...

test-coverage:
	go test -v -count=1 -coverprofile=coverage.out ./internal/...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

install-formatters:
	@[ -f $(GOFUMPT) ] || { \
		mkdir -p $(BIN_DIR); \
		echo '📦 Installing gofumpt...'; \
		GOBIN=$(BIN_DIR) go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION); \
	}
	@[ -f $(GCI) ] || { \
		mkdir -p $(BIN_DIR); \
		echo '📦 Installing gci...'; \
		GOBIN=$(BIN_DIR) go install github.com/daixiang0/gci@$(GCI_VERSION); \
	}

format: install-formatters
	@echo "🧼 Formatting with gofumpt..."
	@find . -type f -name '*.go' ! -path '*/mocks/*' -exec $(GOFUMPT) -extra -w {} +
	@echo "🎯 Sorting imports with gci..."
	@find . -type f -name '*.go' ! -path '*/mocks/*' -exec $(GCI) write -s standard -s default -s "prefix(prod-pobeda-2026/)" {} +

# Install golangci-lint
install-golangci-lint:
	@echo "golangci-lint will be run via go run ($(GOLANGCI_LINT_VERSION))"

# Lint
lint: install-golangci-lint
	@echo "Linting..."
	@go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./... --config=.golangci.yml --new-from-rev=$(LINT_BASE)

# Swagger
swagger:
	swag init -g cmd/app/main.go -o docs

# Migrations
PG_DSN ?= "postgres://postgres:postgres@localhost:5432/app_db?sslmode=disable"

migrate-up:
	migrate -path migrations -database $(PG_DSN) up

migrate-down:
	migrate -path migrations -database $(PG_DSN) down 1

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

# Docker — полный стек (app + инфра)
docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

# Docker — только инфра (postgres, redis)
docker-dev-up:
	docker compose -f docker-compose.dev.yml up -d

docker-dev-down:
	docker compose -f docker-compose.dev.yml down

# Monitoring
monitoring-up: ## Start Prometheus + Grafana
	docker compose up -d prometheus grafana

monitoring-down: ## Stop monitoring stack
	docker compose stop prometheus grafana

monitoring-restart: ## Restart monitoring
	docker compose restart prometheus grafana

monitoring-logs: ## View monitoring logs
	docker compose logs -f prometheus grafana

stack-up: ## Start full stack (app + infra + monitoring)
	docker compose up -d --build

stack-down: ## Stop full stack
	docker compose down

# Dependencies
deps:
	go mod tidy
	go mod download
