SHELL := bash

MIGRATE_DIR := src/internal/db/migrations
MIGRATE_URL := postgres://postgres:postgres@localhost:5432/emc_lb?sslmode=disable

.PHONY: help dev run sqlc docker-up docker-down migrate-create migrate-up migrate-down migrate-version clean

help:
	@echo "Available targets:"
	@echo "  make dev             # run application with air live reload"
	@echo "  make run             # start application directly"
	@echo "  make sqlc            # generate Go types from SQL using sqlc"
	@echo "  make docker-up       # start Postgres and MongoDB"
	@echo "  make docker-down     # stop containers"
	@echo "  make migrate-create  # create a new SQL migration"
	@echo "  make migrate-up      # apply migrations to Postgres"
	@echo "  make migrate-down    # rollback last migration"
	@echo "  make migrate-version # show current migration version"
	@echo "  make clean           # remove temporary files, clean cache"

dev:
	air -c .air.toml

run:
	go build ./src/cmd/server
	go run ./src/cmd/server

sqlc:
	sqlc generate

docker-up:
	docker compose up -d

docker-down:
	docker compose down

migrate-create:
	migrate create -ext sql -dir $(MIGRATE_DIR) -seq $(NAME)

migrate-up:
	migrate -path $(MIGRATE_DIR) -database "$(MIGRATE_URL)" up

migrate-down:
	migrate -path $(MIGRATE_DIR) -database "$(MIGRATE_URL)" down

migrate-version:
	migrate -path $(MIGRATE_DIR) -database "$(MIGRATE_URL)" version

clean:
	go clean -cache

