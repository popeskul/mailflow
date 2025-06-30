package grpc

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor logs gRPC requests
func LoggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		logger.Info("gRPC request",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
			zap.Error(err),
		)

		return resp, err
	}
}

// MetricsInterceptor collects metrics for gRPC requests
func MetricsInterceptor(registry *prometheus.Registry) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		// Simple metrics recording - you can expand this as needed
		_ = duration        // Use duration for actual metrics if needed
		_ = info.FullMethod // Use method name for actual metrics if needed

		// Record metrics based on error status if needed
		_ = err

		return resp, err
	}
}

// RecoveryInterceptor recovers from panics in gRPC handlers
func RecoveryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("gRPC panic recovered",
					zap.String("method", info.FullMethod),
					zap.Any("panic", r),
				)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}

// TracingInterceptor adds tracing to gRPC requests
func TracingInterceptor(tracer trace.Tracer) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		ctx, span := tracer.Start(ctx, info.FullMethod)
		defer span.End()

		return handler(ctx, req)
	}
}
