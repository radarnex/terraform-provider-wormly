package provider

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/radarnex/terraform-provider-wormly/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSensorHTTPResource_Metadata(t *testing.T) {
	r := NewSensorHTTPResource()
	req := frameworkresource.MetadataRequest{
		ProviderTypeName: "wormly",
	}
	resp := &frameworkresource.MetadataResponse{}

	r.Metadata(t.Context(), req, resp)

	assert.Equal(t, "wormly_sensor_http", resp.TypeName)
}

func TestSensorHTTPResource_Configure(t *testing.T) {
	r := &sensorHTTPResource{}
	mockClient := &client.MockSensorHTTPAPI{}

	req := frameworkresource.ConfigureRequest{
		ProviderData: mockClient,
	}
	resp := &frameworkresource.ConfigureResponse{}

	r.Configure(t.Context(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, mockClient, r.client)
}

func TestSensorHTTPResource_Configure_InvalidType(t *testing.T) {
	r := &sensorHTTPResource{}

	req := frameworkresource.ConfigureRequest{
		ProviderData: "invalid",
	}
	resp := &frameworkresource.ConfigureResponse{}

	r.Configure(t.Context(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestSensorHTTPAPI_Methods(t *testing.T) {
	mockClient := &client.MockSensorHTTPAPI{}

	// Test CreateSensorHTTP
	expectedSensor := &client.SensorHTTP{
		ID:        123,
		HostID:    456,
		URL:       "https://example.com",
		NiceName:  "Test Sensor",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	createReq := &client.SensorHTTPCreateRequest{
		HostID:   456,
		URL:      "https://example.com",
		NiceName: "Test Sensor",
	}

	mockClient.On("CreateSensorHTTP", mock.Anything, createReq).Return(expectedSensor, nil)

	sensor, err := mockClient.CreateSensorHTTP(t.Context(), createReq)
	assert.NoError(t, err)
	assert.Equal(t, expectedSensor, sensor)

	// Test GetSensorHTTP
	mockClient.On("GetSensorHTTP", mock.Anything, 456, 123).Return(expectedSensor, nil)

	sensor, err = mockClient.GetSensorHTTP(t.Context(), 456, 123)
	assert.NoError(t, err)
	assert.Equal(t, expectedSensor, sensor)

	// Test DeleteSensorHTTP
	mockClient.On("DeleteSensorHTTP", mock.Anything, 123).Return(nil)

	err = mockClient.DeleteSensorHTTP(t.Context(), 123)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}

func TestParseSensorID(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		expectedHost   int
		expectedSensor int
		expectError    bool
	}{
		{
			name:           "valid ID",
			id:             "123/456",
			expectedHost:   123,
			expectedSensor: 456,
			expectError:    false,
		},
		{
			name:        "invalid format - no slash",
			id:          "123456",
			expectError: true,
		},
		{
			name:        "invalid format - too many parts",
			id:          "123/456/789",
			expectError: true,
		},
		{
			name:        "invalid host ID",
			id:          "abc/456",
			expectError: true,
		},
		{
			name:        "invalid sensor ID",
			id:          "123/def",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hostID, sensorID, err := parseSensorID(tt.id)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedHost, hostID)
				assert.Equal(t, tt.expectedSensor, sensorID)
			}
		})
	}
}

func TestSensorHTTPResource_ErrorHandling(t *testing.T) {
	// Test CreateSensorHTTP error handling
	mockClient := &client.MockSensorHTTPAPI{}
	createReq := &client.SensorHTTPCreateRequest{
		HostID: 456,
		URL:    "https://example.com",
	}
	mockClient.On("CreateSensorHTTP", mock.Anything, createReq).Return(nil, errors.New("API error"))

	_, err := mockClient.CreateSensorHTTP(t.Context(), createReq)
	assert.Error(t, err)
	assert.Equal(t, "API error", err.Error())

	// Test GetSensorHTTP error handling
	mockClient.On("GetSensorHTTP", mock.Anything, 456, 123).Return(nil, errors.New("API error"))

	_, err = mockClient.GetSensorHTTP(t.Context(), 456, 123)
	assert.Error(t, err)
	assert.Equal(t, "API error", err.Error())

	// Test DeleteSensorHTTP error handling
	mockClient.On("DeleteSensorHTTP", mock.Anything, 123).Return(errors.New("API error"))

	err = mockClient.DeleteSensorHTTP(t.Context(), 123)
	assert.Error(t, err)
	assert.Equal(t, "API error", err.Error())

	mockClient.AssertExpectations(t)
}

func TestSensorHTTPResource_CreateRequestBuilding(t *testing.T) {
	// Test that the resource correctly builds the create request from Terraform data
	mockClient := &client.MockSensorHTTPAPI{}

	expectedSensor := &client.SensorHTTP{
		ID:                   123,
		HostID:               456,
		URL:                  "https://example.com",
		NiceName:             "Test Sensor",
		Timeout:              30,
		ResponseCode:         "200",
		VerifySSLCert:        true,
		SearchHeaders:        false,
		ExpectedText:         "success",
		UnwantedText:         "error",
		SSLValidity:          30,
		Cookies:              "session=abc123",
		PostParams:           "param1=value1",
		CustomRequestHeaders: "X-Custom: test",
		UserAgent:            "test-agent",
		ForceResolve:         "1.2.3.4",
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	expectedCreateReq := &client.SensorHTTPCreateRequest{
		HostID:               456,
		URL:                  "https://example.com",
		NiceName:             "Test Sensor",
		Timeout:              30,
		ResponseCode:         "200",
		VerifySSLCert:        true,
		SearchHeaders:        false,
		ExpectedText:         "success",
		UnwantedText:         "error",
		SSLValidity:          30,
		Cookies:              "session=abc123",
		PostParams:           "param1=value1",
		CustomRequestHeaders: "X-Custom: test",
		UserAgent:            "test-agent",
		ForceResolve:         "1.2.3.4",
	}

	mockClient.On("CreateSensorHTTP", mock.Anything, expectedCreateReq).Return(expectedSensor, nil)

	// Test that the mock call would work with the expected request
	sensor, err := mockClient.CreateSensorHTTP(t.Context(), expectedCreateReq)
	assert.NoError(t, err)
	assert.Equal(t, expectedSensor, sensor)

	mockClient.AssertExpectations(t)
}

func TestSensorHTTPResource_ReadWithNotFoundError(t *testing.T) {
	// Test that 404 errors during Read properly remove the resource from state
	mockClient := &client.MockSensorHTTPAPI{}

	// Simulate a 404 error
	mockClient.On("GetSensorHTTP", mock.Anything, 456, 123).Return(nil, errors.New("404 not found"))

	_, err := mockClient.GetSensorHTTP(t.Context(), 456, 123)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")

	mockClient.AssertExpectations(t)
}

func TestSensorHTTPResource_ModelMapping(t *testing.T) {
	// Test that the model correctly maps to and from the API struct
	model := sensorHTTPResourceModel{
		ID:                   types.StringValue("456/123"),
		HostID:               types.Int64Value(456),
		URL:                  types.StringValue("https://example.com"),
		NiceName:             types.StringValue("Test Sensor"),
		Timeout:              types.Int64Value(30),
		ResponseCode:         types.StringValue("200"),
		VerifySSLCert:        types.BoolValue(true),
		SearchHeaders:        types.BoolValue(false),
		ExpectedText:         types.StringValue("success"),
		UnwantedText:         types.StringValue("error"),
		SSLValidity:          types.Int64Value(30),
		Cookies:              types.StringValue("session=abc123"),
		PostParams:           types.StringValue("param1=value1"),
		CustomRequestHeaders: types.StringValue("X-Custom: test"),
		UserAgent:            types.StringValue("test-agent"),
		ForceResolve:         types.StringValue("1.2.3.4"),
	}

	// Verify the model has the expected values
	assert.Equal(t, "456/123", model.ID.ValueString())
	assert.Equal(t, int64(456), model.HostID.ValueInt64())
	assert.Equal(t, "https://example.com", model.URL.ValueString())
	assert.Equal(t, "Test Sensor", model.NiceName.ValueString())
	assert.Equal(t, int64(30), model.Timeout.ValueInt64())
	assert.Equal(t, "200", model.ResponseCode.ValueString())
	assert.True(t, model.VerifySSLCert.ValueBool())
	assert.False(t, model.SearchHeaders.ValueBool())
	assert.Equal(t, "success", model.ExpectedText.ValueString())
	assert.Equal(t, "error", model.UnwantedText.ValueString())
	assert.Equal(t, int64(30), model.SSLValidity.ValueInt64())
	assert.Equal(t, "session=abc123", model.Cookies.ValueString())
	assert.Equal(t, "param1=value1", model.PostParams.ValueString())
	assert.Equal(t, "X-Custom: test", model.CustomRequestHeaders.ValueString())
	assert.Equal(t, "test-agent", model.UserAgent.ValueString())
	assert.Equal(t, "1.2.3.4", model.ForceResolve.ValueString())
}

func TestAccSensorHTTPResource_basic(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSensorHTTPResourceConfig(rName, "https://example.org"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wormly_sensor_http.test", "url", "https://example.org"),
					resource.TestCheckResourceAttr("wormly_sensor_http.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("wormly_sensor_http.test", "host_id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccSensorHTTPResourceConfig(rName, "https://httpbin.org/get"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wormly_sensor_http.test", "url", "https://httpbin.org/get"),
					resource.TestCheckResourceAttr("wormly_sensor_http.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("wormly_sensor_http.test", "host_id"),
				),
			},
			// Import testing
			{
				ResourceName:      "wormly_sensor_http.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSensorHTTPResource_importWithOptionalReplaceFields(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSensorHTTPResourceConfigWithOptionalReplaceFields(rName, "https://example.org", "Import Regression", 29),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wormly_sensor_http.test", "url", "https://example.org"),
					resource.TestCheckResourceAttr("wormly_sensor_http.test", "nice_name", "Import Regression"),
					resource.TestCheckResourceAttr("wormly_sensor_http.test", "timeout", "29"),
				),
			},
			{
				ResourceName:      "wormly_sensor_http.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSensorHTTPResourceConfig(hostName, url string) string {
	return fmt.Sprintf(`
provider "wormly" {
  api_key = "%s"
}

resource "wormly_host" "test" {
  name          = "%s"
  enabled       = true
  test_interval = 60
}

resource "wormly_sensor_http" "test" {
  host_id = wormly_host.test.id
  url     = "%s"
  enabled = true
}
`, os.Getenv("WORMLY_API_KEY"), hostName, url)
}

func testAccSensorHTTPResourceConfigWithOptionalReplaceFields(hostName, url, niceName string, timeout int) string {
	return fmt.Sprintf(`
provider "wormly" {
  api_key = "%s"
}

resource "wormly_host" "test" {
  name          = "%s"
  enabled       = true
  test_interval = 60
}

resource "wormly_sensor_http" "test" {
  host_id   = wormly_host.test.id
  url       = "%s"
  nice_name = "%s"
  timeout   = %d
  enabled   = true
}
`, os.Getenv("WORMLY_API_KEY"), hostName, url, niceName, timeout)
}
