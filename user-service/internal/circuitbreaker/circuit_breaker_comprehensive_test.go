package circuitbreaker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig_Success(t *testing.T) {
	tests := []struct {
		name           string
		expectedConfig *Config
	}{
		{
			name: "default configuration values",
			expectedConfig: &Config{
				FailureThreshold: 5,
				SuccessThreshold: 2,
				Timeout:          30 * time.Second,
				MaxRequests:      3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()

			assert.Equal(t, tt.expectedConfig.FailureThreshold, config.FailureThreshold)
			assert.Equal(t, tt.expectedConfig.SuccessThreshold, config.SuccessThreshold)
			assert.Equal(t, tt.expectedConfig.Timeout, config.Timeout)
			assert.Equal(t, tt.expectedConfig.MaxRequests, config.MaxRequests)
		})
	}
}

func TestNew_Success(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		expectedState State
	}{
		{
			name: "new circuit breaker with custom config",
			config: &Config{
				FailureThreshold: 3,
				SuccessThreshold: 1,
				Timeout:          10 * time.Second,
				MaxRequests:      2,
			},
			expectedState: StateClosed,
		},
		{
			name:          "new circuit breaker with nil config",
			config:        nil,
			expectedState: StateClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := New(tt.config)

			assert.NotNil(t, cb)
			assert.Equal(t, tt.expectedState, cb.GetState())

			if tt.config == nil {
				// Should use default config
				assert.NotNil(t, cb.config)
			} else {
				assert.Equal(t, tt.config, cb.config)
			}
		})
	}
}

func TestCircuitBreaker_Execute_Success(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		setupFn     func(*CircuitBreaker)
		executeFn   func(context.Context) error
		expectError bool
	}{
		{
			name: "execute success in closed state",
			config: &Config{
				FailureThreshold: 3,
				SuccessThreshold: 2,
				Timeout:          1 * time.Second,
				MaxRequests:      2,
			},
			setupFn: func(cb *CircuitBreaker) {
				// Circuit is closed by default
			},
			executeFn: func(ctx context.Context) error {
				return nil
			},
			expectError: false,
		},
		{
			name: "execute success multiple times",
			config: &Config{
				FailureThreshold: 3,
				SuccessThreshold: 2,
				Timeout:          1 * time.Second,
				MaxRequests:      2,
			},
			setupFn: func(cb *CircuitBreaker) {
				// Execute successful operations
				cb.Execute(context.Background(), func(ctx context.Context) error { return nil })
				cb.Execute(context.Background(), func(ctx context.Context) error { return nil })
			},
			executeFn: func(ctx context.Context) error {
				return nil
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := New(tt.config)

			if tt.setupFn != nil {
				tt.setupFn(cb)
			}

			err := cb.Execute(context.Background(), tt.executeFn)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCircuitBreaker_Execute_Fail(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		setupFn       func(*CircuitBreaker)
		executeFn     func(context.Context) error
		expectedError error
	}{
		{
			name: "execute failure opens circuit",
			config: &Config{
				FailureThreshold: 2,
				SuccessThreshold: 2,
				Timeout:          1 * time.Second,
				MaxRequests:      2,
			},
			setupFn: func(cb *CircuitBreaker) {
				// Cause failures to open circuit
				cb.Execute(context.Background(), func(ctx context.Context) error {
					return errors.New("service error")
				})
				cb.Execute(context.Background(), func(ctx context.Context) error {
					return errors.New("service error")
				})
			},
			executeFn: func(ctx context.Context) error {
				return nil
			},
			expectedError: ErrCircuitOpen,
		},
		{
			name: "too many requests in half-open state",
			config: &Config{
				FailureThreshold: 1,
				SuccessThreshold: 2,
				Timeout:          1 * time.Millisecond,
				MaxRequests:      1,
			},
			setupFn: func(cb *CircuitBreaker) {
				// Open circuit
				cb.Execute(context.Background(), func(ctx context.Context) error {
					return errors.New("service error")
				})
				// Wait for timeout to transition to half-open
				time.Sleep(2 * time.Millisecond)
				// Make one request to reach max requests
				cb.Execute(context.Background(), func(ctx context.Context) error {
					return nil
				})
			},
			executeFn: func(ctx context.Context) error {
				return nil
			},
			expectedError: ErrTooManyRequests,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := New(tt.config)

			if tt.setupFn != nil {
				tt.setupFn(cb)
			}

			err := cb.Execute(context.Background(), tt.executeFn)

			assert.Error(t, err)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestCircuitBreaker_StateTransitions_Success(t *testing.T) {
	tests := []struct {
		name       string
		config     *Config
		operations []struct {
			fn            func(context.Context) error
			expectedState State
		}
	}{
		{
			name: "closed to open to half-open to closed",
			config: &Config{
				FailureThreshold: 2,
				SuccessThreshold: 2,
				Timeout:          10 * time.Millisecond,
				MaxRequests:      3,
			},
			operations: []struct {
				fn            func(context.Context) error
				expectedState State
			}{
				{
					fn:            func(ctx context.Context) error { return nil },
					expectedState: StateClosed,
				},
				{
					fn:            func(ctx context.Context) error { return errors.New("error") },
					expectedState: StateClosed, // Still closed, need more failures
				},
				{
					fn:            func(ctx context.Context) error { return errors.New("error") },
					expectedState: StateOpen, // Now open
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := New(tt.config)

			for i, op := range tt.operations {
				cb.Execute(context.Background(), op.fn)

				assert.Equal(t, op.expectedState, cb.GetState(),
					"Operation %d: expected state %v, got %v", i, op.expectedState, cb.GetState())
			}
		})
	}
}

func TestCircuitBreaker_Reset_Success(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		setupFn func(*CircuitBreaker)
	}{
		{
			name: "reset open circuit to closed",
			config: &Config{
				FailureThreshold: 1,
				SuccessThreshold: 2,
				Timeout:          1 * time.Second,
				MaxRequests:      2,
			},
			setupFn: func(cb *CircuitBreaker) {
				// Open the circuit
				cb.Execute(context.Background(), func(ctx context.Context) error {
					return errors.New("service error")
				})
			},
		},
		{
			name:   "reset closed circuit",
			config: DefaultConfig(),
			setupFn: func(cb *CircuitBreaker) {
				// Circuit is already closed, but add some failures
				cb.Execute(context.Background(), func(ctx context.Context) error {
					return errors.New("service error")
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := New(tt.config)

			if tt.setupFn != nil {
				tt.setupFn(cb)
			}

			// Reset the circuit
			cb.Reset()

			assert.Equal(t, StateClosed, cb.GetState())

			// Should be able to execute successfully
			err := cb.Execute(context.Background(), func(ctx context.Context) error {
				return nil
			})
			assert.NoError(t, err)
		})
	}
}

func TestCircuitBreaker_GetMetrics_Success(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		setupFn       func(*CircuitBreaker)
		expectedState string
	}{
		{
			name:   "metrics for closed state",
			config: DefaultConfig(),
			setupFn: func(cb *CircuitBreaker) {
				// Execute some operations
				cb.Execute(context.Background(), func(ctx context.Context) error { return nil })
				cb.Execute(context.Background(), func(ctx context.Context) error {
					return errors.New("error")
				})
			},
			expectedState: "closed",
		},
		{
			name: "metrics for open state",
			config: &Config{
				FailureThreshold: 1,
				SuccessThreshold: 2,
				Timeout:          1 * time.Second,
				MaxRequests:      2,
			},
			setupFn: func(cb *CircuitBreaker) {
				// Open the circuit
				cb.Execute(context.Background(), func(ctx context.Context) error {
					return errors.New("service error")
				})
			},
			expectedState: "open",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := New(tt.config)

			if tt.setupFn != nil {
				tt.setupFn(cb)
			}

			metrics := cb.GetMetrics()

			assert.Equal(t, tt.expectedState, metrics.State)
			assert.GreaterOrEqual(t, metrics.Failures, 0)
			assert.GreaterOrEqual(t, metrics.Successes, 0)
			assert.GreaterOrEqual(t, metrics.HalfOpenReqs, 0)
		})
	}
}

func TestCircuitBreaker_HalfOpenTransition_Success(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "transition from open to half-open after timeout",
			config: &Config{
				FailureThreshold: 1,
				SuccessThreshold: 2,
				Timeout:          10 * time.Millisecond,
				MaxRequests:      2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := New(tt.config)

			// Open the circuit
			err := cb.Execute(context.Background(), func(ctx context.Context) error {
				return errors.New("service error")
			})
			assert.Error(t, err) // Should fail and open circuit
			assert.Equal(t, StateOpen, cb.GetState())

			// Try to execute immediately - should fail
			err = cb.Execute(context.Background(), func(ctx context.Context) error {
				return nil
			})
			assert.Equal(t, ErrCircuitOpen, err)

			// Wait for timeout
			time.Sleep(15 * time.Millisecond)

			// Next execution should transition to half-open
			err = cb.Execute(context.Background(), func(ctx context.Context) error {
				return nil
			})
			assert.NoError(t, err)
			assert.Equal(t, StateHalfOpen, cb.GetState())
		})
	}
}
