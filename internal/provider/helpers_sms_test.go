package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSMSAccountNone(t *testing.T) {
	cases := []struct {
		name string
		v    types.String
		want bool
	}{
		{name: "empty", v: types.StringValue(""), want: true},
		{name: "zero", v: types.StringValue("0"), want: true},
		{name: "spaces", v: types.StringValue(" 0 "), want: true},
		{name: "null", v: types.StringNull(), want: true},
		{name: "account", v: types.StringValue("150060"), want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := smsAccountNone(tc.v); got != tc.want {
				t.Fatalf("smsAccountNone(%q) = %v, want %v", tc.v, got, tc.want)
			}
		})
	}
}

func TestKeepSMSAccountNone(t *testing.T) {
	got := types.StringValue("")
	keepSMSAccountNone(&got, types.StringValue("0"))
	if got.ValueString() != "0" {
		t.Fatalf("keep after API empty: got %q, want 0", got.ValueString())
	}

	got = types.StringValue("0")
	keepSMSAccountNone(&got, types.StringValue(""))
	if got.ValueString() != "" {
		t.Fatalf("keep after config empty: got %q, want empty", got.ValueString())
	}

	got = types.StringValue("")
	keepSMSAccountNone(&got, types.StringValue("150060"))
	if got.ValueString() != "" {
		t.Fatalf("must not keep a real account over API empty: got %q", got.ValueString())
	}

	got = types.StringValue("150060_common-fs")
	keepSMSAccountNone(&got, types.StringValue("0"))
	if got.ValueString() != "150060_common-fs" {
		t.Fatalf("must not overwrite a real account: got %q", got.ValueString())
	}
}
