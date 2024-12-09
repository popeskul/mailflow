.PHONY: all
all: build run

SERVICE_PREFIX := email-service-platform
USER_SERVICE_IMAGE := $(SERVICE_PREFIX)/user-service:latest
EMAIL_SERVICE_IMAGE := $(SERVICE_PREFIX)/email-service:latest

.PHONY: build
build: build-user build-email

.PHONY: build-user
build-user:
	@echo "Building user service..."
	@cd user-service && make build

.PHONY: build-email
build-email:
	@echo "Building email service..."
	@cd email-service && make build

.PHONY: run
run:
	@echo "Starting all services..."
	docker-compose up -d

.PHONY: down
down:
	@echo "Stopping all services..."
	docker-compose down

.PHONY: gen
gen: gen-user gen-email

.PHONY: gen-user
gen-user:
	@echo "Generating user service protos..."
	@cd user-service && make generate

.PHONY: gen-email
gen-email:
	@echo "Generating email service protos..."
	@cd email-service && make generate

.PHONY: clean
clean:
	@echo "Cleaning up..."
	@cd user-service && make clean
	@cd email-service && make clean
	docker-compose down --volumes --remove-orphans

.PHONY: test
test:
	@echo "Running tests..."
	@cd user-service && make test
	@cd email-service && make test

.PHONY: lint
lint:
	@echo "Running linters..."
	@cd user-service && make lint
	@cd email-service && make lint

.PHONY: monitoring
monitoring:
	@echo "Opening monitoring dashboards..."
	@echo "Grafana: http://localhost:3000"
	@echo "Prometheus: http://localhost:9090"
	@echo "User service metrics: http://localhost:9101/metrics"
	@echo "Email service metrics: http://localhost:9102/metrics"

.PHONY: logs
logs:
	docker-compose logs -f

.PHONY: logs-user
logs-user:
	docker-compose logs -f user-service

.PHONY: logs-email
logs-email:
	docker-compose logs -f email-service

.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make build       - Build all services"
	@echo "  make run         - Run all services with docker-compose"
	@echo "  make down        - Stop all services"
	@echo "  make gen         - Generate proto files for all services"
	@echo "  make test        - Run tests for all services"
	@echo "  make lint        - Run linters for all services"
	@echo "  make clean       - Clean up all services"
	@echo "  make monitoring  - Show monitoring dashboard URLs"
	@echo "  make logs        - Show logs from all services"
	@echo "  make logs-user   - Show logs from user service"
	@echo "  make logs-email  - Show logs from email service"