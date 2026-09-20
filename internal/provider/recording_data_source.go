package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vetal-ca-org/terraform-provider-voipms/internal/client"
)

func NewRecordingDataSource() datasource.DataSource  { return &recordingDataSource{} }
func NewRecordingsDataSource() datasource.DataSource { return &recordingsDataSource{} }

type recordingDataSource struct{ client *client.Client }
type recordingsDataSource struct{ client *client.Client }

type recordingModel struct {
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
}

func recordingDataSourceAttributes(lookup bool) map[string]schema.Attribute {
	id := schema.StringAttribute{
		MarkdownDescription: "Recording id. This is the value other resources want: " +
			"`voipms_voicemail` `unavailable_message_recording`, `voipms_ring_group` `caller_announcement`.",
		Computed: true,
	}
	description := schema.StringAttribute{
		MarkdownDescription: "Recording description, as typed in the portal (e.g. `Main Greeting`).",
		Computed:            true,
	}
	if lookup {
		id.Optional = true
		description.Optional = true
	}
	return map[string]schema.Attribute{"id": id, "description": description}
}

func flattenRecording(src *client.Recording, dst *recordingModel) {
	dst.ID = strVal(src.Recording)
	dst.Description = strVal(src.Description)
}

func flattenRecordingCopy(src *client.Recording) recordingModel {
	var m recordingModel
	flattenRecording(src, &m)
	return m
}

func (d *recordingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_recording"
}
func (d *recordingDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads an audio recording by id or `description` (`getRecordings`). " +
			"Recordings are uploaded in the VoIP.ms portal; this provider only reads them, so look one up " +
			"by description and pass its `id` wherever a recording is wanted instead of hardcoding the number.",
		Attributes: recordingDataSourceAttributes(true),
	}
}
func (d *recordingDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}
func (d *recordingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data recordingModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	query := exactlyOneLookup(&resp.Diagnostics, "recording", []lookupField{
		{Name: "id", Value: configuredString(data.ID)},
		{Name: "description", Value: configuredString(data.Description)},
	})
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := d.client.FindRecording(ctx, query)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VoIP.ms recording", err.Error())
		return
	}
	flattenRecording(got, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type recordingsModel struct {
	ID         types.String     `tfsdk:"id"`
	Recordings []recordingModel `tfsdk:"recordings"`
}

func (d *recordingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_recordings"
}
func (d *recordingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists audio recordings (`getRecordings`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{MarkdownDescription: "Placeholder identifier (`recordings`).", Computed: true},
			"recordings": schema.ListNestedAttribute{
				MarkdownDescription: "Recordings on this account.",
				Computed:            true,
				NestedObject:        schema.NestedAttributeObject{Attributes: recordingDataSourceAttributes(false)},
			},
		},
	}
}
func (d *recordingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = configureClient(req.ProviderData, &resp.Diagnostics)
}
func (d *recordingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data recordingsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	items, err := d.client.GetRecordings(ctx, "")
	if err != nil {
		resp.Diagnostics.AddError("Unable to list VoIP.ms recordings", err.Error())
		return
	}
	data.ID = types.StringValue("recordings")
	data.Recordings = make([]recordingModel, 0, len(items))
	for i := range items {
		data.Recordings = append(data.Recordings, flattenRecordingCopy(&items[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
