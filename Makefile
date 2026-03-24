-include .env
export

.PHONY: help
help: ## Display this help screen
	@echo Usage:
	@echo "make up              - Start services with Docker Compose"
	@echo "make down            - Stop services with Docker Compose"
	@echo "make seed            - Seed"
	@echo "make test            - Run tests"
	@echo "make cover           - Run tests with coverage"
	@echo "make linter-golangci - Run golangci-lint"
	@echo "make swagger         - Generate swagger docs"

up: ### Run docker-compose
	docker compose up -d --build
.PHONY: up

down: ### Down docker-compose
	docker compose down -v
.PHONY: compose-down

seed:
	bash ./scripts/seed.sh
.PHONY: seed

test: ### run test
	go test ./... -coverprofile=coverage.out
.PHONY: test

cover: ### run test with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	rm coverage.out
.PHONY: coverage

linter-golangci: ### check by golangci linter
	golangci-lint run ./...
.PHONY: linter-golangci

swagger: ### generate swagger docs
	swag init -g cmd/api/main.go -o docs
.PHONY: swagger