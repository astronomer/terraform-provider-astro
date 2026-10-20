package schemas

import (
	platform_v1 "github.com/astronomer/terraform-provider-astro/internal/clients/platform_v1"
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ApiTokensElementAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":          types.StringType,
		"name":        types.StringType,
		"description": types.StringType,
		"short_token": types.StringType,
		"type":        types.StringType,
		"kind":        types.StringType,
		"start_at":    types.StringType,
		"end_at":      types.StringType,
		"created_at":  types.StringType,
		"updated_at":  types.StringType,
		"created_by": types.ObjectType{
			AttrTypes: SubjectProfileAttributeTypes(),
		},
		"updated_by": types.ObjectType{
			AttrTypes: SubjectProfileAttributeTypes(),
		},
		"expiry_period_in_days": types.Int64Type,
		"last_used_at":          types.StringType,
		"roles": types.SetType{
			ElemType: types.ObjectType{
				AttrTypes: ApiTokenRoleAttributeTypes(),
			},
		},
	}
}

func ApiTokensDataSourceSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"api_tokens": schema.SetNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: ApiTokenDataSourceSchemaAttributes(),
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
		"include_only_organization_tokens": schema.BoolAttribute{
			Optional: true,
		},
		"kind": schema.StringAttribute{
			MarkdownDescription: "Only list API tokens of this kind - 'STANDARD' or 'DIRECT_ACCESS'",
			Optional:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(platform_v1.ApiTokenKindSTANDARD),
					string(platform_v1.ApiTokenKindDIRECTACCESS),
				),
			},
		},
	}
}
