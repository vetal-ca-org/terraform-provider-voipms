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
		{name: "account", v: types.StringValue("100001"), want: false},
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
	keepSMSAccountNone(&got, types.StringValue("100001"))
	if got.ValueString() != "" {
		t.Fatalf("must not keep a real account over API empty: got %q", got.ValueString())
	}

	got = types.StringValue("100001_gateway")
	keepSMSAccountNone(&got, types.StringValue("0"))
	if got.ValueString() != "100001_gateway" {
		t.Fatalf("must not overwrite a real account: got %q", got.ValueString())
	}
}

func TestKeepWebhookEnabled(t *testing.T) {
	got := types.BoolValue(false)
	keepWebhookEnabled(&got, types.BoolValue(true))
	if !got.ValueBool() {
		t.Fatal("keep planned true over API false")
	}

	got = types.BoolValue(true)
	keepWebhookEnabled(&got, types.BoolValue(true))
	if !got.ValueBool() {
		t.Fatal("keep true when API is true")
	}

	got = types.BoolValue(false)
	keepWebhookEnabled(&got, types.BoolValue(false))
	if got.ValueBool() {
		t.Fatal("must not invent true")
	}

	got = types.BoolValue(true)
	keepWebhookEnabled(&got, types.BoolValue(false))
	if !got.ValueBool() {
		t.Fatal("must not hide API true")
	}
}

func TestDidSMSParamsNoneAccount(t *testing.T) {
	for _, none := range []types.String{types.StringValue("0"), types.StringValue(""), types.StringValue(" 0 "), types.StringNull()} {
		got := didSMSParams(didModel{SMSSIPAccount: none})
		if got["sms_sipaccount"] != "" {
			t.Fatalf("none %q: sms_sipaccount=%q, want empty", none, got["sms_sipaccount"])
		}
	}
	got := didSMSParams(didModel{SMSSIPAccount: types.StringValue("100001")})
	if got["sms_sipaccount"] != "100001" {
		t.Fatalf("account: sms_sipaccount=%q, want 100001", got["sms_sipaccount"])
	}
}
