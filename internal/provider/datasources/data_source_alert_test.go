package datasources_test

import (
	"fmt"
	"os"
	"testing"

	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_DataSource_Alert(t *testing.T) {
	alertId := os.Getenv("HOSTED_ALERT_ID")
	tfVarName := "test_data_alert"
	resourceVar := fmt.Sprintf("data.astro_alert.%v", tfVarName)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertId, tfVarName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttrSet(resourceVar, "name"),
					resource.TestCheckResourceAttrSet(resourceVar, "entity_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "entity_type"),
					resource.TestCheckResourceAttrSet(resourceVar, "entity_name"),
					resource.TestCheckResourceAttrSet(resourceVar, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "workspace_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "deployment_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "severity"),
					resource.TestCheckResourceAttrSet(resourceVar, "type"),
					resource.TestCheckResourceAttrWith(resourceVar, "rules.properties.%", CheckAttributeLengthIsNotEmpty),
					resource.TestCheckResourceAttrWith(resourceVar, "rules.pattern_matches.#", CheckAttributeLengthIsNotEmpty),
					resource.TestCheckResourceAttrSet(resourceVar, "rules.pattern_matches.0.entity_type"),
					resource.TestCheckResourceAttrSet(resourceVar, "rules.pattern_matches.0.operator_type"),
					resource.TestCheckResourceAttrWith(resourceVar, "rules.pattern_matches.0.values.#", CheckAttributeLengthIsNotEmpty),
					resource.TestCheckResourceAttrSet(resourceVar, "rules.pattern_matches.1.entity_type"),
					resource.TestCheckResourceAttrSet(resourceVar, "rules.pattern_matches.1.operator_type"),
					resource.TestCheckResourceAttrWith(resourceVar, "rules.pattern_matches.1.values.#", CheckAttributeLengthIsNotEmpty),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by.id"),
				),
			},
		},
	})
}

func alert(alertId string, tfVarName string) string {
	return fmt.Sprintf(`
data astro_alert "%v" {
	id = "%v"
}`, tfVarName, alertId)
}
