# Variables
GO_WORKSPACE_DIR := $(shell pwd)
CMD_DIR := $(GO_WORKSPACE_DIR)/cmd
BUILD_DIR := $(GO_WORKSPACE_DIR)/bin

# Applications
APPS := foo pep pdp pip

# Default target
.PHONY: all
all: build

# Build all applications
.PHONY: build
build: $(APPS)

# Build each application
$(APPS):
	@echo "Building $@..."
	@mkdir -p $(BUILD_DIR)
	@cd $(CMD_DIR)/$@ && go build -o $(BUILD_DIR)/$@-service main.go
	@echo "Built $@ successfully!"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete!"

# Run specific application
.PHONY: run
run:
	@echo "Usage: make run APP=<app_name> or make run-all"
	@if [ -z "$(APP)" ]; then \
		echo "Error: APP is not set!"; \
		exit 1; \
	fi
	@$(BUILD_DIR)/$(APP)-service & echo $$! >> pids.txt
	@echo "$(APP)-service is running with PID: $$!"

.PHONY: run-all
run-all: build
	@echo "Starting all applications in the background..."
	@> pids.txt # Clear the PID file
	@for app in $(APPS); do \
		echo "Starting $$app..."; \
		$(BUILD_DIR)/$$app-service & echo $$! >> pids.txt; \
	done
	@echo "All applications are running in the background. PIDs saved to pids.txt."

.PHONY: stop
stop:
	@if [ ! -f pids.txt ]; then \
		echo "No running processes found."; \
		exit 1; \
	fi
	@echo "Stopping all applications..."
	@cat pids.txt | xargs kill -9
	@rm -f pids.txt
	@echo "All applications have been stopped."

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@for module in ./cmd/foo ./cmd/pep ./cmd/pdp ./cmd/pip ./internal; do \
		echo "Testing $$module..."; \
		go test $$module/... || exit 1; \
	done

# Tidy Go modules in each app
.PHONY: tidy
tidy:
	@echo "Tidying Go modules in each app..."
	@for module in ./cmd/foo ./cmd/pep ./cmd/pdp ./cmd/pip ./internal; do \
		echo "Tidying $$module..."; \
		cd $$module && go mod tidy && cd -; \
	done

# Initialize Go workspace
.PHONY: init
init:
	@echo "Initializing Go workspace..."
	@go work init ./cmd/foo ./cmd/pep ./cmd/pdp ./cmd/pip ./internal
	@echo "Tidying Go modules in each app..."
	$(MAKE) tidy

# Start docker-compose
.PHONY: docker-compose-up
docker-compose-up:
	@echo "Starting docker compose..."
	@docker compose up -d

# Stop docker-compose
.PHONY: docker-compose-down
docker-compose-down:
	@echo "Stopping docker compose..."
	@docker compose down


# Execute a command in the foo database container
.PHONY: foo-db
foo-db:
	docker compose exec foo psql -U postgres -d foo

# Execute a command in the prp database container
.PHONY: prp-db
prp-db:
	docker compose exec prp psql -U postgres -d prp