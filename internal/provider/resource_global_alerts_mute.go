package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/radarnex/terraform-provider-wormly/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &globalAlertsMuteResource{}
	_ resource.ResourceWithConfigure = &globalAlertsMuteResource{}
)

// globalAlertsMuteResourceModel represents the resource data model.
type globalAlertsMuteResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

// globalAlertsMuteResource defines the resource implementation.
type globalAlertsMuteResource struct {
	client *client.Client
}

// NewGlobalAlertsMuteResource creates a new global alerts mute resource.
func NewGlobalAlertsMuteResource() resource.Resource {
	return &globalAlertsMuteResource{}
}

func (r *globalAlertsMuteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global_alerts_mute"
}

func (r *globalAlertsMuteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Wormly global alerts mute resource. Controls the global alert mute setting.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Resource identifier (always 'global_alerts_mute')",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether global alerts mute is enabled",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *globalAlertsMuteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *globalAlertsMuteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data globalAlertsMuteResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the ID to a constant value since this is a singleton resource
	data.ID = types.StringValue("global_alerts_mute")

	// Apply the global alerts mute setting
	enabled := data.Enabled.ValueBool()
	if err := r.client.SetGlobalAlertMute(ctx, enabled); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set global alerts mute, got error: %s", err))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *globalAlertsMuteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data globalAlertsMuteResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Since there's no API to read the current state, we keep the current state as-is
	// The resource represents the desired state that was last applied

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *globalAlertsMuteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data globalAlertsMuteResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve the ID from the prior state
	data.ID = types.StringValue("global_alerts_mute")

	// Apply the updated global alerts mute setting
	enabled := data.Enabled.ValueBool()
	if err := r.client.SetGlobalAlertMute(ctx, enabled); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update global alerts mute, got error: %s", err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *globalAlertsMuteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data globalAlertsMuteResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// On delete, disable global alerts mute (set to false)
	if err := r.client.SetGlobalAlertMute(ctx, false); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to disable global alerts mute, got error: %s", err))
		return
	}

	// The resource is now deleted from state automatically
}
