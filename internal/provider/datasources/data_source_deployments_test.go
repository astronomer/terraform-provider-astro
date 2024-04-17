package datasources_test

import (
	"fmt"
	"testing"

	"github.com/lucsky/cuid"

	astronomerprovider "github.com/astronomer/astronomer-terraform-provider/internal/provider"
	"github.com/astronomer/astronomer-terraform-provider/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAcc_DataSourceDeployments(t *testing.T) {
	deploymentName := utils.GenerateTestResourceName(10)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			//Check the data source for deployments for a hosted organization
			{
				Config: astronomerprovider.ProviderConfig(t, true) + hostedDeployments(deploymentName),
				Check: resource.ComposeTestCheckFunc(
					// Doing all checks in one step because we do not want to unnecessarily create multiple deployments for the data sources test

					// These checks are for the deployment data source (singular)
					resource.TestCheckResourceAttrSet("data.astronomer_deployment.test_data_deployment_kubernetes", "id"),
					resource.TestCheckResourceAttr("data.astronomer_deployment.test_data_deployment_kubernetes", "name", fmt.Sprintf("%v-1", deploymentName)),
					resource.TestCheckResourceAttrSet("data.astronomer_deployment.test_data_deployment_kubernetes", "description"),
					resource.TestCheckResourceAttrSet("data.astronomer_deployment.test_data_deployment_kubernetes", "workspace_id"),
					resource.TestCheckResourceAttrSet("data.astronomer_deployment.test_data_deployment_kubernetes", "created_by.id"),
					resource.TestCheckResourceAttrSet("data.astronomer_deployment.test_data_deployment_kubernetes", "created_at"),
					resource.TestCheckResourceAttrSet("data.astronomer_deployment.test_data_deployment_kubernetes", "updated_by.id"),
					resource.TestCheckResourceAttrSet("data.astronomer_deployment.test_data_deployment_kubernetes", "updated_at"),
					resource.TestCheckResourceAttr("data.astronomer_deployment.test_data_deployment_kubernetes", "region", "us-east4"),
					resource.TestCheckResourceAttr("data.astronomer_deployment.test_data_deployment_kubernetes", "cloud_provider", "GCP"),
					resource.TestCheckResourceAttrSet("data.astronomer_deployment.test_data_deployment_kubernetes", "astro_runtime_version"),
					resource.TestCheckResourceAttrSet("data.astronomer_deployment.test_data_deployment_kubernetes", "airflow_version"),
					resource.TestCheckResourceAttrSet("data.astronomer_deployment.test_data_deployment_kubernetes", "namespace"),
					resource.TestCheckResourceAttr("data.astronomer_deployment.test_data_deployment_kubernetes", "contact_emails.0", "preview@astronomer.test"),
					resource.TestCheckResourceAttr("data.astronomer_deployment.test_data_deployment_kubernetes", "executor", "KUBERNETES"),
					resource.TestCheckNoResourceAttr("data.astronomer_deployment.test_data_deployment_kubernetes", "worker_queues"),
					resource.TestCheckResourceAttrSet("data.astronomer_deployment.test_data_deployment_kubernetes", "scheduler_replicas"),
					resource.TestCheckResourceAttrSet("data.astronomer_deployment.test_data_deployment_kubernetes", "image_tag"),
					resource.TestCheckResourceAttrSet("data.astronomer_deployment.test_data_deployment_kubernetes", "image_repository"),
					resource.TestCheckResourceAttr("data.astronomer_deployment.test_data_deployment_kubernetes", "environment_variables.0.key", "key1"),
					resource.TestCheckResourceAttr("data.astronomer_deployment.test_data_deployment_kubernetes", "environment_variables.0.value", "value1"),
					resource.TestCheckResourceAttr("data.astronomer_deployment.test_data_deployment_kubernetes", "environment_variables.0.is_secret", "false"),
					resource.TestCheckResourceAttrSet("data.astronomer_deployment.test_data_deployment_kubernetes", "webserver_ingress_hostname"),
					resource.TestCheckResourceAttrSet("data.astronomer_deployment.test_data_deployment_kubernetes", "webserver_url"),
					resource.TestCheckResourceAttrSet("data.astronomer_deployment.test_data_deployment_kubernetes", "webserver_airflow_api_url"),
					resource.TestCheckResourceAttrSet("data.astronomer_deployment.test_data_deployment_kubernetes", "status"),
					resource.TestCheckResourceAttr("data.astronomer_deployment.test_data_deployment_kubernetes", "is_cicd_enforced", "true"),
					resource.TestCheckResourceAttr("data.astronomer_deployment.test_data_deployment_kubernetes", "type", "STANDARD"),
					resource.TestCheckResourceAttr("data.astronomer_deployment.test_data_deployment_kubernetes", "is_dag_deploy_enabled", "true"),
					resource.TestCheckResourceAttr("data.astronomer_deployment.test_data_deployment_kubernetes", "scheduler_size", "SMALL"),
					resource.TestCheckResourceAttr("data.astronomer_deployment.test_data_deployment_kubernetes", "is_high_availability", "true"),
					resource.TestCheckResourceAttr("data.astronomer_deployment.test_data_deployment_kubernetes", "is_development_mode", "false"),
					resource.TestCheckResourceAttrSet("data.astronomer_deployment.test_data_deployment_kubernetes", "workload_identity"),
					resource.TestCheckResourceAttrSet("data.astronomer_deployment.test_data_deployment_kubernetes", "external_ips.0"),
					resource.TestCheckResourceAttr("data.astronomer_deployment.test_data_deployment_kubernetes", "resource_quota_cpu", "10"),
					resource.TestCheckResourceAttr("data.astronomer_deployment.test_data_deployment_kubernetes", "resource_quota_memory", "20Gi"),
					resource.TestCheckResourceAttr("data.astronomer_deployment.test_data_deployment_kubernetes", "default_task_pod_cpu", "0.25"),
					resource.TestCheckResourceAttr("data.astronomer_deployment.test_data_deployment_kubernetes", "default_task_pod_memory", "0.5Gi"),

					resource.TestCheckResourceAttr("data.astronomer_deployment.test_data_deployment_celery", "executor", "CELERY"),
					resource.TestCheckResourceAttr("data.astronomer_deployment.test_data_deployment_celery", "worker_queues.0.name", "default"),

					// These checks are for the deployments data source (plural)
					checkDeployments("test_data_deployments_no_filters", deploymentName+"-1"),
					checkDeployments("test_data_deployments_no_filters", deploymentName+"-2"),
					checkDeployments("test_data_deployments_workspace_ids_filter", deploymentName+"-1"),
					checkDeployments("test_data_deployments_workspace_ids_filter", deploymentName+"-2"),
					checkDeployments("test_data_deployments_deployment_ids_filter", deploymentName+"-1"),
					checkDeployments("test_data_deployments_names_filter", deploymentName+"-1"),
					checkDeploymentsAreEmpty("test_data_deployments_incorrect_deployment_ids_filter"),
					checkDeploymentsAreEmpty("test_data_deployments_incorrect_deployment_ids_filter"),
					checkDeploymentsAreEmpty("test_data_deployments_incorrect_names_filter"),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			//Check the data source for deployments for a hybrid organization
			{
				Config: astronomerprovider.ProviderConfig(t, false) + hybridDeployments(),
				Check: resource.ComposeTestCheckFunc(
					// Checks that the deployments data source is not empty and checks the first deployment in the list
					// has some of the expected attributes
					checkDeployments("test_data_deployments_hybrid_no_filters", ""),
				),
			},
		},
	})
}

func hybridDeployments() string {
	return `
data astronomer_deployments "test_data_deployments_hybrid_no_filters" {}`
}

func hostedDeployments(name string) string {
	return fmt.Sprintf(`
resource "astronomer_workspace" "test_workspace" {
	name = "%v"
	description = "%v"
	cicd_enforced_default = true
}

resource "astronomer_deployment" "test_deployment_kubernetes" {
	name = "%v-1"
	description = "%v"
	type = "STANDARD"
	region = "us-east4"
	cloud_provider = "GCP"
	contact_emails = ["preview@astronomer.test"]
	default_task_pod_cpu = "0.25"
	default_task_pod_memory = "0.5Gi"
	executor = "KUBERNETES"
	is_cicd_enforced = true
	is_dag_deploy_enabled = true
	is_development_mode = false
	is_high_availability = true
	resource_quota_cpu = "10"
	resource_quota_memory = "20Gi"
	scheduler_size = "SMALL"
	workspace_id = astronomer_workspace.test_workspace.id
	environment_variables = [{
		key = "key1"
		value = "value1"
		is_secret = false
	}]
}

resource "astronomer_deployment" "test_deployment_celery" {
	name = "%v-2"
	description = "%v"
	type = "STANDARD"
	region = "us-east-1"
	cloud_provider = "AWS"
	contact_emails = []
	default_task_pod_cpu = "0.25"
	default_task_pod_memory = "0.5Gi"
	executor = "CELERY"
	is_cicd_enforced = true
	is_dag_deploy_enabled = true
	is_development_mode = false
	is_high_availability = false
	resource_quota_cpu = "10"
	resource_quota_memory = "20Gi"
	scheduler_size = "SMALL"
	workspace_id = astronomer_workspace.test_workspace.id
	environment_variables = []
	worker_queues = [{
		name = "default"
		is_default = true
		astro_machine = "A5"
		max_worker_count = 10
		min_worker_count = 0
		worker_concurrency = 1
	}]
}

data astronomer_deployment "test_data_deployment_kubernetes" {
	depends_on = [astronomer_deployment.test_deployment_kubernetes]
	id = astronomer_deployment.test_deployment_kubernetes.id
}

data astronomer_deployment "test_data_deployment_celery" {
	depends_on = [astronomer_deployment.test_deployment_celery]
	id = astronomer_deployment.test_deployment_celery.id
}

data astronomer_deployments "test_data_deployments_no_filters" {
	depends_on = [astronomer_deployment.test_deployment_kubernetes, astronomer_deployment.test_deployment_celery]
}

data astronomer_deployments "test_data_deployments_workspace_ids_filter" {
	depends_on = [astronomer_deployment.test_deployment_kubernetes, astronomer_deployment.test_deployment_celery]
	workspace_ids = [astronomer_workspace.test_workspace.id]
}

data astronomer_deployments "test_data_deployments_deployment_ids_filter" {
	depends_on = [astronomer_deployment.test_deployment_kubernetes, astronomer_deployment.test_deployment_celery]
	deployment_ids = [astronomer_deployment.test_deployment_kubernetes.id]
}

data astronomer_deployments "test_data_deployments_names_filter" {
	depends_on = [astronomer_deployment.test_deployment_kubernetes, astronomer_deployment.test_deployment_celery]
	names = ["%v-1"]
}

data astronomer_deployments "test_data_deployments_incorrect_workspace_ids_filter" {
	depends_on = [astronomer_deployment.test_deployment_kubernetes, astronomer_deployment.test_deployment_celery]
	workspace_ids = ["%v"]
}

data astronomer_deployments "test_data_deployments_incorrect_deployment_ids_filter" {
	depends_on = [astronomer_deployment.test_deployment_kubernetes, astronomer_deployment.test_deployment_celery]
	deployment_ids = ["%v"]
}

data astronomer_deployments "test_data_deployments_incorrect_names_filter" {
	depends_on = [astronomer_deployment.test_deployment_kubernetes, astronomer_deployment.test_deployment_celery]
	names = ["%v"]
}
`, name, utils.TestResourceDescription, name, utils.TestResourceDescription, name, utils.TestResourceDescription, name, cuid.New(), cuid.New(), cuid.New())
}

func checkDeploymentsAreEmpty(tfVarName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		instanceState, numDeployments, err := utils.GetDataSourcesLength(s, tfVarName, "deployments")
		if err != nil {
			return err
		}
		if numDeployments != 0 {
			return fmt.Errorf("expected deployments to be 0, got %s", instanceState.Attributes["deployments.#"])
		}
		return nil
	}
}

func checkDeployments(tfVarName, deploymentName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		instanceState, numDeployments, err := utils.GetDataSourcesLength(s, tfVarName, "deployments")
		if err != nil {
			return err
		}
		if numDeployments == 0 {
			return fmt.Errorf("expected deployments to be greater or equal to 1, got %s", instanceState.Attributes["deployments.#"])
		}

		// If deploymentName is not set, we will check the first deployment
		var deploymentIdx int
		if deploymentName == "" {
			deploymentIdx = 0
		} else {
			for i := 0; i < numDeployments; i++ {
				idxName := fmt.Sprintf("deployments.%d.name", i)
				if instanceState.Attributes[idxName] == deploymentName {
					deploymentIdx = i
					break
				}
			}
			if deploymentIdx == -1 {
				return fmt.Errorf("deployment %s not found", deploymentName)
			}
		}

		description := fmt.Sprintf("deployments.%d.description", deploymentIdx)
		if instanceState.Attributes[description] == "" {
			return fmt.Errorf("expected 'description' to be set")
		}
		createdAt := fmt.Sprintf("deployments.%d.created_at", deploymentIdx)
		if instanceState.Attributes[createdAt] == "" {
			return fmt.Errorf("expected 'created_at' to be set")
		}
		updatedAt := fmt.Sprintf("deployments.%d.updated_at", deploymentIdx)
		if instanceState.Attributes[updatedAt] == "" {
			return fmt.Errorf("expected 'updated_at' to be set")
		}
		createdById := fmt.Sprintf("deployments.%d.created_by.id", deploymentIdx)
		if instanceState.Attributes[createdById] == "" {
			return fmt.Errorf("expected 'created_by.id' to be set")
		}
		updatedById := fmt.Sprintf("deployments.%d.updated_by.id", deploymentIdx)
		if instanceState.Attributes[updatedById] == "" {
			return fmt.Errorf("expected 'updated_by.id' to be set")
		}
		workspaceId := fmt.Sprintf("deployments.%d.workspace_id", deploymentIdx)
		if instanceState.Attributes[workspaceId] == "" {
			return fmt.Errorf("expected 'workspace_id' to be set")
		}
		astroRuntimeVersion := fmt.Sprintf("deployments.%d.astro_runtime_version", deploymentIdx)
		if instanceState.Attributes[astroRuntimeVersion] == "" {
			return fmt.Errorf("expected 'astro_runtime_version' to be set")
		}
		airflowVersion := fmt.Sprintf("deployments.%d.airflow_version", deploymentIdx)
		if instanceState.Attributes[airflowVersion] == "" {
			return fmt.Errorf("expected 'airflow_version' to be set")
		}
		namespace := fmt.Sprintf("deployments.%d.namespace", deploymentIdx)
		if instanceState.Attributes[namespace] == "" {
			return fmt.Errorf("expected 'namespace' to be set")
		}
		executor := fmt.Sprintf("deployments.%d.executor", deploymentIdx)
		if instanceState.Attributes[executor] == "" {
			return fmt.Errorf("expected 'executor' to be set")
		} else if instanceState.Attributes[executor] == "KUBERNETES" {
			workerQueues := fmt.Sprintf("deployments.%d.worker_queues", deploymentIdx)
			if instanceState.Attributes[workerQueues] != "" {
				return fmt.Errorf("expected 'worker_queues' to be empty")
			}
		} else if instanceState.Attributes[executor] == "CELERY" {
			workerQueues := fmt.Sprintf("deployments.%d.worker_queues.0.name", deploymentIdx)
			if instanceState.Attributes[workerQueues] == "" {
				return fmt.Errorf("expected 'worker_queues.0.name' to be set")
			}
		}
		typ := fmt.Sprintf("deployments.%d.type", deploymentIdx)
		if instanceState.Attributes[typ] == "" {
			return fmt.Errorf("expected 'type' to be set")
		}
		isCiCdEnforced := fmt.Sprintf("deployments.%d.is_cicd_enforced", deploymentIdx)
		if instanceState.Attributes[isCiCdEnforced] == "" {
			return fmt.Errorf("expected 'is_cicd_enforced' to be set")
		}

		return nil
	}
}
