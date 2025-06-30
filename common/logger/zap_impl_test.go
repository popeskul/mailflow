package logger_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"

	"github.com/popeskul/mailflow/common/logger"
)

func TestZapLoggerMethods_Fail(t *testing.T) {
	testCases := []struct {
		name     string
		logFunc  func(logger.Logger)
		validate func(*testing.T, string)
	}{
		{
			name: "Info with nil fields",
			logFunc: func(l logger.Logger) {
				l.Info("Test info", logger.Field{Key: "nil_value", Value: nil})
			},
			validate: func(t *testing.T, output string) {
				assert.NotEmpty(t, output, "Log buffer should not be empty")
				assert.Contains(t, output, "Test info")
				assert.Contains(t, output, "nil_value")
			},
		},
		{
			name: "Error with empty fields",
			logFunc: func(l logger.Logger) {
				l.Error("Test error")
			},
			validate: func(t *testing.T, output string) {
				assert.NotEmpty(t, output, "Log buffer should not be empty")
				assert.Contains(t, output, "Test error")
				assert.Contains(t, output, "\"level\":\"ERROR\"")
			},
		},
		{
			name: "Warn with invalid field value",
			logFunc: func(l logger.Logger) {
				l.Warn("Test warning", logger.Field{Key: "chan", Value: make(chan int)})
			},
			validate: func(t *testing.T, output string) {
				assert.NotEmpty(t, output, "Log buffer should not be empty")
				assert.Contains(t, output, "Test warning")
				assert.Contains(t, output, "\"level\":\"WARN\"")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := createTestLogger(&buf)

			tc.logFunc(l)
			err := l.Sync()
			assert.NoError(t, err)

			time.Sleep(10 * time.Millisecond)
			tc.validate(t, buf.String())
		})
	}
}

func TestZapLoggerMethods_Success(t *testing.T) {
	testCases := []struct {
		name     string
		logFunc  func(logger.Logger)
		validate func(*testing.T, string)
	}{
		{
			name: "Debug log with fields",
			logFunc: func(l logger.Logger) {
				l.Debug("Test debug",
					logger.Field{Key: "user", Value: "admin"},
					logger.Field{Key: "action", Value: "login"},
				)
			},
			validate: func(t *testing.T, output string) {
				assert.NotEmpty(t, output, "Log buffer should not be empty")
				assert.Contains(t, output, "Test debug")
				assert.Contains(t, output, "\"level\":\"DEBUG\"")
				assert.Contains(t, output, "user")
				assert.Contains(t, output, "admin")
				assert.Contains(t, output, "action")
				assert.Contains(t, output, "login")
			},
		},
		{
			name: "Info log with fields",
			logFunc: func(l logger.Logger) {
				l.Info("Test info",
					logger.Field{Key: "status", Value: "success"},
					logger.Field{Key: "duration", Value: 100},
				)
			},
			validate: func(t *testing.T, output string) {
				assert.NotEmpty(t, output, "Log buffer should not be empty")
				assert.Contains(t, output, "Test info")
				assert.Contains(t, output, "\"level\":\"INFO\"")
				assert.Contains(t, output, "status")
				assert.Contains(t, output, "success")
				assert.Contains(t, output, "duration")
				assert.Contains(t, output, "100")
			},
		},
		{
			name: "Warn log with fields",
			logFunc: func(l logger.Logger) {
				l.Warn("Test warning",
					logger.Field{Key: "component", Value: "cache"},
					logger.Field{Key: "latency", Value: 500},
				)
			},
			validate: func(t *testing.T, output string) {
				assert.NotEmpty(t, output, "Log buffer should not be empty")
				assert.Contains(t, output, "Test warning")
				assert.Contains(t, output, "\"level\":\"WARN\"")
				assert.Contains(t, output, "component")
				assert.Contains(t, output, "cache")
				assert.Contains(t, output, "latency")
				assert.Contains(t, output, "500")
			},
		},
		{
			name: "Error log with fields",
			logFunc: func(l logger.Logger) {
				l.Error("Test error",
					logger.Field{Key: "error_code", Value: "ERR001"},
					logger.Field{Key: "details", Value: "Connection failed"},
				)
			},
			validate: func(t *testing.T, output string) {
				assert.NotEmpty(t, output, "Log buffer should not be empty")
				assert.Contains(t, output, "Test error")
				assert.Contains(t, output, "\"level\":\"ERROR\"")
				assert.Contains(t, output, "error_code")
				assert.Contains(t, output, "ERR001")
				assert.Contains(t, output, "details")
				assert.Contains(t, output, "Connection failed")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := createTestLogger(&buf)

			tc.logFunc(l)
			err := l.Sync()
			assert.NoError(t, err)

			time.Sleep(10 * time.Millisecond)
			tc.validate(t, buf.String())
		})
	}
}

func TestZapLoggerLevels(t *testing.T) {
	var buf bytes.Buffer

	l := logger.NewZapLogger(
		logger.WithOutputs(&buf),
		logger.WithLogLevel(logger.DebugLevel),
		logger.WithJSONFormat(),
	)

	testCases := []struct {
		level   string
		logFunc func(string, ...logger.Field)
		message string
		fields  []logger.Field
	}{
		{
			level:   "DEBUG",
			logFunc: l.Debug,
			message: "Debug message",
			fields: []logger.Field{
				{Key: "debug_key", Value: "debug_value"},
			},
		},
		{
			level:   "INFO",
			logFunc: l.Info,
			message: "Info message",
			fields: []logger.Field{
				{Key: "info_key", Value: "info_value"},
			},
		},
		{
			level:   "WARN",
			logFunc: l.Warn,
			message: "Warning message",
			fields: []logger.Field{
				{Key: "warn_key", Value: "warn_value"},
			},
		},
		{
			level:   "ERROR",
			logFunc: l.Error,
			message: "Error message",
			fields: []logger.Field{
				{Key: "error_key", Value: "error_value"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.level, func(t *testing.T) {
			buf.Reset()
			tc.logFunc(tc.message, tc.fields...)

			output := buf.String()
			assert.NotEmpty(t, output, "Log buffer should not be empty")
			assert.Contains(t, output, tc.message)
			assert.Contains(t, output, fmt.Sprintf("\"level\":\"%s\"", tc.level))

			for _, field := range tc.fields {
				assert.Contains(t, output, field.Key)
				assert.Contains(t, output, fmt.Sprint(field.Value))
			}
		})
	}
}

func TestZapLoggerFatal(t *testing.T) {
	// Calling Fatal in a separate process
	if os.Getenv("BE_FATAL") == "1" {
		l := logger.NewZapLogger(
			logger.WithOutputs(os.Stdout), // Write to stdout instead of the buffer
			logger.WithLogLevel(logger.DebugLevel),
			logger.WithJSONFormat(),
		)

		l.Fatal("Fatal message",
			logger.Field{Key: "fatal_key", Value: "fatal_value"},
		)
		return
	}

	// Run the test in a separate process
	cmd := exec.Command(os.Args[0], "-test.run=TestZapLoggerFatal")
	cmd.Env = append(os.Environ(), "BE_FATAL=1")
	output, err := cmd.CombinedOutput()

	// Check that the process failed with an error (as it should be for Fatal)
	var e *exec.ExitError
	if !errors.As(err, &e) || e.Success() {
		t.Fatalf("Process ran with err %v, want exit status 1", err)
	}

	outputStr := string(output)
	assert.Contains(t, outputStr, "Fatal message")
	assert.Contains(t, outputStr, "\"level\":\"FATAL\"")
	assert.Contains(t, outputStr, "fatal_key")
	assert.Contains(t, outputStr, "fatal_value")
}

func TestSync_Fail(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{"Nil logger sync"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l := logger.NewZapLogger()
			err := l.Sync()
			assert.NoError(t, err)
		})
	}
}

func createTestLogger(writer io.Writer) logger.Logger {
	// Create a MultiWriter that writes to both the test buffer and stdout
	// This will help with debugging while ensuring the test buffer gets the output
	multiWriter := io.MultiWriter(writer, os.Stdout)

	return logger.NewZapLogger(
		logger.WithOutputs(multiWriter),
		logger.WithLogLevel(logger.DebugLevel),
		logger.WithJSONFormat(),
	)
}

func TestLoggerSync(t *testing.T) {
	errorWriter := &errorWriter{
		err: fmt.Errorf("test sync error"),
	}

	l := logger.NewZapLogger(
		logger.WithOutputs(errorWriter),
		logger.WithLogLevel(logger.DebugLevel),
	)

	err := l.Sync()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "logger sync: test sync error")
	assert.Contains(t, err.Error(), "sugar logger sync: test sync error")
}

func TestDebugEmptyMessage(t *testing.T) {
	var buf bytes.Buffer

	l := logger.NewZapLogger(
		logger.WithOutputs(&buf),
		logger.WithLogLevel(logger.DebugLevel),
		logger.WithJSONFormat(),
	)

	l.Debug("", logger.Field{Key: "test", Value: "value"})

	assert.Empty(t, buf.String(), "Buffer should be empty when message is empty")

	l.Debug("not empty", logger.Field{Key: "test", Value: "value"})

	assert.NotEmpty(t, buf.String(), "Buffer should not be empty when message is not empty")
}

func TestConvertLogLevel(t *testing.T) {
	testCases := []struct {
		name     string
		level    logger.LogLevel
		expected zapcore.Level
	}{
		{
			name:     "DebugLevel",
			level:    logger.DebugLevel,
			expected: zapcore.DebugLevel,
		},
		{
			name:     "WarnLevel",
			level:    logger.WarnLevel,
			expected: zapcore.WarnLevel,
		},
		{
			name:     "ErrorLevel",
			level:    logger.ErrorLevel,
			expected: zapcore.ErrorLevel,
		},
		{
			name:     "FatalLevel",
			level:    logger.FatalLevel,
			expected: zapcore.FatalLevel,
		},
		{
			name:     "InvalidLevel",
			level:    logger.LogLevel(999),
			expected: zapcore.InfoLevel, // default
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := logger.NewZapLogger(
				logger.WithOutputs(&buf),
				logger.WithLogLevel(tc.level),
				logger.WithJSONFormat(),
			)

			l.Debug("test")
			l.Info("test")
			l.Warn("test")
			l.Error("test")

			output := buf.String()

			hasDebug := strings.Contains(output, "\"level\":\"DEBUG\"")
			hasInfo := strings.Contains(output, "\"level\":\"INFO\"")
			hasWarn := strings.Contains(output, "\"level\":\"WARN\"")
			hasError := strings.Contains(output, "\"level\":\"ERROR\"")

			switch tc.expected {
			case zapcore.DebugLevel:
				assert.True(t, hasDebug && hasInfo && hasWarn && hasError)
			case zapcore.InfoLevel:
				assert.False(t, hasDebug)
				assert.True(t, hasInfo && hasWarn && hasError)
			case zapcore.WarnLevel:
				assert.False(t, hasDebug || hasInfo)
				assert.True(t, hasWarn && hasError)
			case zapcore.ErrorLevel:
				assert.False(t, hasDebug || hasInfo || hasWarn)
				assert.True(t, hasError)
			case zapcore.FatalLevel:
				assert.False(t, hasDebug || hasInfo || hasWarn || hasError)
			default:
				t.Errorf("Unexpected log level: %v", tc.expected)
			}
		})
	}
}

// errorWriter implements io.Writer with an error on Sync
type errorWriter struct {
	err error
}

func (w *errorWriter) Write(p []byte) (n int, err error) {
	return len(p), nil // Write works fine
}

func (w *errorWriter) Sync() error {
	return w.err // When Syncing, we return an error
}
