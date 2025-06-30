package memory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/popeskul/mailflow/common/logger"
	"github.com/popeskul/mailflow/user-service/internal/domain"
)

func createTestUserRepository() *UserRepository {
	testLogger := logger.NewZapLogger()
	return newUserRepository(testLogger)
}

func createTestUser(email, name string) *domain.User {
	return domain.NewUser(email, name)
}

func TestUserRepository_Create_Success(t *testing.T) {
	tests := []struct {
		name string
		user *domain.User
	}{
		{
			name: "create user successfully",
			user: createTestUser("test@example.com", "Test User"),
		},
		{
			name: "create user with empty name",
			user: createTestUser("test2@example.com", ""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := createTestUserRepository()

			err := repo.Create(context.Background(), tt.user)

			assert.NoError(t, err)

			// Verify user is stored
			storedUser, err := repo.GetByID(context.Background(), tt.user.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.user.ID, storedUser.ID)
			assert.Equal(t, tt.user.Email, storedUser.Email)
			assert.Equal(t, tt.user.Name, storedUser.Name)
		})
	}
}

func TestUserRepository_Create_Fail(t *testing.T) {
	tests := []struct {
		name string
		user *domain.User
	}{
		{
			name: "create duplicate user",
			user: createTestUser("test@example.com", "Test User"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := createTestUserRepository()

			// Create user first time
			err := repo.Create(context.Background(), tt.user)
			require.NoError(t, err)

			// Try to create the same user again
			err = repo.Create(context.Background(), tt.user)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "user already exists")
		})
	}
}

func TestUserRepository_GetByID_Success(t *testing.T) {
	tests := []struct {
		name string
		user *domain.User
	}{
		{
			name: "get existing user",
			user: createTestUser("test@example.com", "Test User"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := createTestUserRepository()

			// Create user
			err := repo.Create(context.Background(), tt.user)
			require.NoError(t, err)

			// Get user
			user, err := repo.GetByID(context.Background(), tt.user.ID)

			assert.NoError(t, err)
			assert.NotNil(t, user)
			assert.Equal(t, tt.user.ID, user.ID)
			assert.Equal(t, tt.user.Email, user.Email)
			assert.Equal(t, tt.user.Name, user.Name)
		})
	}
}

func TestUserRepository_GetByID_Fail(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{
			name: "get non-existent user",
			id:   "non-existent-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := createTestUserRepository()

			user, err := repo.GetByID(context.Background(), tt.id)

			assert.Error(t, err)
			assert.Nil(t, user)
			assert.Contains(t, err.Error(), "email not found")
		})
	}
}

func TestUserRepository_Update_Success(t *testing.T) {
	tests := []struct {
		name         string
		originalUser *domain.User
	}{
		{
			name:         "update user successfully",
			originalUser: createTestUser("test@example.com", "Test User"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := createTestUserRepository()

			// Create user
			err := repo.Create(context.Background(), tt.originalUser)
			require.NoError(t, err)

			// Update user
			tt.originalUser.Name = "Updated Name"
			tt.originalUser.Email = "updated@example.com"
			err = repo.Update(context.Background(), tt.originalUser)

			assert.NoError(t, err)

			// Verify update
			user, err := repo.GetByID(context.Background(), tt.originalUser.ID)
			require.NoError(t, err)
			assert.Equal(t, "Updated Name", user.Name)
			assert.Equal(t, "updated@example.com", user.Email)
		})
	}
}

func TestUserRepository_Update_Fail(t *testing.T) {
	tests := []struct {
		name string
		user *domain.User
	}{
		{
			name: "update non-existent user",
			user: createTestUser("test@example.com", "Test User"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := createTestUserRepository()

			err := repo.Update(context.Background(), tt.user)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "email not found")
		})
	}
}

func TestUserRepository_Delete_Success(t *testing.T) {
	tests := []struct {
		name string
		user *domain.User
	}{
		{
			name: "delete user successfully",
			user: createTestUser("test@example.com", "Test User"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := createTestUserRepository()

			// Create user
			err := repo.Create(context.Background(), tt.user)
			require.NoError(t, err)

			// Delete user
			err = repo.Delete(context.Background(), tt.user.ID)

			assert.NoError(t, err)

			// Verify deletion
			user, err := repo.GetByID(context.Background(), tt.user.ID)
			assert.Error(t, err)
			assert.Nil(t, user)
		})
	}
}

func TestUserRepository_Delete_Fail(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{
			name: "delete non-existent user",
			id:   "non-existent-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := createTestUserRepository()

			err := repo.Delete(context.Background(), tt.id)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "email not found")
		})
	}
}

func TestUserRepository_List_Success(t *testing.T) {
	tests := []struct {
		name      string
		users     []*domain.User
		pageSize  int
		pageToken string
		expected  int
	}{
		{
			name: "list all users",
			users: []*domain.User{
				createTestUser("user1@example.com", "User 1"),
				createTestUser("user2@example.com", "User 2"),
				createTestUser("user3@example.com", "User 3"),
			},
			pageSize:  10,
			pageToken: "",
			expected:  3,
		},
		{
			name: "list with pagination",
			users: []*domain.User{
				createTestUser("user1@example.com", "User 1"),
				createTestUser("user2@example.com", "User 2"),
				createTestUser("user3@example.com", "User 3"),
			},
			pageSize:  2,
			pageToken: "",
			expected:  2,
		},
		{
			name:      "list empty repository",
			users:     []*domain.User{},
			pageSize:  10,
			pageToken: "",
			expected:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := createTestUserRepository()

			// Create users
			for _, user := range tt.users {
				err := repo.Create(context.Background(), user)
				require.NoError(t, err)
			}

			// Wait a bit to ensure different timestamps
			time.Sleep(time.Millisecond)

			users, nextToken, err := repo.List(context.Background(), tt.pageSize, tt.pageToken)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, len(users))

			if tt.expected > 0 && tt.pageSize < len(tt.users) {
				assert.NotEmpty(t, nextToken)
			}

			// Verify users are sorted by creation time
			for i := 1; i < len(users); i++ {
				assert.True(t, users[i-1].CreatedAt.Before(users[i].CreatedAt) ||
					(users[i-1].CreatedAt.Equal(users[i].CreatedAt) && users[i-1].ID < users[i].ID))
			}
		})
	}
}

func TestUserRepository_List_WithPageToken(t *testing.T) {
	repo := createTestUserRepository()

	// Create multiple users
	users := []*domain.User{
		createTestUser("user1@example.com", "User 1"),
		createTestUser("user2@example.com", "User 2"),
		createTestUser("user3@example.com", "User 3"),
		createTestUser("user4@example.com", "User 4"),
	}

	for _, user := range users {
		err := repo.Create(context.Background(), user)
		require.NoError(t, err)
		time.Sleep(time.Millisecond) // Ensure different timestamps
	}

	// Get first page
	firstPage, nextToken, err := repo.List(context.Background(), 2, "")
	require.NoError(t, err)
	assert.Equal(t, 2, len(firstPage))
	assert.NotEmpty(t, nextToken)

	// Get second page using token
	secondPage, finalToken, err := repo.List(context.Background(), 2, nextToken)
	require.NoError(t, err)
	assert.Equal(t, 2, len(secondPage))
	assert.Empty(t, finalToken) // Should be empty as this is the last page

	// Verify no overlap between pages
	firstPageIDs := make(map[string]bool)
	for _, user := range firstPage {
		firstPageIDs[user.ID] = true
	}

	for _, user := range secondPage {
		assert.False(t, firstPageIDs[user.ID], "User %s should not appear in both pages", user.ID)
	}
}

func TestUserRepository_List_InvalidPageToken(t *testing.T) {
	repo := createTestUserRepository()

	// Create a user
	user := createTestUser("test@example.com", "Test User")
	err := repo.Create(context.Background(), user)
	require.NoError(t, err)

	// Use invalid page token - this will start from beginning since token not found
	users, nextToken, err := repo.List(context.Background(), 10, "invalid-token")

	assert.NoError(t, err)
	assert.Equal(t, 1, len(users)) // Should return all users since invalid token means start from beginning
	assert.Empty(t, nextToken)
}

func TestUserRepository_SortUsers(t *testing.T) {
	repo := createTestUserRepository()

	// Create users with different creation times
	user1 := createTestUser("user1@example.com", "User 1")
	time.Sleep(time.Millisecond)
	user2 := createTestUser("user2@example.com", "User 2")
	time.Sleep(time.Millisecond)
	user3 := createTestUser("user3@example.com", "User 3")

	// Set the same creation time for two users to test ID sorting
	user2.CreatedAt = user1.CreatedAt

	repo.sortedUsers = []*domain.User{user1, user2, user3}

	// Test sort function
	result := repo.sortUsers(0, 1) // Compare user1 and user2

	// Since they have the same creation time, should sort by ID
	expected := user1.ID < user2.ID
	assert.Equal(t, expected, result)
}

func TestUserRepository_ConcurrentAccess(t *testing.T) {
	repo := createTestUserRepository()

	// Test concurrent writes
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			user := createTestUser(fmt.Sprintf("user%d@example.com", i), fmt.Sprintf("User %d", i))
			err := repo.Create(context.Background(), user)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all users were created
	users, _, err := repo.List(context.Background(), 20, "")
	require.NoError(t, err)
	assert.Equal(t, 10, len(users))
}
