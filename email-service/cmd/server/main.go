package main

import (
	"context"
	"errors"
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
	"github.com/popeskul/ratelimiter"
)

func main() {
	initialLogger := logger.NewZapLogger(
		logger.WithLogLevel(logger.InfoLevel),
		logger.WithJSONFormat(),
	).Named("email_service")

	cfg, err := config.LoadConfig()
	if err != nil {
		initialLogger.Fatal("failed to load config",
			logger.Field{Key: "error", Value: err},
		)
	}

	logConfig := cfg.Log.ToConfig()
	l := logger.NewZapLogger(
		logger.WithLogLevel(logConfig.Level),
		logger.WithJSONFormat(),
		logger.WithOutputPath(logConfig.OutputPath),
	).Named("email_service")
	defer func() {
		_ = l.Sync() // Unable to sync logger at shutdown
	}()

	emailMetrics := metrics.NewEmailMetrics("email_service")

	limiter, err := ratelimiter.New(
		ratelimiter.WithRate(cfg.Email.RateLimit.EmailsPerMinute),
		ratelimiter.WithBurst(cfg.Email.RateLimit.MaxBurst),
		ratelimiter.WithAlgorithm(ratelimiter.TokenBucketAlgorithm),
	)
	if err != nil {
		l.Fatal("failed to create rate limiter",
			logger.Field{Key: "error", Value: err},
		)
	}

	repos := memory.NewRepositories(l)
	emailSender := smtp.NewSMTPSender(cfg.Email.SMTP, l)

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
		if err := tp.Shutdown(context.Background()); err != nil {
			l.Error("failed to shutdown tracer",
				logger.Field{Key: "error", Value: err},
			)
		}
	}()

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			grpc2.RecoveryInterceptor(l),
			// TODO: Replace with NewServerHandler when available
			// otelgrpc.UnaryServerInterceptor(),
			grpc2.LoggingInterceptor(l),
			grpc2.MetricsInterceptor(emailMetrics),
		),
	}
	server := grpc.NewServer(opts...)
	pb.RegisterEmailServiceServer(server, emailServer)

	lis, err := net.Listen("tcp", cfg.Server.GRPCPort)
	if err != nil {
		l.Fatal("failed to listen",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "port", Value: cfg.Server.GRPCPort},
		)
	}

	go func() {
		l.Info("starting grpc server",
			logger.Field{Key: "port", Value: cfg.Server.GRPCPort},
		)
		if err := server.Serve(lis); err != nil {
			l.Fatal("failed to serve grpc",
				logger.Field{Key: "error", Value: err},
				logger.Field{Key: "port", Value: cfg.Server.GRPCPort},
			)
		}
	}()

	metricsServer := &http.Server{
		Addr:    cfg.Monitor.MetricsPort,
		Handler: promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{}),
	}

	go func() {
		l.Info("starting metrics server",
			logger.Field{Key: "port", Value: cfg.Monitor.MetricsPort},
		)
		if err := metricsServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			l.Fatal("failed to serve metrics",
				logger.Field{Key: "error", Value: err},
				logger.Field{Key: "port", Value: cfg.Monitor.MetricsPort},
			)
		}
	}()

	// Run shutdown simulation if enabled
	if cfg.Email.Maintenance.Enabled {
		go simulateDowntime(
			emailServer,
			cfg.Email.Maintenance,
			l,
		)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	l.Info("initiating graceful shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	server.GracefulStop()
	if err := metricsServer.Shutdown(ctx); err != nil {
		l.Error("failed to shutdown metrics server",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "port", Value: cfg.Monitor.MetricsPort},
		)
	}

	l.Info("service stopped")
}

func simulateDowntime(
	server *grpc2.EmailServer,
	maintenance config.MaintenanceConfig,
	l logger.Logger,
) {
	ticker := time.NewTicker(maintenance.Frequency)
	defer ticker.Stop()

	for range ticker.C {
		l.Info("service is going down for maintenance",
			logger.Field{Key: "duration", Value: maintenance.DowntimePeriod},
		)
		server.SetDowntime(true)
		time.Sleep(maintenance.DowntimePeriod)
		server.SetDowntime(false)
		l.Info("service is back up")
	}
}
