package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/popeskul/mailflow/common/logger"
)

func TestNewRepositories(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "create repositories successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger := logger.NewZapLogger()
			repos := NewRepositories(testLogger)

			assert.NotNil(t, repos)
			assert.NotNil(t, repos.User())
		})
	}
}

func TestRepositories_User(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "get user repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger := logger.NewZapLogger()
			repos := NewRepositories(testLogger)

			userRepo := repos.User()

			assert.NotNil(t, userRepo)
			// Verify we can call methods on the repository
			assert.NotPanics(t, func() {
				_, _, _ = userRepo.List(context.TODO(), 10, "")
			})
		})
	}
}
