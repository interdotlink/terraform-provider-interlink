package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-interlink/internal/portal"
)

type portServicesDataSource struct {
	client *portal.ClientWithResponses
}

type portServicesDataSourceModel struct {
	Services []portServiceModel `tfsdk:"services"`
}

type portServiceModel struct {
	serviceBaseModel

	// PortService extra: its list of components.
	Components []portComponentModel `tfsdk:"components"`
}

// portComponentModel is a curated view of one component; every component
// subtype shares the ServiceComponent base envelope, so AsServiceComponent
// decodes any of them.
type portComponentModel struct {
	ComponentType types.String `tfsdk:"component_type"`
	Name          types.String `tfsdk:"name"`
	ResponseType  types.String `tfsdk:"response_type"`
	Price         types.String `tfsdk:"price"`
}

func NewPortServicesDataSource() datasource.DataSource {
	return &portServicesDataSource{}
}

func (d *portServicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_port_services"
}

func (d *portServicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := baseServiceAttributes()
	attrs["components"] = schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"component_type": schema.StringAttribute{Computed: true},
				"name":           schema.StringAttribute{Computed: true},
				"response_type":  schema.StringAttribute{Computed: true},
				"price":          schema.StringAttribute{Computed: true},
			},
		},
	}

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

func (d *portServicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *portServicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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

	var state portServicesDataSourceModel
	for _, item := range *apiResp.JSON200 {
		rt, err := item.Discriminator()
		if err != nil {
			resp.Diagnostics.AddError("Unable to read Inter.link service type", err.Error())
			return
		}
		if rt != "PortService" {
			continue
		}

		// Base fields shared with interlink_services; extras from the Port schema.
		base, err := item.AsService()
		if err != nil {
			resp.Diagnostics.AddError("Unable to decode Inter.link service", err.Error())
			return
		}
		s, err := item.AsPortService()
		if err != nil {
			resp.Diagnostics.AddError("Unable to decode Inter.link port service", err.Error())
			return
		}

		// Each component is itself a union; decode via the shared base envelope.
		var components []portComponentModel
		for _, item := range s.Components {
			c, err := item.AsServiceComponent()
			if err != nil {
				resp.Diagnostics.AddError("Unable to decode Inter.link port component", err.Error())
				return
			}
			components = append(components, portComponentModel{
				ComponentType: types.StringValue(c.ComponentType),
				Name:          types.StringValue(c.Name),
				ResponseType:  types.StringValue(string(c.ResponseType)),
				Price:         types.StringValue(c.Price.Display),
			})
		}

		state.Services = append(state.Services, portServiceModel{
			serviceBaseModel: mapBaseService(base),
			Components:        components,
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
