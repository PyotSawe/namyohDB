# Makefile for Relational Database project

.PHONY: help build test clean run dev lint format deps install-tools coverage bench docker

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME=relational-db
BUILD_DIR=bin
MAIN_PATH=./cmd/relational-db
GO_FILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*")

# Build information
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Linker flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"

help: ## Show this help message
	@echo "Relational Database - Available Commands:"
	@echo
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

build: deps ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

build-release: deps ## Build optimized release version
	@echo "Building $(BINARY_NAME) (release)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -ldflags "-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Release build completed: $(BUILD_DIR)/$(BINARY_NAME)"

test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./tests/unit/...
	@go test -v ./tests/integration/...
	@echo "All tests passed!"

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	@go test -v ./tests/unit/...

test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	@go test -v ./tests/integration/...

coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	@go test -coverprofile=coverage.out ./tests/unit/... ./tests/integration/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./tests/unit/...

clean: ## Clean build artifacts and test data
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf data
	@rm -f coverage.out coverage.html
	@go clean -cache
	@echo "Clean completed"

run: build ## Build and run the application
	@echo "Starting $(BINARY_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

dev: ## Run the application in development mode
	@echo "Running in development mode..."
	@go run $(MAIN_PATH)

deps: ## Download and tidy dependencies
	@echo "Managing dependencies..."
	@go mod download
	@go mod tidy

format: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted"

lint: install-tools ## Run linter (requires golangci-lint)
	@echo "Running linter..."
	@golangci-lint run

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@echo "Development tools installed"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME):$(VERSION) .

docker-run: docker-build ## Build and run Docker container
	@echo "Running Docker container..."
	@docker run --rm -p 5432:5432 $(BINARY_NAME):$(VERSION)

# Cross-compilation targets
build-linux: deps ## Build for Linux
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

build-windows: deps ## Build for Windows
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

build-darwin: deps ## Build for macOS
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)

build-all: build-linux build-windows build-darwin ## Build for all platforms

# Database management
db-reset: clean ## Reset database (clean data directory)
	@echo "Resetting database..."
	@rm -rf data
	@echo "Database reset completed"

# Development workflow
full-test: clean deps test lint vet ## Full test suite with linting and vetting
	@echo "Full test suite completed successfully!"

ci: deps test vet ## CI pipeline target
	@echo "CI pipeline completed"

# Info targets
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"

env: ## Show environment information
	@echo "Go version: $(shell go version)"
	@echo "GOPATH: $(GOPATH)"
	@echo "GOROOT: $(GOROOT)"
	@echo "GOOS: $(GOOS)"
	@echo "GOARCH: $(GOARCH)"