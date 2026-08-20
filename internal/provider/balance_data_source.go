package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/vetal-ca-org/terraform-provider-voipms/internal/client"
)

var _ datasource.DataSource = &balanceDataSource{}

func NewBalanceDataSource() datasource.DataSource {
	return &balanceDataSource{}
}

type balanceDataSource struct {
	client *client.Client
}

type balanceDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	CurrentBalance types.String `tfsdk:"current_balance"`
	SpentTotal     types.String `tfsdk:"spent_total"`
	CallsTotal     types.String `tfsdk:"calls_total"`
	TimeTotal      types.String `tfsdk:"time_total"`
	SpentToday     types.String `tfsdk:"spent_today"`
	CallsToday     types.String `tfsdk:"calls_today"`
	TimeToday      types.String `tfsdk:"time_today"`
}

func (d *balanceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_balance"
}

func (d *balanceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Account balance from the VoIP.ms `getBalance` API (advanced breakdown).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier (always `balance`).",
				Computed:            true,
			},
			"current_balance": schema.StringAttribute{
				MarkdownDescription: "Current account balance.",
				Computed:            true,
			},
			"spent_total": schema.StringAttribute{
				MarkdownDescription: "Total amount spent.",
				Computed:            true,
			},
			"calls_total": schema.StringAttribute{
				MarkdownDescription: "Total call charges.",
				Computed:            true,
			},
			"time_total": schema.StringAttribute{
				MarkdownDescription: "Total call time.",
				Computed:            true,
			},
			"spent_today": schema.StringAttribute{
				MarkdownDescription: "Amount spent today.",
				Computed:            true,
			},
			"calls_today": schema.StringAttribute{
				MarkdownDescription: "Call charges today.",
				Computed:            true,
			},
			"time_today": schema.StringAttribute{
				MarkdownDescription: "Call time today.",
				Computed:            true,
			},
		},
	}
}

func (d *balanceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected data source configure type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = c
}

func (d *balanceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data balanceDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider client was nil while reading voipms_balance.")
		return
	}

	balance, err := d.client.GetBalance(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VoIP.ms balance", err.Error())
		return
	}

	data.ID = types.StringValue("balance")
	data.CurrentBalance = types.StringValue(balance.CurrentBalance)
	data.SpentTotal = types.StringValue(balance.SpentTotal)
	data.CallsTotal = types.StringValue(balance.CallsTotal)
	data.TimeTotal = types.StringValue(balance.TimeTotal)
	data.SpentToday = types.StringValue(balance.SpentToday)
	data.CallsToday = types.StringValue(balance.CallsToday)
	data.TimeToday = types.StringValue(balance.TimeToday)

	tflog.Trace(ctx, "read voipms_balance")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
