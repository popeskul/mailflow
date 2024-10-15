package metrics

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/popeskul/mailflow/user-service/internal/circuitbreaker"
	"github.com/popeskul/mailflow/user-service/internal/domain"
	"github.com/popeskul/mailflow/user-service/internal/queue"
)

func TestRegistry_Init(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "registry should be initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, Registry)

			// Test that we can gather metrics
			metrics, err := Registry.Gather()
			assert.NoError(t, err)
			assert.NotEmpty(t, metrics)
		})
	}
}

func TestNewCircuitBreakerCollector(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
	}{
		{
			name:      "create circuit breaker collector",
			namespace: "test_service",
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

			cb := circuitbreaker.New(nil)
			collector := NewCircuitBreakerCollector(tt.namespace, cb)

			assert.NotNil(t, collector)
			assert.Equal(t, cb, collector.cb)
			assert.NotNil(t, collector.stateGauge)
			assert.NotNil(t, collector.failuresGauge)
			assert.NotNil(t, collector.successesGauge)
			assert.NotNil(t, collector.halfOpenReqsGauge)
		})
	}
}

func TestCircuitBreakerCollector_Describe(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "describe circuit breaker metrics",
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

			cb := circuitbreaker.New(nil)
			collector := NewCircuitBreakerCollector("test", cb)

			ch := make(chan *prometheus.Desc, 10)
			collector.Describe(ch)
			close(ch)

			descs := make([]*prometheus.Desc, 0)
			for desc := range ch {
				descs = append(descs, desc)
			}

			assert.Equal(t, 4, len(descs)) // state gauge (1) + failures + successes + half_open_reqs
		})
	}
}

func TestCircuitBreakerCollector_Collect(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "collect circuit breaker metrics",
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

			cb := circuitbreaker.New(nil)
			collector := NewCircuitBreakerCollector("test", cb)

			ch := make(chan prometheus.Metric, 10)
			collector.Collect(ch)
			close(ch)

			metrics := make([]prometheus.Metric, 0)
			for metric := range ch {
				metrics = append(metrics, metric)
			}

			// Should have metrics for state (3) + failures + successes + half_open_reqs
			assert.Equal(t, 6, len(metrics))
		})
	}
}

func TestNewQueueCollector(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
	}{
		{
			name:      "create queue collector",
			namespace: "test_service",
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

			q := queue.NewEmailQueue(100, zap.NewNop())
			collector := NewQueueCollector(tt.namespace, q)

			assert.NotNil(t, collector)
			assert.Equal(t, q, collector.queue)
			assert.NotNil(t, collector.sizeGauge)
			assert.NotNil(t, collector.processingGauge)
			assert.NotNil(t, collector.totalGauge)
		})
	}
}

func TestQueueCollector_Describe(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "describe queue metrics",
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

			q := queue.NewEmailQueue(100, zap.NewNop())
			collector := NewQueueCollector("test", q)

			ch := make(chan *prometheus.Desc, 10)
			collector.Describe(ch)
			close(ch)

			descs := make([]*prometheus.Desc, 0)
			for desc := range ch {
				descs = append(descs, desc)
			}

			assert.Equal(t, 3, len(descs)) // size + processing + total
		})
	}
}

func TestQueueCollector_Collect(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "collect queue metrics",
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

			q := queue.NewEmailQueue(100, zap.NewNop())
			collector := NewQueueCollector("test", q)

			// Add some test data to queue
			email := &domain.Email{
				ID:      "test-1",
				To:      "test@example.com",
				Subject: "Test Subject",
				Body:    "Test Body",
			}
			err := q.Enqueue(email)
			require.NoError(t, err)

			ch := make(chan prometheus.Metric, 10)
			collector.Collect(ch)
			close(ch)

			metrics := make([]prometheus.Metric, 0)
			for metric := range ch {
				metrics = append(metrics, metric)
			}

			assert.Equal(t, 3, len(metrics)) // size + processing + total
		})
	}
}

func TestQueueCollector_Collect_NonMemoryQueue(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "collect metrics from non-memory queue",
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

			// Create a mock queue that's not a memory queue
			mockQueue := &mockQueue{}
			collector := NewQueueCollector("test", mockQueue)

			ch := make(chan prometheus.Metric, 10)
			collector.Collect(ch)
			close(ch)

			metrics := make([]prometheus.Metric, 0)
			for metric := range ch {
				metrics = append(metrics, metric)
			}

			// Should still have 3 metrics but with default values (0)
			assert.Equal(t, 3, len(metrics))
		})
	}
}

// mockQueue implements queue.Queue interface for testing
type mockQueue struct{}

func (m *mockQueue) Enqueue(email *domain.Email) error                              { return nil }
func (m *mockQueue) Start(ctx context.Context, processor func(*domain.Email) error) {}
func (m *mockQueue) Stop()                                                          {}
func (m *mockQueue) Size() int                                                      { return 0 }
