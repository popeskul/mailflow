package services

import (
	"context"
	"github.com/popeskul/email-service-platform/email-service/internal/domain"
	"github.com/popeskul/ratelimiter"
	"time"
)

type EmailService interface {
	SendEmail(ctx context.Context, to, subject, body string) (*domain.Email, error)
	GetEmailStatus(ctx context.Context, id string) (*domain.Email, error)
	ListEmails(ctx context.Context, pageSize int, pageToken string) ([]*domain.Email, string, error)
	ResendFailedEmails(ctx context.Context) error
}

type Repositories interface {
	Email() EmailRepository
}

type EmailRepository interface {
	Save(ctx context.Context, email *domain.Email) error
	GetByID(ctx context.Context, id string) (*domain.Email, error)
	UpdateStatus(ctx context.Context, id, status string, sentAt *time.Time) error
	List(ctx context.Context, pageSize int, pageToken string) ([]*domain.Email, string, error)
	DeleteByID(ctx context.Context, id string) error
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
