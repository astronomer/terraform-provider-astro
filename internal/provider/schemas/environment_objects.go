package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/clients/platform"
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// EnvironmentObjectsElementAttributeTypes returns the attribute type map for a
// single element of the environment_objects list. Mirrors the flat shape of
// EnvironmentObjectDataSourceSchemaAttributes — type-specific fields sit at the
// top level, populated according to object_type.
func EnvironmentObjectsElementAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		// Identity / common
		"id":                     types.StringType,
		"object_key":             types.StringType,
		"object_type":            types.StringType,
		"scope":                  types.StringType,
		"scope_entity_id":        types.StringType,
		"source_scope":           types.StringType,
		"source_scope_entity_id": types.StringType,
		"auto_link_deployments":  types.BoolType,
		// AIRFLOW_VARIABLE
		"value":     types.StringType,
		"is_secret": types.BoolType,
		// CONNECTION
		"type":         types.StringType,
		"host":         types.StringType,
		"port":         types.Int64Type,
		"schema":       types.StringType,
		"login":        types.StringType,
		"extra":        types.StringType,
		"auth_type_id": types.StringType,
		"connection_auth_type": types.ObjectType{
			AttrTypes: EnvironmentObjectConnectionAuthTypeAttributeTypes(),
		},
		// METRICS_EXPORT
		"auth_type":     types.StringType,
		"endpoint":      types.StringType,
		"basic_token":   types.StringType,
		"exporter_type": types.StringType,
		"username":      types.StringType,
		"headers":       types.MapType{ElemType: types.StringType},
		"labels":        types.MapType{ElemType: types.StringType},
		// Polymorphic
		"password": types.StringType,
		// Links
		"links":         types.SetType{ElemType: types.ObjectType{AttrTypes: EnvironmentObjectLinkAttributeTypes()}},
		"exclude_links": types.SetType{ElemType: types.ObjectType{AttrTypes: EnvironmentObjectExcludeLinkAttributeTypes()}},
		// Metadata
		"created_at": types.StringType,
		"updated_at": types.StringType,
		"created_by": types.ObjectType{AttrTypes: SubjectProfileAttributeTypes()},
		"updated_by": types.ObjectType{AttrTypes: SubjectProfileAttributeTypes()},
	}
}

func EnvironmentObjectsDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"workspace_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Filter by Workspace ID",
			Optional:            true,
			Validators:          []validator.String{validators.IsCuid()},
		},
		"deployment_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Filter by Deployment ID",
			Optional:            true,
			Validators:          []validator.String{validators.IsCuid()},
		},
		"object_type": datasourceSchema.StringAttribute{
			MarkdownDescription: "Filter by object type (AIRFLOW_VARIABLE, CONNECTION, METRICS_EXPORT)",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(platform.CreateEnvironmentObjectRequestObjectTypeAIRFLOWVARIABLE),
					string(platform.CreateEnvironmentObjectRequestObjectTypeCONNECTION),
					string(platform.CreateEnvironmentObjectRequestObjectTypeMETRICSEXPORT),
				),
			},
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
