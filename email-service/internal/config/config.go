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
	Email   EmailConfig            `mapstructure:"email"`
	Monitor MonitorConfig          `mapstructure:"monitor"`
	Trace   TraceConfig            `mapstructure:"trace"`
	Log     logger.UnmarshalConfig `mapstructure:"logger"`
}

type ServerConfig struct {
	GRPCPort        string        `mapstructure:"grpc_port"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type EmailConfig struct {
	SMTP        SMTPConfig        `mapstructure:"smtp"`
	RateLimit   RateLimitConfig   `mapstructure:"rate_limit"`
	Maintenance MaintenanceConfig `mapstructure:"maintenance"`
}

type SMTPConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Host        string `mapstructure:"host"`
	Port        string `mapstructure:"port"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	SenderEmail string `mapstructure:"sender_email"`
}

type RateLimitConfig struct {
	EmailsPerMinute int `mapstructure:"emails_per_minute"`
	MaxBurst        int `mapstructure:"max_burst"`
}

type MaintenanceConfig struct {
	Enabled        bool          `mapstructure:"enabled"`
	Frequency      time.Duration `mapstructure:"frequency"`
	DowntimePeriod time.Duration `mapstructure:"downtime_period"`
}

type MonitorConfig struct {
	MetricsPort string `mapstructure:"metrics_port"`
}

type TraceConfig struct {
	ServiceName string `mapstructure:"service_name"`
	JaegerURL   string `mapstructure:"jaeger_url"`
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
	viper.SetDefault("server.grpc_port", ":50052")
	viper.SetDefault("server.shutdown_timeout", "30s")

	viper.SetDefault("email.smtp.enabled", false)
	viper.SetDefault("email.rate_limit.emails_per_minute", 60)
	viper.SetDefault("email.rate_limit.max_burst", 10)
	viper.SetDefault("email.maintenance.enabled", true)
	viper.SetDefault("email.maintenance.frequency", "5m")
	viper.SetDefault("email.maintenance.downtime_period", "30s")

	viper.SetDefault("monitor.metrics_port", ":9102")

	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.encoding", "json")
	viper.SetDefault("logger.output_path", "stdout")

	viper.SetDefault("trace.service_name", "email-service")
	viper.SetDefault("trace.version", "1.0.0")
	viper.SetDefault("trace.jaeger_url", "http://jaeger:14268/api/traces")

	viper.AutomaticEnv()
}

func validateConfig(config *Config) error {
	var errors []string

	if config.Server.GRPCPort == "" {
		errors = append(errors, "server.grpc_port is required")
	}

	if config.Email.SMTP.Enabled {
		if config.Email.SMTP.Host == "" {
			errors = append(errors, "email.smtp.host is required when SMTP is enabled")
		}
		if config.Email.SMTP.Port == "" {
			errors = append(errors, "email.smtp.port is required when SMTP is enabled")
		}
		if config.Email.SMTP.SenderEmail == "" {
			errors = append(errors, "email.smtp.sender_email is required when SMTP is enabled")
		}
	}

	if config.Email.RateLimit.EmailsPerMinute <= 0 {
		errors = append(errors, "email.rate_limit.emails_per_minute must be greater than 0")
	}
	if config.Email.RateLimit.MaxBurst <= 0 {
		errors = append(errors, "email.rate_limit.max_burst must be greater than 0")
	}

	if config.Monitor.MetricsPort == "" {
		errors = append(errors, "monitor.metrics_port is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf("config validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}
