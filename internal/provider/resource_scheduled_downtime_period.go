package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/radarnex/terraform-provider-wormly/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &scheduledDowntimePeriodResource{}
	_ resource.ResourceWithConfigure   = &scheduledDowntimePeriodResource{}
	_ resource.ResourceWithImportState = &scheduledDowntimePeriodResource{}
)

// scheduledDowntimePeriodResourceModel represents the resource data model.
type scheduledDowntimePeriodResourceModel struct {
	ID         types.String `tfsdk:"id"`
	HostID     types.Int64  `tfsdk:"hostid"`
	Start      types.String `tfsdk:"start"`
	End        types.String `tfsdk:"end"`
	Timezone   types.String `tfsdk:"timezone"`
	Recurrence types.String `tfsdk:"recurrence"`
	On         types.String `tfsdk:"on"`
}

// scheduledDowntimePeriodResource defines the resource implementation.
type scheduledDowntimePeriodResource struct {
	client client.ScheduledDowntimePeriodAPI
}

// NewScheduledDowntimePeriodResource creates a new scheduled downtime period resource.
func NewScheduledDowntimePeriodResource() resource.Resource {
	return &scheduledDowntimePeriodResource{}
}

func (r *scheduledDowntimePeriodResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scheduled_downtime_period"
}

func (r *scheduledDowntimePeriodResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Wormly scheduled downtime period resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Scheduled downtime period identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hostid": schema.Int64Attribute{
				MarkdownDescription: "The ID of the host to schedule downtime for",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"start": schema.StringAttribute{
				MarkdownDescription: "The starting time of the period in HH:mm format (24-hour clock)",
				Required:            true,
			},
			"end": schema.StringAttribute{
				MarkdownDescription: "The ending time of the period in HH:mm format (24-hour clock)",
				Required:            true,
			},
			"timezone": schema.StringAttribute{
				MarkdownDescription: "The POSIX standard timezone of the start and end times (e.g., 'GMT', 'Europe/London')",
				Required:            true,
			},
			"recurrence": schema.StringAttribute{
				MarkdownDescription: "The recurrence pattern. Must be one of ONCEONLY, DAILY, WEEKLY, or MONTHLY",
				Required:            true,
			},
			"on": schema.StringAttribute{
				MarkdownDescription: "The specific day for the downtime. For ONCEONLY recurrence, this is a date in YYYY-MM-DD format. For WEEKLY recurrence, this is the day of the week (e.g., 'Sunday'). For MONTHLY recurrence, this is the day of the month (1-31 or 'LASTDAY'). This argument should be omitted for DAILY recurrence.",
				Optional:            true,
			},
		},
	}
}

func (r *scheduledDowntimePeriodResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(client.ScheduledDowntimePeriodAPI)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected client.ScheduledDowntimePeriodAPI, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *scheduledDowntimePeriodResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data scheduledDowntimePeriodResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the scheduled downtime period
	period, err := r.client.CreateScheduledDowntimePeriod(
		ctx,
		int(data.HostID.ValueInt64()),
		data.Start.ValueString(),
		data.End.ValueString(),
		data.Timezone.ValueString(),
		data.Recurrence.ValueString(),
		data.On.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create scheduled downtime period, got error: %s", err))
		return
	}

	// Set the resource state
	data.ID = types.StringValue(strconv.Itoa(period.ID))
	data.HostID = types.Int64Value(int64(period.HostID))
	data.Start = types.StringValue(period.Start)
	data.End = types.StringValue(period.End)
	data.Timezone = types.StringValue(period.Timezone)
	data.Recurrence = types.StringValue(period.Recurrence)
	if period.On != "" {
		data.On = types.StringValue(period.On)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *scheduledDowntimePeriodResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data scheduledDowntimePeriodResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the ID
	id, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse scheduled downtime period ID: %s", err))
		return
	}

	// Get the scheduled downtime period
	period, err := r.client.GetScheduledDowntimePeriod(ctx, int(data.HostID.ValueInt64()), id)
	if err != nil {
		// Check if this is a not found error
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read scheduled downtime period, got error: %s", err))
		return
	}

	// Update the model with the latest data
	data.HostID = types.Int64Value(int64(period.HostID))
	data.Start = types.StringValue(period.Start)
	data.End = types.StringValue(period.End)
	data.Timezone = types.StringValue(period.Timezone)
	data.Recurrence = types.StringValue(period.Recurrence)
	if period.On != "" {
		data.On = types.StringValue(period.On)
	} else {
		data.On = types.StringNull()
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *scheduledDowntimePeriodResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state scheduledDowntimePeriodResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read current state data
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the period ID from the current state (not from plan, since ID is computed)
	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse scheduled downtime period ID: %s", err))
		return
	}

	// Update the scheduled downtime period
	period, err := r.client.UpdateScheduledDowntimePeriod(
		ctx,
		int(data.HostID.ValueInt64()),
		id,
		data.Start.ValueString(),
		data.End.ValueString(),
		data.Timezone.ValueString(),
		data.Recurrence.ValueString(),
		data.On.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update scheduled downtime period, got error: %s", err))
		return
	}

	// Update the model with the response data
	data.ID = types.StringValue(strconv.Itoa(period.ID))
	data.HostID = types.Int64Value(int64(period.HostID))
	data.Start = types.StringValue(period.Start)
	data.End = types.StringValue(period.End)
	data.Timezone = types.StringValue(period.Timezone)
	data.Recurrence = types.StringValue(period.Recurrence)
	if period.On != "" {
		data.On = types.StringValue(period.On)
	} else {
		data.On = types.StringNull()
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *scheduledDowntimePeriodResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data scheduledDowntimePeriodResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the ID
	id, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse scheduled downtime period ID: %s", err))
		return
	}

	// Delete the scheduled downtime period
	err = r.client.DeleteScheduledDowntimePeriod(ctx, int(data.HostID.ValueInt64()), id)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete scheduled downtime period, got error: %s", err))
		return
	}
}

func (r *scheduledDowntimePeriodResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Parse the import ID in the format "host_id/period_id"
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be in the format 'host_id/period_id'",
		)
		return
	}

	// Parse host ID
	hostID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Host ID",
			fmt.Sprintf("Unable to parse host ID '%s': %s", parts[0], err),
		)
		return
	}

	// Validate period ID is numeric
	_, err = strconv.Atoi(parts[1])
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Period ID",
			fmt.Sprintf("Unable to parse period ID '%s': %s", parts[1], err),
		)
		return
	}

	// Set the hostid and id in the state
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("hostid"), hostID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
