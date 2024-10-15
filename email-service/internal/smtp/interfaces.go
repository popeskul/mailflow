package smtp

import (
	"context"

	"github.com/popeskul/mailflow/email-service/internal/domain"
)

type EmailSender interface {
	Send(ctx context.Context, email *domain.Email) error
}
