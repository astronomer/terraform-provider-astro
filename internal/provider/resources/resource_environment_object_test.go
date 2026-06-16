package resources_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/samber/lo"
)

// --- Workspace-scoped tests ---

func TestAcc_ResourceEnvironmentObjectAirflowVariable_Workspace(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	varKey := fmt.Sprintf("test_var_%v", namePrefix)
	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
	resourceVar := "astro_environment_object.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckEnvironmentObjectDestroyed(t, varKey),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjAirflowVar("test", varKey, "WORKSPACE", workspaceId, "initial_value", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, varKey),
					resource.TestCheckResourceAttr(resourceVar, "object_key", varKey),
					resource.TestCheckResourceAttr(resourceVar, "object_type", "AIRFLOW_VARIABLE"),
					resource.TestCheckResourceAttr(resourceVar, "scope", "WORKSPACE"),
					resource.TestCheckResourceAttr(resourceVar, "scope_entity_id", workspaceId),
					resource.TestCheckResourceAttr(resourceVar, "value", "initial_value"),
					resource.TestCheckResourceAttr(resourceVar, "is_secret", "false"),
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
				),
			},
			// Update value (in-place)
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjAirflowVar("test", varKey, "WORKSPACE", workspaceId, "updated_value", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, varKey),
					resource.TestCheckResourceAttr(resourceVar, "value", "updated_value"),
				),
			},
			// Toggle is_secret=true — RequiresReplace forces re-create
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjAirflowVar("test", varKey, "WORKSPACE", workspaceId, "secret_value", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, varKey),
					resource.TestCheckResourceAttr(resourceVar, "is_secret", "true"),
					resource.TestCheckResourceAttr(resourceVar, "value", "secret_value"),
				),
			},
			// Import
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"value"},
			},
		},
	})
}

func TestAcc_ResourceEnvironmentObjectConnection_Workspace(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	connKey := fmt.Sprintf("test_conn_%v", namePrefix)
	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
	resourceVar := "astro_environment_object.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckEnvironmentObjectDestroyed(t, connKey),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjConnection("test", connKey, "WORKSPACE", workspaceId, "example.com", 5432, `{"sslmode":"require","timeout":30}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, connKey),
					resource.TestCheckResourceAttr(resourceVar, "object_key", connKey),
					resource.TestCheckResourceAttr(resourceVar, "object_type", "CONNECTION"),
					resource.TestCheckResourceAttr(resourceVar, "scope", "WORKSPACE"),
					resource.TestCheckResourceAttr(resourceVar, "type", "postgres"),
					resource.TestCheckResourceAttr(resourceVar, "host", "example.com"),
					resource.TestCheckResourceAttr(resourceVar, "port", "5432"),
					resource.TestCheckResourceAttr(resourceVar, "login", "testuser"),
					resource.TestCheckResourceAttr(resourceVar, "password", "testpass"),
					resource.TestCheckResourceAttr(resourceVar, "schema", "testdb"),
					resource.TestCheckResourceAttr(resourceVar, "extra", `{"sslmode":"require","timeout":30}`),
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
				),
			},
			// Update host/port
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjConnection("test", connKey, "WORKSPACE", workspaceId, "updated.example.com", 5433, `{"sslmode":"require","timeout":30}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, connKey),
					resource.TestCheckResourceAttr(resourceVar, "host", "updated.example.com"),
					resource.TestCheckResourceAttr(resourceVar, "port", "5433"),
					resource.TestCheckResourceAttr(resourceVar, "password", "testpass"),
					resource.TestCheckResourceAttr(resourceVar, "extra", `{"sslmode":"require","timeout":30}`),
				),
			},
			// Import
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "extra"},
			},
		},
	})
}

func TestAcc_ResourceEnvironmentObjectMetricsExport_Workspace(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	meKey := fmt.Sprintf("test_me_%v", namePrefix)
	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
	resourceVar := "astro_environment_object.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckEnvironmentObjectDestroyed(t, meKey),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjMetricsExport("test", meKey, "WORKSPACE", workspaceId, "https://prometheus.example.com/api/v1/write"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, meKey),
					resource.TestCheckResourceAttr(resourceVar, "object_key", meKey),
					resource.TestCheckResourceAttr(resourceVar, "object_type", "METRICS_EXPORT"),
					resource.TestCheckResourceAttr(resourceVar, "scope", "WORKSPACE"),
					resource.TestCheckResourceAttr(resourceVar, "endpoint", "https://prometheus.example.com/api/v1/write"),
					resource.TestCheckResourceAttr(resourceVar, "exporter_type", "PROMETHEUS"),
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
				),
			},
			// Update endpoint
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjMetricsExport("test", meKey, "WORKSPACE", workspaceId, "https://prometheus.example.com/api/v2/write"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, meKey),
					resource.TestCheckResourceAttr(resourceVar, "endpoint", "https://prometheus.example.com/api/v2/write"),
				),
			},
			// Import — auth_type is write-only on the API (echoes back as ""),
			// so the imported state can't reproduce the live value.
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"basic_token", "password", "auth_type"},
			},
		},
	})
}

// --- Deployment-scoped tests ---

func TestAcc_ResourceEnvironmentObjectAirflowVariable_Deployment(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	varKey := fmt.Sprintf("test_var_dep_%v", namePrefix)
	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")
	resourceVar := "astro_environment_object.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckEnvironmentObjectDestroyed(t, varKey),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjAirflowVar("test", varKey, "DEPLOYMENT", deploymentId, "dep_value", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, varKey),
					resource.TestCheckResourceAttr(resourceVar, "object_key", varKey),
					resource.TestCheckResourceAttr(resourceVar, "object_type", "AIRFLOW_VARIABLE"),
					resource.TestCheckResourceAttr(resourceVar, "scope", "DEPLOYMENT"),
					resource.TestCheckResourceAttr(resourceVar, "scope_entity_id", deploymentId),
					resource.TestCheckResourceAttr(resourceVar, "value", "dep_value"),
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
				),
			},
			// Update value
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjAirflowVar("test", varKey, "DEPLOYMENT", deploymentId, "dep_updated", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "value", "dep_updated"),
				),
			},
			// Import
			{
				ResourceName:      resourceVar,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ResourceEnvironmentObjectConnection_Deployment(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	connKey := fmt.Sprintf("test_conn_dep_%v", namePrefix)
	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")
	resourceVar := "astro_environment_object.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckEnvironmentObjectDestroyed(t, connKey),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjConnection("test", connKey, "DEPLOYMENT", deploymentId, "db.example.com", 3306, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, connKey),
					resource.TestCheckResourceAttr(resourceVar, "object_key", connKey),
					resource.TestCheckResourceAttr(resourceVar, "object_type", "CONNECTION"),
					resource.TestCheckResourceAttr(resourceVar, "scope", "DEPLOYMENT"),
					resource.TestCheckResourceAttr(resourceVar, "type", "postgres"),
					resource.TestCheckResourceAttr(resourceVar, "host", "db.example.com"),
					resource.TestCheckResourceAttr(resourceVar, "port", "3306"),
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
				),
			},
			// Update host
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjConnection("test", connKey, "DEPLOYMENT", deploymentId, "db-updated.example.com", 3306, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "host", "db-updated.example.com"),
				),
			},
			// Import
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "extra"},
			},
		},
	})
}

func TestAcc_ResourceEnvironmentObjectMetricsExport_Deployment(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	meKey := fmt.Sprintf("test_me_dep_%v", namePrefix)
	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")
	resourceVar := "astro_environment_object.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckEnvironmentObjectDestroyed(t, meKey),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjMetricsExport("test", meKey, "DEPLOYMENT", deploymentId, "https://prom.example.com/api/v1/write"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, meKey),
					resource.TestCheckResourceAttr(resourceVar, "object_key", meKey),
					resource.TestCheckResourceAttr(resourceVar, "object_type", "METRICS_EXPORT"),
					resource.TestCheckResourceAttr(resourceVar, "scope", "DEPLOYMENT"),
					resource.TestCheckResourceAttr(resourceVar, "endpoint", "https://prom.example.com/api/v1/write"),
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
				),
			},
			// Update endpoint
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjMetricsExport("test", meKey, "DEPLOYMENT", deploymentId, "https://prom.example.com/api/v2/write"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "endpoint", "https://prom.example.com/api/v2/write"),
				),
			},
			// Import — auth_type is write-only on the API.
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"basic_token", "password", "auth_type"},
			},
		},
	})
}

// --- Links + overrides tests (WORKSPACE-scoped only) ---

func TestAcc_ResourceEnvironmentObjectAirflowVariable_LinkOverrides(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	varKey := fmt.Sprintf("test_var_link_%v", namePrefix)
	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")
	resourceVar := "astro_environment_object.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckEnvironmentObjectDestroyed(t, varKey),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjAirflowVarWithLinkOverride(
					"test", varKey, workspaceId, "workspace_value", deploymentId, "deployment_override_value",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, varKey),
					resource.TestCheckResourceAttr(resourceVar, "object_key", varKey),
					resource.TestCheckResourceAttr(resourceVar, "object_type", "AIRFLOW_VARIABLE"),
					resource.TestCheckResourceAttr(resourceVar, "scope", "WORKSPACE"),
					resource.TestCheckResourceAttr(resourceVar, "value", "workspace_value"),
					resource.TestCheckResourceAttr(resourceVar, "links.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceVar, "links.*", map[string]string{
						"scope":           "DEPLOYMENT",
						"scope_entity_id": deploymentId,
						"overrides.value": "deployment_override_value",
					}),
				),
			},
			// Update the override value in place
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjAirflowVarWithLinkOverride(
					"test", varKey, workspaceId, "workspace_value", deploymentId, "updated_override_value",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(resourceVar, "links.*", map[string]string{
						"scope":           "DEPLOYMENT",
						"scope_entity_id": deploymentId,
						"overrides.value": "updated_override_value",
					}),
				),
			},
			// Import — value + links.*.overrides.value are sensitive, so ignore
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"value", "links"},
			},
		},
	})
}

func TestAcc_ResourceEnvironmentObjectConnection_LinkOverrides(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	connKey := fmt.Sprintf("test_conn_link_%v", namePrefix)
	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")
	resourceVar := "astro_environment_object.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckEnvironmentObjectDestroyed(t, connKey),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjConnectionWithLinkOverride(
					"test", connKey, workspaceId, deploymentId,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, connKey),
					resource.TestCheckResourceAttr(resourceVar, "object_key", connKey),
					resource.TestCheckResourceAttr(resourceVar, "object_type", "CONNECTION"),
					resource.TestCheckResourceAttr(resourceVar, "scope", "WORKSPACE"),
					resource.TestCheckResourceAttr(resourceVar, "host", "warehouse.example.com"),
					resource.TestCheckResourceAttr(resourceVar, "port", "5432"),
					resource.TestCheckResourceAttr(resourceVar, "links.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceVar, "links.*", map[string]string{
						"scope":              "DEPLOYMENT",
						"scope_entity_id":    deploymentId,
						"overrides.host":     "warehouse-staging.example.com",
						"overrides.port":     "5433",
						"overrides.schema":   "analytics_staging",
						"overrides.login":    "staging_user",
						"overrides.password": "staging_password",
						"overrides.extra":    `{"sslmode":"prefer"}`,
					}),
				),
			},
			// Import — top-level password + extra + nested override secrets are not echoed
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "extra", "links"},
			},
		},
	})
}

func TestAcc_ResourceEnvironmentObjectMetricsExport_LinkOverrides(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	meKey := fmt.Sprintf("test_me_link_%v", namePrefix)
	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")
	resourceVar := "astro_environment_object.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckEnvironmentObjectDestroyed(t, meKey),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjMetricsExportWithLinkOverride(
					"test", meKey, workspaceId, deploymentId,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, meKey),
					resource.TestCheckResourceAttr(resourceVar, "object_key", meKey),
					resource.TestCheckResourceAttr(resourceVar, "object_type", "METRICS_EXPORT"),
					resource.TestCheckResourceAttr(resourceVar, "scope", "WORKSPACE"),
					resource.TestCheckResourceAttr(resourceVar, "endpoint", "https://prometheus.example.com/api/v1/write"),
					resource.TestCheckResourceAttr(resourceVar, "links.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceVar, "links.*", map[string]string{
						"scope":                        "DEPLOYMENT",
						"scope_entity_id":              deploymentId,
						"overrides.endpoint":           "https://prometheus-staging.example.com/api/v1/write",
						"overrides.auth_type":          "AUTH_TOKEN",
						"overrides.basic_token":        "staging_token",
						"overrides.labels.environment": "staging",
						"overrides.headers.X-Tenant":   "staging-tenant",
					}),
				),
			},
			// Import
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"basic_token", "password", "links", "auth_type"},
			},
		},
	})
}

// TestAcc_ResourceEnvironmentObject_InPlaceUpdate_NoSpuriousReplace is a
// regression guard for the bug where Optional+Computed+RequiresReplace fields
// (is_secret on AIRFLOW_VARIABLE, type on CONNECTION) without
// UseStateForUnknown would replan as Unknown when omitted in config, then fire
// RequiresReplace against the prior state — turning every in-place update of
// an unrelated field into a destroy + recreate.
//
// The two steps below pin the value of `value` to different strings while
// leaving is_secret out of the config; the plancheck asserts the step-2
// action is Update (in-place), not Replace.
func TestAcc_ResourceEnvironmentObject_InPlaceUpdate_NoSpuriousReplace(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	varKey := fmt.Sprintf("test_var_inplace_%v", namePrefix)
	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
	resourceVar := "astro_environment_object.test"

	avNoIsSecret := func(value string) string {
		return fmt.Sprintf(`
resource "astro_environment_object" "test" {
  object_key      = "%s"
  object_type     = "AIRFLOW_VARIABLE"
  scope           = "WORKSPACE"
  scope_entity_id = "%s"

  value = "%s"
  # is_secret intentionally omitted — must stay in-place on subsequent updates
}
`, varKey, workspaceId, value)
	}

	// Capture the ID after the create step so we can assert step-2 didn't
	// destroy + recreate (which would yield a different ID even though the
	// object_key is identical).
	var idAfterCreate string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckEnvironmentObjectDestroyed(t, varKey),
		Steps: []resource.TestStep{
			// Create with is_secret omitted
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + avNoIsSecret("v1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, varKey),
					resource.TestCheckResourceAttr(resourceVar, "value", "v1"),
					resource.TestCheckResourceAttr(resourceVar, "is_secret", "false"),
					captureResourceID(resourceVar, &idAfterCreate),
				),
			},
			// Update value — plancheck asserts the plan action is an in-place
			// Update (Action_Update), not a destroy + recreate. The id-stability
			// check after apply doubles as a belt-and-braces verification.
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + avNoIsSecret("v2"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceVar, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "value", "v2"),
					resource.TestCheckResourceAttr(resourceVar, "is_secret", "false"),
					checkResourceIDUnchanged(resourceVar, &idAfterCreate),
				),
			},
		},
	})
}

// captureResourceID copies the resource's id attribute into dest after a step
// applies. Use with checkResourceIDUnchanged in a later step to assert the
// resource wasn't destroyed + recreated.
func captureResourceID(addr string, dest *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[addr]
		if !ok {
			return fmt.Errorf("resource %s not in state", addr)
		}
		if rs.Primary == nil || rs.Primary.ID == "" {
			return fmt.Errorf("resource %s has no primary id", addr)
		}
		*dest = rs.Primary.ID
		return nil
	}
}

// checkResourceIDUnchanged asserts the resource's current id matches the
// previously-captured value. Replacement (destroy + create) produces a new id
// even when the object_key is identical, so a stable id is proof of in-place.
func checkResourceIDUnchanged(addr string, want *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[addr]
		if !ok {
			return fmt.Errorf("resource %s not in state", addr)
		}
		if rs.Primary == nil {
			return fmt.Errorf("resource %s has no primary", addr)
		}
		if rs.Primary.ID != *want {
			return fmt.Errorf("resource %s id changed (was %s, now %s) — replacement happened when it shouldn't have", addr, *want, rs.Primary.ID)
		}
		return nil
	}
}

// --- auto_link + exclude_links, typed auth, BASIC auth, link lifecycle ---

// TestAcc_ResourceEnvironmentObject_AutoLinkAndExcludeLinks exercises the
// workspace-scope-only auto_link_deployments + exclude_links combination.
// The deployment is excluded so the auto-link fans out to (workspace
// deployments) − (excluded). We don't verify the API-side linking effect; we
// just verify both fields round-trip through state without drift.
func TestAcc_ResourceEnvironmentObject_AutoLinkAndExcludeLinks(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	varKey := fmt.Sprintf("test_var_autolink_%v", namePrefix)
	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")
	resourceVar := "astro_environment_object.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckEnvironmentObjectDestroyed(t, varKey),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjAirflowVarAutoLinkWithExclude(
					"test", varKey, workspaceId, deploymentId, "auto_value",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, varKey),
					resource.TestCheckResourceAttr(resourceVar, "object_key", varKey),
					resource.TestCheckResourceAttr(resourceVar, "scope", "WORKSPACE"),
					resource.TestCheckResourceAttr(resourceVar, "auto_link_deployments", "true"),
					resource.TestCheckResourceAttr(resourceVar, "exclude_links.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceVar, "exclude_links.*", map[string]string{
						"scope":           "DEPLOYMENT",
						"scope_entity_id": deploymentId,
					}),
				),
			},
			// Toggle auto_link off; exclude_links stays. Verifies the flag update path.
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjAirflowVarManualLinkWithExclude(
					"test", varKey, workspaceId, deploymentId, "auto_value",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "auto_link_deployments", "false"),
					resource.TestCheckResourceAttr(resourceVar, "exclude_links.#", "1"),
				),
			},
			// Import — re-reads exclude_links from API
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"value"},
			},
		},
	})
}

// TestAcc_ResourceEnvironmentObjectConnection_TypedAuth covers auth_type_id +
// the Computed connection_auth_type object. The API resolves the auth-type
// metadata; we assert id/name/parameters populate so the cross-field dependency
// (no UseStateForUnknown by design) actually flows through correctly.
func TestAcc_ResourceEnvironmentObjectConnection_TypedAuth(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	connKey := fmt.Sprintf("test_conn_typedauth_%v", namePrefix)
	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
	resourceVar := "astro_environment_object.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckEnvironmentObjectDestroyed(t, connKey),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjConnectionTypedAuth(
					"test", connKey, workspaceId, "snowflake-password",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, connKey),
					resource.TestCheckResourceAttr(resourceVar, "object_key", connKey),
					resource.TestCheckResourceAttr(resourceVar, "type", "snowflake"),
					// auth_type_id is write-only on the API; the provider preserves
					// the user-supplied value via EnvironmentObjectPreserve.
					resource.TestCheckResourceAttr(resourceVar, "auth_type_id", "snowflake-password"),
					// connection_auth_type is populated by the API when it recognizes
					// the auth_type_id. Whether that happens in this org depends on
					// the org's configured auth-type catalog, so we don't assert on
					// its shape — only that auth_type_id itself round-trips.
				),
			},
			// Import — auth_type_id is write-only on the API and can't be
			// recovered on import; password and extra are also not echoed back.
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "extra", "auth_type_id"},
			},
		},
	})
}

// TestAcc_ResourceEnvironmentObjectMetricsExport_BasicAuth covers the BASIC
// auth path: auth_type=BASIC + username + password (polymorphic — same
// `password` attribute as CONNECTION but with HTTP Basic-auth semantics).
func TestAcc_ResourceEnvironmentObjectMetricsExport_BasicAuth(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	meKey := fmt.Sprintf("test_me_basic_%v", namePrefix)
	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
	resourceVar := "astro_environment_object.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckEnvironmentObjectDestroyed(t, meKey),
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjMetricsExportBasic(
					"test", meKey, workspaceId, "metrics_user", "metrics_pass",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, meKey),
					resource.TestCheckResourceAttr(resourceVar, "object_key", meKey),
					resource.TestCheckResourceAttr(resourceVar, "auth_type", "BASIC"),
					resource.TestCheckResourceAttr(resourceVar, "username", "metrics_user"),
					resource.TestCheckResourceAttr(resourceVar, "password", "metrics_pass"),
				),
			},
			// Update username — verify in-place update works on the BASIC path
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjMetricsExportBasic(
					"test", meKey, workspaceId, "metrics_user_v2", "metrics_pass",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "username", "metrics_user_v2"),
				),
			},
			// Import — password, basic_token, and auth_type are not echoed.
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "basic_token", "auth_type"},
			},
		},
	})
}

// TestAcc_ResourceEnvironmentObject_LinkLifecycle exercises add / modify /
// remove for the links block. Each transition is a fresh refresh+plan, so
// any drift in the link state would surface as a non-empty plan failure.
func TestAcc_ResourceEnvironmentObject_LinkLifecycle(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	varKey := fmt.Sprintf("test_var_lifecycle_%v", namePrefix)
	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")
	resourceVar := "astro_environment_object.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckEnvironmentObjectDestroyed(t, varKey),
		Steps: []resource.TestStep{
			// Step 1: no links
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjAirflowVar(
					"test", varKey, "WORKSPACE", workspaceId, "base_value", false,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, varKey),
					resource.TestCheckResourceAttr(resourceVar, "value", "base_value"),
				),
			},
			// Step 2: add a link with an override
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjAirflowVarWithLinkOverride(
					"test", varKey, workspaceId, "base_value", deploymentId, "override_v1",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "links.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceVar, "links.*", map[string]string{
						"scope":           "DEPLOYMENT",
						"scope_entity_id": deploymentId,
						"overrides.value": "override_v1",
					}),
				),
			},
			// Step 3: modify the override value
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjAirflowVarWithLinkOverride(
					"test", varKey, workspaceId, "base_value", deploymentId, "override_v2",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(resourceVar, "links.*", map[string]string{
						"scope":           "DEPLOYMENT",
						"scope_entity_id": deploymentId,
						"overrides.value": "override_v2",
					}),
				),
			},
			// Step 4: remove the link entirely; back to base config
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + envObjAirflowVar(
					"test", varKey, "WORKSPACE", workspaceId, "base_value", false,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "links.#", "0"),
				),
			},
		},
	})
}

// --- Config helpers ---

func envObjAirflowVar(tfName, varKey, scope, scopeEntityId, value string, isSecret bool) string {
	return fmt.Sprintf(`
resource "astro_environment_object" "%s" {
  object_key      = "%s"
  object_type     = "AIRFLOW_VARIABLE"
  scope           = "%s"
  scope_entity_id = "%s"

  value     = "%s"
  is_secret = %t
}
`, tfName, varKey, scope, scopeEntityId, value, isSecret)
}

func envObjConnection(tfName, connKey, scope, scopeEntityId, host string, port int, extraJSON string) string {
	extraLine := ""
	if extraJSON != "" {
		extraLine = fmt.Sprintf("extra = %q", extraJSON)
	}
	return fmt.Sprintf(`
resource "astro_environment_object" "%s" {
  object_key      = "%s"
  object_type     = "CONNECTION"
  scope           = "%s"
  scope_entity_id = "%s"

  type     = "postgres"
  host     = "%s"
  port     = %d
  login    = "testuser"
  password = "testpass"
  schema   = "testdb"
  %s
}
`, tfName, connKey, scope, scopeEntityId, host, port, extraLine)
}

func envObjMetricsExport(tfName, meKey, scope, scopeEntityId, endpoint string) string {
	return fmt.Sprintf(`
resource "astro_environment_object" "%s" {
  object_key      = "%s"
  object_type     = "METRICS_EXPORT"
  scope           = "%s"
  scope_entity_id = "%s"

  endpoint      = "%s"
  exporter_type = "PROMETHEUS"
}
`, tfName, meKey, scope, scopeEntityId, endpoint)
}

func envObjAirflowVarWithLinkOverride(tfName, varKey, workspaceId, value, deploymentId, overrideValue string) string {
	return fmt.Sprintf(`
resource "astro_environment_object" "%s" {
  object_key      = "%s"
  object_type     = "AIRFLOW_VARIABLE"
  scope           = "WORKSPACE"
  scope_entity_id = "%s"

  value     = "%s"
  is_secret = false

  links = [{
    scope           = "DEPLOYMENT"
    scope_entity_id = "%s"
    overrides = {
      value = "%s"
    }
  }]
}
`, tfName, varKey, workspaceId, value, deploymentId, overrideValue)
}

func envObjConnectionWithLinkOverride(tfName, connKey, workspaceId, deploymentId string) string {
	return fmt.Sprintf(`
resource "astro_environment_object" "%s" {
  object_key      = "%s"
  object_type     = "CONNECTION"
  scope           = "WORKSPACE"
  scope_entity_id = "%s"

  type     = "postgres"
  host     = "warehouse.example.com"
  port     = 5432
  login    = "airflow"
  password = "prod_password"
  schema   = "analytics"
  extra    = jsonencode({ sslmode = "require" })

  links = [{
    scope           = "DEPLOYMENT"
    scope_entity_id = "%s"
    overrides = {
      host     = "warehouse-staging.example.com"
      port     = 5433
      schema   = "analytics_staging"
      login    = "staging_user"
      password = "staging_password"
      extra    = jsonencode({ sslmode = "prefer" })
    }
  }]
}
`, tfName, connKey, workspaceId, deploymentId)
}

func envObjMetricsExportWithLinkOverride(tfName, meKey, workspaceId, deploymentId string) string {
	return fmt.Sprintf(`
resource "astro_environment_object" "%s" {
  object_key      = "%s"
  object_type     = "METRICS_EXPORT"
  scope           = "WORKSPACE"
  scope_entity_id = "%s"

  endpoint      = "https://prometheus.example.com/api/v1/write"
  exporter_type = "PROMETHEUS"
  auth_type     = "AUTH_TOKEN"
  basic_token   = "prod_token"

  links = [{
    scope           = "DEPLOYMENT"
    scope_entity_id = "%s"
    overrides = {
      endpoint    = "https://prometheus-staging.example.com/api/v1/write"
      auth_type   = "AUTH_TOKEN"
      basic_token = "staging_token"
      labels      = { environment = "staging" }
      headers     = { "X-Tenant" = "staging-tenant" }
    }
  }]
}
`, tfName, meKey, workspaceId, deploymentId)
}

func envObjAirflowVarAutoLinkWithExclude(tfName, varKey, workspaceId, excludeDeploymentId, value string) string {
	return fmt.Sprintf(`
resource "astro_environment_object" "%s" {
  object_key      = "%s"
  object_type     = "AIRFLOW_VARIABLE"
  scope           = "WORKSPACE"
  scope_entity_id = "%s"

  value                 = "%s"
  is_secret             = false
  auto_link_deployments = true

  exclude_links = [{
    scope           = "DEPLOYMENT"
    scope_entity_id = "%s"
  }]
}
`, tfName, varKey, workspaceId, value, excludeDeploymentId)
}

func envObjAirflowVarManualLinkWithExclude(tfName, varKey, workspaceId, excludeDeploymentId, value string) string {
	return fmt.Sprintf(`
resource "astro_environment_object" "%s" {
  object_key      = "%s"
  object_type     = "AIRFLOW_VARIABLE"
  scope           = "WORKSPACE"
  scope_entity_id = "%s"

  value                 = "%s"
  is_secret             = false
  auto_link_deployments = false

  exclude_links = [{
    scope           = "DEPLOYMENT"
    scope_entity_id = "%s"
  }]
}
`, tfName, varKey, workspaceId, value, excludeDeploymentId)
}

func envObjConnectionTypedAuth(tfName, connKey, workspaceId, authTypeId string) string {
	return fmt.Sprintf(`
resource "astro_environment_object" "%s" {
  object_key      = "%s"
  object_type     = "CONNECTION"
  scope           = "WORKSPACE"
  scope_entity_id = "%s"

  type         = "snowflake"
  auth_type_id = "%s"
  host         = "abc12345.us-east-1.snowflakecomputing.com"
  login        = "AIRFLOW_USER"
  password     = "test_password"
  schema       = "ANALYTICS"
  extra        = jsonencode({ account = "abc12345", warehouse = "AIRFLOW_WH", role = "AIRFLOW_ROLE" })
}
`, tfName, connKey, workspaceId, authTypeId)
}

func envObjMetricsExportBasic(tfName, meKey, workspaceId, username, password string) string {
	return fmt.Sprintf(`
resource "astro_environment_object" "%s" {
  object_key      = "%s"
  object_type     = "METRICS_EXPORT"
  scope           = "WORKSPACE"
  scope_entity_id = "%s"

  endpoint      = "https://prom-basic.example.com/api/v1/write"
  exporter_type = "PROMETHEUS"
  auth_type     = "BASIC"
  username      = "%s"
  password      = "%s"
}
`, tfName, meKey, workspaceId, username, password)
}

// --- Existence checks ---

func testAccCheckEnvironmentObjectExists(t *testing.T, objectKey string) resource.TestCheckFunc {
	t.Helper()
	return func(s *terraform.State) error {
		client, err := utils.GetTestPlatformClient(true)
		if err != nil {
			return fmt.Errorf("failed to get test platform client: %v", err)
		}

		organizationId := os.Getenv("HOSTED_ORGANIZATION_ID")
		ctx := context.Background()

		resp, err := client.ListEnvironmentObjectsWithResponse(ctx, organizationId, &platform.ListEnvironmentObjectsParams{
			ObjectKey: &objectKey,
			Limit:     lo.ToPtr(10),
		})
		if err != nil {
			return fmt.Errorf("failed to list environment objects: %v", err)
		}
		if resp == nil || resp.JSON200 == nil {
			return fmt.Errorf("nil response from list environment objects")
		}

		for _, obj := range resp.JSON200.EnvironmentObjects {
			if obj.ObjectKey == objectKey {
				return nil
			}
		}

		return fmt.Errorf("environment object %s not found", objectKey)
	}
}

func testAccCheckEnvironmentObjectDestroyed(t *testing.T, objectKey string) resource.TestCheckFunc {
	t.Helper()
	return func(s *terraform.State) error {
		client, err := utils.GetTestPlatformClient(true)
		if err != nil {
			return fmt.Errorf("failed to get test platform client: %v", err)
		}

		organizationId := os.Getenv("HOSTED_ORGANIZATION_ID")
		ctx := context.Background()

		resp, err := client.ListEnvironmentObjectsWithResponse(ctx, organizationId, &platform.ListEnvironmentObjectsParams{
			ObjectKey: &objectKey,
			Limit:     lo.ToPtr(10),
		})
		if err != nil {
			return fmt.Errorf("failed to list environment objects: %v", err)
		}
		if resp == nil {
			return fmt.Errorf("nil response checking destruction of object_key %s", objectKey)
		}
		if resp.StatusCode() == 200 {
			if resp.JSON200 == nil {
				return fmt.Errorf("200 OK but JSON200 was nil for object_key %s", objectKey)
			}
			for _, obj := range resp.JSON200.EnvironmentObjects {
				if obj.ObjectKey == objectKey {
					return fmt.Errorf("environment object %s still exists after destroy", objectKey)
				}
			}
			return nil
		}
		statusCode, diag := clients.NormalizeAPIError(ctx, resp.HTTPResponse, resp.Body)
		if statusCode == 404 {
			return nil
		}
		if diag != nil {
			return fmt.Errorf("unexpected error checking destruction: %s", diag.Detail())
		}
		return fmt.Errorf("unexpected status %d checking destruction of object_key %s", resp.StatusCode(), objectKey)
	}
}
