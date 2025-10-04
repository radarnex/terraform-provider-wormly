package client

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockHostAPI is a mock implementation of the HostAPI interface.
type MockHostAPI struct {
	mock.Mock
}

// CreateHost mocks the CreateHost method.
func (m *MockHostAPI) CreateHost(ctx context.Context, name string, testInterval int, enabled bool) (*Host, error) {
	args := m.Called(ctx, name, testInterval, enabled)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	if host, ok := args.Get(0).(*Host); ok {
		return host, args.Error(1)
	}
	return nil, args.Error(1)
}

// GetHost mocks the GetHost method.
func (m *MockHostAPI) GetHost(ctx context.Context, id int) (*Host, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	if host, ok := args.Get(0).(*Host); ok {
		return host, args.Error(1)
	}
	return nil, args.Error(1)
}

// DeleteHost mocks the DeleteHost method.
func (m *MockHostAPI) DeleteHost(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// DisableHostUptimeMonitoring mocks the DisableHostUptimeMonitoring method.
func (m *MockHostAPI) DisableHostUptimeMonitoring(ctx context.Context, hostID int) error {
	args := m.Called(ctx, hostID)
	return args.Error(0)
}

// EnableHostUptimeMonitoring mocks the EnableHostUptimeMonitoring method.
func (m *MockHostAPI) EnableHostUptimeMonitoring(ctx context.Context, hostID int) error {
	args := m.Called(ctx, hostID)
	return args.Error(0)
}
