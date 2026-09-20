package provider

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
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
	_ resource.ResourceWithModifyPlan  = &didResource{}
)

func NewDIDResource() resource.Resource {
	return &didResource{}
}

type didResource struct {
	client     *client.Client
	routes     client.RouteTables
	routesOnce sync.Once
	routesErr  error
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
	VoicemailName         types.String `tfsdk:"voicemail_name"`
	POP                   types.Int64  `tfsdk:"pop"`
	POPHostname           types.String `tfsdk:"pop_hostname"`
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
		"description":              schema.StringAttribute{MarkdownDescription: "Rate-center / city description from VoIP.ms (read-only).", Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"routing":                  optStr("Inbound route. Set from a resource or data source `route` (`voipms_subaccount.this.route`, `voipms_voicemail.this.route`, `voipms_forwarding.this.route`) or a system action such as `sys:hangup`. Do not paste a raw API id or a display name (`vm:Alex`)."),
		"failover_busy":            optStr("Busy failover route. Same `route` reference as `routing`."),
		"failover_unreachable":     optStr("Unreachable failover route. Same `route` reference as `routing`."),
		"failover_noanswer":        optStr("No-answer failover route. Same `route` reference as `routing`."),
		"voicemail":                optStr("Mailbox attached to the DID (`0` means none). Set from `voipms_voicemail.this.id` or `data.voipms_voicemail.this.id`. Do not paste a raw mailbox number."),
		"voicemail_name":           optStr("Voicemail display name. Prefer `voicemail = voipms_voicemail.this.id` so the mailbox is a resource or data source. Resolved to `voicemail` when applying if set."),
		"pop":                      optIntAttr("Point-of-presence id. Prefer `data.voipms_server` (look up by `hostname` or `name`) rather than a raw POP number."),
		"pop_hostname":             optStr("POP as a SIP hostname (`newyork7.voip.ms`) or display name (`New York 7`). Prefer `data.voipms_server.this.hostname` after a hostname lookup. Resolved to `pop` when applying."),
		"dialtime":                 optIntAttr("Ring time in seconds before failover/voicemail."),
		"cnam":                     optBoolAttr("Enable CNAM lookup on inbound calls."),
		"e911":                     schema.BoolAttribute{MarkdownDescription: "Whether E911 is provisioned (read-only; set in the portal).", Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"callerid_prefix":          optStr("Caller ID prefix."),
		"record_calls":             optBoolAttr("Record inbound calls."),
		"note":                     optStr("Free-form DID note (e.g. `Home line`)."),
		"billing_type":             optStr("DID billing. Use `per_minute` (API `1`) or `flat` (API `2`). Numeric ids still work."),
		"next_billing":             schema.StringAttribute{MarkdownDescription: "Next billing date.", Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"order_date":               schema.StringAttribute{MarkdownDescription: "Date the DID was ordered.", Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"voicemail_threshold":      optIntAttr("Voicemail threshold."),
		"sms_available":            schema.BoolAttribute{MarkdownDescription: "Whether the DID can use SMS.", Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"sms_enabled":              optBoolAttr("Enable SMS/MMS on the DID (`setSMS`)."),
		"mms_available":            schema.BoolAttribute{MarkdownDescription: "Whether the DID can use MMS.", Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"sms_email":                optStr("Email address for inbound SMS."),
		"sms_email_enabled":        optBoolAttr("Deliver inbound SMS to `sms_email`."),
		"sms_forward":              optStr("Phone number to forward SMS to."),
		"sms_forward_enabled":      optBoolAttr("Forward inbound SMS to `sms_forward`."),
		"sms_url_callback":         optStr("Legacy SMS URL callback (supports `{TO}`, `{FROM}`, `{MESSAGE}`)."),
		"sms_url_callback_enabled": optBoolAttr("Enable the legacy URL callback."),
		"sms_url_callback_retry":   optBoolAttr("Retry the legacy URL callback on failure."),
		"webhook":                  optStr("Modern SMS webhook URL."),
		"webhook_enabled":          optBoolAttr("Enable the modern SMS webhook."),
		"dialmode":                 schema.StringAttribute{MarkdownDescription: "SMS dialing mode.", Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
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
	if err := r.applyDID(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Unable to configure VoIP.ms DID", err.Error())
		return
	}
	got, err := r.client.GetDID(ctx, plan.DID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("DID updated but could not be read back", err.Error())
		return
	}
	flattenDID(got, &plan)
	if err := r.keepOrFillPOPHostname(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Unable to resolve DID POP hostname", err.Error())
		return
	}
	if err := r.keepOrFillVoicemailName(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Unable to resolve DID voicemail name", err.Error())
		return
	}
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
	if err := r.keepOrFillPOPHostname(ctx, &state); err != nil {
		resp.Diagnostics.AddError("Unable to resolve DID POP hostname", err.Error())
		return
	}
	if err := r.keepOrFillVoicemailName(ctx, &state); err != nil {
		resp.Diagnostics.AddError("Unable to resolve DID voicemail name", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *didResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan didModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.applyDID(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Unable to update VoIP.ms DID", err.Error())
		return
	}
	got, err := r.client.GetDID(ctx, plan.DID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("DID updated but could not be read back", err.Error())
		return
	}
	flattenDID(got, &plan)
	if err := r.keepOrFillPOPHostname(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Unable to resolve DID POP hostname", err.Error())
		return
	}
	if err := r.keepOrFillVoicemailName(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Unable to resolve DID voicemail name", err.Error())
		return
	}
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

func (r *didResource) applyDID(ctx context.Context, plan *didModel) error {
	if err := r.resolvePOP(ctx, plan); err != nil {
		return err
	}
	if err := r.resolveVoicemail(ctx, plan); err != nil {
		return err
	}
	if err := r.resolveRoutes(ctx, plan); err != nil {
		return err
	}
	current, err := r.client.GetDID(ctx, plan.DID.ValueString())
	if err != nil {
		return err
	}
	info := overlayParams(current.SetInfoParams(), didInfoParams(*plan))
	info["did"] = plan.DID.ValueString()
	if err := r.client.SetDIDInfo(ctx, info); err != nil {
		return err
	}
	if !current.SMSAvailable.Bool() && (current.SMSAvailable.String() == "0" || current.SMSAvailable.String() == "") {
		return nil
	}
	sms := overlayParams(current.SetSMSParams(), didSMSParams(*plan))
	sms["did"] = plan.DID.ValueString()
	return r.client.SetSMS(ctx, sms)
}

func (r *didResource) resolvePOP(ctx context.Context, plan *didModel) error {
	if plan.POPHostname.IsNull() || plan.POPHostname.IsUnknown() || plan.POPHostname.ValueString() == "" {
		return nil
	}
	srv, err := r.client.FindServer(ctx, plan.POPHostname.ValueString())
	if err != nil {
		return err
	}
	n, ok := srv.POP.Int64()
	if !ok {
		return fmt.Errorf("POP %q has no numeric id", plan.POPHostname.ValueString())
	}
	if !plan.POP.IsNull() && !plan.POP.IsUnknown() && plan.POP.ValueInt64() != n {
		return fmt.Errorf("pop (%d) does not match pop_hostname %q (id %d)", plan.POP.ValueInt64(), plan.POPHostname.ValueString(), n)
	}
	plan.POP = types.Int64Value(n)
	return nil
}

func (r *didResource) keepOrFillPOPHostname(ctx context.Context, m *didModel) error {
	return fillDIDPOPHostname(ctx, r.client, m)
}

func (r *didResource) resolveVoicemail(ctx context.Context, plan *didModel) error {
	if plan.VoicemailName.IsNull() || plan.VoicemailName.IsUnknown() || plan.VoicemailName.ValueString() == "" {
		return nil
	}
	box, err := r.client.FindVoicemail(ctx, plan.VoicemailName.ValueString())
	if err != nil {
		return err
	}
	mailbox := box.Mailbox.String()
	if !plan.Voicemail.IsNull() && !plan.Voicemail.IsUnknown() && plan.Voicemail.ValueString() != "" && plan.Voicemail.ValueString() != mailbox {
		return fmt.Errorf("voicemail %q does not match voicemail_name %q (mailbox %s)", plan.Voicemail.ValueString(), plan.VoicemailName.ValueString(), mailbox)
	}
	plan.Voicemail = types.StringValue(mailbox)
	return nil
}

func (r *didResource) keepOrFillVoicemailName(ctx context.Context, m *didModel) error {
	return fillDIDVoicemailName(ctx, r.client, m)
}

func (r *didResource) routeTables(ctx context.Context) (client.RouteTables, error) {
	if r.client == nil {
		return client.RouteTables{}, fmt.Errorf("client not configured")
	}
	r.routesOnce.Do(func() {
		fwds, err := r.client.GetForwardings(ctx, "")
		if err != nil {
			r.routesErr = err
			return
		}
		vms, err := r.client.GetVoicemails(ctx, "")
		if err != nil {
			r.routesErr = err
			return
		}
		r.routes = client.RouteTables{Forwardings: fwds, Voicemails: vms}
	})
	return r.routes, r.routesErr
}

func (r *didResource) resolveRoutes(ctx context.Context, plan *didModel) error {
	tables, err := r.routeTables(ctx)
	if err != nil {
		return err
	}
	return resolveDIDRoutes(plan, tables)
}

func resolveDIDRoutes(plan *didModel, tables client.RouteTables) error {
	fields := []*types.String{&plan.Routing, &plan.FailoverBusy, &plan.FailoverUnreachable, &plan.FailoverNoanswer}
	for _, f := range fields {
		if f.IsNull() || f.IsUnknown() || f.ValueString() == "" {
			continue
		}
		canon, err := client.CanonicalRoute(f.ValueString(), tables)
		if err != nil {
			return err
		}
		*f = types.StringValue(canon)
	}
	return nil
}

func (r *didResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() || r.client == nil {
		return
	}
	var plan, state didModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tables, err := r.routeTables(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to resolve DID route names", err.Error())
		return
	}
	keepEquivalentRoute(&plan.Routing, state.Routing, tables)
	keepEquivalentRoute(&plan.FailoverBusy, state.FailoverBusy, tables)
	keepEquivalentRoute(&plan.FailoverUnreachable, state.FailoverUnreachable, tables)
	keepEquivalentRoute(&plan.FailoverNoanswer, state.FailoverNoanswer, tables)
	keepNamed(&plan.BillingType, state.BillingType, client.BillingType)
	// Keep computed-only attributes from state so Set() does not mark them unknown.
	plan.Description = state.Description
	plan.E911 = state.E911
	plan.NextBilling = state.NextBilling
	plan.OrderDate = state.OrderDate
	plan.SMSAvailable = state.SMSAvailable
	plan.MMSAvailable = state.MMSAvailable
	plan.Dialmode = state.Dialmode
	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

func keepEquivalentRoute(plan *types.String, state types.String, tables client.RouteTables) {
	if plan.IsNull() || plan.IsUnknown() || state.IsNull() || state.IsUnknown() {
		return
	}
	if client.RoutesEqual(plan.ValueString(), state.ValueString(), tables) {
		*plan = state
	}
}

func fillDIDVoicemailName(ctx context.Context, c *client.Client, m *didModel) error {
	if c == nil || m.Voicemail.IsNull() || m.Voicemail.IsUnknown() || m.Voicemail.ValueString() == "" || m.Voicemail.ValueString() == "0" {
		return nil
	}
	items, err := c.GetVoicemails(ctx, "")
	if err != nil {
		return err
	}
	if !m.VoicemailName.IsNull() && !m.VoicemailName.IsUnknown() && m.VoicemailName.ValueString() != "" {
		if box, err := client.MatchVoicemail(items, m.VoicemailName.ValueString()); err == nil && box.Mailbox.String() == m.Voicemail.ValueString() {
			return nil
		}
	}
	if name := client.NameForMailbox(items, m.Voicemail.ValueString()); name != "" {
		m.VoicemailName = types.StringValue(name)
	}
	return nil
}

func fillDIDVoicemailNames(ctx context.Context, c *client.Client, dids []didModel) error {
	if c == nil || len(dids) == 0 {
		return nil
	}
	items, err := c.GetVoicemails(ctx, "")
	if err != nil {
		return err
	}
	for i := range dids {
		if name := client.NameForMailbox(items, dids[i].Voicemail.ValueString()); name != "" {
			dids[i].VoicemailName = types.StringValue(name)
		}
	}
	return nil
}

func fillDIDPOPHostname(ctx context.Context, c *client.Client, m *didModel) error {
	if c == nil || m.POP.IsNull() || m.POP.IsUnknown() {
		return nil
	}
	servers, err := c.GetServersInfo(ctx, "")
	if err != nil {
		return err
	}
	if !m.POPHostname.IsNull() && !m.POPHostname.IsUnknown() && m.POPHostname.ValueString() != "" {
		if srv, err := client.MatchServer(servers, m.POPHostname.ValueString()); err == nil {
			if n, ok := srv.POP.Int64(); ok && n == m.POP.ValueInt64() {
				return nil
			}
		}
	}
	if h := client.HostnameForPOP(servers, m.POP.ValueInt64()); h != "" {
		m.POPHostname = types.StringValue(h)
	}
	return nil
}

func fillDIDPOPHostnames(ctx context.Context, c *client.Client, dids []didModel) error {
	if c == nil || len(dids) == 0 {
		return nil
	}
	servers, err := c.GetServersInfo(ctx, "")
	if err != nil {
		return err
	}
	for i := range dids {
		if h := client.HostnameForPOP(servers, dids[i].POP.ValueInt64()); h != "" {
			dids[i].POPHostname = types.StringValue(h)
		}
	}
	return nil
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
	setNamed(params, "billing_type", m.BillingType, client.BillingType)
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
	dst.BillingType = namedVal(src.BillingType, client.BillingType)
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
