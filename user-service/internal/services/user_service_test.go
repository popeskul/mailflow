package services

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/popeskul/mailflow/common/logger"
	emailv1 "github.com/popeskul/mailflow/email-service/pkg/api/email/v1"
	"github.com/popeskul/mailflow/user-service/internal/circuitbreaker"
	"github.com/popeskul/mailflow/user-service/internal/domain"
	"github.com/popeskul/mailflow/user-service/internal/queue"
	"github.com/popeskul/mailflow/user-service/internal/services/mocks"
)

func createTestLogger() logger.Logger {
	return logger.NewZapLogger(logger.WithOutputs(io.Discard))
}

func TestNewUserService_Success(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "create user service with valid dependencies",
		},
		{
			name: "create user service with nil email client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockUserRepository(ctrl)
			var emailClient emailv1.EmailServiceClient

			if tt.name != "create user service with nil email client" {
				emailClient = mocks.NewMockEmailServiceClient(ctrl)
			}

			service := NewUserService(repo, emailClient, createTestLogger())

			assert.NotNil(t, service)
		})
	}
}

func TestNewUserServiceWithWrapper_Success(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "create user service with email wrapper",
		},
		{
			name: "create user service with nil wrapper",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockUserRepository(ctrl)
			var wrapper *EmailClientWrapper

			if tt.name != "create user service with nil wrapper" {
				emailClient := mocks.NewMockEmailServiceClient(ctrl)
				cb := circuitbreaker.New(circuitbreaker.DefaultConfig())
				q := queue.NewEmailQueue(100, zap.NewNop())
				wrapper = NewEmailClientWrapper(emailClient, cb, q, createTestLogger())
			}

			service := NewUserServiceWithWrapper(repo, wrapper, createTestLogger())

			assert.NotNil(t, service)
		})
	}
}

func TestUserService_Create_Success(t *testing.T) {
	tests := []struct {
		name            string
		email           string
		userName        string
		withEmailClient bool
		withWrapper     bool
	}{
		{
			name:            "create user successfully with email client",
			email:           "test@example.com",
			userName:        "Test User",
			withEmailClient: true,
			withWrapper:     false,
		},
		{
			name:            "create user successfully without email client",
			email:           "test@example.com",
			userName:        "Test User",
			withEmailClient: false,
			withWrapper:     false,
		},
		{
			name:            "create user successfully with wrapper",
			email:           "test@example.com",
			userName:        "Test User",
			withEmailClient: false,
			withWrapper:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockUserRepository(ctrl)
			repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

			var service *UserService

			if tt.withEmailClient {
				emailClient := mocks.NewMockEmailServiceClient(ctrl)
				emailClient.EXPECT().SendEmail(gomock.Any(), gomock.Any()).Return(&emailv1.SendEmailResponse{}, nil)
				service = NewUserService(repo, emailClient, createTestLogger())
			} else if tt.withWrapper {
				emailClient := mocks.NewMockEmailServiceClient(ctrl)
				emailClient.EXPECT().SendEmail(gomock.Any(), gomock.Any()).Return(&emailv1.SendEmailResponse{}, nil)
				cb := circuitbreaker.New(circuitbreaker.DefaultConfig())
				q := queue.NewEmailQueue(100, zap.NewNop())
				wrapper := NewEmailClientWrapper(emailClient, cb, q, createTestLogger())
				service = NewUserServiceWithWrapper(repo, wrapper, createTestLogger())
			} else {
				service = NewUserService(repo, nil, createTestLogger())
			}

			user, err := service.Create(context.Background(), tt.email, tt.userName)

			assert.NoError(t, err)
			assert.NotNil(t, user)
			assert.Equal(t, tt.email, user.Email)
			assert.Equal(t, tt.userName, user.Name)
			assert.NotEmpty(t, user.ID)
		})
	}
}

func TestUserService_Create_Fail(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		userName      string
		expectedError string
	}{
		{
			name:          "repository create failure",
			email:         "test@example.com",
			userName:      "Test User",
			expectedError: "failed to create user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockUserRepository(ctrl)
			repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("database error"))

			service := NewUserService(repo, nil, createTestLogger())

			user, err := service.Create(context.Background(), tt.email, tt.userName)

			assert.Error(t, err)
			assert.Nil(t, user)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestUserService_Get_Success(t *testing.T) {
	tests := []struct {
		name         string
		userID       string
		expectedUser *domain.User
	}{
		{
			name:   "get user successfully",
			userID: "user-123",
			expectedUser: &domain.User{
				ID:    "user-123",
				Email: "test@example.com",
				Name:  "Test User",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockUserRepository(ctrl)
			repo.EXPECT().GetByID(gomock.Any(), tt.userID).Return(tt.expectedUser, nil)

			service := NewUserService(repo, nil, createTestLogger())

			user, err := service.Get(context.Background(), tt.userID)

			assert.NoError(t, err)
			assert.NotNil(t, user)
			assert.Equal(t, tt.expectedUser.ID, user.ID)
			assert.Equal(t, tt.expectedUser.Email, user.Email)
			assert.Equal(t, tt.expectedUser.Name, user.Name)
		})
	}
}

func TestUserService_Get_Fail(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		expectedError string
	}{
		{
			name:          "user not found",
			userID:        "nonexistent",
			expectedError: "failed to get user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockUserRepository(ctrl)
			repo.EXPECT().GetByID(gomock.Any(), tt.userID).Return(nil, errors.New("user not found"))

			service := NewUserService(repo, nil, createTestLogger())

			user, err := service.Get(context.Background(), tt.userID)

			assert.Error(t, err)
			assert.Nil(t, user)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestUserService_Update_Success(t *testing.T) {
	tests := []struct {
		name         string
		userID       string
		email        string
		userName     string
		existingUser *domain.User
	}{
		{
			name:     "update user successfully",
			userID:   "user-123",
			email:    "updated@example.com",
			userName: "Updated User",
			existingUser: &domain.User{
				ID:    "user-123",
				Email: "old@example.com",
				Name:  "Old User",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockUserRepository(ctrl)
			repo.EXPECT().GetByID(gomock.Any(), tt.userID).Return(tt.existingUser, nil)
			repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

			service := NewUserService(repo, nil, createTestLogger())

			user, err := service.Update(context.Background(), tt.userID, tt.email, tt.userName)

			assert.NoError(t, err)
			assert.NotNil(t, user)
			assert.Equal(t, tt.email, user.Email)
			assert.Equal(t, tt.userName, user.Name)
		})
	}
}

func TestUserService_Update_Fail(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		email         string
		userName      string
		setupMocks    func(*mocks.MockUserRepository)
		expectedError string
	}{
		{
			name:     "user not found",
			userID:   "nonexistent",
			email:    "test@example.com",
			userName: "Test User",
			setupMocks: func(repo *mocks.MockUserRepository) {
				repo.EXPECT().GetByID(gomock.Any(), "nonexistent").Return(nil, errors.New("user not found"))
			},
			expectedError: "failed to get user",
		},
		{
			name:     "update failure",
			userID:   "user-123",
			email:    "test@example.com",
			userName: "Test User",
			setupMocks: func(repo *mocks.MockUserRepository) {
				existingUser := &domain.User{ID: "user-123", Email: "old@example.com", Name: "Old User"}
				repo.EXPECT().GetByID(gomock.Any(), "user-123").Return(existingUser, nil)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("update failed"))
			},
			expectedError: "failed to update user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockUserRepository(ctrl)
			tt.setupMocks(repo)

			service := NewUserService(repo, nil, createTestLogger())

			user, err := service.Update(context.Background(), tt.userID, tt.email, tt.userName)

			assert.Error(t, err)
			assert.Nil(t, user)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestUserService_Delete_Success(t *testing.T) {
	tests := []struct {
		name   string
		userID string
	}{
		{
			name:   "delete user successfully",
			userID: "user-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockUserRepository(ctrl)
			repo.EXPECT().Delete(gomock.Any(), tt.userID).Return(nil)

			service := NewUserService(repo, nil, createTestLogger())

			err := service.Delete(context.Background(), tt.userID)

			assert.NoError(t, err)
		})
	}
}

func TestUserService_Delete_Fail(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		expectedError string
	}{
		{
			name:          "delete failure",
			userID:        "user-123",
			expectedError: "failed to delete user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockUserRepository(ctrl)
			repo.EXPECT().Delete(gomock.Any(), tt.userID).Return(errors.New("delete failed"))

			service := NewUserService(repo, nil, createTestLogger())

			err := service.Delete(context.Background(), tt.userID)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestUserService_List_Success(t *testing.T) {
	tests := []struct {
		name              string
		pageSize          int
		pageToken         string
		expectedUsers     []*domain.User
		expectedNextToken string
	}{
		{
			name:      "list users successfully",
			pageSize:  10,
			pageToken: "",
			expectedUsers: []*domain.User{
				{ID: "1", Email: "test1@example.com", Name: "User 1"},
				{ID: "2", Email: "test2@example.com", Name: "User 2"},
			},
			expectedNextToken: "next_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockUserRepository(ctrl)
			repo.EXPECT().List(gomock.Any(), tt.pageSize, tt.pageToken).Return(tt.expectedUsers, tt.expectedNextToken, nil)

			service := NewUserService(repo, nil, createTestLogger())

			users, nextToken, err := service.List(context.Background(), tt.pageSize, tt.pageToken)

			assert.NoError(t, err)
			assert.Equal(t, len(tt.expectedUsers), len(users))
			assert.Equal(t, tt.expectedNextToken, nextToken)
		})
	}
}

func TestUserService_List_Fail(t *testing.T) {
	tests := []struct {
		name          string
		pageSize      int
		pageToken     string
		expectedError string
	}{
		{
			name:          "list failure",
			pageSize:      10,
			pageToken:     "",
			expectedError: "failed to list users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockUserRepository(ctrl)
			repo.EXPECT().List(gomock.Any(), tt.pageSize, tt.pageToken).Return(nil, "", errors.New("list failed"))

			service := NewUserService(repo, nil, createTestLogger())

			users, nextToken, err := service.List(context.Background(), tt.pageSize, tt.pageToken)

			assert.Error(t, err)
			assert.Nil(t, users)
			assert.Empty(t, nextToken)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestEmailClientWrapper_SendEmail_Success(t *testing.T) {
	tests := []struct {
		name    string
		request *emailv1.SendEmailRequest
	}{
		{
			name: "send email successfully",
			request: &emailv1.SendEmailRequest{
				To:      "test@example.com",
				Subject: "Test Subject",
				Body:    "Test Body",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client := mocks.NewMockEmailServiceClient(ctrl)
			client.EXPECT().SendEmail(gomock.Any(), tt.request).Return(&emailv1.SendEmailResponse{}, nil)

			cb := circuitbreaker.New(circuitbreaker.DefaultConfig())
			q := queue.NewEmailQueue(100, zap.NewNop())

			wrapper := NewEmailClientWrapper(client, cb, q, createTestLogger())

			err := wrapper.SendEmail(context.Background(), tt.request)

			assert.NoError(t, err)
		})
	}
}

func TestEmailClientWrapper_SendEmail_Fail(t *testing.T) {
	tests := []struct {
		name        string
		request     *emailv1.SendEmailRequest
		setupMocks  func(*mocks.MockEmailServiceClient)
		shouldQueue bool
	}{
		{
			name: "service unavailable - should queue",
			request: &emailv1.SendEmailRequest{
				To:      "test@example.com",
				Subject: "Test Subject",
				Body:    "Test Body",
			},
			setupMocks: func(client *mocks.MockEmailServiceClient) {
				unavailableErr := status.Error(codes.Unavailable, "service unavailable")
				client.EXPECT().SendEmail(gomock.Any(), gomock.Any()).Return(nil, unavailableErr).AnyTimes()
			},
			shouldQueue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client := mocks.NewMockEmailServiceClient(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(client)
			}

			cb := circuitbreaker.New(&circuitbreaker.Config{
				FailureThreshold: 1,
				SuccessThreshold: 2,
				Timeout:          1,
				MaxRequests:      2,
			})
			q := queue.NewEmailQueue(100, zap.NewNop())

			wrapper := NewEmailClientWrapper(client, cb, q, createTestLogger())

			err := wrapper.SendEmail(context.Background(), tt.request)

			if tt.shouldQueue {
				assert.NoError(t, err) // Should queue successfully
			} else {
				assert.Error(t, err)
			}
		})
	}
}
