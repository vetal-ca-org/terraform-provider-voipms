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
)

func NewVoicemailResource() resource.Resource { return &voicemailResource{} }

type voicemailResource struct{ client *client.Client }

type voicemailModel struct {
	ID                          types.String `tfsdk:"id"`
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
}

func (r *voicemailResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_voicemail"
}

func (r *voicemailResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a voicemail box (`createVoicemail` / `setVoicemail` / `delVoicemail`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Same as `mailbox`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"mailbox": schema.StringAttribute{
				MarkdownDescription: "Mailbox number (digits). Changing this forces a new resource.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name":                          schema.StringAttribute{MarkdownDescription: "Display name.", Required: true},
			"password":                      schema.StringAttribute{MarkdownDescription: "Mailbox PIN.", Required: true, Sensitive: true},
			"skip_password":                 optBoolAttr("Skip the PIN prompt when checking voicemail from a trusted DID."),
			"email":                         optStr("Notification email; comma-separated for multiple addresses."),
			"attach_message":                optBoolAttr("Attach the recording to the notification email."),
			"delete_message":                optBoolAttr("Delete the message from the portal after emailing it."),
			"say_time":                      optBoolAttr("Announce the message time."),
			"timezone":                      optStr("Timezone (e.g. `America/Montreal`)."),
			"say_callerid":                  optBoolAttr("Announce the caller ID."),
			"play_instructions":             optStr("When to play instructions (`u` = unavailable greeting, etc.)."),
			"language":                      optStr("Prompt language (e.g. `en`)."),
			"email_attachment_format":       optStr("Attachment format (e.g. `wav49`)."),
			"unavailable_message_recording": optStr("Unavailable greeting recording id."),
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
	setString(params, "play_instructions", m.PlayInstructions)
	setString(params, "language", m.Language)
	setString(params, "email_attachment_format", m.EmailAttachmentFormat)
	setString(params, "unavailable_message_recording", m.UnavailableMessageRecording)
	return params
}

func flattenVoicemail(src *client.Voicemail, dst *voicemailModel) {
	dst.ID = strVal(src.Mailbox)
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
	dst.PlayInstructions = strVal(src.PlayInstructions)
	dst.Language = strVal(src.Language)
	dst.EmailAttachmentFormat = strVal(src.EmailAttachmentFormat)
	dst.UnavailableMessageRecording = strVal(src.UnavailableMessageRecording)
}

func flattenVoicemailCopy(src *client.Voicemail) voicemailModel {
	var m voicemailModel
	flattenVoicemail(src, &m)
	return m
}
