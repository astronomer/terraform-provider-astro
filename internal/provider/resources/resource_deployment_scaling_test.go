package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
)

func scalingSpecObject(t *testing.T, autoscaling types.Object) types.Object {
	t.Helper()
	obj, diags := types.ObjectValueFrom(context.Background(), schemas.ScalingSpecAttributeTypes(), models.DeploymentScalingSpec{
		Autoscaling:     autoscaling,
		HibernationSpec: types.ObjectNull(schemas.HibernationSpecAttributeTypes()),
	})
	require.False(t, diags.HasError())
	return obj
}

func TestUnit_RequestScalingSpec_Autoscaling(t *testing.T) {
	ctx := context.Background()

	apiServer, diags := types.ObjectValueFrom(ctx, schemas.ReplicaScalingAttributeTypes(), models.ReplicaScaling{
		IsEnabled:   types.BoolValue(false),
		MaxReplicas: types.Int64Value(5),
		MinReplicas: types.Int64Value(2),
	})
	require.False(t, diags.HasError())
	autoscaling, diags := types.ObjectValueFrom(ctx, schemas.AutoscalingAttributeTypes(), models.Autoscaling{ApiServer: apiServer})
	require.False(t, diags.HasError())

	t.Run("autoscaling without hibernation is sent as-is", func(t *testing.T) {
		req, diags := RequestScalingSpec(ctx, scalingSpecObject(t, autoscaling))
		require.False(t, diags.HasError())
		require.NotNil(t, req.Autoscaling)
		require.NotNil(t, req.Autoscaling.ApiServer)
		assert.False(t, req.Autoscaling.ApiServer.IsEnabled)
		assert.Equal(t, 2, req.Autoscaling.ApiServer.MinReplicas)
		assert.Equal(t, 5, req.Autoscaling.ApiServer.MaxReplicas)
		assert.Nil(t, req.HibernationSpec)
	})

	t.Run("absent autoscaling is omitted", func(t *testing.T) {
		req, diags := RequestScalingSpec(ctx, scalingSpecObject(t, types.ObjectNull(schemas.AutoscalingAttributeTypes())))
		require.False(t, diags.HasError())
		assert.Nil(t, req.Autoscaling)
	})
}
