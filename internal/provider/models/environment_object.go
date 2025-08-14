package models

import (
	"context"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// EnvironmentObject describes the data source data model.
type EnvironmentObject struct {
	Id                  types.String `tfsdk:"id"`
	ObjectKey           types.String `tfsdk:"object_key"`
	ScopeEntityId       types.String `tfsdk:"scope_entity_id"`
	SourceScopeEntityId types.String `tfsdk:"source_scope_entity_id"`
	AutoLinkDeployments types.Bool   `tfsdk:"auto_link_deployments"`
	AirflowVariable     types.Object `tfsdk:"airflow_variable"`
	Connection          types.Object `tfsdk:"connection"`
	ExcludeLinks        types.Object `tfsdk:"exclude_links"`
	Links               types.Object `tfsdk:"links"`
	MetricsExport       types.Object `tfsdk:"metrics_export"`
	ObjectType          types.String `tfsdk:"object_type"`
	Scope               types.String `tfsdk:"scope"`
	SourceScope         types.String `tfsdk:"source_scope"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
	CreatedBy           types.Object `tfsdk:"created_by"`
	UpdatedBy           types.Object `tfsdk:"updated_by"`
}

func (data *EnvironmentObject) ReadFromResponse(ctx context.Context, EnvironmentObject *platform.EnvironmentObject) diag.Diagnostics {
	var diags diag.Diagnostics
	data.Id = types.StringValue(*EnvironmentObject.Id)
	data.ObjectKey = types.StringValue(EnvironmentObject.ObjectKey)
	data.ScopeEntityId = types.StringValue(EnvironmentObject.ScopeEntityId)
	data.SourceScopeEntityId = types.StringValue(*EnvironmentObject.SourceScopeEntityId)
	data.AutoLinkDeployments = types.BoolValue(*EnvironmentObject.AutoLinkDeployments)

	data.AirflowVariable, diags = EnvironmentObjectAirflowVariableObject(ctx, EnvironmentObject.AirflowVariable)
	if diags.HasError() {
		return diags
	}
	data.Connection, diags = EnvironmentObjectConnectionObject(ctx, EnvironmentObject.Connection)
	if diags.HasError() {
		return diags
	}
	data.ExcludeLinks, diags = EnvironmentObjectExcludeLinksObject(ctx, EnvironmentObject.ExcludeLinks)
	if diags.HasError() {
		return diags
	}
	data.Links, diags = EnvironmentObjectLinksObject(ctx, EnvironmentObject.Links)
	if diags.HasError() {
		return diags
	}
	data.MetricsExport, diags = EnvironmentObjectMetricsExportObject(ctx, EnvironmentObject.MetricsExport)
	if diags.HasError() {
		return diags
	}

	data.ObjectType = types.StringValue(string(EnvironmentObject.ObjectType))
	data.Scope = types.StringValue(string(EnvironmentObject.Scope))
	data.SourceScope = types.StringValue(string(*EnvironmentObject.SourceScope))

	data.CreatedAt = types.StringValue(*EnvironmentObject.CreatedAt)
	data.UpdatedAt = types.StringValue(*EnvironmentObject.UpdatedAt)
	data.CreatedBy, diags = SubjectProfileTypesObject(ctx, EnvironmentObject.CreatedBy)
	if diags.HasError() {
		return diags
	}
	data.UpdatedBy, diags = SubjectProfileTypesObject(ctx, EnvironmentObject.UpdatedBy)
	if diags.HasError() {
		return diags
	}

	return nil
}

func EnvironmentObjectLinksObject(
	ctx context.Context,
	environmentObjectLinks any,
) (types.Object, diag.Diagnostics) {
	var environmentObjectLinksPtr *platform.EnvironmentObjectLink

	switch v := environmentObjectLinks.(type) {
	case platform.EnvironmentObjectLink:
		environmentObjectLinksPtr = &v
	case *platform.EnvironmentObjectLink:
		environmentObjectLinksPtr = v
	default:
		tflog.Error(
			ctx,
			"Unexpected type passed into environmentObjectLinks",
			map[string]interface{}{"value": environmentObjectLinks},
		)
		return types.Object{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Internal Error",
				"EnvironmentObjectLinkObject expects a platform.EnvironmentObjectLink type but did not receive one",
			),
		}
	}

	connectionObj, diags := EnvironmentObjectLinksConnectionOverridesObject(ctx, environmentObjectLinksPtr.ConnectionOverrides)
	if diags.HasError() {
		return types.Object{}, diags
	}
	connection, diags := types.ObjectValueFrom(ctx, schemas.EnvironmentObjectLinksConnectionOverridesAttributeTypes(), connectionObj)
	if diags.HasError() {
		return types.Object{}, diags
	}

	airflowVariableObj, diags := EnvironmentObjectLinksAirflowVariableOverridesObject(ctx, environmentObjectLinksPtr.AirflowVariableOverrides)
	if diags.HasError() {
		return types.Object{}, diags
	}
	airflowVariable, diags := types.ObjectValueFrom(ctx, schemas.EnvironmentObjectLinksAirflowVariableOverridesAttributeTypes(), airflowVariableObj)
	if diags.HasError() {
		return types.Object{}, diags
	}

	metricsExportObj, diags := EnvironmentObjectLinksMetricsExportOverridesObject(ctx, environmentObjectLinksPtr.MetricsExportOverrides)
	if diags.HasError() {
		return types.Object{}, diags
	}
	metricsExport, diags := types.ObjectValueFrom(ctx, schemas.EnvironmentObjectLinksMetricsExportOverridesAttributeTypes(), metricsExportObj)
	if diags.HasError() {
		return types.Object{}, diags
	}

	return types.ObjectValue(schemas.EnvironmentObjectLinksAttributeTypes(), map[string]attr.Value{
		"scope":            types.StringValue(string(environmentObjectLinksPtr.Scope)),
		"scope_entity_id":  types.StringValue(environmentObjectLinksPtr.ScopeEntityId),
		"connection":       connection,
		"airflow_variable": airflowVariable,
		"metrics_export":   metricsExport,
	})
}

func EnvironmentObjectMetricsExportObject(
	ctx context.Context,
	metricsExportObject any,
) (types.Object, diag.Diagnostics) {
	var metricsExportPtr *platform.EnvironmentObjectMetricsExport

	switch v := metricsExportObject.(type) {
	case platform.EnvironmentObjectMetricsExport:
		metricsExportPtr = &v
	case *platform.EnvironmentObjectMetricsExport:
		metricsExportPtr = v
	default:
		tflog.Error(
			ctx,
			"Unexpected type passed into metricsExportObject",
			map[string]interface{}{"value": metricsExportObject},
		)
		return types.Object{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Internal Error",
				"MetricsExportObject expects a platform.EnvironmentObjectMetricsExport type but did not receive one",
			),
		}
	}
	headersMap := make(map[string]attr.Value, len(*metricsExportPtr.Headers))
	for s, s2 := range *metricsExportPtr.Headers {
		headersMap[s] = types.StringValue(s2)
	}
	headers, diags := types.MapValue(types.StringType, headersMap)
	if diags.HasError() {
		return types.Object{}, diags
	}

	labelsMap := make(map[string]attr.Value, len(*metricsExportPtr.Headers))
	for s, s2 := range *metricsExportPtr.Headers {
		labelsMap[s] = types.StringValue(s2)
	}
	labels, diags := types.MapValue(types.StringType, labelsMap)
	if diags.HasError() {
		return types.Object{}, diags
	}

	return types.ObjectValue(schemas.EnvironmentObjectMetricsExportAttributeTypes(), map[string]attr.Value{
		"auth_type":     types.StringValue(string(*metricsExportPtr.AuthType)),
		"endpoint":      types.StringValue(metricsExportPtr.Endpoint),
		"basic_token":   types.StringPointerValue(metricsExportPtr.BasicToken),
		"exporter_type": types.StringValue(string(metricsExportPtr.ExporterType)),
		"username":      types.StringPointerValue(metricsExportPtr.Username),
		"password":      types.StringPointerValue(metricsExportPtr.Password),
		"headers":       headers,
		"labels":        labels,
	})
}

func EnvironmentObjectLinksMetricsExportOverridesObject(
	ctx context.Context,
	metricsExportOverridesObject any,
) (types.Object, diag.Diagnostics) {
	var metricsExportOverridesPtr *platform.EnvironmentObjectMetricsExportOverrides

	switch v := metricsExportOverridesObject.(type) {
	case platform.EnvironmentObjectMetricsExportOverrides:
		metricsExportOverridesPtr = &v
	case *platform.EnvironmentObjectMetricsExportOverrides:
		metricsExportOverridesPtr = v
	default:
		tflog.Error(
			ctx,
			"Unexpected type passed into metricsExportOverridesObject",
			map[string]interface{}{"value": metricsExportOverridesObject},
		)
		return types.Object{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Internal Error",
				"MetricsExportObject expects a platform.EnvironmentObjectMetricsExport type but did not receive one",
			),
		}
	}
	headersMap := make(map[string]attr.Value, len(*metricsExportOverridesPtr.Headers))
	for s, s2 := range *metricsExportOverridesPtr.Headers {
		headersMap[s] = types.StringValue(s2)
	}
	headers, diags := types.MapValue(types.StringType, headersMap)
	if diags.HasError() {
		return types.Object{}, diags
	}

	labelsMap := make(map[string]attr.Value, len(*metricsExportOverridesPtr.Headers))
	for s, s2 := range *metricsExportOverridesPtr.Headers {
		labelsMap[s] = types.StringValue(s2)
	}
	labels, diags := types.MapValue(types.StringType, labelsMap)
	if diags.HasError() {
		return types.Object{}, diags
	}

	return types.ObjectValue(schemas.EnvironmentObjectLinksMetricsExportOverridesAttributeTypes(), map[string]attr.Value{
		"auth_type":     types.StringValue(string(*metricsExportOverridesPtr.AuthType)),
		"endpoint":      types.StringPointerValue(metricsExportOverridesPtr.Endpoint),
		"basic_token":   types.StringPointerValue(metricsExportOverridesPtr.BasicToken),
		"exporter_type": types.StringValue(string(*metricsExportOverridesPtr.ExporterType)),
		"username":      types.StringPointerValue(metricsExportOverridesPtr.Username),
		"password":      types.StringPointerValue(metricsExportOverridesPtr.Password),
		"headers":       headers,
		"labels":        labels,
	})
}

func EnvironmentObjectLinksAirflowVariableOverridesObject(
	ctx context.Context,
	airflowVariableOverridesObject any,
) (types.Object, diag.Diagnostics) {
	var airflowVariableOverridesPtr *platform.EnvironmentObjectAirflowVariableOverrides

	switch v := airflowVariableOverridesObject.(type) {
	case platform.EnvironmentObjectAirflowVariableOverrides:
		airflowVariableOverridesPtr = &v
	case *platform.EnvironmentObjectAirflowVariableOverrides:
		airflowVariableOverridesPtr = v
	default:
		tflog.Error(
			ctx,
			"Unexpected type passed into airflowVariableOverridesObject",
			map[string]interface{}{"value": airflowVariableOverridesObject},
		)
		return types.Object{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Internal Error",
				"AirflowVariableOverridesObject expects a platform.EnvironmentObjectAirflowVariableOverrides type but did not receive one",
			),
		}
	}

	return types.ObjectValue(schemas.EnvironmentObjectLinksAirflowVariableOverridesAttributeTypes(), map[string]attr.Value{
		"value": types.StringValue(airflowVariableOverridesPtr.Value),
	})
}

func EnvironmentObjectAirflowVariableObject(
	ctx context.Context,
	airflowVariableObject any,
) (types.Object, diag.Diagnostics) {
	var airflowVariablePtr *platform.EnvironmentObjectAirflowVariable

	switch v := airflowVariableObject.(type) {
	case platform.EnvironmentObjectAirflowVariable:
		airflowVariablePtr = &v
	case *platform.EnvironmentObjectAirflowVariable:
		airflowVariablePtr = v
	default:
		tflog.Error(
			ctx,
			"Unexpected type passed into airflowVariableObject",
			map[string]interface{}{"value": airflowVariableObject},
		)
		return types.Object{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Internal Error",
				"AirflowVariableObject expects a platform.EnvironmentObjectAirflowVariable type but did not receive one",
			),
		}
	}

	return types.ObjectValue(schemas.EnvironmentObjectAirflowVariableAttributeTypes(), map[string]attr.Value{
		"value":     types.StringValue(airflowVariablePtr.Value),
		"is_secret": types.BoolValue(airflowVariablePtr.IsSecret),
	})
}

func EnvironmentObjectLinksConnectionOverridesObject(
	ctx context.Context,
	connectionOverridesObject any,
) (types.Object, diag.Diagnostics) {
	var connectionOverridesPtr *platform.EnvironmentObjectConnectionOverrides

	switch v := connectionOverridesObject.(type) {
	case platform.EnvironmentObjectConnectionOverrides:
		connectionOverridesPtr = &v
	case *platform.EnvironmentObjectConnectionOverrides:
		connectionOverridesPtr = v
	default:
		tflog.Error(
			ctx,
			"Unexpected type passed into connectionOverridesObject",
			map[string]interface{}{"value": connectionOverridesObject},
		)
		return types.Object{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Internal Error",
				"ConnectionOverrideObject expects a platform.EnvironmentObjectConnectionOverrides type but did not receive one",
			),
		}
	}

	return types.ObjectValue(schemas.EnvironmentObjectLinksConnectionOverridesAttributeTypes(), map[string]attr.Value{
		"type":     types.StringPointerValue(connectionOverridesPtr.Type),
		"host":     types.StringPointerValue(connectionOverridesPtr.Host),
		"login":    types.StringPointerValue(connectionOverridesPtr.Login),
		"password": types.StringPointerValue(connectionOverridesPtr.Password),
		"port":     types.Int64Value(int64(*connectionOverridesPtr.Port)),
		"schema":   types.StringPointerValue(connectionOverridesPtr.Schema),
		//"extra":    types.Object{}(connectionOverridesPtr.Extra), // TODO not sure how to fix extra as it uses interface and terraform is strongly typed
	})
}

func EnvironmentObjectConnectionObject(
	ctx context.Context,
	connectionObject any,
) (types.Object, diag.Diagnostics) {
	var connectionPtr *platform.EnvironmentObjectConnection

	switch v := connectionObject.(type) {
	case platform.EnvironmentObjectConnection:
		connectionPtr = &v
	case *platform.EnvironmentObjectConnection:
		connectionPtr = v
	default:
		tflog.Error(
			ctx,
			"Unexpected type passed into connectionObject",
			map[string]interface{}{"value": connectionObject},
		)
		return types.Object{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Internal Error",
				"ConnectionObject expects a platform.EnvironmentObjectConnection type but did not receive one",
			),
		}
	}

	obj, diags := EnvironmentObjectConnectionAuthTypeObject(ctx, connectionPtr.ConnectionAuthType)
	if diags.HasError() {
		return types.Object{}, diags
	}
	attrObject, diags := types.ObjectValueFrom(ctx, schemas.EnvironmentObjectConnectionAuthTypeAttributeTypes(), obj)
	if diags.HasError() {
		return types.Object{}, diags
	}

	return types.ObjectValue(schemas.EnvironmentObjectConnectionAttributeTypes(), map[string]attr.Value{
		"connection_auth_type": attrObject,
		"host":                 types.StringPointerValue(connectionPtr.Host),
		"login":                types.StringPointerValue(connectionPtr.Login),
		"password":             types.StringPointerValue(connectionPtr.Password),
		"port":                 types.Int64Value(int64(*connectionPtr.Port)),
		"schema":               types.StringPointerValue(connectionPtr.Schema),
		"type":                 types.StringValue(connectionPtr.Type),
	})
}

func EnvironmentObjectExcludeLinksObject(
	ctx context.Context,
	excludeLinks any,
) (types.Object, diag.Diagnostics) {
	var excludeLinkPtr *platform.EnvironmentObjectExcludeLink

	switch v := excludeLinks.(type) {
	case platform.EnvironmentObjectExcludeLink:
		excludeLinkPtr = &v
	case *platform.EnvironmentObjectExcludeLink:
		excludeLinkPtr = v
	default:
		tflog.Error(
			ctx,
			"Unexpected type passed into excludeLinks",
			map[string]interface{}{"value": excludeLinks},
		)
		return types.Object{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Internal Error",
				"ExcludeLinkObject expects a platform.EnvironmentObjectExcludeLink type but did not receive one",
			),
		}
	}

	return types.ObjectValue(schemas.EnvironmentObjectExcludeLinkAttributeTypes(), map[string]attr.Value{
		"scope":           types.StringValue(string(excludeLinkPtr.Scope)),
		"scope_entity_id": types.StringValue(excludeLinkPtr.ScopeEntityId),
	})
}

func EnvironmentObjectConnectionAuthTypeObject(
	ctx context.Context,
	connectionAuthTypeObject any,
) (types.Object, diag.Diagnostics) {
	var connectionAuthTypePtr *platform.ConnectionAuthType

	switch v := connectionAuthTypeObject.(type) {
	case platform.ConnectionAuthType:
		connectionAuthTypePtr = &v
	case *platform.ConnectionAuthType:
		connectionAuthTypePtr = v
	default:
		tflog.Error(
			ctx,
			"Unexpected type passed into connectionAuthTypeObject",
			map[string]interface{}{"value": connectionAuthTypeObject},
		)
		return types.Object{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Internal Error",
				"ConnectionAuthTypeObject expects a platform.ConnectionAuthType type but did not receive one",
			),
		}
	}

	obj, diags := EnvironmentObjectConnectionAuthTypeParametersObject(ctx, connectionAuthTypePtr.Parameters)
	if diags.HasError() {
		return types.Object{}, diags
	}
	paramsObject, diags := types.ObjectValueFrom(ctx, schemas.EnvironmentObjectConnectionAuthTypeParametersAttributeTypes(), obj)
	if diags.HasError() {
		return types.Object{}, diags
	}

	return types.ObjectValue(schemas.EnvironmentObjectConnectionAuthTypeAttributeTypes(), map[string]attr.Value{
		"parameters":            paramsObject,
		"id":                    types.StringValue(connectionAuthTypePtr.Id),
		"name":                  types.StringValue(connectionAuthTypePtr.Name),
		"auth_method_name":      types.StringValue(connectionAuthTypePtr.AuthMethodName),
		"airflow_type":          types.StringValue(connectionAuthTypePtr.AirflowType),
		"description":           types.StringValue(connectionAuthTypePtr.Description),
		"provider_package_name": types.StringValue(connectionAuthTypePtr.ProviderPackageName),
		"provider_logo":         types.StringPointerValue(connectionAuthTypePtr.ProviderLogo),
		"guide_path":            types.StringPointerValue(connectionAuthTypePtr.GuidePath),
	})
}

func EnvironmentObjectConnectionAuthTypeParametersObject(
	ctx context.Context,
	connectionAuthTypeParametersObject any,
) (types.Object, diag.Diagnostics) {
	var connectionAuthTypeParametersPtr *platform.ConnectionAuthTypeParameter

	switch v := connectionAuthTypeParametersObject.(type) {
	case platform.ConnectionAuthTypeParameter:
		connectionAuthTypeParametersPtr = &v
	case *platform.ConnectionAuthTypeParameter:
		connectionAuthTypeParametersPtr = v
	default:
		tflog.Error(
			ctx,
			"Unexpected type passed into connectionAuthTypeParametersObject",
			map[string]interface{}{"value": connectionAuthTypeParametersObject},
		)
		return types.Object{}, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Internal Error",
				"connectionAuthTypeParametersObject expects a platform.ConnectionAuthTypeParameter type but did not receive one",
			),
		}
	}

	return types.ObjectValue(schemas.EnvironmentObjectConnectionAuthTypeAttributeTypes(), map[string]attr.Value{
		"airflow_param_name": types.StringValue(connectionAuthTypeParametersPtr.AirflowParamName),
		"friendly_name":      types.StringValue(connectionAuthTypeParametersPtr.FriendlyName),
		"data_type":          types.StringValue(connectionAuthTypeParametersPtr.DataType),
		"is_required":        types.BoolValue(connectionAuthTypeParametersPtr.IsRequired),
		"is_secret":          types.BoolValue(connectionAuthTypeParametersPtr.IsSecret),
		"description":        types.StringValue(connectionAuthTypeParametersPtr.Description),
		"example":            types.StringPointerValue(connectionAuthTypeParametersPtr.Example),
		"is_in_extra":        types.BoolValue(connectionAuthTypeParametersPtr.IsInExtra),
	})
}
