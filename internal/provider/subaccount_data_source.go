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
	username := schema.StringAttribute{
		MarkdownDescription: "Sub-account username suffix.",
		Computed:            true,
	}
	if lookup {
		id.Optional = true
		account.Optional = true
		username.Optional = true
	}
	return map[string]schema.Attribute{
		"id":                     id,
		"route":                  dsString("DID routing value (`account:{account}`). Use this for `voipms_did` `routing` / failover."),
		"account":                account,
		"username":               username,
		"description":            dsString("Portal description."),
		"protocol":               dsString("Protocol (`sip` or `iax2`)."),
		"auth_type":              dsString("Authentication type (`password` or `ip`)."),
		"password":               dsSensitiveString("SIP password."),
		"ip":                     dsString("IP/FQDN used for IP authentication."),
		"device_type":            dsString("Device type (`ip_pbx` or `ata`)."),
		"callerid_number":        dsString("Outbound caller ID number."),
		"canada_routing":         dsString("Canada routing (`value` or `premium`)."),
		"lock_international":     dsString("International calling (`allow` or `deny`)."),
		"international_route":    dsString("International route (`value` or `premium`)."),
		"music_on_hold":          dsString("Music on hold class."),
		"language":               dsString("Language code."),
		"allowed_codecs":         dsString("Allowed codecs."),
		"dtmf_mode":              dsString("DTMF mode."),
		"nat":                    dsString("NAT setting."),
		"encrypted_sip_traffic":  dsBool(encryptedSIPTrafficDescription),
		"max_expiry":             dsInt("Maximum registration expiry in seconds."),
		"rtp_timeout":            dsInt("RTP timeout in seconds."),
		"rtp_hold_timeout":       dsInt("RTP hold timeout in seconds."),
		"ip_restriction":         dsString("IP restriction list."),
		"enable_ip_restriction":  dsBool("Whether IP restriction is enabled."),
		"pop_restriction":        dsString("POP restriction list (unset when restriction is off)."),
		"enable_pop_restriction": dsBool("Whether POP restriction is enabled."),
		"record_calls":           dsBool("Whether calls are recorded."),
		"allow_225_balance":      dsBool("Whether `*225` / `*BAL` balance check is allowed."),
		"internal_extension":     dsString("Internal extension."),
		"internal_voicemail":     dsString("Internal voicemail mailbox. Set from a `voipms_voicemail` `id`."),
		"internal_dialtime":      dsString("Internal ring time."),
		"enable_internal_cnam":   dsBool("Whether internal CNAM is enabled."),
		"dialing_mode":           dsString("Outbound dialing mode (`main_account`, `e164`, or `nanpa`)."),
		"default_e911":           dsString("Default E911 DID."),
		"call_pickup_behavior":   dsString("Call pickup permissions (`pickup_and_be_picked_up`, `pickup_only`, `be_picked_up_only`, or `disabled`)."),
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
		MarkdownDescription: "Reads a single VoIP.ms sub-account by `id`, `account`, or `username` (`getSubAccounts`). " +
			"Look up by `username`, then link with `route` — do not paste a raw SIP login into a DID.",
		Attributes: subaccountDataSourceAttributes(true),
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
	lookup := exactlyOneLookup(&resp.Diagnostics, "sub-account", []lookupField{
		{Name: "id", Value: configuredString(data.ID)},
		{Name: "account", Value: configuredString(data.Account)},
		{Name: "username", Value: configuredString(data.Username)},
	})
	if resp.Diagnostics.HasError() {
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
