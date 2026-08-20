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
	_ resource.Resource                = &phonebookEntryResource{}
	_ resource.ResourceWithConfigure   = &phonebookEntryResource{}
	_ resource.ResourceWithImportState = &phonebookEntryResource{}
	_ resource.Resource                = &phonebookGroupResource{}
	_ resource.ResourceWithConfigure   = &phonebookGroupResource{}
	_ resource.ResourceWithImportState = &phonebookGroupResource{}
)

func NewPhonebookEntryResource() resource.Resource { return &phonebookEntryResource{} }
func NewPhonebookGroupResource() resource.Resource { return &phonebookGroupResource{} }

type phonebookEntryResource struct{ client *client.Client }
type phonebookGroupResource struct{ client *client.Client }

type phonebookEntryModel struct {
	ID        types.String `tfsdk:"id"`
	SpeedDial types.String `tfsdk:"speed_dial"`
	Name      types.String `tfsdk:"name"`
	Number    types.String `tfsdk:"number"`
	CallerID  types.String `tfsdk:"callerid"`
	Note      types.String `tfsdk:"note"`
	Group     types.String `tfsdk:"group"`
	GroupName types.String `tfsdk:"group_name"`
}

type phonebookGroupModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Members types.String `tfsdk:"members"`
}

func (r *phonebookEntryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_phonebook_entry"
}

func (r *phonebookEntryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a phonebook entry (`setPhonebook` / `delPhonebook`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "VoIP.ms phonebook entry id.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name":       schema.StringAttribute{MarkdownDescription: "Contact name.", Required: true},
			"number":     schema.StringAttribute{MarkdownDescription: "Phone number or prefix.", Required: true},
			"speed_dial": optStr("Speed-dial code."),
			"callerid":   optStr("Caller ID name override."),
			"note":       optStr("Note."),
			"group":      optStr("Phonebook group id."),
			"group_name": schema.StringAttribute{MarkdownDescription: "Group name (read-only).", Computed: true},
		},
	}
}

func (r *phonebookEntryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

func (r *phonebookEntryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan phonebookEntryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	params := phonebookEntryWriteParams(plan)
	delete(params, "phonebook")
	if err := r.client.SetPhonebook(ctx, params); err != nil {
		resp.Diagnostics.AddError("Unable to create VoIP.ms phonebook entry", err.Error())
		return
	}
	items, err := r.client.GetPhonebook(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError("Phonebook entry created but could not be listed", err.Error())
		return
	}
	found := client.FindPhonebookAfterCreate(items, plan.Number.ValueString(), plan.Name.ValueString())
	if found == nil {
		resp.Diagnostics.AddError("Phonebook entry created but could not be found", "The API did not return the new entry id.")
		return
	}
	flattenPhonebookEntry(found, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *phonebookEntryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state phonebookEntryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetPhonebookEntry(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read VoIP.ms phonebook entry", err.Error())
		return
	}
	flattenPhonebookEntry(got, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *phonebookEntryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan phonebookEntryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	current, err := r.client.GetPhonebookEntry(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VoIP.ms phonebook entry before update", err.Error())
		return
	}
	params := overlayParams(current.SetParams(), phonebookEntryWriteParams(plan))
	params["phonebook"] = plan.ID.ValueString()
	if err := r.client.SetPhonebook(ctx, params); err != nil {
		resp.Diagnostics.AddError("Unable to update VoIP.ms phonebook entry", err.Error())
		return
	}
	got, err := r.client.GetPhonebookEntry(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Phonebook entry updated but could not be read back", err.Error())
		return
	}
	flattenPhonebookEntry(got, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *phonebookEntryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state phonebookEntryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeletePhonebook(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete VoIP.ms phonebook entry", err.Error())
	}
}

func (r *phonebookEntryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func phonebookEntryWriteParams(m phonebookEntryModel) map[string]string {
	params := map[string]string{}
	setString(params, "speed_dial", m.SpeedDial)
	setString(params, "name", m.Name)
	setString(params, "number", m.Number)
	setString(params, "callerid", m.CallerID)
	setString(params, "note", m.Note)
	setString(params, "group", m.Group)
	return params
}

func flattenPhonebookEntry(src *client.PhonebookEntry, dst *phonebookEntryModel) {
	dst.ID = strVal(src.Phonebook)
	dst.SpeedDial = strVal(src.SpeedDial)
	dst.Name = strVal(src.Name)
	dst.Number = strVal(src.Number)
	dst.CallerID = strVal(src.CallerID)
	dst.Note = strVal(src.Note)
	dst.Group = strVal(src.Group)
	dst.GroupName = strVal(src.GroupName)
}

func flattenPhonebookEntryCopy(src *client.PhonebookEntry) phonebookEntryModel {
	var m phonebookEntryModel
	flattenPhonebookEntry(src, &m)
	return m
}

func (r *phonebookGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_phonebook_group"
}

func (r *phonebookGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a phonebook group (`setPhonebookGroup` / `delPhonebookGroup`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "VoIP.ms phonebook group id.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name":    schema.StringAttribute{MarkdownDescription: "Group name.", Required: true},
			"members": optStr("Comma-separated phonebook entry ids."),
		},
	}
}

func (r *phonebookGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

func (r *phonebookGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan phonebookGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	params := phonebookGroupWriteParams(plan)
	delete(params, "group")
	if err := r.client.SetPhonebookGroup(ctx, params); err != nil {
		resp.Diagnostics.AddError("Unable to create VoIP.ms phonebook group", err.Error())
		return
	}
	items, err := r.client.GetPhonebookGroups(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError("Phonebook group created but could not be listed", err.Error())
		return
	}
	found := client.FindPhonebookGroupAfterCreate(items, plan.Name.ValueString())
	if found == nil {
		resp.Diagnostics.AddError("Phonebook group created but could not be found", "The API did not return the new group id.")
		return
	}
	flattenPhonebookGroup(found, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *phonebookGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state phonebookGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetPhonebookGroup(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read VoIP.ms phonebook group", err.Error())
		return
	}
	flattenPhonebookGroup(got, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *phonebookGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan phonebookGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	current, err := r.client.GetPhonebookGroup(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VoIP.ms phonebook group before update", err.Error())
		return
	}
	params := overlayParams(current.SetParams(), phonebookGroupWriteParams(plan))
	params["group"] = plan.ID.ValueString()
	if err := r.client.SetPhonebookGroup(ctx, params); err != nil {
		resp.Diagnostics.AddError("Unable to update VoIP.ms phonebook group", err.Error())
		return
	}
	got, err := r.client.GetPhonebookGroup(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Phonebook group updated but could not be read back", err.Error())
		return
	}
	flattenPhonebookGroup(got, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *phonebookGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state phonebookGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeletePhonebookGroup(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete VoIP.ms phonebook group", err.Error())
	}
}

func (r *phonebookGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func phonebookGroupWriteParams(m phonebookGroupModel) map[string]string {
	params := map[string]string{}
	setString(params, "name", m.Name)
	setString(params, "members", m.Members)
	return params
}

func flattenPhonebookGroup(src *client.PhonebookGroup, dst *phonebookGroupModel) {
	dst.ID = strVal(src.PhonebookGroup)
	dst.Name = strVal(src.Name)
	dst.Members = strVal(src.Members)
}

func flattenPhonebookGroupCopy(src *client.PhonebookGroup) phonebookGroupModel {
	var m phonebookGroupModel
	flattenPhonebookGroup(src, &m)
	return m
}
