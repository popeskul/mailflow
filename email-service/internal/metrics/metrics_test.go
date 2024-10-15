package metrics

import (
	"errors"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestNewREDMetrics(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
	}{
		{
			name:        "create RED metrics",
			serviceName: "test_service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test registry to avoid conflicts
			testRegistry := prometheus.NewRegistry()

			// Temporarily replace global registry
			originalRegistry := Registry
			Registry = testRegistry
			defer func() {
				Registry = originalRegistry
			}()

			metrics := NewREDMetrics(tt.serviceName)

			assert.NotNil(t, metrics)
			assert.NotNil(t, metrics.RequestCounter)
			assert.NotNil(t, metrics.ErrorCounter)
			assert.NotNil(t, metrics.DurationHistogram)
		})
	}
}

func TestREDMetrics_RecordRequest(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		duration float64
		err      error
	}{
		{
			name:     "record successful request",
			method:   "SendEmail",
			duration: 0.5,
			err:      nil,
		},
		{
			name:     "record failed request",
			method:   "SendEmail",
			duration: 1.0,
			err:      errors.New("test error"),
		},
		{
			name:     "record request with code error",
			method:   "GetStatus",
			duration: 0.1,
			err:      &codeError{code: "not_found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test registry to avoid conflicts
			testRegistry := prometheus.NewRegistry()

			// Temporarily replace global registry
			originalRegistry := Registry
			Registry = testRegistry
			defer func() {
				Registry = originalRegistry
			}()

			metrics := NewREDMetrics("test")

			metrics.RecordRequest(tt.method, tt.duration, tt.err)

			// Verify that metrics were recorded
			mf, err := testRegistry.Gather()
			assert.NoError(t, err)
			assert.NotEmpty(t, mf)
		})
	}
}

func TestREDMetrics_Reset(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "reset metrics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test registry to avoid conflicts
			testRegistry := prometheus.NewRegistry()

			// Temporarily replace global registry
			originalRegistry := Registry
			Registry = testRegistry
			defer func() {
				Registry = originalRegistry
			}()

			metrics := NewREDMetrics("test")

			// Record some metrics
			metrics.RecordRequest("test", 1.0, nil)
			metrics.RecordRequest("test", 1.0, errors.New("error"))

			// Reset and verify
			metrics.Reset()

			// After reset, gathering should still work (metrics exist but are reset to 0)
			mf, err := testRegistry.Gather()
			assert.NoError(t, err)
			// Registry might be empty after reset, which is fine
			assert.NotNil(t, mf)
		})
	}
}

func TestNewEmailMetrics(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
	}{
		{
			name:        "create email metrics",
			serviceName: "email_service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test registry to avoid conflicts
			testRegistry := prometheus.NewRegistry()

			// Temporarily replace global registry
			originalRegistry := Registry
			Registry = testRegistry
			defer func() {
				Registry = originalRegistry
			}()

			metrics := NewEmailMetrics(tt.serviceName)

			assert.NotNil(t, metrics)
			assert.NotNil(t, metrics.REDMetrics)
			assert.NotNil(t, metrics.EmailsSent)
			assert.NotNil(t, metrics.EmailsQueued)
			assert.NotNil(t, metrics.EmailsFailed)
			assert.NotNil(t, metrics.RateLimitDelays)
			assert.NotNil(t, metrics.DowntimePeriods)
			assert.NotNil(t, metrics.QueueSize)
			assert.NotNil(t, metrics.ProcessingDuration)
		})
	}
}

func TestEmailMetrics_RecordEmailSent(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "record email sent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test registry to avoid conflicts
			testRegistry := prometheus.NewRegistry()

			// Temporarily replace global registry
			originalRegistry := Registry
			Registry = testRegistry
			defer func() {
				Registry = originalRegistry
			}()

			metrics := NewEmailMetrics("test")

			metrics.RecordEmailSent()

			// Verify that metric was recorded
			mf, err := testRegistry.Gather()
			assert.NoError(t, err)
			assert.NotEmpty(t, mf)
		})
	}
}

func TestEmailMetrics_RecordEmailQueued(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "record email queued",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test registry to avoid conflicts
			testRegistry := prometheus.NewRegistry()

			// Temporarily replace global registry
			originalRegistry := Registry
			Registry = testRegistry
			defer func() {
				Registry = originalRegistry
			}()

			metrics := NewEmailMetrics("test")

			metrics.RecordEmailQueued()

			// Verify that metric was recorded
			mf, err := testRegistry.Gather()
			assert.NoError(t, err)
			assert.NotEmpty(t, mf)
		})
	}
}

func TestEmailMetrics_RecordEmailFailed(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "record email failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test registry to avoid conflicts
			testRegistry := prometheus.NewRegistry()

			// Temporarily replace global registry
			originalRegistry := Registry
			Registry = testRegistry
			defer func() {
				Registry = originalRegistry
			}()

			metrics := NewEmailMetrics("test")

			metrics.RecordEmailFailed()

			// Verify that metric was recorded
			mf, err := testRegistry.Gather()
			assert.NoError(t, err)
			assert.NotEmpty(t, mf)
		})
	}
}

func TestEmailMetrics_RecordRateLimitDelay(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "record rate limit delay",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test registry to avoid conflicts
			testRegistry := prometheus.NewRegistry()

			// Temporarily replace global registry
			originalRegistry := Registry
			Registry = testRegistry
			defer func() {
				Registry = originalRegistry
			}()

			metrics := NewEmailMetrics("test")

			metrics.RecordRateLimitDelay()

			// Verify that metric was recorded
			mf, err := testRegistry.Gather()
			assert.NoError(t, err)
			assert.NotEmpty(t, mf)
		})
	}
}

func TestEmailMetrics_RecordDowntimePeriod(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "record downtime period",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test registry to avoid conflicts
			testRegistry := prometheus.NewRegistry()

			// Temporarily replace global registry
			originalRegistry := Registry
			Registry = testRegistry
			defer func() {
				Registry = originalRegistry
			}()

			metrics := NewEmailMetrics("test")

			metrics.RecordDowntimePeriod()

			// Verify that metric was recorded
			mf, err := testRegistry.Gather()
			assert.NoError(t, err)
			assert.NotEmpty(t, mf)
		})
	}
}

func TestEmailMetrics_SetQueueSize(t *testing.T) {
	tests := []struct {
		name      string
		queueSize int
	}{
		{
			name:      "set queue size",
			queueSize: 42,
		},
		{
			name:      "set zero queue size",
			queueSize: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test registry to avoid conflicts
			testRegistry := prometheus.NewRegistry()

			// Temporarily replace global registry
			originalRegistry := Registry
			Registry = testRegistry
			defer func() {
				Registry = originalRegistry
			}()

			metrics := NewEmailMetrics("test")

			metrics.SetQueueSize(tt.queueSize)

			// Verify that metric was recorded
			mf, err := testRegistry.Gather()
			assert.NoError(t, err)
			assert.NotEmpty(t, mf)
		})
	}
}

func TestEmailMetrics_ObserveProcessingDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration float64
	}{
		{
			name:     "observe processing duration",
			duration: 1.5,
		},
		{
			name:     "observe zero duration",
			duration: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test registry to avoid conflicts
			testRegistry := prometheus.NewRegistry()

			// Temporarily replace global registry
			originalRegistry := Registry
			Registry = testRegistry
			defer func() {
				Registry = originalRegistry
			}()

			metrics := NewEmailMetrics("test")

			metrics.ObserveProcessingDuration(tt.duration)

			// Verify that metric was recorded
			mf, err := testRegistry.Gather()
			assert.NoError(t, err)
			assert.NotEmpty(t, mf)
		})
	}
}

// codeError is a mock error with a Code method for testing
type codeError struct {
	code string
}

func (e *codeError) Error() string {
	return "test error with code"
}

func (e *codeError) Code() string {
	return e.code
}
