package resources_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"

	"github.com/astronomer/terraform-provider-astro/internal/clients"

	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_ResourceTeamRoles(t *testing.T) {
	testName := utils.GenerateTestResourceName(10)
	deploymentName := fmt.Sprintf("deployment-%v", testName)
	teamId := os.Getenv("HOSTED_TEAM_ID")
	tfVarName := fmt.Sprintf("astro_team_roles.%v", teamId)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					teamRoles(string(iam.TeamOrganizationRoleORGANIZATIONBILLINGADMIN), "[]", "", ""),
				ExpectError: regexp.MustCompile("Attribute workspace_roles set must contain at least 1 elements"),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					teamRoles(string(iam.TeamOrganizationRoleORGANIZATIONBILLINGADMIN), "", "[]", ""),
				ExpectError: regexp.MustCompile("Attribute deployment_roles set must contain at least 1 elements"),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					teamRoles(string(iam.TeamOrganizationRoleORGANIZATIONBILLINGADMIN), "", "", "[]"),
				ExpectError: regexp.MustCompile("Attribute dag_roles set must contain at least 1 elements"),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					teamRoles("", "", "", ""),
				ExpectError: regexp.MustCompile("Attribute organization_role value must be one of"),
			},
			// Test failure: dag_roles with neither dag_id nor tag specified
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					teamRoles(string(iam.TeamOrganizationRoleORGANIZATIONBILLINGADMIN), "", "",
						teamDagRoles([]dagRoleInput{
							{
								DeploymentId: os.Getenv("HOSTED_DEPLOYMENT_ID"),
								Role:         "DAG_VIEWER",
								// Neither dag_id nor tag specified
							},
						})),
				ExpectError: regexp.MustCompile("Invalid DAG role configuration"),
			},
			// Test failure: duplicate dag_role keys (same dag_id+deployment_id)
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					teamRoles(string(iam.TeamOrganizationRoleORGANIZATIONBILLINGADMIN), "", "",
						teamDagRoles([]dagRoleInput{
							{
								DeploymentId: os.Getenv("HOSTED_DEPLOYMENT_ID"),
								DagId:        "test_dag_id",
								Role:         "DAG_VIEWER",
							},
							{
								DeploymentId: os.Getenv("HOSTED_DEPLOYMENT_ID"),
								DagId:        "test_dag_id",
								Role:         "DAG_AUTHOR",
							},
						})),
				ExpectError: regexp.MustCompile("Invalid Configuration: Cannot have multiple DAG roles with the same dag_id/tag and deployment_id combination"),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					teamRoles(string(iam.TeamOrganizationRoleORGANIZATIONBILLINGADMIN), "", "", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfVarName, "team_id", teamId),
					resource.TestCheckResourceAttr(tfVarName, "organization_role", string(iam.TeamOrganizationRoleORGANIZATIONBILLINGADMIN)),
					resource.TestCheckNoResourceAttr(tfVarName, "workspace_roles"),
					resource.TestCheckNoResourceAttr(tfVarName, "deployment_roles"),
					// Check via API that team has correct roles
					testAccCheckTeamRolesCorrect(t, string(iam.TeamOrganizationRoleORGANIZATIONBILLINGADMIN), 0, 0, 0),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					standardDeployment(standardDeploymentInput{
						Name:                        deploymentName,
						Description:                 utils.TestResourceDescription,
						Region:                      "us-east4",
						CloudProvider:               string(platform.DeploymentCloudProviderGCP),
						Executor:                    string(platform.DeploymentExecutorCELERY),
						IncludeEnvironmentVariables: false,
						SchedulerSize:               string(platform.DeploymentSchedulerSizeSMALL),
						IsDevelopmentMode:           false,
						WorkerQueuesStr:             workerQueuesStr(""),
					}) +
					teamRoles(string(iam.TeamOrganizationRoleORGANIZATIONMEMBER),
						fmt.Sprintf(`[{workspace_id = %s
									   role = "WORKSPACE_OWNER"}]`, "astro_workspace."+deploymentName+"_workspace.id"),
						fmt.Sprintf(`[{deployment_id = %s
										role = "DEPLOYMENT_ADMIN"}]`, "astro_deployment."+deploymentName+".id"),
						""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfVarName, "team_id", teamId),
					resource.TestCheckResourceAttr(tfVarName, "organization_role", string(iam.TeamOrganizationRoleORGANIZATIONMEMBER)),
					resource.TestCheckResourceAttr(tfVarName, "workspace_roles.#", "1"),
					resource.TestCheckResourceAttr(tfVarName, "deployment_roles.#", "1"),
					resource.TestCheckResourceAttr(tfVarName, "workspace_roles.0.role", "WORKSPACE_OWNER"),
					resource.TestCheckResourceAttr(tfVarName, "deployment_roles.0.role", "DEPLOYMENT_ADMIN"),

					// Check via API that team has correct roles
					testAccCheckTeamRolesCorrect(t, string(iam.TeamOrganizationRoleORGANIZATIONMEMBER), 1, 1, 0),
				),
			},
			// Create team with dag_roles using dag_id
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					teamRoles(string(iam.TeamOrganizationRoleORGANIZATIONMEMBER),
						fmt.Sprintf(`[{workspace_id = "%s"
									   role = "WORKSPACE_OWNER"}]`, os.Getenv("HOSTED_WORKSPACE_ID")),
						"",
						teamDagRoles([]dagRoleInput{
							{
								DeploymentId: os.Getenv("HOSTED_DEPLOYMENT_ID"),
								DagId:        "test_dag_id",
								Role:         "DAG_VIEWER",
							},
						})),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfVarName, "team_id", teamId),
					resource.TestCheckResourceAttr(tfVarName, "organization_role", string(iam.TeamOrganizationRoleORGANIZATIONMEMBER)),
					resource.TestCheckResourceAttr(tfVarName, "workspace_roles.#", "1"),
					resource.TestCheckResourceAttr(tfVarName, "dag_roles.#", "1"),
					// Check via API that team has correct roles
					testAccCheckTeamRolesCorrect(t, string(iam.TeamOrganizationRoleORGANIZATIONMEMBER), 1, 0, 1),
				),
			},
			// Create team with dag_roles using tag
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					teamRoles(string(iam.TeamOrganizationRoleORGANIZATIONMEMBER),
						fmt.Sprintf(`[{workspace_id = "%s"
									   role = "WORKSPACE_OWNER"}]`, os.Getenv("HOSTED_WORKSPACE_ID")),
						"",
						teamDagRoles([]dagRoleInput{
							{
								DeploymentId: os.Getenv("HOSTED_DEPLOYMENT_ID"),
								Tag:          "production",
								Role:         "DAG_AUTHOR",
							},
						})),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfVarName, "team_id", teamId),
					resource.TestCheckResourceAttr(tfVarName, "organization_role", string(iam.TeamOrganizationRoleORGANIZATIONMEMBER)),
					resource.TestCheckResourceAttr(tfVarName, "workspace_roles.#", "1"),
					resource.TestCheckResourceAttr(tfVarName, "dag_roles.#", "1"),
					// Check via API that team has correct roles
					testAccCheckTeamRolesCorrect(t, string(iam.TeamOrganizationRoleORGANIZATIONMEMBER), 1, 0, 1),
				),
			},
			// Create team with multiple dag_roles (mixed dag_id and tag)
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					teamRoles(string(iam.TeamOrganizationRoleORGANIZATIONMEMBER),
						fmt.Sprintf(`[{workspace_id = "%s"
									   role = "WORKSPACE_OWNER"}]`, os.Getenv("HOSTED_WORKSPACE_ID")),
						"",
						teamDagRoles([]dagRoleInput{
							{
								DeploymentId: os.Getenv("HOSTED_DEPLOYMENT_ID"),
								DagId:        "test_dag_id",
								Role:         "DAG_VIEWER",
							},
							{
								DeploymentId: os.Getenv("HOSTED_DEPLOYMENT_ID"),
								Tag:          "production",
								Role:         "DAG_AUTHOR",
							},
						})),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfVarName, "team_id", teamId),
					resource.TestCheckResourceAttr(tfVarName, "organization_role", string(iam.TeamOrganizationRoleORGANIZATIONMEMBER)),
					resource.TestCheckResourceAttr(tfVarName, "workspace_roles.#", "1"),
					resource.TestCheckResourceAttr(tfVarName, "dag_roles.#", "2"),
					// Check via API that team has correct roles
					testAccCheckTeamRolesCorrect(t, string(iam.TeamOrganizationRoleORGANIZATIONMEMBER), 1, 0, 2),
				),
			},
			// Remove dag_roles and verify they are removed
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					teamRoles(string(iam.TeamOrganizationRoleORGANIZATIONMEMBER),
						fmt.Sprintf(`[{workspace_id = "%s"
									   role = "WORKSPACE_OWNER"}]`, os.Getenv("HOSTED_WORKSPACE_ID")),
						"", ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfVarName, "team_id", teamId),
					resource.TestCheckResourceAttr(tfVarName, "organization_role", string(iam.TeamOrganizationRoleORGANIZATIONMEMBER)),
					resource.TestCheckResourceAttr(tfVarName, "workspace_roles.#", "1"),
					resource.TestCheckNoResourceAttr(tfVarName, "dag_roles"),
					// Check via API that team has correct roles
					testAccCheckTeamRolesCorrect(t, string(iam.TeamOrganizationRoleORGANIZATIONMEMBER), 1, 0, 0),
				),
			},
			// Import existing team_roles and check it is correctly imported - https://stackoverflow.com/questions/68824711/how-can-i-test-terraform-import-in-acceptance-tests
			{
				ResourceName:                         tfVarName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        teamId,
				ImportStateVerifyIdentifierAttribute: "team_id",
			},
		},
	})
}

func teamRoles(orgRole, workspaceRoles, deploymentRoles, dagRoles string) string {
	teamId := os.Getenv("HOSTED_TEAM_ID")
	var workspaceRolesStr, deploymentRolesStr, dagRolesStr string
	if workspaceRoles != "" {
		workspaceRolesStr = fmt.Sprintf("workspace_roles = %s", workspaceRoles)
	}
	if deploymentRoles != "" {
		deploymentRolesStr = fmt.Sprintf("deployment_roles = %s", deploymentRoles)
	}
	if dagRoles != "" {
		dagRolesStr = fmt.Sprintf("dag_roles = %s", dagRoles)
	}
	return fmt.Sprintf(`
resource "astro_team_roles" "%s" {
	team_id = "%s"
	organization_role = "%s"	
	%s
	%s
	%s
}
`, teamId, teamId, orgRole, workspaceRolesStr, deploymentRolesStr, dagRolesStr)
}

func teamDagRoles(roles []dagRoleInput) string {
	if len(roles) == 0 {
		return ""
	}
	entries := lo.Map(roles, func(role dagRoleInput, _ int) string {
		var dagIdOrTag string
		if role.DagId != "" {
			dagIdOrTag = fmt.Sprintf(`dag_id = "%v"`, role.DagId)
		} else if role.Tag != "" {
			dagIdOrTag = fmt.Sprintf(`tag = "%v"`, role.Tag)
		}
		return fmt.Sprintf(`
		{
			deployment_id = "%v"
			%v
			role = "%v"
		}`, role.DeploymentId, dagIdOrTag, role.Role)
	})
	return fmt.Sprintf("[%v]", strings.Join(entries, ","))
}

func testAccCheckTeamRolesCorrect(t *testing.T, organizationRole string, numWorkspaceRoles, numDeploymentRoles, numDagRoles int) func(state *terraform.State) error {
	t.Helper()
	return func(state *terraform.State) error {
		client, err := utils.GetTestHostedIamClient()
		assert.NoError(t, err)

		ctx := context.Background()
		resp, err := client.GetTeamWithResponse(ctx, os.Getenv("HOSTED_ORGANIZATION_ID"), os.Getenv("HOSTED_TEAM_ID"))
		if err != nil {
			return fmt.Errorf("failed to get team: %w", err)
		}
		if resp.JSON200 == nil {
			status, diag := clients.NormalizeAPIError(ctx, resp.HTTPResponse, resp.Body)
			return fmt.Errorf("response JSON200 is nil status: %v, err: %v", status, diag.Detail())
		}
		if string(resp.JSON200.OrganizationRole) != organizationRole {
			return fmt.Errorf("organization role from API '%s' does not match expected value: '%s'", resp.JSON200.OrganizationRole, organizationRole)
		}
		// If numWorkspaceRoles or numDeploymentRoles is not 0 then we need to check the length of the roles
		// If it is nil then that is an error
		// If the length does not match the expected value then that is an error
		if numWorkspaceRoles != 0 && (resp.JSON200.WorkspaceRoles == nil || len(*resp.JSON200.WorkspaceRoles) != numWorkspaceRoles) {
			return fmt.Errorf("workspace roles does not match expected value: expected: %d, actual: %d, roles: %+v", numWorkspaceRoles, len(*resp.JSON200.WorkspaceRoles), *resp.JSON200.WorkspaceRoles)
		}
		if numDeploymentRoles != 0 && (resp.JSON200.DeploymentRoles == nil || len(*resp.JSON200.DeploymentRoles) != numDeploymentRoles) {
			return fmt.Errorf("deployment roles does not match expected value: expected: %d, actual: %dm roles: %+v", numDeploymentRoles, len(*resp.JSON200.DeploymentRoles), *resp.JSON200.DeploymentRoles)
		}
		if numDagRoles != 0 && (resp.JSON200.DagRoles == nil || len(*resp.JSON200.DagRoles) != numDagRoles) {
			return fmt.Errorf("dag roles does not match expected value: expected: %d, actual: %d, roles: %+v", numDagRoles, len(*resp.JSON200.DagRoles), *resp.JSON200.DagRoles)
		}
		return nil
	}
}
