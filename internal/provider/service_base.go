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
