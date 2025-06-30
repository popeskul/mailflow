package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Monitor MonitorConfig `mapstructure:"monitor"`
	Trace   TraceConfig   `mapstructure:"trace"`
	Email   EmailConfig   `mapstructure:"email"`
}

type ServerConfig struct {
	HTTPPort string `mapstructure:"http_port"`
	GRPCPort string `mapstructure:"grpc_port"`
}

type MonitorConfig struct {
	MetricsPort string `mapstructure:"metrics_port"`
}

type EmailConfig struct {
	ServiceAddress string `mapstructure:"service_address"`
	Timeout        string `mapstructure:"timeout"`
}

type TraceConfig struct {
	ServiceName string `mapstructure:"service_name"`
	JaegerURL   string `mapstructure:"jaeger_url"` // Keep for backwards compatibility with config
	Version     string `mapstructure:"version"`
}

func LoadConfig() (*Config, error) {
	setDefaultConfig()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/app/configs")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, use defaults
			fmt.Println("Config file not found, using defaults")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	return &cfg, nil
}

func setDefaultConfig() {
	viper.SetDefault("server.http_port", ":8080")
	viper.SetDefault("server.grpc_port", ":50051")
	viper.SetDefault("monitor.metrics_port", ":9101")
	viper.SetDefault("email.service_address", "email-service:50052")
	viper.SetDefault("email.timeout", "30s")
	viper.SetDefault("trace.service_name", "user-service")
	viper.SetDefault("trace.jaeger_url", "http://jaeger:14268/api/traces")
	viper.SetDefault("trace.version", "1.0.0")
}
