package domain

import (
	"time"

	"github.com/google/uuid"
)

// Email represents an email message in the system
type Email struct {
	ID        string
	To        string
	Subject   string
	Body      string
	CreatedAt time.Time
	SentAt    *time.Time
	Status    EmailStatus
}

// EmailStatus represents the status of an email
type EmailStatus string

const (
	EmailStatusPending EmailStatus = "pending"
	EmailStatusSent    EmailStatus = "sent"
	EmailStatusFailed  EmailStatus = "failed"
)

// NewEmail creates a new email
func NewEmail(to, subject, body string) *Email {
	return &Email{
		ID:        generateID(),
		To:        to,
		Subject:   subject,
		Body:      body,
		CreatedAt: time.Now(),
		Status:    EmailStatusPending,
	}
}

// MarkAsSent marks the email as sent
func (e *Email) MarkAsSent() {
	now := time.Now()
	e.SentAt = &now
	e.Status = EmailStatusSent
}

// MarkAsFailed marks the email as failed
func (e *Email) MarkAsFailed() {
	e.Status = EmailStatusFailed
}

// generateID generates a unique ID for the email
func generateID() string {
	return uuid.New().String()
}
