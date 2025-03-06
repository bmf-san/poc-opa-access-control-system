# Variables
DOCKER_COMPOSE := docker compose

# Default target
.PHONY: all
all: setup up

.PHONY: setup
setup:
	@echo "Setting up dependencies..."
	@./scripts/setup-deps.sh

# Start all services
.PHONY: up
up:
	@echo "Building and starting all services..."
	@$(DOCKER_COMPOSE) up -d --build

# Stop all services
.PHONY: down
down:
	@echo "Stopping all services..."
	@$(DOCKER_COMPOSE) down

# Restart all services
.PHONY: restart
restart: down up

# Show logs
.PHONY: logs
logs:
	@$(DOCKER_COMPOSE) logs -f

# Show logs for specific service
.PHONY: log
log:
	@if [ -z "$(SERVICE)" ]; then \
		echo "Usage: make log SERVICE=<service_name>"; \
		exit 1; \
	fi
	@$(DOCKER_COMPOSE) logs -f $(SERVICE)

# Execute a command in the employee database container
.PHONY: employee-db
employee-db:
	@$(DOCKER_COMPOSE) exec -it employee-db psql -U postgres -d employee

# Execute a command in the prp database container
.PHONY: prp-db
prp-db:
	@$(DOCKER_COMPOSE) exec -it prp-db psql -U postgres -d prp

# Run tests in containers
.PHONY: test
test: test-opa test-go

.PHONY: test-go
test-go: vet-go
	@echo "Running Go tests..."
	@cd cmd/pep && go test -v -race ./... || exit 1
	@cd cmd/pdp && go test -v -race ./... || exit 1
	@cd cmd/pip && go test -v -race ./... || exit 1
	@cd internal && go test -v -race ./... || exit 1
	@echo "All Go tests passed!"

.PHONY: fmt-go
fmt-go:
	@echo "Checking Go formatting..."
	@if [ "$$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "The following files are not formatted correctly:"; \
		gofmt -s -l .; \
		exit 1; \
	fi
	@echo "Go formatting check passed!"

.PHONY: vet-go
vet-go:
	@echo "Running go vet..."
	@cd cmd/pep && go vet ./... || exit 1
	@cd cmd/pdp && go vet ./... || exit 1
	@cd cmd/pip && go vet ./... || exit 1
	@cd internal && go vet ./... || exit 1
	@echo "Go vet check passed!"

.PHONY: test-opa
test-opa: check-opa
	@echo "Running OPA tests..."
	@cd cmd/pdp/policy && opa test . -v
	@echo "All OPA tests passed!"

.PHONY: fmt-opa
fmt-opa:
	@echo "Formatting OPA policies..."
	@cd cmd/pdp/policy && opa fmt -w .
	@echo "OPA formatting complete!"

.PHONY: check-opa
check-opa:
	@echo "Checking OPA policies..."
	@cd cmd/pdp/policy && opa check .
	@echo "OPA check complete!"

# Generate database documentation by tbls
.PHONY: gen-dbdocs
gen-dbdocs:
	@tbls doc postgres://postgres:postgres@localhost:5433/prp?sslmode=disable docs/db/prp --force
	@tbls doc postgres://postgres:postgres@localhost:5434/employee?sslmode=disable docs/db/employee --force
