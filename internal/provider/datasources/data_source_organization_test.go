package datasources_test

import (
	"os"
	"testing"

	astronomerprovider "github.com/astronomer/astronomer-terraform-provider/internal/provider"

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
				Config: astronomerprovider.ProviderConfig(t, false) + organization(),
				Check: resource.ComposeTestCheckFunc(
					// TODO: Add more once OrganizationDataSource fully implemented
					resource.TestCheckResourceAttr("data.astronomer_organization.t", "id", os.Getenv("HOSTED_ORGANIZATION_ID")),
					resource.TestCheckResourceAttrSet("data.astronomer_organization.t", "name"),
					resource.TestCheckResourceAttrSet("data.astronomer_organization.t", "created_by.id"),
					resource.TestCheckResourceAttrSet("data.astronomer_organization.t", "created_at"),
					resource.TestCheckResourceAttrSet("data.astronomer_organization.t", "updated_by.id"),
					resource.TestCheckResourceAttrSet("data.astronomer_organization.t", "updated_at"),
				),
			},
		},
	})
}

func organization() string {
	return `
data astronomer_organization "t" {}`
}
