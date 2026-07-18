package provider

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-interlink/internal/portal"
)

// Enum values accepted by the resource, derived from the generated client
// types. GRE Tunnel is intentionally excluded from the port types. IP Transit
// over GRE is out of scope for this resource.
var nonGREPortTypes = []string{
	string(portal.N100GCUSTOMQSFP28),
	string(portal.N100GLR1QSFP28),
	string(portal.N100GLR4QSFP28),
	string(portal.N10GCUSTOMSFP),
	string(portal.N10GLRSFP),
	string(portal.N1GCUSTOMSFP),
	string(portal.N1GLXSFP),
	string(portal.N25GCUSTOMSFP28),
	string(portal.N25GLRSFP28),
	string(portal.N400GCUSTOMQSFP28),
	string(portal.N400GLR4OSFP),
	string(portal.N400GLR4OSFPQSFPDD),
}

var vlanTypeValues = []string{
	string(portal.VlanTypesTagged),
	string(portal.VlanTypesUntagged),
}

var outboundAdvertisementValues = []string{
	string(portal.DefaultRoute),
	string(portal.FullRoutingTable),
	string(portal.FullRoutingTableAndDefaultRoute),
	string(portal.NoneOutboundOnly),
	string(portal.NotSet),
}

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

	// Port union: exactly one of these blocks is set. Modelled as
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

// ConfigValidators enforces that exactly one of the three port blocks is set.
func (r *ipTransitResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("new_port"),
			path.MatchRoot("existing_port"),
			path.MatchRoot("existing_lag"),
		),
	}
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
				Validators:  []validator.String{stringvalidator.OneOf(vlanTypeValues...)},
			},
		}
	}

	newPortAttributes := vlanAttributes()
	newPortAttributes["location"] = schema.StringAttribute{
		Description:   "Location name where the new port is provisioned.",
		Required:      true,
		PlanModifiers: []planmodifier.String{immutableString()},
	}
	newPortAttributes["bandwidth"] = schema.Int64Attribute{
		Description: "Committed data rate (CDR) in Mbps, in the range 0-400000. This is the service speed.",
		Required:    true,
	}
	newPortAttributes["port_type"] = schema.StringAttribute{
		Description:   "Physical port type, e.g. `100G-LR4 (QSFP28)` or `10G-LR (SFP+)`.",
		Required:      true,
		Validators:    []validator.String{stringvalidator.OneOf(nonGREPortTypes...)},
		PlanModifiers: []planmodifier.String{immutableString()},
	}
	newPortAttributes["lag_member_count"] = schema.Int64Attribute{
		Description:   "Number of member ports when the new port is a LAG.",
		Optional:      true,
		PlanModifiers: []planmodifier.Int64{immutableInt64()},
	}
	newPortAttributes["lag_name"] = schema.StringAttribute{
		Description:   "Name for the LAG when the new port is a LAG.",
		Optional:      true,
		PlanModifiers: []planmodifier.String{immutableString()},
	}

	existingPortAttributes := vlanAttributes()
	existingPortAttributes["id"] = schema.Int64Attribute{
		Description:   "Numeric ID of the existing port to attach to.",
		Required:      true,
		PlanModifiers: []planmodifier.Int64{immutableInt64()},
	}
	existingPortAttributes["bandwidth"] = schema.Int64Attribute{
		Description: "Committed data rate (CDR) in Mbps, in the range 0-400000.",
		Required:    true,
	}

	existingLagAttributes := vlanAttributes()
	existingLagAttributes["lag_id"] = schema.Int64Attribute{
		Description:   "Numeric ID of the existing LAG to attach to.",
		Required:      true,
		PlanModifiers: []planmodifier.Int64{immutableInt64()},
	}
	existingLagAttributes["bandwidth"] = schema.Int64Attribute{
		Description: "Committed data rate (CDR) in Mbps, in the range 0-400000.",
		Required:    true,
	}

	resp.Schema = schema.Schema{
		Description: "Manages an Inter.link IP Transit service. Creating the resource submits the order and returns immediately with the service at status `DraftQuote`. The order is placed but the service is not provisioned yet; provisioning continues asynchronously out of band, and `status` advances (towards `Live`) and catches up on later refreshes. In-place updates are not supported in this release; attributes cannot be changed after creation. Note: this service cannot be cancelled through the API, so `terraform destroy` will fail by design (see the delete behaviour). Use `terraform state rm` to stop managing it without cancelling.",
		Attributes: map[string]schema.Attribute{
			// Required create arguments.
			"bgpsession_asn": schema.Int64Attribute{
				Description:   "Customer BGP autonomous system number, in the range 1-4294967295.",
				Required:      true,
				PlanModifiers: []planmodifier.Int64{immutableInt64()},
			},
			"bgpsession_as_set": schema.StringAttribute{
				Description:   "AS-SET name according to RFC2622.",
				Required:      true,
				PlanModifiers: []planmodifier.String{immutableString()},
			},
			"bgpsession_prefix_limit_v4": schema.Int64Attribute{
				Description:   "IPv4 prefix limit, in the range 0-140000.",
				Required:      true,
				PlanModifiers: []planmodifier.Int64{immutableInt64()},
			},
			"bgpsession_prefix_limit_v6": schema.Int64Attribute{
				Description:   "IPv6 prefix limit, in the range 0-70000.",
				Required:      true,
				PlanModifiers: []planmodifier.Int64{immutableInt64()},
			},
			"prefix_v4_size": schema.Int64Attribute{
				Description:   "Requested IPv4 prefix size (CIDR length): `30` or `31`.",
				Required:      true,
				Validators:    []validator.Int64{int64validator.OneOf(30, 31)},
				PlanModifiers: []planmodifier.Int64{immutableInt64()},
			},
			"prefix_v6_size": schema.Int64Attribute{
				Description:   "Requested IPv6 prefix size (CIDR length): `126` or `127`.",
				Required:      true,
				Validators:    []validator.Int64{int64validator.OneOf(126, 127)},
				PlanModifiers: []planmodifier.Int64{immutableInt64()},
			},
			"term": schema.Int64Attribute{
				Description:   "Contract term in months.",
				Required:      true,
				PlanModifiers: []planmodifier.Int64{immutableInt64()},
			},
			"sync_from_pdb": schema.BoolAttribute{
				Description: "Whether to sync the BGP configuration from PeeringDB. Write-only: supplied on create and never stored in state.",
				Required:    true,
				WriteOnly:   true,
			},

			// Optional create arguments.
			"bgpsession_password": schema.StringAttribute{
				Description: "BGP session password (MD5). Required by the API. Write-only: supplied on create and never stored in state or read back.",
				Required:    true,
				Sensitive:   true,
				WriteOnly:   true,
			},
			"aggregated_billing": schema.BoolAttribute{
				Description:   "Whether to bill this service as part of an aggregated commit.",
				Optional:      true,
				PlanModifiers: []planmodifier.Bool{immutableBool()},
			},
			"outbound_advertisement": schema.StringAttribute{
				Description:   "Outbound BGP advertisement policy: `Default Route`, `Full Routing Table`, `Full Routing Table and Default Route`, `None - Outbound Only`, or `not set`.",
				Optional:      true,
				Validators:    []validator.String{stringvalidator.OneOf(outboundAdvertisementValues...)},
				PlanModifiers: []planmodifier.String{immutableString()},
			},
			"purchase_reference": schema.StringAttribute{
				Description:   "Free-text purchase reference recorded against the order.",
				Optional:      true,
				PlanModifiers: []planmodifier.String{immutableString()},
			},

			// Computed read-back attributes.
			"id": schema.Int64Attribute{
				Description: "Numeric service ID and primary identifier for the service.",
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
				Validators: []validator.List{listvalidator.SizeAtMost(1)},
			},
			"existing_port": schema.ListNestedBlock{
				Description: "Attach this service to an existing port. Exactly one of `new_port`, `existing_port`, or `existing_lag` must be set.",
				NestedObject: schema.NestedBlockObject{
					Attributes: existingPortAttributes,
				},
				Validators: []validator.List{listvalidator.SizeAtMost(1)},
			},
			"existing_lag": schema.ListNestedBlock{
				Description: "Attach this service to an existing LAG. Exactly one of `new_port`, `existing_port`, or `existing_lag` must be set.",
				NestedObject: schema.NestedBlockObject{
					Attributes: existingLagAttributes,
				},
				Validators: []validator.List{listvalidator.SizeAtMost(1)},
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
	var m ipTransitResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Write-only arguments (bgpsession_password, sync_from_pdb) are never stored
	// in state and so are absent from the plan; read their values from config.
	var config ipTransitResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	port, err := m.toCreatePort()
	if err != nil {
		resp.Diagnostics.AddError("Unable to build IP Transit port", err.Error())
		return
	}

	body := portal.IPTransitCreate{
		BgpsessionAsn:           int(m.BgpsessionAsn.ValueInt64()),
		BgpsessionAsSet:         m.BgpsessionAsSet.ValueString(),
		BgpsessionPrefixLimitV4: int(m.BgpsessionPrefixLimitV4.ValueInt64()),
		BgpsessionPrefixLimitV6: int(m.BgpsessionPrefixLimitV6.ValueInt64()),
		PrefixV4Size:            int(m.PrefixV4Size.ValueInt64()),
		PrefixV6Size:            int(m.PrefixV6Size.ValueInt64()),
		Term:                    int(m.Term.ValueInt64()),
		SyncFromPdb:             config.SyncFromPdb.ValueBool(),
		Port:                    port,
		BgpsessionPassword:      stringPtrOrNil(config.BgpsessionPassword),
		PurchaseReference:       stringPtrOrNil(m.PurchaseReference),
	}
	if !m.AggregatedBilling.IsNull() {
		ab := m.AggregatedBilling.ValueBool()
		body.AggregatedBilling = &ab
	}
	if !m.OutboundAdvertisement.IsNull() {
		oa := portal.BgpAdvertisementTypes(m.OutboundAdvertisement.ValueString())
		body.OutboundAdvertisement = &oa
	}

	apiResp, err := r.client.IpTransitApiCreateIpTransitWithResponse(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Inter.link IP Transit service", err.Error())
		return
	}
	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unexpected response creating Inter.link IP Transit service",
			fmt.Sprintf("HTTP %s: %s", apiResp.Status(), string(apiResp.Body)),
		)
		return
	}

	s, err := apiResp.JSON200.AsIPTransit()
	if err != nil {
		resp.Diagnostics.AddError("Unable to decode created IP Transit service", err.Error())
		return
	}
	mapIPTransitComputed(&m, s)

	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *ipTransitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var m ipTransitResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if m.Id.IsNull() {
		return
	}

	serviceId := strconv.FormatInt(m.Id.ValueInt64(), 10)
	apiResp, err := r.client.CommonServicesApiGetServiceWithResponse(ctx, serviceId)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Inter.link IP Transit service", err.Error())
		return
	}
	if apiResp.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if apiResp.JSON200 == nil {
		resp.Diagnostics.AddError(
			"Unexpected response reading Inter.link IP Transit service",
			fmt.Sprintf("HTTP %s: %s", apiResp.Status(), string(apiResp.Body)),
		)
		return
	}

	s, err := apiResp.JSON200.AsIPTransit()
	if err != nil {
		resp.Diagnostics.AddError("Unable to decode Inter.link IP Transit service", err.Error())
		return
	}
	// Only the computed read-back attributes are refreshed from the API; the
	// create arguments and the port block stay config-authoritative. (The VLAN
	// round-trip into the port block is added with Import in a later step.)
	mapIPTransitComputed(&m, s)

	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func (r *ipTransitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"In-place updates are not supported",
		"The interlink_ip_transit resource does not support in-place updates in this release. "+
			"Changes to mutable attributes (bandwidth, vlan_id, vlan_type) are not yet available; "+
			"revert them to their previous values.",
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

// toCreatePort builds the port union for the create body from whichever port
// block is set. Exactly one is set (guaranteed by the ExactlyOneOf validator).
func (m ipTransitResourceModel) toCreatePort() (portal.IPTransitCreate_Port, error) {
	var port portal.IPTransitCreate_Port
	switch {
	case len(m.NewPort) == 1:
		b := m.NewPort[0]
		err := port.FromUserNetworkInterface(portal.UserNetworkInterface{
			Location:       b.Location.ValueString(),
			Bandwidth:      int(b.Bandwidth.ValueInt64()),
			PortType:       portal.PortTypes(b.PortType.ValueString()),
			VlanId:         int64PtrOrNil(b.VlanId),
			VlanType:       vlanTypePtr(b.VlanType),
			LagMemberCount: int64PtrOrNil(b.LagMemberCount),
			LagName:        stringPtrOrNil(b.LagName),
		})
		return port, err
	case len(m.ExistingPort) == 1:
		b := m.ExistingPort[0]
		err := port.FromUserNetworkInterfaceOnExistingPort(portal.UserNetworkInterfaceOnExistingPort{
			Id:        int(b.Id.ValueInt64()),
			Bandwidth: int(b.Bandwidth.ValueInt64()),
			VlanId:    int64PtrOrNil(b.VlanId),
			VlanType:  vlanTypePtr(b.VlanType),
		})
		return port, err
	case len(m.ExistingLag) == 1:
		b := m.ExistingLag[0]
		err := port.FromUserNetworkInterfaceOnExistingLAG(portal.UserNetworkInterfaceOnExistingLAG{
			LagId:     int(b.LagId.ValueInt64()),
			Bandwidth: int(b.Bandwidth.ValueInt64()),
			VlanId:    int64PtrOrNil(b.VlanId),
			VlanType:  vlanTypePtr(b.VlanType),
		})
		return port, err
	}
	return port, fmt.Errorf("exactly one of new_port, existing_port, or existing_lag must be set")
}

// mapIPTransitComputed fills the computed read-back attributes from an IPTransit
// response. Shared by Create and Read.
func mapIPTransitComputed(m *ipTransitResourceModel, s portal.IPTransit) {
	m.Id = optionalInt(s.Id)
	m.Sid = optionalString(s.Sid)
	m.Name = types.StringValue(s.Name)
	m.Status = types.StringValue(string(s.Status))
	m.ServiceSpeed = optionalInt(s.ServiceSpeed)
	m.CustomerGid = types.StringValue(s.CustomerGid)
	if s.EndDate != nil {
		m.EndDate = types.StringValue(s.EndDate.Format("2006-01-02"))
	} else {
		m.EndDate = types.StringNull()
	}
	m.NoticePeriod = optionalInt(s.NoticePeriod)
	m.RenewalPeriod = optionalInt(s.RenewalPeriod)
}

// int64PtrOrNil / stringPtrOrNil / vlanTypePtr convert a known framework value
// to a client pointer, or nil when the value is null or unknown.
func int64PtrOrNil(v types.Int64) *int {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := int(v.ValueInt64())
	return &i
}

func stringPtrOrNil(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

func vlanTypePtr(v types.String) *portal.VlanTypes {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	vt := portal.VlanTypes(v.ValueString())
	return &vt
}
