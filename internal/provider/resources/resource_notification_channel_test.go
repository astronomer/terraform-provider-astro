package resources_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestAcc_ResourceNotificationChannelEmail(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	channelName := fmt.Sprintf("%v_email", namePrefix)
	resourceVar := fmt.Sprintf("astro_notification_channel.%v", channelName)

	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckNotificationChannelDestroyed(t, channelName),
		),
		Steps: []resource.TestStep{
			// Validate: invalid entity type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannel(notificationChannelInput{
					Name:       channelName,
					Type:       "EMAIL",
					EntityId:   deploymentId,
					EntityType: "INVALID_ENTITY_TYPE",
					Definition: map[string]interface{}{
						"recipients": []string{"test@example.com"},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Validate: invalid notification channel type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannel(notificationChannelInput{
					Name:       channelName,
					Type:       "INVALID_TYPE",
					EntityId:   deploymentId,
					EntityType: "DEPLOYMENT",
					Definition: map[string]interface{}{
						"recipients": []string{"test@example.com"},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Validate: empty recipients
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannel(notificationChannelInput{
					Name:       channelName,
					Type:       "EMAIL",
					EntityId:   deploymentId,
					EntityType: "DEPLOYMENT",
					Definition: map[string]interface{}{
						"recipients": []string{},
					},
				}),
				ExpectError: regexp.MustCompile("must have at least 1 elements"),
			},
			// Create: EMAIL notification channel for deployment
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannel(notificationChannelInput{
					Name:       channelName,
					Type:       "EMAIL",
					EntityId:   deploymentId,
					EntityType: "DEPLOYMENT",
					Definition: map[string]interface{}{
						"recipients": []string{"test@example.com", "admin@example.com"},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttr(resourceVar, "name", channelName),
					resource.TestCheckResourceAttr(resourceVar, "type", "EMAIL"),
					resource.TestCheckResourceAttr(resourceVar, "entity_id", deploymentId),
					resource.TestCheckResourceAttr(resourceVar, "entity_type", "DEPLOYMENT"),
					resource.TestCheckResourceAttr(resourceVar, "definition.recipients.#", "2"),
					resource.TestCheckResourceAttr(resourceVar, "definition.recipients.0", "test@example.com"),
					resource.TestCheckResourceAttr(resourceVar, "definition.recipients.1", "admin@example.com"),
					resource.TestCheckResourceAttr(resourceVar, "is_shared", "false"),
					resource.TestCheckResourceAttrSet(resourceVar, "organization_id"),
					resource.TestCheckResourceAttr(resourceVar, "organization_id", os.Getenv("HOSTED_ORGANIZATION_ID")),
					resource.TestCheckResourceAttr(resourceVar, "deployment_id", deploymentId),
					resource.TestCheckResourceAttrSet(resourceVar, "workspace_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by.id"),
					testAccCheckNotificationChannelExists(t, channelName),
				),
			},
			// Update: recipients list
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannel(notificationChannelInput{
					Name:       channelName,
					Type:       "EMAIL",
					EntityId:   deploymentId,
					EntityType: "DEPLOYMENT",
					Definition: map[string]interface{}{
						"recipients": []string{"newuser@example.com"},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "definition.recipients.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "definition.recipients.0", "newuser@example.com"),
					testAccCheckNotificationChannelExists(t, channelName),
				),
			},
			// Import: test import functionality
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"entity_name"},
			},
		},
	})
}

func TestAcc_ResourceNotificationChannelEmailWorkspace(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	channelName := fmt.Sprintf("%v_email_workspace", namePrefix)
	resourceVar := fmt.Sprintf("astro_notification_channel.%v", channelName)

	workspaceId := os.Getenv("HOSTED_WORKSPACE_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckNotificationChannelDestroyed(t, channelName),
		),
		Steps: []resource.TestStep{
			// Create: EMAIL notification channel for workspace with is_shared
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannel(notificationChannelInput{
					Name:       channelName,
					Type:       "EMAIL",
					EntityId:   workspaceId,
					EntityType: "WORKSPACE",
					IsShared:   lo.ToPtr(true),
					Definition: map[string]interface{}{
						"recipients": []string{"workspace@example.com"},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttr(resourceVar, "name", channelName),
					resource.TestCheckResourceAttr(resourceVar, "type", "EMAIL"),
					resource.TestCheckResourceAttr(resourceVar, "entity_id", workspaceId),
					resource.TestCheckResourceAttr(resourceVar, "entity_type", "WORKSPACE"),
					resource.TestCheckResourceAttr(resourceVar, "definition.recipients.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "is_shared", "true"),
					resource.TestCheckResourceAttr(resourceVar, "workspace_id", workspaceId),
					resource.TestCheckResourceAttrSet(resourceVar, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by.id"),
					testAccCheckNotificationChannelExists(t, channelName),
				),
			},
			// Update: is_shared to false
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannel(notificationChannelInput{
					Name:       channelName,
					Type:       "EMAIL",
					EntityId:   workspaceId,
					EntityType: "WORKSPACE",
					IsShared:   lo.ToPtr(false),
					Definition: map[string]interface{}{
						"recipients": []string{"workspace@example.com"},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "is_shared", "false"),
					testAccCheckNotificationChannelExists(t, channelName),
				),
			},
		},
	})
}

func TestAcc_ResourceNotificationChannelSlack(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	channelName := fmt.Sprintf("%v_slack", namePrefix)
	resourceVar := fmt.Sprintf("astro_notification_channel.%v", channelName)

	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckNotificationChannelDestroyed(t, channelName),
		),
		Steps: []resource.TestStep{
			// Validate: empty webhook URL
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannel(notificationChannelInput{
					Name:       channelName,
					Type:       "SLACK",
					EntityId:   deploymentId,
					EntityType: "DEPLOYMENT",
					Definition: map[string]interface{}{
						"webhook_url": "",
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Length"),
			},
			// Create: SLACK notification channel
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannel(notificationChannelInput{
					Name:       channelName,
					Type:       "SLACK",
					EntityId:   deploymentId,
					EntityType: "DEPLOYMENT",
					Definition: map[string]interface{}{
						"webhook_url": "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXXXXXX",
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttr(resourceVar, "name", channelName),
					resource.TestCheckResourceAttr(resourceVar, "type", "SLACK"),
					resource.TestCheckResourceAttr(resourceVar, "entity_id", deploymentId),
					resource.TestCheckResourceAttr(resourceVar, "entity_type", "DEPLOYMENT"),
					resource.TestCheckResourceAttr(resourceVar, "definition.webhook_url", "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXXXXXX"),
					resource.TestCheckResourceAttr(resourceVar, "is_shared", "false"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by.id"),
					testAccCheckNotificationChannelExists(t, channelName),
				),
			},
			// Update: webhook URL
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannel(notificationChannelInput{
					Name:       channelName,
					Type:       "SLACK",
					EntityId:   deploymentId,
					EntityType: "DEPLOYMENT",
					Definition: map[string]interface{}{
						"webhook_url": "https://hooks.slack.com/services/T11111111/B11111111/YYYYYYYYYYYYYYYYYYYYYYYYYYYY",
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "definition.webhook_url", "https://hooks.slack.com/services/T11111111/B11111111/YYYYYYYYYYYYYYYYYYYYYYYYYYYY"),
					testAccCheckNotificationChannelExists(t, channelName),
				),
			},
			// Import: test import functionality
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"entity_name"},
			},
		},
	})
}

func TestAcc_ResourceNotificationChannelDagTrigger(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	channelName := fmt.Sprintf("%v_dagtrigger", namePrefix)
	resourceVar := fmt.Sprintf("astro_notification_channel.%v", channelName)

	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckNotificationChannelDestroyed(t, channelName),
		),
		Steps: []resource.TestStep{
			// Validate: empty dag_id
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannel(notificationChannelInput{
					Name:       channelName,
					Type:       "DAGTRIGGER",
					EntityId:   deploymentId,
					EntityType: "DEPLOYMENT",
					Definition: map[string]interface{}{
						"dag_id":               "",
						"deployment_api_token": "test-token",
						"deployment_id":        deploymentId,
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Length"),
			},
			// Validate: empty deployment_api_token
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannel(notificationChannelInput{
					Name:       channelName,
					Type:       "DAGTRIGGER",
					EntityId:   deploymentId,
					EntityType: "DEPLOYMENT",
					Definition: map[string]interface{}{
						"dag_id":               "test_dag",
						"deployment_api_token": "",
						"deployment_id":        deploymentId,
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Length"),
			},
			// Validate: empty deployment_id
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannel(notificationChannelInput{
					Name:       channelName,
					Type:       "DAGTRIGGER",
					EntityId:   deploymentId,
					EntityType: "DEPLOYMENT",
					Definition: map[string]interface{}{
						"dag_id":               "test_dag",
						"deployment_api_token": "test-token",
						"deployment_id":        "",
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Length"),
			},
			// Create: DAGTRIGGER notification channel
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannel(notificationChannelInput{
					Name:       channelName,
					Type:       "DAGTRIGGER",
					EntityId:   deploymentId,
					EntityType: "DEPLOYMENT",
					Definition: map[string]interface{}{
						"dag_id":               "notification_dag",
						"deployment_api_token": "test-api-token-12345",
						"deployment_id":        deploymentId,
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttr(resourceVar, "name", channelName),
					resource.TestCheckResourceAttr(resourceVar, "type", "DAGTRIGGER"),
					resource.TestCheckResourceAttr(resourceVar, "entity_id", deploymentId),
					resource.TestCheckResourceAttr(resourceVar, "entity_type", "DEPLOYMENT"),
					resource.TestCheckResourceAttr(resourceVar, "definition.dag_id", "notification_dag"),
					resource.TestCheckResourceAttr(resourceVar, "definition.deployment_api_token", "test-api-token-12345"),
					resource.TestCheckResourceAttr(resourceVar, "definition.deployment_id", deploymentId),
					resource.TestCheckResourceAttr(resourceVar, "is_shared", "false"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by.id"),
					testAccCheckNotificationChannelExists(t, channelName),
				),
			},
			// Update: dag_id and token
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannel(notificationChannelInput{
					Name:       channelName,
					Type:       "DAGTRIGGER",
					EntityId:   deploymentId,
					EntityType: "DEPLOYMENT",
					Definition: map[string]interface{}{
						"dag_id":               "updated_notification_dag",
						"deployment_api_token": "updated-api-token-67890",
						"deployment_id":        deploymentId,
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "definition.dag_id", "updated_notification_dag"),
					resource.TestCheckResourceAttr(resourceVar, "definition.deployment_api_token", "updated-api-token-67890"),
					testAccCheckNotificationChannelExists(t, channelName),
				),
			},
			// Import: test import functionality
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"entity_name"},
			},
		},
	})
}

func TestAcc_ResourceNotificationChannelPagerDuty(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	channelName := fmt.Sprintf("%v_pagerduty", namePrefix)
	resourceVar := fmt.Sprintf("astro_notification_channel.%v", channelName)

	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckNotificationChannelDestroyed(t, channelName),
		),
		Steps: []resource.TestStep{
			// Validate: empty integration_key
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannel(notificationChannelInput{
					Name:       channelName,
					Type:       "PAGERDUTY",
					EntityId:   deploymentId,
					EntityType: "DEPLOYMENT",
					Definition: map[string]interface{}{
						"integration_key": "",
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Length"),
			},
			// Create: PAGERDUTY notification channel
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannel(notificationChannelInput{
					Name:       channelName,
					Type:       "PAGERDUTY",
					EntityId:   deploymentId,
					EntityType: "DEPLOYMENT",
					Definition: map[string]interface{}{
						"integration_key": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttr(resourceVar, "name", channelName),
					resource.TestCheckResourceAttr(resourceVar, "type", "PAGERDUTY"),
					resource.TestCheckResourceAttr(resourceVar, "entity_id", deploymentId),
					resource.TestCheckResourceAttr(resourceVar, "entity_type", "DEPLOYMENT"),
					resource.TestCheckResourceAttr(resourceVar, "definition.integration_key", "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"),
					resource.TestCheckResourceAttr(resourceVar, "is_shared", "false"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by.id"),
					testAccCheckNotificationChannelExists(t, channelName),
				),
			},
			// Update: integration_key
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannel(notificationChannelInput{
					Name:       channelName,
					Type:       "PAGERDUTY",
					EntityId:   deploymentId,
					EntityType: "DEPLOYMENT",
					Definition: map[string]interface{}{
						"integration_key": "p6o5n4m3l2k1j0i9h8g7f6e5d4c3b2a1",
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "definition.integration_key", "p6o5n4m3l2k1j0i9h8g7f6e5d4c3b2a1"),
					testAccCheckNotificationChannelExists(t, channelName),
				),
			},
			// Import: test import functionality
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"entity_name"},
			},
		},
	})
}

func TestAcc_ResourceNotificationChannelOpsGenie(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	channelName := fmt.Sprintf("%v_opsgenie", namePrefix)
	resourceVar := fmt.Sprintf("astro_notification_channel.%v", channelName)

	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckNotificationChannelDestroyed(t, channelName),
		),
		Steps: []resource.TestStep{
			// Validate: empty api_key
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannel(notificationChannelInput{
					Name:       channelName,
					Type:       "OPSGENIE",
					EntityId:   deploymentId,
					EntityType: "DEPLOYMENT",
					Definition: map[string]interface{}{
						"api_key": "",
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Length"),
			},
			// Create: OPSGENIE notification channel
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannel(notificationChannelInput{
					Name:       channelName,
					Type:       "OPSGENIE",
					EntityId:   deploymentId,
					EntityType: "DEPLOYMENT",
					Definition: map[string]interface{}{
						"api_key": "00000000-0000-0000-0000-000000000000",
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttr(resourceVar, "name", channelName),
					resource.TestCheckResourceAttr(resourceVar, "type", "OPSGENIE"),
					resource.TestCheckResourceAttr(resourceVar, "entity_id", deploymentId),
					resource.TestCheckResourceAttr(resourceVar, "entity_type", "DEPLOYMENT"),
					resource.TestCheckResourceAttr(resourceVar, "definition.api_key", "00000000-0000-0000-0000-000000000000"),
					resource.TestCheckResourceAttr(resourceVar, "is_shared", "false"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by.id"),
					testAccCheckNotificationChannelExists(t, channelName),
				),
			},
			// Update: api_key
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannel(notificationChannelInput{
					Name:       channelName,
					Type:       "OPSGENIE",
					EntityId:   deploymentId,
					EntityType: "DEPLOYMENT",
					Definition: map[string]interface{}{
						"api_key": "11111111-1111-1111-1111-111111111111",
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "definition.api_key", "11111111-1111-1111-1111-111111111111"),
					testAccCheckNotificationChannelExists(t, channelName),
				),
			},
			// Import: test import functionality
			{
				ResourceName:            resourceVar,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"entity_name"},
			},
		},
	})
}

func TestAcc_ResourceNotificationChannelOrganization(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	channelName := fmt.Sprintf("%v_org_email", namePrefix)
	resourceVar := fmt.Sprintf("astro_notification_channel.%v", channelName)

	organizationId := os.Getenv("HOSTED_ORGANIZATION_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckNotificationChannelDestroyed(t, channelName),
		),
		Steps: []resource.TestStep{
			// Create: EMAIL notification channel for organization with is_shared
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannel(notificationChannelInput{
					Name:       channelName,
					Type:       "EMAIL",
					EntityId:   organizationId,
					EntityType: "ORGANIZATION",
					IsShared:   lo.ToPtr(true),
					Definition: map[string]interface{}{
						"recipients": []string{"org@example.com"},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttr(resourceVar, "name", channelName),
					resource.TestCheckResourceAttr(resourceVar, "type", "EMAIL"),
					resource.TestCheckResourceAttr(resourceVar, "entity_id", organizationId),
					resource.TestCheckResourceAttr(resourceVar, "entity_type", "ORGANIZATION"),
					resource.TestCheckResourceAttr(resourceVar, "is_shared", "true"),
					resource.TestCheckResourceAttr(resourceVar, "organization_id", organizationId),
					resource.TestCheckResourceAttr(resourceVar, "workspace_id", ""),
					resource.TestCheckResourceAttr(resourceVar, "deployment_id", ""),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by.id"),
					testAccCheckNotificationChannelExists(t, channelName),
				),
			},
			// Update: is_shared to false
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + notificationChannel(notificationChannelInput{
					Name:       channelName,
					Type:       "EMAIL",
					EntityId:   organizationId,
					EntityType: "ORGANIZATION",
					IsShared:   lo.ToPtr(false),
					Definition: map[string]interface{}{
						"recipients": []string{"org@example.com"},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "is_shared", "false"),
					testAccCheckNotificationChannelExists(t, channelName),
				),
			},
		},
	})
}

// Helper types and functions

type notificationChannelInput struct {
	Name       string
	Type       string
	EntityId   string
	EntityType string
	IsShared   *bool
	Definition map[string]interface{}
}

func notificationChannel(input notificationChannelInput) string {
	// Build definition string
	definitionStr := "definition = {\n"
	for k, v := range input.Definition {
		switch val := v.(type) {
		case string:
			definitionStr += fmt.Sprintf("\t\t%s = \"%s\"\n", k, val)
		case []string:
			definitionStr += fmt.Sprintf("\t\t%s = [", k)
			for i, s := range val {
				if i > 0 {
					definitionStr += ", "
				}
				definitionStr += fmt.Sprintf("\"%s\"", s)
			}
			definitionStr += "]\n"
		}
	}
	definitionStr += "\t}"

	// Build is_shared string
	isSharedStr := ""
	if input.IsShared != nil {
		isSharedStr = fmt.Sprintf("\n\tis_shared = %t", *input.IsShared)
	}

	return fmt.Sprintf(`
resource "astro_notification_channel" "%s" {
	name = "%s"
	type = "%s"
	entity_id = "%s"
	entity_type = "%s"
	%s
	%s
}`, input.Name, input.Name, input.Type, input.EntityId, input.EntityType, isSharedStr, definitionStr)
}

func testAccCheckNotificationChannelExists(t *testing.T, channelName string) func(s *terraform.State) error {
	t.Helper()
	return func(state *terraform.State) error {
		client, err := utils.GetTestPlatformClient(true)
		assert.NoError(t, err)

		organizationId := os.Getenv("HOSTED_ORGANIZATION_ID")
		ctx := context.Background()

		// List all notification channels in the organization
		t.Logf("Listing all notification channels in organization: %s", organizationId)
		resp, err := client.ListNotificationChannelsWithResponse(ctx, organizationId, &platform.ListNotificationChannelsParams{
			Limit: lo.ToPtr(0),
		})
		if err != nil {
			return fmt.Errorf("failed to list notification channels: %v", err)
		}
		if resp == nil {
			return fmt.Errorf("nil response from list notification channels")
		}
		if resp.JSON200 == nil {
			status, diag := clients.NormalizeAPIError(ctx, resp.HTTPResponse, resp.Body)
			return fmt.Errorf("response JSON200 is nil status: %v, err: %v", status, diag.Detail())
		}

		t.Logf("Found %d total notification channels in organization", len(resp.JSON200.NotificationChannels))

		// Check in unfiltered list first
		for _, channel := range resp.JSON200.NotificationChannels {
			if channel.Name == channelName {
				t.Logf("Found notification channel %s in unfiltered list", channelName)
				return nil
			}
		}

		return fmt.Errorf("notification channel %s not found", channelName)
	}
}

func testAccCheckNotificationChannelDestroyed(t *testing.T, channelName string) func(s *terraform.State) error {
	t.Helper()
	return func(state *terraform.State) error {
		client, err := utils.GetTestPlatformClient(true)
		assert.NoError(t, err)

		organizationId := os.Getenv("HOSTED_ORGANIZATION_ID")
		ctx := context.Background()

		// List all notification channels to check if it still exists
		resp, err := client.ListNotificationChannelsWithResponse(ctx, organizationId, &platform.ListNotificationChannelsParams{
			Limit: lo.ToPtr(0),
		})
		if err != nil {
			return fmt.Errorf("failed to list notification channels: %v", err)
		}
		if resp == nil || resp.JSON200 == nil {
			status, diag := clients.NormalizeAPIError(ctx, resp.HTTPResponse, resp.Body)
			return fmt.Errorf("response JSON200 is nil status: %v, err: %v", status, diag.Detail())
		}

		for _, channel := range resp.JSON200.NotificationChannels {
			if channel.Name == channelName {
				return fmt.Errorf("notification channel %s still exists", channelName)
			}
		}

		return nil
	}
}
