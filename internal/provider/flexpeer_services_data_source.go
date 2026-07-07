package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-interlink/internal/portal"
)

type flexpeerServicesDataSource struct {
	client *portal.ClientWithResponses
}

type flexpeerServicesDataSourceModel struct {
	Services []flexpeerServiceModel `tfsdk:"services"`
}

type flexpeerServiceModel struct {
	serviceBaseModel

	// FlexPeerEndpoint extras.
	FlexconnectKey     types.String            `tfsdk:"flexconnect_key"`
	RemoteCustomer     types.String            `tfsdk:"remote_customer"`
	VirtualNetworkName types.String            `tfsdk:"virtual_network_name"`
	VlanId             types.Int64             `tfsdk:"vlan_id"`
	Locations          []types.String          `tfsdk:"locations"`
	Components         []serviceComponentModel `tfsdk:"components"`
}

func NewFlexpeerServicesDataSource() datasource.DataSource {
	return &flexpeerServicesDataSource{}
}

func (d *flexpeerServicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_flexpeer_services"
}

func (d *flexpeerServicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := baseServiceAttributes()
	attrs["flexconnect_key"] = schema.StringAttribute{Description: "FlexConnect key used to establish the peering. Sensitive.", Computed: true, Sensitive: true}
	attrs["remote_customer"] = schema.StringAttribute{Description: "Name of the remote peering customer.", Computed: true}
	attrs["virtual_network_name"] = schema.StringAttribute{Description: "Name of the virtual network this endpoint belongs to.", Computed: true}
	attrs["vlan_id"] = schema.Int64Attribute{Description: "VLAN ID, if tagged.", Computed: true}
	attrs["locations"] = schema.ListAttribute{Description: "Names of the locations (endpoints) in the same virtual network.", Computed: true, ElementType: types.StringType}
	attrs["components"] = componentsAttribute()

	resp.Schema = schema.Schema{
		Description: "Detailed view of FlexPeer endpoint services, including the remote peer and billable components.",
		Attributes: map[string]schema.Attribute{
			"services": schema.ListNestedAttribute{
				Description: "FlexPeer endpoint services.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: attrs,
				},
			},
		},
	}
}

func (d *flexpeerServicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *flexpeerServicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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

	var state flexpeerServicesDataSourceModel
	for _, item := range *apiResp.JSON200 {
		rt, err := item.Discriminator()
		if err != nil {
			resp.Diagnostics.AddError("Unable to read Inter.link service type", err.Error())
			return
		}
		// Only our own endpoints; FlexPeerRemoteEndpoint (the other party's
		// side) carries no service data of ours.
		if rt != "FlexPeerEndpoint" {
			continue
		}

		// Base fields shared with interlink_services; extras from the FlexPeer schema.
		base, err := item.AsService()
		if err != nil {
			resp.Diagnostics.AddError("Unable to decode Inter.link service", err.Error())
			return
		}
		s, err := item.AsFlexPeerEndpoint()
		if err != nil {
			resp.Diagnostics.AddError("Unable to decode Inter.link FlexPeer service", err.Error())
			return
		}

		vlanId := types.Int64Null()
		if s.Vlan != nil {
			vlanId = optionalInt(s.Vlan.VlanId)
		}

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
				resp.Diagnostics.AddError("Unable to decode Inter.link FlexPeer component", err.Error())
				return
			}
			components = append(components, mapServiceComponent(c))
		}

		state.Services = append(state.Services, flexpeerServiceModel{
			serviceBaseModel:   mapBaseService(base),
			FlexconnectKey:     optionalString(s.FlexconnectKey),
			RemoteCustomer:     optionalString(s.RemoteCustomer),
			VirtualNetworkName: optionalString(s.VirtualNetworkName),
			VlanId:             vlanId,
			Locations:          locations,
			Components:         components,
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
