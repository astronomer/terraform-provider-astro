package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeploymentResourceSchemaAttributes_SchedulerAuIsOptionalAndComputed(t *testing.T) {
	attributes := DeploymentResourceSchemaAttributes()

	schedulerAu, ok := attributes["scheduler_au"]
	assert.True(t, ok, "scheduler_au attribute should exist in the deployment resource schema")
	assert.True(t, schedulerAu.IsOptional(), "scheduler_au should be Optional")
	assert.True(t, schedulerAu.IsComputed(), "scheduler_au should be Computed to avoid drift/inconsistent-result errors when omitted from config, e.g. after import")
}

// pod_ephemeral_storage must be Optional so Celery/Astro executor worker queues can override the
// platform default, and Computed so omitting it does not conflict with the value the API returns.
func TestDeploymentWorkerQueueResourceSchemaAttributes_PodEphemeralStorageIsOptionalAndComputed(t *testing.T) {
	attributes := WorkerQueueResourceSchemaAttributes()

	podEphemeralStorage, ok := attributes["pod_ephemeral_storage"]
	assert.True(t, ok, "pod_ephemeral_storage attribute should exist in the worker queue resource schema")
	assert.True(t, podEphemeralStorage.IsOptional(), "pod_ephemeral_storage should be Optional")
	assert.True(t, podEphemeralStorage.IsComputed(), "pod_ephemeral_storage should be Computed so the platform default is accepted when omitted")
}

func TestDeploymentWorkerQueueDataSourceSchemaAttributes_PodEphemeralStorageIsComputed(t *testing.T) {
	attributes := WorkerQueueDataSourceSchemaAttributes()

	podEphemeralStorage, ok := attributes["pod_ephemeral_storage"]
	assert.True(t, ok, "pod_ephemeral_storage attribute should exist in the worker queue data source schema")
	assert.True(t, podEphemeralStorage.IsComputed(), "pod_ephemeral_storage should be Computed")
}

// The attribute-type maps are what types.ObjectValueFrom validates the models against, so a
// schema attribute that is missing from them fails at runtime rather than at compile time.
func TestDeploymentWorkerQueueAttributeTypes_IncludePodEphemeralStorage(t *testing.T) {
	_, ok := WorkerQueueResourceAttributeTypes()["pod_ephemeral_storage"]
	assert.True(t, ok, "pod_ephemeral_storage should exist in the worker queue resource attribute types")

	_, ok = WorkerQueueDataSourceAttributeTypes()["pod_ephemeral_storage"]
	assert.True(t, ok, "pod_ephemeral_storage should exist in the worker queue data source attribute types")
}

// Every attribute in the resource/data source schemas must have a matching entry in the
// attribute-type map, otherwise conversions in the models package fail at runtime.
func TestDeploymentWorkerQueueSchemaAndAttributeTypesStayInSync(t *testing.T) {
	resourceAttributeTypes := WorkerQueueResourceAttributeTypes()
	for name := range WorkerQueueResourceSchemaAttributes() {
		_, ok := resourceAttributeTypes[name]
		assert.True(t, ok, "worker queue resource attribute %q is missing from WorkerQueueResourceAttributeTypes", name)
	}
	assert.Len(t, resourceAttributeTypes, len(WorkerQueueResourceSchemaAttributes()))

	dataSourceAttributeTypes := WorkerQueueDataSourceAttributeTypes()
	for name := range WorkerQueueDataSourceSchemaAttributes() {
		_, ok := dataSourceAttributeTypes[name]
		assert.True(t, ok, "worker queue data source attribute %q is missing from WorkerQueueDataSourceAttributeTypes", name)
	}
	assert.Len(t, dataSourceAttributeTypes, len(WorkerQueueDataSourceSchemaAttributes()))
}
