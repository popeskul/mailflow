package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending = "pending"
	StatusSent    = "sent"
	StatusFailed  = "failed"
)

type Email struct {
	ID        string
	To        string
	Subject   string
	Body      string
	Status    string
	CreatedAt time.Time
	SentAt    *time.Time
}

func NewEmail(to, subject, body string) *Email {
	return &Email{
		ID:        uuid.New().String(),
		To:        to,
		Subject:   subject,
		Body:      body,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}
}
