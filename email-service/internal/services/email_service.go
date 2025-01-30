package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/popeskul/email-service-platform/email-service/internal/domain"
	"github.com/popeskul/email-service-platform/logger"
)

type emailService struct {
	repo        EmailRepository
	sender      EmailSender
	rateLimiter Limiter
	metrics     Metrics
	retryQueue  chan *domain.Email
	logger      logger.Logger
	mu          sync.Mutex
}

func NewEmailService(
	repo EmailRepository,
	sender EmailSender,
	limiter Limiter,
	metrics Metrics,
	l logger.Logger,
) EmailService {
	svc := &emailService{
		repo:        repo,
		sender:      sender,
		rateLimiter: limiter,
		metrics:     metrics,
		retryQueue:  make(chan *domain.Email, 1000),
		logger:      l.Named("email_service"),
	}

	go svc.processRetryQueue()

	return svc
}

func (s *emailService) SendEmail(ctx context.Context, to, subject, body string) (*domain.Email, error) {
	l := s.logger.WithFields(logger.Fields{
		"to":      to,
		"subject": subject,
	})

	email := domain.NewEmail(to, subject, body)

	l.Info("attempting to save email",
		logger.Field{Key: "email_id", Value: email.ID},
	)
	if err := s.repo.Save(ctx, email); err != nil {
		l.Error("failed to save email",
			logger.Field{Key: "error", Value: err},
		)
		return nil, fmt.Errorf("failed to save email: %w", err)
	}

	l.Info("attempting to send email")
	if err := s.rateLimiter.Wait(ctx); err != nil {
		l.Warn("rate limit exceeded, queueing email for retry",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "email_id", Value: email.ID},
		)
		s.metrics.RecordRateLimitDelay()
		s.queueForRetry(email)
		return email, nil
	}

	if err := s.sender.Send(ctx, email); err != nil {
		l.Error("failed to send email",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "email_id", Value: email.ID},
		)
		s.metrics.RecordEmailFailed()
		s.queueForRetry(email)
		return email, nil
	}

	now := time.Now()
	email.Status = domain.StatusSent
	email.SentAt = &now

	l.Info("email sent successfully",
		logger.Field{Key: "email_id", Value: email.ID},
	)
	s.metrics.RecordEmailSent()

	if err := s.repo.UpdateStatus(ctx, email.ID, email.Status, email.SentAt); err != nil {
		l.Error("failed to update email status",
			logger.Field{Key: "error", Value: err},
		)
		return email, fmt.Errorf("failed to update email status: %w", err)
	}

	return email, nil
}

func (s *emailService) GetEmailStatus(ctx context.Context, id string) (*domain.Email, error) {
	l := s.logger.WithFields(logger.Fields{
		"email_id": id,
	})

	email, err := s.repo.GetByID(ctx, id)
	if err != nil {
		l.Error("failed to get email status",
			logger.Field{Key: "error", Value: err},
		)
		return nil, fmt.Errorf("failed to get email status: %w", err)
	}

	return email, nil
}

func (s *emailService) ListEmails(ctx context.Context, pageSize int, pageToken string) ([]*domain.Email, string, error) {
	l := s.logger.WithFields(logger.Fields{
		"page_size":  pageSize,
		"page_token": pageToken,
	})

	emails, nextToken, err := s.repo.List(ctx, pageSize, pageToken)
	if err != nil {
		l.Error("failed to list emails",
			logger.Field{Key: "error", Value: err},
		)
		return nil, "", fmt.Errorf("failed to list emails: %w", err)
	}

	return emails, nextToken, nil
}

func (s *emailService) ResendFailedEmails(ctx context.Context) error {
	l := s.logger.WithFields(logger.Fields{
		"operation": "resend_failed",
	})

	l.Info("starting resend of failed emails")

	emails, _, err := s.repo.List(ctx, 0, "")
	if err != nil {
		l.Error("failed to list emails",
			logger.Field{Key: "error", Value: err},
		)
		return fmt.Errorf("failed to list emails: %w", err)
	}

	var resendCount int
	for _, email := range emails {
		if email.Status == domain.StatusFailed {
			l.Info("requeueing failed email",
				logger.Field{Key: "email_id", Value: email.ID},
				logger.Field{Key: "to", Value: email.To},
			)
			s.queueForRetry(email)
			resendCount++
		}
	}

	l.Info("finished requeueing failed emails",
		logger.Field{Key: "resend_count", Value: resendCount},
	)
	return nil
}

func (s *emailService) queueForRetry(email *domain.Email) {
	l := s.logger.WithFields(logger.Fields{
		"email_id": email.ID,
		"status":   email.Status,
	})

	email.Status = domain.StatusPending
	s.metrics.RecordEmailQueued()

	select {
	case s.retryQueue <- email:
		l.Info("email successfully queued for retry")

		if err := s.repo.UpdateStatus(context.Background(), email.ID, email.Status, nil); err != nil {
			l.Error("failed to update email status after queuing",
				logger.Field{Key: "error", Value: err},
			)
		}

	default:
		l.Warn("retry queue is full, marking email as failed")

		email.Status = domain.StatusFailed
		if err := s.repo.UpdateStatus(context.Background(), email.ID, email.Status, nil); err != nil {
			l.Error("failed to update email status when queue full",
				logger.Field{Key: "error", Value: err},
			)
		}

		s.metrics.RecordEmailFailed()
		s.metrics.SetQueueSize(len(s.retryQueue))
	}
}

func (s *emailService) processRetryQueue() {
	l := s.logger.Named("retry_queue")

	for email := range s.retryQueue {
		ctx := context.Background()
		emailLogger := l.WithFields(logger.Fields{
			"email_id": email.ID,
		})

		emailLogger.Info("processing queued email")

		if err := s.rateLimiter.Wait(ctx); err != nil {
			emailLogger.Warn("rate limit still exceeded, requeueing email",
				logger.Field{Key: "error", Value: err},
			)
			s.queueForRetry(email)
			continue
		}

		if err := s.sender.Send(ctx, email); err != nil {
			emailLogger.Error("failed to send queued email",
				logger.Field{Key: "error", Value: err},
			)
			s.queueForRetry(email)
			continue
		}

		now := time.Now()
		email.Status = domain.StatusSent
		email.SentAt = &now

		emailLogger.Info("queued email sent successfully")
		s.metrics.RecordEmailSent()

		if err := s.repo.UpdateStatus(ctx, email.ID, email.Status, email.SentAt); err != nil {
			emailLogger.Error("failed to update queued email status",
				logger.Field{Key: "error", Value: err},
			)
			s.queueForRetry(email)
			continue
		}
	}
}
