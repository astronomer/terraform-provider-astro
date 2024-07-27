package resources

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
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

// test failures:
// disable team resource if org is isScimEnabled
// mismatch in role and entity type (eg. organization role for workspace)
// create team with no roles, with only deployment, with only workspace
// update team
// delete team
// import team
func TestAcc_ResourceTeam(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)

	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")
	userId := os.Getenv("HOSTED_USER_ID")

	teamName := fmt.Sprintf("%v_team", namePrefix)
	resourceVar := fmt.Sprintf("astro_team.%v", teamName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckTeamExistence(t, teamName, false),
		),
		Steps: []resource.TestStep{
			// Test failure: disable team resource if org is isScimEnabled
			{
				Config: astronomerprovider.ProviderConfig(t, true, true) + team(teamInput{
					Name:             teamName,
					Description:      utils.TestResourceDescription,
					MemberIds:        []string{userId},
					OrganizationRole: string(iam.ORGANIZATIONOWNER),
					DeploymentRoles: []role{
						{
							Role: "DEPLOYMENT_ADMIN",
							Id:   deploymentId,
						},
					},
					WorkspaceRoles: []role{
						{
							Role: string(iam.WORKSPACEOWNER),
							Id:   workspaceId,
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Configuration: Cannot create, update or delete a Team resource when SCIM is enabled"),
			},
			// Test failure: check for mismatch in role and entity type
			{
				Config: astronomerprovider.ProviderConfig(t, true, true) + team(teamInput{
					Name:             teamName,
					Description:      utils.TestResourceDescription,
					MemberIds:        []string{userId},
					OrganizationRole: string(iam.ORGANIZATIONOWNER),
					WorkspaceRoles: []role{
						{
							Role: string(iam.ORGANIZATIONOWNER),
							Id:   workspaceId,
						},
					},
				}),
				ExpectError: regexp.MustCompile(fmt.Sprintf("Role '%s' is not valid for role type '%s'", string(iam.ORGANIZATIONOWNER), string(iam.WORKSPACE))),
			},
			// Create team with all fields
			{
				Config: astronomerprovider.ProviderConfig(t, true, false) + team(teamInput{
					Name:             teamName,
					Description:      utils.TestResourceDescription,
					MemberIds:        []string{userId},
					OrganizationRole: string(iam.ORGANIZATIONOWNER),
					DeploymentRoles: []role{
						{
							Role: "DEPLOYMENT_ADMIN",
							Id:   deploymentId,
						},
					},
					WorkspaceRoles: []role{
						{
							Role: string(iam.WORKSPACEOWNER),
							Id:   workspaceId,
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttr(resourceVar, "name", teamName),
					resource.TestCheckResourceAttr(resourceVar, "description", utils.TestResourceDescription),
					resource.TestCheckResourceAttr(resourceVar, "organization_role", string(iam.ORGANIZATIONOWNER)),
					resource.TestCheckResourceAttr(resourceVar, "member_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "member_ids.0", userId),
					resource.TestCheckResourceAttr(resourceVar, "deployment_roles.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "deployment_roles.0.role", "DEPLOYMENT_ADMIN"),
					resource.TestCheckResourceAttr(resourceVar, "deployment_roles.0.deployment_id", deploymentId),
					resource.TestCheckResourceAttr(resourceVar, "workspace_roles.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "workspace_roles.0.role", string(iam.WORKSPACEOWNER)),
					resource.TestCheckResourceAttr(resourceVar, "workspace_roles.0.workspace_id", workspaceId),
					resource.TestCheckResourceAttrSet(resourceVar, "is_idp_managed"),
					resource.TestCheckResourceAttrSet(resourceVar, "roles_count"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by.id"),
					// Check via API that team exists
					testAccCheckTeamExistence(t, teamName, true),
				),
			},
			// Update team
			{
				Config: astronomerprovider.ProviderConfig(t, true, false) + team(teamInput{
					Name:             teamName,
					Description:      "new description",
					MemberIds:        []string{},
					OrganizationRole: string(iam.ORGANIZATIONOWNER),
					WorkspaceRoles: []role{
						{
							Role: string(iam.WORKSPACEACCESSOR),
							Id:   workspaceId,
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "description", "new description"),
					resource.TestCheckResourceAttr(resourceVar, "member_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceVar, "workspace_roles.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "workspace_roles.0.role", string(iam.WORKSPACEACCESSOR)),
					resource.TestCheckResourceAttr(resourceVar, "workspace_roles.0.workspace_id", workspaceId),
					// Check via API that team exists
					testAccCheckTeamExistence(t, teamName, true),
				),
			},
			// Import existing team and check it is correctly imported
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

type role struct {
	Role string
	Id   string
}

type teamInput struct {
	Name             string
	Description      string
	MemberIds        []string
	OrganizationRole string
	DeploymentRoles  []role
	WorkspaceRoles   []role
}

func team(input teamInput) string {
	deploymentRoles := lo.Map(input.DeploymentRoles, func(role role, _ int) string {
		return fmt.Sprintf(`
		{
			deployment_id = "%v"
			role = "%v"
		}`, role.Id, role.Role)
	})

	workspaceRoles := lo.Map(input.WorkspaceRoles, func(role role, _ int) string {
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
resource "astro_team" "%v" {
	name = "%v"
	description = "%v"
	member_ids = %v
	organization_role = "%v"
	%v
	%v
}`, input.Name, input.Name, input.Description, input.MemberIds, input.OrganizationRole, deploymentRolesStr, workspaceRolesStr)
}

func testAccCheckTeamExistence(t *testing.T, name string, shouldExist bool) func(s *terraform.State) error {
	t.Helper()
	return func(s *terraform.State) error {
		client, err := utils.GetTestIamClient(true)
		assert.NoError(t, err)

		organizationId := os.Getenv("HOSTED_ORGANIZATION_ID")

		ctx := context.Background()

		resp, err := client.ListTeamsWithResponse(ctx, organizationId, &iam.ListTeamsParams{
			Names: &[]string{name},
		})
		if err != nil {
			return fmt.Errorf("failed to list teams: %w", err)
		}
		if resp.JSON200 == nil {
			status, diag := clients.NormalizeAPIError(ctx, resp.HTTPResponse, resp.Body)
			return fmt.Errorf("response JSON200 is nil status: %v, err: %v", status, diag.Detail())
		}
		if shouldExist {
			if len(resp.JSON200.Teams) != 1 {
				return fmt.Errorf("team %s should exist", name)
			}
		} else {
			if len(resp.JSON200.Teams) != 0 {
				return fmt.Errorf("team %s should not exist", name)
			}
		}
		return nil
	}
}
