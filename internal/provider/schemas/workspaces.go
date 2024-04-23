package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func WorkspacesElementAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                    types.StringType,
		"name":                  types.StringType,
		"description":           types.StringType,
		"cicd_enforced_default": types.BoolType,
		"created_at":            types.StringType,
		"updated_at":            types.StringType,
		"created_by": types.ObjectType{
			AttrTypes: SubjectProfileAttributeTypes(),
		},
		"updated_by": types.ObjectType{
			AttrTypes: SubjectProfileAttributeTypes(),
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
			Validators: []validator.List{
				listvalidator.ValueStringsAre(validators.IsCuid()),
				listvalidator.UniqueValues(),
			},
			Optional: true,
		},
		"names": schema.ListAttribute{
			ElementType: types.StringType,
			Validators: []validator.List{
				listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				listvalidator.UniqueValues(),
			},
			Optional: true,
		},
	}
}
