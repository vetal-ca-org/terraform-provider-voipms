package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vetal-ca-org/terraform-provider-voipms/internal/client"
)

var (
	_ resource.Resource                = &forwardingResource{}
	_ resource.ResourceWithConfigure   = &forwardingResource{}
	_ resource.ResourceWithImportState = &forwardingResource{}
)

func NewForwardingResource() resource.Resource { return &forwardingResource{} }

type forwardingResource struct{ client *client.Client }

type forwardingModel struct {
	ID               types.String `tfsdk:"id"`
	PhoneNumber      types.String `tfsdk:"phone_number"`
	CallerIDOverride types.String `tfsdk:"callerid_override"`
	Description      types.String `tfsdk:"description"`
	DTMFDigits       types.String `tfsdk:"dtmf_digits"`
	Pause            types.String `tfsdk:"pause"`
	DiversionHeader  types.Bool   `tfsdk:"diversion_header"`
}

func (r *forwardingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_forwarding"
}

func (r *forwardingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a call-forwarding destination (`setForwarding` / `delForwarding`). Reference it from a DID as `fwd:{id}`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "VoIP.ms forwarding id (used in DID routing as `fwd:<id>`).",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"phone_number":      schema.StringAttribute{MarkdownDescription: "Destination number.", Required: true},
			"callerid_override": optStr("Caller ID override when forwarding."),
			"description":       optStr("Description."),
			"dtmf_digits":       optStr("DTMF digits to send after answer."),
			"pause":             optStr("Pause in seconds before DTMF (0–10, steps of 0.5)."),
			"diversion_header":  optBoolAttr("Send a SIP Diversion header."),
		},
	}
}

func (r *forwardingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

func (r *forwardingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan forwardingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	params := forwardingWriteParams(plan)
	delete(params, "forwarding")
	if err := r.client.SetForwarding(ctx, params); err != nil {
		resp.Diagnostics.AddError("Unable to create VoIP.ms forwarding", err.Error())
		return
	}
	items, err := r.client.GetForwardings(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError("Forwarding created but could not be listed", err.Error())
		return
	}
	found := client.FindForwardingAfterCreate(items, plan.PhoneNumber.ValueString(), plan.Description.ValueString())
	if found == nil {
		resp.Diagnostics.AddError("Forwarding created but could not be found", "The API did not return the new forwarding id.")
		return
	}
	flattenForwarding(found, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *forwardingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state forwardingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetForwarding(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read VoIP.ms forwarding", err.Error())
		return
	}
	flattenForwarding(got, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *forwardingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan forwardingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	current, err := r.client.GetForwarding(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VoIP.ms forwarding before update", err.Error())
		return
	}
	params := overlayParams(current.SetParams(), forwardingWriteParams(plan))
	params["forwarding"] = plan.ID.ValueString()
	if err := r.client.SetForwarding(ctx, params); err != nil {
		resp.Diagnostics.AddError("Unable to update VoIP.ms forwarding", err.Error())
		return
	}
	got, err := r.client.GetForwarding(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Forwarding updated but could not be read back", err.Error())
		return
	}
	flattenForwarding(got, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *forwardingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state forwardingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteForwarding(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete VoIP.ms forwarding", err.Error())
	}
}

func (r *forwardingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func forwardingWriteParams(m forwardingModel) map[string]string {
	params := map[string]string{}
	setString(params, "phone_number", m.PhoneNumber)
	setString(params, "callerid_override", m.CallerIDOverride)
	setString(params, "description", m.Description)
	setString(params, "dtmf_digits", m.DTMFDigits)
	setString(params, "pause", m.Pause)
	setBool01(params, "diversion_header", m.DiversionHeader)
	return params
}

func flattenForwarding(src *client.Forwarding, dst *forwardingModel) {
	dst.ID = strVal(src.Forwarding)
	dst.PhoneNumber = strVal(src.PhoneNumber)
	dst.CallerIDOverride = strVal(src.CallerIDOverride)
	dst.Description = strVal(src.Description)
	dst.DTMFDigits = strVal(src.DTMFDigits)
	dst.Pause = strVal(src.Pause)
	dst.DiversionHeader = boolVal(src.DiversionHeader)
}

func flattenForwardingCopy(src *client.Forwarding) forwardingModel {
	var m forwardingModel
	flattenForwarding(src, &m)
	return m
}
