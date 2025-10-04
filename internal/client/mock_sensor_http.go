package client

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockSensorHTTPAPI is a mock implementation of SensorHTTPAPI for testing.
type MockSensorHTTPAPI struct {
	mock.Mock
}

func (m *MockSensorHTTPAPI) CreateSensorHTTP(ctx context.Context, req *SensorHTTPCreateRequest) (*SensorHTTP, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	if sensor, ok := args.Get(0).(*SensorHTTP); ok {
		return sensor, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSensorHTTPAPI) GetSensorHTTP(ctx context.Context, hostID, sensorID int) (*SensorHTTP, error) {
	args := m.Called(ctx, hostID, sensorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	if sensor, ok := args.Get(0).(*SensorHTTP); ok {
		return sensor, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSensorHTTPAPI) DeleteSensorHTTP(ctx context.Context, sensorID int) error {
	args := m.Called(ctx, sensorID)
	return args.Error(0)
}

func (m *MockSensorHTTPAPI) ListSensorHTTP(ctx context.Context, hostID int) ([]*SensorHTTP, error) {
	args := m.Called(ctx, hostID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	if sensors, ok := args.Get(0).([]*SensorHTTP); ok {
		return sensors, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockSensorHTTPAPI) EnableSensorHTTP(ctx context.Context, hsid int) error {
	args := m.Called(ctx, hsid)
	return args.Error(0)
}

func (m *MockSensorHTTPAPI) DisableSensorHTTP(ctx context.Context, hsid int) error {
	args := m.Called(ctx, hsid)
	return args.Error(0)
}
