package schemas

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func EnvironmentObjectDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Environment Object identifier",
			Required:            true,
		},
		"object_key": datasourceSchema.StringAttribute{
			MarkdownDescription: "Key of the environment object",
			Computed:            true,
		},
		"object_type": datasourceSchema.StringAttribute{
			MarkdownDescription: "Type of the environment object",
			Computed:            true,
		},
		"scope": datasourceSchema.StringAttribute{
			MarkdownDescription: "Scope of the environment object",
			Computed:            true,
		},
		"scope_entity_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Scope entity ID of the environment object",
			Computed:            true,
		},
		"connection": datasourceSchema.ObjectAttribute{
			MarkdownDescription: "Connection details for the environment object",
			Computed:            true,
			AttributeTypes: map[string]attr.Type{
				"host":     types.StringType,
				"login":    types.StringType,
				"password": types.StringType,
				"port":     types.Int64Type,
				"schema":   types.StringType,
				"type":     types.StringType,
			},
		},
		"airflow_variable": datasourceSchema.ObjectAttribute{
			MarkdownDescription: "Airflow variable details for the environment object",
			Computed:            true,
			AttributeTypes: map[string]attr.Type{
				"is_secret": types.BoolType,
				"value":     types.StringType,
			},
		},
		"auto_link_deployments": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Auto link deployments flag",
			Computed:            true,
		},
		"exclude_links": datasourceSchema.SetAttribute{
			MarkdownDescription: "Links to exclude",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"links": datasourceSchema.SetAttribute{
			MarkdownDescription: "Links associated with the environment object",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"metrics_export": datasourceSchema.ObjectAttribute{
			MarkdownDescription: "Metrics export details for the environment object",
			Computed:            true,
			AttributeTypes: map[string]attr.Type{
				"endpoint":      types.StringType,
				"exporter_type": types.StringType,
			},
		},
	}
}

func EnvironmentObjectResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Environment Object ID",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"object_key": resourceSchema.StringAttribute{
			MarkdownDescription: "Key of the environment object",
			Required:            true,
		},
		"object_type": resourceSchema.StringAttribute{
			MarkdownDescription: "Type of the environment object",
			Required:            true,
		},
		"scope": resourceSchema.StringAttribute{
			MarkdownDescription: "Scope of the environment object",
			Required:            true,
		},
		"scope_entity_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Scope entity ID of the environment object",
			Optional:            true,
		},
		"connection": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Connection details for the environment object",
			Optional:            true,
			Attributes: map[string]resourceSchema.Attribute{
				"host": resourceSchema.StringAttribute{
					MarkdownDescription: "Host of the connection",
					Required:            true,
				},
				"login": resourceSchema.StringAttribute{
					MarkdownDescription: "Login for the connection",
					Required:            true,
				},
				"password": resourceSchema.StringAttribute{
					MarkdownDescription: "Password for the connection",
					Required:            true,
					Sensitive:           true,
				},
				"port": resourceSchema.Int64Attribute{
					MarkdownDescription: "Port for the connection",
					Required:            true,
				},
				"schema": resourceSchema.StringAttribute{
					MarkdownDescription: "Schema for the connection",
					Optional:            true,
				},
				"type": resourceSchema.StringAttribute{
					MarkdownDescription: "Type of the connection",
					Required:            true,
				},
			},
		},
		"airflow_variable": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Airflow variable details for the environment object",
			Optional:            true,
			Attributes: map[string]resourceSchema.Attribute{
				"is_secret": resourceSchema.BoolAttribute{
					MarkdownDescription: "Whether the variable is secret",
					Required:            true,
				},
				"value": resourceSchema.StringAttribute{
					MarkdownDescription: "Value of the variable",
					Required:            true,
					Sensitive:           true,
				},
			},
		},
		"auto_link_deployments": resourceSchema.BoolAttribute{
			MarkdownDescription: "Auto link deployments flag",
			Optional:            true,
		},
		"exclude_links": resourceSchema.SetAttribute{
			MarkdownDescription: "Links to exclude",
			Optional:            true,
			ElementType:         types.StringType,
		},
		"links": resourceSchema.SetAttribute{
			MarkdownDescription: "Links associated with the environment object",
			Optional:            true,
			ElementType:         types.StringType,
		},
		"metrics_export": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Metrics export details for the environment object",
			Optional:            true,
			Attributes: map[string]resourceSchema.Attribute{
				"endpoint": resourceSchema.StringAttribute{
					MarkdownDescription: "Endpoint for metrics export",
					Required:            true,
				},
				"exporter_type": resourceSchema.StringAttribute{
					MarkdownDescription: "Type of the exporter",
					Required:            true,
				},
			},
		},
	}
}
