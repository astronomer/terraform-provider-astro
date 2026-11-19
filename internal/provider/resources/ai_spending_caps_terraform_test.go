package resources_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os/exec"
	"regexp"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients/labs"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/provider/resources"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
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

// beforeAndAfter checks a map attribute's value before and after the planned change, so a test can
// see that a plan lists a removal.
type beforeAndAfter struct {
	attr          string
	before, after map[string]any
}

func (c beforeAndAfter) CheckPlan(_ context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	if len(req.Plan.ResourceChanges) == 0 {
		resp.Error = fmt.Errorf("%s: no resource change in the plan", c.attr)
		return
	}
	for _, rc := range req.Plan.ResourceChanges {
		before, _ := rc.Change.Before.(map[string]any)
		after, _ := rc.Change.After.(map[string]any)
		norm := func(v any) map[string]any {
			b, _ := json.Marshal(v)
			out := map[string]any{}
			_ = json.Unmarshal(b, &out)
			return out
		}
		gotBefore, gotAfter := norm(before[c.attr]), norm(after[c.attr])
		wantBefore, wantAfter := norm(c.before), norm(c.after)
		if fmt.Sprint(gotBefore) != fmt.Sprint(wantBefore) || fmt.Sprint(gotAfter) != fmt.Sprint(wantAfter) {
			resp.Error = fmt.Errorf("%s: plan before %v after %v, want before %v after %v", c.attr, gotBefore, gotAfter, wantBefore, wantAfter)
		}
	}
}

// TestUnit_ResourceAiSpendingCaps_Terraform runs the real Terraform CLI against the fake: refresh,
// plan, apply, import and destroy through the provider server. It checks what calling the
// resource's methods directly cannot: that an attribute left out shows no change in the plan, that
// the plan lists every removal, that one apply converges (the testing library fails a step whose
// plan after the apply is not empty), and that a create which would change existing caps fails.
func TestUnit_ResourceAiSpendingCaps_Terraform(t *testing.T) {
	// The test needs a Terraform binary on the PATH, which CI installs. Without one, the testing
	// library would try to download Terraform, so skip instead.
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("terraform is not on the PATH; skipping the test that runs the real Terraform CLI")
	}
	api := resources.NewFakeAiSpendingCapsAPI()
	srv := httptest.NewServer(api)
	t.Cleanup(srv.Close)
	labsClient, err := labs.NewLabsClient(srv.URL, "token", "test")
	require.NoError(t, err)

	a, uiUser, uiUser2, tok := cuid.New(), cuid.New(), cuid.New(), cuid.New()
	const name = "astro_ai_spending_caps.test"
	cfg := func(body string) string {
		return fmt.Sprintf("provider \"astro\" {}\nresource \"astro_ai_spending_caps\" \"test\" {\n%s\n}\n", body)
	}
	holds := func(defaults, users, tokens map[string]int) tfresource.TestCheckFunc {
		return func(*terraform.State) error {
			assert.Equal(t, defaults, api.Defaults, "defaults")
			assert.Equal(t, users, api.Caps["USER"], "user caps")
			assert.Equal(t, tokens, api.Caps["SERVICEKEY"], "API token caps")
			return nil
		}
	}
	// Set in the UI before Terraform: a user's override of 0 and a user default.
	api.Caps["USER"][uiUser] = 0
	api.Defaults["USER"] = 1000

	tokMap := fmt.Sprintf("api_token_cap_cents = { %q = 300 }", tok)

	tfresource.UnitTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"astro": providerserver.NewProtocol6WithError(&capsTestProvider{labsClient: labsClient}),
		},
		// Destroy removes nothing.
		CheckDestroy: func(*terraform.State) error {
			return holds(map[string]int{}, map[string]int{}, map[string]int{tok: 400})(nil)
		},
		Steps: []tfresource.TestStep{
			// Validation, before any request: a key that is not an ID, and caps out of range.
			{
				Config:      cfg(`user_cap_cents = { "not-an-id" = 100 }`),
				ExpectError: regexp.MustCompile(`(?s)user_cap_cents\["not-an-id"\] value must be a cuid`),
			},
			{
				Config:      cfg(`default_cap_cents = { user = -1 }`),
				ExpectError: regexp.MustCompile(`(?s)default_cap_cents.user value must be between 0 and 10000000`),
			},
			{
				Config:      cfg(fmt.Sprintf(`api_token_cap_cents = { %q = 10000001 }`, tok)),
				ExpectError: regexp.MustCompile(`(?s)api_token_cap_cents\[.*\] value must be\s+between 0 and 10000000`),
			},
			// 1. Create refuses to adopt: managing users would remove uiUser's 0, which no plan shows.
			{
				Config:      cfg(fmt.Sprintf("user_cap_cents = { %q = 100 }", a)),
				ExpectError: regexp.MustCompile(`(?s)already has AI spending caps.*override of user\s+` + uiUser),
			},
			// 2. Create managing only API tokens: allowed. Users and defaults are read, not touched.
			{
				Config: cfg(tokMap),
				Check: tfresource.ComposeTestCheckFunc(
					tfresource.TestCheckResourceAttr(name, "user_cap_cents."+uiUser, "0"),
					tfresource.TestCheckResourceAttr(name, "default_cap_cents.user", "1000"),
					holds(map[string]int{"USER": 1000}, map[string]int{uiUser: 0}, map[string]int{tok: 300}),
				),
			},
			// 3. Start managing users: the first plan shows uiUser's removal, and one apply converges
			// (the testing library fails the step if the plan after the apply is not empty).
			{
				Config: cfg(tokMap + fmt.Sprintf("\nuser_cap_cents = { %q = 100 }", a)),
				ConfigPlanChecks: tfresource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					beforeAndAfter{"user_cap_cents", map[string]any{uiUser: 0}, map[string]any{a: 100}},
				}},
				Check: holds(map[string]int{"USER": 1000}, map[string]int{a: 100}, map[string]int{tok: 300}),
			},
			// 4. Take the user map out of the config, while someone sets uiUser2 in the UI: empty plan,
			// nothing removed.
			{
				PreConfig:        func() { api.Caps["USER"][uiUser2] = 5 },
				Config:           cfg(tokMap),
				ConfigPlanChecks: tfresource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
				Check:            holds(map[string]int{"USER": 1000}, map[string]int{a: 100, uiUser2: 5}, map[string]int{tok: 300}),
			},
			// 5. Change something else while users are not managed: the user map stays known (its state
			// value), not "(known after apply)", and nothing is sent for users.
			{
				Config: cfg(fmt.Sprintf("api_token_cap_cents = { %q = 400 }", tok)),
				ConfigPlanChecks: tfresource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(name, plancheck.ResourceActionUpdate),
					plancheck.ExpectKnownValue(name, tfjsonpath.New("user_cap_cents"), knownvalue.MapExact(map[string]knownvalue.Check{
						a: knownvalue.Int64Exact(100), uiUser2: knownvalue.Int64Exact(5),
					})),
					plancheck.ExpectKnownValue(name, tfjsonpath.New("default_cap_cents"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"user": knownvalue.Int64Exact(1000), "api_token": knownvalue.Null(),
					})),
				}},
				Check: holds(map[string]int{"USER": 1000}, map[string]int{a: 100, uiUser2: 5}, map[string]int{tok: 400}),
			},
			// 6. Unmanaged to {}: the plan lists every override, and one apply removes them.
			{
				Config: cfg(fmt.Sprintf("api_token_cap_cents = { %q = 400 }\nuser_cap_cents = {}", tok)),
				ConfigPlanChecks: tfresource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					beforeAndAfter{"user_cap_cents", map[string]any{a: 100, uiUser2: 5}, map[string]any{}},
				}},
				Check: holds(map[string]int{"USER": 1000}, map[string]int{}, map[string]int{tok: 400}),
			},
			// 7. Defaults: set both.
			{
				Config: cfg(fmt.Sprintf("api_token_cap_cents = { %q = 400 }\ndefault_cap_cents = { user = 5000, api_token = 2500 }", tok)),
				Check:  holds(map[string]int{"USER": 5000, "SERVICEKEY": 2500}, map[string]int{}, map[string]int{tok: 400}),
			},
			// 8. A field left out of a set object is removed.
			{
				Config: cfg(fmt.Sprintf("api_token_cap_cents = { %q = 400 }\ndefault_cap_cents = { user = 5000 }", tok)),
				Check:  holds(map[string]int{"USER": 5000}, map[string]int{}, map[string]int{tok: 400}),
			},
			// 9. The object left out, while the UI changes the user default: empty plan.
			{
				PreConfig:        func() { api.Defaults["USER"] = 7000 },
				Config:           cfg(fmt.Sprintf("api_token_cap_cents = { %q = 400 }", tok)),
				ConfigPlanChecks: tfresource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
				Check:            holds(map[string]int{"USER": 7000}, map[string]int{}, map[string]int{tok: 400}),
			},
			// 10. {} removes both defaults in one apply.
			{
				Config: cfg(fmt.Sprintf("api_token_cap_cents = { %q = 400 }\ndefault_cap_cents = {}", tok)),
				Check:  holds(map[string]int{}, map[string]int{}, map[string]int{tok: 400}),
			},
			// 11. Import reads everything and matches the state.
			{ResourceName: name, ImportState: true, ImportStateId: "org", ImportStateVerify: true},
			// 12. The labs API not being turned on for the organization is an error, not a missing
			// resource that Terraform would plan to create again.
			{
				PreConfig:   func() { api.GateOff = true },
				Config:      cfg(fmt.Sprintf("api_token_cap_cents = { %q = 400 }", tok)),
				ExpectError: regexp.MustCompile(`(?s)AI spending caps are not available`),
			},
			// 13. Available again: the resource is still there, with nothing to change.
			{
				PreConfig:        func() { api.GateOff = false },
				Config:           cfg(fmt.Sprintf("api_token_cap_cents = { %q = 400 }", tok)),
				ConfigPlanChecks: tfresource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
			},
		},
	})
}

// TestUnit_ResourceAiSpendingCaps_TerraformUnknownValues covers the create check when a key or value
// in the configuration comes from another resource and is not known until apply. The plan cannot
// check it, so the check runs when Terraform plans again during the apply, before Create sends
// anything: a conflict fails the apply and leaves the caps as they were.
func TestUnit_ResourceAiSpendingCaps_TerraformUnknownValues(t *testing.T) {
	if _, err := exec.LookPath("terraform"); err != nil {
		t.Skip("terraform is not on the PATH; skipping the test that runs the real Terraform CLI")
	}
	existing, other := cuid.New(), cuid.New()
	// terraform_data's output is not known until it is created: lateId is other's ID, lateCap is 5000.
	lateValues := fmt.Sprintf(`
resource "terraform_data" "late_id" { input = %q }
resource "terraform_data" "late_cap" { input = 5000 }
`, other)
	for _, tc := range []struct {
		name         string
		defaults     map[string]int
		tokens       map[string]int
		body         string
		wantError    string
		wantDefaults map[string]int
		wantTokens   map[string]int
	}{
		{
			name:       "a key known at apply, with an existing override the map changes",
			tokens:     map[string]int{existing: 0},
			body:       fmt.Sprintf("api_token_cap_cents = {\n(terraform_data.late_id.output) = 100\n%q = 5000\n}", existing),
			wantError:  `(?s)Error running apply.*already has AI spending caps.*override of API token\s+` + existing,
			wantTokens: map[string]int{existing: 0},
		},
		{
			name:       "a value known at apply that changes an existing override",
			tokens:     map[string]int{existing: 0},
			body:       fmt.Sprintf("api_token_cap_cents = { %q = terraform_data.late_cap.output }", existing),
			wantError:  `(?s)Error running apply.*already has AI spending caps.*override of API token\s+` + existing,
			wantTokens: map[string]int{existing: 0},
		},
		{
			name:       "a value known at apply that matches the existing override",
			tokens:     map[string]int{existing: 5000},
			body:       fmt.Sprintf("api_token_cap_cents = { %q = terraform_data.late_cap.output }", existing),
			wantTokens: map[string]int{existing: 5000},
		},
		{
			name:       "a value known at apply, with no existing caps",
			body:       fmt.Sprintf("api_token_cap_cents = { %q = terraform_data.late_cap.output }", other),
			wantTokens: map[string]int{other: 5000},
		},
		{
			name:         "a default known at apply that changes the existing default",
			defaults:     map[string]int{"SERVICEKEY": 0},
			body:         "default_cap_cents = { api_token = terraform_data.late_cap.output }",
			wantError:    `(?s)Error running apply.*already has AI spending caps.*the API token default`,
			wantDefaults: map[string]int{"SERVICEKEY": 0},
		},
		{
			name:         "a default known at apply that matches the existing default",
			defaults:     map[string]int{"SERVICEKEY": 5000},
			body:         "default_cap_cents = { api_token = terraform_data.late_cap.output }",
			wantDefaults: map[string]int{"SERVICEKEY": 5000},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api := resources.NewFakeAiSpendingCapsAPI()
			srv := httptest.NewServer(api)
			t.Cleanup(srv.Close)
			labsClient, err := labs.NewLabsClient(srv.URL, "token", "test")
			require.NoError(t, err)
			for kind, value := range tc.defaults {
				api.Defaults[kind] = value
			}
			for id, value := range tc.tokens {
				api.Caps["SERVICEKEY"][id] = value
			}
			step := tfresource.TestStep{
				Config: lateValues + fmt.Sprintf("resource \"astro_ai_spending_caps\" \"test\" {\n%s\n}\n", tc.body),
			}
			if tc.wantError != "" {
				step.ExpectError = regexp.MustCompile(tc.wantError)
			}
			tfresource.UnitTest(t, tfresource.TestCase{
				ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
					"astro": providerserver.NewProtocol6WithError(&capsTestProvider{labsClient: labsClient}),
				},
				Steps: []tfresource.TestStep{step},
			})
			want := func(m map[string]int) map[string]int {
				if m == nil {
					return map[string]int{}
				}
				return m
			}
			assert.Equal(t, want(tc.wantDefaults), api.Defaults, "defaults")
			assert.Equal(t, want(tc.wantTokens), api.Caps["SERVICEKEY"], "API token caps")
		})
	}
}
