package resources

import (
	"context"
	"fmt"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// patternMatchRequests converts the decoded Terraform pattern matches into the API's PatternMatchRequest slice.
func patternMatchRequests(input models.ResourceAlertRulesInput) []platform.PatternMatchRequest {
	pmReqs := make([]platform.PatternMatchRequest, len(input.PatternMatches))
	for i, pm := range input.PatternMatches {
		pmReqs[i] = platform.PatternMatchRequest{
			EntityType:   platform.PatternMatchRequestEntityType(pm.EntityType),
			OperatorType: platform.PatternMatchRequestOperatorType(pm.OperatorType),
			Values:       pm.Values,
		}
	}
	return pmReqs
}

// BuildCreateAlertRequest builds a platform CreateAlert request body from an AlertResource model.
// It is shared by the singular astro_alert resource and the bulk astro_alerts resource.
func BuildCreateAlertRequest(ctx context.Context, data models.AlertResource) (platform.CreateAlertJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createAlertRequest platform.CreateAlertJSONRequestBody

	notificationChannelIds, ncDiags := utils.TypesSetToStringSlice(ctx, data.NotificationChannelIds)
	if ncDiags.HasError() {
		return createAlertRequest, ncDiags
	}

	var alertRulesInput models.ResourceAlertRulesInput
	if rulesDiags := data.Rules.As(ctx, &alertRulesInput, basetypes.ObjectAsOptions{}); rulesDiags.HasError() {
		return createAlertRequest, rulesDiags
	}
	pmReqs := patternMatchRequests(alertRulesInput)

	switch data.Type.ValueString() {
	case string(platform.AlertTypeDAGFAILURE):
		req := platform.CreateDagFailureAlertRequest{
			EntityId:               data.EntityId.ValueString(),
			EntityType:             platform.CreateDagFailureAlertRequestEntityType(data.EntityType.ValueString()),
			Name:                   data.Name.ValueString(),
			NotificationChannelIds: notificationChannelIds,
			Severity:               platform.CreateDagFailureAlertRequestSeverity(data.Severity.ValueString()),
			Type:                   platform.CreateDagFailureAlertRequestType(data.Type.ValueString()),
			Rules: platform.CreateDagFailureAlertRules{
				PatternMatches: pmReqs,
				Properties: platform.CreateDagFailureAlertProperties{
					DeploymentId: alertRulesInput.Properties.DeploymentId.ValueString(),
				},
			},
		}
		if err := createAlertRequest.FromCreateDagFailureAlertRequest(req); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build DAG_FAILURE request: %s", err))
			return createAlertRequest, diags
		}

	case string(platform.AlertTypeDAGSUCCESS):
		req := platform.CreateDagSuccessAlertRequest{
			EntityId:               data.EntityId.ValueString(),
			EntityType:             platform.CreateDagSuccessAlertRequestEntityType(data.EntityType.ValueString()),
			Name:                   data.Name.ValueString(),
			NotificationChannelIds: notificationChannelIds,
			Severity:               platform.CreateDagSuccessAlertRequestSeverity(data.Severity.ValueString()),
			Type:                   platform.CreateDagSuccessAlertRequestType(data.Type.ValueString()),
			Rules: platform.CreateDagSuccessAlertRules{
				PatternMatches: pmReqs,
				Properties: platform.CreateDagSuccessAlertProperties{
					DeploymentId: alertRulesInput.Properties.DeploymentId.ValueString(),
				},
			},
		}
		if err := createAlertRequest.FromCreateDagSuccessAlertRequest(req); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build DAG_SUCCESS request: %s", err))
			return createAlertRequest, diags
		}

	case string(platform.AlertTypeDAGDURATION):
		req := platform.CreateDagDurationAlertRequest{
			EntityId:               data.EntityId.ValueString(),
			EntityType:             platform.CreateDagDurationAlertRequestEntityType(data.EntityType.ValueString()),
			Name:                   data.Name.ValueString(),
			NotificationChannelIds: notificationChannelIds,
			Severity:               platform.CreateDagDurationAlertRequestSeverity(data.Severity.ValueString()),
			Type:                   platform.CreateDagDurationAlertRequestType(data.Type.ValueString()),
			Rules: platform.CreateDagDurationAlertRules{
				PatternMatches: pmReqs,
				Properties: platform.CreateDagDurationAlertProperties{
					DeploymentId:       alertRulesInput.Properties.DeploymentId.ValueString(),
					DagDurationSeconds: int(alertRulesInput.Properties.DagDurationSeconds.ValueInt64()),
				},
			},
		}
		if err := createAlertRequest.FromCreateDagDurationAlertRequest(req); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build DAG_DURATION request: %s", err))
			return createAlertRequest, diags
		}

	case string(platform.AlertTypeDAGTIMELINESS):
		var days []string
		if errList := alertRulesInput.Properties.DaysOfWeek.ElementsAs(ctx, &days, false); errList.HasError() {
			return createAlertRequest, errList
		}
		req := platform.CreateDagTimelinessAlertRequest{
			EntityId:               data.EntityId.ValueString(),
			EntityType:             platform.CreateDagTimelinessAlertRequestEntityType(data.EntityType.ValueString()),
			Name:                   data.Name.ValueString(),
			NotificationChannelIds: notificationChannelIds,
			Severity:               platform.CreateDagTimelinessAlertRequestSeverity(data.Severity.ValueString()),
			Type:                   platform.CreateDagTimelinessAlertRequestType(data.Type.ValueString()),
			Rules: platform.CreateDagTimelinessAlertRules{
				PatternMatches: pmReqs,
				Properties: platform.CreateDagTimelinessAlertProperties{
					DeploymentId:          alertRulesInput.Properties.DeploymentId.ValueString(),
					DagDeadline:           alertRulesInput.Properties.DagDeadline.ValueString(),
					DaysOfWeek:            days,
					LookBackPeriodSeconds: int(alertRulesInput.Properties.LookBackPeriodSeconds.ValueInt64()),
				},
			},
		}
		if err := createAlertRequest.FromCreateDagTimelinessAlertRequest(req); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build DAG_TIMELINESS request: %s", err))
			return createAlertRequest, diags
		}

	case string(platform.AlertTypeTASKFAILURE):
		req := platform.CreateTaskFailureAlertRequest{
			EntityId:               data.EntityId.ValueString(),
			EntityType:             platform.CreateTaskFailureAlertRequestEntityType(data.EntityType.ValueString()),
			Name:                   data.Name.ValueString(),
			NotificationChannelIds: notificationChannelIds,
			Severity:               platform.CreateTaskFailureAlertRequestSeverity(data.Severity.ValueString()),
			Type:                   platform.CreateTaskFailureAlertRequestType(data.Type.ValueString()),
			Rules: platform.CreateTaskFailureAlertRules{
				PatternMatches: pmReqs,
				Properties: platform.CreateTaskFailureAlertProperties{
					DeploymentId: alertRulesInput.Properties.DeploymentId.ValueString(),
				},
			},
		}
		if err := createAlertRequest.FromCreateTaskFailureAlertRequest(req); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build TASK_FAILURE request: %s", err))
			return createAlertRequest, diags
		}

	case string(platform.AlertTypeTASKDURATION):
		req := platform.CreateTaskDurationAlertRequest{
			EntityId:               data.EntityId.ValueString(),
			EntityType:             platform.CreateTaskDurationAlertRequestEntityType(data.EntityType.ValueString()),
			Name:                   data.Name.ValueString(),
			NotificationChannelIds: notificationChannelIds,
			Severity:               platform.CreateTaskDurationAlertRequestSeverity(data.Severity.ValueString()),
			Type:                   platform.CreateTaskDurationAlertRequestType(data.Type.ValueString()),
			Rules: platform.CreateTaskDurationAlertRules{
				PatternMatches: pmReqs,
				Properties: platform.CreateTaskDurationAlertProperties{
					DeploymentId:        alertRulesInput.Properties.DeploymentId.ValueString(),
					TaskDurationSeconds: int(alertRulesInput.Properties.TaskDurationSeconds.ValueInt64()),
				},
			},
		}
		if err := createAlertRequest.FromCreateTaskDurationAlertRequest(req); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build TASK_DURATION request: %s", err))
			return createAlertRequest, diags
		}

	default:
		diags.AddError("Invalid alert type", fmt.Sprintf("Unsupported alert type: %s", data.Type.ValueString()))
		return createAlertRequest, diags
	}

	return createAlertRequest, diags
}

// BuildUpdateAlertRequest builds a platform UpdateAlert request body from an AlertResource model.
// When data.Id is set, it is included in the request body — required by the bulk update endpoint and
// ignored by the single-alert update endpoint (which takes the ID from the path).
func BuildUpdateAlertRequest(ctx context.Context, data models.AlertResource) (platform.UpdateAlertJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updateBody platform.UpdateAlertJSONRequestBody

	ncIds, ncDiags := utils.TypesSetToStringSlice(ctx, data.NotificationChannelIds)
	if ncDiags.HasError() {
		return updateBody, ncDiags
	}

	var alertRulesInput models.ResourceAlertRulesInput
	if rulesDiags := data.Rules.As(ctx, &alertRulesInput, basetypes.ObjectAsOptions{}); rulesDiags.HasError() {
		return updateBody, rulesDiags
	}
	pmReqs := patternMatchRequests(alertRulesInput)

	name := data.Name.ValueString()

	switch data.Type.ValueString() {
	case string(platform.AlertTypeDAGFAILURE):
		sev := platform.UpdateDagFailureAlertRequestSeverity(data.Severity.ValueString())
		alertType := platform.UpdateDagFailureAlertRequestType(data.Type.ValueString())
		reqModel := platform.UpdateDagFailureAlertRequest{
			Name:                   &name,
			Severity:               &sev,
			Type:                   &alertType,
			NotificationChannelIds: &ncIds,
			Rules: &platform.UpdateDagFailureAlertRules{
				PatternMatches: &pmReqs,
			},
		}
		if err := updateBody.FromUpdateDagFailureAlertRequest(reqModel); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build update for DAG_FAILURE: %s", err))
			return updateBody, diags
		}

	case string(platform.AlertTypeDAGSUCCESS):
		sev := platform.UpdateDagSuccessAlertRequestSeverity(data.Severity.ValueString())
		alertType := platform.UpdateDagSuccessAlertRequestType(data.Type.ValueString())
		reqModel := platform.UpdateDagSuccessAlertRequest{
			Name:                   &name,
			NotificationChannelIds: &ncIds,
			Severity:               &sev,
			Type:                   &alertType,
			Rules: &platform.UpdateDagSuccessAlertRules{
				PatternMatches: &pmReqs,
			},
		}
		if err := updateBody.FromUpdateDagSuccessAlertRequest(reqModel); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build update for DAG_SUCCESS: %s", err))
			return updateBody, diags
		}

	case string(platform.AlertTypeDAGDURATION):
		sev := platform.UpdateDagDurationAlertRequestSeverity(data.Severity.ValueString())
		alertType := platform.UpdateDagDurationAlertRequestType(data.Type.ValueString())
		dagDurationSeconds := int(alertRulesInput.Properties.DagDurationSeconds.ValueInt64())
		reqModel := platform.UpdateDagDurationAlertRequest{
			Name:                   &name,
			NotificationChannelIds: &ncIds,
			Severity:               &sev,
			Type:                   &alertType,
			Rules: &platform.UpdateDagDurationAlertRules{
				PatternMatches: &pmReqs,
				Properties: &platform.UpdateDagDurationAlertProperties{
					DagDurationSeconds: &dagDurationSeconds,
				},
			},
		}
		if err := updateBody.FromUpdateDagDurationAlertRequest(reqModel); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build update for DAG_DURATION: %s", err))
			return updateBody, diags
		}

	case string(platform.AlertTypeDAGTIMELINESS):
		sev := platform.UpdateDagTimelinessAlertRequestSeverity(data.Severity.ValueString())
		alertType := platform.UpdateDagTimelinessAlertRequestType(data.Type.ValueString())
		dagDeadline := alertRulesInput.Properties.DagDeadline.ValueString()
		var days []string
		if errList := alertRulesInput.Properties.DaysOfWeek.ElementsAs(ctx, &days, false); errList.HasError() {
			return updateBody, errList
		}
		lookBackPeriodSeconds := int(alertRulesInput.Properties.LookBackPeriodSeconds.ValueInt64())
		reqModel := platform.UpdateDagTimelinessAlertRequest{
			Name:                   &name,
			NotificationChannelIds: &ncIds,
			Severity:               &sev,
			Type:                   &alertType,
			Rules: &platform.UpdateDagTimelinessAlertRules{
				PatternMatches: &pmReqs,
				Properties: &platform.UpdateDagTimelinessAlertProperties{
					DagDeadline:           &dagDeadline,
					DaysOfWeek:            &days,
					LookBackPeriodSeconds: &lookBackPeriodSeconds,
				},
			},
		}
		if err := updateBody.FromUpdateDagTimelinessAlertRequest(reqModel); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build update for DAG_TIMELINESS: %s", err))
			return updateBody, diags
		}

	case string(platform.AlertTypeTASKFAILURE):
		sev := platform.UpdateTaskFailureAlertRequestSeverity(data.Severity.ValueString())
		alertType := platform.UpdateTaskFailureAlertRequestType(data.Type.ValueString())
		reqModel := platform.UpdateTaskFailureAlertRequest{
			Name:                   &name,
			NotificationChannelIds: &ncIds,
			Severity:               &sev,
			Type:                   &alertType,
			Rules: &platform.UpdateTaskFailureAlertRules{
				PatternMatches: &pmReqs,
			},
		}
		if err := updateBody.FromUpdateTaskFailureAlertRequest(reqModel); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build update for TASK_FAILURE: %s", err))
			return updateBody, diags
		}

	case string(platform.AlertTypeTASKDURATION):
		sev := platform.UpdateTaskDurationAlertRequestSeverity(data.Severity.ValueString())
		alertType := platform.UpdateTaskDurationAlertRequestType(data.Type.ValueString())
		taskDurationSeconds := int(alertRulesInput.Properties.TaskDurationSeconds.ValueInt64())
		reqModel := platform.UpdateTaskDurationAlertRequest{
			Name:                   &name,
			NotificationChannelIds: &ncIds,
			Severity:               &sev,
			Type:                   &alertType,
			Rules: &platform.UpdateTaskDurationAlertRules{
				PatternMatches: &pmReqs,
				Properties: &platform.UpdateTaskDurationAlertProperties{
					TaskDurationSeconds: &taskDurationSeconds,
				},
			},
		}
		if err := updateBody.FromUpdateTaskDurationAlertRequest(reqModel); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build update for TASK_DURATION: %s", err))
			return updateBody, diags
		}

	default:
		diags.AddError("Invalid alert type", fmt.Sprintf("Unsupported alert type: %s", data.Type.ValueString()))
		return updateBody, diags
	}

	return updateBody, diags
}
