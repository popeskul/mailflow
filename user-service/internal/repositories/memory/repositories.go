package memory

import (
	"github.com/popeskul/mailflow/common/logger"
	"github.com/popeskul/mailflow/user-service/internal/domain"
)

type Repositories struct {
	user domain.UserRepository
}

func NewRepositories(logger logger.Logger) *Repositories {
	return &Repositories{
		user: newUserRepository(logger),
	}
}

func (r Repositories) User() domain.UserRepository {
	return r.user
}
