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

// Alert describes the data source data model.
type Alert struct {
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

func (data *Alert) ReadFromResponse(ctx context.Context, Alert *platform.Alert) diag.Diagnostics {
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
	propMap := make(map[string]interface{})
	if m, ok := rulesPtr.Properties.(map[string]interface{}); ok {
		propMap = m
	}
	propAttrMap := make(map[string]attr.Value, len(propMap))
	for k, v := range propMap {
		propAttrMap[k] = types.StringValue(fmt.Sprintf("%v", v))
	}
	properties, propDiags := types.MapValue(types.StringType, propAttrMap)
	if propDiags.HasError() {
		return types.Object{}, propDiags
	}
	// Convert pattern matches to types.Set
	var pmVals []attr.Value
	if rulesPtr.PatternMatches != nil {
		for _, pm := range *rulesPtr.PatternMatches {
			// Convert values slice to []attr.Value
			vals := make([]attr.Value, len(pm.Values))
			for j, val := range pm.Values {
				vals[j] = types.StringValue(val)
			}
			valuesSet, valDiags := types.SetValue(types.StringType, vals)
			if valDiags.HasError() {
				return types.Object{}, valDiags
			}
			// Build pattern match object
			pmObj, pmDiags := types.ObjectValueFrom(ctx, schemas.AlertRulesPatternMatchAttributeTypes(), AlertRulesPatternMatch{
				EntityType:   types.StringValue(string(pm.EntityType)),
				OperatorType: types.StringValue(string(pm.OperatorType)),
				Values:       valuesSet,
			})
			if pmDiags.HasError() {
				return types.Object{}, pmDiags
			}
			pmVals = append(pmVals, pmObj)
		}
	}
	pmSet, pmDiags := types.SetValue(types.ObjectType{AttrTypes: schemas.AlertRulesPatternMatchAttributeTypes()}, pmVals)
	if pmDiags.HasError() {
		return types.Object{}, pmDiags
	}
	alertRules := AlertRules{
		Properties:     properties,
		PatternMatches: pmSet,
	}

	return types.ObjectValueFrom(ctx, schemas.AlertRulesAttributeTypes(), alertRules)
}
