package datasources_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/samber/lo"
)

func TestAcc_DataSource_NotificationChannels(t *testing.T) {
	tfVarName := "test_data_notification_channels"
	tfWorkspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
	tfDeploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")
	tfNotificationChannelId := os.Getenv("HOSTED_NOTIFICATION_CHANNEL_ID")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannels(tfVarName),
				Check: resource.ComposeTestCheckFunc(
					checkNotificationChannels(tfVarName),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannelsFilterWorkspaceIds(tfVarName, []string{tfWorkspaceId}),
				Check: resource.ComposeTestCheckFunc(
					checkNotificationChannels(tfVarName),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannelsFilterDeploymentIds(tfVarName, []string{tfDeploymentId}),
				Check: resource.ComposeTestCheckFunc(
					checkNotificationChannels(tfVarName),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannelsFilterNotificationChannelIds(tfVarName, []string{tfNotificationChannelId}),
				Check: resource.ComposeTestCheckFunc(
					checkNotificationChannels(tfVarName),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannelsFilterChannelTypes(tfVarName, []string{string(platform.AlertNotificationChannelTypeEMAIL)}),
				Check: resource.ComposeTestCheckFunc(
					checkNotificationChannels(tfVarName),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannelsFilterEntityType(tfVarName, string(platform.AlertNotificationChannelEntityTypeDEPLOYMENT)),
				Check: resource.ComposeTestCheckFunc(
					checkNotificationChannels(tfVarName),
				),
			},
		},
	})
}

func notificationChannels(tfVarName string) string {
	return fmt.Sprintf(`
 data astro_notification_channels "%v" {}`, tfVarName)
}

func notificationChannelsFilterWorkspaceIds(tfVarName string, workspaceIds []string) string {
	quoted := lo.Map(workspaceIds, func(id string, _ int) string {
		return fmt.Sprintf("%q", id)
	})
	return fmt.Sprintf(`
 data astro_notification_channels "%v" {
 	workspace_ids = [%s]
 }`, tfVarName, strings.Join(quoted, ","))
}

func notificationChannelsFilterDeploymentIds(tfVarName string, deploymentIds []string) string {
	quoted := lo.Map(deploymentIds, func(id string, _ int) string {
		return fmt.Sprintf("%q", id)
	})
	return fmt.Sprintf(`
 data astro_notification_channels "%v" {
 	deployment_ids = [%s]
 }`, tfVarName, strings.Join(quoted, ","))
}

func notificationChannelsFilterNotificationChannelIds(tfVarName string, ids []string) string {
	quoted := lo.Map(ids, func(id string, _ int) string {
		return fmt.Sprintf("%q", id)
	})
	return fmt.Sprintf(`
 data astro_notification_channels "%v" {
 	notification_channel_ids = [%s]
 }`, tfVarName, strings.Join(quoted, ","))
}

func notificationChannelsFilterChannelTypes(tfVarName string, channelTypes []string) string {
	quoted := lo.Map(channelTypes, func(t string, _ int) string {
		return fmt.Sprintf("%q", t)
	})
	return fmt.Sprintf(`
 data astro_notification_channels "%v" {
 	channel_types = [%s]
 }`, tfVarName, strings.Join(quoted, ","))
}

func notificationChannelsFilterEntityType(tfVarName, entityType string) string {
	return fmt.Sprintf(`
 data astro_notification_channels "%v" {
 	entity_type = "%v"
 }`, tfVarName, entityType)
}

func checkNotificationChannels(tfVarName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		instanceState, numChannels, err := utils.GetDataSourcesLength(s, tfVarName, "notification_channels")
		if err != nil {
			return err
		}
		if numChannels == 0 {
			return fmt.Errorf("expected notification_channels to be greater or equal to 1, got %s", instanceState.Attributes["notification_channels.#"])
		}

		idx := 0

		id := fmt.Sprintf("notification_channels.%d.id", idx)
		if instanceState.Attributes[id] == "" {
			return fmt.Errorf("expected 'id' to be set")
		}
		name := fmt.Sprintf("notification_channels.%d.name", idx)
		if instanceState.Attributes[name] == "" {
			return fmt.Errorf("expected 'name' to be set")
		}
		// definition map count
		defCountKey := fmt.Sprintf("notification_channels.%d.definition.%%", idx)
		if instanceState.Attributes[defCountKey] == "" {
			return fmt.Errorf("expected 'definition' to have at least one entry")
		}
		channelType := fmt.Sprintf("notification_channels.%d.type", idx)
		if instanceState.Attributes[channelType] == "" {
			return fmt.Errorf("expected 'type' to be set")
		}
		orgId := fmt.Sprintf("notification_channels.%d.organization_id", idx)
		if instanceState.Attributes[orgId] == "" {
			return fmt.Errorf("expected 'organization_id' to be set")
		}
		workspaceId := fmt.Sprintf("notification_channels.%d.workspace_id", idx)
		if instanceState.Attributes[workspaceId] == "" {
			return fmt.Errorf("expected 'workspace_id' to be set")
		}
		deploymentId := fmt.Sprintf("notification_channels.%d.deployment_id", idx)
		if instanceState.Attributes[deploymentId] == "" {
			return fmt.Errorf("expected 'deployment_id' to be set")
		}
		entityId := fmt.Sprintf("notification_channels.%d.entity_id", idx)
		if instanceState.Attributes[entityId] == "" {
			return fmt.Errorf("expected 'entity_id' to be set")
		}
		entityType := fmt.Sprintf("notification_channels.%d.entity_type", idx)
		if instanceState.Attributes[entityType] == "" {
			return fmt.Errorf("expected 'entity_type' to be set")
		}
		isShared := fmt.Sprintf("notification_channels.%d.is_shared", idx)
		if instanceState.Attributes[isShared] == "" {
			return fmt.Errorf("expected 'is_shared' to be set")
		}
		createdAt := fmt.Sprintf("notification_channels.%d.created_at", idx)
		if instanceState.Attributes[createdAt] == "" {
			return fmt.Errorf("expected 'created_at' to be set")
		}
		updatedAt := fmt.Sprintf("notification_channels.%d.updated_at", idx)
		if instanceState.Attributes[updatedAt] == "" {
			return fmt.Errorf("expected 'updated_at' to be set")
		}
		createdBy := fmt.Sprintf("notification_channels.%d.created_by.id", idx)
		if instanceState.Attributes[createdBy] == "" {
			return fmt.Errorf("expected 'created_by.id' to be set")
		}
		updatedBy := fmt.Sprintf("notification_channels.%d.updated_by.id", idx)
		if instanceState.Attributes[updatedBy] == "" {
			return fmt.Errorf("expected 'updated_by.id' to be set")
		}
		return nil
	}
}
