package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/popeskul/email-service-platform/email-service/internal/core/domain"
	"github.com/popeskul/email-service-platform/email-service/internal/core/ports"
)

var (
	ErrEmailNotFound = errors.New("email not found")
)

type emailRepository struct {
	emails map[string]*domain.Email
	mu     sync.RWMutex
}

func newEmailRepository() ports.EmailRepository {
	return &emailRepository{
		emails: make(map[string]*domain.Email),
	}
}

func (r *emailRepository) Save(ctx context.Context, email *domain.Email) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.emails[email.ID] = email
	return nil
}

func (r *emailRepository) GetByID(ctx context.Context, id string) (*domain.Email, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	email, exists := r.emails[id]
	if !exists {
		return nil, ErrEmailNotFound
	}

	return email, nil
}

func (r *emailRepository) UpdateStatus(ctx context.Context, id, status string, sentAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	email, exists := r.emails[id]
	if !exists {
		return ErrEmailNotFound
	}

	email.Status = status
	email.SentAt = sentAt
	return nil
}

func (r *emailRepository) List(ctx context.Context, pageSize int, pageToken string) ([]*domain.Email, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var emails []*domain.Email
	var nextPageToken string

	// Simplified pagination for in-memory storage
	for _, email := range r.emails {
		emails = append(emails, email)
		if pageSize > 0 && len(emails) == pageSize {
			nextPageToken = email.ID
			break
		}
	}

	return emails, nextPageToken, nil
}

func (r *emailRepository) DeleteByID(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.emails[id]; !exists {
		return ErrEmailNotFound
	}

	delete(r.emails, id)
	return nil
}
