package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vetal-ca-org/terraform-provider-voipms/internal/client"
)

func NewCallbackDataSource() datasource.DataSource  { return &callbackDataSource{} }
func NewCallbacksDataSource() datasource.DataSource { return &callbacksDataSource{} }

type callbackDataSource struct{ client *client.Client }
type callbacksDataSource struct{ client *client.Client }

func callbackDataSourceAttributes(lookup bool) map[string]schema.Attribute {
	id := schema.StringAttribute{MarkdownDescription: "Callback id.", Computed: true}
	if lookup {
		id.Required = true
		id.Computed = false
	}
	return map[string]schema.Attribute{
		"id":               id,
		"description":      dsString("Description."),
		"number":           dsString("Callback destination number."),
		"delay_before":     dsInt("Delay before prompt, in seconds."),
		"response_timeout": dsInt("Response timeout in seconds."),
		"digit_timeout":    dsInt("Digit timeout in seconds."),
		"callerid_number":  dsString("Caller ID presented to the callback number."),
	}
}

func (d *callbackDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_callback"
}
func (d *callbackDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a callback by id (`getCallbacks`).",
		Attributes:          callbackDataSourceAttributes(true),
	}
}
func (d *callbackDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}
func (d *callbackDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data callbackModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := d.client.GetCallback(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VoIP.ms callback", err.Error())
		return
	}
	flattenCallback(got, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type callbacksModel struct {
	ID        types.String    `tfsdk:"id"`
	Callbacks []callbackModel `tfsdk:"callbacks"`
}

func (d *callbacksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_callbacks"
}
func (d *callbacksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists callbacks (`getCallbacks`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{MarkdownDescription: "Placeholder identifier (`callbacks`).", Computed: true},
			"callbacks": schema.ListNestedAttribute{
				MarkdownDescription: "Callbacks on this account.",
				Computed:            true,
				NestedObject:        schema.NestedAttributeObject{Attributes: callbackDataSourceAttributes(false)},
			},
		},
	}
}
func (d *callbacksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}
func (d *callbacksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data callbacksModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	items, err := d.client.GetCallbacks(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError("Unable to list VoIP.ms callbacks", err.Error())
		return
	}
	data.ID = types.StringValue("callbacks")
	data.Callbacks = make([]callbackModel, 0, len(items))
	for i := range items {
		data.Callbacks = append(data.Callbacks, flattenCallbackCopy(&items[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
