package main

import (
	"strings"
	"testing"

	platform_v1 "github.com/astronomer/terraform-provider-astro/internal/clients/platform_v1"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func hostedWorkerQueue(podEphemeralStorage *string) []platform_v1.WorkerQueue {
	return []platform_v1.WorkerQueue{{
		Id:                  "clt1y2z3a4b5c6d7e8f9g0h1i",
		Name:                "default",
		AstroMachine:        lo.ToPtr("A5"),
		IsDefault:           true,
		MaxWorkerCount:      10,
		MinWorkerCount:      0,
		PodCpu:              "1",
		PodMemory:           "2Gi",
		PodEphemeralStorage: podEphemeralStorage,
		WorkerConcurrency:   5,
	}}
}

// A Celery/Astro worker queue that overrides ephemeral storage must carry that override into
// the generated config, otherwise importing a Deployment silently drops the setting.
func TestFormatWorkerQueues_EmitsPodEphemeralStorage(t *testing.T) {
	got := formatWorkerQueues(
		lo.ToPtr(hostedWorkerQueue(lo.ToPtr("20Gi"))),
		lo.ToPtr("CELERY"),
		lo.ToPtr(string(platform_v1.DeploymentTypeSTANDARD)),
	)

	assert.Contains(t, got, `pod_ephemeral_storage = "20Gi"`)
}

// A queue on the platform default must not have the value pinned into generated config.
func TestFormatWorkerQueues_OmitsUnsetPodEphemeralStorage(t *testing.T) {
	got := formatWorkerQueues(
		lo.ToPtr(hostedWorkerQueue(nil)),
		lo.ToPtr("CELERY"),
		lo.ToPtr(string(platform_v1.DeploymentTypeSTANDARD)),
	)

	assert.NotContains(t, got, "pod_ephemeral_storage")
}

// Hybrid worker queues do not support ephemeral storage, so the attribute must never appear in
// generated hybrid config even if the API happens to report a value.
func TestFormatWorkerQueues_NeverEmitsPodEphemeralStorageForHybrid(t *testing.T) {
	queues := hostedWorkerQueue(lo.ToPtr("20Gi"))
	queues[0].AstroMachine = nil
	queues[0].NodePoolId = lo.ToPtr("clnp86ly5000301ndzfxz895w")

	got := formatWorkerQueues(
		lo.ToPtr(queues),
		lo.ToPtr("CELERY"),
		lo.ToPtr(string(platform_v1.DeploymentTypeHYBRID)),
	)

	assert.NotContains(t, got, "pod_ephemeral_storage")
	assert.True(t, strings.Contains(got, "node_pool_id"), "hybrid worker queues are keyed by node_pool_id")
}
