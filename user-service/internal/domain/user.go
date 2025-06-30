package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        string
	Email     string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(email, name string) *User {
	now := time.Now()
	return &User{
		ID:        uuid.New().String(),
		Email:     email,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
