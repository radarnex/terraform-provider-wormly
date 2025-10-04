package client

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockScheduledDowntimePeriodAPI is a mock implementation of the ScheduledDowntimePeriodAPI interface.
type MockScheduledDowntimePeriodAPI struct {
	mock.Mock
}

// CreateScheduledDowntimePeriod mocks the CreateScheduledDowntimePeriod method.
func (m *MockScheduledDowntimePeriodAPI) CreateScheduledDowntimePeriod(ctx context.Context, hostID int, start, end, timezone, recurrence, on string) (*ScheduledDowntimePeriod, error) {
	args := m.Called(ctx, hostID, start, end, timezone, recurrence, on)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	if period, ok := args.Get(0).(*ScheduledDowntimePeriod); ok {
		return period, args.Error(1)
	}
	return nil, args.Error(1)
}

// GetScheduledDowntimePeriod mocks the GetScheduledDowntimePeriod method.
func (m *MockScheduledDowntimePeriodAPI) GetScheduledDowntimePeriod(ctx context.Context, hostID, periodID int) (*ScheduledDowntimePeriod, error) {
	args := m.Called(ctx, hostID, periodID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	if period, ok := args.Get(0).(*ScheduledDowntimePeriod); ok {
		return period, args.Error(1)
	}
	return nil, args.Error(1)
}

// UpdateScheduledDowntimePeriod mocks the UpdateScheduledDowntimePeriod method.
func (m *MockScheduledDowntimePeriodAPI) UpdateScheduledDowntimePeriod(ctx context.Context, hostID, periodID int, start, end, timezone, recurrence, on string) (*ScheduledDowntimePeriod, error) {
	args := m.Called(ctx, hostID, periodID, start, end, timezone, recurrence, on)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	if period, ok := args.Get(0).(*ScheduledDowntimePeriod); ok {
		return period, args.Error(1)
	}
	return nil, args.Error(1)
}

// DeleteScheduledDowntimePeriod mocks the DeleteScheduledDowntimePeriod method.
func (m *MockScheduledDowntimePeriodAPI) DeleteScheduledDowntimePeriod(ctx context.Context, hostID, periodID int) error {
	args := m.Called(ctx, hostID, periodID)
	return args.Error(0)
}

// GetScheduledDowntimePeriods mocks the GetScheduledDowntimePeriods method.
func (m *MockScheduledDowntimePeriodAPI) GetScheduledDowntimePeriods(ctx context.Context, hostID int) ([]ScheduledDowntimePeriod, error) {
	args := m.Called(ctx, hostID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	if periods, ok := args.Get(0).([]ScheduledDowntimePeriod); ok {
		return periods, args.Error(1)
	}
	return nil, args.Error(1)
}
