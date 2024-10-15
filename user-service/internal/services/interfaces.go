package services

import (
	"github.com/popeskul/mailflow/user-service/internal/domain"
)

type Repositories interface {
	User() domain.UserRepository
}
