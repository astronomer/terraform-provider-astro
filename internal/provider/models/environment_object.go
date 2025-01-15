package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EnvironmentObject struct {
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
