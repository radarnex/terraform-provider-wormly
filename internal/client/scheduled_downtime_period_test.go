package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient_CreateScheduledDowntimePeriod(t *testing.T) {
	tests := []struct {
		name           string
		hostID         int
		start          string
		end            string
		timezone       string
		recurrence     string
		on             string
		responseBody   string
		expectedError  bool
		expectedResult *ScheduledDowntimePeriod
	}{
		{
			name:         "successful creation",
			hostID:       12345,
			start:        "22:00",
			end:          "06:00",
			timezone:     "GMT",
			recurrence:   "DAILY",
			on:           "",
			responseBody: `{"errorcode": 0, "periodid": 123}`,
			expectedResult: &ScheduledDowntimePeriod{
				ID:         123,
				HostID:     12345,
				Start:      "22:00",
				End:        "06:00",
				Timezone:   "GMT",
				Recurrence: "DAILY",
				On:         "",
			},
		},
		{
			name:         "successful creation with on parameter",
			hostID:       12345,
			start:        "10:00",
			end:          "11:00",
			timezone:     "Europe/London",
			recurrence:   "ONCEONLY",
			on:           "2025-12-25",
			responseBody: `{"errorcode": 0, "periodid": 456}`,
			expectedResult: &ScheduledDowntimePeriod{
				ID:         456,
				HostID:     12345,
				Start:      "10:00",
				End:        "11:00",
				Timezone:   "Europe/London",
				Recurrence: "ONCEONLY",
				On:         "2025-12-25",
			},
		},
		{
			name:          "API error",
			hostID:        12345,
			start:         "22:00",
			end:           "06:00",
			timezone:      "GMT",
			recurrence:    "DAILY",
			on:            "",
			responseBody:  `{"errorcode": 1, "message": "Invalid parameter"}`,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, tt.responseBody)
			}))
			defer server.Close()

			client, err := New(
				&http.Client{Timeout: 30 * time.Second},
				"test-api-key",
				server.URL,
				"test-agent/1.0",
				10.0, 3, time.Second, 2.0, 30*time.Second,
				NoOpLogger{}, false,
			)
			assert.NoError(err, "Failed to create client")

			result, err := client.CreateScheduledDowntimePeriod(
				t.Context(),
				tt.hostID,
				tt.start,
				tt.end,
				tt.timezone,
				tt.recurrence,
				tt.on,
			)

			if tt.expectedError {
				assert.Error(err, "Expected error but got none")
				return
			}

			assert.NoError(err, "Unexpected error")
			assert.Equal(tt.expectedResult.ID, result.ID)
			assert.Equal(tt.expectedResult.HostID, result.HostID)
			assert.Equal(tt.expectedResult.Start, result.Start)
			assert.Equal(tt.expectedResult.End, result.End)
			assert.Equal(tt.expectedResult.Timezone, result.Timezone)
			assert.Equal(tt.expectedResult.Recurrence, result.Recurrence)
			assert.Equal(tt.expectedResult.On, result.On)
		})
	}
}

func TestClient_GetScheduledDowntimePeriods(t *testing.T) {
	tests := []struct {
		name           string
		hostID         int
		responseBody   string
		expectedError  bool
		expectedResult []ScheduledDowntimePeriod
	}{
		{
			name:   "successful retrieval",
			hostID: 12345,
			responseBody: `{
				"errorcode": 0,
				"periods": [
					{
						"periodid": 123,
						"start": "22:00",
						"end": "06:00",
						"timezone": "GMT",
						"recurrence": "DAILY",
						"on": null
					},
					{
						"periodid": 456,
						"start": "10:00",
						"end": "11:00",
						"timezone": "Europe/London",
						"recurrence": "ONCEONLY",
						"on": "2025-12-25"
					}
				]
			}`,
			expectedResult: []ScheduledDowntimePeriod{
				{
					ID:         123,
					HostID:     12345,
					Start:      "22:00",
					End:        "06:00",
					Timezone:   "GMT",
					Recurrence: "DAILY",
					On:         "",
				},
				{
					ID:         456,
					HostID:     12345,
					Start:      "10:00",
					End:        "11:00",
					Timezone:   "Europe/London",
					Recurrence: "ONCEONLY",
					On:         "2025-12-25",
				},
			},
		},
		{
			name:           "empty result",
			hostID:         12345,
			responseBody:   `{"errorcode": 0, "periods": []}`,
			expectedResult: []ScheduledDowntimePeriod{},
		},
		{
			name:          "API error",
			hostID:        12345,
			responseBody:  `{"errorcode": 1}`,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, tt.responseBody)
			}))
			defer server.Close()

			client, err := New(
				&http.Client{Timeout: 30 * time.Second},
				"test-api-key",
				server.URL,
				"test-agent/1.0",
				10.0, 3, time.Second, 2.0, 30*time.Second,
				NoOpLogger{}, false,
			)
			assert.NoError(err, "Failed to create client")

			result, err := client.GetScheduledDowntimePeriods(t.Context(), tt.hostID)

			if tt.expectedError {
				assert.Error(err, "Expected error but got none")
				return
			}

			assert.NoError(err, "Unexpected error")
			assert.Len(result, len(tt.expectedResult))

			for i, expected := range tt.expectedResult {
				actual := result[i]
				assert.Equal(expected.ID, actual.ID, "Period %d: ID mismatch", i)
				assert.Equal(expected.HostID, actual.HostID, "Period %d: HostID mismatch", i)
				assert.Equal(expected.Start, actual.Start, "Period %d: Start mismatch", i)
				assert.Equal(expected.End, actual.End, "Period %d: End mismatch", i)
				assert.Equal(expected.Timezone, actual.Timezone, "Period %d: Timezone mismatch", i)
				assert.Equal(expected.Recurrence, actual.Recurrence, "Period %d: Recurrence mismatch", i)
				assert.Equal(expected.On, actual.On, "Period %d: On mismatch", i)
			}
		})
	}
}

func TestClient_DeleteScheduledDowntimePeriod(t *testing.T) {
	tests := []struct {
		name          string
		hostID        int
		periodID      int
		responseBody  string
		expectedError bool
	}{
		{
			name:         "successful deletion",
			hostID:       12345,
			periodID:     123,
			responseBody: `{"errorcode": 0}`,
		},
		{
			name:          "API error",
			hostID:        12345,
			periodID:      123,
			responseBody:  `{"errorcode": 1, "message": "Period not found"}`,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, tt.responseBody)
			}))
			defer server.Close()

			client, err := New(
				&http.Client{Timeout: 30 * time.Second},
				"test-api-key",
				server.URL,
				"test-agent/1.0",
				10.0, 3, time.Second, 2.0, 30*time.Second,
				NoOpLogger{}, false,
			)
			assert.NoError(err, "Failed to create client")

			err = client.DeleteScheduledDowntimePeriod(t.Context(), tt.hostID, tt.periodID)

			if tt.expectedError {
				assert.Error(err, "Expected error but got none")
				return
			}

			assert.NoError(err, "Unexpected error")
		})
	}
}

func TestScheduledDowntimePeriod_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		expectedID  int
		expectError bool
	}{
		{
			name:       "periodid as string",
			jsonData:   `{"periodid": "123", "hostid": 456, "start": "22:00", "end": "06:00", "timezone": "GMT", "recurrence": "DAILY"}`,
			expectedID: 123,
		},
		{
			name:       "periodid as integer",
			jsonData:   `{"periodid": 123, "hostid": 456, "start": "22:00", "end": "06:00", "timezone": "GMT", "recurrence": "DAILY"}`,
			expectedID: 123,
		},
		{
			name:       "periodid as float",
			jsonData:   `{"periodid": 123.0, "hostid": 456, "start": "22:00", "end": "06:00", "timezone": "GMT", "recurrence": "DAILY"}`,
			expectedID: 123,
		},
		{
			name:       "periodid as empty string",
			jsonData:   `{"periodid": "", "hostid": 456, "start": "22:00", "end": "06:00", "timezone": "GMT", "recurrence": "DAILY"}`,
			expectedID: 0,
		},
		{
			name:        "periodid as invalid string",
			jsonData:    `{"periodid": "invalid", "hostid": 456, "start": "22:00", "end": "06:00", "timezone": "GMT", "recurrence": "DAILY"}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			var period ScheduledDowntimePeriod
			err := json.Unmarshal([]byte(tt.jsonData), &period)

			if tt.expectError {
				assert.Error(err, "Expected error but got none")
				return
			}

			assert.NoError(err, "Unexpected error")
			assert.Equal(tt.expectedID, period.ID)
		})
	}
}
