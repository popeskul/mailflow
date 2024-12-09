package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/popeskul/email-service-platform/user-service/internal/domain"
	"github.com/popeskul/email-service-platform/user-service/internal/ports"
)

type userRepository struct {
	users map[string]*domain.User
	mu    sync.RWMutex
}

func newUserRepository() ports.UserRepository {
	return &userRepository{
		users: make(map[string]*domain.User),
	}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; exists {
		return errors.New("email already exists")
	}

	r.users[user.ID] = user
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, errors.New("email not found")
	}

	return user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; !exists {
		return errors.New("email not found")
	}

	r.users[user.ID] = user
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[id]; !exists {
		return errors.New("email not found")
	}

	delete(r.users, id)
	return nil
}

func (r *userRepository) List(ctx context.Context, pageSize int, pageToken string) ([]*domain.User, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var users []*domain.User
	var nextPageToken string

	for _, user := range r.users {
		users = append(users, user)
		if len(users) == pageSize {
			break
		}
	}

	if len(users) == pageSize && len(r.users) > pageSize {
		nextPageToken = "next_page"
	}

	return users, nextPageToken, nil
}
