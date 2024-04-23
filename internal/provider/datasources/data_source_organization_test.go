package datasources_test

import (
	"os"
	"testing"

	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_DataSourceOrganization(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, true) + organization(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.astro_organization.t", "id", os.Getenv("HOSTED_ORGANIZATION_ID")),
					resource.TestCheckResourceAttrSet("data.astro_organization.t", "name"),
					resource.TestCheckResourceAttrSet("data.astro_organization.t", "support_plan"),
					resource.TestCheckResourceAttrSet("data.astro_organization.t", "product"),
					resource.TestCheckResourceAttrSet("data.astro_organization.t", "created_by.id"),
					resource.TestCheckResourceAttrSet("data.astro_organization.t", "created_at"),
					resource.TestCheckResourceAttrSet("data.astro_organization.t", "updated_by.id"),
					resource.TestCheckResourceAttrSet("data.astro_organization.t", "updated_at"),
					resource.TestCheckResourceAttrSet("data.astro_organization.t", "status"),
				),
			},
		},
	})
}

func organization() string {
	return `
data astro_organization "t" {}`
}
