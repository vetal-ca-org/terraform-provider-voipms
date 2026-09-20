package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vetal-ca-org/terraform-provider-voipms/internal/client"
)

func NewRingGroupDataSource() datasource.DataSource  { return &ringGroupDataSource{} }
func NewRingGroupsDataSource() datasource.DataSource { return &ringGroupsDataSource{} }

type ringGroupDataSource struct{ client *client.Client }
type ringGroupsDataSource struct{ client *client.Client }

func ringGroupDataSourceAttributes(lookup bool) map[string]schema.Attribute {
	id := schema.StringAttribute{MarkdownDescription: "Ring group id.", Computed: true}
	name := schema.StringAttribute{MarkdownDescription: "Ring group name.", Computed: true}
	if lookup {
		id.Optional = true
		name.Optional = true
	}
	return map[string]schema.Attribute{
		"id":                  id,
		"name":                name,
		"route":               dsString("DID routing value (`grp:{id}`). Use this for `voipms_did` `routing` / failover."),
		"voicemail":           dsString("Mailbox that takes the call when nobody answers."),
		"caller_announcement": dsString("Recording code played to the caller."),
		"music_on_hold":       dsString("Music on hold class."),
		"language":            dsString("Language for system messages."),
		"members": schema.ListNestedAttribute{
			MarkdownDescription: "Destinations rung by this group, in order.",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"route":     dsString("Member route (`account:` or `fwd:`)."),
					"ring_time": dsInt("Seconds this member rings."),
					"press1":    dsBool("Member must press 1 to connect."),
				},
			},
		},
	}
}

func (d *ringGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ring_group"
}
func (d *ringGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a ring group by id or `name` (`getRingGroups`). " +
			"Look up by `name`, then link with `route` — do not paste a raw ring group id into a DID.",
		Attributes: ringGroupDataSourceAttributes(true),
	}
}
func (d *ringGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}
func (d *ringGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ringGroupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	query := exactlyOneLookup(&resp.Diagnostics, "ring group", []lookupField{
		{Name: "id", Value: configuredString(data.ID)},
		{Name: "name", Value: configuredString(data.Name)},
	})
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := d.client.FindRingGroup(ctx, query)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VoIP.ms ring group", err.Error())
		return
	}
	flattenRingGroup(got, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type ringGroupsModel struct {
	ID         types.String     `tfsdk:"id"`
	RingGroups []ringGroupModel `tfsdk:"ring_groups"`
}

func (d *ringGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ring_groups"
}
func (d *ringGroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists ring groups (`getRingGroups`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{MarkdownDescription: "Placeholder identifier (`ring_groups`).", Computed: true},
			"ring_groups": schema.ListNestedAttribute{
				MarkdownDescription: "Ring groups on this account.",
				Computed:            true,
				NestedObject:        schema.NestedAttributeObject{Attributes: ringGroupDataSourceAttributes(false)},
			},
		},
	}
}
func (d *ringGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}
func (d *ringGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ringGroupsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	items, err := d.client.GetRingGroups(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError("Unable to list VoIP.ms ring groups", err.Error())
		return
	}
	data.ID = types.StringValue("ring_groups")
	data.RingGroups = make([]ringGroupModel, 0, len(items))
	for i := range items {
		data.RingGroups = append(data.RingGroups, flattenRingGroupCopy(&items[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
