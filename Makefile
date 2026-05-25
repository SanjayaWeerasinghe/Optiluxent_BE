# Makefile for ERP System

# Variables
APP_NAME=erp-system
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR=bin
MAIN_PATH=cmd/api/main.go
DOCKER_IMAGE=$(APP_NAME):$(VERSION)
DOCKER_IMAGE_LATEST=$(APP_NAME):latest

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=$(BUILD_DIR)/$(APP_NAME)

# Database
DB_HOST?=localhost
DB_PORT?=5432
DB_USER?=postgres
DB_PASSWORD?=postgres
DB_NAME?=erp_db
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Migration
MIGRATE=migrate
MIGRATIONS_PATH=./migrations/core

#colors
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

.PHONY: help
help: ## Show this help
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2}' $(MAKEFILE_LIST)

## Development

.PHONY: run
run: ## Run the application
	@echo "${GREEN}Running application...${RESET}"
	@$(GOCMD) run $(MAIN_PATH)

.PHONY: dev
dev: ## Run with hot reload (requires air)
	@echo "${GREEN}Running with hot reload...${RESET}"
	@air

.PHONY: build
build: ## Build the application
	@echo "${GREEN}Building $(APP_NAME) $(VERSION)...${RESET}"
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) -o $(BINARY_NAME) -ldflags="-X 'main.Version=$(VERSION)'" $(MAIN_PATH)
	@echo "${GREEN}Build complete: $(BINARY_NAME)${RESET}"

.PHONY: build-all
build-all: ## Build for all platforms
	@echo "${GREEN}Building for multiple platforms...${RESET}"
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(MAIN_PATH)
	@GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "${GREEN}Multi-platform build complete${RESET}"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "${YELLOW}Cleaning...${RESET}"
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "${GREEN}Clean complete${RESET}"

## Dependencies

.PHONY: deps
deps: ## Download dependencies
	@echo "${GREEN}Downloading dependencies...${RESET}"
	@$(GOMOD) download
	@$(GOMOD) tidy

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "${GREEN}Updating dependencies...${RESET}"
	@$(GOGET) -u ./...
	@$(GOMOD) tidy

.PHONY: deps-vendor
deps-vendor: ## Vendor dependencies
	@echo "${GREEN}Vendoring dependencies...${RESET}"
	@$(GOMOD) vendor

## Testing

.PHONY: test
test: ## Run tests
	@echo "${GREEN}Running tests...${RESET}"
	@$(GOTEST) -v -race ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "${GREEN}Running tests with coverage...${RESET}"
	@$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "${GREEN}Coverage report generated: coverage.html${RESET}"

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "${GREEN}Running integration tests...${RESET}"
	@$(GOTEST) -v -tags=integration ./tests/integration/...

.PHONY: test-e2e
test-e2e: ## Run end-to-end tests
	@echo "${GREEN}Running end-to-end tests...${RESET}"
	@$(GOTEST) -v -tags=e2e ./tests/e2e/...

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "${GREEN}Running benchmarks...${RESET}"
	@$(GOTEST) -bench=. -benchmem ./...

## Code Quality

.PHONY: lint
lint: ## Run linter
	@echo "${GREEN}Running linter...${RESET}"
	@golangci-lint run ./...

.PHONY: lint-fix
lint-fix: ## Run linter and fix issues
	@echo "${GREEN}Running linter with auto-fix...${RESET}"
	@golangci-lint run --fix ./...

.PHONY: fmt
fmt: ## Format code
	@echo "${GREEN}Formatting code...${RESET}"
	@$(GOCMD) fmt ./...
	@gofumpt -l -w .

.PHONY: vet
vet: ## Run go vet
	@echo "${GREEN}Running go vet...${RESET}"
	@$(GOCMD) vet ./...

.PHONY: check
check: fmt vet lint test ## Run all checks (fmt, vet, lint, test)
	@echo "${GREEN}All checks passed!${RESET}"

## Database

.PHONY: db-create
db-create: ## Create database
	@echo "${GREEN}Creating database $(DB_NAME)...${RESET}"
	@docker exec -it erp-postgres createdb -U $(DB_USER) $(DB_NAME) || true

.PHONY: db-drop
db-drop: ## Drop database
	@echo "${YELLOW}Dropping database $(DB_NAME)...${RESET}"
	@docker exec -it erp-postgres dropdb -U $(DB_USER) $(DB_NAME) || true

.PHONY: db-reset
db-reset: db-drop db-create migrate-up seed ## Reset database (drop, create, migrate, seed)
	@echo "${GREEN}Database reset complete${RESET}"

## Migrations

.PHONY: migrate-install
migrate-install: ## Install golang-migrate tool
	@echo "${GREEN}Installing golang-migrate...${RESET}"
	@$(GOCMD) install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

.PHONY: migrate-create
migrate-create: ## Create new migration (usage: make migrate-create name=create_users_table)
	@if [ -z "$(name)" ]; then \
		echo "${YELLOW}Usage: make migrate-create name=migration_name${RESET}"; \
		exit 1; \
	fi
	@echo "${GREEN}Creating migration: $(name)${RESET}"
	@$(MIGRATE) create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)

.PHONY: migrate-up
migrate-up: ## Run all up migrations
	@echo "${GREEN}Running migrations...${RESET}"
	@$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DB_URL)" -verbose up

.PHONY: migrate-down
migrate-down: ## Rollback last migration
	@echo "${YELLOW}Rolling back migration...${RESET}"
	@$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DB_URL)" -verbose down 1

.PHONY: migrate-down-all
migrate-down-all: ## Rollback all migrations
	@echo "${YELLOW}Rolling back all migrations...${RESET}"
	@$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DB_URL)" -verbose down

.PHONY: migrate-force
migrate-force: ## Force migration version (usage: make migrate-force version=1)
	@if [ -z "$(version)" ]; then \
		echo "${YELLOW}Usage: make migrate-force version=N${RESET}"; \
		exit 1; \
	fi
	@echo "${YELLOW}Forcing migration to version $(version)...${RESET}"
	@$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DB_URL)" force $(version)

.PHONY: migrate-version
migrate-version: ## Show current migration version
	@$(MIGRATE) -path $(MIGRATIONS_PATH) -database "$(DB_URL)" version

## Seeding

.PHONY: seed
seed: ## Seed database with default data
	@echo "${GREEN}Seeding database...${RESET}"
	@$(GOCMD) run scripts/seed/main.go

.PHONY: seed-dev
seed-dev: ## Seed database with development data
	@echo "${GREEN}Seeding development data...${RESET}"
	@$(GOCMD) run scripts/seed/main.go --env=development

## Docker

.PHONY: docker-build
docker-build: ## Build production Docker image
	@echo "${GREEN}Building Docker image...${RESET}"
	@docker build -t $(DOCKER_IMAGE) -t $(DOCKER_IMAGE_LATEST) .
	@echo "${GREEN}Docker image built: $(DOCKER_IMAGE)${RESET}"

.PHONY: docker-run
docker-run: ## Run production container standalone
	@echo "${GREEN}Running Docker container...${RESET}"
	@docker run -p 3000:3000 --env-file .env $(DOCKER_IMAGE_LATEST)

.PHONY: docker-up
docker-up: ## Start all services (production) with docker-compose
	@echo "${GREEN}Starting production services...${RESET}"
	@docker-compose up -d --build
	@echo "${GREEN}Services started — API at http://localhost:3000${RESET}"

.PHONY: docker-down
docker-down: ## Stop production services
	@echo "${YELLOW}Stopping services...${RESET}"
	@docker-compose down
	@echo "${GREEN}Services stopped${RESET}"

.PHONY: docker-logs
docker-logs: ## Tail logs for all production services
	@docker-compose logs -f

.PHONY: docker-ps
docker-ps: ## Show running production containers
	@docker-compose ps

.PHONY: docker-clean
docker-clean: ## Remove containers, volumes, and dangling images
	@echo "${YELLOW}Cleaning Docker resources...${RESET}"
	@docker-compose down -v
	@docker-compose -f docker-compose.dev.yml down -v 2>/dev/null || true
	@docker system prune -f
	@echo "${GREEN}Docker cleanup complete${RESET}"

## Dev Docker (hot-reload)

.PHONY: dev-up
dev-up: ## Start dev environment with hot-reload (Dockerfile.dev + Air)
	@echo "${GREEN}Starting dev services with hot-reload...${RESET}"
	@docker-compose -f docker-compose.dev.yml up --build
	@echo "${GREEN}Dev services started — API at http://localhost:3000${RESET}"

.PHONY: dev-up-detached
dev-up-detached: ## Start dev environment in background
	@echo "${GREEN}Starting dev services (detached)...${RESET}"
	@docker-compose -f docker-compose.dev.yml up -d --build

.PHONY: dev-down
dev-down: ## Stop dev environment
	@echo "${YELLOW}Stopping dev services...${RESET}"
	@docker-compose -f docker-compose.dev.yml down

.PHONY: dev-logs
dev-logs: ## Tail dev logs
	@docker-compose -f docker-compose.dev.yml logs -f

.PHONY: dev-logs-api
dev-logs-api: ## Tail dev API logs only
	@docker-compose -f docker-compose.dev.yml logs -f api

.PHONY: dev-restart
dev-restart: ## Restart only the API container (keeps DB/Redis running)
	@docker-compose -f docker-compose.dev.yml restart api

.PHONY: dev-rebuild
dev-rebuild: ## Force rebuild the dev API image
	@docker-compose -f docker-compose.dev.yml up -d --build api

## Code Generation

.PHONY: generate
generate: ## Run go generate
	@echo "${GREEN}Running code generation...${RESET}"
	@$(GOCMD) generate ./...

.PHONY: swagger
swagger: ## Generate Swagger documentation
	@echo "${GREEN}Generating Swagger docs...${RESET}"
	@swag init -g cmd/api/main.go -o docs/swagger
	@echo "${GREEN}Swagger docs generated${RESET}"

.PHONY: mocks
mocks: ## Generate mocks
	@echo "${GREEN}Generating mocks...${RESET}"
	@mockery --all --output=internal/mocks

.PHONY: proto
proto: ## Generate protobuf code
	@echo "${GREEN}Generating protobuf code...${RESET}"
	@protoc --go_out=. --go-grpc_out=. proto/**/*.proto

## Module Generation

.PHONY: new-module
new-module: ## Create new module scaffold (usage: make new-module name=sales)
	@if [ -z "$(name)" ]; then \
		echo "${YELLOW}Usage: make new-module name=module_name${RESET}"; \
		exit 1; \
	fi
	@echo "${GREEN}Creating module: $(name)${RESET}"
	@$(GOCMD) run scripts/generate/module.go --name=$(name)

## Installation

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "${GREEN}Installing development tools...${RESET}"
	@$(GOCMD) install github.com/cosmtrek/air@latest
	@$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GOCMD) install github.com/swaggo/swag/cmd/swag@latest
	@$(GOCMD) install github.com/vektra/mockery/v2@latest
	@$(GOCMD) install mvdan.cc/gofumpt@latest
	@$(GOCMD) install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "${GREEN}Tools installed${RESET}"

## Project Management

.PHONY: init
init: deps install-tools docker-up migrate-up seed ## Initialize project (deps, tools, docker, migrate, seed)
	@echo "${GREEN}Project initialized successfully!${RESET}"

.PHONY: setup
setup: ## Setup development environment
	@echo "${GREEN}Setting up development environment...${RESET}"
	@cp .env.example .env || true
	@echo "${GREEN}Edit .env file with your configuration${RESET}"

.PHONY: status
status: ## Show project status
	@echo "${CYAN}=== Project Status ===${RESET}"
	@echo "Version: $(VERSION)"
	@echo "Go version: $$(go version)"
	@echo "Docker: $$(docker --version 2>/dev/null || echo 'not installed')"
	@echo "Docker Compose: $$(docker-compose --version 2>/dev/null || echo 'not installed')"
	@echo "${CYAN}=== Docker Services ===${RESET}"
	@docker-compose ps || echo "Docker compose not running"
	@echo "${CYAN}=== Database ===${RESET}"
	@make migrate-version 2>/dev/null || echo "Cannot connect to database"

## Security

.PHONY: security-scan
security-scan: ## Run security scan
	@echo "${GREEN}Running security scan...${RESET}"
	@gosec ./...

.PHONY: deps-audit
deps-audit: ## Audit dependencies for vulnerabilities
	@echo "${GREEN}Auditing dependencies...${RESET}"
	@$(GOCMD) list -json -m all | nancy sleuth

## Performance

.PHONY: profile-cpu
profile-cpu: ## Run CPU profiling
	@echo "${GREEN}Running CPU profiling...${RESET}"
	@$(GOCMD) test -cpuprofile=cpu.prof -bench=. ./...
	@$(GOCMD) tool pprof -http=:8080 cpu.prof

.PHONY: profile-mem
profile-mem: ## Run memory profiling
	@echo "${GREEN}Running memory profiling...${RESET}"
	@$(GOCMD) test -memprofile=mem.prof -bench=. ./...
	@$(GOCMD) tool pprof -http=:8080 mem.prof

## Deployment

.PHONY: deploy-staging
deploy-staging: ## Deploy to staging
	@echo "${GREEN}Deploying to staging...${RESET}"
	@kubectl apply -f deployments/kubernetes/ -n staging

.PHONY: deploy-prod
deploy-prod: ## Deploy to production
	@echo "${YELLOW}Deploying to production...${RESET}"
	@kubectl apply -f deployments/kubernetes/ -n production

## Utilities

.PHONY: logs
logs: ## Show application logs (docker)
	@docker-compose logs -f api

.PHONY: shell-api
shell-api: ## Open shell in API container
	@docker-compose exec api /bin/sh

.PHONY: shell-db
shell-db: ## Open PostgreSQL shell
	@docker-compose exec postgres psql -U $(DB_USER) -d $(DB_NAME)

.PHONY: shell-redis
shell-redis: ## Open Redis CLI
	@docker-compose exec redis redis-cli

.PHONY: backup-db
backup-db: ## Backup database
	@echo "${GREEN}Backing up database...${RESET}"
	@docker-compose exec -T postgres pg_dump -U $(DB_USER) $(DB_NAME) > backup_$$(date +%Y%m%d_%H%M%S).sql
	@echo "${GREEN}Backup complete${RESET}"

.PHONY: restore-db
restore-db: ## Restore database (usage: make restore-db file=backup.sql)
	@if [ -z "$(file)" ]; then \
		echo "${YELLOW}Usage: make restore-db file=backup.sql${RESET}"; \
		exit 1; \
	fi
	@echo "${GREEN}Restoring database from $(file)...${RESET}"
	@docker-compose exec -T postgres psql -U $(DB_USER) -d $(DB_NAME) < $(file)
	@echo "${GREEN}Restore complete${RESET}"

## CI/CD

.PHONY: ci
ci: deps lint test ## Run CI pipeline
	@echo "${GREEN}CI pipeline complete${RESET}"

.PHONY: pre-commit
pre-commit: fmt lint test ## Run pre-commit checks
	@echo "${GREEN}Pre-commit checks passed${RESET}"

.DEFAULT_GOAL := help
