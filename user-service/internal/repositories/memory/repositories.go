package memory

import "github.com/popeskul/email-service-platform/user-service/internal/services"

type Repositories struct {
	user *userRepository
}

func NewRepositories() *Repositories {
	return &Repositories{
		user: newUserRepository(),
	}
}

func (r Repositories) UserRepository() services.Repository {
	return r.user
}
