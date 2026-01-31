.PHONY: build test clean install run help coverage lint demo

# Binary name
BINARY_NAME=chatgpt-cli

# Build variables
BUILD_DIR=./bin
GO=go
GOFLAGS=-v

help: ## Show this help message
	@echo 'ChatGPT CLI  - Makefile Commands'
	@echo ''
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building $(BINARY_NAME) ..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "✓ Binary created at $(BUILD_DIR)/$(BINARY_NAME)"

install: ## Install the binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(GOFLAGS)
	@echo "✓ Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

test: ## Run all tests
	@echo "Running tests..."
	$(GO) test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out
	@echo ""
	@echo "For HTML coverage report, run: make coverage-html"

coverage-html: test-coverage ## Generate HTML coverage report
	@echo "Generating HTML coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated at coverage.html"

benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem ./...

clean: ## Remove built binaries and test artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "✓ Clean complete"

fmt: ## Format Go code
	@echo "Formatting code..."
	$(GO) fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	$(GO) vet ./...

lint: fmt vet ## Run formatting and vetting

check: lint test ## Run all checks (lint and test)

watch-test: ## Watch for changes and run tests (requires entr)
	@echo "Watching for changes... (Press Ctrl+C to stop)"
	@find . -name '*.go' | entr -c make quick-test

# Release targets
release-check: ## Check if ready for release
	@echo "Release Check"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@make lint
	@make test
	@echo ""
	@echo "✓ All checks passed! Ready for release."

version: ## Show version info
	@echo "ChatGPT CLI "
	@echo "Go version: $(shell go version)"
	@echo "Build directory: $(BUILD_DIR)"

.DEFAULT_GOAL := help
