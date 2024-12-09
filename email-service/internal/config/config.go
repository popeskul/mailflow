package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"

	"github.com/popeskul/email-service-platform/email-service/internal/logger"
	"github.com/popeskul/email-service-platform/email-service/internal/tracing"
)

type Config struct {
	GRPC struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"grpc"`

	Metrics struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"metrics"`

	SMTP struct {
		Enabled  bool   `mapstructure:"enabled"`
		Host     string `mapstructure:"host"`
		Port     string `mapstructure:"port"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		From     string `mapstructure:"from"`
	} `mapstructure:"smtp"`

	RateLimit struct {
		RequestsPerMinute int `mapstructure:"rpm"`
		BurstSize         int `mapstructure:"burst"`
	} `mapstructure:"rate_limit"`

	Downtime struct {
		Interval time.Duration `mapstructure:"interval"`
		Duration time.Duration `mapstructure:"duration"`
		Enabled  bool          `mapstructure:"enabled"`
	} `mapstructure:"downtime"`

	Logger logger.Config `mapstructure:"logger"`

	Tracing tracing.Config `mapstructure:"tracing"`
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
	viper.SetDefault("grpc.port", ":50052")
	viper.SetDefault("metrics.port", ":9102")

	viper.SetDefault("rate_limit.rpm", 60)
	viper.SetDefault("rate_limit.burst", 10)

	viper.SetDefault("downtime.interval", "5m")
	viper.SetDefault("downtime.duration", "30s")
	viper.SetDefault("downtime.enabled", true)

	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.encoding", "json")
	viper.SetDefault("logger.output_path", "stdout")
	viper.SetDefault("smtp.enabled", false)

	viper.AutomaticEnv()
}

func validateConfig(config *Config) error {
	if config.SMTP.Enabled {
		if config.SMTP.Host == "" || config.SMTP.Port == "" {
			return fmt.Errorf("smtp host and port are required when SMTP is enabled")
		}
	}
	if config.GRPC.Port == "" {
		return fmt.Errorf("grpc port is required")
	}
	if config.Metrics.Port == "" {
		return fmt.Errorf("metrics port is required")
	}
	return nil
}
