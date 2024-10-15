package memory

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/popeskul/mailflow/common/logger"
	"github.com/popeskul/mailflow/user-service/internal/domain"
)

type UserRepository struct {
	users       map[string]*domain.User
	sortedUsers []*domain.User
	mu          *sync.RWMutex
	logger      logger.Logger
}

func newUserRepository(logger logger.Logger) *UserRepository {
	return &UserRepository{
		users:  make(map[string]*domain.User),
		mu:     &sync.RWMutex{},
		logger: logger.Named("user_repository"),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; exists {
		return errors.New("user already exists")
	}

	r.users[user.ID] = user
	r.sortedUsers = append(r.sortedUsers, user)
	sort.Slice(r.sortedUsers, r.sortUsers)

	return nil
}

func (r *UserRepository) sortUsers(i, j int) bool {
	if r.sortedUsers[i].CreatedAt.Equal(r.sortedUsers[j].CreatedAt) {
		return r.sortedUsers[i].ID < r.sortedUsers[j].ID
	}
	return r.sortedUsers[i].CreatedAt.Before(r.sortedUsers[j].CreatedAt)
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, errors.New("email not found")
	}

	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; !exists {
		return errors.New("email not found")
	}

	r.users[user.ID] = user
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[id]; !exists {
		return errors.New("email not found")
	}

	delete(r.users, id)
	return nil
}

func (r *UserRepository) List(ctx context.Context, pageSize int, pageToken string) ([]*domain.User, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	startIndex := 0
	if pageToken != "" {
		for i, user := range r.sortedUsers {
			if user.ID == pageToken {
				startIndex = i + 1
				break
			}
		}
	}

	if startIndex >= len(r.sortedUsers) {
		return nil, "", nil
	}

	endIndex := startIndex + pageSize
	if endIndex > len(r.sortedUsers) {
		endIndex = len(r.sortedUsers)
	}

	result := r.sortedUsers[startIndex:endIndex]

	var nextPageToken string
	if endIndex < len(r.sortedUsers) {
		nextPageToken = r.sortedUsers[endIndex-1].ID
	}

	return result, nextPageToken, nil
}
