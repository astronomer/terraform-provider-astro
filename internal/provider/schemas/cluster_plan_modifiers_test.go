package schemas

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

// isDrEnabledFailedOverSchema is a minimal schema covering just the two attributes
// nullWhenDrDisabledBoolPlanModifier reads/writes, enough to exercise it directly
// without building the full cluster resource schema.
func isDrEnabledFailedOverSchema() rschema.Schema {
	return rschema.Schema{
		Attributes: map[string]rschema.Attribute{
			"is_dr_enabled":  rschema.BoolAttribute{Optional: true, Computed: true},
			"is_failed_over": rschema.BoolAttribute{Optional: true, Computed: true},
		},
	}
}

func isDrEnabledFailedOverObjectType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"is_dr_enabled":  tftypes.Bool,
		"is_failed_over": tftypes.Bool,
	}}
}

// TestNullWhenDrDisabledBoolPlanModifier_DrNewlyEnabled_LeavesUnknown is a regression
// test for a "Provider produced inconsistent result after apply: .is_failed_over: was
// null, but now cty.False" failure seen enabling DR on an existing Azure cluster: when
// DR transitions from disabled to enabled, the prior state's is_failed_over (null,
// because DR was off) must not be reused as the plan value, since the API will return
// a real true/false once DR is actually enabled.
func TestNullWhenDrDisabledBoolPlanModifier_DrNewlyEnabled_LeavesUnknown(t *testing.T) {
	ctx := context.Background()
	s := isDrEnabledFailedOverSchema()
	ot := isDrEnabledFailedOverObjectType()

	stateRaw := tftypes.NewValue(ot, map[string]tftypes.Value{
		"is_dr_enabled":  tftypes.NewValue(tftypes.Bool, false),
		"is_failed_over": tftypes.NewValue(tftypes.Bool, nil),
	})
	planRaw := tftypes.NewValue(ot, map[string]tftypes.Value{
		"is_dr_enabled":  tftypes.NewValue(tftypes.Bool, true),
		"is_failed_over": tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
	})

	req := planmodifier.BoolRequest{
		Path:        path.Root("is_failed_over"),
		State:       tfsdk.State{Schema: s, Raw: stateRaw},
		Plan:        tfsdk.Plan{Schema: s, Raw: planRaw},
		ConfigValue: types.BoolNull(),
		StateValue:  types.BoolNull(),
		PlanValue:   types.BoolUnknown(),
	}
	resp := &planmodifier.BoolResponse{PlanValue: req.PlanValue}

	nullWhenDrDisabledBoolPlanModifier{}.PlanModifyBool(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.True(t, resp.PlanValue.IsUnknown(), "expected plan value to stay unknown when DR is newly enabled, got %v", resp.PlanValue)
}

// TestNullWhenDrDisabledBoolPlanModifier_DrAlreadyEnabled_UsesState verifies the
// UseStateForUnknown-like fallback still applies when DR was already enabled (no
// disabled->enabled transition), so unrelated updates to an already-DR-enabled
// cluster don't show a spurious diff on is_failed_over.
func TestNullWhenDrDisabledBoolPlanModifier_DrAlreadyEnabled_UsesState(t *testing.T) {
	ctx := context.Background()
	s := isDrEnabledFailedOverSchema()
	ot := isDrEnabledFailedOverObjectType()

	stateRaw := tftypes.NewValue(ot, map[string]tftypes.Value{
		"is_dr_enabled":  tftypes.NewValue(tftypes.Bool, true),
		"is_failed_over": tftypes.NewValue(tftypes.Bool, false),
	})
	planRaw := tftypes.NewValue(ot, map[string]tftypes.Value{
		"is_dr_enabled":  tftypes.NewValue(tftypes.Bool, true),
		"is_failed_over": tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
	})

	req := planmodifier.BoolRequest{
		Path:        path.Root("is_failed_over"),
		State:       tfsdk.State{Schema: s, Raw: stateRaw},
		Plan:        tfsdk.Plan{Schema: s, Raw: planRaw},
		ConfigValue: types.BoolNull(),
		StateValue:  types.BoolValue(false),
		PlanValue:   types.BoolUnknown(),
	}
	resp := &planmodifier.BoolResponse{PlanValue: req.PlanValue}

	nullWhenDrDisabledBoolPlanModifier{}.PlanModifyBool(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.False(t, resp.PlanValue.IsUnknown())
	assert.False(t, resp.PlanValue.IsNull())
	assert.False(t, resp.PlanValue.ValueBool())
}

// TestNullWhenDrDisabledBoolPlanModifier_DrDisabling_SetsNull confirms disabling DR
// still nulls out is_failed_over regardless of the prior state.
func TestNullWhenDrDisabledBoolPlanModifier_DrDisabling_SetsNull(t *testing.T) {
	ctx := context.Background()
	s := isDrEnabledFailedOverSchema()
	ot := isDrEnabledFailedOverObjectType()

	stateRaw := tftypes.NewValue(ot, map[string]tftypes.Value{
		"is_dr_enabled":  tftypes.NewValue(tftypes.Bool, true),
		"is_failed_over": tftypes.NewValue(tftypes.Bool, false),
	})
	planRaw := tftypes.NewValue(ot, map[string]tftypes.Value{
		"is_dr_enabled":  tftypes.NewValue(tftypes.Bool, false),
		"is_failed_over": tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
	})

	req := planmodifier.BoolRequest{
		Path:        path.Root("is_failed_over"),
		State:       tfsdk.State{Schema: s, Raw: stateRaw},
		Plan:        tfsdk.Plan{Schema: s, Raw: planRaw},
		ConfigValue: types.BoolNull(),
		StateValue:  types.BoolValue(false),
		PlanValue:   types.BoolUnknown(),
	}
	resp := &planmodifier.BoolResponse{PlanValue: req.PlanValue}

	nullWhenDrDisabledBoolPlanModifier{}.PlanModifyBool(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.True(t, resp.PlanValue.IsNull())
}

// TestNullWhenDrDisabledBoolPlanModifier_Create verifies a brand new resource (no
// prior state) leaves the plan value unknown to be computed, rather than erroring.
func TestNullWhenDrDisabledBoolPlanModifier_Create(t *testing.T) {
	ctx := context.Background()
	s := isDrEnabledFailedOverSchema()
	ot := isDrEnabledFailedOverObjectType()

	planRaw := tftypes.NewValue(ot, map[string]tftypes.Value{
		"is_dr_enabled":  tftypes.NewValue(tftypes.Bool, true),
		"is_failed_over": tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
	})

	req := planmodifier.BoolRequest{
		Path:        path.Root("is_failed_over"),
		State:       tfsdk.State{Schema: s, Raw: tftypes.NewValue(ot, nil)},
		Plan:        tfsdk.Plan{Schema: s, Raw: planRaw},
		ConfigValue: types.BoolNull(),
		StateValue:  types.BoolNull(),
		PlanValue:   types.BoolUnknown(),
	}
	resp := &planmodifier.BoolResponse{PlanValue: req.PlanValue}

	nullWhenDrDisabledBoolPlanModifier{}.PlanModifyBool(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.True(t, resp.PlanValue.IsUnknown())
}
