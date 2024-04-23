package datasources_test

import (
	"fmt"
	"strconv"
	"testing"

	astronomerprovider "github.com/astronomer/astronomer-terraform-provider/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_DataSourceDeploymentOptionsHosted(t *testing.T) {
	resourceName := "test_hosted"
	resourceVar := fmt.Sprintf("data.astronomer_deployment_options.%v", resourceName)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, true) + deploymentOptions(resourceName, ""),
				Check: resource.ComposeTestCheckFunc(
					CheckDeploymentOptions(resourceVar)...,
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, true) + deploymentOptions(resourceName, `deployment_type = "STANDARD"`),
				Check: resource.ComposeTestCheckFunc(
					CheckDeploymentOptions(resourceVar)...,
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, true) + deploymentOptions(resourceName, `deployment_type = "DEDICATED"`),
				Check: resource.ComposeTestCheckFunc(
					CheckDeploymentOptions(resourceVar)...,
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, true) + deploymentOptions(resourceName, `executor = "CELERY"`),
				Check: resource.ComposeTestCheckFunc(
					CheckDeploymentOptions(resourceVar)...,
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, true) + deploymentOptions(resourceName, `executor = "KUBERNETES"`),
				Check: resource.ComposeTestCheckFunc(
					CheckDeploymentOptions(resourceVar)...,
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, true) + deploymentOptions(resourceName, `cloud_provider = "AWS"`),
				Check: resource.ComposeTestCheckFunc(
					CheckDeploymentOptions(resourceVar)...,
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, true) + deploymentOptions(resourceName, `cloud_provider = "GCP"`),
				Check: resource.ComposeTestCheckFunc(
					CheckDeploymentOptions(resourceVar)...,
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, true) + deploymentOptions(resourceName, `cloud_provider = "AZURE"`),
				Check: resource.ComposeTestCheckFunc(
					CheckDeploymentOptions(resourceVar)...,
				),
			},
		},
	})
}

func TestAcc_DataSourceDeploymentOptionsHybrid(t *testing.T) {
	resourceName := "test_hybrid"
	resourceVar := fmt.Sprintf("data.astronomer_deployment_options.%v", resourceName)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, false) + deploymentOptions(resourceName, ""),
				Check: resource.ComposeTestCheckFunc(
					CheckDeploymentOptions(resourceVar)...,
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, false) + deploymentOptions(resourceName, `deployment_type = "HYBRID"`),
				Check: resource.ComposeTestCheckFunc(
					CheckDeploymentOptions(resourceVar)...,
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, false) + deploymentOptions(resourceName, `executor = "CELERY"`),
				Check: resource.ComposeTestCheckFunc(
					CheckDeploymentOptions(resourceVar)...,
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, false) + deploymentOptions(resourceName, `executor = "KUBERNETES"`),
				Check: resource.ComposeTestCheckFunc(
					CheckDeploymentOptions(resourceVar)...,
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, false) + deploymentOptions(resourceName, `cloud_provider = "AWS"`),
				Check: resource.ComposeTestCheckFunc(
					CheckDeploymentOptions(resourceVar)...,
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, false) + deploymentOptions(resourceName, `cloud_provider = "GCP"`),
				Check: resource.ComposeTestCheckFunc(
					CheckDeploymentOptions(resourceVar)...,
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, false) + deploymentOptions(resourceName, `cloud_provider = "AZURE"`),
				Check: resource.ComposeTestCheckFunc(
					CheckDeploymentOptions(resourceVar)...,
				),
			},
		},
	})
}

func CheckDeploymentOptions(resourceVar string) []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(resourceVar, "executors.#", "2"),

		resource.TestCheckResourceAttrSet(resourceVar, "resource_quotas.resource_quota.cpu.floor"),
		resource.TestCheckResourceAttrSet(resourceVar, "resource_quotas.resource_quota.cpu.default"),
		resource.TestCheckResourceAttrSet(resourceVar, "resource_quotas.resource_quota.cpu.ceiling"),
		resource.TestCheckResourceAttrSet(resourceVar, "resource_quotas.resource_quota.memory.floor"),
		resource.TestCheckResourceAttrSet(resourceVar, "resource_quotas.resource_quota.memory.default"),
		resource.TestCheckResourceAttrSet(resourceVar, "resource_quotas.resource_quota.memory.ceiling"),
		resource.TestCheckResourceAttrSet(resourceVar, "resource_quotas.default_pod_size.cpu.floor"),
		resource.TestCheckResourceAttrSet(resourceVar, "resource_quotas.default_pod_size.cpu.default"),
		resource.TestCheckResourceAttrSet(resourceVar, "resource_quotas.default_pod_size.cpu.ceiling"),
		resource.TestCheckResourceAttrSet(resourceVar, "resource_quotas.default_pod_size.memory.floor"),
		resource.TestCheckResourceAttrSet(resourceVar, "resource_quotas.default_pod_size.memory.default"),
		resource.TestCheckResourceAttrSet(resourceVar, "resource_quotas.default_pod_size.memory.ceiling"),

		resource.TestCheckResourceAttrWith(resourceVar, "runtime_releases.#", CheckAttributeLengthIsNotEmpty),
		resource.TestCheckResourceAttrWith(resourceVar, "scheduler_machines.#", CheckAttributeLengthIsNotEmpty),
		resource.TestCheckResourceAttrWith(resourceVar, "worker_machines.#", CheckAttributeLengthIsNotEmpty),

		resource.TestCheckResourceAttrSet(resourceVar, "worker_queues.max_workers.ceiling"),
		resource.TestCheckResourceAttrSet(resourceVar, "worker_queues.max_workers.default"),
		resource.TestCheckResourceAttrSet(resourceVar, "worker_queues.max_workers.floor"),
		resource.TestCheckResourceAttrSet(resourceVar, "worker_queues.min_workers.ceiling"),
		resource.TestCheckResourceAttrSet(resourceVar, "worker_queues.min_workers.default"),
		resource.TestCheckResourceAttrSet(resourceVar, "worker_queues.min_workers.floor"),
		resource.TestCheckResourceAttrSet(resourceVar, "worker_queues.worker_concurrency.ceiling"),
		resource.TestCheckResourceAttrSet(resourceVar, "worker_queues.worker_concurrency.default"),
		resource.TestCheckResourceAttrSet(resourceVar, "worker_queues.worker_concurrency.floor"),
	}
}

func CheckAttributeLengthIsNotEmpty(value string) error {
	if value == "" {
		return fmt.Errorf("expected value to be non-empty")
	}
	parseInt, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	if parseInt == 0 {
		return fmt.Errorf("expected value to be non-zero")
	}
	return nil
}

func deploymentOptions(tfVarName, queryParams string) string {
	return fmt.Sprintf(`
data astronomer_deployment_options "%v" {
	  %v
}`, tfVarName, queryParams)
}
