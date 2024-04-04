package schemas

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func WorkspacesElementAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                    types.StringType,
		"name":                  types.StringType,
		"description":           types.StringType,
		"organization_name":     types.StringType,
		"cicd_enforced_default": types.BoolType,
		"created_at":            types.StringType,
		"updated_at":            types.StringType,
		"created_by": types.ObjectType{
			AttrTypes: SubjectProfileTF,
		},
		"updated_by": types.ObjectType{
			AttrTypes: SubjectProfileTF,
		},
	}
}

func WorkspacesDataSourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"workspaces": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: WorkspaceDataSourceSchemaAttributes(),
			},
			Computed: true,
		},
		"workspace_ids": schema.ListAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
		"names": schema.ListAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
	}
}
