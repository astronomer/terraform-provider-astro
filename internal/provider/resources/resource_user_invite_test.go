package resources_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAcc_ResourceUserInvite(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	email := "astro-terraform-test@astronomer.test"

	userInviteName := fmt.Sprintf("%v_user_invite", namePrefix)
	tfVarName := fmt.Sprintf("astro_user_invite.%v", userInviteName)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckUserInviteExistence(t, email, false),
		),
		Steps: []resource.TestStep{
			// Test failure: check for invalid email
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					userInvite(userInviteInput{
						Name:  userInviteName,
						Email: "invalid-email",
						Role:  string(iam.CreateUserInviteRequestRoleORGANIZATIONOWNER),
					}),
				ExpectError: regexp.MustCompile("must be a valid email address"),
			},
			// Test failure: check for invalid role
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					userInvite(userInviteInput{
						Name:  userInviteName,
						Email: email,
						Role:  "invalid-role",
					}),
				ExpectError: regexp.MustCompile("must be one of"),
			},
			// Create user invite
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					userInvite(userInviteInput{
						Name:  userInviteName,
						Email: email,
						Role:  string(iam.CreateUserInviteRequestRoleORGANIZATIONOWNER),
					}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfVarName, "email", email),
					resource.TestCheckResourceAttr(tfVarName, "role", string(iam.CreateUserInviteRequestRoleORGANIZATIONOWNER)),
					resource.TestCheckResourceAttrSet(tfVarName, "invite_id"),
					resource.TestCheckResourceAttrSet(tfVarName, "expires_at"),
					resource.TestCheckResourceAttrSet(tfVarName, "invitee.id"),
					resource.TestCheckResourceAttrSet(tfVarName, "inviter.id"),
					resource.TestCheckResourceAttrSet(tfVarName, "user_id"),
					// Check via API that user invite exists
					testAccCheckUserInviteExistence(t, email, true),
				),
			},
			// Update user invite
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
					userInvite(userInviteInput{
						Name:  userInviteName,
						Email: email,
						Role:  string(iam.CreateUserInviteRequestRoleORGANIZATIONMEMBER),
					}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfVarName, "email", email),
					resource.TestCheckResourceAttr(tfVarName, "role", string(iam.CreateUserInviteRequestRoleORGANIZATIONMEMBER)),
					resource.TestCheckResourceAttrSet(tfVarName, "invite_id"),
					// Check via API that user invite exists
					testAccCheckUserInviteExistence(t, email, true),
				),
			},
		},
	})
}

type userInviteInput struct {
	Name  string
	Email string
	Role  string
}

func userInvite(input userInviteInput) string {
	return fmt.Sprintf(`
resource "astro_user_invite" "%v" {
	email = "%v"
	role = "%v"
}
`, input.Name, input.Email, input.Role)
}

func testAccCheckUserInviteExistence(t *testing.T, email string, shouldExist bool) func(state *terraform.State) error {
	t.Helper()
	return func(state *terraform.State) error {
		client, err := utils.GetTestHostedIamClient()
		assert.NoError(t, err)

		ctx := context.Background()
		resp, err := client.ListUsersWithResponse(ctx, os.Getenv("HOSTED_ORGANIZATION_ID"), nil)
		if err != nil {
			return fmt.Errorf("failed to list users: %w", err)
		}
		if resp.JSON200 == nil {
			status, diag := clients.NormalizeAPIError(ctx, resp.HTTPResponse, resp.Body)
			return fmt.Errorf("response JSON200 is nil status: %v, err: %v", status, diag.Detail())
		}

		var userInvitee *iam.User
		for _, user := range resp.JSON200.Users {
			if user.Username == email {
				userInvitee = &user
			}
		}

		if shouldExist {
			if userInvitee == nil {
				return fmt.Errorf("user invite %s should exist", email)
			}
		} else {
			if userInvitee != nil {
				return fmt.Errorf("user invite %s should not exist", email)
			}
		}

		return nil
	}
}
