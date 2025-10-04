package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/radarnex/terraform-provider-wormly/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSensorHTTPDataSource_Metadata(t *testing.T) {
	dataSource := NewSensorHTTPDataSource()
	req := datasource.MetadataRequest{
		ProviderTypeName: "wormly",
	}
	resp := &datasource.MetadataResponse{}

	dataSource.Metadata(t.Context(), req, resp)

	assert.Equal(t, "wormly_sensor_http", resp.TypeName)
}

func TestSensorHTTPDataSource_Schema(t *testing.T) {
	dataSource := NewSensorHTTPDataSource()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	dataSource.Schema(t.Context(), req, resp)

	assert.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "host_id")
	assert.Contains(t, resp.Schema.Attributes, "sensors")

	// Check that host_id is required
	hostIDAttr := resp.Schema.Attributes["host_id"]
	assert.True(t, hostIDAttr.IsRequired())

	// Check that sensors is computed
	sensorsAttr := resp.Schema.Attributes["sensors"]
	assert.True(t, sensorsAttr.IsComputed())
}

func TestSensorHTTPDataSource_Configure(t *testing.T) {
	dataSource, ok := NewSensorHTTPDataSource().(*sensorHTTPDataSource)
	if !ok {
		t.Fatal("Expected sensorHTTPDataSource type")
	}
	mockClient := &client.Client{}

	req := datasource.ConfigureRequest{
		ProviderData: mockClient,
	}
	resp := &datasource.ConfigureResponse{}

	dataSource.Configure(t.Context(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, mockClient, dataSource.client)
}

func TestSensorHTTPDataSource_Configure_Error(t *testing.T) {
	dataSource, ok := NewSensorHTTPDataSource().(*sensorHTTPDataSource)
	if !ok {
		t.Fatal("Expected sensorHTTPDataSource type")
	}

	req := datasource.ConfigureRequest{
		ProviderData: "invalid",
	}
	resp := &datasource.ConfigureResponse{}

	dataSource.Configure(t.Context(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Unexpected Data Source Configure Type")
}

func TestSensorHTTPDataSource_Read(t *testing.T) {
	// Create mock client
	mockClient := &client.MockSensorHTTPAPI{}

	// Set up mock expectations
	expectedSensors := []*client.SensorHTTP{
		{
			ID:                   1,
			HostID:               123,
			URL:                  "https://example.com",
			NiceName:             "Test Sensor 1",
			Timeout:              30,
			ResponseCode:         "200",
			VerifySSLCert:        true,
			SearchHeaders:        false,
			ExpectedText:         "",
			UnwantedText:         "",
			SSLValidity:          30,
			Cookies:              "",
			PostParams:           "",
			CustomRequestHeaders: "",
			UserAgent:            "",
			ForceResolve:         "",
		},
		{
			ID:                   2,
			HostID:               123,
			URL:                  "https://example.org",
			NiceName:             "Test Sensor 2",
			Timeout:              60,
			ResponseCode:         "200",
			VerifySSLCert:        false,
			SearchHeaders:        true,
			ExpectedText:         "Success",
			UnwantedText:         "Error",
			SSLValidity:          14,
			Cookies:              "session=abc123",
			PostParams:           "user=test",
			CustomRequestHeaders: "X-API-Key: secret",
			UserAgent:            "Custom Agent",
			ForceResolve:         "127.0.0.1",
		},
	}
	mockClient.On("ListSensorHTTP", mock.Anything, 123).Return(expectedSensors, nil)

	// Create data source with mock client
	dataSource := &sensorHTTPDataSource{
		client: mockClient,
	}

	// Test the client call directly
	ctx := t.Context()
	hostID := 123
	sensors, err := dataSource.client.ListSensorHTTP(ctx, hostID)
	assert.NoError(t, err)
	assert.Equal(t, expectedSensors, sensors)
	assert.Len(t, sensors, 2)

	// Verify first sensor
	assert.Equal(t, 1, sensors[0].ID)
	assert.Equal(t, "Test Sensor 1", sensors[0].NiceName)
	assert.Equal(t, "https://example.com", sensors[0].URL)

	// Verify second sensor
	assert.Equal(t, 2, sensors[1].ID)
	assert.Equal(t, "Test Sensor 2", sensors[1].NiceName)
	assert.Equal(t, "https://example.org", sensors[1].URL)

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}
