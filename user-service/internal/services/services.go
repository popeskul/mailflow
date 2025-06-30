package services

import (
	"github.com/popeskul/mailflow/common/logger"
	emailv1 "github.com/popeskul/mailflow/email-service/pkg/api/email/v1"
	"github.com/popeskul/mailflow/user-service/internal/domain"
)

type Services struct {
	user *UserService
}

func NewServices(
	repos Repositories,
	emailClient emailv1.EmailServiceClient,
	logger logger.Logger,
) *Services {
	return &Services{
		user: NewUserService(repos.User(), emailClient, logger),
	}
}

// NewServicesWithWrapper creates services with email client wrapper
func NewServicesWithWrapper(
	repos Repositories,
	emailWrapper *EmailClientWrapper,
	logger logger.Logger,
) *Services {
	return &Services{
		user: NewUserServiceWithWrapper(repos.User(), emailWrapper, logger),
	}
}

func (s Services) User() domain.UserService {
	return s.user
}
