package memory

import "github.com/popeskul/email-service-platform/logger"

type RepositoryContainer struct {
	email EmailRepository
}

func NewRepositories(logger logger.Logger) *RepositoryContainer {
	return &RepositoryContainer{
		email: newEmailRepository(logger),
	}
}

func (r *RepositoryContainer) Email() EmailRepository {
	return r.email
}
