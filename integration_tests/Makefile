# Makefile for integration_tests

# Try to use parent's golangci-lint, fallback to global
GOLANGCI_LINT := $(shell if [ -f ../bin/golangci-lint ]; then echo ../bin/golangci-lint; else echo golangci-lint; fi)

.PHONY: help lint fmt clean test

help:
	@echo "Available targets:"
	@echo "  lint         - Run golangci-lint"
	@echo "  fmt          - Format code"
	@echo "  test         - Run integration tests"
	@echo "  clean        - Clean test artifacts"

# Run linter
.PHONY: lint
lint:
	@echo "Running golangci-lint for integration_tests..."
	@$(GOLANGCI_LINT) run ./... --timeout=5m

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting integration_tests..."
	@go fmt ./...
	@goimports -w . || echo "goimports not found"

# Run integration tests
.PHONY: test
test:
	@echo "Running integration tests..."
	@go test -v ./suites/...

# Alias for integration tests (CI compatibility)
.PHONY: integration-test
integration-test: test

# Run specific test suite
.PHONY: test-suite
test-suite:
	@echo "Running test suite: $(SUITE)"
	@go test -v ./suites/ -run $(SUITE)

# Performance tests
.PHONY: performance-test
performance-test:
	@echo "Running performance tests..."
	@go test -v ./performance/...

# Stress tests
.PHONY: stress-test
stress-test:
	@echo "Running stress tests..."
	@go test -v ./performance/ -run Stress

# Setup dependencies
.PHONY: setup
setup:
	@echo "Setting up integration test dependencies..."
	@go mod download

# Setup integration test environment
.PHONY: deps
deps: setup
	@echo "Integration test dependencies ready"

# Clean artifacts
.PHONY: clean
clean:
	@echo "Cleaning integration test artifacts..."
	@rm -rf coverage/ test-results/
	@go clean -testcache

# CI pipeline for integration tests
.PHONY: ci
ci: lint test
	@echo "Integration tests CI completed!"
