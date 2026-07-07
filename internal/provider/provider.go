package provider

import (
	"context"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-interlink/internal/portal"
)

type interlinkProvider struct{}

type interlinkProviderModel struct {
	ApiKey types.String `tfsdk:"api_key"`
	ApiUrl types.String `tfsdk:"api_url"`
}

func New() provider.Provider {
	return &interlinkProvider{}
}

func (p *interlinkProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "interlink"
}

func (p *interlinkProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with the Inter.link Portal API.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "API key for the Inter.link Portal, sent as the `X-API-Key` header. Create one under API Keys in the portal. May also be set via the `INTERLINK_API_KEY` environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"api_url": schema.StringAttribute{
				Description: "Base URL of the Inter.link Portal API. Defaults to `https://portal.inter.link`. May also be set via the `INTERLINK_API_URL` environment variable.",
				Optional:    true,
			},
		},
	}
}

func (p *interlinkProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config interlinkProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.ApiKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown Inter.link API key",
			"api_key is not yet known at this point of the plan. "+
				"Use a static value, a variable, or a value known before apply.",
		)
	}
	if config.ApiUrl.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_url"),
			"Unknown Inter.link API URL",
			"api_url is not yet known at this point of the plan. "+
				"Use a static value, a variable, or a value known before apply.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Precedence for both settings: config attribute, then environment
	// variable, then (for the URL) the default.
	baseURL := os.Getenv("INTERLINK_API_URL")
	if !config.ApiUrl.IsNull() {
		baseURL = config.ApiUrl.ValueString()
	}
	if baseURL == "" {
		baseURL = "https://portal.inter.link"
	}

	apiKey := config.ApiKey.ValueString()
	if apiKey == "" {
		apiKey = os.Getenv("INTERLINK_API_KEY")
	}
	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Inter.link API key",
			"Set the api_key attribute or the INTERLINK_API_KEY environment variable. "+
				"Create an API key under API Keys in the Inter.link portal.",
		)
		return
	}
	withAPIKey := portal.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("X-API-Key", apiKey)
		return nil
	})

	client, err := portal.NewClientWithResponses(baseURL, withAPIKey)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Inter.link API client", err.Error())
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *interlinkProvider) Resources(ctx context.Context) []func() resource.Resource {
	return nil
}

func (p *interlinkProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewStatusDataSource,
		NewProductsDataSource,
		NewLocationsDataSource,
		NewProjectsDataSource,
		NewServicesDataSource,
		NewIpServicesDataSource,
		NewDdosServicesDataSource,
		NewPortServicesDataSource,
		NewElanServicesDataSource,
		NewFlexpeerServicesDataSource,
	}
}
