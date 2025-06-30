package logger

import (
	"context"
	"io"
	"strings"
)

// Log level string constants
const (
	DebugLevelStr = "debug"
	InfoLevelStr  = "info"
	WarnLevelStr  = "warn"
	ErrorLevelStr = "error"
	FatalLevelStr = "fatal"
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

// ParseLogLevel converts a string to LogLevel
func ParseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case DebugLevelStr:
		return DebugLevel
	case InfoLevelStr:
		return InfoLevel
	case WarnLevelStr, "warning":
		return WarnLevel
	case ErrorLevelStr:
		return ErrorLevel
	case FatalLevelStr:
		return FatalLevel
	default:
		return InfoLevel
	}
}

// String returns the string representation of LogLevel
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return DebugLevelStr
	case InfoLevel:
		return InfoLevelStr
	case WarnLevel:
		return WarnLevelStr
	case ErrorLevel:
		return ErrorLevelStr
	case FatalLevel:
		return FatalLevelStr
	default:
		return InfoLevelStr
	}
}

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

// Config contains the settings for the logger
type Config struct {
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
type Option func(*Config)