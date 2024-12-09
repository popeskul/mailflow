SWAGGER_UI_VERSION := v5.4.2
SWAGGER_UI_DIR := ./static/swagger-ui
SWAGGER_UI_URL := https://github.com/swagger-api/swagger-ui/archive/v$(SWAGGER_UI_VERSION).tar.gz

PROTOC_GEN_GO := $(shell which protoc-gen-go)
PROTOC_GEN_GO_GRPC := $(shell which protoc-gen-go-grpc)
PROTOC := $(shell which protoc)
PROTO_PATH := $(CURDIR)/proto
PKG_PROTO_PATH := $(CURDIR)/pkg

LOCAL_BIN := $(CURDIR)/bin

VENDOR_PROTO_PATH := $(CURDIR)/vendor.protobuf

vendor: .vendor-reset .vendor-googleapis .vendor-google-protobuf .vendor-protovalidate \
		.vendor-protoc-gen-openapiv2 .vendor-tidy

.vendor-reset:
	rm -rf $(VENDOR_PROTO_PATH)
	mkdir -p $(VENDOR_PROTO_PATH)

.vendor-tidy:
	find $(VENDOR_PROTO_PATH) -type f ! -name "*.proto" -delete
	find $(VENDOR_PROTO_PATH) -empty -type d -delete

# Set up the vendor directory with the google/protobuf proto files
.vendor-google-protobuf:
	git clone -b main --single-branch -n --depth=1 --filter=tree:0 \
		https://github.com/protocolbuffers/protobuf $(VENDOR_PROTO_PATH)/protobuf &&\
	cd $(VENDOR_PROTO_PATH)/protobuf &&\
	git sparse-checkout set --no-cone src/google/protobuf &&\
	git checkout
	mkdir -p $(VENDOR_PROTO_PATH)/google
	mv $(VENDOR_PROTO_PATH)/protobuf/src/google/protobuf $(VENDOR_PROTO_PATH)/google
	rm -rf $(VENDOR_PROTO_PATH)/protobuf

# Set up the vendor directory with the bufbuild/protovalidate proto files
.vendor-protovalidate:
	git clone -b main --single-branch --depth=1 --filter=tree:0 \
		https://github.com/bufbuild/protovalidate $(VENDOR_PROTO_PATH)/protovalidate && \
	cd $(VENDOR_PROTO_PATH)/protovalidate
	git checkout
	mv $(VENDOR_PROTO_PATH)/protovalidate/proto/protovalidate/buf $(VENDOR_PROTO_PATH)
	rm -rf $(VENDOR_PROTO_PATH)/protovalidate

# Set up the vendor directory with the googleapis proto files
.vendor-googleapis:
	git clone -b master --single-branch -n --depth=1 --filter=tree:0 \
		https://github.com/googleapis/googleapis $(VENDOR_PROTO_PATH)/googleapis &&\
	cd $(VENDOR_PROTO_PATH)/googleapis &&\
	git checkout
	mv $(VENDOR_PROTO_PATH)/googleapis/google $(VENDOR_PROTO_PATH)
	rm -rf $(VENDOR_PROTO_PATH)/googleapis

# Set up the vendor directory with the grpc-gateway protoc-gen-openapiv2 proto files
.vendor-protoc-gen-openapiv2:
	git clone -b main --single-branch -n --depth=1 --filter=tree:0 \
 		https://github.com/grpc-ecosystem/grpc-gateway $(VENDOR_PROTO_PATH)/grpc-gateway && \
 	cd $(VENDOR_PROTO_PATH)/grpc-gateway && \
	git sparse-checkout set --no-cone protoc-gen-openapiv2/options && \
	git checkout
	mkdir -p $(VENDOR_PROTO_PATH)/protoc-gen-openapiv2
	mv $(VENDOR_PROTO_PATH)/grpc-gateway/protoc-gen-openapiv2/options $(VENDOR_PROTO_PATH)/protoc-gen-openapiv2
	rm -rf $(VENDOR_PROTO_PATH)/grpc-gatewa

generate: .bin-deps .buf-generate proto-format generate-swagger gogenerate

gogenerate:
	go generate ./...

.PHONY: generate-swagger
generate-swagger: generate-swagger-api install-swagger-ui
	@echo "Generating Swagger documentation..."
	@cp ./api/openapi/api.yaml $(SWAGGER_UI_DIR)/swagger.yaml
	@awk '{gsub("https://petstore.swagger.io/v2/swagger.json", "./swagger.yaml")}1' $(SWAGGER_UI_DIR)/swagger-initializer.js > $(SWAGGER_UI_DIR)/swagger-initializer.js.tmp && mv $(SWAGGER_UI_DIR)/swagger-initializer.js.tmp $(SWAGGER_UI_DIR)/swagger-initializer.js
	@echo "Swagger documentation generated successfully"

.PHONY: generate-swagger-api
generate-swagger-api: .ensure-dir
	@echo "Generating API code..."
	@which oapi-codegen > /dev/null || go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest
	@oapi-codegen -config api/openapi/cfg.yaml api/openapi/api.yaml > gen/api/openapi.gen.go
	@echo "API code generated successfully."

.PHONY: install-swagger-ui
install-swagger-ui:
	@echo "Installing Swagger UI..."
	@mkdir -p $(SWAGGER_UI_DIR)
	@curl -sSL https://github.com/swagger-api/swagger-ui/archive/$(SWAGGER_UI_VERSION).tar.gz | tar -xz -C /tmp
	@cp -R /tmp/swagger-ui-$(SWAGGER_UI_VERSION:v%=%)/dist/* $(SWAGGER_UI_DIR)
	@rm -rf /tmp/swagger-ui-$(SWAGGER_UI_VERSION:v%=%)
	@echo "Swagger UI installed successfully"

.PHONY: .ensure-dir
.ensure-dir:
	@mkdir -p gen/api/

.PHONY: test
test:
	@go test ./...

.buf-generate:
	$(info run buf generate...)
	PATH="$(LOCAL_BIN):$(PATH)" $(LOCAL_BIN)/buf generate

proto-format:
	$(info run buf format...)
	$(LOCAL_BIN)/buf format -w

# Set up the local bin directory for installing protoc plugins
.bin-deps: export GOBIN := $(LOCAL_BIN)
.bin-deps:
	$(info Installing binary dependencies...)

	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go install github.com/bufbuild/buf/cmd/buf@v1.32.2
	go install github.com/yoheimuta/protolint/cmd/protolint@latest

.PHONY: lint
lint:
	@golangci-lint run

lint: .proto-lint

.proto-lint:
	$(LOCAL_BIN)/protolint -config_path ./.protolint.yaml ./proto/
