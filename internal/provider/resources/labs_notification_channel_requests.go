package resources

import (
	"context"
	"fmt"

	"github.com/astronomer/terraform-provider-astro/internal/clients/labs"
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/models"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// BuildLabsCreateNotificationChannelRequest builds a labs CreateNotificationChannelRequest body
// from a bulk resource element. The labs bulk create endpoint accepts a slice of these.
func BuildLabsCreateNotificationChannelRequest(
	ctx context.Context,
	elem models.NotificationChannelsResourceElementModel,
) (labs.CreateNotificationChannelRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	var createRequest labs.CreateNotificationChannelRequest

	var definition models.NotificationChannelDefinition
	if errList := elem.Definition.As(ctx, &definition, basetypes.ObjectAsOptions{}); errList.HasError() {
		return createRequest, errList
	}

	channelType := elem.Type.ValueString()
	entityId := elem.EntityId.ValueString()
	isShared := elem.IsShared.ValueBoolPointer()
	name := elem.Name.ValueString()

	switch channelType {
	case string(labs.EMAIL):
		recipients, ncDiags := utils.TypesSetToStringSlice(ctx, definition.Recipients)
		if ncDiags.HasError() {
			return createRequest, ncDiags
		}
		req := labs.CreateEmailNotificationChannelRequest{
			Name:       name,
			Definition: labs.EmailNotificationChannelDefinition{Recipients: recipients},
			Type:       labs.CreateEmailNotificationChannelRequestType(channelType),
			EntityId:   entityId,
			EntityType: labs.CreateEmailNotificationChannelRequestEntityType(elem.EntityType.ValueString()),
			IsShared:   isShared,
		}
		if err := createRequest.FromCreateEmailNotificationChannelRequest(req); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build EMAIL notification channel request: %s", err))
			return createRequest, diags
		}

	case string(labs.SLACK):
		req := labs.CreateSlackNotificationChannelRequest{
			Name:       name,
			Definition: labs.SlackNotificationChannelDefinition{WebhookUrl: definition.WebhookUrl.ValueString()},
			Type:       labs.CreateSlackNotificationChannelRequestType(channelType),
			EntityId:   entityId,
			EntityType: labs.CreateSlackNotificationChannelRequestEntityType(elem.EntityType.ValueString()),
			IsShared:   isShared,
		}
		if err := createRequest.FromCreateSlackNotificationChannelRequest(req); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build SLACK notification channel request: %s", err))
			return createRequest, diags
		}

	case string(labs.DAGTRIGGER):
		req := labs.CreateDagTriggerNotificationChannelRequest{
			Name: name,
			Definition: labs.DagTriggerNotificationChannelDefinition{
				DagId:              definition.DagId.ValueString(),
				DeploymentApiToken: definition.DeploymentApiToken.ValueString(),
				DeploymentId:       definition.DeploymentId.ValueString(),
			},
			Type:       labs.CreateDagTriggerNotificationChannelRequestType(channelType),
			EntityId:   entityId,
			EntityType: labs.CreateDagTriggerNotificationChannelRequestEntityType(elem.EntityType.ValueString()),
			IsShared:   isShared,
		}
		if err := createRequest.FromCreateDagTriggerNotificationChannelRequest(req); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build DAG_TRIGGER notification channel request: %s", err))
			return createRequest, diags
		}

	case string(labs.PAGERDUTY):
		req := labs.CreatePagerDutyNotificationChannelRequest{
			Name:       name,
			Definition: labs.PagerDutyNotificationChannelDefinition{IntegrationKey: definition.IntegrationKey.ValueString()},
			Type:       labs.CreatePagerDutyNotificationChannelRequestType(channelType),
			EntityId:   entityId,
			EntityType: labs.CreatePagerDutyNotificationChannelRequestEntityType(elem.EntityType.ValueString()),
			IsShared:   isShared,
		}
		if err := createRequest.FromCreatePagerDutyNotificationChannelRequest(req); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build PAGERDUTY notification channel request: %s", err))
			return createRequest, diags
		}

	case string(labs.OPSGENIE):
		req := labs.CreateOpsgenieNotificationChannelRequest{
			Name:       name,
			Definition: labs.OpsgenieNotificationChannelDefinition{ApiKey: definition.ApiKey.ValueString()},
			Type:       labs.CreateOpsgenieNotificationChannelRequestType(channelType),
			EntityId:   entityId,
			EntityType: labs.CreateOpsgenieNotificationChannelRequestEntityType(elem.EntityType.ValueString()),
			IsShared:   isShared,
		}
		if err := createRequest.FromCreateOpsgenieNotificationChannelRequest(req); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build OPSGENIE notification channel request: %s", err))
			return createRequest, diags
		}

	default:
		diags.AddError("Invalid notification channel type", fmt.Sprintf("Unsupported notification channel type: %s", channelType))
		return createRequest, diags
	}

	return createRequest, diags
}

// BuildPlatformUpdateNotificationChannelBody builds a platform UpdateNotificationChannel body from a
// bulk resource element. Labs has no bulk-update endpoint for notification channels (unlike alerts),
// so the bulk resource updates each changed channel through the single-channel platform endpoint.
// Tracked in CPP-968.
func BuildPlatformUpdateNotificationChannelBody(
	ctx context.Context,
	elem models.NotificationChannelsResourceElementModel,
) (platform.UpdateNotificationChannelJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics
	var updateBody platform.UpdateNotificationChannelJSONRequestBody

	var definition models.NotificationChannelDefinition
	if errList := elem.Definition.As(ctx, &definition, basetypes.ObjectAsOptions{}); errList.HasError() {
		return updateBody, errList
	}

	name := elem.Name.ValueString()
	isShared := elem.IsShared.ValueBoolPointer()
	channelType := elem.Type.ValueString()

	switch channelType {
	case string(platform.AlertNotificationChannelTypeEMAIL):
		recipients, ncDiags := utils.TypesSetToStringSlice(ctx, definition.Recipients)
		if ncDiags.HasError() {
			return updateBody, ncDiags
		}
		t := platform.UpdateEmailNotificationChannelRequestType(channelType)
		req := platform.UpdateEmailNotificationChannelRequest{
			Name:       &name,
			Definition: &platform.EmailNotificationChannelDefinition{Recipients: recipients},
			Type:       &t,
			IsShared:   isShared,
		}
		if err := updateBody.FromUpdateEmailNotificationChannelRequest(req); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build EMAIL notification channel update request: %s", err))
			return updateBody, diags
		}

	case string(platform.AlertNotificationChannelTypeSLACK):
		t := platform.UpdateSlackNotificationChannelRequestType(channelType)
		req := platform.UpdateSlackNotificationChannelRequest{
			Name:       &name,
			Definition: &platform.SlackNotificationChannelDefinition{WebhookUrl: definition.WebhookUrl.ValueString()},
			Type:       &t,
			IsShared:   isShared,
		}
		if err := updateBody.FromUpdateSlackNotificationChannelRequest(req); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build SLACK notification channel update request: %s", err))
			return updateBody, diags
		}

	case string(platform.AlertNotificationChannelTypeDAGTRIGGER):
		t := platform.UpdateDagTriggerNotificationChannelRequestType(channelType)
		req := platform.UpdateDagTriggerNotificationChannelRequest{
			Name: &name,
			Definition: &platform.DagTriggerNotificationChannelDefinition{
				DagId:              definition.DagId.ValueString(),
				DeploymentApiToken: definition.DeploymentApiToken.ValueString(),
				DeploymentId:       definition.DeploymentId.ValueString(),
			},
			Type:     &t,
			IsShared: isShared,
		}
		if err := updateBody.FromUpdateDagTriggerNotificationChannelRequest(req); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build DAG_TRIGGER notification channel update request: %s", err))
			return updateBody, diags
		}

	case string(platform.AlertNotificationChannelTypePAGERDUTY):
		t := platform.UpdatePagerDutyNotificationChannelRequestType(channelType)
		req := platform.UpdatePagerDutyNotificationChannelRequest{
			Name:       &name,
			Definition: &platform.PagerDutyNotificationChannelDefinition{IntegrationKey: definition.IntegrationKey.ValueString()},
			Type:       &t,
			IsShared:   isShared,
		}
		if err := updateBody.FromUpdatePagerDutyNotificationChannelRequest(req); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build PAGERDUTY notification channel update request: %s", err))
			return updateBody, diags
		}

	case string(platform.AlertNotificationChannelTypeOPSGENIE):
		t := platform.UpdateOpsgenieNotificationChannelRequestType(channelType)
		req := platform.UpdateOpsgenieNotificationChannelRequest{
			Name:       &name,
			Definition: &platform.OpsgenieNotificationChannelDefinition{ApiKey: definition.ApiKey.ValueString()},
			Type:       &t,
			IsShared:   isShared,
		}
		if err := updateBody.FromUpdateOpsgenieNotificationChannelRequest(req); err != nil {
			diags.AddError("Internal Error", fmt.Sprintf("failed to build OPSGENIE notification channel update request: %s", err))
			return updateBody, diags
		}

	default:
		diags.AddError("Invalid notification channel type", fmt.Sprintf("Unsupported notification channel type: %s", channelType))
		return updateBody, diags
	}

	return updateBody, diags
}
