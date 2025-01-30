package memory

import (
	"context"
	"github.com/popeskul/email-service-platform/email-service/internal/domain"
	"time"
)

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
