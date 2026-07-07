package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-interlink/internal/portal"
)

type locationsDataSource struct {
	client *portal.ClientWithResponses
}

type locationsDataSourceModel struct {
	Locations []locationModel `tfsdk:"locations"`
}

type locationModel struct {
	Id          types.Int64   `tfsdk:"id"`
	Name        types.String  `tfsdk:"name"`
	Description types.String  `tfsdk:"description"`
	Line1       types.String  `tfsdk:"line1"`
	Line2       types.String  `tfsdk:"line2"`
	PostalCode  types.String  `tfsdk:"postal_code"`
	City        types.String  `tfsdk:"city"`
	Country     types.String  `tfsdk:"country"`
	Latitude    types.Float64 `tfsdk:"latitude"`
	Longitude   types.Float64 `tfsdk:"longitude"`
	Status      types.String  `tfsdk:"status"`
	Type        types.String  `tfsdk:"type"`
}

func NewLocationsDataSource() datasource.DataSource {
	return &locationsDataSource{}
}

func (d *locationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_locations"
}

func (d *locationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the Inter.link data-center locations (points of presence).",
		Attributes: map[string]schema.Attribute{
			"locations": schema.ListNestedAttribute{
				Description: "All locations.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.Int64Attribute{Description: "Internal numeric location ID.", Computed: true},
						"name":        schema.StringAttribute{Description: "Location name (e.g. `FRA1-DE`).", Computed: true},
						"description": schema.StringAttribute{Description: "Location description.", Computed: true},
						"line1":       schema.StringAttribute{Description: "Street address, line 1.", Computed: true},
						"line2":       schema.StringAttribute{Description: "Street address, line 2.", Computed: true},
						"postal_code": schema.StringAttribute{Description: "Postal code.", Computed: true},
						"city":        schema.StringAttribute{Description: "City name.", Computed: true},
						"country":     schema.StringAttribute{Description: "Country name.", Computed: true},
						"latitude":    schema.Float64Attribute{Description: "Latitude in decimal degrees.", Computed: true},
						"longitude":   schema.Float64Attribute{Description: "Longitude in decimal degrees.", Computed: true},
						"status":      schema.StringAttribute{Description: "Location status.", Computed: true},
						"type":        schema.StringAttribute{Description: "Location type.", Computed: true},
					},
				},
			},
		},
	}
}

func (d *locationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *locationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	apiResp, err := d.client.CommonServicesApiListLocationsWithResponse(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Inter.link locations", err.Error())
		return
	}
	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unexpected response reading Inter.link locations",
			fmt.Sprintf("HTTP %s: %s", apiResp.Status(), string(apiResp.Body)),
		)
		return
	}

	var state locationsDataSourceModel
	for _, l := range *apiResp.JSON200 {
		state.Locations = append(state.Locations, locationModel{
			Id:          types.Int64Value(int64(l.Id)),
			Name:        types.StringValue(l.Name),
			Description: types.StringValue(l.Description),
			Line1:       types.StringValue(l.Line1),
			Line2:       optionalString(l.Line2),
			PostalCode:  types.StringValue(l.PostalCode),
			City:        types.StringValue(l.City.Name),
			Country:     types.StringValue(l.City.Country.Name),
			Latitude:    optionalFloat(l.Latitude),
			Longitude:   optionalFloat(l.Longitude),
			Status:      types.StringValue(string(l.Status)),
			Type:        types.StringValue(string(l.Type)),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func optionalString(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

func optionalFloat(f *float32) types.Float64 {
	if f == nil {
		return types.Float64Null()
	}
	return types.Float64Value(float64(*f))
}
