package grpc

import (
	"context"

	"github.com/popeskul/email-service-platform/logger"
	"github.com/popeskul/email-service-platform/user-service/internal/domain"
)

type UserService interface {
	Create(ctx context.Context, email, name string) (*domain.User, error)
	Get(ctx context.Context, id string) (*domain.User, error)
	Update(ctx context.Context, id, email, name string) (*domain.User, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, pageSize int, pageToken string) ([]*domain.User, string, error)
}

type Logger interface {
	logger.Logger
}
