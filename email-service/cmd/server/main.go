package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"

	"github.com/popeskul/mailflow/common/logger"
	"github.com/popeskul/mailflow/common/tracing"
	"github.com/popeskul/mailflow/email-service/internal/config"
	grpc2 "github.com/popeskul/mailflow/email-service/internal/grpc"
	"github.com/popeskul/mailflow/email-service/internal/metrics"
	"github.com/popeskul/mailflow/email-service/internal/repositories/memory"
	"github.com/popeskul/mailflow/email-service/internal/services"
	"github.com/popeskul/mailflow/email-service/internal/smtp"
	pb "github.com/popeskul/mailflow/email-service/pkg/api/email/v1"
	"github.com/popeskul/mailflow/ratelimiter"
)

func main() {
	ctx := context.Background()

	// Initialize configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	l := logger.NewZapLogger()

	// Create rate limiter
	limiter := ratelimiter.NewTokenBucket(
		cfg.RateLimit.EmailsPerMinute,
		time.Minute,
	)

	// Initialize metrics
	emailMetrics := metrics.NewEmailMetrics()

	// Initialize email sender
	emailSender := smtp.NewSMTPSender(cfg.SMTP, l)

	// Initialize repositories
	repos := memory.NewRepositories(l)

	// Initialize services
	services := services.NewServices(repos, emailSender, limiter, emailMetrics, l)
	emailServer := grpc2.NewEmailServer(services.Email(), emailMetrics, l)

	tracingConfig := tracing.Config{
		ServiceName:  cfg.Trace.ServiceName,
		OTLPEndpoint: cfg.Trace.JaegerURL,
		Version:      cfg.Trace.Version,
		Enabled:      true,
	}

	tp, err := tracing.InitTracer(tracingConfig)
	if err != nil {
		l.Fatal("failed to init tracer",
			logger.Field{Key: "error", Value: err},
		)
	}
	l.Info("tracer initialized successfully",
		logger.Field{Key: "service_name", Value: tracingConfig.ServiceName},
		logger.Field{Key: "otlp_endpoint", Value: tracingConfig.OTLPEndpoint},
	)

	defer func() {
		if shutdownErr := tp.Shutdown(ctx); shutdownErr != nil {
			l.Error("failed to shutdown tracer",
				logger.Field{Key: "error", Value: shutdownErr},
			)
		}
	}()

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterEmailServiceServer(grpcServer, emailServer)

	// Start gRPC server
	grpcLis, err := net.Listen("tcp", cfg.Server.GRPCPort)
	if err != nil {
		l.Fatal("failed to listen on gRPC port",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "port", Value: cfg.Server.GRPCPort},
		)
	}

	go func() {
		l.Info("starting gRPC server",
			logger.Field{Key: "port", Value: cfg.Server.GRPCPort},
		)
		if grpcErr := grpcServer.Serve(grpcLis); grpcErr != nil {
			l.Fatal("failed to serve gRPC",
				logger.Field{Key: "error", Value: grpcErr},
			)
		}
	}()

	// Start metrics server
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())

	metricsServer := &http.Server{
		Addr:    cfg.Monitor.MetricsPort,
		Handler: metricsMux,
	}

	go func() {
		l.Info("starting metrics server",
			logger.Field{Key: "port", Value: cfg.Monitor.MetricsPort},
		)
		if metricsErr := metricsServer.ListenAndServe(); metricsErr != nil && metricsErr != http.ErrServerClosed {
			l.Fatal("failed to serve metrics",
				logger.Field{Key: "error", Value: metricsErr},
			)
		}
	}()

	// Start periodic downtime simulation
	if cfg.Downtime.Enabled {
		go func() {
			for {
				time.Sleep(time.Duration(cfg.Downtime.IntervalMinutes) * time.Minute)
				l.Info("simulating downtime",
					logger.Field{Key: "duration_minutes", Value: cfg.Downtime.DurationMinutes},
				)
				time.Sleep(time.Duration(cfg.Downtime.DurationMinutes) * time.Minute)
				l.Info("downtime simulation ended")
			}
		}()
	}

	// Wait for interrupt signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	l.Info("shutting down servers")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if metricsErr := metricsServer.Shutdown(shutdownCtx); metricsErr != nil {
		l.Error("metrics server shutdown error",
			logger.Field{Key: "error", Value: metricsErr},
		)
	}

	grpcServer.GracefulStop()
	l.Info("all servers stopped")
}
