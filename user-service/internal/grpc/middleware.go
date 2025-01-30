package grpc

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"

	"github.com/popeskul/email-service-platform/logger"
)

// LoggingInterceptor logs all requests with a trace ID
func LoggingInterceptor(l logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		l = l.Named("grpc")
		startTime := time.Now()

		// Get trace ID from context
		spanCtx := trace.SpanContextFromContext(ctx)
		traceID := "no-trace-id"
		if spanCtx.IsValid() {
			traceID = spanCtx.TraceID().String()
		}

		// Logging with trace ID
		l = l.WithFields(logger.Fields{
			"trace_id": traceID,
			"method":   info.FullMethod,
		})

		l.Info("processing request")

		resp, err := handler(ctx, req)

		duration := time.Since(startTime)
		if err != nil {
			l.Error("request failed",
				logger.Field{Key: "duration", Value: duration},
				logger.Field{Key: "error", Value: err},
			)
		} else {
			l.Info("request successful",
				logger.Field{Key: "duration", Value: duration},
			)
		}

		return resp, err
	}
}
