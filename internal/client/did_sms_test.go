package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestSMSWait(t *testing.T) {
	if !SMSWait(&APIError{Method: "setSMS", Status: "sms_wait_message"}) {
		t.Fatal("sms_wait_message")
	}
	if SMSWait(&APIError{Method: "setSMS", Status: "invalid_account"}) {
		t.Fatal("other status")
	}
}

func TestSetSMSRetriesWait(t *testing.T) {
	calls := 0
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("method") != "setSMS" {
			t.Fatalf("method = %s", r.URL.Query().Get("method"))
		}
		calls++
		status := "sms_wait_message"
		if calls >= 2 {
			status = "success"
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": status})
	})
	c.UseInstantSMSWait()

	if err := c.SetSMS(context.Background(), map[string]string{"did": "5550001001", "webhook_enabled": "1"}); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}
