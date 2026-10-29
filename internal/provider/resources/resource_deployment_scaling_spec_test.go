package resources

import (
	"context"
	"testing"
	"time"

	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// scalingSpecNow is the instant RequestScalingSpec validates override_until against in these tests.
var scalingSpecNow = time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

// overrideObject builds a scaling_spec.hibernation_spec.override object as it arrives from a plan.
// is_active is Computed: unknown for an override that is new in the plan, a known value when the
// object was carried over from the prior state.
func overrideObject(t *testing.T, isActive types.Bool, isHibernating bool, overrideUntil types.String) types.Object {
	t.Helper()
	obj, diags := types.ObjectValue(schemas.HibernationOverrideAttributeTypes(), map[string]attr.Value{
		"is_active":      isActive,
		"is_hibernating": types.BoolValue(isHibernating),
		"override_until": overrideUntil,
	})
	require.False(t, diags.HasError())
	return obj
}

func nullOverride() types.Object {
	return types.ObjectNull(schemas.HibernationOverrideAttributeTypes())
}

func nullSchedules() types.Set {
	return types.SetNull(types.ObjectType{AttrTypes: schemas.HibernationScheduleAttributeTypes()})
}

// nightlySchedules builds a schedules set with the single schedule from issue #350.
func nightlySchedules(t *testing.T) types.Set {
	t.Helper()
	schedule, diags := types.ObjectValue(schemas.HibernationScheduleAttributeTypes(), map[string]attr.Value{
		"description":       types.StringValue("nightly"),
		"hibernate_at_cron": types.StringValue("0 20 * * *"),
		"is_enabled":        types.BoolValue(true),
		"wake_at_cron":      types.StringValue("0 8 * * *"),
	})
	require.False(t, diags.HasError())
	set, diags := types.SetValue(types.ObjectType{AttrTypes: schemas.HibernationScheduleAttributeTypes()}, []attr.Value{schedule})
	require.False(t, diags.HasError())
	return set
}

// hibernationScalingSpecObject wraps an override and schedules into a scaling_spec object; either may be null.
func hibernationScalingSpecObject(t *testing.T, override types.Object, schedules types.Set) types.Object {
	t.Helper()
	hibernationSpec, diags := types.ObjectValue(schemas.HibernationSpecAttributeTypes(), map[string]attr.Value{
		"override":  override,
		"schedules": schedules,
	})
	require.False(t, diags.HasError())
	scalingSpec, diags := types.ObjectValue(schemas.ScalingSpecAttributeTypes(), map[string]attr.Value{
		"autoscaling":      types.ObjectNull(schemas.AutoscalingAttributeTypes()),
		"hibernation_spec": hibernationSpec,
	})
	require.False(t, diags.HasError())
	return scalingSpec
}

func TestUnit_RequestScalingSpec_FutureOverrideUntilIsSent(t *testing.T) {
	spec := hibernationScalingSpecObject(t, overrideObject(t, types.BoolUnknown(), true, types.StringValue("2075-01-01T00:00:00Z")), nullSchedules())

	req, diags := RequestScalingSpec(context.Background(), spec, scalingSpecNow)
	require.False(t, diags.HasError(), "unexpected diags: %v", diags)
	require.NotNil(t, req.HibernationSpec)
	require.NotNil(t, req.HibernationSpec.Override)
	require.NotNil(t, req.HibernationSpec.Override.OverrideUntil)
	assert.Equal(t, "2075-01-01T00:00:00Z", *req.HibernationSpec.Override.OverrideUntil)
	require.NotNil(t, req.HibernationSpec.Override.IsHibernating)
	assert.True(t, *req.HibernationSpec.Override.IsHibernating)
	assert.Nil(t, req.HibernationSpec.Schedules)
}

// Issue #350: an override whose override_until has passed is rejected by the API with
// "Override until time must be in the future" (400). It is reported before the request is sent,
// against the attribute, rather than silently dropped: the API replaces hibernationSpec wholesale,
// so dropping it would clear the override and leave the applied state inconsistent with the plan.
func TestUnit_RequestScalingSpec_ExpiredOverrideUntilIsAnError(t *testing.T) {
	// The shape ignore_changes produces: the override object comes from the prior state, so
	// is_active is known (and false, as the API reports it once the override has ended).
	spec := hibernationScalingSpecObject(t, overrideObject(t, types.BoolValue(false), false, types.StringValue("2025-12-31T23:00:00Z")), nightlySchedules(t))

	req, diags := RequestScalingSpec(context.Background(), spec, scalingSpecNow)
	require.True(t, diags.HasError())
	assert.Nil(t, req)
	require.Len(t, diags, 1)

	assert.Equal(t, "scaling_spec.hibernation_spec.override.override_until must be in the future", diags[0].Summary())
	assert.Contains(t, diags[0].Detail(), "2025-12-31T23:00:00Z")
	assert.Contains(t, diags[0].Detail(), "remove the override")

	withPath, ok := diags[0].(diag.DiagnosticWithPath)
	require.True(t, ok, "diagnostic should carry the attribute path")
	assert.True(t, withPath.Path().Equal(path.Root("scaling_spec").AtName("hibernation_spec").AtName("override").AtName("override_until")))
}

// An override that ends exactly now has not expired yet; the boundary belongs to the API.
func TestUnit_RequestScalingSpec_OverrideUntilAtNowIsSent(t *testing.T) {
	spec := hibernationScalingSpecObject(t, overrideObject(t, types.BoolUnknown(), true, types.StringValue(scalingSpecNow.Format(time.RFC3339))), nullSchedules())

	req, diags := RequestScalingSpec(context.Background(), spec, scalingSpecNow)
	require.False(t, diags.HasError(), "unexpected diags: %v", diags)
	require.NotNil(t, req.HibernationSpec.Override)
	require.NotNil(t, req.HibernationSpec.Override.OverrideUntil)
	assert.Equal(t, scalingSpecNow.Format(time.RFC3339), *req.HibernationSpec.Override.OverrideUntil)
}

// An override with no override_until lasts until it is removed; nothing to validate.
func TestUnit_RequestScalingSpec_OverrideWithoutOverrideUntilIsSent(t *testing.T) {
	spec := hibernationScalingSpecObject(t, overrideObject(t, types.BoolUnknown(), true, types.StringNull()), nullSchedules())

	req, diags := RequestScalingSpec(context.Background(), spec, scalingSpecNow)
	require.False(t, diags.HasError(), "unexpected diags: %v", diags)
	require.NotNil(t, req.HibernationSpec.Override)
	assert.Nil(t, req.HibernationSpec.Override.OverrideUntil)
	require.NotNil(t, req.HibernationSpec.Override.IsHibernating)
	assert.True(t, *req.HibernationSpec.Override.IsHibernating)
}

// A value that is not RFC3339 is passed through unchanged; the API validates the format.
func TestUnit_RequestScalingSpec_UnparseableOverrideUntilIsPassedThrough(t *testing.T) {
	spec := hibernationScalingSpecObject(t, overrideObject(t, types.BoolUnknown(), true, types.StringValue("not-a-timestamp")), nullSchedules())

	req, diags := RequestScalingSpec(context.Background(), spec, scalingSpecNow)
	require.False(t, diags.HasError(), "unexpected diags: %v", diags)
	require.NotNil(t, req.HibernationSpec.Override)
	require.NotNil(t, req.HibernationSpec.Override.OverrideUntil)
	assert.Equal(t, "not-a-timestamp", *req.HibernationSpec.Override.OverrideUntil)
}

// With no override in the plan, none is sent and the schedules go through as they are. After Read
// drops an expired override from state this is the request the #350 configuration produces.
func TestUnit_RequestScalingSpec_SchedulesOnlyOmitsOverride(t *testing.T) {
	spec := hibernationScalingSpecObject(t, nullOverride(), nightlySchedules(t))

	req, diags := RequestScalingSpec(context.Background(), spec, scalingSpecNow)
	require.False(t, diags.HasError(), "unexpected diags: %v", diags)
	require.NotNil(t, req.HibernationSpec)
	assert.Nil(t, req.HibernationSpec.Override)
	require.NotNil(t, req.HibernationSpec.Schedules)
	require.Len(t, *req.HibernationSpec.Schedules, 1)
	assert.Equal(t, "0 20 * * *", (*req.HibernationSpec.Schedules)[0].HibernateAtCron)
	assert.Equal(t, "0 8 * * *", (*req.HibernationSpec.Schedules)[0].WakeAtCron)
}

func TestUnit_RequestScalingSpec_RequiresOverrideOrSchedules(t *testing.T) {
	spec := hibernationScalingSpecObject(t, nullOverride(), nullSchedules())

	_, diags := RequestScalingSpec(context.Background(), spec, scalingSpecNow)
	require.True(t, diags.HasError())
	require.Len(t, diags, 1)
	assert.Equal(t, "scaling_spec.hibernation_spec must have either override or schedules", diags[0].Summary())
}
