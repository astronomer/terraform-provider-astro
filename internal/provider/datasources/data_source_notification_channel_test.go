package datasources_test

import (
	"fmt"
	"os"
	"testing"

	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_DataSource_NotificationChannel(t *testing.T) {
	notificationChannelId := os.Getenv("HOSTED_NOTIFICATION_CHANNEL_ID")
	tfVarName := "test_data_notification_channel"
	resourceVar := fmt.Sprintf("data.astro_notification_channel.%v", tfVarName)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannel(notificationChannelId, tfVarName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttrSet(resourceVar, "name"),
					resource.TestCheckResourceAttrWith(resourceVar, "definition.%", CheckAttributeLengthIsNotEmpty),
					resource.TestCheckResourceAttrSet(resourceVar, "type"),
					resource.TestCheckResourceAttrSet(resourceVar, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "entity_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "entity_type"),
					resource.TestCheckResourceAttrSet(resourceVar, "is_shared"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by.id"),
				),
			},
		},
	})
}

func notificationChannel(notificationChannelId string, tfVarName string) string {
	return fmt.Sprintf(`
data astro_notification_channel "%v" {
	id = "%v"
}`, tfVarName, notificationChannelId)
}
