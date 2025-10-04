package client

import (
	"context"
	"fmt"
)

// GlobalAlertMuteResponse represents the API response for setGlobalAlertMute.
type GlobalAlertMuteResponse struct {
	ErrorCode int `json:"errorcode"`
}

// GlobalAlertMuteAPI defines the interface for global alert mute operations.
type GlobalAlertMuteAPI interface {
	SetGlobalAlertMute(ctx context.Context, enabled bool) error
}

// Ensure Client implements GlobalAlertMuteAPI.
var _ GlobalAlertMuteAPI = (*Client)(nil)

// SetGlobalAlertMute sets the global alert mute status.
func (c *Client) SetGlobalAlertMute(ctx context.Context, enabled bool) error {
	alertsMuted := "0"
	if enabled {
		alertsMuted = "1"
	}

	params := map[string]string{
		"alertsmuted": alertsMuted,
	}

	var response GlobalAlertMuteResponse
	if err := c.makeFormRequest(ctx, "setGlobalAlertMute", params, &response); err != nil {
		return fmt.Errorf("failed to set global alert mute: %w", err)
	}

	if response.ErrorCode != 0 {
		return fmt.Errorf("API returned error code %d", response.ErrorCode)
	}

	return nil
}
