# MailFlow - Resilient Email Service Platform

A microservices-based email service platform with built-in resilience patterns including Circuit Breaker, Rate Limiting, and Message Queuing.

## Architecture Overview

The platform consists of two main services:

1. **User Service** - Manages user registration and sends welcome emails
2. **Email Service** - Handles email delivery with rate limiting

### Key Features

- **Circuit Breaker Pattern**: Prevents cascading failures when email service is unavailable
- **Rate Limiting**: Controls email sending rate using token bucket algorithm
- **Message Queue**: Persists failed email requests for retry
- **Retry Mechanism**: Exponential backoff for transient failures
- **Service Downtime Simulation**: Email service periodically goes offline for testing
- **Comprehensive Metrics**: RED metrics + custom circuit breaker and queue metrics
- **API Gateway**: KrakenD for unified API access
- **Optimized Build System**: Centralized configurations and efficient development workflow

## Quick Start

### Prerequisites

- Go 1.23+
- Docker and Docker Compose
- Make

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd mailflow
```

2. Install development tools:
```bash
make tools
```

### Running the Services

#### Using Docker (Recommended)

```bash
# Complete setup with monitoring
make quick-start
```

This will start:
- User Service (ports: 8080 HTTP, 50051 gRPC, 9101 metrics)
- Email Service (ports: 50052 gRPC, 9102 metrics)
- KrakenD API Gateway (port: 8000)
- Prometheus (port: 9090)
- Grafana (port: 3000)
- Jaeger (port: 16686)

#### Running Locally

```bash
# Start both services
make docker-up

# Or build and run locally (in separate terminals)
cd user-service && make build && ./bin/server
cd email-service && make build && ./bin/server
```

### Testing the System

1. Create a user (triggers welcome email):
```bash
curl -X POST http://localhost:8000/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "name": "Test User"}'

# Or use the convenient make command
make create-user
```

2. Check service health:
```bash
make health-check
```

3. View logs:
```bash
make docker-logs
# Or specific service logs
make logs-user
make logs-email
make logs-gateway
```

4. Open monitoring dashboards:
```bash
make monitor-metrics
```

## Optimized Development with Make

The project uses an optimized Makefile structure with centralized configurations, efficient workflows, and **parallel execution** for faster builds.

### Parallel Execution Features ⚡

All multi-service operations run in parallel by default:
- **Automatic CPU detection**: Uses all available cores
- **Customizable parallelism**: Control job count with `JOBS` parameter  
- **Fast execution**: Significantly reduces build and test times
- **Error handling**: Stops on first failure across all parallel jobs

```bash
# Default parallel execution (uses all CPU cores)
make build             # Builds all services simultaneously
make test              # Runs tests in parallel
make lint              # Runs linters in parallel

# Custom parallel job count
make JOBS=4 build      # Use 4 parallel jobs
make JOBS=2 test       # Use 2 parallel jobs  
make JOBS=1 lint       # Sequential execution

# Check current configuration
make show-jobs         # Show parallel job settings
```

### Project-Level Commands

```bash
make help              # Show all available commands
make quick-start       # Complete setup: tools + build + start
make build             # Build all services
make test              # Run tests for all services
make lint              # Run linters for all services
make proto             # Generate protocol buffers for all services
make fmt               # Format code for all services
make clean             # Clean build artifacts for all services
make tools             # Install development tools
make versions          # Show installed tool versions
```

### Docker Operations

```bash
make docker-up         # Start all services
make docker-down       # Stop all services
make docker-build      # Build Docker images
make docker-logs       # Show all service logs
make docker-clean      # Clean Docker artifacts
make status            # Show service status
```

### Service-Specific Commands

Work with individual services when needed:

```bash
# User Service
cd user-service
make build             # Build user service only
make test              # Test user service only
make lint-go           # Lint Go code
make lint-proto        # Lint Proto files
make proto             # Generate protobuf

# Email Service
cd email-service
make build             # Build email service only
make test              # Test email service only
make lint-go           # Lint Go code
make lint-proto        # Lint Proto files
make proto             # Generate protobuf
```

### Code Quality & Testing

```bash
make test              # Run all tests
make test-cover        # Run tests with coverage
make lint              # Run all linters (Go + Proto)
make lint-go           # Run Go linter only
make lint-proto        # Run Proto linter only
make fmt               # Format all Go code
make check-all         # Run tests + linters
```

### Development Workflow Commands

```bash
make dev               # Development cycle: down -> build -> up -> logs
make restart-all       # Restart all services
make restart-user      # Restart user service
make restart-email     # Restart email service
make rebuild           # Full rebuild from scratch
```

### Sequential Execution (Debugging)

For debugging or when parallel execution causes issues:

```bash
make build-seq         # Build services sequentially
make test-seq          # Run tests sequentially  
make lint-seq          # Run linters sequentially
make show-jobs         # Show current parallel configuration
```

### Maintenance & Cleanup

```bash
make clean             # Clean build artifacts
make clean-all         # Deep clean (including Docker volumes)
make docker-clean      # Clean Docker artifacts
make reset             # Reset project to initial state
```

## Optimized Project Structure

```
mailflow/
├── .golangci.yaml           # 🆕 Centralized Go linter config
├── .protolint.yaml          # 🆕 Centralized Proto linter config  
├── Makefile                 # 🔄 Optimized project-level commands
├── MAKEFILE_STRUCTURE.md    # 🆕 Documentation for build system
├── user-service/
│   ├── Makefile             # 🔄 Service-specific commands only
│   ├── cmd/server/          # Application entry point
│   ├── internal/
│   │   ├── circuitbreaker/  # Circuit breaker implementation
│   │   ├── config/          # Configuration management
│   │   ├── domain/          # Domain models
│   │   ├── grpc/            # gRPC server implementation
│   │   ├── grpc_gateway/    # HTTP/gRPC gateway
│   │   ├── metrics/         # Prometheus metrics collectors
│   │   ├── queue/           # Message queue implementation
│   │   ├── retry/           # Retry mechanism
│   │   └── services/        # Business logic
│   └── proto/               # Protocol buffer definitions
├── email-service/
│   ├── Makefile             # 🔄 Service-specific commands only
│   ├── cmd/server/          # Application entry point
│   ├── internal/            # Similar structure to user-service
│   └── proto/               # Protocol buffer definitions
├── common/                  # Shared components
│   ├── logger/              # Structured logging (Zap)
│   ├── tracing/             # OpenTelemetry tracing
│   └── metrics/             # Prometheus metrics helpers
├── bin/                     # 🆕 Centralized development tools
├── docker-compose.yaml      # Container orchestration
├── krakend.json            # API Gateway configuration
└── prometheus.yml          # Prometheus configuration
```

### Key Improvements ✨

1. **🎯 DRY Principle**: No duplicate configurations
2. **🏗️ Centralized Management**: All configs in root directory
3. **⚡ Parallel Execution**: Fast builds using all CPU cores
4. **🔧 Easy Maintenance**: Single point of configuration updates
5. **📁 Clear Hierarchy**: Logical separation of general vs specific commands
6. **🎛️ Flexible Control**: Customizable parallel job count
7. **🐛 Debug Support**: Sequential execution fallback for troubleshooting

## Resilience Patterns Implementation

### Circuit Breaker

The circuit breaker has three states:
- **Closed**: Normal operation, requests pass through
- **Open**: Service failures exceeded threshold, requests fail fast
- **Half-Open**: Testing if service recovered, limited requests allowed

Configuration:
- Failure Threshold: 5 failures to open circuit
- Success Threshold: 2 successes to close circuit
- Timeout: 30 seconds before attempting recovery
- Max Requests in Half-Open: 3

### Rate Limiter

Email service implements rate limiting using the `/Users/ppopeskul/dev/ratelimiter` library:
- Algorithm: Token Bucket
- Rate: 60 emails per minute (configurable)
- Burst: 10 emails (configurable)

### Message Queue

Failed email requests are queued for retry:
- In-memory queue with configurable size (default: 1000)
- Max retries per message: 3
- Queue processor runs every 10 seconds

### Retry Mechanism

Exponential backoff with jitter:
- Initial delay: 100ms
- Max delay: 30s
- Multiplier: 2.0
- Max attempts: 5

## Simulating Failures

The email service automatically simulates downtime:
- Frequency: Every 5 minutes (configurable via `DOWNTIME_INTERVAL`)
- Duration: 30 seconds (configurable via `DOWNTIME_DURATION`)

During downtime:
1. Circuit breaker opens after 5 failures
2. New email requests are queued
3. Queue processor waits for circuit to close
4. Queued emails are sent when service recovers

To check circuit breaker status:
```bash
curl http://localhost:9101/circuit-breaker
```

To check queue status:
```bash
curl http://localhost:9101/queue
```

## Metrics & Monitoring

### RED Metrics
- **Rate**: `*_requests_total`
- **Errors**: `*_errors_total`
- **Duration**: `*_request_duration_seconds`

### Circuit Breaker Metrics
- `user_service_circuit_breaker_state`
- `user_service_circuit_breaker_failures_total`
- `user_service_circuit_breaker_successes_total`
- `user_service_circuit_breaker_half_open_requests`

### Queue Metrics
- `user_service_queue_size`
- `user_service_queue_processing`
- `user_service_queue_total`

View metrics at:
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)
- User Service Metrics: http://localhost:9101/metrics
- Email Service Metrics: http://localhost:9102/metrics

## Configuration

### User Service Environment Variables
- `GRPC_PORT`: gRPC server port (default: :50051)
- `HTTP_PORT`: HTTP gateway port (default: :8080)
- `METRICS_PORT`: Metrics endpoint port (default: :9101)
- `EMAIL_SERVICE_ADDRESS`: Email service address
- `CLIENT_EMAIL_SERVICE_TIMEOUT`: Request timeout
- `CLIENT_EMAIL_SERVICE_RETRY_ATTEMPTS`: Max retry attempts
- `CLIENT_EMAIL_SERVICE_RETRY_DELAY`: Initial retry delay

### Email Service Environment Variables
- `GRPC_PORT`: gRPC server port (default: :50052)
- `METRICS_PORT`: Metrics endpoint port (default: :9102)
- `RATE_LIMIT_RPM`: Emails per minute
- `RATE_LIMIT_BURST`: Burst capacity
- `DOWNTIME_INTERVAL`: Downtime frequency
- `DOWNTIME_DURATION`: Downtime duration
- `DOWNTIME_ENABLED`: Enable downtime simulation

## Development Workflow

### Typical Development Cycle

1. Make changes to code
2. Run tests: `make test`
3. Check code quality: `make lint`
4. Format code: `make fmt`
5. Build services: `make build`
6. Test with Docker: `make dev`
7. Check logs: `make docker-logs`
8. Monitor metrics: `make monitor-metrics`

### Advanced Development Patterns

```bash
# Complete development cycle
make dev

# Work with specific service
cd user-service
make build test lint

# Monitor system in real-time
make monitor-metrics

# Reset and rebuild everything
make rebuild

# Check system status
make status
make health-check
```

## Rate Limiter Library

This project uses a custom rate limiter library located at `/Users/ppopeskul/dev/ratelimiter`. The library implements various rate limiting algorithms:

- **Token Bucket**: Used in the email service
- **Sliding Window**: Available for other use cases
- **Fixed Window**: Simple rate limiting
- **Leaky Bucket**: Traffic shaping

See the rate limiter documentation for implementation details and usage examples.

## Best Practices Implemented

1. **🏗️ Clean Architecture**: Separation of concerns with clear layers
2. **🔌 Dependency Injection**: Interfaces for all major components
3. **🚨 Error Handling**: Consistent error propagation and logging
4. **👁️ Observability**: Structured logging, metrics, and tracing
5. **🧪 Testing**: Comprehensive unit tests
6. **⚙️ Configuration**: Environment-based configuration management
7. **🛑 Graceful Shutdown**: Proper cleanup of resources
8. **🔒 Concurrent Safety**: Thread-safe implementations
9. **📋 Build Optimization**: Centralized configurations and efficient workflows
10. **📚 Documentation**: Clear README and inline documentation

## Troubleshooting

### Common Issues

1. **Port already in use**
   ```bash
   make docker-down
   # Wait a few seconds
   make docker-up
   ```

2. **Proto generation fails**
   ```bash
   make tools
   make proto
   ```

3. **Linting errors**
   ```bash
   make fmt
   make lint
   ```

4. **Services not responding**
   ```bash
   make health-check
   make docker-logs
   ```

5. **Build tool issues**
   ```bash
   make clean-tools
   make tools
   ```

6. **Docker issues**
   ```bash
   make docker-clean
   make docker-build
   make docker-up
   ```

### Debug Commands

```bash
make status              # Check service status
make health-check        # Verify service health
make versions            # Show tool versions
make docker-ps           # Show running containers
```

## Contributing

1. Follow the established project structure
2. Use the provided Make commands for consistency
3. Ensure all tests pass: `make check-all`
4. Format code before committing: `make fmt`
5. Update documentation as needed

## License

MIT License
