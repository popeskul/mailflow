package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type EmailMetrics struct {
	*REDMetrics
	EmailsSent         prometheus.Counter
	EmailsQueued       prometheus.Counter
	EmailsFailed       prometheus.Counter
	RateLimitDelays    prometheus.Counter
	DowntimePeriods    prometheus.Counter
	QueueSize          prometheus.Gauge
	ProcessingDuration prometheus.Histogram
}

func NewEmailMetrics(serviceName string) *EmailMetrics {
	return &EmailMetrics{
		REDMetrics: NewREDMetrics(serviceName),
		EmailsSent: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: serviceName,
			Name:      "emails_sent_total",
			Help:      "The total number of successfully sent emails",
		}),
		EmailsQueued: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: serviceName,
			Name:      "emails_queued_total",
			Help:      "The total number of emails queued for sending",
		}),
		EmailsFailed: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: serviceName,
			Name:      "emails_failed_total",
			Help:      "The total number of failed email sends",
		}),
		RateLimitDelays: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: serviceName,
			Name:      "rate_limit_delays_total",
			Help:      "The total number of rate limit induced delays",
		}),
		DowntimePeriods: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: serviceName,
			Name:      "downtime_periods_total",
			Help:      "The total number of planned downtime periods",
		}),
		QueueSize: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: serviceName,
			Name:      "email_queue_size",
			Help:      "The current size of the email queue",
		}),
		ProcessingDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: serviceName,
			Name:      "email_processing_duration_seconds",
			Help:      "The duration of email processing in seconds",
			Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30},
		}),
	}
}

// RecordEmailSent increases the counter of sent letters
func (m *EmailMetrics) RecordEmailSent() {
	m.EmailsSent.Inc()
}

// RecordEmailQueued increases the counter of letters in the queue
func (m *EmailMetrics) RecordEmailQueued() {
	m.EmailsQueued.Inc()
}

// RecordEmailFailed increases the counter of unsuccessful sends
func (m *EmailMetrics) RecordEmailFailed() {
	m.EmailsFailed.Inc()
}

// RecordRateLimitDelay increases the rate limit delay counter
func (m *EmailMetrics) RecordRateLimitDelay() {
	m.RateLimitDelays.Inc()
}

// RecordDowntimePeriod increases the unavailability period counter
func (m *EmailMetrics) RecordDowntimePeriod() {
	m.DowntimePeriods.Inc()
}

// SetQueueSize sets the current queue size
func (m *EmailMetrics) SetQueueSize(size int) {
	m.QueueSize.Set(float64(size))
}

// ObserveProcessingDuration records the duration of email processing
func (m *EmailMetrics) ObserveProcessingDuration(duration float64) {
	m.ProcessingDuration.Observe(duration)
}
