package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/popeskul/mailflow/common/logger"
)

type Config struct {
	Server  ServerConfig           `mapstructure:"server"`
	Client  ClientConfig           `mapstructure:"client"`
	Monitor MonitorConfig          `mapstructure:"monitor"`
	Trace   TraceConfig            `mapstructure:"trace"`
	Log     logger.UnmarshalConfig `mapstructure:"logger"`
}

type ServerConfig struct {
	GRPCPort        string        `mapstructure:"grpc_port"`
	HTTPPort        string        `mapstructure:"http_port"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type ClientConfig struct {
	EmailService EmailServiceConfig `mapstructure:"email_service"`
}

type EmailServiceConfig struct {
	Address       string        `mapstructure:"address"`
	Timeout       time.Duration `mapstructure:"timeout"`
	RetryAttempts int           `mapstructure:"retry_attempts"`
	RetryDelay    time.Duration `mapstructure:"retry_delay"`
}

type MonitorConfig struct {
	MetricsPort string `mapstructure:"metrics_port"`
}

type TraceConfig struct {
	ServiceName string `mapstructure:"service_name"`
	JaegerURL   string `mapstructure:"jaeger_url"` // Keep for backwards compatibility with config
	Version     string `mapstructure:"version"`
}

func LoadConfig() (*Config, error) {
	setDefaultConfig()

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	return &config, validateConfig(&config)
}

func setDefaultConfig() {
	// Server defaults
	viper.SetDefault("server.grpc_port", ":50051")
	viper.SetDefault("server.http_port", ":8080")
	viper.SetDefault("server.shutdown_timeout", "30s")

	// Client defaults
	viper.SetDefault("client.email_service.address", "email-service:50052")
	viper.SetDefault("client.email_service.timeout", "5s")
	viper.SetDefault("client.email_service.retry_attempts", 3)
	viper.SetDefault("client.email_service.retry_delay", "1s")

	// Monitor defaults
	viper.SetDefault("monitor.metrics_port", ":9101")

	// Logger defaults
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.encoding", "json")
	viper.SetDefault("logger.output_path", "stdout")

	// Trace defaults
	viper.SetDefault("trace.service_name", "user-service")
	viper.SetDefault("trace.version", "1.0.0")
	viper.SetDefault("trace.jaeger_url", "http://jaeger:14268/api/traces")

	viper.AutomaticEnv()
}

func validateConfig(config *Config) error {
	var errors []string

	// Validate Server config
	if config.Server.GRPCPort == "" {
		errors = append(errors, "server.grpc_port is required")
	}
	if config.Server.HTTPPort == "" {
		errors = append(errors, "server.http_port is required")
	}

	// Validate Client config
	if config.Client.EmailService.Address == "" {
		errors = append(errors, "client.email_service.address is required")
	}
	if config.Client.EmailService.Timeout <= 0 {
		errors = append(errors, "client.email_service.timeout must be greater than 0")
	}
	if config.Client.EmailService.RetryAttempts <= 0 {
		errors = append(errors, "client.email_service.retry_attempts must be greater than 0")
	}
	if config.Client.EmailService.RetryDelay <= 0 {
		errors = append(errors, "client.email_service.retry_delay must be greater than 0")
	}

	// Validate Monitor config
	if config.Monitor.MetricsPort == "" {
		errors = append(errors, "monitor.metrics_port is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf("config validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}
