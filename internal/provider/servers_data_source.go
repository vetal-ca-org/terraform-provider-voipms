package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vetal-ca-org/terraform-provider-voipms/internal/client"
)

func NewServerDataSource() datasource.DataSource  { return &serverDataSource{} }
func NewServersDataSource() datasource.DataSource { return &serversDataSource{} }

type serverDataSource struct{ client *client.Client }
type serversDataSource struct{ client *client.Client }

type serverModel struct {
	ID          types.String `tfsdk:"id"`
	POP         types.String `tfsdk:"pop"`
	Name        types.String `tfsdk:"name"`
	Shortname   types.String `tfsdk:"shortname"`
	Hostname    types.String `tfsdk:"hostname"`
	IP          types.String `tfsdk:"ip"`
	Country     types.String `tfsdk:"country"`
	Recommended types.Bool   `tfsdk:"recommended"`
}

func serverDataSourceAttributes(lookup bool) map[string]schema.Attribute {
	pop := schema.StringAttribute{MarkdownDescription: "Point-of-presence id (`server_pop`).", Computed: true}
	if lookup {
		pop.Required = true
		pop.Computed = false
	}
	return map[string]schema.Attribute{
		"id":          schema.StringAttribute{MarkdownDescription: "Same as `pop`.", Computed: true},
		"pop":         pop,
		"name":        dsString("Server display name."),
		"shortname":   dsString("Short server name."),
		"hostname":    dsString("SIP hostname (e.g. `newyork7.voip.ms`)."),
		"ip":          dsString("Server IP address."),
		"country":     dsString("Server country."),
		"recommended": dsBool("Whether VoIP.ms marks this POP as recommended."),
	}
}

func flattenServer(src *client.Server, dst *serverModel) {
	dst.ID = strVal(src.POP)
	dst.POP = strVal(src.POP)
	dst.Name = strVal(src.Name)
	dst.Shortname = strVal(src.Shortname)
	dst.Hostname = strVal(src.Hostname)
	dst.IP = strVal(src.IP)
	dst.Country = strVal(src.Country)
	dst.Recommended = boolVal(src.Recommended)
}

func flattenServerCopy(src *client.Server) serverModel {
	var m serverModel
	flattenServer(src, &m)
	return m
}

func (d *serverDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}
func (d *serverDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up a VoIP.ms POP by id (`getServersInfo`). Use this to map `pop = 73` to `newyork7.voip.ms`.",
		Attributes:          serverDataSourceAttributes(true),
	}
}
func (d *serverDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}
func (d *serverDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data serverModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := d.client.GetServer(ctx, data.POP.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VoIP.ms server", err.Error())
		return
	}
	flattenServer(got, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type serversModel struct {
	ID      types.String  `tfsdk:"id"`
	Servers []serverModel `tfsdk:"servers"`
}

func (d *serversDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_servers"
}
func (d *serversDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists VoIP.ms points of presence (`getServersInfo`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{MarkdownDescription: "Placeholder identifier (`servers`).", Computed: true},
			"servers": schema.ListNestedAttribute{
				MarkdownDescription: "Available POP servers.",
				Computed:            true,
				NestedObject:        schema.NestedAttributeObject{Attributes: serverDataSourceAttributes(false)},
			},
		},
	}
}
func (d *serversDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}
func (d *serversDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data serversModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	items, err := d.client.GetServersInfo(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError("Unable to list VoIP.ms servers", err.Error())
		return
	}
	data.ID = types.StringValue("servers")
	data.Servers = make([]serverModel, 0, len(items))
	for i := range items {
		data.Servers = append(data.Servers, flattenServerCopy(&items[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
