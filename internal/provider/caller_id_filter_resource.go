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
	_ resource.Resource                = &callerIDFilterResource{}
	_ resource.ResourceWithConfigure   = &callerIDFilterResource{}
	_ resource.ResourceWithImportState = &callerIDFilterResource{}
)

func NewCallerIDFilterResource() resource.Resource { return &callerIDFilterResource{} }

type callerIDFilterResource struct{ client *client.Client }

type callerIDFilterModel struct {
	ID                  types.String `tfsdk:"id"`
	CallerID            types.String `tfsdk:"callerid"`
	DID                 types.String `tfsdk:"did"`
	Routing             types.String `tfsdk:"routing"`
	FailoverUnreachable types.String `tfsdk:"failover_unreachable"`
	FailoverBusy        types.String `tfsdk:"failover_busy"`
	FailoverNoanswer    types.String `tfsdk:"failover_noanswer"`
	Note                types.String `tfsdk:"note"`
}

func (r *callerIDFilterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_caller_id_filter"
}

func (r *callerIDFilterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a caller-ID filter rule (`setCallerIDFiltering` / `delCallerIDFiltering`). Use `X` as a wildcard digit in `callerid`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "VoIP.ms filter id.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"callerid":             schema.StringAttribute{MarkdownDescription: "Caller ID pattern (`X` matches any digit).", Required: true},
			"did":                  schema.StringAttribute{MarkdownDescription: "DID this rule applies to, or `all`.", Required: true},
			"routing":              schema.StringAttribute{MarkdownDescription: "Route for matched calls, e.g. `sys:hangup`.", Required: true},
			"failover_unreachable": optStr("Unreachable failover route."),
			"failover_busy":        optStr("Busy failover route."),
			"failover_noanswer":    optStr("No-answer failover route."),
			"note":                 optStr("Note."),
		},
	}
}

func (r *callerIDFilterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

func (r *callerIDFilterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan callerIDFilterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	params := callerIDFilterWriteParams(plan)
	delete(params, "filtering")
	if err := r.client.SetCallerIDFilter(ctx, params); err != nil {
		resp.Diagnostics.AddError("Unable to create VoIP.ms caller ID filter", err.Error())
		return
	}
	items, err := r.client.GetCallerIDFilters(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError("Caller ID filter created but could not be listed", err.Error())
		return
	}
	found := client.FindCallerIDFilterAfterCreate(items, plan.CallerID.ValueString(), plan.Note.ValueString())
	if found == nil {
		resp.Diagnostics.AddError("Caller ID filter created but could not be found", "The API did not return the new filter id.")
		return
	}
	flattenCallerIDFilter(found, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *callerIDFilterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state callerIDFilterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetCallerIDFilter(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read VoIP.ms caller ID filter", err.Error())
		return
	}
	flattenCallerIDFilter(got, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *callerIDFilterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan callerIDFilterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	current, err := r.client.GetCallerIDFilter(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VoIP.ms caller ID filter before update", err.Error())
		return
	}
	params := overlayParams(current.SetParams(), callerIDFilterWriteParams(plan))
	params["filtering"] = plan.ID.ValueString()
	if err := r.client.SetCallerIDFilter(ctx, params); err != nil {
		resp.Diagnostics.AddError("Unable to update VoIP.ms caller ID filter", err.Error())
		return
	}
	got, err := r.client.GetCallerIDFilter(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Caller ID filter updated but could not be read back", err.Error())
		return
	}
	flattenCallerIDFilter(got, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *callerIDFilterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state callerIDFilterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteCallerIDFilter(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete VoIP.ms caller ID filter", err.Error())
	}
}

func (r *callerIDFilterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func callerIDFilterWriteParams(m callerIDFilterModel) map[string]string {
	params := map[string]string{}
	setString(params, "callerid", m.CallerID)
	setString(params, "did", m.DID)
	setString(params, "routing", m.Routing)
	setString(params, "failover_unreachable", m.FailoverUnreachable)
	setString(params, "failover_busy", m.FailoverBusy)
	setString(params, "failover_noanswer", m.FailoverNoanswer)
	setString(params, "note", m.Note)
	return params
}

func flattenCallerIDFilter(src *client.CallerIDFilter, dst *callerIDFilterModel) {
	dst.ID = strVal(src.Filtering)
	dst.CallerID = strVal(src.CallerID)
	dst.DID = strVal(src.DID)
	dst.Routing = strVal(src.Routing)
	dst.FailoverUnreachable = strVal(src.FailoverUnreachable)
	dst.FailoverBusy = strVal(src.FailoverBusy)
	dst.FailoverNoanswer = strVal(src.FailoverNoanswer)
	dst.Note = strVal(src.Note)
}

func flattenCallerIDFilterCopy(src *client.CallerIDFilter) callerIDFilterModel {
	var m callerIDFilterModel
	flattenCallerIDFilter(src, &m)
	return m
}
