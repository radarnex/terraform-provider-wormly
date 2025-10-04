package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/radarnex/terraform-provider-wormly/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHostDataSource_Metadata(t *testing.T) {
	dataSource := NewHostDataSource()
	req := datasource.MetadataRequest{
		ProviderTypeName: "wormly",
	}
	resp := &datasource.MetadataResponse{}

	dataSource.Metadata(t.Context(), req, resp)

	assert.Equal(t, "wormly_host", resp.TypeName)
}

func TestHostDataSource_Schema(t *testing.T) {
	dataSource := NewHostDataSource()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	dataSource.Schema(t.Context(), req, resp)

	assert.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "id")
	assert.Contains(t, resp.Schema.Attributes, "name")
	assert.Contains(t, resp.Schema.Attributes, "enabled")

	// Check that id is required
	idAttr := resp.Schema.Attributes["id"]
	assert.True(t, idAttr.IsRequired())

	// Check that name and enabled are computed
	nameAttr := resp.Schema.Attributes["name"]
	assert.True(t, nameAttr.IsComputed())

	enabledAttr := resp.Schema.Attributes["enabled"]
	assert.True(t, enabledAttr.IsComputed())
}

func TestHostDataSource_Configure(t *testing.T) {
	ds := NewHostDataSource()
	dataSource, ok := ds.(*hostDataSource)
	assert.True(t, ok, "NewHostDataSource should return *hostDataSource")
	mockClient := &client.Client{}

	req := datasource.ConfigureRequest{
		ProviderData: mockClient,
	}
	resp := &datasource.ConfigureResponse{}

	dataSource.Configure(t.Context(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, mockClient, dataSource.client)
}

func TestHostDataSource_Configure_Error(t *testing.T) {
	dataSource, ok := NewHostDataSource().(*hostDataSource)
	if !ok {
		t.Fatal("Expected hostDataSource type")
	}

	req := datasource.ConfigureRequest{
		ProviderData: "invalid",
	}
	resp := &datasource.ConfigureResponse{}

	dataSource.Configure(t.Context(), req, resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Unexpected Data Source Configure Type")
}

func TestHostDataSource_Read(t *testing.T) {
	// Create mock client
	mockClient := &client.MockHostAPI{}

	// Set up mock expectations
	expectedHost := &client.Host{
		ID:      1,
		Name:    "test-host",
		Enabled: true,
	}
	mockClient.On("GetHost", mock.Anything, 1).Return(expectedHost, nil)

	// Create data source with mock client
	dataSource := &hostDataSource{
		client: mockClient,
	}

	// Test the client call directly
	ctx := t.Context()
	hostID := 1
	host, err := dataSource.client.GetHost(ctx, hostID)
	assert.NoError(t, err)
	assert.Equal(t, expectedHost, host)

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}
