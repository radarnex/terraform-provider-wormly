package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/radarnex/terraform-provider-wormly/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &hostDataSource{}
	_ datasource.DataSourceWithConfigure = &hostDataSource{}
)

// NewHostDataSource is a helper function to simplify the provider implementation.
func NewHostDataSource() datasource.DataSource {
	return &hostDataSource{}
}

// hostDataSource is the data source implementation.
type hostDataSource struct {
	client client.HostAPI
}

// hostDataSourceModel describes the data source data model.
type hostDataSourceModel struct {
	ID      types.Int64  `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

func (d *hostDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

func (d *hostDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Wormly host data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Host identifier",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Host name",
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the host is enabled",
				Computed:            true,
			},
		},
	}
}

func (d *hostDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *hostDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data hostDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	hostID := int(data.ID.ValueInt64())
	host, err := d.client.GetHost(ctx, hostID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read host, got error: %s", err))
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.Name = types.StringValue(host.Name)
	data.Enabled = types.BoolValue(host.Enabled)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
