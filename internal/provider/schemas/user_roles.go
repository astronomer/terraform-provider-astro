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
					string(iam.UserOrganizationRoleORGANIZATIONOWNER),
					string(iam.UserOrganizationRoleORGANIZATIONMEMBER),
					string(iam.UserOrganizationRoleORGANIZATIONBILLINGADMIN),
					string(iam.UserOrganizationRoleORGANIZATIONOBSERVEADMIN),
					string(iam.UserOrganizationRoleORGANIZATIONOBSERVEMEMBER),
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
			MarkdownDescription: "The roles to assign to the deployments. Required for any deployment referenced in `dag_roles`.",
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
		},
		"dag_roles": resourceSchema.SetNestedAttribute{
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: ResourceDagRoleSchemaAttributes(),
			},
			Optional:            true,
			MarkdownDescription: "The DAG roles to assign to the user. Each role grants permissions to a specific DAG or DAGs with a specific tag within a deployment. Each deployment referenced in `dag_roles` must also have a corresponding entry in `deployment_roles` (e.g. with `DEPLOYMENT_ACCESSOR` role).",
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
		},
	}
}
