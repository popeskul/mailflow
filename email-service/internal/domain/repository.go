package domain

import (
	"context"
	"time"
)

type EmailRepository interface {
	Save(ctx context.Context, email *Email) error
	GetByID(ctx context.Context, id string) (*Email, error)
	UpdateStatus(ctx context.Context, id, status string, sentAt *time.Time) error
	List(ctx context.Context, pageSize int, pageToken string) ([]*Email, string, error)
	DeleteByID(ctx context.Context, id string) error
}
