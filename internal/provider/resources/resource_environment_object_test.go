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
	"github.com/stretchr/testify/assert"
)

func TestAcc_ResourceEnvironmentObjectAirflowVariable(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	varKey := fmt.Sprintf("test_var_%v", namePrefix)
	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckEnvironmentObjectDestroyed(t, varKey),
		Steps: []resource.TestStep{
			// Create
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + environmentObjectAirflowVariable("test", varKey, workspaceId, "initial_value", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("astro_environment_object.test", "object_key", varKey),
					resource.TestCheckResourceAttr("astro_environment_object.test", "object_type", "AIRFLOW_VARIABLE"),
					resource.TestCheckResourceAttr("astro_environment_object.test", "scope", "WORKSPACE"),
					resource.TestCheckResourceAttr("astro_environment_object.test", "scope_entity_id", workspaceId),
					resource.TestCheckResourceAttrSet("astro_environment_object.test", "id"),
					resource.TestCheckResourceAttrSet("astro_environment_object.test", "created_at"),
					resource.TestCheckResourceAttrSet("astro_environment_object.test", "updated_at"),
				),
			},
			// Update
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + environmentObjectAirflowVariable("test", varKey, workspaceId, "updated_value", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("astro_environment_object.test", "object_key", varKey),
					resource.TestCheckResourceAttrSet("astro_environment_object.test", "id"),
				),
			},
			// Import
			{
				ResourceName:      "astro_environment_object.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ResourceEnvironmentObjectConnection(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	connKey := fmt.Sprintf("test_conn_%v", namePrefix)
	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckEnvironmentObjectDestroyed(t, connKey),
		Steps: []resource.TestStep{
			// Create
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + environmentObjectConnection("test", connKey, workspaceId, "example.com", 5432),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("astro_environment_object.test", "object_key", connKey),
					resource.TestCheckResourceAttr("astro_environment_object.test", "object_type", "CONNECTION"),
					resource.TestCheckResourceAttr("astro_environment_object.test", "scope", "WORKSPACE"),
					resource.TestCheckResourceAttrSet("astro_environment_object.test", "id"),
					resource.TestCheckResourceAttrSet("astro_environment_object.test", "created_at"),
				),
			},
			// Update
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + environmentObjectConnection("test", connKey, workspaceId, "updated.example.com", 5433),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("astro_environment_object.test", "object_key", connKey),
					resource.TestCheckResourceAttrSet("astro_environment_object.test", "id"),
				),
			},
			// Import
			{
				ResourceName:            "astro_environment_object.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"connection_config.auth_type_id", "connection_config.password"},
			},
		},
	})
}

func TestAcc_ResourceEnvironmentObjectMetricsExport(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	meKey := fmt.Sprintf("test_me_%v", namePrefix)
	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy:             testAccCheckEnvironmentObjectDestroyed(t, meKey),
		Steps: []resource.TestStep{
			// Create
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + environmentObjectMetricsExport("test", meKey, workspaceId, "https://prometheus.example.com/api/v1/write"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("astro_environment_object.test", "object_key", meKey),
					resource.TestCheckResourceAttr("astro_environment_object.test", "object_type", "METRICS_EXPORT"),
					resource.TestCheckResourceAttr("astro_environment_object.test", "scope", "WORKSPACE"),
					resource.TestCheckResourceAttrSet("astro_environment_object.test", "id"),
					resource.TestCheckResourceAttrSet("astro_environment_object.test", "created_at"),
				),
			},
			// Update endpoint
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + environmentObjectMetricsExport("test", meKey, workspaceId, "https://prometheus.example.com/api/v2/write"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("astro_environment_object.test", "object_key", meKey),
					resource.TestCheckResourceAttrSet("astro_environment_object.test", "id"),
				),
			},
			// Import
			{
				ResourceName:            "astro_environment_object.test",
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

func environmentObjectConnection(tfName, connKey, workspaceId, host string, port int) string {
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
  }
}
`, tfName, connKey, workspaceId, host, port)
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

func testAccCheckEnvironmentObjectDestroyed(t *testing.T, objectKey string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := utils.GetTestPlatformClient(true)
		if err != nil {
			return err
		}

		orgId := os.Getenv("HOSTED_ORGANIZATION_ID")
		ctx := context.Background()

		resp, err := client.ListEnvironmentObjectsWithResponse(ctx, orgId, &platform.ListEnvironmentObjectsParams{
			ObjectKey: &objectKey,
			Limit:     lo.ToPtr(10),
		})
		if err != nil {
			return err
		}
		if resp.StatusCode() != 200 {
			statusCode, diag := clients.NormalizeAPIError(ctx, resp.HTTPResponse, resp.Body)
			if statusCode == 404 {
				return nil
			}
			if diag != nil {
				return fmt.Errorf("unexpected error checking environment object destruction: %s", diag.Detail())
			}
		}
		if resp.JSON200 != nil && resp.JSON200.TotalCount > 0 {
			for _, obj := range resp.JSON200.EnvironmentObjects {
				if obj.ObjectKey == objectKey {
					assert.Fail(t, "environment object %s still exists after destroy", objectKey)
				}
			}
		}
		return nil
	}
}
