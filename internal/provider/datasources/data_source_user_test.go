package datasources_test

import (
	"fmt"
	"os"
	"testing"

	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_DataSourceUser(t *testing.T) {
	userId := os.Getenv("HOSTED_USER_ID")
	userName := "user"
	resourceVar := fmt.Sprintf("data.astro_user.%v", userName)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + user(userId, userName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttrSet(resourceVar, "username"),
					resource.TestCheckResourceAttrSet(resourceVar, "full_name"),
					resource.TestCheckResourceAttrSet(resourceVar, "status"),
					resource.TestCheckResourceAttrSet(resourceVar, "avatar_url"),
					resource.TestCheckResourceAttrSet(resourceVar, "organization_role"),
					resource.TestCheckResourceAttrWith(resourceVar, "workspace_roles.#", CheckAttributeLengthIsNotEmpty),
					resource.TestCheckResourceAttrWith(resourceVar, "deployment_roles.#", CheckAttributeLengthIsNotEmpty),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
				),
			},
		},
	})
}

func user(userId string, tfVarName string) string {
	return fmt.Sprintf(`
data astro_user "%v" {
	id = "%v"
}`, tfVarName, userId)
}
