package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-interlink/internal/portal"
)

// ipFamily are the service types that share the IPTransit schema, so AsIPTransit
// decodes any of them.
var ipFamily = map[string]bool{
	"IPTransit":  true,
	"IPAccess":   true,
	"FlexTunnel": true,
	"PniPeering": true,
}

type ipServicesDataSource struct {
	client *portal.ClientWithResponses
}

type ipServicesDataSourceModel struct {
	Services []ipServiceModel `tfsdk:"services"`
}

type ipServiceModel struct {
	// Base fields, shared with interlink_services.
	Id           types.Int64  `tfsdk:"id"`
	Sid          types.String `tfsdk:"sid"`
	Name         types.String `tfsdk:"name"`
	ResponseType types.String `tfsdk:"response_type"`
	Status       types.String `tfsdk:"status"`
	Product      types.String `tfsdk:"product"`
	Location     types.String `tfsdk:"location"`
	ServiceSpeed types.Int64  `tfsdk:"service_speed"`
	Term         types.Int64  `tfsdk:"term"`
	Mrc          types.String `tfsdk:"mrc"`
	CreatedDate  types.String `tfsdk:"created_date"`
	CustomerGid  types.String `tfsdk:"customer_gid"`
	Description  types.String `tfsdk:"description"`

	// IP-family extras.
	Port     types.String `tfsdk:"port"`
	VlanId   types.Int64  `tfsdk:"vlan_id"`
	PrefixV4 types.String `tfsdk:"prefix_v4"`
	PrefixV6 types.String `tfsdk:"prefix_v6"`
	BgpV4Asn types.Int64  `tfsdk:"bgp_v4_asn"`
	BgpV6Asn types.Int64  `tfsdk:"bgp_v6_asn"`
}

func NewIpServicesDataSource() datasource.DataSource {
	return &ipServicesDataSource{}
}

func (d *ipServicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_services"
}

func (d *ipServicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"services": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":            schema.Int64Attribute{Computed: true},
						"sid":           schema.StringAttribute{Computed: true},
						"name":          schema.StringAttribute{Computed: true},
						"response_type": schema.StringAttribute{Computed: true},
						"status":        schema.StringAttribute{Computed: true},
						"product":       schema.StringAttribute{Computed: true},
						"location":      schema.StringAttribute{Computed: true},
						"service_speed": schema.Int64Attribute{Computed: true},
						"term":          schema.Int64Attribute{Computed: true},
						"mrc":           schema.StringAttribute{Computed: true},
						"created_date":  schema.StringAttribute{Computed: true},
						"customer_gid":  schema.StringAttribute{Computed: true},
						"description":   schema.StringAttribute{Computed: true},
						"port":          schema.StringAttribute{Computed: true},
						"vlan_id":       schema.Int64Attribute{Computed: true},
						"prefix_v4":     schema.StringAttribute{Computed: true},
						"prefix_v6":     schema.StringAttribute{Computed: true},
						"bgp_v4_asn":    schema.Int64Attribute{Computed: true},
						"bgp_v6_asn":    schema.Int64Attribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *ipServicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ipServicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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

	var state ipServicesDataSourceModel
	for _, item := range *apiResp.JSON200 {
		rt, err := item.Discriminator()
		if err != nil {
			resp.Diagnostics.AddError("Unable to read Inter.link service type", err.Error())
			return
		}
		if !ipFamily[rt] {
			continue
		}

		s, err := item.AsIPTransit()
		if err != nil {
			resp.Diagnostics.AddError("Unable to decode Inter.link IP service", err.Error())
			return
		}

		// Each extra is an optional nested component; guard the pointer first.
		port := types.StringNull()
		if s.Port != nil {
			port = types.StringValue(s.Port.Name)
		}
		vlanId := types.Int64Null()
		if s.Vlan != nil {
			vlanId = optionalInt(s.Vlan.VlanId)
		}
		prefixV4 := types.StringNull()
		if s.PrefixV4 != nil {
			prefixV4 = types.StringValue(s.PrefixV4.Name)
		}
		prefixV6 := types.StringNull()
		if s.PrefixV6 != nil {
			prefixV6 = types.StringValue(s.PrefixV6.Name)
		}
		bgpV4Asn := types.Int64Null()
		if s.BgpsessionV4 != nil {
			bgpV4Asn = optionalInt(s.BgpsessionV4.BgpsessionAsn)
		}
		bgpV6Asn := types.Int64Null()
		if s.BgpsessionV6 != nil {
			bgpV6Asn = optionalInt(s.BgpsessionV6.BgpsessionAsn)
		}

		state.Services = append(state.Services, ipServiceModel{
			Id:           optionalInt(s.Id),
			Sid:          optionalString(s.Sid),
			Name:         types.StringValue(s.Name),
			ResponseType: types.StringValue(string(s.ResponseType)),
			Status:       types.StringValue(string(s.Status)),
			Product:      types.StringValue(s.Product.Name),
			Location:     types.StringValue(s.Location.Name),
			ServiceSpeed: optionalInt(s.ServiceSpeed),
			Term:         types.Int64Value(int64(s.Term)),
			Mrc:          types.StringValue(s.Mrc.Display),
			CreatedDate:  types.StringValue(s.CreatedDate.Format(time.RFC3339)),
			CustomerGid:  types.StringValue(s.CustomerGid),
			Description:  optionalString(s.Description),
			Port:         port,
			VlanId:       vlanId,
			PrefixV4:     prefixV4,
			PrefixV6:     prefixV6,
			BgpV4Asn:     bgpV4Asn,
			BgpV6Asn:     bgpV6Asn,
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
