package resources_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	astronomerprovider "github.com/astronomer/astronomer-terraform-provider/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"

	"github.com/astronomer/astronomer-terraform-provider/internal/clients"
	"github.com/astronomer/astronomer-terraform-provider/internal/clients/platform"
	"github.com/astronomer/astronomer-terraform-provider/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAcc_ResourceClusterWithDedicatedDeployments(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)

	workspaceName := fmt.Sprintf("%v_workspace", namePrefix)
	awsDeploymentName := fmt.Sprintf("%v_deployment_aws", namePrefix)
	azureDeploymentName := fmt.Sprintf("%v_deployment_azure", namePrefix)
	gcpDeploymentName := fmt.Sprintf("%v_deployment_gcp", namePrefix)

	workspaceResourceVar := fmt.Sprintf("astronomer_workspace.%v", workspaceName)
	awsDeploymentResourceVar := fmt.Sprintf("astronomer_deployment.%v", awsDeploymentName)
	azureDeploymentResourceVar := fmt.Sprintf("astronomer_deployment.%v", azureDeploymentName)
	gcpDeploymentResourceVar := fmt.Sprintf("astronomer_deployment.%v", gcpDeploymentName)

	// AWS cluster will switch executors during our tests
	awsClusterName := fmt.Sprintf("%v_aws", namePrefix)
	azureClusterName := fmt.Sprintf("%v_azure", namePrefix)
	gcpClusterName := fmt.Sprintf("%v_gcp", namePrefix)

	awsResourceVar := fmt.Sprintf("astronomer_cluster.%v", awsClusterName)
	azureResourceVar := fmt.Sprintf("astronomer_cluster.%v", azureClusterName)
	gcpResourceVar := fmt.Sprintf("astronomer_cluster.%v", gcpClusterName)

	// aws cluster
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			// Check that clusters have been removed
			testAccCheckClusterExistence(t, awsClusterName, true, false),
		),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, true) +
					workspace(workspaceName, workspaceName, utils.TestResourceDescription, false) +
					cluster(clusterInput{
						Name:           awsClusterName,
						Description:    "bad description",
						Region:         "us-east-1",
						CloudProvider:  "AWS",
						DbInstanceType: "db.m6g.large",
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
					resource.TestCheckResourceAttr(awsResourceVar, "description", "bad description"),
					resource.TestCheckResourceAttr(awsResourceVar, "region", "us-east-1"),
					resource.TestCheckResourceAttr(awsResourceVar, "cloud_provider", "AWS"),
					resource.TestCheckResourceAttr(awsResourceVar, "db_instance_type", "db.m6g.large"),
					resource.TestCheckResourceAttrSet(awsResourceVar, "vpc_subnet_range"),
					resource.TestCheckNoResourceAttr(gcpResourceVar, "pod_subnet_range"),
					resource.TestCheckNoResourceAttr(gcpResourceVar, "service_peering_range"),
					resource.TestCheckNoResourceAttr(gcpResourceVar, "service_subnet_range"),
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
			// Just update cluster
			{
				Config: astronomerprovider.ProviderConfig(t, true) +
					workspace(workspaceName, workspaceName, utils.TestResourceDescription, false) +
					cluster(clusterInput{
						Name:           awsClusterName,
						Description:    "bad description - updated",
						Region:         "us-east-1",
						CloudProvider:  "AWS",
						DbInstanceType: "db.r5.xlarge",
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
					resource.TestCheckResourceAttr(awsResourceVar, "description", "bad description - updated"),
					resource.TestCheckResourceAttr(awsResourceVar, "region", "us-east-1"),
					resource.TestCheckResourceAttr(awsResourceVar, "cloud_provider", "AWS"),
					resource.TestCheckResourceAttr(awsResourceVar, "db_instance_type", "db.r5.xlarge"),
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
			{
				Config: astronomerprovider.ProviderConfig(t, true) +
					workspace(workspaceName, workspaceName, utils.TestResourceDescription, false) +
					cluster(clusterInput{
						Name:                               awsClusterName,
						Description:                        utils.TestResourceDescription,
						Region:                             "us-east-1",
						CloudProvider:                      "AWS",
						DbInstanceType:                     "db.m6g.large",
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
					resource.TestCheckResourceAttr(awsResourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(awsResourceVar, "region", "us-east-1"),
					resource.TestCheckResourceAttr(awsResourceVar, "cloud_provider", "AWS"),
					resource.TestCheckResourceAttr(awsResourceVar, "db_instance_type", "db.m6g.large"),
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
				Config: astronomerprovider.ProviderConfig(t, true) +
					workspace(workspaceName, workspaceName, utils.TestResourceDescription, false) +
					cluster(clusterInput{
						Name:                               awsClusterName,
						Description:                        utils.TestResourceDescription,
						Region:                             "us-east-1",
						CloudProvider:                      "AWS",
						DbInstanceType:                     "db.m6g.large",
						RestrictedWorkspaceResourceVarName: workspaceResourceVar,
					}),
				Check: resource.ComposeTestCheckFunc(
					// Check cluster
					resource.TestCheckResourceAttr(awsResourceVar, "name", awsClusterName),
					resource.TestCheckResourceAttr(awsResourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(awsResourceVar, "region", "us-east-1"),
					resource.TestCheckResourceAttr(awsResourceVar, "cloud_provider", "AWS"),
					resource.TestCheckResourceAttr(awsResourceVar, "db_instance_type", "db.m6g.large"),
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
				ResourceName:      awsResourceVar,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})

	// azure cluster
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			// Check that clusters have been removed
			testAccCheckClusterExistence(t, azureClusterName, true, false),
		),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, true) +
					workspace(workspaceName, workspaceName, utils.TestResourceDescription, false) +
					cluster(clusterInput{
						Name:                               azureClusterName,
						Description:                        utils.TestResourceDescription,
						Region:                             "westus2",
						CloudProvider:                      "AZURE",
						DbInstanceType:                     "Standard_D2ds_v4",
						RestrictedWorkspaceResourceVarName: workspaceResourceVar,
					}) +
					dedicatedDeployment(dedicatedDeploymentInput{
						ClusterResourceVar:   azureResourceVar,
						WorkspaceResourceVar: workspaceResourceVar,
						Name:                 azureDeploymentName,
						Description:          "deployment description",
						SchedulerSize:        "SMALL",
					}),
				Check: resource.ComposeTestCheckFunc(
					// Check cluster
					resource.TestCheckResourceAttr(azureResourceVar, "name", azureClusterName),
					resource.TestCheckResourceAttr(azureResourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(azureResourceVar, "region", "westus2"),
					resource.TestCheckResourceAttr(azureResourceVar, "cloud_provider", "AZURE"),
					resource.TestCheckResourceAttr(azureResourceVar, "db_instance_type", "Standard_D2ds_v4"),
					resource.TestCheckResourceAttrSet(azureResourceVar, "vpc_subnet_range"),
					resource.TestCheckNoResourceAttr(gcpResourceVar, "pod_subnet_range"),
					resource.TestCheckNoResourceAttr(gcpResourceVar, "service_peering_range"),
					resource.TestCheckNoResourceAttr(gcpResourceVar, "service_subnet_range"),
					resource.TestCheckResourceAttr(awsResourceVar, "workspace_ids.#", "1"),

					// Check via API that cluster exists
					testAccCheckClusterExistence(t, azureClusterName, true, true),

					// Check dedicated deployment
					resource.TestCheckResourceAttr(azureDeploymentResourceVar, "name", azureDeploymentName),
					resource.TestCheckResourceAttr(azureDeploymentResourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(azureDeploymentResourceVar, "type", "DEDICATED"),
					resource.TestCheckResourceAttr(azureDeploymentResourceVar, "scheduler_size", "SMALL"),

					// Check via API that deployment exists
					testAccCheckDeploymentExistence(t, azureDeploymentName, true, true),
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

	// gcp cluster
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			// Check that clusters have been removed
			testAccCheckClusterExistence(t, gcpClusterName, true, false),
		),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, true) +
					workspace(workspaceName, workspaceName, utils.TestResourceDescription, false) +
					cluster(clusterInput{
						Name:                               gcpClusterName,
						Description:                        utils.TestResourceDescription,
						Region:                             "us-central1",
						CloudProvider:                      "GCP",
						DbInstanceType:                     "Small General Purpose",
						RestrictedWorkspaceResourceVarName: workspaceResourceVar,
					}) +
					dedicatedDeployment(dedicatedDeploymentInput{
						ClusterResourceVar:   gcpResourceVar,
						WorkspaceResourceVar: workspaceResourceVar,
						Name:                 gcpDeploymentName,
						Description:          utils.TestResourceDescription,
						SchedulerSize:        "SMALL",
					}),
				Check: resource.ComposeTestCheckFunc(
					// Check cluster
					resource.TestCheckResourceAttr(gcpResourceVar, "name", gcpClusterName),
					resource.TestCheckResourceAttr(gcpResourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(gcpResourceVar, "region", "westus2"),
					resource.TestCheckResourceAttr(gcpResourceVar, "cloud_provider", "AZURE"),
					resource.TestCheckResourceAttr(gcpResourceVar, "db_instance_type", "Standard_D2ds_v4"),
					resource.TestCheckResourceAttrSet(gcpResourceVar, "vpc_subnet_range"),
					resource.TestCheckResourceAttrSet(gcpResourceVar, "pod_subnet_range"),
					resource.TestCheckResourceAttrSet(gcpResourceVar, "service_peering_range"),
					resource.TestCheckResourceAttrSet(gcpResourceVar, "service_subnet_range"),
					resource.TestCheckResourceAttr(gcpResourceVar, "workspace_ids.#", "1"),

					// Check via API that cluster exists
					testAccCheckClusterExistence(t, awsClusterName, true, true),

					// Check dedicated deployment
					resource.TestCheckResourceAttr(gcpDeploymentResourceVar, "name", gcpDeploymentName),
					resource.TestCheckResourceAttr(gcpDeploymentResourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(gcpDeploymentResourceVar, "type", "DEDICATED"),
					resource.TestCheckResourceAttr(gcpDeploymentResourceVar, "scheduler_size", "SMALL"),

					// Check via API that deployment exists
					testAccCheckDeploymentExistence(t, gcpDeploymentName, true, true),
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
	clusterName := utils.GenerateTestResourceName(10)
	clusterResource := fmt.Sprintf("astronomer_cluster.%v", clusterName)
	depInput := clusterInput{
		Name:           clusterName,
		Description:    utils.TestResourceDescription,
		Region:         "us-east-1",
		CloudProvider:  "AWS",
		DbInstanceType: "db.m6g.large",
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckClusterExistence(t, clusterName, true, false),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, true) + clusterWithVariableName(depInput),
				ConfigVariables: map[string]config.Variable{
					"name": config.StringVariable(clusterName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(clusterResource, "name", clusterName),
					resource.TestCheckResourceAttr(clusterResource, "description", utils.TestResourceDescription),
					// Check via API that workspace exists
					testAccCheckClusterExistence(t, clusterName, true, true),
				),
			},
			{
				PreConfig: func() { deleteClusterOutsideOfTerraform(t, clusterName) },
				Config:    astronomerprovider.ProviderConfig(t, true) + clusterWithVariableName(depInput),
				ConfigVariables: map[string]config.Variable{
					"name": config.StringVariable(clusterName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(clusterResource, "name", clusterName),
					resource.TestCheckResourceAttr(clusterResource, "description", utils.TestResourceDescription),
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
	return fmt.Sprintf(`resource "astronomer_deployment" "%v" {
	name = "%s"
	description = "%s"
	type = "DEDICATED"
	cluster_id = %s
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
	workspace_id = %s
}
`, input.Name, input.Name, input.Description, input.ClusterResourceVar, input.SchedulerSize, input.WorkspaceResourceVar)
}

type clusterInput struct {
	Name                               string
	Description                        string
	Region                             string
	CloudProvider                      string
	DbInstanceType                     string
	RestrictedWorkspaceResourceVarName string
}

func cluster(input clusterInput) string {
	gcpNetworkFields := ""
	if input.CloudProvider == string(platform.ClusterCloudProviderGCP) {
		gcpNetworkFields = `
pod_subnet_range = "172.21.0.0/19",
service_peering_range = "172.23.0.0/20",
service_subnet_range =  "172.22.0.0/22",`
	}
	return fmt.Sprintf(`resource "astronomer_cluster" "%v" {
	name = "%s"
	description = "%s"
	type = "DEDICATED"
	region = "%v"
	cloud_provider = "%v"
	db_instance_type = "%v"
	vpc_subnet_range = "172.20.0.0/20"
	%v
	workspace_ids = [%v]
}
`, input.Name, input.Name, input.Description, input.Region, input.CloudProvider, input.DbInstanceType, gcpNetworkFields, input.RestrictedWorkspaceResourceVarName)
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

	client, err := utils.GetTestPlatformClient()
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
		client, err := utils.GetTestPlatformClient()
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
