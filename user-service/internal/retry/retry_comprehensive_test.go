package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultExponentialBackoff_Success(t *testing.T) {
	tests := []struct {
		name           string
		expectedConfig *ExponentialBackoff
	}{
		{
			name: "default exponential backoff configuration",
			expectedConfig: &ExponentialBackoff{
				InitialDelay: 100 * time.Millisecond,
				MaxDelay:     30 * time.Second,
				Multiplier:   2.0,
				MaxAttempts:  5,
				Jitter:       true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultExponentialBackoff()

			assert.Equal(t, tt.expectedConfig.InitialDelay, config.InitialDelay)
			assert.Equal(t, tt.expectedConfig.MaxDelay, config.MaxDelay)
			assert.Equal(t, tt.expectedConfig.Multiplier, config.Multiplier)
			assert.Equal(t, tt.expectedConfig.MaxAttempts, config.MaxAttempts)
			assert.Equal(t, tt.expectedConfig.Jitter, config.Jitter)
		})
	}
}

func TestExponentialBackoff_NextDelay_Success(t *testing.T) {
	tests := []struct {
		name             string
		config           *ExponentialBackoff
		attempt          int
		expectedMinDelay time.Duration
		expectedMaxDelay time.Duration
	}{
		{
			name: "first attempt delay",
			config: &ExponentialBackoff{
				InitialDelay: 100 * time.Millisecond,
				MaxDelay:     30 * time.Second,
				Multiplier:   2.0,
				Jitter:       false,
			},
			attempt:          0,
			expectedMinDelay: 100 * time.Millisecond,
			expectedMaxDelay: 100 * time.Millisecond,
		},
		{
			name: "second attempt delay",
			config: &ExponentialBackoff{
				InitialDelay: 100 * time.Millisecond,
				MaxDelay:     30 * time.Second,
				Multiplier:   2.0,
				Jitter:       false,
			},
			attempt:          1,
			expectedMinDelay: 100 * time.Millisecond,
			expectedMaxDelay: 100 * time.Millisecond,
		},
		{
			name: "third attempt delay",
			config: &ExponentialBackoff{
				InitialDelay: 100 * time.Millisecond,
				MaxDelay:     30 * time.Second,
				Multiplier:   2.0,
				Jitter:       false,
			},
			attempt:          2,
			expectedMinDelay: 200 * time.Millisecond,
			expectedMaxDelay: 200 * time.Millisecond,
		},
		{
			name: "max delay reached",
			config: &ExponentialBackoff{
				InitialDelay: 1 * time.Second,
				MaxDelay:     5 * time.Second,
				Multiplier:   2.0,
				Jitter:       false,
			},
			attempt:          10,
			expectedMinDelay: 5 * time.Second,
			expectedMaxDelay: 5 * time.Second,
		},
		{
			name: "with jitter",
			config: &ExponentialBackoff{
				InitialDelay: 100 * time.Millisecond,
				MaxDelay:     30 * time.Second,
				Multiplier:   2.0,
				Jitter:       true,
			},
			attempt:          2,
			expectedMinDelay: 200 * time.Millisecond,
			expectedMaxDelay: 260 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := tt.config.NextDelay(tt.attempt)

			assert.GreaterOrEqual(t, delay, tt.expectedMinDelay)
			assert.LessOrEqual(t, delay, tt.expectedMaxDelay)
		})
	}
}

func TestExponentialBackoff_ShouldRetry_Success(t *testing.T) {
	tests := []struct {
		name     string
		config   *ExponentialBackoff
		attempt  int
		expected bool
	}{
		{
			name:     "should retry within max attempts",
			config:   &ExponentialBackoff{MaxAttempts: 5},
			attempt:  3,
			expected: true,
		},
		{
			name:     "should retry at last attempt",
			config:   &ExponentialBackoff{MaxAttempts: 5},
			attempt:  4,
			expected: true,
		},
		{
			name:     "first attempt should retry",
			config:   &ExponentialBackoff{MaxAttempts: 5},
			attempt:  0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.ShouldRetry(tt.attempt)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExponentialBackoff_ShouldRetry_Fail(t *testing.T) {
	tests := []struct {
		name     string
		config   *ExponentialBackoff
		attempt  int
		expected bool
	}{
		{
			name:     "should not retry beyond max attempts",
			config:   &ExponentialBackoff{MaxAttempts: 5},
			attempt:  5,
			expected: false,
		},
		{
			name:     "should not retry far beyond max attempts",
			config:   &ExponentialBackoff{MaxAttempts: 3},
			attempt:  10,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.ShouldRetry(tt.attempt)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNew_Success(t *testing.T) {
	tests := []struct {
		name     string
		strategy Strategy
	}{
		{
			name: "create retrier with custom strategy",
			strategy: &ExponentialBackoff{
				InitialDelay: 50 * time.Millisecond,
				MaxDelay:     10 * time.Second,
				Multiplier:   1.5,
				MaxAttempts:  3,
				Jitter:       false,
			},
		},
		{
			name:     "create retrier with nil strategy",
			strategy: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrier := New(tt.strategy)

			assert.NotNil(t, retrier)
			assert.NotNil(t, retrier.strategy)

			if tt.strategy != nil {
				assert.Equal(t, tt.strategy, retrier.strategy)
			}
		})
	}
}

func TestRetrier_Do_Success(t *testing.T) {
	tests := []struct {
		name          string
		strategy      Strategy
		fn            RetryableFunc
		expectedCalls int
	}{
		{
			name: "succeed on first attempt",
			strategy: &ExponentialBackoff{
				InitialDelay: 1 * time.Millisecond,
				MaxAttempts:  3,
				Jitter:       false,
			},
			fn: func(ctx context.Context) error {
				return nil
			},
			expectedCalls: 1,
		},
		{
			name: "succeed on second attempt",
			strategy: &ExponentialBackoff{
				InitialDelay: 1 * time.Millisecond,
				MaxAttempts:  3,
				Jitter:       false,
			},
			fn: func() RetryableFunc {
				calls := 0
				return func(ctx context.Context) error {
					calls++
					if calls == 1 {
						return errors.New("temporary error")
					}
					return nil
				}
			}(),
			expectedCalls: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrier := New(tt.strategy)

			err := retrier.Do(context.Background(), tt.fn)

			assert.NoError(t, err)
		})
	}
}

func TestRetrier_Do_Fail(t *testing.T) {
	tests := []struct {
		name          string
		strategy      Strategy
		fn            RetryableFunc
		expectedError string
	}{
		{
			name: "fail after max attempts",
			strategy: &ExponentialBackoff{
				InitialDelay: 1 * time.Millisecond,
				MaxAttempts:  2,
				Jitter:       false,
			},
			fn: func(ctx context.Context) error {
				return errors.New("persistent error")
			},
			expectedError: "persistent error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrier := New(tt.strategy)

			ctx := context.Background()

			err := retrier.Do(ctx, tt.fn)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

type testRetryableError struct {
	msg       string
	retryable bool
}

func (e *testRetryableError) Error() string {
	return e.msg
}

func (e *testRetryableError) Retryable() bool {
	return e.retryable
}

func TestRetrier_Do_RetryableError_Success(t *testing.T) {
	tests := []struct {
		name     string
		strategy Strategy
		fn       RetryableFunc
	}{
		{
			name: "retryable error should be retried",
			strategy: &ExponentialBackoff{
				InitialDelay: 1 * time.Millisecond,
				MaxAttempts:  3,
				Jitter:       false,
			},
			fn: func() RetryableFunc {
				calls := 0
				return func(ctx context.Context) error {
					calls++
					if calls == 1 {
						return &testRetryableError{msg: "retryable error", retryable: true}
					}
					return nil
				}
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrier := New(tt.strategy)

			err := retrier.Do(context.Background(), tt.fn)

			assert.NoError(t, err)
		})
	}
}

func TestRetrier_Do_RetryableError_Fail(t *testing.T) {
	tests := []struct {
		name          string
		strategy      Strategy
		fn            RetryableFunc
		expectedError string
	}{
		{
			name: "non-retryable error should not be retried",
			strategy: &ExponentialBackoff{
				InitialDelay: 1 * time.Millisecond,
				MaxAttempts:  3,
				Jitter:       false,
			},
			fn: func(ctx context.Context) error {
				return &testRetryableError{msg: "non-retryable error", retryable: false}
			},
			expectedError: "non-retryable error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrier := New(tt.strategy)

			err := retrier.Do(context.Background(), tt.fn)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestWithRetry_Success(t *testing.T) {
	tests := []struct {
		name string
		fn   RetryableFunc
		opts []Option
	}{
		{
			name: "with default options",
			fn: func(ctx context.Context) error {
				return nil
			},
			opts: []Option{},
		},
		{
			name: "with custom strategy",
			fn: func(ctx context.Context) error {
				return nil
			},
			opts: []Option{
				WithStrategy(&ExponentialBackoff{
					InitialDelay: 10 * time.Millisecond,
					MaxAttempts:  2,
					Jitter:       false,
				}),
			},
		},
		{
			name: "with max attempts option",
			fn: func(ctx context.Context) error {
				return nil
			},
			opts: []Option{
				WithMaxAttempts(2),
			},
		},
		{
			name: "with initial delay option",
			fn: func(ctx context.Context) error {
				return nil
			},
			opts: []Option{
				WithInitialDelay(5 * time.Millisecond),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WithRetry(context.Background(), tt.fn, tt.opts...)

			assert.NoError(t, err)
		})
	}
}

func TestWithRetry_Fail(t *testing.T) {
	tests := []struct {
		name          string
		fn            RetryableFunc
		opts          []Option
		expectedError string
	}{
		{
			name: "fail after retries",
			fn: func(ctx context.Context) error {
				return errors.New("persistent failure")
			},
			opts: []Option{
				WithMaxAttempts(2),
				WithInitialDelay(1 * time.Millisecond),
			},
			expectedError: "persistent failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WithRetry(context.Background(), tt.fn, tt.opts...)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestWithStrategy_Success(t *testing.T) {
	tests := []struct {
		name     string
		strategy Strategy
	}{
		{
			name: "set custom strategy",
			strategy: &ExponentialBackoff{
				InitialDelay: 10 * time.Millisecond,
				MaxAttempts:  2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{Strategy: DefaultExponentialBackoff()}
			option := WithStrategy(tt.strategy)
			option(config)

			assert.Equal(t, tt.strategy, config.Strategy)
		})
	}
}

func TestWithMaxAttempts_Success(t *testing.T) {
	tests := []struct {
		name        string
		maxAttempts int
	}{
		{
			name:        "set max attempts to 3",
			maxAttempts: 3,
		},
		{
			name:        "set max attempts to 1",
			maxAttempts: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{Strategy: DefaultExponentialBackoff()}
			option := WithMaxAttempts(tt.maxAttempts)
			option(config)

			if eb, ok := config.Strategy.(*ExponentialBackoff); ok {
				assert.Equal(t, tt.maxAttempts, eb.MaxAttempts)
			}
		})
	}
}

func TestWithInitialDelay_Success(t *testing.T) {
	tests := []struct {
		name         string
		initialDelay time.Duration
	}{
		{
			name:         "set initial delay to 50ms",
			initialDelay: 50 * time.Millisecond,
		},
		{
			name:         "set initial delay to 1s",
			initialDelay: 1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{Strategy: DefaultExponentialBackoff()}
			option := WithInitialDelay(tt.initialDelay)
			option(config)

			if eb, ok := config.Strategy.(*ExponentialBackoff); ok {
				assert.Equal(t, tt.initialDelay, eb.InitialDelay)
			}
		})
	}
}
