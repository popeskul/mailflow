package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/popeskul/mailflow/common/logger"
	"github.com/popeskul/mailflow/user-service/internal/config"
	grpcserver "github.com/popeskul/mailflow/user-service/internal/grpc"
	"github.com/popeskul/mailflow/user-service/internal/repositories/memory"
	"github.com/popeskul/mailflow/user-service/internal/services"
	pb "github.com/popeskul/mailflow/user-service/pkg/api/user/v1"
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

	// Initialize repositories
	repos := memory.NewRepositories(l)

	// Initialize services (without email client for now)
	srvs := services.NewServices(repos, nil, l)

	// Start gRPC server
	grpcServer := grpc.NewServer()

	userGrpcServer := grpcserver.NewUserServer(srvs, l)
	pb.RegisterUserServiceServer(grpcServer, userGrpcServer)

	grpcLis, err := net.Listen("tcp", cfg.Server.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port: %v", err)
	}

	go func() {
		log.Printf("Starting gRPC server on %s", cfg.Server.GRPCPort)
		if grpcErr := grpcServer.Serve(grpcLis); grpcErr != nil {
			log.Fatalf("Failed to serve gRPC: %v", grpcErr)
		}
	}()

	// Start HTTP server (gRPC-Gateway)
	mux := runtime.NewServeMux()

	conn, err := grpc.NewClient(
		"localhost"+cfg.Server.GRPCPort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			log.Printf("Failed to close gRPC connection: %v", closeErr)
		}
	}()

	if err := pb.RegisterUserServiceHandler(ctx, mux, conn); err != nil {
		log.Fatalf("Failed to register gRPC-Gateway handler: %v", err)
	}

	httpServer := &http.Server{
		Addr:    cfg.Server.HTTPPort,
		Handler: mux,
	}

	go func() {
		log.Printf("Starting HTTP server on %s", cfg.Server.HTTPPort)
		if httpErr := httpServer.ListenAndServe(); httpErr != nil && httpErr != http.ErrServerClosed {
			log.Fatalf("Failed to serve HTTP: %v", httpErr)
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
		log.Printf("Starting metrics server on %s", cfg.Monitor.MetricsPort)
		if metricsErr := metricsServer.ListenAndServe(); metricsErr != nil && metricsErr != http.ErrServerClosed {
			log.Fatalf("Failed to serve metrics: %v", metricsErr)
		}
	}()

	// Wait for interrupt signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("Shutting down servers...")

	// Shutdown servers gracefully
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Metrics server shutdown error: %v", err)
	}

	grpcServer.GracefulStop()

	log.Println("All servers stopped")
}
