package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/popeskul/mailflow/user-service/internal/domain"
)

// ErrQueueEmpty indicates that the queue is empty
var ErrQueueEmpty = fmt.Errorf("queue is empty")

// Queue defines the interface for email queue operations
type Queue interface {
	Enqueue(email *domain.Email) error
	Start(ctx context.Context, processor func(*domain.Email) error)
	Stop()
	Size() int
}

// EmailQueue represents an email queue for retry logic
type EmailQueue struct {
	queue  chan *domain.Email
	logger *zap.Logger
	done   chan struct{}
	wg     sync.WaitGroup
}

// NewEmailQueue creates a new email queue
func NewEmailQueue(bufferSize int, logger *zap.Logger) *EmailQueue {
	return &EmailQueue{
		queue:  make(chan *domain.Email, bufferSize),
		logger: logger,
		done:   make(chan struct{}),
	}
}

// Enqueue adds an email to the retry queue
func (q *EmailQueue) Enqueue(email *domain.Email) error {
	select {
	case q.queue <- email:
		q.logger.Debug("Email enqueued for retry", zap.String("email_id", email.ID))
		return nil
	default:
		return fmt.Errorf("queue is full")
	}
}

// Start begins processing the queue
func (q *EmailQueue) Start(ctx context.Context, processor func(*domain.Email) error) {
	q.wg.Add(1)
	go func() {
		defer q.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-q.done:
				return
			case email := <-q.queue:
				if email != nil {
					if err := processor(email); err != nil {
						q.logger.Error("Failed to process email from queue",
							zap.String("email_id", email.ID),
							zap.Error(err))
						// Retry after delay
						time.Sleep(5 * time.Second)
						if retryErr := q.Enqueue(email); retryErr != nil {
							q.logger.Error("Failed to re-enqueue email",
								zap.String("email_id", email.ID),
								zap.Error(retryErr))
						}
					}
				}
			}
		}
	}()
}

// Stop stops the queue processing
func (q *EmailQueue) Stop() {
	close(q.done)
	q.wg.Wait()
}

// Size returns the current queue size
func (q *EmailQueue) Size() int {
	return len(q.queue)
}

// MockEmailQueue for testing
type MockEmailQueue struct {
	emails []*domain.Email
}

// NewMockEmailQueue creates a new mock email queue
func NewMockEmailQueue() *MockEmailQueue {
	return &MockEmailQueue{
		emails: make([]*domain.Email, 0),
	}
}

// Enqueue adds an email to the mock queue
func (m *MockEmailQueue) Enqueue(email *domain.Email) error {
	m.emails = append(m.emails, email)
	return nil
}

// Start does nothing for mock
func (m *MockEmailQueue) Start(_ context.Context, _ func(*domain.Email) error) {
	// Mock implementation - does nothing
}

// Stop does nothing for mock
func (m *MockEmailQueue) Stop() {
	// Mock implementation - does nothing
}

// Size returns the mock queue size
func (m *MockEmailQueue) Size() int {
	return len(m.emails)
}

// GetEmails returns all emails in mock queue
func (m *MockEmailQueue) GetEmails() []*domain.Email {
	return m.emails
}

// Clear empties the mock queue
func (m *MockEmailQueue) Clear() {
	m.emails = m.emails[:0]
}
