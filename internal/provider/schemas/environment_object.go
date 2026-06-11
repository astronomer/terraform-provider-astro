package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// EnvironmentObjectConnectionAuthTypeAttributeTypes describes the read-only
// `connection_auth_type` nested object that the API resolves from `auth_type_id`.
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
		"pattern":            types.StringType,
	}
}

func EnvironmentObjectExcludeLinkAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"scope":           types.StringType,
		"scope_entity_id": types.StringType,
	}
}

// EnvironmentObjectLinkAttributeTypes describes a single per-link element with
// a flat `overrides` block (one optional field per overridable attribute,
// discriminated by the parent's object_type — same flat-polymorphic pattern as
// `astro_notification_channel.definition`).
func EnvironmentObjectLinkAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"scope":           types.StringType,
		"scope_entity_id": types.StringType,
		"overrides":       types.ObjectType{AttrTypes: EnvironmentObjectOverridesAttributeTypes()},
	}
}

// EnvironmentObjectOverridesAttributeTypes describes the per-link `overrides`
// flat union: every overridable field across all object_types, all optional.
// The parent's object_type determines which fields are meaningful.
func EnvironmentObjectOverridesAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		// AIRFLOW_VARIABLE
		"value": types.StringType,
		// CONNECTION
		"type":     types.StringType,
		"host":     types.StringType,
		"port":     types.Int64Type,
		"schema":   types.StringType,
		"login":    types.StringType,
		"extra":    types.StringType,
		"password": types.StringType, // polymorphic: connection password OR basic-auth password
		// METRICS_EXPORT
		"auth_type":     types.StringType,
		"endpoint":      types.StringType,
		"basic_token":   types.StringType,
		"exporter_type": types.StringType,
		"username":      types.StringType,
		"headers":       types.MapType{ElemType: types.StringType},
		"labels":        types.MapType{ElemType: types.StringType},
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
		"pattern": datasourceSchema.StringAttribute{
			MarkdownDescription: "A regex pattern that the parameter value must match",
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

// environmentObjectOverridesDataSourceSchemaAttributes is the flat `overrides`
// shape for the data source — every overridable field across all object_types.
func environmentObjectOverridesDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"value": datasourceSchema.StringAttribute{
			MarkdownDescription: "Override value (only used when object_type=AIRFLOW_VARIABLE)",
			Computed:            true,
			Sensitive:           true,
		},
		"type": datasourceSchema.StringAttribute{
			MarkdownDescription: "Override connection type (only used when object_type=CONNECTION)",
			Computed:            true,
		},
		"host": datasourceSchema.StringAttribute{
			MarkdownDescription: "Override host address (only used when object_type=CONNECTION)",
			Computed:            true,
		},
		"port": datasourceSchema.Int64Attribute{
			MarkdownDescription: "Override port (only used when object_type=CONNECTION)",
			Computed:            true,
		},
		"schema": datasourceSchema.StringAttribute{
			MarkdownDescription: "Override schema (only used when object_type=CONNECTION)",
			Computed:            true,
		},
		"login": datasourceSchema.StringAttribute{
			MarkdownDescription: "Override login (only used when object_type=CONNECTION)",
			Computed:            true,
		},
		"extra": datasourceSchema.StringAttribute{
			MarkdownDescription: "Override extra JSON (only used when object_type=CONNECTION)",
			Computed:            true,
		},
		"password": datasourceSchema.StringAttribute{
			MarkdownDescription: "Override password — the connection password when object_type=CONNECTION, the HTTP Basic-auth password when object_type=METRICS_EXPORT",
			Computed:            true,
			Sensitive:           true,
		},
		"auth_type": datasourceSchema.StringAttribute{
			MarkdownDescription: "Override auth type (only used when object_type=METRICS_EXPORT)",
			Computed:            true,
		},
		"endpoint": datasourceSchema.StringAttribute{
			MarkdownDescription: "Override Prometheus endpoint (only used when object_type=METRICS_EXPORT)",
			Computed:            true,
		},
		"basic_token": datasourceSchema.StringAttribute{
			MarkdownDescription: "Override bearer token (only used when object_type=METRICS_EXPORT)",
			Computed:            true,
			Sensitive:           true,
		},
		"exporter_type": datasourceSchema.StringAttribute{
			MarkdownDescription: "Override exporter type (only used when object_type=METRICS_EXPORT)",
			Computed:            true,
		},
		"username": datasourceSchema.StringAttribute{
			MarkdownDescription: "Override username (only used when object_type=METRICS_EXPORT)",
			Computed:            true,
		},
		"headers": datasourceSchema.MapAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Override HTTP request headers (only used when object_type=METRICS_EXPORT)",
			Computed:            true,
		},
		"labels": datasourceSchema.MapAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Override metrics labels (only used when object_type=METRICS_EXPORT)",
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
		"overrides": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Per-link overrides. Only the fields matching the parent object_type are populated",
			Computed:            true,
			Attributes:          environmentObjectOverridesDataSourceSchemaAttributes(),
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

// EnvironmentObjectDataSourceSchemaAttributes is the flat data source schema
// matching the resource shape. Every type-specific field is Computed and
// populated only when the parent object_type matches.
func EnvironmentObjectDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		// Identity / common
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
			MarkdownDescription: "The ID of the scope entity where the environment object is created",
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
		// AIRFLOW_VARIABLE fields
		"value": datasourceSchema.StringAttribute{
			MarkdownDescription: "The value of the Airflow variable (only used when object_type=AIRFLOW_VARIABLE)",
			Computed:            true,
			Sensitive:           true,
		},
		"is_secret": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Whether the value is a secret (only used when object_type=AIRFLOW_VARIABLE)",
			Computed:            true,
		},
		// CONNECTION fields
		"type": datasourceSchema.StringAttribute{
			MarkdownDescription: "The connection type (only used when object_type=CONNECTION)",
			Computed:            true,
		},
		"host": datasourceSchema.StringAttribute{
			MarkdownDescription: "The host address for the connection (only used when object_type=CONNECTION)",
			Computed:            true,
		},
		"port": datasourceSchema.Int64Attribute{
			MarkdownDescription: "The port for the connection (only used when object_type=CONNECTION)",
			Computed:            true,
		},
		"schema": datasourceSchema.StringAttribute{
			MarkdownDescription: "The schema for the connection (only used when object_type=CONNECTION)",
			Computed:            true,
		},
		"login": datasourceSchema.StringAttribute{
			MarkdownDescription: "The username used for the connection (only used when object_type=CONNECTION)",
			Computed:            true,
		},
		"extra": datasourceSchema.StringAttribute{
			MarkdownDescription: "Extra connection details as JSON string (only used when object_type=CONNECTION)",
			Computed:            true,
		},
		"auth_type_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "The ID for the connection auth type (only used when object_type=CONNECTION)",
			Computed:            true,
		},
		"connection_auth_type": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The resolved auth type of the connection (only used when object_type=CONNECTION)",
			Computed:            true,
			Attributes:          environmentObjectConnectionAuthTypeDataSourceSchemaAttributes(),
		},
		// METRICS_EXPORT fields
		"endpoint": datasourceSchema.StringAttribute{
			MarkdownDescription: "The Prometheus endpoint where the metrics are exported (only used when object_type=METRICS_EXPORT)",
			Computed:            true,
		},
		"exporter_type": datasourceSchema.StringAttribute{
			MarkdownDescription: "The type of exporter (only used when object_type=METRICS_EXPORT)",
			Computed:            true,
		},
		"auth_type": datasourceSchema.StringAttribute{
			MarkdownDescription: "The type of authentication (only used when object_type=METRICS_EXPORT)",
			Computed:            true,
		},
		"basic_token": datasourceSchema.StringAttribute{
			MarkdownDescription: "The bearer token to connect to the remote endpoint (only used when object_type=METRICS_EXPORT)",
			Computed:            true,
			Sensitive:           true,
		},
		"username": datasourceSchema.StringAttribute{
			MarkdownDescription: "The username to connect to the remote endpoint (only used when object_type=METRICS_EXPORT)",
			Computed:            true,
		},
		"headers": datasourceSchema.MapAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "HTTP request headers for the remote endpoint (only used when object_type=METRICS_EXPORT)",
			Computed:            true,
		},
		"labels": datasourceSchema.MapAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Key-value pair metrics labels (only used when object_type=METRICS_EXPORT)",
			Computed:            true,
		},
		// Polymorphic — used by both CONNECTION and METRICS_EXPORT
		"password": datasourceSchema.StringAttribute{
			MarkdownDescription: "The password — the connection password when object_type=CONNECTION, the HTTP Basic-auth password when object_type=METRICS_EXPORT",
			Computed:            true,
			Sensitive:           true,
		},
		// Links
		"links": datasourceSchema.SetNestedAttribute{
			MarkdownDescription: "The Deployments linked to the environment object",
			Computed:            true,
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: environmentObjectLinkDataSourceSchemaAttributes(),
			},
		},
		"exclude_links": datasourceSchema.SetNestedAttribute{
			MarkdownDescription: "The excluded links for the environment object",
			Computed:            true,
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: environmentObjectExcludeLinkDataSourceSchemaAttributes(),
			},
		},
		// Metadata
		"created_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "Environment Object creation timestamp",
			Computed:            true,
		},
		"updated_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "Environment Object last updated timestamp",
			Computed:            true,
		},
		"created_by": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Environment Object creator",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchemaAttributes(),
		},
		"updated_by": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Environment Object updater",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchemaAttributes(),
		},
	}
}

func environmentObjectConnectionAuthTypeParameterResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"airflow_param_name": resourceSchema.StringAttribute{
			MarkdownDescription: "The name of the parameter in Airflow",
			Computed:            true,
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
			MarkdownDescription: "Whether the parameter is included in the extra field",
			Computed:            true,
		},
		"pattern": resourceSchema.StringAttribute{
			MarkdownDescription: "A regex pattern that the parameter value must match",
			Computed:            true,
		},
	}
}

func environmentObjectConnectionAuthTypeResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"parameters": resourceSchema.ListNestedAttribute{
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: environmentObjectConnectionAuthTypeParameterResourceSchemaAttributes(),
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
			MarkdownDescription: "The URL of the provider logo",
			Computed:            true,
		},
		"guide_path": resourceSchema.StringAttribute{
			MarkdownDescription: "The URL to the guide for the connection auth type",
			Computed:            true,
		},
	}
}

// environmentObjectOverridesResourceSchemaAttributes is the flat per-link
// `overrides` block. All fields Optional; ValidateConfig enforces which apply
// based on parent object_type.
func environmentObjectOverridesResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		// AIRFLOW_VARIABLE
		"value": resourceSchema.StringAttribute{
			MarkdownDescription: "Override value (only valid when object_type=AIRFLOW_VARIABLE)",
			Optional:            true,
			Sensitive:           true,
		},
		// CONNECTION
		"type": resourceSchema.StringAttribute{
			MarkdownDescription: "Override connection type (only valid when object_type=CONNECTION)",
			Optional:            true,
		},
		"host": resourceSchema.StringAttribute{
			MarkdownDescription: "Override host address (only valid when object_type=CONNECTION)",
			Optional:            true,
		},
		"port": resourceSchema.Int64Attribute{
			MarkdownDescription: "Override port (only valid when object_type=CONNECTION)",
			Optional:            true,
		},
		"schema": resourceSchema.StringAttribute{
			MarkdownDescription: "Override schema (only valid when object_type=CONNECTION)",
			Optional:            true,
		},
		"login": resourceSchema.StringAttribute{
			MarkdownDescription: "Override login (only valid when object_type=CONNECTION)",
			Optional:            true,
		},
		"extra": resourceSchema.StringAttribute{
			MarkdownDescription: "Override extra JSON (only valid when object_type=CONNECTION)",
			Optional:            true,
		},
		// METRICS_EXPORT
		"auth_type": resourceSchema.StringAttribute{
			MarkdownDescription: "Override auth type (only valid when object_type=METRICS_EXPORT)",
			Optional:            true,
		},
		"endpoint": resourceSchema.StringAttribute{
			MarkdownDescription: "Override Prometheus endpoint (only valid when object_type=METRICS_EXPORT)",
			Optional:            true,
		},
		"basic_token": resourceSchema.StringAttribute{
			MarkdownDescription: "Override bearer token (only valid when object_type=METRICS_EXPORT)",
			Optional:            true,
			Sensitive:           true,
		},
		"exporter_type": resourceSchema.StringAttribute{
			MarkdownDescription: "Override exporter type (only valid when object_type=METRICS_EXPORT)",
			Optional:            true,
		},
		"username": resourceSchema.StringAttribute{
			MarkdownDescription: "Override username (only valid when object_type=METRICS_EXPORT)",
			Optional:            true,
		},
		"headers": resourceSchema.MapAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Override HTTP request headers (only valid when object_type=METRICS_EXPORT)",
			Optional:            true,
		},
		"labels": resourceSchema.MapAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Override metrics labels (only valid when object_type=METRICS_EXPORT)",
			Optional:            true,
		},
		// Polymorphic
		"password": resourceSchema.StringAttribute{
			MarkdownDescription: "Override password — the connection password when object_type=CONNECTION, the HTTP Basic-auth password when object_type=METRICS_EXPORT",
			Optional:            true,
			Sensitive:           true,
		},
	}
}

func environmentObjectExcludeLinkResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"scope": resourceSchema.StringAttribute{
			MarkdownDescription: "Scope of the excluded entity (DEPLOYMENT)",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(string(platform.ExcludeLinkEnvironmentObjectRequestScopeDEPLOYMENT)),
			},
		},
		"scope_entity_id": resourceSchema.StringAttribute{
			MarkdownDescription: "ID of the excluded entity",
			Required:            true,
			Validators:          []validator.String{validators.IsCuid()},
		},
	}
}

func environmentObjectLinkResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"scope": resourceSchema.StringAttribute{
			MarkdownDescription: "Scope of the linked entity (DEPLOYMENT)",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(string(platform.CreateEnvironmentObjectLinkRequestScopeDEPLOYMENT)),
			},
		},
		"scope_entity_id": resourceSchema.StringAttribute{
			MarkdownDescription: "Linked entity ID",
			Required:            true,
			Validators:          []validator.String{validators.IsCuid()},
		},
		"overrides": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Per-link overrides. Set only the fields matching the parent object_type.",
			Optional:            true,
			Attributes:          environmentObjectOverridesResourceSchemaAttributes(),
		},
	}
}

// EnvironmentObjectResourceSchemaAttributes is the flat top-level schema.
// Type-specific fields sit directly on the resource and are gated by ValidateConfig
// against the `object_type` discriminator — same flat-polymorphic shape as
// `astro_notification_channel.definition` writ large.
func EnvironmentObjectResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		// Identity / common
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
			MarkdownDescription: "The type of environment object (AIRFLOW_VARIABLE, CONNECTION, METRICS_EXPORT). Determines which type-specific fields are required.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(platform.CreateEnvironmentObjectRequestObjectTypeAIRFLOWVARIABLE),
					string(platform.CreateEnvironmentObjectRequestObjectTypeCONNECTION),
					string(platform.CreateEnvironmentObjectRequestObjectTypeMETRICSEXPORT),
				),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"scope": resourceSchema.StringAttribute{
			MarkdownDescription: "The scope of the environment object (WORKSPACE, DEPLOYMENT)",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(platform.CreateEnvironmentObjectRequestScopeWORKSPACE),
					string(platform.CreateEnvironmentObjectRequestScopeDEPLOYMENT),
				),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"scope_entity_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The ID of the scope entity where the environment object is created",
			Required:            true,
			Validators:          []validator.String{validators.IsCuid()},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"source_scope": resourceSchema.StringAttribute{
			MarkdownDescription: "The source scope, if resolved from a link",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"source_scope_entity_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The source scope entity ID, if resolved from a link",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"auto_link_deployments": resourceSchema.BoolAttribute{
			MarkdownDescription: "Whether to automatically link Deployments to the environment object. Only applicable for WORKSPACE scope",
			Optional:            true,
			Computed:            true,
			// The model normalizes API nil → false (the API treats false as
			// the absence value). Without UseStateForUnknown, an omitted
			// auto_link_deployments would refresh as unknown each plan and
			// re-introduce the same drift after the first apply.
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
		// AIRFLOW_VARIABLE fields (only valid when object_type=AIRFLOW_VARIABLE)
		"value": resourceSchema.StringAttribute{
			MarkdownDescription: "The value of the Airflow variable (only valid when object_type=AIRFLOW_VARIABLE)",
			Optional:            true,
			Sensitive:           true,
		},
		"is_secret": resourceSchema.BoolAttribute{
			MarkdownDescription: "Whether the value is a secret (only valid when object_type=AIRFLOW_VARIABLE). Immutable on the API; toggling forces resource replacement.",
			Optional:            true,
			Computed:            true,
			// No Default — is_secret is type-specific. The API returns null
			// for CONNECTION / METRICS_EXPORT, so a blanket false default
			// would cause "inconsistent result after apply" drift.
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.RequiresReplace(),
			},
		},
		// CONNECTION fields (only valid when object_type=CONNECTION)
		"type": resourceSchema.StringAttribute{
			MarkdownDescription: "The connection type (required when object_type=CONNECTION). Immutable on the API; changing it forces resource replacement.",
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"host": resourceSchema.StringAttribute{
			MarkdownDescription: "The host address for the connection (only valid when object_type=CONNECTION)",
			Optional:            true,
			Computed:            true,
		},
		"port": resourceSchema.Int64Attribute{
			MarkdownDescription: "The port for the connection (only valid when object_type=CONNECTION)",
			Optional:            true,
			Computed:            true,
		},
		"schema": resourceSchema.StringAttribute{
			MarkdownDescription: "The schema for the connection (only valid when object_type=CONNECTION)",
			Optional:            true,
			Computed:            true,
		},
		"login": resourceSchema.StringAttribute{
			MarkdownDescription: "The username used for the connection (only valid when object_type=CONNECTION)",
			Optional:            true,
			Computed:            true,
		},
		"extra": resourceSchema.StringAttribute{
			MarkdownDescription: "Extra connection details as JSON string (only valid when object_type=CONNECTION). Use jsonencode({...})",
			Optional:            true,
			Computed:            true,
		},
		"auth_type_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The ID for the connection auth type (only valid when object_type=CONNECTION). Provided on create/update; not returned by the API",
			Optional:            true,
		},
		"connection_auth_type": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The resolved auth type of the connection, populated from auth_type_id (only set when object_type=CONNECTION)",
			Computed:            true,
			// No UseStateForUnknown: the API recomputes this object whenever
			// auth_type_id changes, but the framework doesn't know about that
			// cross-field dependency. Pinning to prior state would produce
			// "Provider produced inconsistent result after apply" on any
			// auth_type_id change. Leave it (known after apply) instead.
			Attributes: environmentObjectConnectionAuthTypeResourceSchemaAttributes(),
		},
		// METRICS_EXPORT fields (only valid when object_type=METRICS_EXPORT)
		"endpoint": resourceSchema.StringAttribute{
			MarkdownDescription: "The Prometheus endpoint where the metrics are exported (required when object_type=METRICS_EXPORT)",
			Optional:            true,
			Computed:            true,
		},
		"exporter_type": resourceSchema.StringAttribute{
			MarkdownDescription: "The type of exporter (required when object_type=METRICS_EXPORT)",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(platform.CreateEnvironmentObjectMetricsExportRequestExporterTypePROMETHEUS),
				),
			},
		},
		"auth_type": resourceSchema.StringAttribute{
			MarkdownDescription: "The type of authentication (only valid when object_type=METRICS_EXPORT). Values: AUTH_TOKEN, BASIC",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(platform.CreateEnvironmentObjectMetricsExportRequestAuthTypeAUTHTOKEN),
					string(platform.CreateEnvironmentObjectMetricsExportRequestAuthTypeBASIC),
				),
			},
		},
		"basic_token": resourceSchema.StringAttribute{
			MarkdownDescription: "The bearer token to connect to the remote endpoint (only valid when object_type=METRICS_EXPORT)",
			Optional:            true,
			Sensitive:           true,
		},
		"username": resourceSchema.StringAttribute{
			MarkdownDescription: "The username to connect to the remote endpoint (only valid when object_type=METRICS_EXPORT)",
			Optional:            true,
			Computed:            true,
		},
		"headers": resourceSchema.MapAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "HTTP request headers for the remote endpoint (only valid when object_type=METRICS_EXPORT)",
			Optional:            true,
			Computed:            true,
		},
		"labels": resourceSchema.MapAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Key-value pair metrics labels for your export (only valid when object_type=METRICS_EXPORT)",
			Optional:            true,
			Computed:            true,
		},
		// Polymorphic — applies to both CONNECTION and METRICS_EXPORT
		"password": resourceSchema.StringAttribute{
			MarkdownDescription: "The password — the connection password when object_type=CONNECTION, the HTTP Basic-auth password when object_type=METRICS_EXPORT",
			Optional:            true,
			Sensitive:           true,
		},
		// Links
		"links": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "The Deployments linked to the environment object. Only applicable for WORKSPACE scope",
			Optional:            true,
			Computed:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: environmentObjectLinkResourceSchemaAttributes(),
			},
		},
		"exclude_links": resourceSchema.SetNestedAttribute{
			MarkdownDescription: "The excluded links for the environment object. Only applicable for WORKSPACE scope",
			Optional:            true,
			Computed:            true,
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: environmentObjectExcludeLinkResourceSchemaAttributes(),
			},
		},
		// Metadata
		"created_at": resourceSchema.StringAttribute{
			MarkdownDescription: "Environment Object creation timestamp",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_at": resourceSchema.StringAttribute{
			MarkdownDescription: "Environment Object last updated timestamp",
			Computed:            true,
		},
		"created_by": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Environment Object creator",
			Computed:            true,
			Attributes:          ResourceSubjectProfileSchemaAttributes(),
			PlanModifiers: []planmodifier.Object{
				objectplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_by": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Environment Object updater",
			Computed:            true,
			Attributes:          ResourceSubjectProfileSchemaAttributes(),
		},
	}
}
