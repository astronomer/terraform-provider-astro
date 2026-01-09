package datasources_test

import (
	"fmt"
	"os"
	"testing"

	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_DataSourceCustomRole(t *testing.T) {
	customRoleId := os.Getenv("HOSTED_CUSTOM_ROLE_ID")
	roleName := "custom_role"
	resourceVar := fmt.Sprintf("data.astro_custom_role.%v", roleName)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + customRole(customRoleId, roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttrSet(resourceVar, "name"),
					resource.TestCheckResourceAttrSet(resourceVar, "scope_type"),
					resource.TestCheckResourceAttrSet(resourceVar, "permissions.#"),
					resource.TestCheckResourceAttrSet(resourceVar, "restricted_workspace_ids.#"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by.id"),
				),
			},
		},
	})
}

func customRole(customRoleId string, tfVarName string) string {
	return fmt.Sprintf(`
data astro_custom_role "%v" {
	id = "%v"
}`, tfVarName, customRoleId)
}
