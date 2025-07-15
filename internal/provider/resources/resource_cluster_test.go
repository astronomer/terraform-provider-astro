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

// These acceptance tests are testing the creation of dedicated clusters in the Astronomer platform.
// We are also testing 'DEDICATED' deployment resources in these tests since they will be created in the clusters we create.

const SKIP_CLUSTER_RESOURCE_TESTS = "SKIP_CLUSTER_RESOURCE_TESTS"
const SKIP_CLUSTER_RESOURCE_TESTS_REASON = "Skipping dedicated cluster (and dedicated deployment) resource tests. To run these tests, unset the SKIP_CLUSTER_RESOURCE_TESTS environment variable."

func TestAcc_ResourceClusterAwsWithDedicatedDeployments(t *testing.T) {
	if os.Getenv(SKIP_CLUSTER_RESOURCE_TESTS) == "True" {
		t.Skip(SKIP_CLUSTER_RESOURCE_TESTS_REASON)
	}
	namePrefix := utils.GenerateTestResourceName(10)

	workspaceName := fmt.Sprintf("%v_workspace", namePrefix)
	awsDeploymentName := fmt.Sprintf("%v_deployment_aws", namePrefix)

	workspaceResourceVar := fmt.Sprintf("astro_workspace.%v", workspaceName)
	awsDeploymentResourceVar := fmt.Sprintf("astro_deployment.%v", awsDeploymentName)

	// deployments in AWS cluster will switch executors during our tests
	awsClusterName := fmt.Sprintf("%v_aws", namePrefix)
	awsResourceVar := fmt.Sprintf("astro_cluster.%v", awsClusterName)

	// aws cluster
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			// Check that clusters have been removed
			testAccCheckClusterExistence(t, awsClusterName, true, false),
		),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					workspace(workspaceName, workspaceName, utils.TestResourceDescription, false) +
					cluster(clusterInput{
						Name:                               awsClusterName,
						Region:                             "us-east-1",
						CloudProvider:                      "AWS",
						RestrictedWorkspaceResourceVarName: workspaceResourceVar,
					}) +
					dedicatedDeployment(dedicatedDeploymentInput{
						ClusterResourceVar:   awsResourceVar,
						WorkspaceResourceVar: workspaceResourceVar,
						Name:                 awsDeploymentName,
						Description:          "deployment description",
						SchedulerSize:        "SMALL",
					}) +
					dedicatedDeploymentWithAstroExecutor(dedicatedDeploymentInput{
						ClusterResourceVar:   awsResourceVar,
						WorkspaceResourceVar: workspaceResourceVar,
						Name:                 awsDeploymentName,
						Description:          "deployment description",
						SchedulerSize:        "SMALL",
					}),
				Check: resource.ComposeTestCheckFunc(
					// Check cluster
					resource.TestCheckResourceAttr(awsResourceVar, "name", awsClusterName),
					resource.TestCheckResourceAttr(awsResourceVar, "region", "us-east-1"),
					resource.TestCheckResourceAttr(awsResourceVar, "cloud_provider", "AWS"),
					resource.TestCheckResourceAttrSet(awsResourceVar, "vpc_subnet_range"),
					resource.TestCheckResourceAttr(awsResourceVar, "workspace_ids.#", "1"),

					// Check via API that cluster exists
					testAccCheckClusterExistence(t, awsClusterName, true, true),

					// Check dedicated deployment
					resource.TestCheckResourceAttr(awsDeploymentResourceVar, "name", awsDeploymentName),
					resource.TestCheckResourceAttr(awsDeploymentResourceVar, "description", "deployment description"),
					resource.TestCheckResourceAttr(awsDeploymentResourceVar, "type", "DEDICATED"),
					resource.TestCheckResourceAttr(awsDeploymentResourceVar, "scheduler_size", "SMALL"),
					// Check dedicated deployment with Astro executor
					resource.TestCheckResourceAttr(awsDeploymentResourceVar+"_astro", "name", awsDeploymentName+"_astro"),
					resource.TestCheckResourceAttr(awsDeploymentResourceVar+"_astro", "description", "deployment description"),
					resource.TestCheckResourceAttr(awsDeploymentResourceVar+"_astro", "type", "DEDICATED"),
					resource.TestCheckResourceAttr(awsDeploymentResourceVar+"_astro", "scheduler_size", "SMALL"),
					resource.TestCheckResourceAttr(awsDeploymentResourceVar+"_astro", "executor", "ASTRO"),

					// Check via API that deployment exists
					testAccCheckDeploymentExistence(t, awsDeploymentName, true, true),
					testAccCheckDeploymentExistence(t, awsDeploymentName+"_astro", true, true),
				),
			},
			// Just update cluster and remove workspace restrictions
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					workspace(workspaceName, workspaceName, utils.TestResourceDescription, false) +
					cluster(clusterInput{
						Name:          awsClusterName,
						Region:        "us-east-1",
						CloudProvider: "AWS",
					}) +
					dedicatedDeployment(dedicatedDeploymentInput{
						ClusterResourceVar:   awsResourceVar,
						WorkspaceResourceVar: workspaceResourceVar,
						Name:                 awsDeploymentName,
						Description:          "deployment description",
						SchedulerSize:        "SMALL",
					}),
				Check: resource.ComposeTestCheckFunc(
					// Check cluster
					resource.TestCheckResourceAttr(awsResourceVar, "name", awsClusterName),
					resource.TestCheckResourceAttr(awsResourceVar, "region", "us-east-1"),
					resource.TestCheckResourceAttr(awsResourceVar, "cloud_provider", "AWS"),
					resource.TestCheckResourceAttrSet(awsResourceVar, "vpc_subnet_range"),
					resource.TestCheckResourceAttr(awsResourceVar, "workspace_ids.#", "0"),

					// Check via API that cluster exists
					testAccCheckClusterExistence(t, awsClusterName, true, true),

					// Check dedicated deployment
					resource.TestCheckResourceAttr(awsDeploymentResourceVar, "name", awsDeploymentName),
					resource.TestCheckResourceAttr(awsDeploymentResourceVar, "description", "deployment description"),
					resource.TestCheckResourceAttr(awsDeploymentResourceVar, "type", "DEDICATED"),
					resource.TestCheckResourceAttr(awsDeploymentResourceVar, "scheduler_size", "SMALL"),

					// Check via API that deployment exists
					testAccCheckDeploymentExistence(t, awsDeploymentName, true, true),
				),
			},
			// Change properties of cluster and deployment and check they have been updated in terraform state
			// Add back workspace restrictions
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					workspace(workspaceName, workspaceName, utils.TestResourceDescription, false) +
					cluster(clusterInput{
						Name:                               awsClusterName,
						Region:                             "us-east-1",
						CloudProvider:                      "AWS",
						RestrictedWorkspaceResourceVarName: workspaceResourceVar,
					}) +
					dedicatedDeployment(dedicatedDeploymentInput{
						ClusterResourceVar:   awsResourceVar,
						WorkspaceResourceVar: workspaceResourceVar,
						Name:                 awsDeploymentName,
						Description:          utils.TestResourceDescription,
						SchedulerSize:        "MEDIUM",
					}),
				Check: resource.ComposeTestCheckFunc(
					// Check cluster
					resource.TestCheckResourceAttr(awsResourceVar, "name", awsClusterName),
					resource.TestCheckResourceAttr(awsResourceVar, "region", "us-east-1"),
					resource.TestCheckResourceAttr(awsResourceVar, "cloud_provider", "AWS"),
					resource.TestCheckResourceAttrSet(awsResourceVar, "vpc_subnet_range"),
					resource.TestCheckResourceAttr(awsResourceVar, "workspace_ids.#", "1"),

					// Check via API that cluster exists
					testAccCheckClusterExistence(t, awsClusterName, true, true),

					// Check dedicated deployment
					resource.TestCheckResourceAttr(awsDeploymentResourceVar, "name", awsDeploymentName),
					resource.TestCheckResourceAttr(awsDeploymentResourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(awsDeploymentResourceVar, "type", "DEDICATED"),
					resource.TestCheckResourceAttr(awsDeploymentResourceVar, "scheduler_size", "MEDIUM"),

					// Check via API that deployment exists
					testAccCheckDeploymentExistence(t, awsDeploymentName, true, true),
				),
			},
			// Remove deployment
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					workspace(workspaceName, workspaceName, utils.TestResourceDescription, false) +
					cluster(clusterInput{
						Name:                               awsClusterName,
						Region:                             "us-east-1",
						CloudProvider:                      "AWS",
						RestrictedWorkspaceResourceVarName: workspaceResourceVar,
					}),
				Check: resource.ComposeTestCheckFunc(
					// Check cluster
					resource.TestCheckResourceAttr(awsResourceVar, "name", awsClusterName),
					resource.TestCheckResourceAttr(awsResourceVar, "region", "us-east-1"),
					resource.TestCheckResourceAttr(awsResourceVar, "cloud_provider", "AWS"),
					resource.TestCheckResourceAttrSet(awsResourceVar, "vpc_subnet_range"),
					resource.TestCheckResourceAttr(awsResourceVar, "workspace_ids.#", "1"),

					// Check via API that cluster exists
					testAccCheckClusterExistence(t, awsClusterName, true, true),

					// Check via API that deployment does not exist
					testAccCheckDeploymentExistence(t, awsDeploymentName, true, false),
				),
			},
			// Import existing cluster and check it is correctly imported - https://stackoverflow.com/questions/68824711/how-can-i-test-terraform-import-in-acceptance-tests
			{
				ResourceName:            awsResourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"health_status", "health_status.value"},
			},
		},
	})
}

func TestAcc_ResourceClusterAzureWithDedicatedDeployments(t *testing.T) {
	if os.Getenv(SKIP_CLUSTER_RESOURCE_TESTS) == "True" {
		t.Skip(SKIP_CLUSTER_RESOURCE_TESTS_REASON)
	}
	namePrefix := utils.GenerateTestResourceName(10)

	workspaceName := fmt.Sprintf("%v_workspace", namePrefix)
	azureDeploymentName := fmt.Sprintf("%v_deployment_azure", namePrefix)

	workspaceResourceVar := fmt.Sprintf("astro_workspace.%v", workspaceName)
	azureDeploymentResourceVar := fmt.Sprintf("astro_deployment.%v", azureDeploymentName)

	azureClusterName := fmt.Sprintf("%v_azure", namePrefix)
	azureResourceVar := fmt.Sprintf("astro_cluster.%v", azureClusterName)

	// azure cluster
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			// Check that clusters have been removed
			testAccCheckClusterExistence(t, azureClusterName, true, false),
		),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					workspace(workspaceName, workspaceName, utils.TestResourceDescription, false) +
					cluster(clusterInput{
						Name:                               azureClusterName,
						Region:                             "westus2",
						CloudProvider:                      "AZURE",
						RestrictedWorkspaceResourceVarName: workspaceResourceVar,
					}) +
					dedicatedDeployment(dedicatedDeploymentInput{
						ClusterResourceVar:   azureResourceVar,
						WorkspaceResourceVar: workspaceResourceVar,
						Name:                 azureDeploymentName,
						Description:          utils.TestResourceDescription,
						SchedulerSize:        "SMALL",
					}) +
					dedicatedDeploymentWithAstroExecutor(dedicatedDeploymentInput{
						ClusterResourceVar:   azureResourceVar,
						WorkspaceResourceVar: workspaceResourceVar,
						Name:                 azureDeploymentName,
						Description:          utils.TestResourceDescription,
						SchedulerSize:        "SMALL",
					}),
				Check: resource.ComposeTestCheckFunc(
					// Check cluster
					resource.TestCheckResourceAttr(azureResourceVar, "name", azureClusterName),
					resource.TestCheckResourceAttr(azureResourceVar, "region", "westus2"),
					resource.TestCheckResourceAttr(azureResourceVar, "cloud_provider", "AZURE"),
					resource.TestCheckResourceAttrSet(azureResourceVar, "vpc_subnet_range"),
					resource.TestCheckResourceAttr(azureResourceVar, "workspace_ids.#", "1"),

					// Check via API that cluster exists
					testAccCheckClusterExistence(t, azureClusterName, true, true),

					// Check dedicated deployment
					resource.TestCheckResourceAttr(azureDeploymentResourceVar, "name", azureDeploymentName),
					resource.TestCheckResourceAttr(azureDeploymentResourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(azureDeploymentResourceVar, "type", "DEDICATED"),
					resource.TestCheckResourceAttr(azureDeploymentResourceVar, "scheduler_size", "SMALL"),
					// Check dedicated deployment with Astro executor
					resource.TestCheckResourceAttr(azureDeploymentResourceVar+"_astro", "name", azureDeploymentName+"_astro"),
					resource.TestCheckResourceAttr(azureDeploymentResourceVar+"_astro", "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(azureDeploymentResourceVar+"_astro", "type", "DEDICATED"),
					resource.TestCheckResourceAttr(azureDeploymentResourceVar+"_astro", "scheduler_size", "SMALL"),
					resource.TestCheckResourceAttr(azureDeploymentResourceVar+"_astro", "executor", "ASTRO"),

					// Check via API that deployment exists
					testAccCheckDeploymentExistence(t, azureDeploymentName, true, true),
					testAccCheckDeploymentExistence(t, azureDeploymentName+"_astro", true, true),
				),
			},
			// Import existing cluster and check it is correctly imported - https://stackoverflow.com/questions/68824711/how-can-i-test-terraform-import-in-acceptance-tests
			{
				ResourceName:      azureResourceVar,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ResourceClusterGcpWithDedicatedDeployments(t *testing.T) {
	if os.Getenv(SKIP_CLUSTER_RESOURCE_TESTS) == "True" {
		t.Skip(SKIP_CLUSTER_RESOURCE_TESTS_REASON)
	}
	namePrefix := utils.GenerateTestResourceName(10)

	workspaceName := fmt.Sprintf("%v_workspace", namePrefix)
	gcpDeploymentName := fmt.Sprintf("%v_deployment_gcp", namePrefix)

	workspaceResourceVar := fmt.Sprintf("astro_workspace.%v", workspaceName)
	gcpDeploymentResourceVar := fmt.Sprintf("astro_deployment.%v", gcpDeploymentName)

	gcpClusterName := fmt.Sprintf("%v_gcp", namePrefix)
	gcpResourceVar := fmt.Sprintf("astro_cluster.%v", gcpClusterName)

	// gcp cluster
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			// Check that clusters have been removed
			testAccCheckClusterExistence(t, gcpClusterName, true, false),
		),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					workspace(workspaceName, workspaceName, utils.TestResourceDescription, false) +
					cluster(clusterInput{
						Name:                               gcpClusterName,
						Region:                             "us-central1",
						CloudProvider:                      "GCP",
						RestrictedWorkspaceResourceVarName: workspaceResourceVar,
					}) +
					dedicatedDeployment(dedicatedDeploymentInput{
						ClusterResourceVar:   gcpResourceVar,
						WorkspaceResourceVar: workspaceResourceVar,
						Name:                 gcpDeploymentName,
						Description:          utils.TestResourceDescription,
						SchedulerSize:        "SMALL",
					}) +
					dedicatedDeploymentWithAstroExecutor(dedicatedDeploymentInput{
						ClusterResourceVar:   gcpResourceVar,
						WorkspaceResourceVar: workspaceResourceVar,
						Name:                 gcpDeploymentName,
						Description:          utils.TestResourceDescription,
						SchedulerSize:        "SMALL",
					}),
				Check: resource.ComposeTestCheckFunc(
					// Check cluster
					resource.TestCheckResourceAttr(gcpResourceVar, "name", gcpClusterName),
					resource.TestCheckResourceAttr(gcpResourceVar, "region", "us-central1"),
					resource.TestCheckResourceAttr(gcpResourceVar, "cloud_provider", "GCP"),
					resource.TestCheckResourceAttrSet(gcpResourceVar, "vpc_subnet_range"),
					resource.TestCheckResourceAttrSet(gcpResourceVar, "pod_subnet_range"),
					resource.TestCheckResourceAttrSet(gcpResourceVar, "service_peering_range"),
					resource.TestCheckResourceAttrSet(gcpResourceVar, "service_subnet_range"),
					resource.TestCheckResourceAttr(gcpResourceVar, "workspace_ids.#", "1"),

					// Check via API that cluster exists
					testAccCheckClusterExistence(t, gcpClusterName, true, true),

					// Check dedicated deployment
					resource.TestCheckResourceAttr(gcpDeploymentResourceVar, "name", gcpDeploymentName),
					resource.TestCheckResourceAttr(gcpDeploymentResourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(gcpDeploymentResourceVar, "type", "DEDICATED"),
					resource.TestCheckResourceAttr(gcpDeploymentResourceVar, "scheduler_size", "SMALL"),
					// Check dedicated deployment with Astro executor
					resource.TestCheckResourceAttr(gcpDeploymentResourceVar+"_astro", "name", gcpDeploymentName+"_astro"),
					resource.TestCheckResourceAttr(gcpDeploymentResourceVar+"_astro", "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(gcpDeploymentResourceVar+"_astro", "type", "DEDICATED"),
					resource.TestCheckResourceAttr(gcpDeploymentResourceVar+"_astro", "scheduler_size", "SMALL"),
					resource.TestCheckResourceAttr(gcpDeploymentResourceVar+"_astro", "executor", "ASTRO"),

					// Check via API that deployment exists
					testAccCheckDeploymentExistence(t, gcpDeploymentName, true, true),
					testAccCheckDeploymentExistence(t, gcpDeploymentName+"_astro", true, true),
				),
			},
			// Import existing cluster and check it is correctly imported - https://stackoverflow.com/questions/68824711/how-can-i-test-terraform-import-in-acceptance-tests
			{
				ResourceName:      gcpResourceVar,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ResourceClusterRemovedOutsideOfTerraform(t *testing.T) {
	if os.Getenv(SKIP_CLUSTER_RESOURCE_TESTS) == "True" {
		t.Skip(SKIP_CLUSTER_RESOURCE_TESTS_REASON)
	}
	clusterName := utils.GenerateTestResourceName(10)
	clusterResource := fmt.Sprintf("astro_cluster.%v", clusterName)
	depInput := clusterInput{
		Name:          clusterName,
		Region:        "us-central1",
		CloudProvider: "GCP",
	}
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckClusterExistence(t, clusterName, true, false),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + clusterWithVariableName(depInput),
				ConfigVariables: map[string]config.Variable{
					"name": config.StringVariable(clusterName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(clusterResource, "name", clusterName),
					// Check via API that workspace exists
					testAccCheckClusterExistence(t, clusterName, true, true),
				),
			},
			{
				PreConfig: func() { deleteClusterOutsideOfTerraform(t, clusterName) },
				Config:    astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + clusterWithVariableName(depInput),
				ConfigVariables: map[string]config.Variable{
					"name": config.StringVariable(clusterName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(clusterResource, "name", clusterName),
					// Check via API that workspace exists
					testAccCheckClusterExistence(t, clusterName, true, true),
				),
			},
		},
	})
}

type dedicatedDeploymentInput struct {
	ClusterResourceVar   string
	WorkspaceResourceVar string
	Name                 string
	Description          string
	SchedulerSize        string
}

func dedicatedDeployment(input dedicatedDeploymentInput) string {
	return fmt.Sprintf(`
resource "astro_deployment" "%v" {
	name = "%s"
	description = "%s"
	type = "DEDICATED"
	cluster_id = %s.id
	contact_emails = []
	default_task_pod_cpu = "0.25"
	default_task_pod_memory = "0.5Gi"
	executor = "KUBERNETES"
	is_cicd_enforced = true
	is_dag_deploy_enabled = true
	is_development_mode = false
	is_high_availability = false
	resource_quota_cpu = "10"
	resource_quota_memory = "20Gi"
	scheduler_size = "%v"
	workspace_id = %s.id
	environment_variables = []
}
`, input.Name, input.Name, input.Description, input.ClusterResourceVar, input.SchedulerSize, input.WorkspaceResourceVar)
}

func dedicatedDeploymentWithAstroExecutor(input dedicatedDeploymentInput) string {
	return fmt.Sprintf(`
resource "astro_deployment" "%v_astro" {
	name = "%v_astro"
	description = "%s"
	type = "DEDICATED"
	cluster_id = %s.id
	contact_emails = []
	default_task_pod_cpu = "0.25"
	default_task_pod_memory = "0.5Gi"
	executor = "ASTRO"
	is_cicd_enforced = true
	is_dag_deploy_enabled = true
	is_development_mode = false
	is_high_availability = false
	resource_quota_cpu = "10"
	resource_quota_memory = "20Gi"
	scheduler_size = "%v"
	workspace_id = %s.id
	environment_variables = []
	worker_queues = [{ name = "default", is_default = true, astro_machine = "A5", max_worker_count = 2, min_worker_count = 1, worker_concurrency = 5 }]
}
`, input.Name, input.Name, input.Description, input.ClusterResourceVar, input.SchedulerSize, input.WorkspaceResourceVar)
}

type clusterInput struct {
	Name                               string
	Region                             string
	CloudProvider                      string
	RestrictedWorkspaceResourceVarName string
}

func cluster(input clusterInput) string {
	gcpNetworkFields := ""
	workspaceId := ""
	if input.RestrictedWorkspaceResourceVarName != "" {
		workspaceId = fmt.Sprintf("%v.id", input.RestrictedWorkspaceResourceVarName)
	}
	if input.CloudProvider == string(platform.ClusterCloudProviderGCP) {
		gcpNetworkFields = `
	pod_subnet_range = "172.21.0.0/19"
	service_peering_range = "172.23.0.0/20"
	service_subnet_range =  "172.22.0.0/22"`
	}
	return fmt.Sprintf(`resource "astro_cluster" "%v" {
	name = "%s"
	type = "DEDICATED"
	region = "%v"
	cloud_provider = "%v"
	vpc_subnet_range = "172.20.0.0/20"
	%v
	workspace_ids = [%v]
}
`, input.Name, input.Name, input.Region, input.CloudProvider, gcpNetworkFields, workspaceId)
}

func clusterWithVariableName(input clusterInput) string {
	tfConfig := fmt.Sprintf(`
variable "name" {
	type = string
}

%v`, cluster(input))
	return strings.Replace(tfConfig, fmt.Sprintf(`name = "%v"`, input.Name), "name = var.name", -1)
}

func deleteClusterOutsideOfTerraform(t *testing.T, name string) {
	t.Helper()

	client, err := utils.GetTestHostedPlatformClient()
	assert.NoError(t, err)

	organizationId := os.Getenv("HOSTED_ORGANIZATION_ID")

	ctx := context.Background()
	resp, err := client.ListClustersWithResponse(ctx, organizationId, &platform.ListClustersParams{
		Names: &[]string{name},
	})
	if err != nil {
		assert.NoError(t, err)
	}
	assert.True(t, len(resp.JSON200.Clusters) >= 1, "cluster should exist but list clusters did not find it")
	_, err = client.DeleteClusterWithResponse(ctx, organizationId, resp.JSON200.Clusters[0].Id)
	assert.NoError(t, err)
}

func testAccCheckClusterExistence(t *testing.T, name string, isHosted, shouldExist bool) func(state *terraform.State) error {
	t.Helper()
	return func(state *terraform.State) error {
		client, err := utils.GetTestHostedPlatformClient()
		assert.NoError(t, err)

		organizationId := os.Getenv("HYBRID_ORGANIZATION_ID")
		if isHosted {
			organizationId = os.Getenv("HOSTED_ORGANIZATION_ID")
		}

		ctx := context.Background()
		resp, err := client.ListClustersWithResponse(ctx, organizationId, &platform.ListClustersParams{
			Names: &[]string{name},
		})
		if err != nil {
			return fmt.Errorf("failed to list clusters: %w", err)
		}
		if resp == nil {
			return fmt.Errorf("response is nil")
		}
		if resp.JSON200 == nil {
			status, diag := clients.NormalizeAPIError(ctx, resp.HTTPResponse, resp.Body)
			return fmt.Errorf("response JSON200 is nil status: %v, err: %v", status, diag.Detail())
		}
		if shouldExist {
			if len(resp.JSON200.Clusters) != 1 {
				return fmt.Errorf("cluster %s should exist", name)
			}
		} else {
			if len(resp.JSON200.Clusters) != 0 {
				return fmt.Errorf("cluster %s should not exist", name)
			}
		}
		return nil
	}
}
