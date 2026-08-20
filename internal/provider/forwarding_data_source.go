package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vetal-ca-org/terraform-provider-voipms/internal/client"
)

func NewForwardingDataSource() datasource.DataSource  { return &forwardingDataSource{} }
func NewForwardingsDataSource() datasource.DataSource { return &forwardingsDataSource{} }

type forwardingDataSource struct{ client *client.Client }
type forwardingsDataSource struct{ client *client.Client }

func forwardingDataSourceAttributes(lookup bool) map[string]schema.Attribute {
	id := schema.StringAttribute{MarkdownDescription: "Forwarding id.", Computed: true}
	if lookup {
		id.Required = true
		id.Computed = false
	}
	return map[string]schema.Attribute{
		"id":                id,
		"phone_number":      dsString("Destination number."),
		"callerid_override": dsString("Caller ID override."),
		"description":       dsString("Description."),
		"dtmf_digits":       dsString("DTMF digits sent after answer."),
		"pause":             dsString("Pause before DTMF."),
		"diversion_header":  dsBool("SIP Diversion header enabled."),
	}
}

func (d *forwardingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_forwarding"
}
func (d *forwardingDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a call forwarding by id (`getForwardings`).",
		Attributes:          forwardingDataSourceAttributes(true),
	}
}
func (d *forwardingDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}
func (d *forwardingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data forwardingModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := d.client.GetForwarding(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VoIP.ms forwarding", err.Error())
		return
	}
	flattenForwarding(got, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type forwardingsModel struct {
	ID          types.String      `tfsdk:"id"`
	Forwardings []forwardingModel `tfsdk:"forwardings"`
}

func (d *forwardingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_forwardings"
}
func (d *forwardingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists call forwardings (`getForwardings`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{MarkdownDescription: "Placeholder identifier (`forwardings`).", Computed: true},
			"forwardings": schema.ListNestedAttribute{
				MarkdownDescription: "Forwarding destinations on this account.",
				Computed:            true,
				NestedObject:        schema.NestedAttributeObject{Attributes: forwardingDataSourceAttributes(false)},
			},
		},
	}
}
func (d *forwardingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}
func (d *forwardingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data forwardingsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	items, err := d.client.GetForwardings(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError("Unable to list VoIP.ms forwardings", err.Error())
		return
	}
	data.ID = types.StringValue("forwardings")
	data.Forwardings = make([]forwardingModel, 0, len(items))
	for i := range items {
		data.Forwardings = append(data.Forwardings, flattenForwardingCopy(&items[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
