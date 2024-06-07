package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Team describes the data source data model.
type Team struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	IsIdpManaged     types.Bool   `tfsdk:"is_idp_managed"`
	OrganizationRole types.String `tfsdk:"organization_role"`
	DeploymentRoles  types.Set    `tfsdk:"deployment_roles"`
	WorkspaceRoles   types.Set    `tfsdk:"workspace_roles"`
	RolesCount       types.Int64  `tfsdk:"roles_count"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
	CreatedBy        types.Object `tfsdk:"created_by"`
	UpdatedBy        types.Object `tfsdk:"updated_by"`
}

func (data *Team) ReadFromResponse(ctx context.Context, team *iam.Team) diag.Diagnostics {
	var diags diag.Diagnostics
	data.Id = types.StringValue(team.Id)
	data.Name = types.StringValue(team.Name)
	if team.Description != nil {
		data.Description = types.StringValue(*team.Description)
	} else {
		data.Description = types.StringValue("")
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
