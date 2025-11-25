# --- CONFIGURATIONS ---
PACKAGES = $(shell go list ./...)
PACKAGES_PATH = $(shell go list -f '{{ .Dir }}' ./...)

# Database Config for Migrations
DB_URL=postgres://postgres:1234@localhost:5432/ecommerce?sslmode=disable
MIGRATE_PATH=internal/database/migrations

.PHONY: all
all: help

help:
	@echo "FinHub Makefile commands"
	@echo "----------------------------------------------------------------"
	@echo "  DEV ENVIRONMENT (GoLand / VSCode)"
	@echo "  setup              Configures Git Hooks and permissions"
	@echo "  db                 Start ONLY the Database (Required for Debugging)"
	@echo ""
	@echo "  ALTERNATIVE RUN MODES (CLI)"
	@echo "  run                Run application via terminal (Don't use with Debugger)"
	@echo "  air                Run with hot-reload (Don't use with Debugger)"
	@echo ""
	@echo "  DOCKER / INFRA"
	@echo "  up                 Start ALL containers (DB + API)"
	@echo "  down               Stop and remove all containers"
	@echo "  logs               Show container logs"
	@echo "  ps                 Show container status"
	@echo ""
	@echo "  DATABASE MIGRATIONS"
	@echo "  migrate-create     Create a new migration (usage: NAME=foo)"
	@echo "  migrate-up         Run migrations UP"
	@echo "  migrate-down       Run migrations DOWN (undo last)"
	@echo "  migrate-force      Force version (fix dirty state)"
	@echo ""
	@echo "  TESTING & QUALITY"
	@echo "  test               Run unit tests"
	@echo "  linter             Run code checks"
	@echo "  fmt                Format code"
	@echo "----------------------------------------------------------------"

# --- SETUP ---
.PHONY: setup
setup:
	@echo "Configuring development environment..."
	@chmod +x scripts/*.sh
	@./scripts/install-hooks.sh

# --- DOCKER / INFRASTRUCTURE ---
.PHONY: up
up:
	@echo "Starting full environment (DB + API)..."
	@docker compose up -d

.PHONY: down
down:
	@echo "Stopping environment..."
	@docker compose down

.PHONY: db
db:
	@echo "Starting Database only..."
	@docker compose up -d db
	@# Wait to ensure DB is ready
	@echo "Waiting for Database to be ready..."
	@sleep 2

.PHONY: logs
logs:
	@docker compose logs -f

.PHONY: ps
ps:
	@docker compose ps

# --- DATABASE MIGRATIONS ---
.PHONY: migrate-create
migrate-create:
	@if [ -z "$(NAME)" ]; then echo "Error: Define migration name. Ex: make migrate-create NAME=users"; exit 1; fi
	@echo "Creating migration file for $(NAME)..."
	@migrate create -ext sql -dir $(MIGRATE_PATH) -seq $(NAME)

.PHONY: migrate-up
migrate-up:
	@echo "Running migrations UP..."
	@migrate -path $(MIGRATE_PATH) -database "$(DB_URL)" -verbose up

.PHONY: migrate-down
migrate-down:
	@echo "Running migration DOWN..."
	@migrate -path $(MIGRATE_PATH) -database "$(DB_URL)" -verbose down 1

.PHONY: migrate-force
migrate-force:
	@echo "Forcing migration version..."
	@migrate -path $(MIGRATE_PATH) -database "$(DB_URL)" force $(VERSION)

# --- SEED---
.PHONY: seed
seed:
	@echo "Seeding database..."
	@go run cmd/seed/main.go


# --- GO / APP ---
.PHONY: run
run:
	@echo "Running App locally..."
	@export SCOPE=local && go run cmd/api/main.go

.PHONY: air
air:
	@air -c .air.toml -build.args_bin serve

.PHONY: check_tools
check_tools:
	@type "git" > /dev/null 2>&1 || echo 'Missing: git'
	@type "go" > /dev/null 2>&1 || echo 'Missing: go'
	@type "docker" > /dev/null 2>&1 || echo 'Missing: docker'
	@type "migrate" > /dev/null 2>&1 || echo 'Missing: migrate (golang-migrate)'

.PHONY: ensure-deps
ensure-deps:
	@go mod tidy

.PHONY: fmt
fmt:
	@go fmt $(PACKAGES)
	@goimports -w $(PACKAGES_PATH)

.PHONY: linter
linter:
	@golangci-lint run ./...

.PHONY: test
test:
	@./scripts/run-tests.sh

.PHONY: clean
clean:
	@rm -rf ./coverage
	@go clean -testcache


.PHONY: docs
docs:
	@echo "Generating Swagger documentation..."
	@# --parseDependency: Lê arquivos do GORM e outras libs externas
	@# --parseInternal: Garante que leia seus pacotes internal
	@swag init -g cmd/api/docs.go --output docs --parseDependency --parseInternal