package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-interlink/internal/portal"
)

type ddosServicesDataSource struct {
	client *portal.ClientWithResponses
}

type ddosServicesDataSourceModel struct {
	Services []ddosServiceModel `tfsdk:"services"`
}

type ddosServiceModel struct {
	serviceBaseModel

	// DdosProtection extras (curated scalars).
	DdosProtectionType types.String `tfsdk:"ddos_protection_type"`
	AttackBandwidth    types.Int64  `tfsdk:"attack_bandwidth"`
	ScrubbingCapacity  types.Int64  `tfsdk:"scrubbing_capacity"`
	AdditionalNetworks types.Int64  `tfsdk:"additional_networks"`
	EnterpriseReports  types.Int64  `tfsdk:"enterprise_reports"`
}

func NewDdosServicesDataSource() datasource.DataSource {
	return &ddosServicesDataSource{}
}

func (d *ddosServicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ddos_services"
}

func (d *ddosServicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := baseServiceAttributes()
	attrs["ddos_protection_type"] = schema.StringAttribute{Computed: true}
	attrs["attack_bandwidth"] = schema.Int64Attribute{Computed: true}
	attrs["scrubbing_capacity"] = schema.Int64Attribute{Computed: true}
	attrs["additional_networks"] = schema.Int64Attribute{Computed: true}
	attrs["enterprise_reports"] = schema.Int64Attribute{Computed: true}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"services": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: attrs,
				},
			},
		},
	}
}

func (d *ddosServicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ddosServicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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

	var state ddosServicesDataSourceModel
	for _, item := range *apiResp.JSON200 {
		rt, err := item.Discriminator()
		if err != nil {
			resp.Diagnostics.AddError("Unable to read Inter.link service type", err.Error())
			return
		}
		if rt != "DdosProtection" {
			continue
		}

		// Base fields shared with interlink_services; extras from the DDoS schema.
		base, err := item.AsService()
		if err != nil {
			resp.Diagnostics.AddError("Unable to decode Inter.link service", err.Error())
			return
		}
		s, err := item.AsDdosProtection()
		if err != nil {
			resp.Diagnostics.AddError("Unable to decode Inter.link DDoS service", err.Error())
			return
		}

		// Each extra is an optional nested component; guard the pointer first.
		protType := types.StringNull()
		if s.DdosProtection != nil {
			protType = optionalString(s.DdosProtection.DdosProtectionType)
		}
		attackBandwidth := types.Int64Null()
		if s.DdosProtectionAttackBandwidth != nil {
			attackBandwidth = types.Int64Value(int64(s.DdosProtectionAttackBandwidth.Quantity))
		}
		scrubbingCapacity := types.Int64Null()
		if s.DdosProtectionScrubbingCapacity != nil {
			scrubbingCapacity = types.Int64Value(int64(s.DdosProtectionScrubbingCapacity.Quantity))
		}
		additionalNetworks := types.Int64Null()
		if s.DdosProtectionAdditionalNetworks != nil {
			additionalNetworks = types.Int64Value(int64(s.DdosProtectionAdditionalNetworks.Quantity))
		}
		enterpriseReports := types.Int64Null()
		if s.DdosProtectionEnterpriseReports != nil {
			enterpriseReports = types.Int64Value(int64(s.DdosProtectionEnterpriseReports.Quantity))
		}

		state.Services = append(state.Services, ddosServiceModel{
			serviceBaseModel:   mapBaseService(base),
			DdosProtectionType: protType,
			AttackBandwidth:    attackBandwidth,
			ScrubbingCapacity:  scrubbingCapacity,
			AdditionalNetworks: additionalNetworks,
			EnterpriseReports:  enterpriseReports,
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
