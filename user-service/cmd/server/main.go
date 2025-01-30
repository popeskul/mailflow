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

	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	emailv1 "github.com/popeskul/email-service-platform/email-service/pkg/api/email/v1"
	"github.com/popeskul/email-service-platform/logger"
	"github.com/popeskul/email-service-platform/user-service/internal/config"
	grpc2 "github.com/popeskul/email-service-platform/user-service/internal/grpc"
	"github.com/popeskul/email-service-platform/user-service/internal/grpc_gateway"
	"github.com/popeskul/email-service-platform/user-service/internal/repositories/memory"
	"github.com/popeskul/email-service-platform/user-service/internal/services"
	pbv1 "github.com/popeskul/email-service-platform/user-service/pkg/api/user/v1"
)

func main() {
	l := logger.NewZapLogger(
		logger.WithLogLevel(logger.InfoLevel),
		logger.WithJSONFormat(),
	)
	defer l.Sync()

	cfg, err := config.LoadConfig()
	if err != nil {
		l.Fatal("failed to load config",
			logger.Field{Key: "error", Value: err},
		)
	}

	emailConn, err := grpc.NewClient(
		cfg.Email.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		l.Fatal("failed to connect to email service",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "address", Value: cfg.Email.Address},
		)
	}
	defer func() {
		if err := emailConn.Close(); err != nil {
			l.Error("failed to close email connection",
				logger.Field{Key: "error", Value: err},
			)
		}
	}()

	emailClient := emailv1.NewEmailServiceClient(emailConn)

	repos := memory.NewRepositories()
	services := services.NewServices(repos, emailClient, l)
	userServer := grpc2.NewUserServer(services, l)

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			grpc_recovery.UnaryServerInterceptor(),
			otelgrpc.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			grpc2.LoggingInterceptor(l),
		),
	}
	server := grpc.NewServer(opts...)
	grpc_prometheus.Register(server)
	pbv1.RegisterUserServiceServer(server, userServer)

	lis, err := net.Listen("tcp", cfg.GRPC.Port)
	if err != nil {
		l.Fatal("failed to start tcp listener",
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := grpc_gateway.NewGatewayMux()
	err = pbv1.RegisterUserServiceHandlerServer(ctx, mux, userServer)
	if err != nil {
		l.Fatal("failed to register gateway",
			logger.Field{Key: "error", Value: err},
		)
	}

	// Setup HTTP server (gRPC-Gateway)
	httpServer := &http.Server{
		Addr:    cfg.HTTP.Port,
		Handler: mux,
	}

	go func() {
		l.Info("starting http server",
			logger.Field{Key: "port", Value: cfg.HTTP.Port},
		)
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			l.Fatal("failed to serve http",
				logger.Field{Key: "error", Value: err},
				logger.Field{Key: "port", Value: cfg.HTTP.Port},
			)
		}
	}()

	// Setup server metrics
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

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	l.Info("initiating graceful shutdown")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	server.GracefulStop()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		l.Error("failed to shutdown http server",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "port", Value: cfg.HTTP.Port},
		)
	}

	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		l.Error("failed to shutdown metrics server",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "port", Value: cfg.Metrics.Port},
		)
	}

	l.Info("service stopped")
}
