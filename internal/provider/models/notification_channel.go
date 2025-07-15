package models

import (
	"context"
	"fmt"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NotificationChannelDataSource describes the data source data model.
type NotificationChannelDataSource struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Definition     types.Object `tfsdk:"definition"`
	Type           types.String `tfsdk:"type"`
	OrganizationId types.String `tfsdk:"organization_id"`
	WorkspaceId    types.String `tfsdk:"workspace_id"`
	DeploymentId   types.String `tfsdk:"deployment_id"`
	EntityId       types.String `tfsdk:"entity_id"`
	EntityType     types.String `tfsdk:"entity_type"`
	EntityName     types.String `tfsdk:"entity_name"`
	IsShared       types.Bool   `tfsdk:"is_shared"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	CreatedBy      types.Object `tfsdk:"created_by"`
	UpdatedBy      types.Object `tfsdk:"updated_by"`
}

// NotificationChannelResource describes the resource data model.
type NotificationChannelResource struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Definition     types.Object `tfsdk:"definition"`
	Type           types.String `tfsdk:"type"`
	OrganizationId types.String `tfsdk:"organization_id"`
	WorkspaceId    types.String `tfsdk:"workspace_id"`
	DeploymentId   types.String `tfsdk:"deployment_id"`
	EntityId       types.String `tfsdk:"entity_id"`
	EntityType     types.String `tfsdk:"entity_type"`
	EntityName     types.String `tfsdk:"entity_name"`
	IsShared       types.Bool   `tfsdk:"is_shared"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	CreatedBy      types.Object `tfsdk:"created_by"`
	UpdatedBy      types.Object `tfsdk:"updated_by"`
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
	data.Definition, diags = NotificationChannelDefinitionResourceTypesObject(ctx, notificationChannel.Definition)
	if diags.HasError() {
		return diags
	}
	data.Type = types.StringValue(notificationChannel.Type)
	data.OrganizationId = types.StringValue(notificationChannel.OrganizationId)
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
	data.CreatedAt = types.StringValue(notificationChannel.CreatedAt)
	data.UpdatedAt = types.StringValue(notificationChannel.UpdatedAt)
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
	data.Id = types.StringValue(notificationChannel.Id)
	data.Name = types.StringValue(notificationChannel.Name)
	// Load definition into Terraform object
	data.Definition, diags = NotificationChannelDefinitionResourceTypesObject(ctx, notificationChannel.Definition)
	if diags.HasError() {
		return diags
	}
	data.Type = types.StringValue(notificationChannel.Type)
	data.OrganizationId = types.StringValue(notificationChannel.OrganizationId)
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
	data.CreatedAt = types.StringValue(notificationChannel.CreatedAt)
	data.UpdatedAt = types.StringValue(notificationChannel.UpdatedAt)
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

// NotificationChannelDefinitionResourceTypesObject converts a generic interface{} to a Terraform types.Object
func NotificationChannelDefinitionResourceTypesObject(ctx context.Context, def interface{}) (types.Object, diag.Diagnostics) {
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

	// Override with actual values when present
	for k, v := range defMap {
		switch val := v.(type) {
		case string:
			if val != "" {
				defAttrMap[k] = types.StringValue(val)
			}
		case []interface{}:
			// Handle array values (like recipients)
			var stringValues []attr.Value
			for _, item := range val {
				if str, ok := item.(string); ok && str != "" {
					stringValues = append(stringValues, types.StringValue(str))
				}
			}
			if len(stringValues) > 0 {
				set, diags := types.SetValue(types.StringType, stringValues)
				if diags.HasError() {
					return types.Object{}, diags
				}
				defAttrMap[k] = set
			}
		default:
			if v != nil {
				defAttrMap[k] = types.StringValue(fmt.Sprintf("%v", v))
			}
		}
	}

	// Create Terraform object using the definition attribute types
	return types.ObjectValue(schemas.NotificationChannelDefinitionResourceAttributeTypes(), defAttrMap)
}
