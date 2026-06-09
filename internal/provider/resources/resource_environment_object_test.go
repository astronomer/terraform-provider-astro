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
			// Import
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"basic_token", "password"},
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
			// Import
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"basic_token", "password"},
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
			return nil
		}
		if resp.StatusCode() != 200 {
			statusCode, diag := clients.NormalizeAPIError(ctx, resp.HTTPResponse, resp.Body)
			if statusCode == 404 {
				return nil
			}
			if diag != nil {
				return fmt.Errorf("unexpected error checking destruction: %s", diag.Detail())
			}
		}
		if resp.JSON200 != nil {
			for _, obj := range resp.JSON200.EnvironmentObjects {
				if obj.ObjectKey == objectKey {
					return fmt.Errorf("environment object %s still exists after destroy", objectKey)
				}
			}
		}
		return nil
	}
}
