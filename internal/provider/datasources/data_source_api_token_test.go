package datasources_test

import (
	"fmt"
	"os"
	"testing"

	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_DataSource_ApiToken(t *testing.T) {
	apiTokenId := os.Getenv("HOSTED_ORGANIZATION_API_TOKEN")
	tfVarName := "test_data_api_token"
	resourceVar := fmt.Sprintf("data.astro_api_token.%v", apiTokenId)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, true) + apiToken(apiTokenId, tfVarName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttrSet(resourceVar, "name"),
					resource.TestCheckResourceAttrSet(resourceVar, "description"),
					resource.TestCheckResourceAttrSet(resourceVar, "short_token"),
					resource.TestCheckResourceAttrSet(resourceVar, "type"),
					resource.TestCheckResourceAttrSet(resourceVar, "start_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "end_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by"),
					resource.TestCheckResourceAttrSet(resourceVar, "expiry_period_in_days"),
					resource.TestCheckResourceAttrSet(resourceVar, "last_used_at"),
					resource.TestCheckResourceAttrWith(resourceVar, "roles.#", CheckAttributeLengthIsNotEmpty),
					resource.TestCheckResourceAttrSet(resourceVar, "token"),
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
