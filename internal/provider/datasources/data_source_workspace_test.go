package datasources_test

import (
	"fmt"
	astronomerprovider "github.com/astronomer/astronomer-terraform-provider/internal/provider"
	"github.com/astronomer/astronomer-terraform-provider/internal/utils"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_DataSourceWorkspace(t *testing.T) {
	workspaceName := utils.GenerateTestResourceName(10)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig() + workspace(workspaceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.astronomer_workspace.t", "id"),
					resource.TestCheckResourceAttr("data.astronomer_workspace.t", "name", workspaceName),
					resource.TestCheckResourceAttrSet("data.astronomer_workspace.t", "description"),
					resource.TestCheckResourceAttr("data.astronomer_workspace.t", "cicd_enforced_default", "true"),
					resource.TestCheckResourceAttrSet("data.astronomer_workspace.t", "created_by.id"),
					resource.TestCheckResourceAttrSet("data.astronomer_workspace.t", "created_at"),
					resource.TestCheckResourceAttrSet("data.astronomer_workspace.t", "updated_by.id"),
					resource.TestCheckResourceAttrSet("data.astronomer_workspace.t", "updated_at"),
				),
			},
		},
	})
}

func workspace(name string) string {
	return fmt.Sprintf(`
resource "astronomer_workspace" "test_workspace" {
	name = "%v"
	description = "%v"
	cicd_enforced_default = true
}

data astronomer_workspace "t" {
	depends_on = [astronomer_workspace.test_workspace]
	id = astronomer_workspace.test_workspace.id
}`, name, utils.TestResourceDescription)
}
