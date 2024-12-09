package grpc

import (
	"context"
	"runtime/debug"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/popeskul/email-service-platform/user-service/internal/metrics"
)

// TracingInterceptor добавляет трейсинг к запросам
func TracingInterceptor(tracer trace.Tracer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		spanCtx, span := tracer.Start(ctx, info.FullMethod)
		defer span.End()

		// Добавляем атрибуты к спану
		span.SetAttributes(
			attribute.String("rpc.method", info.FullMethod),
			attribute.String("rpc.system", "grpc"),
		)

		// Выполняем запрос
		resp, err := handler(spanCtx, req)

		if err != nil {
			span.RecordError(err)
			span.SetAttributes(attribute.String("error", err.Error()))
		}

		return resp, err
	}
}

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

// MetricsInterceptor собирает метрики для каждого запроса
func MetricsInterceptor(metrics *metrics.REDMetrics) grpc.UnaryServerInterceptor {
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

// RecoveryInterceptor восстанавливается после паники
func RecoveryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				// Получаем trace ID для логирования
				spanCtx := trace.SpanContextFromContext(ctx)
				traceID := "no-trace-id"
				if spanCtx.IsValid() {
					traceID = spanCtx.TraceID().String()
				}

				logger.Error("recovered from panic",
					zap.String("trace_id", traceID),
					zap.Any("panic", r),
					zap.String("method", info.FullMethod),
					zap.String("stack", string(debug.Stack())),
				)
				err = status.Error(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}
