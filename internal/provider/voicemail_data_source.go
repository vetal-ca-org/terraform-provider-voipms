package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vetal-ca-org/terraform-provider-voipms/internal/client"
)

func NewVoicemailDataSource() datasource.DataSource  { return &voicemailDataSource{} }
func NewVoicemailsDataSource() datasource.DataSource { return &voicemailsDataSource{} }

type voicemailDataSource struct{ client *client.Client }
type voicemailsDataSource struct{ client *client.Client }

func voicemailDataSourceAttributes(lookup bool) map[string]schema.Attribute {
	mailbox := schema.StringAttribute{MarkdownDescription: "Mailbox number.", Computed: true}
	name := schema.StringAttribute{MarkdownDescription: "Display name (e.g. `John`).", Computed: true}
	if lookup {
		mailbox.Optional = true
		name.Optional = true
	}
	return map[string]schema.Attribute{
		"id":                            schema.StringAttribute{MarkdownDescription: "VoIP.ms identifier (same as `mailbox`). Use this to attach the box to a DID: `voicemail = data.voipms_voicemail.this.id`.", Computed: true},
		"route":                         dsString("DID routing value (`vm:{mailbox}`). Use this for `voipms_did` `routing` / failover."),
		"mailbox":                       mailbox,
		"name":                          name,
		"password":                      dsSensitiveString("Mailbox PIN."),
		"skip_password":                 dsBool("Skip PIN prompt."),
		"email":                         dsString("Notification email."),
		"attach_message":                dsBool("Attach recording to email."),
		"delete_message":                dsBool("Delete after emailing."),
		"say_time":                      dsBool("Announce message time."),
		"timezone":                      dsString("Timezone."),
		"say_callerid":                  dsBool("Announce caller ID."),
		"play_instructions":             dsString("When to play mailbox instructions (`unread` or `skip_unread`)."),
		"language":                      dsString("Prompt language."),
		"email_attachment_format":       dsString("Email attachment format."),
		"unavailable_message_recording": dsString("Unavailable greeting recording id."),
		"transcription":                 dsBool("Voicemail transcription enabled."),
		"transcription_locale":          dsString("Transcription locale."),
		"transcription_format":          dsString("Transcription format (`text` or `html`)."),
		"transcription_redaction":       dsBool("Transcript redaction enabled."),
		"transcription_summary":         dsBool("Transcript summary enabled."),
		"transcription_sentiment":       dsBool("Transcript sentiment analysis enabled."),
	}
}

func (d *voicemailDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_voicemail"
}
func (d *voicemailDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a voicemail box by mailbox number or display `name` (`getVoicemails`). " +
			"Look up by `name`, then link with `id` or `route` — do not paste a raw mailbox into a DID.",
		Attributes: voicemailDataSourceAttributes(true),
	}
}
func (d *voicemailDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}
func (d *voicemailDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data voicemailModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	query := exactlyOneLookup(&resp.Diagnostics, "voicemail", []lookupField{
		{Name: "mailbox", Value: configuredString(data.Mailbox)},
		{Name: "name", Value: configuredString(data.Name)},
	})
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := d.client.FindVoicemail(ctx, query)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VoIP.ms voicemail", err.Error())
		return
	}
	flattenVoicemail(got, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type voicemailsModel struct {
	ID         types.String     `tfsdk:"id"`
	Voicemails []voicemailModel `tfsdk:"voicemails"`
}

func (d *voicemailsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_voicemails"
}
func (d *voicemailsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists voicemail boxes (`getVoicemails`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{MarkdownDescription: "Placeholder identifier (`voicemails`).", Computed: true},
			"voicemails": schema.ListNestedAttribute{
				MarkdownDescription: "Voicemail boxes on this account.",
				Computed:            true,
				NestedObject:        schema.NestedAttributeObject{Attributes: voicemailDataSourceAttributes(false)},
			},
		},
	}
}
func (d *voicemailsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}
func (d *voicemailsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data voicemailsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	items, err := d.client.GetVoicemails(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError("Unable to list VoIP.ms voicemails", err.Error())
		return
	}
	data.ID = types.StringValue("voicemails")
	data.Voicemails = make([]voicemailModel, 0, len(items))
	for i := range items {
		data.Voicemails = append(data.Voicemails, flattenVoicemailCopy(&items[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
