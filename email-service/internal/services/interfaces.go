package services

import (
	"context"
	"time"

	"github.com/popeskul/mailflow/email-service/internal/domain"
	"github.com/popeskul/ratelimiter"
)

type EmailService interface {
	SendEmail(ctx context.Context, to, subject, body string) (*domain.Email, error)
	GetEmailStatus(ctx context.Context, id string) (*domain.Email, error)
	ListEmails(ctx context.Context, pageSize int, pageToken string) ([]*domain.Email, string, error)
	ResendFailedEmails(ctx context.Context) error
}

type EmailRepository interface {
	Save(ctx context.Context, email *domain.Email) error
	GetByID(ctx context.Context, id string) (*domain.Email, error)
	UpdateStatus(ctx context.Context, id string, status string, sentAt *time.Time) error
	List(ctx context.Context, pageSize int, pageToken string) ([]*domain.Email, string, error)
}

type Repositories interface {
	Email() domain.EmailRepository
}

type EmailSender interface {
	Send(ctx context.Context, email *domain.Email) error
}

type Metrics interface {
	RecordEmailSent()
	RecordEmailQueued()
	RecordEmailFailed()
	RecordRateLimitDelay()
	RecordDowntimePeriod()
	SetQueueSize(size int)
	ObserveProcessingDuration(duration float64)
}

type Limiter interface {
	ratelimiter.Limiter
}
