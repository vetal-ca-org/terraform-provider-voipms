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
	_ resource.Resource                = &voicemailResource{}
	_ resource.ResourceWithConfigure   = &voicemailResource{}
	_ resource.ResourceWithImportState = &voicemailResource{}
	_ resource.ResourceWithModifyPlan  = &voicemailResource{}
)

func NewVoicemailResource() resource.Resource { return &voicemailResource{} }

type voicemailResource struct{ client *client.Client }

type voicemailModel struct {
	ID                          types.String `tfsdk:"id"`
	Route                       types.String `tfsdk:"route"`
	Mailbox                     types.String `tfsdk:"mailbox"`
	Name                        types.String `tfsdk:"name"`
	Password                    types.String `tfsdk:"password"`
	SkipPassword                types.Bool   `tfsdk:"skip_password"`
	Email                       types.String `tfsdk:"email"`
	AttachMessage               types.Bool   `tfsdk:"attach_message"`
	DeleteMessage               types.Bool   `tfsdk:"delete_message"`
	SayTime                     types.Bool   `tfsdk:"say_time"`
	Timezone                    types.String `tfsdk:"timezone"`
	SayCallerID                 types.Bool   `tfsdk:"say_callerid"`
	PlayInstructions            types.String `tfsdk:"play_instructions"`
	Language                    types.String `tfsdk:"language"`
	EmailAttachmentFormat       types.String `tfsdk:"email_attachment_format"`
	UnavailableMessageRecording types.String `tfsdk:"unavailable_message_recording"`
	Transcription               types.Bool   `tfsdk:"transcription"`
	TranscriptionLocale         types.String `tfsdk:"transcription_locale"`
	TranscriptionFormat         types.String `tfsdk:"transcription_format"`
	TranscriptionRedaction      types.Bool   `tfsdk:"transcription_redaction"`
	TranscriptionSummary        types.Bool   `tfsdk:"transcription_summary"`
	TranscriptionSentiment      types.Bool   `tfsdk:"transcription_sentiment"`
}

func (r *voicemailResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_voicemail"
}

func (r *voicemailResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a voicemail box (`createVoicemail` / `setVoicemail` / `delVoicemail`). " +
			"Link a DID with `voicemail = voipms_voicemail.this.id` or `routing = voipms_voicemail.this.route`. " +
			"Look up an existing box with `data.voipms_voicemail` (by `name` or `mailbox`) and use that object's `id` / `route`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "VoIP.ms identifier (same as `mailbox`). Computed. Attach this mailbox to a DID with `voicemail = voipms_voicemail.this.id` (or `data.voipms_voicemail.this.id`). Do not paste a raw mailbox number.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"route": computedRouteAttr("DID routing value (`vm:{mailbox}`). Use this for `voipms_did` `routing` / failover, not `vm:` plus a display name."),
			"mailbox": schema.StringAttribute{
				MarkdownDescription: "Mailbox number people dial. You choose it; VoIP.ms does not assign it. This is also the API identifier. Changing this forces a new resource. Link other objects with `id` or `route`, not a mailbox typed as a literal.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name":                          schema.StringAttribute{MarkdownDescription: "Display name. Used to create or look up this box, not to link other resources.", Required: true},
			"password":                      schema.StringAttribute{MarkdownDescription: "Mailbox PIN.", Required: true, Sensitive: true},
			"skip_password":                 optBoolAttr("Skip the PIN prompt when checking voicemail from a trusted DID."),
			"email":                         optStr("Notification email; comma-separated for multiple addresses."),
			"attach_message":                optBoolAttr("Attach the recording to the notification email."),
			"delete_message":                optBoolAttr("Delete the message from the portal after emailing it."),
			"say_time":                      optBoolAttr("Announce the message time."),
			"timezone":                      optStr("Timezone (e.g. `America/Montreal`)."),
			"say_callerid":                  optBoolAttr("Announce the caller ID."),
			"play_instructions":             optStr("When to play mailbox instructions. Use `unread` (API `u`) or `skip_unread` (API `su`). Short codes still work."),
			"language":                      optStr("Prompt language (e.g. `en`)."),
			"email_attachment_format":       optStr("Attachment format (e.g. `wav49`)."),
			"unavailable_message_recording": optStr("Unavailable greeting recording id. Set from `data.voipms_recording.this.id`."),
			"transcription":                 optBoolAttr("Transcribe voicemail messages. Billed per minute by VoIP.ms."),
			"transcription_locale":          optStr("Transcription locale (values from `getLocales`), comma-separated for up to 10."),
			"transcription_format":          optStr("Transcription format (`text` or `html`)."),
			"transcription_redaction":       optBoolAttr("Redact sensitive data in the transcript."),
			"transcription_summary":         optBoolAttr("Add a summary to the transcript. Billed per minute."),
			"transcription_sentiment":       optBoolAttr("Add sentiment analysis to the transcript. Billed per minute."),
		},
	}
}

func (r *voicemailResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

func (r *voicemailResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan voicemailModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	params := voicemailWriteParams(plan)
	params["digits"] = plan.Mailbox.ValueString()
	delete(params, "mailbox")
	if err := r.client.CreateVoicemail(ctx, params); err != nil {
		resp.Diagnostics.AddError("Unable to create VoIP.ms voicemail", err.Error())
		return
	}
	got, err := r.client.GetVoicemail(ctx, plan.Mailbox.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Voicemail created but could not be read back", err.Error())
		return
	}
	flattenVoicemail(got, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *voicemailResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state voicemailModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetVoicemail(ctx, state.Mailbox.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read VoIP.ms voicemail", err.Error())
		return
	}
	flattenVoicemail(got, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *voicemailResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan voicemailModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	current, err := r.client.GetVoicemail(ctx, plan.Mailbox.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VoIP.ms voicemail before update", err.Error())
		return
	}
	params := overlayParams(current.SetParams(), voicemailWriteParams(plan))
	params["mailbox"] = plan.Mailbox.ValueString()
	if err := r.client.UpdateVoicemail(ctx, params); err != nil {
		resp.Diagnostics.AddError("Unable to update VoIP.ms voicemail", err.Error())
		return
	}
	got, err := r.client.GetVoicemail(ctx, plan.Mailbox.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Voicemail updated but could not be read back", err.Error())
		return
	}
	flattenVoicemail(got, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *voicemailResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state voicemailModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteVoicemail(ctx, state.Mailbox.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete VoIP.ms voicemail", err.Error())
	}
}

func (r *voicemailResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("mailbox"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func voicemailWriteParams(m voicemailModel) map[string]string {
	params := map[string]string{}
	setString(params, "name", m.Name)
	setString(params, "password", m.Password)
	setBoolYesNo(params, "skip_password", m.SkipPassword)
	setString(params, "email", m.Email)
	setBoolYesNo(params, "attach_message", m.AttachMessage)
	setBoolYesNo(params, "delete_message", m.DeleteMessage)
	setBoolYesNo(params, "say_time", m.SayTime)
	setString(params, "timezone", m.Timezone)
	setBoolYesNo(params, "say_callerid", m.SayCallerID)
	setNamed(params, "play_instructions", m.PlayInstructions, client.PlayInstructions)
	setString(params, "language", m.Language)
	setString(params, "email_attachment_format", m.EmailAttachmentFormat)
	setString(params, "unavailable_message_recording", m.UnavailableMessageRecording)
	setBoolYesNo(params, "transcription", m.Transcription)
	setString(params, "transcription_locale", m.TranscriptionLocale)
	setString(params, "transcription_format", m.TranscriptionFormat)
	setBoolYesNo(params, "transcription_redaction", m.TranscriptionRedaction)
	setBoolYesNo(params, "transcription_summary", m.TranscriptionSummary)
	setBoolYesNo(params, "transcription_sentiment", m.TranscriptionSentiment)
	return params
}

func (r *voicemailResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}
	var plan voicemailModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !plan.Mailbox.IsNull() && !plan.Mailbox.IsUnknown() && plan.Mailbox.ValueString() != "" {
		plan.ID = types.StringValue(plan.Mailbox.ValueString())
		plan.Route = types.StringValue(client.VoicemailRoute(plan.Mailbox.ValueString()))
	}
	if !req.State.Raw.IsNull() {
		var state voicemailModel
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
		keepNamed(&plan.PlayInstructions, state.PlayInstructions, client.PlayInstructions)
	}
	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

func flattenVoicemail(src *client.Voicemail, dst *voicemailModel) {
	dst.ID = strVal(src.Mailbox)
	dst.Route = types.StringValue(client.VoicemailRoute(src.Mailbox.String()))
	dst.Mailbox = strVal(src.Mailbox)
	dst.Name = strVal(src.Name)
	dst.Password = strVal(src.Password)
	dst.SkipPassword = boolVal(src.SkipPassword)
	dst.Email = strVal(src.Email)
	dst.AttachMessage = boolVal(src.AttachMessage)
	dst.DeleteMessage = boolVal(src.DeleteMessage)
	dst.SayTime = boolVal(src.SayTime)
	dst.Timezone = strVal(src.Timezone)
	dst.SayCallerID = boolVal(src.SayCallerID)
	dst.PlayInstructions = namedVal(src.PlayInstructions, client.PlayInstructions)
	dst.Language = strVal(src.Language)
	dst.EmailAttachmentFormat = strVal(src.EmailAttachmentFormat)
	dst.UnavailableMessageRecording = strVal(src.UnavailableMessageRecording)
	dst.Transcription = boolVal(src.Transcription)
	dst.TranscriptionLocale = strVal(src.TranscriptionLocale)
	dst.TranscriptionFormat = strVal(src.TranscriptionFormat)
	dst.TranscriptionRedaction = boolVal(src.TranscriptionRedaction)
	dst.TranscriptionSummary = boolVal(src.TranscriptionSummary)
	dst.TranscriptionSentiment = boolVal(src.TranscriptionSentiment)
}

func flattenVoicemailCopy(src *client.Voicemail) voicemailModel {
	var m voicemailModel
	flattenVoicemail(src, &m)
	return m
}
