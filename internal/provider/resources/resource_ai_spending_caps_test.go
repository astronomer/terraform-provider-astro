package resources_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/labs"
	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAcc_ResourceAiSpendingCaps exercises the AI spending caps end to end against a real
// environment: setting the user default and a user's override, changing them, importing the
// singleton by organization ID, leaving the user map out (no change in the plan, the override
// stays), emptying the default object and the user map (both removed), and destroying, which must
// remove nothing.
//
// It changes the organization's real caps, so it is opt-in: set ACC_TEST_AI_SPENDING_CAPS=1, and run
// it against an environment where the labs AI spending caps endpoints are deployed and the
// enableLabsAiSpendingCaps flag covers HOSTED_ORGANIZATION_ID, and against a disposable
// organization on token pricing. HOSTED_USER_ID is the user that gets an override. The caps are
// high enough ($100,000 and $50,000) never to block anyone. The organization must start with no
// defaults and no user overrides, or the create fails and asks for an import. A destroy removes no
// default or override, so the check after it removes them to leave the organization as it was.
func TestAcc_ResourceAiSpendingCaps(t *testing.T) {
	if os.Getenv("ACC_TEST_AI_SPENDING_CAPS") == "" {
		t.Skip("ACC_TEST_AI_SPENDING_CAPS not set; skipping because the test changes the organization's real AI spending caps")
	}
	userId := os.Getenv("HOSTED_USER_ID")
	resourceVar := "astro_ai_spending_caps.test"

	withBody := func(body string) string {
		return astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + fmt.Sprintf(`
resource "astro_ai_spending_caps" "test" {
%s
}
`, body)
	}
	config := func(defaultCents, userCents int) string {
		return withBody(fmt.Sprintf(`  default_cap_cents = { user = %d }
  user_cap_cents    = { %q = %d }`, defaultCents, userId, userCents))
	}
	overrideIs := func(want bool) resource.TestCheckFunc {
		return func(*terraform.State) error {
			has, err := userHasAiSpendingCapOverride(userId)
			if err != nil {
				return err
			}
			if has != want {
				return fmt.Errorf("user %s has an override: %v, want %v", userId, has, want)
			}
			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckAiSpendingCapsDestroyed(t, userId),
		Steps: []resource.TestStep{
			// Create: the user default and the user's override.
			{
				Config: config(10_000_000, 5_000_000),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "default_cap_cents.user", "10000000"),
					resource.TestCheckNoResourceAttr(resourceVar, "default_cap_cents.api_token"),
					resource.TestCheckResourceAttr(resourceVar, "user_cap_cents."+userId, "5000000"),
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
				),
			},
			// Update: change both.
			{
				Config: config(9_000_000, 4_000_000),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "default_cap_cents.user", "9000000"),
					resource.TestCheckResourceAttr(resourceVar, "user_cap_cents."+userId, "4000000"),
				),
			},
			// Import: the resource is a singleton; import by organization ID.
			{
				ResourceName:      resourceVar,
				ImportState:       true,
				ImportStateId:     os.Getenv("HOSTED_ORGANIZATION_ID"),
				ImportStateVerify: true,
			},
			// Leaving the user map out shows no change and keeps the override.
			{
				Config:           withBody(`  default_cap_cents = { user = 9000000 }`),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
				Check:            overrideIs(true),
			},
			// {} removes everything it covers: the user default and the user's override.
			{
				Config: withBody(`  default_cap_cents = {}
  user_cap_cents    = {}`),
				Check: resource.ComposeTestCheckFunc(
					overrideIs(false),
					func(*terraform.State) error {
						defaults, err := aiSpendingCapDefaults()
						if err != nil {
							return err
						}
						if defaults.UserSpendingCap != nil {
							return fmt.Errorf("the user default is still %d after default_cap_cents = {}", *defaults.UserSpendingCap)
						}
						return nil
					},
				),
			},
			// Set them again, for the destroy to leave in place.
			{
				Config: config(9_000_000, 4_000_000),
				Check:  overrideIs(true),
			},
		},
	})
}

// aiSpendingCapDefaults reads the organization's defaults straight from the API.
func aiSpendingCapDefaults() (*labs.AiSpendingCapDefaults, error) {
	client, err := labs.NewLabsClient(os.Getenv("ASTRO_API_HOST"), os.Getenv("HOSTED_ORGANIZATION_API_TOKEN"), "test")
	if err != nil {
		return nil, fmt.Errorf("failed to create the labs client: %v", err)
	}
	ctx := context.Background()
	defaults, err := client.LabsGetAiSpendingCapDefaultsWithResponse(ctx, os.Getenv("HOSTED_ORGANIZATION_ID"))
	if err != nil {
		return nil, fmt.Errorf("failed to get the default AI spending caps: %v", err)
	}
	if _, d := clients.NormalizeAPIResponseWithBody(ctx, defaults.HTTPResponse, defaults.Body, defaults.JSON200, "get the default AI spending caps"); d != nil {
		return nil, fmt.Errorf("%s: %s", d.Summary(), d.Detail())
	}
	return defaults.JSON200, nil
}

// userHasAiSpendingCapOverride reports whether the user has an override, read straight from the API.
func userHasAiSpendingCapOverride(userId string) (bool, error) {
	client, err := labs.NewLabsClient(os.Getenv("ASTRO_API_HOST"), os.Getenv("HOSTED_ORGANIZATION_API_TOKEN"), "test")
	if err != nil {
		return false, fmt.Errorf("failed to create the labs client: %v", err)
	}
	ctx := context.Background()
	hasOverride := true
	list, err := client.LabsListAiSpendingCapsWithResponse(ctx, os.Getenv("HOSTED_ORGANIZATION_ID"), &labs.LabsListAiSpendingCapsParams{SubjectType: labs.USER, HasOverride: &hasOverride})
	if err != nil {
		return false, fmt.Errorf("failed to list the AI spending caps: %v", err)
	}
	if _, d := clients.NormalizeAPIResponseWithBody(ctx, list.HTTPResponse, list.Body, list.JSON200, "list the AI spending caps"); d != nil {
		return false, fmt.Errorf("%s: %s", d.Summary(), d.Detail())
	}
	for _, entry := range list.JSON200.AiSpendingCaps {
		if entry.SubjectId == userId {
			return true, nil
		}
	}
	return false, nil
}

func testAccCheckAiSpendingCapsDestroyed(t *testing.T, userId string) func(s *terraform.State) error {
	t.Helper()
	return func(state *terraform.State) error {
		client, err := labs.NewLabsClient(os.Getenv("ASTRO_API_HOST"), os.Getenv("HOSTED_ORGANIZATION_API_TOKEN"), "test")
		if err != nil {
			return fmt.Errorf("failed to create the labs client: %v", err)
		}
		ctx := context.Background()
		organizationId := os.Getenv("HOSTED_ORGANIZATION_ID")

		defaults, err := aiSpendingCapDefaults()
		if err != nil {
			return err
		}
		// A destroy removes no default or override, so both are still there. Remove them, so the test
		// leaves the organization as it found it.
		if defaults.UserSpendingCap == nil || *defaults.UserSpendingCap != 9_000_000 {
			return fmt.Errorf("the user default changed on destroy, which leaves defaults in place: %v", defaults.UserSpendingCap)
		}
		removedDefault, err := client.LabsDeleteAiSpendingCapDefaultsWithResponse(ctx, organizationId,
			labs.DeleteAiSpendingCapDefaultsRequest{SubjectTypes: []labs.DeleteAiSpendingCapDefaultsRequestSubjectTypes{labs.DeleteAiSpendingCapDefaultsRequestSubjectTypesUSER}})
		if err != nil {
			return fmt.Errorf("failed to remove the user default: %v", err)
		}
		if _, d := clients.NormalizeAPIResponseWithBody(ctx, removedDefault.HTTPResponse, removedDefault.Body, removedDefault.JSON200, "remove the user default"); d != nil {
			return fmt.Errorf("%s: %s", d.Summary(), d.Detail())
		}

		has, err := userHasAiSpendingCapOverride(userId)
		if err != nil {
			return err
		}
		if !has {
			return fmt.Errorf("user %s lost its override on destroy, which leaves overrides in place", userId)
		}
		removed, err := client.LabsDeleteAiSpendingCapsWithResponse(ctx, organizationId, labs.DeleteAiSpendingCapsRequest{
			Subjects: []labs.AiSpendingCapSubject{{SubjectId: userId, SubjectType: labs.AiSpendingCapSubjectSubjectTypeUSER}},
		})
		if err != nil {
			return fmt.Errorf("failed to remove the user's override: %v", err)
		}
		if _, d := clients.NormalizeAPIResponseWithBody(ctx, removed.HTTPResponse, removed.Body, removed.JSON200, "remove the user's override"); d != nil {
			return fmt.Errorf("%s: %s", d.Summary(), d.Detail())
		}
		return nil
	}
}
