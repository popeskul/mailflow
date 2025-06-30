package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/popeskul/mailflow/user-service/internal/circuitbreaker"
	"github.com/popeskul/mailflow/user-service/internal/queue"
)

// CircuitBreakerCollector collects circuit breaker metrics
type CircuitBreakerCollector struct {
	cb                *circuitbreaker.CircuitBreaker
	stateGauge        *prometheus.GaugeVec
	failuresGauge     prometheus.Gauge
	successesGauge    prometheus.Gauge
	halfOpenReqsGauge prometheus.Gauge
}

// NewCircuitBreakerCollector creates a new circuit breaker collector
func NewCircuitBreakerCollector(namespace string, cb *circuitbreaker.CircuitBreaker) *CircuitBreakerCollector {
	collector := &CircuitBreakerCollector{
		cb: cb,
		stateGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "circuit_breaker",
				Name:      "state",
				Help:      "Current state of circuit breaker (0=closed, 1=open, 2=half-open)",
			},
			[]string{"state"},
		),
		failuresGauge: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "circuit_breaker",
				Name:      "failures_total",
				Help:      "Total number of failures",
			},
		),
		successesGauge: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "circuit_breaker",
				Name:      "successes_total",
				Help:      "Total number of successes in half-open state",
			},
		),
		halfOpenReqsGauge: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "circuit_breaker",
				Name:      "half_open_requests",
				Help:      "Number of requests in half-open state",
			},
		),
	}

	// Register the collector with our custom registry
	Registry.MustRegister(collector)

	return collector
}

// Describe implements prometheus.Collector
func (c *CircuitBreakerCollector) Describe(ch chan<- *prometheus.Desc) {
	c.stateGauge.Describe(ch)
	ch <- c.failuresGauge.Desc()
	ch <- c.successesGauge.Desc()
	ch <- c.halfOpenReqsGauge.Desc()
}

// Collect implements prometheus.Collector
func (c *CircuitBreakerCollector) Collect(ch chan<- prometheus.Metric) {
	metrics := c.cb.GetMetrics()

	// Reset all state gauges
	c.stateGauge.Reset()

	// Set the current state
	switch metrics.State {
	case "closed":
		c.stateGauge.WithLabelValues("closed").Set(1)
		c.stateGauge.WithLabelValues("open").Set(0)
		c.stateGauge.WithLabelValues("half_open").Set(0)
	case "open":
		c.stateGauge.WithLabelValues("closed").Set(0)
		c.stateGauge.WithLabelValues("open").Set(1)
		c.stateGauge.WithLabelValues("half_open").Set(0)
	case "half-open":
		c.stateGauge.WithLabelValues("closed").Set(0)
		c.stateGauge.WithLabelValues("open").Set(0)
		c.stateGauge.WithLabelValues("half_open").Set(1)
	}

	c.failuresGauge.Set(float64(metrics.Failures))
	c.successesGauge.Set(float64(metrics.Successes))
	c.halfOpenReqsGauge.Set(float64(metrics.HalfOpenReqs))

	// Collect all metrics
	c.stateGauge.Collect(ch)
	ch <- c.failuresGauge
	ch <- c.successesGauge
	ch <- c.halfOpenReqsGauge
}

// QueueCollector collects queue metrics
type QueueCollector struct {
	queue           queue.Queue
	sizeGauge       prometheus.Gauge
	processingGauge prometheus.Gauge
	totalGauge      prometheus.Gauge
}

// NewQueueCollector creates a new queue collector
func NewQueueCollector(namespace string, q queue.Queue) *QueueCollector {
	collector := &QueueCollector{
		queue: q,
		sizeGauge: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "queue",
				Name:      "size",
				Help:      "Current number of messages in queue",
			},
		),
		processingGauge: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "queue",
				Name:      "processing",
				Help:      "Number of messages currently being processed",
			},
		),
		totalGauge: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "queue",
				Name:      "total",
				Help:      "Total number of messages (queued + processing)",
			},
		),
	}

	// Register the collector with our custom registry
	Registry.MustRegister(collector)

	return collector
}

// Describe implements prometheus.Collector
func (c *QueueCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.sizeGauge.Desc()
	ch <- c.processingGauge.Desc()
	ch <- c.totalGauge.Desc()
}

// Collect implements prometheus.Collector
func (c *QueueCollector) Collect(ch chan<- prometheus.Metric) {
	if eq, ok := c.queue.(*queue.EmailQueue); ok {
		size := eq.Size()
		c.sizeGauge.Set(float64(size))
		c.processingGauge.Set(0) // EmailQueue doesn't track processing separately
		c.totalGauge.Set(float64(size))
	}

	ch <- c.sizeGauge
	ch <- c.processingGauge
	ch <- c.totalGauge
}
