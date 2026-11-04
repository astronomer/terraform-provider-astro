package planmodifiers_test

import (
	"context"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/provider/planmodifiers"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

// resourceType is a stand-in for a resource with one computed attribute.
var resourceType = tftypes.Object{AttributeTypes: map[string]tftypes.Type{"image_version": tftypes.String}}

func rawNull() tftypes.Value {
	return tftypes.NewValue(resourceType, nil)
}

func rawSet() tftypes.Value {
	return tftypes.NewValue(resourceType, map[string]tftypes.Value{
		"image_version": tftypes.NewValue(tftypes.String, "3.3-11"),
	})
}

func TestUnknownStringOnUpdate(t *testing.T) {
	priorValue := types.StringValue("3.3-11")

	testCases := map[string]struct {
		state       tftypes.Value
		plan        tftypes.Value
		expectKnown bool
	}{
		"update plans the value as unknown": {
			state:       rawSet(),
			plan:        rawSet(),
			expectKnown: false,
		},
		"create is left to the framework": {
			state:       rawNull(),
			plan:        rawSet(),
			expectKnown: true,
		},
		"destroy is left alone": {
			state:       rawSet(),
			plan:        rawNull(),
			expectKnown: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			req := planmodifier.StringRequest{
				State:      tfsdk.State{Raw: testCase.state},
				Plan:       tfsdk.Plan{Raw: testCase.plan},
				PlanValue:  priorValue,
				StateValue: priorValue,
			}
			resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}

			planmodifiers.UnknownStringOnUpdate().PlanModifyString(context.Background(), req, resp)

			if testCase.expectKnown {
				assert.False(t, resp.PlanValue.IsUnknown())
				assert.Equal(t, priorValue, resp.PlanValue)
				return
			}
			assert.True(t, resp.PlanValue.IsUnknown())
		})
	}
}

func TestUnknownObjectOnUpdate(t *testing.T) {
	attributeTypes := map[string]attr.Type{"next_event_at": types.StringType}
	priorValue := types.ObjectValueMust(attributeTypes, map[string]attr.Value{
		"next_event_at": types.StringValue("2026-10-29T17:59:00Z"),
	})

	testCases := map[string]struct {
		state       tftypes.Value
		plan        tftypes.Value
		planValue   types.Object
		expectKnown bool
	}{
		"update plans the object as unknown": {
			state:       rawSet(),
			plan:        rawSet(),
			planValue:   priorValue,
			expectKnown: false,
		},
		"a null object still becomes unknown on update": {
			state:       rawSet(),
			plan:        rawSet(),
			planValue:   types.ObjectNull(attributeTypes),
			expectKnown: false,
		},
		"create is left to the framework": {
			state:       rawNull(),
			plan:        rawSet(),
			planValue:   priorValue,
			expectKnown: true,
		},
		"destroy is left alone": {
			state:       rawSet(),
			plan:        rawNull(),
			planValue:   priorValue,
			expectKnown: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			req := planmodifier.ObjectRequest{
				State:      tfsdk.State{Raw: testCase.state},
				Plan:       tfsdk.Plan{Raw: testCase.plan},
				PlanValue:  testCase.planValue,
				StateValue: priorValue,
			}
			resp := &planmodifier.ObjectResponse{PlanValue: req.PlanValue}

			planmodifiers.UnknownObjectOnUpdate().PlanModifyObject(context.Background(), req, resp)

			if testCase.expectKnown {
				assert.False(t, resp.PlanValue.IsUnknown())
				assert.Equal(t, testCase.planValue, resp.PlanValue)
				return
			}
			assert.True(t, resp.PlanValue.IsUnknown())
			// The object keeps its shape so dependent references still type-check.
			assert.Equal(t, attributeTypes, resp.PlanValue.AttributeTypes(context.Background()))
		})
	}
}
