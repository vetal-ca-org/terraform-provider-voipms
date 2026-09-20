package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vetal-ca-org/terraform-provider-voipms/internal/client"
)

func window(startDay, endDay string, startHour, startMinute, endHour, endMinute int64) timeConditionWindowModel {
	return timeConditionWindowModel{
		StartDay:    types.StringValue(startDay),
		EndDay:      types.StringValue(endDay),
		StartHour:   types.Int64Value(startHour),
		StartMinute: types.Int64Value(startMinute),
		EndHour:     types.Int64Value(endHour),
		EndMinute:   types.Int64Value(endMinute),
	}
}

func TestEncodeTimeConditionWindows(t *testing.T) {
	got, err := encodeTimeConditionWindows([]timeConditionWindowModel{
		window("mon", "fri", 8, 0, 16, 0),
		window("sat", "sat", 9, 30, 12, 0),
	})
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	for key, want := range map[string]string{
		"weekdaystart": "mon;sat",
		"weekdayend":   "fri;sat",
		"starthour":    "8;9",
		"startminute":  "0;30",
		"endhour":      "16;12",
		"endminute":    "0;0",
	} {
		if got[key] != want {
			t.Errorf("%s = %q, want %q", key, got[key], want)
		}
	}
}

// VoIP.ms accepts any case but stores lowercase, so an uppercase day in the
// config would otherwise drift on every read.
func TestEncodeTimeConditionWindowsLowercasesWeekdays(t *testing.T) {
	got, err := encodeTimeConditionWindows([]timeConditionWindowModel{window("MON", "Fri", 8, 0, 16, 0)})
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if got["weekdaystart"] != "mon" || got["weekdayend"] != "fri" {
		t.Errorf("weekdays = %q/%q, want mon/fri", got["weekdaystart"], got["weekdayend"])
	}
}

func TestEncodeTimeConditionWindowsRejectsBadInput(t *testing.T) {
	for name, windows := range map[string][]timeConditionWindowModel{
		"empty":         {},
		"unknown day":   {window("monday", "fri", 8, 0, 16, 0)},
		"blank day":     {window("", "fri", 8, 0, 16, 0)},
		"hour 24":       {window("mon", "fri", 24, 0, 16, 0)},
		"hour negative": {window("mon", "fri", -1, 0, 16, 0)},
		"minute 60":     {window("mon", "fri", 8, 60, 16, 0)},
		"end hour 24":   {window("mon", "fri", 8, 0, 24, 0)},
	} {
		if _, err := encodeTimeConditionWindows(windows); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestTimeConditionWriteParamsRejectsLongName(t *testing.T) {
	model := timeConditionModel{
		Name:           types.StringValue("123456789012345"),
		RoutingMatch:   types.StringValue("vm:1001"),
		RoutingNomatch: types.StringValue("sys:hangup"),
		Windows:        []timeConditionWindowModel{window("mon", "fri", 8, 0, 16, 0)},
	}
	if _, err := timeConditionWriteParams(model); err != nil {
		t.Fatalf("15 characters should be accepted: %v", err)
	}
	model.Name = types.StringValue("1234567890123456")
	if _, err := timeConditionWriteParams(model); err == nil {
		t.Error("16 characters should be rejected; VoIP.ms truncates it silently")
	}
}

func TestDecodeTimeConditionWindows(t *testing.T) {
	got := decodeTimeConditionWindows(&client.TimeCondition{
		WeekdayStart: "mon;sat",
		WeekdayEnd:   "fri;sat",
		StartHour:    "8;9",
		StartMinute:  "0;30",
		EndHour:      "16;12",
		EndMinute:    "0;0",
	})
	if len(got) != 2 {
		t.Fatalf("decoded %d windows, want 2", len(got))
	}
	if got[1].StartDay.ValueString() != "sat" || got[1].StartHour.ValueInt64() != 9 || got[1].StartMinute.ValueInt64() != 30 {
		t.Errorf("windows[1] = %+v", got[1])
	}
	if len(decodeTimeConditionWindows(&client.TimeCondition{})) != 0 {
		t.Error("an empty time condition should decode to no windows")
	}
}

// The portal zero-pads hours it writes; the API stores whatever it is sent, so
// "08" has to decode to the same window as "8" or every plan shows a diff.
func TestDecodeTimeConditionWindowsAcceptsZeroPadding(t *testing.T) {
	got := decodeTimeConditionWindows(&client.TimeCondition{
		WeekdayStart: "mon", WeekdayEnd: "fri",
		StartHour: "08", StartMinute: "05", EndHour: "16", EndMinute: "00",
	})
	if len(got) != 1 {
		t.Fatalf("decoded %d windows, want 1", len(got))
	}
	if got[0].StartHour.ValueInt64() != 8 || got[0].StartMinute.ValueInt64() != 5 {
		t.Errorf("window = %+v, want 8:05", got[0])
	}
}

func TestTimeConditionWindowsRoundTrip(t *testing.T) {
	src := &client.TimeCondition{
		WeekdayStart: "mon;sat",
		WeekdayEnd:   "fri;sun",
		StartHour:    "8;22",
		StartMinute:  "0;15",
		EndHour:      "16;6",
		EndMinute:    "30;0",
	}
	encoded, err := encodeTimeConditionWindows(decodeTimeConditionWindows(src))
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	for key, want := range map[string]string{
		"weekdaystart": src.WeekdayStart.String(),
		"weekdayend":   src.WeekdayEnd.String(),
		"starthour":    src.StartHour.String(),
		"startminute":  src.StartMinute.String(),
		"endhour":      src.EndHour.String(),
		"endminute":    src.EndMinute.String(),
	} {
		if encoded[key] != want {
			t.Errorf("round trip %s = %q, want %q", key, encoded[key], want)
		}
	}
}

// A DID pointed at a time condition by name has to reach the API as tc:{id}.
func TestResolveDIDRoutesResolvesTimeConditionByName(t *testing.T) {
	tables := client.RouteTables{TimeConditions: []client.TimeCondition{
		{TimeCondition: "1830", Name: "Office Hours"},
	}}
	plan := didModel{
		Routing:          types.StringValue("tc:Office Hours"),
		FailoverNoanswer: types.StringValue("tc:1830"),
	}
	if err := resolveDIDRoutes(&plan, tables); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if plan.Routing.ValueString() != "tc:1830" {
		t.Errorf("routing = %q, want tc:1830", plan.Routing.ValueString())
	}
	if plan.FailoverNoanswer.ValueString() != "tc:1830" {
		t.Errorf("failover_noanswer = %q, want tc:1830", plan.FailoverNoanswer.ValueString())
	}

	missing := didModel{Routing: types.StringValue("tc:nope")}
	if err := resolveDIDRoutes(&missing, tables); err == nil {
		t.Error("an unknown time condition name should be an error")
	}
}
