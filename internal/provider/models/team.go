package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TeamDataSource describes the data source data model.
type TeamDataSource struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	IsIdpManaged     types.Bool   `tfsdk:"is_idp_managed"`
	TeamMembers      types.Set    `tfsdk:"team_members"`
	OrganizationRole types.String `tfsdk:"organization_role"`
	DeploymentRoles  types.Set    `tfsdk:"deployment_roles"`
	WorkspaceRoles   types.Set    `tfsdk:"workspace_roles"`
	RolesCount       types.Int64  `tfsdk:"roles_count"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
	CreatedBy        types.Object `tfsdk:"created_by"`
	UpdatedBy        types.Object `tfsdk:"updated_by"`
}

type TeamResource struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	IsIdpManaged     types.Bool   `tfsdk:"is_idp_managed"`
	MemberIds        types.Set    `tfsdk:"member_ids"`
	OrganizationRole types.String `tfsdk:"organization_role"`
	DeploymentRoles  types.Set    `tfsdk:"deployment_roles"`
	WorkspaceRoles   types.Set    `tfsdk:"workspace_roles"`
	RolesCount       types.Int64  `tfsdk:"roles_count"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
	CreatedBy        types.Object `tfsdk:"created_by"`
	UpdatedBy        types.Object `tfsdk:"updated_by"`
}

func (data *TeamDataSource) ReadFromResponse(ctx context.Context, team *iam.Team, teamMembers *[]iam.TeamMember) diag.Diagnostics {
	var diags diag.Diagnostics
	data.Id = types.StringValue(team.Id)
	data.Name = types.StringValue(team.Name)
	if team.Description != nil {
		data.Description = types.StringValue(*team.Description)
	} else {
		data.Description = types.StringValue("")
	}
	data.TeamMembers, diags = utils.ObjectSet(ctx, teamMembers, schemas.TeamMemberAttributeTypes(), TeamMemberTypesObject)
	if diags.HasError() {
		return diags
	}
	data.IsIdpManaged = types.BoolValue(team.IsIdpManaged)
	data.OrganizationRole = types.StringValue(string(team.OrganizationRole))
	data.DeploymentRoles, diags = utils.ObjectSet(ctx, team.DeploymentRoles, schemas.DeploymentRoleAttributeTypes(), DeploymentRoleTypesObject)
	if diags.HasError() {
		return diags
	}
	data.WorkspaceRoles, diags = utils.ObjectSet(ctx, team.WorkspaceRoles, schemas.WorkspaceRoleAttributeTypes(), WorkspaceRoleTypesObject)
	if diags.HasError() {
		return diags
	}
	if team.RolesCount != nil {
		data.RolesCount = types.Int64Value(int64(*team.RolesCount))
	} else {
		data.RolesCount = types.Int64Value(0)
	}

	data.CreatedAt = types.StringValue(team.CreatedAt.String())
	data.UpdatedAt = types.StringValue(team.UpdatedAt.String())
	data.CreatedBy, diags = SubjectProfileTypesObject(ctx, team.CreatedBy)
	if diags.HasError() {
		return diags
	}
	data.UpdatedBy, diags = SubjectProfileTypesObject(ctx, team.UpdatedBy)
	if diags.HasError() {
		return diags
	}

	return nil
}

func (data *TeamResource) ReadFromResponse(ctx context.Context, team *iam.Team, memberIds *[]string) diag.Diagnostics {
	var diags diag.Diagnostics
	data.Id = types.StringValue(team.Id)
	data.Name = types.StringValue(team.Name)
	if team.Description != nil && *team.Description != "" {
		data.Description = types.StringValue(*team.Description)
	} else {
		data.Description = types.StringNull()
	}
	data.MemberIds, diags = utils.StringSet(memberIds)
	if diags.HasError() {
		return diags
	}
	data.IsIdpManaged = types.BoolValue(team.IsIdpManaged)
	data.OrganizationRole = types.StringValue(string(team.OrganizationRole))
	data.DeploymentRoles, diags = utils.ObjectSet(ctx, team.DeploymentRoles, schemas.DeploymentRoleAttributeTypes(), DeploymentRoleTypesObject)
	if diags.HasError() {
		return diags
	}
	data.WorkspaceRoles, diags = utils.ObjectSet(ctx, team.WorkspaceRoles, schemas.WorkspaceRoleAttributeTypes(), WorkspaceRoleTypesObject)
	if diags.HasError() {
		return diags
	}
	if team.RolesCount != nil {
		data.RolesCount = types.Int64Value(int64(*team.RolesCount))
	} else {
		data.RolesCount = types.Int64Value(0)
	}

	data.CreatedAt = types.StringValue(team.CreatedAt.String())
	data.UpdatedAt = types.StringValue(team.UpdatedAt.String())
	data.CreatedBy, diags = SubjectProfileTypesObject(ctx, team.CreatedBy)
	if diags.HasError() {
		return diags
	}
	data.UpdatedBy, diags = SubjectProfileTypesObject(ctx, team.UpdatedBy)
	if diags.HasError() {
		return diags
	}

	return nil
}

type TeamMember struct {
	UserId    types.String `tfsdk:"user_id"`
	Username  types.String `tfsdk:"username"`
	FullName  types.String `tfsdk:"full_name"`
	AvatarUrl types.String `tfsdk:"avatar_url"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func TeamMemberTypesObject(ctx context.Context, member iam.TeamMember) (types.Object, diag.Diagnostics) {
	obj := TeamMember{
		UserId:    types.StringValue(member.UserId),
		Username:  types.StringValue(member.Username),
		FullName:  types.StringValue(*member.FullName),
		AvatarUrl: types.StringValue(*member.AvatarUrl),
		CreatedAt: types.StringValue(member.CreatedAt.String()),
	}
	return types.ObjectValueFrom(ctx, schemas.TeamMemberAttributeTypes(), obj)
}
