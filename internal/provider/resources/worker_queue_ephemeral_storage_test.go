package resources

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func workerQueueSet(t *testing.T, podEphemeralStorage types.String) types.Set {
	t.Helper()
	obj, diags := types.ObjectValueFrom(context.Background(), schemas.WorkerQueueResourceAttributeTypes(), models.WorkerQueueResource{
		Name:                types.StringValue("default"),
		AstroMachine:        types.StringValue("A5"),
		IsDefault:           types.BoolValue(true),
		MaxWorkerCount:      types.Int64Value(10),
		MinWorkerCount:      types.Int64Value(0),
		NodePoolId:          types.StringNull(),
		PodCpu:              types.StringValue("1"),
		PodMemory:           types.StringValue("2Gi"),
		PodEphemeralStorage: podEphemeralStorage,
		WorkerConcurrency:   types.Int64Value(5),
	})
	require.False(t, diags.HasError())

	set, diags := types.SetValue(types.ObjectType{AttrTypes: schemas.WorkerQueueResourceAttributeTypes()}, []attr.Value{obj})
	require.False(t, diags.HasError())
	return set
}

// A configured pod_ephemeral_storage reaches the create request for Celery/Astro worker queues.
func TestUnit_RequestHostedWorkerQueues_SendsConfiguredPodEphemeralStorage(t *testing.T) {
	queues, diags := RequestHostedWorkerQueues(context.Background(), workerQueueSet(t, types.StringValue("20Gi")))
	require.False(t, diags.HasError())
	require.NotNil(t, queues)
	require.Len(t, *queues, 1)

	require.NotNil(t, (*queues)[0].PodEphemeralStorage)
	assert.Equal(t, "20Gi", *(*queues)[0].PodEphemeralStorage)
}

// Leaving pod_ephemeral_storage out must omit it from the request so the platform default for
// the astro_machine applies, rather than pinning whatever placeholder the plan carried.
func TestUnit_RequestHostedWorkerQueues_OmitsUnsetPodEphemeralStorage(t *testing.T) {
	for name, value := range map[string]types.String{
		"null":    types.StringNull(),
		"unknown": types.StringUnknown(),
	} {
		t.Run(name, func(t *testing.T) {
			queues, diags := RequestHostedWorkerQueues(context.Background(), workerQueueSet(t, value))
			require.False(t, diags.HasError())
			require.NotNil(t, queues)
			require.Len(t, *queues, 1)

			assert.Nil(t, (*queues)[0].PodEphemeralStorage)
		})
	}
}

func TestUnit_RequestHostedUpdateWorkerQueues_SendsConfiguredPodEphemeralStorage(t *testing.T) {
	queues, diags := RequestHostedUpdateWorkerQueues(context.Background(), workerQueueSet(t, types.StringValue("20Gi")))
	require.False(t, diags.HasError())
	require.NotNil(t, queues)
	require.Len(t, *queues, 1)

	require.NotNil(t, (*queues)[0].PodEphemeralStorage)
	assert.Equal(t, "20Gi", *(*queues)[0].PodEphemeralStorage)
}

func TestUnit_RequestHostedUpdateWorkerQueues_OmitsUnsetPodEphemeralStorage(t *testing.T) {
	queues, diags := RequestHostedUpdateWorkerQueues(context.Background(), workerQueueSet(t, types.StringUnknown()))
	require.False(t, diags.HasError())
	require.NotNil(t, queues)
	require.Len(t, *queues, 1)

	assert.Nil(t, (*queues)[0].PodEphemeralStorage)
}

// Hybrid worker queues are sized by their node pool and the v1 hybrid create payload has no
// ephemeral storage field, so the value must never be smuggled into a hybrid request.
func TestUnit_RequestHybridUpdateWorkerQueues_NeverSendsPodEphemeralStorage(t *testing.T) {
	queues, diags := RequestHybridUpdateWorkerQueues(context.Background(), workerQueueSet(t, types.StringValue("20Gi")))
	require.False(t, diags.HasError())
	require.NotNil(t, queues)
	require.Len(t, *queues, 1)

	assert.Nil(t, (*queues)[0].PodEphemeralStorage)
}
