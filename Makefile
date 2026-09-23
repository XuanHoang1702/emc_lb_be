SHELL := bash

MIGRATE_DIR := src/internal/db/migrations
MIGRATE_URL := postgres://postgres:postgres@localhost:5433/emc_lb?sslmode=disable

COMPOSE_FILE := docker/docker-compose.yml
ENV_FILE     := .env.development

# Docker service names
API_SERVICES   := emc_api_node_1 emc_api_node_2 emc_api_node_3
WORKER_SERVICE := emc_worker
APP_SERVICES   := $(API_SERVICES) $(WORKER_SERVICE)

.PHONY: help dev run build lint sqlc swag mock clean test test-cover \
        docker-up docker-down docker-logs docker-status \
        deploy rebuild restart watch \
        migrate-create migrate-up migrate-down migrate-version

# ============================================================
# Help
# ============================================================

help:
	@echo ""
	@echo "╔══════════════════════════════════════════════════════════╗"
	@echo "║                    EMC LB Makefile                      ║"
	@echo "╚══════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "  Development:"
	@echo "    make dev             Run with air live reload (local)"
	@echo "    make run             Start application directly (local)"
	@echo "    make build           Build project (go build)"
	@echo "    make test            Run all unit tests"
	@echo "    make test-cover      Run tests with HTML coverage report"
	@echo "    make lint            Run golangci-lint"
	@echo "    make sqlc            Generate Go types from SQL"
	@echo "    make mock            Regenerate mocks"
	@echo "    make clean           Remove temp files, clean cache"
	@echo ""
	@echo "  Docker:"
	@echo "    make deploy          Build & start all containers"
	@echo "    make rebuild         Rebuild API + Worker (no cache) & restart"
	@echo "    make restart         Restart API + Worker containers only"
	@echo "    make watch           Auto-rebuild containers on code save"
	@echo "    make docker-up       Start all containers"
	@echo "    make docker-down     Stop all containers"
	@echo "    make docker-logs     Tail logs from API + Worker"
	@echo "    make docker-status   Show container status"
	@echo ""
	@echo "  Migrations:"
	@echo "    make migrate-create NAME=xxx   Create a new migration"
	@echo "    make migrate-up                Apply migrations"
	@echo "    make migrate-down              Rollback last migration"
	@echo "    make migrate-version           Show current version"
	@echo ""

# ============================================================
# Development (local)
# ============================================================

dev:
	air -c .air.toml

run:
	go build -o tmp_server ./src/cmd/server
	./tmp_server

build:
	go build -v ./...

lint:
	golangci-lint run ./...

sqlc:
	sqlc generate

mock:
	go run github.com/vektra/mockery/v2@latest --all --keeptree --dir=src/internal/repository --output=src/tests/mocks/repository
	go run github.com/vektra/mockery/v2@latest --all --keeptree --dir=src/pkg/storage --output=src/tests/mocks/storage
	go run github.com/vektra/mockery/v2@latest --all --keeptree --dir=src/pkg/mail --output=src/tests/mocks/mail
	go run github.com/vektra/mockery/v2@latest --all --keeptree --dir=src/pkg/cache --output=src/tests/mocks/cache

swag:
	go run github.com/swaggo/swag/cmd/swag@latest init -g src/cmd/server/main.go -o src/docs

clean:
	go clean -cache
	rm -f tmp_server coverage.out coverage.html

test:
	go test -v -race ./...

test-cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# ============================================================
# Docker Compose
# ============================================================

docker-up:
	docker compose --env-file $(ENV_FILE) -f $(COMPOSE_FILE) up -d

docker-down:
	docker compose --env-file $(ENV_FILE) -f $(COMPOSE_FILE) down

docker-logs:
	docker compose --env-file $(ENV_FILE) -f $(COMPOSE_FILE) logs -f $(APP_SERVICES)

docker-status:
	docker compose --env-file $(ENV_FILE) -f $(COMPOSE_FILE) ps

# ============================================================
# Deploy & Rebuild
# ============================================================

## Build images & start all containers (first-time or full deploy)
deploy:
	@echo "🚀 Building & deploying all services..."
	docker compose --env-file $(ENV_FILE) -f $(COMPOSE_FILE) up -d --build
	@echo "✅ All services are running"
	@$(MAKE) docker-status

## Rebuild only API + Worker containers (no cache) and restart
rebuild:
	@echo "🔨 Rebuilding API + Worker..."
	docker compose --env-file $(ENV_FILE) -f $(COMPOSE_FILE) build --no-cache $(APP_SERVICES)
	docker compose --env-file $(ENV_FILE) -f $(COMPOSE_FILE) up -d --no-deps $(APP_SERVICES)
	@echo "✅ Rebuild complete"
	@$(MAKE) docker-status

## Quick restart API + Worker (no rebuild, reuses existing image)
restart:
	@echo "🔄 Restarting API + Worker..."
	docker compose --env-file $(ENV_FILE) -f $(COMPOSE_FILE) restart $(APP_SERVICES)
	@echo "✅ Restart complete"

# ============================================================
# Watch Mode — auto-rebuild on code save
# ============================================================

## Watch for .go/.html file changes in src/ and auto-rebuild containers
watch:
	@echo "👀 Watching src/ for changes... (Ctrl+C to stop)"
	@echo "   Triggers rebuild of API + Worker on every save"
	@echo ""
	@if command -v inotifywait >/dev/null 2>&1; then \
		while true; do \
			inotifywait -r -e modify,create,delete \
				--include '\.(go|html|tmpl|tpl)$$' \
				src/ 2>/dev/null; \
			echo ""; \
			echo "📦 Change detected! Rebuilding..."; \
			echo "─────────────────────────────────────"; \
			$(MAKE) rebuild; \
			echo ""; \
		done; \
	elif command -v fswatch >/dev/null 2>&1; then \
		fswatch -r -e '.*' -i '\\.go$$' -i '\\.html$$' -i '\\.tmpl$$' -i '\\.tpl$$' \
			--one-per-batch src/ | while read -r _; do \
			echo ""; \
			echo "📦 Change detected! Rebuilding..."; \
			echo "─────────────────────────────────────"; \
			$(MAKE) rebuild; \
			echo ""; \
		done; \
	else \
		echo "❌ Cần cài inotifywait hoặc fswatch:"; \
		echo "   Ubuntu/Debian: sudo apt install inotify-tools"; \
		echo "   macOS:         brew install fswatch"; \
		exit 1; \
	fi

# ============================================================
# Migrations
# ============================================================

migrate-create:
	migrate create -ext sql -dir $(MIGRATE_DIR) -seq $(NAME)

migrate-up:
	migrate -path $(MIGRATE_DIR) -database "$(MIGRATE_URL)" up

migrate-down:
	migrate -path $(MIGRATE_DIR) -database "$(MIGRATE_URL)" down

migrate-version:
	migrate -path $(MIGRATE_DIR) -database "$(MIGRATE_URL)" version
