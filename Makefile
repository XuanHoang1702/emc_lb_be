SHELL := bash

MIGRATE_DIR := src/internal/db/migrations
MIGRATE_URL := postgres://postgres:postgres@localhost:5432/emc_lb?sslmode=disable

.PHONY: help dev run build lint sqlc swag docker-up docker-down migrate-create migrate-up migrate-down migrate-version clean test test-cover

help:
	@echo "Available targets:"
	@echo "  make dev             # run application with air live reload"
	@echo "  make run             # start application directly"
	@echo "  make build           # build project"
	@echo "  make test            # run all unit tests"
	@echo "  make test-cover      # run tests with HTML coverage report"
	@echo "  make lint            # run golangci-lint"
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
	go build -o tmp_server ./src/cmd/server
	./tmp_server

build:
	go build -v ./...

lint:
	golangci-lint run ./...

sqlc:
	sqlc generate

mock:
	go run github.com/vektra/mockery/v2@v2.42.1 --all --keeptree --dir=src/internal/repository --output=src/tests/mocks/repository
	go run github.com/vektra/mockery/v2@v2.42.1 --all --keeptree --dir=src/pkg/storage --output=src/tests/mocks/storage
	go run github.com/vektra/mockery/v2@v2.42.1 --all --keeptree --dir=src/pkg/mail --output=src/tests/mocks/mail
	go run github.com/vektra/mockery/v2@v2.42.1 --all --keeptree --dir=src/pkg/cache --output=src/tests/mocks/cache

swag:
	go run github.com/swaggo/swag/cmd/swag@latest init -g src/cmd/server/main.go -o src/docs

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

test:
	go test -v -race ./...

test-cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"
