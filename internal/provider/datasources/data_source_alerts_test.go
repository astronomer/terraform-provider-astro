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
)

func TestAcc_DataSourceAlerts(t *testing.T) {
	tfVarName := "test_data_alerts"
	tfWorkspaceId := os.Getenv("HOSTED_WORKSPACE_ID")
	tfDeploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")
	tfAlertId := os.Getenv("HOSTED_ALERT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			astronomerprovider.TestAccPreCheck(t)
		},
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alerts(tfVarName),
				Check: resource.ComposeTestCheckFunc(
					checkAlerts(tfVarName),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alertsFilterWorkspaceIds(tfVarName, []string{tfWorkspaceId}),
				Check: resource.ComposeTestCheckFunc(
					checkAlerts(tfVarName),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alertsFilterDeploymentIds(tfVarName, []string{tfDeploymentId}),
				Check: resource.ComposeTestCheckFunc(
					checkAlerts(tfVarName),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alertsFilterAlertIds(tfVarName, []string{tfAlertId}),
				Check: resource.ComposeTestCheckFunc(
					checkAlerts(tfVarName),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alertsFilterAlertTypes(tfVarName, []string{string(platform.CreateTaskFailureAlertRequestTypeTASKFAILURE)}),
				Check: resource.ComposeTestCheckFunc(
					checkAlerts(tfVarName),
				),
			},
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alertsFilterEntityType(tfVarName, string(platform.AlertEntityTypeDEPLOYMENT)),
				Check: resource.ComposeTestCheckFunc(
					checkAlerts(tfVarName),
				),
			},
		},
	})
}

func alerts(tfVarName string) string {
	return fmt.Sprintf(`
data astro_alerts "%v" {}`, tfVarName)
}

func alertsFilterWorkspaceIds(tfVarName string, workspaceIds []string) string {
	var quoted []string
	for _, id := range workspaceIds {
		quoted = append(quoted, fmt.Sprintf("%q", id))
	}
	return fmt.Sprintf(`
data astro_alerts "%v" {
	workspace_ids = [%s]
}`, tfVarName, strings.Join(quoted, ","))
}

func alertsFilterDeploymentIds(tfVarName string, deploymentIds []string) string {
	var quoted []string
	for _, id := range deploymentIds {
		quoted = append(quoted, fmt.Sprintf("%q", id))
	}
	return fmt.Sprintf(`
data astro_alerts "%v" {
	deployment_ids = [%s]
}`, tfVarName, strings.Join(quoted, ","))
}

func alertsFilterAlertIds(tfVarName string, alertIds []string) string {
	var quoted []string
	for _, id := range alertIds {
		quoted = append(quoted, fmt.Sprintf("%q", id))
	}
	return fmt.Sprintf(`
data astro_alerts "%v" {
	alert_ids = [%s]
}`, tfVarName, strings.Join(quoted, ","))
}

func alertsFilterAlertTypes(tfVarName string, alertTypes []string) string {
	var quoted []string
	for _, t := range alertTypes {
		quoted = append(quoted, fmt.Sprintf("%q", t))
	}
	return fmt.Sprintf(`
data astro_alerts "%v" {
	alert_types = [%s]
}`, tfVarName, strings.Join(quoted, ","))
}

func alertsFilterEntityType(tfVarName string, entityType string) string {
	return fmt.Sprintf(`
data astro_alerts "%v" {
	entity_type = "%v"
}`, tfVarName, entityType)
}

func checkAlerts(tfVarName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		instanceState, numalerts, err := utils.GetDataSourcesLength(s, tfVarName, "alerts")
		if err != nil {
			return err
		}
		if numalerts == 0 {
			return fmt.Errorf("expected alerts to be greater or equal to 1, got %s", instanceState.Attributes["alerts.#"])
		}

		// Check the first alert
		alertsIdx := 0

		id := fmt.Sprintf("alerts.%d.id", alertsIdx)
		if instanceState.Attributes[id] == "" {
			return fmt.Errorf("expected 'id' to be set")
		}
		name := fmt.Sprintf("alerts.%d.name", alertsIdx)
		if instanceState.Attributes[name] == "" {
			return fmt.Errorf("expected 'name' to be set")
		}
		entityId := fmt.Sprintf("alerts.%d.entity_id", alertsIdx)
		if instanceState.Attributes[entityId] == "" {
			return fmt.Errorf("expected 'entity_id' to be set")
		}
		entityType := fmt.Sprintf("alerts.%d.entity_type", alertsIdx)
		if instanceState.Attributes[entityType] == "" {
			return fmt.Errorf("expected 'entity_type' to be set")
		}
		organizationId := fmt.Sprintf("alerts.%d.organization_id", alertsIdx)
		if instanceState.Attributes[organizationId] == "" {
			return fmt.Errorf("expected 'organization_id' to be set")
		}
		workspaceId := fmt.Sprintf("alerts.%d.workspace_id", alertsIdx)
		if instanceState.Attributes[workspaceId] == "" {
			return fmt.Errorf("expected 'workspace_id' to be set")
		}
		deploymentId := fmt.Sprintf("alerts.%d.deployment_id", alertsIdx)
		if instanceState.Attributes[deploymentId] == "" {
			return fmt.Errorf("expected 'deployment_id' to be set")
		}
		severity := fmt.Sprintf("alerts.%d.severity", alertsIdx)
		if instanceState.Attributes[severity] == "" {
			return fmt.Errorf("expected 'severity' to be set")
		}
		alertType := fmt.Sprintf("alerts.%d.type", alertsIdx)
		if instanceState.Attributes[alertType] == "" {
			return fmt.Errorf("expected 'type' to be set")
		}
		propCountKey := fmt.Sprintf("alerts.%d.rules.properties.%%", alertsIdx)
		if instanceState.Attributes[propCountKey] == "" {
			return fmt.Errorf("expected 'rules.properties' to have at least one entry")
		}
		patternCountKey := fmt.Sprintf("alerts.%d.rules.pattern_matches.#", alertsIdx)
		if instanceState.Attributes[patternCountKey] == "" {
			return fmt.Errorf("expected 'rules.pattern_matches' to have at least one entry")
		}
		createdAt := fmt.Sprintf("alerts.%d.created_at", alertsIdx)
		if instanceState.Attributes[createdAt] == "" {
			return fmt.Errorf("expected 'created_at' to be set")
		}
		updatedAt := fmt.Sprintf("alerts.%d.updated_at", alertsIdx)
		if instanceState.Attributes[updatedAt] == "" {
			return fmt.Errorf("expected 'updated_at' to be set")
		}
		createdBy := fmt.Sprintf("alerts.%d.created_by.id", alertsIdx)
		if instanceState.Attributes[createdBy] == "" {
			return fmt.Errorf("expected 'created_by.id' to be set")
		}
		updatedBy := fmt.Sprintf("alerts.%d.updated_by.id", alertsIdx)
		if instanceState.Attributes[updatedBy] == "" {
			return fmt.Errorf("expected 'updated_by.id' to be set")
		}
		return nil
	}
}
