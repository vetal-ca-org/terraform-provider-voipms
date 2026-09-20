package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vetal-ca-org/terraform-provider-voipms/internal/client"
)

func NewTimeConditionDataSource() datasource.DataSource  { return &timeConditionDataSource{} }
func NewTimeConditionsDataSource() datasource.DataSource { return &timeConditionsDataSource{} }

type timeConditionDataSource struct{ client *client.Client }
type timeConditionsDataSource struct{ client *client.Client }

func timeConditionDataSourceAttributes(lookup bool) map[string]schema.Attribute {
	id := schema.StringAttribute{MarkdownDescription: "Time condition id.", Computed: true}
	name := schema.StringAttribute{MarkdownDescription: "Time condition name.", Computed: true}
	if lookup {
		id.Optional = true
		name.Optional = true
	}
	return map[string]schema.Attribute{
		"id":              id,
		"name":            name,
		"route":           dsString("DID routing value (`tc:{id}`). Use this for `voipms_did` `routing` / failover."),
		"routing_match":   dsString("Where the call goes inside any of the windows."),
		"routing_nomatch": dsString("Where the call goes outside every window."),
		"windows": schema.ListNestedAttribute{
			MarkdownDescription: "Periods the condition matches.",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"start_day":    dsString("First weekday of the range."),
					"end_day":      dsString("Last weekday of the range, inclusive."),
					"start_hour":   dsInt("Hour the window opens."),
					"start_minute": dsInt("Minute the window opens."),
					"end_hour":     dsInt("Hour the window closes."),
					"end_minute":   dsInt("Minute the window closes."),
				},
			},
		},
	}
}

func (d *timeConditionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_time_condition"
}
func (d *timeConditionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a time condition by id or `name` (`getTimeConditions`). " +
			"Look up by `name`, then link with `route` — do not paste a raw time condition id into a DID.",
		Attributes: timeConditionDataSourceAttributes(true),
	}
}
func (d *timeConditionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}
func (d *timeConditionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data timeConditionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	query := exactlyOneLookup(&resp.Diagnostics, "time condition", []lookupField{
		{Name: "id", Value: configuredString(data.ID)},
		{Name: "name", Value: configuredString(data.Name)},
	})
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := d.client.FindTimeCondition(ctx, query)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VoIP.ms time condition", err.Error())
		return
	}
	flattenTimeCondition(got, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type timeConditionsModel struct {
	ID             types.String         `tfsdk:"id"`
	TimeConditions []timeConditionModel `tfsdk:"time_conditions"`
}

func (d *timeConditionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_time_conditions"
}
func (d *timeConditionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists time conditions (`getTimeConditions`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{MarkdownDescription: "Placeholder identifier (`time_conditions`).", Computed: true},
			"time_conditions": schema.ListNestedAttribute{
				MarkdownDescription: "Time conditions on this account.",
				Computed:            true,
				NestedObject:        schema.NestedAttributeObject{Attributes: timeConditionDataSourceAttributes(false)},
			},
		},
	}
}
func (d *timeConditionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}
func (d *timeConditionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data timeConditionsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	items, err := d.client.GetTimeConditions(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError("Unable to list VoIP.ms time conditions", err.Error())
		return
	}
	data.ID = types.StringValue("time_conditions")
	data.TimeConditions = make([]timeConditionModel, 0, len(items))
	for i := range items {
		data.TimeConditions = append(data.TimeConditions, flattenTimeConditionCopy(&items[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
