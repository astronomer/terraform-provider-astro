package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func EnvironmentObjectMetricsExportAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"auth_type":     types.StringType,
		"endpoint":      types.StringType,
		"basic_token":   types.StringType,
		"exporter_type": types.StringType,
		"username":      types.StringType,
		"password":      types.StringType,
		"headers":       types.MapType{ElemType: types.StringType},
		"labels": types.MapType{
			ElemType: types.StringType,
		},
	}
}

func EnvironmentObjectMetricsExportDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
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
		},
		"headers": datasourceSchema.MapAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Add key-value pairs to the HTTP request headers made by Astro when connecting to the remote endpoint",
			Computed:            true,
		},
		"labels": datasourceSchema.MapAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Any key-value pair metrics labels for your export. You can use these to filter your metrics in downstream applications.",
			Computed:            true,
		},
	}
}

func EnvironmentObjectMetricsExportOverridesDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
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
		},
		"headers": datasourceSchema.MapAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Add key-value pairs to the HTTP request headers made by Astro when connecting to the remote endpoint",
			Computed:            true,
		},
		"labels": datasourceSchema.MapAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Any key-value pair metrics labels for your export. You can use these to filter your metrics in downstream applications.",
			Computed:            true,
		},
	}
}

func EnvironmentObjectAirflowVariableDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"value": datasourceSchema.StringAttribute{
			MarkdownDescription: "The value of the Airflow variable. If the value is a secret, the value returned is empty",
			Computed:            true,
		},
		"is_secret": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Whether the value is a secret or not",
			Computed:            true,
		},
	}
}

func EnvironmentObjectAirflowVariableOverridesDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"value": datasourceSchema.StringAttribute{
			MarkdownDescription: "The value of the Airflow variable. If the value is a secret, the value returned is empty",
			Computed:            true,
		},
	}
}

func EnvironmentObjectConnectionAuthTypeParametersAttributeTypes() map[string]attr.Type {
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

func EnvironmentObjectConnectionAuthTypeParametersDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"airflow_param_name": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The name of the parameter in Airflow",
			Computed:            true,
			Attributes:          EnvironmentObjectConnectionAuthTypeDataSourceSchemaAttributes(),
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
			MarkdownDescription: "Whether or not the parameter is included in the \"extra\" field",
			Computed:            true,
		},
	}
}

func EnvironmentObjectConnectionAuthTypeAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"parameters": types.ObjectType{
			AttrTypes: EnvironmentObjectConnectionAuthTypeParametersAttributeTypes(),
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

func EnvironmentObjectConnectionAuthTypeDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"parameters": datasourceSchema.SetNestedAttribute{
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: EnvironmentObjectConnectionAuthTypeParametersDataSourceSchemaAttributes(),
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
			MarkdownDescription: "he URL of the provider logo",
			Computed:            true,
		},
		"guide_path": datasourceSchema.StringAttribute{
			MarkdownDescription: "The URL to the guide for the connection auth type",
			Computed:            true,
		},
	}
}

func EnvironmentObjectConnectionAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
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

func EnvironmentObjectConnectionDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"connection_auth_type": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The auth type of the connection",
			Computed:            true,
			Attributes:          EnvironmentObjectConnectionAuthTypeDataSourceSchemaAttributes(),
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
		},
		"extra": datasourceSchema.StringAttribute{
			MarkdownDescription: "Extra connection details, if any",
			Computed:            true,
		},
	}
}

func EnvironmentObjectConnectionOverridesDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
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
		},
		"extra": datasourceSchema.StringAttribute{
			MarkdownDescription: "Extra connection details, if any",
			Computed:            true,
		},
	}
}

func EnvironmentObjectExcludeLinkAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"scope":           types.StringType,
		"scope_entity_id": types.StringType,
	}
}

func EnvironmentObjectExcludeLinksDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"scope": datasourceSchema.StringAttribute{
			MarkdownDescription: "Scope of the excluded entity for environment object",
			Computed:            true,
		},
		"scope_entity_id": datasourceSchema.BoolAttribute{
			MarkdownDescription: "ID for the excluded entity for the environment object",
			Computed:            true,
		},
	}
}

func EnvironmentObjectLinksAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"scope":           types.StringType,
		"scope_entity_id": types.StringType,
		"connection": types.ObjectType{
			AttrTypes: EnvironmentObjectLinksConnectionOverridesAttributeTypes(),
		},
		"airflow_variable": types.ObjectType{
			AttrTypes: EnvironmentObjectLinksAirflowVariableOverridesAttributeTypes(),
		},
		"metrics_export": types.ObjectType{
			AttrTypes: EnvironmentObjectLinksMetricsExportOverridesAttributeTypes(),
		},
	}
}

func EnvironmentObjectLinksAirflowVariableOverridesAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"value": types.StringType,
	}
}

func EnvironmentObjectLinksConnectionOverridesAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":     types.StringType,
		"host":     types.StringType,
		"port":     types.Int64Type,
		"schema":   types.StringType,
		"login":    types.StringType,
		"password": types.StringType,
		"extra":    types.MapType{ElemType: types.StringType},
	}
}

func EnvironmentObjectLinksMetricsExportOverridesAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"auth_type":     types.StringType,
		"endpoint":      types.StringType,
		"basic_token":   types.StringType,
		"exporter_type": types.StringType,
		"username":      types.StringType,
		"password":      types.StringType,
		"extra":         types.StringType,
		"headers":       types.MapType{ElemType: types.StringType},
		"labels": types.MapType{
			ElemType: types.StringType,
		},
	}
}

func EnvironmentObjectLinksDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"scope": datasourceSchema.StringAttribute{
			MarkdownDescription: "Scope of the linked entity for the environment object",
			Computed:            true,
		},
		"scope_entity_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Linked entity ID the environment object",
			Computed:            true,
		},
		"connection": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The connection object, if the object type is CONNECTION",
			Computed:            true,
			Attributes:          EnvironmentObjectConnectionOverridesDataSourceSchemaAttributes(),
		},
		"airflow_variable": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The Airflow variable object, if the object type is AIRFLOW_VARIABLE",
			Computed:            true,
			Attributes:          EnvironmentObjectAirflowVariableOverridesDataSourceSchemaAttributes(),
		},
		"metrics_export": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The metrics export object, if the object type is METRICS_EXPORT",
			Computed:            true,
			Attributes:          EnvironmentObjectMetricsExportOverridesDataSourceSchemaAttributes(),
		},
	}
}

func EnvironmentObjectDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"id": datasourceSchema.StringAttribute{
			MarkdownDescription: "EnvironmentObject identifier",
			Required:            true,
			Validators:          []validator.String{validators.IsCuid()},
		},
		"connection": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The connection object, if the object type is CONNECTION",
			Computed:            true,
			Attributes:          EnvironmentObjectConnectionDataSourceSchemaAttributes(),
		},
		"airflow_variable": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The Airflow variable object, if the object type is AIRFLOW_VARIABLE",
			Computed:            true,
			Attributes:          EnvironmentObjectAirflowVariableDataSourceSchemaAttributes(),
		},
		"metrics_export": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The metrics export object, if the object type is METRICS_EXPORT",
			Computed:            true,
			Attributes:          EnvironmentObjectMetricsExportDataSourceSchemaAttributes(),
		},
		"links": datasourceSchema.SetNestedAttribute{
			MarkdownDescription: "The Deployments linked to the environment object",
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: EnvironmentObjectLinksDataSourceSchemaAttributes(),
			},
			Computed: true,
		},
		"exclude_links": datasourceSchema.SetNestedAttribute{
			MarkdownDescription: "The excluded links for the environment object",
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: EnvironmentObjectExcludeLinksDataSourceSchemaAttributes(),
			},
			Computed: true,
		},
		"scope": datasourceSchema.StringAttribute{
			MarkdownDescription: "The scope of the environment object",
			Computed:            true,
		},
		"scope_entity_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "The ID of the scope entity where the environment object is created",
			Computed:            true,
		},
		"object_type": datasourceSchema.StringAttribute{
			MarkdownDescription: "The type of environment object",
			Computed:            true,
		},
		"object_key": datasourceSchema.StringAttribute{
			MarkdownDescription: "The key for the environment object",
			Computed:            true,
		},
		"source_scope": datasourceSchema.StringAttribute{
			MarkdownDescription: "The source scope of the environment object, if it is resolved from a link",
			Computed:            true,
		},
		"source_scope_entity_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "The source scope entity ID of the environment object, if it is resolved from a link",
			Computed:            true,
		},
		"auto_link_deployments": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Whether or not to automatically link Deployments to the environment object",
			Computed:            true,
		},
		"created_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "EnvironmentObject creation timestamp",
			Computed:            true,
		},
		"updated_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "EnvironmentObject last updated timestamp",
			Computed:            true,
		},
		"created_by": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "EnvironmentObject creator",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchemaAttributes(),
		},
		"updated_by": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "EnvironmentObject updater",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchemaAttributes(),
		},
	}
}

func EnvironmentObjectMetricsExportResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"auth_type": resourceSchema.StringAttribute{
			MarkdownDescription: "The type of authentication to use when connecting to the remote endpoint",
		},
		"endpoint": resourceSchema.StringAttribute{
			MarkdownDescription: "The Prometheus endpoint where the metrics are exported",
			Required:            true,
		},
		"basic_token": resourceSchema.StringAttribute{
			MarkdownDescription: "The bearer token to connect to the remote endpoint",
		},
		"exporter_type": resourceSchema.StringAttribute{
			MarkdownDescription: "The type of exporter",
		},
		"username": resourceSchema.StringAttribute{
			MarkdownDescription: "The username to connect to the remote endpoint",
		},
		"password": resourceSchema.StringAttribute{
			MarkdownDescription: "The password to connect to the remote endpoint",
		},
		"headers": resourceSchema.MapAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Add key-value pairs to the HTTP request headers made by Astro when connecting to the remote endpoint",
		},
		"labels": resourceSchema.MapAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Any key-value pair metrics labels for your export. You can use these to filter your metrics in downstream applications.",
		},
	}
}

func EnvironmentObjectMetricsExportOverridesResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"auth_type": resourceSchema.StringAttribute{
			MarkdownDescription: "The type of authentication to use when connecting to the remote endpoint",
		},
		"endpoint": resourceSchema.StringAttribute{
			MarkdownDescription: "The Prometheus endpoint where the metrics are exported",
		},
		"basic_token": resourceSchema.StringAttribute{
			MarkdownDescription: "The bearer token to connect to the remote endpoint",
		},
		"exporter_type": resourceSchema.StringAttribute{
			MarkdownDescription: "The type of exporter",
		},
		"username": resourceSchema.StringAttribute{
			MarkdownDescription: "The username to connect to the remote endpoint",
		},
		"password": resourceSchema.StringAttribute{
			MarkdownDescription: "The password to connect to the remote endpoint",
		},
		"headers": resourceSchema.MapAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Add key-value pairs to the HTTP request headers made by Astro when connecting to the remote endpoint",
		},
		"labels": resourceSchema.MapAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Any key-value pair metrics labels for your export. You can use these to filter your metrics in downstream applications.",
		},
	}
}

func EnvironmentObjectAirflowVariableAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"value":     types.StringType,
		"is_secret": types.BoolType,
	}
}

func EnvironmentObjectAirflowVariableResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"value": resourceSchema.StringAttribute{
			MarkdownDescription: "The value of the Airflow variable. If the value is a secret, the value returned is empty",
		},
		"is_secret": resourceSchema.BoolAttribute{
			MarkdownDescription: "Whether the value is a secret or not",
		},
	}
}

func EnvironmentObjectAirflowVariableOverridesResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"value": resourceSchema.StringAttribute{
			MarkdownDescription: "The value of the Airflow variable. If the value is a secret, the value returned is empty",
		},
	}
}

func EnvironmentObjectConnectionAuthTypeParametersResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"airflow_param_name": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The name of the parameter in Airflow",
			Computed:            true,
			Attributes:          EnvironmentObjectConnectionAuthTypeResourceSchemaAttributes(),
		},
		"friendly_name": resourceSchema.StringAttribute{
			MarkdownDescription: "The UI-friendly name for the parameter",
			Computed:            true,
		},
		"data_type": resourceSchema.StringAttribute{
			MarkdownDescription: "The data type of the parameter",
			Computed:            true,
		},
		"is_required": resourceSchema.BoolAttribute{
			MarkdownDescription: "Whether the parameter is required",
			Computed:            true,
		},
		"is_secret": resourceSchema.BoolAttribute{
			MarkdownDescription: "Whether the parameter is a secret",
			Computed:            true,
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "A description of the parameter",
			Computed:            true,
		},
		"example": resourceSchema.StringAttribute{
			MarkdownDescription: "An example value for the parameter",
			Computed:            true,
		},
		"is_in_extra": resourceSchema.BoolAttribute{
			MarkdownDescription: "Whether or not the parameter is included in the \"extra\" field",
			Computed:            true,
		},
	}
}

func EnvironmentObjectConnectionAuthTypeResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"parameters": resourceSchema.SetNestedAttribute{
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: EnvironmentObjectConnectionAuthTypeParametersResourceSchemaAttributes(),
			},
			MarkdownDescription: "The parameters for the connection auth type",
			Computed:            true,
		},
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "The ID of the connection auth type",
			Computed:            true,
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "The name of the connection auth type",
			Computed:            true,
		},
		"auth_method_name": resourceSchema.StringAttribute{
			MarkdownDescription: "The name of the auth method used in the connection",
			Computed:            true,
		},
		"airflow_type": resourceSchema.StringAttribute{
			MarkdownDescription: "The type of connection in Airflow",
			Computed:            true,
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "A description of the connection auth type",
			Computed:            true,
		},
		"provider_package_name": resourceSchema.StringAttribute{
			MarkdownDescription: "The name of the provider package",
			Computed:            true,
		},
		"provider_logo": resourceSchema.StringAttribute{
			MarkdownDescription: "he URL of the provider logo",
			Computed:            true,
		},
		"guide_path": resourceSchema.StringAttribute{
			MarkdownDescription: "The URL to the guide for the connection auth type",
			Computed:            true,
		},
	}
}

func EnvironmentObjectConnectionResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"auth_type_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The ID for the connection auth type",
		},
		"type": resourceSchema.StringAttribute{
			MarkdownDescription: "The type of connection",
			Required:            true,
		},
		"host": resourceSchema.StringAttribute{
			MarkdownDescription: "The host address for the connection",
		},
		"port": resourceSchema.Int64Attribute{
			MarkdownDescription: "The port for the connection",
		},
		"schema": resourceSchema.StringAttribute{
			MarkdownDescription: "The schema for the connection",
		},
		"login": resourceSchema.StringAttribute{
			MarkdownDescription: "The username used for the connection",
		},
		"password": resourceSchema.StringAttribute{
			MarkdownDescription: "The password used for the connection",
		},
		"extra": resourceSchema.StringAttribute{
			MarkdownDescription: "Extra connection details, if any",
		},
	}
}

func EnvironmentObjectConnectionOverridesResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"type": resourceSchema.StringAttribute{
			MarkdownDescription: "The type of connection",
		},
		"host": resourceSchema.StringAttribute{
			MarkdownDescription: "The host address for the connection",
		},
		"port": resourceSchema.Int64Attribute{
			MarkdownDescription: "The port for the connection",
		},
		"schema": resourceSchema.StringAttribute{
			MarkdownDescription: "The schema for the connection",
		},
		"login": resourceSchema.StringAttribute{
			MarkdownDescription: "The username used for the connection",
		},
		"password": resourceSchema.StringAttribute{
			MarkdownDescription: "The password used for the connection",
		},
		"extra": resourceSchema.StringAttribute{
			MarkdownDescription: "Extra connection details, if any",
		},
	}
}

func EnvironmentObjectExcludeLinksResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"scope": resourceSchema.StringAttribute{
			MarkdownDescription: "Scope of the excluded entity for environment object",
			Required:            true,
		},
		"scope_entity_id": resourceSchema.BoolAttribute{
			MarkdownDescription: "ID for the excluded entity for the environment object",
			Required:            true,
		},
	}
}

func EnvironmentObjectOverridesResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"connection": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The connection object, if the object type is CONNECTION",
			Computed:            true,
			Attributes:          EnvironmentObjectConnectionOverridesResourceSchemaAttributes(),
		},
		"airflow_variable": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The Airflow variable object, if the object type is AIRFLOW_VARIABLE",
			Computed:            true,
			Attributes:          EnvironmentObjectAirflowVariableOverridesResourceSchemaAttributes(),
		},
		"metrics_export": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The metrics export object, if the object type is METRICS_EXPORT",
			Computed:            true,
			Attributes:          EnvironmentObjectMetricsExportOverridesResourceSchemaAttributes(),
		},
	}
}

func EnvironmentObjectLinksResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"scope": resourceSchema.StringAttribute{
			MarkdownDescription: "Scope of the linked entity for the environment object",
			Required:            true,
		},
		"scope_entity_id": resourceSchema.BoolAttribute{
			MarkdownDescription: "Linked entity ID the environment object",
			Required:            true,
		},
		"overrides": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Linked entity ID the environment object",
			Attributes:          EnvironmentObjectOverridesResourceSchemaAttributes(),
		},
	}
}

func EnvironmentObjectResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "EnvironmentObject identifier",
			Computed:            true,
			Validators:          []validator.String{validators.IsCuid()},
		},
		"connection": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The connection object, if the object type is CONNECTION",
			Attributes:          EnvironmentObjectConnectionResourceSchemaAttributes(),
		},
		"airflow_variable": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The Airflow variable object, if the object type is AIRFLOW_VARIABLE",
			Attributes:          EnvironmentObjectAirflowVariableResourceSchemaAttributes(),
		},
		"metrics_export": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The metrics export object, if the object type is METRICS_EXPORT",
			Attributes:          EnvironmentObjectMetricsExportResourceSchemaAttributes(),
		},
		"links": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "The Deployments linked to the environment object",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: EnvironmentObjectLinksResourceSchemaAttributes(),
			},
		},
		"exclude_links": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "The excluded links for the environment object",
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: EnvironmentObjectExcludeLinksResourceSchemaAttributes(),
			},
		},
		"scope": resourceSchema.StringAttribute{
			MarkdownDescription: "The scope of the environment object",
			Required:            true,
		},
		"scope_entity_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The ID of the scope entity where the environment object is created",
			Required:            true,
		},
		"object_type": resourceSchema.StringAttribute{
			MarkdownDescription: "The type of environment object",
			Required:            true,
		},
		"object_key": resourceSchema.StringAttribute{
			MarkdownDescription: "The key for the environment object",
			Required:            true,
		},
		"source_scope": resourceSchema.StringAttribute{
			MarkdownDescription: "The source scope of the environment object, if it is resolved from a link",
			Computed:            true,
		},
		"source_scope_entity_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The source scope entity ID of the environment object, if it is resolved from a link",
			Computed:            true,
		},
		"auto_link_deployments": resourceSchema.BoolAttribute{
			MarkdownDescription: "Whether or not to automatically link Deployments to the environment object",
		},
		"created_at": resourceSchema.StringAttribute{
			MarkdownDescription: "EnvironmentObject creation timestamp",
			Computed:            true,
		},
		"updated_at": resourceSchema.StringAttribute{
			MarkdownDescription: "EnvironmentObject last updated timestamp",
			Computed:            true,
		},
		"created_by": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "EnvironmentObject creator",
			Computed:            true,
			Attributes:          ResourceSubjectProfileSchemaAttributes(),
		},
		"updated_by": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "EnvironmentObject updater",
			Computed:            true,
			Attributes:          ResourceSubjectProfileSchemaAttributes(),
		},
	}
}
