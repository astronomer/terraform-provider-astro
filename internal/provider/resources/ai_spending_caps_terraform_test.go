package resources_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os/exec"
	"regexp"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients/labs"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/resources"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/lucsky/cuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// capsTestProvider is a provider with only the AI spending caps resource, handing it a client for
// the fake. The real provider only accepts Astronomer's own hosts, so it cannot be pointed at a
// test server; the resource, its schema and its validators are the real ones.
type capsTestProvider struct{ labsClient *labs.ClientWithResponses }

func (p *capsTestProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "astro"
}

func (p *capsTestProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = providerschema.Schema{}
}

func (p *capsTestProvider) Configure(_ context.Context, _ provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	resp.ResourceData = models.ApiClientsModel{OrganizationId: "org", LabsClient: p.labsClient}
}

func (p *capsTestProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{resources.NewAiSpendingCapsResource}
}

func (p *capsTestProvider) DataSources(context.Context) []func() datasource.DataSource { return nil }

// TestUnit_ResourceAiSpendingCaps_Terraform runs the real Terraform CLI against a fake of Core's
// labs AI spending caps endpoints: plan, apply, refresh and import through the provider server, so
// what the direct tests of the resource's methods cannot see is covered here: the schema's
// validators, a plan that is empty after every apply, and drift.
func TestUnit_ResourceAiSpendingCaps_Terraform(t *testing.T) {
	// The test needs a Terraform binary on the PATH, which CI installs. Without one, the testing
	// library would try to download Terraform, so skip instead.
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("terraform is not on the PATH; skipping the test that runs the real Terraform CLI")
	}
	api := utils.NewFakeAiSpendingCapsAPI()
	srv := httptest.NewServer(api)
	t.Cleanup(srv.Close)

	userA, userB, userC, uiUser, tokenA := cuid.New(), cuid.New(), cuid.New(), cuid.New(), cuid.New()
	const name = "astro_ai_spending_caps.test"
	labsClient, err := labs.NewLabsClient(srv.URL, "token", "test")
	require.NoError(t, err)
	config := func(body string) string {
		return fmt.Sprintf(`
provider "astro" {}

resource "astro_ai_spending_caps" "test" {
%s
}
`, body)
	}
	// holds checks what the fake API holds, which is what Core would be enforcing.
	holds := func(defaults map[string]int, users, tokens map[string]int) tfresource.TestCheckFunc {
		return func(*terraform.State) error {
			assert.Equal(t, defaults, api.Defaults, "defaults")
			assert.Equal(t, users, api.Caps["USER"], "user caps")
			assert.Equal(t, tokens, api.Caps["SERVICEKEY"], "API token caps")
			return nil
		}
	}

	tfresource.UnitTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"astro": providerserver.NewProtocol6WithError(&capsTestProvider{labsClient: labsClient}),
		},
		CheckDestroy: func(*terraform.State) error { return holds(map[string]int{}, map[string]int{}, map[string]int{})(nil) },
		Steps: []tfresource.TestStep{
			// Validation, before any request: a key that is not an ID, a cap below 0 and one above
			// $100,000.
			{
				Config:      config(`user_cap_cents = { "not-an-id" = 100 }`),
				ExpectError: regexp.MustCompile(`(?s)user_cap_cents\["not-an-id"\] value must be a cuid`),
			},
			{
				Config: config(fmt.Sprintf(`default_user_cap_cents = -1
  user_cap_cents = { %q = 100 }`, userA)),
				ExpectError: regexp.MustCompile(`(?s)default_user_cap_cents value must be between 0 and 10000000`),
			},
			{
				Config:      config(fmt.Sprintf(`api_token_cap_cents = { %q = 10000001 }`, tokenA)),
				ExpectError: regexp.MustCompile(`(?s)api_token_cap_cents\[.*\] value must be\s+between 0 and 10000000`),
			},
			// Create: both defaults, two users (one of them 0, which is a cap) and an API token.
			{
				Config: config(fmt.Sprintf(`
  default_user_cap_cents      = 5000
  default_api_token_cap_cents = 2500
  user_cap_cents              = { %q = 100, %q = 0 }
  api_token_cap_cents         = { %q = 300 }`, userA, userB, tokenA)),
				Check: tfresource.ComposeTestCheckFunc(
					tfresource.TestCheckResourceAttr(name, "id", "org"),
					tfresource.TestCheckResourceAttr(name, "default_user_cap_cents", "5000"),
					tfresource.TestCheckResourceAttr(name, "user_cap_cents.%", "2"),
					tfresource.TestCheckResourceAttr(name, "user_cap_cents."+userB, "0"),
					tfresource.TestCheckResourceAttr(name, "api_token_cap_cents."+tokenA, "300"),
					holds(map[string]int{"USER": 5000, "SERVICEKEY": 2500}, map[string]int{userA: 100, userB: 0}, map[string]int{tokenA: 300}),
				),
			},
			// Update: change one cap, drop one, add one, and leave the API token default out, which
			// removes it. The plan after the apply is empty, or the step fails.
			{
				Config: config(fmt.Sprintf(`
  default_user_cap_cents = 5000
  user_cap_cents         = { %q = 150, %q = 20 }
  api_token_cap_cents    = { %q = 300 }`, userA, userC, tokenA)),
				Check: tfresource.ComposeTestCheckFunc(
					tfresource.TestCheckNoResourceAttr(name, "default_api_token_cap_cents"),
					holds(map[string]int{"USER": 5000}, map[string]int{userA: 150, userC: 20}, map[string]int{tokenA: 300}),
				),
			},
			// Drift: a cap set in the UI for a user the map does not list is removed, since the map
			// is authoritative.
			{
				PreConfig: func() { api.Caps["USER"][uiUser] = 700 },
				Config: config(fmt.Sprintf(`
  default_user_cap_cents = 5000
  user_cap_cents         = { %q = 150, %q = 20 }
  api_token_cap_cents    = { %q = 300 }`, userA, userC, tokenA)),
				Check: holds(map[string]int{"USER": 5000}, map[string]int{userA: 150, userC: 20}, map[string]int{tokenA: 300}),
			},
			// Import: the singleton is imported by the organization ID, and reads everything.
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateId:     "org",
				ImportStateVerify: true,
			},
			// The labs API not being turned on for the organization is an error, not a missing
			// resource that Terraform would plan to create again.
			{
				PreConfig: func() { api.GateOff = true },
				Config: config(fmt.Sprintf(`
  default_user_cap_cents = 5000
  user_cap_cents         = { %q = 150, %q = 20 }
  api_token_cap_cents    = { %q = 300 }`, userA, userC, tokenA)),
				ExpectError: regexp.MustCompile(`(?s)AI spending caps are not available`),
			},
			// The labs API is available again, and the resource is back to what it held.
			{
				PreConfig: func() { api.GateOff = false },
				Config: config(fmt.Sprintf(`
  default_user_cap_cents = 5000
  user_cap_cents         = { %q = 150, %q = 20 }
  api_token_cap_cents    = { %q = 300 }`, userA, userC, tokenA)),
				Check: holds(map[string]int{"USER": 5000}, map[string]int{userA: 150, userC: 20}, map[string]int{tokenA: 300}),
			},
		},
	})
}
