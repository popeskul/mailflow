# Common Module

This module contains shared components used across microservices in the Email Service Platform.

## Components

### Logger
A structured logging package based on Uber's Zap logger.

```go
import "github.com/popeskul/mailflow/common/logger"

// Create logger
l := logger.NewZapLogger(
    logger.WithLogLevel(logger.InfoLevel),
    logger.WithJSONFormat(),
)

// Use logger
l.Info("message", 
    logger.Field{Key: "user_id", Value: "123"},
    logger.Field{Key: "action", Value: "created"},
)
```

### Tracing
OpenTelemetry tracing with Jaeger exporter.

```go
import "github.com/popeskul/mailflow/common/tracing"

// Initialize tracer
config := tracing.Config{
    ServiceName: "my-service",
    JaegerURL:   "http://jaeger:14268/api/traces",
    Version:     "1.0.0",
    Enabled:     true,
}

tp, err := tracing.InitTracer(config)
defer tp.Shutdown(context.Background())

// Use tracer
tracer := tp.Tracer("my-component")
ctx, span := tracer.Start(ctx, "operation-name")
defer span.End()
```

### Metrics
Prometheus metrics helpers with RED pattern support.

```go
import "github.com/popeskul/mailflow/common/metrics"

// Create RED metrics
redMetrics := metrics.NewREDMetrics("my_service", "api")

// Use timer
timer := redMetrics.StartTimer("create_user")
// ... do work ...
timer.ObserveDuration(err)

// Or record manually
redMetrics.RecordRequest("create_user", duration, err)

// For gRPC services
grpcMetrics := metrics.NewGRPCMetrics("my_service", "grpc")
server := grpc.NewServer(
    grpc.UnaryInterceptor(grpcMetrics.UnaryServerInterceptor()),
)
```

## Usage in Services

1. Add to go.work:
```
use (
    ./common
    ./my-service
)
```

2. Import in your service:
```go
import (
    "github.com/popeskul/mailflow/common/logger"
    "github.com/popeskul/mailflow/common/tracing"
    "github.com/popeskul/mailflow/common/metrics"
)
```

## Testing

Run tests:
```bash
cd common
go test ./...
```

## Structure

```
common/
├── logger/          # Structured logging
│   ├── interfaces.go
│   ├── options.go
│   └── zap_impl.go
├── tracing/         # OpenTelemetry tracing
│   └── tracer.go
├── metrics/         # Prometheus metrics
│   ├── red.go      # RED pattern implementation
│   └── grpc.go     # gRPC interceptors
└── go.mod
```
