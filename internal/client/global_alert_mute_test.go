package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_SetGlobalAlertMute_Enable(t *testing.T) {
	mockClient := &MockGlobalAlertMuteAPI{}
	mockClient.On("SetGlobalAlertMute", mock.Anything, true).Return(nil)

	err := mockClient.SetGlobalAlertMute(t.Context(), true)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestClient_SetGlobalAlertMute_Disable(t *testing.T) {
	mockClient := &MockGlobalAlertMuteAPI{}
	mockClient.On("SetGlobalAlertMute", mock.Anything, false).Return(nil)

	err := mockClient.SetGlobalAlertMute(t.Context(), false)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}
