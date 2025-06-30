package circuitbreaker

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// State represents the state of the circuit breaker
type State int32

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

var (
	ErrCircuitOpen     = errors.New("circuit breaker is open")
	ErrTooManyRequests = errors.New("too many requests in half-open state")
)

// Config holds the configuration for the circuit breaker
type Config struct {
	// FailureThreshold is the number of failures before opening the circuit
	FailureThreshold int
	// SuccessThreshold is the number of successes in half-open state before closing the circuit
	SuccessThreshold int
	// Timeout is the duration the circuit stays open before switching to half-open
	Timeout time.Duration
	// MaxRequests is the maximum number of requests allowed in half-open state
	MaxRequests int
}

// DefaultConfig returns default circuit breaker configuration
func DefaultConfig() *Config {
	return &Config{
		FailureThreshold: 5,
		SuccessThreshold: 2,
		Timeout:          30 * time.Second,
		MaxRequests:      3,
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	config *Config
	state  atomic.Value // State
	mu     sync.Mutex

	failures        int
	successes       int
	lastFailureTime time.Time
	halfOpenReqs    int
}

// New creates a new circuit breaker
func New(config *Config) *CircuitBreaker {
	if config == nil {
		config = DefaultConfig()
	}

	cb := &CircuitBreaker{
		config: config,
	}
	cb.state.Store(StateClosed)
	return cb
}

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(context.Context) error) error {
	if err := cb.canExecute(); err != nil {
		return err
	}

	err := fn(ctx)
	cb.recordResult(err)
	return err
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() State {
	return cb.state.Load().(State)
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state.Store(StateClosed)
	cb.failures = 0
	cb.successes = 0
	cb.halfOpenReqs = 0
}

func (cb *CircuitBreaker) canExecute() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	state := cb.state.Load().(State)

	switch state {
	case StateClosed:
		return nil

	case StateOpen:
		if time.Since(cb.lastFailureTime) > cb.config.Timeout {
			cb.state.Store(StateHalfOpen)
			cb.halfOpenReqs = 1 // Count this request
			cb.successes = 0
			return nil
		}
		return ErrCircuitOpen

	case StateHalfOpen:
		if cb.halfOpenReqs >= cb.config.MaxRequests {
			return ErrTooManyRequests
		}
		cb.halfOpenReqs++
		return nil

	default:
		return nil
	}
}

func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	state := cb.state.Load().(State)

	switch state {
	case StateClosed:
		if err != nil {
			cb.failures++
			if cb.failures >= cb.config.FailureThreshold {
				cb.state.Store(StateOpen)
				cb.lastFailureTime = time.Now()
			}
		} else {
			cb.failures = 0
		}

	case StateHalfOpen:
		if err != nil {
			cb.state.Store(StateOpen)
			cb.lastFailureTime = time.Now()
			cb.failures = cb.config.FailureThreshold
		} else {
			cb.successes++
			if cb.successes >= cb.config.SuccessThreshold {
				cb.state.Store(StateClosed)
				cb.failures = 0
			}
		}
	}
}

// Metrics represents circuit breaker metrics
type Metrics struct {
	State           string
	Failures        int
	Successes       int
	LastFailureTime time.Time
	HalfOpenReqs    int
}

// GetMetrics returns current circuit breaker metrics
func (cb *CircuitBreaker) GetMetrics() Metrics {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	state := cb.state.Load().(State)
	var stateStr string
	switch state {
	case StateClosed:
		stateStr = "closed"
	case StateOpen:
		stateStr = "open"
	case StateHalfOpen:
		stateStr = "half-open"
	}

	return Metrics{
		State:           stateStr,
		Failures:        cb.failures,
		Successes:       cb.successes,
		LastFailureTime: cb.lastFailureTime,
		HalfOpenReqs:    cb.halfOpenReqs,
	}
}
