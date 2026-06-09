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

func TestAcc_ResourceEnvironmentObjectAirflowVariable(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	varKey := fmt.Sprintf("test_var_%v", namePrefix)
	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
	resourceVar := "astro_environment_object.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             resource.ComposeTestCheckFunc(testAccCheckEnvironmentObjectDestroyed(t, varKey)),
		Steps: []resource.TestStep{
			// Create
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + environmentObjectAirflowVariable("test", varKey, workspaceId, "initial_value", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, varKey),
					resource.TestCheckResourceAttr(resourceVar, "object_key", varKey),
					resource.TestCheckResourceAttr(resourceVar, "object_type", "AIRFLOW_VARIABLE"),
					resource.TestCheckResourceAttr(resourceVar, "scope", "WORKSPACE"),
					resource.TestCheckResourceAttr(resourceVar, "scope_entity_id", workspaceId),
					resource.TestCheckResourceAttr(resourceVar, "airflow_variable.value", "initial_value"),
					resource.TestCheckResourceAttr(resourceVar, "airflow_variable.is_secret", "false"),
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
				),
			},
			// Update value (in-place)
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + environmentObjectAirflowVariable("test", varKey, workspaceId, "updated_value", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, varKey),
					resource.TestCheckResourceAttr(resourceVar, "airflow_variable.value", "updated_value"),
					resource.TestCheckResourceAttr(resourceVar, "airflow_variable.is_secret", "false"),
				),
			},
			// Toggle is_secret=true — immutable on the API, so RequiresReplace forces re-create.
			// The model preserves the plan value so state stays consistent with the user's input
			// even though the API returns an empty value for secrets.
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + environmentObjectAirflowVariable("test", varKey, workspaceId, "secret_value", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, varKey),
					resource.TestCheckResourceAttr(resourceVar, "airflow_variable.is_secret", "true"),
					resource.TestCheckResourceAttr(resourceVar, "airflow_variable.value", "secret_value"),
				),
			},
			// Import — value cannot be verified after import because the secret value isn't
			// returned by the API and there's no plan to preserve from during a raw import.
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"airflow_variable.value"},
			},
		},
	})
}

func TestAcc_ResourceEnvironmentObjectConnection(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	connKey := fmt.Sprintf("test_conn_%v", namePrefix)
	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
	resourceVar := "astro_environment_object.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             resource.ComposeTestCheckFunc(testAccCheckEnvironmentObjectDestroyed(t, connKey)),
		Steps: []resource.TestStep{
			// Create with extra JSON to exercise the round-trip preservation logic.
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + environmentObjectConnection("test", connKey, workspaceId, "example.com", 5432, `{"sslmode":"require","timeout":30}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, connKey),
					resource.TestCheckResourceAttr(resourceVar, "object_key", connKey),
					resource.TestCheckResourceAttr(resourceVar, "object_type", "CONNECTION"),
					resource.TestCheckResourceAttr(resourceVar, "scope", "WORKSPACE"),
					resource.TestCheckResourceAttr(resourceVar, "connection_config.type", "postgres"),
					resource.TestCheckResourceAttr(resourceVar, "connection_config.host", "example.com"),
					resource.TestCheckResourceAttr(resourceVar, "connection_config.port", "5432"),
					resource.TestCheckResourceAttr(resourceVar, "connection_config.login", "testuser"),
					resource.TestCheckResourceAttr(resourceVar, "connection_config.password", "testpass"),
					resource.TestCheckResourceAttr(resourceVar, "connection_config.schema", "testdb"),
					resource.TestCheckResourceAttr(resourceVar, "connection_config.extra", `{"sslmode":"require","timeout":30}`),
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
				),
			},
			// Update host/port — preserved password and extra must follow through.
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + environmentObjectConnection("test", connKey, workspaceId, "updated.example.com", 5433, `{"sslmode":"require","timeout":30}`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, connKey),
					resource.TestCheckResourceAttr(resourceVar, "connection_config.host", "updated.example.com"),
					resource.TestCheckResourceAttr(resourceVar, "connection_config.port", "5433"),
					resource.TestCheckResourceAttr(resourceVar, "connection_config.password", "testpass"),
					resource.TestCheckResourceAttr(resourceVar, "connection_config.extra", `{"sslmode":"require","timeout":30}`),
				),
			},
			// Import — password and extra are unrecoverable on import: the API does not echo
			// them back on GET, and there's no plan to preserve from during a raw import.
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"connection_config.password", "connection_config.extra"},
			},
		},
	})
}

func TestAcc_ResourceEnvironmentObjectMetricsExport(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	meKey := fmt.Sprintf("test_me_%v", namePrefix)
	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
	resourceVar := "astro_environment_object.test"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             resource.ComposeTestCheckFunc(testAccCheckEnvironmentObjectDestroyed(t, meKey)),
		Steps: []resource.TestStep{
			// Create
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + environmentObjectMetricsExport("test", meKey, workspaceId, "https://prometheus.example.com/api/v1/write"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, meKey),
					resource.TestCheckResourceAttr(resourceVar, "object_key", meKey),
					resource.TestCheckResourceAttr(resourceVar, "object_type", "METRICS_EXPORT"),
					resource.TestCheckResourceAttr(resourceVar, "scope", "WORKSPACE"),
					resource.TestCheckResourceAttr(resourceVar, "metrics_export.endpoint", "https://prometheus.example.com/api/v1/write"),
					resource.TestCheckResourceAttr(resourceVar, "metrics_export.exporter_type", "PROMETHEUS"),
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
				),
			},
			// Update endpoint
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + environmentObjectMetricsExport("test", meKey, workspaceId, "https://prometheus.example.com/api/v2/write"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentObjectExists(t, meKey),
					resource.TestCheckResourceAttr(resourceVar, "metrics_export.endpoint", "https://prometheus.example.com/api/v2/write"),
				),
			},
			// Import — basic_token & password are unrecoverable on import.
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metrics_export.basic_token", "metrics_export.password"},
			},
		},
	})
}

func environmentObjectAirflowVariable(tfName, varKey, workspaceId, value string, isSecret bool) string {
	return fmt.Sprintf(`
resource "astro_environment_object" "%s" {
  object_key      = "%s"
  object_type     = "AIRFLOW_VARIABLE"
  scope           = "WORKSPACE"
  scope_entity_id = "%s"

  airflow_variable = {
    value     = "%s"
    is_secret = %t
  }
}
`, tfName, varKey, workspaceId, value, isSecret)
}

func environmentObjectConnection(tfName, connKey, workspaceId, host string, port int, extraJSON string) string {
	extraLine := ""
	if extraJSON != "" {
		// %q quotes the JSON string verbatim so it round-trips byte-identical through
		// the preservation logic in models/environment_object.go.
		extraLine = fmt.Sprintf("extra = %q", extraJSON)
	}
	return fmt.Sprintf(`
resource "astro_environment_object" "%s" {
  object_key      = "%s"
  object_type     = "CONNECTION"
  scope           = "WORKSPACE"
  scope_entity_id = "%s"

  connection_config = {
    type     = "postgres"
    host     = "%s"
    port     = %d
    login    = "testuser"
    password = "testpass"
    schema   = "testdb"
    %s
  }
}
`, tfName, connKey, workspaceId, host, port, extraLine)
}

func environmentObjectMetricsExport(tfName, meKey, workspaceId, endpoint string) string {
	return fmt.Sprintf(`
resource "astro_environment_object" "%s" {
  object_key      = "%s"
  object_type     = "METRICS_EXPORT"
  scope           = "WORKSPACE"
  scope_entity_id = "%s"

  metrics_export = {
    endpoint      = "%s"
    exporter_type = "PROMETHEUS"
  }
}
`, tfName, meKey, workspaceId, endpoint)
}

// testAccCheckEnvironmentObjectExists verifies that an environment object with
// the given key is reachable via the API (not just present in Terraform state).
// Matches the pattern used by testAccCheckNotificationChannelExists et al.
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
		if resp == nil {
			return fmt.Errorf("nil response from list environment objects")
		}
		if resp.JSON200 == nil {
			status, diag := clients.NormalizeAPIError(ctx, resp.HTTPResponse, resp.Body)
			return fmt.Errorf("response JSON200 is nil status: %v, err: %v", status, diag.Detail())
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
			return fmt.Errorf("nil response from list environment objects")
		}
		if resp.StatusCode() != 200 {
			status, diag := clients.NormalizeAPIError(ctx, resp.HTTPResponse, resp.Body)
			if status == 404 {
				return nil
			}
			if diag != nil {
				return fmt.Errorf("unexpected error checking environment object destruction: %s", diag.Detail())
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
