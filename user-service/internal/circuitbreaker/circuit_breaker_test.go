package circuitbreaker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/popeskul/mailflow/user-service/internal/circuitbreaker"
)

func TestCircuitBreaker_ClosedState(t *testing.T) {
	cb := circuitbreaker.New(&circuitbreaker.Config{
		FailureThreshold: 3,
		SuccessThreshold: 2,
		Timeout:          1 * time.Second,
		MaxRequests:      2,
	})

	ctx := context.Background()

	// Successful calls should work
	for i := 0; i < 5; i++ {
		err := cb.Execute(ctx, func(ctx context.Context) error {
			return nil
		})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	}

	if cb.GetState() != circuitbreaker.StateClosed {
		t.Error("Circuit should remain closed after successful calls")
	}
}

func TestCircuitBreaker_OpenState(t *testing.T) {
	cb := circuitbreaker.New(&circuitbreaker.Config{
		FailureThreshold: 3,
		SuccessThreshold: 2,
		Timeout:          1 * time.Second,
		MaxRequests:      2,
	})

	ctx := context.Background()
	testErr := errors.New("test error")

	// Cause failures to open the circuit
	for i := 0; i < 3; i++ {
		cb.Execute(ctx, func(ctx context.Context) error {
			return testErr
		})
	}

	if cb.GetState() != circuitbreaker.StateOpen {
		t.Error("Circuit should be open after reaching failure threshold")
	}

	// Calls should fail immediately when circuit is open
	err := cb.Execute(ctx, func(ctx context.Context) error {
		return nil
	})
	if err != circuitbreaker.ErrCircuitOpen {
		t.Errorf("Expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreaker_HalfOpenState(t *testing.T) {
	cb := circuitbreaker.New(&circuitbreaker.Config{
		FailureThreshold: 2,
		SuccessThreshold: 2,
		Timeout:          100 * time.Millisecond,
		MaxRequests:      2,
	})

	ctx := context.Background()
	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Execute(ctx, func(ctx context.Context) error {
			return testErr
		})
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// First call should be allowed (half-open state)
	err := cb.Execute(ctx, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected no error in half-open state, got %v", err)
	}

	// Second successful call should close the circuit
	err = cb.Execute(ctx, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if cb.GetState() != circuitbreaker.StateClosed {
		t.Error("Circuit should be closed after successful calls in half-open state")
	}
}

func TestCircuitBreaker_HalfOpenToOpen(t *testing.T) {
	cb := circuitbreaker.New(&circuitbreaker.Config{
		FailureThreshold: 2,
		SuccessThreshold: 2,
		Timeout:          100 * time.Millisecond,
		MaxRequests:      2,
	})

	ctx := context.Background()
	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Execute(ctx, func(ctx context.Context) error {
			return testErr
		})
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Failure in half-open state should reopen the circuit
	cb.Execute(ctx, func(ctx context.Context) error {
		return testErr
	})

	if cb.GetState() != circuitbreaker.StateOpen {
		t.Error("Circuit should be open after failure in half-open state")
	}
}

func TestCircuitBreaker_MaxRequests(t *testing.T) {
	cb := circuitbreaker.New(&circuitbreaker.Config{
		FailureThreshold: 2,
		SuccessThreshold: 3, // Higher than MaxRequests to keep in half-open
		Timeout:          100 * time.Millisecond,
		MaxRequests:      2,
	})

	ctx := context.Background()
	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Execute(ctx, func(ctx context.Context) error {
			return testErr
		})
	}

	if cb.GetState() != circuitbreaker.StateOpen {
		t.Error("Circuit should be open after failures")
	}

	// Wait for timeout to transition to half-open
	time.Sleep(150 * time.Millisecond)

	// First successful call in half-open
	err := cb.Execute(ctx, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected no error for first request, got %v", err)
	}

	// Check state is still half-open after one success
	if cb.GetState() != circuitbreaker.StateHalfOpen {
		t.Errorf("Circuit should still be half-open after one success, got %v", cb.GetState())
	}

	// Second successful call (still within MaxRequests)
	err = cb.Execute(ctx, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected no error for second request, got %v", err)
	}

	// Third call should be rejected (exceeds MaxRequests)
	// First check metrics to debug
	metrics := cb.GetMetrics()
	t.Logf("Before third call - State: %s, Successes: %d, HalfOpenReqs: %d",
		metrics.State, metrics.Successes, metrics.HalfOpenReqs)

	err = cb.Execute(ctx, func(ctx context.Context) error {
		return nil
	})
	if err != circuitbreaker.ErrTooManyRequests {
		metrics = cb.GetMetrics()
		t.Logf("After third call - State: %s, Successes: %d, HalfOpenReqs: %d",
			metrics.State, metrics.Successes, metrics.HalfOpenReqs)
		t.Errorf("Expected ErrTooManyRequests, got %v", err)
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := circuitbreaker.New(&circuitbreaker.Config{
		FailureThreshold: 2,
		SuccessThreshold: 2,
		Timeout:          1 * time.Second,
		MaxRequests:      2,
	})

	ctx := context.Background()
	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Execute(ctx, func(ctx context.Context) error {
			return testErr
		})
	}

	if cb.GetState() != circuitbreaker.StateOpen {
		t.Error("Circuit should be open")
	}

	// Reset the circuit
	cb.Reset()

	if cb.GetState() != circuitbreaker.StateClosed {
		t.Error("Circuit should be closed after reset")
	}

	// Should be able to execute calls again
	err := cb.Execute(ctx, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected no error after reset, got %v", err)
	}
}
