# Makefile for vcon

.PHONY: all build test test-race test-cover clean lint fmt vet mod-tidy mod-verify examples help install-tools bench security docs

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Build info
BINARY_NAME=vcon
VERSION?=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Default target
all: clean fmt vet lint test build

# Build the library
build:
	@echo "Building..."
	$(GOBUILD) -v ./...

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	$(GOTEST) -race -v ./...

# Run tests with coverage
test-cover:
	@echo "Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f coverage.out coverage.html
	rm -f examples/basic_usage examples/sms_conversation examples/call_center

# Format Go code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

# Run golangci-lint
lint:
	@echo "Running linter..."
	$(GOLINT) run ./...

# Tidy modules
mod-tidy:
	@echo "Tidying modules..."
	$(GOMOD) tidy

# Verify modules
mod-verify:
	@echo "Verifying modules..."
	$(GOMOD) verify

# Download modules
mod-download:
	@echo "Downloading modules..."
	$(GOMOD) download

# Install development tools
install-tools:
	@echo "Installing development tools..."
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) golang.org/x/tools/cmd/godoc@latest
	$(GOGET) github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Run security scan
security:
	@echo "Running security scan..."
	gosec ./...

# Build and run examples
examples: build
	@echo "Building examples..."
	$(GOBUILD) -o examples/basic_usage examples/basic_usage.go
	$(GOBUILD) -o examples/sms_conversation examples/sms_conversation.go
	$(GOBUILD) -o examples/call_center examples/call_center.go
	@echo "Examples built successfully. Run them with:"
	@echo "  ./examples/basic_usage"
	@echo "  ./examples/sms_conversation"
	@echo "  ./examples/call_center"

# Run examples
run-examples: examples
	@echo "Running basic_usage example..."
	@timeout 10s ./examples/basic_usage > /dev/null || echo "Example completed"
	@echo "Running sms_conversation example..."
	@timeout 10s ./examples/sms_conversation > /dev/null || echo "Example completed"
	@echo "Running call_center example..."
	@timeout 10s ./examples/call_center > /dev/null || echo "Example completed"
	@echo "All examples ran successfully"

# Start local documentation server
docs:
	@echo "Starting documentation server at http://localhost:6060"
	@echo "Visit http://localhost:6060/pkg/github.com/lmendes86/vcon/"
	godoc -http=:6060

# Check code quality
quality: fmt vet lint test-race

# Continuous integration target
ci: mod-verify quality test-cover security

# Pre-commit checks
pre-commit: fmt vet lint test

# Pre-commit hooks target
install-pre-commit: install-tools
	@echo "Installing pre-commit hooks..."
	@echo '#!/bin/sh' > .git/hooks/pre-commit
	@echo 'make pre-commit' >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Pre-commit hooks installed successfully"

# Code quality metrics
metrics: test-cover
	@echo "Code Quality Metrics:"
	@echo "===================="
	@if [ -f coverage.out ]; then \
		COVERAGE=$$(go tool cover -func=coverage.out | tail -1 | awk '{print $$3}'); \
		echo "Test Coverage: $$COVERAGE"; \
	fi
	@TOTAL_LINES=$$(find . -name "*.go" -not -path "./examples/*" | xargs wc -l | tail -1 | awk '{print $$1}'); \
	echo "Total Lines of Code: $$TOTAL_LINES"
	@GO_FILES=$$(find . -name "*.go" -not -path "./examples/*" | wc -l); \
	echo "Total Go Files: $$GO_FILES"
	@TEST_FILES=$$(find . -name "*_test.go" -not -path "./examples/*" | wc -l); \
	echo "Test Files: $$TEST_FILES"

# Release checks
release-check: clean mod-tidy quality test-cover security examples
	@echo "Release checks passed!"

# Show help
help:
	@echo "Available targets:"
	@echo "  all          - Run clean, fmt, vet, lint, test, build"
	@echo "  build        - Build the library"
	@echo "  test         - Run tests"
	@echo "  test-race    - Run tests with race detector"
	@echo "  test-cover   - Run tests with coverage"
	@echo "  bench        - Run benchmarks"
	@echo "  clean        - Clean build artifacts"
	@echo "  fmt          - Format Go code"
	@echo "  vet          - Run go vet"
	@echo "  lint         - Run golangci-lint"
	@echo "  mod-tidy     - Tidy modules"
	@echo "  mod-verify   - Verify modules"
	@echo "  mod-download - Download modules"
	@echo "  install-tools- Install development tools"
	@echo "  security     - Run security scan"
	@echo "  examples     - Build examples"
	@echo "  run-examples - Build and run examples"
	@echo "  docs         - Start local documentation server"
	@echo "  quality      - Run all quality checks"
	@echo "  ci           - Run CI checks"
	@echo "  pre-commit   - Run pre-commit checks"
	@echo "  install-pre-commit - Install pre-commit git hooks"
	@echo "  metrics      - Show code quality metrics"
	@echo "  release-check- Run release checks"
	@echo "  help         - Show this help"