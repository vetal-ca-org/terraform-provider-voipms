package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vetal-ca-org/terraform-provider-voipms/internal/client"
)

func NewPhonebookEntryDataSource() datasource.DataSource   { return &phonebookEntryDataSource{} }
func NewPhonebookEntriesDataSource() datasource.DataSource { return &phonebookEntriesDataSource{} }
func NewPhonebookGroupDataSource() datasource.DataSource   { return &phonebookGroupDataSource{} }
func NewPhonebookGroupsDataSource() datasource.DataSource  { return &phonebookGroupsDataSource{} }

type phonebookEntryDataSource struct{ client *client.Client }
type phonebookEntriesDataSource struct{ client *client.Client }
type phonebookGroupDataSource struct{ client *client.Client }
type phonebookGroupsDataSource struct{ client *client.Client }

func phonebookEntryDataSourceAttributes(lookup bool) map[string]schema.Attribute {
	id := schema.StringAttribute{MarkdownDescription: "Phonebook entry id.", Computed: true}
	if lookup {
		id.Required = true
		id.Computed = false
	}
	return map[string]schema.Attribute{
		"id":         id,
		"speed_dial": dsString("Speed-dial code."),
		"name":       dsString("Contact name."),
		"number":     dsString("Phone number or prefix."),
		"callerid":   dsString("Caller ID name override."),
		"note":       dsString("Note."),
		"group":      dsString("Phonebook group id."),
		"group_name": dsString("Phonebook group name."),
	}
}

func phonebookGroupDataSourceAttributes(lookup bool) map[string]schema.Attribute {
	id := schema.StringAttribute{MarkdownDescription: "Phonebook group id.", Computed: true}
	if lookup {
		id.Required = true
		id.Computed = false
	}
	return map[string]schema.Attribute{
		"id":      id,
		"name":    dsString("Group name."),
		"members": dsString("Comma-separated phonebook entry ids."),
	}
}

func (d *phonebookEntryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_phonebook_entry"
}
func (d *phonebookEntryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a phonebook entry by id (`getPhonebook`).",
		Attributes:          phonebookEntryDataSourceAttributes(true),
	}
}
func (d *phonebookEntryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}
func (d *phonebookEntryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data phonebookEntryModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := d.client.GetPhonebookEntry(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VoIP.ms phonebook entry", err.Error())
		return
	}
	flattenPhonebookEntry(got, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type phonebookEntriesModel struct {
	ID      types.String          `tfsdk:"id"`
	Entries []phonebookEntryModel `tfsdk:"entries"`
}

func (d *phonebookEntriesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_phonebook_entries"
}
func (d *phonebookEntriesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists phonebook entries (`getPhonebook`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{MarkdownDescription: "Placeholder identifier (`phonebook_entries`).", Computed: true},
			"entries": schema.ListNestedAttribute{
				MarkdownDescription: "Phonebook entries on this account.",
				Computed:            true,
				NestedObject:        schema.NestedAttributeObject{Attributes: phonebookEntryDataSourceAttributes(false)},
			},
		},
	}
}
func (d *phonebookEntriesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}
func (d *phonebookEntriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data phonebookEntriesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	items, err := d.client.GetPhonebook(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError("Unable to list VoIP.ms phonebook entries", err.Error())
		return
	}
	data.ID = types.StringValue("phonebook_entries")
	data.Entries = make([]phonebookEntryModel, 0, len(items))
	for i := range items {
		data.Entries = append(data.Entries, flattenPhonebookEntryCopy(&items[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *phonebookGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_phonebook_group"
}
func (d *phonebookGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a phonebook group by id (`getPhonebookGroups`).",
		Attributes:          phonebookGroupDataSourceAttributes(true),
	}
}
func (d *phonebookGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}
func (d *phonebookGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data phonebookGroupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := d.client.GetPhonebookGroup(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VoIP.ms phonebook group", err.Error())
		return
	}
	flattenPhonebookGroup(got, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type phonebookGroupsModel struct {
	ID     types.String          `tfsdk:"id"`
	Groups []phonebookGroupModel `tfsdk:"groups"`
}

func (d *phonebookGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_phonebook_groups"
}
func (d *phonebookGroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists phonebook groups (`getPhonebookGroups`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{MarkdownDescription: "Placeholder identifier (`phonebook_groups`).", Computed: true},
			"groups": schema.ListNestedAttribute{
				MarkdownDescription: "Phonebook groups on this account.",
				Computed:            true,
				NestedObject:        schema.NestedAttributeObject{Attributes: phonebookGroupDataSourceAttributes(false)},
			},
		},
	}
}
func (d *phonebookGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}
func (d *phonebookGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data phonebookGroupsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	items, err := d.client.GetPhonebookGroups(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError("Unable to list VoIP.ms phonebook groups", err.Error())
		return
	}
	data.ID = types.StringValue("phonebook_groups")
	data.Groups = make([]phonebookGroupModel, 0, len(items))
	for i := range items {
		data.Groups = append(data.Groups, flattenPhonebookGroupCopy(&items[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
