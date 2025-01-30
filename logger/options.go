package logger

import (
	"io"
	"os"
)

// WithLogLevel sets the logging level
func WithLogLevel(level LogLevel) Option {
	return func(c *LoggerConfig) {
		c.Level = level
	}
}

// WithOutputs adds outputs for logging
func WithOutputs(outputs ...io.Writer) Option {
	return func(c *LoggerConfig) {
		c.Output = outputs
	}
}

// WithFileRotation configures log rotation
func WithFileRotation(filePath string, maxSize, maxBackups, maxAge int) Option {
	return func(c *LoggerConfig) {
		c.FilePath = filePath
		if maxSize < 0 {
			maxSize = 0
		}
		c.MaxSize = maxSize
		c.MaxBackups = maxBackups
		c.MaxAge = maxAge
	}
}

// WithJSONFormat sets the JSON format
func WithJSONFormat() Option {
	return func(c *LoggerConfig) {
		c.Format = "json"
	}
}

// WithOutputPath sets the output path
func WithOutputPath(path string) Option {
	return func(c *LoggerConfig) {
		c.OutputPath = path
	}
}

// DefaultOptions returns the default settings
func DefaultOptions() []Option {
	return []Option{
		WithLogLevel(InfoLevel),
		WithOutputs(os.Stdout),
	}
}

// StandardFields creates standard fields for logging
func StandardFields(traceID, userID, requestID string) []Field {
	return []Field{
		{Key: FieldKeyTraceID, Value: traceID},
		{Key: FieldKeyUserID, Value: userID},
		{Key: FieldKeyRequestID, Value: requestID},
	}
}
