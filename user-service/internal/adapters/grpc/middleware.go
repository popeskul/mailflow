package grpc

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// LoggingInterceptor логирует все запросы с trace ID
func LoggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()

		// Получаем trace ID из контекста
		spanCtx := trace.SpanContextFromContext(ctx)
		traceID := "no-trace-id"
		if spanCtx.IsValid() {
			traceID = spanCtx.TraceID().String()
		}

		// Логируем с trace ID
		logger := logger.With(
			zap.String("trace_id", traceID),
			zap.String("method", info.FullMethod),
		)

		logger.Info("processing request")

		resp, err := handler(ctx, req)

		duration := time.Since(startTime)
		if err != nil {
			logger.Error("request failed",
				zap.Duration("duration", duration),
				zap.Error(err),
			)
		} else {
			logger.Info("request successful",
				zap.Duration("duration", duration),
			)
		}

		return resp, err
	}
}
