//go:generate go run go.uber.org/mock/mockgen -destination=mocks/mock_user_repository.go -package=mocks github.com/popeskul/mailflow/user-service/internal/domain UserRepository
//go:generate go run go.uber.org/mock/mockgen -destination=mocks/mock_email_client.go -package=mocks github.com/popeskul/mailflow/email-service/pkg/api/email/v1 EmailServiceClient
//go:generate go run go.uber.org/mock/mockgen -destination=mocks/mock_queue.go -package=mocks github.com/popeskul/mailflow/user-service/internal/queue Queue

package services
