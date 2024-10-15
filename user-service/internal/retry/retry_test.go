package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/popeskul/mailflow/user-service/internal/retry"
)

func TestRetrier_Success(t *testing.T) {
	retrier := retry.New(retry.DefaultExponentialBackoff())

	callCount := 0
	err := retrier.Do(context.Background(), func(ctx context.Context) error {
		callCount++
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestRetrier_RetryOnError(t *testing.T) {
	strategy := &retry.ExponentialBackoff{
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		MaxAttempts:  3,
		Jitter:       false,
	}
	retrier := retry.New(strategy)

	callCount := 0
	testErr := errors.New("test error")

	start := time.Now()
	err := retrier.Do(context.Background(), func(ctx context.Context) error {
		callCount++
		if callCount < 3 {
			return testErr
		}
		return nil
	})
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Expected no error after retries, got %v", err)
	}
	if callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}
	// Should have delayed at least 10ms + 20ms = 30ms
	if elapsed < 30*time.Millisecond {
		t.Errorf("Expected at least 30ms elapsed, got %v", elapsed)
	}
}

func TestRetrier_MaxAttempts(t *testing.T) {
	strategy := &retry.ExponentialBackoff{
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		MaxAttempts:  3,
		Jitter:       false,
	}
	retrier := retry.New(strategy)

	callCount := 0
	testErr := errors.New("persistent error")

	err := retrier.Do(context.Background(), func(ctx context.Context) error {
		callCount++
		return testErr
	})

	if err != testErr {
		t.Errorf("Expected test error, got %v", err)
	}
	if callCount != 3 {
		t.Errorf("Expected 3 calls (max attempts), got %d", callCount)
	}
}

func TestRetrier_ContextCancellation(t *testing.T) {
	strategy := &retry.ExponentialBackoff{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		MaxAttempts:  5,
		Jitter:       false,
	}
	retrier := retry.New(strategy)

	ctx, cancel := context.WithCancel(context.Background())
	callCount := 0
	testErr := errors.New("test error")

	// Cancel context after first attempt
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := retrier.Do(ctx, func(ctx context.Context) error {
		callCount++
		return testErr
	})

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call before cancellation, got %d", callCount)
	}
}

// TestRetryableError tests custom retryable error handling
type testRetryableError struct {
	retryable bool
}

func (e testRetryableError) Error() string {
	return "test retryable error"
}

func (e testRetryableError) Retryable() bool {
	return e.retryable
}

func TestRetrier_NonRetryableError(t *testing.T) {
	retrier := retry.New(retry.DefaultExponentialBackoff())

	callCount := 0
	nonRetryableErr := testRetryableError{retryable: false}

	err := retrier.Do(context.Background(), func(ctx context.Context) error {
		callCount++
		return nonRetryableErr
	})

	if err != nonRetryableErr {
		t.Errorf("Expected non-retryable error, got %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call (no retry for non-retryable error), got %d", callCount)
	}
}

func TestExponentialBackoff_NextDelay(t *testing.T) {
	eb := &retry.ExponentialBackoff{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		MaxAttempts:  5,
		Jitter:       false,
	}

	tests := []struct {
		attempt       int
		expectedDelay time.Duration
	}{
		{0, 100 * time.Millisecond},
		{1, 100 * time.Millisecond},
		{2, 200 * time.Millisecond},
		{3, 400 * time.Millisecond},
		{4, 800 * time.Millisecond},
		{5, 1 * time.Second}, // Max delay
	}

	for _, tt := range tests {
		delay := eb.NextDelay(tt.attempt)
		if delay != tt.expectedDelay {
			t.Errorf("Attempt %d: expected delay %v, got %v", tt.attempt, tt.expectedDelay, delay)
		}
	}
}
