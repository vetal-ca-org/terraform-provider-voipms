package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vetal-ca-org/terraform-provider-voipms/internal/client"
)

func testDIDJSON(webhookEnabled string) map[string]any {
	return map[string]any{
		"did":                    "5550001001",
		"routing":                "account:100001_gateway",
		"failover_busy":          "sys:hangup",
		"failover_unreachable":   "sys:hangup",
		"failover_noanswer":      "sys:hangup",
		"voicemail":              "0",
		"pop":                    "73",
		"dialtime":               "60",
		"cnam":                   "0",
		"e911":                   "0",
		"callerid_prefix":        "",
		"record_calls":           "0",
		"note":                   "",
		"billing_type":           "1",
		"voicemail_threshold":    "0",
		"sms_available":          "1",
		"sms_enabled":            "1",
		"email_enabled":          "0",
		"email_address":          "",
		"sms_forward_enable":     "0",
		"sms_forward":            "",
		"url_callback_enable":    "0",
		"url_callback":           "",
		"url_callback_retry":     "0",
		"sms_sipaccount":         "",
		"sms_sipaccount_enabled": "0",
		"webhook":                "https://example.com/hook",
		"webhook_enabled":        webhookEnabled,
	}
}

func TestApplyDIDSkipsUnchangedInfo(t *testing.T) {
	webhookReadDelay = 0
	var mu sync.Mutex
	methods := map[string]int{}
	webhook := "0"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := r.URL.Query().Get("method")
		mu.Lock()
		methods[method]++
		if method == "setSMS" {
			webhook = r.URL.Query().Get("webhook_enable")
			if webhook == "" {
				t.Error("setSMS missing webhook_enable")
				webhook = "0"
			}
		}
		enabled := webhook
		mu.Unlock()

		switch method {
		case "getForwardings":
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "no_forwarding"})
		case "getVoicemails":
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "no_voicemail"})
		case "getRingGroups":
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "no_ring_group"})
		case "getTimeConditions":
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "no_timecondition"})
		case "getDIDsInfo":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"dids":   []map[string]any{testDIDJSON(enabled)},
			})
		case "setDIDInfo", "setSMS":
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
		default:
			t.Errorf("unexpected method %s", method)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
		}
	}))
	t.Cleanup(server.Close)

	c := client.New(client.Config{
		BaseURL:    server.URL,
		Username:   "user@example.com",
		Password:   "secret",
		HTTPClient: server.Client(),
	})
	c.UseInstantSMSWait()

	r := &didResource{client: c}
	plan := didModel{
		DID:            types.StringValue("5550001001"),
		Webhook:        types.StringValue("https://example.com/hook"),
		WebhookEnabled: types.BoolValue(true),
	}
	if err := r.applyDID(context.Background(), &plan); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if methods["setDIDInfo"] != 0 {
		t.Fatalf("setDIDInfo calls = %d, want 0", methods["setDIDInfo"])
	}
	if methods["setSMS"] < 1 {
		t.Fatalf("setSMS calls = %d, want at least 1", methods["setSMS"])
	}
}

func TestApplyDIDWritesSMSAfterInfoChange(t *testing.T) {
	webhookReadDelay = 0
	var mu sync.Mutex
	methods := map[string]int{}
	webhook := "1"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := r.URL.Query().Get("method")
		mu.Lock()
		methods[method]++
		if method == "setDIDInfo" {
			webhook = "0"
		}
		if method == "setSMS" {
			if r.URL.Query().Get("webhook_enable") != "1" {
				t.Error("setSMS missing webhook_enable=1")
			}
			webhook = "1"
		}
		enabled := webhook
		mu.Unlock()

		switch method {
		case "getForwardings":
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "no_forwarding"})
		case "getVoicemails":
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "no_voicemail"})
		case "getRingGroups":
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "no_ring_group"})
		case "getTimeConditions":
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "no_timecondition"})
		case "getDIDsInfo":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"dids":   []map[string]any{testDIDJSON(enabled)},
			})
		case "setDIDInfo", "setSMS":
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
		default:
			t.Errorf("unexpected method %s", method)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
		}
	}))
	t.Cleanup(server.Close)

	c := client.New(client.Config{
		BaseURL:    server.URL,
		Username:   "user@example.com",
		Password:   "secret",
		HTTPClient: server.Client(),
	})
	c.UseInstantSMSWait()

	r := &didResource{client: c}
	plan := didModel{
		DID:            types.StringValue("5550001001"),
		Routing:        types.StringValue("sys:hangup"),
		Webhook:        types.StringValue("https://example.com/hook"),
		WebhookEnabled: types.BoolValue(true),
	}
	if err := r.applyDID(context.Background(), &plan); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if methods["setDIDInfo"] != 1 {
		t.Fatalf("setDIDInfo calls = %d, want 1", methods["setDIDInfo"])
	}
	if methods["setSMS"] < 1 {
		t.Fatalf("setSMS calls = %d, want at least 1", methods["setSMS"])
	}
}
