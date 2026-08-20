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

var (
	_ resource.Resource                = &subaccountResource{}
	_ resource.ResourceWithConfigure   = &subaccountResource{}
	_ resource.ResourceWithImportState = &subaccountResource{}
)

func NewSubaccountResource() resource.Resource {
	return &subaccountResource{}
}

type subaccountResource struct {
	client *client.Client
}

type subaccountModel struct {
	ID                   types.String `tfsdk:"id"`
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
	SIPTraffic           types.Bool   `tfsdk:"sip_traffic"`
	MaxExpiry            types.Int64  `tfsdk:"max_expiry"`
	RTPTimeout           types.Int64  `tfsdk:"rtp_timeout"`
	RTPHoldTimeout       types.Int64  `tfsdk:"rtp_hold_timeout"`
	IPRestriction        types.String `tfsdk:"ip_restriction"`
	EnableIPRestriction  types.Bool   `tfsdk:"enable_ip_restriction"`
	POPRestriction       types.String `tfsdk:"pop_restriction"`
	EnablePOPRestriction types.Bool   `tfsdk:"enable_pop_restriction"`
	RecordCalls          types.Bool   `tfsdk:"record_calls"`
	Allow225             types.Bool   `tfsdk:"allow225"`
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
			MarkdownDescription: "Numeric VoIP.ms sub-account id.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"account": schema.StringAttribute{
			MarkdownDescription: "Full SIP login (`{main}_{username}`, e.g. `100001_gateway`).",
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
			MarkdownDescription: "Protocol id from `getProtocols` (`1` = SIP).",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"auth_type": schema.StringAttribute{
			MarkdownDescription: "Authentication type from `getAuthTypes` (`1` = user/password, `2` = IP).",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"ip": schema.StringAttribute{
			MarkdownDescription: "Allowed IP or FQDN when `auth_type` is IP authentication.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"device_type": schema.StringAttribute{
			MarkdownDescription: "Device type from `getDeviceTypes` (`1` = IP PBX, `2` = ATA/softphone).",
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
			MarkdownDescription: "Canada routing from `getRoutes` (`1` = Value, `2` = Premium).",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"lock_international": schema.StringAttribute{
			MarkdownDescription: "International lock from `getLockInternational`.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"international_route": schema.StringAttribute{
			MarkdownDescription: "International route from `getRoutes`.",
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
		"sip_traffic": schema.BoolAttribute{
			MarkdownDescription: "Whether encrypted SIP traffic is enabled.",
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
			MarkdownDescription: "Comma-separated POP ids when POP restriction is enabled.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"enable_pop_restriction": schema.BoolAttribute{
			MarkdownDescription: "Restrict this sub-account to `pop_restriction` servers.",
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
		"allow225": schema.BoolAttribute{
			MarkdownDescription: "Allow `*225` balance check.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.Bool{optBool()},
		},
		"internal_extension": schema.StringAttribute{
			MarkdownDescription: "Internal extension digits.",
			Optional:            true,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{optString()},
		},
		"internal_voicemail": schema.StringAttribute{
			MarkdownDescription: "Internal voicemail mailbox.",
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
			MarkdownDescription: "Dialing mode (`0` = use main account setting).",
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
			MarkdownDescription: "Call pickup behavior.",
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
		MarkdownDescription: "Manages a VoIP.ms sub-account (`createSubAccount` / `setSubAccount` / `delSubAccount`). " +
			"Use this for SIP trunks (for example a FreeSWITCH gateway) and softphones.",
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
	setString(params, "protocol", m.Protocol)
	setString(params, "auth_type", m.AuthType)
	setString(params, "password", m.Password)
	setString(params, "ip", m.IP)
	setString(params, "device_type", m.DeviceType)
	setString(params, "callerid_number", m.CallerIDNumber)
	setString(params, "canada_routing", m.CanadaRouting)
	setString(params, "lock_international", m.LockInternational)
	setString(params, "international_route", m.InternationalRoute)
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
	setBool01(params, "allow225", m.Allow225)
	setString(params, "internal_extension", m.InternalExtension)
	setString(params, "internal_voicemail", m.InternalVoicemail)
	setString(params, "internal_dialtime", m.InternalDialtime)
	setBool01(params, "enable_internal_cnam", m.EnableInternalCNAM)
	setString(params, "dialing_mode", m.DialingMode)
	setString(params, "default_e911", m.DefaultE911)
	setString(params, "call_pickup_behavior", m.CallPickupBehavior)
	return params
}

func flattenSubaccount(src *client.SubAccount, dst *subaccountModel) {
	dst.ID = strVal(src.ID)
	dst.Account = strVal(src.Account)
	dst.Username = strVal(src.Username)
	dst.Description = strVal(src.Description)
	dst.Protocol = strVal(src.Protocol)
	dst.AuthType = strVal(src.AuthType)
	dst.Password = strVal(src.Password)
	dst.IP = strVal(src.IP)
	dst.DeviceType = strVal(src.DeviceType)
	dst.CallerIDNumber = strVal(src.CallerIDNumber)
	dst.CanadaRouting = strVal(src.CanadaRouting)
	dst.LockInternational = strVal(src.LockInternational)
	dst.InternationalRoute = strVal(src.InternationalRoute)
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
	dst.POPRestriction = strVal(src.POPRestriction)
	dst.EnablePOPRestriction = boolVal(src.EnablePOPRestriction)
	dst.RecordCalls = boolVal(src.RecordCalls)
	dst.Allow225 = boolVal(src.Allow225)
	dst.InternalExtension = strVal(src.InternalExtension)
	dst.InternalVoicemail = strVal(src.InternalVoicemail)
	dst.InternalDialtime = strVal(src.InternalDialtime)
	dst.EnableInternalCNAM = boolVal(src.EnableInternalCNAM)
	dst.DialingMode = strVal(src.DialingMode)
	dst.DefaultE911 = strVal(src.DefaultE911)
	dst.CallPickupBehavior = strVal(src.CallPickupBehavior)
}

func flattenSubaccountCopy(src *client.SubAccount) subaccountModel {
	var m subaccountModel
	flattenSubaccount(src, &m)
	return m
}
