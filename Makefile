# Makefile for EESA (Efficacious Executive Summary Assistant)
# Cross-platform desktop application built with Go and Fyne

# Variables
BINARY_NAME=eesa
MAIN_PATH=./cmd/eesa
MODULE=$(shell go list -m)
VERSION ?= 1.0.0
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT_HASH=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commitHash=$(COMMIT_HASH)"

# Directories
BUILD_DIR=build
DIST_DIR=dist
COVERAGE_DIR=coverage

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Colors for output removed for compatibility

# Default target
.PHONY: all
all: clean deps fmt vet test build

# Help target
.PHONY: help
help:
	@echo "EESA Development Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  help           - Show this help message"
	@echo ""
	@echo "Development:"
	@echo "  build          - Build the application for current platform"
	@echo "  run            - Run the application"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Download and tidy dependencies"
	@echo ""
	@echo "Testing:"
	@echo "  test           - Run all tests"
	@echo "  test-short     - Run tests excluding integration tests"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  test-coverage-html - Generate HTML coverage report"
	@echo "  benchmark      - Run benchmarks"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt            - Format code"
	@echo "  vet            - Run go vet"
	@echo "  lint           - Run golangci-lint (if available)"
	@echo "  security       - Run security scan (if gosec available)"
	@echo "  quality        - Run all quality checks"
	@echo ""
	@echo "Cross-Platform Builds:"
	@echo "  build-all      - Build for current platform (Fyne GUI limitation)"
	@echo "  build-current  - Build for current platform"
	@echo "  build-docker-linux - Build for Linux using Docker"
	@echo "  build-docker-windows - Build for Windows using Docker"
	@echo "  build-linux-simple - Simple Linux build (may fail)"
	@echo "  build-windows-simple - Simple Windows build (may fail)"
	@echo "  build-macos-simple - Simple macOS build (may fail)"
	@echo ""
	@echo "Distribution:"
	@echo "  dist           - Create distribution package for current platform"
	@echo "  dist-all       - Create distribution packages for all platforms (requires Docker)"
	@echo "  dist-clean     - Clean distribution artifacts"
	@echo ""
	@echo "Utilities:"
	@echo "  install-tools  - Install development tools"
	@echo "  update-deps    - Update all dependencies"
	@echo "  mod-graph      - Show dependency graph"
	@echo "  check-env      - Check development environment"

# Build targets
.PHONY: build
build: create-build-dir
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: build-debug
build-debug: create-build-dir
	@echo "Building $(BINARY_NAME) with debug symbols..."
	$(GOBUILD) -gcflags="all=-N -l" -o $(BUILD_DIR)/$(BINARY_NAME)-debug $(MAIN_PATH)

# Cross-platform builds
# Note: Fyne GUI applications require CGO and platform-specific libraries
# Cross-compilation requires either native libraries or Docker environments

.PHONY: build-all
build-all: build-current build-docker-warning
	@echo "All available platform builds complete"

.PHONY: build-current
build-current: create-build-dir
	@echo "Building for current platform ($(shell go env GOOS)/$(shell go env GOARCH))..."
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(shell go env GOOS)-$(shell go env GOARCH) $(MAIN_PATH)

.PHONY: build-docker-warning
build-docker-warning:
	@echo ""
	@echo "WARNING: Cross-platform builds for Fyne GUI applications require:"
	@echo "1. Native system libraries for each target platform"
	@echo "2. Docker containers with appropriate build environments"
	@echo "3. Platform-specific build systems"
	@echo ""
	@echo "Use 'make build-docker-*' targets for proper cross-compilation"
	@echo "or build natively on each target platform."

# Docker-based cross-compilation (recommended for Fyne apps)
.PHONY: build-docker-linux
build-docker-linux: create-build-dir
	@echo "Building for Linux using Docker..."
	@if command -v docker >/dev/null 2>&1; then \
		docker run --rm -v $(PWD):/app -w /app golang:1.23-alpine sh -c \
			"apk add --no-cache gcc musl-dev pkgconfig mesa-dev libxrandr-dev libxcursor-dev libxinerama-dev libxi-dev libxext-dev libxfixes-dev && \
			go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)"; \
	else \
		echo "Docker not found. Install Docker to use this target."; \
		exit 1; \
	fi

.PHONY: build-docker-windows
build-docker-windows: create-build-dir
	@echo "Building for Windows using Docker..."
	@if command -v docker >/dev/null 2>&1; then \
		docker run --rm -v $(PWD):/app -w /app golang:1.23-windowsservercore sh -c \
			"go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)"; \
	else \
		echo "Docker not found. Install Docker to use this target."; \
		exit 1; \
	fi

# Simple cross-compilation (may fail due to CGO dependencies)
.PHONY: build-linux-simple
build-linux-simple: create-build-dir
	@echo "Attempting simple cross-compilation for Linux..."
	@echo "Note: This may fail due to CGO dependencies. Use build-docker-linux instead."
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

.PHONY: build-windows-simple
build-windows-simple: create-build-dir
	@echo "Attempting simple cross-compilation for Windows..."
	@echo "Note: This may fail due to CGO dependencies. Use build-docker-windows instead."
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

.PHONY: build-macos-simple
build-macos-simple: create-build-dir
	@echo "Attempting simple cross-compilation for macOS Intel..."
	@echo "Note: This may fail due to CGO dependencies."
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)

.PHONY: build-macos-arm-simple
build-macos-arm-simple: create-build-dir
	@echo "Attempting simple cross-compilation for macOS ARM64..."
	@echo "Note: This may fail due to CGO dependencies."
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

# Run targets
.PHONY: run
run:
	@echo "Running $(BINARY_NAME)..."
	$(GOCMD) run $(MAIN_PATH)

.PHONY: run-debug
run-debug:
	@echo "Running $(BINARY_NAME) with debug logging..."
	ESA_LOG_LEVEL=debug $(GOCMD) run $(MAIN_PATH)

# Test targets
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

.PHONY: test-short
test-short:
	@echo "Running short tests (excluding integration)..."
	$(GOTEST) -short -v ./...

.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v ./internal/mocks/

.PHONY: test-coverage
test-coverage: create-coverage-dir
	@echo "Running tests with coverage..."
	$(GOTEST) -cover -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GOCMD) tool cover -func=$(COVERAGE_DIR)/coverage.out
	@echo "Coverage report saved to $(COVERAGE_DIR)/coverage.out"

.PHONY: test-coverage-html
test-coverage-html: test-coverage
	@echo "Generating HTML coverage report..."
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "HTML coverage report: $(COVERAGE_DIR)/coverage.html"

.PHONY: benchmark
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Code quality targets
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

.PHONY: vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

.PHONY: lint
lint:
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with: make install-tools"; \
	fi

.PHONY: security
security:
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not found. Install with: make install-tools"; \
	fi

.PHONY: quality
quality: fmt vet lint security
	@echo "All quality checks complete"

# Dependency management
.PHONY: deps
deps:
	@echo "Downloading and tidying dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

.PHONY: update-deps
update-deps:
	@echo "Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy

.PHONY: mod-graph
mod-graph:
	@echo "Dependency graph:"
	$(GOMOD) graph

# Distribution targets
.PHONY: dist
dist: clean build-current create-dist-dir
	@echo "Creating distribution package for current platform..."
	# Current platform
	@if [ -f "$(BUILD_DIR)/$(BINARY_NAME)-$(shell go env GOOS)-$(shell go env GOARCH)" ]; then \
		if [ "$(shell go env GOOS)" = "windows" ]; then \
			zip -j $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-$(shell go env GOOS)-$(shell go env GOARCH).zip $(BUILD_DIR)/$(BINARY_NAME)-$(shell go env GOOS)-$(shell go env GOARCH); \
		else \
			tar -czf $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-$(shell go env GOOS)-$(shell go env GOARCH).tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-$(shell go env GOOS)-$(shell go env GOARCH); \
		fi; \
	else \
		echo "No build found for current platform. Run 'make build-current' first."; \
		exit 1; \
	fi
	@echo "Distribution package created in $(DIST_DIR)/"

.PHONY: dist-all
dist-all: clean create-dist-dir
	@echo "Creating distribution packages for all platforms..."
	@echo "Note: This requires Docker for cross-compilation"
	@echo "Building for Linux..."
	@$(MAKE) build-docker-linux
	@echo "Building for Windows..."
	@$(MAKE) build-docker-windows
	@echo "Packaging distributions..."
	@if [ -f "$(BUILD_DIR)/$(BINARY_NAME)-linux-amd64" ]; then \
		tar -czf $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-linux-amd64; \
	fi
	@if [ -f "$(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe" ]; then \
		zip -j $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe; \
	fi
	@echo "Distribution packages created in $(DIST_DIR)/"

.PHONY: dist-clean
dist-clean:
	@echo "Cleaning distribution artifacts..."
	rm -rf $(DIST_DIR)

# Utility targets
.PHONY: install-tools
install-tools:
	@echo "Installing development tools..."
	@echo "Installing golangci-lint..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin; \
	fi
	@echo "Installing gosec..."
	@if ! command -v gosec >/dev/null 2>&1; then \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi

.PHONY: check-env
check-env:
	@echo "Checking development environment..."
	@echo "Go version: $(shell go version)"
	@echo "Module: $(MODULE)"
	@echo "GOPATH: $(shell go env GOPATH)"
	@echo "GOROOT: $(shell go env GOROOT)"
	@echo "Platform: $(shell go env GOOS)/$(shell go env GOARCH)"
	@echo ""
	@echo "Available tools:"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "✓ golangci-lint: $(shell golangci-lint version)"; \
	else \
		echo "✗ golangci-lint not found"; \
	fi
	@if command -v gosec >/dev/null 2>&1; then \
		echo "✓ gosec: $(shell gosec --version)"; \
	else \
		echo "✗ gosec not found"; \
	fi

# Clean targets
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(COVERAGE_DIR)

.PHONY: clean-all
clean-all: clean dist-clean
	@echo "All artifacts cleaned"

# Directory creation rules
.PHONY: create-build-dir create-dist-dir create-coverage-dir

create-build-dir:
	@mkdir -p $(BUILD_DIR)

create-dist-dir:
	@mkdir -p $(DIST_DIR)

create-coverage-dir:
	@mkdir -p $(COVERAGE_DIR)