package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func CustomRoleResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "The name of the custom role",
			Required:            true,
		},
		"scope_type": resourceSchema.StringAttribute{
			MarkdownDescription: "The scope the custom role can be used in - For now Deployment is the only available option",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(string(iam.DEPLOYMENT)),
			},
		},
		"restricted_workspace_ids": resourceSchema.ListAttribute{
			MarkdownDescription: "Optional list of workspaces this custom role can be used in",
			Required:            false,
			ElementType:         types.StringType,
			Validators: []validator.List{
				listvalidator.ValueStringsAre(validators.IsCuid()),
			},
		},
		"permissions": resourceSchema.ListAttribute{
			MarkdownDescription: "The permissions that this custom role is enabled with",
			Computed:            true,
			ElementType:         types.StringType,
		},
	}
}
