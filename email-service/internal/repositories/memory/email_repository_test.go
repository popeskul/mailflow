package memory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/popeskul/mailflow/common/logger"
	"github.com/popeskul/mailflow/email-service/internal/domain"
)

func createTestEmailRepository() *EmailRepositoryContainer {
	testLogger := logger.NewZapLogger()
	return newEmailRepository(testLogger)
}

func createTestEmail(to, subject, body string) *domain.Email {
	return domain.NewEmail(to, subject, body)
}

func TestEmailRepository_Save_Success(t *testing.T) {
	tests := []struct {
		name  string
		email *domain.Email
	}{
		{
			name:  "save email successfully",
			email: createTestEmail("test@example.com", "Test Subject", "Test Body"),
		},
		{
			name:  "save email with empty body",
			email: createTestEmail("test2@example.com", "Test Subject 2", ""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := createTestEmailRepository()

			err := repo.Save(context.Background(), tt.email)

			assert.NoError(t, err)

			// Verify email is stored
			storedEmail, err := repo.GetByID(context.Background(), tt.email.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.email.ID, storedEmail.ID)
			assert.Equal(t, tt.email.To, storedEmail.To)
			assert.Equal(t, tt.email.Subject, storedEmail.Subject)
			assert.Equal(t, tt.email.Body, storedEmail.Body)
		})
	}
}

func TestEmailRepository_Save_Overwrite(t *testing.T) {
	repo := createTestEmailRepository()
	email := createTestEmail("test@example.com", "Original Subject", "Original Body")

	// Save email first time
	err := repo.Save(context.Background(), email)
	require.NoError(t, err)

	// Update and save again
	email.Subject = "Updated Subject"
	email.Body = "Updated Body"
	err = repo.Save(context.Background(), email)

	assert.NoError(t, err)

	// Verify update
	storedEmail, err := repo.GetByID(context.Background(), email.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Subject", storedEmail.Subject)
	assert.Equal(t, "Updated Body", storedEmail.Body)
}

func TestEmailRepository_GetByID_Success(t *testing.T) {
	tests := []struct {
		name  string
		email *domain.Email
	}{
		{
			name:  "get existing email",
			email: createTestEmail("test@example.com", "Test Subject", "Test Body"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := createTestEmailRepository()

			// Save email
			err := repo.Save(context.Background(), tt.email)
			require.NoError(t, err)

			// Get email
			email, err := repo.GetByID(context.Background(), tt.email.ID)

			assert.NoError(t, err)
			assert.NotNil(t, email)
			assert.Equal(t, tt.email.ID, email.ID)
			assert.Equal(t, tt.email.To, email.To)
			assert.Equal(t, tt.email.Subject, email.Subject)
			assert.Equal(t, tt.email.Body, email.Body)
		})
	}
}

func TestEmailRepository_GetByID_Fail(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{
			name: "get non-existent email",
			id:   "non-existent-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := createTestEmailRepository()

			email, err := repo.GetByID(context.Background(), tt.id)

			assert.Error(t, err)
			assert.Nil(t, email)
			assert.Equal(t, ErrEmailNotFound, err)
		})
	}
}

func TestEmailRepository_UpdateStatus_Success(t *testing.T) {
	tests := []struct {
		name      string
		email     *domain.Email
		newStatus string
		sentAt    *time.Time
	}{
		{
			name:      "update status to sent",
			email:     createTestEmail("test@example.com", "Test Subject", "Test Body"),
			newStatus: domain.StatusSent,
			sentAt:    &time.Time{},
		},
		{
			name:      "update status to failed",
			email:     createTestEmail("test2@example.com", "Test Subject 2", "Test Body 2"),
			newStatus: domain.StatusFailed,
			sentAt:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := createTestEmailRepository()

			// Save email
			err := repo.Save(context.Background(), tt.email)
			require.NoError(t, err)

			// Update status
			err = repo.UpdateStatus(context.Background(), tt.email.ID, tt.newStatus, tt.sentAt)

			assert.NoError(t, err)

			// Verify update
			email, err := repo.GetByID(context.Background(), tt.email.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.newStatus, email.Status)
			assert.Equal(t, tt.sentAt, email.SentAt)
		})
	}
}

func TestEmailRepository_UpdateStatus_Fail(t *testing.T) {
	tests := []struct {
		name   string
		id     string
		status string
	}{
		{
			name:   "update status of non-existent email",
			id:     "non-existent-id",
			status: domain.StatusSent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := createTestEmailRepository()

			err := repo.UpdateStatus(context.Background(), tt.id, tt.status, nil)

			assert.Error(t, err)
			assert.Equal(t, ErrEmailNotFound, err)
		})
	}
}

func TestEmailRepository_List_Success(t *testing.T) {
	tests := []struct {
		name      string
		emails    []*domain.Email
		pageSize  int
		pageToken string
		expected  int
	}{
		{
			name: "list all emails",
			emails: []*domain.Email{
				createTestEmail("user1@example.com", "Subject 1", "Body 1"),
				createTestEmail("user2@example.com", "Subject 2", "Body 2"),
				createTestEmail("user3@example.com", "Subject 3", "Body 3"),
			},
			pageSize:  10,
			pageToken: "",
			expected:  3,
		},
		{
			name: "list with pagination",
			emails: []*domain.Email{
				createTestEmail("user1@example.com", "Subject 1", "Body 1"),
				createTestEmail("user2@example.com", "Subject 2", "Body 2"),
				createTestEmail("user3@example.com", "Subject 3", "Body 3"),
			},
			pageSize:  2,
			pageToken: "",
			expected:  2,
		},
		{
			name:      "list empty repository",
			emails:    []*domain.Email{},
			pageSize:  10,
			pageToken: "",
			expected:  0,
		},
		{
			name: "list with zero page size",
			emails: []*domain.Email{
				createTestEmail("user1@example.com", "Subject 1", "Body 1"),
			},
			pageSize:  0,
			pageToken: "",
			expected:  1, // Should default to pageSize 10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := createTestEmailRepository()

			// Save emails
			for _, email := range tt.emails {
				err := repo.Save(context.Background(), email)
				require.NoError(t, err)
			}

			// Wait a bit to ensure different timestamps
			time.Sleep(time.Millisecond)

			emails, nextToken, err := repo.List(context.Background(), tt.pageSize, tt.pageToken)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, len(emails))

			if tt.expected > 0 && tt.pageSize > 0 && tt.pageSize < len(tt.emails) {
				assert.NotEmpty(t, nextToken)
			}

			// Verify emails are sorted by creation time
			for i := 1; i < len(emails); i++ {
				assert.True(t, emails[i-1].CreatedAt.Before(emails[i].CreatedAt) ||
					(emails[i-1].CreatedAt.Equal(emails[i].CreatedAt) && emails[i-1].ID < emails[i].ID))
			}
		})
	}
}

func TestEmailRepository_List_WithPageToken(t *testing.T) {
	repo := createTestEmailRepository()

	// Create multiple emails
	emails := []*domain.Email{
		createTestEmail("user1@example.com", "Subject 1", "Body 1"),
		createTestEmail("user2@example.com", "Subject 2", "Body 2"),
		createTestEmail("user3@example.com", "Subject 3", "Body 3"),
		createTestEmail("user4@example.com", "Subject 4", "Body 4"),
	}

	for _, email := range emails {
		err := repo.Save(context.Background(), email)
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
	for _, email := range firstPage {
		firstPageIDs[email.ID] = true
	}

	for _, email := range secondPage {
		assert.False(t, firstPageIDs[email.ID], "Email %s should not appear in both pages", email.ID)
	}
}

func TestEmailRepository_List_InvalidPageToken(t *testing.T) {
	repo := createTestEmailRepository()

	// Create an email
	email := createTestEmail("test@example.com", "Test Subject", "Test Body")
	err := repo.Save(context.Background(), email)
	require.NoError(t, err)

	// Use invalid page token - this will start from beginning since token not found
	emails, nextToken, err := repo.List(context.Background(), 10, "invalid-token")

	assert.NoError(t, err)
	assert.Equal(t, 1, len(emails)) // Should return all emails since invalid token means start from beginning
	assert.Empty(t, nextToken)
}

func TestEmailRepository_DeleteByID_Success(t *testing.T) {
	tests := []struct {
		name  string
		email *domain.Email
	}{
		{
			name:  "delete email successfully",
			email: createTestEmail("test@example.com", "Test Subject", "Test Body"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := createTestEmailRepository()

			// Save email
			err := repo.Save(context.Background(), tt.email)
			require.NoError(t, err)

			// Delete email
			err = repo.DeleteByID(context.Background(), tt.email.ID)

			assert.NoError(t, err)

			// Verify deletion
			email, err := repo.GetByID(context.Background(), tt.email.ID)
			assert.Error(t, err)
			assert.Nil(t, email)
			assert.Equal(t, ErrEmailNotFound, err)
		})
	}
}

func TestEmailRepository_DeleteByID_Fail(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{
			name: "delete non-existent email",
			id:   "non-existent-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := createTestEmailRepository()

			err := repo.DeleteByID(context.Background(), tt.id)

			assert.Error(t, err)
			assert.Equal(t, ErrEmailNotFound, err)
		})
	}
}

func TestEmailRepository_ConcurrentAccess(t *testing.T) {
	repo := createTestEmailRepository()

	// Test concurrent writes
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			email := createTestEmail(fmt.Sprintf("user%d@example.com", i),
				fmt.Sprintf("Subject %d", i), fmt.Sprintf("Body %d", i))
			err := repo.Save(context.Background(), email)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all emails were saved
	emails, _, err := repo.List(context.Background(), 20, "")
	require.NoError(t, err)
	assert.Equal(t, 10, len(emails))
}
