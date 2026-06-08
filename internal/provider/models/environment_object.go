package models

import (
	"context"
	"encoding/json"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// EnvironmentObject describes the resource and data source data model.
type EnvironmentObject struct {
	Id                  types.String `tfsdk:"id"`
	ObjectKey           types.String `tfsdk:"object_key"`
	ObjectType          types.String `tfsdk:"object_type"`
	Scope               types.String `tfsdk:"scope"`
	ScopeEntityId       types.String `tfsdk:"scope_entity_id"`
	SourceScope         types.String `tfsdk:"source_scope"`
	SourceScopeEntityId types.String `tfsdk:"source_scope_entity_id"`
	AutoLinkDeployments types.Bool   `tfsdk:"auto_link_deployments"`
	AirflowVariable     types.Object `tfsdk:"airflow_variable"`
	ConnectionConfig    types.Object `tfsdk:"connection_config"`
	MetricsExport       types.Object `tfsdk:"metrics_export"`
	Links               types.List   `tfsdk:"links"`
	ExcludeLinks        types.List   `tfsdk:"exclude_links"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
	CreatedBy           types.Object `tfsdk:"created_by"`
	UpdatedBy           types.Object `tfsdk:"updated_by"`
}

func (data *EnvironmentObject) ReadFromResponse(ctx context.Context, obj *platform.EnvironmentObject) diag.Diagnostics {
	var diags diag.Diagnostics

	data.Id = types.StringPointerValue(obj.Id)
	data.ObjectKey = types.StringValue(obj.ObjectKey)
	data.ObjectType = types.StringValue(string(obj.ObjectType))
	data.Scope = types.StringValue(string(obj.Scope))
	data.ScopeEntityId = types.StringValue(obj.ScopeEntityId)

	if obj.SourceScope != nil {
		data.SourceScope = types.StringValue(string(*obj.SourceScope))
	} else {
		data.SourceScope = types.StringNull()
	}
	data.SourceScopeEntityId = types.StringPointerValue(obj.SourceScopeEntityId)

	if obj.AutoLinkDeployments != nil {
		data.AutoLinkDeployments = types.BoolValue(*obj.AutoLinkDeployments)
	} else {
		data.AutoLinkDeployments = types.BoolNull()
	}

	data.CreatedAt = types.StringPointerValue(obj.CreatedAt)
	data.UpdatedAt = types.StringPointerValue(obj.UpdatedAt)

	if obj.CreatedBy != nil {
		data.CreatedBy, diags = SubjectProfileTypesObject(ctx, obj.CreatedBy)
		if diags.HasError() {
			return diags
		}
	} else {
		data.CreatedBy = types.ObjectNull(schemas.SubjectProfileAttributeTypes())
	}

	if obj.UpdatedBy != nil {
		data.UpdatedBy, diags = SubjectProfileTypesObject(ctx, obj.UpdatedBy)
		if diags.HasError() {
			return diags
		}
	} else {
		data.UpdatedBy = types.ObjectNull(schemas.SubjectProfileAttributeTypes())
	}

	// Airflow Variable
	if obj.AirflowVariable != nil {
		data.AirflowVariable, diags = types.ObjectValue(schemas.EnvironmentObjectAirflowVariableAttributeTypes(), map[string]attr.Value{
			"value":     types.StringValue(obj.AirflowVariable.Value),
			"is_secret": types.BoolValue(obj.AirflowVariable.IsSecret),
		})
		if diags.HasError() {
			return diags
		}
	} else {
		data.AirflowVariable = types.ObjectNull(schemas.EnvironmentObjectAirflowVariableAttributeTypes())
	}

	// Connection
	if obj.Connection != nil {
		data.ConnectionConfig, diags = environmentObjectConnectionToObject(ctx, obj.Connection)
		if diags.HasError() {
			return diags
		}
	} else {
		data.ConnectionConfig = types.ObjectNull(schemas.EnvironmentObjectConnectionAttributeTypes())
	}

	// Metrics Export
	if obj.MetricsExport != nil {
		data.MetricsExport, diags = environmentObjectMetricsExportToObject(ctx, obj.MetricsExport)
		if diags.HasError() {
			return diags
		}
	} else {
		data.MetricsExport = types.ObjectNull(schemas.EnvironmentObjectMetricsExportAttributeTypes())
	}

	// Links
	if obj.Links != nil && len(*obj.Links) > 0 {
		linkObjects := make([]attr.Value, len(*obj.Links))
		for i, link := range *obj.Links {
			linkObjects[i], diags = environmentObjectLinkToObject(ctx, &link)
			if diags.HasError() {
				return diags
			}
		}
		data.Links, diags = types.ListValue(types.ObjectType{AttrTypes: schemas.EnvironmentObjectLinkAttributeTypes()}, linkObjects)
		if diags.HasError() {
			return diags
		}
	} else {
		data.Links = types.ListNull(types.ObjectType{AttrTypes: schemas.EnvironmentObjectLinkAttributeTypes()})
	}

	// Exclude Links
	if obj.ExcludeLinks != nil && len(*obj.ExcludeLinks) > 0 {
		excludeLinkObjects := make([]attr.Value, len(*obj.ExcludeLinks))
		for i, el := range *obj.ExcludeLinks {
			excludeLinkObjects[i], diags = types.ObjectValue(schemas.EnvironmentObjectExcludeLinkAttributeTypes(), map[string]attr.Value{
				"scope":           types.StringValue(string(el.Scope)),
				"scope_entity_id": types.StringValue(el.ScopeEntityId),
			})
			if diags.HasError() {
				return diags
			}
		}
		data.ExcludeLinks, diags = types.ListValue(types.ObjectType{AttrTypes: schemas.EnvironmentObjectExcludeLinkAttributeTypes()}, excludeLinkObjects)
		if diags.HasError() {
			return diags
		}
	} else {
		data.ExcludeLinks = types.ListNull(types.ObjectType{AttrTypes: schemas.EnvironmentObjectExcludeLinkAttributeTypes()})
	}

	return nil
}

func environmentObjectConnectionToObject(ctx context.Context, conn *platform.EnvironmentObjectConnection) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	var authTypeObj types.Object
	if conn.ConnectionAuthType != nil {
		authTypeObj, diags = environmentObjectConnectionAuthTypeToObject(ctx, conn.ConnectionAuthType)
		if diags.HasError() {
			return types.Object{}, diags
		}
	} else {
		authTypeObj = types.ObjectNull(schemas.EnvironmentObjectConnectionAuthTypeAttributeTypes())
	}

	var extraVal types.String
	if conn.Extra != nil {
		extraBytes, err := json.Marshal(conn.Extra)
		if err != nil {
			return types.Object{}, diag.Diagnostics{diag.NewErrorDiagnostic("Internal Error", "Failed to marshal connection extra to JSON")}
		}
		extraVal = types.StringValue(string(extraBytes))
	} else {
		extraVal = types.StringNull()
	}

	var portVal types.Int64
	if conn.Port != nil {
		portVal = types.Int64Value(int64(*conn.Port))
	} else {
		portVal = types.Int64Null()
	}

	return types.ObjectValue(schemas.EnvironmentObjectConnectionAttributeTypes(), map[string]attr.Value{
		"auth_type_id":         types.StringNull(),
		"connection_auth_type": authTypeObj,
		"type":                 types.StringValue(conn.Type),
		"host":                 types.StringPointerValue(conn.Host),
		"port":                 portVal,
		"schema":               types.StringPointerValue(conn.Schema),
		"login":                types.StringPointerValue(conn.Login),
		"password":             types.StringPointerValue(conn.Password),
		"extra":                extraVal,
	})
}

func environmentObjectConnectionAuthTypeToObject(ctx context.Context, cat *platform.ConnectionAuthType) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	paramAttrType := schemas.EnvironmentObjectConnectionAuthTypeParameterAttributeTypes()
	paramObjects := make([]attr.Value, len(cat.Parameters))
	for i, p := range cat.Parameters {
		paramObjects[i], diags = types.ObjectValue(paramAttrType, map[string]attr.Value{
			"airflow_param_name": types.StringValue(p.AirflowParamName),
			"friendly_name":      types.StringValue(p.FriendlyName),
			"data_type":          types.StringValue(p.DataType),
			"is_required":        types.BoolValue(p.IsRequired),
			"is_secret":          types.BoolValue(p.IsSecret),
			"description":        types.StringValue(p.Description),
			"example":            types.StringPointerValue(p.Example),
			"is_in_extra":        types.BoolValue(p.IsInExtra),
		})
		if diags.HasError() {
			return types.Object{}, diags
		}
	}

	paramsList, diags := types.ListValue(types.ObjectType{AttrTypes: paramAttrType}, paramObjects)
	if diags.HasError() {
		return types.Object{}, diags
	}

	return types.ObjectValue(schemas.EnvironmentObjectConnectionAuthTypeAttributeTypes(), map[string]attr.Value{
		"parameters":            paramsList,
		"id":                    types.StringValue(cat.Id),
		"name":                  types.StringValue(cat.Name),
		"auth_method_name":      types.StringValue(cat.AuthMethodName),
		"airflow_type":          types.StringValue(cat.AirflowType),
		"description":           types.StringValue(cat.Description),
		"provider_package_name": types.StringValue(cat.ProviderPackageName),
		"provider_logo":         types.StringPointerValue(cat.ProviderLogo),
		"guide_path":            types.StringPointerValue(cat.GuidePath),
	})
}

func environmentObjectMetricsExportToObject(ctx context.Context, me *platform.EnvironmentObjectMetricsExport) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	var authTypeVal types.String
	if me.AuthType != nil {
		authTypeVal = types.StringValue(string(*me.AuthType))
	} else {
		authTypeVal = types.StringNull()
	}

	headers, diags := stringMapToTFMap(me.Headers)
	if diags.HasError() {
		return types.Object{}, diags
	}

	labels, diags := stringMapToTFMap(me.Labels)
	if diags.HasError() {
		return types.Object{}, diags
	}

	return types.ObjectValue(schemas.EnvironmentObjectMetricsExportAttributeTypes(), map[string]attr.Value{
		"auth_type":     authTypeVal,
		"endpoint":      types.StringValue(me.Endpoint),
		"basic_token":   types.StringPointerValue(me.BasicToken),
		"exporter_type": types.StringValue(string(me.ExporterType)),
		"username":      types.StringPointerValue(me.Username),
		"password":      types.StringPointerValue(me.Password),
		"headers":       headers,
		"labels":        labels,
	})
}

func environmentObjectLinkToObject(ctx context.Context, link *platform.EnvironmentObjectLink) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Airflow variable overrides
	var avOverrides types.Object
	if link.AirflowVariableOverrides != nil {
		avOverrides, diags = types.ObjectValue(schemas.EnvironmentObjectAirflowVariableOverridesAttributeTypes(), map[string]attr.Value{
			"value": types.StringValue(link.AirflowVariableOverrides.Value),
		})
		if diags.HasError() {
			return types.Object{}, diags
		}
	} else {
		avOverrides = types.ObjectNull(schemas.EnvironmentObjectAirflowVariableOverridesAttributeTypes())
	}

	// Connection overrides
	var connOverrides types.Object
	if link.ConnectionOverrides != nil {
		co := link.ConnectionOverrides
		var extraVal types.String
		if co.Extra != nil {
			extraBytes, err := json.Marshal(co.Extra)
			if err != nil {
				return types.Object{}, diag.Diagnostics{diag.NewErrorDiagnostic("Internal Error", "Failed to marshal connection overrides extra")}
			}
			extraVal = types.StringValue(string(extraBytes))
		} else {
			extraVal = types.StringNull()
		}
		var portVal types.Int64
		if co.Port != nil {
			portVal = types.Int64Value(int64(*co.Port))
		} else {
			portVal = types.Int64Null()
		}
		connOverrides, diags = types.ObjectValue(schemas.EnvironmentObjectConnectionOverridesAttributeTypes(), map[string]attr.Value{
			"type":     types.StringPointerValue(co.Type),
			"host":     types.StringPointerValue(co.Host),
			"port":     portVal,
			"schema":   types.StringPointerValue(co.Schema),
			"login":    types.StringPointerValue(co.Login),
			"password": types.StringPointerValue(co.Password),
			"extra":    extraVal,
		})
		if diags.HasError() {
			return types.Object{}, diags
		}
	} else {
		connOverrides = types.ObjectNull(schemas.EnvironmentObjectConnectionOverridesAttributeTypes())
	}

	// Metrics export overrides
	var meOverrides types.Object
	if link.MetricsExportOverrides != nil {
		mo := link.MetricsExportOverrides
		var authTypeVal types.String
		if mo.AuthType != nil {
			authTypeVal = types.StringValue(string(*mo.AuthType))
		} else {
			authTypeVal = types.StringNull()
		}
		var exporterTypeVal types.String
		if mo.ExporterType != nil {
			exporterTypeVal = types.StringValue(string(*mo.ExporterType))
		} else {
			exporterTypeVal = types.StringNull()
		}
		headers, d := stringMapToTFMap(mo.Headers)
		if d.HasError() {
			return types.Object{}, d
		}
		labels, d := stringMapToTFMap(mo.Labels)
		if d.HasError() {
			return types.Object{}, d
		}
		meOverrides, diags = types.ObjectValue(schemas.EnvironmentObjectMetricsExportOverridesAttributeTypes(), map[string]attr.Value{
			"auth_type":     authTypeVal,
			"endpoint":      types.StringPointerValue(mo.Endpoint),
			"basic_token":   types.StringPointerValue(mo.BasicToken),
			"exporter_type": exporterTypeVal,
			"username":      types.StringPointerValue(mo.Username),
			"password":      types.StringPointerValue(mo.Password),
			"headers":       headers,
			"labels":        labels,
		})
		if diags.HasError() {
			return types.Object{}, diags
		}
	} else {
		meOverrides = types.ObjectNull(schemas.EnvironmentObjectMetricsExportOverridesAttributeTypes())
	}

	return types.ObjectValue(schemas.EnvironmentObjectLinkAttributeTypes(), map[string]attr.Value{
		"scope":                      types.StringValue(string(link.Scope)),
		"scope_entity_id":            types.StringValue(link.ScopeEntityId),
		"airflow_variable_overrides": avOverrides,
		"connection_overrides":       connOverrides,
		"metrics_export_overrides":   meOverrides,
	})
}

func stringMapToTFMap(m *map[string]string) (types.Map, diag.Diagnostics) {
	if m == nil || len(*m) == 0 {
		return types.MapNull(types.StringType), nil
	}
	elems := make(map[string]attr.Value, len(*m))
	for k, v := range *m {
		elems[k] = types.StringValue(v)
	}
	return types.MapValue(types.StringType, elems)
}
