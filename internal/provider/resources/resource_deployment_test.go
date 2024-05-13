package resources_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
)

// We will test dedicated deployment resources once dedicated_cluster_resource is implemented

func TestAcc_ResourceDeploymentHybrid(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)

	clusterId := os.Getenv("HYBRID_CLUSTER_ID")
	nodePoolId := os.Getenv("HYBRID_NODE_POOL_ID")
	deploymentName := fmt.Sprintf("%v_hybrid", namePrefix)
	resourceVar := fmt.Sprintf("astro_deployment.%v", deploymentName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			// Check that deployments have been removed
			testAccCheckDeploymentExistence(t, deploymentName, false, false),
		),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, false) + hybridDeployment(hybridDeploymentInput{
					Name:                        deploymentName,
					Description:                 utils.TestResourceDescription,
					ClusterId:                   clusterId,
					Executor:                    "KUBERNETES",
					IncludeEnvironmentVariables: true,
					SchedulerAu:                 5,
					NodePoolId:                  nodePoolId,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "name", deploymentName),
					resource.TestCheckResourceAttr(resourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(resourceVar, "cluster_id", clusterId),
					resource.TestCheckResourceAttrSet(resourceVar, "region"),
					resource.TestCheckResourceAttrSet(resourceVar, "cloud_provider"),
					resource.TestCheckResourceAttr(resourceVar, "executor", "KUBERNETES"),
					resource.TestCheckNoResourceAttr(resourceVar, "worker_queues"),
					resource.TestCheckResourceAttr(resourceVar, "scheduler_au", "5"),
					resource.TestCheckResourceAttrSet(resourceVar, "environment_variables.0.key"),
					// Check via API that deployment exists
					testAccCheckDeploymentExistence(t, deploymentName, false, true),
				),
			},
			// Change properties and check they have been updated in terraform state including executor change
			{
				Config: astronomerprovider.ProviderConfig(t, false) + hybridDeployment(hybridDeploymentInput{
					Name:                        deploymentName,
					Description:                 utils.TestResourceDescription,
					ClusterId:                   clusterId,
					Executor:                    "CELERY",
					IncludeEnvironmentVariables: false,
					SchedulerAu:                 6,
					NodePoolId:                  nodePoolId,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(resourceVar, "worker_queues.0.name", "default"),
					resource.TestCheckResourceAttr(resourceVar, "environment_variables.#", "0"),
					resource.TestCheckResourceAttr(resourceVar, "executor", "CELERY"),
					resource.TestCheckResourceAttr(resourceVar, "scheduler_au", "6"),
					// Check via API that deployment exists
					testAccCheckDeploymentExistence(t, deploymentName, false, true),
				),
			},
			// Change executor back to KUBERNETES and check it is correctly updated in terraform state
			{
				Config: astronomerprovider.ProviderConfig(t, false) + hybridDeployment(hybridDeploymentInput{
					Name:                        deploymentName,
					Description:                 utils.TestResourceDescription,
					ClusterId:                   clusterId,
					Executor:                    "KUBERNETES",
					SchedulerAu:                 6,
					IncludeEnvironmentVariables: false,
					NodePoolId:                  nodePoolId,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "executor", "KUBERNETES"),
					resource.TestCheckNoResourceAttr(resourceVar, "worker_queues"),
					// Check via API that deployment exists
					testAccCheckDeploymentExistence(t, deploymentName, false, true),
				),
			},
			// Import existing deployment and check it is correctly imported - https://stackoverflow.com/questions/68824711/how-can-i-test-terraform-import-in-acceptance-tests
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"external_ips", "oidc_issuer_url"},
			},
		},
	})
}

func TestAcc_ResourceDeploymentStandard(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)

	// AWS deployment will switch executors during our tests
	awsDeploymentName := fmt.Sprintf("%v_aws", namePrefix)
	azureCeleryDeploymentName := fmt.Sprintf("%v_azure_celery", namePrefix)
	gcpKubernetesDeploymentName := fmt.Sprintf("%v_gcp_kubernetes", namePrefix)

	awsResourceVar := fmt.Sprintf("astro_deployment.%v", awsDeploymentName)
	azureCeleryResourceVar := fmt.Sprintf("astro_deployment.%v", azureCeleryDeploymentName)
	gcpKubernetesResourceVar := fmt.Sprintf("astro_deployment.%v", gcpKubernetesDeploymentName)

	// standard aws deployment
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			// Check that deployments have been removed
			testAccCheckDeploymentExistence(t, awsDeploymentName, true, false),
		),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, true) + standardDeployment(standardDeploymentInput{
					Name:                        awsDeploymentName,
					Description:                 utils.TestResourceDescription,
					Region:                      "us-east-1",
					CloudProvider:               "AWS",
					Executor:                    "KUBERNETES",
					SchedulerSize:               "SMALL",
					IncludeEnvironmentVariables: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(awsResourceVar, "name", awsDeploymentName),
					resource.TestCheckResourceAttr(awsResourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(awsResourceVar, "region", "us-east-1"),
					resource.TestCheckResourceAttr(awsResourceVar, "cloud_provider", "AWS"),
					resource.TestCheckResourceAttr(awsResourceVar, "executor", "KUBERNETES"),
					resource.TestCheckNoResourceAttr(awsResourceVar, "worker_queues"),
					resource.TestCheckResourceAttr(awsResourceVar, "scheduler_size", "SMALL"),
					resource.TestCheckResourceAttrSet(awsResourceVar, "environment_variables.0.key"),
					// Check via API that deployment exists
					testAccCheckDeploymentExistence(t, awsDeploymentName, true, true),
				),
			},
			// Change properties and check they have been updated in terraform state including executor change
			{
				Config: astronomerprovider.ProviderConfig(t, true) + standardDeployment(standardDeploymentInput{
					Name:                        awsDeploymentName,
					Description:                 utils.TestResourceDescription,
					Region:                      "us-east-1",
					CloudProvider:               "AWS",
					Executor:                    "CELERY",
					SchedulerSize:               "MEDIUM",
					IncludeEnvironmentVariables: false,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(awsResourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(awsResourceVar, "scheduler_size", "MEDIUM"),
					resource.TestCheckResourceAttr(awsResourceVar, "worker_queues.0.name", "default"),
					resource.TestCheckNoResourceAttr(awsResourceVar, "environment_variables.0.key"),
					resource.TestCheckResourceAttr(awsResourceVar, "executor", "CELERY"),
					// Check via API that deployment exists
					testAccCheckDeploymentExistence(t, awsDeploymentName, true, true),
				),
			},
			// Change executor back to KUBERNETES and check it is correctly updated in terraform state
			{
				Config: astronomerprovider.ProviderConfig(t, true) + standardDeployment(standardDeploymentInput{
					Name:                        awsDeploymentName,
					Description:                 utils.TestResourceDescription,
					Region:                      "us-east-1",
					CloudProvider:               "AWS",
					Executor:                    "KUBERNETES",
					SchedulerSize:               "MEDIUM",
					IncludeEnvironmentVariables: false,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(awsResourceVar, "executor", "KUBERNETES"),
					resource.TestCheckNoResourceAttr(awsResourceVar, "worker_queues"),
					// Check via API that deployment exists
					testAccCheckDeploymentExistence(t, awsDeploymentName, true, true),
				),
			},
			// Change property that requires destroy and recreate (currently: is_development_mode)
			{
				Config: astronomerprovider.ProviderConfig(t, true) + standardDeployment(standardDeploymentInput{
					Name:                        awsDeploymentName,
					Description:                 utils.TestResourceDescription,
					Region:                      "us-east-1",
					CloudProvider:               "AWS",
					Executor:                    "KUBERNETES",
					SchedulerSize:               "SMALL",
					IncludeEnvironmentVariables: false,
					IsDevelopmentMode:           true,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(awsResourceVar, "scheduler_size", "SMALL"),
					resource.TestCheckResourceAttr(awsResourceVar, "is_development_mode", "true"),
					// Check via API that deployment exists
					testAccCheckDeploymentExistence(t, awsDeploymentName, true, true),
				),
			},
			// Change is_development_mode back to false (will not recreate)
			{
				Config: astronomerprovider.ProviderConfig(t, true) + standardDeployment(standardDeploymentInput{
					Name:                        awsDeploymentName,
					Description:                 utils.TestResourceDescription,
					Region:                      "us-east-1",
					CloudProvider:               "AWS",
					Executor:                    "KUBERNETES",
					SchedulerSize:               "SMALL",
					IncludeEnvironmentVariables: false,
					IsDevelopmentMode:           false,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(awsResourceVar, "scheduler_size", "SMALL"),
					resource.TestCheckResourceAttr(awsResourceVar, "is_development_mode", "false"),
					// Check via API that deployment exists
					testAccCheckDeploymentExistence(t, awsDeploymentName, true, true),
				),
			},
			// Import existing deployment and check it is correctly imported - https://stackoverflow.com/questions/68824711/how-can-i-test-terraform-import-in-acceptance-tests
			{
				ResourceName:            awsResourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"external_ips"},
			},
		},
	})

	// standard azure celery deployment
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			// Check that deployments have been removed
			testAccCheckDeploymentExistence(t, azureCeleryDeploymentName, true, false),
		),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, true) + standardDeployment(standardDeploymentInput{
					Name:                        azureCeleryDeploymentName,
					Description:                 utils.TestResourceDescription,
					Region:                      "westus2",
					CloudProvider:               "AZURE",
					Executor:                    "CELERY",
					SchedulerSize:               "SMALL",
					IncludeEnvironmentVariables: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(azureCeleryResourceVar, "name", azureCeleryDeploymentName),
					resource.TestCheckResourceAttr(azureCeleryResourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(azureCeleryResourceVar, "region", "westus2"),
					resource.TestCheckResourceAttr(azureCeleryResourceVar, "cloud_provider", "AZURE"),
					resource.TestCheckResourceAttr(azureCeleryResourceVar, "executor", "CELERY"),
					resource.TestCheckResourceAttr(azureCeleryResourceVar, "worker_queues.0.name", "default"),
					resource.TestCheckResourceAttr(azureCeleryResourceVar, "scheduler_size", "SMALL"),
					resource.TestCheckResourceAttrSet(azureCeleryResourceVar, "environment_variables.0.key"),
					// Check via API that deployment exists
					testAccCheckDeploymentExistence(t, azureCeleryDeploymentName, true, true),
				),
			},
			// Import existing deployment and check it is correctly imported - https://stackoverflow.com/questions/68824711/how-can-i-test-terraform-import-in-acceptance-tests
			{
				ResourceName:            azureCeleryResourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"external_ips", "oidc_issuer_url"},
			},
		},
	})

	// standard gcp kubernetes deployment
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			// Check that deployments have been removed
			testAccCheckDeploymentExistence(t, gcpKubernetesDeploymentName, true, false),
		),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, true) + standardDeployment(standardDeploymentInput{
					Name:                        gcpKubernetesDeploymentName,
					Description:                 utils.TestResourceDescription,
					Region:                      "westus2",
					CloudProvider:               "AZURE",
					Executor:                    "KUBERNETES",
					SchedulerSize:               "SMALL",
					IncludeEnvironmentVariables: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(gcpKubernetesResourceVar, "name", gcpKubernetesDeploymentName),
					resource.TestCheckResourceAttr(gcpKubernetesResourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(gcpKubernetesResourceVar, "region", "westus2"),
					resource.TestCheckResourceAttr(gcpKubernetesResourceVar, "cloud_provider", "AZURE"),
					resource.TestCheckResourceAttr(gcpKubernetesResourceVar, "executor", "KUBERNETES"),
					resource.TestCheckResourceAttr(gcpKubernetesResourceVar, "worker_queues.#", "0"),
					resource.TestCheckResourceAttr(gcpKubernetesResourceVar, "scheduler_size", "SMALL"),
					resource.TestCheckResourceAttrSet(gcpKubernetesResourceVar, "environment_variables.0.key"),
					// Check via API that deployment exists
					testAccCheckDeploymentExistence(t, gcpKubernetesDeploymentName, true, true),
				),
			},
			// Import existing deployment and check it is correctly imported - https://stackoverflow.com/questions/68824711/how-can-i-test-terraform-import-in-acceptance-tests
			{
				ResourceName:            gcpKubernetesResourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"external_ips", "oidc_issuer_url"},
			},
		},
	})
}

func TestAcc_ResourceDeploymentStandardRemovedOutsideOfTerraform(t *testing.T) {
	standardDeploymentName := utils.GenerateTestResourceName(10)
	standardDeploymentResource := fmt.Sprintf("astro_deployment.%v", standardDeploymentName)
	depInput := standardDeploymentInput{
		Name:                        standardDeploymentName,
		Description:                 utils.TestResourceDescription,
		Region:                      "us-east-1",
		CloudProvider:               "AWS",
		Executor:                    "KUBERNETES",
		IncludeEnvironmentVariables: true,
		SchedulerSize:               "SMALL",
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckDeploymentExistence(t, standardDeploymentName, true, false),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, true) + standardDeploymentWithVariableName(depInput),
				ConfigVariables: map[string]config.Variable{
					"name": config.StringVariable(standardDeploymentName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(standardDeploymentResource, "name", standardDeploymentName),
					resource.TestCheckResourceAttr(standardDeploymentResource, "description", utils.TestResourceDescription),
					// Check via API that workspace exists
					testAccCheckDeploymentExistence(t, standardDeploymentName, true, true),
				),
			},
			{
				PreConfig: func() { deleteDeploymentOutsideOfTerraform(t, standardDeploymentName, true) },
				Config:    astronomerprovider.ProviderConfig(t, true) + standardDeploymentWithVariableName(depInput),
				ConfigVariables: map[string]config.Variable{
					"name": config.StringVariable(standardDeploymentName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(standardDeploymentResource, "name", standardDeploymentName),
					resource.TestCheckResourceAttr(standardDeploymentResource, "description", utils.TestResourceDescription),
					// Check via API that workspace exists
					testAccCheckDeploymentExistence(t, standardDeploymentName, true, true),
				),
			},
		},
	})
}

func workerQueuesStr(nodePoolId string) string {
	workerStr := `astro_machine = "A5"`
	if nodePoolId != "" {
		workerStr = fmt.Sprintf(`node_pool_id = "%v"`, nodePoolId)
	}
	return fmt.Sprintf(`worker_queues = [{
	name = "default"
	is_default = true
	max_worker_count = 10
	min_worker_count = 0
	worker_concurrency = 1
	%v
}]`, workerStr)
}

func envVarsStr(includeEnvVars bool) string {
	environmentVariables := "[]"
	if includeEnvVars {
		environmentVariables = `[{
		key = "key1"
		value = "value1"
		is_secret = false
	}]`
	}
	return fmt.Sprintf("environment_variables = %v", environmentVariables)
}

type hybridDeploymentInput struct {
	Name                        string
	Description                 string
	ClusterId                   string
	Executor                    string
	IncludeEnvironmentVariables bool
	SchedulerAu                 int
	NodePoolId                  string
}

func hybridDeployment(input hybridDeploymentInput) string {
	wqStr := ""
	taskPodNodePoolIdStr := ""
	if input.Executor == string(platform.DeploymentExecutorCELERY) {
		wqStr = workerQueuesStr(input.NodePoolId)
	} else {
		taskPodNodePoolIdStr = fmt.Sprintf(`task_pod_node_pool_id = "%v"`, input.NodePoolId)
	}
	return fmt.Sprintf(`
resource "astro_workspace" "%v_workspace" {
	name = "%s"
	description = "%s"
	cicd_enforced_default = true
}

data "astro_deployment_options" "deployment_options" {}

resource "astro_deployment" "%v" {
	astro_runtime_version = data.astro_deployment_options.deployment_options.runtime_releases[0].version
	name = "%s"
	description = "%s"
	type = "HYBRID"
	cluster_id = "%v"
	contact_emails = []
	executor = "%v"
	is_cicd_enforced = true
	is_dag_deploy_enabled = true
	scheduler_au = %v
	scheduler_replicas = 1
	workspace_id = astro_workspace.%v_workspace.id
	%v
	%v
	%v
  }
`,
		input.Name, input.Name, utils.TestResourceDescription, input.Name, input.Name, input.Description, input.ClusterId, input.Executor, input.SchedulerAu, input.Name,
		envVarsStr(input.IncludeEnvironmentVariables), wqStr, taskPodNodePoolIdStr)
}

type standardDeploymentInput struct {
	Name                        string
	Description                 string
	Region                      string
	CloudProvider               string
	Executor                    string
	IncludeEnvironmentVariables bool
	SchedulerSize               string
	IsDevelopmentMode           bool
}

func standardDeployment(input standardDeploymentInput) string {
	wqStr := ""
	if input.Executor == string(platform.DeploymentExecutorCELERY) {
		wqStr = workerQueuesStr("")
	}
	var scalingSpecStr string

	if input.IsDevelopmentMode == true {
		scalingSpecStr = `scaling_spec            = {
      hibernation_spec      = {
        schedules             = [{
          hibernate_at_cron    = "1 * * * *"
          is_enabled           = true
          wake_at_cron         = "59 * * * *"
        }]
        override            = {
          is_hibernating      = true
          override_until     = "2030-04-25T12:58:00+05:30"
        }
      }
    }`
	}
	return fmt.Sprintf(`
resource "astro_workspace" "%v_workspace" {
	name = "%s"
	description = "%s"
	cicd_enforced_default = true
}

data "astro_deployment_options" "deployment_options" {}

resource "astro_deployment" "%v" {
	astro_runtime_version = data.astro_deployment_options.deployment_options.runtime_releases[0].version
	name = "%s"
	description = "%s"
	type = "STANDARD"
	region = "%v"
	cloud_provider = "%v"
	contact_emails = []
	default_task_pod_cpu = "0.25"
	default_task_pod_memory = "0.5Gi"
	executor = "%v"
	is_cicd_enforced = true
	is_dag_deploy_enabled = true
	is_development_mode = %v
	is_high_availability = false
	resource_quota_cpu = "10"
	resource_quota_memory = "20Gi"
	scheduler_size = "%v"
	workspace_id = astro_workspace.%v_workspace.id
	%v
	%v
    %v
}
`,
		input.Name, input.Name, utils.TestResourceDescription, input.Name, input.Name, input.Description, input.Region, input.CloudProvider, input.Executor, input.IsDevelopmentMode, input.SchedulerSize, input.Name,
		envVarsStr(input.IncludeEnvironmentVariables), wqStr, scalingSpecStr)
}

func standardDeploymentWithVariableName(input standardDeploymentInput) string {
	tfConfig := fmt.Sprintf(`
variable "name" {
	type = string
}

%v`, standardDeployment(input))
	return strings.Replace(tfConfig, fmt.Sprintf(`name = "%v"`, input.Name), "name = var.name", -1)
}

func deleteDeploymentOutsideOfTerraform(t *testing.T, name string, isHosted bool) {
	t.Helper()

	client, err := utils.GetTestPlatformClient(isHosted)
	assert.NoError(t, err)

	organizationId := os.Getenv("HYBRID_ORGANIZATION_ID")
	if isHosted {
		organizationId = os.Getenv("HOSTED_ORGANIZATION_ID")
	}

	ctx := context.Background()
	resp, err := client.ListDeploymentsWithResponse(ctx, organizationId, &platform.ListDeploymentsParams{
		Names: &[]string{name},
	})
	if err != nil {
		assert.NoError(t, err)
	}
	assert.True(t, len(resp.JSON200.Deployments) >= 1, "deployment should exist but list deployments did not find it")
	_, err = client.DeleteDeploymentWithResponse(ctx, organizationId, resp.JSON200.Deployments[0].Id)
	assert.NoError(t, err)
}

func testAccCheckDeploymentExistence(t *testing.T, name string, isHosted, shouldExist bool) func(state *terraform.State) error {
	t.Helper()
	return func(state *terraform.State) error {
		client, err := utils.GetTestPlatformClient(isHosted)
		assert.NoError(t, err)

		organizationId := os.Getenv("HYBRID_ORGANIZATION_ID")
		if isHosted {
			organizationId = os.Getenv("HOSTED_ORGANIZATION_ID")
		}

		ctx := context.Background()
		resp, err := client.ListDeploymentsWithResponse(ctx, organizationId, &platform.ListDeploymentsParams{
			Names: &[]string{name},
		})
		if err != nil {
			return fmt.Errorf("failed to list deployments: %w", err)
		}
		if resp == nil {
			return fmt.Errorf("response is nil")
		}
		if resp.JSON200 == nil {
			status, diag := clients.NormalizeAPIError(ctx, resp.HTTPResponse, resp.Body)
			return fmt.Errorf("response JSON200 is nil status: %v, err: %v", status, diag.Detail())
		}
		if shouldExist {
			if len(resp.JSON200.Deployments) != 1 {
				return fmt.Errorf("deployment %s should exist", name)
			}
		} else {
			if len(resp.JSON200.Deployments) != 0 {
				return fmt.Errorf("deployment %s should not exist", name)
			}
		}
		return nil
	}
}
