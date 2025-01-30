package smtp

import (
	"context"
	"github.com/popeskul/email-service-platform/email-service/internal/domain"
)

type EmailSender interface {
	Send(ctx context.Context, email *domain.Email) error
}
