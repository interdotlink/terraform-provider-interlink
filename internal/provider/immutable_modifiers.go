package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// Immutable attributes error at plan time when changed, rather than forcing a
// replace. A replace would cancel the existing (billed) service and order a new
// one from a one-line diff, never acceptable for this resource. Changing an
// immutable attribute is only possible by deliberately replacing the service
// (cancelling the old one, a contractual/billed action) outside Terraform.
//
// The check allows the value when there is no prior state (a fresh create, or an
// import adopting a field the API cannot return), so those paths are not blocked.
const immutableDetail = "This attribute cannot be changed after the service is created. " +
	"Changing it would require replacing the service, cancelling the existing one " +
	"(a contractual, billed action) and ordering a new one, which this provider will " +
	"never do automatically. Revert this attribute to its previous value."

type immutableStringModifier struct{}

func immutableString() planmodifier.String { return immutableStringModifier{} }

func (m immutableStringModifier) Description(context.Context) string {
	return "Errors if the value changes after creation."
}

func (m immutableStringModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m immutableStringModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// No prior value (create or import adoption), unknown plan value, or no
	// change: nothing to enforce.
	if req.StateValue.IsNull() || req.PlanValue.IsUnknown() || req.StateValue.Equal(req.PlanValue) {
		return
	}
	resp.Diagnostics.AddAttributeError(req.Path, "Attribute is immutable", immutableDetail)
}

type immutableInt64Modifier struct{}

func immutableInt64() planmodifier.Int64 { return immutableInt64Modifier{} }

func (m immutableInt64Modifier) Description(context.Context) string {
	return "Errors if the value changes after creation."
}

func (m immutableInt64Modifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m immutableInt64Modifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if req.StateValue.IsNull() || req.PlanValue.IsUnknown() || req.StateValue.Equal(req.PlanValue) {
		return
	}
	resp.Diagnostics.AddAttributeError(req.Path, "Attribute is immutable", immutableDetail)
}

type immutableBoolModifier struct{}

func immutableBool() planmodifier.Bool { return immutableBoolModifier{} }

func (m immutableBoolModifier) Description(context.Context) string {
	return "Errors if the value changes after creation."
}

func (m immutableBoolModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m immutableBoolModifier) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsUnknown() || req.StateValue.Equal(req.PlanValue) {
		return
	}
	resp.Diagnostics.AddAttributeError(req.Path, "Attribute is immutable", immutableDetail)
}

// Ensure the interfaces are satisfied.
var (
	_ planmodifier.String = immutableStringModifier{}
	_ planmodifier.Int64  = immutableInt64Modifier{}
	_ planmodifier.Bool   = immutableBoolModifier{}
)
