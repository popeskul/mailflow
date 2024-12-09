package services

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	emailv1 "github.com/popeskul/email-service-platform/email-service/pkg/api/email/v1"
	"github.com/popeskul/email-service-platform/user-service/internal/domain"
	"github.com/popeskul/email-service-platform/user-service/internal/ports"
)

type userService struct {
	repo        ports.UserRepository
	emailClient emailv1.EmailServiceClient
	logger      *zap.Logger
}

func NewUserService(
	repo ports.UserRepository,
	emailClient emailv1.EmailServiceClient,
	logger *zap.Logger,
) ports.UserService {
	return &userService{
		repo:        repo,
		emailClient: emailClient,
		logger:      logger.Named("user_service"),
	}
}

func (s *userService) Create(ctx context.Context, email, name string) (*domain.User, error) {
	logger := s.logger.With(
		zap.String("email", email),
		zap.String("name", name),
	)

	user := domain.NewUser(email, name)

	logger.Info("creating new user", zap.String("user_id", user.ID))

	if err := s.repo.Create(ctx, user); err != nil {
		logger.Error("failed to create user", zap.Error(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	logger.Info("sending welcome email")
	_, err := s.emailClient.SendEmail(ctx, &emailv1.SendEmailRequest{
		To:      user.Email,
		Subject: "Welcome to our service!",
		Body:    fmt.Sprintf("Hello %s,\n\nWelcome to our service! We're glad to have you here.", user.Name),
	})
	if err != nil {
		logger.Error("failed to send welcome email", zap.Error(err))
		// Не возвращаем ошибку, так как пользователь уже создан
	}

	return user, nil
}

func (s *userService) Get(ctx context.Context, id string) (*domain.User, error) {
	logger := s.logger.With(zap.String("user_id", id))

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.Error("failed to get user", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (s *userService) Update(ctx context.Context, id, email, name string) (*domain.User, error) {
	logger := s.logger.With(
		zap.String("user_id", id),
		zap.String("email", email),
		zap.String("name", name),
	)

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.Error("failed to get user for update", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.Email = email
	user.Name = name

	if err := s.repo.Update(ctx, user); err != nil {
		logger.Error("failed to update user", zap.Error(err))
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

func (s *userService) Delete(ctx context.Context, id string) error {
	logger := s.logger.With(zap.String("user_id", id))

	if err := s.repo.Delete(ctx, id); err != nil {
		logger.Error("failed to delete user", zap.Error(err))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (s *userService) List(ctx context.Context, pageSize int, pageToken string) ([]*domain.User, string, error) {
	logger := s.logger.With(
		zap.Int("page_size", pageSize),
		zap.String("page_token", pageToken),
	)

	users, nextToken, err := s.repo.List(ctx, pageSize, pageToken)
	if err != nil {
		logger.Error("failed to list users", zap.Error(err))
		return nil, "", fmt.Errorf("failed to list users: %w", err)
	}

	return users, nextToken, nil
}
