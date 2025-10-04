package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

// ScheduledDowntimePeriod represents a Wormly scheduled downtime period.
type ScheduledDowntimePeriod struct {
	ID         int    `json:"periodid"`
	HostID     int    `json:"hostid"`
	Start      string `json:"start"`
	End        string `json:"end"`
	Timezone   string `json:"timezone"`
	Recurrence string `json:"recurrence"`
	On         string `json:"on,omitempty"`
}

// UnmarshalJSON implements custom JSON unmarshaling to handle periodid as string or int.
func (s *ScheduledDowntimePeriod) UnmarshalJSON(data []byte) error {
	// Define a temporary struct that accepts periodid as either string or int
	type Alias ScheduledDowntimePeriod
	aux := &struct {
		PeriodID interface{} `json:"periodid"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Handle periodid conversion
	switch v := aux.PeriodID.(type) {
	case string:
		if v != "" {
			id, err := strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("failed to convert periodid string '%s' to int: %w", v, err)
			}
			s.ID = id
		}
	case float64:
		s.ID = int(v)
	case int:
		s.ID = v
	case nil:
		s.ID = 0
	default:
		return fmt.Errorf("unexpected type for periodid: %T", v)
	}

	return nil
}

// WormlyScheduledDowntimePeriodResponse represents the API response for scheduled downtime period operations.
type WormlyScheduledDowntimePeriodResponse struct {
	ErrorCode int    `json:"errorcode"`
	Message   string `json:"message,omitempty"`
	PeriodID  int    `json:"periodid,omitempty"`
}

// WormlyGetScheduledDowntimePeriodsResponse represents the API response for getScheduledDowntimePeriods.
type WormlyGetScheduledDowntimePeriodsResponse struct {
	ErrorCode int                       `json:"errorcode"`
	Periods   []ScheduledDowntimePeriod `json:"periods"`
}

// ScheduledDowntimePeriodAPI defines the interface for scheduled downtime period-related operations.
type ScheduledDowntimePeriodAPI interface {
	CreateScheduledDowntimePeriod(ctx context.Context, hostID int, start, end, timezone, recurrence, on string) (*ScheduledDowntimePeriod, error)
	GetScheduledDowntimePeriod(ctx context.Context, hostID, periodID int) (*ScheduledDowntimePeriod, error)
	UpdateScheduledDowntimePeriod(ctx context.Context, hostID, periodID int, start, end, timezone, recurrence, on string) (*ScheduledDowntimePeriod, error)
	DeleteScheduledDowntimePeriod(ctx context.Context, hostID, periodID int) error
	GetScheduledDowntimePeriods(ctx context.Context, hostID int) ([]ScheduledDowntimePeriod, error)
}

// Ensure Client implements ScheduledDowntimePeriodAPI.
var _ ScheduledDowntimePeriodAPI = (*Client)(nil)

// CreateScheduledDowntimePeriod creates a new scheduled downtime period.
func (c *Client) CreateScheduledDowntimePeriod(ctx context.Context, hostID int, start, end, timezone, recurrence, on string) (*ScheduledDowntimePeriod, error) {
	params := map[string]string{
		"hostid":     strconv.Itoa(hostID),
		"start":      start,
		"end":        end,
		"timezone":   timezone,
		"recurrence": recurrence,
	}

	// Only include "on" parameter if it's not empty
	if on != "" {
		params["on"] = on
	}

	var response WormlyScheduledDowntimePeriodResponse
	if err := c.makeFormRequest(ctx, "setScheduledDowntimePeriod", params, &response); err != nil {
		return nil, fmt.Errorf("failed to create scheduled downtime period: %w", err)
	}

	if response.ErrorCode != 0 {
		c.DebugLog("CreateScheduledDowntimePeriod API error response: %+v", response)
		return nil, fmt.Errorf("API returned error code %d: %s", response.ErrorCode, response.Message)
	}

	return &ScheduledDowntimePeriod{
		ID:         response.PeriodID,
		HostID:     hostID,
		Start:      start,
		End:        end,
		Timezone:   timezone,
		Recurrence: recurrence,
		On:         on,
	}, nil
}

// GetScheduledDowntimePeriod retrieves a scheduled downtime period by host ID and period ID.
func (c *Client) GetScheduledDowntimePeriod(ctx context.Context, hostID, periodID int) (*ScheduledDowntimePeriod, error) {
	periods, err := c.GetScheduledDowntimePeriods(ctx, hostID)
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled downtime periods: %w", err)
	}

	// Find the period with the matching ID
	for _, period := range periods {
		if period.ID == periodID {
			return &period, nil
		}
	}

	return nil, fmt.Errorf("scheduled downtime period with ID %d not found", periodID)
}

// UpdateScheduledDowntimePeriod updates an existing scheduled downtime period.
func (c *Client) UpdateScheduledDowntimePeriod(ctx context.Context, hostID, periodID int, start, end, timezone, recurrence, on string) (*ScheduledDowntimePeriod, error) {
	params := map[string]string{
		"hostid":     strconv.Itoa(hostID),
		"periodid":   strconv.Itoa(periodID),
		"start":      start,
		"end":        end,
		"timezone":   timezone,
		"recurrence": recurrence,
	}

	// Only include "on" parameter if it's not empty
	if on != "" {
		params["on"] = on
	}

	var response WormlyScheduledDowntimePeriodResponse
	if err := c.makeFormRequest(ctx, "setScheduledDowntimePeriod", params, &response); err != nil {
		return nil, fmt.Errorf("failed to update scheduled downtime period: %w", err)
	}

	if response.ErrorCode != 0 {
		c.DebugLog("UpdateScheduledDowntimePeriod API error response: %+v", response)
		return nil, fmt.Errorf("API returned error code %d: %s", response.ErrorCode, response.Message)
	}

	return &ScheduledDowntimePeriod{
		ID:         response.PeriodID,
		HostID:     hostID,
		Start:      start,
		End:        end,
		Timezone:   timezone,
		Recurrence: recurrence,
		On:         on,
	}, nil
}

// DeleteScheduledDowntimePeriod deletes a scheduled downtime period.
func (c *Client) DeleteScheduledDowntimePeriod(ctx context.Context, hostID, periodID int) error {
	params := map[string]string{
		"hostid":   strconv.Itoa(hostID),
		"periodid": strconv.Itoa(periodID),
	}

	var response WormlyScheduledDowntimePeriodResponse
	if err := c.makeFormRequest(ctx, "deleteScheduledDowntimePeriod", params, &response); err != nil {
		return fmt.Errorf("failed to delete scheduled downtime period: %w", err)
	}

	if response.ErrorCode != 0 {
		return fmt.Errorf("API returned error code %d: %s", response.ErrorCode, response.Message)
	}

	return nil
}

// GetScheduledDowntimePeriods retrieves all scheduled downtime periods for a host.
func (c *Client) GetScheduledDowntimePeriods(ctx context.Context, hostID int) ([]ScheduledDowntimePeriod, error) {
	params := map[string]string{
		"hostid": strconv.Itoa(hostID),
	}

	var response WormlyGetScheduledDowntimePeriodsResponse
	if err := c.makeFormRequest(ctx, "getScheduledDowntimePeriods", params, &response); err != nil {
		return nil, fmt.Errorf("failed to get scheduled downtime periods: %w", err)
	}

	if response.ErrorCode != 0 {
		return nil, fmt.Errorf("API returned error code %d", response.ErrorCode)
	}

	// Set the HostID for all periods since the API response doesn't include it
	for i := range response.Periods {
		response.Periods[i].HostID = hostID
	}

	return response.Periods, nil
}
