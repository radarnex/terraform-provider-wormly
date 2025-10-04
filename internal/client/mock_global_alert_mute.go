package client

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockGlobalAlertMuteAPI is a mock implementation of the GlobalAlertMuteAPI interface.
type MockGlobalAlertMuteAPI struct {
	mock.Mock
}

// SetGlobalAlertMute mocks the SetGlobalAlertMute method.
func (m *MockGlobalAlertMuteAPI) SetGlobalAlertMute(ctx context.Context, enabled bool) error {
	args := m.Called(ctx, enabled)
	return args.Error(0)
}
