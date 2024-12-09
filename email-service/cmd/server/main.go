package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	grpcServer "github.com/popeskul/email-service-platform/email-service/internal/adapters/grpc"
	"github.com/popeskul/email-service-platform/email-service/internal/adapters/repositories/memory"
	"github.com/popeskul/email-service-platform/email-service/internal/adapters/smtp"
	"github.com/popeskul/email-service-platform/email-service/internal/config"
	"github.com/popeskul/email-service-platform/email-service/internal/logger"
	"github.com/popeskul/email-service-platform/email-service/internal/metrics"
	"github.com/popeskul/email-service-platform/email-service/internal/services"
	"github.com/popeskul/email-service-platform/email-service/internal/tracing"
	pb "github.com/popeskul/email-service-platform/email-service/pkg/api/email/v1"
	"github.com/popeskul/ratelimiter"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger, err := logger.NewLogger(cfg.Logger, "email_service")
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	emailMetrics := metrics.NewEmailMetrics("email_service")

	limiter, err := ratelimiter.New(
		ratelimiter.WithRate(cfg.RateLimit.RequestsPerMinute),
		ratelimiter.WithBurst(cfg.RateLimit.BurstSize),
		ratelimiter.WithAlgorithm(ratelimiter.TokenBucketAlgorithm),
	)
	if err != nil {
		logger.Fatal("failed to create rate limiter", zap.Error(err))
	}

	repos := memory.NewRepositories()
	emailSender := smtp.NewSMTPSender(
		cfg.SMTP.Enabled,
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.Username,
		cfg.SMTP.Password,
		cfg.SMTP.From,
		logger,
	)

	services := services.NewServices(repos, emailSender, limiter, emailMetrics, logger)
	emailServer := grpcServer.NewEmailServer(services.EmailService, emailMetrics, logger)

	tp, err := tracing.InitTracer(cfg.Tracing)
	if err != nil {
		logger.Fatal("failed to init tracer", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Error("failed to shutdown tracer", zap.Error(err))
		}
	}()

	tracer := tp.Tracer("server")

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			grpcServer.RecoveryInterceptor(logger),
			grpcServer.TracingInterceptor(tracer),
			grpcServer.LoggingInterceptor(logger),
			grpcServer.MetricsInterceptor(emailMetrics),
		),
	}
	server := grpc.NewServer(opts...)
	pb.RegisterEmailServiceServer(server, emailServer)

	lis, err := net.Listen("tcp", cfg.GRPC.Port)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	go func() {
		logger.Info("starting grpc server", zap.String("port", cfg.GRPC.Port))
		if err := server.Serve(lis); err != nil {
			logger.Fatal("failed to serve grpc", zap.Error(err))
		}
	}()

	metricsServer := &http.Server{
		Addr:    cfg.Metrics.Port,
		Handler: promhttp.Handler(),
	}

	go func() {
		logger.Info("starting metrics server", zap.String("port", cfg.Metrics.Port))
		if err := metricsServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("failed to serve metrics", zap.Error(err))
		}
	}()

	// Run shutdown simulation if enabled
	if cfg.Downtime.Enabled {
		go simulateDowntime(emailServer, cfg.Downtime.Interval, cfg.Downtime.Duration, logger)
	}

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("initiating graceful shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	server.GracefulStop()
	if err := metricsServer.Shutdown(ctx); err != nil {
		logger.Error("failed to shutdown metrics server", zap.Error(err))
	}

	logger.Info("service stopped")
}

func simulateDowntime(
	server *grpcServer.EmailServer,
	interval, duration time.Duration,
	logger *zap.Logger,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		logger.Info("service is going down for maintenance",
			zap.Duration("duration", duration),
		)
		server.SetDowntime(true)
		time.Sleep(duration)
		server.SetDowntime(false)
		logger.Info("service is back up")
	}
}
