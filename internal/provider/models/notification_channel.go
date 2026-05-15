package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// NotificationChannelDataSource describes the data source data model.
type NotificationChannelDataSource struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Definition   types.Object `tfsdk:"definition"`
	Type         types.String `tfsdk:"type"`
	WorkspaceId  types.String `tfsdk:"workspace_id"`
	DeploymentId types.String `tfsdk:"deployment_id"`
	EntityId     types.String `tfsdk:"entity_id"`
	EntityType   types.String `tfsdk:"entity_type"`
	EntityName   types.String `tfsdk:"entity_name"`
	IsShared     types.Bool   `tfsdk:"is_shared"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
	CreatedBy    types.Object `tfsdk:"created_by"`
	UpdatedBy    types.Object `tfsdk:"updated_by"`
}

// NotificationChannelResource describes the resource data model.
type NotificationChannelResource struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Definition   types.Object `tfsdk:"definition"`
	Type         types.String `tfsdk:"type"`
	WorkspaceId  types.String `tfsdk:"workspace_id"`
	DeploymentId types.String `tfsdk:"deployment_id"`
	EntityId     types.String `tfsdk:"entity_id"`
	EntityType   types.String `tfsdk:"entity_type"`
	EntityName   types.String `tfsdk:"entity_name"`
	IsShared     types.Bool   `tfsdk:"is_shared"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
	CreatedBy    types.Object `tfsdk:"created_by"`
	UpdatedBy    types.Object `tfsdk:"updated_by"`
}

type NotificationChannelDefinition struct {
	DagId              types.String `tfsdk:"dag_id"`
	DeploymentApiToken types.String `tfsdk:"deployment_api_token"`
	DeploymentId       types.String `tfsdk:"deployment_id"`
	Recipients         types.Set    `tfsdk:"recipients"`
	ApiKey             types.String `tfsdk:"api_key"`
	IntegrationKey     types.String `tfsdk:"integration_key"`
	WebhookUrl         types.String `tfsdk:"webhook_url"`
}

func (data *NotificationChannelDataSource) ReadFromResponse(ctx context.Context, notificationChannel *platform.NotificationChannel) diag.Diagnostics {
	var diags diag.Diagnostics
	data.Id = types.StringValue(notificationChannel.Id)
	data.Name = types.StringValue(notificationChannel.Name)
	// Load definition into Terraform object
	data.Definition, diags = NotificationChannelDefinitionDataSourceTypesObject(ctx, notificationChannel.Definition)
	if diags.HasError() {
		return diags
	}
	data.Type = types.StringValue(notificationChannel.Type)
	if notificationChannel.WorkspaceId != nil {
		data.WorkspaceId = types.StringValue(*notificationChannel.WorkspaceId)
	} else {
		data.WorkspaceId = types.StringValue("")
	}
	if notificationChannel.DeploymentId != nil {
		data.DeploymentId = types.StringValue(*notificationChannel.DeploymentId)
	} else {
		data.DeploymentId = types.StringValue("")
	}
	data.EntityId = types.StringValue(notificationChannel.EntityId)
	data.EntityType = types.StringValue(notificationChannel.EntityType)
	if notificationChannel.EntityName != nil {
		data.EntityName = types.StringValue(*notificationChannel.EntityName)
	} else {
		data.EntityName = types.StringValue("")
	}
	data.IsShared = types.BoolValue(notificationChannel.IsShared)
	data.CreatedAt = types.StringValue(notificationChannel.CreatedAt.String())
	data.UpdatedAt = types.StringValue(notificationChannel.UpdatedAt.String())
	data.CreatedBy, diags = SubjectProfileTypesObject(ctx, notificationChannel.CreatedBy)
	if diags.HasError() {
		return diags
	}
	data.UpdatedBy, diags = SubjectProfileTypesObject(ctx, notificationChannel.UpdatedBy)
	if diags.HasError() {
		return diags
	}
	return diags
}

func (data *NotificationChannelResource) ReadFromResponse(ctx context.Context, notificationChannel *platform.NotificationChannel) diag.Diagnostics {
	var diags diag.Diagnostics
	// Preserve current definition to maintain sensitive field values
	currentDefinition := data.Definition

	data.Id = types.StringValue(notificationChannel.Id)
	data.Name = types.StringValue(notificationChannel.Name)

	// Load definition into Terraform object, preserving current state for sensitive fields
	data.Definition, diags = NotificationChannelDefinitionResourceTypesObject(ctx, notificationChannel.Definition, currentDefinition, notificationChannel.Type)
	if diags.HasError() {
		return diags
	}
	data.Type = types.StringValue(notificationChannel.Type)
	if notificationChannel.WorkspaceId != nil {
		data.WorkspaceId = types.StringValue(*notificationChannel.WorkspaceId)
	} else {
		data.WorkspaceId = types.StringValue("")
	}
	if notificationChannel.DeploymentId != nil {
		data.DeploymentId = types.StringValue(*notificationChannel.DeploymentId)
	} else {
		data.DeploymentId = types.StringValue("")
	}
	data.EntityId = types.StringValue(notificationChannel.EntityId)
	data.EntityType = types.StringValue(notificationChannel.EntityType)
	if notificationChannel.EntityName != nil {
		data.EntityName = types.StringValue(*notificationChannel.EntityName)
	} else {
		data.EntityName = types.StringValue("")
	}
	data.IsShared = types.BoolValue(notificationChannel.IsShared)
	data.CreatedAt = types.StringValue(notificationChannel.CreatedAt.String())
	data.UpdatedAt = types.StringValue(notificationChannel.UpdatedAt.String())
	data.CreatedBy, diags = SubjectProfileTypesObject(ctx, notificationChannel.CreatedBy)
	if diags.HasError() {
		return diags
	}
	data.UpdatedBy, diags = SubjectProfileTypesObject(ctx, notificationChannel.UpdatedBy)
	if diags.HasError() {
		return diags
	}
	return diags
}

// NotificationChannelDefinitionDataSourceTypesObject converts a generic interface{} to a Terraform types.Object
// matching the data source schema. The Astro API returns camelCase keys (e.g. dagId, deploymentId) which must be
// mapped to the snake_case attribute names declared in the Terraform schema before constructing the Object —
// otherwise types.ObjectValue rejects the extra keys with "Extra Object Attribute Name: <key>".
func NotificationChannelDefinitionDataSourceTypesObject(ctx context.Context, def interface{}) (types.Object, diag.Diagnostics) {
	// Cast to map[string]interface{} or default empty
	var defMap map[string]interface{}
	if m, ok := def.(map[string]interface{}); ok {
		defMap = m
	} else {
		defMap = make(map[string]interface{})
	}

	// Initialize all expected attributes with null values
	defAttrMap := map[string]attr.Value{
		"dag_id":               types.StringNull(),
		"deployment_api_token": types.StringNull(),
		"deployment_id":        types.StringNull(),
		"recipients":           types.SetNull(types.StringType),
		"api_key":              types.StringNull(),
		"integration_key":      types.StringNull(),
		"webhook_url":          types.StringNull(),
	}

	// Map every camelCase key the OpenAPI spec declares for a NotificationChannelDefinition to its snake_case
	// schema attribute. Sensitive fields (deploymentApiToken, apiKey, integrationKey, webhookUrl) are typically
	// stripped by the API for security, but we handle them defensively in case the backend ever returns them.

	// String fields: camelCase API key → snake_case schema attr
	stringKeyMap := map[string]string{
		"dagId":              "dag_id",
		"deploymentApiToken": "deployment_api_token",
		"deploymentId":       "deployment_id",
		"apiKey":             "api_key",
		"integrationKey":     "integration_key",
		"webhookUrl":         "webhook_url",
	}
	for apiKey, attrKey := range stringKeyMap {
		if v, ok := defMap[apiKey]; ok && v != nil {
			if s, ok := v.(string); ok && s != "" {
				defAttrMap[attrKey] = types.StringValue(s)
			}
		}
	}

	// Handle recipients (set of strings, used by EMAIL channels)
	if v, ok := defMap["recipients"]; ok && v != nil {
		if arr, ok := v.([]interface{}); ok {
			recipientVals := make([]attr.Value, 0, len(arr))
			for _, el := range arr {
				if s, ok2 := el.(string); ok2 {
					recipientVals = append(recipientVals, types.StringValue(s))
				}
			}
			if len(recipientVals) > 0 {
				defAttrMap["recipients"] = types.SetValueMust(types.StringType, recipientVals)
			} else {
				defAttrMap["recipients"] = types.SetValueMust(types.StringType, []attr.Value{})
			}
		}
	}

	// Surface unrecognized keys at Warn level so future API additions are visible without breaking apply.
	for k := range defMap {
		if _, known := stringKeyMap[k]; known {
			continue
		}
		if k == "recipients" {
			continue
		}
		tflog.Warn(ctx, "Ignoring unrecognized notification channel definition key from API", map[string]interface{}{"key": k})
	}

	obj, objDiags := types.ObjectValue(schemas.NotificationChannelDefinitionAttributeTypes(), defAttrMap)
	if objDiags.HasError() {
		return types.Object{}, objDiags
	}
	return obj, nil
}

// NotificationChannelDefinitionResourceTypesObject maps platform notification channel definitions into a Terraform types.Object matching the resource schema.
func NotificationChannelDefinitionResourceTypesObject(ctx context.Context, def interface{}, preserveCurrentState types.Object, channelType string) (types.Object, diag.Diagnostics) {
	// Convert def into a map[string]interface{} for processing
	var defMap map[string]interface{}
	switch v := def.(type) {
	case map[string]interface{}:
		defMap = v
	case *map[string]interface{}:
		if v != nil {
			defMap = *v
		} else {
			defMap = make(map[string]interface{})
		}
	default:
		tflog.Error(ctx, "Unexpected type passed into NotificationChannelDefinitionResourceTypesObject", map[string]interface{}{"value": def})
		return types.Object{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Internal Error", "NotificationChannelDefinitionResourceTypesObject expects a map[string]interface{} type"),
		}
	}

	// Initialize all expected attributes with null values
	defAttrMap := map[string]attr.Value{
		"dag_id":               types.StringNull(),
		"deployment_api_token": types.StringNull(),
		"deployment_id":        types.StringNull(),
		"integration_key":      types.StringNull(),
		"api_key":              types.StringNull(),
		"recipients":           types.SetNull(types.StringType),
		"webhook_url":          types.StringNull(),
	}

	// Preserve values from current state for sensitive fields that are relevant to this channel type
	if !preserveCurrentState.IsNull() && !preserveCurrentState.IsUnknown() {
		currentAttrs := preserveCurrentState.Attributes()

		// Only preserve sensitive fields that are relevant to the current channel type
		switch channelType {
		case string(platform.AlertNotificationChannelTypeSLACK):
			if val, exists := currentAttrs["webhook_url"]; exists && !val.IsNull() && !val.IsUnknown() {
				defAttrMap["webhook_url"] = val
			}
		case string(platform.AlertNotificationChannelTypePAGERDUTY):
			if val, exists := currentAttrs["integration_key"]; exists && !val.IsNull() && !val.IsUnknown() {
				defAttrMap["integration_key"] = val
			}
		case string(platform.AlertNotificationChannelTypeOPSGENIE):
			if val, exists := currentAttrs["api_key"]; exists && !val.IsNull() && !val.IsUnknown() {
				defAttrMap["api_key"] = val
			}
		case string(platform.AlertNotificationChannelTypeDAGTRIGGER):
			if val, exists := currentAttrs["dag_id"]; exists && !val.IsNull() && !val.IsUnknown() {
				defAttrMap["dag_id"] = val
			}
			if val, exists := currentAttrs["deployment_api_token"]; exists && !val.IsNull() && !val.IsUnknown() {
				defAttrMap["deployment_api_token"] = val
			}
			if val, exists := currentAttrs["deployment_id"]; exists && !val.IsNull() && !val.IsUnknown() {
				defAttrMap["deployment_id"] = val
			}
		}
	}

	// Handle fields that the API actually returned (non-sensitive fields typically)

	// Handle recipients (from recipients)
	if v, ok := defMap["recipients"]; ok && v != nil {
		if arr, ok := v.([]interface{}); ok && len(arr) > 0 {
			recipientVals := make([]attr.Value, 0, len(arr))
			for _, el := range arr {
				if s, ok2 := el.(string); ok2 {
					recipientVals = append(recipientVals, types.StringValue(s))
				}
			}
			defAttrMap["recipients"] = types.SetValueMust(types.StringType, recipientVals)
		} else {
			// API returned empty recipients array
			defAttrMap["recipients"] = types.SetValueMust(types.StringType, []attr.Value{})
		}
	}

	// Handle deployment_id (from deploymentId)
	if v, ok := defMap["deploymentId"]; ok && v != nil {
		if s, ok := v.(string); ok && s != "" {
			defAttrMap["deployment_id"] = types.StringValue(s)
		}
	}

	// Handle dag_id (from dagId)
	if v, ok := defMap["dagId"]; ok && v != nil {
		if s, ok := v.(string); ok && s != "" {
			defAttrMap["dag_id"] = types.StringValue(s)
		}
	}

	// DO NOT handle these sensitive fields from API response - they're never returned:
	// - webhook_url (SLACK - never returned by API for security)
	// - api_key (OPSGENIE - never returned by API for security)
	// - integration_key (PAGERDUTY - never returned by API for security)
	// - deployment_api_token (DAGTRIGGER - never returned by API for security)

	obj, objDiags := types.ObjectValue(
		schemas.NotificationChannelDefinitionAttributeTypes(),
		defAttrMap,
	)
	if objDiags.HasError() {
		return types.Object{}, objDiags
	}

	return obj, nil
}
