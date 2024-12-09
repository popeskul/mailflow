package config

import (
	"fmt"
	"github.com/spf13/viper"

	"github.com/popeskul/email-service-platform/user-service/internal/logger"
	"github.com/popeskul/email-service-platform/user-service/internal/tracing"
)

type Config struct {
	GRPC struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"grpc"`

	Metrics struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"metrics"`

	HTTP struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"http"`

	Email struct {
		Address string `mapstructure:"address"`
		Timeout string `mapstructure:"timeout"`
	} `mapstructure:"email"`

	Logger logger.Config `mapstructure:"logger"`

	Tracing tracing.Config `mapstructure:"tracing"`
}

func LoadConfig() (*Config, error) {
	setDefaultConfig()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
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
	viper.SetDefault("grpc.port", ":50051")
	viper.SetDefault("metrics.port", ":9101")
	viper.SetDefault("http.port", ":8080")

	viper.SetDefault("email.address", "email-service:50052")
	viper.SetDefault("email.timeout", "5s")

	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.encoding", "json")
	viper.SetDefault("logger.output_path", "stdout")

	viper.AutomaticEnv()
}

func validateConfig(config *Config) error {
	if config.GRPC.Port == "" {
		return fmt.Errorf("grpc port is required")
	}
	if config.Metrics.Port == "" {
		return fmt.Errorf("metrics port is required")
	}
	if config.HTTP.Port == "" {
		return fmt.Errorf("http port is required")
	}
	if config.Email.Address == "" {
		return fmt.Errorf("email service address is required")
	}
	return nil
}
