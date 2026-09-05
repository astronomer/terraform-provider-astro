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

func TestDeploymentResourceSchemaAttributes_DrWorkloadIdentity(t *testing.T) {
	attributes := DeploymentResourceSchemaAttributes()

	desired, ok := attributes["desired_dr_workload_identity"]
	assert.True(t, ok, "desired_dr_workload_identity attribute should exist in the deployment resource schema")
	assert.True(t, desired.IsOptional(), "desired_dr_workload_identity should be Optional, mirroring desired_workload_identity")
	assert.False(t, desired.IsComputed(), "desired_dr_workload_identity should not be Computed, mirroring desired_workload_identity")

	effective, ok := attributes["dr_workload_identity"]
	assert.True(t, ok, "dr_workload_identity attribute should exist in the deployment resource schema")
	assert.True(t, effective.IsComputed(), "dr_workload_identity should be Computed, mirroring workload_identity")
	assert.False(t, effective.IsOptional(), "dr_workload_identity should not be Optional, mirroring workload_identity")
}

func TestDeploymentDataSourceSchemaAttributes_DrWorkloadIdentity(t *testing.T) {
	attributes := DeploymentDataSourceSchemaAttributes()

	effective, ok := attributes["dr_workload_identity"]
	assert.True(t, ok, "dr_workload_identity attribute should exist in the deployment data source schema")
	assert.True(t, effective.IsComputed(), "dr_workload_identity should be Computed")
}

// The deployments data source builds its objects from DeploymentsElementAttributeTypes, which has
// to stay in step with the single-deployment data source schema. A key present in one and missing
// from the other surfaces as a type mismatch at plan time rather than a compile error.
func TestDeploymentsElementAttributeTypes_MatchesDeploymentDataSourceSchema(t *testing.T) {
	schemaAttributes := DeploymentDataSourceSchemaAttributes()
	elementTypes := DeploymentsElementAttributeTypes()

	for name := range schemaAttributes {
		_, ok := elementTypes[name]
		assert.True(t, ok, "attribute %q is in the deployment data source schema but missing from DeploymentsElementAttributeTypes", name)
	}
	for name := range elementTypes {
		_, ok := schemaAttributes[name]
		assert.True(t, ok, "attribute %q is in DeploymentsElementAttributeTypes but missing from the deployment data source schema", name)
	}
}
