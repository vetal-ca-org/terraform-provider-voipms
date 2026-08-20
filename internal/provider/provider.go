package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/vetal-ca-org/terraform-provider-voipms/internal/client"
)

var _ provider.Provider = &voipmsProvider{}

type voipmsProvider struct {
	version string
}

type voipmsProviderModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	APIURL   types.String `tfsdk:"api_url"`
}

func (p *voipmsProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "voipms"
	resp.Version = p.version
}

func (p *voipmsProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage [VoIP.ms](https://voip.ms) accounts through the REST/JSON API. " +
			"Authenticate with an API username (account email) and a dedicated API password from the " +
			"SOAP & REST/JSON API page. The public IP of the machine running Terraform must also be allow-listed there.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				MarkdownDescription: "VoIP.ms API username (account email). May also be set via `VOIPMS_USERNAME` or `voip_ms_username`.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "VoIP.ms API password (not the portal password). May also be set via `VOIPMS_PASSWORD` or `voip_ms_api_key`.",
				Optional:            true,
				Sensitive:           true,
			},
			"api_url": schema.StringAttribute{
				MarkdownDescription: "Override the REST endpoint. Defaults to `https://voip.ms/api/v1/rest.php`. May also be set via `VOIPMS_API_URL`.",
				Optional:            true,
			},
		},
	}
}

func (p *voipmsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config voipmsProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Username.IsUnknown() || config.Password.IsUnknown() || config.APIURL.IsUnknown() {
		return
	}

	username := firstNonEmpty(os.Getenv("VOIPMS_USERNAME"), os.Getenv("voip_ms_username"))
	password := firstNonEmpty(os.Getenv("VOIPMS_PASSWORD"), os.Getenv("voip_ms_api_key"))
	apiURL := os.Getenv("VOIPMS_API_URL")

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}
	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}
	if !config.APIURL.IsNull() {
		apiURL = config.APIURL.ValueString()
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing VoIP.ms API username",
			"Set the username provider argument or the VOIPMS_USERNAME environment variable.",
		)
	}
	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing VoIP.ms API password",
			"Set the password provider argument or the VOIPMS_PASSWORD environment variable. "+
				"Generate an API password in the VoIP.ms portal under SOAP & REST/JSON API; do not use the portal login password.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "password", "api_password")
	tflog.Info(ctx, "configured VoIP.ms client", map[string]any{
		"api_url":  firstNonEmpty(apiURL, client.DefaultBaseURL),
		"username": username,
	})

	c := client.New(client.Config{
		BaseURL:   apiURL,
		Username:  username,
		Password:  password,
		UserAgent: "terraform-provider-voipms/" + p.version,
	})

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *voipmsProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSubaccountResource,
		NewDIDResource,
		NewForwardingResource,
		NewVoicemailResource,
		NewCallbackResource,
		NewCallerIDFilterResource,
		NewPhonebookEntryResource,
		NewPhonebookGroupResource,
	}
}

func (p *voipmsProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewBalanceDataSource,
		NewSubaccountDataSource,
		NewSubaccountsDataSource,
		NewDIDDataSource,
		NewDIDsDataSource,
		NewForwardingDataSource,
		NewForwardingsDataSource,
		NewVoicemailDataSource,
		NewVoicemailsDataSource,
		NewCallbackDataSource,
		NewCallbacksDataSource,
		NewCallerIDFilterDataSource,
		NewCallerIDFiltersDataSource,
		NewPhonebookEntryDataSource,
		NewPhonebookEntriesDataSource,
		NewPhonebookGroupDataSource,
		NewPhonebookGroupsDataSource,
		NewServerDataSource,
		NewServersDataSource,
	}
}

// New is the provider factory used by main and by acceptance tests.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &voipmsProvider{version: version}
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
