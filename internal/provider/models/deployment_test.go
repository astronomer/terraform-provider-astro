package models_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// overrideUntil is the instant 2075-04-25T12:58:00+05:30, which the API returns normalized
// to UTC as 2075-04-25T07:28:00Z.
func overrideUntil(t *testing.T) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, "2075-04-25T07:28:00Z")
	require.NoError(t, err)
	return parsed
}

func readOverrideUntil(t *testing.T, obj types.Object) types.String {
	t.Helper()
	require.False(t, obj.IsNull(), "override object should not be null")
	return obj.Attributes()["override_until"].(types.String)
}

// The API normalizes override_until to UTC. When the configured value denotes the same
// instant in a different offset, the configured spelling must be preserved, otherwise
// Terraform rejects the apply as an inconsistent result.
func TestUnit_HibernationOverrideTypesObject_PreservesConfiguredOffset(t *testing.T) {
	until := overrideUntil(t)

	obj, diags := models.HibernationOverrideTypesObject(
		context.Background(),
		&platform.DeploymentHibernationOverride{OverrideUntil: &until},
		types.StringValue("2075-04-25T12:58:00+05:30"),
	)
	assert.False(t, diags.HasError())
	assert.Equal(t, "2075-04-25T12:58:00+05:30", readOverrideUntil(t, obj).ValueString())
}

// A configured value denoting a different instant must not be preserved — the API value wins
// so real drift is still reported.
func TestUnit_HibernationOverrideTypesObject_DifferentInstantUsesApiValue(t *testing.T) {
	until := overrideUntil(t)

	obj, diags := models.HibernationOverrideTypesObject(
		context.Background(),
		&platform.DeploymentHibernationOverride{OverrideUntil: &until},
		types.StringValue("2075-04-26T12:58:00+05:30"),
	)
	assert.False(t, diags.HasError())
	assert.Equal(t, "2075-04-25T07:28:00Z", readOverrideUntil(t, obj).ValueString())
}

// With nothing configured (import, or the data source), the API value is used as-is.
func TestUnit_HibernationOverrideTypesObject_NoConfiguredValue(t *testing.T) {
	until := overrideUntil(t)

	obj, diags := models.HibernationOverrideTypesObject(
		context.Background(),
		&platform.DeploymentHibernationOverride{OverrideUntil: &until},
		types.StringNull(),
	)
	assert.False(t, diags.HasError())
	assert.Equal(t, "2075-04-25T07:28:00Z", readOverrideUntil(t, obj).ValueString())
}

// An unparseable configured value is ignored rather than propagated.
func TestUnit_HibernationOverrideTypesObject_UnparseableConfiguredValue(t *testing.T) {
	until := overrideUntil(t)

	obj, diags := models.HibernationOverrideTypesObject(
		context.Background(),
		&platform.DeploymentHibernationOverride{OverrideUntil: &until},
		types.StringValue("not-a-timestamp"),
	)
	assert.False(t, diags.HasError())
	assert.Equal(t, "2075-04-25T07:28:00Z", readOverrideUntil(t, obj).ValueString())
}

// ConfiguredOverrideUntil returns a null string at every level where the nesting stops,
// rather than erroring, so callers can pass the result through unconditionally.
func TestUnit_ConfiguredOverrideUntil_NullScalingSpec(t *testing.T) {
	got := models.ConfiguredOverrideUntil(
		context.Background(),
		types.ObjectNull(schemas.ScalingSpecAttributeTypes()),
	)
	assert.True(t, got.IsNull())
}

// A fully populated scaling spec yields the configured override_until.
func TestUnit_ConfiguredOverrideUntil_RoundTrip(t *testing.T) {
	ctx := context.Background()
	until := overrideUntil(t)

	override, diags := models.HibernationOverrideTypesObject(
		ctx,
		&platform.DeploymentHibernationOverride{OverrideUntil: &until},
		types.StringValue("2075-04-25T12:58:00+05:30"),
	)
	require.False(t, diags.HasError())

	hibernationSpec, diags := types.ObjectValueFrom(ctx, schemas.HibernationSpecAttributeTypes(), models.HibernationSpec{
		Override:  override,
		Schedules: types.SetNull(types.ObjectType{AttrTypes: schemas.HibernationScheduleAttributeTypes()}),
	})
	require.False(t, diags.HasError())

	scalingSpec, diags := types.ObjectValueFrom(ctx, schemas.ScalingSpecAttributeTypes(), models.DeploymentScalingSpec{
		HibernationSpec: hibernationSpec,
	})
	require.False(t, diags.HasError())

	assert.Equal(t, "2075-04-25T12:58:00+05:30", models.ConfiguredOverrideUntil(ctx, scalingSpec).ValueString())
}
