package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-interlink/internal/portal"
)

type elanServicesDataSource struct {
	client *portal.ClientWithResponses
}

type elanServicesDataSourceModel struct {
	Services []elanServiceModel `tfsdk:"services"`
}

type elanServiceModel struct {
	serviceBaseModel

	// ElanEndpoint extras.
	VirtualNetworkName types.String            `tfsdk:"virtual_network_name"`
	VlanId             types.Int64             `tfsdk:"vlan_id"`
	Locations          []types.String          `tfsdk:"locations"`
	Components         []serviceComponentModel `tfsdk:"components"`
}

func NewElanServicesDataSource() datasource.DataSource {
	return &elanServicesDataSource{}
}

func (d *elanServicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elan_services"
}

func (d *elanServicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := baseServiceAttributes()
	attrs["virtual_network_name"] = schema.StringAttribute{Description: "Name of the virtual network this endpoint belongs to.", Computed: true}
	attrs["vlan_id"] = schema.Int64Attribute{Description: "VLAN ID, if tagged.", Computed: true}
	attrs["locations"] = schema.ListAttribute{Description: "Names of the other locations (endpoints) in the same virtual network.", Computed: true, ElementType: types.StringType}
	attrs["components"] = componentsAttribute()

	resp.Schema = schema.Schema{
		Description: "Detailed view of Flex Ethernet (ELAN) endpoint services, including their virtual network and billable components.",
		Attributes: map[string]schema.Attribute{
			"services": schema.ListNestedAttribute{
				Description: "ELAN endpoint services.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: attrs,
				},
			},
		},
	}
}

func (d *elanServicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *elanServicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	apiResp, err := d.client.CommonServicesApiListServicesWithResponse(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Inter.link services", err.Error())
		return
	}
	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unexpected response reading Inter.link services",
			fmt.Sprintf("HTTP %s: %s", apiResp.Status(), string(apiResp.Body)),
		)
		return
	}

	var state elanServicesDataSourceModel
	for _, item := range *apiResp.JSON200 {
		rt, err := item.Discriminator()
		if err != nil {
			resp.Diagnostics.AddError("Unable to read Inter.link service type", err.Error())
			return
		}
		if rt != "ElanEndpoint" {
			continue
		}

		// Base fields shared with interlink_services; extras from the Elan schema.
		base, err := item.AsService()
		if err != nil {
			resp.Diagnostics.AddError("Unable to decode Inter.link service", err.Error())
			return
		}
		s, err := item.AsElanEndpoint()
		if err != nil {
			resp.Diagnostics.AddError("Unable to decode Inter.link ELAN service", err.Error())
			return
		}

		vlanId := types.Int64Null()
		if s.Vlan != nil {
			vlanId = optionalInt(s.Vlan.VlanId)
		}

		// The other endpoints of the same virtual network, as location names.
		var locations []types.String
		if s.Locations != nil {
			for _, l := range *s.Locations {
				locations = append(locations, types.StringValue(l.Name))
			}
		}

		var components []serviceComponentModel
		for _, item := range s.Components {
			c, err := item.AsServiceComponent()
			if err != nil {
				resp.Diagnostics.AddError("Unable to decode Inter.link ELAN component", err.Error())
				return
			}
			components = append(components, mapServiceComponent(c))
		}

		state.Services = append(state.Services, elanServiceModel{
			serviceBaseModel:   mapBaseService(base),
			VirtualNetworkName: optionalString(s.VirtualNetworkName),
			VlanId:             vlanId,
			Locations:          locations,
			Components:         components,
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
