package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
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
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"api_url": schema.StringAttribute{
				Optional: true,
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

	baseURL := "https://portal.inter.link"
	if !config.ApiUrl.IsNull() {
		baseURL = config.ApiUrl.ValueString()
	}

	apiKey := config.ApiKey.ValueString()
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
	return nil
}
