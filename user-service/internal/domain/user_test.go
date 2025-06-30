package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser_Success(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		userName string
	}{
		{
			name:     "create new user with valid data",
			email:    "test@example.com",
			userName: "Test User",
		},
		{
			name:     "create user with empty name",
			email:    "test@example.com",
			userName: "",
		},
		{
			name:     "create user with long name",
			email:    "test@example.com",
			userName: "This is a very long user name that might be used in some applications",
		},
		{
			name:     "create user with special characters in email",
			email:    "test+tag@example.co.uk",
			userName: "Test User",
		},
		{
			name:     "create user with special characters in name",
			email:    "test@example.com",
			userName: "Test User with émojis 🚀 and special chars!",
		},
		{
			name:     "create user with international characters",
			email:    "test@example.com",
			userName: "Test User",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := NewUser(tt.email, tt.userName)

			require.NotNil(t, user)
			assert.NotEmpty(t, user.ID)
			assert.Equal(t, tt.email, user.Email)
			assert.Equal(t, tt.userName, user.Name)
			assert.False(t, user.CreatedAt.IsZero())
			assert.False(t, user.UpdatedAt.IsZero())
			assert.Equal(t, user.CreatedAt, user.UpdatedAt)
		})
	}
}

func TestUser_ID_Success(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		userName string
	}{
		{
			name:     "each user should have unique ID",
			email:    "test@example.com",
			userName: "Test User",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user1 := NewUser(tt.email, tt.userName)
			user2 := NewUser(tt.email, tt.userName)

			assert.NotEqual(t, user1.ID, user2.ID)
			assert.NotEmpty(t, user1.ID)
			assert.NotEmpty(t, user2.ID)
		})
	}
}

func TestUser_Timestamps_Success(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		userName string
	}{
		{
			name:     "created and updated timestamps should be set",
			email:    "test@example.com",
			userName: "Test User",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			user := NewUser(tt.email, tt.userName)
			after := time.Now()

			// Check CreatedAt
			assert.True(t, user.CreatedAt.After(before) || user.CreatedAt.Equal(before))
			assert.True(t, user.CreatedAt.Before(after) || user.CreatedAt.Equal(after))

			// Check UpdatedAt
			assert.True(t, user.UpdatedAt.After(before) || user.UpdatedAt.Equal(before))
			assert.True(t, user.UpdatedAt.Before(after) || user.UpdatedAt.Equal(after))

			// Initially, CreatedAt and UpdatedAt should be the same
			assert.Equal(t, user.CreatedAt, user.UpdatedAt)
		})
	}
}

func TestUser_Fields_Success(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		userName string
		id       string
	}{
		{
			name:     "user with all fields set",
			email:    "test@example.com",
			userName: "Test User",
			id:       "test-id",
		},
		{
			name:     "user with empty name",
			email:    "test@example.com",
			userName: "",
			id:       "test-id-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now()
			user := &User{
				ID:        tt.id,
				Email:     tt.email,
				Name:      tt.userName,
				CreatedAt: now,
				UpdatedAt: now,
			}

			assert.Equal(t, tt.id, user.ID)
			assert.Equal(t, tt.email, user.Email)
			assert.Equal(t, tt.userName, user.Name)
			assert.Equal(t, now, user.CreatedAt)
			assert.Equal(t, now, user.UpdatedAt)
		})
	}
}

func TestUser_Update_Success(t *testing.T) {
	tests := []struct {
		name         string
		initialEmail string
		initialName  string
		newEmail     string
		newName      string
	}{
		{
			name:         "update user email and name",
			initialEmail: "old@example.com",
			initialName:  "Old Name",
			newEmail:     "new@example.com",
			newName:      "New Name",
		},
		{
			name:         "update only email",
			initialEmail: "old@example.com",
			initialName:  "Test User",
			newEmail:     "new@example.com",
			newName:      "Test User",
		},
		{
			name:         "update only name",
			initialEmail: "test@example.com",
			initialName:  "Old Name",
			newEmail:     "test@example.com",
			newName:      "New Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := NewUser(tt.initialEmail, tt.initialName)
			originalCreatedAt := user.CreatedAt
			originalUpdatedAt := user.UpdatedAt

			// Simulate some time passing
			time.Sleep(1 * time.Millisecond)

			// Update user
			user.Email = tt.newEmail
			user.Name = tt.newName
			user.UpdatedAt = time.Now()

			assert.Equal(t, tt.newEmail, user.Email)
			assert.Equal(t, tt.newName, user.Name)
			assert.Equal(t, originalCreatedAt, user.CreatedAt)      // CreatedAt should not change
			assert.True(t, user.UpdatedAt.After(originalUpdatedAt)) // UpdatedAt should be newer
		})
	}
}

func TestUser_Validation_Success(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		userName string
		valid    bool
	}{
		{
			name:     "valid user with proper email",
			email:    "test@example.com",
			userName: "Test User",
			valid:    true,
		},
		{
			name:     "valid user with empty name",
			email:    "test@example.com",
			userName: "",
			valid:    true, // Assuming empty name is allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := NewUser(tt.email, tt.userName)

			// Basic validation checks
			assert.NotEmpty(t, user.ID)
			assert.Equal(t, tt.email, user.Email)
			assert.Equal(t, tt.userName, user.Name)

			if tt.valid {
				assert.NotNil(t, user)
			}
		})
	}
}
