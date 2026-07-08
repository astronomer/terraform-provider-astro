package resources

import (
	"context"
	"fmt"

	"github.com/astronomer/terraform-provider-astro/internal/clients/labs"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// labsPatternMatchRequests converts the decoded Terraform pattern matches into the labs API's
// PatternMatchRequest slice.
func labsPatternMatchRequests(input models.ResourceAlertRulesInput) []labs.PatternMatchRequest {
	pmReqs := make([]labs.PatternMatchRequest, len(input.PatternMatches))
	for i, pm := range input.PatternMatches {
		pmReqs[i] = labs.PatternMatchRequest{
			EntityType:   labs.PatternMatchRequestEntityType(pm.EntityType),
			OperatorType: labs.PatternMatchRequestOperatorType(pm.OperatorType),
			Values:       pm.Values,
		}
	}
	return pmReqs
}

// BuildLabsCreateAlertRequest builds a labs CreateAlertRequest body from an AlertResource model.
// The labs bulk create endpoint accepts a slice of these.
func BuildLabsCreateAlertRequest(ctx context.Context, data models.AlertResource) (labs.CreateAlertRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createAlertRequest labs.CreateAlertRequest

	notificationChannelIds, ncDiags := utils.TypesSetToStringSlice(ctx, data.NotificationChannelIds)
	if ncDiags.HasError() {
		return createAlertRequest, ncDiags
	}

	var alertRulesInput models.ResourceAlertRulesInput
	if rulesDiags := data.Rules.As(ctx, &alertRulesInput, basetypes.ObjectAsOptions{}); rulesDiags.HasError() {
		return createAlertRequest, rulesDiags
	}
	pmReqs := labsPatternMatchRequests(alertRulesInput)

	switch data.Type.ValueString() {
	case string(labs.AlertTypeDAGFAILURE):
		req := labs.CreateDagFailureAlertRequest{
			EntityId:               data.EntityId.ValueString(),
			EntityType:             labs.CreateDagFailureAlertRequestEntityType(data.EntityType.ValueString()),
			Name:                   data.Name.ValueString(),
			NotificationChannelIds: notificationChannelIds,
			Severity:               labs.CreateDagFailureAlertRequestSeverity(data.Severity.ValueString()),
			Type:                   labs.CreateDagFailureAlertRequestType(data.Type.ValueString()),
			Rules: labs.CreateDagFailureAlertRules{
				PatternMatches: pmReqs,
				Properties: labs.CreateDagFailureAlertProperties{
					DeploymentId: alertRulesInput.Properties.DeploymentId.ValueString(),
				},
			},
		}
		if err := createAlertRequest.FromCreateDagFailureAlertRequest(req); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build DAG_FAILURE request: %s", err))
			return createAlertRequest, diags
		}

	case string(labs.AlertTypeDAGSUCCESS):
		req := labs.CreateDagSuccessAlertRequest{
			EntityId:               data.EntityId.ValueString(),
			EntityType:             labs.CreateDagSuccessAlertRequestEntityType(data.EntityType.ValueString()),
			Name:                   data.Name.ValueString(),
			NotificationChannelIds: notificationChannelIds,
			Severity:               labs.CreateDagSuccessAlertRequestSeverity(data.Severity.ValueString()),
			Type:                   labs.CreateDagSuccessAlertRequestType(data.Type.ValueString()),
			Rules: labs.CreateDagSuccessAlertRules{
				PatternMatches: pmReqs,
				Properties: labs.CreateDagSuccessAlertProperties{
					DeploymentId: alertRulesInput.Properties.DeploymentId.ValueString(),
				},
			},
		}
		if err := createAlertRequest.FromCreateDagSuccessAlertRequest(req); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build DAG_SUCCESS request: %s", err))
			return createAlertRequest, diags
		}

	case string(labs.AlertTypeDAGDURATION):
		req := labs.CreateDagDurationAlertRequest{
			EntityId:               data.EntityId.ValueString(),
			EntityType:             labs.CreateDagDurationAlertRequestEntityType(data.EntityType.ValueString()),
			Name:                   data.Name.ValueString(),
			NotificationChannelIds: notificationChannelIds,
			Severity:               labs.CreateDagDurationAlertRequestSeverity(data.Severity.ValueString()),
			Type:                   labs.CreateDagDurationAlertRequestType(data.Type.ValueString()),
			Rules: labs.CreateDagDurationAlertRules{
				PatternMatches: pmReqs,
				Properties: labs.CreateDagDurationAlertProperties{
					DeploymentId:       alertRulesInput.Properties.DeploymentId.ValueString(),
					DagDurationSeconds: int(alertRulesInput.Properties.DagDurationSeconds.ValueInt64()),
				},
			},
		}
		if err := createAlertRequest.FromCreateDagDurationAlertRequest(req); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build DAG_DURATION request: %s", err))
			return createAlertRequest, diags
		}

	case string(labs.AlertTypeDAGTIMELINESS):
		var days []string
		if errList := alertRulesInput.Properties.DaysOfWeek.ElementsAs(ctx, &days, false); errList.HasError() {
			return createAlertRequest, errList
		}
		req := labs.CreateDagTimelinessAlertRequest{
			EntityId:               data.EntityId.ValueString(),
			EntityType:             labs.CreateDagTimelinessAlertRequestEntityType(data.EntityType.ValueString()),
			Name:                   data.Name.ValueString(),
			NotificationChannelIds: notificationChannelIds,
			Severity:               labs.CreateDagTimelinessAlertRequestSeverity(data.Severity.ValueString()),
			Type:                   labs.CreateDagTimelinessAlertRequestType(data.Type.ValueString()),
			Rules: labs.CreateDagTimelinessAlertRules{
				PatternMatches: pmReqs,
				Properties: labs.CreateDagTimelinessAlertProperties{
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

	case string(labs.AlertTypeTASKFAILURE):
		req := labs.CreateTaskFailureAlertRequest{
			EntityId:               data.EntityId.ValueString(),
			EntityType:             labs.CreateTaskFailureAlertRequestEntityType(data.EntityType.ValueString()),
			Name:                   data.Name.ValueString(),
			NotificationChannelIds: notificationChannelIds,
			Severity:               labs.CreateTaskFailureAlertRequestSeverity(data.Severity.ValueString()),
			Type:                   labs.CreateTaskFailureAlertRequestType(data.Type.ValueString()),
			Rules: labs.CreateTaskFailureAlertRules{
				PatternMatches: pmReqs,
				Properties: labs.CreateTaskFailureAlertProperties{
					DeploymentId: alertRulesInput.Properties.DeploymentId.ValueString(),
				},
			},
		}
		if err := createAlertRequest.FromCreateTaskFailureAlertRequest(req); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build TASK_FAILURE request: %s", err))
			return createAlertRequest, diags
		}

	case string(labs.AlertTypeTASKDURATION):
		req := labs.CreateTaskDurationAlertRequest{
			EntityId:               data.EntityId.ValueString(),
			EntityType:             labs.CreateTaskDurationAlertRequestEntityType(data.EntityType.ValueString()),
			Name:                   data.Name.ValueString(),
			NotificationChannelIds: notificationChannelIds,
			Severity:               labs.CreateTaskDurationAlertRequestSeverity(data.Severity.ValueString()),
			Type:                   labs.CreateTaskDurationAlertRequestType(data.Type.ValueString()),
			Rules: labs.CreateTaskDurationAlertRules{
				PatternMatches: pmReqs,
				Properties: labs.CreateTaskDurationAlertProperties{
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

// BuildLabsUpdateAlertRequest builds a labs UpdateAlertRequest body from an AlertResource model.
// data.Id must be set — the labs bulk update endpoint identifies each alert by the id in the body.
func BuildLabsUpdateAlertRequest(ctx context.Context, data models.AlertResource) (labs.UpdateAlertRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updateBody labs.UpdateAlertRequest

	ncIds, ncDiags := utils.TypesSetToStringSlice(ctx, data.NotificationChannelIds)
	if ncDiags.HasError() {
		return updateBody, ncDiags
	}

	var alertRulesInput models.ResourceAlertRulesInput
	if rulesDiags := data.Rules.As(ctx, &alertRulesInput, basetypes.ObjectAsOptions{}); rulesDiags.HasError() {
		return updateBody, rulesDiags
	}
	pmReqs := labsPatternMatchRequests(alertRulesInput)

	name := data.Name.ValueString()
	var idPtr *string
	if id := data.Id.ValueString(); id != "" {
		idPtr = &id
	}

	switch data.Type.ValueString() {
	case string(labs.AlertTypeDAGFAILURE):
		sev := labs.UpdateDagFailureAlertRequestSeverity(data.Severity.ValueString())
		alertType := labs.UpdateDagFailureAlertRequestType(data.Type.ValueString())
		reqModel := labs.UpdateDagFailureAlertRequest{
			Id:                     idPtr,
			Name:                   &name,
			Severity:               &sev,
			Type:                   &alertType,
			NotificationChannelIds: &ncIds,
			Rules: &labs.UpdateDagFailureAlertRules{
				PatternMatches: &pmReqs,
			},
		}
		if err := updateBody.FromUpdateDagFailureAlertRequest(reqModel); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build update for DAG_FAILURE: %s", err))
			return updateBody, diags
		}

	case string(labs.AlertTypeDAGSUCCESS):
		sev := labs.UpdateDagSuccessAlertRequestSeverity(data.Severity.ValueString())
		alertType := labs.UpdateDagSuccessAlertRequestType(data.Type.ValueString())
		reqModel := labs.UpdateDagSuccessAlertRequest{
			Id:                     idPtr,
			Name:                   &name,
			NotificationChannelIds: &ncIds,
			Severity:               &sev,
			Type:                   &alertType,
			Rules: &labs.UpdateDagSuccessAlertRules{
				PatternMatches: &pmReqs,
			},
		}
		if err := updateBody.FromUpdateDagSuccessAlertRequest(reqModel); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build update for DAG_SUCCESS: %s", err))
			return updateBody, diags
		}

	case string(labs.AlertTypeDAGDURATION):
		sev := labs.UpdateDagDurationAlertRequestSeverity(data.Severity.ValueString())
		alertType := labs.UpdateDagDurationAlertRequestType(data.Type.ValueString())
		dagDurationSeconds := int(alertRulesInput.Properties.DagDurationSeconds.ValueInt64())
		reqModel := labs.UpdateDagDurationAlertRequest{
			Id:                     idPtr,
			Name:                   &name,
			NotificationChannelIds: &ncIds,
			Severity:               &sev,
			Type:                   &alertType,
			Rules: &labs.UpdateDagDurationAlertRules{
				PatternMatches: &pmReqs,
				Properties: &labs.UpdateDagDurationAlertProperties{
					DagDurationSeconds: &dagDurationSeconds,
				},
			},
		}
		if err := updateBody.FromUpdateDagDurationAlertRequest(reqModel); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build update for DAG_DURATION: %s", err))
			return updateBody, diags
		}

	case string(labs.AlertTypeDAGTIMELINESS):
		sev := labs.UpdateDagTimelinessAlertRequestSeverity(data.Severity.ValueString())
		alertType := labs.UpdateDagTimelinessAlertRequestType(data.Type.ValueString())
		dagDeadline := alertRulesInput.Properties.DagDeadline.ValueString()
		var days []string
		if errList := alertRulesInput.Properties.DaysOfWeek.ElementsAs(ctx, &days, false); errList.HasError() {
			return updateBody, errList
		}
		lookBackPeriodSeconds := int(alertRulesInput.Properties.LookBackPeriodSeconds.ValueInt64())
		reqModel := labs.UpdateDagTimelinessAlertRequest{
			Id:                     idPtr,
			Name:                   &name,
			NotificationChannelIds: &ncIds,
			Severity:               &sev,
			Type:                   &alertType,
			Rules: &labs.UpdateDagTimelinessAlertRules{
				PatternMatches: &pmReqs,
				Properties: &labs.UpdateDagTimelinessAlertProperties{
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

	case string(labs.AlertTypeTASKFAILURE):
		sev := labs.UpdateTaskFailureAlertRequestSeverity(data.Severity.ValueString())
		alertType := labs.UpdateTaskFailureAlertRequestType(data.Type.ValueString())
		reqModel := labs.UpdateTaskFailureAlertRequest{
			Id:                     idPtr,
			Name:                   &name,
			NotificationChannelIds: &ncIds,
			Severity:               &sev,
			Type:                   &alertType,
			Rules: &labs.UpdateTaskFailureAlertRules{
				PatternMatches: &pmReqs,
			},
		}
		if err := updateBody.FromUpdateTaskFailureAlertRequest(reqModel); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build update for TASK_FAILURE: %s", err))
			return updateBody, diags
		}

	case string(labs.AlertTypeTASKDURATION):
		sev := labs.UpdateTaskDurationAlertRequestSeverity(data.Severity.ValueString())
		alertType := labs.UpdateTaskDurationAlertRequestType(data.Type.ValueString())
		taskDurationSeconds := int(alertRulesInput.Properties.TaskDurationSeconds.ValueInt64())
		reqModel := labs.UpdateTaskDurationAlertRequest{
			Id:                     idPtr,
			Name:                   &name,
			NotificationChannelIds: &ncIds,
			Severity:               &sev,
			Type:                   &alertType,
			Rules: &labs.UpdateTaskDurationAlertRules{
				PatternMatches: &pmReqs,
				Properties: &labs.UpdateTaskDurationAlertProperties{
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
