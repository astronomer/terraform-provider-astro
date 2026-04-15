package resources_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"regexp"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAcc_ResourceTeamMembership(t *testing.T) {
	teamName := fmt.Sprintf("%v_membership_team", utils.GenerateTestResourceName(10))
	userId := os.Getenv("HOSTED_USER_ID")
	tfVarName := "astro_team_membership.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckTeamMembershipNotExists(t, teamName, userId),
		Steps: []resource.TestStep{
			// Invalid team_id (not a CUID) — expect plan-time error
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					teamMembership("not-a-cuid", userId),
				ExpectError: testAccCuidValidatorError(),
			},
			// Invalid user_id (not a CUID) — expect plan-time error
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					teamMembershipWithTeamResource(teamName, "not-a-cuid"),
				ExpectError: testAccCuidValidatorError(),
			},
			// Create team + membership
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					teamMembershipWithTeamResource(teamName, userId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfVarName, "id"),
					resource.TestCheckResourceAttrPair(tfVarName, "team_id", "astro_team.membership_team", "id"),
					resource.TestCheckResourceAttr(tfVarName, "user_id", userId),
					testAccCheckTeamMembershipExists(t, teamName, userId),
				),
			},
			// Import via <team_id>/<user_id>
			{
				ResourceName:      tfVarName,
				ImportState:       true,
				ImportStateIdFunc: testAccTeamMembershipImportStateIdFunc(tfVarName),
				ImportStateVerify: true,
			},
		},
	})
}

// teamMembership builds a config for a standalone membership with literal IDs (for validation error tests).
func teamMembership(teamId, userId string) string {
	return fmt.Sprintf(`
resource "astro_team_membership" "test" {
  team_id = %q
  user_id = %q
}`, teamId, userId)
}

// teamMembershipWithTeamResource builds a config that creates a team and adds a member to it.
func teamMembershipWithTeamResource(teamName, userId string) string {
	return fmt.Sprintf(`
resource "astro_team" "membership_team" {
  name              = %q
  organization_role = "ORGANIZATION_MEMBER"
}

resource "astro_team_membership" "test" {
  team_id = astro_team.membership_team.id
  user_id = %q
}`, teamName, userId)
}

// testAccTeamMembershipImportStateIdFunc returns the composite <team_id>/<user_id> import ID from state.
func testAccTeamMembershipImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}
		teamId := rs.Primary.Attributes["team_id"]
		userId := rs.Primary.Attributes["user_id"]
		return teamId + "/" + userId, nil
	}
}

// testAccCheckTeamMembershipExists verifies via the API that userId is a member of the named team.
func testAccCheckTeamMembershipExists(t *testing.T, teamName, userId string) resource.TestCheckFunc {
	t.Helper()
	return func(s *terraform.State) error {
		return checkTeamMembership(t, teamName, userId, true)
	}
}

// testAccCheckTeamMembershipNotExists verifies via the API that userId is NOT a member of the named team.
func testAccCheckTeamMembershipNotExists(t *testing.T, teamName, userId string) resource.TestCheckFunc {
	t.Helper()
	return func(s *terraform.State) error {
		return checkTeamMembership(t, teamName, userId, false)
	}
}

func checkTeamMembership(t *testing.T, teamName, userId string, shouldExist bool) error {
	t.Helper()

	iamClient, err := utils.GetTestHostedIamClient()
	assert.NoError(t, err)

	organizationId := os.Getenv("HOSTED_ORGANIZATION_ID")
	ctx := context.Background()

	// Find the team by name
	teamsResp, err := iamClient.ListTeamsWithResponse(ctx, organizationId, &iam.ListTeamsParams{
		Names: &[]string{teamName},
	})
	if err != nil {
		return fmt.Errorf("failed to list teams: %w", err)
	}
	if teamsResp.JSON200 == nil {
		_, diag := clients.NormalizeAPIError(ctx, teamsResp.HTTPResponse, teamsResp.Body)
		return fmt.Errorf("list teams: %v", diag.Detail())
	}
	if len(teamsResp.JSON200.Teams) == 0 {
		if shouldExist {
			return fmt.Errorf("team %q not found", teamName)
		}
		return nil // team gone, so membership is gone too
	}

	teamId := teamsResp.JSON200.Teams[0].Id

	// Check membership
	membersResp, err := iamClient.ListTeamMembersWithResponse(ctx, organizationId, teamId, nil)
	if err != nil {
		return fmt.Errorf("failed to list team members: %w", err)
	}
	statusCode, diag := clients.NormalizeAPIError(ctx, membersResp.HTTPResponse, membersResp.Body)
	if statusCode == http.StatusNotFound {
		if shouldExist {
			return fmt.Errorf("team %q (id=%s) not found when checking members", teamName, teamId)
		}
		return nil
	}
	if diag != nil {
		return fmt.Errorf("list team members: %v", diag.Detail())
	}

	for _, m := range membersResp.JSON200.TeamMembers {
		if m.UserId == userId {
			if !shouldExist {
				return fmt.Errorf("user %s should NOT be a member of team %q but is", userId, teamName)
			}
			return nil
		}
	}

	if shouldExist {
		return fmt.Errorf("user %s is not a member of team %q", userId, teamName)
	}
	return nil
}

// testAccCuidValidatorError returns a regexp that matches the CUID validator error message.
func testAccCuidValidatorError() *regexp.Regexp {
	return regexp.MustCompile(`(?i)cuid|invalid`)
}
