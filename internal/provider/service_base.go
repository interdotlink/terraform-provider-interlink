package provider

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-interlink/internal/portal"
)

// serviceBaseModel holds the fields common to every service type (the base
// Service schema). Embedded by each service data source's row model; the
// framework promotes its tfsdk-tagged fields into the object.
type serviceBaseModel struct {
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

// baseServiceAttributes returns a fresh map of the base service schema
// attributes. Per-family data sources merge their own extras into it.
func baseServiceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id":            schema.Int64Attribute{Description: "Internal numeric service ID.", Computed: true},
		"sid":           schema.StringAttribute{Description: "Service identifier (e.g. `SID1194`).", Computed: true},
		"name":          schema.StringAttribute{Description: "Human-readable service name.", Computed: true},
		"response_type": schema.StringAttribute{Description: "Service type discriminator (e.g. `IPTransit`, `FlexTunnel`, `DdosProtection`).", Computed: true},
		"status":        schema.StringAttribute{Description: "Current service status.", Computed: true},
		"product":       schema.StringAttribute{Description: "Name of the product this service is an instance of.", Computed: true},
		"location":      schema.StringAttribute{Description: "Name of the primary location for this service.", Computed: true},
		"service_speed": schema.Int64Attribute{Description: "Service speed in Mbps.", Computed: true},
		"term":          schema.Int64Attribute{Description: "Contract term in months.", Computed: true},
		"mrc":           schema.StringAttribute{Description: "Monthly recurring charge, formatted for display (e.g. `€75.00`).", Computed: true},
		"created_date":  schema.StringAttribute{Description: "Service creation timestamp (RFC 3339).", Computed: true},
		"customer_gid":  schema.StringAttribute{Description: "Global identifier of the customer that owns this service.", Computed: true},
		"description":   schema.StringAttribute{Description: "Optional free-text service description.", Computed: true},
	}
}

// serviceComponentModel is a curated view of one service component; every
// component subtype shares the ServiceComponent base envelope, so
// AsServiceComponent decodes any of them.
type serviceComponentModel struct {
	ComponentType types.String `tfsdk:"component_type"`
	Name          types.String `tfsdk:"name"`
	ResponseType  types.String `tfsdk:"response_type"`
	Price         types.String `tfsdk:"price"`
}

// componentsAttribute returns the schema for a curated components list.
func componentsAttribute() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Computed:    true,
		Description: "Billable components that make up this service.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"component_type": schema.StringAttribute{Description: "Component type discriminator.", Computed: true},
				"name":           schema.StringAttribute{Description: "Human-readable component name.", Computed: true},
				"response_type":  schema.StringAttribute{Description: "Component subtype (e.g. `LAG`, `BillingGroup`, `VLAN`).", Computed: true},
				"price":          schema.StringAttribute{Description: "Component price, formatted for display.", Computed: true},
			},
		},
	}
}

// mapServiceComponent maps a decoded base ServiceComponent into the curated model.
func mapServiceComponent(c portal.ServiceComponent) serviceComponentModel {
	return serviceComponentModel{
		ComponentType: types.StringValue(c.ComponentType),
		Name:          types.StringValue(c.Name),
		ResponseType:  types.StringValue(string(c.ResponseType)),
		Price:         types.StringValue(c.Price.Display),
	}
}

// mapBaseService maps a decoded base Service into the shared model.
func mapBaseService(s portal.Service) serviceBaseModel {
	return serviceBaseModel{
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
	}
}
