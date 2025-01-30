package services_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEmailService_SendEmail(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*mockRepo)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(repo *mockRepo) {
				repo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Email")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "fail - service down",
			setup: func(repo *mockRepo) {
				repo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Email")).Return(nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepo)
			tt.setup(repo)

			svc := NewEmailService(repo, nil, nil, nil)

			_, err := svc.SendEmail(context.Background(), "test@test.com", "test", "test")

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}
