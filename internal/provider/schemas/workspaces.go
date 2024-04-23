package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
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
		"workspaces": schema.SetNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: WorkspaceDataSourceSchemaAttributes(),
			},
			Computed: true,
		},
		"workspace_ids": schema.SetAttribute{
			ElementType: types.StringType,
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(validators.IsCuid()),
			},
			Optional: true,
		},
		"names": schema.SetAttribute{
			ElementType: types.StringType,
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
			Optional: true,
		},
	}
}
