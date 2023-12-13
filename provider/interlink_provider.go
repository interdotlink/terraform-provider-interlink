package provider_interlink

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	user "terraform-provider-interlink/user"
	api "github.com/interdotlink/api-client-go"
)

var _ provider.Provider = (*interlinkProvider)(nil)

func New() func() provider.Provider {
	return func() provider.Provider {
		return &interlinkProvider{}
	}
}

type interlinkProvider struct{}

func (p *interlinkProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = InterlinkProviderSchema(ctx)
}

func (p *interlinkProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var providerConfig InterlinkModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &providerConfig)...)

	if resp.Diagnostics.HasError() {
		return
	}

	conf := api.NewConfiguration()
	if !providerConfig.ApiUrl.IsNull() {
		conf.Servers = api.ServerConfigurations{ { URL: providerConfig.ApiUrl.ValueString() } }
	}
	conf.AddDefaultHeader("X-API-Key", providerConfig.ApiKey.ValueString())
	//conf.Debug = true
	client := api.NewAPIClient(conf)

	resp.ResourceData = client
}

func (p *interlinkProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "interlink"
}

func (p *interlinkProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *interlinkProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		user.NewUserResource,
	}
}
