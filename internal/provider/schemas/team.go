package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
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
		"roles_count": datasourceSchema.Int64Attribute{
			MarkdownDescription: "Number of roles assigned to the team",
			Computed:            true,
		},
		"created_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "Team creation timestamp",
			Computed:            true,
		},
		"updated_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "Team last updated timestamp",
			Computed:            true,
		},
		"created_by": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Team creator",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchemaAttributes(),
		},
		"updated_by": datasourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Team updater",
			Computed:            true,
			Attributes:          DataSourceSubjectProfileSchemaAttributes(),
		},
	}
}

func TeamResourceSchemaAttributes() map[string]resourceSchema.Attribute {
	return map[string]resourceSchema.Attribute{
		"id": resourceSchema.StringAttribute{
			MarkdownDescription: "Team identifier",
			Required:            true,
			Validators: []validator.String{
				validators.IsCuid(),
			},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Team name",
			Optional:            true,
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Team description",
			Optional:            true,
		},
		"member_ids": resourceSchema.SetAttribute{
			MarkdownDescription: "The IDs of the users to add to the Team",
			Optional:            true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
		},
		"is_idp_managed": resourceSchema.BoolAttribute{
			MarkdownDescription: "Whether the team is managed by an identity provider",
			Computed:            true,
		},
		"organization_role": resourceSchema.StringAttribute{
			MarkdownDescription: "The role assigned to the organization",
			Computed:            true,
		},
		"workspace_roles": resourceSchema.SetNestedAttribute{
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: ResourceWorkspaceRoleSchemaAttributes(),
			},
			Computed:            true,
			MarkdownDescription: "The roles assigned to the workspaces",
		},
		"deployment_roles": resourceSchema.SetNestedAttribute{
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: ResourceDeploymentRoleSchemaAttributes(),
			},
			Computed:            true,
			MarkdownDescription: "The roles assigned to the deployments",
		},
		"roles_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Number of roles assigned to the team",
			Computed:            true,
		},
		"created_at": resourceSchema.StringAttribute{
			MarkdownDescription: "Team creation timestamp",
			Computed:            true,
		},
		"updated_at": resourceSchema.StringAttribute{
			MarkdownDescription: "Team last updated timestamp",
			Computed:            true,
		},
		"created_by": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Team creator",
			Computed:            true,
			Attributes:          ResourceSubjectProfileSchemaAttributes(),
		},
		"updated_by": resourceSchema.SingleNestedAttribute{
			MarkdownDescription: "Team updater",
			Computed:            true,
			Attributes:          ResourceSubjectProfileSchemaAttributes(),
		},
	}
}
