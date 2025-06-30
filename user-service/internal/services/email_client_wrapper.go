package services

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/popeskul/mailflow/common/logger"
	emailv1 "github.com/popeskul/mailflow/email-service/pkg/api/email/v1"
	"github.com/popeskul/mailflow/user-service/internal/circuitbreaker"
	"github.com/popeskul/mailflow/user-service/internal/domain"
	"github.com/popeskul/mailflow/user-service/internal/queue"
	"github.com/popeskul/mailflow/user-service/internal/retry"
)

// EmailClientWrapper wraps the email client with resilience patterns
type EmailClientWrapper struct {
	client         emailv1.EmailServiceClient
	circuitBreaker *circuitbreaker.CircuitBreaker
	retrier        *retry.Retrier
	queue          *queue.EmailQueue
	logger         logger.Logger
}

// NewEmailClientWrapper creates a new wrapped email client
func NewEmailClientWrapper(
	client emailv1.EmailServiceClient,
	cb *circuitbreaker.CircuitBreaker,
	q *queue.EmailQueue,
	l logger.Logger,
) *EmailClientWrapper {
	return &EmailClientWrapper{
		client:         client,
		circuitBreaker: cb,
		retrier:        retry.New(retry.DefaultExponentialBackoff()),
		queue:          q,
		logger:         l.Named("email_client_wrapper"),
	}
}

// SendEmail sends an email with circuit breaker and retry logic
func (w *EmailClientWrapper) SendEmail(ctx context.Context, req *emailv1.SendEmailRequest) error {
	// First, try to send directly
	err := w.sendWithCircuitBreaker(ctx, req)

	if err == nil {
		return nil
	}

	// If circuit is open or service is unavailable, queue the request
	if err == circuitbreaker.ErrCircuitOpen || isServiceUnavailable(err) {
		w.logger.Info("email service unavailable, queueing request",
			logger.Field{Key: "to", Value: req.To},
			logger.Field{Key: "error", Value: err},
		)

		// Convert to domain.Email for queueing
		email := &domain.Email{
			ID:      fmt.Sprintf("email_%d", time.Now().UnixNano()),
			To:      req.To,
			Subject: req.Subject,
			Body:    req.Body,
		}

		if qErr := w.queue.Enqueue(email); qErr != nil {
			w.logger.Error("failed to queue email request",
				logger.Field{Key: "error", Value: qErr},
			)
			return fmt.Errorf("email service unavailable and failed to queue: %w", qErr)
		}

		w.logger.Info("email request queued successfully",
			logger.Field{Key: "to", Value: req.To},
		)

		return nil // Successfully queued
	}

	return err
}

// sendWithCircuitBreaker sends email with circuit breaker protection
func (w *EmailClientWrapper) sendWithCircuitBreaker(ctx context.Context, req *emailv1.SendEmailRequest) error {
	return w.circuitBreaker.Execute(ctx, func(ctx context.Context) error {
		return w.sendWithRetry(ctx, req)
	})
}

// sendWithRetry sends email with retry logic
func (w *EmailClientWrapper) sendWithRetry(ctx context.Context, req *emailv1.SendEmailRequest) error {
	return w.retrier.Do(ctx, func(ctx context.Context) error {
		_, err := w.client.SendEmail(ctx, req)
		if err != nil {
			w.logger.Debug("email send attempt failed",
				logger.Field{Key: "error", Value: err},
				logger.Field{Key: "to", Value: req.To},
			)
		}
		return err
	})
}

// ProcessQueue processes queued email requests
func (w *EmailClientWrapper) ProcessQueue(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	w.logger.Info("starting queue processor")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("stopping queue processor")
			return
		case <-ticker.C:
			w.processQueuedEmails(ctx)
		}
	}
}

// processQueuedEmails processes emails from the queue
func (w *EmailClientWrapper) processQueuedEmails(ctx context.Context) {
	// Simple implementation - just log that we're checking the queue
	// The actual processing is handled by the queue's Start method
	queueSize := w.queue.Size()
	if queueSize > 0 {
		w.logger.Info("emails in queue waiting for processing",
			logger.Field{Key: "queue_size", Value: queueSize},
		)
	}
}

// isServiceUnavailable checks if the error indicates service unavailability
func isServiceUnavailable(err error) bool {
	if err == nil {
		return false
	}

	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	return st.Code() == codes.Unavailable ||
		st.Code() == codes.DeadlineExceeded ||
		st.Code() == codes.Aborted
}
