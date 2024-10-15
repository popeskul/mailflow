package grpc

import (
	"context"
	"runtime/debug"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/popeskul/mailflow/common/logger"
	"github.com/popeskul/mailflow/email-service/internal/metrics"
)

const (
	noTraceID = "no-trace-id"
)

func TracingInterceptor(tracer trace.Tracer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		spanCtx, span := tracer.Start(ctx, info.FullMethod)
		defer span.End()

		span.SetAttributes(
			attribute.String("rpc.method", info.FullMethod),
			attribute.String("rpc.system", "grpc"),
		)

		resp, err := handler(spanCtx, req)

		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("error", err.Error()))
		}

		return resp, err
	}
}

func LoggingInterceptor(l logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()

		spanCtx := trace.SpanContextFromContext(ctx)
		traceID := noTraceID
		if spanCtx.IsValid() {
			traceID = spanCtx.TraceID().String()
		}

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

func MetricsInterceptor(metrics *metrics.EmailMetrics) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()

		resp, err := handler(ctx, req)

		metrics.RecordRequest(
			info.FullMethod,
			time.Since(startTime).Seconds(),
			err,
		)

		return resp, err
	}
}

func RecoveryInterceptor(l logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				spanCtx := trace.SpanContextFromContext(ctx)
				traceID := noTraceID
				if spanCtx.IsValid() {
					traceID = spanCtx.TraceID().String()
				}

				l.Error("recovered from panic",
					logger.Field{Key: "trace_id", Value: traceID},
					logger.Field{Key: "panic", Value: r},
					logger.Field{Key: "method", Value: info.FullMethod},
					logger.Field{Key: "stack", Value: string(debug.Stack())},
				)
				err = status.Error(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}
