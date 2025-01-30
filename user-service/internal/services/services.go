package services

import (
	emailv1 "github.com/popeskul/email-service-platform/email-service/pkg/api/email/v1"
)

type Services struct {
	user *UserService
}

type Repositories interface {
	UserRepository() Repository
}

func NewServices(
	repos Repositories,
	emailClient emailv1.EmailServiceClient,
	logger Logger,
) *Services {
	return &Services{
		user: NewUserService(repos.UserRepository(), emailClient, logger),
	}
}

func (s Services) User() *UserService {
	return s.user
}
