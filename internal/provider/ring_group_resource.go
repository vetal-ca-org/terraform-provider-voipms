package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vetal-ca-org/terraform-provider-voipms/internal/client"
)

var (
	_ resource.Resource                = &ringGroupResource{}
	_ resource.ResourceWithConfigure   = &ringGroupResource{}
	_ resource.ResourceWithImportState = &ringGroupResource{}
)

func NewRingGroupResource() resource.Resource { return &ringGroupResource{} }

type ringGroupResource struct{ client *client.Client }

type ringGroupMemberModel struct {
	Route    types.String `tfsdk:"route"`
	RingTime types.Int64  `tfsdk:"ring_time"`
	Press1   types.Bool   `tfsdk:"press1"`
}

type ringGroupModel struct {
	ID                 types.String           `tfsdk:"id"`
	Route              types.String           `tfsdk:"route"`
	Name               types.String           `tfsdk:"name"`
	Members            []ringGroupMemberModel `tfsdk:"members"`
	Voicemail          types.String           `tfsdk:"voicemail"`
	CallerAnnouncement types.String           `tfsdk:"caller_announcement"`
	MusicOnHold        types.String           `tfsdk:"music_on_hold"`
	Language           types.String           `tfsdk:"language"`
}

func (r *ringGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ring_group"
}

func (r *ringGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a ring group (`setRingGroup` / `delRingGroup`) — the fan-out that rings " +
			"several sub-accounts and forwardings at once. Point a DID at it with " +
			"`routing = voipms_ring_group.this.route`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "VoIP.ms ring group id, assigned on create. Computed. Do not paste this into a DID; use `route`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"route": computedRouteAttr("DID routing value (`grp:{id}`). Use this for `voipms_did` `routing` / failover."),
			"name":  schema.StringAttribute{MarkdownDescription: "Ring group name shown in the portal.", Required: true},
			"voicemail": schema.StringAttribute{
				MarkdownDescription: "Mailbox that takes the call when nobody answers. Set from `voipms_voicemail.this.id` or `data.voipms_voicemail.this.id`.",
				Required:            true,
			},
			"caller_announcement": optStr("Recording code played to the caller (values from `getRecordings`)."),
			"music_on_hold":       optStr("Music on hold class (values from `getMusicOnHold`)."),
			"language":            optStr("Language for system messages (values from `getLanguages`)."),
			"members": schema.ListNestedAttribute{
				MarkdownDescription: "Destinations rung by this group, in order. VoIP.ms stores these as one " +
					"`account:x,20,0;fwd:y,20,0` string; Terraform assembles it from these blocks.",
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"route": schema.StringAttribute{
							MarkdownDescription: "Member route. Set from `voipms_subaccount.this.route` or " +
								"`voipms_forwarding.this.route`, or a `sip:` URI. Ring groups cannot contain other ring groups.",
							Required: true,
						},
						"ring_time": schema.Int64Attribute{
							MarkdownDescription: "Seconds to ring this member (1–60).",
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(20),
						},
						"press1": schema.BoolAttribute{
							MarkdownDescription: "Require the member to press 1 before the call connects.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
					},
				},
			},
		},
	}
}

func (r *ringGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

func (r *ringGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ringGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	params, err := ringGroupWriteParams(plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ring group members", err.Error())
		return
	}
	id, err := r.client.CreateRingGroup(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create VoIP.ms ring group", err.Error())
		return
	}
	got, err := r.client.GetRingGroup(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Ring group created but could not be read back", err.Error())
		return
	}
	flattenRingGroup(got, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ringGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ringGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetRingGroup(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read VoIP.ms ring group", err.Error())
		return
	}
	flattenRingGroup(got, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ringGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ringGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	current, err := r.client.GetRingGroup(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VoIP.ms ring group before update", err.Error())
		return
	}
	updates, err := ringGroupWriteParams(plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ring group members", err.Error())
		return
	}
	params := overlayParams(current.SetParams(), updates)
	params["ring_group"] = plan.ID.ValueString()
	if err := r.client.UpdateRingGroup(ctx, params); err != nil {
		resp.Diagnostics.AddError("Unable to update VoIP.ms ring group", err.Error())
		return
	}
	got, err := r.client.GetRingGroup(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Ring group updated but could not be read back", err.Error())
		return
	}
	flattenRingGroup(got, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ringGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ringGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteRingGroup(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete VoIP.ms ring group", err.Error())
	}
}

func (r *ringGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func ringGroupWriteParams(m ringGroupModel) (map[string]string, error) {
	members, err := encodeRingGroupMembers(m.Members)
	if err != nil {
		return nil, err
	}
	params := map[string]string{"members": members}
	setString(params, "name", m.Name)
	setString(params, "voicemail", m.Voicemail)
	setString(params, "caller_announcement", m.CallerAnnouncement)
	setString(params, "music_on_hold", m.MusicOnHold)
	setString(params, "language", m.Language)
	return params, nil
}

// encodeRingGroupMembers renders the members blocks as the single
// `account:x,20,0;fwd:y,10,1` string setRingGroup expects.
func encodeRingGroupMembers(members []ringGroupMemberModel) (string, error) {
	parts := make([]string, 0, len(members))
	for i, m := range members {
		route := strings.TrimSpace(m.Route.ValueString())
		if route == "" {
			return "", fmt.Errorf("members[%d]: route is empty", i)
		}
		if strings.HasPrefix(strings.ToLower(route), client.RouteKindGroup+":") {
			return "", fmt.Errorf("members[%d]: %q is a ring group; ring groups cannot contain other ring groups", i, route)
		}
		ring := int64(20)
		if !m.RingTime.IsNull() && !m.RingTime.IsUnknown() {
			ring = m.RingTime.ValueInt64()
		}
		if ring < 1 || ring > 60 {
			return "", fmt.Errorf("members[%d]: ring_time %d is outside 1-60", i, ring)
		}
		press1 := "0"
		if m.Press1.ValueBool() {
			press1 = "1"
		}
		parts = append(parts, route+","+strconv.FormatInt(ring, 10)+","+press1)
	}
	if len(parts) == 0 {
		return "", errors.New("a ring group needs at least one member")
	}
	return strings.Join(parts, ";"), nil
}

// decodeRingGroupMembers parses what getRingGroups returns. Ring time and
// press1 are absent on groups created through the portal.
func decodeRingGroupMembers(raw string) []ringGroupMemberModel {
	out := []ringGroupMemberModel{}
	for _, part := range strings.Split(raw, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		fields := strings.Split(part, ",")
		member := ringGroupMemberModel{
			Route:    types.StringValue(strings.TrimSpace(fields[0])),
			RingTime: types.Int64Value(20),
			Press1:   types.BoolValue(false),
		}
		if len(fields) > 1 {
			if n, err := strconv.ParseInt(strings.TrimSpace(fields[1]), 10, 64); err == nil {
				member.RingTime = types.Int64Value(n)
			}
		}
		if len(fields) > 2 {
			member.Press1 = types.BoolValue(strings.TrimSpace(fields[2]) == "1")
		}
		out = append(out, member)
	}
	return out
}

func flattenRingGroup(src *client.RingGroup, dst *ringGroupModel) {
	dst.ID = strVal(src.RingGroup)
	dst.Route = types.StringValue(client.RingGroupRoute(src.RingGroup.String()))
	dst.Name = strVal(src.Name)
	dst.Members = decodeRingGroupMembers(src.Members.String())
	dst.Voicemail = strVal(src.Voicemail)
	dst.CallerAnnouncement = strVal(src.CallerAnnouncement)
	dst.MusicOnHold = strVal(src.MusicOnHold)
	dst.Language = strVal(src.Language)
}

func flattenRingGroupCopy(src *client.RingGroup) ringGroupModel {
	var m ringGroupModel
	flattenRingGroup(src, &m)
	return m
}
