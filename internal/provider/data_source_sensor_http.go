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
	_ datasource.DataSource              = &sensorHTTPDataSource{}
	_ datasource.DataSourceWithConfigure = &sensorHTTPDataSource{}
)

// NewSensorHTTPDataSource is a helper function to simplify the provider implementation.
func NewSensorHTTPDataSource() datasource.DataSource {
	return &sensorHTTPDataSource{}
}

// sensorHTTPDataSource is the data source implementation.
type sensorHTTPDataSource struct {
	client client.SensorHTTPAPI
}

// sensorHTTPDataSourceModel describes the data source data model.
type sensorHTTPDataSourceModel struct {
	HostID  types.Int64                       `tfsdk:"host_id"`
	Sensors []sensorHTTPDataSourceSensorModel `tfsdk:"sensors"`
}

// sensorHTTPDataSourceSensorModel describes the sensor data model.
type sensorHTTPDataSourceSensorModel struct {
	ID       types.Int64             `tfsdk:"id"`
	NiceName types.String            `tfsdk:"nice_name"`
	Enabled  types.Bool              `tfsdk:"enabled"`
	Params   map[string]types.String `tfsdk:"params"`
}

func (d *sensorHTTPDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sensor_http"
}

func (d *sensorHTTPDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Wormly HTTP sensor data source",

		Attributes: map[string]schema.Attribute{
			"host_id": schema.Int64Attribute{
				MarkdownDescription: "Host identifier",
				Required:            true,
			},
			"sensors": schema.ListNestedAttribute{
				MarkdownDescription: "List of HTTP sensors for the host",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "Sensor identifier",
							Computed:            true,
						},
						"nice_name": schema.StringAttribute{
							MarkdownDescription: "Sensor nice name",
							Computed:            true,
						},
						"enabled": schema.BoolAttribute{
							MarkdownDescription: "Whether the sensor is enabled",
							Computed:            true,
						},
						"params": schema.MapAttribute{
							MarkdownDescription: "Sensor parameters",
							ElementType:         types.StringType,
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *sensorHTTPDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sensorHTTPDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data sensorHTTPDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	hostID := int(data.HostID.ValueInt64())
	sensors, err := d.client.ListSensorHTTP(ctx, hostID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read sensors, got error: %s", err))
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.Sensors = make([]sensorHTTPDataSourceSensorModel, len(sensors))
	for i, sensor := range sensors {
		params := make(map[string]types.String)
		params["url"] = types.StringValue(sensor.URL)
		params["timeout"] = types.StringValue(fmt.Sprintf("%d", sensor.Timeout))
		params["response_code"] = types.StringValue(sensor.ResponseCode)
		params["verify_ssl_cert"] = types.StringValue(fmt.Sprintf("%t", sensor.VerifySSLCert))
		params["search_headers"] = types.StringValue(fmt.Sprintf("%t", sensor.SearchHeaders))
		params["expected_text"] = types.StringValue(sensor.ExpectedText)
		params["unwanted_text"] = types.StringValue(sensor.UnwantedText)
		params["ssl_validity"] = types.StringValue(fmt.Sprintf("%d", sensor.SSLValidity))
		params["cookies"] = types.StringValue(sensor.Cookies)
		params["post_params"] = types.StringValue(sensor.PostParams)
		params["custom_request_headers"] = types.StringValue(sensor.CustomRequestHeaders)
		params["user_agent"] = types.StringValue(sensor.UserAgent)
		params["force_resolve"] = types.StringValue(sensor.ForceResolve)

		data.Sensors[i] = sensorHTTPDataSourceSensorModel{
			ID:       types.Int64Value(int64(sensor.ID)),
			NiceName: types.StringValue(sensor.NiceName),
			Enabled:  types.BoolValue(sensor.Enabled),
			Params:   params,
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
