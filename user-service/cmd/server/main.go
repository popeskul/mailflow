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

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	emailv1 "github.com/popeskul/email-service-platform/email-service/pkg/api/email/v1"
	grpcServer "github.com/popeskul/email-service-platform/user-service/internal/adapters/grpc"
	"github.com/popeskul/email-service-platform/user-service/internal/adapters/repositories/memory"
	"github.com/popeskul/email-service-platform/user-service/internal/config"
	"github.com/popeskul/email-service-platform/user-service/internal/logger"
	"github.com/popeskul/email-service-platform/user-service/internal/metrics"
	"github.com/popeskul/email-service-platform/user-service/internal/services"
	"github.com/popeskul/email-service-platform/user-service/internal/tracing"
	pbv1 "github.com/popeskul/email-service-platform/user-service/pkg/api/user/v1"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger, err := logger.NewLogger(cfg.Logger, "user_service")
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	userMetrics := metrics.NewREDMetrics("user_service")

	emailConn, err := grpc.NewClient(
		cfg.Email.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Fatal("failed to connect to email service", zap.Error(err))
	}
	defer func() {
		if err := emailConn.Close(); err != nil {
			logger.Error("failed to close email connection", zap.Error(err))
		}
	}()

	emailClient := emailv1.NewEmailServiceClient(emailConn)

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

	repos := memory.NewRepositories()
	services := services.NewServices(repos, emailClient, logger)
	userServer := grpcServer.NewUserServer(services, logger)

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			grpcServer.RecoveryInterceptor(logger),
			grpcServer.TracingInterceptor(tracer),
			grpcServer.LoggingInterceptor(logger),
			grpcServer.MetricsInterceptor(userMetrics),
		),
	}
	server := grpc.NewServer(opts...)
	pbv1.RegisterUserServiceServer(server, userServer)

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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := runtime.NewServeMux()
	err = pbv1.RegisterUserServiceHandlerServer(ctx, mux, userServer)
	if err != nil {
		logger.Fatal("failed to register gateway", zap.Error(err))
	}

	// Setup HTTP server (gRPC-Gateway)
	httpServer := &http.Server{
		Addr:    cfg.HTTP.Port,
		Handler: mux,
	}

	go func() {
		logger.Info("starting http server", zap.String("port", cfg.HTTP.Port))
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("failed to serve http", zap.Error(err))
		}
	}()

	// Setup server metrics
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

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("initiating graceful shutdown")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	server.GracefulStop()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("failed to shutdown http server", zap.Error(err))
	}

	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("failed to shutdown metrics server", zap.Error(err))
	}

	logger.Info("service stopped")
}
