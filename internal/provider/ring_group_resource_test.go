package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func member(route string, ring int64, press1 bool) ringGroupMemberModel {
	return ringGroupMemberModel{
		Route:    types.StringValue(route),
		RingTime: types.Int64Value(ring),
		Press1:   types.BoolValue(press1),
	}
}

func TestEncodeRingGroupMembers(t *testing.T) {
	got, err := encodeRingGroupMembers([]ringGroupMemberModel{
		member("account:100001_gateway", 20, false),
		member("fwd:16006", 25, true),
	})
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	want := "account:100001_gateway,20,0;fwd:16006,25,1"
	if got != want {
		t.Errorf("encode = %q, want %q", got, want)
	}
}

func TestEncodeRingGroupMembersRejectsBadInput(t *testing.T) {
	for name, members := range map[string][]ringGroupMemberModel{
		"empty":            {},
		"blank route":      {member("", 20, false)},
		"nested group":     {member("grp:900001", 20, false)},
		"ring time zero":   {member("account:x", 0, false)},
		"ring time over60": {member("account:x", 61, false)},
	} {
		if _, err := encodeRingGroupMembers(members); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

// Groups created in the portal come back without the ring time / press1 fields.
func TestDecodeRingGroupMembers(t *testing.T) {
	got := decodeRingGroupMembers("account:100001_a,20,0;fwd:16006,10,1;account:100001_b")
	if len(got) != 3 {
		t.Fatalf("decoded %d members, want 3", len(got))
	}
	if got[1].Route.ValueString() != "fwd:16006" || got[1].RingTime.ValueInt64() != 10 || !got[1].Press1.ValueBool() {
		t.Errorf("member[1] = %+v", got[1])
	}
	if got[2].RingTime.ValueInt64() != 20 || got[2].Press1.ValueBool() {
		t.Errorf("member[2] should default to 20s / no press1, got %+v", got[2])
	}
	if len(decodeRingGroupMembers("")) != 0 {
		t.Error("empty members string should decode to no members")
	}
}

func TestRingGroupMembersRoundTrip(t *testing.T) {
	const raw = "account:100001_gateway,20,0;fwd:16006,25,1"
	encoded, err := encodeRingGroupMembers(decodeRingGroupMembers(raw))
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if encoded != raw {
		t.Errorf("round trip = %q, want %q", encoded, raw)
	}
}
