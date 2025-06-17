package models

import (
	"context"
	"fmt"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NotificationChannelDataSource describes the data source data model.
type NotificationChannelDataSource struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Definition     types.Map    `tfsdk:"definition"`
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

func (data *NotificationChannelDataSource) ReadFromResponse(ctx context.Context, notificationChannel *platform.NotificationChannel) diag.Diagnostics {
	var diags diag.Diagnostics
	data.Id = types.StringValue(notificationChannel.Id)
	data.Name = types.StringValue(notificationChannel.Name)
	// Load definition into Terraform map
	data.Definition, diags = definitionToMap(notificationChannel.Definition)
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

// definitionToMap converts a generic interface{} to a Terraform types.Map of string values
func definitionToMap(def interface{}) (types.Map, diag.Diagnostics) {
	// Cast to map[string]interface{} or default empty
	var defMap map[string]interface{}
	if m, ok := def.(map[string]interface{}); ok {
		defMap = m
	} else {
		defMap = make(map[string]interface{})
	}
	// Build Terraform attribute map
	defAttrMap := make(map[string]attr.Value, len(defMap))
	for k, v := range defMap {
		defAttrMap[k] = types.StringValue(fmt.Sprintf("%v", v))
	}
	// Create Terraform map
	return types.MapValue(types.StringType, defAttrMap)
}
