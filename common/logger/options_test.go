package logger_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/popeskul/mailflow/common/logger"
)

func TestWithLogLevel_Success(t *testing.T) {
	testCases := []struct {
		name     string
		level    logger.LogLevel
		expected logger.LogLevel
	}{
		{"Debug Level", logger.DebugLevel, logger.DebugLevel},
		{"Info Level", logger.InfoLevel, logger.InfoLevel},
		{"Warn Level", logger.WarnLevel, logger.WarnLevel},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &logger.Config{}
			opt := logger.WithLogLevel(tc.level)
			opt(config)
			assert.Equal(t, tc.expected, config.Level)
		})
	}
}

func TestWithLogLevel_Fail(t *testing.T) {
	testCases := []struct {
		name     string
		level    logger.LogLevel
		expected logger.LogLevel
	}{
		{"Invalid Level", logger.LogLevel(999), logger.LogLevel(999)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &logger.Config{}
			opt := logger.WithLogLevel(tc.level)
			opt(config)
			assert.Equal(t, tc.expected, config.Level)
		})
	}
}

func TestWithOutputs_Success(t *testing.T) {
	testCases := []struct {
		name     string
		writers  []io.Writer
		expected int
	}{
		{"Single Writer", []io.Writer{&bytes.Buffer{}}, 1},
		{"Multiple Writers", []io.Writer{&bytes.Buffer{}, &bytes.Buffer{}}, 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &logger.Config{}
			opt := logger.WithOutputs(tc.writers...)
			opt(config)
			assert.Len(t, config.Output, tc.expected)
		})
	}
}

func TestWithOutputs_Fail(t *testing.T) {
	testCases := []struct {
		name     string
		writers  []io.Writer
		expected int
	}{
		{"Nil Writer", []io.Writer{nil}, 0},
		{"Mixed Valid and Nil Writers", []io.Writer{&bytes.Buffer{}, nil}, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &logger.Config{}
			opt := logger.WithOutputs(tc.writers...)
			opt(config)

			nonNilWriters := filterNonNilWriters(config.Output)
			assert.Len(t, nonNilWriters, tc.expected)
		})
	}
}

func filterNonNilWriters(writers []io.Writer) []io.Writer {
	var filtered []io.Writer
	for _, w := range writers {
		if w != nil {
			filtered = append(filtered, w)
		}
	}
	return filtered
}

func TestWithFileRotation_Success(t *testing.T) {
	testCases := []struct {
		name     string
		filePath string
		maxSize  int
		maxBkp   int
		maxAge   int
	}{
		{
			name:     "Valid Rotation Config",
			filePath: "/tmp/test.log",
			maxSize:  10,
			maxBkp:   3,
			maxAge:   30,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &logger.Config{}
			opt := logger.WithFileRotation(tc.filePath, tc.maxSize, tc.maxBkp, tc.maxAge)
			opt(config)

			assert.Equal(t, tc.filePath, config.FilePath)
			assert.Equal(t, tc.maxSize, config.MaxSize)
			assert.Equal(t, tc.maxBkp, config.MaxBackups)
			assert.Equal(t, tc.maxAge, config.MaxAge)
		})
	}
}

func TestWithFileRotation_Fail(t *testing.T) {
	testCases := []struct {
		name       string
		filePath   string
		maxSize    int
		maxBackups int
		maxAge     int
		expected   int
	}{
		{
			name:       "Negative Max Size",
			filePath:   "/tmp/test.log",
			maxSize:    -1,
			maxBackups: 3,
			maxAge:     30,
			expected:   0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &logger.Config{}
			opt := logger.WithFileRotation(tc.filePath, tc.maxSize, tc.maxBackups, tc.maxAge)
			opt(config)

			assert.Equal(t, tc.expected, config.MaxSize)
		})
	}
}

func TestWithJSONFormat_Success(t *testing.T) {
	testCases := []struct {
		name   string
		format string
	}{
		{"Set JSON Format", "json"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &logger.Config{}
			opt := logger.WithJSONFormat()
			opt(config)
			assert.Equal(t, tc.format, config.Format)
		})
	}
}

func TestWithJSONFormat_Fail(t *testing.T) {
	testCases := []struct {
		name           string
		modifyConfig   func(*logger.Config)
		expectedFormat string
	}{
		{
			"Overwrite Existing Format",
			func(cfg *logger.Config) { cfg.Format = "text" },
			"json",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &logger.Config{}
			if tc.modifyConfig != nil {
				tc.modifyConfig(config)
			}
			opt := logger.WithJSONFormat()
			opt(config)
			assert.Equal(t, tc.expectedFormat, config.Format)
		})
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := logger.DefaultOptions()

	config := &logger.Config{}

	for _, opt := range opts {
		opt(config)
	}

	assert.Equal(t, logger.InfoLevel, config.Level, "Default log level should be InfoLevel")

	assert.Len(t, config.Output, 1, "Should have exactly one output")
	assert.Equal(t, os.Stdout, config.Output[0], "Default output should be os.Stdout")
}
