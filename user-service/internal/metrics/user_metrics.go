package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type UserMetrics struct {
	*REDMetrics
	UsersCreated    prometheus.Counter
	EmailSendErrors prometheus.Counter
}

// NewUserMetrics creates a new UserMetrics instance
func NewUserMetrics(serviceName string) *UserMetrics {
	return &UserMetrics{
		REDMetrics: NewREDMetrics(serviceName),
		UsersCreated: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: serviceName,
			Name:      "users_created_total",
			Help:      "The total number of created users",
		}),
		EmailSendErrors: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: serviceName,
			Name:      "email_send_errors_total",
			Help:      "The total number of email send errors",
		}),
	}
}

// RecordUserCreated increases the counter of created users
func (m *UserMetrics) RecordUserCreated() {
	m.UsersCreated.Inc()
}

// RecordEmailSendError increases the email sending error counter
func (m *UserMetrics) RecordEmailSendError() {
	m.EmailSendErrors.Inc()
}
