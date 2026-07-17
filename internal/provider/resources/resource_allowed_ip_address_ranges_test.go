package resources_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
)

// TestAcc_ResourceAllowedIpAddressRanges exercises the org IP access list end-to-end: seeding the
// first ranges (the empty->first-entry transition the API guards against lockout), replacing a
// range (which exercises create-before-delete), and importing the singleton by organization ID.
//
// This test mutates the organization's real IP access list, so it is opt-in and self-guarding: it
// runs only when ACC_TEST_RUNNER_IP_CIDR is set to a CIDR covering the machine running the test
// (e.g. your public egress IP as a /32). That CIDR is kept in every applied config so the runner is
// never locked out. Run against an environment where the labs allowed-ip-address-ranges routes are
// deployed, and against a disposable organization.
func TestAcc_ResourceAllowedIpAddressRanges(t *testing.T) {
	runnerCidr := os.Getenv("ACC_TEST_RUNNER_IP_CIDR")
	if runnerCidr == "" {
		t.Skip("ACC_TEST_RUNNER_IP_CIDR not set; skipping to avoid locking the runner out of the test organization")
	}

	// TEST-NET CIDRs (RFC 5737) that never route anywhere - throwaway managed ranges.
	const rangeA = "203.0.113.0/24"
	const rangeB = "198.51.100.0/24"
	resourceVar := "astro_allowed_ip_address_ranges.test"

	config := func(ranges ...string) string {
		quoted := make([]string, len(ranges))
		for i, r := range ranges {
			quoted[i] = fmt.Sprintf("%q", r)
		}
		return astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + fmt.Sprintf(`
resource "astro_allowed_ip_address_ranges" "test" {
  ip_address_ranges = [%s]
}
`, strings.Join(quoted, ", "))
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckAllowedIpAddressRangesDestroyed(t, rangeA, rangeB, runnerCidr),
		Steps: []resource.TestStep{
			// Create: seed the first ranges (empty->first-entry). The runner CIDR is included so the
			// API's first-range lockout guard allows the create.
			{
				Config: config(runnerCidr, rangeA),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "ip_address_ranges.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceVar, "ip_address_ranges.*", runnerCidr),
					resource.TestCheckTypeSetElemAttr(resourceVar, "ip_address_ranges.*", rangeA),
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
				),
			},
			// Update: replace rangeA with rangeB while keeping the runner CIDR. Exercises
			// create-before-delete; the runner stays covered throughout the apply.
			{
				Config: config(runnerCidr, rangeB),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "ip_address_ranges.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceVar, "ip_address_ranges.*", runnerCidr),
					resource.TestCheckTypeSetElemAttr(resourceVar, "ip_address_ranges.*", rangeB),
				),
			},
			// Import: the resource is a singleton; import by organization ID.
			{
				ResourceName:      resourceVar,
				ImportState:       true,
				ImportStateId:     os.Getenv("HOSTED_ORGANIZATION_ID"),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAllowedIpAddressRangesDestroyed(t *testing.T, testCidrs ...string) func(s *terraform.State) error {
	t.Helper()
	return func(state *terraform.State) error {
		client, err := utils.GetTestIamClient(true)
		assert.NoError(t, err)

		organizationId := os.Getenv("HOSTED_ORGANIZATION_ID")
		ctx := context.Background()
		limit := 1000
		resp, err := client.ListAllowedIpAddressRangesWithResponse(ctx, organizationId, &iam.ListAllowedIpAddressRangesParams{Limit: &limit})
		if err != nil {
			return fmt.Errorf("failed to list allowed IP address ranges: %v", err)
		}
		if resp == nil || resp.JSON200 == nil {
			status, diag := clients.NormalizeAPIError(ctx, resp.HTTPResponse, resp.Body)
			return fmt.Errorf("response JSON200 is nil status: %v, err: %v", status, diag.Detail())
		}
		testSet := make(map[string]bool, len(testCidrs))
		for _, c := range testCidrs {
			testSet[c] = true
		}
		for _, r := range resp.JSON200.AllowedIpAddressRanges {
			if testSet[r.IpAddressRange] {
				return fmt.Errorf("allowed IP address range %s still exists after destroy", r.IpAddressRange)
			}
		}
		return nil
	}
}
