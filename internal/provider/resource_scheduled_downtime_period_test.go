package provider

import (
	"errors"
	"fmt"
	"os"
	"testing"

	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/radarnex/terraform-provider-wormly/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestScheduledDowntimePeriodResource_Metadata(t *testing.T) {
	r := NewScheduledDowntimePeriodResource()
	req := frameworkresource.MetadataRequest{
		ProviderTypeName: "wormly",
	}
	resp := &frameworkresource.MetadataResponse{}

	r.Metadata(t.Context(), req, resp)

	assert.Equal(t, "wormly_scheduled_downtime_period", resp.TypeName)
}

func TestScheduledDowntimePeriodResource_Configure(t *testing.T) {
	r := &scheduledDowntimePeriodResource{}
	mockClient := &client.MockScheduledDowntimePeriodAPI{}

	req := frameworkresource.ConfigureRequest{
		ProviderData: mockClient,
	}
	resp := &frameworkresource.ConfigureResponse{}

	r.Configure(t.Context(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, mockClient, r.client)
}

func TestScheduledDowntimePeriodResource_Configure_InvalidType(t *testing.T) {
	r := &scheduledDowntimePeriodResource{}

	req := frameworkresource.ConfigureRequest{
		ProviderData: "invalid",
	}
	resp := &frameworkresource.ConfigureResponse{}

	r.Configure(t.Context(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Unexpected Resource Configure Type")
}

func TestScheduledDowntimePeriodAPI_Methods(t *testing.T) {
	mockClient := &client.MockScheduledDowntimePeriodAPI{}

	// Test CreateScheduledDowntimePeriod
	expectedPeriod := &client.ScheduledDowntimePeriod{
		ID:         123,
		HostID:     12345,
		Start:      "22:00",
		End:        "06:00",
		Timezone:   "GMT",
		Recurrence: "DAILY",
		On:         "",
	}

	mockClient.On("CreateScheduledDowntimePeriod",
		mock.Anything, 12345, "22:00", "06:00", "GMT", "DAILY", "").
		Return(expectedPeriod, nil)

	period, err := mockClient.CreateScheduledDowntimePeriod(
		t.Context(), 12345, "22:00", "06:00", "GMT", "DAILY", "")
	assert.NoError(t, err)
	assert.Equal(t, expectedPeriod, period)

	// Test GetScheduledDowntimePeriod
	mockClient.On("GetScheduledDowntimePeriod", mock.Anything, 12345, 123).
		Return(expectedPeriod, nil)

	period, err = mockClient.GetScheduledDowntimePeriod(t.Context(), 12345, 123)
	assert.NoError(t, err)
	assert.Equal(t, expectedPeriod, period)

	// Test UpdateScheduledDowntimePeriod
	updatedPeriod := &client.ScheduledDowntimePeriod{
		ID:         123,
		HostID:     12345,
		Start:      "23:00", // Changed from 22:00
		End:        "07:00", // Changed from 06:00
		Timezone:   "GMT",
		Recurrence: "DAILY",
		On:         "",
	}

	mockClient.On("UpdateScheduledDowntimePeriod",
		mock.Anything, 12345, 123, "23:00", "07:00", "GMT", "DAILY", "").
		Return(updatedPeriod, nil)

	period, err = mockClient.UpdateScheduledDowntimePeriod(
		t.Context(), 12345, 123, "23:00", "07:00", "GMT", "DAILY", "")
	assert.NoError(t, err)
	assert.Equal(t, updatedPeriod, period)

	// Test DeleteScheduledDowntimePeriod
	mockClient.On("DeleteScheduledDowntimePeriod", mock.Anything, 12345, 123).
		Return(nil)

	err = mockClient.DeleteScheduledDowntimePeriod(t.Context(), 12345, 123)
	assert.NoError(t, err)

	// Test GetScheduledDowntimePeriods
	expectedPeriods := []client.ScheduledDowntimePeriod{*expectedPeriod}
	mockClient.On("GetScheduledDowntimePeriods", mock.Anything, 12345).
		Return(expectedPeriods, nil)

	periods, err := mockClient.GetScheduledDowntimePeriods(t.Context(), 12345)
	assert.NoError(t, err)
	assert.Equal(t, expectedPeriods, periods)

	mockClient.AssertExpectations(t)
}

func TestScheduledDowntimePeriodAPI_CreateWithOnParameter(t *testing.T) {
	mockClient := &client.MockScheduledDowntimePeriodAPI{}

	expectedPeriod := &client.ScheduledDowntimePeriod{
		ID:         456,
		HostID:     12345,
		Start:      "10:00",
		End:        "11:00",
		Timezone:   "Europe/London",
		Recurrence: "ONCEONLY",
		On:         "2025-12-25",
	}

	mockClient.On("CreateScheduledDowntimePeriod",
		mock.Anything, 12345, "10:00", "11:00", "Europe/London", "ONCEONLY", "2025-12-25").
		Return(expectedPeriod, nil)

	period, err := mockClient.CreateScheduledDowntimePeriod(
		t.Context(), 12345, "10:00", "11:00", "Europe/London", "ONCEONLY", "2025-12-25")
	assert.NoError(t, err)
	assert.Equal(t, expectedPeriod, period)

	mockClient.AssertExpectations(t)
}

func TestScheduledDowntimePeriodAPI_ErrorHandling(t *testing.T) {
	mockClient := &client.MockScheduledDowntimePeriodAPI{}

	// Test CreateScheduledDowntimePeriod error
	mockClient.On("CreateScheduledDowntimePeriod",
		mock.Anything, 12345, "22:00", "06:00", "GMT", "DAILY", "").
		Return(nil, errors.New("API error"))

	period, err := mockClient.CreateScheduledDowntimePeriod(
		t.Context(), 12345, "22:00", "06:00", "GMT", "DAILY", "")
	assert.Error(t, err)
	assert.Nil(t, period)
	assert.Contains(t, err.Error(), "API error")

	// Test GetScheduledDowntimePeriod not found error
	mockClient.On("GetScheduledDowntimePeriod", mock.Anything, 12345, 999).
		Return(nil, errors.New("scheduled downtime period with ID 999 not found"))

	period, err = mockClient.GetScheduledDowntimePeriod(t.Context(), 12345, 999)
	assert.Error(t, err)
	assert.Nil(t, period)
	assert.Contains(t, err.Error(), "not found")

	// Test DeleteScheduledDowntimePeriod error
	mockClient.On("DeleteScheduledDowntimePeriod", mock.Anything, 12345, 123).
		Return(errors.New("Period not found"))

	err = mockClient.DeleteScheduledDowntimePeriod(t.Context(), 12345, 123)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Period not found")

	mockClient.AssertExpectations(t)
}

func TestAccScheduledDowntimePeriodResource_basic(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccScheduledDowntimePeriodResourceConfig(rName, "14:00", "16:00", "America/New_York"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wormly_scheduled_downtime_period.test", "start", "14:00"),
					resource.TestCheckResourceAttr("wormly_scheduled_downtime_period.test", "end", "16:00"),
					resource.TestCheckResourceAttr("wormly_scheduled_downtime_period.test", "timezone", "America/New_York"),
					resource.TestCheckResourceAttr("wormly_scheduled_downtime_period.test", "recurrence", "DAILY"),
				),
			},
			// Update and Read testing
			{
				Config: testAccScheduledDowntimePeriodResourceConfig(rName, "14:00", "16:00", "America/Los_Angeles"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wormly_scheduled_downtime_period.test", "start", "14:00"),
					resource.TestCheckResourceAttr("wormly_scheduled_downtime_period.test", "end", "16:00"),
					resource.TestCheckResourceAttr("wormly_scheduled_downtime_period.test", "timezone", "America/Los_Angeles"),
					resource.TestCheckResourceAttr("wormly_scheduled_downtime_period.test", "recurrence", "DAILY"),
				),
			},
			// Import testing
			{
				ResourceName:      "wormly_scheduled_downtime_period.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccScheduledDowntimePeriodImportStateIdFunc("wormly_scheduled_downtime_period.test"),
			},
		},
	})
}

func testAccScheduledDowntimePeriodResourceConfig(hostName, start, end, timezone string) string {
	return fmt.Sprintf(`
provider "wormly" {
  api_key = "%s"
}

resource "wormly_host" "test" {
  name          = "%s"
  enabled       = true
  test_interval = 60
}

resource "wormly_scheduled_downtime_period" "test" {
  hostid     = wormly_host.test.id
  start      = "%s"
  end        = "%s"
  timezone   = "%s"
  recurrence = "DAILY"
}
`, os.Getenv("WORMLY_API_KEY"), hostName, start, end, timezone)
}

func testAccScheduledDowntimePeriodImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Resource not found: %s", resourceName)
		}

		hostID := rs.Primary.Attributes["hostid"]
		periodID := rs.Primary.ID

		return fmt.Sprintf("%s/%s", hostID, periodID), nil
	}
}
