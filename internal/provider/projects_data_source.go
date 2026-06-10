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

type projectsDataSource struct {
	client *portal.ClientWithResponses
}

type projectsDataSourceModel struct {
	Projects []projectModel `tfsdk:"projects"`
}

type projectModel struct {
	Id          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	CreatedDate types.String `tfsdk:"created_date"`
	Prj         types.String `tfsdk:"prj"`
	Status      types.String `tfsdk:"status"`
	Description types.String `tfsdk:"description"`
}

func NewProjectsDataSource() datasource.DataSource {
	return &projectsDataSource{}
}

func (d *projectsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects"
}

func (d *projectsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"projects": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.Int64Attribute{Computed: true},
						"name":         schema.StringAttribute{Computed: true},
						"created_date": schema.StringAttribute{Computed: true},
						"prj":          schema.StringAttribute{Computed: true},
						"status":       schema.StringAttribute{Computed: true},
						"description":  schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *projectsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	apiResp, err := d.client.CommonServicesApiListProjectsWithResponse(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Inter.link projects", err.Error())
		return
	}
	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unexpected response reading Inter.link projects",
			fmt.Sprintf("HTTP %s: %s", apiResp.Status(), string(apiResp.Body)),
		)
		return
	}

	var state projectsDataSourceModel
	for _, p := range *apiResp.JSON200 {
		state.Projects = append(state.Projects, projectModel{
			Id:          optionalInt(p.Id),
			Name:        types.StringValue(p.Name),
			CreatedDate: types.StringValue(p.CreatedDate.Format(time.RFC3339)),
			Prj:         types.StringValue(p.Prj),
			Status:      optionalString(p.Status),
			Description: optionalString(p.Description),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func optionalInt(i *int) types.Int64 {
	if i == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*i))
}
