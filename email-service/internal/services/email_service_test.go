package services

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/popeskul/mailflow/common/logger"
	"github.com/popeskul/mailflow/email-service/internal/domain"
	"github.com/popeskul/mailflow/email-service/internal/services/mocks"
)

func createTestLogger() logger.Logger {
	return logger.NewZapLogger(logger.WithOutputs(io.Discard))
}

func createTestEmailService(repo EmailRepository, sender EmailSender, limiter Limiter, metrics Metrics) *emailService {
	service := &emailService{
		repo:        repo,
		sender:      sender,
		rateLimiter: limiter,
		metrics:     metrics,
		retryQueue:  make(chan *domain.Email, 1000),
		logger:      createTestLogger().Named("email_service"),
	}

	// Initialize nil dependencies with no-op implementations for testing
	if service.metrics == nil {
		service.metrics = &noOpMetrics{}
	}

	return service
}

// noOpMetrics provides a no-op implementation for testing
type noOpMetrics struct{}

func (n *noOpMetrics) RecordEmailSent()                           {}
func (n *noOpMetrics) RecordEmailQueued()                         {}
func (n *noOpMetrics) RecordEmailFailed()                         {}
func (n *noOpMetrics) RecordRateLimitDelay()                      {}
func (n *noOpMetrics) RecordDowntimePeriod()                      {}
func (n *noOpMetrics) SetQueueSize(size int)                      {}
func (n *noOpMetrics) ObserveProcessingDuration(duration float64) {}

func TestNewEmailService_Success(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "create email service with all dependencies",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockEmailRepository(ctrl)
			sender := mocks.NewMockEmailSender(ctrl)
			limiter := mocks.NewMockLimiter(ctrl)
			metrics := mocks.NewMockMetrics(ctrl)

			service := createTestEmailService(repo, sender, limiter, metrics)

			assert.NotNil(t, service)
		})
	}
}

func TestEmailService_SendEmail_Success(t *testing.T) {
	tests := []struct {
		name       string
		to         string
		subject    string
		body       string
		setupMocks func(*mocks.MockEmailRepository, *mocks.MockEmailSender, *mocks.MockLimiter, *mocks.MockMetrics)
	}{
		{
			name:    "send email successfully",
			to:      "test@example.com",
			subject: "Test Subject",
			body:    "Test Body",
			setupMocks: func(repo *mocks.MockEmailRepository, sender *mocks.MockEmailSender, limiter *mocks.MockLimiter, metrics *mocks.MockMetrics) {
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
				limiter.EXPECT().Wait(gomock.Any()).Return(nil)
				sender.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)
				metrics.EXPECT().RecordEmailSent()
				repo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name:    "send email with rate limiting delay",
			to:      "test@example.com",
			subject: "Test Subject",
			body:    "Test Body",
			setupMocks: func(repo *mocks.MockEmailRepository, sender *mocks.MockEmailSender, limiter *mocks.MockLimiter, metrics *mocks.MockMetrics) {
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
				limiter.EXPECT().Wait(gomock.Any()).Return(context.DeadlineExceeded)
				metrics.EXPECT().RecordRateLimitDelay()
				metrics.EXPECT().RecordEmailQueued()
				repo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name:    "send email with sender failure",
			to:      "test@example.com",
			subject: "Test Subject",
			body:    "Test Body",
			setupMocks: func(repo *mocks.MockEmailRepository, sender *mocks.MockEmailSender, limiter *mocks.MockLimiter, metrics *mocks.MockMetrics) {
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
				limiter.EXPECT().Wait(gomock.Any()).Return(nil)
				sender.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("send error"))
				metrics.EXPECT().RecordEmailFailed()
				metrics.EXPECT().RecordEmailQueued()
				repo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockEmailRepository(ctrl)
			sender := mocks.NewMockEmailSender(ctrl)
			limiter := mocks.NewMockLimiter(ctrl)
			metrics := mocks.NewMockMetrics(ctrl)

			tt.setupMocks(repo, sender, limiter, metrics)

			service := createTestEmailService(repo, sender, limiter, metrics)

			email, err := service.SendEmail(context.Background(), tt.to, tt.subject, tt.body)

			assert.NoError(t, err)
			assert.NotNil(t, email)
			assert.Equal(t, tt.to, email.To)
			assert.Equal(t, tt.subject, email.Subject)
			assert.Equal(t, tt.body, email.Body)
		})
	}
}

func TestEmailService_SendEmail_Fail(t *testing.T) {
	tests := []struct {
		name          string
		to            string
		subject       string
		body          string
		setupMocks    func(*mocks.MockEmailRepository, *mocks.MockEmailSender, *mocks.MockLimiter, *mocks.MockMetrics)
		expectedError string
	}{
		{
			name:    "repository save failure",
			to:      "test@example.com",
			subject: "Test Subject",
			body:    "Test Body",
			setupMocks: func(repo *mocks.MockEmailRepository, sender *mocks.MockEmailSender, limiter *mocks.MockLimiter, metrics *mocks.MockMetrics) {
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(errors.New("database error"))
			},
			expectedError: "failed to save email",
		},
		{
			name:    "update status failure after successful send",
			to:      "test@example.com",
			subject: "Test Subject",
			body:    "Test Body",
			setupMocks: func(repo *mocks.MockEmailRepository, sender *mocks.MockEmailSender, limiter *mocks.MockLimiter, metrics *mocks.MockMetrics) {
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
				limiter.EXPECT().Wait(gomock.Any()).Return(nil)
				sender.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)
				metrics.EXPECT().RecordEmailSent()
				repo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("update failed"))
			},
			expectedError: "failed to update email status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockEmailRepository(ctrl)
			sender := mocks.NewMockEmailSender(ctrl)
			limiter := mocks.NewMockLimiter(ctrl)
			metrics := mocks.NewMockMetrics(ctrl)

			tt.setupMocks(repo, sender, limiter, metrics)

			service := createTestEmailService(repo, sender, limiter, metrics)

			email, err := service.SendEmail(context.Background(), tt.to, tt.subject, tt.body)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
			if tt.expectedError == "failed to save email" {
				assert.Nil(t, email)
			}
		})
	}
}

func TestEmailService_GetEmailStatus_Success(t *testing.T) {
	tests := []struct {
		name          string
		emailID       string
		expectedEmail *domain.Email
	}{
		{
			name:    "get email status successfully",
			emailID: "email-123",
			expectedEmail: &domain.Email{
				ID:      "email-123",
				To:      "test@example.com",
				Subject: "Test",
				Body:    "Test body",
				Status:  domain.StatusSent,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockEmailRepository(ctrl)
			repo.EXPECT().GetByID(gomock.Any(), tt.emailID).Return(tt.expectedEmail, nil)

			service := createTestEmailService(repo, nil, nil, nil)

			email, err := service.GetEmailStatus(context.Background(), tt.emailID)

			assert.NoError(t, err)
			assert.NotNil(t, email)
			assert.Equal(t, tt.expectedEmail.ID, email.ID)
			assert.Equal(t, tt.expectedEmail.To, email.To)
			assert.Equal(t, tt.expectedEmail.Status, email.Status)
		})
	}
}

func TestEmailService_GetEmailStatus_Fail(t *testing.T) {
	tests := []struct {
		name          string
		emailID       string
		expectedError string
	}{
		{
			name:          "email not found",
			emailID:       "nonexistent",
			expectedError: "failed to get email status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockEmailRepository(ctrl)
			repo.EXPECT().GetByID(gomock.Any(), tt.emailID).Return(nil, errors.New("email not found"))

			service := createTestEmailService(repo, nil, nil, nil)

			email, err := service.GetEmailStatus(context.Background(), tt.emailID)

			assert.Error(t, err)
			assert.Nil(t, email)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestEmailService_ListEmails_Success(t *testing.T) {
	tests := []struct {
		name              string
		pageSize          int
		pageToken         string
		expectedEmails    []*domain.Email
		expectedNextToken string
	}{
		{
			name:      "list emails successfully",
			pageSize:  10,
			pageToken: "",
			expectedEmails: []*domain.Email{
				{ID: "1", To: "test1@example.com"},
				{ID: "2", To: "test2@example.com"},
			},
			expectedNextToken: "next_token",
		},
		{
			name:              "list empty emails",
			pageSize:          10,
			pageToken:         "",
			expectedEmails:    []*domain.Email{},
			expectedNextToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockEmailRepository(ctrl)
			repo.EXPECT().List(gomock.Any(), tt.pageSize, tt.pageToken).Return(tt.expectedEmails, tt.expectedNextToken, nil)

			service := createTestEmailService(repo, nil, nil, nil)

			emails, nextToken, err := service.ListEmails(context.Background(), tt.pageSize, tt.pageToken)

			assert.NoError(t, err)
			assert.NotNil(t, emails)
			assert.Equal(t, len(tt.expectedEmails), len(emails))
			assert.Equal(t, tt.expectedNextToken, nextToken)
		})
	}
}

func TestEmailService_ListEmails_Fail(t *testing.T) {
	tests := []struct {
		name          string
		pageSize      int
		pageToken     string
		expectedError string
	}{
		{
			name:          "repository list failure",
			pageSize:      10,
			pageToken:     "",
			expectedError: "failed to list emails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockEmailRepository(ctrl)
			repo.EXPECT().List(gomock.Any(), tt.pageSize, tt.pageToken).Return(nil, "", errors.New("database error"))

			service := createTestEmailService(repo, nil, nil, nil)

			emails, nextToken, err := service.ListEmails(context.Background(), tt.pageSize, tt.pageToken)

			assert.Error(t, err)
			assert.Nil(t, emails)
			assert.Empty(t, nextToken)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestEmailService_ResendFailedEmails_Success(t *testing.T) {
	tests := []struct {
		name         string
		failedEmails []*domain.Email
	}{
		{
			name: "resend failed emails successfully",
			failedEmails: []*domain.Email{
				{ID: "1", Status: domain.StatusFailed},
				{ID: "2", Status: domain.StatusFailed},
				{ID: "3", Status: domain.StatusSent},
			},
		},
		{
			name: "no failed emails to resend",
			failedEmails: []*domain.Email{
				{ID: "1", Status: domain.StatusSent},
				{ID: "2", Status: domain.StatusSent},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockEmailRepository(ctrl)
			repo.EXPECT().List(gomock.Any(), 0, "").Return(tt.failedEmails, "", nil)

			// Add expectations for queueForRetry if there are failed emails
			failedCount := 0
			for _, email := range tt.failedEmails {
				if email.Status == domain.StatusFailed {
					failedCount++
				}
			}
			if failedCount > 0 {
				repo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), domain.StatusPending, nil).Times(failedCount)
			}

			service := createTestEmailService(repo, nil, nil, nil)
			service.retryQueue = make(chan *domain.Email, 100)

			err := service.ResendFailedEmails(context.Background())

			assert.NoError(t, err)
		})
	}
}

func TestEmailService_ResendFailedEmails_Fail(t *testing.T) {
	tests := []struct {
		name          string
		expectedError string
	}{
		{
			name:          "repository list failure",
			expectedError: "failed to list emails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockEmailRepository(ctrl)
			repo.EXPECT().List(gomock.Any(), 0, "").Return(nil, "", errors.New("database error"))

			service := createTestEmailService(repo, nil, nil, nil)

			err := service.ResendFailedEmails(context.Background())

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestEmailService_ProcessQueue_Success(t *testing.T) {
	tests := []struct {
		name        string
		queueEmails []*domain.Email
	}{
		{
			name: "process queue with emails",
			queueEmails: []*domain.Email{
				{ID: "1", To: "test1@example.com", Status: domain.StatusPending},
				{ID: "2", To: "test2@example.com", Status: domain.StatusPending},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockEmailRepository(ctrl)
			sender := mocks.NewMockEmailSender(ctrl)
			limiter := mocks.NewMockLimiter(ctrl)
			metrics := mocks.NewMockMetrics(ctrl)

			service := createTestEmailService(repo, sender, limiter, metrics)
			emailSvc := service

			for _, email := range tt.queueEmails {
				emailSvc.retryQueue <- email
			}
			close(emailSvc.retryQueue)

			limiter.EXPECT().Wait(gomock.Any()).Return(nil).AnyTimes()
			sender.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			metrics.EXPECT().RecordEmailSent().AnyTimes()
			repo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

			emailSvc.processRetryQueue()

			assert.Equal(t, 0, len(emailSvc.retryQueue))
		})
	}
}

func TestEmailService_QueueForRetry_Success(t *testing.T) {
	tests := []struct {
		name  string
		email *domain.Email
	}{
		{
			name: "queue email for retry",
			email: &domain.Email{
				ID:     "test-email",
				To:     "test@example.com",
				Status: domain.StatusFailed,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockEmailRepository(ctrl)
			sender := mocks.NewMockEmailSender(ctrl)
			limiter := mocks.NewMockLimiter(ctrl)
			metrics := mocks.NewMockMetrics(ctrl)

			metrics.EXPECT().RecordEmailQueued()
			metrics.EXPECT().SetQueueSize(gomock.Any()).AnyTimes()
			repo.EXPECT().UpdateStatus(gomock.Any(), tt.email.ID, domain.StatusPending, nil).Return(nil)

			service := createTestEmailService(repo, sender, limiter, metrics)
			emailSvc := service

			emailSvc.queueForRetry(tt.email)

			assert.Equal(t, domain.StatusPending, tt.email.Status)
			assert.Equal(t, 1, len(emailSvc.retryQueue))
		})
	}
}

func TestEmailService_QueueForRetry_QueueFull_Fail(t *testing.T) {
	tests := []struct {
		name  string
		email *domain.Email
	}{
		{
			name: "queue full marks email as failed",
			email: &domain.Email{
				ID:     "test-email",
				To:     "test@example.com",
				Status: domain.StatusFailed,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockEmailRepository(ctrl)
			sender := mocks.NewMockEmailSender(ctrl)
			limiter := mocks.NewMockLimiter(ctrl)
			metrics := mocks.NewMockMetrics(ctrl)

			metrics.EXPECT().RecordEmailQueued()
			metrics.EXPECT().RecordEmailFailed()
			metrics.EXPECT().SetQueueSize(gomock.Any()).AnyTimes()
			repo.EXPECT().UpdateStatus(gomock.Any(), tt.email.ID, domain.StatusFailed, nil).Return(nil)

			service := createTestEmailService(repo, sender, limiter, metrics)
			emailSvc := service

			emailSvc.retryQueue = make(chan *domain.Email)

			emailSvc.queueForRetry(tt.email)

			assert.Equal(t, domain.StatusFailed, tt.email.Status)
		})
	}
}
