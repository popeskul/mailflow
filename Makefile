# Makefile for mailflow project

# Binary dependencies management
LOCAL_BIN := $(CURDIR)/bin
export PATH := $(LOCAL_BIN):$(PATH)

# Tool versions
GOLANGCI_VERSION := v1.64.8
BUF_VERSION := v1.32.2
PROTOC_VERSION := 25.1

# Tool paths
GOLANGCI_LINT := $(LOCAL_BIN)/golangci-lint
BUF := $(LOCAL_BIN)/buf
PROTOC := $(LOCAL_BIN)/protoc
PROTOC_GEN_GO := $(LOCAL_BIN)/protoc-gen-go
PROTOC_GEN_GO_GRPC := $(LOCAL_BIN)/protoc-gen-go-grpc
PROTOC_GEN_GRPC_GATEWAY := $(LOCAL_BIN)/protoc-gen-grpc-gateway
PROTOC_GEN_OPENAPIV2 := $(LOCAL_BIN)/protoc-gen-openapiv2
PROTOLINT := $(LOCAL_BIN)/protolint

# Services
SERVICES := user-service email-service

# Parallel execution support
NPROC := $(shell nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)
JOBS ?= $(NPROC)
MAKEFLAGS += --jobs=$(JOBS)

# Helper function for parallel execution
define run_parallel
	@echo "$(1) (using $(JOBS) parallel jobs)..."
	@echo "Services to process: $(SERVICES)"
	@set -e; \
	pids=""; \
	start_time=$$(date +%s); \
	for service in $(SERVICES); do \
		echo "[$$start_time] Starting $(2) for $$service (PID will be captured)..."; \
		$(MAKE) -C $$service $(2) & \
		pid=$$!; \
		echo "[$$start_time] $$service $(2) started with PID $$pid"; \
		pids="$$pids $$pid"; \
	done; \
	echo "Waiting for processes: $$pids"; \
	for pid in $$pids; do \
		echo "Waiting for PID $$pid..."; \
		wait $$pid || exit 1; \
		echo "PID $$pid completed"; \
	done; \
	end_time=$$(date +%s); \
	duration=$$((end_time - start_time)); \
	echo "✓ $(1) completed successfully in $${duration}s"
endef

# OS detection for protoc download
UNAME_OS := $(shell uname -s)
UNAME_ARCH := $(shell uname -m)

ifeq ($(UNAME_OS),Darwin)
	PROTOC_OS := osx
ifeq ($(UNAME_ARCH),arm64)
	PROTOC_ARCH := aarch_64
else
	PROTOC_ARCH := x86_64
endif
else ifeq ($(UNAME_OS),Linux)
	PROTOC_OS := linux
	PROTOC_ARCH := x86_64
endif

PROTOC_ZIP := protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip

.PHONY: help build test run clean docker-up docker-down proto lint

# Default target
help:
	@echo "Available targets:"
	@echo ""
	@echo "QUICK START:"
	@echo "  quick-start      - Full setup: install tools, build, and start services"
	@echo ""
	@echo "BUILD & RUN:"
	@echo "  build            - Build all services"
	@echo "  docker-build     - Build docker images"
	@echo "  docker-up        - Start all services with docker-compose"
	@echo "  docker-down      - Stop all services"
	@echo "  run              - Run services locally"
	@echo ""
	@echo "TESTING:"
	@echo "  test             - Run all tests"
	@echo "  test-cover       - Run tests with coverage"
	@echo "  coverage-html    - Generate HTML coverage reports"
	@echo "  coverage-view    - Generate and open coverage report"
	@echo "  create-user      - Create a test user via API"
	@echo ""
	@echo "MONITORING:"
	@echo "  monitor-metrics  - Open Prometheus, Grafana, Jaeger"
	@echo "  monitor-live     - Live metrics monitoring in terminal"
	@echo "  status           - Show service status"
	@echo "  health-check     - Check services health"
	@echo ""
	@echo "LOGS & DEBUG:"
	@echo "  docker-logs      - Show all logs"
	@echo "  logs-user        - Show user-service logs"
	@echo "  logs-email       - Show email-service logs"
	@echo "  logs-gateway     - Show API gateway logs"
	@echo ""
	@echo "MAINTENANCE:"
	@echo "  clean            - Clean build artifacts"
	@echo "  clean-all        - Deep clean (including Docker volumes)"
	@echo "  docker-clean     - Clean Docker artifacts"
	@echo "  reset            - Reset project to initial state"
	@echo "  rebuild          - Full rebuild from scratch"
	@echo ""
	@echo "DEVELOPMENT:"
	@echo "  proto            - Generate protocol buffers for all services (parallel)"
	@echo "  lint             - Run linters for all services (parallel)"
	@echo "  lint-go          - Run Go linters (parallel)"
	@echo "  lint-proto       - Run Proto linters (parallel)"
	@echo "  lint-all         - Run linters for services"
	@echo "  fmt              - Format code for all services (parallel)"
	@echo "  fmt-all          - Format code for services"
	@echo "  tools            - Install development tools"
	@echo "  versions         - Show tool versions"
	@echo ""
	@echo "PARALLEL CONTROL:"
	@echo "  build-seq        - Build all services (sequential)"
	@echo "  test-seq         - Run tests (sequential)"
	@echo "  lint-seq         - Run linters (sequential)"
	@echo "  show-jobs        - Show current parallel job count"
	@echo ""
	@echo "EXAMPLES:"
	@echo "  make JOBS=2 build    - Build with 2 parallel jobs"
	@echo "  make JOBS=1 lint     - Run linters sequentially"

# Create bin directory
$(LOCAL_BIN):
	@mkdir -p $(LOCAL_BIN)

# Install golangci-lint
$(GOLANGCI_LINT): | $(LOCAL_BIN)
	@echo "Installing golangci-lint $(GOLANGCI_VERSION)..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(LOCAL_BIN) $(GOLANGCI_VERSION)

# Install buf
$(BUF): | $(LOCAL_BIN)
	@echo "Installing buf $(BUF_VERSION)..."
	@GOBIN=$(LOCAL_BIN) go install github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION)

# Install protoc
$(PROTOC): | $(LOCAL_BIN)
	@echo "Installing protoc $(PROTOC_VERSION)..."
	@curl -sLO https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/$(PROTOC_ZIP)
	@unzip -q $(PROTOC_ZIP) -d $(LOCAL_BIN)/protoc-tmp
	@mv $(LOCAL_BIN)/protoc-tmp/bin/protoc $(LOCAL_BIN)/
	@rm -rf $(LOCAL_BIN)/protoc-tmp $(PROTOC_ZIP)
	@chmod +x $(PROTOC)

# Install protoc plugins
$(PROTOC_GEN_GO): | $(LOCAL_BIN)
	@echo "Installing protoc-gen-go..."
	@GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

$(PROTOC_GEN_GO_GRPC): | $(LOCAL_BIN)
	@echo "Installing protoc-gen-go-grpc..."
	@GOBIN=$(LOCAL_BIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

$(PROTOC_GEN_GRPC_GATEWAY): | $(LOCAL_BIN)
	@echo "Installing protoc-gen-grpc-gateway..."
	@GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

$(PROTOC_GEN_OPENAPIV2): | $(LOCAL_BIN)
	@echo "Installing protoc-gen-openapiv2..."
	@GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

$(PROTOLINT): | $(LOCAL_BIN)
	@echo "Installing protolint..."
	@GOBIN=$(LOCAL_BIN) go install github.com/yoheimuta/protolint/cmd/protolint@latest

# Install all tools
.PHONY: tools
tools: $(GOLANGCI_LINT) $(BUF) $(PROTOC) $(PROTOC_GEN_GO) $(PROTOC_GEN_GO_GRPC) $(PROTOC_GEN_GRPC_GATEWAY) $(PROTOC_GEN_OPENAPIV2) $(PROTOLINT)
	@echo "All tools installed successfully!"

# Build all services
.PHONY: build
build:
	$(call run_parallel,Building all services,build)

# Run tests for all services
.PHONY: test
test:
	$(call run_parallel,Running tests for all services,test)

.PHONY: test-cover
test-cover:
	$(call run_parallel,Running tests with coverage for all services,test-cover)

.PHONY: coverage-html
coverage-html:
	@echo "Generating HTML coverage reports..."
	@./coverage_report.sh

.PHONY: coverage-view
coverage-view: coverage-html
	@echo "Opening coverage report..."
	@open coverage/html/combined.html

# Centralized vendor dependencies download
# Generate protocol buffers for all services
.PHONY: proto
proto: $(BUF) $(PROTOC_GEN_GO) $(PROTOC_GEN_GO_GRPC) $(PROTOC_GEN_GRPC_GATEWAY) $(PROTOC_GEN_OPENAPIV2)
	@echo "Generating protocol buffers sequentially..."
	@for service in $(SERVICES); do \
		echo "Generating proto for $$service..."; \
		$(MAKE) -C $$service proto || exit 1; \
		echo "✓ Proto generated for $$service"; \
	done
	@echo "✓ Protocol buffers generated for all services"
	@$(MAKE) tidy

# Run linters for all services
.PHONY: lint
lint: lint-go lint-proto

.PHONY: lint-go
lint-go: $(GOLANGCI_LINT)
	$(call run_parallel,Running Go linters for all services,lint-go)

.PHONY: lint-proto
lint-proto: $(PROTOLINT)
	$(call run_parallel,Running Proto linters for all services,lint-proto)

# Lint everything (services only)
.PHONY: lint-all
lint-all: lint

# Format code for all services
.PHONY: fmt
fmt:
	$(call run_parallel,Formatting code for all services,fmt)

# Format everything (services only)
.PHONY: fmt-all
fmt-all: fmt

# Docker commands
.PHONY: docker-up
docker-up:
	@echo "Starting services with docker-compose..."
	@docker-compose up -d

.PHONY: docker-down
docker-down:
	@echo "Stopping services..."
	@docker-compose down

.PHONY: docker-build
docker-build:
	@echo "Building docker images..."
	@docker-compose build

.PHONY: docker-logs
docker-logs:
	@docker-compose logs -f

.PHONY: docker-ps
docker-ps:
	@docker-compose ps

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@$(foreach service,$(SERVICES),$(MAKE) -C $(service) clean &&) true
	@rm -rf $(LOCAL_BIN)
	@rm -rf common/vendor.protobuf

# Deep clean - removes all generated files, docker volumes, etc.
.PHONY: clean-all
clean-all: clean docker-down
	@echo "Performing deep clean..."
	@docker-compose down -v --remove-orphans
	@docker system prune -f
	@$(foreach service,$(SERVICES),$(MAKE) -C $(service) clean-all &&) true
	@find . -name ".DS_Store" -delete
	@find . -name "*.log" -delete
	@echo "✓ Deep clean completed"

# Sync Go workspace after code generation
.PHONY: tidy
tidy:
	@echo "Syncing Go workspace..."
	@go work sync
	@echo "✓ Go workspace synced"

# Remove all Docker artifacts
.PHONY: docker-clean
docker-clean:
	@echo "Cleaning Docker artifacts..."
	@docker-compose down -v --remove-orphans
	@docker rmi mailflow-user-service mailflow-email-service -f 2>/dev/null || true
	@docker system prune -f
	@echo "✓ Docker cleanup completed"

# Reset project to initial state
.PHONY: reset
reset: clean-all
	@echo "Resetting project to initial state..."
	@git clean -fdx -e .idea -e .vscode
	@echo "✓ Project reset completed"

# Development helpers
.PHONY: dev-setup
dev-setup: tools
	@echo "Development environment is ready!"

# Monitoring
.PHONY: monitor-metrics
monitor-metrics:
	@echo "Opening monitoring dashboards..."
	@open http://localhost:9090 2>/dev/null || xdg-open http://localhost:9090 2>/dev/null || echo "Prometheus: http://localhost:9090"
	@open http://localhost:3000 2>/dev/null || xdg-open http://localhost:3000 2>/dev/null || echo "Grafana: http://localhost:3000"
	@open http://localhost:16686 2>/dev/null || xdg-open http://localhost:16686 2>/dev/null || echo "Jaeger: http://localhost:16686"

# Health checks
.PHONY: health-check
health-check:
	@echo "Checking service health..."
	@curl -sf http://localhost:9101/metrics > /dev/null && echo "✓ User service is healthy" || echo "✗ User service is down"
	@curl -sf http://localhost:9102/metrics > /dev/null && echo "✓ Email service is healthy" || echo "✗ Email service is down"
	@curl -sf http://localhost:8000/__health > /dev/null && echo "✓ API Gateway is healthy" || echo "✗ API Gateway is down"

# Quick start
.PHONY: quick-start
quick-start: tools docker-build docker-up
	@echo "Waiting for services to start..."
	@sleep 10
	@$(MAKE) health-check
	@echo ""
	@echo "Services are ready! You can now:"
	@echo "  - View metrics: make monitor-metrics"
	@echo "  - Check logs: make docker-logs"

# Show all running services
.PHONY: status
status:
	@echo "Service Status:"
	@docker-compose ps
	@echo ""
	@$(MAKE) health-check

# Restart specific service
.PHONY: restart-user
restart-user:
	@docker-compose restart user-service

.PHONY: restart-email
restart-email:
	@docker-compose restart email-service

.PHONY: restart-all
restart-all:
	@docker-compose restart

# View logs for specific services
.PHONY: logs-user
logs-user:
	@docker-compose logs -f user-service

.PHONY: logs-email
logs-email:
	@docker-compose logs -f email-service

.PHONY: logs-gateway
logs-gateway:
	@docker-compose logs -f krakend

# Create a user via API Gateway
.PHONY: create-user
create-user:
	@curl -X POST http://localhost:8000/user/create \
		-H "Content-Type: application/json" \
		-d '{"username": "testuser", "email": "test@example.com"}' | jq . 2>/dev/null || \
	curl -X POST http://localhost:8000/user/create \
		-H "Content-Type: application/json" \
		-d '{"username": "testuser", "email": "test@example.com"}'

# Development workflow
.PHONY: dev
dev: docker-down docker-build docker-up docker-logs

# Run all tests and checks
.PHONY: check-all
check-all: test lint
	@echo "All checks passed!"

# Show installed tools versions
.PHONY: versions
versions: tools
	@echo "Installed tool versions:"
	@$(GOLANGCI_LINT) --version
	@$(BUF) --version
	@$(PROTOC) --version
	@echo "protoc-gen-go: $$($(PROTOC_GEN_GO) --version 2>&1 | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+')"
	@echo "protolint: $$($(PROTOLINT) version)"

# Full project reset and rebuild
.PHONY: rebuild
rebuild: clean-all tools docker-build docker-up
	@echo "Project rebuilt from scratch!"

# Sequential execution commands (fallback for debugging)
.PHONY: build-seq
build-seq:
	@echo "Building all services (sequential)..."
	@$(foreach service,$(SERVICES),echo "Building $(service)..." && $(MAKE) -C $(service) build &&) true
	@echo "✓ Build completed"

.PHONY: test-seq
test-seq:
	@echo "Running tests for all services (sequential)..."
	@$(foreach service,$(SERVICES),echo "Testing $(service)..." && $(MAKE) -C $(service) test &&) true
	@echo "✓ Tests completed"

.PHONY: lint-seq
lint-seq:
	@echo "Running linters for all services (sequential)..."
	@$(foreach service,$(SERVICES),echo "Linting $(service)..." && $(MAKE) -C $(service) lint &&) true
	@echo "✓ Linting completed"

# Show parallel job configuration
.PHONY: show-jobs
show-jobs:
	@echo "Parallel execution configuration:"
	@echo "  CPU cores detected: $(NPROC)"
	@echo "  Current JOBS setting: $(JOBS)"
	@echo "  Services: $(SERVICES)"
	@echo ""
	@echo "Usage examples:"
	@echo "  make JOBS=1 build    # Sequential execution"
	@echo "  make JOBS=4 test     # 4 parallel jobs"
	@echo "  make build           # Default ($(JOBS) jobs)"

# CI/CD pipeline
.PHONY: ci
ci: tools build test lint
	@echo "CI pipeline completed successfully!"
