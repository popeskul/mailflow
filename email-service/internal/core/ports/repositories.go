package ports

import (
	"context"
	"time"

	"github.com/popeskul/email-service-platform/email-service/internal/core/domain"
)

type EmailRepository interface {
	Save(ctx context.Context, email *domain.Email) error
	GetByID(ctx context.Context, id string) (*domain.Email, error)
	UpdateStatus(ctx context.Context, id, status string, sentAt *time.Time) error
	List(ctx context.Context, pageSize int, pageToken string) ([]*domain.Email, string, error)
	DeleteByID(ctx context.Context, id string) error
}
