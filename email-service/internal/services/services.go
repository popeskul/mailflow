package services

import (
	"go.uber.org/zap"

	"github.com/popeskul/email-service-platform/email-service/internal/core/ports"
	"github.com/popeskul/email-service-platform/email-service/internal/metrics"
	"github.com/popeskul/ratelimiter"
)

type Services struct {
	EmailService ports.EmailService
}

type Repositories interface {
	Email() ports.EmailRepository
}

func NewServices(
	repos Repositories,
	emailSender ports.EmailSender,
	limiter ratelimiter.Limiter,
	metrics *metrics.EmailMetrics,
	logger *zap.Logger,
) *Services {
	return &Services{
		EmailService: NewEmailService(repos.Email(), emailSender, limiter, metrics, logger),
	}
}
