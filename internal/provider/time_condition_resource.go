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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vetal-ca-org/terraform-provider-voipms/internal/client"
)

var (
	_ resource.Resource                = &timeConditionResource{}
	_ resource.ResourceWithConfigure   = &timeConditionResource{}
	_ resource.ResourceWithImportState = &timeConditionResource{}
)

func NewTimeConditionResource() resource.Resource { return &timeConditionResource{} }

type timeConditionResource struct{ client *client.Client }

type timeConditionWindowModel struct {
	StartDay    types.String `tfsdk:"start_day"`
	EndDay      types.String `tfsdk:"end_day"`
	StartHour   types.Int64  `tfsdk:"start_hour"`
	StartMinute types.Int64  `tfsdk:"start_minute"`
	EndHour     types.Int64  `tfsdk:"end_hour"`
	EndMinute   types.Int64  `tfsdk:"end_minute"`
}

type timeConditionModel struct {
	ID             types.String               `tfsdk:"id"`
	Route          types.String               `tfsdk:"route"`
	Name           types.String               `tfsdk:"name"`
	RoutingMatch   types.String               `tfsdk:"routing_match"`
	RoutingNomatch types.String               `tfsdk:"routing_nomatch"`
	Windows        []timeConditionWindowModel `tfsdk:"windows"`
}

func (r *timeConditionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_time_condition"
}

func (r *timeConditionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a time condition (`setTimeCondition` / `delTimeCondition`) — the branch that " +
			"sends a call one way inside business hours and another way outside them. Point a DID at it with " +
			"`routing = voipms_time_condition.this.route`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "VoIP.ms time condition id, assigned on create. Computed. Do not paste this into a DID; use `route`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"route": computedRouteAttr("DID routing value (`tc:{id}`). Use this for `voipms_did` `routing` / failover."),
			"name": schema.StringAttribute{
				MarkdownDescription: "Time condition name shown in the portal. VoIP.ms stores at most 15 characters " +
					"and silently truncates anything longer, so this provider rejects it instead.",
				Required: true,
			},
			"routing_match": schema.StringAttribute{
				MarkdownDescription: "Where the call goes inside any of the `windows`. Set from a resource or data source " +
					"`route` (`voipms_ring_group.this.route`, `voipms_subaccount.this.route`) or a system action such as `sys:hangup`.",
				Required: true,
			},
			"routing_nomatch": schema.StringAttribute{
				MarkdownDescription: "Where the call goes outside every window. Same `route` reference as `routing_match`.",
				Required:            true,
			},
			"windows": schema.ListNestedAttribute{
				MarkdownDescription: "Periods the condition matches, as weekday range plus time-of-day range. " +
					"VoIP.ms stores these as six parallel `8;9` / `mon;sat` lists; Terraform assembles them from these blocks.",
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"start_day": schema.StringAttribute{
							MarkdownDescription: "First weekday of the range: `mon` `tue` `wed` `thu` `fri` `sat` `sun`.",
							Required:            true,
						},
						"end_day": schema.StringAttribute{
							MarkdownDescription: "Last weekday of the range, inclusive. Same day as `start_day` for a single day.",
							Required:            true,
						},
						"start_hour": schema.Int64Attribute{
							MarkdownDescription: "Hour the window opens (0-23).",
							Required:            true,
						},
						"start_minute": schema.Int64Attribute{
							MarkdownDescription: "Minute the window opens (0-59).",
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(0),
						},
						"end_hour": schema.Int64Attribute{
							MarkdownDescription: "Hour the window closes (0-23).",
							Required:            true,
						},
						"end_minute": schema.Int64Attribute{
							MarkdownDescription: "Minute the window closes (0-59).",
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(0),
						},
					},
				},
			},
		},
	}
}

func (r *timeConditionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureClient(req.ProviderData, &resp.Diagnostics)
}

func (r *timeConditionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan timeConditionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	params, err := timeConditionWriteParams(plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid time condition", err.Error())
		return
	}
	id, err := r.client.CreateTimeCondition(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create VoIP.ms time condition", err.Error())
		return
	}
	got, err := r.client.GetTimeCondition(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Time condition created but could not be read back", err.Error())
		return
	}
	flattenTimeCondition(got, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *timeConditionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state timeConditionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	got, err := r.client.GetTimeCondition(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read VoIP.ms time condition", err.Error())
		return
	}
	flattenTimeCondition(got, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *timeConditionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan timeConditionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	current, err := r.client.GetTimeCondition(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read VoIP.ms time condition before update", err.Error())
		return
	}
	updates, err := timeConditionWriteParams(plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid time condition", err.Error())
		return
	}
	params := overlayParams(current.SetParams(), updates)
	params["timecondition"] = plan.ID.ValueString()
	if err := r.client.UpdateTimeCondition(ctx, params); err != nil {
		resp.Diagnostics.AddError("Unable to update VoIP.ms time condition", err.Error())
		return
	}
	got, err := r.client.GetTimeCondition(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Time condition updated but could not be read back", err.Error())
		return
	}
	flattenTimeCondition(got, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *timeConditionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state timeConditionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteTimeCondition(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to delete VoIP.ms time condition", err.Error())
	}
}

func (r *timeConditionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// setTimeCondition cuts a longer name down to 15 characters and reports success,
// which surfaces as Terraform's "provider produced inconsistent result" rather
// than as anything to do with the name.
const timeConditionNameMax = 15

func timeConditionWriteParams(m timeConditionModel) (map[string]string, error) {
	if name := m.Name.ValueString(); len(name) > timeConditionNameMax {
		return nil, fmt.Errorf("name %q is %d characters; VoIP.ms stores at most %d",
			name, len(name), timeConditionNameMax)
	}
	params, err := encodeTimeConditionWindows(m.Windows)
	if err != nil {
		return nil, err
	}
	setString(params, "name", m.Name)
	setString(params, "routing_match", m.RoutingMatch)
	setString(params, "routing_nomatch", m.RoutingNomatch)
	return params, nil
}

var weekdays = []string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"}

// encodeWeekday lowercases and checks one weekday. VoIP.ms accepts any case but
// stores lowercase, so anything else would drift on the next read.
func encodeWeekday(field string, v types.String) (string, error) {
	day := strings.ToLower(strings.TrimSpace(v.ValueString()))
	for _, known := range weekdays {
		if day == known {
			return day, nil
		}
	}
	return "", fmt.Errorf("%s %q is not one of %s", field, v.ValueString(), strings.Join(weekdays, ", "))
}

func encodeClock(field string, v types.Int64, max int64) (string, error) {
	n := v.ValueInt64()
	if n < 0 || n > max {
		return "", fmt.Errorf("%s %d is outside 0-%d", field, n, max)
	}
	return strconv.FormatInt(n, 10), nil
}

type encodedTimeConditionWindow struct {
	StartDay    string
	EndDay      string
	StartHour   string
	StartMinute string
	EndHour     string
	EndMinute   string
}

func encodeTimeConditionWindow(w timeConditionWindowModel) (encodedTimeConditionWindow, error) {
	var out encodedTimeConditionWindow
	var err error
	if out.StartDay, err = encodeWeekday("start_day", w.StartDay); err != nil {
		return out, err
	}
	if out.EndDay, err = encodeWeekday("end_day", w.EndDay); err != nil {
		return out, err
	}
	if out.StartHour, err = encodeClock("start_hour", w.StartHour, 23); err != nil {
		return out, err
	}
	if out.StartMinute, err = encodeClock("start_minute", w.StartMinute, 59); err != nil {
		return out, err
	}
	if out.EndHour, err = encodeClock("end_hour", w.EndHour, 23); err != nil {
		return out, err
	}
	if out.EndMinute, err = encodeClock("end_minute", w.EndMinute, 59); err != nil {
		return out, err
	}
	return out, nil
}

// encodeTimeConditionWindows renders the windows blocks as the six parallel
// semicolon lists setTimeCondition expects (`starthour` "8;9", `weekdaystart`
// "mon;sat"). Building all six from one list keeps their lengths equal, which
// the API requires.
func encodeTimeConditionWindows(windows []timeConditionWindowModel) (map[string]string, error) {
	if len(windows) == 0 {
		return nil, errors.New("a time condition needs at least one window")
	}
	var startDay, endDay, startHour, startMinute, endHour, endMinute []string
	for i, w := range windows {
		enc, err := encodeTimeConditionWindow(w)
		if err != nil {
			return nil, fmt.Errorf("windows[%d]: %w", i, err)
		}
		startDay = append(startDay, enc.StartDay)
		endDay = append(endDay, enc.EndDay)
		startHour = append(startHour, enc.StartHour)
		startMinute = append(startMinute, enc.StartMinute)
		endHour = append(endHour, enc.EndHour)
		endMinute = append(endMinute, enc.EndMinute)
	}
	return map[string]string{
		"weekdaystart": strings.Join(startDay, ";"),
		"weekdayend":   strings.Join(endDay, ";"),
		"starthour":    strings.Join(startHour, ";"),
		"startminute":  strings.Join(startMinute, ";"),
		"endhour":      strings.Join(endHour, ";"),
		"endminute":    strings.Join(endMinute, ";"),
	}, nil
}

func splitWindowList(raw string) []string {
	out := []string{}
	for _, part := range strings.Split(raw, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func windowInt(values []string, i int) types.Int64 {
	n, err := strconv.ParseInt(values[i], 10, 64)
	if err != nil {
		return types.Int64Value(0)
	}
	return types.Int64Value(n)
}

// decodeTimeConditionWindows zips the six parallel lists back into windows. The
// shortest list decides how many there are; VoIP.ms will not store a condition
// whose lists differ in length, so that only bites on a hand-edited account.
func decodeTimeConditionWindows(src *client.TimeCondition) []timeConditionWindowModel {
	startDay := splitWindowList(src.WeekdayStart.String())
	endDay := splitWindowList(src.WeekdayEnd.String())
	startHour := splitWindowList(src.StartHour.String())
	startMinute := splitWindowList(src.StartMinute.String())
	endHour := splitWindowList(src.EndHour.String())
	endMinute := splitWindowList(src.EndMinute.String())

	count := min(len(startDay), len(endDay), len(startHour), len(startMinute), len(endHour), len(endMinute))
	out := make([]timeConditionWindowModel, 0, count)
	for i := range count {
		out = append(out, timeConditionWindowModel{
			StartDay:    types.StringValue(startDay[i]),
			EndDay:      types.StringValue(endDay[i]),
			StartHour:   windowInt(startHour, i),
			StartMinute: windowInt(startMinute, i),
			EndHour:     windowInt(endHour, i),
			EndMinute:   windowInt(endMinute, i),
		})
	}
	return out
}

func flattenTimeCondition(src *client.TimeCondition, dst *timeConditionModel) {
	dst.ID = strVal(src.TimeCondition)
	dst.Route = types.StringValue(client.TimeConditionRoute(src.TimeCondition.String()))
	dst.Name = strVal(src.Name)
	dst.RoutingMatch = strVal(src.RoutingMatch)
	dst.RoutingNomatch = strVal(src.RoutingNomatch)
	dst.Windows = decodeTimeConditionWindows(src)
}

func flattenTimeConditionCopy(src *client.TimeCondition) timeConditionModel {
	var m timeConditionModel
	flattenTimeCondition(src, &m)
	return m
}
