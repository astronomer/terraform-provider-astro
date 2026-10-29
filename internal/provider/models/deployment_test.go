package models_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	platform_v1 "github.com/astronomer/terraform-provider-astro/internal/clients/platform_v1"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/samber/lo"
)

// overrideUntil is the instant 2075-04-25T12:58:00+05:30, which the API returns normalized
// to UTC as 2075-04-25T07:28:00Z.
func overrideUntil(t *testing.T) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, "2075-04-25T07:28:00Z")
	require.NoError(t, err)
	return parsed
}

// fixedNow is the instant the tests measure override expiry against; overrideUntil is after it
// and expiredOverrideUntil before it.
var fixedNow = time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

func expiredOverrideUntil(t *testing.T) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, "2025-12-31T23:00:00Z")
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
		&platform_v1.DeploymentHibernationOverride{OverrideUntil: &until},
		types.StringValue("2075-04-25T12:58:00+05:30"),
		fixedNow,
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
		&platform_v1.DeploymentHibernationOverride{OverrideUntil: &until},
		types.StringValue("2075-04-26T12:58:00+05:30"),
		fixedNow,
	)
	assert.False(t, diags.HasError())
	assert.Equal(t, "2075-04-25T07:28:00Z", readOverrideUntil(t, obj).ValueString())
}

// With nothing configured (import, or the data source), the API value is used as-is.
func TestUnit_HibernationOverrideTypesObject_NoConfiguredValue(t *testing.T) {
	until := overrideUntil(t)

	obj, diags := models.HibernationOverrideTypesObject(
		context.Background(),
		&platform_v1.DeploymentHibernationOverride{OverrideUntil: &until},
		types.StringNull(),
		fixedNow,
	)
	assert.False(t, diags.HasError())
	assert.Equal(t, "2075-04-25T07:28:00Z", readOverrideUntil(t, obj).ValueString())
}

// An unparseable configured value is ignored rather than propagated.
func TestUnit_HibernationOverrideTypesObject_UnparseableConfiguredValue(t *testing.T) {
	until := overrideUntil(t)

	obj, diags := models.HibernationOverrideTypesObject(
		context.Background(),
		&platform_v1.DeploymentHibernationOverride{OverrideUntil: &until},
		types.StringValue("not-a-timestamp"),
		fixedNow,
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
		&platform_v1.DeploymentHibernationOverride{OverrideUntil: &until},
		types.StringValue("2075-04-25T12:58:00+05:30"),
		fixedNow,
	)
	require.False(t, diags.HasError())

	hibernationSpec, diags := types.ObjectValueFrom(ctx, schemas.HibernationSpecAttributeTypes(), models.HibernationSpec{
		Override:  override,
		Schedules: types.SetNull(types.ObjectType{AttrTypes: schemas.HibernationScheduleAttributeTypes()}),
	})
	require.False(t, diags.HasError())

	scalingSpec, diags := types.ObjectValueFrom(ctx, schemas.ScalingSpecAttributeTypes(), models.DeploymentScalingSpec{
		Autoscaling:     types.ObjectNull(schemas.AutoscalingAttributeTypes()),
		HibernationSpec: hibernationSpec,
	})
	require.False(t, diags.HasError())

	assert.Equal(t, "2075-04-25T12:58:00+05:30", models.ConfiguredOverrideUntil(ctx, scalingSpec).ValueString())
}

// Issue #350: the API keeps returning an override after its override_until has passed, with
// is_active = false. The resource must not write it to state, otherwise lifecycle.ignore_changes
// carries it into the next plan and the update request resends a timestamp the API rejects.
func TestUnit_HibernationOverrideTypesObject_ExpiredIsDropped(t *testing.T) {
	until := expiredOverrideUntil(t)

	obj, diags := models.HibernationOverrideTypesObject(
		context.Background(),
		&platform_v1.DeploymentHibernationOverride{IsHibernating: lo.ToPtr(false), IsActive: lo.ToPtr(false), OverrideUntil: &until},
		types.StringNull(),
		fixedNow,
	)
	assert.False(t, diags.HasError())
	assert.True(t, obj.IsNull())
}

// The data source passes the zero time and keeps reporting an expired override, is_active = false
// included.
func TestUnit_HibernationOverrideTypesObject_ExpiredIsKeptForDataSource(t *testing.T) {
	until := expiredOverrideUntil(t)

	obj, diags := models.HibernationOverrideTypesObject(
		context.Background(),
		&platform_v1.DeploymentHibernationOverride{IsHibernating: lo.ToPtr(false), IsActive: lo.ToPtr(false), OverrideUntil: &until},
		types.StringNull(),
		time.Time{},
	)
	assert.False(t, diags.HasError())
	assert.Equal(t, "2025-12-31T23:00:00Z", readOverrideUntil(t, obj).ValueString())
	assert.False(t, obj.Attributes()["is_active"].(types.Bool).ValueBool())
}

// An override that has not ended yet is kept.
func TestUnit_HibernationOverrideTypesObject_FutureIsKept(t *testing.T) {
	until := overrideUntil(t)

	obj, diags := models.HibernationOverrideTypesObject(
		context.Background(),
		&platform_v1.DeploymentHibernationOverride{IsHibernating: lo.ToPtr(true), IsActive: lo.ToPtr(true), OverrideUntil: &until},
		types.StringNull(),
		fixedNow,
	)
	assert.False(t, diags.HasError())
	assert.Equal(t, "2075-04-25T07:28:00Z", readOverrideUntil(t, obj).ValueString())
	assert.True(t, obj.Attributes()["is_active"].(types.Bool).ValueBool())
}

// An override without override_until lasts until it is removed, so it never expires.
func TestUnit_HibernationOverrideTypesObject_NoOverrideUntilIsKept(t *testing.T) {
	obj, diags := models.HibernationOverrideTypesObject(
		context.Background(),
		&platform_v1.DeploymentHibernationOverride{IsHibernating: lo.ToPtr(true), IsActive: lo.ToPtr(true)},
		types.StringNull(),
		fixedNow,
	)
	assert.False(t, diags.HasError())
	require.False(t, obj.IsNull())
	assert.True(t, obj.Attributes()["override_until"].(types.String).IsNull())
}

func TestUnit_HibernationOverrideExpired(t *testing.T) {
	past := fixedNow.Add(-time.Minute)
	future := fixedNow.Add(time.Minute)
	tests := []struct {
		name  string
		until *time.Time
		now   time.Time
		want  bool
	}{
		{name: "no override_until never expires", until: nil, now: fixedNow, want: false},
		{name: "zero now disables the check", until: &past, now: time.Time{}, want: false},
		{name: "before now has expired", until: &past, now: fixedNow, want: true},
		{name: "exactly now has not expired", until: &fixedNow, now: fixedNow, want: false},
		{name: "after now has not expired", until: &future, now: fixedNow, want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, models.HibernationOverrideExpired(tc.until, tc.now))
		})
	}
}

// An expired override with no schedules leaves nothing in the spec, so hibernation_spec is null,
// which is what a configuration without one expects.
func TestUnit_HibernationSpecTypesObject_ExpiredOverrideWithoutSchedulesIsNull(t *testing.T) {
	until := expiredOverrideUntil(t)

	obj, diags := models.HibernationSpecTypesObject(
		context.Background(),
		&platform_v1.DeploymentHibernationSpec{
			Override: &platform_v1.DeploymentHibernationOverride{IsHibernating: lo.ToPtr(false), IsActive: lo.ToPtr(false), OverrideUntil: &until},
		},
		types.StringNull(),
		fixedNow,
	)
	assert.False(t, diags.HasError())
	assert.True(t, obj.IsNull())
}

// The same spec read by the data source keeps the override.
func TestUnit_HibernationSpecTypesObject_ExpiredOverrideWithoutSchedulesKeptForDataSource(t *testing.T) {
	until := expiredOverrideUntil(t)

	obj, diags := models.HibernationSpecTypesObject(
		context.Background(),
		&platform_v1.DeploymentHibernationSpec{
			Override: &platform_v1.DeploymentHibernationOverride{IsHibernating: lo.ToPtr(false), IsActive: lo.ToPtr(false), OverrideUntil: &until},
		},
		types.StringNull(),
		time.Time{},
	)
	assert.False(t, diags.HasError())
	require.False(t, obj.IsNull())
	assert.Equal(t, "2025-12-31T23:00:00Z", readOverrideUntil(t, obj.Attributes()["override"].(types.Object)).ValueString())
}

// The #350 shape: schedules in configuration, an expired override set outside Terraform. The
// schedules stay and the override is null, so a plan carries no override to resend.
func TestUnit_HibernationSpecTypesObject_ExpiredOverrideWithSchedulesKeepsSchedules(t *testing.T) {
	until := expiredOverrideUntil(t)

	obj, diags := models.HibernationSpecTypesObject(
		context.Background(),
		&platform_v1.DeploymentHibernationSpec{
			Override: &platform_v1.DeploymentHibernationOverride{IsHibernating: lo.ToPtr(false), IsActive: lo.ToPtr(false), OverrideUntil: &until},
			Schedules: &[]platform_v1.DeploymentHibernationSchedule{{
				Description:     lo.ToPtr("nightly"),
				HibernateAtCron: "0 20 * * *",
				IsEnabled:       true,
				WakeAtCron:      "0 8 * * *",
			}},
		},
		types.StringNull(),
		fixedNow,
	)
	assert.False(t, diags.HasError())
	require.False(t, obj.IsNull())
	assert.True(t, obj.Attributes()["override"].(types.Object).IsNull())

	schedules := obj.Attributes()["schedules"].(types.Set)
	require.Len(t, schedules.Elements(), 1)
	schedule := schedules.Elements()[0].(types.Object).Attributes()
	assert.Equal(t, "0 20 * * *", schedule["hibernate_at_cron"].(types.String).ValueString())
	assert.Equal(t, "0 8 * * *", schedule["wake_at_cron"].(types.String).ValueString())
}

// ScalingSpecTypesObject is what the resource's ReadFromResponse calls; the clock reaches the
// override through it.
func TestUnit_ScalingSpecTypesObject_DropsExpiredOverride(t *testing.T) {
	until := expiredOverrideUntil(t)

	obj, diags := models.ScalingSpecTypesObject(
		context.Background(),
		&platform_v1.DeploymentScalingSpec{
			HibernationSpec: &platform_v1.DeploymentHibernationSpec{
				Override: &platform_v1.DeploymentHibernationOverride{IsHibernating: lo.ToPtr(false), IsActive: lo.ToPtr(false), OverrideUntil: &until},
				Schedules: &[]platform_v1.DeploymentHibernationSchedule{{
					HibernateAtCron: "0 20 * * *",
					IsEnabled:       true,
					WakeAtCron:      "0 8 * * *",
				}},
			},
		},
		types.StringNull(),
		fixedNow,
	)
	assert.False(t, diags.HasError())
	require.False(t, obj.IsNull())
	hibernationSpec := obj.Attributes()["hibernation_spec"].(types.Object)
	require.False(t, hibernationSpec.IsNull())
	assert.True(t, hibernationSpec.Attributes()["override"].(types.Object).IsNull())
	assert.Len(t, hibernationSpec.Attributes()["schedules"].(types.Set).Elements(), 1)
}

// The worker queue read path must surface pod_ephemeral_storage, otherwise a value set in
// config never lands in state and every plan shows drift.
func TestUnit_WorkerQueueTypesObject_ReadsPodEphemeralStorage(t *testing.T) {
	ctx := context.Background()
	workerQueue := platform_v1.WorkerQueue{
		Id:                  "clt1y2z3a4b5c6d7e8f9g0h1i",
		Name:                "default",
		AstroMachine:        lo.ToPtr("A5"),
		IsDefault:           true,
		MaxWorkerCount:      10,
		MinWorkerCount:      0,
		PodCpu:              "1",
		PodMemory:           "2Gi",
		PodEphemeralStorage: lo.ToPtr("20Gi"),
		WorkerConcurrency:   5,
	}

	resourceObj, diags := models.WorkerQueueResourceTypesObject(ctx, workerQueue)
	require.False(t, diags.HasError())
	assert.Equal(t, "20Gi", resourceObj.Attributes()["pod_ephemeral_storage"].(types.String).ValueString())

	dataSourceObj, diags := models.WorkerQueueDataSourceTypesObject(ctx, workerQueue)
	require.False(t, diags.HasError())
	assert.Equal(t, "20Gi", dataSourceObj.Attributes()["pod_ephemeral_storage"].(types.String).ValueString())
}

// A worker queue the API reports without ephemeral storage leaves the attribute null rather
// than inventing an empty string.
func TestUnit_WorkerQueueTypesObject_NullPodEphemeralStorage(t *testing.T) {
	obj, diags := models.WorkerQueueResourceTypesObject(context.Background(), platform_v1.WorkerQueue{
		Id:                "clt1y2z3a4b5c6d7e8f9g0h1i",
		Name:              "default",
		AstroMachine:      lo.ToPtr("A5"),
		IsDefault:         true,
		MaxWorkerCount:    10,
		MinWorkerCount:    0,
		PodCpu:            "1",
		PodMemory:         "2Gi",
		WorkerConcurrency: 5,
	})
	require.False(t, diags.HasError())
	assert.True(t, obj.Attributes()["pod_ephemeral_storage"].(types.String).IsNull())
}

func TestUnit_ScalingSpecTypesObject_Autoscaling(t *testing.T) {
	ctx := context.Background()

	t.Run("disabled autoscaling keeps its bounds", func(t *testing.T) {
		obj, diags := models.ScalingSpecTypesObject(ctx, &platform_v1.DeploymentScalingSpec{
			Autoscaling: &platform_v1.AutoscalingSpec{
				ApiServer: &platform_v1.ScalingSpec{IsEnabled: false, MinReplicas: 2, MaxReplicas: 5},
			},
		}, types.StringNull(), fixedNow)
		require.False(t, diags.HasError())

		var scalingSpec models.DeploymentScalingSpec
		require.False(t, obj.As(ctx, &scalingSpec, basetypes.ObjectAsOptions{}).HasError())
		assert.True(t, scalingSpec.HibernationSpec.IsNull())

		var autoscaling models.Autoscaling
		require.False(t, scalingSpec.Autoscaling.As(ctx, &autoscaling, basetypes.ObjectAsOptions{}).HasError())
		var apiServer models.ReplicaScaling
		require.False(t, autoscaling.ApiServer.As(ctx, &apiServer, basetypes.ObjectAsOptions{}).HasError())
		assert.False(t, apiServer.IsEnabled.ValueBool())
		assert.Equal(t, int64(2), apiServer.MinReplicas.ValueInt64())
		assert.Equal(t, int64(5), apiServer.MaxReplicas.ValueInt64())
	})

	t.Run("absent autoscaling is null", func(t *testing.T) {
		obj, diags := models.ScalingSpecTypesObject(ctx, &platform_v1.DeploymentScalingSpec{}, types.StringNull(), fixedNow)
		require.False(t, diags.HasError())

		var scalingSpec models.DeploymentScalingSpec
		require.False(t, obj.As(ctx, &scalingSpec, basetypes.ObjectAsOptions{}).HasError())
		assert.True(t, scalingSpec.Autoscaling.IsNull())
	})
}

func TestUnit_ManagedScalingSpec(t *testing.T) {
	ctx := context.Background()

	apiAutoscaling := &platform_v1.AutoscalingSpec{
		ApiServer: &platform_v1.ScalingSpec{IsEnabled: true, MinReplicas: 2, MaxReplicas: 5},
	}
	apiHibernation := &platform_v1.DeploymentHibernationSpec{
		Schedules: &[]platform_v1.DeploymentHibernationSchedule{{
			HibernateAtCron: "0 20 * * *",
			IsEnabled:       true,
			WakeAtCron:      "0 8 * * *",
		}},
	}

	priorWith := func(t *testing.T, autoscaling types.Object) types.Object {
		t.Helper()
		obj, diags := types.ObjectValueFrom(ctx, schemas.ScalingSpecAttributeTypes(), models.DeploymentScalingSpec{
			Autoscaling:     autoscaling,
			HibernationSpec: types.ObjectNull(schemas.HibernationSpecAttributeTypes()),
		})
		require.False(t, diags.HasError())
		return obj
	}
	managedAutoscaling, diags := types.ObjectValueFrom(ctx, schemas.AutoscalingAttributeTypes(), models.Autoscaling{
		ApiServer: types.ObjectNull(schemas.ReplicaScalingAttributeTypes()),
	})
	require.False(t, diags.HasError())

	t.Run("autoscaling the config does not manage is ignored", func(t *testing.T) {
		got := models.ManagedScalingSpec(ctx, types.ObjectNull(schemas.ScalingSpecAttributeTypes()), &platform_v1.DeploymentScalingSpec{Autoscaling: apiAutoscaling})
		assert.Nil(t, got)
	})

	t.Run("hibernation is kept when autoscaling is ignored", func(t *testing.T) {
		got := models.ManagedScalingSpec(ctx, priorWith(t, types.ObjectNull(schemas.AutoscalingAttributeTypes())), &platform_v1.DeploymentScalingSpec{
			Autoscaling:     apiAutoscaling,
			HibernationSpec: apiHibernation,
		})
		require.NotNil(t, got)
		assert.Nil(t, got.Autoscaling)
		assert.NotNil(t, got.HibernationSpec)
	})

	t.Run("autoscaling is kept once the config manages it", func(t *testing.T) {
		got := models.ManagedScalingSpec(ctx, priorWith(t, managedAutoscaling), &platform_v1.DeploymentScalingSpec{Autoscaling: apiAutoscaling})
		require.NotNil(t, got)
		assert.Equal(t, apiAutoscaling, got.Autoscaling)
	})

	t.Run("no scaling spec stays nil", func(t *testing.T) {
		assert.Nil(t, models.ManagedScalingSpec(ctx, priorWith(t, managedAutoscaling), nil))
	})
}
