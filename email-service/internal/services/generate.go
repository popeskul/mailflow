//go:generate go run go.uber.org/mock/mockgen -destination=mocks/mock_email_repository.go -package=mocks github.com/popeskul/mailflow/email-service/internal/services EmailRepository
//go:generate go run go.uber.org/mock/mockgen -destination=mocks/mock_email_sender.go -package=mocks github.com/popeskul/mailflow/email-service/internal/services EmailSender
//go:generate go run go.uber.org/mock/mockgen -destination=mocks/mock_limiter.go -package=mocks github.com/popeskul/mailflow/email-service/internal/services Limiter
//go:generate go run go.uber.org/mock/mockgen -destination=mocks/mock_metrics.go -package=mocks github.com/popeskul/mailflow/email-service/internal/services Metrics

package services
