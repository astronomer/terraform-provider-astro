package datasources_test

import (
	"fmt"
	"os"
	"testing"

	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_DataSource_ApiToken(t *testing.T) {
	apiTokenId := os.Getenv("HOSTED_API_TOKEN_ID")
	tfVarName := "test_data_api_token"
	resourceVar := fmt.Sprintf("data.astro_api_token.%v", tfVarName)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, true, false) + apiToken(apiTokenId, tfVarName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttrSet(resourceVar, "name"),
					resource.TestCheckResourceAttrSet(resourceVar, "short_token"),
					resource.TestCheckResourceAttrSet(resourceVar, "type"),
					resource.TestCheckResourceAttrSet(resourceVar, "start_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "last_used_at"),
					resource.TestCheckResourceAttrWith(resourceVar, "roles.#", CheckAttributeLengthIsNotEmpty),
					resource.TestCheckResourceAttrSet(resourceVar, "roles.0.entity_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "roles.0.entity_type"),
					resource.TestCheckResourceAttrSet(resourceVar, "roles.0.role"),
				),
			},
		},
	})
}

func apiToken(apiTokenId string, tfVarName string) string {
	return fmt.Sprintf(`
data astro_api_token "%v" {
	id = "%v"
}`, tfVarName, apiTokenId)
}
