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
	_ resource.Resource                = &callbackResource{}
	_ resource.ResourceWithConfigure   = &callbackResource{}
	_ resource.ResourceWithImportState = &callbackResource{}
)

func NewCallbackResource() resource.Resource { return &callbackResource{} }

type callbackResource struct{ client *client.Client }

type callbackModel struct {
	ID              types.String `tfsdk:"id"`
	Description     types.String `tfsdk:"description"`
	Number          types.String `tfsdk:"number"`
	DelayBefore     types.Int64  `tfsdk:"delay_before"`
	ResponseTimeout types.Int64  `tfsdk:"response_timeout"`
	DigitTimeout    types.Int64  `tfsdk:"digit_timeout"`
	CallerIDNumber  types.String `tfsdk:"callerid_number"`
}

func (r *callbackResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_callback"
}

func (r *callbackResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a callback (`setCallback` / `delCallback`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "VoIP.ms callback id.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"number":           schema.StringAttribute{MarkdownDescription: "Number VoIP.ms calls to start the callback.", Required: true},
			"description":      optStr("Description."),
			"delay_before":     optIntAttr("Seconds to wait before sending the callback prompt."),
			"response_timeout": optIntAttr("Seconds to wait for DTMF after the prompt."),
			"digit_timeout":    optIntAttr("Seconds to wait between DTMF digits."),
			"callerid_number":  optStr("Caller ID presented to the callback number."),
		},
	}
}

func (r *callbackResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

func (r *callbackResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan callbackModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	params := callbackWriteParams(plan)
	delete(params, "callback")
	if err := r.client.SetCallback(ctx, params); err != nil {
		resp.Diagnostics.AddError("Unable to create VoIP.ms callback", err.Error())
		return
	}
	items, err := r.client.GetCallbacks(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError("Callback created but could not be listed", err.Error())
		return
	}
	found := client.FindCallbackAfterCreate(items, plan.Number.ValueString(), plan.Description.ValueString())
	if found == nil {
		resp.Diagnostics.AddError("Callback created but could not be found", "The API did not return the new callback id.")
		return
	}
	flattenCallback(found, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *callbackResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state callbackModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetCallback(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read VoIP.ms callback", err.Error())
		return
	}
	flattenCallback(got, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *callbackResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan callbackModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	current, err := r.client.GetCallback(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VoIP.ms callback before update", err.Error())
		return
	}
	params := overlayParams(current.SetParams(), callbackWriteParams(plan))
	params["callback"] = plan.ID.ValueString()
	if err := r.client.SetCallback(ctx, params); err != nil {
		resp.Diagnostics.AddError("Unable to update VoIP.ms callback", err.Error())
		return
	}
	got, err := r.client.GetCallback(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Callback updated but could not be read back", err.Error())
		return
	}
	flattenCallback(got, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *callbackResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state callbackModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteCallback(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete VoIP.ms callback", err.Error())
	}
}

func (r *callbackResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func callbackWriteParams(m callbackModel) map[string]string {
	params := map[string]string{}
	setString(params, "description", m.Description)
	setString(params, "number", m.Number)
	setInt(params, "delay_before", m.DelayBefore)
	setInt(params, "response_timeout", m.ResponseTimeout)
	setInt(params, "digit_timeout", m.DigitTimeout)
	setString(params, "callerid_number", m.CallerIDNumber)
	return params
}

func flattenCallback(src *client.Callback, dst *callbackModel) {
	dst.ID = strVal(src.Callback)
	dst.Description = strVal(src.Description)
	dst.Number = strVal(src.Number)
	dst.DelayBefore = intVal(src.DelayBefore)
	dst.ResponseTimeout = intVal(src.ResponseTimeout)
	dst.DigitTimeout = intVal(src.DigitTimeout)
	dst.CallerIDNumber = strVal(src.CallerIDNumber)
}

func flattenCallbackCopy(src *client.Callback) callbackModel {
	var m callbackModel
	flattenCallback(src, &m)
	return m
}
