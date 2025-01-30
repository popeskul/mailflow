package services

import (
	"context"

	"github.com/popeskul/email-service-platform/logger"
	"github.com/popeskul/email-service-platform/user-service/internal/domain"
)

// UserService is a service for managing users
type UserService interface {
}

type Repository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, pageSize int, pageToken string) ([]*domain.User, string, error)
}

// Logger is an abstraction for logging
type Logger interface {
	logger.Logger
}
