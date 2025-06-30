package domain

import "context"

type UserService interface {
	Create(ctx context.Context, email, username string) (*User, error)
	Get(ctx context.Context, id string) (*User, error)
	List(ctx context.Context, pageSize int, pageToken string) ([]*User, string, error)
}
