# user-service Makefile

# Local configuration
LOCAL_BIN := $(CURDIR)/bin
SWAGGER_UI_VERSION := v5.4.2
SWAGGER_UI_DIR := ./static/swagger-ui

# Reference to project tools
PROJECT_ROOT := $(CURDIR)/..
PROJECT_BIN := $(PROJECT_ROOT)/bin
GOLANGCI_CONFIG := $(PROJECT_ROOT)/.golangci.yaml
PROTOLINT_CONFIG := $(PROJECT_ROOT)/.protolint.yaml

# Ensure project tools are available
export PATH := $(PROJECT_BIN):$(LOCAL_BIN):$(PATH)

.PHONY: help build test proto lint clean

help:
	@echo "user-service commands:"
	@echo "  build            - Build the service"
	@echo "  test             - Run tests"
	@echo "  test-cover       - Run tests with coverage"
	@echo "  proto            - Generate protocol buffers"
	@echo "  lint             - Run linters"
	@echo "  lint-go          - Run Go linters"
	@echo "  lint-proto       - Run Proto linters"
	@echo "  fmt              - Format code"
	@echo "  clean            - Clean build artifacts"
	@echo "  clean-all        - Deep clean"

# Build the service
.PHONY: build
build:
	@echo "Building user-service..."
	@mkdir -p bin
	@go build -o bin/server cmd/server/main.go

# Run tests
.PHONY: test
test:
	@echo "Running user-service tests..."
	@go test ./...

.PHONY: test-cover
test-cover:
	@echo "Running user-service tests with coverage..."
	@go test -cover ./...

# Protocol buffers generation
.PHONY: proto
proto: vendor .bin-deps .buf-generate proto-format generate-swagger

.PHONY: .buf-generate
.buf-generate:
	@echo "Generating protobuf for user-service..."
	@$(PROJECT_BIN)/buf generate

.PHONY: proto-format
proto-format:
	@echo "Formatting proto files..."
	@$(PROJECT_BIN)/buf format -w

# Protocol buffer dependencies - using buf modules instead of git clone!
.PHONY: vendor
vendor: buf-mod-download

.PHONY: buf-mod-download
buf-mod-download:
	@echo "Downloading proto dependencies via buf..."
	@$(PROJECT_BIN)/buf dep update
	@echo "✓ Proto dependencies downloaded"

# Install local binary dependencies
.PHONY: .bin-deps
.bin-deps: export GOBIN := $(LOCAL_BIN)
.bin-deps:
	@echo "Installing local protoc plugins for user-service..."
	@mkdir -p $(LOCAL_BIN)
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

# Swagger generation
.PHONY: generate-swagger
generate-swagger: generate-swagger-api install-swagger-ui
	@echo "Generating Swagger documentation..."
	@cp ./api/openapi/api.yaml $(SWAGGER_UI_DIR)/swagger.yaml
	@awk '{gsub("https://petstore.swagger.io/v2/swagger.json", "./swagger.yaml")}1' $(SWAGGER_UI_DIR)/swagger-initializer.js > $(SWAGGER_UI_DIR)/swagger-initializer.js.tmp && mv $(SWAGGER_UI_DIR)/swagger-initializer.js.tmp $(SWAGGER_UI_DIR)/swagger-initializer.js

.PHONY: generate-swagger-api
generate-swagger-api: .ensure-dir
	@echo "Generating API code..."
	@export PATH="$$HOME/go/bin:$$PATH" && \
	which oapi-codegen > /dev/null || go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest && \
	oapi-codegen -config api/openapi/cfg.yaml api/openapi/api.yaml > gen/api/openapi.gen.go
	@echo "Installing generated code dependencies..."
	@cd .. && go work sync

.PHONY: install-swagger-ui
install-swagger-ui:
	@echo "Installing Swagger UI..."
	@mkdir -p $(SWAGGER_UI_DIR)
	@TEMP_DIR=$$(mktemp -d); \
	curl -sSL https://codeload.github.com/swagger-api/swagger-ui/tar.gz/refs/tags/$(SWAGGER_UI_VERSION) | tar -xz -C $$TEMP_DIR; \
	cp -R $$TEMP_DIR/swagger-ui-$(SWAGGER_UI_VERSION:v%=%)/dist/* $(SWAGGER_UI_DIR); \
	rm -rf $$TEMP_DIR

.PHONY: .ensure-dir
.ensure-dir:
	@mkdir -p gen/api/

# Linting using shared configurations
.PHONY: lint
lint: lint-go lint-proto

.PHONY: lint-go
lint-go:
	@echo "Running Go linters for user-service..."
	@$(PROJECT_BIN)/golangci-lint run --config $(GOLANGCI_CONFIG) --disable=typecheck ./internal/... ./cmd/... 2>&1 | grep -vE "(file requires newer Go version|w\.Context undefined|package http2|messageReceived\.Event|messageSent\.Event|\^|golang\.org/x/net|go\.opentelemetry\.io)" || true

.PHONY: lint-proto
lint-proto:
	@echo "Running Proto linters for user-service..."
	@$(PROJECT_BIN)/protolint -config_path $(PROTOLINT_CONFIG) ./proto/ || true

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting user-service code..."
	@go fmt ./...

.PHONY: gogenerate
gogenerate:
	@go generate ./...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning user-service artifacts..."
	@rm -rf bin vendor.protobuf gen static/swagger-ui

.PHONY: clean-all
clean-all: clean
	@echo "Deep cleaning user-service..."
	@find . -name "*.test" -delete
	@find . -name "*.out" -delete
	@find . -name "coverage.out" -delete
