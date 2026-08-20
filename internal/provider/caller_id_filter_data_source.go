package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vetal-ca-org/terraform-provider-voipms/internal/client"
)

func NewCallerIDFilterDataSource() datasource.DataSource  { return &callerIDFilterDataSource{} }
func NewCallerIDFiltersDataSource() datasource.DataSource { return &callerIDFiltersDataSource{} }

type callerIDFilterDataSource struct{ client *client.Client }
type callerIDFiltersDataSource struct{ client *client.Client }

func callerIDFilterDataSourceAttributes(lookup bool) map[string]schema.Attribute {
	id := schema.StringAttribute{MarkdownDescription: "Filter id.", Computed: true}
	if lookup {
		id.Required = true
		id.Computed = false
	}
	return map[string]schema.Attribute{
		"id":                   id,
		"callerid":             dsString("Caller ID pattern."),
		"did":                  dsString("DID this rule applies to, or `all`."),
		"routing":              dsString("Route for matched calls."),
		"failover_unreachable": dsString("Unreachable failover route."),
		"failover_busy":        dsString("Busy failover route."),
		"failover_noanswer":    dsString("No-answer failover route."),
		"note":                 dsString("Note."),
	}
}

func (d *callerIDFilterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_caller_id_filter"
}
func (d *callerIDFilterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a caller-ID filter by id (`getCallerIDFiltering`).",
		Attributes:          callerIDFilterDataSourceAttributes(true),
	}
}
func (d *callerIDFilterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}
func (d *callerIDFilterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data callerIDFilterModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := d.client.GetCallerIDFilter(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VoIP.ms caller ID filter", err.Error())
		return
	}
	flattenCallerIDFilter(got, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type callerIDFiltersModel struct {
	ID      types.String          `tfsdk:"id"`
	Filters []callerIDFilterModel `tfsdk:"filters"`
}

func (d *callerIDFiltersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_caller_id_filters"
}
func (d *callerIDFiltersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists caller-ID filter rules (`getCallerIDFiltering`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{MarkdownDescription: "Placeholder identifier (`caller_id_filters`).", Computed: true},
			"filters": schema.ListNestedAttribute{
				MarkdownDescription: "Caller-ID filters on this account.",
				Computed:            true,
				NestedObject:        schema.NestedAttributeObject{Attributes: callerIDFilterDataSourceAttributes(false)},
			},
		},
	}
}
func (d *callerIDFiltersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}
func (d *callerIDFiltersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data callerIDFiltersModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	items, err := d.client.GetCallerIDFilters(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError("Unable to list VoIP.ms caller ID filters", err.Error())
		return
	}
	data.ID = types.StringValue("caller_id_filters")
	data.Filters = make([]callerIDFilterModel, 0, len(items))
	for i := range items {
		data.Filters = append(data.Filters, flattenCallerIDFilterCopy(&items[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
