package memory

import "github.com/popeskul/email-service-platform/email-service/internal/core/ports"

type Repositories struct {
	EmailRepository ports.EmailRepository
}

func NewRepositories() *Repositories {
	return &Repositories{
		EmailRepository: newEmailRepository(),
	}
}

func (r *Repositories) Email() ports.EmailRepository {
	return r.EmailRepository
}
