package datasources_test

import (
	"fmt"
	"os"
	"testing"

	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_DataSourceTeam(t *testing.T) {
	teamId := os.Getenv("HOSTED_TEAM_ID")
	teamName := "terraform_acceptance_tests_dnd"
	resourceVar := fmt.Sprintf("data.astro_team.%v", teamName)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, true) + team(teamId, teamName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttrSet(resourceVar, "name"),
					resource.TestCheckResourceAttrSet(resourceVar, "description"),
					resource.TestCheckResourceAttrSet(resourceVar, "is_idp_managed"),
					resource.TestCheckResourceAttrSet(resourceVar, "organization_role"),
					resource.TestCheckResourceAttrSet(resourceVar, "workspace_roles"),
					resource.TestCheckResourceAttrSet(resourceVar, "deployment_roles"),
					resource.TestCheckResourceAttrSet(resourceVar, "roles_count"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by"),
				),
			},
		},
	})
}

func team(teamId string, teamName string) string {
	return fmt.Sprintf(`
data astro_team "%v" {
	id = "%v"
}`, teamName, teamId)
}
