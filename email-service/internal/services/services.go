package services

import (
	"github.com/popeskul/mailflow/common/logger"
	"github.com/popeskul/mailflow/email-service/internal/metrics"
)

type ServiceContainer struct {
	email EmailService
}

func NewServices(
	repos Repositories,
	emailSender EmailSender,
	limiter Limiter,
	metrics *metrics.EmailMetrics,
	logger logger.Logger,
) *ServiceContainer {
	return &ServiceContainer{
		email: NewEmailService(repos.Email(), emailSender, limiter, metrics, logger),
	}
}

func (s *ServiceContainer) Email() EmailService {
	return s.email
}
