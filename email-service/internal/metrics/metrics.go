package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type REDMetrics struct {
	RequestCounter    *prometheus.CounterVec
	ErrorCounter      *prometheus.CounterVec
	DurationHistogram *prometheus.HistogramVec
}

func NewREDMetrics(serviceName string) *REDMetrics {
	return &REDMetrics{
		RequestCounter: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: serviceName,
				Name:      "requests_total",
				Help:      "The total number of processed requests",
			},
			[]string{"method"},
		),
		ErrorCounter: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: serviceName,
				Name:      "errors_total",
				Help:      "The total number of errors",
			},
			[]string{"method", "code"},
		),
		DurationHistogram: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: serviceName,
				Name:      "request_duration_seconds",
				Help:      "The duration of requests in seconds",
				Buckets:   []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5},
			},
			[]string{"method"},
		),
	}
}

// RecordRequest records request metrics
func (m *REDMetrics) RecordRequest(method string, duration float64, err error) {
	m.RequestCounter.WithLabelValues(method).Inc()
	m.DurationHistogram.WithLabelValues(method).Observe(duration)
	if err != nil {
		code := "internal_error"
		if statusErr, ok := err.(interface{ Code() string }); ok {
			code = statusErr.Code()
		}
		m.ErrorCounter.WithLabelValues(method, code).Inc()
	}
}

// Reset resets all metrics
func (m *REDMetrics) Reset() {
	m.RequestCounter.Reset()
	m.ErrorCounter.Reset()
	m.DurationHistogram.Reset()
}
