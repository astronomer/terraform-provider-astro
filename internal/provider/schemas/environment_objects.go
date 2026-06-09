package schemas

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// EnvironmentObjectsElementAttributeTypes returns the attribute type map for a
// single element of the environment_objects list, matching the sibling pattern
// (DeploymentsElementAttributeTypes, AlertsElementAttributeTypes, etc.).
func EnvironmentObjectsElementAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                     types.StringType,
		"object_key":             types.StringType,
		"object_type":            types.StringType,
		"scope":                  types.StringType,
		"scope_entity_id":        types.StringType,
		"source_scope":           types.StringType,
		"source_scope_entity_id": types.StringType,
		"auto_link_deployments":  types.BoolType,
		"airflow_variable":       types.ObjectType{AttrTypes: EnvironmentObjectAirflowVariableAttributeTypes()},
		"connection_config":      types.ObjectType{AttrTypes: EnvironmentObjectConnectionAttributeTypes()},
		"metrics_export":         types.ObjectType{AttrTypes: EnvironmentObjectMetricsExportAttributeTypes()},
		"links":                  types.SetType{ElemType: types.ObjectType{AttrTypes: EnvironmentObjectLinkAttributeTypes()}},
		"exclude_links":          types.SetType{ElemType: types.ObjectType{AttrTypes: EnvironmentObjectExcludeLinkAttributeTypes()}},
		"created_at":             types.StringType,
		"updated_at":             types.StringType,
		"created_by":             types.ObjectType{AttrTypes: SubjectProfileAttributeTypes()},
		"updated_by":             types.ObjectType{AttrTypes: SubjectProfileAttributeTypes()},
	}
}

func EnvironmentObjectsDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"workspace_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Filter by Workspace ID",
			Optional:            true,
		},
		"deployment_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Filter by Deployment ID",
			Optional:            true,
		},
		"object_type": datasourceSchema.StringAttribute{
			MarkdownDescription: "Filter by object type (AIRFLOW_VARIABLE, CONNECTION, METRICS_EXPORT)",
			Optional:            true,
		},
		"object_key": datasourceSchema.StringAttribute{
			MarkdownDescription: "Filter by object key",
			Optional:            true,
		},
		"show_secrets": datasourceSchema.BoolAttribute{
			MarkdownDescription: "If true, returns the actual values of secret fields in the response",
			Optional:            true,
		},
		"resolve_linked": datasourceSchema.BoolAttribute{
			MarkdownDescription: "If true, resolves and returns environment objects linked to the specified Deployment or Workspace",
			Optional:            true,
		},
		"environment_objects": datasourceSchema.SetNestedAttribute{
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: EnvironmentObjectDataSourceSchemaAttributes(),
			},
			Computed: true,
		},
	}
}
