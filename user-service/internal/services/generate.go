package services

//go:generate go run go.uber.org/mock/mockgen -destination=mocks/mock_email_service_test.go -package=mocks github.com/popeskul/email-service-platform/email-service/internal/core/ports EmailService
