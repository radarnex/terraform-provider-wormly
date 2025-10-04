package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/radarnex/terraform-provider-wormly/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &hostResource{}
	_ resource.ResourceWithConfigure   = &hostResource{}
	_ resource.ResourceWithImportState = &hostResource{}
)

// hostResourceModel represents the resource data model.
type hostResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	TestInterval types.Int64  `tfsdk:"test_interval"`
	Enabled      types.Bool   `tfsdk:"enabled"`
}

// hostResource defines the resource implementation.
type hostResource struct {
	client client.HostAPI
}

// NewHostResource creates a new host resource.
func NewHostResource() resource.Resource {
	return &hostResource{}
}

func (r *hostResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

func (r *hostResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Wormly host resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Host identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Host name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"test_interval": schema.Int64Attribute{
				MarkdownDescription: "Test interval in seconds",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(60),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the host is enabled",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
		},
	}
}

func (r *hostResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(client.HostAPI)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected client.HostAPI, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *hostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data hostResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the host
	host, err := r.client.CreateHost(ctx, data.Name.ValueString(), int(data.TestInterval.ValueInt64()), data.Enabled.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create host, got error: %s", err))
		return
	}

	// Set the basic resource state
	data.ID = types.StringValue(strconv.Itoa(host.ID))
	data.Name = types.StringValue(host.Name)
	data.TestInterval = types.Int64Value(int64(host.TestInterval))

	// Apply the desired enabled state through the monitoring APIs
	desiredEnabled := data.Enabled.ValueBool()
	if desiredEnabled {
		// Enable monitoring to ensure the host is in the desired state
		err := r.client.EnableHostUptimeMonitoring(ctx, host.ID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to enable host uptime monitoring: %s", err))
			return
		}
		data.Enabled = types.BoolValue(true)
	} else {
		// Disable monitoring to set to the desired state
		err := r.client.DisableHostUptimeMonitoring(ctx, host.ID)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to disable host uptime monitoring: %s", err))
			return
		}
		data.Enabled = types.BoolValue(false)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *hostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data hostResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the ID
	id, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse host ID: %s", err))
		return
	}

	// Get the host
	host, err := r.client.GetHost(ctx, id)
	if err != nil {
		// Check if this is a 404 error
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read host, got error: %s", err))
		return
	}

	// Update the model with the latest data
	data.Name = types.StringValue(host.Name)
	data.TestInterval = types.Int64Value(int64(host.TestInterval))
	data.Enabled = types.BoolValue(host.Enabled)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *hostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state hostResourceModel

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

	// Parse the host ID from the current state (not from plan, since ID is computed)
	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse host ID: %s", err))
		return
	}

	// Handle enabled state changes
	if !data.Enabled.Equal(state.Enabled) {
		if !data.Enabled.ValueBool() {
			// Host is being disabled - disable uptime monitoring
			err := r.client.DisableHostUptimeMonitoring(ctx, id)
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to disable host uptime monitoring: %s", err))
				return
			}
		} else {
			// Host is being enabled - enable uptime monitoring
			err := r.client.EnableHostUptimeMonitoring(ctx, id)
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to enable host uptime monitoring: %s", err))
				return
			}
		}
	}

	// Preserve all values from the current state and only update the enabled field from the plan
	// Note: name and test_interval have RequiresReplace, so they should not change in an update
	updatedState := hostResourceModel{
		ID:           state.ID,
		Name:         state.Name,
		TestInterval: state.TestInterval,
		Enabled:      data.Enabled, // This is the only field that can change in an update
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &updatedState)...)
}

func (r *hostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data hostResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the ID
	id, err := strconv.Atoi(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse host ID: %s", err))
		return
	}

	// Delete the host
	err = r.client.DeleteHost(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete host, got error: %s", err))
		return
	}
}

func (r *hostResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Set the ID from the import identifier
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// isNotFoundError checks if an error represents a 404 Not Found response.
func isNotFoundError(err error) bool {
	// This is a simple implementation - in a real scenario, you would check
	// the actual HTTP response status code
	return err != nil && err.Error() == "404 Not Found"
}
