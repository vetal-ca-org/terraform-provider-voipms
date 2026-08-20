package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/vetal-ca-org/terraform-provider-voipms/internal/client"
)

var (
	_ resource.Resource                = &didResource{}
	_ resource.ResourceWithConfigure   = &didResource{}
	_ resource.ResourceWithImportState = &didResource{}
)

func NewDIDResource() resource.Resource {
	return &didResource{}
}

type didResource struct {
	client *client.Client
}

type didModel struct {
	ID                    types.String `tfsdk:"id"`
	DID                   types.String `tfsdk:"did"`
	Description           types.String `tfsdk:"description"`
	Routing               types.String `tfsdk:"routing"`
	FailoverBusy          types.String `tfsdk:"failover_busy"`
	FailoverUnreachable   types.String `tfsdk:"failover_unreachable"`
	FailoverNoanswer      types.String `tfsdk:"failover_noanswer"`
	Voicemail             types.String `tfsdk:"voicemail"`
	POP                   types.Int64  `tfsdk:"pop"`
	Dialtime              types.Int64  `tfsdk:"dialtime"`
	CNAM                  types.Bool   `tfsdk:"cnam"`
	E911                  types.Bool   `tfsdk:"e911"`
	CallerIDPrefix        types.String `tfsdk:"callerid_prefix"`
	RecordCalls           types.Bool   `tfsdk:"record_calls"`
	Note                  types.String `tfsdk:"note"`
	BillingType           types.String `tfsdk:"billing_type"`
	NextBilling           types.String `tfsdk:"next_billing"`
	OrderDate             types.String `tfsdk:"order_date"`
	VoicemailThreshold    types.Int64  `tfsdk:"voicemail_threshold"`
	SMSAvailable          types.Bool   `tfsdk:"sms_available"`
	SMSEnabled            types.Bool   `tfsdk:"sms_enabled"`
	MMSAvailable          types.Bool   `tfsdk:"mms_available"`
	SMSEmail              types.String `tfsdk:"sms_email"`
	SMSEmailEnabled       types.Bool   `tfsdk:"sms_email_enabled"`
	SMSForward            types.String `tfsdk:"sms_forward"`
	SMSForwardEnabled     types.Bool   `tfsdk:"sms_forward_enabled"`
	SMSURLCallback        types.String `tfsdk:"sms_url_callback"`
	SMSURLCallbackEnabled types.Bool   `tfsdk:"sms_url_callback_enabled"`
	SMSURLCallbackRetry   types.Bool   `tfsdk:"sms_url_callback_retry"`
	Webhook               types.String `tfsdk:"webhook"`
	WebhookEnabled        types.Bool   `tfsdk:"webhook_enabled"`
	Dialmode              types.String `tfsdk:"dialmode"`
	SMSSIPAccount         types.String `tfsdk:"sms_sipaccount"`
	SMSSIPAccountEnabled  types.Bool   `tfsdk:"sms_sipaccount_enabled"`
}

func (r *didResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_did"
}

func (r *didResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages routing, failover, voicemail, POP, and SMS settings for a DID that is already on the account (`setDIDInfo` / `setSMS`). " +
			"Terraform will not order or cancel the phone number. Destroy only removes the resource from state.",
		Attributes: didResourceAttributes(),
	}
}

func didResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Same as `did`.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"did": schema.StringAttribute{
			MarkdownDescription: "Phone number already on the account. Changing this forces a new resource.",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"description":              schema.StringAttribute{MarkdownDescription: "Rate-center / city description from VoIP.ms (read-only).", Computed: true},
		"routing":                  optStr("Inbound route, e.g. `account:100001_gateway`, `fwd:1001`, `vm:101`."),
		"failover_busy":            optStr("Busy failover route."),
		"failover_unreachable":     optStr("Unreachable failover route."),
		"failover_noanswer":        optStr("No-answer failover route."),
		"voicemail":                optStr("Mailbox attached to the DID (`0` means none)."),
		"pop":                      optIntAttr("Point-of-presence id (see `voipms_servers`)."),
		"dialtime":                 optIntAttr("Ring time in seconds before failover/voicemail."),
		"cnam":                     optBoolAttr("Enable CNAM lookup on inbound calls."),
		"e911":                     schema.BoolAttribute{MarkdownDescription: "Whether E911 is provisioned (read-only; set in the portal).", Computed: true},
		"callerid_prefix":          optStr("Caller ID prefix."),
		"record_calls":             optBoolAttr("Record inbound calls."),
		"note":                     optStr("Free-form DID note (e.g. `Home line`)."),
		"billing_type":             optStr("`1` = per minute, `2` = flat rate."),
		"next_billing":             schema.StringAttribute{MarkdownDescription: "Next billing date.", Computed: true},
		"order_date":               schema.StringAttribute{MarkdownDescription: "Date the DID was ordered.", Computed: true},
		"voicemail_threshold":      optIntAttr("Voicemail threshold."),
		"sms_available":            schema.BoolAttribute{MarkdownDescription: "Whether the DID can use SMS.", Computed: true},
		"sms_enabled":              optBoolAttr("Enable SMS/MMS on the DID (`setSMS`)."),
		"mms_available":            schema.BoolAttribute{MarkdownDescription: "Whether the DID can use MMS.", Computed: true},
		"sms_email":                optStr("Email address for inbound SMS."),
		"sms_email_enabled":        optBoolAttr("Deliver inbound SMS to `sms_email`."),
		"sms_forward":              optStr("Phone number to forward SMS to."),
		"sms_forward_enabled":      optBoolAttr("Forward inbound SMS to `sms_forward`."),
		"sms_url_callback":         optStr("Legacy SMS URL callback (supports `{TO}`, `{FROM}`, `{MESSAGE}`)."),
		"sms_url_callback_enabled": optBoolAttr("Enable the legacy URL callback."),
		"sms_url_callback_retry":   optBoolAttr("Retry the legacy URL callback on failure."),
		"webhook":                  optStr("Modern SMS webhook URL."),
		"webhook_enabled":          optBoolAttr("Enable the modern SMS webhook."),
		"dialmode":                 schema.StringAttribute{MarkdownDescription: "SMS dialing mode.", Computed: true},
		"sms_sipaccount":           optStr("Sub-account used to send SMS via SIP."),
		"sms_sipaccount_enabled":   optBoolAttr("Enable SIP-account SMS sending."),
	}
}

func optStr(desc string) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: desc,
		Optional:            true,
		Computed:            true,
		PlanModifiers:       []planmodifier.String{optString()},
	}
}

func optBoolAttr(desc string) schema.BoolAttribute {
	return schema.BoolAttribute{
		MarkdownDescription: desc,
		Optional:            true,
		Computed:            true,
		PlanModifiers:       []planmodifier.Bool{optBool()},
	}
}

func optIntAttr(desc string) schema.Int64Attribute {
	return schema.Int64Attribute{
		MarkdownDescription: desc,
		Optional:            true,
		Computed:            true,
		PlanModifiers:       []planmodifier.Int64{optInt()},
	}
}

func (r *didResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

func (r *didResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan didModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.applyDID(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Unable to configure VoIP.ms DID", err.Error())
		return
	}
	got, err := r.client.GetDID(ctx, plan.DID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("DID updated but could not be read back", err.Error())
		return
	}
	flattenDID(got, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *didResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state didModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetDID(ctx, state.DID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read VoIP.ms DID", err.Error())
		return
	}
	flattenDID(got, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *didResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan didModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.applyDID(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Unable to update VoIP.ms DID", err.Error())
		return
	}
	got, err := r.client.GetDID(ctx, plan.DID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("DID updated but could not be read back", err.Error())
		return
	}
	flattenDID(got, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *didResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Warn(ctx, "removing voipms_did from state without cancelling the number at VoIP.ms")
	resp.State.RemoveResource(ctx)
}

func (r *didResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("did"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *didResource) applyDID(ctx context.Context, plan didModel) error {
	current, err := r.client.GetDID(ctx, plan.DID.ValueString())
	if err != nil {
		return err
	}
	info := overlayParams(current.SetInfoParams(), didInfoParams(plan))
	info["did"] = plan.DID.ValueString()
	if err := r.client.SetDIDInfo(ctx, info); err != nil {
		return err
	}
	if !current.SMSAvailable.Bool() && (current.SMSAvailable.String() == "0" || current.SMSAvailable.String() == "") {
		return nil
	}
	sms := overlayParams(current.SetSMSParams(), didSMSParams(plan))
	sms["did"] = plan.DID.ValueString()
	return r.client.SetSMS(ctx, sms)
}

func didInfoParams(m didModel) map[string]string {
	params := map[string]string{}
	setString(params, "routing", m.Routing)
	setString(params, "failover_busy", m.FailoverBusy)
	setString(params, "failover_unreachable", m.FailoverUnreachable)
	setString(params, "failover_noanswer", m.FailoverNoanswer)
	setString(params, "voicemail", m.Voicemail)
	setInt(params, "pop", m.POP)
	setInt(params, "dialtime", m.Dialtime)
	setBool01(params, "cnam", m.CNAM)
	setString(params, "callerid_prefix", m.CallerIDPrefix)
	setString(params, "note", m.Note)
	setString(params, "billing_type", m.BillingType)
	setBool01(params, "record_calls", m.RecordCalls)
	setInt(params, "voicemail_threshold", m.VoicemailThreshold)
	return params
}

func didSMSParams(m didModel) map[string]string {
	params := map[string]string{}
	setBool01(params, "enable", m.SMSEnabled)
	setBool01(params, "email_enabled", m.SMSEmailEnabled)
	setString(params, "email_address", m.SMSEmail)
	setBool01(params, "sms_forward_enable", m.SMSForwardEnabled)
	setString(params, "sms_forward", m.SMSForward)
	setBool01(params, "url_callback_enable", m.SMSURLCallbackEnabled)
	setString(params, "url_callback", m.SMSURLCallback)
	setBool01(params, "url_callback_retry", m.SMSURLCallbackRetry)
	setString(params, "sms_sipaccount", m.SMSSIPAccount)
	setBool01(params, "sms_sipaccount_enabled", m.SMSSIPAccountEnabled)
	setString(params, "webhook", m.Webhook)
	setBool01(params, "webhook_enabled", m.WebhookEnabled)
	return params
}

func flattenDID(src *client.DID, dst *didModel) {
	dst.ID = strVal(src.DID)
	dst.DID = strVal(src.DID)
	dst.Description = strVal(src.Description)
	dst.Routing = strVal(src.Routing)
	dst.FailoverBusy = strVal(src.FailoverBusy)
	dst.FailoverUnreachable = strVal(src.FailoverUnreachable)
	dst.FailoverNoanswer = strVal(src.FailoverNoanswer)
	dst.Voicemail = strVal(src.Voicemail)
	dst.POP = intVal(src.POP)
	dst.Dialtime = intVal(src.Dialtime)
	dst.CNAM = boolVal(src.CNAM)
	dst.E911 = boolVal(src.E911)
	dst.CallerIDPrefix = strVal(src.CallerIDPrefix)
	dst.RecordCalls = boolVal(src.RecordCalls)
	dst.Note = strVal(src.Note)
	dst.BillingType = strVal(src.BillingType)
	dst.NextBilling = strVal(src.NextBilling)
	dst.OrderDate = strVal(src.OrderDate)
	dst.VoicemailThreshold = intVal(src.VoicemailThreshold)
	dst.SMSAvailable = boolVal(src.SMSAvailable)
	dst.SMSEnabled = boolVal(src.SMSEnabled)
	dst.MMSAvailable = boolVal(src.MMSAvailable)
	dst.SMSEmail = strVal(src.SMSEmail)
	dst.SMSEmailEnabled = boolVal(src.SMSEmailEnabled)
	dst.SMSForward = strVal(src.SMSForward)
	dst.SMSForwardEnabled = boolVal(src.SMSForwardEnabled)
	dst.SMSURLCallback = strVal(src.SMSURLCallback)
	dst.SMSURLCallbackEnabled = boolVal(src.SMSURLCallbackEnabled)
	dst.SMSURLCallbackRetry = boolVal(src.SMSURLCallbackRetry)
	dst.Webhook = strVal(src.Webhook)
	dst.WebhookEnabled = boolVal(src.WebhookEnabled)
	dst.Dialmode = strVal(src.Dialmode)
	dst.SMSSIPAccount = strVal(src.SMSSIPAccount)
	dst.SMSSIPAccountEnabled = boolVal(src.SMSSIPAccountEnabled)
}

func flattenDIDCopy(src *client.DID) didModel {
	var m didModel
	flattenDID(src, &m)
	return m
}
