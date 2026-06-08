package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- Attribute Types (used by models for types.ObjectValue / types.ListValueFrom) ---

func EnvironmentObjectAirflowVariableAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"value":     types.StringType,
		"is_secret": types.BoolType,
	}
}

func EnvironmentObjectMetricsExportAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"auth_type":     types.StringType,
		"endpoint":      types.StringType,
		"basic_token":   types.StringType,
		"exporter_type": types.StringType,
		"username":      types.StringType,
		"password":      types.StringType,
		"headers":       types.MapType{ElemType: types.StringType},
		"labels":        types.MapType{ElemType: types.StringType},
	}
}

func EnvironmentObjectConnectionAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"auth_type_id": types.StringType,
		"connection_auth_type": types.ObjectType{
			AttrTypes: EnvironmentObjectConnectionAuthTypeAttributeTypes(),
		},
		"type":     types.StringType,
		"host":     types.StringType,
		"port":     types.Int64Type,
		"schema":   types.StringType,
		"login":    types.StringType,
		"password": types.StringType,
		"extra":    types.StringType,
	}
}

func EnvironmentObjectConnectionAuthTypeAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"parameters": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: EnvironmentObjectConnectionAuthTypeParameterAttributeTypes(),
			},
		},
		"id":                    types.StringType,
		"name":                  types.StringType,
		"auth_method_name":      types.StringType,
		"airflow_type":          types.StringType,
		"description":           types.StringType,
		"provider_package_name": types.StringType,
		"provider_logo":         types.StringType,
		"guide_path":            types.StringType,
	}
}

func EnvironmentObjectConnectionAuthTypeParameterAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"airflow_param_name": types.StringType,
		"friendly_name":      types.StringType,
		"data_type":          types.StringType,
		"is_required":        types.BoolType,
		"is_secret":          types.BoolType,
		"description":        types.StringType,
		"example":            types.StringType,
		"is_in_extra":        types.BoolType,
	}
}

func EnvironmentObjectExcludeLinkAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"scope":           types.StringType,
		"scope_entity_id": types.StringType,
	}
}

func EnvironmentObjectLinkAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"scope":                      types.StringType,
		"scope_entity_id":            types.StringType,
		"airflow_variable_overrides": types.ObjectType{AttrTypes: EnvironmentObjectAirflowVariableOverridesAttributeTypes()},
		"connection_overrides":       types.ObjectType{AttrTypes: EnvironmentObjectConnectionOverridesAttributeTypes()},
		"metrics_export_overrides":   types.ObjectType{AttrTypes: EnvironmentObjectMetricsExportOverridesAttributeTypes()},
	}
}

func EnvironmentObjectAirflowVariableOverridesAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"value": types.StringType,
	}
}

func EnvironmentObjectConnectionOverridesAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":     types.StringType,
		"host":     types.StringType,
		"port":     types.Int64Type,
		"schema":   types.StringType,
		"login":    types.StringType,
		"password": types.StringType,
		"extra":    types.StringType,
	}
}

func EnvironmentObjectMetricsExportOverridesAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"auth_type":     types.StringType,
		"endpoint":      types.StringType,
		"basic_token":   types.StringType,
		"exporter_type": types.StringType,
		"username":      types.StringType,
		"password":      types.StringType,
		"headers":       types.MapType{ElemType: types.StringType},
		"labels":        types.MapType{ElemType: types.StringType},
	}
}

// --- Data Source Schema Attributes ---

func environmentObjectAirflowVariableDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"value": datasourceSchema.StringAttribute{
			MarkdownDescription: "The value of the Airflow variable",
			Computed:            true,
		},
		"is_secret": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Whether the value is a secret",
			Computed:            true,
		},
	}
}

func environmentObjectMetricsExportDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"auth_type": datasourceSchema.StringAttribute{
			MarkdownDescription: "The type of authentication to use when connecting to the remote endpoint",
			Computed:            true,
		},
		"endpoint": datasourceSchema.StringAttribute{
			MarkdownDescription: "The Prometheus endpoint where the metrics are exported",
			Computed:            true,
		},
		"basic_token": datasourceSchema.StringAttribute{
			MarkdownDescription: "The bearer token to connect to the remote endpoint",
			Computed:            true,
			Sensitive:           true,
		},
		"exporter_type": datasourceSchema.StringAttribute{
			MarkdownDescription: "The type of exporter",
			Computed:            true,
		},
		"username": datasourceSchema.StringAttribute{
			MarkdownDescription: "The username to connect to the remote endpoint",
			Computed:            true,
		},
		"password": datasourceSchema.StringAttribute{
			MarkdownDescription: "The password to connect to the remote endpoint",
			Computed:            true,
			Sensitive:           true,
		},
		"headers": datasourceSchema.MapAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "HTTP request headers for the remote endpoint",
			Computed:            true,
		},
		"labels": datasourceSchema.MapAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Key-value pair metrics labels for your export",
			Computed:            true,
		},
	}
}

func environmentObjectConnectionAuthTypeParameterDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"airflow_param_name": datasourceSchema.StringAttribute{
			MarkdownDescription: "The name of the parameter in Airflow",
			Computed:            true,
		},
		"friendly_name": datasourceSchema.StringAttribute{
			MarkdownDescription: "The UI-friendly name for the parameter",
			Computed:            true,
		},
		"data_type": datasourceSchema.StringAttribute{
			MarkdownDescription: "The data type of the parameter",
			Computed:            true,
		},
		"is_required": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Whether the parameter is required",
			Computed:            true,
		},
		"is_secret": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Whether the parameter is a secret",
			Computed:            true,
		},
		"description": datasourceSchema.StringAttribute{
			MarkdownDescription: "A description of the parameter",
			Computed:            true,
		},
		"example": datasourceSchema.StringAttribute{
			MarkdownDescription: "An example value for the parameter",
			Computed:            true,
		},
		"is_in_extra": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Whether the parameter is included in the extra field",
			Computed:            true,
		},
	}
}

func environmentObjectConnectionAuthTypeDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"parameters": datasourceSchema.ListNestedAttribute{
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: environmentObjectConnectionAuthTypeParameterDataSourceSchemaAttributes(),
			},
			MarkdownDescription: "The parameters for the connection auth type",
			Computed:            true,
		},
		"id": datasourceSchema.StringAttribute{
			MarkdownDescription: "The ID of the connection auth type",
			Computed:            true,
		},
		"name": datasourceSchema.StringAttribute{
			MarkdownDescription: "The name of the connection auth type",
			Computed:            true,
		},
		"auth_method_name": datasourceSchema.StringAttribute{
			MarkdownDescription: "The name of the auth method used in the connection",
			Computed:            true,
		},
		"airflow_type": datasourceSchema.StringAttribute{
			MarkdownDescription: "The type of connection in Airflow",
			Computed:            true,
		},
		"description": datasourceSchema.StringAttribute{
			MarkdownDescription: "A description of the connection auth type",
			Computed:            true,
		},
		"provider_package_name": datasourceSchema.StringAttribute{
			MarkdownDescription: "The name of the provider package",
			Computed:            true,
		},
		"provider_logo": datasourceSchema.StringAttribute{
			MarkdownDescription: "The URL of the provider logo",
			Computed:            true,
		},
		"guide_path": datasourceSchema.StringAttribute{
			MarkdownDescription: "The URL to the guide for the connection auth type",
			Computed:            true,
		},
	}
}

func environmentObjectConnectionDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"auth_type_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "The ID for the connection auth type",
			Computed:            true,
		},
		"connection_auth_type": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The auth type of the connection",
			Computed:            true,
			Attributes:          environmentObjectConnectionAuthTypeDataSourceSchemaAttributes(),
		},
		"type": datasourceSchema.StringAttribute{
			MarkdownDescription: "The type of connection",
			Computed:            true,
		},
		"host": datasourceSchema.StringAttribute{
			MarkdownDescription: "The host address for the connection",
			Computed:            true,
		},
		"port": datasourceSchema.Int64Attribute{
			MarkdownDescription: "The port for the connection",
			Computed:            true,
		},
		"schema": datasourceSchema.StringAttribute{
			MarkdownDescription: "The schema for the connection",
			Computed:            true,
		},
		"login": datasourceSchema.StringAttribute{
			MarkdownDescription: "The username used for the connection",
			Computed:            true,
		},
		"password": datasourceSchema.StringAttribute{
			MarkdownDescription: "The password used for the connection",
			Computed:            true,
			Sensitive:           true,
		},
		"extra": datasourceSchema.StringAttribute{
			MarkdownDescription: "Extra connection details as JSON string",
			Computed:            true,
		},
	}
}

func environmentObjectAirflowVariableOverridesDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"value": datasourceSchema.StringAttribute{
			MarkdownDescription: "The value of the Airflow variable",
			Computed:            true,
		},
	}
}

func environmentObjectConnectionOverridesDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"type": datasourceSchema.StringAttribute{
			MarkdownDescription: "The type of connection",
			Computed:            true,
		},
		"host": datasourceSchema.StringAttribute{
			MarkdownDescription: "The host address for the connection",
			Computed:            true,
		},
		"port": datasourceSchema.Int64Attribute{
			MarkdownDescription: "The port for the connection",
			Computed:            true,
		},
		"schema": datasourceSchema.StringAttribute{
			MarkdownDescription: "The schema for the connection",
			Computed:            true,
		},
		"login": datasourceSchema.StringAttribute{
			MarkdownDescription: "The username used for the connection",
			Computed:            true,
		},
		"password": datasourceSchema.StringAttribute{
			MarkdownDescription: "The password used for the connection",
			Computed:            true,
			Sensitive:           true,
		},
		"extra": datasourceSchema.StringAttribute{
			MarkdownDescription: "Extra connection details as JSON string",
			Computed:            true,
		},
	}
}

func environmentObjectMetricsExportOverridesDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"auth_type": datasourceSchema.StringAttribute{
			MarkdownDescription: "The type of authentication",
			Computed:            true,
		},
		"endpoint": datasourceSchema.StringAttribute{
			MarkdownDescription: "The Prometheus endpoint",
			Computed:            true,
		},
		"basic_token": datasourceSchema.StringAttribute{
			MarkdownDescription: "The bearer token",
			Computed:            true,
			Sensitive:           true,
		},
		"exporter_type": datasourceSchema.StringAttribute{
			MarkdownDescription: "The type of exporter",
			Computed:            true,
		},
		"username": datasourceSchema.StringAttribute{
			MarkdownDescription: "The username",
			Computed:            true,
		},
		"password": datasourceSchema.StringAttribute{
			MarkdownDescription: "The password",
			Computed:            true,
			Sensitive:           true,
		},
		"headers": datasourceSchema.MapAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "HTTP request headers",
			Computed:            true,
		},
		"labels": datasourceSchema.MapAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Metrics labels",
			Computed:            true,
		},
	}
}

func environmentObjectLinkDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"scope": datasourceSchema.StringAttribute{
			MarkdownDescription: "Scope of the linked entity",
			Computed:            true,
		},
		"scope_entity_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Linked entity ID",
			Computed:            true,
		},
		"airflow_variable_overrides": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Airflow variable overrides for this link",
			Computed:            true,
			Attributes:          environmentObjectAirflowVariableOverridesDataSourceSchemaAttributes(),
		},
		"connection_overrides": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Connection overrides for this link",
			Computed:            true,
			Attributes:          environmentObjectConnectionOverridesDataSourceSchemaAttributes(),
		},
		"metrics_export_overrides": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Metrics export overrides for this link",
			Computed:            true,
			Attributes:          environmentObjectMetricsExportOverridesDataSourceSchemaAttributes(),
		},
	}
}

func environmentObjectExcludeLinkDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"scope": datasourceSchema.StringAttribute{
			MarkdownDescription: "Scope of the excluded entity",
			Computed:            true,
		},
		"scope_entity_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "ID of the excluded entity",
			Computed:            true,
		},
	}
}

func EnvironmentObjectDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Environment object identifier",
			Required:            true,
			Validators:          []validator.String{validators.IsCuid()},
		},
		"object_key": datasourceSchema.StringAttribute{
			MarkdownDescription: "The key for the environment object",
			Computed:            true,
		},
		"object_type": datasourceSchema.StringAttribute{
			MarkdownDescription: "The type of environment object (AIRFLOW_VARIABLE, CONNECTION, METRICS_EXPORT)",
			Computed:            true,
		},
		"scope": datasourceSchema.StringAttribute{
			MarkdownDescription: "The scope of the environment object (WORKSPACE, DEPLOYMENT)",
			Computed:            true,
		},
		"scope_entity_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "The ID of the scope entity",
			Computed:            true,
		},
		"source_scope": datasourceSchema.StringAttribute{
			MarkdownDescription: "The source scope, if resolved from a link",
			Computed:            true,
		},
		"source_scope_entity_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "The source scope entity ID, if resolved from a link",
			Computed:            true,
		},
		"auto_link_deployments": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Whether to automatically link Deployments to the environment object",
			Computed:            true,
		},
		"airflow_variable": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The Airflow variable definition",
			Computed:            true,
			Attributes:          environmentObjectAirflowVariableDataSourceSchemaAttributes(),
		},
		"connection_config": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The connection definition",
			Computed:            true,
			Attributes:          environmentObjectConnectionDataSourceSchemaAttributes(),
		},
		"metrics_export": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The metrics export definition",
			Computed:            true,
			Attributes:          environmentObjectMetricsExportDataSourceSchemaAttributes(),
		},
		"links": datasourceSchema.ListNestedAttribute{
			MarkdownDescription: "The Deployments linked to the environment object",
			Computed:            true,
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: environmentObjectLinkDataSourceSchemaAttributes(),
			},
		},
		"exclude_links": datasourceSchema.ListNestedAttribute{
			MarkdownDescription: "The excluded links for the environment object",
			Computed:            true,
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: environmentObjectExcludeLinkDataSourceSchemaAttributes(),
			},
		},
		"created_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "Creation timestamp",
			Computed:            true,
		},
		"updated_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "Last updated timestamp",
			Computed:            true,
		},
		"created_by": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Creator",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchemaAttributes(),
		},
		"updated_by": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Updater",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchemaAttributes(),
		},
	}
}

// --- Resource Schema Attributes ---

func environmentObjectAirflowVariableResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"value": resourceSchema.StringAttribute{
			MarkdownDescription: "The value of the Airflow variable",
			Optional:            true,
			Sensitive:           true,
		},
		"is_secret": resourceSchema.BoolAttribute{
			MarkdownDescription: "Whether the value is a secret",
			Optional:            true,
			Computed:            true,
		},
	}
}

func environmentObjectMetricsExportResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"auth_type": resourceSchema.StringAttribute{
			MarkdownDescription: "The type of authentication (AUTH_TOKEN, BASIC)",
			Optional:            true,
			Computed:            true,
		},
		"endpoint": resourceSchema.StringAttribute{
			MarkdownDescription: "The Prometheus endpoint where the metrics are exported",
			Required:            true,
		},
		"basic_token": resourceSchema.StringAttribute{
			MarkdownDescription: "The bearer token to connect to the remote endpoint",
			Optional:            true,
			Sensitive:           true,
		},
		"exporter_type": resourceSchema.StringAttribute{
			MarkdownDescription: "The type of exporter (PROMETHEUS)",
			Required:            true,
		},
		"username": resourceSchema.StringAttribute{
			MarkdownDescription: "The username to connect to the remote endpoint",
			Optional:            true,
		},
		"password": resourceSchema.StringAttribute{
			MarkdownDescription: "The password to connect to the remote endpoint",
			Optional:            true,
			Sensitive:           true,
		},
		"headers": resourceSchema.MapAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "HTTP request headers for the remote endpoint",
			Optional:            true,
		},
		"labels": resourceSchema.MapAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Key-value pair metrics labels for your export",
			Optional:            true,
		},
	}
}

func environmentObjectConnectionResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"auth_type_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The ID for the connection auth type. Write-only field used during create/update.",
			Optional:            true,
		},
		"connection_auth_type": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The resolved auth type of the connection. Read-only, populated from auth_type_id.",
			Computed:            true,
			Attributes: map[string]resourceSchema.Attribute{
				"parameters": resourceSchema.ListNestedAttribute{
					NestedObject: resourceSchema.NestedAttributeObject{
						Attributes: map[string]resourceSchema.Attribute{
							"airflow_param_name": resourceSchema.StringAttribute{Computed: true},
							"friendly_name":      resourceSchema.StringAttribute{Computed: true},
							"data_type":          resourceSchema.StringAttribute{Computed: true},
							"is_required":        resourceSchema.BoolAttribute{Computed: true},
							"is_secret":          resourceSchema.BoolAttribute{Computed: true},
							"description":        resourceSchema.StringAttribute{Computed: true},
							"example":            resourceSchema.StringAttribute{Computed: true},
							"is_in_extra":        resourceSchema.BoolAttribute{Computed: true},
						},
					},
					Computed: true,
				},
				"id":                    resourceSchema.StringAttribute{Computed: true},
				"name":                  resourceSchema.StringAttribute{Computed: true},
				"auth_method_name":      resourceSchema.StringAttribute{Computed: true},
				"airflow_type":          resourceSchema.StringAttribute{Computed: true},
				"description":           resourceSchema.StringAttribute{Computed: true},
				"provider_package_name": resourceSchema.StringAttribute{Computed: true},
				"provider_logo":         resourceSchema.StringAttribute{Computed: true},
				"guide_path":            resourceSchema.StringAttribute{Computed: true},
			},
		},
		"type": resourceSchema.StringAttribute{
			MarkdownDescription: "The type of connection",
			Required:            true,
		},
		"host": resourceSchema.StringAttribute{
			MarkdownDescription: "The host address for the connection",
			Optional:            true,
		},
		"port": resourceSchema.Int64Attribute{
			MarkdownDescription: "The port for the connection",
			Optional:            true,
		},
		"schema": resourceSchema.StringAttribute{
			MarkdownDescription: "The schema for the connection",
			Optional:            true,
		},
		"login": resourceSchema.StringAttribute{
			MarkdownDescription: "The username used for the connection",
			Optional:            true,
		},
		"password": resourceSchema.StringAttribute{
			MarkdownDescription: "The password used for the connection",
			Optional:            true,
			Sensitive:           true,
		},
		"extra": resourceSchema.StringAttribute{
			MarkdownDescription: "Extra connection details as JSON string",
			Optional:            true,
		},
	}
}

func environmentObjectExcludeLinkResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"scope": resourceSchema.StringAttribute{
			MarkdownDescription: "Scope of the excluded entity (DEPLOYMENT)",
			Required:            true,
		},
		"scope_entity_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the excluded entity",
			Required:            true,
		},
	}
}

func environmentObjectLinkResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"scope": resourceSchema.StringAttribute{
			MarkdownDescription: "Scope of the linked entity (DEPLOYMENT)",
			Required:            true,
		},
		"scope_entity_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Linked entity ID",
			Required:            true,
		},
		"airflow_variable_overrides": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Airflow variable overrides for this link",
			Optional:            true,
			Attributes: map[string]resourceSchema.Attribute{
				"value": resourceSchema.StringAttribute{
					MarkdownDescription: "The value of the Airflow variable",
					Optional:            true,
					Sensitive:           true,
				},
			},
		},
		"connection_overrides": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Connection overrides for this link",
			Optional:            true,
			Attributes: map[string]resourceSchema.Attribute{
				"type": resourceSchema.StringAttribute{
					MarkdownDescription: "The type of connection",
					Optional:            true,
				},
				"host": resourceSchema.StringAttribute{
					MarkdownDescription: "The host address",
					Optional:            true,
				},
				"port": resourceSchema.Int64Attribute{
					MarkdownDescription: "The port",
					Optional:            true,
				},
				"schema": resourceSchema.StringAttribute{
					MarkdownDescription: "The schema",
					Optional:            true,
				},
				"login": resourceSchema.StringAttribute{
					MarkdownDescription: "The username",
					Optional:            true,
				},
				"password": resourceSchema.StringAttribute{
					MarkdownDescription: "The password",
					Optional:            true,
					Sensitive:           true,
				},
				"extra": resourceSchema.StringAttribute{
					MarkdownDescription: "Extra connection details as JSON string",
					Optional:            true,
				},
			},
		},
		"metrics_export_overrides": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Metrics export overrides for this link",
			Optional:            true,
			Attributes: map[string]resourceSchema.Attribute{
				"auth_type": resourceSchema.StringAttribute{
					MarkdownDescription: "The type of authentication",
					Optional:            true,
				},
				"endpoint": resourceSchema.StringAttribute{
					MarkdownDescription: "The Prometheus endpoint",
					Optional:            true,
				},
				"basic_token": resourceSchema.StringAttribute{
					MarkdownDescription: "The bearer token",
					Optional:            true,
					Sensitive:           true,
				},
				"exporter_type": resourceSchema.StringAttribute{
					MarkdownDescription: "The type of exporter",
					Optional:            true,
				},
				"username": resourceSchema.StringAttribute{
					MarkdownDescription: "The username",
					Optional:            true,
				},
				"password": resourceSchema.StringAttribute{
					MarkdownDescription: "The password",
					Optional:            true,
					Sensitive:           true,
				},
				"headers": resourceSchema.MapAttribute{
					ElementType:         types.StringType,
					MarkdownDescription: "HTTP request headers",
					Optional:            true,
				},
				"labels": resourceSchema.MapAttribute{
					ElementType:         types.StringType,
					MarkdownDescription: "Metrics labels",
					Optional:            true,
				},
			},
		},
	}
}

func EnvironmentObjectResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Environment object identifier",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"object_key": resourceSchema.StringAttribute{
			MarkdownDescription: "The key for the environment object",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"object_type": resourceSchema.StringAttribute{
			MarkdownDescription: "The type of environment object (AIRFLOW_VARIABLE, CONNECTION, METRICS_EXPORT)",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"scope": resourceSchema.StringAttribute{
			MarkdownDescription: "The scope of the environment object (WORKSPACE, DEPLOYMENT)",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"scope_entity_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The ID of the scope entity where the environment object is created",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"source_scope": resourceSchema.StringAttribute{
			MarkdownDescription: "The source scope, if resolved from a link",
			Computed:            true,
		},
		"source_scope_entity_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The source scope entity ID, if resolved from a link",
			Computed:            true,
		},
		"auto_link_deployments": resourceSchema.BoolAttribute{
			MarkdownDescription: "Whether to automatically link Deployments to the environment object. Only applicable for WORKSPACE scope",
			Optional:            true,
			Computed:            true,
		},
		"airflow_variable": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The Airflow variable definition. Required when object_type is AIRFLOW_VARIABLE",
			Optional:            true,
			Computed:            true,
			Attributes:          environmentObjectAirflowVariableResourceSchemaAttributes(),
		},
		"connection_config": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The connection definition. Required when object_type is CONNECTION",
			Optional:            true,
			Computed:            true,
			Attributes:          environmentObjectConnectionResourceSchemaAttributes(),
		},
		"metrics_export": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The metrics export definition. Required when object_type is METRICS_EXPORT",
			Optional:            true,
			Computed:            true,
			Attributes:          environmentObjectMetricsExportResourceSchemaAttributes(),
		},
		"links": resourceSchema.ListNestedAttribute{
			MarkdownDescription: "The Deployments linked to the environment object. Only applicable for WORKSPACE scope",
			Optional:            true,
			Computed:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: environmentObjectLinkResourceSchemaAttributes(),
			},
		},
		"exclude_links": resourceSchema.ListNestedAttribute{
			MarkdownDescription: "The excluded links for the environment object. Only applicable for WORKSPACE scope",
			Optional:            true,
			Computed:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: environmentObjectExcludeLinkResourceSchemaAttributes(),
			},
		},
		"created_at": resourceSchema.StringAttribute{
			MarkdownDescription: "Creation timestamp",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_at": resourceSchema.StringAttribute{
			MarkdownDescription: "Last updated timestamp",
			Computed:            true,
		},
		"created_by": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Creator",
			Computed:            true,
			Attributes:          ResourceSubjectProfileSchemaAttributes(),
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_by": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Updater",
			Computed:            true,
			Attributes:          ResourceSubjectProfileSchemaAttributes(),
		},
	}
}
