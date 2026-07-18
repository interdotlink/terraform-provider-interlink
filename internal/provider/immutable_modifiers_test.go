package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The immutable modifiers allow the value when there is no prior state (a fresh
// create or an import adopting a field the API cannot return), when the planned
// value is unknown, or when it is unchanged; they error only on an actual change
// to an existing value.

func TestImmutableStringModifier(t *testing.T) {
	cases := []struct {
		name    string
		state   types.String
		plan    types.String
		wantErr bool
	}{
		{"null state (create or import)", types.StringNull(), types.StringValue("new"), false},
		{"unknown plan", types.StringValue("old"), types.StringUnknown(), false},
		{"unchanged", types.StringValue("same"), types.StringValue("same"), false},
		{"changed", types.StringValue("old"), types.StringValue("new"), true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp := &planmodifier.StringResponse{PlanValue: c.plan}
			immutableString().PlanModifyString(context.Background(), planmodifier.StringRequest{
				Path:       path.Root("test"),
				StateValue: c.state,
				PlanValue:  c.plan,
			}, resp)
			if got := resp.Diagnostics.HasError(); got != c.wantErr {
				t.Errorf("HasError() = %v, want %v", got, c.wantErr)
			}
		})
	}
}

func TestImmutableInt64Modifier(t *testing.T) {
	cases := []struct {
		name    string
		state   types.Int64
		plan    types.Int64
		wantErr bool
	}{
		{"null state (create or import)", types.Int64Null(), types.Int64Value(2), false},
		{"unknown plan", types.Int64Value(1), types.Int64Unknown(), false},
		{"unchanged", types.Int64Value(1), types.Int64Value(1), false},
		{"changed", types.Int64Value(1), types.Int64Value(2), true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp := &planmodifier.Int64Response{PlanValue: c.plan}
			immutableInt64().PlanModifyInt64(context.Background(), planmodifier.Int64Request{
				Path:       path.Root("test"),
				StateValue: c.state,
				PlanValue:  c.plan,
			}, resp)
			if got := resp.Diagnostics.HasError(); got != c.wantErr {
				t.Errorf("HasError() = %v, want %v", got, c.wantErr)
			}
		})
	}
}

func TestImmutableBoolModifier(t *testing.T) {
	cases := []struct {
		name    string
		state   types.Bool
		plan    types.Bool
		wantErr bool
	}{
		{"null state (create or import)", types.BoolNull(), types.BoolValue(true), false},
		{"unknown plan", types.BoolValue(true), types.BoolUnknown(), false},
		{"unchanged", types.BoolValue(true), types.BoolValue(true), false},
		{"changed", types.BoolValue(true), types.BoolValue(false), true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp := &planmodifier.BoolResponse{PlanValue: c.plan}
			immutableBool().PlanModifyBool(context.Background(), planmodifier.BoolRequest{
				Path:       path.Root("test"),
				StateValue: c.state,
				PlanValue:  c.plan,
			}, resp)
			if got := resp.Diagnostics.HasError(); got != c.wantErr {
				t.Errorf("HasError() = %v, want %v", got, c.wantErr)
			}
		})
	}
}
