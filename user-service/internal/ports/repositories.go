package ports

import (
	"context"
	"github.com/popeskul/email-service-platform/user-service/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, pageSize int, pageToken string) ([]*domain.User, string, error)
}
