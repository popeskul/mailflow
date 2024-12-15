package memory

import (
	"context"
	"errors"
	"sort"
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
	mu     *sync.RWMutex
}

func newEmailRepository() ports.EmailRepository {
	return &emailRepository{
		emails: make(map[string]*domain.Email),
		mu:     &sync.RWMutex{},
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

	if pageSize <= 0 {
		pageSize = 10
	}

	emails := make([]*domain.Email, len(r.emails))
	i := 0
	for _, email := range r.emails {
		emails[i] = email
		i++
	}

	sort.Slice(emails, func(i, j int) bool {
		if emails[i].CreatedAt.Equal(emails[j].CreatedAt) {
			return emails[i].ID < emails[j].ID
		}
		return emails[i].CreatedAt.Before(emails[j].CreatedAt)
	})

	startIndex := 0
	if pageToken != "" {
		for i, email := range emails {
			if email.ID == pageToken {
				startIndex = i + 1
				break
			}
		}
	}

	if startIndex >= len(emails) {
		return nil, "", nil
	}

	endIndex := startIndex + pageSize
	if endIndex > len(emails) {
		endIndex = len(emails)
	}

	result := emails[startIndex:endIndex]

	var nextPageToken string
	if endIndex < len(emails) {
		nextPageToken = emails[endIndex-1].ID
	}

	return result, nextPageToken, nil
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
