package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return New(Config{
		BaseURL:    server.URL,
		Username:   "user@example.com",
		Password:   "secret",
		HTTPClient: server.Client(),
	})
}

func TestGetBalanceSuccess(t *testing.T) {
	t.Parallel()

	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("method") != "getBalance" {
			t.Errorf("method = %q, want getBalance", r.URL.Query().Get("method"))
		}
		if r.URL.Query().Get("api_username") != "user@example.com" {
			t.Errorf("api_username = %q", r.URL.Query().Get("api_username"))
		}
		if r.URL.Query().Get("api_password") != "secret" {
			t.Errorf("api_password mismatch")
		}
		if r.URL.Query().Get("advanced") != "true" {
			t.Errorf("advanced = %q, want true", r.URL.Query().Get("advanced"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"balance": map[string]string{
				"current_balance": "12.34",
				"spent_today":     "1.00",
			},
		})
	})

	got, err := c.GetBalance(context.Background())
	if err != nil {
		t.Fatalf("GetBalance() error = %v", err)
	}
	if got.CurrentBalance != "12.34" {
		t.Errorf("CurrentBalance = %q, want 12.34", got.CurrentBalance)
	}
	if got.SpentToday != "1.00" {
		t.Errorf("SpentToday = %q, want 1.00", got.SpentToday)
	}
}

func TestCallAPIError(t *testing.T) {
	t.Parallel()

	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "invalid_credentials",
			"message": "bad password",
		})
	})

	err := c.Call(context.Background(), "getBalance", nil, nil)
	if err == nil {
		t.Fatal("expected API error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.Status != "invalid_credentials" {
		t.Errorf("Status = %q, want invalid_credentials", apiErr.Status)
	}
	if apiErr.Message != "bad password" {
		t.Errorf("Message = %q, want bad password", apiErr.Message)
	}
	if apiErr.EmptyResult() {
		t.Error("invalid_credentials should not be EmptyResult")
	}
}

func TestCallHTTPError(t *testing.T) {
	t.Parallel()

	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusForbidden)
	})

	err := c.Call(context.Background(), "getBalance", nil, nil)
	if err == nil {
		t.Fatal("expected HTTP error, got nil")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("error type = %T, want *HTTPError", err)
	}
	if httpErr.Status != http.StatusForbidden {
		t.Errorf("Status = %d, want %d", httpErr.Status, http.StatusForbidden)
	}
	if httpErr.Retryable() {
		t.Error("403 should not be retryable")
	}
}

func TestCallRetriesCloudflare522(t *testing.T) {
	t.Parallel()

	var hits int
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(statusCloudflareTimeout)
			_, _ = w.Write([]byte("error code: 522"))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})
	c.retryBackoff = func(context.Context, time.Duration) error { return nil }

	if err := c.Call(context.Background(), "getBalance", nil, nil); err != nil {
		t.Fatalf("Call() after retries: %v", err)
	}
	if hits != 3 {
		t.Errorf("hits = %d, want 3", hits)
	}
}

func TestCallGivesUpOnPersistent522(t *testing.T) {
	t.Parallel()

	var hits int
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(statusCloudflareTimeout)
		_, _ = w.Write([]byte("error code: 522"))
	})
	c.maxAttempts = 4
	c.retryBackoff = func(context.Context, time.Duration) error { return nil }

	err := c.Call(context.Background(), "getServersInfo", nil, nil)
	if err == nil {
		t.Fatal("expected HTTP error after retries")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) || httpErr.Status != statusCloudflareTimeout {
		t.Fatalf("err = %v, want HTTP 522", err)
	}
	if hits != 4 {
		t.Errorf("hits = %d, want 4", hits)
	}
}

func TestCallOmitsEmptyParams(t *testing.T) {
	t.Parallel()

	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Has("note") {
			t.Errorf("empty note should be omitted, got %q", r.URL.Query().Get("note"))
		}
		if r.URL.Query().Get("did") != "5550001001" {
			t.Errorf("did = %q", r.URL.Query().Get("did"))
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})

	err := c.Call(context.Background(), "setDIDInfo", map[string]string{
		"did":  "5550001001",
		"note": "",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCallWriteSendsEmptyParams(t *testing.T) {
	t.Parallel()

	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if !r.URL.Query().Has("note") {
			t.Error("CallWrite should send empty note")
		}
		if r.URL.Query().Get("note") != "" {
			t.Errorf("note = %q, want empty", r.URL.Query().Get("note"))
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})

	err := c.CallWrite(context.Background(), "setDIDInfo", map[string]string{
		"did":  "5550001001",
		"note": "",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewDefaults(t *testing.T) {
	t.Parallel()

	c := New(Config{Username: "a", Password: "b"})
	if c.baseURL != DefaultBaseURL {
		t.Errorf("baseURL = %q, want %s", c.baseURL, DefaultBaseURL)
	}
	if c.httpClient == nil {
		t.Fatal("httpClient is nil")
	}
	if c.userAgent == "" {
		t.Fatal("userAgent is empty")
	}
}

func TestFlexStringUnmarshal(t *testing.T) {
	t.Parallel()

	var got struct {
		A FlexString `json:"a"`
		B FlexString `json:"b"`
		C FlexString `json:"c"`
		D FlexString `json:"d"`
	}
	if err := json.Unmarshal([]byte(`{"a":"1","b":0,"c":true,"d":null}`), &got); err != nil {
		t.Fatal(err)
	}
	if got.A.String() != "1" || !got.A.Bool() {
		t.Errorf("A = %q bool=%v", got.A, got.A.Bool())
	}
	if got.B.String() != "0" || got.B.Bool() {
		t.Errorf("B = %q bool=%v", got.B, got.B.Bool())
	}
	if got.C.String() != "1" || !got.C.Bool() {
		t.Errorf("C = %q bool=%v", got.C, got.C.Bool())
	}
	if got.D.String() != "" {
		t.Errorf("D = %q, want empty", got.D)
	}

	yes := FlexString("yes")
	if !yes.Bool() {
		t.Error("yes should be true")
	}
	n, ok := FlexString("73").Int64()
	if !ok || n != 73 {
		t.Errorf("Int64 = %d, %v", n, ok)
	}
}

func TestGetSubAccountsEmpty(t *testing.T) {
	t.Parallel()

	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "no_subaccount"})
	})
	got, err := c.GetSubAccounts(context.Background(), "")
	if err != nil {
		t.Fatalf("GetSubAccounts() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

func TestGetSubAccountByLogin(t *testing.T) {
	t.Parallel()

	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("method") != "getSubAccounts" {
			t.Errorf("method = %q", r.URL.Query().Get("method"))
		}
		if r.URL.Query().Get("account") != "100001_gateway" {
			t.Errorf("account = %q", r.URL.Query().Get("account"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"accounts": []map[string]any{
				{"id": "2001", "account": "100001_gateway", "username": "gateway", "sip_traffic": 1, "nat": "no"},
			},
		})
	})

	got, err := c.GetSubAccount(context.Background(), "100001_gateway")
	if err != nil {
		t.Fatalf("GetSubAccount() error = %v", err)
	}
	if got.ID.String() != "2001" || got.Username.String() != "gateway" {
		t.Errorf("got id=%s username=%s", got.ID, got.Username)
	}
	if !got.SIPTraffic.Bool() {
		t.Error("sip_traffic should be true")
	}
}

func TestGetSubAccountByNumericID(t *testing.T) {
	t.Parallel()

	var calls int
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		account := r.URL.Query().Get("account")
		switch account {
		case "2001":
			// VoIP.ms rejects numeric id filters.
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "no_subaccount"})
		case "":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
				"accounts": []map[string]any{
					{"id": "2001", "account": "100001_gateway", "username": "gateway"},
					{"id": "2002", "account": "100001_other", "username": "other"},
				},
			})
		default:
			t.Errorf("unexpected account filter %q", account)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "no_subaccount"})
		}
	})

	got, err := c.GetSubAccount(context.Background(), "2001")
	if err != nil {
		t.Fatalf("GetSubAccount() error = %v", err)
	}
	if got.ID.String() != "2001" || got.Username.String() != "gateway" {
		t.Errorf("got id=%s username=%s", got.ID, got.Username)
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2 (filtered miss + full list)", calls)
	}
}

func TestGetSubAccountNotFound(t *testing.T) {
	t.Parallel()

	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "no_subaccount"})
	})
	_, err := c.GetSubAccount(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestGetDIDsInfoMixedTypes(t *testing.T) {
	t.Parallel()

	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("did") != "5550001001" {
			t.Errorf("did = %q", r.URL.Query().Get("did"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"dids": []map[string]any{
				{
					"did":             "5550001001",
					"routing":         "account:100001_gateway",
					"pop":             "73",
					"e911":            "1",
					"sms_available":   1,
					"sms_enabled":     "1",
					"webhook_enabled": "1",
					"webhook":         "https://example.com/hook",
				},
			},
		})
	})

	did, err := c.GetDID(context.Background(), "5550001001")
	if err != nil {
		t.Fatalf("GetDID() error = %v", err)
	}
	if did.POP.String() != "73" {
		t.Errorf("POP = %q", did.POP)
	}
	if !did.E911.Bool() || !did.SMSAvailable.Bool() || !did.WebhookEnabled.Bool() {
		t.Errorf("bool fields: e911=%v sms=%v webhook=%v", did.E911.Bool(), did.SMSAvailable.Bool(), did.WebhookEnabled.Bool())
	}
}

func TestSetDIDInfoAndSMS(t *testing.T) {
	t.Parallel()

	methods := map[string]int{}
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		methods[r.URL.Query().Get("method")]++
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})

	if err := c.SetDIDInfo(context.Background(), map[string]string{"did": "5550001001", "routing": "sys:hangup"}); err != nil {
		t.Fatal(err)
	}
	if err := c.SetSMS(context.Background(), map[string]string{"did": "5550001001", "enable": "1"}); err != nil {
		t.Fatal(err)
	}
	if methods["setDIDInfo"] != 1 || methods["setSMS"] != 1 {
		t.Errorf("methods = %#v", methods)
	}
}

func TestGetForwardingsEmpty(t *testing.T) {
	t.Parallel()

	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "no_forwarding"})
	})
	got, err := c.GetForwardings(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("len = %d", len(got))
	}
}

func TestGetVoicemail(t *testing.T) {
	t.Parallel()

	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("mailbox") != "101" {
			t.Errorf("mailbox = %q", r.URL.Query().Get("mailbox"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"voicemails": []map[string]any{
				{"mailbox": "101", "name": "Main", "skip_password": "1", "attach_message": "yes"},
			},
		})
	})
	got, err := c.GetVoicemail(context.Background(), "101")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name.String() != "Main" || !got.SkipPassword.Bool() || !got.AttachMessage.Bool() {
		t.Errorf("got %+v", got)
	}
}

func TestGetServersInfo(t *testing.T) {
	t.Parallel()

	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("method") != "getServersInfo" {
			t.Errorf("method = %q", r.URL.Query().Get("method"))
		}
		if r.URL.Query().Get("server_pop") != "73" {
			t.Errorf("server_pop = %q", r.URL.Query().Get("server_pop"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"servers": []map[string]any{
				{"server_pop": "73", "server_hostname": "newyork7.voip.ms", "server_name": "New York 7", "server_country": "US", "server_recommended": 1},
			},
		})
	})
	got, err := c.GetServer(context.Background(), "73")
	if err != nil {
		t.Fatal(err)
	}
	if got.Hostname.String() != "newyork7.voip.ms" {
		t.Errorf("hostname = %q", got.Hostname)
	}
	if !got.Recommended.Bool() {
		t.Error("recommended should be true")
	}
}

func TestMatchServer(t *testing.T) {
	t.Parallel()

	servers := []Server{
		{POP: "73", Hostname: "newyork7.voip.ms", Name: "New York 7", Country: "US", Shortname: "NY7"},
		{POP: "1", Hostname: "atlanta.voip.ms", Name: "Atlanta", Country: "US"},
	}
	cases := []string{"73", "newyork7.voip.ms", "New York 7", "NY7", "New York 7, US"}
	for _, q := range cases {
		got, err := MatchServer(servers, q)
		if err != nil {
			t.Errorf("MatchServer(%q) error = %v", q, err)
			continue
		}
		if got.POP.String() != "73" {
			t.Errorf("MatchServer(%q) pop = %s", q, got.POP)
		}
	}
	if _, err := MatchServer(servers, "missing"); !errors.Is(err, ErrNotFound) {
		t.Errorf("missing query error = %v", err)
	}
}

func TestMatchNamedObjects(t *testing.T) {
	t.Parallel()

	vms := []Voicemail{{Mailbox: "101", Name: "John"}}
	got, err := MatchVoicemail(vms, "John")
	if err != nil || got.Mailbox.String() != "101" {
		t.Fatalf("MatchVoicemail: %v %+v", err, got)
	}
	if NameForMailbox(vms, "101") != "John" {
		t.Errorf("NameForMailbox = %q", NameForMailbox(vms, "101"))
	}

	fwds := []Forwarding{{Forwarding: "1001", Description: "Mobile"}}
	fwd, err := MatchForwarding(fwds, "Mobile")
	if err != nil || fwd.Forwarding.String() != "1001" {
		t.Fatalf("MatchForwarding: %v %+v", err, fwd)
	}

	cbs := []Callback{{Callback: "3001", Description: "Mobile"}}
	cb, err := MatchCallback(cbs, "Mobile")
	if err != nil || cb.Callback.String() != "3001" {
		t.Fatalf("MatchCallback: %v %+v", err, cb)
	}

	grps := []PhonebookGroup{{PhonebookGroup: "5001", Name: "Spam"}}
	grp, err := MatchPhonebookGroup(grps, "Spam")
	if err != nil || grp.PhonebookGroup.String() != "5001" {
		t.Fatalf("MatchPhonebookGroup: %v %+v", err, grp)
	}
}

func TestFindServerByHostname(t *testing.T) {
	t.Parallel()

	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("server_pop") != "" {
			t.Errorf("FindServer should list all POPs, got server_pop=%q", r.URL.Query().Get("server_pop"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"servers": []map[string]any{
				{"server_pop": "1", "server_hostname": "atlanta.voip.ms", "server_name": "Atlanta"},
				{"server_pop": "73", "server_hostname": "newyork7.voip.ms", "server_name": "New York 7", "server_country": "US"},
			},
		})
	})
	got, err := c.FindServer(context.Background(), "newyork7.voip.ms")
	if err != nil {
		t.Fatal(err)
	}
	if got.POP.String() != "73" {
		t.Errorf("pop = %s", got.POP)
	}
	if HostnameForPOP([]Server{*got}, 73) != "newyork7.voip.ms" {
		t.Errorf("HostnameForPOP = %q", HostnameForPOP([]Server{*got}, 73))
	}
}

func TestGetCallerIDFilters(t *testing.T) {
	t.Parallel()

	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"filtering": []map[string]any{
				{"filtering": "4001", "callerid": "999XXXXXXXX", "did": "all", "routing": "sys:hangup"},
			},
		})
	})
	got, err := c.GetCallerIDFilter(context.Background(), "4001")
	if err != nil {
		t.Fatal(err)
	}
	if got.CallerID.String() != "999XXXXXXXX" {
		t.Errorf("callerid = %q", got.CallerID)
	}
}

func TestCreateAndDeleteSubAccount(t *testing.T) {
	t.Parallel()

	methods := map[string]int{}
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		methods[r.URL.Query().Get("method")]++
		if r.URL.Query().Get("method") == "delSubAccount" && r.URL.Query().Get("id") != "2001" {
			t.Errorf("id = %q", r.URL.Query().Get("id"))
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})
	if err := c.CreateSubAccount(context.Background(), map[string]string{"username": "gateway", "password": "x"}); err != nil {
		t.Fatal(err)
	}
	if err := c.DeleteSubAccount(context.Background(), "2001"); err != nil {
		t.Fatal(err)
	}
	if methods["createSubAccount"] != 1 || methods["delSubAccount"] != 1 {
		t.Errorf("methods = %#v", methods)
	}
}

func TestFindForwardingAfterCreate(t *testing.T) {
	t.Parallel()

	items := []Forwarding{
		{Forwarding: "1", PhoneNumber: "111", Description: "old"},
		{Forwarding: "1001", PhoneNumber: "5550002001", Description: "Mobile"},
	}
	got := FindForwardingAfterCreate(items, "5550002001", "Mobile")
	if got == nil || got.Forwarding.String() != "1001" {
		t.Fatalf("got %+v", got)
	}
}

func TestFindHelpers(t *testing.T) {
	t.Parallel()

	cb := FindCallbackAfterCreate([]Callback{
		{Callback: "1", Number: "111", Description: "other"},
		{Callback: "3001", Number: "15550002001", Description: "Mobile"},
	}, "15550002001", "Mobile")
	if cb == nil || cb.Callback.String() != "3001" {
		t.Fatalf("callback = %+v", cb)
	}

	grp := FindPhonebookGroupAfterCreate([]PhonebookGroup{
		{PhonebookGroup: "10", Name: "General"},
		{PhonebookGroup: "5001", Name: "Spam"},
	}, "Spam")
	if grp == nil || grp.PhonebookGroup.String() != "5001" {
		t.Fatalf("group = %+v", grp)
	}
}

func TestSlug(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"Kate Fizz Cell": "kate-fizz-cell",
		"Main":           "main",
		"  Kate  Fizz  ": "kate-fizz",
		"":               "",
	}
	for in, want := range cases {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCanadaRoute(t *testing.T) {
	t.Parallel()
	name, ok := CanadaRouteName("2")
	if !ok || name != CanadaRoutePremium {
		t.Errorf("CanadaRouteName(2) = %q %v", name, ok)
	}
	id, ok := CanadaRouteID("premium")
	if !ok || id != "2" {
		t.Errorf("CanadaRouteID(premium) = %q %v", id, ok)
	}
	if !CanadaRoutesEqual("2", "Premium") {
		t.Error("2 should equal Premium")
	}
	if CanadaRoutesEqual("1", "2") {
		t.Error("value should not equal premium")
	}
}

func TestNamedCodes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		code NamedCode
		id   string
		name string
	}{
		{DeviceType, "1", DeviceTypeIPPBX},
		{DeviceType, "2", DeviceTypeATA},
		{AuthType, "1", AuthTypePassword},
		{AuthType, "2", AuthTypeIP},
		{Protocol, "1", ProtocolSIP},
		{Protocol, "3", ProtocolIAX2},
		{LockInternational, "0", LockInternationalAllow},
		{LockInternational, "1", LockInternationalDeny},
		{DialingMode, "0", DialingModeMainAccount},
		{DialingMode, "1", DialingModeE164},
		{DialingMode, "2", DialingModeNANPA},
		{CallPickupBehavior, "1", CallPickupBoth},
		{CallPickupBehavior, "4", CallPickupDisabled},
		{BillingType, "1", BillingTypePerMinute},
		{BillingType, "2", BillingTypeFlat},
		{PlayInstructions, "u", PlayInstructionsUnread},
		{PlayInstructions, "su", PlayInstructionsSkipUnread},
	}
	for _, tc := range cases {
		name, ok := tc.code.Name(tc.id)
		if !ok || name != tc.name {
			t.Errorf("Name(%q) = %q %v, want %q", tc.id, name, ok, tc.name)
		}
		id, ok := tc.code.ID(tc.name)
		if !ok || id != tc.id {
			t.Errorf("ID(%q) = %q %v, want %q", tc.name, id, ok, tc.id)
		}
		if !tc.code.Equal(tc.id, strings.ToUpper(tc.name)) {
			t.Errorf("%q should equal %q", tc.id, tc.name)
		}
	}
	if _, ok := DeviceType.Name("9"); ok {
		t.Error("unknown device type should not map")
	}
}

func TestRouteFormat(t *testing.T) {
	t.Parallel()
	if got := AccountRoute("150060_gateway"); got != "account:150060_gateway" {
		t.Errorf("AccountRoute = %q", got)
	}
	if got := ForwardingRoute("186772"); got != "fwd:186772" {
		t.Errorf("ForwardingRoute = %q", got)
	}
	if got := VoicemailRoute("101"); got != "vm:101" {
		t.Errorf("VoicemailRoute = %q", got)
	}
}

func TestCanonicalRoute(t *testing.T) {
	t.Parallel()
	tables := RouteTables{
		Forwardings: []Forwarding{{Forwarding: "186772", Description: "Kate Fizz Cell"}},
		Voicemails:  []Voicemail{{Mailbox: "500601", Name: "Main"}},
	}
	cases := []struct {
		in, want string
	}{
		{"fwd:186772", "fwd:186772"},
		{"fwd:Kate Fizz Cell", "fwd:186772"},
		{"fwd:kate-fizz-cell", "fwd:186772"},
		{"vm:500601", "vm:500601"},
		{"vm:Main", "vm:500601"},
		{"vm:main", "vm:500601"},
		{"account:150060_common-fs", "account:150060_common-fs"},
		{"sys:hangup", "sys:hangup"},
		{"none:", "none:"},
	}
	for _, tc := range cases {
		got, err := CanonicalRoute(tc.in, tables)
		if err != nil {
			t.Errorf("CanonicalRoute(%q) error = %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("CanonicalRoute(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
	if !RoutesEqual("fwd:Kate Fizz Cell", "fwd:186772", tables) {
		t.Error("named forwarding should equal id form")
	}
	if _, err := CanonicalRoute("fwd:missing", tables); err == nil {
		t.Error("expected error for unknown forwarding")
	}
}
