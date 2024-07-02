package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func UserDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"id": datasourceSchema.StringAttribute{
			MarkdownDescription: "User identifier",
			Required:            true,
			Validators:          []validator.String{validators.IsCuid()},
		},
		"username": datasourceSchema.StringAttribute{
			MarkdownDescription: "User username",
			Computed:            true,
		},
		"full_name": datasourceSchema.StringAttribute{
			MarkdownDescription: "User full name",
			Computed:            true,
		},
		"status": datasourceSchema.StringAttribute{
			MarkdownDescription: "User status",
			Computed:            true,
		},
		"avatar_url": datasourceSchema.StringAttribute{
			MarkdownDescription: "User avatar URL",
			Computed:            true,
		},
		"organization_role": datasourceSchema.StringAttribute{
			MarkdownDescription: "The role assigned to the organization",
			Computed:            true,
		},
		"workspace_roles": datasourceSchema.SetNestedAttribute{
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: DataSourceWorkspaceRoleSchemaAttributes(),
			},
			Computed:            true,
			MarkdownDescription: "The roles assigned to the workspaces",
		},
		"deployment_roles": datasourceSchema.SetNestedAttribute{
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: DataSourceDeploymentRoleSchemaAttributes(),
			},
			Computed:            true,
			MarkdownDescription: "The roles assigned to the deployments",
		},
		"created_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "User creation timestamp",
			Computed:            true,
		},
		"updated_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "User last updated timestamp",
			Computed:            true,
		},
	}
}
