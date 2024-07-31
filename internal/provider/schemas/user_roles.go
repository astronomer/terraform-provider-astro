package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func ResourceUserRolesSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"user_id": resourceSchema.StringAttribute{
			MarkdownDescription: "The ID of the user to assign the roles to",
			Required:            true,
			Validators: []validator.String{
				validators.IsCuid(),
			},
		},
		"organization_role": resourceSchema.StringAttribute{
			MarkdownDescription: "The role to assign to the organization",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(
					string(iam.ORGANIZATIONOWNER),
					string(iam.ORGANIZATIONMEMBER),
					string(iam.ORGANIZATIONBILLINGADMIN),
				),
			},
		},
		"workspace_roles": resourceSchema.SetNestedAttribute{
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: ResourceWorkspaceRoleSchemaAttributes(),
			},
			Optional:            true,
			MarkdownDescription: "The roles to assign to the workspaces",
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
		},
		"deployment_roles": resourceSchema.SetNestedAttribute{
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: ResourceDeploymentRoleSchemaAttributes(),
			},
			Optional:            true,
			MarkdownDescription: "The roles to assign to the deployments",
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
		},
	}
}
