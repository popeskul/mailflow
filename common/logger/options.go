package logger

import (
	"io"
	"os"
)

// Format constants
const (
	JSONFormat = "json"
)

// WithLogLevel sets the logging level
func WithLogLevel(level LogLevel) Option {
	return func(c *Config) {
		c.Level = level
	}
}

// WithOutputs adds outputs for logging
func WithOutputs(outputs ...io.Writer) Option {
	return func(c *Config) {
		c.Output = outputs
	}
}

// WithFileRotation configures log rotation
func WithFileRotation(filePath string, maxSize, maxBackups, maxAge int) Option {
	return func(c *Config) {
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
	return func(c *Config) {
		c.Format = JSONFormat
	}
}

// WithOutputPath sets the output path
func WithOutputPath(path string) Option {
	return func(c *Config) {
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