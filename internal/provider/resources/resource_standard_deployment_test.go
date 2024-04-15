package resources_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients/platform"
	astronomerprovider "github.com/astronomer/astronomer-terraform-provider/internal/provider"
	"github.com/astronomer/astronomer-terraform-provider/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_ResourceStandardDeployment(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	awsDeploymentName := fmt.Sprintf("%v_aws", namePrefix)
	azureCeleryDeploymentName := fmt.Sprintf("%v_azure_celery", namePrefix)
	gcpKubernetesDeploymentName := fmt.Sprintf("%v_gcp_celery", namePrefix)

	awsDeploymentResource := fmt.Sprintf("astronomer_standard_deployment.%v", awsDeploymentName)
	azureCeleryDeploymentResource := fmt.Sprintf("astronomer_standard_deployment.%v", azureCeleryDeploymentName)
	gcpKubernetesDeploymentResource := fmt.Sprintf("astronomer_standard_deployment.%v", gcpKubernetesDeploymentName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			// Check that deployments have been removed
			testAccCheckDeploymentExistence(t, awsDeploymentName, false),
		),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, true) + standardDeployment(standardDeploymentInput{
					Name:                        awsDeploymentName,
					Description:                 "test",
					Region:                      "us-east-1",
					CloudProvider:               "AWS",
					Executor:                    "KUBERNETES",
					SchedulerSize:               "SMALL",
					IncludeEnvironmentVariables: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(awsDeploymentResource, "name", awsDeploymentName),
					resource.TestCheckResourceAttr(awsDeploymentResource, "description", "test"),
					resource.TestCheckResourceAttr(awsDeploymentResource, "region", "us-east-1"),
					resource.TestCheckResourceAttr(awsDeploymentResource, "cloud_provider", "AWS"),
					resource.TestCheckResourceAttr(awsDeploymentResource, "executor", "KUBERNETES"),
					resource.TestCheckNoResourceAttr(awsDeploymentResource, "worker_queues"),
					resource.TestCheckResourceAttr(awsDeploymentResource, "scheduler_size", "SMALL"),
					resource.TestCheckResourceAttrSet(awsDeploymentResource, "environment_variables.0.key"),
					// Check via API that deployment exists
					testAccCheckDeploymentExistence(t, awsDeploymentName, true),
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
					resource.TestCheckResourceAttr(awsDeploymentResource, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(awsDeploymentResource, "scheduler_size", "MEDIUM"),
					resource.TestCheckResourceAttr(awsDeploymentResource, "worker_queues.0.name", "default"),
					resource.TestCheckNoResourceAttr(awsDeploymentResource, "environment_variables.0.key"),
					resource.TestCheckResourceAttr(awsDeploymentResource, "executor", "CELERY"),
					// Check via API that deployment exists
					testAccCheckDeploymentExistence(t, awsDeploymentName, true),
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
					resource.TestCheckResourceAttr(awsDeploymentResource, "executor", "KUBERNETES"),
					resource.TestCheckNoResourceAttr(awsDeploymentResource, "worker_queues"),
					// Check via API that deployment exists
					testAccCheckDeploymentExistence(t, awsDeploymentName, true),
				),
			},
			// Import existing deployment and check it is correctly imported - https://stackoverflow.com/questions/68824711/how-can-i-test-terraform-import-in-acceptance-tests
			{
				ResourceName:      awsDeploymentResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			// Check that deployments have been removed
			testAccCheckDeploymentExistence(t, azureCeleryDeploymentName, false),
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
					resource.TestCheckResourceAttr(azureCeleryDeploymentResource, "name", azureCeleryDeploymentName),
					resource.TestCheckResourceAttr(azureCeleryDeploymentResource, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(azureCeleryDeploymentResource, "region", "westus2"),
					resource.TestCheckResourceAttr(azureCeleryDeploymentResource, "cloud_provider", "AZURE"),
					resource.TestCheckResourceAttr(azureCeleryDeploymentResource, "executor", "CELERY"),
					resource.TestCheckResourceAttr(azureCeleryDeploymentResource, "worker_queues.0.name", "default"),
					resource.TestCheckResourceAttr(azureCeleryDeploymentResource, "scheduler_size", "SMALL"),
					resource.TestCheckResourceAttrSet(azureCeleryDeploymentResource, "environment_variables.0.key"),
					// Check via API that deployment exists
					testAccCheckDeploymentExistence(t, azureCeleryDeploymentName, true),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			// Check that deployments have been removed
			testAccCheckDeploymentExistence(t, gcpKubernetesDeploymentName, false),
		),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, true) + standardDeployment(standardDeploymentInput{
					Name:                        gcpKubernetesDeploymentName,
					Description:                 utils.TestResourceDescription,
					Region:                      "us-east4",
					CloudProvider:               "GCP",
					Executor:                    "KUBERNETES",
					SchedulerSize:               "SMALL",
					IncludeEnvironmentVariables: true,
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(gcpKubernetesDeploymentResource, "name", gcpKubernetesDeploymentName),
					resource.TestCheckResourceAttr(gcpKubernetesDeploymentResource, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(gcpKubernetesDeploymentResource, "region", "us-east4"),
					resource.TestCheckResourceAttr(gcpKubernetesDeploymentResource, "cloud_provider", "GCP"),
					resource.TestCheckResourceAttr(gcpKubernetesDeploymentResource, "executor", "KUBERNETES"),
					resource.TestCheckNoResourceAttr(gcpKubernetesDeploymentResource, "worker_queues"),
					resource.TestCheckResourceAttr(gcpKubernetesDeploymentResource, "scheduler_size", "SMALL"),
					resource.TestCheckResourceAttrSet(gcpKubernetesDeploymentResource, "environment_variables.0.key"),
					// Check via API that deployment exists
					testAccCheckDeploymentExistence(t, gcpKubernetesDeploymentName, true),
				),
			},
		},
	})
}

func TestAcc_StandardDeploymentRemovedOutsideOfTerraform(t *testing.T) {
	standardDeploymentName := utils.GenerateTestResourceName(10)
	standardDeploymentResource := fmt.Sprintf("astronomer_standard_deployment.%v", standardDeploymentName)
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
		CheckDestroy:             testAccCheckDeploymentExistence(t, standardDeploymentName, false),
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
					testAccCheckDeploymentExistence(t, standardDeploymentName, true),
				),
			},
			{
				PreConfig: func() { deleteDeploymentOutsideOfTerraform(t, standardDeploymentName) },
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
					testAccCheckDeploymentExistence(t, standardDeploymentName, true),
				),
			},
		},
	})
}

type standardDeploymentInput struct {
	Name                        string
	Description                 string
	Region                      string
	CloudProvider               string
	Executor                    string
	IncludeEnvironmentVariables bool
	SchedulerSize               string
}

func standardDeploymentWithVariableName(input standardDeploymentInput) string {
	tfConfig := fmt.Sprintf(`
variable "name" {
	type = string
}

%v`, standardDeployment(input))
	return strings.Replace(tfConfig, fmt.Sprintf(`name = "%v"`, input.Name), "name = var.name", -1)
}

func standardDeployment(input standardDeploymentInput) string {
	workerQueuesStr := ""
	if input.Executor == string(platform.DeploymentExecutorCELERY) {
		workerQueuesStr = `
			worker_queues = [{
				name = "default"
				is_default = true
				astro_machine = "A5"
				max_worker_count = 10
				min_worker_count = 0
				worker_concurrency = 1
			}]`
	}
	environmentVariables := "[]"
	if input.IncludeEnvironmentVariables {
		environmentVariables = `[{
				key = "key1"
				value = "value1"
				is_secret = false
			}]`
	}
	return fmt.Sprintf(`
resource "astronomer_workspace" "%v_workspace" {
	name = "%s"
	description = "%s"
	cicd_enforced_default = true
}

resource "astronomer_standard_deployment" "%v" {
	name = "%s"
	description = "%s"
	region = "%v"
	cloud_provider = "%v"
	contact_emails = []
	default_task_pod_cpu = "0.25"
	default_task_pod_memory = "0.5Gi"
	executor = "%v"
	is_cicd_enforced = true
	is_dag_deploy_enabled = true
	is_development_mode = false
	is_high_availability = false
	resource_quota_cpu = "10"
	resource_quota_memory = "20Gi"
	scheduler_size = "%v"
	workspace_id = astronomer_workspace.%v_workspace.id
	environment_variables = %v
	%v
}
`, input.Name, input.Name, utils.TestResourceDescription, input.Name, input.Name, input.Description, input.Region, input.CloudProvider, input.Executor, input.SchedulerSize, input.Name, environmentVariables, workerQueuesStr)
}

func deleteDeploymentOutsideOfTerraform(t *testing.T, name string) {
	t.Helper()

	client, err := utils.GetTestPlatformClient()
	assert.NoError(t, err)

	ctx := context.Background()
	resp, err := client.ListDeploymentsWithResponse(ctx, os.Getenv("HOSTED_ORGANIZATION_ID"), &platform.ListDeploymentsParams{
		Names: &[]string{name},
	})
	if err != nil {
		assert.NoError(t, err)
	}
	assert.True(t, len(resp.JSON200.Deployments) >= 1, "deployment should exist but list deployments did not find it")
	_, err = client.DeleteDeploymentWithResponse(ctx, os.Getenv("HOSTED_ORGANIZATION_ID"), resp.JSON200.Deployments[0].Id)
	assert.NoError(t, err)
}

func testAccCheckDeploymentExistence(t *testing.T, name string, shouldExist bool) func(state *terraform.State) error {
	t.Helper()
	return func(state *terraform.State) error {
		client, err := utils.GetTestPlatformClient()
		assert.NoError(t, err)

		ctx := context.Background()
		resp, err := client.ListDeploymentsWithResponse(ctx, os.Getenv("HOSTED_ORGANIZATION_ID"), &platform.ListDeploymentsParams{
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
