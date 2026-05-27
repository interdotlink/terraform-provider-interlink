package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type interlinkProvider struct{}

func New() provider.Provider {
	return &interlinkProvider{}
}

func (p *interlinkProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "interlink"
}

func (p *interlinkProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{}
}

func (p *interlinkProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
}

func (p *interlinkProvider) Resources(ctx context.Context) []func() resource.Resource {
	return nil
}

func (p *interlinkProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}
