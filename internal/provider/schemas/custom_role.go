package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func CustomRoleAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                       types.StringType,
		"name":                     types.StringType,
		"description":              types.StringType,
		"permissions":              types.SetType{ElemType: types.StringType},
		"scope_type":               types.StringType,
		"restricted_workspace_ids": types.SetType{ElemType: types.StringType},
		"created_at":               types.StringType,
		"created_by":               types.ObjectType{AttrTypes: SubjectProfileAttributeTypes()},
		"updated_at":               types.StringType,
		"updated_by":               types.ObjectType{AttrTypes: SubjectProfileAttributeTypes()},
	}
}

func CustomRoleDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Custom role identifier",
			Required:            true,
			Validators:          []validator.String{validators.IsCuid()},
		},
		"name": datasourceSchema.StringAttribute{
			MarkdownDescription: "The custom role's name",
			Computed:            true,
		},
		"description": datasourceSchema.StringAttribute{
			MarkdownDescription: "The custom role's description",
			Computed:            true,
		},
		"permissions": datasourceSchema.SetAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "The custom role's permissions",
			Computed:            true,
		},
		"scope_type": datasourceSchema.StringAttribute{
			MarkdownDescription: "The custom role's scope (DEPLOYMENT, ORGANIZATION, or WORKSPACE)",
			Computed:            true,
		},
		"restricted_workspace_ids": datasourceSchema.SetAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "The IDs of Workspaces that the custom role is restricted to",
			Computed:            true,
		},
		"created_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "The time the custom role was created",
			Computed:            true,
		},
		"created_by": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The subject who created the custom role",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchemaAttributes(),
		},
		"updated_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "The time the custom role was last updated",
			Computed:            true,
		},
		"updated_by": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "The subject who last updated the custom role",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchemaAttributes(),
		},
	}
}
