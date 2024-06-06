package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func TeamDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Team identifier",
			Required:            true,
			Validators:          []validator.String{validators.IsCuid()},
		},
		"name": datasourceSchema.StringAttribute{
			MarkdownDescription: "Team name",
			Computed:            true,
		},
		"description": datasourceSchema.StringAttribute{
			MarkdownDescription: "Team description",
			Computed:            true,
		},
		"is_idp_managed": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Whether the team is managed by an identity provider",
			Computed:            true,
		},
		"organization_role": datasourceSchema.StringAttribute{
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
		"created_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "Workspace creation timestamp",
			Computed:            true,
		},
		"updated_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "Workspace last updated timestamp",
			Computed:            true,
		},
		"created_by": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Workspace creator",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchemaAttributes(),
		},
		"updated_by": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Workspace updater",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchemaAttributes(),
		},
	}
}
