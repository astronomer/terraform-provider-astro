package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EnvironmentObjectDataSource struct {
	Id            types.String `tfsdk:"id"`
	ObjectKey     types.String `tfsdk:"object_key"`
	ObjectType    types.String `tfsdk:"object_type"`
	Scope         types.String `tfsdk:"scope"`
	ScopeEntityId types.String `tfsdk:"scope_entity_id"`
	// Optional fields for CONNECTION type
	Connection types.Object `tfsdk:"connection"`
	// Optional fields for AIRFLOW_VARIABLE type
	AirflowVariable types.Object `tfsdk:"airflow_variable"`
	// Optional fields for WORKSPACE scope
	AutoLinkDeployments types.Bool `tfsdk:"auto_link_deployments"`
	ExcludeLinks        types.Set  `tfsdk:"exclude_links"`
	Links               types.Set  `tfsdk:"links"`
	// Optional fields for METRICS_EXPORT type
	MetricsExport types.Object `tfsdk:"metrics_export"`
}

type EnvironmentObjectResource struct {
	Id            types.String `tfsdk:"id"`
	ObjectKey     types.String `tfsdk:"object_key"`
	ObjectType    types.String `tfsdk:"object_type"`
	Scope         types.String `tfsdk:"scope"`
	ScopeEntityId types.String `tfsdk:"scope_entity_id"`
	// Optional fields for CONNECTION type
	Connection types.Object `tfsdk:"connection"`
	// Optional fields for AIRFLOW_VARIABLE type
	AirflowVariable types.Object `tfsdk:"airflow_variable"`
	// Optional fields for WORKSPACE scope
	AutoLinkDeployments types.Bool `tfsdk:"auto_link_deployments"`
	ExcludeLinks        types.Set  `tfsdk:"exclude_links"`
	Links               types.Set  `tfsdk:"links"`
	// Optional fields for METRICS_EXPORT type
	MetricsExport types.Object `tfsdk:"metrics_export"`
}

func (data *EnvironmentObjectDataSource) ReadFromResponse(ctx context.Context, user *platform.EnvironmentObject) diag.Diagnostics {
	var diags diag.Diagnostics

	if user.Id != nil {
		data.Id = types.StringValue(*user.Id)
	}
	data.ObjectKey = types.StringValue(user.ObjectKey)
	data.ObjectType = types.StringValue(string(user.ObjectType))
	data.Scope = types.StringValue(string(user.Scope))
	data.ScopeEntityId = types.StringValue(user.ScopeEntityId)

	if user.Connection != nil {
		connectionType := map[string]attr.Type{
			"host":     types.StringType,
			"login":    types.StringType,
			"password": types.StringType,
			"port":     types.Int64Type,
			"schema":   types.StringType,
			"type":     types.StringType,
		}
		connectionValue := map[string]attr.Value{
			"host":     types.StringValue(*user.Connection.Host),
			"login":    types.StringValue(*user.Connection.Login),
			"password": types.StringValue(*user.Connection.Password),
			"port":     types.Int64Value(int64(*user.Connection.Port)),
			"schema":   types.StringValue(*user.Connection.Schema),
			"type":     types.StringValue(user.Connection.Type),
		}
		data.Connection, _ = types.ObjectValue(connectionType, connectionValue)
	}
	if user.AirflowVariable != nil {
		afVarType := map[string]attr.Type{
			"is_secret": types.BoolType,
			"value":     types.StringType,
		}
		afVarValue := map[string]attr.Value{
			"is_secret": types.BoolValue(user.AirflowVariable.IsSecret),
			"value":     types.StringValue(user.AirflowVariable.Value),
		}
		data.AirflowVariable, _ = types.ObjectValue(afVarType, afVarValue)
	}
	if user.AutoLinkDeployments != nil {
		data.AutoLinkDeployments = types.BoolValue(*user.AutoLinkDeployments)
	}
	if user.ExcludeLinks != nil {
		excludeLinkType := types.SetType{ElemType: types.StringType}
		excludeLinkValues := make([]attr.Value, len(*user.ExcludeLinks))
		for i, link := range *user.ExcludeLinks {
			excludeLinkValues[i] = types.StringValue(link.ScopeEntityId)
		}
		data.ExcludeLinks, _ = types.SetValue(excludeLinkType, excludeLinkValues)
	}
	if user.Links != nil {
		linkType := types.SetType{ElemType: types.StringType}
		linkValues := make([]attr.Value, len(*user.Links))
		for i, link := range *user.Links {
			linkValues[i] = types.StringValue(link.ScopeEntityId)
		}
		data.Links, _ = types.SetValue(linkType, linkValues)
	}
	if user.MetricsExport != nil {
		metricsExportType := map[string]attr.Type{
			"endpoint":      types.StringType,
			"exporter_type": types.StringType,
		}
		metricsExportValue := map[string]attr.Value{
			"endpoint":      types.StringValue(user.MetricsExport.Endpoint),
			"exporter_type": types.StringValue(string(user.MetricsExport.ExporterType)),
		}
		data.MetricsExport, _ = types.ObjectValue(metricsExportType, metricsExportValue)
	}

	return diags
}

func (data *EnvironmentObjectResource) ReadFromResponse(ctx context.Context, user *platform.EnvironmentObject) diag.Diagnostics {
	var diags diag.Diagnostics

	if user.Id != nil {
		data.Id = types.StringValue(*user.Id)
	}
	data.ObjectKey = types.StringValue(user.ObjectKey)
	data.ObjectType = types.StringValue(string(user.ObjectType))
	data.Scope = types.StringValue(string(user.Scope))
	data.ScopeEntityId = types.StringValue(user.ScopeEntityId)

	if user.Connection != nil {
		connectionType := map[string]attr.Type{
			"host":     types.StringType,
			"login":    types.StringType,
			"password": types.StringType,
			"port":     types.Int64Type,
			"schema":   types.StringType,
			"type":     types.StringType,
		}
		connectionValue := map[string]attr.Value{
			"host":     types.StringValue(*user.Connection.Host),
			"login":    types.StringValue(*user.Connection.Login),
			"password": types.StringValue(*user.Connection.Password),
			"port":     types.Int64Value(int64(*user.Connection.Port)),
			"schema":   types.StringValue(*user.Connection.Schema),
			"type":     types.StringValue(user.Connection.Type),
		}
		data.Connection, _ = types.ObjectValue(connectionType, connectionValue)
	}
	if user.AirflowVariable != nil {
		afVarType := map[string]attr.Type{
			"is_secret": types.BoolType,
			"value":     types.StringType,
		}
		afVarValue := map[string]attr.Value{
			"is_secret": types.BoolValue(user.AirflowVariable.IsSecret),
			"value":     types.StringValue(user.AirflowVariable.Value),
		}
		data.AirflowVariable, _ = types.ObjectValue(afVarType, afVarValue)
	}
	if user.AutoLinkDeployments != nil {
		data.AutoLinkDeployments = types.BoolValue(*user.AutoLinkDeployments)
	}
	if user.ExcludeLinks != nil {
		excludeLinkType := types.SetType{ElemType: types.StringType}
		excludeLinkValues := make([]attr.Value, len(*user.ExcludeLinks))
		for i, link := range *user.ExcludeLinks {
			excludeLinkValues[i] = types.StringValue(link.ScopeEntityId)
		}
		data.ExcludeLinks, _ = types.SetValue(excludeLinkType, excludeLinkValues)
	}
	if user.Links != nil {
		linkType := types.SetType{ElemType: types.StringType}
		linkValues := make([]attr.Value, len(*user.Links))
		for i, link := range *user.Links {
			linkValues[i] = types.StringValue(link.ScopeEntityId)
		}
		data.Links, _ = types.SetValue(linkType, linkValues)
	}
	if user.MetricsExport != nil {
		metricsExportType := map[string]attr.Type{
			"endpoint":      types.StringType,
			"exporter_type": types.StringType,
		}
		metricsExportValue := map[string]attr.Value{
			"endpoint":      types.StringValue(user.MetricsExport.Endpoint),
			"exporter_type": types.StringValue(string(user.MetricsExport.ExporterType)),
		}
		data.MetricsExport, _ = types.ObjectValue(metricsExportType, metricsExportValue)
	}

	return diags
}
