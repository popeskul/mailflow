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

	"github.com/popeskul/email-service-platform/email-service/internal/config"
	grpc2 "github.com/popeskul/email-service-platform/email-service/internal/grpc"
	"github.com/popeskul/email-service-platform/email-service/internal/metrics"
	"github.com/popeskul/email-service-platform/email-service/internal/repositories/memory"
	"github.com/popeskul/email-service-platform/email-service/internal/services"
	"github.com/popeskul/email-service-platform/email-service/internal/smtp"
	"github.com/popeskul/email-service-platform/email-service/internal/tracing"
	pb "github.com/popeskul/email-service-platform/email-service/pkg/api/email/v1"
	"github.com/popeskul/email-service-platform/logger"
	"github.com/popeskul/ratelimiter"
)

func main() {
	l := logger.NewZapLogger(
		logger.WithLogLevel(logger.InfoLevel),
		logger.WithJSONFormat(),
	).Named("email_service")
	defer l.Sync()

	cfg, err := config.LoadConfig()
	if err != nil {
		l.Fatal("failed to load config",
			logger.Field{Key: "error", Value: err},
		)
	}

	emailMetrics := metrics.NewEmailMetrics("email_service")

	limiter, err := ratelimiter.New(
		ratelimiter.WithRate(cfg.RateLimit.RequestsPerMinute),
		ratelimiter.WithBurst(cfg.RateLimit.BurstSize),
		ratelimiter.WithAlgorithm(ratelimiter.TokenBucketAlgorithm),
	)
	if err != nil {
		l.Fatal("failed to create rate limiter",
			logger.Field{Key: "error", Value: err},
		)
	}

	repos := memory.NewRepositories(l)
	emailSender := smtp.NewSMTPSender(
		cfg.SMTP.Enabled,
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.Username,
		cfg.SMTP.Password,
		cfg.SMTP.From,
		l,
	)

	services := services.NewServices(repos, emailSender, limiter, emailMetrics, l)
	emailServer := grpc2.NewEmailServer(services.Email(), emailMetrics, l)

	tp, err := tracing.InitTracer(cfg.Tracing)
	if err != nil {
		l.Fatal("failed to init tracer",
			logger.Field{Key: "error", Value: err},
		)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			l.Error("failed to shutdown tracer",
				logger.Field{Key: "error", Value: err},
			)
		}
	}()

	tracer := tp.Tracer("server")

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			grpc2.RecoveryInterceptor(l),
			grpc2.TracingInterceptor(tracer),
			grpc2.LoggingInterceptor(l),
			grpc2.MetricsInterceptor(emailMetrics),
		),
	}
	server := grpc.NewServer(opts...)
	pb.RegisterEmailServiceServer(server, emailServer)

	lis, err := net.Listen("tcp", cfg.GRPC.Port)
	if err != nil {
		l.Fatal("failed to listen",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "port", Value: cfg.GRPC.Port},
		)
	}

	go func() {
		l.Info("starting grpc server",
			logger.Field{Key: "port", Value: cfg.GRPC.Port},
		)
		if err := server.Serve(lis); err != nil {
			l.Fatal("failed to serve grpc",
				logger.Field{Key: "error", Value: err},
				logger.Field{Key: "port", Value: cfg.GRPC.Port},
			)
		}
	}()

	metricsServer := &http.Server{
		Addr:    cfg.Metrics.Port,
		Handler: promhttp.Handler(),
	}

	go func() {
		l.Info("starting metrics server",
			logger.Field{Key: "port", Value: cfg.Metrics.Port},
		)
		if err := metricsServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			l.Fatal("failed to serve metrics",
				logger.Field{Key: "error", Value: err},
				logger.Field{Key: "port", Value: cfg.Metrics.Port},
			)
		}
	}()

	// Run shutdown simulation if enabled
	if cfg.Downtime.Enabled {
		go simulateDowntime(emailServer, cfg.Downtime.Interval, cfg.Downtime.Duration, l)
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
			logger.Field{Key: "port", Value: cfg.Metrics.Port},
		)
	}

	l.Info("service stopped")
}

func simulateDowntime(
	server *grpc2.EmailServer,
	interval, duration time.Duration,
	l logger.Logger,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		l.Info("service is going down for maintenance",
			logger.Field{Key: "duration", Value: duration},
		)
		server.SetDowntime(true)
		time.Sleep(duration)
		server.SetDowntime(false)
		l.Info("service is back up")
	}
}
