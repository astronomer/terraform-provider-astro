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
			// Create an Airflow variable
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
			// Update the value
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

func environmentObjectAirflowVariable(tfName, varKey, workspaceId, value string, isSecret bool) string {
	return fmt.Sprintf(`
resource "astro_environment_object" "%s" {
  object_key    = "%s"
  object_type   = "AIRFLOW_VARIABLE"
  scope         = "WORKSPACE"
  scope_entity_id = "%s"

  airflow_variable = {
    value     = "%s"
    is_secret = %t
  }
}
`, tfName, varKey, workspaceId, value, isSecret)
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
