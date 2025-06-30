package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

// NewZapLogger creates a new logger based on Zap
func NewZapLogger(opts ...Option) Logger {
	config := &Config{
		Level:  InfoLevel,
		Output: []io.Writer{os.Stdout},
		Format: "text",
	}

	for _, opt := range opts {
		opt(config)
	}

	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = convertLogLevel(config.Level)

	if config.Format == JSONFormat {
		zapConfig.Encoding = JSONFormat
	}

	var writers []zapcore.WriteSyncer
	for _, w := range config.Output {
		writers = append(writers, zapcore.AddSync(w))
	}

	core := zapcore.NewCore(
		getEncoder(config),
		zapcore.NewMultiWriteSyncer(writers...),
		zapConfig.Level,
	)

	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	return &zapLogger{
		logger: logger,
		sugar:  logger.Sugar(),
	}
}

func getEncoder(config *Config) zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if config.Format == JSONFormat {
		return zapcore.NewJSONEncoder(encoderConfig)
	}
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func convertLogLevel(level LogLevel) zap.AtomicLevel {
	switch level {
	case DebugLevel:
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case InfoLevel:
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	case WarnLevel:
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case ErrorLevel:
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	case FatalLevel:
		return zap.NewAtomicLevelAt(zap.FatalLevel)
	default:
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	}
}

func (z *zapLogger) Debug(msg string, fields ...Field) {
	if msg == "" {
		return
	}
	z.logger.Debug(msg, convertFields(fields)...)
}

func (z *zapLogger) Info(msg string, fields ...Field) {
	z.logger.Info(msg, convertFields(fields)...)
}

func (z *zapLogger) Warn(msg string, fields ...Field) {
	z.logger.Warn(msg, convertFields(fields)...)
}

func (z *zapLogger) Error(msg string, fields ...Field) {
	z.logger.Error(msg, convertFields(fields)...)
}

func (z *zapLogger) Fatal(msg string, fields ...Field) {
	z.logger.Fatal(msg, convertFields(fields)...)
}

func (z *zapLogger) WithContext(ctx context.Context) Logger {
	if traceID := ctx.Value(TraceIDKey); traceID != nil {
		return &zapLogger{
			logger: z.logger.With(zap.Any("trace_id", traceID)),
			sugar:  z.sugar,
		}
	}
	return z
}

func (z *zapLogger) WithFields(fields Fields) Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return &zapLogger{
		logger: z.logger.With(zapFields...),
		sugar:  z.sugar,
	}
}

func (z *zapLogger) Named(name string) Logger {
	return &zapLogger{
		logger: z.logger.Named(name),
		sugar:  z.sugar.Named(name),
	}
}

func (z *zapLogger) Sync() error {
	var errs []error

	if z.logger != nil {
		if err := z.logger.Sync(); err != nil {
			if !strings.Contains(err.Error(), "bad file descriptor") {
				errs = append(errs, fmt.Errorf("logger sync: %v", err))
			}
		}
	}

	if z.sugar != nil {
		if err := z.sugar.Sync(); err != nil {
			if !strings.Contains(err.Error(), "bad file descriptor") {
				errs = append(errs, fmt.Errorf("sugar logger sync: %v", err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("sync errors: %v", errs)
	}

	return nil
}

func convertFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zapFields[i] = zap.Any(f.Key, f.Value)
	}
	return zapFields
}
