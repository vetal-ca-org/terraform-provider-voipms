package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vetal-ca-org/terraform-provider-voipms/internal/client"
)

var _ datasource.DataSource = &subaccountDataSource{}
var _ datasource.DataSource = &subaccountsDataSource{}

func NewSubaccountDataSource() datasource.DataSource {
	return &subaccountDataSource{}
}

func NewSubaccountsDataSource() datasource.DataSource {
	return &subaccountsDataSource{}
}

type subaccountDataSource struct {
	client *client.Client
}

type subaccountsDataSource struct {
	client *client.Client
}

func subaccountDataSourceAttributes(lookup bool) map[string]schema.Attribute {
	id := schema.StringAttribute{
		MarkdownDescription: "Numeric VoIP.ms sub-account id, full SIP login, or username suffix.",
		Computed:            true,
	}
	account := schema.StringAttribute{
		MarkdownDescription: "Full SIP login (`{main}_{username}`).",
		Computed:            true,
	}
	if lookup {
		id.Optional = true
		account.Optional = true
	}
	return map[string]schema.Attribute{
		"id":                     id,
		"account":                account,
		"username":               dsString("Sub-account username suffix."),
		"description":            dsString("Portal description."),
		"protocol":               dsString("Protocol id."),
		"auth_type":              dsString("Authentication type."),
		"password":               dsSensitiveString("SIP password."),
		"ip":                     dsString("IP/FQDN used for IP authentication."),
		"device_type":            dsString("Device type id."),
		"callerid_number":        dsString("Outbound caller ID number."),
		"canada_routing":         dsString("Canada routing id."),
		"lock_international":     dsString("International lock setting."),
		"international_route":    dsString("International route id."),
		"music_on_hold":          dsString("Music on hold class."),
		"language":               dsString("Language code."),
		"allowed_codecs":         dsString("Allowed codecs."),
		"dtmf_mode":              dsString("DTMF mode."),
		"nat":                    dsString("NAT setting."),
		"sip_traffic":            dsBool("Encrypted SIP traffic enabled."),
		"max_expiry":             dsInt("Maximum registration expiry in seconds."),
		"rtp_timeout":            dsInt("RTP timeout in seconds."),
		"rtp_hold_timeout":       dsInt("RTP hold timeout in seconds."),
		"ip_restriction":         dsString("IP restriction list."),
		"enable_ip_restriction":  dsBool("Whether IP restriction is enabled."),
		"pop_restriction":        dsString("POP restriction list."),
		"enable_pop_restriction": dsBool("Whether POP restriction is enabled."),
		"record_calls":           dsBool("Whether calls are recorded."),
		"allow225":               dsBool("Whether `*225` is allowed."),
		"internal_extension":     dsString("Internal extension."),
		"internal_voicemail":     dsString("Internal voicemail mailbox."),
		"internal_dialtime":      dsString("Internal ring time."),
		"enable_internal_cnam":   dsBool("Whether internal CNAM is enabled."),
		"dialing_mode":           dsString("Dialing mode."),
		"default_e911":           dsString("Default E911 DID."),
		"call_pickup_behavior":   dsString("Call pickup behavior."),
	}
}

func dsString(desc string) schema.StringAttribute {
	return schema.StringAttribute{MarkdownDescription: desc, Computed: true}
}

func dsSensitiveString(desc string) schema.StringAttribute {
	return schema.StringAttribute{MarkdownDescription: desc, Computed: true, Sensitive: true}
}

func dsBool(desc string) schema.BoolAttribute {
	return schema.BoolAttribute{MarkdownDescription: desc, Computed: true}
}

func dsInt(desc string) schema.Int64Attribute {
	return schema.Int64Attribute{MarkdownDescription: desc, Computed: true}
}

func (d *subaccountDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subaccount"
}

func (d *subaccountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a single VoIP.ms sub-account by `id` or `account` (`getSubAccounts`).",
		Attributes:          subaccountDataSourceAttributes(true),
	}
}

func (d *subaccountDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

func (d *subaccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data subaccountModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider client was nil while reading voipms_subaccount.")
		return
	}
	lookup := data.ID.ValueString()
	if lookup == "" {
		lookup = data.Account.ValueString()
	}
	if lookup == "" {
		resp.Diagnostics.AddError("Missing sub-account identifier", "Set id or account to look up a sub-account.")
		return
	}
	acct, err := d.client.GetSubAccount(ctx, lookup)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VoIP.ms sub-account", err.Error())
		return
	}
	flattenSubaccount(acct, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type subaccountsModel struct {
	ID          types.String      `tfsdk:"id"`
	Subaccounts []subaccountModel `tfsdk:"subaccounts"`
}

func (d *subaccountsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subaccounts"
}

func (d *subaccountsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all VoIP.ms sub-accounts (`getSubAccounts`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier (`subaccounts`).",
				Computed:            true,
			},
			"subaccounts": schema.ListNestedAttribute{
				MarkdownDescription: "Sub-accounts on this VoIP.ms account.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: subaccountDataSourceAttributes(false),
				},
			},
		},
	}
}

func (d *subaccountsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

func (d *subaccountsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data subaccountsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider client was nil while reading voipms_subaccounts.")
		return
	}
	items, err := d.client.GetSubAccounts(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError("Unable to list VoIP.ms sub-accounts", err.Error())
		return
	}
	data.ID = types.StringValue("subaccounts")
	data.Subaccounts = make([]subaccountModel, 0, len(items))
	for i := range items {
		data.Subaccounts = append(data.Subaccounts, flattenSubaccountCopy(&items[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
