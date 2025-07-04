// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/popeskul/mailflow/email-service/internal/services (interfaces: EmailRepository)
//
// Generated by this command:
//
//	mockgen -destination=mocks/mock_email_repository.go -package=mocks github.com/popeskul/mailflow/email-service/internal/services EmailRepository
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"
	time "time"

	domain "github.com/popeskul/mailflow/email-service/internal/domain"
	gomock "go.uber.org/mock/gomock"
)

// MockEmailRepository is a mock of EmailRepository interface.
type MockEmailRepository struct {
	ctrl     *gomock.Controller
	recorder *MockEmailRepositoryMockRecorder
	isgomock struct{}
}

// MockEmailRepositoryMockRecorder is the mock recorder for MockEmailRepository.
type MockEmailRepositoryMockRecorder struct {
	mock *MockEmailRepository
}

// NewMockEmailRepository creates a new mock instance.
func NewMockEmailRepository(ctrl *gomock.Controller) *MockEmailRepository {
	mock := &MockEmailRepository{ctrl: ctrl}
	mock.recorder = &MockEmailRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEmailRepository) EXPECT() *MockEmailRepositoryMockRecorder {
	return m.recorder
}

// GetByID mocks base method.
func (m *MockEmailRepository) GetByID(ctx context.Context, id string) (*domain.Email, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*domain.Email)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID.
func (mr *MockEmailRepositoryMockRecorder) GetByID(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockEmailRepository)(nil).GetByID), ctx, id)
}

// List mocks base method.
func (m *MockEmailRepository) List(ctx context.Context, pageSize int, pageToken string) ([]*domain.Email, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx, pageSize, pageToken)
	ret0, _ := ret[0].([]*domain.Email)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// List indicates an expected call of List.
func (mr *MockEmailRepositoryMockRecorder) List(ctx, pageSize, pageToken any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockEmailRepository)(nil).List), ctx, pageSize, pageToken)
}

// Save mocks base method.
func (m *MockEmailRepository) Save(ctx context.Context, email *domain.Email) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", ctx, email)
	ret0, _ := ret[0].(error)
	return ret0
}

// Save indicates an expected call of Save.
func (mr *MockEmailRepositoryMockRecorder) Save(ctx, email any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockEmailRepository)(nil).Save), ctx, email)
}

// UpdateStatus mocks base method.
func (m *MockEmailRepository) UpdateStatus(ctx context.Context, id, status string, sentAt *time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateStatus", ctx, id, status, sentAt)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateStatus indicates an expected call of UpdateStatus.
func (mr *MockEmailRepositoryMockRecorder) UpdateStatus(ctx, id, status, sentAt any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateStatus", reflect.TypeOf((*MockEmailRepository)(nil).UpdateStatus), ctx, id, status, sentAt)
}
