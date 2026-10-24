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
		{name: "subaccount", v: types.StringValue("100001_gateway"), want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := smsAccountNone(tc.v); got != tc.want {
				t.Fatalf("smsAccountNone(%q) = %v, want %v", tc.v, got, tc.want)
			}
		})
	}
}

func TestLeftoverSMSAccountNone(t *testing.T) {
	cases := []struct {
		name string
		v    types.String
		want bool
	}{
		{name: "bare account", v: types.StringValue("100001"), want: true},
		{name: "spaces", v: types.StringValue(" 100001 "), want: true},
		{name: "empty", v: types.StringValue(""), want: false},
		{name: "zero", v: types.StringValue("0"), want: false},
		{name: "null", v: types.StringNull(), want: false},
		{name: "subaccount", v: types.StringValue("100001_gateway"), want: false},
		{name: "name", v: types.StringValue("main"), want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := leftoverSMSAccountNone(tc.v); got != tc.want {
				t.Fatalf("leftoverSMSAccountNone(%q) = %v, want %v", tc.v, got, tc.want)
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

	got = types.StringValue("100001")
	keepSMSAccountNone(&got, types.StringValue(""))
	if got.ValueString() != "" {
		t.Fatalf("keep after API leftover main account: got %q, want empty", got.ValueString())
	}

	got = types.StringValue("100001")
	keepSMSAccountNone(&got, types.StringValue("0"))
	if got.ValueString() != "0" {
		t.Fatalf("keep after API leftover when config is 0: got %q, want 0", got.ValueString())
	}

	got = types.StringValue("100001")
	keepSMSAccountNone(&got, types.StringValue("100001"))
	if got.ValueString() != "100001" {
		t.Fatalf("explicit main account must round-trip: got %q", got.ValueString())
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

	got = types.StringValue("100001_gateway")
	keepSMSAccountNone(&got, types.StringValue(""))
	if got.ValueString() != "100001_gateway" {
		t.Fatalf("must not treat a subaccount as leftover none: got %q", got.ValueString())
	}
}

func TestParamsEqual(t *testing.T) {
	a := map[string]string{"did": "5550001001", "webhook_enabled": "1"}
	b := map[string]string{"did": "5550001001", "webhook_enabled": "1"}
	if !paramsEqual(a, b) {
		t.Fatal("equal maps")
	}
	b["webhook_enabled"] = "0"
	if paramsEqual(a, b) {
		t.Fatal("different value")
	}
	if paramsEqual(a, map[string]string{"did": "5550001001"}) {
		t.Fatal("different length")
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

func TestDidSMSParamsWebhookEnable(t *testing.T) {
	got := didSMSParams(didModel{WebhookEnabled: types.BoolValue(true)})
	if got["webhook_enable"] != "1" {
		t.Fatalf("webhook_enable=%q, want 1", got["webhook_enable"])
	}
	if _, ok := got["webhook_enabled"]; ok {
		t.Fatal("setSMS uses webhook_enable, not webhook_enabled")
	}
}
