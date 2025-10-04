package client

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

// Host represents a Wormly host.
type Host struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	TestInterval int       `json:"test_interval"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// WormlyHostResponse represents the API response for host operations.
type WormlyHostResponse struct {
	ErrorCode int    `json:"errorcode"`
	Message   string `json:"message,omitempty"`
	HostID    int    `json:"hostid,omitempty"`
	Data      struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		Enabled bool   `json:"enabled"`
	} `json:"data,omitempty"`
}

// WormlyHostStatusResponse represents the API response for getHostStatus.
type WormlyHostStatusResponse struct {
	ErrorCode int `json:"errorcode"`
	Status    []struct {
		HostID          int    `json:"hostid"`
		Name            string `json:"name"`
		UptimeMonitored bool   `json:"uptimemonitored"`
		HealthMonitored bool   `json:"healthmonitored"`
		UptimeErrors    bool   `json:"uptimeerrors"`
		HealthErrors    bool   `json:"healtherrors"`
		LastUptimeCheck *int64 `json:"lastuptimecheck"` // Can be null, -1, or timestamp
		LastHealthCheck *int64 `json:"lasthealthcheck"` // Can be null, -1, or timestamp
		LastUptimeError *int64 `json:"lastuptimeerror"` // Can be null, -1, or timestamp
	} `json:"status"`
}

// HostAPI defines the interface for host-related operations.
type HostAPI interface {
	CreateHost(ctx context.Context, name string, testInterval int, enabled bool) (*Host, error)
	GetHost(ctx context.Context, id int) (*Host, error)
	DeleteHost(ctx context.Context, id int) error
	DisableHostUptimeMonitoring(ctx context.Context, hostID int) error
	EnableHostUptimeMonitoring(ctx context.Context, hostID int) error
}

// Ensure Client implements HostAPI.
var _ HostAPI = (*Client)(nil)

// CreateHost creates a new host.
func (c *Client) CreateHost(ctx context.Context, name string, testInterval int, enabled bool) (*Host, error) {
	params := map[string]string{
		"name":         name,
		"testinterval": strconv.Itoa(testInterval),
	}

	// Note: The Wormly API doesn't support an 'enabled' parameter in createHost.
	// The 'enabled' state is managed through disable/enable monitoring API calls at the provider level.

	var response WormlyHostResponse
	if err := c.makeFormRequest(ctx, "createHost", params, &response); err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	if response.ErrorCode != 0 {
		c.DebugLog("CreateHost API error response: %+v", response)
		return nil, fmt.Errorf("API returned error code %d: %s", response.ErrorCode, response.Message)
	}

	return &Host{
		ID:           response.HostID,
		Name:         name,
		TestInterval: testInterval,
		Enabled:      enabled,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

// GetHost retrieves a host by ID.
func (c *Client) GetHost(ctx context.Context, id int) (*Host, error) {
	params := map[string]string{
		"hostid": strconv.Itoa(id),
	}

	var response WormlyHostStatusResponse
	if err := c.makeFormRequest(ctx, "getHostStatus", params, &response); err != nil {
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	if response.ErrorCode != 0 {
		return nil, fmt.Errorf("API returned error code %d", response.ErrorCode)
	}

	if len(response.Status) == 0 {
		return nil, fmt.Errorf("host with ID %d not found", id)
	}

	// Find the host with the matching ID
	for _, host := range response.Status {
		if host.HostID == id {
			return &Host{
				ID:           host.HostID,
				Name:         host.Name,
				TestInterval: 60,                                           // Default value, API doesn't return this in getHostStatus
				Enabled:      host.UptimeMonitored || host.HealthMonitored, // Consider host enabled if either monitoring is active
				CreatedAt:    time.Now(),                                   // API doesn't return timestamps
				UpdatedAt:    time.Now(),                                   // API doesn't return timestamps
			}, nil
		}
	}

	return nil, fmt.Errorf("host with ID %d not found", id)
}

// DeleteHost deletes a host by ID.
func (c *Client) DeleteHost(ctx context.Context, id int) error {
	params := map[string]string{
		"hostid": strconv.Itoa(id),
	}

	var response WormlyHostResponse
	if err := c.makeFormRequest(ctx, "deleteHost", params, &response); err != nil {
		return fmt.Errorf("failed to delete host: %w", err)
	}

	if response.ErrorCode != 0 {
		return fmt.Errorf("API returned error code %d: %s", response.ErrorCode, response.Message)
	}

	return nil
}

// DisableHostUptimeMonitoring disables uptime monitoring for a host.
func (c *Client) DisableHostUptimeMonitoring(ctx context.Context, hostID int) error {
	params := map[string]string{
		"hostid": strconv.Itoa(hostID),
	}

	var response WormlyHostResponse
	if err := c.makeFormRequest(ctx, "disableHostUptimeMonitoring", params, &response); err != nil {
		return fmt.Errorf("failed to disable host uptime monitoring: %w", err)
	}

	if response.ErrorCode != 0 {
		return fmt.Errorf("API returned error code %d: %s", response.ErrorCode, response.Message)
	}

	return nil
}

// EnableHostUptimeMonitoring enables uptime monitoring for a host.
func (c *Client) EnableHostUptimeMonitoring(ctx context.Context, hostID int) error {
	params := map[string]string{
		"hostid": strconv.Itoa(hostID),
	}

	var response WormlyHostResponse
	if err := c.makeFormRequest(ctx, "enableHostUptimeMonitoring", params, &response); err != nil {
		return fmt.Errorf("failed to enable host uptime monitoring: %w", err)
	}

	if response.ErrorCode != 0 {
		return fmt.Errorf("API returned error code %d: %s", response.ErrorCode, response.Message)
	}

	return nil
}
