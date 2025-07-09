package models

import (
	"context"
	"fmt"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// AlertDataSource describes the data source data model.
type AlertDataSource struct {
	Id                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Type                 types.String `tfsdk:"type"`
	Rules                types.Object `tfsdk:"rules"`
	EntityId             types.String `tfsdk:"entity_id"`
	EntityType           types.String `tfsdk:"entity_type"`
	EntityName           types.String `tfsdk:"entity_name"`
	NotificationChannels types.Set    `tfsdk:"notification_channels"`
	OrganizationId       types.String `tfsdk:"organization_id"`
	WorkspaceId          types.String `tfsdk:"workspace_id"`
	DeploymentId         types.String `tfsdk:"deployment_id"`
	Severity             types.String `tfsdk:"severity"`
	CreatedAt            types.String `tfsdk:"created_at"`
	UpdatedAt            types.String `tfsdk:"updated_at"`
	CreatedBy            types.Object `tfsdk:"created_by"`
	UpdatedBy            types.Object `tfsdk:"updated_by"`
}

// AlertListModel is used for listing alerts without notification channels.
type AlertListModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	Rules          types.Object `tfsdk:"rules"`
	EntityId       types.String `tfsdk:"entity_id"`
	EntityType     types.String `tfsdk:"entity_type"`
	EntityName     types.String `tfsdk:"entity_name"`
	OrganizationId types.String `tfsdk:"organization_id"`
	WorkspaceId    types.String `tfsdk:"workspace_id"`
	DeploymentId   types.String `tfsdk:"deployment_id"`
	Severity       types.String `tfsdk:"severity"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	CreatedBy      types.Object `tfsdk:"created_by"`
	UpdatedBy      types.Object `tfsdk:"updated_by"`
}

// AlertResource describes the data source data model.
type AlertResource struct {
	Id                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Type                   types.String `tfsdk:"type"`
	Rules                  types.Object `tfsdk:"rules"`
	EntityId               types.String `tfsdk:"entity_id"`
	EntityType             types.String `tfsdk:"entity_type"`
	EntityName             types.String `tfsdk:"entity_name"`
	NotificationChannelIds types.Set    `tfsdk:"notification_channel_ids"`
	NotificationChannels   types.Set    `tfsdk:"notification_channels"`
	OrganizationId         types.String `tfsdk:"organization_id"`
	WorkspaceId            types.String `tfsdk:"workspace_id"`
	DeploymentId           types.String `tfsdk:"deployment_id"`
	Severity               types.String `tfsdk:"severity"`
	CreatedAt              types.String `tfsdk:"created_at"`
	UpdatedAt              types.String `tfsdk:"updated_at"`
	CreatedBy              types.Object `tfsdk:"created_by"`
	UpdatedBy              types.Object `tfsdk:"updated_by"`
}

type AlertRules struct {
	Properties     types.Map `tfsdk:"properties"`
	PatternMatches types.Set `tfsdk:"pattern_matches"`
}

// AlertRulesPatternMatch describes element type for pattern_matches in AlertRules.
type AlertRulesPatternMatch struct {
	EntityType   types.String `tfsdk:"entity_type"`
	OperatorType types.String `tfsdk:"operator_type"`
	Values       types.Set    `tfsdk:"values"`
}

// ResourceAlertPatternMatch describes element type for pattern_matches in Alert resource.
type ResourceAlertPatternMatchInput struct {
	EntityType   string   `tfsdk:"entity_type"`
	OperatorType string   `tfsdk:"operator_type"`
	Values       []string `tfsdk:"values"`
}

// ResourceAlertPropertiesInput decodes the Terraform 'properties' nested block and can handle nulls.
type ResourceAlertPropertiesInput struct {
	DeploymentId          types.String `tfsdk:"deployment_id"`            // always present
	DagDurationSeconds    types.Int64  `tfsdk:"dag_duration_seconds"`     // optional
	DagDeadline           types.String `tfsdk:"dag_deadline"`             // optional
	DaysOfWeek            types.Set    `tfsdk:"days_of_week"`             // optional
	LookBackPeriodSeconds types.Int64  `tfsdk:"look_back_period_seconds"` // optional
	TaskDurationSeconds   types.Int64  `tfsdk:"task_duration_seconds"`    // optional
}

// ResourceAlertRulesPatternMatch is used to build the Terraform object for each pattern match in AlertRules for resource.
type ResourceAlertRulesPatternMatch struct {
	EntityType   types.String `tfsdk:"entity_type"`
	OperatorType types.String `tfsdk:"operator_type"`
	Values       types.Set    `tfsdk:"values"`
}

// ResourceAlertRulesInput is used to decode the Terraform 'rules' block when creating or updating an Alert resource.
type ResourceAlertRulesInput struct {
	PatternMatches []ResourceAlertPatternMatchInput `tfsdk:"pattern_matches"`
	Properties     ResourceAlertPropertiesInput     `tfsdk:"properties"`
}

func (data *AlertDataSource) ReadFromResponse(ctx context.Context, Alert *platform.Alert) diag.Diagnostics {
	var diags diag.Diagnostics
	data.Id = types.StringValue(Alert.Id)
	data.Name = types.StringValue(Alert.Name)
	data.Type = types.StringValue(string(Alert.Type))
	data.Rules, diags = AlertRulesTypesObject(ctx, Alert.Rules)
	if diags.HasError() {
		return diags
	}
	data.EntityId = types.StringValue(Alert.EntityId)
	data.EntityType = types.StringValue(string(Alert.EntityType))
	if Alert.EntityName != nil {
		data.EntityName = types.StringValue(*Alert.EntityName)
	} else {
		data.EntityName = types.StringValue("")
	}
	data.NotificationChannels, diags = AlertNotificationChannelsTypesSet(ctx, Alert.NotificationChannels)
	if diags.HasError() {
		return diags
	}
	data.OrganizationId = types.StringValue(Alert.OrganizationId)
	if Alert.WorkspaceId != nil {
		data.WorkspaceId = types.StringValue(*Alert.WorkspaceId)
	} else {
		data.WorkspaceId = types.StringValue("")
	}
	if Alert.DeploymentId != nil {
		data.DeploymentId = types.StringValue(*Alert.DeploymentId)
	} else {
		data.DeploymentId = types.StringValue("")
	}
	data.Severity = types.StringValue(string(Alert.Severity))
	data.CreatedAt = types.StringValue(Alert.CreatedAt.String())
	data.UpdatedAt = types.StringValue(Alert.UpdatedAt.String())
	data.CreatedBy, diags = SubjectProfileTypesObject(ctx, Alert.CreatedBy)
	if diags.HasError() {
		return diags
	}
	data.UpdatedBy, diags = SubjectProfileTypesObject(ctx, Alert.UpdatedBy)
	if diags.HasError() {
		return diags
	}

	return nil
}

// ReadFromAlertListResponse populates AlertListModel from a platform.Alert, omitting notification_channels.
func (data *AlertListModel) ReadFromAlertListResponse(ctx context.Context, alert *platform.Alert) diag.Diagnostics {
	var diags diag.Diagnostics
	data.Id = types.StringValue(alert.Id)
	data.Name = types.StringValue(alert.Name)
	data.Type = types.StringValue(string(alert.Type))
	data.Rules, diags = AlertRulesTypesObject(ctx, alert.Rules)
	if diags.HasError() {
		return diags
	}
	data.EntityId = types.StringValue(alert.EntityId)
	data.EntityType = types.StringValue(string(alert.EntityType))
	if alert.EntityName != nil {
		data.EntityName = types.StringValue(*alert.EntityName)
	} else {
		data.EntityName = types.StringValue("")
	}
	data.OrganizationId = types.StringValue(alert.OrganizationId)
	if alert.WorkspaceId != nil {
		data.WorkspaceId = types.StringValue(*alert.WorkspaceId)
	} else {
		data.WorkspaceId = types.StringValue("")
	}
	if alert.DeploymentId != nil {
		data.DeploymentId = types.StringValue(*alert.DeploymentId)
	} else {
		data.DeploymentId = types.StringValue("")
	}
	data.Severity = types.StringValue(string(alert.Severity))
	data.CreatedAt = types.StringValue(alert.CreatedAt.String())
	data.UpdatedAt = types.StringValue(alert.UpdatedAt.String())
	data.CreatedBy, diags = SubjectProfileTypesObject(ctx, alert.CreatedBy)
	if diags.HasError() {
		return diags
	}
	data.UpdatedBy, diags = SubjectProfileTypesObject(ctx, alert.UpdatedBy)
	if diags.HasError() {
		return diags
	}
	return nil
}

func (data *AlertResource) ReadFromResponse(ctx context.Context, Alert *platform.Alert) diag.Diagnostics {
	var diags diag.Diagnostics
	data.Id = types.StringValue(Alert.Id)
	data.Name = types.StringValue(Alert.Name)
	data.Type = types.StringValue(string(Alert.Type))
	data.Rules, diags = AlertRulesResourceTypesObject(ctx, Alert.Rules)
	if diags.HasError() {
		return diags
	}
	data.EntityId = types.StringValue(Alert.EntityId)
	data.EntityType = types.StringValue(string(Alert.EntityType))
	if Alert.EntityName != nil {
		data.EntityName = types.StringValue(*Alert.EntityName)
	} else {
		data.EntityName = types.StringValue("")
	}
	data.NotificationChannels, diags = AlertNotificationChannelsTypesSet(ctx, Alert.NotificationChannels)
	if diags.HasError() {
		return diags
	}
	data.OrganizationId = types.StringValue(Alert.OrganizationId)
	if Alert.WorkspaceId != nil {
		data.WorkspaceId = types.StringValue(*Alert.WorkspaceId)
	} else {
		data.WorkspaceId = types.StringValue("")
	}
	if Alert.DeploymentId != nil {
		data.DeploymentId = types.StringValue(*Alert.DeploymentId)
	} else {
		data.DeploymentId = types.StringValue("")
	}
	data.Severity = types.StringValue(string(Alert.Severity))
	data.CreatedAt = types.StringValue(Alert.CreatedAt.String())
	data.UpdatedAt = types.StringValue(Alert.UpdatedAt.String())
	data.CreatedBy, diags = SubjectProfileTypesObject(ctx, Alert.CreatedBy)
	if diags.HasError() {
		return diags
	}
	data.UpdatedBy, diags = SubjectProfileTypesObject(ctx, Alert.UpdatedBy)
	if diags.HasError() {
		return diags
	}
	return nil
}

func AlertRulesTypesObject(
	ctx context.Context,
	rules any,
) (types.Object, diag.Diagnostics) {
	// Attempt to convert rules to *platform.AlertRules
	var rulesPtr *platform.AlertRules

	switch v := rules.(type) {
	case platform.AlertRules:
		rulesPtr = &v
	case *platform.AlertRules:
		rulesPtr = v
	default:
		tflog.Error(
			ctx,
			"Unexpected type passed into alert rules",
			map[string]interface{}{"value": rules},
		)
		return types.Object{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Internal Error",
				"AlertRulesTypesObject expects a platform.AlertRules type but did not receive one",
			),
		}
	}

	// Convert properties to types.Map
	propertiesMap := make(map[string]interface{})
	if m, ok := rulesPtr.Properties.(map[string]interface{}); ok {
		propertiesMap = m
	}
	propertiesAttrMap := make(map[string]attr.Value, len(propertiesMap))
	for k, v := range propertiesMap {
		propertiesAttrMap[k] = types.StringValue(fmt.Sprintf("%v", v))
	}
	properties, propDiags := types.MapValue(types.StringType, propertiesAttrMap)
	if propDiags.HasError() {
		return types.Object{}, propDiags
	}
	// Convert pattern matches to types.Set
	var patternMatchValues []attr.Value
	if rulesPtr.PatternMatches != nil {
		for _, patternMatch := range *rulesPtr.PatternMatches {
			// Convert values slice to []attr.Value
			values := make([]attr.Value, len(patternMatch.Values))
			for j, val := range patternMatch.Values {
				values[j] = types.StringValue(val)
			}
			valuesSet, valDiags := types.SetValue(types.StringType, values)
			if valDiags.HasError() {
				return types.Object{}, valDiags
			}
			// Build pattern match object
			pmObj, pmDiags := types.ObjectValue(schemas.AlertRulesPatternMatchAttributeTypes(), map[string]attr.Value{
				"entity_type":   types.StringValue(string(patternMatch.EntityType)),
				"operator_type": types.StringValue(string(patternMatch.OperatorType)),
				"values":        valuesSet,
			})
			if pmDiags.HasError() {
				return types.Object{}, pmDiags
			}
			patternMatchValues = append(patternMatchValues, pmObj)
		}
	}
	patternMatchSet, patternMatchDiags := types.SetValue(types.ObjectType{AttrTypes: schemas.AlertRulesPatternMatchAttributeTypes()}, patternMatchValues)
	if patternMatchDiags.HasError() {
		return types.Object{}, patternMatchDiags
	}
	return types.ObjectValue(schemas.AlertRulesAttributeTypes(), map[string]attr.Value{
		"properties":      properties,
		"pattern_matches": patternMatchSet,
	})
}

// AlertNotificationChannelsTypesSet converts a slice of platform.AlertNotificationChannel into a Terraform types.Set of nested NotificationChannelDataSource objects
func AlertNotificationChannelsTypesSet(ctx context.Context, channels any) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	// Attempt to convert channels to slice
	var slice []platform.AlertNotificationChannel
	switch v := channels.(type) {
	case []platform.AlertNotificationChannel:
		slice = v
	case *[]platform.AlertNotificationChannel:
		slice = *v
	default:
		tflog.Error(ctx, "Unexpected type passed into alert notification channels", map[string]interface{}{"value": channels})
		return types.Set{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Internal Error",
				"AlertNotificationChannelsTypesSet expects a slice of platform.AlertNotificationChannel",
			),
		}
	}
	var notificationChannelValues []attr.Value
	for _, anc := range slice {
		// Map AlertNotificationChannel fields into NotificationChannelDataSource via temporary platform.NotificationChannel
		pc := platform.NotificationChannel{
			CreatedAt:      anc.CreatedAt,
			CreatedBy:      platform.BasicSubjectProfile{},
			Definition:     anc.Definition,
			DeploymentId:   anc.DeploymentId,
			EntityId:       anc.EntityId,
			EntityName:     nil,
			EntityType:     string(anc.EntityType),
			Id:             anc.Id,
			IsShared:       false,
			Name:           anc.Name,
			OrganizationId: anc.OrganizationId,
			Type:           string(anc.Type),
			UpdatedAt:      anc.UpdatedAt,
			UpdatedBy:      platform.BasicSubjectProfile{},
			WorkspaceId:    anc.WorkspaceId,
		}
		var single NotificationChannelDataSource
		diagsC := single.ReadFromResponse(ctx, &pc)
		if diagsC.HasError() {
			return types.Set{}, diagsC
		}
		obj, diagsC := types.ObjectValueFrom(ctx, schemas.NotificationChannelsElementAttributeTypes(), single)
		if diagsC.HasError() {
			return types.Set{}, diagsC
		}
		notificationChannelValues = append(notificationChannelValues, obj)
	}
	setVal, diagsSet := types.SetValue(types.ObjectType{AttrTypes: schemas.NotificationChannelsElementAttributeTypes()}, notificationChannelValues)
	diags = append(diags, diagsSet...)
	return setVal, diags
}

// AlertRulesResourceTypesObject maps platform.AlertRules into a Terraform types.Object matching the resource schema.
func AlertRulesResourceTypesObject(ctx context.Context, rules any) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	// Convert rules into *platform.AlertRules
	var rulesPtr *platform.AlertRules
	switch v := rules.(type) {
	case platform.AlertRules:
		rulesPtr = &v
	case *platform.AlertRules:
		rulesPtr = v
	default:
		tflog.Error(ctx, "Unexpected type passed into AlertRulesResourceTypesObject", map[string]interface{}{"value": rules})
		return types.Object{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Internal Error", "AlertRulesResourceTypesObject expects a platform.AlertRules type"),
		}
	}
	// Build properties nested object from raw map
	propMap := make(map[string]interface{})
	if m, ok := rulesPtr.Properties.(map[string]interface{}); ok {
		propMap = m
	}
	// extract values
	depsId, _ := propMap["deploymentId"].(string)

	// Handle optional fields - only set if present in the map
	var dagDurationSeconds types.Int64
	if v, ok := propMap["dagDurationSeconds"]; ok && v != nil {
		if f, ok := v.(float64); ok {
			dagDurationSeconds = types.Int64Value(int64(f))
		} else {
			dagDurationSeconds = types.Int64Null()
		}
	} else {
		dagDurationSeconds = types.Int64Null()
	}

	var dagDeadline types.String
	if v, ok := propMap["dagDeadline"]; ok && v != nil {
		if s, ok := v.(string); ok && s != "" {
			dagDeadline = types.StringValue(s)
		} else {
			dagDeadline = types.StringNull()
		}
	} else {
		dagDeadline = types.StringNull()
	}

	var daysOfWeek types.Set
	if v, ok := propMap["daysOfWeek"]; ok && v != nil {
		if arr, ok := v.([]interface{}); ok && len(arr) > 0 {
			dayVals := make([]attr.Value, 0, len(arr))
			for _, el := range arr {
				if s, ok2 := el.(string); ok2 {
					dayVals = append(dayVals, types.StringValue(s))
				}
			}
			daysOfWeek = types.SetValueMust(types.StringType, dayVals)
		} else {
			daysOfWeek = types.SetNull(types.StringType)
		}
	} else {
		daysOfWeek = types.SetNull(types.StringType)
	}

	var lookBackPeriodSeconds types.Int64
	if v, ok := propMap["lookBackPeriodSeconds"]; ok && v != nil {
		if f, ok := v.(float64); ok {
			lookBackPeriodSeconds = types.Int64Value(int64(f))
		} else {
			lookBackPeriodSeconds = types.Int64Null()
		}
	} else {
		lookBackPeriodSeconds = types.Int64Null()
	}

	var taskDurationSeconds types.Int64
	if v, ok := propMap["taskDurationSeconds"]; ok && v != nil {
		if f, ok := v.(float64); ok {
			taskDurationSeconds = types.Int64Value(int64(f))
		} else {
			taskDurationSeconds = types.Int64Null()
		}
	} else {
		taskDurationSeconds = types.Int64Null()
	}
	props, propDiags := types.ObjectValue(
		schemas.AlertRulesResourceAttributeTypes()["properties"].(types.ObjectType).AttrTypes,
		map[string]attr.Value{
			"deployment_id":            types.StringValue(depsId),
			"dag_duration_seconds":     dagDurationSeconds,
			"dag_deadline":             dagDeadline,
			"days_of_week":             daysOfWeek,
			"look_back_period_seconds": lookBackPeriodSeconds,
			"task_duration_seconds":    taskDurationSeconds,
		},
	)
	if propDiags.HasError() {
		return types.Object{}, propDiags
	}
	// Build pattern_matches list
	pmAttrTypes := schemas.AlertRulesPatternMatchAttributeTypes()
	var patternMatchValues []attr.Value
	if rulesPtr.PatternMatches != nil {
		for _, patternMatch := range *rulesPtr.PatternMatches {
			values := make([]attr.Value, len(patternMatch.Values))
			for i, val := range patternMatch.Values {
				values[i] = types.StringValue(val)
			}
			valuesSet, valuesSetDiags := types.SetValue(types.StringType, values)
			if valuesSetDiags.HasError() {
				return types.Object{}, valuesSetDiags
			}
			// use resource-level pattern match struct
			pmObj, pmDiags2 := types.ObjectValue(pmAttrTypes, map[string]attr.Value{
				"entity_type":   types.StringValue(string(patternMatch.EntityType)),
				"operator_type": types.StringValue(string(patternMatch.OperatorType)),
				"values":        valuesSet,
			})
			if pmDiags2.HasError() {
				return types.Object{}, pmDiags2
			}
			patternMatchValues = append(patternMatchValues, pmObj)
		}
	}
	patternMatchSet, patternMatchDiags := types.SetValue(types.ObjectType{AttrTypes: pmAttrTypes}, patternMatchValues)
	if patternMatchDiags.HasError() {
		return types.Object{}, patternMatchDiags
	}
	diags = append(diags, patternMatchDiags...)
	// Final object
	obj, objDiags := types.ObjectValue(schemas.AlertRulesResourceAttributeTypes(), map[string]attr.Value{
		"properties":      props,
		"pattern_matches": patternMatchSet,
	})
	diags = append(diags, objDiags...)
	return obj, diags
}
