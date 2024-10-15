package services

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/popeskul/mailflow/common/logger"
	"github.com/popeskul/mailflow/email-service/internal/domain"
)

type emailService struct {
	repo        EmailRepository
	sender      EmailSender
	rateLimiter Limiter
	metrics     Metrics
	retryQueue  chan *domain.Email
	logger      logger.Logger
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
	// Get tracer from global provider
	tracer := otel.GetTracerProvider().Tracer("email-service")
	ctx, span := tracer.Start(ctx, "SendEmail",
		trace.WithAttributes(
			attribute.String("email.to", to),
			attribute.String("email.subject", subject),
		))
	defer span.End()

	l := s.logger.WithFields(logger.Fields{
		"to":      to,
		"subject": subject,
	})

	email := domain.NewEmail(to, subject, body)
	span.SetAttributes(attribute.String("email.id", email.ID))

	// Save email to repository
	saveCtx, saveSpan := tracer.Start(ctx, "SaveEmailToRepository")
	l.Info("attempting to save email",
		logger.Field{Key: "email_id", Value: email.ID},
	)
	if err := s.repo.Save(saveCtx, email); err != nil {
		saveSpan.RecordError(err)
		saveSpan.SetStatus(codes.Error, err.Error())
		saveSpan.End()
		l.Error("failed to save email",
			logger.Field{Key: "error", Value: err},
		)
		return nil, fmt.Errorf("failed to save email: %w", err)
	}
	saveSpan.End()

	// Check rate limit
	rateLimitCtx, rateLimitSpan := tracer.Start(ctx, "RateLimitCheck")
	l.Info("attempting to send email")
	if err := s.rateLimiter.Wait(rateLimitCtx); err != nil {
		rateLimitSpan.RecordError(err)
		rateLimitSpan.SetStatus(codes.Error, "rate limit exceeded")
		rateLimitSpan.End()
		l.Warn("rate limit exceeded, queueing email for retry",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "email_id", Value: email.ID},
		)
		s.metrics.RecordRateLimitDelay()
		s.queueForRetry(email)
		span.SetAttributes(attribute.Bool("email.queued", true))
		return email, nil
	}
	rateLimitSpan.End()

	// Send email
	sendCtx, sendSpan := tracer.Start(ctx, "SendEmailViaSMTP")
	if err := s.sender.Send(sendCtx, email); err != nil {
		sendSpan.RecordError(err)
		sendSpan.SetStatus(codes.Error, err.Error())
		sendSpan.End()
		l.Error("failed to send email",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "email_id", Value: email.ID},
		)
		s.metrics.RecordEmailFailed()
		s.queueForRetry(email)
		span.SetAttributes(attribute.Bool("email.failed", true))
		return email, nil
	}
	sendSpan.End()

	now := time.Now()
	email.Status = domain.StatusSent
	email.SentAt = &now

	s.metrics.RecordEmailSent()
	span.SetAttributes(
		attribute.String("email.status", email.Status),
		attribute.String("email.sent_at", now.Format(time.RFC3339)),
	)

	l.Info("email sent successfully",
		logger.Field{Key: "email_id", Value: email.ID},
		logger.Field{Key: "metrics_sent", Value: true},
	)

	// Update status in repository
	updateCtx, updateSpan := tracer.Start(ctx, "UpdateEmailStatus")
	if err := s.repo.UpdateStatus(updateCtx, email.ID, email.Status, email.SentAt); err != nil {
		updateSpan.RecordError(err)
		updateSpan.SetStatus(codes.Error, err.Error())
		updateSpan.End()
		l.Error("failed to update email status",
			logger.Field{Key: "error", Value: err},
		)
		return email, fmt.Errorf("failed to update email status: %w", err)
	}
	updateSpan.End()

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
	startTime := time.Now()

	l := s.logger.WithFields(logger.Fields{
		"email_id": email.ID,
		"status":   email.Status,
	})

	email.Status = domain.StatusPending
	s.metrics.RecordEmailQueued()

	l.Info("email successfully queued for retry",
		logger.Field{Key: "queue_time", Value: time.Since(startTime)},
	)

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
		queueSize := len(s.retryQueue)
		l := l.WithFields(logger.Fields{
			"email_id":   email.ID,
			"queue_size": queueSize,
		})

		l.Info("processing queued email")

		if err := s.rateLimiter.Wait(ctx); err != nil {
			l.Warn("rate limit still exceeded, requeueing email",
				logger.Field{Key: "error", Value: err},
			)
			s.queueForRetry(email)
			continue
		}

		if err := s.sender.Send(ctx, email); err != nil {
			l.Error("failed to send queued email",
				logger.Field{Key: "error", Value: err},
			)
			s.queueForRetry(email)
			continue
		}

		now := time.Now()
		email.Status = domain.StatusSent
		email.SentAt = &now

		l.Info("queued email sent successfully")
		s.metrics.RecordEmailSent()

		if err := s.repo.UpdateStatus(ctx, email.ID, email.Status, email.SentAt); err != nil {
			l.Error("failed to update queued email status",
				logger.Field{Key: "error", Value: err},
			)
			s.queueForRetry(email)
			continue
		}
	}
}
