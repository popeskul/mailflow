package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEmail_Success(t *testing.T) {
	tests := []struct {
		name    string
		to      string
		subject string
		body    string
	}{
		{
			name:    "create new email with valid data",
			to:      "test@example.com",
			subject: "Test Subject",
			body:    "Test Body",
		},
		{
			name:    "create email with empty subject",
			to:      "test@example.com",
			subject: "",
			body:    "Test Body",
		},
		{
			name:    "create email with empty body",
			to:      "test@example.com",
			subject: "Test Subject",
			body:    "",
		},
		{
			name:    "create email with long subject",
			to:      "test@example.com",
			subject: "This is a very long subject line that might be used in some email clients",
			body:    "Test Body",
		},
		{
			name:    "create email with special characters",
			to:      "test+tag@example.com",
			subject: "Test Subject with éspecial chars & symbols!",
			body:    "Test Body with 中文 and émojis 🚀",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email := NewEmail(tt.to, tt.subject, tt.body)

			require.NotNil(t, email)
			assert.NotEmpty(t, email.ID)
			assert.Equal(t, tt.to, email.To)
			assert.Equal(t, tt.subject, email.Subject)
			assert.Equal(t, tt.body, email.Body)
			assert.Equal(t, StatusPending, email.Status)
			assert.False(t, email.CreatedAt.IsZero())
			assert.Nil(t, email.SentAt)
		})
	}
}

func TestEmail_ID_Success(t *testing.T) {
	tests := []struct {
		name    string
		to      string
		subject string
		body    string
	}{
		{
			name:    "each email should have unique ID",
			to:      "test@example.com",
			subject: "Test Subject",
			body:    "Test Body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email1 := NewEmail(tt.to, tt.subject, tt.body)
			email2 := NewEmail(tt.to, tt.subject, tt.body)

			assert.NotEqual(t, email1.ID, email2.ID)
			assert.NotEmpty(t, email1.ID)
			assert.NotEmpty(t, email2.ID)
		})
	}
}

func TestEmail_CreatedAt_Success(t *testing.T) {
	tests := []struct {
		name    string
		to      string
		subject string
		body    string
	}{
		{
			name:    "created at should be set to current time",
			to:      "test@example.com",
			subject: "Test Subject",
			body:    "Test Body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			email := NewEmail(tt.to, tt.subject, tt.body)
			after := time.Now()

			assert.True(t, email.CreatedAt.After(before) || email.CreatedAt.Equal(before))
			assert.True(t, email.CreatedAt.Before(after) || email.CreatedAt.Equal(after))
		})
	}
}

func TestEmail_Status_Constants_Success(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{
			name:     "pending status constant",
			status:   StatusPending,
			expected: "pending",
		},
		{
			name:     "sent status constant",
			status:   StatusSent,
			expected: "sent",
		},
		{
			name:     "failed status constant",
			status:   StatusFailed,
			expected: "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status)
		})
	}
}

func TestEmail_Fields_Success(t *testing.T) {
	tests := []struct {
		name    string
		to      string
		subject string
		body    string
		status  string
		sentAt  *time.Time
	}{
		{
			name:    "email with all fields set",
			to:      "test@example.com",
			subject: "Test Subject",
			body:    "Test Body",
			status:  StatusSent,
			sentAt:  func() *time.Time { t := time.Now(); return &t }(),
		},
		{
			name:    "email with pending status and no sent time",
			to:      "test@example.com",
			subject: "Test Subject",
			body:    "Test Body",
			status:  StatusPending,
			sentAt:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email := &Email{
				ID:        "test-id",
				To:        tt.to,
				Subject:   tt.subject,
				Body:      tt.body,
				Status:    tt.status,
				CreatedAt: time.Now(),
				SentAt:    tt.sentAt,
			}

			assert.Equal(t, "test-id", email.ID)
			assert.Equal(t, tt.to, email.To)
			assert.Equal(t, tt.subject, email.Subject)
			assert.Equal(t, tt.body, email.Body)
			assert.Equal(t, tt.status, email.Status)
			assert.Equal(t, tt.sentAt, email.SentAt)
		})
	}
}

func TestEmail_StatusTransitions_Success(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus string
		finalStatus   string
	}{
		{
			name:          "pending to sent",
			initialStatus: StatusPending,
			finalStatus:   StatusSent,
		},
		{
			name:          "pending to failed",
			initialStatus: StatusPending,
			finalStatus:   StatusFailed,
		},
		{
			name:          "failed to sent",
			initialStatus: StatusFailed,
			finalStatus:   StatusSent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email := NewEmail("test@example.com", "Test", "Body")
			email.Status = tt.initialStatus

			// Simulate status transition
			email.Status = tt.finalStatus
			if tt.finalStatus == StatusSent {
				now := time.Now()
				email.SentAt = &now
			}

			assert.Equal(t, tt.finalStatus, email.Status)
			if tt.finalStatus == StatusSent {
				assert.NotNil(t, email.SentAt)
			}
		})
	}
}
