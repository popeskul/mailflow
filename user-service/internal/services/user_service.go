package services

import (
	"context"
	"fmt"

	"github.com/popeskul/mailflow/common/logger"
	emailv1 "github.com/popeskul/mailflow/email-service/pkg/api/email/v1"
	"github.com/popeskul/mailflow/user-service/internal/domain"
)

type UserService struct {
	repo         domain.UserRepository
	emailClient  emailv1.EmailServiceClient
	emailWrapper *EmailClientWrapper
	logger       logger.Logger
}

func NewUserService(
	repo domain.UserRepository,
	emailClient emailv1.EmailServiceClient,
	l logger.Logger,
) *UserService {
	return &UserService{
		repo:        repo,
		emailClient: emailClient,
		logger:      l.Named("user_service"),
	}
}

// NewUserServiceWithWrapper creates user service with email wrapper
func NewUserServiceWithWrapper(
	repo domain.UserRepository,
	emailWrapper *EmailClientWrapper,
	l logger.Logger,
) *UserService {
	return &UserService{
		repo:         repo,
		emailWrapper: emailWrapper,
		logger:       l.Named("user_service"),
	}
}

func (s *UserService) Create(ctx context.Context, email, name string) (*domain.User, error) {
	l := s.logger.WithFields(logger.Fields{
		"email": email,
		"name":  name,
	})

	user := domain.NewUser(email, name)

	l.Info("creating new user",
		logger.Field{Key: "user_id", Value: user.ID},
	)

	if err := s.repo.Create(ctx, user); err != nil {
		l.Error("failed to create user",
			logger.Field{Key: "error", Value: err},
		)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	l.Info("sending welcome email")

	// Use wrapper if available, otherwise use direct client
	if s.emailWrapper != nil {
		err := s.emailWrapper.SendEmail(ctx, &emailv1.SendEmailRequest{
			To:      user.Email,
			Subject: "Welcome to our service!",
			Body:    fmt.Sprintf("Hello %s,\n\nWelcome to our service! We're glad to have you here.", user.Name),
		})
		if err != nil {
			l.Error("failed to send welcome email",
				logger.Field{Key: "error", Value: err},
			)
			// Don't return error as user is already created
		}
	} else if s.emailClient != nil {
		_, err := s.emailClient.SendEmail(ctx, &emailv1.SendEmailRequest{
			To:      user.Email,
			Subject: "Welcome to our service!",
			Body:    fmt.Sprintf("Hello %s,\n\nWelcome to our service! We're glad to have you here.", user.Name),
		})
		if err != nil {
			l.Error("failed to send welcome email",
				logger.Field{Key: "error", Value: err},
			)
			// Don't return error as user is already created
		}
	}

	return user, nil
}

func (s *UserService) Get(ctx context.Context, id string) (*domain.User, error) {
	l := s.logger.WithFields(logger.Fields{
		"user_id": id,
	})

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		l.Error("failed to get user",
			logger.Field{Key: "error", Value: err},
		)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (s *UserService) Update(ctx context.Context, id, email, name string) (*domain.User, error) {
	l := s.logger.WithFields(logger.Fields{
		"user_id": id,
		"email":   email,
		"name":    name,
	})

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		l.Error("failed to get user for update",
			logger.Field{Key: "error", Value: err},
		)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.Email = email
	user.Name = name

	if err := s.repo.Update(ctx, user); err != nil {
		l.Error("failed to update user",
			logger.Field{Key: "error", Value: err},
		)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	l := s.logger.WithFields(logger.Fields{
		"user_id": id,
	})

	if err := s.repo.Delete(ctx, id); err != nil {
		l.Error("failed to delete user",
			logger.Field{Key: "error", Value: err},
		)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (s *UserService) List(ctx context.Context, pageSize int, pageToken string) ([]*domain.User, string, error) {
	l := s.logger.WithFields(logger.Fields{
		"page_size":  pageSize,
		"page_token": pageToken,
	})

	users, nextToken, err := s.repo.List(ctx, pageSize, pageToken)
	if err != nil {
		l.Error("failed to list users",
			logger.Field{Key: "error", Value: err},
		)
		return nil, "", fmt.Errorf("failed to list users: %w", err)
	}

	return users, nextToken, nil
}
