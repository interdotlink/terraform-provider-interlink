package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-interlink/internal/portal"
)

type ipTransitResource struct {
	client *portal.ClientWithResponses
}

// ipTransitResourceModel is the top-level state/plan for interlink_ip_transit.
// The port union is modelled as three sibling optional blocks; exactly one is
// set (enforced by a validator wired in a later step).
type ipTransitResourceModel struct {
	// Required create arguments.
	BgpsessionAsn           types.Int64  `tfsdk:"bgpsession_asn"`
	BgpsessionAsSet         types.String `tfsdk:"bgpsession_as_set"`
	BgpsessionPrefixLimitV4 types.Int64  `tfsdk:"bgpsession_prefix_limit_v4"`
	BgpsessionPrefixLimitV6 types.Int64  `tfsdk:"bgpsession_prefix_limit_v6"`
	PrefixV4Size            types.Int64  `tfsdk:"prefix_v4_size"`
	PrefixV6Size            types.Int64  `tfsdk:"prefix_v6_size"`
	Term                    types.Int64  `tfsdk:"term"`
	SyncFromPdb             types.Bool   `tfsdk:"sync_from_pdb"`

	// Optional create arguments.
	BgpsessionPassword    types.String `tfsdk:"bgpsession_password"`
	AggregatedBilling     types.Bool   `tfsdk:"aggregated_billing"`
	OutboundAdvertisement types.String `tfsdk:"outbound_advertisement"`
	PurchaseReference     types.String `tfsdk:"purchase_reference"`

	// Port union — exactly one of these blocks is set. Modelled as
	// at-most-one-element lists (ListNestedBlock) so required inner
	// attributes are enforced only when the block is present.
	NewPort      []newPortModel      `tfsdk:"new_port"`
	ExistingPort []existingPortModel `tfsdk:"existing_port"`
	ExistingLag  []existingLagModel  `tfsdk:"existing_lag"`

	// Computed read-back attributes (from the IPTransit response).
	Id            types.Int64  `tfsdk:"id"`
	Sid           types.String `tfsdk:"sid"`
	Name          types.String `tfsdk:"name"`
	Status        types.String `tfsdk:"status"`
	ServiceSpeed  types.Int64  `tfsdk:"service_speed"`
	CustomerGid   types.String `tfsdk:"customer_gid"`
	EndDate       types.String `tfsdk:"end_date"`
	NoticePeriod  types.Int64  `tfsdk:"notice_period"`
	RenewalPeriod types.Int64  `tfsdk:"renewal_period"`
}

// newPortModel provisions a brand-new port (UserNetworkInterface).
type newPortModel struct {
	Location       types.String `tfsdk:"location"`
	Bandwidth      types.Int64  `tfsdk:"bandwidth"`
	PortType       types.String `tfsdk:"port_type"`
	VlanId         types.Int64  `tfsdk:"vlan_id"`
	VlanType       types.String `tfsdk:"vlan_type"`
	LagMemberCount types.Int64  `tfsdk:"lag_member_count"`
	LagName        types.String `tfsdk:"lag_name"`
}

// existingPortModel attaches to an existing port (UserNetworkInterfaceOnExistingPort).
type existingPortModel struct {
	Id        types.Int64  `tfsdk:"id"`
	Bandwidth types.Int64  `tfsdk:"bandwidth"`
	VlanId    types.Int64  `tfsdk:"vlan_id"`
	VlanType  types.String `tfsdk:"vlan_type"`
}

// existingLagModel attaches to an existing LAG (UserNetworkInterfaceOnExistingLAG).
type existingLagModel struct {
	LagId     types.Int64  `tfsdk:"lag_id"`
	Bandwidth types.Int64  `tfsdk:"bandwidth"`
	VlanId    types.Int64  `tfsdk:"vlan_id"`
	VlanType  types.String `tfsdk:"vlan_type"`
}

func NewIpTransitResource() resource.Resource {
	return &ipTransitResource{}
}

func (r *ipTransitResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_transit"
}

func (r *ipTransitResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	// VLAN attributes shared by all three port blocks.
	vlanAttributes := func() map[string]schema.Attribute {
		return map[string]schema.Attribute{
			"vlan_id": schema.Int64Attribute{
				Description: "VLAN ID in the range 1-4094. Configured here but returned at the top level on read.",
				Optional:    true,
			},
			"vlan_type": schema.StringAttribute{
				Description: "VLAN tagging mode: `Tagged` or `Untagged`.",
				Optional:    true,
			},
		}
	}

	newPortAttributes := vlanAttributes()
	newPortAttributes["location"] = schema.StringAttribute{
		Description: "Location name where the new port is provisioned.",
		Required:    true,
	}
	newPortAttributes["bandwidth"] = schema.Int64Attribute{
		Description: "Committed data rate (CDR) in Mbps, in the range 0-400000. This is the service speed.",
		Required:    true,
	}
	newPortAttributes["port_type"] = schema.StringAttribute{
		Description: "Physical port type, e.g. `100G-LR4 (QSFP28)` or `10G-LR (SFP+)`.",
		Required:    true,
	}
	newPortAttributes["lag_member_count"] = schema.Int64Attribute{
		Description: "Number of member ports when the new port is a LAG.",
		Optional:    true,
	}
	newPortAttributes["lag_name"] = schema.StringAttribute{
		Description: "Name for the LAG when the new port is a LAG.",
		Optional:    true,
	}

	existingPortAttributes := vlanAttributes()
	existingPortAttributes["id"] = schema.Int64Attribute{
		Description: "Numeric ID of the existing port to attach to.",
		Required:    true,
	}
	existingPortAttributes["bandwidth"] = schema.Int64Attribute{
		Description: "Committed data rate (CDR) in Mbps, in the range 0-400000.",
		Required:    true,
	}

	existingLagAttributes := vlanAttributes()
	existingLagAttributes["lag_id"] = schema.Int64Attribute{
		Description: "Numeric ID of the existing LAG to attach to.",
		Required:    true,
	}
	existingLagAttributes["bandwidth"] = schema.Int64Attribute{
		Description: "Committed data rate (CDR) in Mbps, in the range 0-400000.",
		Required:    true,
	}

	resp.Schema = schema.Schema{
		Description: "Manages an Inter.link IP Transit service. Note: this service cannot be cancelled through the API — `terraform destroy` will fail by design (see the delete behaviour). Use `terraform state rm` to stop managing it without cancelling.",
		Attributes: map[string]schema.Attribute{
			// Required create arguments.
			"bgpsession_asn": schema.Int64Attribute{
				Description: "Customer BGP autonomous system number, in the range 1-4294967295.",
				Required:    true,
			},
			"bgpsession_as_set": schema.StringAttribute{
				Description: "AS-SET name according to RFC2622.",
				Required:    true,
			},
			"bgpsession_prefix_limit_v4": schema.Int64Attribute{
				Description: "IPv4 prefix limit, in the range 0-140000.",
				Required:    true,
			},
			"bgpsession_prefix_limit_v6": schema.Int64Attribute{
				Description: "IPv6 prefix limit, in the range 0-70000.",
				Required:    true,
			},
			"prefix_v4_size": schema.Int64Attribute{
				Description: "Requested IPv4 prefix size (CIDR length).",
				Required:    true,
			},
			"prefix_v6_size": schema.Int64Attribute{
				Description: "Requested IPv6 prefix size (CIDR length).",
				Required:    true,
			},
			"term": schema.Int64Attribute{
				Description: "Contract term in months.",
				Required:    true,
			},
			"sync_from_pdb": schema.BoolAttribute{
				Description: "Whether to sync the BGP configuration from PeeringDB.",
				Required:    true,
			},

			// Optional create arguments.
			"bgpsession_password": schema.StringAttribute{
				Description: "BGP session password (MD5). Write-only — never read back from the API.",
				Optional:    true,
				Sensitive:   true,
			},
			"aggregated_billing": schema.BoolAttribute{
				Description: "Whether to bill this service as part of an aggregated commit.",
				Optional:    true,
			},
			"outbound_advertisement": schema.StringAttribute{
				Description: "Outbound BGP advertisement policy: `Default Route`, `Full Routing Table`, `Full Routing Table and Default Route`, `None - Outbound Only`, or `not set`.",
				Optional:    true,
			},
			"purchase_reference": schema.StringAttribute{
				Description: "Free-text purchase reference recorded against the order.",
				Optional:    true,
			},

			// Computed read-back attributes.
			"id": schema.Int64Attribute{
				Description: "Numeric service ID. Primary identifier used for read, update, and import.",
				Computed:    true,
			},
			"sid": schema.StringAttribute{
				Description: "Human-readable service ID (e.g. `SID1194`).",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Service name.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Service status. A new order starts at `DraftQuote` and advances to `Live` out of band; it catches up on later refreshes.",
				Computed:    true,
			},
			"service_speed": schema.Int64Attribute{
				Description: "Effective committed data rate (CDR) in Mbps.",
				Computed:    true,
			},
			"customer_gid": schema.StringAttribute{
				Description: "Customer global ID that owns the service.",
				Computed:    true,
			},
			"end_date": schema.StringAttribute{
				Description: "Contract end date. Answers when the service can be cancelled without early-termination charges.",
				Computed:    true,
			},
			"notice_period": schema.Int64Attribute{
				Description: "Cancellation notice period in days.",
				Computed:    true,
			},
			"renewal_period": schema.Int64Attribute{
				Description: "Automatic renewal period in months.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"new_port": schema.ListNestedBlock{
				Description: "Provision a new port for this service. Exactly one of `new_port`, `existing_port`, or `existing_lag` must be set.",
				NestedObject: schema.NestedBlockObject{
					Attributes: newPortAttributes,
				},
			},
			"existing_port": schema.ListNestedBlock{
				Description: "Attach this service to an existing port. Exactly one of `new_port`, `existing_port`, or `existing_lag` must be set.",
				NestedObject: schema.NestedBlockObject{
					Attributes: existingPortAttributes,
				},
			},
			"existing_lag": schema.ListNestedBlock{
				Description: "Attach this service to an existing LAG. Exactly one of `new_port`, `existing_port`, or `existing_lag` must be set.",
				NestedObject: schema.NestedBlockObject{
					Attributes: existingLagAttributes,
				},
			},
		},
	}
}

func (r *ipTransitResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*portal.ClientWithResponses)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *portal.ClientWithResponses, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *ipTransitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError(
		"Not implemented",
		"Create for interlink_ip_transit is implemented in a later step (17c-2).",
	)
}

func (r *ipTransitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.AddError(
		"Not implemented",
		"Read for interlink_ip_transit is implemented in a later step (17c-2).",
	)
}

func (r *ipTransitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Not implemented",
		"Update for interlink_ip_transit is implemented in a later step (17d).",
	)
}

// Delete always fails: an IP Transit service cannot be cancelled through the
// API, and cancellation is a commercial act (end of term, or early termination
// at the remaining contract value less 5%). Destroying must never silently
// cancel a billing service. Later steps refine this message to quote the
// service's own end_date / notice_period from state.
func (r *ipTransitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError(
		"IP Transit services cannot be destroyed",
		"An Inter.link IP Transit service cannot be cancelled through the API, and this "+
			"provider will never cancel one on destroy. Cancellation is contractual: it is "+
			"only possible towards the end of the term, or early against payment of the "+
			"remaining contract value less 5%.\n\n"+
			"To stop managing this service with Terraform without cancelling it, run "+
			"`terraform state rm` on this resource. To actually cancel it, contact Inter.link.",
	)
}
