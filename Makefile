# Variables
GO_WORKSPACE_DIR := $(shell pwd)
CMD_DIR := $(GO_WORKSPACE_DIR)/cmd
BUILD_DIR := $(GO_WORKSPACE_DIR)/bin

# Applications
APPS := pep pdp pip

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
	@echo "Usage: make run APP=<app_name>"
	@if [ -z "$(APP)" ]; then \
		echo "Error: APP is not set!"; \
		exit 1; \
	fi
	@$(BUILD_DIR)/$(APP)-service

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test ./...

# Initialize Go workspace
.PHONY: init
init:
	@echo "Initializing Go workspace..."
	@go work init ./cmd/pep ./cmd/pdp ./cmd/pip ./internal
