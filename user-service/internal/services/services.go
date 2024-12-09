package services

import (
	"go.uber.org/zap"

	emailv1 "github.com/popeskul/email-service-platform/email-service/pkg/api/email/v1"
	"github.com/popeskul/email-service-platform/user-service/internal/ports"
)

type Services struct {
	User ports.UserService
}

type Repositories interface {
	UserRepository() ports.UserRepository
}

func NewServices(
	repos Repositories,
	emailClient emailv1.EmailServiceClient,
	logger *zap.Logger,
) *Services {
	return &Services{
		User: NewUserService(repos.UserRepository(), emailClient, logger),
	}
}

func (s Services) UserService() ports.UserService {
	return s.User
}
