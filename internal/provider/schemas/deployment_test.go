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
