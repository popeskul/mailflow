package retry

import (
	"context"
	"math"
	"math/rand"
	"time"
)

// Strategy defines the retry strategy interface
type Strategy interface {
	// NextDelay returns the delay before the next retry
	NextDelay(attempt int) time.Duration
	// ShouldRetry determines if we should retry based on the attempt number
	ShouldRetry(attempt int) bool
}

// ExponentialBackoff implements exponential backoff with jitter
type ExponentialBackoff struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	MaxAttempts  int
	Jitter       bool
}

// DefaultExponentialBackoff returns default exponential backoff configuration
func DefaultExponentialBackoff() *ExponentialBackoff {
	return &ExponentialBackoff{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		MaxAttempts:  5,
		Jitter:       true,
	}
}

// NextDelay calculates the next delay with exponential backoff
func (e *ExponentialBackoff) NextDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return e.InitialDelay
	}

	delay := float64(e.InitialDelay) * math.Pow(e.Multiplier, float64(attempt-1))
	if delay > float64(e.MaxDelay) {
		delay = float64(e.MaxDelay)
	}

	if e.Jitter {
		// Add jitter: random value between 0 and delay
		jitter := rand.Float64() * delay * 0.3 // 30% jitter
		delay = delay + jitter
	}

	return time.Duration(delay)
}

// ShouldRetry checks if we should retry based on attempt count
func (e *ExponentialBackoff) ShouldRetry(attempt int) bool {
	return attempt < e.MaxAttempts
}

// RetryableFunc is a function that can be retried
type RetryableFunc func(ctx context.Context) error

// RetryableError is an error that indicates if retry should be attempted
type RetryableError interface {
	error
	Retryable() bool
}

// Retrier handles retry logic
type Retrier struct {
	strategy Strategy
}

// New creates a new Retrier with the given strategy
func New(strategy Strategy) *Retrier {
	if strategy == nil {
		strategy = DefaultExponentialBackoff()
	}
	return &Retrier{strategy: strategy}
}

// Do executes the function with retry logic
func (r *Retrier) Do(ctx context.Context, fn RetryableFunc) error {
	var lastErr error

	for attempt := 0; r.strategy.ShouldRetry(attempt); attempt++ {
		if attempt > 0 {
			delay := r.strategy.NextDelay(attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		err := fn(ctx)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if retryable, ok := err.(RetryableError); ok && !retryable.Retryable() {
			return err
		}
	}

	return lastErr
}

// WithRetry is a helper function for simple retry logic
func WithRetry(ctx context.Context, fn RetryableFunc, opts ...Option) error {
	config := &Config{
		Strategy: DefaultExponentialBackoff(),
	}

	for _, opt := range opts {
		opt(config)
	}

	retrier := New(config.Strategy)
	return retrier.Do(ctx, fn)
}

// Config holds retry configuration
type Config struct {
	Strategy Strategy
}

// Option is a function that configures retry behavior
type Option func(*Config)

// WithStrategy sets the retry strategy
func WithStrategy(strategy Strategy) Option {
	return func(c *Config) {
		c.Strategy = strategy
	}
}

// WithMaxAttempts sets the maximum number of retry attempts
func WithMaxAttempts(attempts int) Option {
	return func(c *Config) {
		if eb, ok := c.Strategy.(*ExponentialBackoff); ok {
			eb.MaxAttempts = attempts
		}
	}
}

// WithInitialDelay sets the initial retry delay
func WithInitialDelay(delay time.Duration) Option {
	return func(c *Config) {
		if eb, ok := c.Strategy.(*ExponentialBackoff); ok {
			eb.InitialDelay = delay
		}
	}
}
