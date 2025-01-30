package logger

import (
	"context"
	"io"
)

// LogLevel represents the importance level of the log message
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// Keys for logging fields
const (
	TraceIDKey        = contextKey("trace_id")
	FieldKeyTraceID   = "trace_id"
	FieldKeyUserID    = "user_id"
	FieldKeyRequestID = "request_id"
	FieldKeyOperation = "operation"
	FieldKeyComponent = "component"
)

// contextKey - type for context keys
type contextKey string

// Logger is an abstraction for logging
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)

	WithContext(ctx context.Context) Logger
	WithFields(fields Fields) Logger
	Named(name string) Logger

	Sync() error
}

// Field - a structure for adding additional information to logs
type Field struct {
	Key   string
	Value interface{}
}

// Fields - a collection of fields
type Fields map[string]interface{}

// LoggerConfig contains the settings for the logger
type LoggerConfig struct {
	Level             LogLevel
	Output            []io.Writer
	Format            string
	DisableCaller     bool
	DisableStacktrace bool
	FilePath          string
	MaxSize           int // MB
	MaxBackups        int
	MaxAge            int // days
	OutputPath        string
}

// Option - type for configuring the logger
type Option func(*LoggerConfig)
