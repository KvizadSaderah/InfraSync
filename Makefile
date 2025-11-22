.PHONY: build test lint clean install help

# Build variables
BINARY_NAME=infrasync
VERSION?=0.2.0
BUILD_DIR=dist
CMD_DIR=cmd/infrasync

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BINARY_NAME) -ldflags="-s -w -X main.version=$(VERSION)" ./$(CMD_DIR)
	@echo "✓ Build complete: ./$(BINARY_NAME)"

install: build ## Install the binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@mv $(BINARY_NAME) $(GOPATH)/bin/
	@echo "✓ Installed to $(GOPATH)/bin/$(BINARY_NAME)"

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "✓ Tests complete"

test-coverage: test ## Run tests with coverage report
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

lint: ## Run linters
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run --timeout=5m
	@echo "✓ Linting complete"

fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) ./...
	@echo "✓ Formatting complete"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "✓ Clean complete"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) verify
	@echo "✓ Dependencies downloaded"

tidy: ## Tidy dependencies
	@echo "Tidying dependencies..."
	$(GOMOD) tidy
	@echo "✓ Dependencies tidied"

release: clean ## Build release binaries for all platforms
	@echo "Building release binaries..."
	@mkdir -p $(BUILD_DIR)

	@echo "  → Linux amd64"
	@GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 -ldflags="-s -w -X main.version=$(VERSION)" ./$(CMD_DIR)

	@echo "  → Linux arm64"
	@GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 -ldflags="-s -w -X main.version=$(VERSION)" ./$(CMD_DIR)

	@echo "  → macOS amd64"
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 -ldflags="-s -w -X main.version=$(VERSION)" ./$(CMD_DIR)

	@echo "  → macOS arm64"
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 -ldflags="-s -w -X main.version=$(VERSION)" ./$(CMD_DIR)

	@echo "  → Windows amd64"
	@GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe -ldflags="-s -w -X main.version=$(VERSION)" ./$(CMD_DIR)

	@echo "  → Generating checksums"
	@cd $(BUILD_DIR) && sha256sum * > checksums.txt

	@echo "✓ Release build complete: ./$(BUILD_DIR)/"

run-example: build ## Build and run with example file
	@echo "Running example..."
	./$(BINARY_NAME) examples/simple/tfplan.json

all: clean deps lint test build ## Run all checks and build
	@echo "✓ All tasks complete"
