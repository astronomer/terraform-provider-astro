package resources_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/astronomer/terraform-provider-astro/internal/provider/common"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestAcc_ResourceUserRoles(t *testing.T) {
	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")
	userId := os.Getenv("HOSTED_DUMMY_USER_ID")
	tfVarName := fmt.Sprintf("astro_user_roles.%v", userId)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		Steps: []resource.TestStep{
			// Test failure: check for mismatch in role and entity type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					userRoles(userRolesInput{
						OrganizationRole: string(iam.ORGANIZATIONOWNER),
						WorkspaceRoles: []common.Role{
							{
								Role: string(iam.ORGANIZATIONOWNER),
								Id:   workspaceId,
							},
						},
					}),
				ExpectError: regexp.MustCompile(fmt.Sprintf("Role '%s' is not valid for role type '%s'", string(iam.ORGANIZATIONOWNER), string(iam.WORKSPACE))),
			},
			// Test failure: check for missing corresponding workspace role if deployment role is present
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					userRoles(userRolesInput{
						OrganizationRole: string(iam.ORGANIZATIONOWNER),
						DeploymentRoles: []common.Role{
							{
								Role: "DEPLOYMENT_ADMIN",
								Id:   deploymentId,
							},
						},
					}),
				ExpectError: regexp.MustCompile("Unable to mutate Team roles, not every deployment role has a corresponding workspace role"),
			},
			// Test failure: check for multiple roles with same entity id
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					userRoles(userRolesInput{
						OrganizationRole: string(iam.ORGANIZATIONOWNER),
						WorkspaceRoles: []common.Role{
							{
								Role: string(iam.WORKSPACEOWNER),
								Id:   workspaceId,
							},
							{
								Role: string(iam.WORKSPACEACCESSOR),
								Id:   workspaceId,
							},
						},
					}),
				ExpectError: regexp.MustCompile("Invalid Configuration: Cannot have multiple roles with the same workspace id"),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					userRoles(userRolesInput{
						OrganizationRole: string(iam.ORGANIZATIONOWNER),
					}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfVarName, "user_id", userId),
					resource.TestCheckResourceAttr(tfVarName, "organization_role", string(iam.ORGANIZATIONOWNER)),
					resource.TestCheckNoResourceAttr(tfVarName, "workspace_roles"),
					resource.TestCheckNoResourceAttr(tfVarName, "deployment_roles"),
					// Check via API that user has correct roles
					testAccCheckUserRolesCorrect(t, string(iam.ORGANIZATIONOWNER), nil, nil),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					userRoles(userRolesInput{
						OrganizationRole: string(iam.ORGANIZATIONOWNER),
						WorkspaceRoles: []common.Role{
							{
								Role: string(iam.WORKSPACEOWNER),
								Id:   workspaceId,
							},
						},
						DeploymentRoles: []common.Role{
							{
								Role: "DEPLOYMENT_ADMIN",
								Id:   deploymentId,
							},
						},
					}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfVarName, "user_id", userId),
					resource.TestCheckResourceAttr(tfVarName, "organization_role", string(iam.ORGANIZATIONOWNER)),
					resource.TestCheckResourceAttr(tfVarName, "workspace_roles.#", "1"),
					resource.TestCheckResourceAttr(tfVarName, "deployment_roles.#", "1"),
					resource.TestCheckResourceAttr(tfVarName, "workspace_roles.0.role", string(iam.WORKSPACEOWNER)),
					resource.TestCheckResourceAttr(tfVarName, "deployment_roles.0.role", "DEPLOYMENT_ADMIN"),

					// Check via API that user has correct roles
					testAccCheckUserRolesCorrect(t,
						string(iam.ORGANIZATIONOWNER),
						[]common.Role{
							{
								Role: string(iam.WORKSPACEOWNER),
								Id:   workspaceId,
							},
						},
						[]common.Role{
							{
								Role: "DEPLOYMENT_ADMIN",
								Id:   deploymentId,
							},
						},
					),
				),
			},
			// Import existing user_roles and check it is correctly imported - https://stackoverflow.com/questions/68824711/how-can-i-test-terraform-import-in-acceptance-tests
			{
				ResourceName:                         tfVarName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateId:                        userId,
				ImportStateVerifyIdentifierAttribute: "user_id",
			},
		},
	})
}

type userRolesInput struct {
	OrganizationRole string
	DeploymentRoles  []common.Role
	WorkspaceRoles   []common.Role
}

func userRoles(input userRolesInput) string {
	userId := os.Getenv("HOSTED_DUMMY_USER_ID")
	deploymentRoles := lo.Map(input.DeploymentRoles, func(role common.Role, _ int) string {
		return fmt.Sprintf(`
		{
			deployment_id = "%v"
			role = "%v"
		}`, role.Id, role.Role)
	})

	workspaceRoles := lo.Map(input.WorkspaceRoles, func(role common.Role, _ int) string {
		return fmt.Sprintf(`
		{
			workspace_id = "%v"
			role = "%v"
		}`, role.Id, role.Role)
	})

	var deploymentRolesStr string
	if len(deploymentRoles) > 0 {
		deploymentRolesStr = fmt.Sprintf("deployment_roles = [%v]", strings.Join(deploymentRoles, ","))
	}

	var workspaceRolesStr string
	if len(workspaceRoles) > 0 {
		workspaceRolesStr = fmt.Sprintf("workspace_roles = [%v]", strings.Join(workspaceRoles, ","))
	}
	return fmt.Sprintf(`
resource "astro_user_roles" "%v" {
  	user_id = "%v"
  	organization_role = "%v"
  	%s
	%s
}
`, userId, userId, input.OrganizationRole, workspaceRolesStr, deploymentRolesStr)
}

func testAccCheckUserRolesCorrect(t *testing.T, organizationRole string, workspaceRoles, deploymentRoles []common.Role) func(state *terraform.State) error {
	t.Helper()
	return func(state *terraform.State) error {
		client, err := utils.GetTestHostedIamClient()
		assert.NoError(t, err)

		ctx := context.Background()
		resp, err := client.GetUserWithResponse(ctx, os.Getenv("HOSTED_ORGANIZATION_ID"), os.Getenv("HOSTED_DUMMY_USER_ID"))
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}
		if resp.JSON200 == nil {
			status, diag := clients.NormalizeAPIError(ctx, resp.HTTPResponse, resp.Body)
			return fmt.Errorf("response JSON200 is nil status: %v, err: %v", status, diag.Detail())
		}
		if string(*resp.JSON200.OrganizationRole) != organizationRole {
			return fmt.Errorf("organization role from API '%s' does not match expected value: '%s'", *resp.JSON200.OrganizationRole, organizationRole)
		}
		// If numWorkspaceRoles or numDeploymentRoles is not 0 then we need to check the length of the roles
		// If it is nil then that is an error
		// If the length does not match the expected value then that is an error
		if len(workspaceRoles) != 0 && (resp.JSON200.WorkspaceRoles == nil) {
			missingRoles := common.ContainsWorkspaceRoles(*resp.JSON200.WorkspaceRoles, workspaceRoles)
			if len(missingRoles) > 0 {
				return fmt.Errorf("workspace roles does not contain expected role: expected: %v, missing: %v, roles: %+v", workspaceRoles, missingRoles, *resp.JSON200.WorkspaceRoles)
			}
		}
		if len(deploymentRoles) != 0 && (resp.JSON200.DeploymentRoles == nil) {
			missingRoles := common.ContainsDeploymentRoles(*resp.JSON200.DeploymentRoles, deploymentRoles)
			if len(missingRoles) > 0 {
				return fmt.Errorf("deployment roles does not match expected value: expected: %v, missing: %v, roles: %+v", deploymentRoles, missingRoles, *resp.JSON200.DeploymentRoles)
			}
		}
		return nil
	}
}
