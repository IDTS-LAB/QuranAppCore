# Development workflow: `make setup` (db + migrations), then `make dev`.
# Override the DB URL when needed:  make migrate-up DB_URL=postgres://...

.DEFAULT_GOAL := help

APP := ./cmd/app
BIN := ./bin/app
COMPOSE := docker compose -f docker-compose.dev.yml

DB_URL ?= postgres://gorm:gorm@localhost:5432/gorm?sslmode=disable
MIGRATE_BIN := ./bin/migrate
# golang-migrate gates drivers behind build tags; go tool(1) cannot pass
# -tags, so build a local binary with the postgres driver instead.
MIGRATE := $(MIGRATE_BIN) -path migrations -database "$(DB_URL)"

##@ General

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2} /^##@/ {printf "\n%s\n", substr($$0, 5)}' $(MAKEFILE_LIST)

##@ App

.PHONY: setup
setup: db-up minio-up migrate-up ## Start dev DB + MinIO and apply migrations
	@echo "Run 'make dev' to start the API."

.PHONY: dev
dev: ## Run with live reload (air)
	@command -v air >/dev/null || go install github.com/air-verse/air@latest
	air

.PHONY: run
run: ## Run without reload
	go run $(APP)

.PHONY: build
build: ## Build binary to ./bin/app
	@mkdir -p bin
	go build -o $(BIN) $(APP)

.PHONY: test
test: ## Run all tests
	go test ./...

.PHONY: vet
vet: ## Vet + gofmt check
	go vet ./...
	@test -z "$$(gofmt -l cmd internal pkg 2>/dev/null)" || (echo "gofmt diff:"; gofmt -d cmd internal pkg; exit 1)

.PHONY: fmt
fmt: ## Format code
	gofmt -w cmd internal pkg

.PHONY: tidy
tidy: ## Tidy go.mod
	go mod tidy

.PHONY: tools
tools: $(MIGRATE_BIN) ## Build local helper binaries (migrate CLI)

.PHONY: swag
swag: ## Format annotations + regenerate Swagger docs
	go tool swag fmt ./...
	go tool swag init -g cmd/app/main.go -o internal/core/swagger/docs

$(MIGRATE_BIN): go.mod ## Build migrate CLI with postgres driver
	@mkdir -p bin
	go build -tags postgres -o $(MIGRATE_BIN) github.com/golang-migrate/migrate/v4/cmd/migrate

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf bin tmp/main

##@ Database (docker-compose.dev.yml)

.PHONY: db-up
db-up: ## Start dev Postgres
	$(COMPOSE) up -d db

.PHONY: minio-up
minio-up: ## Start dev MinIO (S3-compatible storage: API :9000, console :9001)
	$(COMPOSE) up -d minio

.PHONY: db-down
db-down: ## Stop dev Postgres (keep data)
	$(COMPOSE) stop db

.PHONY: db-logs
db-logs: ## Tail dev Postgres logs
	$(COMPOSE) logs -f db

.PHONY: db-reset
db-reset: ## Destroy dev DB container AND its data volume
	$(COMPOSE) down -v

##@ Migrations (golang-migrate, make targets only — never auto-run on boot)

.PHONY: migrate-new
migrate-new: $(MIGRATE_BIN) ## New migration pair: make migrate-new NAME=create_users
	@test -n "$(NAME)" || (echo "Usage: make migrate-new NAME=<snake_case_name>"; exit 1)
	$(MIGRATE_BIN) create -ext sql -dir migrations -seq $(NAME)

.PHONY: migrate-up
migrate-up: $(MIGRATE_BIN) ## Apply all pending migrations
	$(MIGRATE) up

.PHONY: migrate-down
migrate-down: $(MIGRATE_BIN) ## Revert migrations (all, or N=1 for one step)
	$(MIGRATE) down $(N)

.PHONY: migrate-goto
migrate-goto: $(MIGRATE_BIN) ## Migrate to version: make migrate-goto V=3
	@test -n "$(V)" || (echo "Usage: make migrate-goto V=<version>"; exit 1)
	$(MIGRATE) goto $(V)

.PHONY: migrate-force
migrate-force: $(MIGRATE_BIN) ## Force version (fixes dirty state): make migrate-force V=3
	@test -n "$(V)" || (echo "Usage: make migrate-force V=<version>"; exit 1)
	$(MIGRATE) force $(V)

.PHONY: migrate-version
migrate-version: $(MIGRATE_BIN) ## Show current migration version
	$(MIGRATE) version

.PHONY: migrate-drop
migrate-drop: $(MIGRATE_BIN) ## Drop EVERYTHING (dev only!)
	$(MIGRATE) drop -f
