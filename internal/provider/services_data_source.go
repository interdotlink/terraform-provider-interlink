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

type servicesDataSource struct {
	client *portal.ClientWithResponses
}

type servicesDataSourceModel struct {
	Services []serviceModel `tfsdk:"services"`
}

type serviceModel struct {
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
}

func NewServicesDataSource() datasource.DataSource {
	return &servicesDataSource{}
}

func (d *servicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_services"
}

func (d *servicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
					},
				},
			},
		},
	}
}

func (d *servicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *servicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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

	var state servicesDataSourceModel
	for _, item := range *apiResp.JSON200 {
		// Every service type extends the base Service and its extra fields are
		// optional, so AsService decodes any item into the shared base fields.
		// Type-specific extras are deferred to later steps.
		s, err := item.AsService()
		if err != nil {
			resp.Diagnostics.AddError("Unable to decode Inter.link service", err.Error())
			return
		}

		state.Services = append(state.Services, serviceModel{
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
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
