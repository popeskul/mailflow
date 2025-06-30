package memory

import (
	"github.com/popeskul/mailflow/common/logger"
	"github.com/popeskul/mailflow/email-service/internal/domain"
)

type Repositories struct {
	email domain.EmailRepository
}

func NewRepositories(logger logger.Logger) *Repositories {
	return &Repositories{
		email: newEmailRepository(logger),
	}
}

func (r *Repositories) Email() domain.EmailRepository {
	return r.email
}
