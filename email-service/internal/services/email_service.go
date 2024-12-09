package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/popeskul/email-service-platform/email-service/internal/core/domain"
	"github.com/popeskul/email-service-platform/email-service/internal/core/ports"
	"github.com/popeskul/email-service-platform/email-service/internal/metrics"
	"github.com/popeskul/ratelimiter"
)

type emailService struct {
	repo        ports.EmailRepository
	sender      ports.EmailSender
	rateLimiter ratelimiter.Limiter
	metrics     *metrics.EmailMetrics
	retryQueue  chan *domain.Email
	logger      *zap.Logger
	mu          sync.Mutex
}

func NewEmailService(
	repo ports.EmailRepository,
	sender ports.EmailSender,
	limiter ratelimiter.Limiter,
	metrics *metrics.EmailMetrics,
	logger *zap.Logger,
) ports.EmailService {
	svc := &emailService{
		repo:        repo,
		sender:      sender,
		rateLimiter: limiter,
		metrics:     metrics,
		retryQueue:  make(chan *domain.Email, 1000),
		logger:      logger.Named("email_service"),
	}

	go svc.processRetryQueue()

	return svc
}

func (s *emailService) SendEmail(ctx context.Context, to, subject, body string) (*domain.Email, error) {
	logger := s.logger.With(
		zap.String("to", to),
		zap.String("subject", subject),
	)

	email := domain.NewEmail(to, subject, body)

	logger.Info("attempting to save email", zap.String("email_id", email.ID))
	if err := s.repo.Save(ctx, email); err != nil {
		logger.Error("failed to save email", zap.Error(err))
		return nil, fmt.Errorf("failed to save email: %w", err)
	}

	logger.Info("attempting to send email")
	if err := s.rateLimiter.Wait(ctx); err != nil {
		logger.Warn("rate limit exceeded, queueing email for retry",
			zap.Error(err),
			zap.String("email_id", email.ID),
		)
		s.metrics.RecordRateLimitDelay()
		s.queueForRetry(email)
		return email, nil
	}

	if err := s.sender.Send(ctx, email); err != nil {
		logger.Error("failed to send email",
			zap.Error(err),
			zap.String("email_id", email.ID),
		)
		s.metrics.RecordEmailFailed()
		s.queueForRetry(email)
		return email, nil
	}

	now := time.Now()
	email.Status = domain.StatusSent
	email.SentAt = &now

	logger.Info("email sent successfully", zap.String("email_id", email.ID))
	s.metrics.RecordEmailSent()

	if err := s.repo.UpdateStatus(ctx, email.ID, email.Status, email.SentAt); err != nil {
		logger.Error("failed to update email status", zap.Error(err))
		return email, fmt.Errorf("failed to update email status: %w", err)
	}

	return email, nil
}

func (s *emailService) GetEmailStatus(ctx context.Context, id string) (*domain.Email, error) {
	logger := s.logger.With(zap.String("email_id", id))

	email, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.Error("failed to get email status", zap.Error(err))
		return nil, fmt.Errorf("failed to get email status: %w", err)
	}

	return email, nil
}

func (s *emailService) ListEmails(ctx context.Context, pageSize int, pageToken string) ([]*domain.Email, string, error) {
	logger := s.logger.With(
		zap.Int("page_size", pageSize),
		zap.String("page_token", pageToken),
	)

	emails, nextToken, err := s.repo.List(ctx, pageSize, pageToken)
	if err != nil {
		logger.Error("failed to list emails", zap.Error(err))
		return nil, "", fmt.Errorf("failed to list emails: %w", err)
	}

	return emails, nextToken, nil
}

func (s *emailService) ResendFailedEmails(ctx context.Context) error {
	logger := s.logger.With(zap.String("operation", "resend_failed"))

	logger.Info("starting resend of failed emails")

	// We receive all letters
	emails, _, err := s.repo.List(ctx, 0, "")
	if err != nil {
		logger.Error("failed to list emails", zap.Error(err))
		return fmt.Errorf("failed to list emails: %w", err)
	}

	var resendCount int
	// Find and resend emails with FAILED status
	for _, email := range emails {
		if email.Status == domain.StatusFailed {
			logger.Info("requeueing failed email",
				zap.String("email_id", email.ID),
				zap.String("to", email.To),
			)
			s.queueForRetry(email)
			resendCount++
		}
	}

	logger.Info("finished requeueing failed emails",
		zap.Int("resend_count", resendCount),
	)
	return nil
}

func (s *emailService) queueForRetry(email *domain.Email) {
	logger := s.logger.With(
		zap.String("email_id", email.ID),
		zap.String("status", email.Status),
	)

	// Update the status to "pending" to try again
	email.Status = domain.StatusPending

	// Increase the counter in the queue
	s.metrics.RecordEmailQueued()

	// Trying to add to queue
	select {
	case s.retryQueue <- email:
		logger.Info("email successfully queued for retry")

		// Updating the status in the database
		if err := s.repo.UpdateStatus(context.Background(), email.ID, email.Status, nil); err != nil {
			logger.Error("failed to update email status after queuing",
				zap.Error(err),
			)
		}

	default:
		// The queue is overcrowded
		logger.Warn("retry queue is full, marking email as failed")

		email.Status = domain.StatusFailed
		if err := s.repo.UpdateStatus(context.Background(), email.ID, email.Status, nil); err != nil {
			logger.Error("failed to update email status when queue full",
				zap.Error(err),
			)
		}

		// Increase the error counter
		s.metrics.RecordEmailFailed()

		// You can also add a metric for queue overflow
		s.metrics.QueueSize.Set(float64(len(s.retryQueue)))
	}
}

func (s *emailService) processRetryQueue() {
	logger := s.logger.Named("retry_queue")

	for email := range s.retryQueue {
		ctx := context.Background()
		logger := logger.With(zap.String("email_id", email.ID))

		logger.Info("processing queued email")

		if err := s.rateLimiter.Wait(ctx); err != nil {
			logger.Warn("rate limit still exceeded, requeueing email", zap.Error(err))
			s.queueForRetry(email)
			continue
		}

		if err := s.sender.Send(ctx, email); err != nil {
			logger.Error("failed to send queued email", zap.Error(err))
			s.queueForRetry(email)
			continue
		}

		now := time.Now()
		email.Status = domain.StatusSent
		email.SentAt = &now

		logger.Info("queued email sent successfully")
		s.metrics.RecordEmailSent()

		if err := s.repo.UpdateStatus(ctx, email.ID, email.Status, email.SentAt); err != nil {
			logger.Error("failed to update queued email status", zap.Error(err))
			s.queueForRetry(email)
			continue
		}
	}
}
