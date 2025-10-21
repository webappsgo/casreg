.PHONY: all build build-ui clean run dev test lint help

# Build variables
VERSION ?= 1.0.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)

# Directories
BUILD_DIR := build
DIST_DIR := dist
UI_DIR := ui
UI_DIST_DIR := $(UI_DIR)/dist

# Binary names
BINARY_NAME := casreg
BINARY_LINUX_AMD64 := $(BINARY_NAME)-linux-amd64
BINARY_LINUX_ARM64 := $(BINARY_NAME)-linux-arm64
BINARY_DARWIN_AMD64 := $(BINARY_NAME)-darwin-amd64
BINARY_DARWIN_ARM64 := $(BINARY_NAME)-darwin-arm64
BINARY_WINDOWS_AMD64 := $(BINARY_NAME)-windows-amd64.exe

# Default target
all: build

# Help target
help:
	@echo "casreg Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  all            - Build everything (default)"
	@echo "  build          - Build server binary"
	@echo "  build-ui       - Build UI assets"
	@echo "  build-all      - Build for all platforms"
	@echo "  clean          - Clean build artifacts"
	@echo "  run            - Run the server"
	@echo "  dev            - Run in development mode"
	@echo "  test           - Run tests"
	@echo "  lint           - Run linters"
	@echo "  help           - Show this help message"

# Build UI
build-ui:
	@echo "Building UI assets..."
	@cd $(UI_DIR) && npm install && npm run build
	@echo "UI build complete"

# Build server
build: build-ui
	@echo "Building casreg server..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for all platforms
build-all: build-ui
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)

	# Linux AMD64
	@echo "Building for Linux AMD64..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_LINUX_AMD64) .

	# Linux ARM64
	@echo "Building for Linux ARM64..."
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_LINUX_ARM64) .

	# macOS AMD64
	@echo "Building for macOS AMD64..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_DARWIN_AMD64) .

	# macOS ARM64
	@echo "Building for macOS ARM64..."
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_DARWIN_ARM64) .

	# Windows AMD64
	@echo "Building for Windows AMD64..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_WINDOWS_AMD64) .

	@echo "Cross-compilation complete. Binaries in $(DIST_DIR)/"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR) $(UI_DIST_DIR)
	@go clean
	@echo "Clean complete"

# Run the server
run: build
	@echo "Starting casreg server..."
	@$(BUILD_DIR)/$(BINARY_NAME) serve

# Development mode (auto-reload would require additional tools)
dev:
	@echo "Starting development server..."
	@go run . serve

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

# Run linters
lint:
	@echo "Running linters..."
	@golangci-lint run ./...
	@cd $(UI_DIR) && npm run lint

# Install dependencies
deps:
	@echo "Installing Go dependencies..."
	@go mod download
	@go mod tidy
	@echo "Installing UI dependencies..."
	@cd $(UI_DIR) && npm install

# Generate documentation
docs:
	@echo "Generating API documentation..."
	@swag init
	@echo "Documentation generated"

# Docker build
docker-build:
	@echo "Building Docker image..."
	@docker build -t casreg:$(VERSION) -t casreg:latest .
	@echo "Docker image built"

# Docker run
docker-run:
	@echo "Running Docker container..."
	@docker run -p 8080:8080 casreg:latest
