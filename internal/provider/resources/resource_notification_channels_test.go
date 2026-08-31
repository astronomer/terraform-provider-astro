package resources_test

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAcc_ResourceNotificationChannels exercises the bulk astro_notification_channels resource
// end-to-end: create a map of channels, then update it (rename one, add another), and finally let
// CheckDestroy confirm every channel is gone.
func TestAcc_ResourceNotificationChannels(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	emailName := fmt.Sprintf("%v_email", namePrefix)
	slackName := fmt.Sprintf("%v_slack", namePrefix)
	pagerdutyName := fmt.Sprintf("%v_pagerduty", namePrefix)
	emailRenamed := fmt.Sprintf("%v_email_v2", namePrefix)

	resourceVar := "astro_notification_channels.test"
	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckNotificationChannelDestroyed(t, emailName),
			testAccCheckNotificationChannelDestroyed(t, slackName),
			testAccCheckNotificationChannelDestroyed(t, pagerdutyName),
			testAccCheckNotificationChannelDestroyed(t, emailRenamed),
		),
		Steps: []resource.TestStep{
			// Create: two channels (EMAIL + SLACK) in one bulk resource.
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannels(map[string]notificationChannelInput{
					"email": {
						Name:       emailName,
						Type:       string(platform.AlertNotificationChannelTypeEMAIL),
						EntityId:   deploymentId,
						EntityType: "DEPLOYMENT",
						Definition: map[string]interface{}{
							"recipients": []string{"test@example.com", "admin@example.com"},
						},
					},
					"slack": {
						Name:       slackName,
						Type:       string(platform.AlertNotificationChannelTypeSLACK),
						EntityId:   deploymentId,
						EntityType: "DEPLOYMENT",
						Definition: map[string]interface{}{
							"webhook_url": "https://hooks.slack.com/services/T000/B000/XXX",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "notification_channels.%", "2"),
					resource.TestCheckResourceAttrSet(resourceVar, "notification_channels.email.id"),
					resource.TestCheckResourceAttr(resourceVar, "notification_channels.email.name", emailName),
					resource.TestCheckResourceAttr(resourceVar, "notification_channels.email.type", string(platform.AlertNotificationChannelTypeEMAIL)),
					resource.TestCheckResourceAttr(resourceVar, "notification_channels.email.definition.recipients.#", "2"),
					resource.TestCheckResourceAttrSet(resourceVar, "notification_channels.slack.id"),
					resource.TestCheckResourceAttr(resourceVar, "notification_channels.slack.type", string(platform.AlertNotificationChannelTypeSLACK)),
					resource.TestCheckResourceAttr(resourceVar, "notification_channels.slack.definition.webhook_url", "https://hooks.slack.com/services/T000/B000/XXX"),
					testAccCheckNotificationChannelExists(t, emailName),
					testAccCheckNotificationChannelExists(t, slackName),
				),
			},
			// Update: rename the email channel (update path via single-channel endpoint), drop
			// nothing, and add a PagerDuty channel (bulk create for the new key).
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannels(map[string]notificationChannelInput{
					"email": {
						Name:       emailRenamed,
						Type:       string(platform.AlertNotificationChannelTypeEMAIL),
						EntityId:   deploymentId,
						EntityType: "DEPLOYMENT",
						Definition: map[string]interface{}{
							"recipients": []string{"test@example.com"},
						},
					},
					"slack": {
						Name:       slackName,
						Type:       string(platform.AlertNotificationChannelTypeSLACK),
						EntityId:   deploymentId,
						EntityType: "DEPLOYMENT",
						Definition: map[string]interface{}{
							"webhook_url": "https://hooks.slack.com/services/T000/B000/XXX",
						},
					},
					"pagerduty": {
						Name:       pagerdutyName,
						Type:       string(platform.AlertNotificationChannelTypePAGERDUTY),
						EntityId:   deploymentId,
						EntityType: "DEPLOYMENT",
						Definition: map[string]interface{}{
							"integration_key": "0123456789abcdef0123456789abcdef",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "notification_channels.%", "3"),
					resource.TestCheckResourceAttr(resourceVar, "notification_channels.email.name", emailRenamed),
					resource.TestCheckResourceAttr(resourceVar, "notification_channels.email.definition.recipients.#", "1"),
					resource.TestCheckResourceAttrSet(resourceVar, "notification_channels.pagerduty.id"),
					resource.TestCheckResourceAttr(resourceVar, "notification_channels.pagerduty.type", string(platform.AlertNotificationChannelTypePAGERDUTY)),
					testAccCheckNotificationChannelExists(t, emailRenamed),
					testAccCheckNotificationChannelExists(t, pagerdutyName),
					// The old email name should no longer resolve to a channel.
					testAccCheckNotificationChannelDestroyed(t, emailName),
				),
			},
		},
	})
}

// notificationChannels renders an astro_notification_channels resource named "test" from a map of
// channel inputs keyed by their resource-map key. Keys are emitted in sorted order for a stable
// config string.
func notificationChannels(channels map[string]notificationChannelInput) string {
	keys := make([]string, 0, len(channels))
	for k := range channels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteString("resource \"astro_notification_channels\" \"test\" {\n")
	b.WriteString("\tnotification_channels = {\n")
	for _, k := range keys {
		input := channels[k]

		definitionStr := "definition = {\n"
		for dk, dv := range input.Definition {
			switch val := dv.(type) {
			case string:
				definitionStr += fmt.Sprintf("\t\t\t\t%s = \"%s\"\n", dk, val)
			case []string:
				definitionStr += fmt.Sprintf("\t\t\t\t%s = [", dk)
				for i, s := range val {
					if i > 0 {
						definitionStr += ", "
					}
					definitionStr += fmt.Sprintf("\"%s\"", s)
				}
				definitionStr += "]\n"
			}
		}
		definitionStr += "\t\t\t}"

		isSharedStr := ""
		if input.IsShared != nil {
			isSharedStr = fmt.Sprintf("\n\t\t\tis_shared = %t", *input.IsShared)
		}

		b.WriteString(fmt.Sprintf(`		"%s" = {
			name        = "%s"
			type        = "%s"
			entity_id   = "%s"
			entity_type = "%s"%s
			%s
		}
`, k, input.Name, input.Type, input.EntityId, input.EntityType, isSharedStr, definitionStr))
	}
	b.WriteString("\t}\n}\n")
	return b.String()
}
