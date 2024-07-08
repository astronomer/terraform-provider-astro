package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func UsersElementAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                types.StringType,
		"username":          types.StringType,
		"full_name":         types.StringType,
		"status":            types.StringType,
		"avatar_url":        types.StringType,
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
		"created_at": types.StringType,
		"updated_at": types.StringType,
	}
}

func UsersDataSourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"users": schema.SetNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: UserDataSourceSchemaAttributes(),
			},
			Computed: true,
		},
		"workspace_id": schema.StringAttribute{
			Optional:   true,
			Validators: []validator.String{validators.IsCuid()},
		},
		"deployment_id": schema.StringAttribute{
			Optional:   true,
			Validators: []validator.String{validators.IsCuid()},
		},
	}
}
