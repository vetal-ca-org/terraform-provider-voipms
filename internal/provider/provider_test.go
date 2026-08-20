package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"voipms": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	t.Helper()

	if os.Getenv("VOIPMS_USERNAME") == "" || os.Getenv("VOIPMS_PASSWORD") == "" {
		t.Skip("VOIPMS_USERNAME and VOIPMS_PASSWORD must be set for acceptance tests")
	}
}

func TestProviderRegistersInventory(t *testing.T) {
	t.Parallel()

	p := New("test")()
	if got := len(p.Resources(context.Background())); got != 8 {
		t.Errorf("Resources() count = %d, want 8", got)
	}
	if got := len(p.DataSources(context.Background())); got != 19 {
		t.Errorf("DataSources() count = %d, want 19", got)
	}
}

func TestFirstNonEmpty(t *testing.T) {
	t.Parallel()

	if got := firstNonEmpty("", "voip_ms_user", "other"); got != "voip_ms_user" {
		t.Errorf("got %q", got)
	}
	if got := firstNonEmpty(); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}
