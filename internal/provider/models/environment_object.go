package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EnvironmentObjectConnection struct {
	Host     types.String `tfsdk:"host"`
	Login    types.String `tfsdk:"login"`
	Password types.String `tfsdk:"password"`
	Port     types.Int64  `tfsdk:"port"`
	Schema   types.String `tfsdk:"schema"`
	Type     types.String `tfsdk:"type"`
}

type EnvironmentObjectAirflowVariable struct {
	IsSecret types.Bool   `tfsdk:"is_secret"`
	Value    types.String `tfsdk:"value"`
}

type EnvironmentObjectMetricsExport struct {
	Endpoint     types.String `tfsdk:"endpoint"`
	ExporterType types.String `tfsdk:"exporter_type"`
}

type EnvironmentObjectExcludeLink struct {
	ScopeEntityId types.String `tfsdk:"scope_entity_id"`
	Scope         types.String `tfsdk:"scope"`
}

type EnvironmentObjectResource struct {
	Id            types.String `tfsdk:"id"`
	ObjectKey     types.String `tfsdk:"object_key"`
	ObjectType    types.String `tfsdk:"object_type"`
	Scope         types.String `tfsdk:"scope"`
	ScopeEntityId types.String `tfsdk:"scope_entity_id"`
	// Optional fields for CONNECTION, AIRFLOW_VARIABLE, and METRICS_EXPORT type
	Connection      *EnvironmentObjectConnection      `tfsdk:"airflow_connection"`
	AirflowVariable *EnvironmentObjectAirflowVariable `tfsdk:"airflow_variable"`
	MetricsExport   *EnvironmentObjectMetricsExport   `tfsdk:"metrics_export"`
	// Optional fields for WORKSPACE scope
	AutoLinkDeployments types.Bool `tfsdk:"auto_link_deployments"`
	// TODO: Handle ExcludeLinks and Links later
	// ExcludeLinks        types.Set  `tfsdk:"exclude_links"`
	// Links               types.Set  `tfsdk:"links"`
}

func (data *EnvironmentObjectResource) ReadFromResponse(ctx context.Context, envObject *platform.EnvironmentObject) diag.Diagnostics {
	var diags diag.Diagnostics

	// Id is not returned for resolved environment objects
	if envObject.Id != nil {
		data.Id = types.StringValue(*envObject.Id)
	}
	data.ObjectKey = types.StringValue(envObject.ObjectKey)
	data.ObjectType = types.StringValue(string(envObject.ObjectType))
	data.Scope = types.StringValue(string(envObject.Scope))
	data.ScopeEntityId = types.StringValue(envObject.ScopeEntityId)

	// Set one of the optional fields
	switch envObject.ObjectType {
	case platform.EnvironmentObjectObjectTypeCONNECTION:
		if envObject.Connection != nil {
			data.Connection = &EnvironmentObjectConnection{
				Host:     types.StringValue(*envObject.Connection.Host),
				Login:    types.StringValue(*envObject.Connection.Login),
				Password: types.StringValue(*envObject.Connection.Password),
				Port:     types.Int64Value(int64(*envObject.Connection.Port)),
				Schema:   types.StringValue(*envObject.Connection.Schema),
				Type:     types.StringValue(envObject.Connection.Type),
			}
		}
	case platform.EnvironmentObjectObjectTypeAIRFLOWVARIABLE:
		if envObject.AirflowVariable != nil {
			data.AirflowVariable = &EnvironmentObjectAirflowVariable{
				IsSecret: types.BoolValue(envObject.AirflowVariable.IsSecret),
				Value:    types.StringValue(envObject.AirflowVariable.Value),
			}
		}
	case platform.EnvironmentObjectObjectTypeMETRICSEXPORT:
		if envObject.MetricsExport != nil {
			data.MetricsExport = &EnvironmentObjectMetricsExport{
				Endpoint:     types.StringValue(envObject.MetricsExport.Endpoint),
				ExporterType: types.StringValue(string(envObject.MetricsExport.ExporterType)),
			}
		}
	}

	if envObject.AutoLinkDeployments != nil {
		data.AutoLinkDeployments = types.BoolValue(*envObject.AutoLinkDeployments)
	}

	return diags
}
