package schemas

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TeamsElementAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                types.StringType,
		"name":              types.StringType,
		"description":       types.StringType,
		"is_idp_managed":    types.BoolType,
		"organization_id":   types.StringType,
		"organization_role": types.StringType,
		"deployment_roles": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: DeploymentRoleAttributeTypes(),
			},
		},
		"workspace_roles": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: WorkspaceRoleAttributeTypes(),
			},
		},
		"roles_count": types.Int64Type,
		"created_at":  types.StringType,
		"updated_at":  types.StringType,
		"created_by": types.ObjectType{
			AttrTypes: SubjectProfileAttributeTypes(),
		},
		"updated_by": types.ObjectType{
			AttrTypes: SubjectProfileAttributeTypes(),
		},
	}
}

func TeamsDataSourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"teams": schema.SetNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: TeamDataSourceSchemaAttributes(),
			},
			Computed: true,
		},
		"names": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
			},
		},
	}
}
