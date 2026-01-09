package schemas

import (
	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TeamDataSourceSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"id": datasourceSchema.StringAttribute{
			MarkdownDescription: "Team ID",
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
		"team_members": datasourceSchema.SetNestedAttribute{
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: DataSourceTeamMemberSchemaAttributes(),
			},
			MarkdownDescription: "The members of the Team",
			Computed:            true,
		},
		"is_idp_managed": datasourceSchema.BoolAttribute{
			MarkdownDescription: "Whether the Team is managed by an identity provider",
			Computed:            true,
		},
		"organization_role": datasourceSchema.StringAttribute{
			MarkdownDescription: "The role assigned to the Organization",
			Computed:            true,
		},
		"workspace_roles": datasourceSchema.SetNestedAttribute{
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: DataSourceWorkspaceRoleSchemaAttributes(),
			},
			Computed:            true,
			MarkdownDescription: "The roles assigned to the Workspaces",
		},
		"deployment_roles": datasourceSchema.SetNestedAttribute{
			NestedObject: datasourceSchema.NestedAttributeObject{
				Attributes: DataSourceDeploymentRoleSchemaAttributes(),
			},
			Computed:            true,
			MarkdownDescription: "The roles assigned to the Deployments",
		},
		"roles_count": datasourceSchema.Int64Attribute{
			MarkdownDescription: "Number of roles assigned to the Team",
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
			MarkdownDescription: "Team ID",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": resourceSchema.StringAttribute{
			MarkdownDescription: "Team name",
			Required:            true,
		},
		"description": resourceSchema.StringAttribute{
			MarkdownDescription: "Team description",
			Optional:            true,
		},
		"member_ids": resourceSchema.SetAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "The IDs of the users to add to the Team",
			Optional:            true,
			Validators: []validator.Set{
				setvalidator.ValueStringsAre(validators.IsCuid()),
			},
		},
		"is_idp_managed": resourceSchema.BoolAttribute{
			MarkdownDescription: "Whether the Team is managed by an identity provider",
			Computed:            true,
		},
		"organization_role": resourceSchema.StringAttribute{
			MarkdownDescription: "The role to assign to the Organization",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.OneOf(string(iam.TeamOrganizationRoleORGANIZATIONOWNER),
					string(iam.TeamOrganizationRoleORGANIZATIONMEMBER),
					string(iam.TeamOrganizationRoleORGANIZATIONBILLINGADMIN),
				),
			},
		},
		"workspace_roles": resourceSchema.SetNestedAttribute{
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: ResourceWorkspaceRoleSchemaAttributes(),
			},
			Optional:            true,
			MarkdownDescription: "The roles to assign to the Workspaces",
		},
		"deployment_roles": resourceSchema.SetNestedAttribute{
			NestedObject: resourceSchema.NestedAttributeObject{
				Attributes: ResourceDeploymentRoleSchemaAttributes(),
			},
			Optional:            true,
			MarkdownDescription: "The roles to assign to the Deployments",
		},
		"roles_count": resourceSchema.Int64Attribute{
			MarkdownDescription: "Number of roles assigned to the Team",
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

func TeamMemberAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"user_id":    types.StringType,
		"username":   types.StringType,
		"full_name":  types.StringType,
		"avatar_url": types.StringType,
		"created_at": types.StringType,
	}
}

func DataSourceTeamMemberSchemaAttributes() map[string]datasourceSchema.Attribute {
	return map[string]datasourceSchema.Attribute{
		"user_id": datasourceSchema.StringAttribute{
			MarkdownDescription: "The ID of the user in the Team",
			Computed:            true,
		},
		"username": datasourceSchema.StringAttribute{
			MarkdownDescription: "Team member username",
			Computed:            true,
		},
		"full_name": datasourceSchema.StringAttribute{
			MarkdownDescription: "Team member full name",
			Computed:            true,
		},
		"avatar_url": datasourceSchema.StringAttribute{
			MarkdownDescription: "Team member avatar URL",
			Computed:            true,
		},
		"created_at": datasourceSchema.StringAttribute{
			MarkdownDescription: "Team member creation timestamp",
			Computed:            true,
		},
	}
}
