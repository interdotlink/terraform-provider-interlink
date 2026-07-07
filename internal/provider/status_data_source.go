package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-interlink/internal/portal"
)

type statusDataSource struct {
	client *portal.ClientWithResponses
}

type statusDataSourceModel struct {
	Status                types.String `tfsdk:"status"`
	HasPlannedMaintenance types.Bool   `tfsdk:"has_planned_maintenance"`
	HasOngoingMaintenance types.Bool   `tfsdk:"has_ongoing_maintenance"`
}

func NewStatusDataSource() datasource.DataSource {
	return &statusDataSource{}
}

func (d *statusDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_status"
}

func (d *statusDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the current operational status of the Inter.link platform.",
		Attributes: map[string]schema.Attribute{
			"status": schema.StringAttribute{
				Description: "Overall platform status (e.g. `success`).",
				Computed:    true,
			},
			"has_planned_maintenance": schema.BoolAttribute{
				Description: "Whether there is planned maintenance scheduled.",
				Computed:    true,
			},
			"has_ongoing_maintenance": schema.BoolAttribute{
				Description: "Whether there is maintenance currently in progress.",
				Computed:    true,
			},
		},
	}
}

func (d *statusDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *statusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	apiResp, err := d.client.PortalStatusApiGetStatusWithResponse(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Inter.link status", err.Error())
		return
	}
	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unexpected response reading Inter.link status",
			fmt.Sprintf("HTTP %s: %s", apiResp.Status(), string(apiResp.Body)),
		)
		return
	}

	body := *apiResp.JSON200
	var state statusDataSourceModel

	if v, ok := body["status"].(string); ok {
		state.Status = types.StringValue(v)
	}
	if v, ok := body["has_planned_maintenance"].(bool); ok {
		state.HasPlannedMaintenance = types.BoolValue(v)
	}
	if v, ok := body["has_ongoing_maintenance"].(bool); ok {
		state.HasOngoingMaintenance = types.BoolValue(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
