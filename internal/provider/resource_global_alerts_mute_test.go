package provider

import (
	"fmt"
	"os"
	"testing"

	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/radarnex/terraform-provider-wormly/internal/client"
	"github.com/stretchr/testify/assert"
)

func TestGlobalAlertsMuteResource_Metadata(t *testing.T) {
	r := NewGlobalAlertsMuteResource()
	req := frameworkresource.MetadataRequest{
		ProviderTypeName: "wormly",
	}
	resp := &frameworkresource.MetadataResponse{}

	r.Metadata(t.Context(), req, resp)

	assert.Equal(t, "wormly_global_alerts_mute", resp.TypeName)
}

func TestGlobalAlertsMuteResource_Configure(t *testing.T) {
	r := &globalAlertsMuteResource{}
	mockClient := &client.Client{}

	req := frameworkresource.ConfigureRequest{
		ProviderData: mockClient,
	}
	resp := &frameworkresource.ConfigureResponse{}

	r.Configure(t.Context(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, mockClient, r.client)
}

func TestGlobalAlertsMuteResource_Configure_InvalidType(t *testing.T) {
	r := &globalAlertsMuteResource{}

	req := frameworkresource.ConfigureRequest{
		ProviderData: "invalid",
	}
	resp := &frameworkresource.ConfigureResponse{}

	r.Configure(t.Context(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Unexpected Resource Configure Type")
}

func TestGlobalAlertsMuteResource_Schema(t *testing.T) {
	r := &globalAlertsMuteResource{}
	req := frameworkresource.SchemaRequest{}
	resp := &frameworkresource.SchemaResponse{}

	r.Schema(t.Context(), req, resp)

	assert.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "id")
	assert.Contains(t, resp.Schema.Attributes, "enabled")
	assert.True(t, resp.Schema.Attributes["id"].IsComputed())
	assert.True(t, resp.Schema.Attributes["enabled"].IsOptional())
}

func TestAccGlobalAlertsMuteResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccGlobalAlertsMuteResourceConfig(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wormly_global_alerts_mute.test", "enabled", "true"),
				),
			},
			// Update and Read testing
			{
				Config: testAccGlobalAlertsMuteResourceConfig(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wormly_global_alerts_mute.test", "enabled", "false"),
				),
			},
		},
	})
}

func testAccGlobalAlertsMuteResourceConfig(enabled bool) string {
	return fmt.Sprintf(`
provider "wormly" {
  api_key = "%s"
}

resource "wormly_global_alerts_mute" "test" {
  enabled = %t
}
`, os.Getenv("WORMLY_API_KEY"), enabled)
}
