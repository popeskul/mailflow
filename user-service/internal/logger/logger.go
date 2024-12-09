package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	FieldKeyTraceID   = "trace_id"
	FieldKeyUserID    = "user_id"
	FieldKeyRequestID = "request_id"
	FieldKeyOperation = "operation"
	FieldKeyComponent = "component"
)

type Config struct {
	Level      string `mapstructure:"level"`
	Encoding   string `mapstructure:"encoding"`
	OutputPath string `mapstructure:"output_path"`
}

func NewLogger(cfg Config, serviceName string) (*zap.Logger, error) {
	level := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	if cfg.Level != "" {
		err := level.UnmarshalText([]byte(cfg.Level))
		if err != nil {
			return nil, err
		}
	}

	zapConfig := zap.Config{
		Level:       level,
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         cfg.Encoding,
		EncoderConfig:    getEncoderConfig(),
		OutputPaths:      []string{cfg.OutputPath},
		ErrorOutputPaths: []string{cfg.OutputPath},
	}

	logger, err := zapConfig.Build(
		zap.AddCallerSkip(1),
		zap.Fields(
			zap.String("service", serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	return logger, nil
}

func StandardFields(traceID, userID, requestID string) []zap.Field {
	return []zap.Field{
		zap.String(FieldKeyTraceID, traceID),
		zap.String(FieldKeyUserID, userID),
		zap.String(FieldKeyRequestID, requestID),
	}
}

func getEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}
