package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-interlink/internal/portal"
)

type productsDataSource struct {
	client *portal.ClientWithResponses
}

type productsDataSourceModel struct {
	Products []productModel `tfsdk:"products"`
}

type productModel struct {
	Id   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Pid  types.String `tfsdk:"pid"`
}

func NewProductsDataSource() datasource.DataSource {
	return &productsDataSource{}
}

func (d *productsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_products"
}

func (d *productsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the Inter.link product catalog.",
		Attributes: map[string]schema.Attribute{
			"products": schema.ListNestedAttribute{
				Description: "All products available in the catalog.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "Internal numeric product ID.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Product name (e.g. `Cloud Connect`).",
							Computed:    true,
						},
						"pid": schema.StringAttribute{
							Description: "Product identifier (e.g. `PID25`).",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *productsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*portal.ClientWithResponses)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *portal.ClientWithResponses, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *productsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	apiResp, err := d.client.CommonServicesApiListProductsWithResponse(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Inter.link products", err.Error())
		return
	}
	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unexpected response reading Inter.link products",
			fmt.Sprintf("HTTP %s: %s", apiResp.Status(), string(apiResp.Body)),
		)
		return
	}

	var state productsDataSourceModel
	for _, p := range *apiResp.JSON200 {
		state.Products = append(state.Products, productModel{
			Id:   types.Int64Value(int64(p.Id)),
			Name: types.StringValue(p.Name),
			Pid:  types.StringValue(p.Pid),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
