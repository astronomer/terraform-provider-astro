package schemas

import (
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

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
		"environment_objects": datasourceSchema.ListNestedAttribute{
			MarkdownDescription: "List of environment objects",
			Computed:            true,
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: EnvironmentObjectDataSourceSchemaAttributes(),
			},
		},
	}
}
