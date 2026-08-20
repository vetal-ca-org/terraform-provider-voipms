package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vetal-ca-org/terraform-provider-voipms/internal/client"
)

func NewDIDDataSource() datasource.DataSource  { return &didDataSource{} }
func NewDIDsDataSource() datasource.DataSource { return &didsDataSource{} }

type didDataSource struct{ client *client.Client }
type didsDataSource struct{ client *client.Client }

func didDataSourceAttributes(lookup bool) map[string]schema.Attribute {
	did := schema.StringAttribute{MarkdownDescription: "Phone number.", Computed: true}
	if lookup {
		did.Required = true
		did.Computed = false
	}
	return map[string]schema.Attribute{
		"id":                       schema.StringAttribute{MarkdownDescription: "Same as `did`.", Computed: true},
		"did":                      did,
		"description":              dsString("Rate-center / city description."),
		"routing":                  dsString("Inbound route."),
		"failover_busy":            dsString("Busy failover route."),
		"failover_unreachable":     dsString("Unreachable failover route."),
		"failover_noanswer":        dsString("No-answer failover route."),
		"voicemail":                dsString("Attached mailbox."),
		"pop":                      dsInt("Point-of-presence id."),
		"dialtime":                 dsInt("Ring time in seconds."),
		"cnam":                     dsBool("CNAM lookup enabled."),
		"e911":                     dsBool("E911 provisioned."),
		"callerid_prefix":          dsString("Caller ID prefix."),
		"record_calls":             dsBool("Inbound call recording."),
		"note":                     dsString("DID note."),
		"billing_type":             dsString("Billing type."),
		"next_billing":             dsString("Next billing date."),
		"order_date":               dsString("Order date."),
		"voicemail_threshold":      dsInt("Voicemail threshold."),
		"sms_available":            dsBool("SMS capable."),
		"sms_enabled":              dsBool("SMS enabled."),
		"mms_available":            dsBool("MMS capable."),
		"sms_email":                dsString("SMS email address."),
		"sms_email_enabled":        dsBool("SMS email delivery enabled."),
		"sms_forward":              dsString("SMS forward destination."),
		"sms_forward_enabled":      dsBool("SMS forwarding enabled."),
		"sms_url_callback":         dsString("Legacy SMS URL callback."),
		"sms_url_callback_enabled": dsBool("Legacy URL callback enabled."),
		"sms_url_callback_retry":   dsBool("Legacy URL callback retry."),
		"webhook":                  dsString("Modern SMS webhook URL."),
		"webhook_enabled":          dsBool("Modern webhook enabled."),
		"dialmode":                 dsString("SMS dialing mode."),
		"sms_sipaccount":           dsString("SIP account used to send SMS."),
		"sms_sipaccount_enabled":   dsBool("SIP-account SMS sending enabled."),
	}
}

func (d *didDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_did"
}

func (d *didDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a single DID (`getDIDsInfo`).",
		Attributes:          didDataSourceAttributes(true),
	}
}

func (d *didDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

func (d *didDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data didModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := d.client.GetDID(ctx, data.DID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VoIP.ms DID", err.Error())
		return
	}
	flattenDID(got, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type didsModel struct {
	ID   types.String `tfsdk:"id"`
	DIDs []didModel   `tfsdk:"dids"`
}

func (d *didsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dids"
}

func (d *didsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all DIDs on the account (`getDIDsInfo`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{MarkdownDescription: "Placeholder identifier (`dids`).", Computed: true},
			"dids": schema.ListNestedAttribute{
				MarkdownDescription: "DIDs on this account.",
				Computed:            true,
				NestedObject:        schema.NestedAttributeObject{Attributes: didDataSourceAttributes(false)},
			},
		},
	}
}

func (d *didsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

func (d *didsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data didsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	items, err := d.client.GetDIDsInfo(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError("Unable to list VoIP.ms DIDs", err.Error())
		return
	}
	data.ID = types.StringValue("dids")
	data.DIDs = make([]didModel, 0, len(items))
	for i := range items {
		data.DIDs = append(data.DIDs, flattenDIDCopy(&items[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
