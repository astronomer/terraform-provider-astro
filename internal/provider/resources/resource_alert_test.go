package resources_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/astronomer/terraform-provider-astro/internal/clients"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	astronomerprovider "github.com/astronomer/terraform-provider-astro/internal/provider"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAcc_ResourceAlertDagFailure(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	alertName := fmt.Sprintf("%v_dag_failure", namePrefix)
	resourceVar := fmt.Sprintf("astro_alert.%v", alertName)

	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")
	notificationChannelId := os.Getenv("HOSTED_NOTIFICATION_CHANNEL_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckAlertDestroyed(t, alertName),
		),
		Steps: []resource.TestStep{
			// Validate: invalid entity type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagFailureAlertRequestTypeDAGFAILURE),
					Severity:               string(platform.CreateDagFailureAlertRequestSeverityINFO),
					EntityId:               deploymentId,
					EntityType:             "WORKSPACE",
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"test_dag"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Validate: invalid alert type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   "INVALID_ALERT_TYPE",
					Severity:               string(platform.CreateDagFailureAlertRequestSeverityINFO),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagFailureAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"test_dag"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Validate: invalid severity
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagFailureAlertRequestTypeDAGFAILURE),
					Severity:               "INVALID_SEVERITY",
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagFailureAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"test_dag"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Validate: invalid deployment ID
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagFailureAlertRequestTypeDAGFAILURE),
					Severity:               string(platform.CreateDagFailureAlertRequestSeverityINFO),
					EntityId:               "clx4825jb068z01j9931ib5ga",
					EntityType:             string(platform.CreateDagFailureAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": "clx4825jb068z01j9931ib5ga",
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"test_dag"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("deployment with id .* not found"),
			},
			// Validate: invalid pattern match entity type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagFailureAlertRequestTypeDAGFAILURE),
					Severity:               string(platform.CreateDagFailureAlertRequestSeverityINFO),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagFailureAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   "INVALID_ENTITY_TYPE",
							OperatorType: string(platform.IS),
							Values:       []string{"test_dag"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("(?s).*entityType.*should be one of.*TASK_ID.*DAG_ID"),
			},
			// Validate: invalid pattern match operator type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagFailureAlertRequestTypeDAGFAILURE),
					Severity:               string(platform.CreateDagFailureAlertRequestSeverityINFO),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagFailureAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: "INVALID_OPERATOR_TYPE",
							Values:       []string{"test_dag"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("(?s).*operatorType.*should be one of"),
			},
			// Validate: empty pattern match values
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagFailureAlertRequestTypeDAGFAILURE),
					Severity:               string(platform.CreateDagFailureAlertRequestSeverityINFO),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagFailureAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{""},
						},
					},
				}),
				ExpectError: regexp.MustCompile("(?s).*values[0].*should be min: 1"),
			},
			// Create: DAG_FAILURE alert
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagFailureAlertRequestTypeDAGFAILURE),
					Severity:               string(platform.CreateDagFailureAlertRequestSeverityINFO),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagFailureAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"test_dag", "another_dag"},
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttr(resourceVar, "name", alertName),
					resource.TestCheckResourceAttr(resourceVar, "type", string(platform.CreateDagFailureAlertRequestTypeDAGFAILURE)),
					resource.TestCheckResourceAttr(resourceVar, "severity", string(platform.CreateDagFailureAlertRequestSeverityINFO)),
					resource.TestCheckResourceAttr(resourceVar, "entity_id", deploymentId),
					resource.TestCheckResourceAttr(resourceVar, "entity_type", string(platform.CreateDagFailureAlertRequestEntityTypeDEPLOYMENT)),
					resource.TestCheckResourceAttrSet(resourceVar, "entity_name"),
					resource.TestCheckResourceAttr(resourceVar, "notification_channel_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "rules.properties.deployment_id", deploymentId),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.entity_type", string(platform.DAGID)),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.operator_type", string(platform.IS)),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.values.#", "2"),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.values.0", "test_dag"),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.values.1", "another_dag"),
					resource.TestCheckResourceAttrSet(resourceVar, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "workspace_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "deployment_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by.id"),
					testAccCheckAlertExists(t, alertName),
				),
			},
			// Update: severity and pattern match operator type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagFailureAlertRequestTypeDAGFAILURE),
					Severity:               string(platform.CreateDagFailureAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagFailureAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.INCLUDES),
							Values:       []string{"test"},
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "severity", string(platform.CreateDagFailureAlertRequestSeverityWARNING)),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.#", "1"),
					testAccCheckAlertExists(t, alertName),
				),
			},
			// Update: pattern matches with multiple conditions
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagFailureAlertRequestTypeDAGFAILURE),
					Severity:               string(platform.CreateDagFailureAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagFailureAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.INCLUDES),
							Values:       []string{"test"},
						},
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.EXCLUDES),
							Values:       []string{"abc"},
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.#", "2"),
					testAccCheckAlertExists(t, alertName),
				),
			},
			// Import: test import functionality
			{
				ResourceName:      resourceVar,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ResourceAlertDagSuccess(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	alertName := fmt.Sprintf("%v_dag_success", namePrefix)
	resourceVar := fmt.Sprintf("astro_alert.%v", alertName)

	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")
	notificationChannelId := os.Getenv("HOSTED_NOTIFICATION_CHANNEL_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckAlertDestroyed(t, alertName),
		),
		Steps: []resource.TestStep{
			// Validate: invalid entity type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagSuccessAlertRequestTypeDAGSUCCESS),
					Severity:               string(platform.CreateDagSuccessAlertRequestSeverityINFO),
					EntityId:               deploymentId,
					EntityType:             "WORKSPACE",
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"success_dag"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Validate: invalid alert type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   "INVALID_ALERT_TYPE",
					Severity:               string(platform.CreateDagSuccessAlertRequestSeverityINFO),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagSuccessAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"success_dag"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Validate: invalid severity
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagSuccessAlertRequestTypeDAGSUCCESS),
					Severity:               "INVALID_SEVERITY",
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagSuccessAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"success_dag"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Validate: invalid pattern match entity type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagSuccessAlertRequestTypeDAGSUCCESS),
					Severity:               string(platform.CreateDagSuccessAlertRequestSeverityINFO),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagSuccessAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   "INVALID_ENTITY_TYPE",
							OperatorType: string(platform.IS),
							Values:       []string{"success_dag"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("(?s).*entityType.*should be one of.*TASK_ID.*DAG_ID"),
			},
			// Validate: invalid pattern match operator type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagSuccessAlertRequestTypeDAGSUCCESS),
					Severity:               string(platform.CreateDagSuccessAlertRequestSeverityINFO),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagSuccessAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: "INVALID_OPERATOR_TYPE",
							Values:       []string{"success_dag"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("(?s).*operatorType.*should be one of"),
			},
			// Validate: empty pattern match values
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagSuccessAlertRequestTypeDAGSUCCESS),
					Severity:               string(platform.CreateDagSuccessAlertRequestSeverityINFO),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagSuccessAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{""},
						},
					},
				}),
				ExpectError: regexp.MustCompile("(?s).*values[0].*should be min: 1"),
			},
			// Create: DAG_SUCCESS alert
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagSuccessAlertRequestTypeDAGSUCCESS),
					Severity:               string(platform.CreateDagSuccessAlertRequestSeverityINFO),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagSuccessAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"success_dag"},
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttr(resourceVar, "name", alertName),
					resource.TestCheckResourceAttr(resourceVar, "type", string(platform.CreateDagSuccessAlertRequestTypeDAGSUCCESS)),
					resource.TestCheckResourceAttr(resourceVar, "severity", string(platform.CreateDagSuccessAlertRequestSeverityINFO)),
					resource.TestCheckResourceAttr(resourceVar, "entity_id", deploymentId),
					resource.TestCheckResourceAttr(resourceVar, "entity_type", string(platform.CreateDagSuccessAlertRequestEntityTypeDEPLOYMENT)),
					resource.TestCheckResourceAttrSet(resourceVar, "entity_name"),
					resource.TestCheckResourceAttr(resourceVar, "notification_channel_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "rules.properties.deployment_id", deploymentId),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.entity_type", string(platform.DAGID)),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.operator_type", string(platform.IS)),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.values.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.values.0", "success_dag"),
					resource.TestCheckResourceAttrSet(resourceVar, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "workspace_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "deployment_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by.id"),
					testAccCheckAlertExists(t, alertName),
				),
			},
			// Update: pattern matches with multiple conditions
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagSuccessAlertRequestTypeDAGSUCCESS),
					Severity:               string(platform.CreateDagSuccessAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagSuccessAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.INCLUDES),
							Values:       []string{"success", "complete"},
						},
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.EXCLUDES),
							Values:       []string{"failed", "error"},
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "severity", string(platform.CreateDagSuccessAlertRequestSeverityWARNING)),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.#", "2"),
					testAccCheckAlertExists(t, alertName),
				),
			},
			// Import: test import functionality
			{
				ResourceName:      resourceVar,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ResourceAlertDagDuration(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	alertName := fmt.Sprintf("%v_dag_duration", namePrefix)
	resourceVar := fmt.Sprintf("astro_alert.%v", alertName)

	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")
	notificationChannelId := os.Getenv("HOSTED_NOTIFICATION_CHANNEL_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckAlertDestroyed(t, alertName),
		),
		Steps: []resource.TestStep{
			// Validate: invalid entity type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagDurationAlertRequestTypeDAGDURATION),
					Severity:               string(platform.CreateDagDurationAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             "WORKSPACE",
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":        deploymentId,
						"dag_duration_seconds": 300,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"slow_dag"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Validate: invalid alert type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   "INVALID_ALERT_TYPE",
					Severity:               string(platform.CreateDagDurationAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagDurationAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":        deploymentId,
						"dag_duration_seconds": 300,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"slow_dag"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Validate: invalid severity
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagDurationAlertRequestTypeDAGDURATION),
					Severity:               "INVALID_SEVERITY",
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagDurationAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":        deploymentId,
						"dag_duration_seconds": 300,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"slow_dag"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Validate: negative duration
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagDurationAlertRequestTypeDAGDURATION),
					Severity:               string(platform.CreateDagDurationAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagDurationAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":        deploymentId,
						"dag_duration_seconds": -100,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"slow_dag"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value|must be greater than 0"),
			},
			// Validate: missing required property
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagDurationAlertRequestTypeDAGDURATION),
					Severity:               string(platform.CreateDagDurationAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagDurationAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
						// Missing dag_duration_seconds
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"slow_dag"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("dag_duration_seconds is required for DAG_DURATION alerts"),
			},
			// Validate: invalid pattern match entity type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagDurationAlertRequestTypeDAGDURATION),
					Severity:               string(platform.CreateDagDurationAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagDurationAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":        deploymentId,
						"dag_duration_seconds": 300,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   "INVALID_ENTITY_TYPE",
							OperatorType: string(platform.IS),
							Values:       []string{"slow_dag"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("(?s).*entityType.*should be one of.*TASK_ID.*DAG_ID"),
			},
			// Validate: invalid pattern match operator type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagDurationAlertRequestTypeDAGDURATION),
					Severity:               string(platform.CreateDagDurationAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagDurationAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":        deploymentId,
						"dag_duration_seconds": 300,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: "INVALID_OPERATOR_TYPE",
							Values:       []string{"slow_dag"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("(?s).*operatorType.*should be one of"),
			},
			// Validate: empty pattern match values
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagDurationAlertRequestTypeDAGDURATION),
					Severity:               string(platform.CreateDagDurationAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagDurationAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":        deploymentId,
						"dag_duration_seconds": 300,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{""},
						},
					},
				}),
				ExpectError: regexp.MustCompile("(?s).*values.*should be one of"),
			},
			// Create: DAG_DURATION alert
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagDurationAlertRequestTypeDAGDURATION),
					Severity:               string(platform.CreateDagDurationAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagDurationAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":        deploymentId,
						"dag_duration_seconds": 300,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"slow_dag"},
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttr(resourceVar, "name", alertName),
					resource.TestCheckResourceAttr(resourceVar, "type", string(platform.CreateDagDurationAlertRequestTypeDAGDURATION)),
					resource.TestCheckResourceAttr(resourceVar, "severity", string(platform.CreateDagDurationAlertRequestSeverityWARNING)),
					resource.TestCheckResourceAttr(resourceVar, "rules.properties.deployment_id", deploymentId),
					resource.TestCheckResourceAttr(resourceVar, "rules.properties.dag_duration_seconds", "300"),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.entity_type", string(platform.DAGID)),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.operator_type", string(platform.IS)),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.values.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.values.0", "slow_dag"),
					resource.TestCheckResourceAttrSet(resourceVar, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "workspace_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "deployment_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by.id"),
					testAccCheckAlertExists(t, alertName),
				),
			},
			// Update: duration threshold
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagDurationAlertRequestTypeDAGDURATION),
					Severity:               string(platform.CreateDagDurationAlertRequestSeverityCRITICAL),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagDurationAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":        deploymentId,
						"dag_duration_seconds": 600,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"slow_dag"},
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "severity", string(platform.CreateDagDurationAlertRequestSeverityCRITICAL)),
					resource.TestCheckResourceAttr(resourceVar, "rules.properties.dag_duration_seconds", "600"),
					testAccCheckAlertExists(t, alertName),
				),
			},
			// Update: pattern matches with multiple conditions
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagDurationAlertRequestTypeDAGDURATION),
					Severity:               string(platform.CreateDagDurationAlertRequestSeverityCRITICAL),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagDurationAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":        deploymentId,
						"dag_duration_seconds": 600,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.INCLUDES),
							Values:       []string{"slow", "long"},
						},
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.EXCLUDES),
							Values:       []string{"test"},
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.#", "2"),
					testAccCheckAlertExists(t, alertName),
				),
			},
			// Import: test import functionality
			{
				ResourceName:      resourceVar,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ResourceAlertDagTimeliness(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	alertName := fmt.Sprintf("%v_dag_timeliness", namePrefix)
	resourceVar := fmt.Sprintf("astro_alert.%v", alertName)

	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")
	notificationChannelId := os.Getenv("HOSTED_NOTIFICATION_CHANNEL_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckAlertDestroyed(t, alertName),
		),
		Steps: []resource.TestStep{
			// Validate: invalid entity type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagTimelinessAlertRequestTypeDAGTIMELINESS),
					Severity:               string(platform.CreateDagTimelinessAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             "WORKSPACE",
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":            deploymentId,
						"dag_deadline":             "08:00",
						"days_of_week":             []string{"MONDAY"},
						"look_back_period_seconds": 3600,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"daily_etl"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Validate: invalid dag_deadline format
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagTimelinessAlertRequestTypeDAGTIMELINESS),
					Severity:               string(platform.CreateDagTimelinessAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagTimelinessAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":            deploymentId,
						"dag_deadline":             "25:00", // Invalid hour
						"days_of_week":             []string{"MONDAY"},
						"look_back_period_seconds": 3600,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"daily_etl"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid|invalid time format"),
			},
			// Validate: invalid days_of_week value
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagTimelinessAlertRequestTypeDAGTIMELINESS),
					Severity:               string(platform.CreateDagTimelinessAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagTimelinessAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":            deploymentId,
						"dag_deadline":             "08:00",
						"days_of_week":             []string{"INVALID_DAY"},
						"look_back_period_seconds": 3600,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"daily_etl"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid|invalid day"),
			},
			// Validate: negative look_back_period_seconds
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagTimelinessAlertRequestTypeDAGTIMELINESS),
					Severity:               string(platform.CreateDagTimelinessAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagTimelinessAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":            deploymentId,
						"dag_deadline":             "08:00",
						"days_of_week":             []string{"MONDAY"},
						"look_back_period_seconds": -3600,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"daily_etl"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid|must be greater than 0"),
			},
			// Validate: empty days_of_week
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagTimelinessAlertRequestTypeDAGTIMELINESS),
					Severity:               string(platform.CreateDagTimelinessAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagTimelinessAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":            deploymentId,
						"dag_deadline":             "08:00",
						"days_of_week":             []string{},
						"look_back_period_seconds": 3600,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"daily_etl"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid|at least one day"),
			},
			// Validate: invalid pattern match entity type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagTimelinessAlertRequestTypeDAGTIMELINESS),
					Severity:               string(platform.CreateDagTimelinessAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagTimelinessAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":            deploymentId,
						"dag_deadline":             "08:00",
						"days_of_week":             []string{"MONDAY"},
						"look_back_period_seconds": 3600,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   "INVALID_ENTITY_TYPE",
							OperatorType: string(platform.IS),
							Values:       []string{"daily_etl"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("(?s).*entityType.*should be one of.*TASK_ID.*DAG_ID"),
			},
			// Validate: invalid pattern match operator type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagTimelinessAlertRequestTypeDAGTIMELINESS),
					Severity:               string(platform.CreateDagTimelinessAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagTimelinessAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":            deploymentId,
						"dag_deadline":             "08:00",
						"days_of_week":             []string{"MONDAY"},
						"look_back_period_seconds": 3600,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: "INVALID_OPERATOR_TYPE",
							Values:       []string{"daily_etl"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("(?s).*operatorType.*should be one of"),
			},
			// Validate: empty pattern match values
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagTimelinessAlertRequestTypeDAGTIMELINESS),
					Severity:               string(platform.CreateDagTimelinessAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagTimelinessAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":            deploymentId,
						"dag_deadline":             "08:00",
						"days_of_week":             []string{"MONDAY"},
						"look_back_period_seconds": 3600,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{""},
						},
					},
				}),
				ExpectError: regexp.MustCompile("(?s).*values[0].*should be min: 1"),
			},
			// Create: DAG_TIMELINESS alert
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagTimelinessAlertRequestTypeDAGTIMELINESS),
					Severity:               string(platform.CreateDagTimelinessAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagTimelinessAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":            deploymentId,
						"dag_deadline":             "08:00",
						"days_of_week":             []string{"MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY"},
						"look_back_period_seconds": 3600,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"daily_etl"},
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttr(resourceVar, "name", alertName),
					resource.TestCheckResourceAttr(resourceVar, "type", string(platform.CreateDagTimelinessAlertRequestTypeDAGTIMELINESS)),
					resource.TestCheckResourceAttr(resourceVar, "severity", string(platform.CreateDagTimelinessAlertRequestSeverityWARNING)),
					resource.TestCheckResourceAttr(resourceVar, "rules.properties.deployment_id", deploymentId),
					resource.TestCheckResourceAttr(resourceVar, "rules.properties.dag_deadline", "08:00"),
					resource.TestCheckResourceAttr(resourceVar, "rules.properties.days_of_week.#", "5"),
					resource.TestCheckResourceAttr(resourceVar, "rules.properties.look_back_period_seconds", "3600"),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.entity_type", string(platform.DAGID)),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.operator_type", string(platform.IS)),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.values.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.values.0", "daily_etl"),
					resource.TestCheckResourceAttrSet(resourceVar, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "workspace_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "deployment_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by.id"),
					testAccCheckAlertExists(t, alertName),
				),
			},
			// Update: properties
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagTimelinessAlertRequestTypeDAGTIMELINESS),
					Severity:               string(platform.CreateDagTimelinessAlertRequestSeverityCRITICAL),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagTimelinessAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":            deploymentId,
						"dag_deadline":             "09:00",
						"days_of_week":             []string{"MONDAY", "WEDNESDAY", "FRIDAY"},
						"look_back_period_seconds": 7200,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"daily_etl"},
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "severity", string(platform.CreateDagTimelinessAlertRequestSeverityCRITICAL)),
					resource.TestCheckResourceAttr(resourceVar, "rules.properties.dag_deadline", "09:00"),
					resource.TestCheckResourceAttr(resourceVar, "rules.properties.days_of_week.#", "3"),
					resource.TestCheckResourceAttr(resourceVar, "rules.properties.look_back_period_seconds", "7200"),
					testAccCheckAlertExists(t, alertName),
				),
			},
			// Update: pattern matches with multiple conditions
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateDagTimelinessAlertRequestTypeDAGTIMELINESS),
					Severity:               string(platform.CreateDagTimelinessAlertRequestSeverityCRITICAL),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateDagTimelinessAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":            deploymentId,
						"dag_deadline":             "09:00",
						"days_of_week":             []string{"MONDAY", "WEDNESDAY", "FRIDAY"},
						"look_back_period_seconds": 7200,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.INCLUDES),
							Values:       []string{"etl", "daily"},
						},
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.EXCLUDES),
							Values:       []string{"test", "dev"},
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.#", "2"),
					testAccCheckAlertExists(t, alertName),
				),
			},
			// Import: test import functionality
			{
				ResourceName:      resourceVar,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ResourceAlertTaskFailure(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	alertName := fmt.Sprintf("%v_task_failure", namePrefix)
	resourceVar := fmt.Sprintf("astro_alert.%v", alertName)

	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")
	notificationChannelId := os.Getenv("HOSTED_NOTIFICATION_CHANNEL_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckAlertDestroyed(t, alertName),
		),
		Steps: []resource.TestStep{
			// Validate: invalid entity type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateTaskFailureAlertRequestTypeTASKFAILURE),
					Severity:               string(platform.CreateTaskFailureAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             "WORKSPACE",
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.TASKID),
							OperatorType: string(platform.IS),
							Values:       []string{"critical_task"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Validate: invalid alert type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   "INVALID_ALERT_TYPE",
					Severity:               string(platform.CreateTaskFailureAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateTaskFailureAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.TASKID),
							OperatorType: string(platform.IS),
							Values:       []string{"critical_task"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Validate: invalid severity
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateTaskFailureAlertRequestTypeTASKFAILURE),
					Severity:               "INVALID_SEVERITY",
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateTaskFailureAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.TASKID),
							OperatorType: string(platform.IS),
							Values:       []string{"critical_task"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Validate: invalid pattern match entity type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateTaskFailureAlertRequestTypeTASKFAILURE),
					Severity:               string(platform.CreateTaskFailureAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateTaskFailureAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   "INVALID_ENTITY_TYPE",
							OperatorType: string(platform.IS),
							Values:       []string{"critical_task"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("(?s).*entityType.*should be one of.*TASK_ID.*DAG_ID"),
			},
			// Validate: invalid pattern match operator type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateTaskFailureAlertRequestTypeTASKFAILURE),
					Severity:               string(platform.CreateTaskFailureAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateTaskFailureAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.TASKID),
							OperatorType: "INVALID_OPERATOR_TYPE",
							Values:       []string{"critical_task"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("(?s).*operatorType.*should be one of"),
			},
			// Validate: empty pattern match values
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateTaskFailureAlertRequestTypeTASKFAILURE),
					Severity:               string(platform.CreateTaskFailureAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateTaskFailureAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.TASKID),
							OperatorType: string(platform.IS),
							Values:       []string{""},
						},
					},
				}),
				ExpectError: regexp.MustCompile("(?s).*values[0].*should be min: 1"),
			},
			// Validate: using TASK pattern with invalid entity type (must be TASKID for tasks)
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateTaskFailureAlertRequestTypeTASKFAILURE),
					Severity:               string(platform.CreateTaskFailureAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateTaskFailureAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   "ENVIRONMENT", // Invalid for task alerts
							OperatorType: string(platform.IS),
							Values:       []string{"prod"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid|not allowed"),
			},
			// Create: TASK_FAILURE alert
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateTaskFailureAlertRequestTypeTASKFAILURE),
					Severity:               string(platform.CreateTaskFailureAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateTaskFailureAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.TASKID),
							OperatorType: string(platform.IS),
							Values:       []string{"critical_task"},
						},
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"important_dag"},
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttr(resourceVar, "name", alertName),
					resource.TestCheckResourceAttr(resourceVar, "type", string(platform.CreateTaskFailureAlertRequestTypeTASKFAILURE)),
					resource.TestCheckResourceAttr(resourceVar, "severity", string(platform.CreateTaskFailureAlertRequestSeverityWARNING)),
					resource.TestCheckResourceAttr(resourceVar, "entity_id", deploymentId),
					resource.TestCheckResourceAttr(resourceVar, "entity_type", string(platform.CreateTaskFailureAlertRequestEntityTypeDEPLOYMENT)),
					resource.TestCheckResourceAttr(resourceVar, "rules.properties.deployment_id", deploymentId),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.#", "2"),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.entity_type", string(platform.TASKID)),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.operator_type", string(platform.IS)),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.values.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.values.0", "critical_task"),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.1.entity_type", string(platform.DAGID)),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.1.operator_type", string(platform.IS)),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.1.values.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.1.values.0", "important_dag"),
					resource.TestCheckResourceAttrSet(resourceVar, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "workspace_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "deployment_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by.id"),
					testAccCheckAlertExists(t, alertName),
				),
			},
			// Update: pattern matches with complex conditions
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateTaskFailureAlertRequestTypeTASKFAILURE),
					Severity:               string(platform.CreateTaskFailureAlertRequestSeverityCRITICAL),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateTaskFailureAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id": deploymentId,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.TASKID),
							OperatorType: string(platform.INCLUDES),
							Values:       []string{"critical", "important"},
						},
						{
							EntityType:   string(platform.TASKID),
							OperatorType: string(platform.EXCLUDES),
							Values:       []string{"test"},
						},
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"production_pipeline"},
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.#", "3"),
					testAccCheckAlertExists(t, alertName),
				),
			},
			// Import: test import functionality
			{
				ResourceName:      resourceVar,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAcc_ResourceAlertTaskDuration(t *testing.T) {
	namePrefix := utils.GenerateTestResourceName(10)
	alertName := fmt.Sprintf("%v_task_duration", namePrefix)
	resourceVar := fmt.Sprintf("astro_alert.%v", alertName)

	deploymentId := os.Getenv("HOSTED_DEPLOYMENT_ID")
	notificationChannelId := os.Getenv("HOSTED_NOTIFICATION_CHANNEL_ID")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: astronomerprovider.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { astronomerprovider.TestAccPreCheck(t) },
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckAlertDestroyed(t, alertName),
		),
		Steps: []resource.TestStep{
			// Validate: invalid entity type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateTaskDurationAlertRequestTypeTASKDURATION),
					Severity:               string(platform.CreateTaskDurationAlertRequestSeverityINFO),
					EntityId:               deploymentId,
					EntityType:             "WORKSPACE",
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":         deploymentId,
						"task_duration_seconds": 60,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.TASKID),
							OperatorType: string(platform.IS),
							Values:       []string{"long_running_task"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Validate: invalid alert type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   "INVALID_ALERT_TYPE",
					Severity:               string(platform.CreateTaskDurationAlertRequestSeverityINFO),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateTaskDurationAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":         deploymentId,
						"task_duration_seconds": 60,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.TASKID),
							OperatorType: string(platform.IS),
							Values:       []string{"long_running_task"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Validate: invalid severity
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateTaskDurationAlertRequestTypeTASKDURATION),
					Severity:               "INVALID_SEVERITY",
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateTaskDurationAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":         deploymentId,
						"task_duration_seconds": 60,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.TASKID),
							OperatorType: string(platform.IS),
							Values:       []string{"long_running_task"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Match"),
			},
			// Validate: negative duration
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateTaskDurationAlertRequestTypeTASKDURATION),
					Severity:               string(platform.CreateTaskDurationAlertRequestSeverityINFO),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateTaskDurationAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":         deploymentId,
						"task_duration_seconds": -60,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.TASKID),
							OperatorType: string(platform.IS),
							Values:       []string{"long_running_task"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value|must be greater than 0"),
			},
			// Validate: zero duration
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateTaskDurationAlertRequestTypeTASKDURATION),
					Severity:               string(platform.CreateTaskDurationAlertRequestSeverityINFO),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateTaskDurationAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":         deploymentId,
						"task_duration_seconds": 0,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.TASKID),
							OperatorType: string(platform.IS),
							Values:       []string{"long_running_task"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("Invalid Attribute Value|must be greater than 0"),
			},
			// Validate: invalid pattern match entity type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateTaskDurationAlertRequestTypeTASKDURATION),
					Severity:               string(platform.CreateTaskDurationAlertRequestSeverityINFO),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateTaskDurationAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":         deploymentId,
						"task_duration_seconds": 60,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   "INVALID_ENTITY_TYPE",
							OperatorType: string(platform.IS),
							Values:       []string{"long_running_task"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("(?s).*entityType.*should be one of.*TASK_ID.*DAG_ID"),
			},
			// Validate: invalid pattern match operator type
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateTaskDurationAlertRequestTypeTASKDURATION),
					Severity:               string(platform.CreateTaskDurationAlertRequestSeverityINFO),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateTaskDurationAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":         deploymentId,
						"task_duration_seconds": 60,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.TASKID),
							OperatorType: "INVALID_OPERATOR_TYPE",
							Values:       []string{"long_running_task"},
						},
					},
				}),
				ExpectError: regexp.MustCompile("(?s).*operatorType.*should be one of"),
			},
			// Validate: empty pattern match values
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateTaskDurationAlertRequestTypeTASKDURATION),
					Severity:               string(platform.CreateTaskDurationAlertRequestSeverityINFO),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateTaskDurationAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":         deploymentId,
						"task_duration_seconds": 60,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.TASKID),
							OperatorType: string(platform.IS),
							Values:       []string{""},
						},
					},
				}),
				ExpectError: regexp.MustCompile("(?s).*values[0].*should be min: 1"),
			},
			// Create: TASK_DURATION alert
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateTaskDurationAlertRequestTypeTASKDURATION),
					Severity:               string(platform.CreateTaskDurationAlertRequestSeverityINFO),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateTaskDurationAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":         deploymentId,
						"task_duration_seconds": 60,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.TASKID),
							OperatorType: string(platform.IS),
							Values:       []string{"long_running_task"},
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceVar, "id"),
					resource.TestCheckResourceAttr(resourceVar, "name", alertName),
					resource.TestCheckResourceAttr(resourceVar, "type", string(platform.CreateTaskDurationAlertRequestTypeTASKDURATION)),
					resource.TestCheckResourceAttr(resourceVar, "severity", string(platform.CreateTaskDurationAlertRequestSeverityINFO)),
					resource.TestCheckResourceAttr(resourceVar, "rules.properties.deployment_id", deploymentId),
					resource.TestCheckResourceAttr(resourceVar, "rules.properties.task_duration_seconds", "60"),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.entity_type", string(platform.TASKID)),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.operator_type", string(platform.IS)),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.values.#", "1"),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.0.values.0", "long_running_task"),
					resource.TestCheckResourceAttrSet(resourceVar, "organization_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "workspace_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "deployment_id"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceVar, "created_by.id"),
					resource.TestCheckResourceAttrSet(resourceVar, "updated_by.id"),
					testAccCheckAlertExists(t, alertName),
				),
			},
			// Update: duration threshold
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateTaskDurationAlertRequestTypeTASKDURATION),
					Severity:               string(platform.CreateTaskDurationAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateTaskDurationAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":         deploymentId,
						"task_duration_seconds": 120,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.TASKID),
							OperatorType: string(platform.INCLUDES),
							Values:       []string{"process"},
						},
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.IS),
							Values:       []string{"etl_pipeline"},
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "severity", string(platform.CreateTaskDurationAlertRequestSeverityWARNING)),
					resource.TestCheckResourceAttr(resourceVar, "rules.properties.task_duration_seconds", "120"),
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.#", "2"),
					testAccCheckAlertExists(t, alertName),
				),
			},
			// Update: pattern matches with complex conditions
			{
				Config: astronomerprovider.ProviderConfig(t, astronomerprovider.HOSTED) + alert(alertInput{
					Name:                   alertName,
					Type:                   string(platform.CreateTaskDurationAlertRequestTypeTASKDURATION),
					Severity:               string(platform.CreateTaskDurationAlertRequestSeverityWARNING),
					EntityId:               deploymentId,
					EntityType:             string(platform.CreateTaskDurationAlertRequestEntityTypeDEPLOYMENT),
					NotificationChannelIds: []string{notificationChannelId},
					Properties: map[string]interface{}{
						"deployment_id":         deploymentId,
						"task_duration_seconds": 120,
					},
					PatternMatches: []patternMatch{
						{
							EntityType:   string(platform.TASKID),
							OperatorType: string(platform.INCLUDES),
							Values:       []string{"process", "transform"},
						},
						{
							EntityType:   string(platform.TASKID),
							OperatorType: string(platform.EXCLUDES),
							Values:       []string{"test", "debug"},
						},
						{
							EntityType:   string(platform.DAGID),
							OperatorType: string(platform.ISNOT),
							Values:       []string{"dev_pipeline"},
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceVar, "rules.pattern_matches.#", "3"),
					testAccCheckAlertExists(t, alertName),
				),
			},
			// Import: test import functionality
			{
				ResourceName:      resourceVar,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Helper types and functions

type patternMatch struct {
	EntityType   string
	OperatorType string
	Values       []string
}

type alertInput struct {
	Name                   string
	Type                   string
	Severity               string
	EntityId               string
	EntityType             string
	NotificationChannelIds []string
	Properties             map[string]interface{}
	PatternMatches         []patternMatch
}

func alert(input alertInput) string {
	// Build properties string
	propertiesStr := ""
	if input.Properties != nil {
		propertiesStr = "properties = {\n"
		for k, v := range input.Properties {
			switch val := v.(type) {
			case string:
				propertiesStr += fmt.Sprintf("\t\t\t%s = \"%s\"\n", k, val)
			case int:
				propertiesStr += fmt.Sprintf("\t\t\t%s = %d\n", k, val)
			case []string:
				propertiesStr += fmt.Sprintf("\t\t\t%s = [%s]\n", k, formatStringList(val))
			}
		}
		propertiesStr += "\t\t}"
	}

	// Build pattern matches string
	patternMatchesStr := ""
	if len(input.PatternMatches) > 0 {
		patternMatchesStr = "pattern_matches = [\n"
		for i, pm := range input.PatternMatches {
			patternMatchesStr += fmt.Sprintf(`		{
			entity_type = "%s"
			operator_type = "%s"
			values = [%s]
		}`, pm.EntityType, pm.OperatorType, formatStringList(pm.Values))
			if i < len(input.PatternMatches)-1 {
				patternMatchesStr += ","
			}
			patternMatchesStr += "\n"
		}
		patternMatchesStr += "\t\t]"
	}

	// Build notification channel IDs string
	notificationChannelIdsStr := formatStringList(input.NotificationChannelIds)

	return fmt.Sprintf(`
resource "astro_alert" "%s" {
	name = "%s"
	type = "%s"
	severity = "%s"
	entity_id = "%s"
	entity_type = "%s"
	notification_channel_ids = [%s]
	
	rules = {
		%s
		%s
	}
}`, input.Name, input.Name, input.Type, input.Severity, input.EntityId, input.EntityType,
		notificationChannelIdsStr, propertiesStr, patternMatchesStr)
}

func formatStringList(items []string) string {
	quoted := make([]string, len(items))
	for i, item := range items {
		quoted[i] = fmt.Sprintf("\"%s\"", item)
	}
	return strings.Join(quoted, ", ")
}

func testAccCheckAlertExists(t *testing.T, alertName string) func(s *terraform.State) error {
	t.Helper()
	return func(state *terraform.State) error {
		client, err := utils.GetTestPlatformClient(false)
		assert.NoError(t, err)

		organizationId := os.Getenv("HOSTED_ORGANIZATION_ID")
		ctx := context.Background()

		resp, err := client.ListAlertsWithResponse(ctx, organizationId, &platform.ListAlertsParams{})
		if err != nil {
			return fmt.Errorf("failed to list alerts: %v", err)
		}
		if resp == nil || resp.JSON200 == nil {
			return fmt.Errorf("nil response from list alerts")
		}

		for _, alert := range resp.JSON200.Alerts {
			if alert.Name == alertName {
				return nil
			}
		}

		return fmt.Errorf("alert %s not found", alertName)
	}
}

func testAccCheckAlertDestroyed(t *testing.T, alertName string) func(s *terraform.State) error {
	t.Helper()
	return func(state *terraform.State) error {
		client, err := utils.GetTestPlatformClient(false)
		assert.NoError(t, err)

		organizationId := os.Getenv("HOSTED_ORGANIZATION_ID")
		ctx := context.Background()

		resp, err := client.ListAlertsWithResponse(ctx, organizationId, &platform.ListAlertsParams{})
		if err != nil {
			return fmt.Errorf("failed to list alerts: %v", err)
		}
		if resp == nil || resp.JSON200 == nil {
			status, diag := clients.NormalizeAPIError(ctx, resp.HTTPResponse, resp.Body)
			return fmt.Errorf("response JSON200 is nil status: %v, err: %v", status, diag.Detail())
		}

		for _, alert := range resp.JSON200.Alerts {
			if alert.Name == alertName {
				return fmt.Errorf("alert %s still exists", alertName)
			}
		}

		return nil
	}
}
