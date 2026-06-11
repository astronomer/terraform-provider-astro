package datasources_test

import (
	"fmt"
	"os"
	"testing"

	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAcc_DataSource_EnvironmentObject exercises the single-item data source
// for every object_type × scope combination, plus link-with-overrides on
// WORKSPACE-scoped objects. The data source ID comes from a sibling resource
// declaration in the same plan, so the read happens immediately after create.
func TestAcc_DataSource_EnvironmentObject(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")

	varWsKey := fmt.Sprintf("ds_av_ws_%s", namePrefix)
	varDepKey := fmt.Sprintf("ds_av_dep_%s", namePrefix)
	connWsKey := fmt.Sprintf("ds_conn_ws_%s", namePrefix)
	connDepKey := fmt.Sprintf("ds_conn_dep_%s", namePrefix)
	meWsKey := fmt.Sprintf("ds_me_ws_%s", namePrefix)
	meDepKey := fmt.Sprintf("ds_me_dep_%s", namePrefix)

	connTypedKey := fmt.Sprintf("ds_conn_typed_%s", namePrefix)

	config := astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) +
		envObjAirflowVarResourceAndDS("av_ws", varWsKey, "WORKSPACE", workspaceId, "ws_value", false) +
		envObjAirflowVarResourceAndDS("av_dep", varDepKey, "DEPLOYMENT", deploymentId, "dep_value", false) +
		envObjConnectionResourceAndDS("conn_ws", connWsKey, "WORKSPACE", workspaceId) +
		envObjConnectionResourceAndDS("conn_dep", connDepKey, "DEPLOYMENT", deploymentId) +
		envObjMetricsExportResourceAndDS("me_ws", meWsKey, "WORKSPACE", workspaceId) +
		envObjMetricsExportResourceAndDS("me_dep", meDepKey, "DEPLOYMENT", deploymentId) +
		envObjAirflowVarWithLinkResourceAndDS("av_link", fmt.Sprintf("ds_av_link_%s", namePrefix), workspaceId, deploymentId) +
		envObjConnectionTypedAuthResourceAndDS("conn_typed", connTypedKey, workspaceId, "snowflake-password")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					// AIRFLOW_VARIABLE × WORKSPACE
					checkEnvObjDataSourceCommon("data.astro_environment_object.av_ws", varWsKey, "AIRFLOW_VARIABLE", "WORKSPACE", workspaceId),
					resource.TestCheckResourceAttr("data.astro_environment_object.av_ws", "value", "ws_value"),
					resource.TestCheckResourceAttr("data.astro_environment_object.av_ws", "is_secret", "false"),

					// AIRFLOW_VARIABLE × DEPLOYMENT
					checkEnvObjDataSourceCommon("data.astro_environment_object.av_dep", varDepKey, "AIRFLOW_VARIABLE", "DEPLOYMENT", deploymentId),
					resource.TestCheckResourceAttr("data.astro_environment_object.av_dep", "value", "dep_value"),

					// CONNECTION × WORKSPACE
					checkEnvObjDataSourceCommon("data.astro_environment_object.conn_ws", connWsKey, "CONNECTION", "WORKSPACE", workspaceId),
					resource.TestCheckResourceAttr("data.astro_environment_object.conn_ws", "type", "postgres"),
					resource.TestCheckResourceAttr("data.astro_environment_object.conn_ws", "host", "ds.example.com"),
					resource.TestCheckResourceAttr("data.astro_environment_object.conn_ws", "port", "5432"),
					resource.TestCheckResourceAttr("data.astro_environment_object.conn_ws", "schema", "analytics"),
					resource.TestCheckResourceAttr("data.astro_environment_object.conn_ws", "login", "ds_user"),

					// CONNECTION × DEPLOYMENT
					checkEnvObjDataSourceCommon("data.astro_environment_object.conn_dep", connDepKey, "CONNECTION", "DEPLOYMENT", deploymentId),
					resource.TestCheckResourceAttr("data.astro_environment_object.conn_dep", "type", "postgres"),

					// METRICS_EXPORT × WORKSPACE
					checkEnvObjDataSourceCommon("data.astro_environment_object.me_ws", meWsKey, "METRICS_EXPORT", "WORKSPACE", workspaceId),
					resource.TestCheckResourceAttr("data.astro_environment_object.me_ws", "endpoint", "https://ds-prom.example.com/api/v1/write"),
					resource.TestCheckResourceAttr("data.astro_environment_object.me_ws", "exporter_type", "PROMETHEUS"),

					// METRICS_EXPORT × DEPLOYMENT
					checkEnvObjDataSourceCommon("data.astro_environment_object.me_dep", meDepKey, "METRICS_EXPORT", "DEPLOYMENT", deploymentId),
					resource.TestCheckResourceAttr("data.astro_environment_object.me_dep", "endpoint", "https://ds-prom.example.com/api/v1/write"),

					// AIRFLOW_VARIABLE with link override — verify overrides round-trip via data source
					checkEnvObjDataSourceCommon("data.astro_environment_object.av_link", fmt.Sprintf("ds_av_link_%s", namePrefix), "AIRFLOW_VARIABLE", "WORKSPACE", workspaceId),
					resource.TestCheckResourceAttr("data.astro_environment_object.av_link", "links.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("data.astro_environment_object.av_link", "links.*", map[string]string{
						"scope":           "DEPLOYMENT",
						"scope_entity_id": deploymentId,
						"overrides.value": "ds_override_value",
					}),

					// CONNECTION with typed auth — verify auth_type_id is preserved
					// across the read path. The resolved connection_auth_type
					// nested object depends on the org's auth-type catalog so
					// we don't assert on its shape.
					checkEnvObjDataSourceCommon("data.astro_environment_object.conn_typed", connTypedKey, "CONNECTION", "WORKSPACE", workspaceId),
					resource.TestCheckResourceAttr("data.astro_environment_object.conn_typed", "type", "snowflake"),
				),
			},
		},
	})
}

// checkEnvObjDataSourceCommon asserts the common attributes every data source
// instance should expose. Per-type fields are checked at the call site.
func checkEnvObjDataSourceCommon(dsAddr, objectKey, objectType, scope, scopeEntityId string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttrSet(dsAddr, "id"),
		resource.TestCheckResourceAttr(dsAddr, "object_key", objectKey),
		resource.TestCheckResourceAttr(dsAddr, "object_type", objectType),
		resource.TestCheckResourceAttr(dsAddr, "scope", scope),
		resource.TestCheckResourceAttr(dsAddr, "scope_entity_id", scopeEntityId),
		resource.TestCheckResourceAttrSet(dsAddr, "created_at"),
		resource.TestCheckResourceAttrSet(dsAddr, "updated_at"),
		resource.TestCheckResourceAttrSet(dsAddr, "created_by.id"),
		resource.TestCheckResourceAttrSet(dsAddr, "updated_by.id"),
	)
}

// --- Config helpers (resource + sibling data source) ---

func envObjAirflowVarResourceAndDS(name, key, scope, scopeEntityId, value string, isSecret bool) string {
	return fmt.Sprintf(`
resource "astro_environment_object" "%s" {
  object_key      = "%s"
  object_type     = "AIRFLOW_VARIABLE"
  scope           = "%s"
  scope_entity_id = "%s"

  value     = "%s"
  is_secret = %t
}

data "astro_environment_object" "%s" {
  id = astro_environment_object.%s.id
}
`, name, key, scope, scopeEntityId, value, isSecret, name, name)
}

func envObjConnectionResourceAndDS(name, key, scope, scopeEntityId string) string {
	return fmt.Sprintf(`
resource "astro_environment_object" "%s" {
  object_key      = "%s"
  object_type     = "CONNECTION"
  scope           = "%s"
  scope_entity_id = "%s"

  type     = "postgres"
  host     = "ds.example.com"
  port     = 5432
  login    = "ds_user"
  password = "ds_pass"
  schema   = "analytics"
}

data "astro_environment_object" "%s" {
  id = astro_environment_object.%s.id
}
`, name, key, scope, scopeEntityId, name, name)
}

func envObjMetricsExportResourceAndDS(name, key, scope, scopeEntityId string) string {
	return fmt.Sprintf(`
resource "astro_environment_object" "%s" {
  object_key      = "%s"
  object_type     = "METRICS_EXPORT"
  scope           = "%s"
  scope_entity_id = "%s"

  endpoint      = "https://ds-prom.example.com/api/v1/write"
  exporter_type = "PROMETHEUS"
}

data "astro_environment_object" "%s" {
  id = astro_environment_object.%s.id
}
`, name, key, scope, scopeEntityId, name, name)
}

func envObjConnectionTypedAuthResourceAndDS(name, key, workspaceId, authTypeId string) string {
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
  password     = "ds_pass"
  schema       = "ANALYTICS"
  extra        = jsonencode({ account = "abc12345", warehouse = "AIRFLOW_WH", role = "AIRFLOW_ROLE" })
}

data "astro_environment_object" "%s" {
  id = astro_environment_object.%s.id
}
`, name, key, workspaceId, authTypeId, name, name)
}

func envObjAirflowVarWithLinkResourceAndDS(name, key, workspaceId, deploymentId string) string {
	return fmt.Sprintf(`
resource "astro_environment_object" "%s" {
  object_key      = "%s"
  object_type     = "AIRFLOW_VARIABLE"
  scope           = "WORKSPACE"
  scope_entity_id = "%s"

  value     = "ds_workspace_value"
  is_secret = false

  links = [{
    scope           = "DEPLOYMENT"
    scope_entity_id = "%s"
    overrides = {
      value = "ds_override_value"
    }
  }]
}

data "astro_environment_object" "%s" {
  id = astro_environment_object.%s.id
}
`, name, key, workspaceId, deploymentId, name, name)
}
