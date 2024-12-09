package ports

import (
	"context"

	"github.com/popeskul/email-service-platform/email-service/internal/core/domain"
)

type EmailService interface {
	SendEmail(ctx context.Context, to, subject, body string) (*domain.Email, error)
	GetEmailStatus(ctx context.Context, id string) (*domain.Email, error)
	ListEmails(ctx context.Context, pageSize int, pageToken string) ([]*domain.Email, string, error)
	ResendFailedEmails(ctx context.Context) error
}

type EmailSender interface {
	Send(ctx context.Context, email *domain.Email) error
}
