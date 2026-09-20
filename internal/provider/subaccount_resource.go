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

	"github.com/vetal-ca-org/terraform-provider-voipms/internal/client"
)

// encryptedSIPTrafficDescription is VoIP.ms Encrypted SIP Traffic (API sip_traffic).
const encryptedSIPTrafficDescription = "VoIP.ms **Encrypted SIP Traffic** for this sub-account.\n\n" +
	"`false`: normal, unencrypted SIP signaling/media — typically SIP over UDP or TCP and plain RTP. This is the usual setting for standard Asterisk/FreeSWITCH/ATA configurations.\n\n" +
	"`true`: VoIP.ms requires encrypted calling for the sub-account: SIP over TLS for signaling and SRTP for audio. Devices that still send UDP/TCP SIP or ordinary RTP can be rejected, commonly with SIP error 488."

var (
	_ resource.Resource                 = &subaccountResource{}
	_ resource.ResourceWithConfigure    = &subaccountResource{}
	_ resource.ResourceWithImportState  = &subaccountResource{}
	_ resource.ResourceWithModifyPlan   = &subaccountResource{}
	_ resource.ResourceWithUpgradeState = &subaccountResource{}
)

func NewSubaccountResource() resource.Resource {
	return &subaccountResource{}
}

type subaccountResource struct {
	client *client.Client
}

type subaccountModel struct {
	ID                   types.String `tfsdk:"id"`
	Route                types.String `tfsdk:"route"`
	Account              types.String `tfsdk:"account"`
	Username             types.String `tfsdk:"username"`
	Description          types.String `tfsdk:"description"`
	Protocol             types.String `tfsdk:"protocol"`
	AuthType             types.String `tfsdk:"auth_type"`
	Password             types.String `tfsdk:"password"`
	IP                   types.String `tfsdk:"ip"`
	DeviceType           types.String `tfsdk:"device_type"`
	CallerIDNumber       types.String `tfsdk:"callerid_number"`
	CanadaRouting        types.String `tfsdk:"canada_routing"`
	LockInternational    types.String `tfsdk:"lock_international"`
	InternationalRoute   types.String `tfsdk:"international_route"`
	MusicOnHold          types.String `tfsdk:"music_on_hold"`
	Language             types.String `tfsdk:"language"`
	AllowedCodecs        types.String `tfsdk:"allowed_codecs"`
	DTMFMode             types.String `tfsdk:"dtmf_mode"`
	NAT                  types.String `tfsdk:"nat"`
	SIPTraffic           types.Bool   `tfsdk:"encrypted_sip_traffic"`
	MaxExpiry            types.Int64  `tfsdk:"max_expiry"`
	RTPTimeout           types.Int64  `tfsdk:"rtp_timeout"`
	RTPHoldTimeout       types.Int64  `tfsdk:"rtp_hold_timeout"`
	IPRestriction        types.String `tfsdk:"ip_restriction"`
	EnableIPRestriction  types.Bool   `tfsdk:"enable_ip_restriction"`
	POPRestriction       types.String `tfsdk:"pop_restriction"`
	EnablePOPRestriction types.Bool   `tfsdk:"enable_pop_restriction"`
	RecordCalls          types.Bool   `tfsdk:"record_calls"`
	Allow225Balance      types.Bool   `tfsdk:"allow_225_balance"`
	InternalExtension    types.String `tfsdk:"internal_extension"`
	InternalVoicemail    types.String `tfsdk:"internal_voicemail"`
	InternalDialtime     types.String `tfsdk:"internal_dialtime"`
	EnableInternalCNAM   types.Bool   `tfsdk:"enable_internal_cnam"`
	DialingMode          types.String `tfsdk:"dialing_mode"`
	DefaultE911          types.String `tfsdk:"default_e911"`
	CallPickupBehavior   types.String `tfsdk:"call_pickup_behavior"`
}

func subaccountResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Numeric VoIP.ms sub-account id, assigned on create. Not used in DID routing (VoIP.ms routes by SIP login). Do not paste this into a DID; use `route`.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"route": computedRouteAttr("DID routing value (`account:{account}`). Use this for `voipms_did` `routing` / failover. VoIP.ms expects the SIP login, not the numeric `id`."),
		"account": schema.StringAttribute{
			MarkdownDescription: "Full SIP login (`{main}_{username}`, e.g. `100001_gateway`). Computed. Prefer `route` when linking a DID.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"username": schema.StringAttribute{
			MarkdownDescription: "Sub-account username suffix only (max 12 characters). Changing this forces a new resource.",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"password": schema.StringAttribute{
			MarkdownDescription: "SIP password (required for user/password auth).",
			Optional:            true,
			Computed:            true,
			Sensitive:           true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "Label shown in the portal.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"protocol": schema.StringAttribute{
			MarkdownDescription: "Protocol from `getProtocols`. Use `sip` (API `1`) or `iax2` (API `3`). Numeric ids still work.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"auth_type": schema.StringAttribute{
			MarkdownDescription: "Authentication type from `getAuthTypes`. Use `password` (user/password, API `1`) or `ip` (static IP, API `2`). Numeric ids still work.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"ip": schema.StringAttribute{
			MarkdownDescription: "Allowed IP or FQDN when `auth_type` is `ip`.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"device_type": schema.StringAttribute{
			MarkdownDescription: "Device type from `getDeviceTypes`. Use `ip_pbx` (Asterisk, IP PBX, Gateway or VoIP Switch; API `1`) or `ata` (ATA, IP Phone or Softphone; API `2`). Numeric ids still work.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"callerid_number": schema.StringAttribute{
			MarkdownDescription: "Outbound caller ID number.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"canada_routing": schema.StringAttribute{
			MarkdownDescription: "Canada routing from `getRoutes`. Use `value` (API `1`) or `premium` (API `2`). Numeric `1`/`2` still work.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"lock_international": schema.StringAttribute{
			MarkdownDescription: "Whether international calling is allowed. Use `allow` (API `0`) or `deny` (API `1`). Numeric ids still work.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"international_route": schema.StringAttribute{
			MarkdownDescription: "International route from `getRoutes`. Use `value` (API `1`) or `premium` (API `2`). Numeric ids still work.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"music_on_hold": schema.StringAttribute{
			MarkdownDescription: "Music on hold class (see `getMusicOnHold`).",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"language": schema.StringAttribute{
			MarkdownDescription: "IVR/system language (e.g. `en`).",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"allowed_codecs": schema.StringAttribute{
			MarkdownDescription: "Semicolon-separated codecs (e.g. `ulaw;g722`).",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"dtmf_mode": schema.StringAttribute{
			MarkdownDescription: "DTMF mode from `getDTMFModes` (e.g. `auto`).",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"nat": schema.StringAttribute{
			MarkdownDescription: "NAT setting from `getNAT` (`yes`, `no`, `route`, …).",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"encrypted_sip_traffic": schema.BoolAttribute{
			MarkdownDescription: encryptedSIPTrafficDescription,
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.Bool{optBool()},
		},
		"max_expiry": schema.Int64Attribute{
			MarkdownDescription: "Maximum SIP registration expiry in seconds.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.Int64{optInt()},
		},
		"rtp_timeout": schema.Int64Attribute{
			MarkdownDescription: "RTP timeout in seconds.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.Int64{optInt()},
		},
		"rtp_hold_timeout": schema.Int64Attribute{
			MarkdownDescription: "RTP hold timeout in seconds.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.Int64{optInt()},
		},
		"ip_restriction": schema.StringAttribute{
			MarkdownDescription: "Comma-separated IP allow-list when IP restriction is enabled.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"enable_ip_restriction": schema.BoolAttribute{
			MarkdownDescription: "Restrict registrations to `ip_restriction`.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.Bool{optBool()},
		},
		"pop_restriction": schema.StringAttribute{
			MarkdownDescription: "Comma-separated POP ids when `enable_pop_restriction` is true. " +
				"When restriction is off, VoIP.ms still returns the full POP list; Terraform treats this attribute as unset so configs do not have to store that list.",
			Optional:      true,
			Computed:      true,
			PlanModifiers: []planmodifier.String{optString()},
		},
		"enable_pop_restriction": schema.BoolAttribute{
			MarkdownDescription: "Restrict this sub-account to `pop_restriction` servers. Leave false for no POP restriction.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.Bool{optBool()},
		},
		"record_calls": schema.BoolAttribute{
			MarkdownDescription: "Record calls for this sub-account.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.Bool{optBool()},
		},
		"allow_225_balance": schema.BoolAttribute{
			MarkdownDescription: "Allow this sub-account to dial `*225` (or `*BAL`) to hear the current VoIP.ms account balance. " +
				"If disabled, calls to that feature code from this sub-account are rejected.\n\n" +
				"For a PBX or phone endpoint, leave this disabled unless users of that extension should be able to retrieve the account balance.",
			Optional:      true,
			Computed:      true,
			PlanModifiers: []planmodifier.Bool{optBool()},
		},
		"internal_extension": schema.StringAttribute{
			MarkdownDescription: "Internal extension digits.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"internal_voicemail": schema.StringAttribute{
			MarkdownDescription: "Internal voicemail mailbox. Set from `voipms_voicemail.this.id` or `data.voipms_voicemail.this.id`.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"internal_dialtime": schema.StringAttribute{
			MarkdownDescription: "Internal ring time in seconds.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"enable_internal_cnam": schema.BoolAttribute{
			MarkdownDescription: "Send internal Caller ID name.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.Bool{optBool()},
		},
		"dialing_mode": schema.StringAttribute{
			MarkdownDescription: "Outbound dialing mode. Use `main_account` (API `0`), `e164` (API `1`), or `nanpa` (API `2`). Numeric ids still work.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"default_e911": schema.StringAttribute{
			MarkdownDescription: "Default E911 DID for this sub-account.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"call_pickup_behavior": schema.StringAttribute{
			MarkdownDescription: "Call pickup permissions. Use `pickup_and_be_picked_up` (API `1`), `pickup_only` (API `2`), `be_picked_up_only` (API `3`), or `disabled` (API `4`). Numeric ids still work.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
	}
}

func (r *subaccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subaccount"
}

func (r *subaccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 2,
		MarkdownDescription: "Manages a VoIP.ms sub-account (`createSubAccount` / `setSubAccount` / `delSubAccount`). " +
			"Use this for SIP trunks (for example a FreeSWITCH gateway) and softphones. " +
			"Link a DID with `routing = voipms_subaccount.this.route`. Look up an existing trunk with `data.voipms_subaccount`.",
		Attributes: subaccountResourceAttributes(),
	}
}

func (r *subaccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

func (r *subaccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan subaccountModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := subaccountWriteParams(plan)
	params["username"] = plan.Username.ValueString()
	if err := r.client.CreateSubAccount(ctx, params); err != nil {
		resp.Diagnostics.AddError("Unable to create VoIP.ms sub-account", err.Error())
		return
	}

	acct, err := r.client.GetSubAccount(ctx, plan.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Sub-account created but could not be read back", err.Error())
		return
	}
	flattenSubaccount(acct, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *subaccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state subaccountModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lookup := state.ID.ValueString()
	if lookup == "" {
		lookup = state.Account.ValueString()
	}
	acct, err := r.client.GetSubAccount(ctx, lookup)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read VoIP.ms sub-account", err.Error())
		return
	}
	flattenSubaccount(acct, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *subaccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan subaccountModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetSubAccount(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VoIP.ms sub-account before update", err.Error())
		return
	}
	params := overlayParams(current.SetParams(), subaccountWriteParams(plan))
	params["id"] = plan.ID.ValueString()
	if err := r.client.UpdateSubAccount(ctx, params); err != nil {
		resp.Diagnostics.AddError("Unable to update VoIP.ms sub-account", err.Error())
		return
	}

	acct, err := r.client.GetSubAccount(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Sub-account updated but could not be read back", err.Error())
		return
	}
	flattenSubaccount(acct, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *subaccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state subaccountModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteSubAccount(ctx, state.ID.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return
		}
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.EmptyResult() {
			return
		}
		resp.Diagnostics.AddError("Unable to delete VoIP.ms sub-account", err.Error())
	}
}

func (r *subaccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func subaccountWriteParams(m subaccountModel) map[string]string {
	params := map[string]string{}
	setString(params, "description", m.Description)
	setNamed(params, "protocol", m.Protocol, client.Protocol)
	setNamed(params, "auth_type", m.AuthType, client.AuthType)
	setString(params, "password", m.Password)
	setString(params, "ip", m.IP)
	setNamed(params, "device_type", m.DeviceType, client.DeviceType)
	setString(params, "callerid_number", m.CallerIDNumber)
	setNamed(params, "canada_routing", m.CanadaRouting, client.CanadaRoutes)
	setNamed(params, "lock_international", m.LockInternational, client.LockInternational)
	setNamed(params, "international_route", m.InternationalRoute, client.CanadaRoutes)
	setString(params, "music_on_hold", m.MusicOnHold)
	setString(params, "language", m.Language)
	setString(params, "allowed_codecs", m.AllowedCodecs)
	setString(params, "dtmf_mode", m.DTMFMode)
	setString(params, "nat", m.NAT)
	setBool01(params, "sip_traffic", m.SIPTraffic)
	setInt(params, "max_expiry", m.MaxExpiry)
	setInt(params, "rtp_timeout", m.RTPTimeout)
	setInt(params, "rtp_hold_timeout", m.RTPHoldTimeout)
	setString(params, "ip_restriction", m.IPRestriction)
	setBool01(params, "enable_ip_restriction", m.EnableIPRestriction)
	setString(params, "pop_restriction", m.POPRestriction)
	setBool01(params, "enable_pop_restriction", m.EnablePOPRestriction)
	setBool01(params, "record_calls", m.RecordCalls)
	setBool01(params, "allow225", m.Allow225Balance)
	setString(params, "internal_extension", m.InternalExtension)
	setString(params, "internal_voicemail", m.InternalVoicemail)
	setString(params, "internal_dialtime", m.InternalDialtime)
	setBool01(params, "enable_internal_cnam", m.EnableInternalCNAM)
	setNamed(params, "dialing_mode", m.DialingMode, client.DialingMode)
	setString(params, "default_e911", m.DefaultE911)
	setNamed(params, "call_pickup_behavior", m.CallPickupBehavior, client.CallPickupBehavior)
	return params
}

func flattenSubaccount(src *client.SubAccount, dst *subaccountModel) {
	dst.ID = strVal(src.ID)
	dst.Route = types.StringValue(client.AccountRoute(src.Account.String()))
	dst.Account = strVal(src.Account)
	dst.Username = strVal(src.Username)
	dst.Description = strVal(src.Description)
	dst.Protocol = namedVal(src.Protocol, client.Protocol)
	dst.AuthType = namedVal(src.AuthType, client.AuthType)
	dst.Password = strVal(src.Password)
	dst.IP = strVal(src.IP)
	dst.DeviceType = namedVal(src.DeviceType, client.DeviceType)
	dst.CallerIDNumber = strVal(src.CallerIDNumber)
	dst.CanadaRouting = namedVal(src.CanadaRouting, client.CanadaRoutes)
	dst.LockInternational = namedVal(src.LockInternational, client.LockInternational)
	dst.InternationalRoute = namedVal(src.InternationalRoute, client.CanadaRoutes)
	dst.MusicOnHold = strVal(src.MusicOnHold)
	dst.Language = strVal(src.Language)
	dst.AllowedCodecs = strVal(src.AllowedCodecs)
	dst.DTMFMode = strVal(src.DTMFMode)
	dst.NAT = strVal(src.NAT)
	dst.SIPTraffic = boolVal(src.SIPTraffic)
	dst.MaxExpiry = intVal(src.MaxExpiry)
	dst.RTPTimeout = intVal(src.RTPTimeout)
	dst.RTPHoldTimeout = intVal(src.RTPHoldTimeout)
	dst.IPRestriction = strVal(src.IPRestriction)
	dst.EnableIPRestriction = boolVal(src.EnableIPRestriction)
	dst.EnablePOPRestriction = boolVal(src.EnablePOPRestriction)
	if src.EnablePOPRestriction.Bool() {
		dst.POPRestriction = strVal(src.POPRestriction)
	} else {
		dst.POPRestriction = types.StringNull()
	}
	dst.RecordCalls = boolVal(src.RecordCalls)
	dst.Allow225Balance = boolVal(src.Allow225)
	dst.InternalExtension = strVal(src.InternalExtension)
	dst.InternalVoicemail = strVal(src.InternalVoicemail)
	dst.InternalDialtime = strVal(src.InternalDialtime)
	dst.EnableInternalCNAM = boolVal(src.EnableInternalCNAM)
	dst.DialingMode = namedVal(src.DialingMode, client.DialingMode)
	dst.DefaultE911 = strVal(src.DefaultE911)
	dst.CallPickupBehavior = namedVal(src.CallPickupBehavior, client.CallPickupBehavior)
}

func flattenSubaccountCopy(src *client.SubAccount) subaccountModel {
	var m subaccountModel
	flattenSubaccount(src, &m)
	return m
}

func (r *subaccountResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}
	var plan subaccountModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.EnablePOPRestriction.IsUnknown() && !plan.EnablePOPRestriction.ValueBool() {
		plan.POPRestriction = types.StringNull()
	}
	if !plan.Account.IsNull() && !plan.Account.IsUnknown() && plan.Account.ValueString() != "" {
		plan.Route = types.StringValue(client.AccountRoute(plan.Account.ValueString()))
	}

	if !req.State.Raw.IsNull() {
		var state subaccountModel
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
		keepNamed(&plan.Protocol, state.Protocol, client.Protocol)
		keepNamed(&plan.AuthType, state.AuthType, client.AuthType)
		keepNamed(&plan.DeviceType, state.DeviceType, client.DeviceType)
		keepNamed(&plan.CanadaRouting, state.CanadaRouting, client.CanadaRoutes)
		keepNamed(&plan.LockInternational, state.LockInternational, client.LockInternational)
		keepNamed(&plan.InternationalRoute, state.InternationalRoute, client.CanadaRoutes)
		keepNamed(&plan.DialingMode, state.DialingMode, client.DialingMode)
		keepNamed(&plan.CallPickupBehavior, state.CallPickupBehavior, client.CallPickupBehavior)
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}
