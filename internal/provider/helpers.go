package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vetal-ca-org/terraform-provider-voipms/internal/client"
)

func configureClient(providerData any, diags *diag.Diagnostics) *client.Client {
	if providerData == nil {
		return nil
	}
	c, ok := providerData.(*client.Client)
	if !ok {
		diags.AddError(
			"Unexpected configure type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", providerData),
		)
		return nil
	}
	return c
}

func strVal(v client.FlexString) types.String {
	return types.StringValue(v.String())
}

func boolVal(v client.FlexString) types.Bool {
	return types.BoolValue(v.Bool())
}

func intVal(v client.FlexString) types.Int64 {
	n, ok := v.Int64()
	if !ok {
		return types.Int64Value(0)
	}
	return types.Int64Value(n)
}

func setString(params map[string]string, key string, v types.String) {
	if v.IsNull() || v.IsUnknown() {
		return
	}
	params[key] = v.ValueString()
}

func setBool01(params map[string]string, key string, v types.Bool) {
	if v.IsNull() || v.IsUnknown() {
		return
	}
	if v.ValueBool() {
		params[key] = "1"
	} else {
		params[key] = "0"
	}
}

func setBoolYesNo(params map[string]string, key string, v types.Bool) {
	if v.IsNull() || v.IsUnknown() {
		return
	}
	if v.ValueBool() {
		params[key] = "yes"
	} else {
		params[key] = "no"
	}
}

func setInt(params map[string]string, key string, v types.Int64) {
	if v.IsNull() || v.IsUnknown() {
		return
	}
	params[key] = strconv.FormatInt(v.ValueInt64(), 10)
}

func overlayParams(base, updates map[string]string) map[string]string {
	out := make(map[string]string, len(base)+len(updates))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range updates {
		out[k] = v
	}
	return out
}

// computedOptional keeps omitted attributes unknown on create (so Read can fill
// them) and preserves state on later plans when the practitioner still omits them.
type computedOptionalString struct{}

func (m computedOptionalString) Description(_ context.Context) string {
	return "If unset, Terraform keeps the previous value (or the API default after create)."
}
func (m computedOptionalString) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
func (m computedOptionalString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.ConfigValue.IsNull() {
		return
	}
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		resp.PlanValue = types.StringUnknown()
		return
	}
	resp.PlanValue = req.StateValue
}

type computedOptionalBool struct{}

func (m computedOptionalBool) Description(_ context.Context) string {
	return "If unset, Terraform keeps the previous value (or the API default after create)."
}
func (m computedOptionalBool) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
func (m computedOptionalBool) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if !req.ConfigValue.IsNull() {
		return
	}
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		resp.PlanValue = types.BoolUnknown()
		return
	}
	resp.PlanValue = req.StateValue
}

type computedOptionalInt64 struct{}

func (m computedOptionalInt64) Description(_ context.Context) string {
	return "If unset, Terraform keeps the previous value (or the API default after create)."
}
func (m computedOptionalInt64) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
func (m computedOptionalInt64) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if !req.ConfigValue.IsNull() {
		return
	}
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		resp.PlanValue = types.Int64Unknown()
		return
	}
	resp.PlanValue = req.StateValue
}

func optString() planmodifier.String { return computedOptionalString{} }
func optBool() planmodifier.Bool     { return computedOptionalBool{} }
func optInt() planmodifier.Int64     { return computedOptionalInt64{} }
