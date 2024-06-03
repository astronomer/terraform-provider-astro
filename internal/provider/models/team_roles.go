package models

import (
	"context"

	"github.com/astronomer/terraform-provider-astro/internal/clients/iam"
	"github.com/astronomer/terraform-provider-astro/internal/provider/schemas"
	"github.com/astronomer/terraform-provider-astro/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TeamRoles describes the team_roles resource
type TeamRoles struct {
	TeamId           types.String `tfsdk:"team_id"`
	OrganizationRole types.String `tfsdk:"organization_role"`
	WorkspaceRoles   types.Set    `tfsdk:"workspace_roles"`
	DeploymentRoles  types.Set    `tfsdk:"deployment_roles"`
}

func (data *TeamRoles) ReadFromResponse(
	ctx context.Context,
	teamId string,
	teamRoles *iam.SubjectRoles,
) diag.Diagnostics {
	var diags diag.Diagnostics
	data.TeamId = types.StringValue(teamId)
	data.OrganizationRole = types.StringPointerValue((*string)(teamRoles.OrganizationRole))
	data.WorkspaceRoles, diags = utils.ObjectSet(ctx, teamRoles.WorkspaceRoles, schemas.WorkspaceRoleAttributeTypes(), WorkspaceRoleTypesObject)
	if diags.HasError() {
		return diags
	}
	data.DeploymentRoles, diags = utils.ObjectSet(ctx, teamRoles.DeploymentRoles, schemas.DeploymentRoleAttributeTypes(), DeploymentRoleTypesObject)
	if diags.HasError() {
		return diags
	}
	return nil
}
