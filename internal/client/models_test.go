package client

import (
	"encoding/json"
	"testing"
)

// setVoicemail and setSubAccount replace the whole record. Any writable field
// getVoicemails / getSubAccounts hands back has to survive the round trip, or
// updating one attribute silently resets the rest.

func TestVoicemailSetParamsRoundTripsTranscription(t *testing.T) {
	// Verbatim shape of one mailbox from getVoicemails.
	raw := `{
		"mailbox": "101", "name": "555-123-4567", "password": "9999",
		"skip_password": "1", "email": "", "client": "0",
		"attach_message": "yes", "delete_message": "no", "say_time": "yes",
		"timezone": "America/Chicago", "say_callerid": "yes",
		"play_instructions": "u", "language": "en",
		"email_attachment_format": "wavmp3", "unavailable_message_recording": "7567",
		"transcription": "Y", "transcription_locale": "en-US",
		"transcription_sentiment": "N", "transcription_summary": "N",
		"transcription_format": "text", "transcription_redaction": "N"
	}`
	var box Voicemail
	if err := json.Unmarshal([]byte(raw), &box); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	params := box.SetParams()
	want := map[string]string{
		"transcription":           "yes",
		"transcription_locale":    "en-US",
		"transcription_sentiment": "no",
		"transcription_summary":   "no",
		"transcription_format":    "text",
		"transcription_redaction": "no",
		"skip_password":           "yes",
		"attach_message":          "yes",
		"delete_message":          "no",
	}
	for key, expected := range want {
		if got := params[key]; got != expected {
			t.Errorf("params[%q] = %q, want %q", key, got, expected)
		}
	}
	// Reseller-only: client=0 means no reseller and is rejected if sent.
	if _, ok := params["client"]; ok {
		t.Errorf("params carried client=%q for a non-reseller mailbox", params["client"])
	}
}

func TestSubAccountSetParamsRoundTripsUnmodelledFields(t *testing.T) {
	raw := `{
		"id": "2001", "account": "100001_gateway", "username": "gateway",
		"parking_lot": "323", "tfcarrier": "-1",
		"internal_extension_location": "7", "transcription_start_delay": "10",
		"internal_voicemail": "101"
	}`
	var acct SubAccount
	if err := json.Unmarshal([]byte(raw), &acct); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	params := acct.SetParams()
	want := map[string]string{
		"parking_lot":                 "323",
		"tfcarrier":                   "-1",
		"internal_extension_location": "7",
		"transcription_start_delay":   "10",
		"internal_voicemail":          "101",
	}
	for key, expected := range want {
		if got := params[key]; got != expected {
			t.Errorf("params[%q] = %q, want %q", key, got, expected)
		}
	}
}

func TestCanonicalRouteResolvesRingGroupByName(t *testing.T) {
	tables := RouteTables{RingGroups: []RingGroup{
		{RingGroup: "900001", Name: "555-123-4567"},
		{RingGroup: "900002", Name: "555-123-4568"},
	}}
	for _, tc := range []struct{ in, want string }{
		{"grp:555-123-4567", "grp:900001"},
		{"grp:900002", "grp:900002"},
		{"account:100001_gateway", "account:100001_gateway"},
		{"sys:hangup", "sys:hangup"},
	} {
		got, err := CanonicalRoute(tc.in, tables)
		if err != nil {
			t.Fatalf("CanonicalRoute(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Errorf("CanonicalRoute(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
	if _, err := CanonicalRoute("grp:nope", tables); err == nil {
		t.Error("CanonicalRoute(grp:nope) should fail on an unknown ring group")
	}
}

// port_out_pin is the credential that authorises porting a number away. It is
// not in the Terraform schema, so it only survives an update by being echoed
// back through SetInfoParams.
func TestDIDSetInfoParamsRoundTripsPortOutPIN(t *testing.T) {
	raw := `{
		"did": "5551234567", "routing": "grp:900001", "voicemail": "1001",
		"port_out_pin": "9999", "inbound_dialing_mode": "0",
		"transcribe": "0", "transcription_locale": "en-US",
		"transcription_redaction": "N", "transcription_sentiment": "N",
		"transcription_summary": "N", "transcription_start_delay": "0",
		"transcription_email": ""
	}`
	var did DID
	if err := json.Unmarshal([]byte(raw), &did); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	params := did.SetInfoParams()
	want := map[string]string{
		"port_out_pin":              "9999",
		"inbound_dialing_mode":      "0",
		"transcribe":                "0",
		"transcription_locale":      "en-US",
		"transcription_redaction":   "no",
		"transcription_sentiment":   "no",
		"transcription_summary":     "no",
		"transcription_start_delay": "0",
	}
	for key, expected := range want {
		if got := params[key]; got != expected {
			t.Errorf("params[%q] = %q, want %q", key, got, expected)
		}
	}
}
