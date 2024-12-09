package memory

import (
	"github.com/popeskul/email-service-platform/user-service/internal/ports"
)

type Repositories struct {
	User ports.UserRepository
}

func (r *Repositories) UserRepository() ports.UserRepository {
	return r.User
}

func NewRepositories() *Repositories {
	return &Repositories{
		User: newUserRepository(),
	}
}
