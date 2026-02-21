package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/radarnex/terraform-provider-wormly/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &sensorHTTPResource{}
	_ resource.ResourceWithConfigure   = &sensorHTTPResource{}
	_ resource.ResourceWithImportState = &sensorHTTPResource{}
)

// sensorHTTPResourceModel represents the resource data model.
type sensorHTTPResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	HostID               types.Int64  `tfsdk:"host_id"`
	URL                  types.String `tfsdk:"url"`
	NiceName             types.String `tfsdk:"nice_name"`
	Enabled              types.Bool   `tfsdk:"enabled"`
	Timeout              types.Int64  `tfsdk:"timeout"`
	ResponseCode         types.String `tfsdk:"response_code"`
	VerifySSLCert        types.Bool   `tfsdk:"verify_ssl_cert"`
	SearchHeaders        types.Bool   `tfsdk:"search_headers"`
	ExpectedText         types.String `tfsdk:"expected_text"`
	UnwantedText         types.String `tfsdk:"unwanted_text"`
	SSLValidity          types.Int64  `tfsdk:"ssl_validity"`
	Cookies              types.String `tfsdk:"cookies"`
	PostParams           types.String `tfsdk:"post_params"`
	CustomRequestHeaders types.String `tfsdk:"custom_request_headers"`
	UserAgent            types.String `tfsdk:"user_agent"`
	ForceResolve         types.String `tfsdk:"force_resolve"`
}

// sensorHTTPResource defines the resource implementation.
type sensorHTTPResource struct {
	client client.SensorHTTPAPI
}

// NewSensorHTTPResource creates a new HTTP sensor resource.
func NewSensorHTTPResource() resource.Resource {
	return &sensorHTTPResource{}
}

func (r *sensorHTTPResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sensor_http"
}

func (r *sensorHTTPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Wormly HTTP sensor resource\n\n~> Note: Wormly's public API does not currently provide a dedicated update command for HTTP sensor settings, so changes to attributes other than `enabled` require resource replacement.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Sensor identifier in format <host_id>/<sensor_id>",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"host_id": schema.Int64Attribute{
				MarkdownDescription: "Host ID",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "URL to monitor",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"nice_name": schema.StringAttribute{
				MarkdownDescription: "Nice name for the sensor",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the sensor is enabled",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "Timeout in seconds",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
					int64planmodifier.RequiresReplace(),
				},
			},
			"response_code": schema.StringAttribute{
				MarkdownDescription: "Expected HTTP response code",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"verify_ssl_cert": schema.BoolAttribute{
				MarkdownDescription: "Whether to verify SSL certificate",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			"search_headers": schema.BoolAttribute{
				MarkdownDescription: "Whether to search headers",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			"expected_text": schema.StringAttribute{
				MarkdownDescription: "Expected text in response",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"unwanted_text": schema.StringAttribute{
				MarkdownDescription: "Unwanted text in response",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ssl_validity": schema.Int64Attribute{
				MarkdownDescription: "SSL validity period in days",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
					int64planmodifier.RequiresReplace(),
				},
			},
			"cookies": schema.StringAttribute{
				MarkdownDescription: "Cookies to send with request",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"post_params": schema.StringAttribute{
				MarkdownDescription: "POST parameters",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"custom_request_headers": schema.StringAttribute{
				MarkdownDescription: "Custom request headers",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_agent": schema.StringAttribute{
				MarkdownDescription: "User agent string",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"force_resolve": schema.StringAttribute{
				MarkdownDescription: "Force resolve to specific IP",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *sensorHTTPResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(client.SensorHTTPAPI)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected client.SensorHTTPAPI, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *sensorHTTPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data sensorHTTPResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plannedData := data

	// Build create request
	createReq := &client.SensorHTTPCreateRequest{
		HostID: int(data.HostID.ValueInt64()),
		URL:    data.URL.ValueString(),
	}

	if !data.NiceName.IsNull() && !data.NiceName.IsUnknown() {
		createReq.NiceName = data.NiceName.ValueString()
	}
	if !data.Timeout.IsNull() && !data.Timeout.IsUnknown() {
		createReq.Timeout = int(data.Timeout.ValueInt64())
	}
	if !data.ResponseCode.IsNull() && !data.ResponseCode.IsUnknown() {
		createReq.ResponseCode = data.ResponseCode.ValueString()
	}
	if !data.VerifySSLCert.IsNull() && !data.VerifySSLCert.IsUnknown() {
		createReq.VerifySSLCert = data.VerifySSLCert.ValueBool()
	}
	if !data.SearchHeaders.IsNull() && !data.SearchHeaders.IsUnknown() {
		createReq.SearchHeaders = data.SearchHeaders.ValueBool()
	}
	if !data.ExpectedText.IsNull() && !data.ExpectedText.IsUnknown() {
		createReq.ExpectedText = data.ExpectedText.ValueString()
	}
	if !data.UnwantedText.IsNull() && !data.UnwantedText.IsUnknown() {
		createReq.UnwantedText = data.UnwantedText.ValueString()
	}
	if !data.SSLValidity.IsNull() && !data.SSLValidity.IsUnknown() {
		createReq.SSLValidity = int(data.SSLValidity.ValueInt64())
	}
	if !data.Cookies.IsNull() && !data.Cookies.IsUnknown() {
		createReq.Cookies = data.Cookies.ValueString()
	}
	if !data.PostParams.IsNull() && !data.PostParams.IsUnknown() {
		createReq.PostParams = data.PostParams.ValueString()
	}
	if !data.CustomRequestHeaders.IsNull() && !data.CustomRequestHeaders.IsUnknown() {
		createReq.CustomRequestHeaders = data.CustomRequestHeaders.ValueString()
	}
	if !data.UserAgent.IsNull() && !data.UserAgent.IsUnknown() {
		createReq.UserAgent = data.UserAgent.ValueString()
	}
	if !data.ForceResolve.IsNull() && !data.ForceResolve.IsUnknown() {
		createReq.ForceResolve = data.ForceResolve.ValueString()
	}

	// Create the sensor
	sensor, err := r.client.CreateSensorHTTP(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create HTTP sensor, got error: %s", err))
		return
	}

	// Handle enabled state - ensure sensor matches desired state
	if data.Enabled.ValueBool() {
		// Explicitly enable the sensor to ensure it's enabled
		err = r.client.EnableSensorHTTP(ctx, sensor.ID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to enable HTTP sensor after creation, got error: %s", err))
			return
		}
	} else {
		// Explicitly disable the sensor
		err = r.client.DisableSensorHTTP(ctx, sensor.ID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to disable HTTP sensor after creation, got error: %s", err))
			return
		}
	}

	// Read the created sensor so all computed attributes are known in state.
	sensor, err = r.client.GetSensorHTTP(ctx, sensor.HostID, sensor.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read HTTP sensor after creation, got error: %s", err))
		return
	}

	// Set the computed ID in format <host_id>/<sensor_id>
	data.ID = types.StringValue(fmt.Sprintf("%d/%d", sensor.HostID, sensor.ID))
	setSensorHTTPResourceModelFromAPI(&data, sensor)
	applyKnownSensorHTTPPlanValues(&data, &plannedData)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sensorHTTPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data sensorHTTPResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the ID to get host_id and sensor_id
	hostID, sensorID, err := parseSensorID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse sensor ID: %s", err))
		return
	}

	// Get the sensor
	sensor, err := r.client.GetSensorHTTP(ctx, hostID, sensorID)
	if err != nil {
		// If sensor is not found (404), remove from state
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read HTTP sensor, got error: %s", err))
		return
	}

	// Update the model with the current state from API
	previousSSLValidity := data.SSLValidity
	setSensorHTTPResourceModelFromAPI(&data, sensor)
	preserveReadValuesWhenAPIDoesNotReturnThem(&data, sensor, previousSSLValidity)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *sensorHTTPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state sensorHTTPResourceModel

	// Read Terraform plan and current state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the ID to get sensor information
	_, _, err := parseSensorID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse sensor ID: %s", err))
		return
	}

	// Parse the sensor ID to get the HSID (which is the sensor ID from the client)
	parts := strings.Split(state.ID.ValueString(), "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Parse Error", "Invalid sensor ID format")
		return
	}
	hsid, err := strconv.Atoi(parts[1])
	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Invalid sensor ID: %s", err))
		return
	}

	// Check if enabled state changed
	if !plan.Enabled.Equal(state.Enabled) {
		if plan.Enabled.ValueBool() {
			// Enable the sensor
			err = r.client.EnableSensorHTTP(ctx, hsid)
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to enable HTTP sensor, got error: %s", err))
				return
			}
		} else {
			// Disable the sensor
			err = r.client.DisableSensorHTTP(ctx, hsid)
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to disable HTTP sensor, got error: %s", err))
				return
			}
		}
	}

	// Use the plan values but preserve the ID from state
	plan.ID = state.ID

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sensorHTTPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data sensorHTTPResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the ID to get sensor_id
	_, sensorID, err := parseSensorID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse sensor ID: %s", err))
		return
	}

	// Delete the sensor
	err = r.client.DeleteSensorHTTP(ctx, sensorID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete HTTP sensor, got error: %s", err))
		return
	}
}

func (r *sensorHTTPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Parse the import ID to validate format
	hostID, _, err := parseSensorID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", fmt.Sprintf("Expected import identifier with format: host_id/sensor_id. Got: %s", req.ID))
		return
	}

	// Set the ID and host_id in state
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("host_id"), int64(hostID))...)

	// Trigger a read to populate the rest of the attributes
	// The Read method will be called automatically after import
}

// parseSensorID parses a sensor ID in format "host_id/sensor_id" and returns the components.
func parseSensorID(id string) (hostID int, sensorID int, err error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid ID format, expected host_id/sensor_id")
	}

	hostID, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid host_id: %s", err)
	}

	sensorID, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid sensor_id: %s", err)
	}

	return hostID, sensorID, nil
}

func setSensorHTTPResourceModelFromAPI(data *sensorHTTPResourceModel, sensor *client.SensorHTTP) {
	data.HostID = types.Int64Value(int64(sensor.HostID))
	data.URL = types.StringValue(sensor.URL)
	data.NiceName = types.StringValue(sensor.NiceName)
	data.Enabled = types.BoolValue(sensor.Enabled)
	data.Timeout = types.Int64Value(int64(sensor.Timeout))
	data.ResponseCode = types.StringValue(sensor.ResponseCode)
	data.VerifySSLCert = types.BoolValue(sensor.VerifySSLCert)
	data.SearchHeaders = types.BoolValue(sensor.SearchHeaders)
	data.ExpectedText = types.StringValue(sensor.ExpectedText)
	data.UnwantedText = types.StringValue(sensor.UnwantedText)
	data.SSLValidity = types.Int64Value(int64(sensor.SSLValidity))
	data.Cookies = types.StringValue(sensor.Cookies)
	data.PostParams = types.StringValue(sensor.PostParams)
	data.CustomRequestHeaders = types.StringValue(sensor.CustomRequestHeaders)
	data.UserAgent = types.StringValue(sensor.UserAgent)
	data.ForceResolve = types.StringValue(sensor.ForceResolve)
}

func preserveReadValuesWhenAPIDoesNotReturnThem(data *sensorHTTPResourceModel, sensor *client.SensorHTTP, previousSSLValidity types.Int64) {
	if sensor.SSLValidity == 0 && !previousSSLValidity.IsNull() && !previousSSLValidity.IsUnknown() && previousSSLValidity.ValueInt64() > 0 {
		data.SSLValidity = previousSSLValidity
	}
}

func applyKnownSensorHTTPPlanValues(data *sensorHTTPResourceModel, plan *sensorHTTPResourceModel) {
	if !plan.NiceName.IsUnknown() {
		data.NiceName = plan.NiceName
	}
	if !plan.Timeout.IsUnknown() {
		data.Timeout = plan.Timeout
	}
	if !plan.ResponseCode.IsUnknown() {
		data.ResponseCode = plan.ResponseCode
	}
	if !plan.VerifySSLCert.IsUnknown() {
		data.VerifySSLCert = plan.VerifySSLCert
	}
	if !plan.SearchHeaders.IsUnknown() {
		data.SearchHeaders = plan.SearchHeaders
	}
	if !plan.ExpectedText.IsUnknown() {
		data.ExpectedText = plan.ExpectedText
	}
	if !plan.UnwantedText.IsUnknown() {
		data.UnwantedText = plan.UnwantedText
	}
	if !plan.SSLValidity.IsUnknown() {
		data.SSLValidity = plan.SSLValidity
	}
	if !plan.Cookies.IsUnknown() {
		data.Cookies = plan.Cookies
	}
	if !plan.PostParams.IsUnknown() {
		data.PostParams = plan.PostParams
	}
	if !plan.CustomRequestHeaders.IsUnknown() {
		data.CustomRequestHeaders = plan.CustomRequestHeaders
	}
	if !plan.UserAgent.IsUnknown() {
		data.UserAgent = plan.UserAgent
	}
	if !plan.ForceResolve.IsUnknown() {
		data.ForceResolve = plan.ForceResolve
	}
}
